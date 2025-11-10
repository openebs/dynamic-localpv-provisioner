package app

import (
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	k8sEnv "k8s.io/utils/env"
)

//This file defines the environement variable names that are specific
// to this provisioner. In addition to the variables defined in this file,
// provisioner also uses the following:
//   OPENEBS_NAMESPACE
//   NODE_NAME
//   OPENEBS_SERVICE_ACCOUNT
//   OPENEBS_IO_K8S_MASTER
//   OPENEBS_IO_KUBE_CONFIG

const (
	// ProvisionerHelperImage is the environment variable that provides the
	// container image to be used to launch the help pods managing the
	// host path
	ProvisionerHelperImage menv.ENVKey = "OPENEBS_IO_HELPER_IMAGE"

	// ProvisionerHelperPodHostNetwork is the environment variable that provides the
	// host network mode to be used to launch the help pods
	ProvisionerHelperPodHostNetwork string = "OPENEBS_IO_HELPER_POD_HOST_NETWORK"

	// ProvisionerBasePath is the environment variable that provides the
	// default base path on the node where host-path PVs will be provisioned.
	ProvisionerBasePath menv.ENVKey = "OPENEBS_IO_BASE_PATH"

	// ProvisionerImagePullSecrets is the environment variable that provides the
	// init pod to use as authentication when pulling helper image, it is used in the scene where authentication is required
	ProvisionerImagePullSecrets menv.ENVKey = "OPENEBS_IO_IMAGE_PULL_SECRETS"

	// ProvisionerEnablePVCManager enables the PVC Manager mode instead of helper pods
	ProvisionerEnablePVCManager string = "OPENEBS_IO_ENABLE_PVC_MANAGER"

	// ProvisionerPVCManagerPort is the port where PVC Manager service listens
	ProvisionerPVCManagerPort string = "OPENEBS_IO_PVC_MANAGER_PORT"
)

var (
	defaultHelperImage    = "openebs/linux-utils:latest"
	defaultBasePath       = "/var/openebs/local"
	defaultPVCManagerPort = "8080"
)

func getOpenEBSNamespace() string {
	return menv.Get(menv.OpenEBSNamespace)
}
func getDefaultHelperImage() string {
	return menv.GetOrDefault(ProvisionerHelperImage, string(defaultHelperImage))
}
func getHelperPodHostNetwork() bool {
	val, _ := k8sEnv.GetBool(ProvisionerHelperPodHostNetwork, false)
	return val
}

func getDefaultBasePath() string {
	return menv.GetOrDefault(ProvisionerBasePath, string(defaultBasePath))
}

func getOpenEBSServiceAccountName() string {
	return menv.Get(menv.OpenEBSServiceAccount)
}
func getOpenEBSImagePullSecrets() string {
	return menv.Get(ProvisionerImagePullSecrets)
}

// GetEnv gets an environment variable value
func GetEnv(key string) string {
	return menv.Get(menv.ENVKey(key))
}

// getPVCManagerEnabled returns whether PVC Manager mode is enabled
func getPVCManagerEnabled() bool {
	val, _ := k8sEnv.GetBool(ProvisionerEnablePVCManager, true) // Default to true for new architecture
	return val
}

// getPVCManagerPort returns the port for PVC Manager service
func getPVCManagerPort() string {
	return k8sEnv.GetString(ProvisionerPVCManagerPort, defaultPVCManagerPort)
}
