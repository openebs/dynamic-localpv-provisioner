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
	"strings"

	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	errors "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset dynamic.Interface, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset dynamic.Interface, err error)

// CreateFn is a typed function that abstracts
// creating of unstructured object
type CreateFn func(
	cli dynamic.Interface,
	obj *unstructured.Unstructured,
	opts *CreateOption,
) (*unstructured.Unstructured, error)

// GetFn is a typed function that abstracts
// fetching of unstructured object
type GetFn func(
	cli dynamic.Interface,
	name string,
	namespace string, opt *GetOption,
) (*unstructured.Unstructured, error)

// DeleteFn is a typed function that abstract  deletion
// of unstructured object
type DeleteFn func(
	cli dynamic.Interface,
	obj *unstructured.Unstructured,
	opt *DeleteOption,
) error

// Kubeclient enables kubernetes API operations on catalog
// instance
type Kubeclient struct {
	// clientset refers to clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset dynamic.Interface

	// Kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	create              CreateFn
	get                 GetFn
	delete              DeleteFn
}

// KubeclientBuildOption defines the abstraction to build
// a Kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets default options for Kubeclient
func withDefaults(k *Kubeclient) {
	if k.clientset == nil {
		k.getClientset = func() (dynamic.Interface, error) {
			return client.New().Dynamic()
		}
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (dynamic.Interface, error) {
			return client.New(client.WithKubeConfigPath(k.kubeConfigPath)).Dynamic()
		}
	}
	if k.get == nil {
		k.get = func(
			cli dynamic.Interface,
			name string,
			namespace string, opt *GetOption) (*unstructured.Unstructured, error) {
			return cli.
				Resource(opt.gvr).
				Namespace(namespace).
				Get(name, *opt.GetOptions, opt.subresources...)
		}
	}
	if k.create == nil {
		k.create = func(
			cli dynamic.Interface,
			obj *unstructured.Unstructured,
			opt *CreateOption) (*unstructured.Unstructured, error) {
			return cli.
				Resource(k8s.GroupVersionResourceFromGVK(obj)).
				Namespace(obj.GetNamespace()).
				Create(obj, *opt.CreateOptions, opt.subresources...)
		}
	}
	if k.delete == nil {
		k.delete = func(
			cli dynamic.Interface,
			obj *unstructured.Unstructured, opt *DeleteOption) error {
			return cli.
				Resource(k8s.GroupVersionResourceFromGVK(obj)).
				Namespace(obj.GetNamespace()).
				Delete(obj.GetName(), opt.DeleteOptions, opt.subresources...)
		}
	}
}

// WithClient sets the kubernetes client against
// the Kubeclient instance
func WithClient(c dynamic.Interface) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets kubeconfig path
// against this client instance
func WithKubeConfigPath(kubeConfigPath string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = kubeConfigPath
	}
}

// NewKubeClient returns a new instance of Kubeclient meant for
// catalog operations
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	withDefaults(k)
	return k
}

// getClientsetForPathOrDirect returns new instance of kubernetes client
func (k *Kubeclient) getClientsetForPathOrDirect() (dynamic.Interface, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientsetOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientsetOrCached() (dynamic.Interface, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}
	cs, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get clientset")
	}

	k.clientset = cs
	return k.clientset, nil
}

// Get returns an unstructured instance from kubernetes
// cluster
func (k *Kubeclient) Get(name string, opts ...GetOptionFn) (*unstructured.Unstructured, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get unstructured instance: missing name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, err
	}
	getOptions := NewGetOption(opts...)
	return k.get(
		cli,
		name,
		getOptions.namespace,
		getOptions,
	)
}

// CreateAllOrNone creates all the provided
// unstructured instances at kubernetes cluster
// or none in case of any error
func (k *Kubeclient) CreateAllOrNone(u ...*unstructured.Unstructured) []error {
	var (
		errs    []error
		created []*unstructured.Unstructured
	)
	for _, ustruct := range u {
		err := k.Create(ustruct)
		if err != nil {
			errs = append(errs, err)
			break
		}
		created = append(created, ustruct)
	}
	if len(errs) > 0 {
		k.DeleteAll(created...)
	}
	return errs
}

