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
	"testing"

	ndm "github.com/openebs/api/v3/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilter(t *testing.T) {
	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		blockDeviceList *BlockDeviceList
		filterPredicate []string
		// expectedBlockDeviceListLength holds the length of disk list
		expectedBlockDeviceCount int
	}{
		"EmptyBlockDeviceList1": {
			blockDeviceList:          nil,
			filterPredicate:          []string{FilterInactive},
			expectedBlockDeviceCount: 0,
		},
		"EmptyBlockDeviceList2": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: nil,
				errs:            nil,
			},
			filterPredicate:          []string{FilterInactive},
			expectedBlockDeviceCount: 0,
		},
		"EmptyBlockDeviceList3": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{},
				errs:            nil,
			},
			filterPredicate:          []string{FilterInactive},
			expectedBlockDeviceCount: 0,
		},
		"blockDeviceList3": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:          []string{FilterInactive},
			expectedBlockDeviceCount: 3,
		},
		"blockDeviceList4": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:          []string{FilterNonInactive},
			expectedBlockDeviceCount: 0,
		},
		"blockDeviceList5": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:          []string{FilterNonInactive, FilterInactive},
			expectedBlockDeviceCount: 0,
		},
		"blockDeviceList6": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:          []string{FilterInactive, FilterNonInactive},
			expectedBlockDeviceCount: 0,
		},
		"blockDeviceList7": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:          []string{FilterNonInactive},
			expectedBlockDeviceCount: 1,
		},
		"blockDeviceList8": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
							},
							Status: ndm.DeviceStatus{
								State:      "Inactive",
								ClaimState: ndm.BlockDeviceClaimed,
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State:      "Active",
								ClaimState: ndm.BlockDeviceUnclaimed,
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State:      "Inactive",
								ClaimState: ndm.BlockDeviceClaimed,
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:          []string{FilterInactive, FilterNonInactive},
			expectedBlockDeviceCount: 0,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			filtteredBlockDeviceList := test.blockDeviceList.Filter(nil, test.filterPredicate...)
			if len(filtteredBlockDeviceList.Items) != test.expectedBlockDeviceCount {
				t.Errorf("Test %q failed: expected block device object count %d but got %d", name, test.expectedBlockDeviceCount, len(filtteredBlockDeviceList.Items))
			}
		})
	}
}

func TestHasitems(t *testing.T) {
	tests := map[string]struct {
		blockDeviceList BlockDeviceList
		expectedOutput  bool
	}{
		"Nil block device list": {
			blockDeviceList: BlockDeviceList{nil, nil},
			expectedOutput:  false,
		},
		"Empty block device list": {
			blockDeviceList: BlockDeviceList{
				&ndm.BlockDeviceList{},
				nil,
			},
			expectedOutput: false,
		},
		"Empty block device items": {
			blockDeviceList: BlockDeviceList{
				&ndm.BlockDeviceList{
					Items: []ndm.BlockDevice{},
				},
				nil,
			},
			expectedOutput: true,
		},
		"Valid block device list": {
			blockDeviceList: BlockDeviceList{
				&ndm.BlockDeviceList{
					Items: []ndm.BlockDevice{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "blockdevice",
							},
						},
					},
				},
				nil,
			},
			expectedOutput: true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			_, actual := test.blockDeviceList.Hasitems()
			if actual != test.expectedOutput {
				t.Errorf("Test %q failed expected blockdevice list items: %t and got: %t", name, test.expectedOutput, actual)
			}
		})
	}
}

func TestIsClaimed(t *testing.T) {
	tests := map[string]struct {
		blockDevice    *BlockDevice
		expectedOutput bool
	}{
		"Test Claimed Status": {
			blockDevice: &BlockDevice{
				BlockDevice: &ndm.BlockDevice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "blockdevice",
					},
					Status: ndm.DeviceStatus{
						ClaimState: ndm.BlockDeviceClaimed,
					},
				},
			},
			expectedOutput: true,
		},
		"Test UnClaimed Status": {
			blockDevice: &BlockDevice{
				BlockDevice: &ndm.BlockDevice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "blockdevice",
					},
					Status: ndm.DeviceStatus{
						ClaimState: ndm.BlockDeviceUnclaimed,
					},
				},
			},
			expectedOutput: false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			output := test.blockDevice.IsClaimed()
			if output != test.expectedOutput {
				t.Errorf("Test %q failed expected status: %t and got: %t", name, test.expectedOutput, output)
			}
		})
	}
}

