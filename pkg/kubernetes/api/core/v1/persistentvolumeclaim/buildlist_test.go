/*
Copyright 2020 The OpenEBS Authors

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

package persistentvolumeclaim

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestListBuilderWithAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePVCs  []string
		expectedPVCLen int
	}{
		"PVC set 1":  {[]string{}, 0},
		"PVC set 2":  {[]string{"pvc1"}, 1},
		"PVC set 3":  {[]string{"pvc1", "pvc2"}, 2},
		"PVC set 4":  {[]string{"pvc1", "pvc2", "pvc3"}, 3},
		"PVC set 5":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4"}, 4},
		"PVC set 6":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5"}, 5},
		"PVC set 7":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6"}, 6},
		"PVC set 8":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7"}, 7},
		"PVC set 9":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7", "pvc8"}, 8},
		"PVC set 10": {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7", "pvc8", "pvc9"}, 9},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := ListBuilderForAPIObjects(fakeAPIPVCList(mock.availablePVCs))
			if mock.expectedPVCLen != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPVCLen, len(b.list.items))
			}
		})
	}
}

func TestListBuilderWithAPIObjects(t *testing.T) {
	tests := map[string]struct {
		availablePVCs  []string
		expectedPVCLen int
		expectedErr    bool
	}{
		"PVC set 1": {[]string{}, 0, true},
		"PVC set 2": {[]string{"pvc1"}, 1, false},
		"PVC set 3": {[]string{"pvc1", "pvc2"}, 2, false},
		"PVC set 4": {[]string{"pvc1", "pvc2", "pvc3"}, 3, false},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b, err := ListBuilderForAPIObjects(fakeAPIPVCList(mock.availablePVCs)).APIList()
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
			if !mock.expectedErr && mock.expectedPVCLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.availablePVCs, len(b.Items))
			}
		})
	}
}

func TestListBuilderAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePVCs  []string
		expectedPVCLen int
		expectedErr    bool
	}{
		"PVC set 1": {[]string{}, 0, true},
		"PVC set 2": {[]string{"pvc1"}, 1, false},
		"PVC set 3": {[]string{"pvc1", "pvc2"}, 2, false},
		"PVC set 4": {[]string{"pvc1", "pvc2", "pvc3"}, 3, false},
		"PVC set 5": {[]string{"pvc1", "pvc2", "pvc3", "pvc4"}, 4, false},
	}
	for name, mock := range tests {

		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b, err := ListBuilderForAPIObjects(fakeAPIPVCList(mock.availablePVCs)).APIList()
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
			if err == nil && mock.expectedPVCLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPVCLen, len(b.Items))
			}
		})
	}
}

func TestFilterList(t *testing.T) {
	tests := map[string]struct {
		availablePVCs map[string]corev1.PersistentVolumeClaimPhase
		filteredPVCs  []string
		filters       PredicateList
		expectedErr   bool
	}{
		"PVC Set 1": {
			availablePVCs: map[string]corev1.PersistentVolumeClaimPhase{"PVC5": corev1.ClaimBound, "PVC6": corev1.ClaimPending, "PVC7": corev1.ClaimLost},
			filteredPVCs:  []string{"PVC5"},
			filters:       PredicateList{IsBound()},
			expectedErr:   false,
		},

		"PVC Set 2": {
			availablePVCs: map[string]corev1.PersistentVolumeClaimPhase{"PVC3": corev1.ClaimBound, "PVC4": corev1.ClaimBound},
			filteredPVCs:  []string{"PVC2", "PVC4"},
			filters:       PredicateList{IsBound()},
			expectedErr:   false,
		},

		"PVC Set 3": {
			availablePVCs: map[string]corev1.PersistentVolumeClaimPhase{"PVC1": corev1.ClaimLost, "PVC2": corev1.ClaimPending, "PVC3": corev1.ClaimPending},
			filteredPVCs:  []string{},
			filters:       PredicateList{IsBound()},
			expectedErr:   false,
		},
		"PVC Set 4": {
			availablePVCs: map[string]corev1.PersistentVolumeClaimPhase{},
			filteredPVCs:  []string{},
			filters:       PredicateList{IsBound()},
			expectedErr:   true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			list, err := ListBuilderForAPIObjects(fakeAPIPVCListFromNameStatusMap(mock.availablePVCs)).WithFilter(mock.filters...).List()
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
			if err == nil && len(list.items) != len(mock.filteredPVCs) {
				t.Fatalf("Test %v failed: expected %v got %v", name, len(mock.filteredPVCs), len(list.items))
			}
		})
	}
}
