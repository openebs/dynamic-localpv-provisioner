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
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorrestore

// CStorRestore describes a cstor restore resource created as a custom resource
type CStorRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"` // set name to restore name + volume name + something like csp tag
	Spec              CStorRestoreSpec            `json:"spec"`
	Status            CStorRestoreStatus          `json:"status"`
}

// CStorRestoreSpec is the spec for a CStorRestore resource
type CStorRestoreSpec struct {
	RestoreName   string            `json:"restoreName"` // set restore name
	VolumeName    string            `json:"volumeName"`
	RestoreSrc    string            `json:"restoreSrc"` // it can be ip:port in case of restore from remote or volumeName in case of local restore
	MaxRetryCount int               `json:"maxretrycount"`
	RetryCount    int               `json:"retrycount"`
	StorageClass  string            `json:"storageClass,omitempty"`
	Size          resource.Quantity `json:"size,omitempty"`
	Local         bool              `json:"localRestore,omitempty"` // if restore is from local backup/snapshot
}

// CStorRestoreStatus is to hold result of action.
type CStorRestoreStatus string

// Status written onto CStrorRestore object.
const (
	// RSTCStorStatusEmpty ensures the create operation is to be done, if import fails.
	RSTCStorStatusEmpty CStorRestoreStatus = ""

	// RSTCStorStatusDone , restore operation is completed.
	RSTCStorStatusDone CStorRestoreStatus = "Done"

	// RSTCStorStatusFailed , restore operation is failed.
	RSTCStorStatusFailed CStorRestoreStatus = "Failed"

	// RSTCStorStatusInit , restore operation is initialized.
	RSTCStorStatusInit CStorRestoreStatus = "Init"

	// RSTCStorStatusPending , restore operation is pending.
	RSTCStorStatusPending CStorRestoreStatus = "Pending"

	// RSTCStorStatusInProgress , restore operation is in progress.
	RSTCStorStatusInProgress CStorRestoreStatus = "InProgress"

	// RSTCStorStatusInvalid , restore operation is invalid.
	RSTCStorStatusInvalid CStorRestoreStatus = "Invalid"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorrestore

// CStorRestoreList is a list of CStorRestore resources
type CStorRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorRestore `json:"items"`
}
