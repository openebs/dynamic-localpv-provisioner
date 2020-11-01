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
	ndm "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
)

// ListBuilder is the builder object for BlockDeviceClaimList
type ListBuilder struct {
	BlockDeviceClaimList *BlockDeviceClaimList
	filters              PredicateList
}

// NewListBuilder returns a new instance of ListBuilder object.
func NewListBuilder() *ListBuilder {
	return &ListBuilder{
		BlockDeviceClaimList: &BlockDeviceClaimList{
			ObjectList: &ndm.BlockDeviceClaimList{},
		},
		filters: PredicateList{},
	}
}

// ListBuilderFromList builds the list based on the
// provided *BlockDeviceClaimList instances.
func ListBuilderFromList(bdcl *BlockDeviceClaimList) *ListBuilder {
	lb := NewListBuilder()
	if bdcl == nil {
		return lb
	}
	lb.BlockDeviceClaimList.ObjectList.Items =
		append(lb.BlockDeviceClaimList.ObjectList.Items,
			bdcl.ObjectList.Items...)
	return lb
}

// ListBuilderFromAPIList builds the list based on the provided APIBDC List
func ListBuilderFromAPIList(bdcl *ndm.BlockDeviceClaimList) *ListBuilder {
	lb := NewListBuilder()
	if bdcl == nil {
		return lb
	}
	lb.BlockDeviceClaimList.ObjectList.Items = append(
		lb.BlockDeviceClaimList.ObjectList.Items,
		bdcl.Items...)
	return lb
}

// List returns the list of bdc
// instances that was built by this
// builder
func (lb *ListBuilder) List() *BlockDeviceClaimList {
	if lb.filters == nil || len(lb.filters) == 0 {
		return lb.BlockDeviceClaimList
	}
	filtered := NewListBuilder().List()
	for _, bdcAPI := range lb.BlockDeviceClaimList.ObjectList.Items {
		bdcAPI := bdcAPI // pin it
		bdc := BuilderForAPIObject(&bdcAPI).BDC
		if lb.filters.all(bdc) {
			filtered.ObjectList.Items = append(filtered.ObjectList.Items, *bdc.Object)
		}
	}
	return filtered
}

// WithFilter adds filters on which the bdc's has to be filtered
func (lb *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	lb.filters = append(lb.filters, pred...)
	return lb
}

// GetBlockDeviceClaim returns block device claim object from existing
// ListBuilder
func (lb *ListBuilder) GetBlockDeviceClaim(bdcName string) *ndm.BlockDeviceClaim {
	for _, bdcObj := range lb.BlockDeviceClaimList.ObjectList.Items {
		bdcObj := bdcObj
		if bdcObj.Name == bdcName {
			return &bdcObj
		}
	}
	return nil
}
