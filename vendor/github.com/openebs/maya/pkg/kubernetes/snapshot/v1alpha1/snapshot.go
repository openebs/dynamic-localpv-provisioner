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

package v1alpha1

import (
	snapshot "github.com/openebs/maya/pkg/apis/openebs.io/snapshot/v1"
)

// Snapshot is a wrapper over API based
// volume snapshot instance
type Snapshot struct {
	object *snapshot.VolumeSnapshot
}

// SnapshotList holds the list of Snapshot instances
type SnapshotList struct {
	items *snapshot.VolumeSnapshotList
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided Snapshot instance
type Predicate func(*Snapshot) bool

// predicateList holds the list of predicates
type predicateList []Predicate

// all returns true if all the predicateList
// succeed against the provided Snapshot
// instance
func (l predicateList) all(s *Snapshot) bool {
	for _, pred := range l {
		if !pred(s) {
			return false
		}
	}
	return true
}

// NewForAPIObject returns a new instance of Snapshot
func NewForAPIObject(obj *snapshot.VolumeSnapshot) *Snapshot {
	s := &Snapshot{object: obj}
	return s
}

// Len returns the number of items present
// in the SnapshotList
func (p *SnapshotList) Len() int {
	return len(p.items.Items)
}
