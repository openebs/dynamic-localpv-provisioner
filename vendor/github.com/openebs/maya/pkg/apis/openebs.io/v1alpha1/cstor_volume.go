/*
Copyright 2018 The OpenEBS Authors.

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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolume

// CStorVolume describes a cstor volume resource created as custom resource
type CStorVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorVolumeSpec   `json:"spec"`
	Status            CStorVolumeStatus `json:"status"`
	VersionDetails    VersionDetails    `json:"versionDetails"`
}

// CStorVolumeSpec is the spec for a CStorVolume resource
type CStorVolumeSpec struct {
	// Capacity represents the desired size of the underlying volume.
	Capacity          resource.Quantity `json:"capacity"`
	TargetIP          string            `json:"targetIP"`
	TargetPort        string            `json:"targetPort"`
	Iqn               string            `json:"iqn"`
	TargetPortal      string            `json:"targetPortal"`
	Status            string            `json:"status"`
	NodeBase          string            `json:"nodeBase"`
	ReplicationFactor int               `json:"replicationFactor"`
	ConsistencyFactor int               `json:"consistencyFactor"`
	// DesiredReplicationFactor represents maximum number of replicas
	// that are allowed to connect to the target
	DesiredReplicationFactor int `json:"desiredReplicationFactor"`
	//ReplicaDetails refers to the trusty replica information
	ReplicaDetails CStorVolumeReplicaDetails `json:"replicaDetails,omitempty"`
}

// ReplicaID is to hold replicaID information
type ReplicaID string

// CStorVolumePhase is to hold result of action.
type CStorVolumePhase string

// CStorVolumeStatus is for handling status of cvr.
type CStorVolumeStatus struct {
	Phase           CStorVolumePhase `json:"phase"`
	ReplicaStatuses []ReplicaStatus  `json:"replicaStatuses,omitempty"`
	// Represents the actual resources of the underlying volume.
	Capacity resource.Quantity `json:"capacity,omitempty"`
	// LastTransitionTime refers to the time when the phase changes
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	LastUpdateTime     metav1.Time `json:"lastUpdateTime,omitempty"`
	Message            string      `json:"message,omitempty"`
	// Current Condition of cstorvolume. If underlying persistent volume is being
	// resized then the Condition will be set to 'ResizePending'.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []CStorVolumeCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,4,rep,name=conditions"`
	// ReplicaDetails refers to the trusty replica information which are
	// connected at given time
	ReplicaDetails CStorVolumeReplicaDetails `json:"replicaDetails,omitempty"`
}

// CStorVolumeReplicaDetails contains trusty replica inform which will be
// updated by target
type CStorVolumeReplicaDetails struct {
	// KnownReplicas represents the replicas that target can trust to read data
	KnownReplicas map[ReplicaID]string `json:"knownReplicas,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolume

// CStorVolumeList is a list of CStorVolume resources
type CStorVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorVolume `json:"items"`
}

// CVStatusResponse stores the reponse of istgt replica command output
// It may contain several volumes
type CVStatusResponse struct {
	CVStatuses []CVStatus `json:"volumeStatus"`
}

// CVStatus stores the status of a CstorVolume obtained from response
type CVStatus struct {
	Name            string          `json:"name"`
	Status          string          `json:"status"`
	ReplicaStatuses []ReplicaStatus `json:"replicaStatus"`
}

// ReplicaStatus stores the status of replicas
type ReplicaStatus struct {
	ID                string `json:"replicaId"`
	Mode              string `json:"mode"`
	CheckpointedIOSeq string `json:"checkpointedIOSeq"`
	InflightRead      string `json:"inflightRead"`
	InflightWrite     string `json:"inflightWrite"`
	InflightSync      string `json:"inflightSync"`
	UpTime            int    `json:"upTime"`
	Quorum            string `json:"quorum"`
}

// CStorVolumeCondition contains details about state of cstorvolume
type CStorVolumeCondition struct {
	Type   CStorVolumeConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=CStorVolumeConditionType"`
	Status ConditionStatus          `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// Unique, this should be a short, machine understandable string that gives the reason
	// for condition's last transition. If it reports "ResizePending" that means the underlying
	// cstorvolume is being resized.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// CStorVolumeConditionType is a valid value of CStorVolumeCondition.Type
type CStorVolumeConditionType string

const (
	// CStorVolumeResizing - a user trigger resize of pvc has been started
	CStorVolumeResizing CStorVolumeConditionType = "Resizing"
)

// ConditionStatus states in which state condition is present
type ConditionStatus string

// These are valid condition statuses. "ConditionInProgress" means corresponding
// condition is inprogress. "ConditionSuccess" means corresponding condition is success
const (
	// ConditionInProgress states resize of underlying volumes are in progress
	ConditionInProgress ConditionStatus = "InProgress"
	// ConditionSuccess states resizing underlying volumes are successfull
	ConditionSuccess ConditionStatus = "Success"
)
