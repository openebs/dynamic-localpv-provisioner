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

package v1alpha3

import (
	"fmt"

	"github.com/pkg/errors"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	kclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// listFn is a typed function that abstracts
// listing of cstor pool
type listFn func(
	cli *clientset.Clientset,
	opts metav1.ListOptions,
) (*apis.CStorPoolList, error)

type getFn func(
	cli *clientset.Clientset,
	name string,
	opts metav1.GetOptions,
) (*apis.CStorPool, error)

// deleteFn is a typed function that abstracts
// deletion of cstor pool
type deleteFn func(
	cli *clientset.Clientset,
	name string,
	opts *metav1.DeleteOptions,
) (*apis.CStorPool, error)

// deleteCollectionFn is a typed function that abstracts
// deletion of csp's collection
type deleteCollectionFn func(
	cli *clientset.Clientset,
	listOpts metav1.ListOptions,
	deleteOpts *metav1.DeleteOptions,
) error

type patchFn func(
	cli *clientset.Clientset,
	name string,
	pt types.PatchType,
	data []byte,
) (*apis.CStorPool, error)

// Kubeclient enables kubernetes API operations
// on cstor storage pool instance
type Kubeclient struct {
	// clientset refers to cstor storage pool's
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset

	// functions useful during mocking
	getClientset  getClientsetFn
	get           getFn
	list          listFn
	del           deleteFn
	delCollection deleteCollectionFn
	patch         patchFn
}

// KubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

func defaultGetClientset() (clients *clientset.Clientset, err error) {
	config, err := kclient.New().Config()
	if err != nil {
		return nil, err
	}
	return clientset.NewForConfig(config)
}

func defaultList(
	cli *clientset.Clientset,
	opts metav1.ListOptions,
) (*apis.CStorPoolList, error) {
	return cli.OpenebsV1alpha1().CStorPools().List(opts)
}

func defaultGet(
	cli *clientset.Clientset,
	name string,
	opts metav1.GetOptions,
) (*apis.CStorPool, error) {
	return cli.OpenebsV1alpha1().CStorPools().Get(name, opts)
}

func defaultDel(
	cli *clientset.Clientset,
	name string,
	opts *metav1.DeleteOptions,
) (*apis.CStorPool, error) {
	return nil, cli.OpenebsV1alpha1().CStorPools().Delete(name, opts)
}

func defaultDelCollection(
	cs *clientset.Clientset,
	listOpts metav1.ListOptions,
	deleteOpts *metav1.DeleteOptions,
) error {
	return cs.OpenebsV1alpha1().
		CStorPools().
		DeleteCollection(deleteOpts, listOpts)
}

func defaultPatch(
	cli *clientset.Clientset,
	name string,
	pt types.PatchType,
	data []byte,
) (*apis.CStorPool, error) {
	return cli.OpenebsV1alpha1().
		CStorPools().
		Patch(name, pt, data)
}

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = defaultGetClientset
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
	if k.delCollection == nil {
		k.delCollection = defaultDelCollection
	}
	if k.patch == nil {
		k.patch = defaultPatch
	}
}

// WithKubeClient sets the kubernetes client against
// the kubeclient instance
func WithKubeClient(c *clientset.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets the client using the kubeconfig path
func (k *Kubeclient) WithKubeConfigPath(kubeconfig string) (*Kubeclient, error) {
	cfg, err := getClusterConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("Error building kubeconfig: %s", err.Error())
	}

	// Building OpenEBS Clientset
	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("Error building openebs clientset: %s", err.Error())
	}
	k.clientset = openebsClient
	return k, nil
}

func getClusterConfig(kubeconfig string) (*rest.Config, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("Error building kubeconfig: %s", err.Error())
	}
	return cfg, err
}

// KubeClient returns a new instance of kubeclient meant for
// cstor volume replica operations
func KubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*clientset.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}
	c, err := k.getClientset()
	if err != nil {
		return nil, err
	}
	k.clientset = c
	return k.clientset, nil
}

// Get returns cstor pool
// instances present in kubernetes cluster
func (k *Kubeclient) Get(
	name string,
	opts metav1.GetOptions,
) (*apis.CStorPool, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, name, opts)
}

// List returns a list of cstor pool
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*apis.CStorPoolList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, opts)
}

// Delete deletes a cstor pool
// instances present in kubernetes cluster
func (k *Kubeclient) Delete(name string, opts *metav1.DeleteOptions) (*apis.CStorPool, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.del(cli, name, opts)
}

// DeleteCollection deletes a collection of csp objects.
func (k *Kubeclient) DeleteCollection(listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to delete the collection of csps having label %s",
			listOpts.LabelSelector,
		)
	}
	return k.delCollection(cli, listOpts, deleteOpts)
}

// Patch patches a cstor pool
// instances present in kubernetes cluster
func (k *Kubeclient) Patch(
	name string,
	pt types.PatchType,
	data []byte,
) (*apis.CStorPool, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.patch(cli, name, pt, data)
}
