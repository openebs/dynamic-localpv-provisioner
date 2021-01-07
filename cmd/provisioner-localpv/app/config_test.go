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
