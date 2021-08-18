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

package app

import (
	"os"
	"reflect"
	"testing"

	menv "github.com/openebs/maya/pkg/env/v1alpha1"
)

func TestConverToK(t *testing.T) {
	testCases := map[string]struct {
		value       string
		expectValue string
	}{
		"Missing limit grace": {
			value:       "",
			expectValue: "0k",
		},
		"Present limit grace with value": {
			value:       "0%",
			expectValue: "5000000k",
		},
		"Present limit grace exceeding 100%": {
			value:       "200%",
			expectValue: "10000000k",
		},
	}

	for k, v := range testCases {
		v := v
		t.Run(k, func(t *testing.T) {
			if len(v.value) != 0 {
				os.Setenv(string(menv.OpenEBSNamespace), v.value)
			}
			actualValue, _ := convertToK(v.value, 5000000000)
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
			os.Unsetenv(string(menv.OpenEBSNamespace))
		})
	}
}
