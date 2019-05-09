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
	"fmt"
	"path/filepath"
	//"strings"
	"time"

	"github.com/golang/glog"
	//"github.com/pkg/errors"

	hostpath "github.com/openebs/maya/pkg/hostpath/v1alpha1"

	mContainer "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	mPod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	mVolume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	//CmdTimeoutCounts specifies the duration to wait for cleanup pod to be launched.
	CmdTimeoutCounts = 120
)

// getPathAndNodeForPV inspects the PV spec to determine the host path used
//  and the node (via the NodeAffinity) on which host path exists.
func (p *Provisioner) getPathAndNodeForPV(pv *v1.PersistentVolume) (path, node string, err error) {
	hostPath := pv.Spec.PersistentVolumeSource.HostPath
	if hostPath == nil {
		return "", "", fmt.Errorf("no HostPath set")
	}
	path = hostPath.Path

	nodeAffinity := pv.Spec.NodeAffinity
	if nodeAffinity == nil {
		return "", "", fmt.Errorf("no NodeAffinity set")
	}
	required := nodeAffinity.Required
	if required == nil {
		return "", "", fmt.Errorf("no NodeAffinity.Required set")
	}

	node = ""
	for _, selectorTerm := range required.NodeSelectorTerms {
		for _, expression := range selectorTerm.MatchExpressions {
			if expression.Key == KeyNode && expression.Operator == v1.NodeSelectorOpIn {
				if len(expression.Values) != 1 {
					return "", "", fmt.Errorf("multiple values for the node affinity")
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
		return "", "", fmt.Errorf("cannot find affinited node")
	}
	return path, node, nil
}

// createCleanupPod launches a helper(busybox) pod, to delete the host path.
//  This provisioner expects that the host paths are created using
//  an unique PV path - under a given BasePath. From the absolute path,
//  it extracts the base path and the PV path. The helper pod is then launched
//  by mounting the base path - and performing a delete on the unique PV path.
func (p *Provisioner) createCleanupPod(pOpts *HelperPodOptions) (err error) {
	//func (p *Provisioner) createCleanupPod(cmdsForPath []string, name, path, node string) (err error) {
	if pOpts.name == "" || pOpts.path == "" || pOpts.nodeName == "" {
		return fmt.Errorf("invalid empty name or path or node")
	}

	// Initialize HostPath builder and validate that
	// non-root directories are not passed for delete
	// Extract the base path and the volume unique path.
	parentDir, volumeDir, vErr := hostpath.NewBuilder().WithPath(pOpts.path).
		WithCheckf(hostpath.IsNonRoot(), "path should not be a root directory: path %v", pOpts.path).
		ExtractSubPath()
	if vErr != nil {
		return vErr
	}

	conObj, _ := mContainer.Builder().
		WithName("local-path-cleanup").
		WithImage(p.helperImage).
		WithCommand(append(pOpts.cmdsForPath, filepath.Join("/data/", volumeDir))).
		WithVolumeMounts([]v1.VolumeMount{
			{
				Name:      "data",
				ReadOnly:  false,
				MountPath: "/data/",
			},
		}).
		Build()
	containers := []v1.Container{conObj}

	volObj, _ := mVolume.NewBuilder().
		WithName("data").
		WithHostDirectory(parentDir).
		Build()
	volumes := []v1.Volume{*volObj}

	helperPod, _ := mPod.NewBuilder().
		WithName("cleanup-" + pOpts.name).
		WithRestartPolicy(v1.RestartPolicyNever).
		WithNodeName(pOpts.nodeName).
		WithContainers(containers).
		WithVolumes(volumes).
		Build()

	//Launch the cleanup pod.
	pod, err := p.kubeClient.CoreV1().Pods(p.namespace).Create(helperPod)
	if err != nil {
		return err
	}

	defer func() {
		e := p.kubeClient.CoreV1().Pods(p.namespace).Delete(pod.Name, &metav1.DeleteOptions{})
		if e != nil {
			glog.Errorf("unable to delete the helper pod: %v", e)
		}
	}()

	//Wait for the cleanup pod to complete it job and exit
	completed := false
	for i := 0; i < CmdTimeoutCounts; i++ {
		if pod, err := p.kubeClient.CoreV1().Pods(p.namespace).Get(pod.Name, metav1.GetOptions{}); err != nil {
			return err
		} else if pod.Status.Phase == v1.PodSucceeded {
			completed = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	if !completed {
		return fmt.Errorf("create process timeout after %v seconds", CmdTimeoutCounts)
	}

	glog.Infof("Volume %v has been cleaned on %v:%v", pOpts.name, pOpts.nodeName, pOpts.path)
	return nil
}