// DeleteAll deletes all the provided unstructured
// instances at kubernetes cluster
func (k *Kubeclient) DeleteAll(u ...*unstructured.Unstructured) []error {
	var errs []error
	for _, ustruct := range u {
		err := k.Delete(ustruct)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Create creates an unstructured instance at
// kubernetes cluster
func (k *Kubeclient) Create(u *unstructured.Unstructured, opts ...CreateOptionFn) error {
	if u == nil {
		return errors.Errorf("create failed: nil unstruct instance was provided")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return err
	}
	cOptions := NewCreateOption(opts...)
	_, err = k.create(cli, u, cOptions)
	return err
}

// Delete deletes the unstructured instance from
// kubernetes cluster
func (k *Kubeclient) Delete(u *unstructured.Unstructured, opts ...DeleteOptionFn) error {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return err
	}
	dOptions := NewDeleteOption(opts...)
	return k.delete(cli, u, dOptions)
}

// TODO:
// Implement builder pattern for the below functions

// GetOption holds the kubernetes options
// to get a resource
type GetOption struct {
	namespace string
	*metav1.GetOptions
	gvr          schema.GroupVersionResource
	subresources []string
}

// GetOptionFn abstracts the construction of GetOption
type GetOptionFn func(*GetOption)

// WithGetNamespace is a GetOptionFn to provide
// namespace
func WithGetNamespace(namespace string) GetOptionFn {
	return func(opt *GetOption) {
		opt.namespace = namespace
	}
}

// WithGetOption is a GetOptionsFn to provide
// kubernetes getoption
func WithGetOption(getOption metav1.GetOptions) GetOptionFn {
	return func(opt *GetOption) {
		opt.GetOptions = &getOption
	}
}

// WithGroupVersionResource is a GetOptionFn to provide
// GroupResourceVersion
func WithGroupVersionResource(r schema.GroupVersionResource) GetOptionFn {
	return func(opt *GetOption) {
		opt.gvr = r
	}
}

// WithGetSubResources is a GetOptionFn to provide
// subresources
func WithGetSubResources(r ...string) GetOptionFn {
	return func(opt *GetOption) {
		opt.subresources = r
	}
}

// NewGetOption returns a new instance of GetOption
func NewGetOption(gOpts ...GetOptionFn) *GetOption {
	opts := &GetOption{GetOptions: &metav1.GetOptions{}, gvr: schema.GroupVersionResource{}}
	for _, o := range gOpts {
		o(opts)
	}
	return opts
}

// DeleteOption holds kubernetes options to delete a
// resource
type DeleteOption struct {
	*metav1.DeleteOptions
	subresources []string
}

// NewDeleteOption returns a new instance of
// DeleteOption
func NewDeleteOption(dOpts ...DeleteOptionFn) *DeleteOption {
	opts := &DeleteOption{DeleteOptions: &metav1.DeleteOptions{}}
	for _, o := range dOpts {
		o(opts)
	}
	return opts
}

// DeleteOptionFn abstracts the construvtion of delete
// option
type DeleteOptionFn func(*DeleteOption)

// WithDeleteOption is a DeleteOptionFn to provide
// kubernetes delete option
func WithDeleteOption(deleteOpt *metav1.DeleteOptions) DeleteOptionFn {
	return func(opt *DeleteOption) {
		opt.DeleteOptions = deleteOpt
	}
}

// WithDeleteSubResources is a DeleteOptionFn to provide
// subresources during delete
func WithDeleteSubResources(r ...string) DeleteOptionFn {
	return func(opt *DeleteOption) {
		opt.subresources = r
	}
}

// CreateOption holds the kubernetes option to create a resource
type CreateOption struct {
	*metav1.CreateOptions
	subresources []string
}

// NewCreateOption returns a new instance of CreateOption
func NewCreateOption(cOpts ...CreateOptionFn) *CreateOption {
	opts := &CreateOption{&metav1.CreateOptions{}, []string{}}
	for _, o := range cOpts {
		o(opts)
	}
	return opts
}

// CreateOptionFn abstracts the construction of CreateOption
type CreateOptionFn func(*CreateOption)

// WithCreateOption is CreateOptionFn to provide kubernetes
// createOption for creating a resource
func WithCreateOption(opt metav1.CreateOptions) CreateOptionFn {
	return func(createOpt *CreateOption) {
		createOpt.CreateOptions = &opt
	}
}

// WithCreateSubResources is CreateOptionFn to kubernetes
// subresources during resource creation
func WithCreateSubResources(r ...string) CreateOptionFn {
	return func(createOpt *CreateOption) {
		createOpt.subresources = r
	}
}
