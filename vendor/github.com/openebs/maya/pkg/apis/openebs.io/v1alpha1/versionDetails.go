/*
Copyright 2019 The OpenEBS Authors.

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
	"strings"

	"github.com/openebs/maya/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	validCurrentVersions = map[string]bool{
		"1.0.0": true, "1.1.0": true, "1.2.0": true, "1.3.0": true,
		"1.4.0": true, "1.5.0": true, "1.6.0": true, "1.7.0": true,
		"1.8.0": true, "1.9.0": true, "1.10.0": true, "1.11.0": true,
		"1.12.0": true, "2.0.0": true, "2.1.0": true, "2.2.0": true,
		"2.3.0": true, "2.4.0": true, "2.4.1": true, "2.5.0": true,
		"2.6.0": true, "2.7.0": true, "2.8.0": true, "2.9.0": true,
		"2.10.0": true, "2.11.0": true, "2.12.0": true,
	}
	validDesiredVersion = strings.Split(version.GetVersion(), "-")[0]
)

// IsCurrentVersionValid verifies if the  current version is valid or not
func IsCurrentVersionValid(v string) bool {
	currentVersion := strings.Split(v, "-")[0]
	return validCurrentVersions[currentVersion]
}

// IsDesiredVersionValid verifies the desired version is valid or not
func IsDesiredVersionValid(v string) bool {
	desiredVersion := strings.Split(v, "-")[0]
	return validDesiredVersion == desiredVersion
}

// SetErrorStatus sets the message and reason for the error
func (vs *VersionStatus) SetErrorStatus(msg string, err error) {
	vs.Message = msg
	vs.Reason = err.Error()
	vs.LastUpdateTime = metav1.Now()
}

// SetInProgressStatus sets the state as ReconcileInProgress
func (vs *VersionStatus) SetInProgressStatus() {
	vs.State = ReconcileInProgress
	vs.LastUpdateTime = metav1.Now()
}

// SetSuccessStatus resets the message and reason and sets the state as
// Reconciled
func (vd *VersionDetails) SetSuccessStatus() {
	vd.Status.Current = vd.Desired
	vd.Status.Message = ""
	vd.Status.Reason = ""
	vd.Status.State = ReconcileComplete
	vd.Status.LastUpdateTime = metav1.Now()
}
