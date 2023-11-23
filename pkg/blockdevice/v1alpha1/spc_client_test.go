/*
Copyright 2018 The OpenEBS Authors

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
	"strconv"
	"testing"

	apis "github.com/openebs/api/v3/pkg/apis/openebs.io/v1alpha1"
	ndmFakeClient "github.com/openebs/api/v3/pkg/client/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSpcClientGet(t *testing.T) {
	var blockDeviceK8s BlockDeviceInterface
	var blockDeviceSpc BlockDeviceInterface
	fndmcs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     ndmFakeClient.NewSimpleClientset(),
	}
	spcClient := &SpcObjectClient{
		KubernetesClient: fndmcs,
		Spc: &apis.StoragePoolClaim{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "pool1",
			},
			Spec: apis.StoragePoolClaimSpec{
				BlockDevices: apis.BlockDeviceAttr{
					BlockDeviceList: []string{"myblockDevice1", "myblockDevice4"},
				},
			},
		},
	}
	blockDeviceK8s = fndmcs
	blockDeviceSpc = spcClient
	blockDeviceObj, err := New().WithName("myblockDevice1").Build()
	if err != nil {
		t.Fatalf("Could not build blockDevice object:%s", err)
	}
	blockDeviceK8s.Create(blockDeviceObj)
	blockDeviceObj, err = New().WithName("myblockDevice2").Build()
	if err != nil {
		t.Fatalf("Could not build blockDevice object:%s", err)
	}
	blockDeviceK8s.Create(blockDeviceObj)
	tests := map[string]struct {
		blockDeviceName         string
		expectedBlockDeviceName string
		expectedErr             bool
	}{
		"blockDevice present in spc as well as k8s": {
			blockDeviceName:         "myblockDevice1",
			expectedBlockDeviceName: "myblockDevice1",
			expectedErr:             false,
		},
		"blockDevice not present in both spc and k8s": {
			blockDeviceName:         "myblockDevice2",
			expectedBlockDeviceName: "myblockDevice2",
			expectedErr:             true,
		},
		"blockDevice not present in spc but present in k8s": {
			blockDeviceName:         "myblockDevice3",
			expectedBlockDeviceName: "myblockDevice3",
			expectedErr:             true,
		},
		"blockDevice present in spc but not present in k8s": {
			blockDeviceName:         "myblockDevice4",
			expectedBlockDeviceName: "myblockDevice4",
			expectedErr:             true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			blockDevice, err := blockDeviceSpc.Get(test.blockDeviceName, metav1.GetOptions{})
			if test.expectedErr {
				if err == nil {
					t.Error("Test case failed as got nil error")
				}
			} else {

				if blockDevice.BlockDevice == nil {
					t.Fatalf("Test case failed as got nil blockDevice:%v", err)
				}
				if blockDevice.BlockDevice.Name != test.expectedBlockDeviceName {
					t.Errorf("Test case failed as expected blockDevice name %s but got %s", blockDevice.BlockDevice.Name, test.expectedBlockDeviceName)
				}
			}

		})
	}
}

func TestSpcClientList(t *testing.T) {
	var blockDeviceK8s BlockDeviceInterface
	// Get a fake openebs client set
	fndmcs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     ndmFakeClient.NewSimpleClientset(),
	}
	blockDeviceK8s = fndmcs
	blockDeviceObj, err := New().WithName("myblockDevice1").Build()
	if err != nil {
		t.Fatalf("Could not build blockDevice object:%s", err)
	}
	blockDeviceK8s.Create(blockDeviceObj)
	blockDeviceObj, err = New().WithName("myblockDevice2").Build()
	if err != nil {
		t.Fatalf("Could not build blockDevice object:%s", err)
	}
	blockDeviceK8s.Create(blockDeviceObj)

	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		blockDeviceCount int
		spcClientObj     *SpcObjectClient
		// expectedDiskListLength holds the length of blockDevice list
		expectedErr bool
	}{
		// Test Case #1
		"All blockDevices specified in spc is present in k8s": {
			blockDeviceCount: 2,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: fndmcs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{
						BlockDevices: apis.BlockDeviceAttr{
							BlockDeviceList: []string{"myblockDevice1", "myblockDevice2"},
						},
					},
				},
			},
			expectedErr: false,
		},
		"Some blockDevices specified in spc is present in k8s": {
			blockDeviceCount: 2,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: fndmcs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{
						BlockDevices: apis.BlockDeviceAttr{
							BlockDeviceList: []string{"myblockDevice1", "myblockDevice2", "myblockDevice3", "myblockDevice4"},
						},
					},
				},
			},
			expectedErr: false,
		},
		"All blockDevices specified in spc is not present in k8s": {
			blockDeviceCount: 0,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: fndmcs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{
						BlockDevices: apis.BlockDeviceAttr{
							BlockDeviceList: []string{"myblockDevice3", "myblockDevice4", "myblockDevice5", "myblockDevice6"},
						},
					},
				},
			},
			expectedErr: false,
		},
		"No blockDevices specified in spc": {
			blockDeviceCount: 0,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: fndmcs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{},
				},
			},
			expectedErr: true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			blockDeviceList, err := test.spcClientObj.List(metav1.ListOptions{})
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if test.expectedErr {
				if gotErr != test.expectedErr {
					t.Errorf("Test case failed as expected error %t but got %t", test.expectedErr, gotErr)
				}
			} else {
				if gotErr != test.expectedErr {
					t.Fatalf("Test case failed as the expected error %t but got %t:%v", test.expectedErr, gotErr, err)
				}
				if len(blockDeviceList.BlockDeviceList.Items) != test.blockDeviceCount {
					t.Errorf("Test case failed as expected blockDevice object count %d but got %d", test.blockDeviceCount, len(blockDeviceList.BlockDeviceList.Items))
				}
			}

		})
	}
}

func TestSpcClientFilteredList(t *testing.T) {
	var blockDeviceK8s BlockDeviceInterface
	// Get a fake openebs client set
	fndmcs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     ndmFakeClient.NewSimpleClientset(),
	}
	blockDeviceK8s = fndmcs
	// Create some blockDevice objects
	for i := 1; i <= 5; i++ {
		var blockDeviceState string
		if i%2 == 0 {
			blockDeviceState = "Inactive"
		} else {
			blockDeviceState = "Active"
		}
		blockDeviceObj, errs := New().WithName("myblockDevice" + strconv.Itoa(i)).WithState(blockDeviceState).Build()
		if errs != nil {
			t.Fatalf("Could not build blockDevice object:%s", errs)
		}
		blockDevice, err := blockDeviceK8s.Create(blockDeviceObj)
		if blockDevice == nil {
			t.Fatalf("Failed to create blockDevice object:%v", err)
		}
	}

	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		blockDeviceCount int
		spcClientObj     *SpcObjectClient
		filterWith       []string
		// expectedDiskListLength holds the length of blockDevice list
		expectedErr bool
	}{
		// Test Case #1
		"All blockDevices specified in spc is present in k8s": {
			blockDeviceCount: 2,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: fndmcs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{
						BlockDevices: apis.BlockDeviceAttr{
							BlockDeviceList: []string{"myblockDevice1", "myblockDevice2", "myblockDevice3", "myblockDevice4", "myblockDevice5"},
						},
					},
				},
			},
			expectedErr: false,
			filterWith:  []string{FilterInactive},
		},
		"Some blockDevices specified in spc is present in k8s": {
			blockDeviceCount: 1,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: fndmcs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{
						BlockDevices: apis.BlockDeviceAttr{
							BlockDeviceList: []string{"myblockDevice1", "myblockDevice2", "myblockDevice3", "myblockDevice6"},
						},
					},
				},
			},
			expectedErr: false,
		},
		"All blockDevices specified in spc is not present in k8s": {
			blockDeviceCount: 0,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: fndmcs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{
						BlockDevices: apis.BlockDeviceAttr{
							BlockDeviceList: []string{"myblockDevice6", "myblockDevice7", "myblockDevice8", "myblockDevice9"},
						},
					},
				},
			},
			expectedErr: false,
		},
		"No blockDevices specified in spc": {
			blockDeviceCount: 0,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: fndmcs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{},
				},
			},
			expectedErr: true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			blockDeviceList, err := test.spcClientObj.List(metav1.ListOptions{})
			blockDeviceList = blockDeviceList.Filter(nil, FilterInactive)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.expectedErr {
				t.Fatalf("Test case failed as the expected error %v but got %v", test.expectedErr, gotErr)
			}
			if len(blockDeviceList.BlockDeviceList.Items) != test.blockDeviceCount {
				t.Errorf("Test case failed as expected blockDevice object count %d but got %d", test.blockDeviceCount, len(blockDeviceList.BlockDeviceList.Items))
			}
		})
	}
}
