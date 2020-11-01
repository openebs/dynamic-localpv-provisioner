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
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	errors "github.com/pkg/errors"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset"
	"k8s.io/apimachinery/pkg/types"
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
type listFn func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.BlockDeviceClaimList, error)

// getFn is a typed function that
// abstracts fetching of block device
type getFn func(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*apis.BlockDeviceClaim, error)

// createFn is a typed function that abstracts
// creation of block device
type createFn func(cli *clientset.Clientset, namespace string, bdc *apis.BlockDeviceClaim) (*apis.BlockDeviceClaim, error)

// deleteFn is a typed function that abstracts
// deletion of bdcs
type deleteFn func(cli *clientset.Clientset, namespace string, name string, deleteOpts *metav1.DeleteOptions) error

// deleteFn is a typed function that abstracts
// deletion of bdc's collection
type deleteCollectionFn func(cli *clientset.Clientset, namespace string, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error

// patchFn is a typed function that abstracts
// to patch block device claim
type patchFn func(cli *clientset.Clientset, namespace, name string, pt types.PatchType, data []byte, subresources ...string) (*apis.BlockDeviceClaim, error)

// updateFn is a typed function that abstracts to update
// block device claim
type updateFn func(cli *clientset.Clientset, namespace string, bdc *apis.BlockDeviceClaim) (*apis.BlockDeviceClaim, error)

// make ndm clientset as singleton
var (
	ndmClientset *clientset.Clientset
	once         sync.Once
)

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
	create              createFn
	del                 deleteFn
	delCollection       deleteCollectionFn
	patch               patchFn
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
			if ndmClientset != nil {
				return ndmClientset, nil
			}
			config, err := kclient.New().GetConfigForPathOrDirect()
			if err != nil {
				return nil, err
			}
			ndmCS, err := clientset.NewForConfig(config)
			if err != nil {
				return nil, err
			}
			once.Do(func() {
				ndmClientset = ndmCS
			})
			return ndmClientset, nil
		}
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (clients *clientset.Clientset, err error) {
			config, err := kclient.New(kclient.WithKubeConfigPath(kubeConfigPath)).GetConfigForPathOrDirect()
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}
	if k.list == nil {
		k.list = func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.BlockDeviceClaimList, error) {
			return cli.OpenebsV1alpha1().BlockDeviceClaims(namespace).List(opts)
		}
	}

	if k.get == nil {
		k.get = func(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*apis.BlockDeviceClaim, error) {
			return cli.OpenebsV1alpha1().BlockDeviceClaims(namespace).Get(name, opts)
		}
	}
	if k.create == nil {
		k.create = func(cli *clientset.Clientset, namespace string, bdc *apis.BlockDeviceClaim) (*apis.BlockDeviceClaim, error) {
			return cli.OpenebsV1alpha1().BlockDeviceClaims(namespace).Create(bdc)
		}
	}
	if k.del == nil {
		k.del = func(cli *clientset.Clientset, namespace string, name string, deleteOpts *metav1.DeleteOptions) error {
			return cli.OpenebsV1alpha1().BlockDeviceClaims(namespace).Delete(name, deleteOpts)
		}
	}
	if k.delCollection == nil {
		k.delCollection = func(cli *clientset.Clientset, namespace string, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
			return cli.OpenebsV1alpha1().BlockDeviceClaims(namespace).DeleteCollection(deleteOpts, listOpts)
		}
	}
	if k.patch == nil {
		k.patch = func(cli *clientset.Clientset, namespace, name string, pt types.PatchType, data []byte, subresources ...string) (*apis.BlockDeviceClaim, error) {
			return cli.OpenebsV1alpha1().BlockDeviceClaims(namespace).Patch(name, pt, data, subresources...)
		}
	}
	if k.update == nil {
		k.update = func(cli *clientset.Clientset, namespace string, bdc *apis.BlockDeviceClaim) (*apis.BlockDeviceClaim, error) {
			return cli.OpenebsV1alpha1().BlockDeviceClaims(namespace).Update(bdc)
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
func (k *Kubeclient) getClientsetOrCached() (*clientset.Clientset, error) {
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
func (k *Kubeclient) List(opts metav1.ListOptions) (*apis.BlockDeviceClaimList, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list bdc in namespace {%s}", k.namespace)
	}
	return k.list(cli, k.namespace, opts)
}

// Get returns a disk object
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*apis.BlockDeviceClaim, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get bdc: missing bdc name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get bdc {%s} in namespace {%s}", name, k.namespace)
	}
	return k.get(cli, k.namespace, name, opts)
}

// Create creates a bdc in specified namespace in kubernetes cluster
func (k *Kubeclient) Create(bdc *apis.BlockDeviceClaim) (*apis.BlockDeviceClaim, error) {
	if bdc == nil {
		return nil, errors.New("failed to create bdc: nil bdc object")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create bdc {%s} in namespace {%s}", bdc.Name, bdc.Namespace)
	}
	return k.create(cli, k.namespace, bdc)
}

// DeleteCollection deletes a collection of bdc objects.
func (k *Kubeclient) DeleteCollection(listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete the collection of bdcs")
	}
	return k.delCollection(cli, k.namespace, listOpts, deleteOpts)
}

// Delete deletes a bdc instance from the
// kubecrnetes cluster
func (k *Kubeclient) Delete(name string, deleteOpts *metav1.DeleteOptions) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("failed to delete bdc: missing bdc name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete bdc {%s} in namespace {%s}", name, k.namespace)
	}
	return k.del(cli, k.namespace, name, deleteOpts)
}

// Patch patches the block device claim if present in kubernetes cluster
func (k *Kubeclient) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*apis.BlockDeviceClaim, error) {
	if len(name) == 0 {
		return nil, errors.New("failed to patch block device claim: missing bdc name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to patch bdc: {%s}", name)
	}
	return k.patch(cli, k.namespace, name, pt, data, subresources...)
}

// Update updates the block device claim if present in kubernetes cluster
func (k *Kubeclient) Update(bdc *apis.BlockDeviceClaim) (*apis.BlockDeviceClaim, error) {
	if bdc == nil {
		return nil, errors.New("failed to udpate bdc: nil bdc object")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update bdc {%s} in namespace {%s}", bdc.Name, bdc.Namespace)
	}
	return k.update(cli, k.namespace, bdc)
}