func TestFilterNonPartitions(t *testing.T) {
	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		blockDeviceList *BlockDeviceList
		// expectedBlockDeviceListLength holds the length of disk list
		expectedBlockDeviceCount int
	}{
		"EmptyBlockDeviceList1": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{},
				errs:            nil,
			},
			expectedBlockDeviceCount: 0,
		},
		"blockDeviceList2": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path:        "/dev/sda",
								Partitioned: "YES",
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path:        "/dev/sdb",
								Partitioned: "NO",
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path:        "/dev/sdb",
								Partitioned: "NO",
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
					},
				},
				errs: nil,
			},
			expectedBlockDeviceCount: 2,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			filtteredBlockDeviceList := filterNonPartitions(test.blockDeviceList)
			if len(filtteredBlockDeviceList.Items) != test.expectedBlockDeviceCount {
				t.Errorf("Test %q failed: expected block device object count %d but got %d", name, test.expectedBlockDeviceCount, len(filtteredBlockDeviceList.Items))
			}
		})
	}
}

func TestFilterSparseDevices(t *testing.T) {
	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		blockDeviceList *BlockDeviceList
		// expectedBlockDeviceListLength holds the length of disk list
		expectedBlockDeviceCount int
	}{
		"EmptyBlockDeviceList1": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{},
				errs:            nil,
			},
			expectedBlockDeviceCount: 0,
		},
		"blockDeviceList2": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Details: ndm.DeviceDetails{
									DeviceType: "sparse",
								},
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Details: ndm.DeviceDetails{
									DeviceType: "HDD",
								},
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Details: ndm.DeviceDetails{
									DeviceType: "SSD",
								},
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
					},
				},
				errs: nil,
			},
			expectedBlockDeviceCount: 1,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			filtteredBlockDeviceList := filterSparseDevices(test.blockDeviceList)
			if len(filtteredBlockDeviceList.Items) != test.expectedBlockDeviceCount {
				t.Errorf("Test %q failed: expected block device object count %d but got %d", name, test.expectedBlockDeviceCount, len(filtteredBlockDeviceList.Items))
			}
		})
	}
}

func TestFilterNonSparseDevices(t *testing.T) {
	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		blockDeviceList *BlockDeviceList
		// expectedBlockDeviceListLength holds the length of disk list
		expectedBlockDeviceCount int
	}{
		"EmptyBlockDeviceList1": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{},
				errs:            nil,
			},
			expectedBlockDeviceCount: 0,
		},
		"blockDeviceList2": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Details: ndm.DeviceDetails{
									DeviceType: "sparse",
								},
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Details: ndm.DeviceDetails{
									DeviceType: "HDD",
								},
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Details: ndm.DeviceDetails{
									DeviceType: "SSD",
								},
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
					},
				},
				errs: nil,
			},
			expectedBlockDeviceCount: 2,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			filtteredBlockDeviceList := filterNonSparseDevices(test.blockDeviceList)
			if len(filtteredBlockDeviceList.Items) != test.expectedBlockDeviceCount {
				t.Errorf("Test %q failed: expected block device object count %d but got %d", name, test.expectedBlockDeviceCount, len(filtteredBlockDeviceList.Items))
			}
		})
	}
}

func TestFilterNonFSType(t *testing.T) {
	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		blockDeviceList *BlockDeviceList
		// expectedBlockDeviceListLength holds the length of disk list
		expectedBlockDeviceCount int
	}{
		"EmptyBlockDeviceList1": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{},
				errs:            nil,
			},
			expectedBlockDeviceCount: 0,
		},
		"blockDeviceList2": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
								FileSystem: ndm.FileSystemInfo{
									Type: "ext4",
								},
								Partitioned: "YES",
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path:        "/dev/sdb",
								Partitioned: "NO",
								FileSystem: ndm.FileSystemInfo{
									Type: "ext3",
								},
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path:        "/dev/sdb",
								Partitioned: "NO",
								FileSystem:  ndm.FileSystemInfo{},
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path:        "/dev/sdb",
								Partitioned: "NO",
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
					},
				},
				errs: nil,
			},
			expectedBlockDeviceCount: 2,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			filtteredBlockDeviceList := filterNonFSType(test.blockDeviceList)
			if len(filtteredBlockDeviceList.Items) != test.expectedBlockDeviceCount {
				t.Errorf("Test %q failed: expected block device object count %d but got %d", name, test.expectedBlockDeviceCount, len(filtteredBlockDeviceList.Items))
			}
		})
	}
}

