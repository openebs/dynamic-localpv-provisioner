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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// CStorVolumeClaim describes a cstor volume claim resource created as
// custom resource. CStorVolumeClaim is a request for creating cstor volume
// related resources like deployment, svc etc.
type CStorVolumeClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec defines a specification of a cstor volume claim required
	// to provisione cstor volume resources
	Spec CStorVolumeClaimSpec `json:"spec"`

	// Publish contains info related to attachment of a volume to a node.
	// i.e. NodeId etc.
	Publish CStorVolumeClaimPublish `json:"publish,omitempty"`

	// Status represents the current information/status for the cstor volume
	// claim, populated by the controller.
	Status         CStorVolumeClaimStatus `json:"status"`
	VersionDetails VersionDetails         `json:"versionDetails"`
}

// CStorVolumeClaimSpec is the spec for a CStorVolumeClaim resource
type CStorVolumeClaimSpec struct {
	// Capacity represents the actual resources of the underlying
	// cstor volume.
	Capacity corev1.ResourceList `json:"capacity"`
	// ReplicaCount represents the actual replica count for the underlying
	// cstor volume
	ReplicaCount int `json:"replicaCount"`
	// CStorVolumeRef has the information about where CstorVolumeClaim
	// is created from.
	CStorVolumeRef *corev1.ObjectReference `json:"cstorVolumeRef,omitempty"`
	// CstorVolumeSource contains the source volumeName@snapShotname
	// combaination.  This will be filled only if it is a clone creation.
	CstorVolumeSource string `json:"cstorVolumeSource,omitempty"`
	// Policy contains volume specific required policies target and replicas
	Policy CStorVolumePolicySpec `json:"policy"`
}

// CStorVolumeClaimPublish contains info related to attachment of a volume to a node.
// i.e. NodeId etc.
type CStorVolumeClaimPublish struct {
	// NodeID contains publish info related to attachment of a volume to a node.
	NodeID string `json:"nodeId,omitempty"`
}

// CStorVolumeClaimPhase represents the current phase of CStorVolumeClaim.
type CStorVolumeClaimPhase string

const (
	//CStorVolumeClaimPhasePending indicates that the cvc is still waiting for
	//the cstorvolume to be created and bound
	CStorVolumeClaimPhasePending CStorVolumeClaimPhase = "Pending"

	//CStorVolumeClaimPhaseBound indiacates that the cstorvolume has been
	//provisioned and bound to the cstor volume claim
	CStorVolumeClaimPhaseBound CStorVolumeClaimPhase = "Bound"

	//CStorVolumeClaimPhaseFailed indiacates that the cstorvolume provisioning
	//has failed
	CStorVolumeClaimPhaseFailed CStorVolumeClaimPhase = "Failed"
)

// CStorVolumeClaimStatus is for handling status of CstorVolume Claim.
// defines the observed state of CStorVolumeClaim
type CStorVolumeClaimStatus struct {
	// Phase represents the current phase of CStorVolumeClaim.
	Phase CStorVolumeClaimPhase `json:"phase"`
	// Capacity the actual resources of the underlying volume.
	Capacity   corev1.ResourceList         `json:"capacity,omitempty"`
	Conditions []CStorVolumeClaimCondition `json:"condition,omitempty"`
	// PoolInfo represents current pool names where volume replicas exists
	PoolInfo []string `json:"poolInfo"`
}

// CStorVolumeClaimCondition contains details about state of cstor volume
type CStorVolumeClaimCondition struct {
	// Current Condition of cstor volume claim. If underlying persistent volume is being
	// resized then the Condition will be set to 'ResizeStarted' etc
	Type CStorVolumeClaimConditionType `json:"type"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Reason is a brief CamelCase string that describes any failure
	Reason string `json:"reason"`
	// Human-readable message indicating details about last transition.
	Message string `json:"message"`
}

// CStorVolumeClaimConditionType is a valid value of CstorVolumeClaimCondition.Type
type CStorVolumeClaimConditionType string

// These constants are CVC condition types related to resize operation.
const (
	// CStorVolumeClaimResizePending ...
	CStorVolumeClaimResizing CStorVolumeClaimConditionType = "Resizing"
	// CStorVolumeClaimResizeFailed ...
	CStorVolumeClaimResizeFailed CStorVolumeClaimConditionType = "VolumeResizeFailed"
	// CStorVolumeClaimResizeSuccess ...
	CStorVolumeClaimResizeSuccess CStorVolumeClaimConditionType = "VolumeResizeSuccessful"
	// CStorVolumeClaimResizePending ...
	CStorVolumeClaimResizePending CStorVolumeClaimConditionType = "VolumeResizePending"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// CStorVolumeClaimList is a list of CStorVolumeClaim resources
type CStorVolumeClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorVolumeClaim `json:"items"`
}
