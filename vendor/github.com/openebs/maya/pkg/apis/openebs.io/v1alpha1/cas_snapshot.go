/*
Copyright 2017 The OpenEBS Authors.

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

// CASSnapshotKey is a typed string to represent CAS Snapshot related annotations'
// or labels' keys
//
// Example 1 - Below is a sample StorageClass that makes use of a CASSnapshotKey
// constant i.e. the cas template used to create a cas snapshot
//
// ```yaml
// apiVersion: storage.k8s.io/v1
// kind: StorageClass
// metadata:
//  name: openebs-standard
//  annotations:
//    cas.openebs.io/create-snapshot-template: cast-standard-0.8.0
// provisioner: openebs.io/provisioner-iscsi
// ```
type CASSnapshotKey string

const (
	// CASTemplateKeyForSnapshotCreate is the key to fetch name of CASTemplate
	// to create a CAS Snapshot
	CASTemplateKeyForSnapshotCreate CASSnapshotKey = "cas.openebs.io/create-snapshot-template"

	// CASTemplateKeyForSnapshotRead is the key to fetch name of CASTemplate
	// to read a CAS Snapshot
	CASTemplateKeyForSnapshotRead CASSnapshotKey = "cas.openebs.io/read-snapshot-template"

	// CASTemplateKeyForSnapshotDelete is the key to fetch name of CASTemplate
	// to delete a CAS Snapshot
	CASTemplateKeyForSnapshotDelete CASSnapshotKey = "cas.openebs.io/delete-snapshot-template"

	// CASTemplateKeyForSnapshotList is the key to fetch name of CASTemplate
	// to list CAS Snapshots
	CASTemplateKeyForSnapshotList CASSnapshotKey = "cas.openebs.io/list-snapshot-template"
)

// CASSnapshot represents a cas snapshot
type CASSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec i.e. specifications of this cas snapshot
	Spec SnapshotSpec `json:"spec"`
}

// SnapshotSpec has the properties of a cas snapshot
type SnapshotSpec struct {
	CasType    string `json:"casType"`
	VolumeName string `json:"volumeName"`
}

// SnapshotOptions has the properties of a cas snapshot list
type SnapshotOptions struct {
	CasType    string `json:"casType,omitempty"`
	VolumeName string `json:"volumeName,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name,omitempty"`
}

// CASSnapshotList is a list of CASSnapshot resources
type CASSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	// Items are the list of volumes
	Items []CASSnapshot `json:"items"`
}