func TestFilterNonReleasedDevices(t *testing.T) {
	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		blockDeviceList *BlockDeviceList
		// expectedBlockDeviceListLength holds the length of disk list
		expectedBlockDeviceCount int
	}{
		"EmptyBlockDeviceList1": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{},
				errs:            nil,
			},
			expectedBlockDeviceCount: 0,
		},
		"blockDeviceList2": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
							},
							Status: ndm.DeviceStatus{
								State:      "Active",
								ClaimState: ndm.BlockDeviceReleased,
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State:      "Active",
								ClaimState: ndm.BlockDeviceUnclaimed,
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State:      "Active",
								ClaimState: ndm.BlockDeviceReleased,
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path:        "/dev/sdb",
								Partitioned: "NO",
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
					},
				},
				errs: nil,
			},
			expectedBlockDeviceCount: 2,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			filtteredBlockDeviceList := filterNonRelesedDevices(test.blockDeviceList)
			if len(filtteredBlockDeviceList.Items) != test.expectedBlockDeviceCount {
				t.Errorf("Test %q failed: expected block device object count %d but got %d", name, test.expectedBlockDeviceCount, len(filtteredBlockDeviceList.Items))
			}
		})
	}
}

func TestBDTagFilter(t *testing.T) {
	type fields struct {
		BlockDeviceList *ndm.BlockDeviceList
	}
	type args struct {
		filterOps     *FilterOptions
		predicateKeys []string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantCount int
	}{
		{
			fields: fields{
				BlockDeviceList: &ndm.BlockDeviceList{
					Items: []ndm.BlockDevice{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-1",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-2",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-3",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-4",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-5",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
					},
				},
			},
			args: args{
				filterOps: &FilterOptions{
					AllowedBDTags: map[string]bool{},
				},
				predicateKeys: []string{FilterNotAllowedBDTag},
			},
			wantCount: 0,
		},

		{
			fields: fields{
				BlockDeviceList: &ndm.BlockDeviceList{
					Items: []ndm.BlockDevice{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-1",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-2",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-3",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-4",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-5",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
					},
				},
			},
			args: args{
				filterOps: &FilterOptions{
					AllowedBDTags: map[string]bool{"cstor": true},
				},
				predicateKeys: []string{FilterNotAllowedBDTag},
			},
			wantCount: 5,
		},

		{
			fields: fields{
				BlockDeviceList: &ndm.BlockDeviceList{
					Items: []ndm.BlockDevice{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-1",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-2",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-3",
								Labels: map[string]string{"openebs.io/block-device-tag": "local"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-4",
								Labels: map[string]string{"openebs.io/block-device-tag": "local"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-5",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
					},
				},
			},
			args: args{
				filterOps: &FilterOptions{
					AllowedBDTags: map[string]bool{"cstor": true},
				},
				predicateKeys: []string{FilterNotAllowedBDTag},
			},
			wantCount: 3,
		},

		{
			fields: fields{
				BlockDeviceList: &ndm.BlockDeviceList{
					Items: []ndm.BlockDevice{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-1",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor-cspc"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-2",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor-spc"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-3",
								Labels: map[string]string{"openebs.io/block-device-tag": "localpv"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-4",
								Labels: map[string]string{"openebs.io/block-device-tag": "zfslocal"},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-5",
								Labels: map[string]string{"openebs.io/block-device-tag": "mayastor"},
							},
						},

						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-6",
								Labels: map[string]string{"openebs.io/block-device-tag": "nvme"},
							},
						},

						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-7",
								Labels: map[string]string{"openebs.io/block-device-tag": "ssd"},
							},
						},

						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-8",
								Labels: map[string]string{"openebs.io/block-device-tag": "cstor"},
							},
						},
					},
				},
			},
			args: args{
				filterOps: &FilterOptions{
					AllowedBDTags: map[string]bool{"cstor-cspc": true, "cstor-spc": true},
				},
				predicateKeys: []string{FilterNotAllowedBDTag},
			},
			wantCount: 2,
		},

		{
			fields: fields{
				BlockDeviceList: &ndm.BlockDeviceList{
					Items: []ndm.BlockDevice{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bd-1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bd-2",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bd-3",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bd-4",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bd-5",
							},
						},

						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bd-6",
							},
						},

						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "test-bd-7",
								Labels: map[string]string{},
							},
						},

						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bd-8",
							},
						},
					},
				},
			},
			args: args{
				filterOps: &FilterOptions{
					AllowedBDTags: map[string]bool{"cstor-cspc": true, "cstor-spc": true},
				},
				predicateKeys: []string{FilterNotAllowedBDTag},
			},
			wantCount: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bdl := &BlockDeviceList{
				BlockDeviceList: tt.fields.BlockDeviceList,
			}
			got := bdl.Filter(tt.args.filterOps, tt.args.predicateKeys...)
			if len(got.Items) != tt.wantCount {
				t.Errorf("test case failed as want count %d but got %d", tt.wantCount, len(got.Items))
			}
		})
	}
}
