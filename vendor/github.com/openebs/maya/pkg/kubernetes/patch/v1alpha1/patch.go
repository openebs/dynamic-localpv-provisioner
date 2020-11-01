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
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/template"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

// PatchType is a custom type that holds the patch type
type PatchType string

const (
	// PatchTypeJSON refers to a generic json patch type that is understood
	// by Kubernetes API as well
	PatchTypeJSON PatchType = "json"
	// PatchTypeMerge refers to a generic json merge patch type that is
	// understood by Kubernetes API as well
	PatchTypeMerge PatchType = "merge"
	// PatchTypeStrategic refers to a patch type that is understood
	// by Kubernetes API only
	PatchTypeStrategic PatchType = "strategic"
)

// kubePatchTypes map task patch types
// to k8s patch types
var kubePatchTypes = map[PatchType]types.PatchType{
	PatchTypeJSON:      types.JSONPatchType,
	PatchTypeMerge:     types.MergePatchType,
	PatchTypeStrategic: types.StrategicMergePatchType,
}

// Patch will consist of patch that gets applied
// against a particular resource
type Patch struct {
	// Type determines the type of patch to be applied
	Type types.PatchType `json:"type"`
	// object determines the actual patch object
	// in json format
	Object []byte `json:"object"`
}

// GoString provides the essential Patch struct details
func (p *Patch) GoString() string {
	return fmt.Sprintf("Patch{Type: %s}", p.Type)
}

// String provides the essential Patch details
func (p *Patch) String() string {
	return p.GoString()
}

// Builder returns a new instance of builder
type Builder struct {
	patch  *Patch
	checks map[*Predicate]string
	errors []error
}

// Predicate abstracts conditional logic w.r.t the patch instance
//
// NOTE:
// Predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// Predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*Patch) bool

// NewBuilder returns a new instance of builder
func NewBuilder() *Builder {
	return &Builder{
		patch:  &Patch{},
		checks: make(map[*Predicate]string),
	}
}

// BuilderForObject returns a new instance of builder
// when a patch obj and patch type is given
func BuilderForObject(t types.PatchType, obj []byte) *Builder {
	return &Builder{
		patch: &Patch{
			Type:   t,
			Object: obj,
		},
		checks: make(map[*Predicate]string),
	}
}

// BuilderForRuntask returns a new instance of builder
// for runtask when a runtask patch yaml is given
func BuilderForRuntask(context, templateYaml string,
	templateValues map[string]interface{}) *Builder {
	// runTaskPatch reflects a runtask having
	// patch action
	type runTaskPatch struct {
		// Type determines the type of patch to be applied
		Type PatchType `json:"type"`
		// Spec determines the actual patch object
		Spec string `json:"pspec"`
	}
	t := &runTaskPatch{}
	// This will be used to unmarshal spec
	// field of patch runtask
	m := map[string]interface{}{}
	b := &Builder{
		patch:  &Patch{},
		checks: make(map[*Predicate]string),
	}
	// Here, we are running go templating on the given runtask yaml
	p, err := template.AsTemplatedBytes(context, templateYaml, templateValues)
	if err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	// unmarshal task yaml into taskPatch struct
	// to get the type and the actual patch object
	err = yaml.Unmarshal(p, t)
	if err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	// unmarshal rawSpec into map[string]interface{}{}
	// since the patch operation supports this format
	// for patching a k8s resource
	err = yaml.Unmarshal([]byte(t.Spec), &m)
	if err != nil {
		return b
	}
	raw, err := json.Marshal(m)
	if err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	b.patch = &Patch{
		Type:   kubePatchTypes[t.Type],
		Object: raw,
	}
	return b
}

// validate will run checks against patch instance
func (b *Builder) validate() error {
	for cond := range b.checks {
		pass := (*cond)(b.patch)
		if !pass {
			b.errors = append(b.errors,
				errors.Errorf("validation failed: %s", b.checks[cond]))
		}
	}
	if len(b.errors) == 0 {
		return nil
	}
	return errors.Errorf("%v", b.errors)
}

// Build returns the final instance of patch
func (b *Builder) Build() (*Patch, error) {
	if len(b.errors) != 0 {
		return nil, errors.Errorf("%v", b.errors)
	}
	err := b.validate()
	if err != nil {
		return nil, err
	}
	return b.patch, nil
}

// AddCheckf adds the predicate as a condition to be validated against the
// patch instance and format the message string according to format specifier.
// If only predicate and message string is provided, it will treat it as the
// value for the corresponding predicate.
func (b *Builder) AddCheckf(p Predicate, predicateMsg string, args ...interface{}) *Builder {
	if p == nil {
		b.errors = append(b.errors,
			errors.New("nil predicate given"))
		return b
	}
	b.checks[&p] = fmt.Sprintf(predicateMsg, args...)
	return b
}

// AddCheck adds the predicate as a condition to be validated against the
// patch instance
func (b *Builder) AddCheck(p Predicate) *Builder {
	return b.AddCheckf(p, "")
}

// AddChecks adds the provided predicates as conditions to be validated against
// the patch instance
func (b *Builder) AddChecks(predicates ...Predicate) *Builder {
	for _, check := range predicates {
		b.AddCheck(check)
	}
	return b
}

// IsValidType returns a predicate for
// checking valid patch type
func IsValidType() Predicate {
	return func(p *Patch) bool {
		return p.IsValidType()
	}
}

// IsValidType returns true if provided patch
// type is one of the valid patch types
func (p *Patch) IsValidType() bool {
	return p.Type == types.JSONPatchType || p.Type == types.MergePatchType ||
		p.Type == types.StrategicMergePatchType
}
