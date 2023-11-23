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

package v1alpha2

import (
	"testing"

	ndmapis "github.com/openebs/api/v3/pkg/apis/openebs.io/v1alpha1"
)

func TestValidateBlockDevice(t *testing.T) {
	tests := map[string]struct {
		bd            *BlockDevice
		expectedError bool
		validateList  []Validate
	}{
		"BlockDevice with filesystem": {
			bd: &BlockDevice{
				Object: &ndmapis.BlockDevice{
					Spec: ndmapis.DeviceSpec{
						FileSystem: ndmapis.FileSystemInfo{
							Type: "xfs",
						},
					},
				},
			},
			expectedError: true,
			validateList:  []Validate{CheckIfBDIsNonFsType()},
		},
		"BlockDevice with different node name": {
			bd: &BlockDevice{
				Object: &ndmapis.BlockDevice{
					Spec: ndmapis.DeviceSpec{
						NodeAttributes: ndmapis.NodeAttribute{
							NodeName: "node1",
						},
					},
				},
			},
			expectedError: true,
			validateList:  []Validate{CheckIfBDBelongsToNode("node2")},
		},
		"BlockDevice with InActive state": {
			bd: &BlockDevice{
				Object: &ndmapis.BlockDevice{
					Status: ndmapis.DeviceStatus{
						State: "InActive",
					},
				},
			},
			expectedError: true,
			validateList:  []Validate{CheckIfBDIsActive()},
		},
		"Validate all the changes": {
			bd: &BlockDevice{
				Object: &ndmapis.BlockDevice{
					Spec: ndmapis.DeviceSpec{
						NodeAttributes: ndmapis.NodeAttribute{
							NodeName: "node1",
						},
					},
					Status: ndmapis.DeviceStatus{
						State: "Active",
					},
				},
			},
			expectedError: false,
			validateList:  []Validate{CheckIfBDIsNonFsType(), CheckIfBDBelongsToNode("node1"), CheckIfBDIsActive()},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			err := test.bd.ValidateBlockDevice(test.validateList...)
			if test.expectedError && err == nil {
				t.Errorf("test %s failed expected error but got nil", name)
			}
			if !test.expectedError && err != nil {
				t.Errorf("test %s failed expected error to be nil but got %v", name, err)
			}
		})
	}
}
