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

	snapshotapi "github.com/openebs/maya/pkg/apis/openebs.io/snapshot/v1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/snapshot/v1/clientset/internalclientset"
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	errors "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// getClientsetFn is a typed function that abstracts
// fetching an instance of snapshot clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of snapshot clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *clientset.Clientset, err error)

// listFn is a typed function that abstracts
// listing of volumesnapshotdatas
type listFn func(cli *clientset.Clientset, opts metav1.ListOptions) (*snapshotapi.VolumeSnapshotDataList, error)

// getFn is a typed function that abstracts
// fetching an instance of volumesnapshotdata
type getFn func(cli *clientset.Clientset, name string, opts metav1.GetOptions) (*snapshotapi.VolumeSnapshotData, error)

// deleteFn is a typed function that abstracts
// to delete volumesnapshotdata
type deleteFn func(cli *clientset.Clientset, name string, opts *metav1.DeleteOptions) error

// patchFn is a typed function that abstracts
// to patch volumesnapshotdata
type patchFn func(cli *clientset.Clientset, name string, pt types.PatchType, data []byte, subresources ...string) (*snapshotapi.VolumeSnapshotData, error)

// Kubeclient enables kubernetes API operations on volumesnapshotdata instance
type Kubeclient struct {
	// clientset refers to snapshot clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	list                listFn
	get                 getFn
	del                 deleteFn
	patch               patchFn
}

// KubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

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
		k.list = func(cli *clientset.Clientset, opts metav1.ListOptions) (*snapshotapi.VolumeSnapshotDataList, error) {
			return cli.VolumesnapshotV1().VolumeSnapshotDatas().
				List(context.TODO(), opts)
		}
	}
	if k.get == nil {
		k.get = func(cli *clientset.Clientset, name string, opts metav1.GetOptions) (*snapshotapi.VolumeSnapshotData, error) {
			return cli.VolumesnapshotV1().VolumeSnapshotDatas().
				Get(context.TODO(), name, opts)
		}
	}
	if k.del == nil {
		k.del = func(cli *clientset.Clientset, name string, opts *metav1.DeleteOptions) error {
			return cli.VolumesnapshotV1().VolumeSnapshotDatas().
				Delete(context.TODO(), name, *opts)
		}
	}
	if k.patch == nil {
		k.patch = func(cli *clientset.Clientset, name string, pt types.PatchType, data []byte, subresources ...string) (*snapshotapi.VolumeSnapshotData, error) {
			return cli.VolumesnapshotV1().VolumeSnapshotDatas().
				Patch(context.TODO(), name, pt, data, metav1.PatchOptions{}, subresources...)
		}
	}
}

// NewKubeClient returns a new instance of kubeclient meant for snapshot data
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// WithClientSet sets the snapshot client against
// the kubeclient instance
func WithClientSet(c *clientset.Clientset) KubeclientBuildOption {
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

func (k *Kubeclient) getClientsetForPathOrDirect() (*clientset.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientsetOrCached returns either a new
// instance of snapshot clientset or its
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

// List returns a list of volumesnapshotdata instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*snapshotapi.VolumeSnapshotDataList, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list volumeSnapshotDatas")
	}
	return k.list(cli, opts)
}

// Get return a volumesnapshotdata instance present in kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*snapshotapi.VolumeSnapshotData, error) {
	if len(name) == 0 {
		return nil, errors.New("failed to get volumesnapshotdata: missing snapshotdata name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get VolumeSnapshotData {%s}", name)
	}
	return k.get(cli, name, opts)
}

// Delete deletes the snapshotdata if present in kubernetes cluster
func (k *Kubeclient) Delete(name string, opts *metav1.DeleteOptions) error {
	if len(name) == 0 {
		return errors.New("failed to delete volumesnapshotdata: missing snapshotdata name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete VolumeSnapshotData: {%s}", name)
	}
	return k.del(cli, name, opts)
}

// ListRaw returns volumesnapshotdata object for given name in byte format
func (k *Kubeclient) ListRaw(opts metav1.ListOptions) ([]byte, error) {
	vsdList, err := k.List(opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(vsdList)
}

// GetRaw gets volumesnapshotdata object for a given name present
// in kubernetes cluster and returns result in raw byte.
func (k *Kubeclient) GetRaw(name string, opts metav1.GetOptions) ([]byte, error) {
	vsd, err := k.Get(name, opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(vsd)
}

// Patch patches the snapshotdata if present in kubernetes cluster
func (k *Kubeclient) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*snapshotapi.VolumeSnapshotData, error) {
	if len(name) == 0 {
		return nil, errors.New("failed to patch volumesnapshotdata: missing snapshotdata name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to patch VolumeSnapshotData: {%s}", name)
	}
	return k.patch(cli, name, pt, data, subresources...)
}
