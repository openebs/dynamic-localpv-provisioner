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

func TestDataConfigToMap(t *testing.T) {
	hostpathConfig := Config{Name: "StorageType", Value: "hostpath"}
	xfsQuotaConfig := Config{Name: "XFSQuota", Enabled: "true",
		Data: map[string]RawLiteral{
			"SoftLimitGrace": "20%",
			"HardLimitGrace": "80%",
		},
	}

	testCases := map[string]struct {
		config        []Config
		expectedValue map[string]interface{}
	}{
		"nil 'Data' map": {
			config: []Config{hostpathConfig, xfsQuotaConfig},
			expectedValue: map[string]interface{}{
				"XFSQuota": map[string]RawLiteral{
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
	hostpathConfig := Config{Name: "StorageType", Value: "hostpath"}
	permissionConfig := Config{Name: "FilePermissions",
		Data: map[string]RawLiteral{
			"mode": "0750",
		},
	}

	testCases := map[string]struct {
		config        []Config
		expectedValue map[string]interface{}
	}{
		"nil 'Data' map": {
			config: []Config{hostpathConfig, permissionConfig},
			expectedValue: map[string]interface{}{
				"FilePermissions": map[string]RawLiteral{
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
		pvConfig      []Config
		expectedValue map[string]interface{}
		wantErr       bool
	}{
		"Valid list parameter": {
			pvConfig: []Config{
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

func TestConfigGetImagePullPolicy(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected corev1.PullPolicy
	}{
		"Always Policy": {
			input:    "Always",
			expected: corev1.PullAlways,
		},
		"Never Policy": {
			input:    "Never",
			expected: corev1.PullNever,
		},
		"IfNotPresent Policy": {
			input:    "IfNotPresent",
			expected: corev1.PullIfNotPresent,
		},
		"Empty String Defaults to IfNotPresent": {
			input:    "",
			expected: corev1.PullIfNotPresent,
		},
		"Invalid Value Defaults to IfNotPresent": {
			input:    "invalid",
			expected: corev1.PullIfNotPresent,
		},
		"Whitespace Value Defaults to IfNotPresent": {
			input:    " ",
			expected: corev1.PullIfNotPresent,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			actual := GetImagePullPolicy(tc.input)

			if actual != tc.expected {
				t.Errorf(
					"expected %q, got %q",
					tc.expected,
					actual,
				)
			}
		})
	}
}

func TestFilterPVCConfig(t *testing.T) {
	tests := map[string]struct {
		configs        []Config
		restrictedKeys []string
		expectedNames  []string
	}{
		"filters BasePath": {
			configs: []Config{
				{Name: "BasePath", Value: "/evil/path"},
				{Name: "NodeAffinityLabels", List: []string{"kubernetes.io/hostname"}},
				{Name: "StorageType", Value: "hostpath"},
			},
			restrictedKeys: []string{KeyPVBasePath},
			expectedNames:  []string{"NodeAffinityLabels", "StorageType"},
		},
		"no restricted keys leaves all configs": {
			configs: []Config{
				{Name: "BasePath", Value: "/some/path"},
				{Name: "StorageType", Value: "hostpath"},
			},
			restrictedKeys: []string{},
			expectedNames:  []string{"BasePath", "StorageType"},
		},
		"empty config list": {
			configs:        []Config{},
			restrictedKeys: []string{KeyPVBasePath},
			expectedNames:  []string{},
		},
		"filters with whitespace in name": {
			configs: []Config{
				{Name: " BasePath ", Value: "/evil/path"},
				{Name: "StorageType", Value: "hostpath"},
			},
			restrictedKeys: []string{KeyPVBasePath},
			expectedNames:  []string{"StorageType"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := filterPVCConfig(tc.configs, tc.restrictedKeys)
			if len(got) != len(tc.expectedNames) {
				t.Fatalf("filterPVCConfig returned %d configs, want %d", len(got), len(tc.expectedNames))
			}
			for i, c := range got {
				if c.Name != tc.expectedNames[i] {
					t.Errorf("config[%d].Name = %q, want %q", i, c.Name, tc.expectedNames[i])
				}
			}
		})
	}
}
