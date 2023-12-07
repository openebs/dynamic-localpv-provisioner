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
	"time"

	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	blockdeviceclaim "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	bdcStorageClassAnnotation = "local.openebs.io/blockdeviceclaim"
	// LocalPVFinalizer represents finalizer string used by LocalPV
	LocalPVFinalizer = "local.openebs.io/finalizer"
)

// WaitForBDTimeoutCounts specifies the duration to wait for BDC to be associated with a BD
// The duration is the value specified here multiplied by 5
var WaitForBDTimeoutCounts int

// HelperBlockDeviceOptions contains the options that
// will launch a BDC on a specific node (nodeHostname)
type HelperBlockDeviceOptions struct {
	nodeHostname string
	name         string

	//nodeAffinityLabels represents the labels of the node where pod should be launched.
	nodeAffinityLabels map[string]string

	capacity string
	//	deviceType string
	bdcName string
	//  volumeMode of PVC
	volumeMode corev1.PersistentVolumeMode

	// bdSelectors stores the different fields
	// used for selecting a block device
	bdSelectors map[string]string
}

// BlockDeviceSelectorFields stores the block device selectors
type BlockDeviceSelectorFields map[string]string

// validate checks that the required fields to create BDC
// are available
func (blkDevOpts *HelperBlockDeviceOptions) validate() error {
	klog.V(4).Infof("Validate Block Device Options")
	if blkDevOpts.name == "" || blkDevOpts.nodeHostname == "" {
		return errors.Errorf("invalid empty name or node hostname")
	}
	return nil
}

// hasBDC checks if the bdcName has already been determined
func (blkDevOpts *HelperBlockDeviceOptions) hasBDC() bool {
	klog.V(4).Infof("Already has BDC %t", blkDevOpts.bdcName != "")
	return blkDevOpts.bdcName != ""
}

// setBlcokDeviceClaimFromPV inspects the PV and fetches the BDC associated
//
//	with the Local PV.
func (blkDevOpts *HelperBlockDeviceOptions) setBlockDeviceClaimFromPV(pv *corev1.PersistentVolume) {
	klog.V(4).Infof("Setting Block Device Claim From PV")
	bdc, found := pv.Annotations[bdcStorageClassAnnotation]
	if found {
		blkDevOpts.bdcName = bdc
	}
}

// createBlockDeviceClaim creates a new BlockDeviceClaim for a given
//
//	Local PV
func (p *Provisioner) createBlockDeviceClaim(ctx context.Context, blkDevOpts *HelperBlockDeviceOptions) error {
	klog.V(4).Infof("Creating Block Device Claim")
	if err := blkDevOpts.validate(); err != nil {
		return err
	}
	//Create a BDC for this PV (of type device). NDM will
	//look for the device matching the capacity and node on which
	//pod is being scheduled. Since this BDC is specific to a PV
	//use the name of the bdc to be:  "bdc-<pvname>"
	//TODO: Look into setting the labels and owner references
	//on BDC with PV/PVC details.
	bdcName := "bdc-" + blkDevOpts.name

	//Check if the BDC is already created. This can happen
	//if the previous reconciliation of PVC-PV, resulted in
	//creating a BDC, but BD was not yet available for 60+ seconds
	_, err := blockdeviceclaim.NewKubeClient().
		WithNamespace(p.namespace).
		Get(ctx, bdcName, metav1.GetOptions{})
	if err == nil {
		blkDevOpts.bdcName = bdcName
		klog.Infof("Volume %v has been initialized with BDC:%v", blkDevOpts.name, bdcName)
		return nil
	}

	bdcObjBuilder := blockdeviceclaim.NewBuilder().
		WithNamespace(p.namespace).
		WithName(bdcName).
		WithSelector(blkDevOpts.nodeAffinityLabels).
		WithCapacity(blkDevOpts.capacity).
		WithFinalizer(LocalPVFinalizer).
		WithBlockVolumeMode(blkDevOpts.volumeMode)

	// if block device selectors are present, set it on the BDC
	if blkDevOpts.bdSelectors != nil {
		bdcObjBuilder.WithSelector(blkDevOpts.bdSelectors)
	}

	bdcObj, err := bdcObjBuilder.Build()

	if err != nil {
		//TODO : Need to relook at this error
		return errors.Wrapf(err, "unable to build BDC")
	}

	_, err = blockdeviceclaim.NewKubeClient().
		WithNamespace(p.namespace).
		Create(ctx, bdcObj.Object)

	if err != nil {
		//TODO : Need to relook at this error
		//If the error is about BDC being already present, then return nil
		return errors.Wrapf(err, "failed to create BDC{%v}", bdcName)
	}

	blkDevOpts.bdcName = bdcName

	return nil
}

