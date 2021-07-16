/*
Copyright © 2019 The OpenEBS Authors

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
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

// Namespace is a wrapper over Namespace api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type Namespace struct {
	object *corev1.Namespace
}

// Builder enables building an instance of StorageClass
type Builder struct {
	ns   *Namespace
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{ns: &Namespace{object: &corev1.Namespace{}}}
}

// WithName sets the Name field of namespace with provided argument.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build namespace: missing namespace name"))
		return b
	}
	b.ns.object.Name = name
	return b
}

// WithGenerateName appends a random string after the name
func (b *Builder) WithGenerateName(name string) *Builder {
	b.ns.object.GenerateName = name + "-"
	return b
}

// Build returns the Namespace instance
func (b *Builder) Build() (*Namespace, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.ns, nil
}

// APIObject returns the API Namespace instance
func (b *Builder) APIObject() (*corev1.Namespace, error) {
	ns, err := b.Build()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build APIObject")
	}
	return ns.object, nil
}
