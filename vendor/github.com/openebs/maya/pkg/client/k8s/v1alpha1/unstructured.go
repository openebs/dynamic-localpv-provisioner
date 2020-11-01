/*
Copyright 2018 The OpenEBS Authors

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

// TODO
// Move this to pkg/unstruct/v1alpha1/unstruct.go

// TODO
// Create a new struct called unstruct that wraps unstructured.Unstructured

// An unstructured instance is a JSON compatible structure and is compatible
// to any kubernetes resource (native as well as custom). This instance can be
// used during following kubernetes API invocations:
//
// - Create
// - Update
// - UpdateStatus
// - Delete
// - DeleteCollection
// - Get
// - List
// - Watch
// - Patch

package v1alpha1

import (
	"strings"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/pkg/version"
	kubever "github.com/openebs/maya/pkg/version/kubernetes"
	errors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// kind represents the name of the kubernetes kind
type kind string

// resource converts the name of the kubernetes kind to corresponding kubernetes
// resource's name
//
// NOTE:
//  This may not be the best of approaches to get name of a resource. However,
// this fits the current requirement. This might need a revisit depending on
// future requirements.
func (k kind) resource() (resource string) {
	resource = strings.ToLower(string(k))
	switch resource {
	case "":
		return
	case "storageclass":
		return "storageclasses"
	default:
		return resource + "s"
	}
}

// isNamespaced flags if the kind is namespaced or not
//
// NOTE:
//  This may not be the best of approaches to flag a resource as namespaced or
// not. However, this fits the current requirement. This might need a revisit
// depending on future requirements.
func (k kind) isNamespaced() (no bool) {
	ks := strings.ToLower(string(k))
	switch ks {
	case "customresourcedefinition":
		return no
	case "storageclass":
		return no
	case "persistentvolume":
		return no
	case "castemplate":
		return no
	case "storagepoolclaim":
		return no
	case "cstorpool":
		return no
	case "storagepool":
		return no
	default:
		return !no
	}
}

// GroupVersionResourceFromGVK returns the GroupVersionResource of the provided
// unstructured instance by making use of this instance's GroupVersionKind info
//
// NOTE:
//  Resource is assumed as plural of Kind
func GroupVersionResourceFromGVK(unstructured *unstructured.Unstructured) (gvr schema.GroupVersionResource) {
	if unstructured == nil {
		return
	}

	gvk := unstructured.GroupVersionKind()

	gvr.Group = strings.ToLower(gvk.Group)
	gvr.Version = strings.ToLower(gvk.Version)
	gvr.Resource = kind(gvk.Kind).resource()

	return
}

// WithBytesUnstructuredCreator abstracts creation of unstructured instance
// from the provided bytes
type WithBytesUnstructuredCreator func(raw []byte) (*unstructured.Unstructured, error)

// CreateUnstructuredFromYamlBytes creates an unstructured instance from the
// provided YAML document in bytes
//
// NOTE:
//  This is an implementation of WithBytesUnstructuredCreator
func CreateUnstructuredFromYamlBytes(raw []byte) (*unstructured.Unstructured, error) {
	m := map[string]interface{}{}
	err := yaml.Unmarshal(raw, &m)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create unstructured instance from yaml: %s", raw)
	}

	return &unstructured.Unstructured{
		Object: m,
	}, nil
}

// UnstructuredCreator abstracts creation of unstructured instance from the
// provided document
type UnstructuredCreator func(document string) (*unstructured.Unstructured, error)

// CreateUnstructuredFromYaml creates an unstructured instance from the
// provided YAML document
func CreateUnstructuredFromYaml(document string) (*unstructured.Unstructured, error) {
	return CreateUnstructuredFromYamlBytes([]byte(document))
}

// CreateUnstructuredFromJson creates an unstructured instance from the
// provided JSON document
func CreateUnstructuredFromJson(document string) (*unstructured.Unstructured, error) {
	uncastObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, []byte(document))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create unstructured instance from json document")
	}

	return uncastObj.(*unstructured.Unstructured), nil
}

// UnstructuredMiddleware abstracts updating given unstructured instance
type UnstructuredMiddleware func(given *unstructured.Unstructured) (updated *unstructured.Unstructured)

// UnstructuredPredicate abstracts evaluating a condition against the provided
// unstructured instance
type UnstructuredPredicate func(given *unstructured.Unstructured) bool

// UnstructuredPredicateList represents a list of unstructured predicates
type UnstructuredPredicateList []UnstructuredPredicate

// All evaluates if all the predicates succeed
func (l UnstructuredPredicateList) All(given *unstructured.Unstructured) bool {
	for _, p := range l {
		if !p(given) {
			return false
		}
	}
	return true
}

// Any evaluates if at least one of the predicates succeed
func (l UnstructuredPredicateList) Any(given *unstructured.Unstructured) bool {
	if len(l) == 0 {
		return true
	}
	for _, p := range l {
		if p(given) {
			return true
		}
	}
	return false
}

// IsNamespaceScoped flags if the given unstructured instance is namespace
// scoped
//
// NOTE:
//  This is a UnstructuredPredicate implementation
func IsNamespaceScoped(given *unstructured.Unstructured) bool {
	if given == nil {
		return false
	}
	return kind(given.GetKind()).isNamespaced()
}

// IsCASTemplate flags if the given unstructured instance is a CASTemplate
// scoped
//
// NOTE:
//  This is a UnstructuredPredicate implementation
func IsCASTemplate(given *unstructured.Unstructured) bool {
	if given == nil {
		return false
	}
	return strings.ToLower(given.GetKind()) == "castemplate"
}

// IsNameUnVersioned flags if the given unstructured instance name has
// version as its suffix
//
// NOTE:
//  This is a UnstructuredPredicate implementation
func IsNameUnVersioned(given *unstructured.Unstructured) bool {
	if given == nil {
		return false
	}
	return !IsNameVersioned(given)
}

// IsNameVersioned flags if the given unstructured instance name does not have
// version as its suffix
//
// NOTE:
//  This is a UnstructuredPredicate implementation
func IsNameVersioned(given *unstructured.Unstructured) bool {
	if given == nil {
		return false
	}
	return version.IsVersioned(given.GetName())
}

// IsRunTask flags if the given unstructured instance is a RunTask
// scoped
//
// NOTE:
//  This is a UnstructuredPredicate implementation
func IsRunTask(given *unstructured.Unstructured) bool {
	if given == nil {
		return false
	}
	return strings.ToLower(given.GetKind()) == "runtask"
}

// UpdateNamespace updates the given unstructured instance's namespace with
// a valid namespace i.e. non-empty namespace
func UpdateNamespace(n string) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) (updated *unstructured.Unstructured) {
		namespace := strings.TrimSpace(n)
		if given == nil || len(namespace) == 0 {
			return given
		}
		given.SetNamespace(namespace)
		return given
	}
}

// AddKubeServerVersionToLabels adds kubernetes server version to instance's
// labels
//
// TODO
// Move this override flag to a Predicate. This will follow idiomatic Maya
// convention to separate conditional(s) logic from core business logic. This
// in turn helps in readability & maintainability of the codebase.
//
// e.g.
// func IsLabelSet(key string) UnstructuredPredicate {...}
// func IsLabelUnSet(key string) UnstructuredPredicate {...}
func AddKubeServerVersionToLabels(override bool) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) (updated *unstructured.Unstructured) {
		if given == nil {
			return given
		}
		l := given.GetLabels()
		if l == nil {
			l = map[string]string{}
		}
		// TODO
		// Nested if else clause is bad. This will get handled once IsLabelUnSet
		// predicate is used.
		if len(l[string(v1alpha1.KubeServerVersionPlainKey)]) == 0 || override {
			// TODO
			// error should not be ignored here
			// Need to design a custom struct that wraps unstructured.Unstructured
			// instance and has fields to accomodate error(s), warning(s), etc
			//
			// NOTE:
			//  This will be done along with idiomatic maya refactorings
			vInfo, err := GetServerVersion()
			if err != nil {
				return given
			}
			l[string(v1alpha1.KubeServerVersionPlainKey)] = kubever.AsLabelValue(vInfo.GitVersion)
		}
		given.SetLabels(l)
		return given
	}
}

// AddNameToLabels extracts the instance's name & adds it to the same instance's
// labels mapped by the provided label key
//
// TODO
// Move this override flag to a Predicate. This will follow idiomatic Maya
// convention to separate conditional(s) logic from core business logic. This
// in turn helps in readability & maintainability of the codebase.
//
// e.g.
// func IsLabelSet(key string) UnstructuredPredicate {...}
// func IsLabelUnSet(key string) UnstructuredPredicate {...}
func AddNameToLabels(key string, override bool) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) (updated *unstructured.Unstructured) {
		if given == nil || len(key) == 0 {
			return given
		}
		l := given.GetLabels()
		if l == nil {
			l = map[string]string{}
		}
		if override {
			l[key] = given.GetName()
		}
		if len(l[key]) == 0 {
			l[key] = given.GetName()
		}
		given.SetLabels(l)
		return given
	}
}

// SuffixNameWithVersion suffixes the given unstructured instance's name with
// current version
// Converting to lowercase is required as the version can have custom tags having
// uppercase characters like 1.11.0-ce-RC2 and uppercase characters are not valid
// in names of the k8s resources
func SuffixNameWithVersion() UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) (updated *unstructured.Unstructured) {
		if given == nil || IsNameVersioned(given) {
			return given
		}
		given.SetName(version.WithSuffixLower(given.GetName()))
		return given
	}
}

// SuffixSliceWithVersionAtPath updates the given unstructured instance's
// path with provided slice after suffixing each slice item with version
func SuffixSliceWithVersionAtPath(o []interface{}, path string) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) (updated *unstructured.Unstructured) {
		if given == nil {
			return given
		}
		slice := make([]string, 0, len(o))
		for _, v := range o {
			if s, ok := v.(string); ok {
				slice = append(slice, s)
			}
		}
		if len(slice) != len(o) {
			return given
		}
		u := version.WithSuffixesIf(slice, version.IsNotVersioned)
		util.SetNestedSlice(given.Object, u, strings.Split(path, ".")...)
		return given
	}
}

// SuffixStringSliceWithVersionAtPath updates the given unstructured instance's
// path with provided slice after suffixing each slice item with version
func SuffixStringSliceWithVersionAtPath(s []string, path string) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) (updated *unstructured.Unstructured) {
		if given == nil {
			return given
		}
		u := version.WithSuffixesIf(s, version.IsNotVersioned)
		util.SetNestedSlice(given.Object, u, strings.Split(path, ".")...)
		return given
	}
}

// SuffixStringWithVersionAtPath updates the given unstructured instance's
// path with provided string after suffixing it with version
func SuffixStringWithVersionAtPath(s, path string) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) (updated *unstructured.Unstructured) {
		if given == nil {
			return given
		}
		u := version.WithSuffixIf(s, version.IsNotVersioned)
		util.SetNestedField(given.Object, u, strings.Split(path, ".")...)
		return given
	}
}

// SuffixWithVersionAtPath suffixes the value(s) extracted from the provided path
// with version
//
// NOTE: Currently supports path having following values:
// - string
// - []string
func SuffixWithVersionAtPath(path string) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) (updated *unstructured.Unstructured) {
		if given == nil {
			return given
		}
		o := util.GetNestedField(given.Object, strings.Split(path, ".")...)
		slice, ok := o.([]interface{})
		if ok {
			return SuffixSliceWithVersionAtPath(slice, path)(given)
		}
		sslice, ok := o.([]string)
		if ok {
			return SuffixStringSliceWithVersionAtPath(sslice, path)(given)
		}
		s, ok := o.(string)
		if ok {
			return SuffixStringWithVersionAtPath(s, path)(given)
		}
		return given
	}
}

// UpdateLabels updates the unstructured instance's labels
//
// TODO
// Move this override flag to a Predicate. This will follow idiomatic Maya
// convention to separate conditional(s) logic from core business logic. This
// in turn helps in readability & maintainability of the codebase.
//
// e.g.
// func IsLabelSet(key string) UnstructuredPredicate {...}
// func IsLabelUnSet(key string) UnstructuredPredicate {...}
func UpdateLabels(l map[string]string, override bool) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) (updated *unstructured.Unstructured) {
		if given == nil || len(l) == 0 {
			return given
		}
		orig := given.GetLabels()
		if orig == nil {
			orig = map[string]string{}
		}
		for k, v := range l {
			if override {
				orig[k] = v
			}
			if len(orig[k]) == 0 {
				orig[k] = v
			}
		}
		given.SetLabels(orig)
		return given
	}
}

// TODO
// Convert this into a method of yet to be defined unstruct struct

// UnstructuredMap maps the given unstructured instance by executing the
// provided middleware. Map is considered if all the provided predicate
// evaluations succeed.
func UnstructuredMap(m UnstructuredMiddleware, p ...UnstructuredPredicate) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) *unstructured.Unstructured {
		conds := UnstructuredPredicateList(p)
		if conds.All(given) {
			return m(given)
		}
		return given
	}
}

// TODO
// Convert this into a method of yet to be defined unstruct struct

// UnstructuredMapIfAny maps the given unstructured instance by executing the
// provided middleware. Map is considered if at least one of the provided
// predicate evaluations succeed.
func UnstructuredMapIfAny(m UnstructuredMiddleware, p ...UnstructuredPredicate) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) *unstructured.Unstructured {
		conds := UnstructuredPredicateList(p)
		if conds.Any(given) {
			return m(given)
		}
		return given
	}
}

// TODO
// Convert this into a method of yet to be defined unstruct struct

// UnstructuredMapAll maps the given unstructured instance by executing all
// the provided middlewares. Map is considered if the provided predicate
// evaluations succeed.
func UnstructuredMapAll(m []UnstructuredMiddleware, p ...UnstructuredPredicate) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) *unstructured.Unstructured {
		for _, mid := range m {
			given = UnstructuredMap(mid, p...)(given)
		}
		return given
	}
}

// TODO
// Convert this into a method of yet to be defined unstruct struct

// UnstructuredMapAllIfAny maps the given unstructured instance by executing all
// the provided middlewares. Map is considered if atleast one of the provided
// predicate evaluations succeed.
func UnstructuredMapAllIfAny(m []UnstructuredMiddleware, p ...UnstructuredPredicate) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) *unstructured.Unstructured {
		for _, mid := range m {
			given = UnstructuredMapIfAny(mid, p...)(given)
		}
		return given
	}
}

// UnstructedList represents a list of unstructured instances
type UnstructedList struct {
	Items []*unstructured.Unstructured
}

// Map will map each of its instance against the provided middleware if the
// provided predicate evaluations succeed
func (u UnstructedList) Map(m UnstructuredMiddleware, p ...UnstructuredPredicate) (ul UnstructedList) {
	for _, unstruct := range u.Items {
		unstruct = UnstructuredMap(m, p...)(unstruct)
		if unstruct != nil {
			ul.Items = append(ul.Items, unstruct)
		}
	}
	return
}

// MapAll will map all the provided middlewares against each of its instance if
// the provided predicate evaluations succeed
func (u UnstructedList) MapAll(m []UnstructuredMiddleware, p ...UnstructuredPredicate) (ul UnstructedList) {
	for _, unstruct := range u.Items {
		unstruct = UnstructuredMapAll(m, p...)(unstruct)
		if unstruct != nil {
			ul.Items = append(ul.Items, unstruct)
		}
	}
	return
}

// MapAllIfAny will map all the provided middlewares against each of its
// instance if at least one of the provided predicate evaluations succeed
func (u UnstructedList) MapAllIfAny(m []UnstructuredMiddleware, p ...UnstructuredPredicate) (ul UnstructedList) {
	for _, unstruct := range u.Items {
		unstruct = UnstructuredMapAllIfAny(m, p...)(unstruct)
		if unstruct != nil {
			ul.Items = append(ul.Items, unstruct)
		}
	}
	return
}
