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
//   POD_NAME
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

	// ProvisionerImagePullPolicy is the environment variable that provides the
	// helper pod container the imagePullPolicy to be used
	ProvisionerImagePullPolicy string = "OPENEBS_IO_IMAGE_PULL_POLICY"

	// OpenebsNamespace is the env where we read the kubernetes namespace of the Pod.
	//
	// This environment variable is set via kubernetes downward API
	OpenebsNamespace string = "OPENEBS_NAMESPACE"

	// OpenebsServiceAccount is the environment variable to get openebs
	// serviceaccount.
	//
	// This environment variable is set via kubernetes downward API.
	OpenebsServiceAccount string = "OPENEBS_SERVICE_ACCOUNT"

	// PodName is the name of the current pod, set via the Kubernetes
	// downward API. Used as the leader-election identity for analytics in
	// node-deployment mode so that each DaemonSet pod has a stable, unique
	// identity for lease acquisition.
	PodName string = "POD_NAME"

	// AnalyticsStateCM is the environment variable that names the
	// ConfigMap used to record whether the analytics install event has
	// been sent for the current Helm release. Both deployment modes
	// consult this CM so a pod restart (helper-pod mode) or a leadership
	// transition (node-deployment mode) does not re-emit install.
	AnalyticsStateCM string = "OPENEBS_IO_ANALYTICS_STATE_CM"

	// AnalyticsLease is the environment variable that names the Lease
	// used to elect the single analytics emitter in node-deployment mode.
	// Set by the chart via include "localpv.analyticsLease.name" so the
	// Lease name is release-scoped, matching the rest of the chart's
	// naming convention.
	AnalyticsLease string = "OPENEBS_IO_ANALYTICS_LEASE"

	// ProvisionerWorkerThreads is the environment variable that controls the
	// number of concurrent worker goroutines for processing PVC create and
	// PV delete events. Higher values increase provisioning throughput when
	// many PVCs are created simultaneously. Default is 4.
	ProvisionerWorkerThreads string = "OPENEBS_IO_WORKER_THREADS"

	// ProvisionerHelperPodTimeout is the environment variable that controls
	// the maximum number of seconds to wait for a helper pod (init, cleanup,
	// quota) to complete before timing out. Default is 120.
	ProvisionerHelperPodTimeout string = "OPENEBS_IO_HELPER_POD_TIMEOUT_SECS"
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
func getOpenEBSImagePullPolicy() string {
	return menv.Get(ProvisionerImagePullPolicy)
}

// getNodeName returns the current node name from NODE_NAME environment variable
func getNodeName() string {
	return menv.Get("NODE_NAME")
}

// getPodName returns the current pod name from the POD_NAME environment
// variable (set via the Kubernetes downward API). Returns an empty string if
// POD_NAME is unset; callers should fall back to os.Hostname() in that case.
func getPodName() string {
	return menv.Get(PodName)
}

// getAnalyticsStateCMName returns the analytics-state ConfigMap name from
// OPENEBS_IO_ANALYTICS_STATE_CM. Returns "" when unset, in which case
// maybeEmitInstall treats this as a non-Helm deployment and emits install
// unconditionally (legacy behavior).
func getAnalyticsStateCMName() string {
	return menv.Get(AnalyticsStateCM)
}

// defaultAnalyticsLeaseName is the fallback Lease name when the chart did
// not provide one via OPENEBS_IO_ANALYTICS_LEASE. Matches the pre-Helm-
// templated name so non-Helm deployers see no behavior change.
const defaultAnalyticsLeaseName = "openebs-localpv-analytics"

// getAnalyticsLeaseName returns the analytics Lease name from
// OPENEBS_IO_ANALYTICS_LEASE, or defaultAnalyticsLeaseName if unset. Unlike
// the state CM, the Lease has a hardcoded fallback because an empty name
// would disable leader election entirely — and with it, all analytics in
// node-deployment mode, which is worse than an inconsistent name.
func getAnalyticsLeaseName() string {
	if name := menv.Get(AnalyticsLease); name != "" {
		return name
	}
	return defaultAnalyticsLeaseName
}

func getWorkerThreads() int {
	val, err := k8sEnv.GetInt(ProvisionerWorkerThreads, 4)
	if err != nil || val < 2 {
		return 4
	}
	return val
}

func getHelperPodTimeout() int {
	val, err := k8sEnv.GetInt(ProvisionerHelperPodTimeout, 120)
	if err != nil || val < 1 {
		return 120
	}
	return val
}
