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

	mconfig "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
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

func TestDataConfigToMap(t *testing.T) {
	hostpathConfig := mconfig.Config{Name: "StorageType", Value: "hostpath"}
	xfsQuotaConfig := mconfig.Config{Name: "XFSQuota", Enabled: "true",
		Data: map[string]string{
			"SoftLimitGrace": "20%",
			"HardLimitGrace": "80%",
		},
	}

	testCases := map[string]struct {
		config        []mconfig.Config
		expectedValue map[string]interface{}
	}{
		"nil 'Data' map": {
			config: []mconfig.Config{hostpathConfig, xfsQuotaConfig},
			expectedValue: map[string]interface{}{
				"XFSQuota": map[string]string{
					"SoftLimitGrace": "20%",
					"HardLimitGrace": "80%",
				},
			},
		},
	}

	for k, v := range testCases {
		v := v
		k := k
		t.Run(k, func(t *testing.T) {
			actualValue, err := dataConfigToMap(v.config)
			if err != nil {
				t.Errorf("expected error to be nil, but got %v", err)
			}
			if !reflect.DeepEqual(actualValue, v.expectedValue) {
				t.Errorf("expected %v, but got %v", v.expectedValue, actualValue)
			}
		})
	}
}

func TestPermissionConfigToMap(t *testing.T) {
	hostpathConfig := mconfig.Config{Name: "StorageType", Value: "hostpath"}
	permissionConfig := mconfig.Config{Name: "FilePermissions", Enabled: "true",
		Data: map[string]string{
			"mode": "0750",
		},
	}

	testCases := map[string]struct {
		config        []mconfig.Config
		expectedValue map[string]interface{}
	}{
		"nil 'Data' map": {
			config: []mconfig.Config{hostpathConfig, permissionConfig},
			expectedValue: map[string]interface{}{
				"FilePermissions": map[string]string{
					"mode": "0750",
				},
			},
		},
	}

	for k, v := range testCases {
		v := v
		k := k
		t.Run(k, func(t *testing.T) {
			actualValue, err := dataConfigToMap(v.config)
			if err != nil {
				t.Errorf("expected error to be nil, but got %v", err)
			}
			if !reflect.DeepEqual(actualValue, v.expectedValue) {
				t.Errorf("expected %v, but got %v", v.expectedValue, actualValue)
			}
		})
	}
}

func Test_listConfigToMap(t *testing.T) {
	tests := map[string]struct {
		pvConfig      []mconfig.Config
		expectedValue map[string]interface{}
		wantErr       bool
	}{
		"Valid list parameter": {
			pvConfig: []mconfig.Config{
				{Name: "StorageType", Value: "hostpath"},
				{Name: "NodeAffinityLabels", List: []string{"fake-node-label-key"}},
			},
			expectedValue: map[string]interface{}{
				"NodeAffinityLabels": []string{"fake-node-label-key"},
			},
			wantErr: false,
		},
	}
	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			got, err := listConfigToMap(v.pvConfig)
			if (err != nil) != v.wantErr {
				t.Errorf("listConfigToMap() error = %v, wantErr %v", err, v.wantErr)
				return
			}
			if !reflect.DeepEqual(got, v.expectedValue) {
				t.Errorf("listConfigToMap() got = %v, want %v", got, v.expectedValue)
			}
		})
	}
}
