// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

type labelKey string

const (
	cstorPoolUIDLabel  labelKey = "cstorpool.openebs.io/uid"
	cstorpoolNameLabel labelKey = "cstorpool.openebs.io/name"
	// CStorVolumeReplicaFinalizer is the name of finalizer on CStorVolumeClaim
	CStorVolumeReplicaFinalizer = "cstorvolumereplica.openebs.io/finalizer"
)

// CVR is a wrapper for cstorvolume replica object
type CVR struct {
	// actual cstor volume replica
	// object
	object *apis.CStorVolumeReplica
}

// CVRList is a list of cstorvolume replica objects
type CVRList struct {
	// list of cstor volume replicas
	items []*CVR
}

// GetPoolUIDs returns a list of cstor pool
// UIDs corresponding to cstor volume replica
// instances
func (l *CVRList) GetPoolUIDs() []string {
	var uids []string
	for _, cvr := range l.items {
		uid := cvr.object.GetLabels()[string(cstorPoolUIDLabel)]
		uids = append(uids, uid)
	}
	return uids
}

// GetPoolNames returns a list of cstor pool
// name corresponding to cstor volume replica
// instances
func (l *CVRList) GetPoolNames() []string {
	var pools []string
	for _, cvr := range l.items {
		pool := cvr.object.GetLabels()[string(cstorpoolNameLabel)]
		if pool == "" {
			pools = append(pools, pool)
		}
	}
	return pools
}

// GetUniquePoolNames returns a list of cstor pool
// name corresponding to cstor volume replica
// instances
func (l *CVRList) GetUniquePoolNames() []string {
	registerd := map[string]bool{}
	var unique []string
	for _, cvr := range l.items {
		pool := cvr.object.GetLabels()[string(cstorpoolNameLabel)]
		if pool != "" && !registerd[pool] {
			registerd[pool] = true
			unique = append(unique, pool)
		}
	}
	return unique
}

// ListBuilder enables building
// an instance of CVRList
type ListBuilder struct {
	list    *CVRList
	filters PredicateList
}

// NewListBuilder returns a new instance
// of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &CVRList{}}
}

// WithAPIList builds the list of cvr
// instances based on the provided
// cvr api instances
func (b *ListBuilder) WithAPIList(
	list *apis.CStorVolumeReplicaList,
) *ListBuilder {
	if list == nil {
		return b
	}
	for _, c := range list.Items {
		c := c // pin it
		b.list.items = append(b.list.items, &CVR{object: &c})
	}
	return b
}

// AppendListBuilder append the provided CVR API object into existing ListBuilder
func (b *ListBuilder) AppendListBuilder(
	cvr *apis.CStorVolumeReplica) *ListBuilder {
	b.list.items = append(b.list.items, &CVR{object: cvr})
	return b
}

// List returns the list of cvr
// instances that was built by this
// builder
func (b *ListBuilder) List() *CVRList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := NewListBuilder().List()
	for _, cvr := range b.list.items {
		cvr := cvr // pin it
		if b.filters.all(cvr) {
			filtered.items = append(filtered.items, cvr)
		}
	}
	return filtered
}

// Len returns the number of items present
// in the CVRList
func (l *CVRList) Len() int {
	return len(l.items)
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided cvr instance
type Predicate func(*CVR) bool

// IsHealthy returns true if the CVR is in
// healthy state
func (p *CVR) IsHealthy() bool {
	return p.object.Status.Phase == "Healthy"
}

// IsErrored returns true if the CVR is in
// Error state
func (p *CVR) IsErrored() bool {
	return p.object.Status.Phase == "Error"
}

// HasLabel returns true only if label key value matched to provided
// value.
func (p *CVR) HasLabel(key, value string) bool {
	return p.object.Labels[key] == value
}

// HasLabel returns predicate to filter out CVRs based on the
// provided key and values
func HasLabel(key, value string) Predicate {
	return func(p *CVR) bool {
		return p.HasLabel(key, value)
	}
}

// IsHealthy is a Predicate to filter out cvrs
// which is healthy
func IsHealthy() Predicate {
	return func(p *CVR) bool {
		return p.IsHealthy()
	}
}

// IsErrored is a Predicate to filter out cvrs
// which is in errored state
func IsErrored() Predicate {
	return func(p *CVR) bool {
		return p.IsErrored()
	}
}

// PredicateList holds a list of Predicate
type PredicateList []Predicate

// all returns true if all the Predicates
// succeed against the provided cvr
// instance
func (l PredicateList) all(p *CVR) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// WithFilter adds filters on which the cvr's has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}
