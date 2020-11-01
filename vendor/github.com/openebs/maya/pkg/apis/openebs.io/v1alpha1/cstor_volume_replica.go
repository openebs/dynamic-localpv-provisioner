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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CVRKey represents the properties of a cstorvolumereplica
type CVRKey string

const (
	// CloneEnableKEY is used to enable/disable cloning for a cstorvolumereplica
	CloneEnableKEY CVRKey = "openebs.io/cloned"

	// SourceVolumeKey stores the name of source volume whose snapshot is used to
	// create this cvr
	SourceVolumeKey CVRKey = "openebs.io/source-volume"

	// SnapshotNameKey stores the name of the snapshot being used to restore this replica
	SnapshotNameKey CVRKey = "openebs.io/snapshot"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolumereplica

// CStorVolumeReplica describes a cstor volume resource created as custom resource
type CStorVolumeReplica struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorVolumeReplicaSpec   `json:"spec"`
	Status            CStorVolumeReplicaStatus `json:"status"`
	VersionDetails    VersionDetails           `json:"versionDetails"`
}

// CStorVolumeReplicaSpec is the spec for a CStorVolumeReplica resource
type CStorVolumeReplicaSpec struct {
	TargetIP string `json:"targetIP"`
	Capacity string `json:"capacity"`
	// ZvolWorkers represents number of threads that executes client IOs
	ZvolWorkers string `json:"zvolWorkers"`
	// ReplicaID is unique number to identify the replica
	ReplicaID string `json:"replicaid"`
}

// CStorVolumeReplicaPhase is to hold result of action.
type CStorVolumeReplicaPhase string

// Status written onto CStorVolumeReplica objects.
const (

	// CVRStatusEmpty describes CVR resource is created but not yet monitored by
	// controller(i.e resource is just created)
	CVRStatusEmpty CStorVolumeReplicaPhase = ""

	// CVRStatusOnline describes volume replica is Healthy and data existing on
	// the healthy replica is up to date
	CVRStatusOnline CStorVolumeReplicaPhase = "Healthy"

	// CVRStatusOffline describes volume replica is created but not yet connected
	// to the target
	CVRStatusOffline CStorVolumeReplicaPhase = "Offline"

	// CVRStatusDegraded describes volume replica is connected to the target and
	// rebuilding from other replicas is not yet started but ready for serving
	// IO's
	CVRStatusDegraded CStorVolumeReplicaPhase = "Degraded"

	// CVRStatusNewReplicaDegraded describes replica is recreated (due to pool
	// recreation[underlying disk got changed]/volume replica scaleup cases) and
	// just connected to the target. Volume replica has to start reconstructing
	// entier data from another available healthy replica. Until volume replica
	// becomes healthy whatever data written to it is lost(NewReplica also not part
	// of any quorum decision)
	CVRStatusNewReplicaDegraded CStorVolumeReplicaPhase = "NewReplicaDegraded"

	// CVRStatusRebuilding describes volume replica has missing data and it
	// started rebuilding missing data from other replicas
	CVRStatusRebuilding CStorVolumeReplicaPhase = "Rebuilding"

	// CVRStatusReconstructingNewReplica describes volume replica is recreated
	// and it started reconstructing entier data from other healthy replica
	CVRStatusReconstructingNewReplica CStorVolumeReplicaPhase = "ReconstructingNewReplica"

	// CVRStatusError describes either volume replica is not exist in cstor pool
	CVRStatusError CStorVolumeReplicaPhase = "Error"

	// CVRStatusDeletionFailed describes volume replica deletion is failed
	CVRStatusDeletionFailed CStorVolumeReplicaPhase = "DeletionFailed"

	// CVRStatusInvalid ensures invalid resource(currently not honoring)
	CVRStatusInvalid CStorVolumeReplicaPhase = "Invalid"

	// CVRStatusInit describes CVR resource is newly created but it is not yet
	// created zfs dataset
	CVRStatusInit CStorVolumeReplicaPhase = "Init"

	// CVRStatusRecreate describes the volume replica is recreated due to pool
	// recreation/scaleup
	CVRStatusRecreate CStorVolumeReplicaPhase = "Recreate"
)

// CStorVolumeReplicaStatus is for handling status of cvr.
type CStorVolumeReplicaStatus struct {
	Phase    CStorVolumeReplicaPhase `json:"phase"`
	Capacity CStorVolumeCapacityAttr `json:"capacity"`
	// LastTransitionTime refers to the time when the phase changes
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	LastUpdateTime     metav1.Time `json:"lastUpdateTime,omitempty"`
	Message            string      `json:"message,omitempty"`

	// Snapshots contains list of snapshots, and their properties,
	// created on CVR
	Snapshots map[string]CStorSnapshotInfo `json:"snapshots,omitempty"`

	// PendingSnapshots contains list of pending snapshots that are not yet
	// available on this replica
	PendingSnapshots map[string]CStorSnapshotInfo `json:"pendingSnapshots,omitempty"`
}

type CStorSnapshotInfo struct {
	// LogicalReferenced describes the amount of space that is "logically"
	// accessable by this snapshot. This logical space ignores the
	// effect of the compression and copies properties, giving a quantity
	// closer to the amount of data that application see. It also includes
	// space consumed by metadata.
	LogicalReferenced uint64 `json:"logicalReferenced"`

	// Used is the used bytes for given snapshot
	// Used uint64 `json:"used"`
}

// CStorVolumeCapacityAttr is for storing the volume capacity.
type CStorVolumeCapacityAttr struct {
	TotalAllocated string `json:"totalAllocated"`
	Used           string `json:"used"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolumereplicas

// CStorVolumeReplicaList is a list of CStorVolumeReplica resources
type CStorVolumeReplicaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorVolumeReplica `json:"items"`
}
