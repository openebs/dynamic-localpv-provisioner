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
	//"fmt"
	"path/filepath"
	//"strings"
	"time"

	"github.com/golang/glog"
	//"github.com/pkg/errors"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"

	hostpath "github.com/openebs/maya/pkg/hostpath/v1alpha1"

	container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	volume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	//CmdTimeoutCounts specifies the duration to wait for cleanup pod to be launched.
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

// getPathAndNodeForPV inspects the PV spec to determine the hostpath
//  and the node of OpenEBS Local PV. Both types of OpenEBS Local PV
//  (storage type = hostpath and device) use:
//  -  LocalVolumeSource to specify the path and
//  -  NodeAffinity to specify the node.
//  Note: This function also takes care of deleting OpenEBS Local PVs
//  provisioned in 0.9, which were using HostPathVolumeSource to
//  specify the path.
func (p *Provisioner) getPathAndNodeForPV(pv *corev1.PersistentVolume) (string, string, error) {
	path := ""
	local := pv.Spec.PersistentVolumeSource.Local
	if local == nil {
		//Handle the case of Local PV created in 0.9 using HostPathVolumeSource
		hostPath := pv.Spec.PersistentVolumeSource.HostPath
		if hostPath == nil {
			return "", "", errors.Errorf("no HostPath set")
		}
		path = hostPath.Path
	} else {
		path = local.Path
	}

	nodeAffinity := pv.Spec.NodeAffinity
	if nodeAffinity == nil {
		return "", "", errors.Errorf("no NodeAffinity set")
	}
	required := nodeAffinity.Required
	if required == nil {
		return "", "", errors.Errorf("no NodeAffinity.Required set")
	}

	node := ""
	for _, selectorTerm := range required.NodeSelectorTerms {
		for _, expression := range selectorTerm.MatchExpressions {
			if expression.Key == KeyNode && expression.Operator == corev1.NodeSelectorOpIn {
				if len(expression.Values) != 1 {
					return "", "", errors.Errorf("multiple values for the node affinity")
				}
				node = expression.Values[0]
				break
			}
		}
		if node != "" {
			break
		}
	}
	if node == "" {
		return "", "", errors.Errorf("cannot find affinited node")
	}
	return path, node, nil
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

	conObj, _ := container.NewBuilder().
		WithName("local-path-init").
		WithImage(p.helperImage).
		WithCommand(append(pOpts.cmdsForPath, filepath.Join("/data/", volumeDir))).
		WithVolumeMounts([]corev1.VolumeMount{
			{
				Name:      "data",
				ReadOnly:  false,
				MountPath: "/data/",
			},
		}).
		Build()
	//containers := []v1.Container{conObj}

	volObj, _ := volume.NewBuilder().
		WithName("data").
		WithHostDirectory(parentDir).
		Build()
	//volumes := []v1.Volume{*volObj}

	helperPod, _ := pod.NewBuilder().
		WithName("init-" + pOpts.name).
		WithRestartPolicy(corev1.RestartPolicyNever).
		WithNodeName(pOpts.nodeName).
		WithContainer(conObj).
		WithVolume(*volObj).
		Build()

	//Launch the init pod.
	hPod, err := p.kubeClient.CoreV1().Pods(p.namespace).Create(helperPod)
	if err != nil {
		return err
	}

	defer func() {
		e := p.kubeClient.CoreV1().Pods(p.namespace).Delete(hPod.Name, &metav1.DeleteOptions{})
		if e != nil {
			glog.Errorf("unable to delete the helper pod: %v", e)
		}
	}()

	//Wait for the cleanup pod to complete it job and exit
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

	//glog.Infof("Volume %v has been initialized on %v:%v", pOpts.name, pOpts.nodeName, pOpts.path)
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

	conObj, _ := container.NewBuilder().
		WithName("local-path-cleanup").
		WithImage(p.helperImage).
		WithCommand(append(pOpts.cmdsForPath, filepath.Join("/data/", volumeDir))).
		WithVolumeMounts([]corev1.VolumeMount{
			{
				Name:      "data",
				ReadOnly:  false,
				MountPath: "/data/",
			},
		}).
		Build()
	//containers := []v1.Container{conObj}

	volObj, _ := volume.NewBuilder().
		WithName("data").
		WithHostDirectory(parentDir).
		Build()
	//volumes := []v1.Volume{*volObj}

	helperPod, _ := pod.NewBuilder().
		WithName("cleanup-" + pOpts.name).
		WithRestartPolicy(corev1.RestartPolicyNever).
		WithNodeName(pOpts.nodeName).
		WithContainer(conObj).
		WithVolume(*volObj).
		Build()

	//Launch the cleanup pod.
	hPod, err := p.kubeClient.CoreV1().Pods(p.namespace).Create(helperPod)
	if err != nil {
		return err
	}

	defer func() {
		e := p.kubeClient.CoreV1().Pods(p.namespace).Delete(hPod.Name, &metav1.DeleteOptions{})
		if e != nil {
			glog.Errorf("unable to delete the helper pod: %v", e)
		}
	}()

	//Wait for the cleanup pod to complete it job and exit
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

	glog.Infof("Volume %v has been cleaned on %v:%v", pOpts.name, pOpts.nodeName, pOpts.path)
	return nil
}
