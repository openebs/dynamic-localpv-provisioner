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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// VersionDetails provides the details for upgrade
type VersionDetails struct {
	// If AutoUpgrade is set to true then the resource is
	// upgraded automatically without any manual steps
	AutoUpgrade bool `json:"autoUpgrade"`
	// Desired is the version that we want to
	// upgrade or the control plane version
	Desired string `json:"desired"`
	// Status gives the status of reconciliation triggered
	// when the desired and current version are not same
	Status VersionStatus `json:"status"`
}

// VersionStatus is the status of the reconciliation of versions
type VersionStatus struct {
	// DependentsUpgraded gives the details whether all children
	// of a resource are upgraded to desired version or not
	DependentsUpgraded bool `json:"dependentsUpgraded"`
	// Current is the version of resource
	Current string `json:"current"`
	// State is the state of reconciliation
	State VersionState `json:"state"`
	// Message is a human readable message if some error occurs
	Message string `json:"message,omitempty"`
	// Reason is the actual reason for the error state
	Reason string `json:"reason,omitempty"`
	// LastUpdateTime is the time the status was last  updated
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

// VersionState is the state of reconciliation
type VersionState string

const (
	// ReconcileComplete is the state when desired and current version are equal.
	ReconcileComplete VersionState = "Reconciled"
	// ReconcileInProgress is the state when desired and current version are
	// not same and the reconcile functions is retrying to make them same.
	ReconcileInProgress VersionState = "ReconcileInProgress"
	// ReconcilePending is the state the reconciliation is still not started yet
	ReconcilePending VersionState = "ReconcilePending"
)
