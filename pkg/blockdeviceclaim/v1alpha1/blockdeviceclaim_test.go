// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"testing"

	ndmapis "github.com/openebs/api/v3/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fakeAPIBDCList(bdcNames []string) *ndmapis.BlockDeviceClaimList {
	if len(bdcNames) == 0 {
		return nil
	}
	list := &ndmapis.BlockDeviceClaimList{}
	for _, name := range bdcNames {
		bdc := ndmapis.BlockDeviceClaim{}
		bdc.SetName(name)
		list.Items = append(list.Items, bdc)
	}
	return list
}

func TestGetBDList(t *testing.T) {
	tests := map[string]struct {
		bdcList     *BlockDeviceClaimList
		nodeCount   int
		expectedLen []int
	}{
		"blockDeviceClaimList1": {
			bdcList: &BlockDeviceClaimList{
				ObjectList: &ndmapis.BlockDeviceClaimList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndmapis.BlockDeviceClaim{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndmapis.DeviceClaimSpec{
								BlockDeviceNodeAttributes: ndmapis.BlockDeviceNodeAttributes{
									HostName: "openebs-1234",
								},
								BlockDeviceName: "blockdevice1",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndmapis.DeviceClaimSpec{
								BlockDeviceNodeAttributes: ndmapis.BlockDeviceNodeAttributes{
									HostName: "openebs-1234",
								},
								BlockDeviceName: "blockdevice2",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndmapis.DeviceClaimSpec{
								BlockDeviceNodeAttributes: ndmapis.BlockDeviceNodeAttributes{
									HostName: "openebs-1234",
								},
								BlockDeviceName: "blockdevice3",
							},
						},
					},
				},
			},
			nodeCount:   1,
			expectedLen: []int{3},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			nodeBDList := test.bdcList.GetBlockDeviceNamesByNode()
			if len(nodeBDList) != test.nodeCount {
				t.Errorf("Test %q failed: expected block device object count %d but got %d", name, test.nodeCount, len(nodeBDList))
			}
		})
	}
}

func TestGetHostName(t *testing.T) {
	tests := map[string]struct {
		bdc            *BlockDeviceClaim
		expectedOutput string
	}{
		"Test with blockdevice attribute hostname": {
			bdc: &BlockDeviceClaim{
				Object: &ndmapis.BlockDeviceClaim{
					Spec: ndmapis.DeviceClaimSpec{
						BlockDeviceNodeAttributes: ndmapis.BlockDeviceNodeAttributes{
							HostName: "fakeNode1",
						},
					},
				},
			},
			expectedOutput: "fakeNode1",
		},
		"Test with spec hostName": {
			bdc: &BlockDeviceClaim{
				Object: &ndmapis.BlockDeviceClaim{
					Spec: ndmapis.DeviceClaimSpec{
						HostName: "fakeNode2",
					},
				},
			},
			expectedOutput: "fakeNode2",
		},
		"Test with empty": {
			bdc: &BlockDeviceClaim{
				Object: &ndmapis.BlockDeviceClaim{},
			},
			expectedOutput: "",
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			hostName := test.bdc.GetHostName()
			if hostName != test.expectedOutput {
				t.Errorf("Test %q failed: expected hostName %s but got hostName %s", name, test.expectedOutput, hostName)
			}
		})
	}
}

func TestGetBlockDeviceClaimFromBDName(t *testing.T) {
	tests := map[string]struct {
		bdcList       BlockDeviceClaimList
		bdName        string
		expectedError bool
	}{
		"When block device claim list exists": {
			bdcList: BlockDeviceClaimList{
				ObjectList: &ndmapis.BlockDeviceClaimList{
					Items: []ndmapis.BlockDeviceClaim{
						ndmapis.BlockDeviceClaim{
							ObjectMeta: metav1.ObjectMeta{
								Name: "blockdevice1",
							},
							Spec: ndmapis.DeviceClaimSpec{
								BlockDeviceName: "blockdevice1",
							},
						},
						ndmapis.BlockDeviceClaim{
							Spec: ndmapis.DeviceClaimSpec{
								BlockDeviceName: "blockdevice2",
							},
						},
						ndmapis.BlockDeviceClaim{
							Spec: ndmapis.DeviceClaimSpec{
								BlockDeviceName: "blockdevice3",
							},
						},
					},
				},
			},
			bdName:        "blockdevice3",
			expectedError: false,
		},
		"when block device doesn't exist": {
			bdcList: BlockDeviceClaimList{
				ObjectList: &ndmapis.BlockDeviceClaimList{
					Items: []ndmapis.BlockDeviceClaim{},
				},
			},
			bdName:        "blockdevice2",
			expectedError: true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			_, err := test.bdcList.GetBlockDeviceClaimFromBDName(test.bdName)
			if test.expectedError && err == nil {
				t.Errorf("test %s failed expected error but got nil", name)
			}
			if !test.expectedError && err != nil {
				t.Errorf("test %s failed expected not to get error but got error :%v", name, err)
			}

		})
	}
}
