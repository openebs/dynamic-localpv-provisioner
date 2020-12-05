// Copyright © 2018-2020 The OpenEBS Authors
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

package pod

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListBuilderToAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePods  []string
		expectedPodLen int
	}{
		"Pod set 1": {[]string{}, 0},
		"Pod set 2": {[]string{"pod1"}, 1},
		"Pod set 3": {[]string{"pod1", "pod2"}, 2},
		"Pod set 4": {[]string{"pod1", "pod2", "pod3"}, 3},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := ListBuilderForAPIList(fakeAPIPodList(mock.availablePods)).List().ToAPIList()
			if mock.expectedPodLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPodLen, len(b.Items))
			}
		})
	}
}

func TestHasLabelPredicate(t *testing.T) {
	tests := map[string]struct {
		availableLabels          map[string]string
		checkForKey, checkForVal string
		hasLabels                bool
	}{
		"Test1": {map[string]string{"Label 1": "Key 1"}, "Label 1", "Key 1", true},
		"Test2": {map[string]string{"Label 1": "Key 1", "Label 2": "Key 2"}, "Label 1", "Key 1", true},
		"Test3": {map[string]string{"Label 1": "Key 1", "Label 2": "Key 2"}, "Label 3", "Key 3", false},
		"Test4": {map[string]string{"Label 1": "Key 1", "Label 2": "Key 2"}, "Label 1", "Key 0", false},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			fakePod := &Pod{&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: test.availableLabels}}}
			ok := HasLabel(test.checkForKey, test.checkForVal)(fakePod)
			if ok != test.hasLabels {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.availableLabels, fakePod.object.GetLabels())
			}
		})
	}
}

func TestHasLabels(t *testing.T) {
	tests := map[string]struct {
		availableLabels map[string]string
		checkLabels     map[string]string
		hasLabels       bool
	}{
		"Test1": {
			availableLabels: map[string]string{"Label 1": "Key 1"},
			checkLabels:     map[string]string{"Label 1": "Key 1"},
			hasLabels:       true,
		},
		"Test2": {
			availableLabels: map[string]string{"Label 1": "Key 1", "Label 2": "Key 2", "L1": "K1", "L3": "K3"},
			checkLabels:     map[string]string{"Label 1": "Key 1", "L3": "K3"},
			hasLabels:       true,
		},
		"Test3": {
			availableLabels: map[string]string{"Label 1": "Key 1", "Label 2": "Key 2", "L1": "K1"},
			checkLabels:     map[string]string{"L1": "K1", "Label 3": "Key 3"},
			hasLabels:       false,
		},
		"Test4": {
			availableLabels: map[string]string{"Label 1": "Key 1", "Label 2": "Key 2"},
			checkLabels:     map[string]string{"Label 1": "Key 0"},
			hasLabels:       false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			fakePod := &Pod{&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: test.availableLabels}}}
			ok := HasLabels(test.checkLabels)(fakePod)
			if ok != test.hasLabels {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.availableLabels, fakePod.object.GetLabels())
			}
		})
	}
}
