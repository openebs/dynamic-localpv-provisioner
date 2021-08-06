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
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset"
	"github.com/pkg/errors"
)

//TODO: While using these packages UnitTest must be written to corresponding function

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *clientset.Clientset, err error)

// listFn is a typed function that abstracts
// listing of block device
type listFn func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.BlockDeviceList, error)

// getFn is a typed function that
// abstracts fetching of block deivce
type getFn func(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*apis.BlockDevice, error)

// delFn is a typed function that
// abstracts deleting of block deivce
type delFn func(cli *clientset.Clientset, namespace, name string, opts *metav1.DeleteOptions) error

// updateFn is a typed function that abstracts to update
// block device
type updateFn func(cli *clientset.Clientset, namespace string, bd *apis.BlockDevice) (*apis.BlockDevice, error)

// Kubeclient enables kubernetes API operations
// on block device instance
type Kubeclient struct {
	// clientset refers to block device
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset
	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string
	namespace      string
	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	list                listFn
	get                 getFn
	del                 delFn
	update              updateFn
}

// KubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// WithDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) WithDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *clientset.Clientset, err error) {
			config, err := kclient.New().Config()
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (clients *clientset.Clientset, err error) {
			config, err := kclient.New(kclient.WithKubeConfigPath(kubeConfigPath)).
				GetConfigForPathOrDirect()
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}
	if k.list == nil {
		k.list = func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.BlockDeviceList, error) {
			return cli.OpenebsV1alpha1().BlockDevices(namespace).
				List(context.TODO(), opts)
		}
	}

	if k.get == nil {
		k.get = func(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*apis.BlockDevice, error) {
			return cli.OpenebsV1alpha1().BlockDevices(namespace).
				Get(context.TODO(), name, opts)
		}
	}
	if k.del == nil {
		k.del = func(cli *clientset.Clientset, namespace, name string, opts *metav1.DeleteOptions) error {
			return cli.OpenebsV1alpha1().BlockDevices(namespace).
				Delete(context.TODO(), name, *opts)
		}
	}
	if k.update == nil {
		k.update = func(cli *clientset.Clientset, namespace string, bd *apis.BlockDevice) (*apis.BlockDevice, error) {
			return cli.OpenebsV1alpha1().BlockDevices(namespace).
				Update(context.TODO(), bd, metav1.UpdateOptions{})
		}
	}
}

// WithKubeClient sets the kubernetes client against
// the kubeclient instance
func WithKubeClient(c *clientset.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(kubeConfigPath string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = kubeConfigPath
	}
}

// NewKubeClient returns a new instance of kubeclient meant for
// block device operations
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.WithDefaults()
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*clientset.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// WithNamespace sets the kubernetes namespace against
// the provided namespace
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*clientset.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}
	c, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil, err
	}
	k.clientset = c
	return k.clientset, nil
}

// List returns a list of disk
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*apis.BlockDeviceList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, k.namespace, opts)
}

// Get returns a disk object
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*apis.BlockDevice, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, k.namespace, name, opts)
}

// Delete deletes a disk object
func (k *Kubeclient) Delete(name string, opts *metav1.DeleteOptions) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.del(cli, k.namespace, name, opts)
}

// Update updates the block device claim if present in kubernetes cluster
func (k *Kubeclient) Update(bd *apis.BlockDevice) (*apis.BlockDevice, error) {
	if bd == nil {
		return nil, errors.New("failed to udpate bdc: nil bdc object")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update bdc {%s} in namespace {%s}", bd.Name, bd.Namespace)
	}
	return k.update(cli, k.namespace, bd)
}
