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
	errors "github.com/pkg/errors"
)

// Builder enables building an instance of snapshot
type Builder struct {
	s    *Snapshot
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{s: &Snapshot{object: &snapshot.VolumeSnapshot{}}}
}

// WithName sets the Name field of snapshot with provided argument.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build snapshot: missing snapshot name"))
		return b
	}
	b.s.object.Name = name
	return b
}

// WithNamespace sets the Namespace field of snapshot with provided value.
func (b *Builder) WithNamespace(namespace string) *Builder {
	if len(namespace) == 0 {
		b.errs = append(b.errs, errors.New("failed to build snapshot: missing namespace"))
	}
	b.s.object.Namespace = namespace
	return b
}

// WithPVC sets the PVC field of snapshot with provided value.
func (b *Builder) WithPVC(pvc string) *Builder {
	if len(pvc) == 0 {
		b.errs = append(b.errs, errors.New("failed to build snapshot: missing pvc name"))
	}
	b.s.object.Spec.PersistentVolumeClaimName = pvc
	return b
}

// Build returns the snapshot API instance
func (b *Builder) Build() (*snapshot.VolumeSnapshot, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.s.object, nil
}
