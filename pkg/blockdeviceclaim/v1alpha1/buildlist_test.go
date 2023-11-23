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
)

func TestListBuilderWithAPIList(t *testing.T) {
	tests := map[string]struct {
		availableBDCs  []string
		expectedBDCLen int
	}{
		"BDC set 1": {[]string{}, 0},
		"BDC set 2": {[]string{"bdc1"}, 1},
		"BDC set 3": {[]string{"bdc1", "bdc2"}, 2},
		"BDC set 4": {[]string{"bdc1", "bdc2", "bdc3"}, 3},
		"BDC set 5": {[]string{"bdc1", "bdc2", "bdc3", "bdc4"}, 4},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := ListBuilderFromAPIList(fakeAPIBDCList(mock.availableBDCs)).List()
			itemsLen := len(b.ObjectList.Items)
			if mock.expectedBDCLen != itemsLen {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedBDCLen, itemsLen)
			}
		})
	}
}
