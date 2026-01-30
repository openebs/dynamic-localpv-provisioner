package utils

import (
	menv "github.com/openebs/lib-csi/pkg/common/env"
	k8sEnv "k8s.io/utils/env"
)

// GoogleAnalyticsEnabled parses a boolean value from an env. Defaults to false, in case of an error.
func GoogleAnalyticsEnabled(env string) bool {
	enabled, _ := k8sEnv.GetBool(env, false)
	return enabled
}

// GetStringEnvOrDefault fetches the value of an env. If unset or empty, it returns the default value.
func GetStringEnvOrDefault(key, defaultValue string) string {
	val := menv.Get(key)
	if len(val) == 0 {
		return defaultValue
	}
	return val
}
