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

// createInitPod launches a pod, to set quota on the hostpath.
func (p *Provisioner) createInitQuotaPod(pOpts *HelperPodOptions, bsoft string, bhard string) error {
	var config podConfig
	config.pOpts, config.podName = pOpts, "init-quota"

	var vErr error
	config.parentDir, config.volumeDir, vErr = hostpath.NewBuilder().WithPath(pOpts.path).
		WithCheckf(hostpath.IsNonRoot(), "volume directory {%v} should not be under root directory", pOpts.path).
		ExtractSubPath()
	if vErr != nil {
		return vErr
	}

	//set limits of quota on the project
	config.bsoft = bsoft
	config.bhard = bhard

	//Pass on the taints, to create tolerations.
	config.taints = pOpts.selectedNodeTaints

	iPod, err := p.launchQuotaPod(config)
	if err != nil {
		return err
	}

	if err := p.exitPod(iPod); err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) launchQuotaPod(config podConfig) (*corev1.Pod, error) {
	// the quota pod need to be launched in privileged mode. This is because in CoreOS
	// nodes, pods without privileged access cannot write to the host directory.
	// Quota pods need to setup projectID which requires to write the mapping of projectID and subdirectory.

	lastPid := "PID=`xfs_quota -x -c 'report -h' /data | tail -2 | awk 'NR==1{print substr ($1,2)}+0'` ;"
	newPid := "PID=`expr $PID + 1` ;"
	initializeProject := "xfs_quota -x -c 'project -s -p " + filepath.Join("/data/", config.volumeDir) + " '$PID /data ;"
	setQuota := "xfs_quota -x -c 'limit -p bsoft=" + config.bsoft + " bhard=" + config.bhard + " '$PID /data"

	config.pOpts.cmdsForPath = []string{"sh", "-c", lastPid + newPid + initializeProject + setQuota}

	privileged := true

	klog.Infof("Setting the project quota...")
	klog.Infof("Check filesystem of mount if quota didn't set")

	quotaPod, err := pod.NewBuilder().
		WithName(config.podName+"-"+config.pOpts.name).
		WithRestartPolicy(corev1.RestartPolicyNever).
		WithNodeAffinityNew(config.pOpts.nodeAffinityLabelKey, config.pOpts.nodeAffinityLabelValue).
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

	if err != nil {
		return nil, err
	}

	var qPod *corev1.Pod

	//Launch the quota pod.
	qPod, err = p.kubeClient.CoreV1().Pods(p.namespace).Create(quotaPod)
	return qPod, err
}

func (p *Provisioner) exitQuotaPod(qPod *corev1.Pod) error {
	defer func() {
		e := p.kubeClient.CoreV1().Pods(p.namespace).Delete(qPod.Name, &metav1.DeleteOptions{})
		if e != nil {
			klog.Errorf("unable to delete the quota pod: %v", e)
		}
	}()

	//Wait for the quota pod to complete it job and exit
	completed := false
	for i := 0; i < CmdTimeoutCounts; i++ {
		checkPod, err := p.kubeClient.CoreV1().Pods(p.namespace).Get(qPod.Name, metav1.GetOptions{})
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
