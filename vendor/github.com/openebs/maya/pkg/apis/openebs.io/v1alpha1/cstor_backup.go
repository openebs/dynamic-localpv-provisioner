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

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorbackup

// CStorBackup describes a cstor backup resource created as a custom resource
type CStorBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorBackupSpec   `json:"spec"`
	Status            CStorBackupStatus `json:"status"`
}

// CStorBackupSpec is the spec for a CStorBackup resource
type CStorBackupSpec struct {
	// BackupName is a name of the backup or scheduled backup
	BackupName string `json:"backupName"`

	// VolumeName is a name of the volume for which this backup is destined
	VolumeName string `json:"volumeName"`

	// SnapName is a name of the current backup snapshot
	SnapName string `json:"snapName"`

	// PrevSnapName is the last completed-backup's snapshot name
	PrevSnapName string `json:"prevSnapName"`

	// BackupDest is the remote address for backup transfer
	BackupDest string `json:"backupDest"`

	// LocalSnap is flag to enable local snapshot only
	LocalSnap bool `json:"localSnap"`
}

// CStorBackupStatus is to hold status of backup
type CStorBackupStatus string

// Status written onto CStorBackup objects.
const (
	// BKPCStorStatusEmpty ensures the create operation is to be done, if import fails.
	BKPCStorStatusEmpty CStorBackupStatus = ""

	// BKPCStorStatusDone , backup is completed.
	BKPCStorStatusDone CStorBackupStatus = "Done"

	// BKPCStorStatusFailed , backup is failed.
	BKPCStorStatusFailed CStorBackupStatus = "Failed"

	// BKPCStorStatusInit , backup is initialized.
	BKPCStorStatusInit CStorBackupStatus = "Init"

	// BKPCStorStatusPending , backup is pending.
	BKPCStorStatusPending CStorBackupStatus = "Pending"

	// BKPCStorStatusInProgress , backup is in progress.
	BKPCStorStatusInProgress CStorBackupStatus = "InProgress"

	// BKPCStorStatusInvalid , backup operation is invalid.
	BKPCStorStatusInvalid CStorBackupStatus = "Invalid"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorbackup

// CStorBackupList is a list of CStorBackup resources
type CStorBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorBackup `json:"items"`
}
