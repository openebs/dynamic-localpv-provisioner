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
// listing CStorPool instances
type listFunc func(cs *clientset.Clientset, opts metav1.ListOptions) (*apis.CStorPoolList, error)

// getFunc is a typed function that abstracts
// getting CStorPool instances
type getFunc func(cs *clientset.Clientset, name string, opts metav1.GetOptions) (*apis.CStorPool, error)

// createFunc is a typed function that abstracts
// creating CStorPool instances
type createFunc func(cs *clientset.Clientset, storegePoolObj *apis.CStorPool) (*apis.CStorPool, error)

// patchFunc is a typed function that abstracts
// patching CStorPool instances
type patchFunc func(cs *clientset.Clientset, name string, pt types.PatchType, patchObj []byte) (*apis.CStorPool, error)

// delFn is a typed function that abstracts
// delete of StoragePool instances
type delFn func(cs *clientset.Clientset, name string,
	opts *metav1.DeleteOptions) error

// Kubeclient enables kubernetes API operations
// on CStorPool instance
type Kubeclient struct {
	// clientset refers to CStorPool's
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
			opts metav1.ListOptions) (*apis.CStorPoolList, error) {
			return cs.OpenebsV1alpha1().
				CStorPools().
				List(opts)
		}
	}

	if k.get == nil {
		k.get = func(cs *clientset.Clientset,
			name string, opts metav1.GetOptions) (*apis.CStorPool, error) {
			return cs.OpenebsV1alpha1().
				CStorPools().
				Get(name, opts)
		}
	}

	if k.create == nil {
		k.create = func(cs *clientset.Clientset,
			obj *apis.CStorPool) (*apis.CStorPool, error) {
			return cs.OpenebsV1alpha1().
				CStorPools().
				Create(obj)
		}
	}

	if k.patch == nil {
		k.patch = func(cs *clientset.Clientset, name string,
			pt types.PatchType, patchObj []byte) (*apis.CStorPool, error) {
			return cs.OpenebsV1alpha1().
				CStorPools().
				Patch(name, pt, patchObj)
		}
	}

	if k.del == nil {
		k.del = func(cs *clientset.Clientset, name string,
			opts *metav1.DeleteOptions) error {
			return cs.OpenebsV1alpha1().
				CStorPools().
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
// CStorPool related operations
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

// List returns a list of CStorPool
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*apis.CStorPoolList, error) {
	cs, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list cstor pool")
	}
	return k.list(cs, opts)
}

// Get returns an CStorPool instance from kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*apis.CStorPool, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get cstor pool: missing name")
	}
	cs, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get cstor pool: {%s}", name)
	}
	return k.get(cs, name, opts)
}

// Create creates an CStorPool instance in kubernetes cluster
func (k *Kubeclient) Create(obj *apis.CStorPool) (*apis.CStorPool, error) {
	if obj == nil {
		return nil, errors.New("failed to create cstor pool: missing object")
	}
	cs, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create cstor pool %s", obj)
	}
	return k.create(cs, obj)
}

// Patch returns the patched CStorPool instance
func (k *Kubeclient) Patch(name string, pt types.PatchType,
	patchObj []byte) (*apis.CStorPool, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to patch cstor pool: missing name")
	}
	cs, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrap(err, "failed to patch cstor pool")
	}
	return k.patch(cs, name, pt, patchObj)
}

// Delete deletes CStorPool instance
func (k *Kubeclient) Delete(name string, opts *metav1.DeleteOptions) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("failed to delete cstor pool: missing name")
	}
	cs, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrap(err, "failed to delete cstor pool")
	}
	return k.del(cs, name, opts)
}
