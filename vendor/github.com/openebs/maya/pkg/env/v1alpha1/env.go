/*
Copyright 2018 The OpenEBS Authors

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

package v1alpha1

import (
	"os"
	"strconv"
	"strings"
)

// ENVKey is a typed string to represent various environment keys used by this
// binary
type ENVKey string

const (
	// KubeConfig is the ENV variable to fetch kubernetes kubeconfig
	KubeConfig ENVKey = "OPENEBS_IO_KUBE_CONFIG"

	// KubeMaster is the ENV variable to fetch kubernetes master's address
	KubeMaster ENVKey = "OPENEBS_IO_K8S_MASTER"

	// OpenEBSEnableAnalytics is the environment variable to get user's consent to
	// send usage data to OpenEBS core-developers using the Google Analytics platform
	OpenEBSEnableAnalytics ENVKey = "OPENEBS_IO_ENABLE_ANALYTICS"

	// OpenEBSVersion is the environment variable to get openebs version
	OpenEBSVersion ENVKey = "OPENEBS_IO_VERSION"

	// OpenEBSNamespace is the environment variable to get openebs namespace
	//
	// This environment variable is set via kubernetes downward API
	OpenEBSNamespace ENVKey = "OPENEBS_NAMESPACE"

	// Namespace is the environment variable to get resource namespace
	//
	// This environment variable is set via kubernetes downward API
	Namespace ENVKey = "NAMESPACE"

	// OpenEBSBaseDir is the environment variable to get base directory of
	// openebs
	OpenEBSBaseDir ENVKey = "OPENEBS_IO_BASE_DIR"

	// OpenEBSMayaPodName is the environment variable to get maya-apiserver pod
	// name
	//
	// This environment variable is set via kubernetes downward API
	OpenEBSMayaPodName ENVKey = "OPENEBS_MAYA_POD_NAME"

	// CSPCOperatorPodName is the environment variable to get cspc-operator pod
	// name
	//
	// This environment variable is set via kubernetes downward API
	CSPCOperatorPodName ENVKey = "CSPC_OPERATOR_POD_NAME"

	// OpenEBSServiceAccount is the environment variable to get openebs
	// serviceaccount
	//
	// This environment variable is set via kubernetes downward API
	OpenEBSServiceAccount ENVKey = "OPENEBS_SERVICE_ACCOUNT"

	// TODO:
	//
	// The constants present here should be moved to respective/relevant packages
	// This file will hold env related operations as well as environment variables
	// that is common.
	//
	// All these might be moved to
	//
	// pkg/<entity_name>/v1alpha1/env_<entity_name>_<introduced_at_version>.go
	//
	// Need to discuss with team on above !!!
	//
	// TODO:
	//
	// The names of these variables should also change. It need not be suffixed
	// with ENVK. Need to discuss with team.

	// CASTemplateFeatureGateENVK is the ENV key to fetch cas template feature gate
	// i.e. if cas template based provisioning is enabled or disabled
	CASTemplateFeatureGateENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_FEATURE_GATE"

	// CASTemplateToListVolumeENVK is the ENV key that specifies the CAS Template
	// to list cas volumes
	CASTemplateToListVolumeENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_LIST_VOLUME"

	// CASTemplateToCreateJivaVolumeENVK is the ENV key that specifies the CAS Template
	// to create jiva cas volumes
	CASTemplateToCreateJivaVolumeENVK ENVKey = "OPENEBS_IO_JIVA_CAS_TEMPLATE_TO_CREATE_VOLUME"

	// CASTemplateToReadJivaVolumeENVK is the ENV key that specifies the CAS Template
	// to read jiva cas volumes
	CASTemplateToReadJivaVolumeENVK ENVKey = "OPENEBS_IO_JIVA_CAS_TEMPLATE_TO_READ_VOLUME"

	// CASTemplateToDeleteJivaVolumeENVK is the ENV key that specifies the CAS Template
	// to delete jiva cas volumes
	CASTemplateToDeleteJivaVolumeENVK ENVKey = "OPENEBS_IO_JIVA_CAS_TEMPLATE_TO_DELETE_VOLUME"

	// CASTemplateToCreateCStorVolumeENVK is the ENV key that specifies the CAS Template
	// to create cstor cas volumes
	CASTemplateToCreateCStorVolumeENVK ENVKey = "OPENEBS_IO_CSTOR_CAS_TEMPLATE_TO_CREATE_VOLUME"

	// CASTemplateToReadCStorVolumeENVK is the ENV key that specifies the CAS Template
	// to read cstor cas volumes
	CASTemplateToReadCStorVolumeENVK ENVKey = "OPENEBS_IO_CSTOR_CAS_TEMPLATE_TO_READ_VOLUME"

	// CASTemplateToDeleteCStorVolumeENVK is the ENV key that specifies the CAS Template
	// to delete cstor cas volumes
	CASTemplateToDeleteCStorVolumeENVK ENVKey = "OPENEBS_IO_CSTOR_CAS_TEMPLATE_TO_DELETE_VOLUME"

	// CASTemplateToCreatePoolENVK is the ENV key that specifies the CAS Template
	// to create storage pool
	CASTemplateToCreatePoolENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_CREATE_POOL"

	// CASTemplateToDeletePoolENVK is the ENV key that specifies the CAS Template
	// to delete storage pool
	CASTemplateToDeletePoolENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_DELETE_POOL"

	// CASTemplateToListCStorSnapshotENVK is the ENV key that specifies the CAS Template
	// to list cstor cas snapshots
	CASTemplateToListCStorSnapshotENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_LIST_CSTOR_SNAPSHOT"

	// CASTemplateToListJivaSnapshotENVK is the ENV key that specifies the CAS Template
	// to list jiva cas snapshots
	CASTemplateToListJivaSnapshotENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_LIST_JIVA_SNAPSHOT"

	// CASTemplateToCreateJivaSnapshotENVK is the ENV key that specifies the CAS Template
	// to create jiva cas snapshot
	CASTemplateToCreateJivaSnapshotENVK ENVKey = "OPENEBS_IO_JIVA_CAS_TEMPLATE_TO_CREATE_SNAPSHOT"

	// CASTemplateToReadJivaSnapshotENVK is the ENV key that specifies the CAS Template
	// to read jiva cas snapshot
	CASTemplateToReadJivaSnapshotENVK ENVKey = "OPENEBS_IO_JIVA_CAS_TEMPLATE_TO_READ_SNAPSHOT"

	// CASTemplateToDeleteJivaSnapshotENVK is the ENV key that specifies the CAS Template
	// to delete jiva cas snapshot
	CASTemplateToDeleteJivaSnapshotENVK ENVKey = "OPENEBS_IO_JIVA_CAS_TEMPLATE_TO_DELETE_SNAPSHOT"

	// CASTemplateToCreateCStorSnapshotENVK is the ENV key that specifies the CAS Template
	// to create cstor cas snapshot
	CASTemplateToCreateCStorSnapshotENVK ENVKey = "OPENEBS_IO_CSTOR_CAS_TEMPLATE_TO_CREATE_SNAPSHOT"

	// CASTemplateToReadCStorSnapshotENVK is the ENV key that specifies the CAS Template
	// to read cstor cas snapshot
	CASTemplateToReadCStorSnapshotENVK ENVKey = "OPENEBS_IO_CSTOR_CAS_TEMPLATE_TO_READ_SNAPSHOT"

	// CASTemplateToDeleteCStorSnapshotENVK is the ENV key that specifies the CAS Template
	// to delete cstor cas snapshot
	CASTemplateToDeleteCStorSnapshotENVK ENVKey = "OPENEBS_IO_CSTOR_CAS_TEMPLATE_TO_DELETE_SNAPSHOT"

	// CASTemplateToReadVolumeStatsENVK is the ENV key that specifies the CAS Template
	// to read volume stats
	CASTemplateToReadVolumeStatsENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_READ_CAS_VOLUME_STATS"

	// CASTemplateToListStoragePoolENVK is the ENV key that specifies the CAS Template
	// to list cas sto
	CASTemplateToListStoragePoolENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_LIST_STORAGE_POOL"

	// CASTemplateToReadStoragePoolENVK is the ENV key that specifies the CAS Template
	// to read storagepool
	CASTemplateToReadStoragePoolENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_READ_STORAGE_POOL"
)

// EnvironmentSetter abstracts setting of environment variable
type EnvironmentSetter func(envKey ENVKey, value string) (err error)

// EnvironmentGetter abstracts fetching value from an environment variable
type EnvironmentGetter func(envKey ENVKey) (value string)

// EnvironmentLookup abstracts looking up an environment variable
type EnvironmentLookup func(envKey ENVKey) (value string, present bool)

// Set sets the provided environment variable
//
// NOTE:
//  This is an implementation of EnvironmentSetter
func Set(envKey ENVKey, value string) (err error) {
	return os.Setenv(string(envKey), value)
}

// Get fetches value from the provided environment variable
//
// NOTE:
//  This is an implementation of EnvironmentGetter
func Get(envKey ENVKey) (value string) {
	return getEnv(string(envKey))
}

// GetOrDefault fetches value from the provided environment variable
// which on empty returns the defaultValue
// NOTE: os.Getenv is used here instead of os.LookupEnv because it is
// not required to know if the environment variable is defined on the system
func GetOrDefault(e ENVKey, defaultValue string) (value string) {
	envValue := Get(e)
	if len(envValue) == 0 {
		// ENV not defined or set to ""
		return defaultValue
	}

	return envValue
}

// Lookup looks up an environment variable
//
// NOTE:
//  This is an implementation of EnvironmentLookup
func Lookup(envKey ENVKey) (value string, present bool) {
	return lookupEnv(string(envKey))
}

// Truthy returns boolean based on the environment variable's value
//
// The lookup value can be truthy (i.e. 1, t, TRUE, true) or falsy (0, false,
// etc) based on strconv.ParseBool logic
func Truthy(envKey ENVKey) (truth bool) {
	v, found := Lookup(envKey)
	if !found {
		return
	}
	truth, _ = strconv.ParseBool(v)
	return
}

// LookupOrFalse looks up an environment variable and returns a string "false"
// if environment variable is not present. It returns appropriate values for
// other cases.
func LookupOrFalse(envKey ENVKey) string {
	val, present := Lookup(envKey)
	if !present {
		return "false"
	}
	return strings.TrimSpace(val)
}

// getEnv fetches the provided environment variable's value
func getEnv(envKey string) (value string) {
	return strings.TrimSpace(os.Getenv(envKey))
}

// lookupEnv looks up the provided environment variable
func lookupEnv(envKey string) (value string, present bool) {
	value, present = os.LookupEnv(envKey)
	value = strings.TrimSpace(value)
	return
}

// GetOpenebsBaseDirPath returns the base path to store openebs related files on
// host machine
func GetOpenebsBaseDirPath() string {
	baseDir, isPresent := lookupEnv(string(OpenEBSBaseDir))
	if !isPresent {
		return "/var/openebs"
	}
	return baseDir
}
