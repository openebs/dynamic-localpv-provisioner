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

// ProvisionBlockDevice is invoked by the Provisioner to create a Local PV
//  with a Block Device
func (p *Provisioner) ProvisionBlockDevice(opts pvController.VolumeOptions, volumeConfig *VolumeConfig) (*v1.PersistentVolume, error) {
	pvc := opts.PVC
	node := opts.SelectedNode
	name := opts.PVName
	capacity := opts.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	stgType := volumeConfig.GetStorageType()
	fsType := volumeConfig.GetFSType()

	//Extract the details to create a Block Device Claim
	blkDevOpts := &HelperBlockDeviceOptions{
		nodeName: node.Name,
		name:     name,
		capacity: capacity.String(),
	}

	path, blkPath, err := p.getBlockDevicePath(blkDevOpts)
	if err != nil {
		glog.Infof("Initialize volume %v failed: %v", name, err)
		return nil, err
	}
	glog.Infof("Creating volume %v on %v at %v(%v)", name, node.Name, path, blkPath)
	if path == "" {
		path = blkPath
		glog.Infof("Using block device{%v} with fs{%v}", blkPath, fsType)
	}

	// TODO
	// VolumeMode will always be specified as Filesystem for host path volume,
	// and the value passed in from the PVC spec will be ignored.
	fs := v1.PersistentVolumeFilesystem

	// It is possible that the HostPath doesn't already exist on the node.
	// Set the Local PV to create it.
	//hostPathType := v1.HostPathDirectoryOrCreate

	// TODO initialize the Labels and annotations
	// Use annotations to specify the context using which the PV was created.
	volAnnotations := make(map[string]string)
	volAnnotations[bdcStorageClassAnnotation] = blkDevOpts.bdcName
	//fstype := casVolume.Spec.FSType

	labels := make(map[string]string)
	labels[string(mconfig.CASTypeKey)] = "local-" + stgType
	//labels[string(v1alpha1.StorageClassKey)] = *className

	//TODO Change the following to a builder pattern
	pvObj, err := mPV.NewBuilder().
		WithName(name).
		WithLabels(labels).
		WithAnnotations(volAnnotations).
		WithReclaimPolicy(opts.PersistentVolumeReclaimPolicy).
		WithAccessModes(pvc.Spec.AccessModes).
		WithVolumeMode(fs).
		WithCapacityQty(pvc.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]).
		WithLocalHostPathFormat(path, fsType).
		WithNodeAffinity(node.Name).
		Build()

	if err != nil {
		return nil, err
	}

	return pvObj, nil

}

// DeleteBlockDevice is invoked by the PVC controller to perform clean-up
//  activities before deleteing the PV object. If reclaim policy is
//  set to not-retain, then this function will delete the associated BDC
func (p *Provisioner) DeleteBlockDevice(pv *v1.PersistentVolume) (err error) {
	defer func() {
		err = errors.Wrapf(err, "failed to delete volume %v", pv.Name)
	}()

	blkDevOpts := &HelperBlockDeviceOptions{
		name: pv.Name,
	}

	//Determine if a BDC is set on the PV and save it to BlockDeviceOptions
	blkDevOpts.setBlockDeviceClaimFromPV(pv)

	//Initiate clean up only when reclaim policy is not retain.
	//TODO: this part of the code could be eliminated by setting up
	// BDC owner reference to PVC.
	glog.Infof("Release the Block Device Claim %v for PV %v", blkDevOpts.bdcName, pv.Name)

	if err := p.deleteBlockDeviceClaim(blkDevOpts); err != nil {
		glog.Infof("clean up volume %v failed: %v", pv.Name, err)
		return err
	}
	return nil
}
