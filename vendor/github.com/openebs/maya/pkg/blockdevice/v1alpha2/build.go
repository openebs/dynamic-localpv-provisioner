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

package v1alpha2

import (
	ndm "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
)

//TODO: While using these packages
// UnitTest must be written to corresponding function

// Builder is the builder object for BlockDevice
type Builder struct {
	BlockDevice *BlockDevice
}

// NewBuilder returns an empty instance of the Builder object.
func NewBuilder() *Builder {
	return &Builder{
		BlockDevice: &BlockDevice{&ndm.BlockDevice{}},
	}
}

// BuilderForObject returns an instance of the Builder object based on block
// device object
func BuilderForObject(BlockDevice *BlockDevice) *Builder {
	return &Builder{
		BlockDevice: BlockDevice,
	}
}

// BuilderForAPIObject returns an instance of the Builder object based on block
// device api object.
func BuilderForAPIObject(bd *ndm.BlockDevice) *Builder {
	return &Builder{
		BlockDevice: &BlockDevice{bd},
	}
}

// Build returns the block device object built by this builder.
func (b *Builder) Build() *BlockDevice {
	return b.BlockDevice
}
