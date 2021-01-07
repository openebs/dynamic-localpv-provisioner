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

package app

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestGetImagePullSecrets(t *testing.T) {
	testCases := map[string]struct {
		value         string
		expectedValue []corev1.LocalObjectReference
	}{
		"empty variable": {
			value:         "",
			expectedValue: []corev1.LocalObjectReference{},
		},
		"single value": {
			value:         "image-pull-secret",
			expectedValue: []corev1.LocalObjectReference{{Name: "image-pull-secret"}},
		},
		"multiple value": {
			value:         "image-pull-secret,secret-1",
			expectedValue: []corev1.LocalObjectReference{{Name: "image-pull-secret"}, {Name: "secret-1"}},
		},
		"whitespaces": {
			value:         " ",
			expectedValue: []corev1.LocalObjectReference{},
		},
		"single value with whitespaces": {
			value:         " docker-secret ",
			expectedValue: []corev1.LocalObjectReference{{Name: "docker-secret"}},
		},
		"multiple value with whitespaces": {
			value:         " docker-secret, image-pull-secret ",
			expectedValue: []corev1.LocalObjectReference{{Name: "docker-secret"}, {Name: "image-pull-secret"}},
		},
	}
	for k, v := range testCases {
		v := v
		t.Run(k, func(t *testing.T) {
			actualValue := GetImagePullSecrets(v.value)
			if !reflect.DeepEqual(actualValue, v.expectedValue) {
				t.Errorf("expected %s got %s", v.expectedValue, actualValue)
			}
		})
	}
}
