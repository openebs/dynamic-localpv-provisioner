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

	"github.com/golang/glog"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"

	hostpath "github.com/openebs/maya/pkg/hostpath/v1alpha1"

	container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	volume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	//CmdTimeoutCounts specifies the duration to wait for cleanup pod
	//to be launched.
	CmdTimeoutCounts = 120
)

// HelperPodOptions contains the options that
// will launch a Pod on a specific node (nodeName)
// to execute a command (cmdsForPath) on a given
// volume path (path)
type HelperPodOptions struct {
	//nodeName represents the host where pod should be launched.
	nodeName string
	//name is the name of the PV for which the pod is being launched
	name string
	//cmdsForPath represent either create (mkdir) or delete(rm)
	//commands that need to be executed on the volume path.
	cmdsForPath []string
	//path is the volume hostpath directory
	path string
}

// validate checks that the required fields to launch
// helper pods are valid. helper pods are used to either
// create or delete a directory (path) on a given node (nodeName).
// name refers to the volume being created or deleted.
func (pOpts *HelperPodOptions) validate() error {
	if pOpts.name == "" || pOpts.path == "" || pOpts.nodeName == "" {
		return errors.Errorf("invalid empty name or path or node")
	}
	return nil
}

// createInitPod launches a helper(busybox) pod, to create the host path.
//  The local pv expect the hostpath to be already present before mounting
//  into pod. Validate that the local pv host path is not created under root.
func (p *Provisioner) createInitPod(pOpts *HelperPodOptions) error {
	//err := pOpts.validate()
	if err := pOpts.validate(); err != nil {
		return err
	}

	// Initialize HostPath builder and validate that
	// volume directory is not directly under root.
	// Extract the base path and the volume unique path.
	parentDir, volumeDir, vErr := hostpath.NewBuilder().WithPath(pOpts.path).
		WithCheckf(hostpath.IsNonRoot(), "volume directory {%v} should not be under root directory", pOpts.path).
		ExtractSubPath()
	if vErr != nil {
		return vErr
	}

	initPod, _ := pod.NewBuilder().
		WithName("init-" + pOpts.name).
		WithRestartPolicy(corev1.RestartPolicyNever).
		WithNodeName(pOpts.nodeName).
		WithContainerBuilder(
			container.NewBuilder().
				WithName("local-path-init").
				WithImage(p.helperImage).
				WithCommandNew(append(pOpts.cmdsForPath, filepath.Join("/data/", volumeDir))).
				WithVolumeMountsNew([]corev1.VolumeMount{
					{
						Name:      "data",
						ReadOnly:  false,
						MountPath: "/data/",
					},
				}),
		).
		WithVolumeBuilder(
			volume.NewBuilder().
				WithName("data").
				WithHostDirectory(parentDir),
		).
		Build()

	//Launch the init pod.
	iPod, err := p.kubeClient.CoreV1().Pods(p.namespace).Create(initPod)
	if err != nil {
		return err
	}

	defer func() {
		e := p.kubeClient.CoreV1().Pods(p.namespace).Delete(iPod.Name, &metav1.DeleteOptions{})
		if e != nil {
			glog.Errorf("unable to delete the helper pod: %v", e)
		}
	}()

	//Wait for the cleanup pod to complete it job and exit
	completed := false
	for i := 0; i < CmdTimeoutCounts; i++ {
		checkPod, err := p.kubeClient.CoreV1().Pods(p.namespace).Get(iPod.Name, metav1.GetOptions{})
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

// createCleanupPod launches a helper(busybox) pod, to delete the host path.
//  This provisioner expects that the host paths are created using
//  an unique PV path - under a given BasePath. From the absolute path,
//  it extracts the base path and the PV path. The helper pod is then launched
//  by mounting the base path - and performing a delete on the unique PV path.
func (p *Provisioner) createCleanupPod(pOpts *HelperPodOptions) error {
	//err := pOpts.validate()
	if err := pOpts.validate(); err != nil {
		return err
	}

	// Initialize HostPath builder and validate that
	// volume directory is not directly under root.
	// Extract the base path and the volume unique path.
	parentDir, volumeDir, vErr := hostpath.NewBuilder().WithPath(pOpts.path).
		WithCheckf(hostpath.IsNonRoot(), "volume directory {%v} should not be under root directory", pOpts.path).
		ExtractSubPath()
	if vErr != nil {
		return vErr
	}

	cleanerPod, _ := pod.NewBuilder().
		WithName("cleanup-" + pOpts.name).
		WithRestartPolicy(corev1.RestartPolicyNever).
		WithNodeName(pOpts.nodeName).
		WithContainerBuilder(
			container.NewBuilder().
				WithName("local-path-cleanup").
				WithImage(p.helperImage).
				WithCommandNew(append(pOpts.cmdsForPath, filepath.Join("/data/", volumeDir))).
				WithVolumeMountsNew([]corev1.VolumeMount{
					{
						Name:      "data",
						ReadOnly:  false,
						MountPath: "/data/",
					},
				}),
		).
		WithVolumeBuilder(
			volume.NewBuilder().
				WithName("data").
				WithHostDirectory(parentDir),
		).
		Build()

	//Launch the cleanup pod.
	cPod, err := p.kubeClient.CoreV1().Pods(p.namespace).Create(cleanerPod)
	if err != nil {
		return err
	}

	defer func() {
		e := p.kubeClient.CoreV1().Pods(p.namespace).Delete(cPod.Name, &metav1.DeleteOptions{})
		if e != nil {
			glog.Errorf("unable to delete the helper pod: %v", e)
		}
	}()

	//Wait for the cleanup pod to complete it job and exit
	completed := false
	for i := 0; i < CmdTimeoutCounts; i++ {
		checkPod, err := p.kubeClient.CoreV1().Pods(p.namespace).Get(cPod.Name, metav1.GetOptions{})
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
