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
	"context"
	"encoding/json"

	snapshot "github.com/openebs/maya/pkg/apis/openebs.io/snapshot/v1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/snapshot/v1/clientset/internalclientset"
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	errors "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getClientsetFn is a typed function that abstracts
// fetching an instance of kubernetes clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *clientset.Clientset, err error)

// listFn is a typed function that abstracts
// listing of snapshots
type listFn func(cli *clientset.Clientset, opts metav1.ListOptions) (*snapshot.VolumeSnapshotList, error)

// getFn is a typed function that abstracts
// fetching an instance of snapshot
type getFn func(cli *clientset.Clientset, name string, opts metav1.GetOptions) (*snapshot.VolumeSnapshot, error)

// createFn is a typed function that abstracts
// to create snapshot
type createFn func(cli *clientset.Clientset, snap *snapshot.VolumeSnapshot) (*snapshot.VolumeSnapshot, error)

// deleteFn is a typed function that abstracts
// to delete snapshot
type deleteFn func(cli *clientset.Clientset, name string, opts *metav1.DeleteOptions) error

// Kubeclient enables kubernetes API operations on snapshot instance
type Kubeclient struct {
	// clientset refers to snapshot clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset

	namespace string
	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	list                listFn
	get                 getFn
	create              createFn
	del                 deleteFn
}

// KubeClientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeClientBuildOption func(*Kubeclient)

func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *clientset.Clientset, err error) {
			config, err := client.New().GetConfigForPathOrDirect()
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (clients *clientset.Clientset, err error) {
			config, err := client.New(client.WithKubeConfigPath(kubeConfigPath)).GetConfigForPathOrDirect()
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}
	if k.list == nil {
		k.list = func(cli *clientset.Clientset, opts metav1.ListOptions) (*snapshot.VolumeSnapshotList, error) {
			return cli.VolumesnapshotV1().VolumeSnapshots(k.namespace).
				List(context.TODO(), opts)
		}
	}
	if k.get == nil {
		k.get = func(cli *clientset.Clientset, name string, opts metav1.GetOptions) (*snapshot.VolumeSnapshot, error) {
			return cli.VolumesnapshotV1().VolumeSnapshots(k.namespace).
				Get(context.TODO(), name, opts)
		}
	}
	if k.create == nil {
		k.create = func(cli *clientset.Clientset, snap *snapshot.VolumeSnapshot) (*snapshot.VolumeSnapshot, error) {
			return cli.VolumesnapshotV1().VolumeSnapshots(k.namespace).
				Create(context.TODO(), snap, metav1.CreateOptions{})
		}
	}
	if k.del == nil {
		k.del = func(cli *clientset.Clientset, name string, opts *metav1.DeleteOptions) error {
			return cli.VolumesnapshotV1().VolumeSnapshots(k.namespace).
				Delete(context.TODO(), name, *opts)
		}
	}
}

// NewKubeClient returns a new instance of kubeclient meant for snapshot
func NewKubeClient(opts ...KubeClientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// WithClientSet sets the kubernetes client against
// the kubeclient instance
func WithClientSet(c *clientset.Clientset) KubeClientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeClientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*clientset.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientsetOrCached returns either a new
// instance of kubernetes clientset or its
// cached copy cached copy
func (k *Kubeclient) getClientsetOrCached() (*clientset.Clientset, error) {
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

// List returns a list of snapshot instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*snapshot.VolumeSnapshotList, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list snapshots")
	}
	return k.list(cli, opts)
}

// Get return a snapshot instance present in kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*snapshot.VolumeSnapshot, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get snapshot {%s}", name)
	}
	return k.get(cli, name, opts)
}

// Create creates and returns a snapshot instance
func (k *Kubeclient) Create(snap *snapshot.VolumeSnapshot) (*snapshot.VolumeSnapshot, error) {
	if snap == nil {
		return nil, errors.New("failed to create snapshot: nil snapshot object")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create snapshot")
	}
	return k.create(cli, snap)
}

// Delete deletes the snapshot if present in kubernetes cluster
func (k *Kubeclient) Delete(name string, opts *metav1.DeleteOptions) error {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete snapshot: {%s}", name)
	}
	return k.del(cli, name, opts)
}

// WithNamespace sets the kubernetes client against
// the provided namespace
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

// ListRaw returns volumesnapshot object for given name in byte format
func (k *Kubeclient) ListRaw(opts metav1.ListOptions) ([]byte, error) {
	vsList, err := k.List(opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(vsList)
}
