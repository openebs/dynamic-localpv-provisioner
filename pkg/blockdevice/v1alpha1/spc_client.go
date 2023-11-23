/*
Copyright 2019 The OpenEBS Authors

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

package v1alpha1

import (
	"context"

	ndm "github.com/openebs/api/v3/pkg/apis/openebs.io/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Get is spc client implementation to get blockdevice
func (s *SpcObjectClient) Get(name string, opts metav1.GetOptions) (*BlockDevice, error) {
	spcBDList := s.Spc.Spec.BlockDevices.BlockDeviceList
	var diskName string
	for _, disk := range spcBDList {
		if name == disk {
			diskName = name
		}
	}
	if diskName == "" {
		return nil, errors.Errorf("Disk %s not found in the given SPC %s", diskName, s.Spc.Name)
	}
	namespace := env.Get(env.OpenEBSNamespace)
	d, err := s.Clientset.OpenebsV1alpha1().BlockDevices(namespace).
		Get(context.TODO(), diskName, opts)
	return &BlockDevice{d, nil}, err
}

// List is spc client implementation to list blockdevices
func (s *SpcObjectClient) List(opts metav1.ListOptions) (*BlockDeviceList, error) {
	bdL := &BlockDeviceList{
		BlockDeviceList: &ndm.BlockDeviceList{},
		errs:            nil,
	}
	var err error
	spcBDList := s.Spc.Spec.BlockDevices.BlockDeviceList
	if len(spcBDList) == 0 {
		return nil, errors.Errorf("No disk found in the given SPC %s", s.Spc.Name)
	}
	spcDiskMap := make(map[string]int)
	for _, diskName := range spcBDList {
		spcDiskMap[diskName]++
	}

	namespace := env.Get(env.OpenEBSNamespace)
	getAllDisk, err := s.Clientset.OpenebsV1alpha1().BlockDevices(namespace).
		List(context.TODO(), opts)
	if getAllDisk.Items == nil {
		return nil, errors.Wrapf(err, "Could not get disk from kube apiserver")
	}
	for _, disk := range getAllDisk.Items {
		if spcDiskMap[disk.Name] > 0 {
			bdL.BlockDeviceList.Items = append(bdL.BlockDeviceList.Items, disk)
		}
	}
	return bdL, err
}

// Create is kubernetes client implementation to create blockdevice
func (s *SpcObjectClient) Create(bs *ndm.BlockDevice) (*BlockDevice, error) {
	return nil, errors.New("Disk object creation is not supported through spc client")
}
