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
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//UpgradeResult holds the apis upgraderesult object
type UpgradeResult struct {
	// upgrade result object
	object *apis.UpgradeResult
}

// UpgradeResultList is the list of
// upgradeResults
type UpgradeResultList struct {
	// list of upgrade results
	items []*UpgradeResult
}

// Builder enables building an instance of
// upgradeResult
type Builder struct {
	*UpgradeResult
	checks map[*Predicate]string
	errors []error
}

// Predicate abstracts conditional logic w.r.t the upgradeResult instance
//
// NOTE:
// Predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// Predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*UpgradeResult) bool

// NewBuilder returns a new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		UpgradeResult: &UpgradeResult{
			object: &apis.UpgradeResult{},
		},
		checks: make(map[*Predicate]string),
	}
}

// WithTypeMeta adds typemeta in upgrade result instance.
func (b *Builder) WithTypeMeta(typeMeta metav1.TypeMeta) *Builder {
	b.object.TypeMeta = typeMeta
	return b
}

// WithObjectMeta adds objectMeta in upgrade result instance.
func (b *Builder) WithObjectMeta(objectMeta metav1.ObjectMeta) *Builder {
	b.object.ObjectMeta = objectMeta
	return b
}

// WithTasks adds tasks details in upgrade result instance.
func (b *Builder) WithTasks(tasks ...apis.UpgradeResultTask) *Builder {
	b.object.Tasks = append(b.object.Tasks, tasks...)
	return b
}

// WithResultConfig adds resource details and runtime
// config in upgrade result instance.
func (b *Builder) WithResultConfig(resource apis.ResourceDetails,
	data ...apis.DataItem) *Builder {
	b.object.Config.ResourceDetails = resource
	b.object.Config.Data = append(b.object.Config.Data, data...)
	return b
}

// BuilderForYAMLObject returns a new instance
// of Builder for a given template object
func BuilderForYAMLObject(object []byte) *Builder {
	b := &Builder{
		UpgradeResult: &UpgradeResult{
			object: &apis.UpgradeResult{},
		},
		checks: make(map[*Predicate]string),
	}
	err := yaml.Unmarshal(object, b.object)
	if err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	return b
}

// Build returns the final instance of upgradeResult
func (b *Builder) Build() (*apis.UpgradeResult, error) {
	if len(b.errors) != 0 {
		return nil, errors.Errorf("%v", b.errors)
	}
	err := b.validate()
	if err != nil {
		return nil, err
	}
	return b.UpgradeResult.object, nil
}

// validate will run checks against upgrade
// result instance
func (b *Builder) validate() error {
	for cond := range b.checks {
		pass := (*cond)(b.UpgradeResult)
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

// AddCheckf adds the predicate as a condition to be validated against the
// upgrade result instance and format the message string according to format specifier.
// If only predicate and message string is provided, it will treat it as the
// value for the corresponding predicate.
func (b *Builder) AddCheckf(p Predicate, predicateMsg string, args ...interface{}) *Builder {
	b.checks[&p] = fmt.Sprintf(predicateMsg, args...)
	return b
}

// AddCheck adds the predicate as a condition to be validated against the
// upgrade result instance
func (b *Builder) AddCheck(p Predicate) *Builder {
	return b.AddCheckf(p, "")
}

// AddChecks adds the provided predicates as conditions to be validated against
// the upgrade result instance
func (b *Builder) AddChecks(predicates ...Predicate) *Builder {
	for _, check := range predicates {
		b.AddCheck(check)
	}
	return b
}

// ListBuilder enables building
// an instance of upgradeResultList
type ListBuilder struct {
	list *UpgradeResultList
}

// NewListBuilder returns a new instance
// of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &UpgradeResultList{}}
}

// WithAPIList builds the list of ur
// instances based on the provided
// ur api instances
func (b *ListBuilder) WithAPIList(list *apis.UpgradeResultList) *ListBuilder {
	if list == nil {
		return b
	}
	for i := range list.Items {
		b.list.items = append(b.list.items, &UpgradeResult{object: &list.Items[i]})
	}
	return b
}

// List returns the list of ur
// instances that was built by this
// Builder
func (b *ListBuilder) List() *UpgradeResultList {
	return b.list
}
