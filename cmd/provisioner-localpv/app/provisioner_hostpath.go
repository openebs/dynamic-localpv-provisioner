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
*/

package app

import (
	"github.com/golang/glog"
	"github.com/pkg/errors"

	pvController "github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/controller"
	mconfig "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	mPV "github.com/openebs/maya/pkg/kubernetes/persistentvolume/v1alpha1"
	"k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProvisionHostPath is invoked by the Provisioner which expect HostPath PV
//  to be provisioned and a valid PV spec returned.
func (p *Provisioner) ProvisionHostPath(opts pvController.VolumeOptions, volumeConfig *VolumeConfig) (*v1.PersistentVolume, error) {
	pvc := opts.PVC
	node := opts.SelectedNode
	name := opts.PVName
	stgType := volumeConfig.GetStorageType()

	path, err := volumeConfig.GetPath()
	if err != nil {
		return nil, err
	}

	glog.Infof("Creating volume %v at %v:%v", name, node.Name, path)

	//Before using the path for local PV, make sure it is created.
	initCmdsForPath := []string{"mkdir", "-m", "0777", "-p"}
	podOpts := &HelperPodOptions{
		cmdsForPath: initCmdsForPath,
		name:        name,
		path:        path,
		nodeName:    node.Name,
	}

	iErr := p.createInitPod(podOpts)
	if iErr != nil {
		glog.Infof("Initialize volume %v failed: %v", name, iErr)
		return nil, iErr
	}

	// VolumeMode will always be specified as Filesystem for host path volume,
	// and the value passed in from the PVC spec will be ignored.
	fs := v1.PersistentVolumeFilesystem

	// It is possible that the HostPath doesn't already exist on the node.
	// Set the Local PV to create it.
	//hostPathType := v1.HostPathDirectoryOrCreate

	// TODO initialize the Labels and annotations
	// Use annotations to specify the context using which the PV was created.
	//volAnnotations := make(map[string]string)
	//volAnnotations[string(v1alpha1.CASTypeKey)] = casVolume.Spec.CasType
	//fstype := casVolume.Spec.FSType

	labels := make(map[string]string)
	labels[string(mconfig.CASTypeKey)] = "local-" + stgType
	//labels[string(v1alpha1.StorageClassKey)] = *className

	//TODO Change the following to a builder pattern
	pvObj, err := mPV.NewBuilder().
		WithName(name).
		WithLabels(labels).
		WithReclaimPolicy(opts.PersistentVolumeReclaimPolicy).
		WithAccessModes(pvc.Spec.AccessModes).
		WithVolumeMode(fs).
		WithCapacityQty(pvc.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]).
		WithLocalHostDirectory(path).
		WithNodeAffinity(node.Name).
		Build()

	if err != nil {
		return nil, err
	}

	return pvObj, nil

}

// DeleteHostPath is invoked by the PVC controller to perform clean-up
//  activities before deleteing the PV object. If reclaim policy is
//  set to not-retain, then this function will create a helper pod
//  to delete the host path from the node.
func (p *Provisioner) DeleteHostPath(pv *v1.PersistentVolume) (err error) {
	defer func() {
		err = errors.Wrapf(err, "failed to delete volume %v", pv.Name)
	}()

	//Determine the path and node of the Local PV.
	path, node, err := p.getPathAndNodeForPV(pv)
	if err != nil {
		return err
	}

	//Initiate clean up only when reclaim policy is not retain.
	glog.Infof("Deleting volume %v at %v:%v", pv.Name, node, path)
	cleanupCmdsForPath := []string{"rm", "-rf"}
	podOpts := &HelperPodOptions{
		cmdsForPath: cleanupCmdsForPath,
		name:        pv.Name,
		path:        path,
		nodeName:    node,
	}

	if err := p.createCleanupPod(podOpts); err != nil {
		return errors.Wrapf(err, "clean up volume %v failed", pv.Name)
	}
	return nil
}
