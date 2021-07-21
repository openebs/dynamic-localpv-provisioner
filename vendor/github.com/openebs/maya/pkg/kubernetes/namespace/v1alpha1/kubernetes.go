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
	"strings"

	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *kubernetes.Clientset, err error)

// getFn is a typed function that abstracts
// to get namespace
type getFn func(cli *kubernetes.Clientset, name string, opts metav1.GetOptions) (*corev1.Namespace, error)

// createFn is a typed function that abstracts
// creation of namespace
type createFn func(cli *kubernetes.Clientset, namespace *corev1.Namespace) (*corev1.Namespace, error)

// deleteFn is a typed function that abstracts
// deletion of namespaces
type deleteFn func(cli *kubernetes.Clientset, namespace string, deleteOpts *metav1.DeleteOptions) error

// Kubeclient enables kubernetes API operations
// on namespace instance
type Kubeclient struct {
	// clientset refers to namespace clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *kubernetes.Clientset

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	get                 getFn
	create              createFn
	del                 deleteFn
}

// KubeclientBuildOption abstracts creating an
// instance of kubeclient
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *kubernetes.Clientset, err error) {
			return client.New().Clientset()
		}
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (clients *kubernetes.Clientset, err error) {
			return client.New(client.WithKubeConfigPath(kubeConfigPath)).Clientset()
		}
	}
	if k.get == nil {
		k.get = func(cli *kubernetes.Clientset, name string, opts metav1.GetOptions) (*corev1.Namespace, error) {
			return cli.CoreV1().Namespaces().Get(name, opts)
		}
	}
	if k.del == nil {
		k.del = func(cli *kubernetes.Clientset, name string, deleteOpts *metav1.DeleteOptions) error {
			return cli.CoreV1().Namespaces().Delete(name, deleteOpts)
		}
	}
	if k.create == nil {
		k.create = func(cli *kubernetes.Clientset, namespace *corev1.Namespace) (*corev1.Namespace, error) {
			return cli.CoreV1().Namespaces().Create(namespace)
		}
	}
}

// WithClientSet sets the kubernetes client against
// the kubeclient instance
func WithClientSet(c *kubernetes.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// NewKubeClient returns a new instance of kubeclient meant for
// cstor volume replica operations
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*kubernetes.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientsetOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientsetOrCached() (*kubernetes.Clientset, error) {
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

// Get returns a namespace resource
// instances present in kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*corev1.Namespace, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get namespace: missing namespace name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get namespace {%s}", name)
	}
	return k.get(cli, name, opts)
}

// Delete deletes a namespace instance from the
// kubecrnetes cluster
func (k *Kubeclient) Delete(name string, deleteOpts *metav1.DeleteOptions) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("failed to delete namespace: missing namespace name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete namespace {%s}", name)
	}
	return k.del(cli, name, deleteOpts)
}

// Create creates a namespace in specified namespace in kubernetes cluster
func (k *Kubeclient) Create(namespace *corev1.Namespace) (*corev1.Namespace, error) {
	if namespace == nil {
		return nil, errors.New("failed to create namespace: nil namespace object")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create namespace {%s}", namespace.Name)
	}
	return k.create(cli, namespace)
}
