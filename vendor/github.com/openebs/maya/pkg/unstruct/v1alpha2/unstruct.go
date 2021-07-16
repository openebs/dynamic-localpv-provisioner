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
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
	http "github.com/openebs/maya/pkg/util/http/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Unstruct holds an object of Unstructured
type Unstruct struct {
	Object *unstructured.Unstructured
}

// GetUnstructured converts Unstruct object
// to API's Unstructured
func (u *Unstruct) GetUnstructured() *unstructured.Unstructured {
	return u.Object
}

// Builder enables building of an
// Unstructured instance
type Builder struct {
	unstruct *Unstruct
	errs     []error
}

// NewBuilder returns a new instance of
// empty Builder
func NewBuilder() *Builder {
	return &Builder{
		unstruct: &Unstruct{
			&unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
		},
	}
}

// BuilderForYaml returns a new instance of
// Unstruct Builder by making use of the provided
// YAML
func BuilderForYaml(doc string) *Builder {
	b := NewBuilder()
	err := yaml.Unmarshal([]byte(doc), &b.unstruct.Object)
	if err != nil {
		b.errs = append(b.errs, err)
	}
	return b
}

// BuilderForObject returns a new instance of
// Unstruct Builder by making use of the provided object
func BuilderForObject(obj *unstructured.Unstructured) *Builder {
	b := NewBuilder()
	b.unstruct.Object = obj
	return b
}

// Build returns the Unstruct object created by
// the Builder
func (b *Builder) Build() (*Unstruct, error) {
	if len(b.errs) != 0 {
		return nil, errors.Errorf("errors {%+v}", b.errs)
	}
	return b.unstruct, nil
}

// BuildAPIUnstructured returns the Unstruct object created by
// the Builder
func (b *Builder) BuildAPIUnstructured() (*unstructured.Unstructured, error) {
	if len(b.errs) != 0 {
		return nil, errors.Errorf("errors {%+v}", b.errs)
	}
	return b.unstruct.Object, nil
}

// UnstructList contains a list of Unstructured
// items
type UnstructList struct {
	Items []*Unstruct
}

// ListBuilder enables building a list
// of an Unstruct instance
type ListBuilder struct {
	list *UnstructList
	errs []error
}

// ListBuilderForYamls returns a new instance of
// list Unstruct Builder by making use of the provided YAMLs
func ListBuilderForYamls(yamls ...string) *ListBuilder {
	lb := &ListBuilder{list: &UnstructList{}}
	for _, yaml := range yamls {
		y := strings.Split(strings.Trim(yaml, "---"), "---")
		for _, f := range y {
			f = strings.TrimSpace(f)
			a, err := BuilderForYaml(f).Build()
			if err != nil {
				lb.errs = append(lb.errs, err)
				continue
			}
			lb.list.Items = append(lb.list.Items, a)
		}
	}
	return lb
}

// ListBuilderForObjects returns a mew instance of
// list Unstruct Builder by making use of the provided
// Unstructured object
func ListBuilderForObjects(objs ...*unstructured.Unstructured) *ListBuilder {
	lb := &ListBuilder{list: &UnstructList{}}
	for _, obj := range objs {
		lb.list.Items = append(lb.list.Items, &Unstruct{obj})
	}
	return lb
}

// Build returns the list of Unstruct objects created by
// the Builder
func (l *ListBuilder) Build() ([]*Unstruct, error) {
	if len(l.errs) > 0 {
		return nil, errors.Errorf("errors {%+v}", l.errs)
	}
	return l.list.Items, nil
}

// FromURL provides the unstructured objects from given url
func FromURL(url string) (UnstructList, error) {
	list := UnstructList{}
	// Read yaml file from the url
	read, err := http.Fetch(url)
	if err != nil {
		return list, err
	}
	defer read.Close()
	yamls, err := ioutil.ReadAll(read)
	if err != nil {
		return list, err
	}
	// Connvert the yaml to unstructured objects
	list.Items, err = ListBuilderForYamls(string(yamls)).Build()
	if err != nil {
		return list, err
	}
	return list, nil
}
