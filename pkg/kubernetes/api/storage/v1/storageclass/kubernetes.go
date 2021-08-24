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

package storageclass

import (
	"context"
	"strings"

	errors "github.com/pkg/errors"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/client"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *kubernetes.Clientset, err error)

// getFn is a typed function that
// abstracts fetching of StorageClass
type getFn func(ctx context.Context, cli *kubernetes.Clientset, name string, opts metav1.GetOptions) (*storagev1.StorageClass, error)

// listFn is a typed function that abstracts
// listing of StorageClass
type listFn func(ctx context.Context, cli *kubernetes.Clientset, opts metav1.ListOptions) (*storagev1.StorageClassList, error)

// deleteFn is a typed function that abstracts
// deletion of StorageClass
type deleteFn func(ctx context.Context, cli *kubernetes.Clientset, name string, deleteOpts *metav1.DeleteOptions) error

// deleteFn is a typed function that abstracts
// deletion of StorageClass's collection
type deleteCollectionFn func(ctx context.Context, cli *kubernetes.Clientset, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error

// createFn is a typed function that abstracts
// creation of StorageClass
type createFn func(ctx context.Context, cli *kubernetes.Clientset, sc *storagev1.StorageClass) (*storagev1.StorageClass, error)

// updateFn is a typed function that abstracts
// updation of StorageClass
type updateFn func(ctx context.Context, cli *kubernetes.Clientset, sc *storagev1.StorageClass) (*storagev1.StorageClass, error)

// Kubeclient enables kubernetes API operations
// on StorageClass instance
type Kubeclient struct {
	// clientset will be responsible for
	// making kubernetes API calls
	clientset *kubernetes.Clientset

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	list                listFn
	get                 getFn
	create              createFn
	update              updateFn
	del                 deleteFn
	delCollection       deleteCollectionFn
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
		k.get = func(ctx context.Context, cli *kubernetes.Clientset, name string, opts metav1.GetOptions) (*storagev1.StorageClass, error) {
			return cli.StorageV1().StorageClasses().Get(ctx, name, opts)
		}
	}

	if k.list == nil {
		k.list = func(ctx context.Context, cli *kubernetes.Clientset, opts metav1.ListOptions) (*storagev1.StorageClassList, error) {
			return cli.StorageV1().StorageClasses().List(ctx, opts)
		}
	}

	if k.del == nil {
		k.del = func(ctx context.Context, cli *kubernetes.Clientset, name string, deleteOpts *metav1.DeleteOptions) error {
			return cli.StorageV1().StorageClasses().Delete(ctx, name, *deleteOpts)
		}
	}

	if k.delCollection == nil {
		k.delCollection = func(ctx context.Context, cli *kubernetes.Clientset, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
			return cli.StorageV1().StorageClasses().DeleteCollection(ctx, *deleteOpts, listOpts)
		}
	}

	if k.create == nil {
		k.create = func(ctx context.Context, cli *kubernetes.Clientset, sc *storagev1.StorageClass) (*storagev1.StorageClass, error) {
			return cli.StorageV1().StorageClasses().Create(ctx, sc, metav1.CreateOptions{})
		}
	}

	if k.update == nil {
		k.update = func(ctx context.Context, cli *kubernetes.Clientset, sc *storagev1.StorageClass) (*storagev1.StorageClass, error) {
			return cli.StorageV1().StorageClasses().Update(ctx, sc, metav1.UpdateOptions{})
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

// NewKubeClient returns a new instance of kubeclient
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

// Get returns a StorageClass resource
// instances present in kubernetes cluster
func (k *Kubeclient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*storagev1.StorageClass, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get StorageClass: missing StorageClass name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get StorageClass {%s}", name)
	}
	return k.get(ctx, cli, name, opts)
}

// List returns a list of StorageClasses
// instances present in kubernetes cluster
func (k *Kubeclient) List(ctx context.Context, opts metav1.ListOptions) (*storagev1.StorageClassList, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list StorageClass listoptions: '%v'", opts)
	}
	return k.list(ctx, cli, opts)
}

// Delete deletes a StorageClass instance from the
// kubernetes cluster
func (k *Kubeclient) Delete(ctx context.Context, name string, deleteOpts *metav1.DeleteOptions) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("failed to delete StorageClass: missing StorageClass name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete StorageClass {%s}", name)
	}
	return k.del(ctx, cli, name, deleteOpts)
}

// Create creates a StorageClass in kubernetes cluster
func (k *Kubeclient) Create(ctx context.Context, sc *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	if sc == nil {
		return nil, errors.New("failed to create StorageClass: nil StorageClass object")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create StorageClass {%s}", sc.Name)
	}
	return k.create(ctx, cli, sc)
}

// Update updates a StorageClass in kubernetes cluster
func (k *Kubeclient) Update(ctx context.Context, sc *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	if sc == nil {
		return nil, errors.New("failed to update StorageClass: nil StorageClass object")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update StorageClass {%s}", sc.Name)
	}
	return k.update(ctx, cli, sc)
}

// CreateCollection creates a list of StorageClasses
// in kubernetes cluster
func (k *Kubeclient) CreateCollection(
	ctx context.Context,
	list *storagev1.StorageClassList,
) (*storagev1.StorageClassList, error) {
	if list == nil || len(list.Items) == 0 {
		return nil, errors.New("failed to create list of StorageClasses: nil StorageClass list provided")
	}

	newlist := &storagev1.StorageClassList{}
	for _, item := range list.Items {
		item := item
		obj, err := k.Create(ctx, &item)
		if err != nil {
			return nil, err
		}

		newlist.Items = append(newlist.Items, *obj)
	}

	return newlist, nil
}

// DeleteCollection deletes a collection of StorageClass objects.
func (k *Kubeclient) DeleteCollection(ctx context.Context, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete the collection of StorageClasses")
	}
	return k.delCollection(ctx, cli, listOpts, deleteOpts)
}
