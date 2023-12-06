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
	"context"

	"github.com/openebs/maya/pkg/alertlog"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	mconfig "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	v1 "k8s.io/api/core/v1"
	pvController "sigs.k8s.io/sig-storage-lib-external-provisioner/v9/controller"

	pv "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/persistentvolume"
)

// ProvisionBlockDevice is invoked by the Provisioner to create a Local PV
//
//	with a Block Device
func (p *Provisioner) ProvisionBlockDevice(ctx context.Context, opts pvController.ProvisionOptions, volumeConfig *VolumeConfig) (*v1.PersistentVolume, pvController.ProvisioningState, error) {
	pvc := opts.PVC
	nodeHostname := GetNodeHostname(opts.SelectedNode)
	name := opts.PVName
	capacity := opts.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	stgType := volumeConfig.GetStorageType()
	fsType := volumeConfig.GetFSType()

	bdTagValue := volumeConfig.GetBDTagValue()
	if len(bdTagValue) > 0 {
		klog.Infof("The 'BlockDeviceTag' cas.openebs.io/config key has been deprecated in favor of the 'BlockDeviceSelectors' key.\nSample config:\n" +
			"\t\t" + "- name: BlockDeviceSelectors\n" +
			"\t\t" + "  data:\n" +
			"\t\t" + "    openebs.io/block-device-tag: \"mongo\"\n")
		return nil, pvController.ProvisioningFinished, errors.Errorf("Cannot use deprecated \"BlockDeviceTag\" config option")
	}

	nodeAffinityKey := volumeConfig.GetNodeAffinityLabelKey()
	if len(nodeAffinityKey) != 0 {
		klog.Infof("The 'NodeAffinityLabel' cas.openebs.io/config key has been deprecated in favor of the 'NodeAffinityLabels' key.\nSample config:\n" +
			"\t\t" + "- name: NodeAffinityLabels\n" +
			"\t\t" + "  list:\n" +
			"\t\t" + "    - \"node-label-key\"\n")
		return nil, pvController.ProvisioningFinished, errors.Errorf("Cannot use deprecated \"NodeAffinityLabel\" config option")
	}

	// nodeAffinityLabels contains all the custom node affinity labels.
	// This helps in cases where the hostname changes when the node is removed and
	// added back.
	nodeAffinityLabels := make(map[string]string)

	nodeAffinityKeys := volumeConfig.GetNodeAffinityLabelKeys()
	if nodeAffinityKeys == nil {
		nodeAffinityLabels[k8sNodeLabelKeyHostname] = GetNodeLabelValue(opts.SelectedNode, k8sNodeLabelKeyHostname)
	} else {
		for _, nodeAffinityKey := range nodeAffinityKeys {
			nodeAffinityLabels[nodeAffinityKey] = GetNodeLabelValue(opts.SelectedNode, nodeAffinityKey)
		}
	}

	//Extract the details to create a Block Device Claim
	blkDevOpts := &HelperBlockDeviceOptions{
		nodeHostname:       nodeHostname,
		name:               name,
		nodeAffinityLabels: nodeAffinityLabels,
		capacity:           capacity.String(),
		volumeMode:         *opts.PVC.Spec.VolumeMode,
		bdSelectors:        volumeConfig.GetBlockDeviceSelectors(),
	}

	path, blkPath, err := p.getBlockDevicePath(ctx, blkDevOpts)
	if err != nil {
		klog.Infof("Initialize volume %v failed: %v", name, err)
		alertlog.Logger.Errorw("",
			"eventcode", "local.pv.provision.failure",
			"msg", "Failed to provision Local PV",
			"rname", opts.PVName,
			"reason", "Block device initialization failed",
			"storagetype", stgType,
		)
		return nil, pvController.ProvisioningFinished, err
	}
	klog.Infof("Creating volume %v on %v at %v(%v)", name, nodeHostname, path, blkPath)

	// Over-ride the path, with the blockPath, when path is empty.
	if path == "" {
		path = blkPath
		klog.Infof("Using block device{%v} with fs{%v}", blkPath, fsType)
	}

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
	pvObjBuilder := pv.NewBuilder().
		WithName(name).
		WithLabels(labels).
		WithAnnotations(volAnnotations).
		WithReclaimPolicy(*opts.StorageClass.ReclaimPolicy).
		WithAccessModes(pvc.Spec.AccessModes).
		WithCapacityQty(pvc.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]).
		WithLocalHostPathFormat(path, fsType).
		WithNodeAffinity(nodeAffinityLabels)

	// If volumeMode set to "Block", then provide the appropriate volumeMode, to pvObj
	if *opts.PVC.Spec.VolumeMode == v1.PersistentVolumeBlock {
		pvObjBuilder.WithVolumeMode(v1.PersistentVolumeBlock)
	}

	//Build the pvObject
	pvObj, err := pvObjBuilder.Build()

	if err != nil {
		alertlog.Logger.Errorw("",
			"eventcode", "local.pv.provision.failure",
			"msg", "Failed to provision Local PV",
			"rname", opts.PVName,
			"reason", "Building volume failed",
			"storagetype", stgType,
		)
		return nil, pvController.ProvisioningFinished, err
	}
	alertlog.Logger.Infow("",
		"eventcode", "local.pv.provision.success",
		"msg", "Successfully provisioned Local PV",
		"rname", opts.PVName,
		"storagetype", stgType,
	)
	return pvObj, pvController.ProvisioningFinished, nil
}

// DeleteBlockDevice is invoked by the PVC controller to perform clean-up
//
//	activities before deleteing the PV object. If reclaim policy is
//	set to not-retain, then this function will delete the associated BDC
func (p *Provisioner) DeleteBlockDevice(ctx context.Context, pv *v1.PersistentVolume) (err error) {
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
	klog.Infof("Release the Block Device Claim %v for PV %v", blkDevOpts.bdcName, pv.Name)

	if err := p.deleteBlockDeviceClaim(ctx, blkDevOpts); err != nil {
		klog.Infof("clean up volume %v failed: %v", pv.Name, err)
		return err
	}
	return nil
}
