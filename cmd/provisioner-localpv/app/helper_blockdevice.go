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
	//"path/filepath"
	//"strings"
	"time"

	"github.com/golang/glog"
	//"github.com/pkg/errors"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"

	//hostpath "github.com/openebs/maya/pkg/hostpath/v1alpha1"

	//container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	//pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	//volume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	blockdeviceclaim "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	//ndmv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	bdcStorageClassAnnotation = "local.openebs.io/blockdeviceclaim"
)

//TODO
var (
	//WaitForBDTimeoutCounts specifies the duration to wait for BDC to be associated with a BD
	//The duration is the value specified here multiplied by 5
	WaitForBDTimeoutCounts = 12
)

// HelperBlockDeviceOptions contains the options that
// will launch a BDC on a specific node (nodeName)
type HelperBlockDeviceOptions struct {
	nodeName string
	name     string
	capacity string
	//	deviceType string
	bdcName string
}

// validate checks that the required fields to create BDC
// are available
func (blkDevOpts *HelperBlockDeviceOptions) validate() error {
	glog.Infof("Validate Block Device Options")
	if blkDevOpts.name == "" || blkDevOpts.nodeName == "" {
		return errors.Errorf("invalid empty name or node")
	}
	return nil
}

// hasBDC checks if the bdcName has already been determined
func (blkDevOpts *HelperBlockDeviceOptions) hasBDC() bool {
	glog.Infof("Already has BDC %t", blkDevOpts.bdcName != "")
	return blkDevOpts.bdcName != ""
}

// setBlcokDeviceClaimFromPV inspects the PV and fetches the BDC associated
//  with the Local PV.
func (blkDevOpts *HelperBlockDeviceOptions) setBlockDeviceClaimFromPV(pv *corev1.PersistentVolume) {
	glog.Infof("Setting Block Device Claim From PV")
	bdc, found := pv.Annotations[bdcStorageClassAnnotation]
	if found {
		blkDevOpts.bdcName = bdc
	}
}

// createBlockDeviceClaim creates a new BlockDeviceClaim for a given
//  Local PV
func (p *Provisioner) createBlockDeviceClaim(blkDevOpts *HelperBlockDeviceOptions) error {
	glog.Infof("Creating Block Device Claim")
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
	//if the previous reconcilation of PVC-PV, resulted in
	//creating a BDC, but BD was not yet available for 60+ seconds
	_, err := blockdeviceclaim.NewKubeClient().
		WithNamespace(p.namespace).
		Get(bdcName, metav1.GetOptions{})
	if err == nil {
		blkDevOpts.bdcName = bdcName
		glog.Infof("Volume %v has been initialized with BDC:%v", blkDevOpts.name, bdcName)
		return nil
	}

	bdcObj, err := blockdeviceclaim.NewBuilder().
		WithNamespace(p.namespace).
		WithName(bdcName).
		WithHostName(blkDevOpts.nodeName).
		WithCapacity(blkDevOpts.capacity).
		Build()

	if err != nil {
		//TODO : Need to relook at this error
		return errors.Wrapf(err, "unable to build BDC")
	}

	_, err = blockdeviceclaim.NewKubeClient().
		WithNamespace(p.namespace).
		Create(bdcObj.Object)

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
func (p *Provisioner) getBlockDevicePath(blkDevOpts *HelperBlockDeviceOptions) (string, string, error) {

	glog.Infof("Getting Block Device Path")
	if !blkDevOpts.hasBDC() {
		err := p.createBlockDeviceClaim(blkDevOpts)
		if err != nil {
			return "", "", err
		}
	}

	//TODO
	glog.Infof("Getting Block Device Path from BDC %v", blkDevOpts.bdcName)
	bdName := ""
	//Check if the BDC is created
	for i := 0; i < WaitForBDTimeoutCounts; i++ {

		bdc, err := blockdeviceclaim.NewKubeClient().
			WithNamespace(p.namespace).
			Get(blkDevOpts.bdcName, metav1.GetOptions{})
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
		//TODO : Iterate and get the first path by id.
		blkPath = bd.Spec.DevLinks[0].Links[0]
	}

	return path, blkPath, nil
}

// deleteBlockDeviceClaim deletes the BlockDeviceClaim associated with the
//  PV being deleted.
func (p *Provisioner) deleteBlockDeviceClaim(blkDevOpts *HelperBlockDeviceOptions) error {
	glog.Infof("Delete Block Device Claim")
	if !blkDevOpts.hasBDC() {
		return nil
	}

	//TODO: Issue a delete BDC request
	err := blockdeviceclaim.NewKubeClient().
		WithNamespace(p.namespace).
		Delete(blkDevOpts.bdcName, &metav1.DeleteOptions{})

	if err != nil {
		//TODO : Need to relook at this error
		return errors.Errorf("unable to delete BDC %v associated with PV:%v", blkDevOpts.bdcName, blkDevOpts.name)
	}
	return nil
}
