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

	"k8s.io/apimachinery/pkg/types"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	errors "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getClientsetFunc is a typed function that
// abstracts fetching internal clientset
type getClientsetFunc func() (cs *clientset.Clientset, err error)

// listFunc is a typed function that abstracts
// listing StoragePool instances
type listFunc func(cs *clientset.Clientset,
	opts metav1.ListOptions) (*apis.StoragePoolList, error)

// getFunc is a typed function that abstracts
// getting StoragePool instances
type getFunc func(cs *clientset.Clientset, name string,
	opts metav1.GetOptions) (*apis.StoragePool, error)

// createFunc is a typed function that abstracts
// creating StoragePool instances
type createFunc func(cs *clientset.Clientset,
	storegePoolObj *apis.StoragePool) (*apis.StoragePool, error)

// patchFunc is a typed function that abstracts
// patching StoragePool instances
type patchFunc func(cs *clientset.Clientset, name string,
	pt types.PatchType, patchObj []byte) (*apis.StoragePool, error)

// delFn is a typed function that abstracts
// delete of StoragePool instances
type delFn func(cs *clientset.Clientset, name string,
	opts *metav1.DeleteOptions) error

// Kubeclient enables kubernetes API operations
// on StoragePool instance
type Kubeclient struct {
	// clientset refers to StoragePool's
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset
	// functions useful during mocking
	getClientset getClientsetFunc
	list         listFunc
	get          getFunc
	create       createFunc
	patch        patchFunc
	del          delFn
}

// KubeclientBuildOption defines the abstraction
// to build a Kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (cs *clientset.Clientset, err error) {
			config, err := client.New().Config()
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}

	if k.list == nil {
		k.list = func(cs *clientset.Clientset,
			opts metav1.ListOptions) (*apis.StoragePoolList, error) {
			return cs.OpenebsV1alpha1().
				StoragePools().
				List(opts)
		}
	}

	if k.get == nil {
		k.get = func(cs *clientset.Clientset,
			name string, opts metav1.GetOptions) (*apis.StoragePool, error) {
			return cs.OpenebsV1alpha1().
				StoragePools().
				Get(name, opts)
		}
	}

	if k.create == nil {
		k.create = func(cs *clientset.Clientset,
			storagePoolObj *apis.StoragePool) (*apis.StoragePool, error) {
			return cs.OpenebsV1alpha1().
				StoragePools().
				Create(storagePoolObj)
		}
	}

	if k.patch == nil {
		k.patch = func(cs *clientset.Clientset, name string,
			pt types.PatchType, patchObj []byte) (*apis.StoragePool, error) {
			return cs.OpenebsV1alpha1().
				StoragePools().
				Patch(name, pt, patchObj)
		}
	}

	if k.del == nil {
		k.del = func(cs *clientset.Clientset, name string,
			opts *metav1.DeleteOptions) error {
			return cs.OpenebsV1alpha1().
				StoragePools().
				Delete(name, opts)
		}
	}
}

// WithClientset sets the kubernetes clientset against
// the kubeclient instance
func WithClientset(c *clientset.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// NewKubeClient returns a new instance of kubeclient meant for
// StoragePool related operations
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// getClientsetOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientsetOrCached() (*clientset.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}
	c, err := k.getClientset()
	if err != nil {
		return nil,
			errors.WithStack(errors.Wrap(err, "failed to get clientset"))
	}
	k.clientset = c
	return k.clientset, nil
}

// List returns a list of StoragePool
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*apis.StoragePoolList, error) {
	cs, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list storage pool")
	}
	return k.list(cs, opts)
}

// Get returns an StoragePool instance from kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*apis.StoragePool, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get storage pool: missing name")
	}
	cs, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get storage pool: {%s}", name)
	}
	return k.get(cs, name, opts)
}

// Create creates an StoragePool instance in kubernetes cluster
func (k *Kubeclient) Create(obj *apis.StoragePool) (*apis.StoragePool, error) {
	if obj == nil {
		return nil, errors.New("failed to create storage pool: missing object")
	}
	cs, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create storage pool: {%v}", obj)
	}
	return k.create(cs, obj)
}

// Patch returns the patched StoragePool instance
func (k *Kubeclient) Patch(name string, pt types.PatchType,
	patchObj []byte) (*apis.StoragePool, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to patch storage pool: missing name")
	}
	cs, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to patch storage pool: {%s}", name)
	}
	return k.patch(cs, name, pt, patchObj)
}

// Delete deletes StoragePool instance
func (k *Kubeclient) Delete(name string, opts *metav1.DeleteOptions) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("failed to delete storage pool: missing name")
	}
	cs, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete storage pool: {%s}", name)
	}
	return k.del(cs, name, opts)
}
