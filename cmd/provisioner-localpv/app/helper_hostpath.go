/*
Copyright 2019 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This code was taken from https://github.com/rancher/local-path-provisioner
and modified to work with the configuration options used by OpenEBS
*/

package app

import (
	"context"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	hostpath "github.com/openebs/maya/pkg/hostpath/v1alpha1"
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/container"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/pod"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/volume"
)

type podConfig struct {
	pOpts                         *HelperPodOptions
	parentDir, volumeDir, podName string
	taints                        []corev1.Taint
}

var (
	//CmdTimeoutCounts specifies the duration to wait for cleanup pod
	//to be launched.
	CmdTimeoutCounts = 120
)

// HelperPodOptions contains the options that
// will launch a Pod on a specific node (nodeHostname)
// to execute a command (cmdsForPath) on a given
// volume path (path)
type HelperPodOptions struct {
	//nodeAffinityLabels represents the labels of the node where pod should be launched.
	nodeAffinityLabels map[string]string

	//name is the name of the PV for which the pod is being launched
	name string

	//cmdsForPath represent either create (mkdir) or delete(rm)
	//commands that need to be executed on the volume path.
	cmdsForPath []string

	//path is the volume hostpath directory
	path string

	//serviceAccountName is the service account with which the pod should be launched
	serviceAccountName string

	selectedNodeTaints []corev1.Taint

	imagePullSecrets []corev1.LocalObjectReference

	//softLimitGrace is the soft limit of quota on the project
	softLimitGrace string

	//hardLimitGrace is the hard limit of quota on the project
	hardLimitGrace string

	//pvcStorage is the storage requested for pv
	pvcStorage int64
}

// validate checks that the required fields to launch
// helper pods are valid. helper pods are used to either
// create or delete a directory (path) on a given node hostname (nodeHostname).
// name refers to the volume being created or deleted.
func (pOpts *HelperPodOptions) validate() error {
	if pOpts.name == "" ||
		pOpts.path == "" ||
		pOpts.nodeAffinityLabels == nil ||
		len(pOpts.nodeAffinityLabels) == 0 ||
		pOpts.serviceAccountName == "" {
		return errors.Errorf("invalid empty name or hostpath or hostname or service account name")
	}
	return nil
}

// validateLimits check that the limits to setup qouta are valid
func (pOpts *HelperPodOptions) validateLimits() error {
	if pOpts.softLimitGrace == "0k" &&
		pOpts.hardLimitGrace == "0k" {
		// Hack: using convertToK() style converstion
		// TODO: Refactor this section of the code
		pvcStorageInK := math.Ceil(float64(pOpts.pvcStorage) / 1024)
		pvcStorageInKString := strconv.FormatFloat(pvcStorageInK, 'f', -1, 64) + "k"
		pOpts.softLimitGrace, pOpts.hardLimitGrace = pvcStorageInKString, pvcStorageInKString
		return nil
	}

	if pOpts.softLimitGrace == "0k" ||
		pOpts.hardLimitGrace == "0k" {
		return nil
	}

	if len(pOpts.softLimitGrace) > len(pOpts.hardLimitGrace) ||
		(len(pOpts.softLimitGrace) == len(pOpts.hardLimitGrace) &&
			pOpts.softLimitGrace > pOpts.hardLimitGrace) {
		return errors.Errorf("hard limit cannot be smaller than soft limit")
	}

	return nil
}

// converToK converts the limits to kilobytes
func convertToK(limit string, pvcStorage int64) (string, error) {

	if len(limit) == 0 {
		return "0k", nil
	}

	valueRegex := regexp.MustCompile(`[\d]*[\.]?[\d]*`)
	valueString := valueRegex.FindString(limit)

	if limit != valueString+"%" {
		return "", errors.Errorf("invalid format for limit grace")
	}

	value, err := strconv.ParseFloat(valueString, 64)

	if err != nil {
		return "", errors.Errorf("invalid format, cannot parse")
	}
	if value > 100 {
		value = 100
	}

	value *= float64(pvcStorage)
	value /= 100
	value += float64(pvcStorage)
	value /= 1024

	value = math.Ceil(value)
	valueString = strconv.FormatFloat(value, 'f', -1, 64)
	valueString += "k"
	return valueString, nil
}

