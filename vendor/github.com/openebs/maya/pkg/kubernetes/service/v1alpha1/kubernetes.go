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

	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(
	kubeConfigPath string,
) (*kubernetes.Clientset, error)

// getFn is a typed function that abstracts get of service instances
type getFn func(
	cli *kubernetes.Clientset,
	name, namespace string,
	opts metav1.GetOptions,
) (*corev1.Service, error)

// listFn is a typed function that abstracts list of service instances
type listFn func(
	cli *kubernetes.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*corev1.ServiceList, error)

// delFn is a typed function that abstracts delete of service instances
type delFn func(
	cli *kubernetes.Clientset,
	name, namespace string,
	opts *metav1.DeleteOptions,
) error

// createFn is a typed function that abstracts delete of service instances
type createFn func(
	cli *kubernetes.Clientset,
	service *corev1.Service,
	namespace string,
) (*corev1.Service, error)

// updateFn is a typed function that abstracts delete of service instances
type updateFn func(
	cli *kubernetes.Clientset,
	service *corev1.Service,
	namespace string,
) (*corev1.Service, error)

// patchFn is a typed function that abstracts patch of service instances
type patchFn func(
	cli *kubernetes.Clientset,
	name, namespace string,
	pt types.PatchType,
	data []byte,
	subresources ...string,
) (*corev1.Service, error)

// Kubeclient enables kubernetes API operations on service instance
type Kubeclient struct {
	// clientset refers to kubernetes clientset. It is responsible to
	// make kubernetes API calls for crud op
	clientset *kubernetes.Clientset
	namespace string

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	list                listFn
	get                 getFn
	del                 delFn
	create              createFn
	patch               patchFn
	update              updateFn
}

// KubeclientBuildOption defines the abstraction to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// defaultGetClientset is the default implementation to
// get kubernetes clientset instance
func defaultGetClientset() (clients *kubernetes.Clientset, err error) {
	config, err := client.New().Config()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// defaultGetClientsetForPath is the default implementation to
// get kubernetes clientset instance based on the given
// kubeconfig path
func defaultGetClientsetForPath(
	kubeConfigPath string,
) (clients *kubernetes.Clientset, err error) {
	return client.New(client.WithKubeConfigPath(kubeConfigPath)).
		Clientset()
}

// defaultCreate is the default implementation to get
// a service instance in kubernetes cluster
func defaultGet(
	cli *kubernetes.Clientset,
	name, namespace string,
	opts metav1.GetOptions,
) (r *corev1.Service, err error) {
	r, err = cli.CoreV1().
		Services(namespace).
		Get(name, opts)
	return
}

// defaultCreate is the default implementation to list
// service instances in kubernetes cluster
func defaultList(
	cli *kubernetes.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (rl *corev1.ServiceList, err error) {
	rl, err = cli.CoreV1().
		Services(namespace).
		List(opts)
	return
}

// defaultCreate is the default implementation to delete
// a service instance in kubernetes cluster
func defaultDel(
	cli *kubernetes.Clientset,
	name, namespace string,
	opts *metav1.DeleteOptions,
) (err error) {
	deletePropagation := metav1.DeletePropagationForeground
	opts.PropagationPolicy = &deletePropagation
	err = cli.CoreV1().
		Services(namespace).
		Delete(name, opts)
	return
}

// defaultCreate is the default implementation to create
// a service instance in kubernetes cluster
func defaultCreate(
	cli *kubernetes.Clientset,
	service *corev1.Service,
	namespace string,
) (*corev1.Service, error) {
	return cli.CoreV1().
		Services(namespace).
		Create(service)
}

// defaultUpdate is the default implementation to update
// a service instance in kubernetes cluster
func defaultUpdate(
	cli *kubernetes.Clientset,
	service *corev1.Service,
	namespace string,
) (*corev1.Service, error) {
	return cli.CoreV1().
		Services(namespace).
		Update(service)
}

// defaultPatch is the default implementation to patch
// a service instance in kubernetes cluster
func defaultPatch(
	cli *kubernetes.Clientset,
	name, namespace string,
	pt types.PatchType,
	data []byte,
	subresources ...string,
) (*corev1.Service, error) {
	return cli.CoreV1().
		Services(namespace).
		Patch(name, pt, data, subresources...)
}

// withDefaults sets the default options of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = defaultGetClientset
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = defaultGetClientsetForPath
	}
	if k.get == nil {
		k.get = defaultGet
	}
	if k.list == nil {
		k.list = defaultList
	}
	if k.del == nil {
		k.del = defaultDel
	}
	if k.create == nil {
		k.create = defaultCreate
	}
	if k.patch == nil {
		k.patch = defaultPatch
	}
	if k.update == nil {
		k.update = defaultUpdate
	}
}

// WithClientset sets the kubernetes client against the kubeclient instance
func WithClientset(c *kubernetes.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithNamespace set namespace in kubeclient object
func WithNamespace(namespace string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.namespace = namespace
	}
}

// WithNamespace set namespace in kubeclient object
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// NewKubeClient returns a new instance of kubeclient meant for service,
// caller can configure it with different kubeclientBuildOption
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (
	*kubernetes.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientOrCached returns either a new
// instance of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*kubernetes.Clientset, error) {
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

// Get returns service object for given name
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (
	*corev1.Service, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, name, k.namespace, opts)
}

// GetRaw returns service object for given name in byte format
func (k *Kubeclient) GetRaw(name string, opts metav1.GetOptions) (
	[]byte, error) {
	svc, err := k.Get(name, opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(svc)
}

// List returns list of services
func (k *Kubeclient) List(opts metav1.ListOptions) (
	*corev1.ServiceList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, k.namespace, opts)
}

// ListRaw returns list of services in byte format
func (k *Kubeclient) ListRaw(opts metav1.ListOptions) ([]byte, error) {
	svcList, err := k.List(opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(svcList)
}

// Delete returns service object for given name
func (k *Kubeclient) Delete(name string, options *metav1.DeleteOptions) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.del(cli, name, k.namespace, options)
}

// Create creates a service in specified namespace in kubernetes cluster
func (k *Kubeclient) Create(service *corev1.Service) (*corev1.Service, error) {
	if service == nil {
		return nil, errors.New("failed to create service: nil service object")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to create service {%s} in namespace {%s}",
			service.Name,
			service.Namespace,
		)
	}
	return k.create(cli, service, k.namespace)
}

// Update updates a service in specified namespace in kubernetes cluster
func (k *Kubeclient) Update(service *corev1.Service) (*corev1.Service, error) {
	if service == nil {
		return nil, errors.New("failed to update service: nil service object")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to update service {%s} in namespace {%s}",
			service.Name,
			service.Namespace,
		)
	}
	return k.update(cli, service, k.namespace)
}

// Patch patches service object for given name
func (k *Kubeclient) Patch(
	name string,
	pt types.PatchType,
	data []byte,
	subresources ...string,
) (*corev1.Service, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}

	return k.patch(cli, name, k.namespace, pt, data, subresources...)
}
