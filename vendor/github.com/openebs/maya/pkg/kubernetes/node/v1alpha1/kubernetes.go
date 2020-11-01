// Copyright © 2018-2019 The OpenEBS Authors
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
	"encoding/json"
	"strings"
	"sync"

	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/klog"

	"k8s.io/client-go/kubernetes"
)

const (
	labelKeyHostName string = "kubernetes.io/hostname"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *kubernetes.Clientset, err error)

// getFn is a typed function that abstracts
// get of node
type getFn func(cli *kubernetes.Clientset, name string, opts metav1.GetOptions) (*corev1.Node, error)

// listFn is a typed function that abstracts
// listing of nodes
type listFn func(cli *kubernetes.Clientset, opts metav1.ListOptions) (*corev1.NodeList, error)

// updateFn is a typed function that abstracts
// updating of node
type updateFn func(cli *kubernetes.Clientset, node *corev1.Node) (*corev1.Node, error)

// patchFn is a typed function that abstracts
// patch of a node
type patchFn func(cli *kubernetes.Clientset, name string, pt types.PatchType,
	data []byte, subresources ...string) (*corev1.Node, error)

var (
	kubeClientInst *Kubeclient
	once           sync.Once
)

// Kubeclient enables kubernetes API operations on node instance
type Kubeclient struct {
	// clientset refers to node clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *kubernetes.Clientset

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	get                 getFn
	list                listFn
	patch               patchFn
	update              updateFn
}

// KubeClientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeClientBuildOption func(*Kubeclient)

func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *kubernetes.Clientset, err error) {
			return client.New().Clientset()
		}
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(
			kubeConfigPath string) (clients *kubernetes.Clientset, err error) {
			return client.New(client.WithKubeConfigPath(kubeConfigPath)).Clientset()
		}
	}
	if k.get == nil {
		k.get = func(cli *kubernetes.Clientset, name string, opts metav1.GetOptions) (*corev1.Node, error) {
			return cli.CoreV1().Nodes().Get(name, opts)
		}
	}
	if k.list == nil {
		k.list = func(cli *kubernetes.Clientset,
			opts metav1.ListOptions) (*corev1.NodeList, error) {
			return cli.CoreV1().Nodes().List(opts)
		}
	}
	if k.patch == nil {
		k.patch = func(cli *kubernetes.Clientset, name string, pt types.PatchType, data []byte, subresources ...string) (*corev1.Node, error) {
			return cli.CoreV1().Nodes().Patch(name, pt, data, subresources...)
		}
	}
	if k.update == nil {
		k.update = func(cli *kubernetes.Clientset, node *corev1.Node) (*corev1.Node, error) {
			return cli.CoreV1().Nodes().Update(node)
		}
	}

}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeClientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// NewKubeClient returns a new instance of kubeclient meant for node
func NewKubeClient(opts ...KubeClientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// KubeClientInstanceOrDie returns the singleton instance of Kubeclient
//
// Usage:
//  Caller code will use syntax(-es) as shown below:
//
// ```go
// import (
//  node "github.com/openebs/maya/pkg/kubernetes/node/v1alpha1"
// )
//
// node.KubeClientInstanceOrDie().Get(...)
// node.KubeClientInstanceOrDie().Create(...)
// node.KubeClientInstanceOrDie().Update(...)
// node.KubeClientInstanceOrDie().List(...)
// ```
func KubeClientInstanceOrDie() *Kubeclient {
	once.Do(func() {
		k := NewKubeClient()
		_, err := k.getClientsetOrCached()
		if err != nil {
			klog.Fatalf("failed to initialise kubeclient instance: {%v}", err)
		}
		kubeClientInst = k
	})

	return kubeClientInst
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

// Get returns a node resource
// instances present in kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*corev1.Node, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get node: missing node name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get node {%s}", name)
	}
	return k.get(cli, name, opts)
}

// GetHostName returns a hostname corresponding
// to the provided node name
func (k *Kubeclient) GetHostName(name string, opts metav1.GetOptions) (string, error) {
	n, err := k.Get(name, opts)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get hostname for node {%s}", name)
	}

	if len(n.Labels) == 0 {
		return "",
			errors.Errorf("failed to get hostname for node {%s}: missing labels", name)
	}

	h := n.Labels[labelKeyHostName]
	if h == "" {
		return "",
			errors.Errorf(
				"failed to get hostname for node {%s}: missing label {%s}",
				name,
				labelKeyHostName,
			)
	}
	return h, nil
}

// GetHostNameOrNodeName returns a hostname corresponding
// to the provided node name or returns the node name if
// hostname is not available
func (k *Kubeclient) GetHostNameOrNodeName(
	name string,
	opts metav1.GetOptions,
) (string, error) {
	n, err := k.Get(name, opts)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get hostname for node {%s}", name)
	}

	if len(n.Labels) == 0 || n.Labels[labelKeyHostName] == "" {
		return name, nil
	}

	return n.Labels[labelKeyHostName], nil
}

// List returns a list of nodes instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*corev1.NodeList, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list")
	}
	return k.list(cli, opts)
}

// Patch patches the Node in kubernetes cluster
func (k *Kubeclient) Patch(name string,
	pt types.PatchType,
	data []byte,
	subresources ...string) (*corev1.Node, error) {
	if len(name) == 0 {
		return nil, errors.New("failed to patch node: missing node name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to patch node {%s}", name)
	}
	return k.patch(cli, name, pt, data, subresources...)
}

// Update updates the Node in kubernetes cluster
func (k *Kubeclient) Update(node *corev1.Node) (*corev1.Node, error) {
	if len(node.Name) == 0 {
		return nil, errors.New("failed to patch node: missing node name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to patch node {%s}", node.Name)
	}
	return k.update(cli, node)
}

// CordonViaUpdate cordon/uncordon node in cluster using update
func (k *Kubeclient) CordonViaUpdate(node *corev1.Node, isCordon bool) error {
	node.Spec.Unschedulable = isCordon

	_, err := k.Update(node)
	return err
}

// CordonViaPatch cordon/uncordon node in cluster using patch
func (k *Kubeclient) CordonViaPatch(node *corev1.Node, isCordon bool) error {
	oldData, err := json.Marshal(node)
	if err != nil {
		return err
	}

	node.Spec.Unschedulable = isCordon

	newData, err := json.Marshal(node)
	if err != nil {
		return err
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, node)
	if err == nil {
		_, err = k.Patch(node.Name, types.StrategicMergePatchType, patchBytes)
	}
	return err
}
