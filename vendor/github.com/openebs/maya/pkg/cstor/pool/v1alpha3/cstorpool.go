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

package v1alpha3

import (
	"text/template"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

type annotationKey string

const scheduleOnHostAnnotation annotationKey = "volume.kubernetes.io/selected-node"

// CSP encapsulates CStorPool api object.
type CSP struct {
	// actual cstor pool object
	Object *apis.CStorPool
}

// CSPList holds the list of StoragePoolClaim api.
type CSPList struct {
	// list of cstor pools
	ObjectList *apis.CStorPoolList
}

// Builder is the builder object for CSP.
type Builder struct {
	Csp *CSP
}

// ListBuilder is the builder object for CSPList.
type ListBuilder struct {
	CspList *CSPList
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided csp instance
type Predicate func(*CSP) bool

type predicateList []Predicate

// all returns true if all the predicates
// succeed against the provided csp
// instance
func (l predicateList) all(c *CSP) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// IsNotDeleted returns true if deletion
// timestamp was nil(which means
// object is not deleted) else return false
// i.e CSP is deleted
func IsNotDeleted() Predicate {
	return func(c *CSP) bool {
		return c.Object.DeletionTimestamp.IsZero()
	}
}

// IsNotUID returns true if provided csp
// instance's UID does not match with any
// of the provided UIDs
func IsNotUID(uids ...string) Predicate {
	return func(c *CSP) bool {
		for _, uid := range uids {
			if uid == string(c.Object.GetUID()) {
				return false
			}
		}
		return true
	}
}

// HasAnnotation returns true if provided annotation
// key and value are present in the provided CSP
// instance
func HasAnnotation(key, value string) Predicate {
	return func(c *CSP) bool {
		val, ok := c.Object.GetAnnotations()[key]
		if ok {
			return val == value
		}
		return false
	}
}

// HasLabel returns true if provided label
// key and value are present in the provided CSP
// instance
func HasLabel(key, value string) Predicate {
	return func(c *CSP) bool {
		val, ok := c.Object.GetLabels()[key]
		if ok {
			return val == value
		}
		return false
	}
}

// IsStatus returns true if the status on csp matches with provided status.
func IsStatus(status string) Predicate {
	return func(c *CSP) bool {
		val := c.Object.Status.Phase
		return string(val) == status
	}
}

// Filter will filter the csp instances
// if all the predicates succeed against that
// csp.
func (l *CSPList) Filter(p ...Predicate) *CSPList {
	var plist predicateList
	plist = append(plist, p...)
	if len(plist) == 0 {
		return l
	}

	filtered := NewListBuilder().List()
	for _, cspAPI := range l.ObjectList.Items {
		cspAPI := cspAPI // pin it
		CSP := BuilderForAPIObject(&cspAPI).Csp
		if plist.all(CSP) {
			filtered.ObjectList.Items = append(filtered.ObjectList.Items, *CSP.Object)
		}
	}
	return filtered
}

// NewListBuilder returns a new instance of ListBuilder Object.
func NewListBuilder() *ListBuilder {
	return &ListBuilder{CspList: &CSPList{ObjectList: &apis.CStorPoolList{}}}
}

// ListBuilderForObject returns a new instance of ListBuilderForApiList object based on csp list.
func ListBuilderForObject(cspList *CSPList) *ListBuilder {
	newLb := NewListBuilder()
	for _, obj := range cspList.ObjectList.Items {
		// pin it
		obj := obj
		newLb.CspList.ObjectList.Items = append(newLb.CspList.ObjectList.Items, obj)
	}
	return newLb
}

// ListBuilderForAPIObject returns a new instance of ListBuilderForApiList object based on csp api list.
func ListBuilderForAPIObject(cspAPIList *apis.CStorPoolList) *ListBuilder {
	newLb := NewListBuilder()
	for _, obj := range cspAPIList.Items {
		// pin it
		obj := obj
		newLb.CspList.ObjectList.Items = append(newLb.CspList.ObjectList.Items, obj)
	}
	return newLb
}

// WithUIDs builds a list of cstor pools
// based on the provided pool UIDs
func (b *ListBuilder) WithUIDs(poolUIDs ...string) *ListBuilder {
	for _, uid := range poolUIDs {
		obj := &CSP{&apis.CStorPool{}}
		obj.Object.SetUID(types.UID(uid))
		b.CspList.ObjectList.Items = append(b.CspList.ObjectList.Items, *obj.Object)
	}
	return b
}

// WithUIDNode builds a cspList based on the provided
// map of uid and nodename
func (b *ListBuilder) WithUIDNode(UIDNode map[string]string) *ListBuilder {
	for k, v := range UIDNode {
		obj := &CSP{&apis.CStorPool{}}
		obj.Object.SetUID(types.UID(k))
		obj.Object.SetAnnotations(map[string]string{string(scheduleOnHostAnnotation): v})
		b.CspList.ObjectList.Items = append(b.CspList.ObjectList.Items, *obj.Object)
	}
	return b
}

// List returns the list of csp instances that were built by this builder
func (b *ListBuilder) List() *CSPList {
	return b.CspList
}

// Len returns the length of the CSPList object
func (l *CSPList) Len() int {
	return len(l.ObjectList.Items)
}

// GetPoolUIDs retuns the UIDs of the pools available in the list.
func (l *CSPList) GetPoolUIDs() []string {
	uids := []string{}
	for _, pool := range l.ObjectList.Items {
		uids = append(uids, string(pool.GetUID()))
	}
	return uids
}

// NewBuilder returns an empty instance of the Builder object.
func NewBuilder() *Builder {
	return &Builder{
		Csp: &CSP{},
	}
}

// BuilderForObject returns an instance of the Builder object based on csp object.
func BuilderForObject(csp *CSP) *Builder {
	return &Builder{
		Csp: csp,
	}
}

// BuilderForAPIObject returns an instance of the Builder object based on csp api object
func BuilderForAPIObject(cspAPI *apis.CStorPool) *Builder {
	return &Builder{
		Csp: &CSP{cspAPI},
	}
}

// newListFromUIDNode exposes WithUIDNodeMap to CAS Templates.
func newListFromUIDNode(UIDNodeMap map[string]string) *CSPList {
	return NewListBuilder().WithUIDNode(UIDNodeMap).List()
}

// newListFromUIDs exposes WithUIDs to CASTemplates.
func newListFromUIDs(uids []string) *CSPList {
	return NewListBuilder().WithUIDs(uids...).List()
}

// TemplateFunctions exposes a few functions as go template functions.
func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"createCSPListFromUIDs":       newListFromUIDs,
		"createCSPListFromUIDNodeMap": newListFromUIDNode,
	}
}
