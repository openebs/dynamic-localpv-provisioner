// Copyright © 2019 The OpenEBS Authors
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
	"fmt"

	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	unstruct "github.com/openebs/maya/pkg/unstruct/v1alpha2"
	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// DefaultInstaller enables installing/uninstalling
// of resources in container orchestrator
type DefaultInstaller struct {
	object     *unstructured.Unstructured
	predicates []Predicate
	*unstruct.Kubeclient
}

// Predicate abstracts the verification of
// a resource
type Predicate func(*DefaultInstaller) bool

// String implements Stringer interface
func (i *DefaultInstaller) String() string {
	return StringYaml("default installer", i.object)
}

// StringYaml returns the provided object
// as a yaml formatted string
func StringYaml(ctx string, obj interface{}) string {
	if obj == nil {
		return ""
	}
	b, _ := yaml.Marshal(obj)
	return fmt.Sprintf("\n%s {%s}", ctx, string(b))
}

// GoString implements GoStringer interface
func (i *DefaultInstaller) GoString() string {
	return i.String()
}

func (i *DefaultInstaller) getKubeClientOrCached() *unstruct.Kubeclient {
	if i.Kubeclient != nil {
		return i.Kubeclient
	}
	i.Kubeclient = unstruct.NewKubeClient()
	return i.Kubeclient
}

// GetUnstructuredObject returns Unstructured objecct
func (i *DefaultInstaller) GetUnstructuredObject() *unstructured.Unstructured {
	return i.object
}

// Install triggers the installation of resource
// in the kubernetes cluster
func (i *DefaultInstaller) Install() error {
	return i.getKubeClientOrCached().Create(i.object)
}

// UnInstall triggers deletion of resource from
// kubernetes cluster
func (i *DefaultInstaller) UnInstall() error {
	return i.getKubeClientOrCached().Delete(i.object)
}

// Verify returns the installation of resource.
// It returns true if all the predicates passes
func (i *DefaultInstaller) Verify() (bool, error) {
	k := i.getKubeClientOrCached()
	obj, err := k.Get(
		i.object.GetName(),
		unstruct.WithGetNamespace(i.object.GetNamespace()),
		unstruct.WithGroupVersionResource(k8s.GroupVersionResourceFromGVK(i.object)),
	)
	if err != nil {
		return false, err
	}

	for _, p := range i.predicates {
		if !p(&DefaultInstaller{object: obj}) {
			return false, nil
		}
	}
	return true, nil
}

// Builder abstracts the construction of the builder
type Builder struct {
	installer *DefaultInstaller
	errs      []error
}

// BuilderForObject creates the installer builder from
// the object provided
func BuilderForObject(obj *unstructured.Unstructured) *Builder {
	b := &Builder{}
	if obj == nil {
		b.errs = append(b.errs, errors.Errorf("failed to build for object: nil unstruct instance provided"))
		return b
	}
	b.installer = &DefaultInstaller{object: obj, Kubeclient: nil}
	return b
}

// BuilderForYaml creates the installer builder from the
// YAML provided
func BuilderForYaml(yaml string) *Builder {
	b := &Builder{}
	obj, err := unstruct.BuilderForYaml(yaml).Build()
	if err != nil {
		b.errs = append(b.errs, err)
		return b
	}
	b.installer = &DefaultInstaller{object: obj.GetUnstructured(), Kubeclient: nil}
	return b
}

// WithKubeClient returns a new instance of unstructured kubeclient
func (b *Builder) WithKubeClient(opts ...unstruct.KubeclientBuildOption) *Builder {
	b.installer.Kubeclient = unstruct.NewKubeClient(opts...)
	return b
}

// AddCheck adds verify predicated for the installer
func (b *Builder) AddCheck(p ...Predicate) *Builder {
	for _, o := range p {
		if o != nil {
			b.installer.predicates = append(b.installer.predicates, o)
		}
	}
	return b
}

// Build triggers the building of the installer
func (b *Builder) Build() (*DefaultInstaller, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("errors {%+v}", b.errs)
	}
	return b.installer, nil
}

// IsPodRunning returns true if the pod is in running
// state
func IsPodRunning() Predicate {
	return func(d *DefaultInstaller) bool {
		v := unstruct.GetNestedString(d.object.Object, "status", "phase")
		return v == "Running"
	}
}
