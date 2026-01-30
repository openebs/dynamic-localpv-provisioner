package app

import (
	"os"
	"reflect"
	"strconv"
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
				os.Setenv(OpenebsNamespace, v.value)
			}
			actualValue := getOpenEBSNamespace()
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
			os.Unsetenv(OpenebsNamespace)
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
				os.Setenv(ProvisionerHelperImage, v.value)
			}
			actualValue := getDefaultHelperImage()
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
			os.Unsetenv(ProvisionerHelperImage)
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
				os.Setenv(ProvisionerHelperPodHostNetwork, v.value)
			}
			actualValue := strconv.FormatBool(getHelperPodHostNetwork())
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
			os.Unsetenv(ProvisionerHelperPodHostNetwork)
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
				os.Setenv(ProvisionerBasePath, v.value)
			}
			actualValue := getDefaultBasePath()
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
			os.Unsetenv(ProvisionerBasePath)
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
				os.Setenv(OpenebsServiceAccount, v.value)
			}
			actualValue := getOpenEBSServiceAccountName()
			if !reflect.DeepEqual(actualValue, v.expectedValue) {
				t.Errorf("expected %s got %s", v.expectedValue, actualValue)
			}
			os.Unsetenv(OpenebsServiceAccount)
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
				os.Setenv(ProvisionerImagePullSecrets, v.value)
			}
			actualValue := getOpenEBSImagePullSecrets()
			if !reflect.DeepEqual(actualValue, v.expectedValue) {
				t.Errorf("expected %s got %s", v.expectedValue, actualValue)
			}
			os.Unsetenv(ProvisionerImagePullSecrets)
		})
	}
}
