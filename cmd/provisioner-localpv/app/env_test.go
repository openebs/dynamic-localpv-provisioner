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
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	"os"
	"reflect"
	"testing"
)

func TestGetOpenEBSNamespace(t *testing.T) {
	testCases := map[string]struct {
		value       string
		expectValue string
	}{
		"Missing env variable": {
			value:       "",
			expectValue: "",
		},
		"Present env variable with value": {
			value:       "value1",
			expectValue: "value1",
		},
		"Present env variable with whitespaces": {
			value:       " ",
			expectValue: "",
		},
	}

	for k, v := range testCases {
		v := v
		t.Run(k, func(t *testing.T) {
			if len(v.value) != 0 {
				os.Setenv(string(menv.OpenEBSNamespace), v.value)
			}
			actualValue := getOpenEBSNamespace()
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
			os.Unsetenv(string(menv.OpenEBSNamespace))
		})
	}
}

func TestGetDefaultHelperImage(t *testing.T) {
	testCases := map[string]struct {
		value       string
		expectValue string
	}{
		"Missing env variable": {
			value:       "",
			expectValue: defaultHelperImage,
		},
		"Present env variable with value": {
			value:       "value1",
			expectValue: "value1",
		},
		"Present env variable with whitespaces": {
			value:       " ",
			expectValue: defaultHelperImage,
		},
	}

	for k, v := range testCases {
		v := v
		t.Run(k, func(t *testing.T) {
			if len(v.value) != 0 {
				os.Setenv(string(ProvisionerHelperImage), v.value)
			}
			actualValue := getDefaultHelperImage()
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
			os.Unsetenv(string(ProvisionerHelperImage))
		})
	}
}

func TestGetDefaultBasePath(t *testing.T) {
	testCases := map[string]struct {
		value       string
		expectValue string
	}{
		"Missing env variable": {
			value:       "",
			expectValue: defaultBasePath,
		},
		"Present env variable with value": {
			value:       "value1",
			expectValue: "value1",
		},
		"Present env variable with whitespaces": {
			value:       " ",
			expectValue: defaultBasePath,
		},
	}

	for k, v := range testCases {
		v := v
		t.Run(k, func(t *testing.T) {
			if len(v.value) != 0 {
				os.Setenv(string(ProvisionerBasePath), v.value)
			}
			actualValue := getDefaultBasePath()
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
			os.Unsetenv(string(ProvisionerBasePath))
		})
	}
}

func TestGetOpenEBSServiceAccountName(t *testing.T) {
	testCases := map[string]struct {
		value         string
		expectedValue string
	}{
		"Missing env variable": {
			value:         "",
			expectedValue: "",
		},
		"Present env variable with value": {
			value:         "value1",
			expectedValue: "value1",
		},
		"Present env variable with whitespaces": {
			value:         " ",
			expectedValue: "",
		},
	}
	for k, v := range testCases {
		v := v
		t.Run(k, func(t *testing.T) {
			if len(v.value) != 0 {
				os.Setenv(string(menv.OpenEBSServiceAccount), v.value)
			}
			actualValue := getOpenEBSServiceAccountName()
			if !reflect.DeepEqual(actualValue, v.expectedValue) {
				t.Errorf("expected %s got %s", v.expectedValue, actualValue)
			}
			os.Unsetenv(string(menv.OpenEBSServiceAccount))
		})
	}
}
