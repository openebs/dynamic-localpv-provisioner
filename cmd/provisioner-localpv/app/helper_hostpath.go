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
	"path/filepath"
	"time"

	errors "github.com/pkg/errors"
	"k8s.io/klog"

	hostpath "github.com/openebs/maya/pkg/hostpath/v1alpha1"

	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/container"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/pod"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/volume"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	//nodeAffinityLabelKey represents the label key of the node where pod should be launched.
	nodeAffinityLabelKey string

	//nodeAffinityLabelValue represents the label value of the node where pod should be launched.
	nodeAffinityLabelValue string

	//name is the name of the PV for which the pod is being launched
	name string

	//cmdsForPath represent either create (mkdir) or delete(rm)
	//commands that need to be executed on the volume path.
	cmdsForPath []string

	//path is the volume hostpath directory
	path string

	// serviceAccountName is the service account with which the pod should be launched
	serviceAccountName string

	selectedNodeTaints []corev1.Taint

	imagePullSecrets []corev1.LocalObjectReference
}

// validate checks that the required fields to launch
// helper pods are valid. helper pods are used to either
// create or delete a directory (path) on a given node hostname (nodeHostname).
// name refers to the volume being created or deleted.
func (pOpts *HelperPodOptions) validate() error {
	if pOpts.name == "" ||
		pOpts.path == "" ||
		pOpts.nodeAffinityLabelKey == "" ||
		pOpts.nodeAffinityLabelValue == "" ||
		pOpts.serviceAccountName == "" {
		return errors.Errorf("invalid empty name or hostpath or hostname or service account name")
	}
	return nil
}

// createInitPod launches a helper(busybox) pod, to create the host path.
//  The local pv expect the hostpath to be already present before mounting
//  into pod. Validate that the local pv host path is not created under root.
func (p *Provisioner) createInitPod(pOpts *HelperPodOptions) error {
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

	iPod, err := p.launchPod(config)
	if err != nil {
		return err
	}

	if err := p.exitPod(iPod); err != nil {
		return err
	}

	return nil
}

// createCleanupPod launches a helper(busybox) pod, to delete the host path.
//  This provisioner expects that the host paths are created using
//  an unique PV path - under a given BasePath. From the absolute path,
//  it extracts the base path and the PV path. The helper pod is then launched
//  by mounting the base path - and performing a delete on the unique PV path.
func (p *Provisioner) createCleanupPod(pOpts *HelperPodOptions) error {
	var config podConfig
	config.pOpts, config.podName = pOpts, "cleanup"
	//err := pOpts.validate()
	if err := pOpts.validate(); err != nil {
		return err
	}

	config.taints = pOpts.selectedNodeTaints
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

	cPod, err := p.launchPod(config)
	if err != nil {
		return err
	}

	if err := p.exitPod(cPod); err != nil {
		return err
	}
	return nil
}

func (p *Provisioner) launchPod(config podConfig) (*corev1.Pod, error) {
	// the helper pod need to be launched in privileged mode. This is because in CoreOS
	// nodes, pods without privileged access cannot write to the host directory.
	// Helper pods need to create and delete directories on the host.
	privileged := true

	helperPod, err := pod.NewBuilder().
		WithName(config.podName+"-"+config.pOpts.name).
		WithRestartPolicy(corev1.RestartPolicyNever).
		//WithNodeSelectorHostnameNew(config.pOpts.nodeHostname).
		WithNodeAffinityNew(config.pOpts.nodeAffinityLabelKey, config.pOpts.nodeAffinityLabelValue).
		WithServiceAccountName(config.pOpts.serviceAccountName).
		WithTolerationsForTaints(config.taints...).
		WithContainerBuilder(
			container.NewBuilder().
				WithName("local-path-" + config.podName).
				WithImage(p.helperImage).
				WithCommandNew(append(config.pOpts.cmdsForPath, filepath.Join("/data/", config.volumeDir))).
				WithVolumeMountsNew([]corev1.VolumeMount{
					{
						Name:      "data",
						ReadOnly:  false,
						MountPath: "/data/",
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
		Build()

	//Launch the helper pod.
	hPod, err := p.kubeClient.CoreV1().Pods(p.namespace).Create(helperPod)
	return hPod, err
}

func (p *Provisioner) exitPod(hPod *corev1.Pod) error {
	defer func() {
		e := p.kubeClient.CoreV1().Pods(p.namespace).Delete(hPod.Name, &metav1.DeleteOptions{})
		if e != nil {
			klog.Errorf("unable to delete the helper pod: %v", e)
		}
	}()

	//Wait for the helper pod to complete it job and exit
	completed := false
	for i := 0; i < CmdTimeoutCounts; i++ {
		checkPod, err := p.kubeClient.CoreV1().Pods(p.namespace).Get(hPod.Name, metav1.GetOptions{})
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
