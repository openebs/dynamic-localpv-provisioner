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

package v1alpha2

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"
	"text/template"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

type annotationKey string

const scheduleOnHostAnnotation annotationKey = "volume.kubernetes.io/selected-node"

const (
	// PoolCapacityThresholdPercentage is the threshold percentage for usable pool size.
	// Example:
	// If pool size is 100 Gi and threshold percentae is 70 then
	// 70 Gi is the usable pool capacity.
	PoolCapacityThresholdPercentage = 70
)

// CSP encapsulates the CStorPool API object.
type CSP struct {
	// actual cstor pool Object
	Object *apis.CStorPool
}

// predicate defines an abstraction
// to determine conditional checks
// against the provided CSP instance
type predicate func(*CSP) bool

type predicateList []predicate

// all returns true if all the predicates
// succeed against the provided CSP
// instance
func (l predicateList) all(c *CSP) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// IsNotUID returns true if provided CSP
// instance's UID does not match with any
// of the provided UIDs
func IsNotUID(uids ...string) predicate {
	return func(c *CSP) bool {
		for _, uid := range uids {
			if uid == string(c.Object.GetUID()) {
				return false
			}
		}
		return true
	}
}

// HasSpace returns true if the CSP has free space to accomodate the incoming volume.
func (c *CSP) HasSpace(incomingVolCap, capacityUsedByExistingVols resource.Quantity) bool {
	totalUsableCapacityOnPool, err := c.UsableFreeCapacityOnCSP(capacityUsedByExistingVols)
	if err != nil {
		klog.Error(err)
		return false
	}
	if totalUsableCapacityOnPool.Cmp(incomingVolCap) == 1 {
		return true
	}
	return false
}

// UsableFreeCapacityOnCSP returns the total usable free capacity present on the CSP.
// Example:
// If UsableFreeCapacityOnCSP is 30 Gi than the pool (CSP) can accomodate any volume
// of size less than(or equal to) 30 Gi
func (c *CSP) UsableFreeCapacityOnCSP(capacityUsedByExistingVols resource.Quantity) (resource.Quantity, error) {
	usableCapcityOnCSP, err := c.GetTotalUsableCapacityWithThreshold()
	if err != nil {
		return resource.Quantity{}, errors.Wrapf(err, "failed to get total usable capacity on CSP %s", c.Object.Name)
	}
	usableCapcityOnCSP.Sub(capacityUsedByExistingVols)
	return usableCapcityOnCSP, nil
}

// GetTotalUsableCapacityWithThreshold returns the usable capacity on pool.
// Example:
// If TotalUsableCapacityWithThreshold is 70 Gi than it is possible that some amount
// of space is already being used by some other existing volumes.
func (c *CSP) GetTotalUsableCapacityWithThreshold() (resource.Quantity, error) {
	totalCapacity, err := resource.ParseQuantity(c.Object.Status.Capacity.Total)
	if err != nil {
		return resource.Quantity{}, errors.Wrapf(err, "failed to parse capcity {%s} from CSP %s", c.Object.Status.Capacity.Total, c.Object.Name)
	}
	capacityThreshold := c.GetCapacityThreshold()

	totalCapacityValue := totalCapacity.Value()
	totalUsableCapacityValue := totalCapacityValue * (int64(capacityThreshold)) / 100
	totalCapacity.Set(totalUsableCapacityValue)
	return totalCapacity, nil
}

// GetCapacityThreshold returns the capacity threshold.
// Example :
// If pool size is 100 Gi and capacity threshold is 70 %
// then the effective size of pool is 70 Gi for accommodating volumes.
func (c *CSP) GetCapacityThreshold() int {
	// ToDo: Add capability to override via annotaions
	return PoolCapacityThresholdPercentage
}

// HasAnnotation returns true if provided annotation
// key and value are present in the provided CSP
// instance
func HasAnnotation(key, value string) predicate {
	return func(c *CSP) bool {
		val, ok := c.Object.GetAnnotations()[key]
		if ok {
			return val == value
		}
		return false
	}
}

// CSPList holds the list of cstorpools
type CSPList struct {
	// list of cstor pools
	Items []*CSP
}

// Filter will filter the CSP instances
// if all the predicates succeed against that
// CSP.
func (l *CSPList) Filter(p ...predicate) *CSPList {
	var plist predicateList
	plist = append(plist, p...)
	if len(plist) == 0 {
		return l
	}

	filtered := ListBuilder().List()
	for _, CSP := range l.Items {
		if plist.all(CSP) {
			filtered.Items = append(filtered.Items, CSP)
		}
	}
	return filtered
}

// listBuilder enables building a
// list of CSP instances
type listBuilder struct {
	list *CSPList
}

// ListBuilder returns a new instance of
// listBuilder Object
func ListBuilder() *listBuilder {
	return &listBuilder{list: &CSPList{}}
}

// WithUIDs builds a list of cstor pools
// based on the provided pool UIDs
func (b *listBuilder) WithUIDs(poolUIDs ...string) *listBuilder {
	for _, uid := range poolUIDs {
		obj := &CSP{&apis.CStorPool{}}
		obj.Object.SetUID(types.UID(uid))
		b.list.Items = append(b.list.Items, obj)
	}
	return b
}

// WithUIDNodeMap builds a CSPList based on the provided
// map of uid and nodename
func (b *listBuilder) WithUIDNode(UIDNode, UIDCapacity map[string]string) *listBuilder {
	for k, v := range UIDNode {
		obj := &CSP{&apis.CStorPool{}}
		obj.Object.SetUID(types.UID(k))
		obj.Object.SetAnnotations(map[string]string{string(scheduleOnHostAnnotation): v})
		obj.Object.Status.Capacity.Total = UIDCapacity[k]
		b.list.Items = append(b.list.Items, obj)
	}
	return b
}

// WithList builds the list based on the provided
// *apis.CStorPool instances
func (b *listBuilder) WithList(pools *CSPList) *listBuilder {
	if pools == nil {
		return b
	}
	b.list.Items = append(b.list.Items, pools.Items...)
	return b
}

// WithAPIList builds the list based on the provided
// *apis.CStorPoolList
func (b *listBuilder) WithAPIList(pools *apis.CStorPoolList) *listBuilder {
	if pools == nil {
		return b
	}
	for _, pool := range pools.Items {
		pool := pool //pin it
		b.list.Items = append(b.list.Items, &CSP{&pool})
	}

	return b
}

// List returns the list of CSP
// instances that were built by
// this builder
func (b *listBuilder) List() *CSPList {
	return b.list
}

// GetPoolUIDs retuns the UIDs of the pools
// available in the list
func (l *CSPList) GetPoolUIDs() []string {
	uids := []string{}
	for _, pool := range l.Items {
		uids = append(uids, string(pool.Object.GetUID()))
	}
	return uids
}

// newListFromUIDNode exposes WithUIDNodeMap
// to CAS Templates
func newListFromUIDNode(UIDNodeMap, UIDCapacityMap map[string]string) *CSPList {
	return ListBuilder().WithUIDNode(UIDNodeMap, UIDCapacityMap).List()
}

// newListFromUIDs exposes WithUIDs to CASTemplates
func newListFromUIDs(uids []string) *CSPList {
	return ListBuilder().WithUIDs(uids...).List()
}

// TemplateFunctions exposes a few functions as go
// template functions
func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"createCSPListFromUIDs":       newListFromUIDs,
		"createCSPListFromUIDNodeMap": newListFromUIDNode,
	}
}