// createInitPod launches a helper(busybox) pod, to create the host path.
//
//	The local pv expect the hostpath to be already present before mounting
//	into pod. Validate that the local pv host path is not created under root.
func (p *Provisioner) createInitPod(ctx context.Context, pOpts *HelperPodOptions) error {
	var config podConfig
	config.pOpts, config.podName = pOpts, "init"
	//err := pOpts.validate()
	if err := pOpts.validate(); err != nil {
		return err
	}

	// Initialize HostPath builder and validate that
	// volume directory is not directly under root.
	// Extract the base path and the volume unique path.
	var vErr error
	config.parentDir, config.volumeDir, vErr = hostpath.NewBuilder().WithPath(pOpts.path).
		WithCheckf(hostpath.IsNonRoot(), "volume directory {%v} should not be under root directory", pOpts.path).
		ExtractSubPath()
	if vErr != nil {
		return vErr
	}

	//Pass on the taints, to create tolerations.
	config.taints = pOpts.selectedNodeTaints

	config.pOpts.cmdsForPath = append(config.pOpts.cmdsForPath, filepath.Join("/data/", config.volumeDir))

	iPod, err := p.launchPod(ctx, config)
	if err != nil {
		return err
	}

	if err := p.exitPod(ctx, iPod); err != nil {
		return err
	}

	return nil
}

// createCleanupPod launches a helper(busybox) pod, to delete the host path.
//
//	This provisioner expects that the host paths are created using
//	an unique PV path - under a given BasePath. From the absolute path,
//	it extracts the base path and the PV path. The helper pod is then launched
//	by mounting the base path - and performing a delete on the unique PV path.
func (p *Provisioner) createCleanupPod(ctx context.Context, pOpts *HelperPodOptions) error {
	var config podConfig
	config.pOpts, config.podName = pOpts, "cleanup"
	//err := pOpts.validate()
	if err := pOpts.validate(); err != nil {
		return err
	}

	// Initialize HostPath builder and validate that
	// volume directory is not directly under root.
	// Extract the base path and the volume unique path.
	var vErr error
	config.parentDir, config.volumeDir, vErr = hostpath.NewBuilder().WithPath(pOpts.path).
		WithCheckf(hostpath.IsNonRoot(), "volume directory {%v} should not be under root directory", pOpts.path).
		ExtractSubPath()
	if vErr != nil {
		return vErr
	}

	config.taints = pOpts.selectedNodeTaints

	config.pOpts.cmdsForPath = append(config.pOpts.cmdsForPath, filepath.Join("/data/", config.volumeDir))

	cPod, err := p.launchPod(ctx, config)
	if err != nil {
		return err
	}

	if err := p.exitPod(ctx, cPod); err != nil {
		return err
	}
	return nil
}

