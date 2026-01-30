package app

import (
	menv "github.com/openebs/lib-csi/pkg/common/env"
	k8sEnv "k8s.io/utils/env"

	"github.com/openebs/dynamic-localpv-provisioner/pkg/utils"
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
	ProvisionerHelperImage string = "OPENEBS_IO_HELPER_IMAGE"

	// ProvisionerHelperPodHostNetwork is the environment variable that provides the
	// host network mode to be used to launch the help pods
	ProvisionerHelperPodHostNetwork string = "OPENEBS_IO_HELPER_POD_HOST_NETWORK"

	// ProvisionerBasePath is the environment variable that provides the
	// default base path on the node where host-path PVs will be provisioned.
	ProvisionerBasePath string = "OPENEBS_IO_BASE_PATH"

	// ProvisionerImagePullSecrets is the environment variable that provides the
	// init pod to use as authentication when pulling helper image, it is used in the scene where authentication is required
	ProvisionerImagePullSecrets string = "OPENEBS_IO_IMAGE_PULL_SECRETS"

	// OpenebsNamespace is the env where we read the kubernetes namespace of the Pod.
	//
	// This environment variable is set via kubernetes downward API
	OpenebsNamespace string = "OPENEBS_NAMESPACE"

	// OpenebsServiceAccount is the environment variable to get openebs
	// serviceaccount.
	//
	// This environment variable is set via kubernetes downward API.
	OpenebsServiceAccount string = "OPENEBS_SERVICE_ACCOUNT"
)

var (
	defaultHelperImage = "openebs/linux-utils:latest"
	defaultBasePath    = "/var/openebs/local"
)

func getOpenEBSNamespace() string {
	return menv.Get(OpenebsNamespace)
}
func getDefaultHelperImage() string {
	return utils.GetStringEnvOrDefault(ProvisionerHelperImage, defaultHelperImage)
}
func getHelperPodHostNetwork() bool {
	val, _ := k8sEnv.GetBool(ProvisionerHelperPodHostNetwork, false)
	return val
}

func getDefaultBasePath() string {
	return utils.GetStringEnvOrDefault(ProvisionerBasePath, defaultBasePath)
}

func getOpenEBSServiceAccountName() string {
	return menv.Get(OpenebsServiceAccount)
}
func getOpenEBSImagePullSecrets() string {
	return menv.Get(ProvisionerImagePullSecrets)
}

// getNodeName returns the current node name from NODE_NAME environment variable
func getNodeName() string {
	return menv.Get("NODE_NAME")
}
