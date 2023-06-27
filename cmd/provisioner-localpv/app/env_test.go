package app

import (
	"os"
	"reflect"
	"strconv"
	"testing"

	menv "github.com/openebs/maya/pkg/env/v1alpha1"
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

func TestGetHelperPodHostNetworke(t *testing.T) {
	testCases := map[string]struct {
		value       string
		expectValue string
	}{
		"Missing env variable": {
			value:       "",
			expectValue: "false",
		},
		"Present env variable with value": {
			value:       "value1",
			expectValue: "false",
		},
		"Present env variable with whitespaces": {
			value:       "true",
			expectValue: "true",
		},
	}

	for k, v := range testCases {
		v := v
		t.Run(k, func(t *testing.T) {
			if len(v.value) != 0 {
				os.Setenv(string(ProvisionerHelperPodHostNetwork), v.value)
			}
			actualValue := strconv.FormatBool(getHelperPodHostNetwork())
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
			os.Unsetenv(string(ProvisionerHelperPodHostNetwork))
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

func TestGetOpenEBSImagePullSecrets(t *testing.T) {
	testCases := map[string]struct {
		value         string
		expectedValue string
	}{
		"Missing env variable": {
			value:         "",
			expectedValue: "",
		},
		"Present env variable with value": {
			value:         "image-pull-secret",
			expectedValue: "image-pull-secret",
		},
		"Present env variable with multiple value": {
			value:         "image-pull-secret,secret-1",
			expectedValue: "image-pull-secret,secret-1",
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
				os.Setenv(string(ProvisionerImagePullSecrets), v.value)
			}
			actualValue := getOpenEBSImagePullSecrets()
			if !reflect.DeepEqual(actualValue, v.expectedValue) {
				t.Errorf("expected %s got %s", v.expectedValue, actualValue)
			}
			os.Unsetenv(string(ProvisionerImagePullSecrets))
		})
	}
}
