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

	"k8s.io/client-go/kubernetes/fake"

	ndmFakeClient "github.com/openebs/api/v3/pkg/client/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilteredList(t *testing.T) {
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
		// expectedDiskListLength holds the length of blockDevice list
		err bool
	}{
		// Test Case #1
		"disk": {
			2,
			false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			blockDeviceList, err := blockDeviceK8s.List(metav1.ListOptions{})
			filtteredBlockDeviceList := blockDeviceList.Filter(nil, FilterInactive)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v", test.err, gotErr)
			}
			if len(filtteredBlockDeviceList.Items) != test.blockDeviceCount {
				t.Errorf("Test case failed as expected blockDevice object count %d but got %d", test.blockDeviceCount, len(filtteredBlockDeviceList.Items))
			}
		})
	}
}
