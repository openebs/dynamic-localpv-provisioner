/*
Copyright 2019 The OpenEBS Authors

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

package app

import (
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
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

	// ProvisionerBasePath is the environment variable that provides the
	// default base path on the node where host-path PVs will be provisioned.
	ProvisionerBasePath menv.ENVKey = "OPENEBS_IO_BASE_PATH"

	// ProvisionerImagePullSecrets is the environment variable that provides the
	// init pod to use as authentication when pulling helper image, it is used in the scene where authentication is required
	ProvisionerImagePullSecrets menv.ENVKey = "OPENEBS_IO_IMAGE_PULL_SECRETS"
)

var (
	defaultHelperImage = "quay.io/openebs/linux-utils:latest"
	defaultBasePath    = "/var/openebs/local"
)

func getOpenEBSNamespace() string {
	return menv.Get(menv.OpenEBSNamespace)
}
func getDefaultHelperImage() string {
	return menv.GetOrDefault(ProvisionerHelperImage, string(defaultHelperImage))
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
