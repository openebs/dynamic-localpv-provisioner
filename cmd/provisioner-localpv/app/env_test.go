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

func TestGetPodName(t *testing.T) {
	testCases := map[string]struct {
		value         string
		expectedValue string
	}{
		"Missing env variable": {
			value:         "",
			expectedValue: "",
		},
		"Present env variable with value": {
			value:         "openebs-localpv-abc12",
			expectedValue: "openebs-localpv-abc12",
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
				os.Setenv(PodName, v.value)
			}
			actualValue := getPodName()
			if !reflect.DeepEqual(actualValue, v.expectedValue) {
				t.Errorf("expected %s got %s", v.expectedValue, actualValue)
			}
			os.Unsetenv(PodName)
		})
	}
}

func TestGetAnalyticsStateCMName(t *testing.T) {
	testCases := map[string]struct {
		value         string
		expectedValue string
	}{
		"Missing env variable": {
			value:         "",
			expectedValue: "",
		},
		"Present env variable with value": {
			value:         "openebs-localpv-analytics-state",
			expectedValue: "openebs-localpv-analytics-state",
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
				os.Setenv(AnalyticsStateCM, v.value)
			}
			actualValue := getAnalyticsStateCMName()
			if !reflect.DeepEqual(actualValue, v.expectedValue) {
				t.Errorf("expected %s got %s", v.expectedValue, actualValue)
			}
			os.Unsetenv(AnalyticsStateCM)
		})
	}
}

func TestGetClientQPS(t *testing.T) {
	testCases := map[string]struct {
		value     string
		setEnv    bool
		expectQPS float32
		expectOK  bool
	}{
		"Missing env variable leaves default": {
			setEnv:    false,
			expectQPS: 0,
			expectOK:  false,
		},
		"Empty env variable leaves default": {
			value:     "",
			setEnv:    true,
			expectQPS: 0,
			expectOK:  false,
		},
		"Valid integer value": {
			value:     "50",
			setEnv:    true,
			expectQPS: 50,
			expectOK:  true,
		},
		"Valid fractional value": {
			value:     "7.5",
			setEnv:    true,
			expectQPS: 7.5,
			expectOK:  true,
		},
		"Invalid value leaves default": {
			value:     "not-a-number",
			setEnv:    true,
			expectQPS: 0,
			expectOK:  false,
		},
		"Zero value leaves default": {
			value:     "0",
			setEnv:    true,
			expectQPS: 0,
			expectOK:  false,
		},
		"Negative value leaves default": {
			value:     "-1",
			setEnv:    true,
			expectQPS: 0,
			expectOK:  false,
		},
	}

	for k, v := range testCases {
		v := v
		t.Run(k, func(t *testing.T) {
			if v.setEnv {
				os.Setenv(ProvisionerClientQPS, v.value)
			}
			qps, ok := getClientQPS()
			if qps != v.expectQPS || ok != v.expectOK {
				t.Errorf("expected (%v, %v) got (%v, %v)", v.expectQPS, v.expectOK, qps, ok)
			}
			os.Unsetenv(ProvisionerClientQPS)
		})
	}
}

func TestGetClientBurst(t *testing.T) {
	testCases := map[string]struct {
		value       string
		setEnv      bool
		expectBurst int
		expectOK    bool
	}{
		"Missing env variable leaves default": {
			setEnv:      false,
			expectBurst: 0,
			expectOK:    false,
		},
		"Empty env variable leaves default": {
			value:       "",
			setEnv:      true,
			expectBurst: 0,
			expectOK:    false,
		},
		"Valid value": {
			value:       "100",
			setEnv:      true,
			expectBurst: 100,
			expectOK:    true,
		},
		"Invalid value leaves default": {
			value:       "not-a-number",
			setEnv:      true,
			expectBurst: 0,
			expectOK:    false,
		},
		"Fractional value leaves default": {
			value:       "10.5",
			setEnv:      true,
			expectBurst: 0,
			expectOK:    false,
		},
		"Zero value leaves default": {
			value:       "0",
			setEnv:      true,
			expectBurst: 0,
			expectOK:    false,
		},
		"Negative value leaves default": {
			value:       "-1",
			setEnv:      true,
			expectBurst: 0,
			expectOK:    false,
		},
	}

	for k, v := range testCases {
		v := v
		t.Run(k, func(t *testing.T) {
			if v.setEnv {
				os.Setenv(ProvisionerClientBurst, v.value)
			}
			burst, ok := getClientBurst()
			if burst != v.expectBurst || ok != v.expectOK {
				t.Errorf("expected (%v, %v) got (%v, %v)", v.expectBurst, v.expectOK, burst, ok)
			}
			os.Unsetenv(ProvisionerClientBurst)
		})
	}
}

func TestGetAnalyticsLeaseName(t *testing.T) {
	testCases := map[string]struct {
		value         string
		expectedValue string
	}{
		"Missing env variable falls back to default": {
			value:         "",
			expectedValue: defaultAnalyticsLeaseName,
		},
		"Present env variable with value": {
			value:         "mything-localpv-analytics",
			expectedValue: "mything-localpv-analytics",
		},
		"Present env variable with whitespaces falls back to default": {
			value:         " ",
			expectedValue: defaultAnalyticsLeaseName,
		},
	}
	for k, v := range testCases {
		v := v
		t.Run(k, func(t *testing.T) {
			if len(v.value) != 0 {
				os.Setenv(AnalyticsLease, v.value)
			}
			actualValue := getAnalyticsLeaseName()
			if !reflect.DeepEqual(actualValue, v.expectedValue) {
				t.Errorf("expected %s got %s", v.expectedValue, actualValue)
			}
			os.Unsetenv(AnalyticsLease)
		})
	}
}