// createQuotaPod launches a helper(busybox) pod, to apply the quota.
//
//	The local pv expect the hostpath to be already present before mounting
//	into pod. Validate that the local pv host path is not created under root.
func (p *Provisioner) createQuotaPod(ctx context.Context, pOpts *HelperPodOptions) error {
	var config podConfig
	config.pOpts, config.podName = pOpts, "quota"
	//err := pOpts.validate()
	if err := pOpts.validate(); err != nil {
		return err
	}

	// Initialize HostPath builder and validate that
	// volume directory is not directly under root.
	// Extract the base path and the volume unique path.
	var vErr error
	config.parentDir, config.volumeDir, vErr = hostpath.NewBuilder().WithPath(pOpts.path).
		WithCheckf(hostpath.IsNonRoot(), "volume directory {%v} should not be under root directory", pOpts.path).
		ExtractSubPath()
	if vErr != nil {
		return vErr
	}

	//Pass on the taints, to create tolerations.
	config.taints = pOpts.selectedNodeTaints

	var lErr error
	config.pOpts.softLimitGrace, lErr = convertToK(config.pOpts.softLimitGrace, config.pOpts.pvcStorage)
	if lErr != nil {
		return lErr
	}
	config.pOpts.hardLimitGrace, lErr = convertToK(config.pOpts.hardLimitGrace, config.pOpts.pvcStorage)
	if lErr != nil {
		return lErr
	}

	if err := pOpts.validateLimits(); err != nil {
		return err
	}

	//fs stores the file system of mount
	fs := "FS=`stat -f -c %T /data` ; "
	//check if fs is xfs or ext4 (output of stat is ext2/ext3)
	//PID is the last project Id in the directory
	//xfs_quota project(xfs) or chattr +P (ext4) initializes project with new project id
	//xfs_quota limit(xfs) or repquota (ext4) sets the quota according to limits defined
	checkQuota := "" +
		"if [[ \"$FS\" == \"xfs\" ]]; then " +
		"  PID=`xfs_quota -x -c 'report -h' /data | tail -2 | awk 'NR==1{print substr ($1,2)}+0'` ;" +
		"  PID=`expr $PID + 1` ;" +
		"  xfs_quota -x -c 'project -s -p " + filepath.Join("/data/", config.volumeDir) + " '$PID /data;" +
		"  xfs_quota -x -c 'limit -p bsoft=" + config.pOpts.softLimitGrace + " bhard=" + config.pOpts.hardLimitGrace + " '$PID /data ;" +
		"elif [[ \"$FS\" == \"ext2/ext3\" ]]; then" +
		"  PID=`repquota -P /data | tail -3 | awk 'NR==1{print substr ($1,2)}+0'` ;" +
		"  PID=`expr $PID + 1` ;" +
		"  chattr +P -p $PID " + filepath.Join("/data/", config.volumeDir) + " ;" +
		"  setquota -P $PID " + strings.ToUpper(config.pOpts.softLimitGrace) + " " + strings.ToUpper(config.pOpts.hardLimitGrace) + " 0 0 " + "/data ; " +
		"else " +
		"  rm -rf " + filepath.Join("/data/", config.volumeDir) + " ; exit 1; fi"
	config.pOpts.cmdsForPath = []string{"sh", "-c", fs + checkQuota}

	qPod, err := p.launchPod(ctx, config)
	if err != nil {
		return err
	}

	if err := p.exitPod(ctx, qPod); err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) launchPod(ctx context.Context, config podConfig) (*corev1.Pod, error) {
	// the helper pod need to be launched in privileged mode. This is because in CoreOS
	// nodes, pods without privileged access cannot write to the host directory.
	// Helper pods need to create and delete directories on the host.
	privileged := true

	helperPod, err := pod.NewBuilder().
		WithName(config.podName + "-" + config.pOpts.name).
		WithRestartPolicy(corev1.RestartPolicyNever).
		//WithNodeSelectorHostnameNew(config.pOpts.nodeHostname).
		WithNodeAffinityNew(config.pOpts.nodeAffinityLabels).
		WithServiceAccountName(config.pOpts.serviceAccountName).
		WithTolerationsForTaints(config.taints...).
		WithContainerBuilder(
			container.NewBuilder().
				WithName("local-path-" + config.podName).
				WithImage(p.helperImage).
				WithCommandNew(config.pOpts.cmdsForPath).
				WithVolumeMountsNew([]corev1.VolumeMount{
					{
						Name:      "data",
						ReadOnly:  false,
						MountPath: "/data/",
					},
					{
						Name:      "dev",
						ReadOnly:  false,
						MountPath: "/dev/",
					},
				}).
				WithPrivilegedSecurityContext(&privileged),
		).
		WithImagePullSecrets(config.pOpts.imagePullSecrets).
		WithVolumeBuilder(
			volume.NewBuilder().
				WithName("data").
				WithHostDirectory(config.parentDir),
		).
		WithVolumeBuilder(
			volume.NewBuilder().
				WithName("dev").
				WithHostDirectory("/dev/"),
		).
		Build()

	if err != nil {
		return nil, err
	}

	var hPod *corev1.Pod

	//Launch the helper pod.
	hPod, err = p.kubeClient.CoreV1().Pods(p.namespace).Create(ctx, helperPod, metav1.CreateOptions{})
	return hPod, err
}

func (p *Provisioner) exitPod(ctx context.Context, hPod *corev1.Pod) error {
	defer func() {
		e := p.kubeClient.CoreV1().Pods(p.namespace).Delete(ctx, hPod.Name, metav1.DeleteOptions{})
		if e != nil {
			klog.Errorf("unable to delete the helper pod: %v", e)
		}
	}()

	//Wait for the helper pod to complete it job and exit
	completed := false
	for i := 0; i < CmdTimeoutCounts; i++ {
		checkPod, err := p.kubeClient.CoreV1().Pods(p.namespace).Get(ctx, hPod.Name, metav1.GetOptions{})
		if err != nil {
			return err
		} else if checkPod.Status.Phase == corev1.PodSucceeded {
			completed = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	if !completed {
		return errors.Errorf("create process timeout after %v seconds", CmdTimeoutCounts)
	}
	return nil
}