// getBlockDevicePath fetches the BDC associated with this Local PV
// or creates one. From the BDC, fetch the BD and get the path
func (p *Provisioner) getBlockDevicePath(ctx context.Context, blkDevOpts *HelperBlockDeviceOptions) (string, string, error) {

	klog.V(4).Infof("Getting Block Device Path")
	if !blkDevOpts.hasBDC() {
		err := p.createBlockDeviceClaim(ctx, blkDevOpts)
		if err != nil {
			return "", "", err
		}
	}

	//TODO
	klog.Infof("Getting Block Device Path from BDC %v", blkDevOpts.bdcName)
	bdName := ""
	//Check if the BDC is created
	for i := 0; i < WaitForBDTimeoutCounts; i++ {

		bdc, err := blockdeviceclaim.NewKubeClient().
			WithNamespace(p.namespace).
			Get(ctx, blkDevOpts.bdcName, metav1.GetOptions{})
		if err != nil {
			//TODO : Need to relook at this error
			//If the error is about BDC being already present, then return nil
			return "", "", errors.Errorf("unable to get BDC %v associated with PV:%v %v", blkDevOpts.bdcName, blkDevOpts.name, err)
		}

		bdName = bdc.Spec.BlockDeviceName
		//Check if the BDC is associated with a BD
		if bdName == "" {
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	// if bdName not found should delete BDC and return err
	if bdName == "" {
		err := errors.Errorf("unable to find BD for BDC:%v associated with PV:%v and try to delete BDC", blkDevOpts.bdcName, blkDevOpts.name)
		delErr := p.deleteBlockDeviceClaim(ctx, blkDevOpts)
		if delErr != nil {
			return "", "", delErr
		} else {
			return "", "", err
		}
	}

	//Get the BD Path.
	bd, err := blockdevice.NewKubeClient().
		WithNamespace(p.namespace).
		Get(bdName, metav1.GetOptions{})
	if err != nil {
		//TODO : Need to relook at this error
		//If the error is about BDC being already present, then return nil
		return "", "", errors.Errorf("unable to find BD:%v for BDC:%v associated with PV:%v", bdName, blkDevOpts.bdcName, blkDevOpts.name)
	}
	path := bd.Spec.FileSystem.Mountpoint

	blkPath := bd.Spec.Path
	if len(bd.Spec.DevLinks) > 0 {

		blkPath = bd.Spec.DevLinks[0].Links[0]

		//Iterate and get the first path by id.
		for _, v := range bd.Spec.DevLinks {
			if v.Kind == "by-id" {
				blkPath = v.Links[0]
			}
		}
	}

	return path, blkPath, nil
}

// deleteBlockDeviceClaim deletes the BlockDeviceClaim associated with the
//
//	PV being deleted.
func (p *Provisioner) deleteBlockDeviceClaim(ctx context.Context, blkDevOpts *HelperBlockDeviceOptions) error {
	klog.V(4).Infof("Delete Block Device Claim")
	if !blkDevOpts.hasBDC() {
		return nil
	}

	err := p.removeFinalizer(ctx, blkDevOpts)
	if err != nil {
		// if finalizer is not removed, donot proceed with deletion
		return errors.Errorf("unable to remove finalizer on BDC %v : %v", blkDevOpts.name, err)
	}

	err = blockdeviceclaim.NewKubeClient().
		WithNamespace(p.namespace).
		Delete(ctx, blkDevOpts.bdcName, &metav1.DeleteOptions{})

	if err != nil {
		//TODO : Need to relook at this error
		return errors.Errorf("unable to delete BDC %v associated with PV:%v", blkDevOpts.bdcName, blkDevOpts.name)
	}
	return nil
}

func (p *Provisioner) removeFinalizer(ctx context.Context, blkDevOpts *HelperBlockDeviceOptions) error {
	klog.V(4).Info("removing local-pv finalizer on the BDC")

	bdc, err := blockdeviceclaim.NewKubeClient().
		WithNamespace(p.namespace).
		Get(ctx, blkDevOpts.bdcName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("unable to get BDC %s for removing finalizer", blkDevOpts.bdcName)
	}

	// edit the finalizer in the copy of the BDC
	bdc.Finalizers = util.RemoveString(bdc.Finalizers, LocalPVFinalizer)

	// udpate the BDC with the new finalizer array
	_, err = blockdeviceclaim.NewKubeClient().
		WithNamespace(p.namespace).
		Update(ctx, bdc)

	return err
}
