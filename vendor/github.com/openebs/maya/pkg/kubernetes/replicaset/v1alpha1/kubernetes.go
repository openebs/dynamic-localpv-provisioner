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
	extnv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getFn is a typed function that abstracts get of replicaset instances
type getFn func(cli *kubernetes.Clientset, name, namespace string,
	opts metav1.GetOptions) (*extnv1beta1.ReplicaSet, error)

// listFn is a typed function that abstracts get of replicaset instances
type listFn func(cli *kubernetes.Clientset, namespace string,
	opts metav1.ListOptions) (*extnv1beta1.ReplicaSetList, error)

// delFn is a typed function that abstracts get of replicaset instances
type delFn func(cli *kubernetes.Clientset, name, namespace string,
	opts *metav1.DeleteOptions) error

// kubeclient enables kubernetes API operations on replicaset instance
type kubeclient struct {
	// clientset refers to kubernetes clientset. It is responsible to
	// make kubernetes API calls for crud op
	clientset *kubernetes.Clientset
	namespace string

	// functions useful during mocking
	getClientset getClientsetFn
	list         listFn
	get          getFn
	del          delFn
}

// kubeclientBuildOption defines the abstraction to build a kubeclient instance
type kubeclientBuildOption func(*kubeclient)

// withDefaults sets the default options of kubeclient instance
func (k *kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *kubernetes.Clientset, err error) {
			config, err := client.GetConfig(client.New())
			if err != nil {
				return nil, err
			}
			return kubernetes.NewForConfig(config)
		}
	}

	if k.get == nil {
		k.get = func(cli *kubernetes.Clientset, name,
			namespace string, opts metav1.GetOptions) (
			r *extnv1beta1.ReplicaSet, err error) {
			r, err = cli.ExtensionsV1beta1().
				ReplicaSets(namespace).
				Get(name, opts)
			return
		}
	}

	if k.list == nil {
		k.list = func(cli *kubernetes.Clientset,
			namespace string, opts metav1.ListOptions) (
			rl *extnv1beta1.ReplicaSetList, err error) {
			rl, err = cli.ExtensionsV1beta1().
				ReplicaSets(namespace).
				List(opts)
			return
		}
	}

	if k.del == nil {
		k.del = func(cli *kubernetes.Clientset, name,
			namespace string, opts *metav1.DeleteOptions) (err error) {
			deletePropagation := metav1.DeletePropagationForeground
			opts.PropagationPolicy = &deletePropagation
			err = cli.ExtensionsV1beta1().
				ReplicaSets(namespace).
				Delete(name, opts)
			return
		}
	}

}

// WithClientset sets the kubernetes client against the kubeclient instance
func WithClientset(c *kubernetes.Clientset) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.clientset = c
	}
}

// WithNamespace set namespace in kubeclient object
func WithNamespace(namespace string) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.namespace = namespace
	}
}

// KubeClient returns a new instance of kubeclient meant for deployment.
// caller can configure it with different kubeclientBuildOption
func KubeClient(opts ...kubeclientBuildOption) *kubeclient {
	k := &kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// getClientOrCached returns either a new
// instance of kubernetes client or its cached copy
func (k *kubeclient) getClientOrCached() (*kubernetes.Clientset, error) {
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

// Get returns deployment object for given name
func (k *kubeclient) Get(name string) (*extnv1beta1.ReplicaSet, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, name, k.namespace, metav1.GetOptions{})
}

// GetRaw returns deployment object for given name in byte format
func (k *kubeclient) GetRaw(name string) ([]byte, error) {
	rs, err := k.Get(name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(rs)
}

// List returns deployment object for given name
func (k *kubeclient) List(opts metav1.ListOptions) (*extnv1beta1.ReplicaSetList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, k.namespace, opts)
}

// ListRaw returns deployment object for given name in byte format
func (k *kubeclient) ListRaw(opts metav1.ListOptions) ([]byte, error) {
	rsList, err := k.List(opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(rsList)
}

// Delete returns deployment object for given name
func (k *kubeclient) Delete(name string) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.del(cli, name, k.namespace, &metav1.DeleteOptions{})
}
