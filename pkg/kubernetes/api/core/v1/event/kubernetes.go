/*
Copyright 2021 The OpenEBS Authors

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

package event

import (
	"context"

	client "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/client"
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (*clientset.Clientset, error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (*clientset.Clientset, error)

// getKubeConfigFn is a typed function that
// abstracts fetching of config
type getKubeConfigFn func() (*rest.Config, error)

// getKubeConfigForPathFn is a typed function that
// abstracts fetching of config from kubeConfigPath
type getKubeConfigForPathFn func(kubeConfigPath string) (*rest.Config, error)

// listFn is a typed function that abstracts
// listing of Events
type listFn func(ctx context.Context, cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*corev1.EventList, error)

// KubeClient enables kubernetes API operations
// on Event instance
type KubeClient struct {
	// clientset refers to Event clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset

	// namespace holds the namespace on which
	// KubeClient has to operate
	namespace string

	// kubeConfig represents kubernetes config
	kubeConfig *rest.Config

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getKubeConfig        getKubeConfigFn
	getKubeConfigForPath getKubeConfigForPathFn
	getClientset         getClientsetFn
	getClientsetForPath  getClientsetForPathFn
	list                 listFn
}

// KubeClientBuildOption defines the abstraction
// to build a KubeClient instance
type KubeClientBuildOption func(*KubeClient)

// withDefaults sets the default options
// of KubeClient instance
func (k *KubeClient) withDefaults() {
	if k.getKubeConfig == nil {
		k.getKubeConfig = func() (config *rest.Config, err error) {
			return client.New().Config()
		}
	}
	if k.getKubeConfigForPath == nil {
		k.getKubeConfigForPath = func(kubeConfigPath string) (
			config *rest.Config, err error) {
			return client.New(client.WithKubeConfigPath(kubeConfigPath)).
				GetConfigForPathOrDirect()
		}
	}
	if k.getClientset == nil {
		k.getClientset = func() (clients *clientset.Clientset, err error) {
			return client.New().Clientset()
		}
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (
			clients *clientset.Clientset, err error) {
			return client.New(client.WithKubeConfigPath(kubeConfigPath)).Clientset()
		}
	}
	if k.list == nil {
		k.list = func(ctx context.Context, cli *clientset.Clientset,
			namespace string, opts metav1.ListOptions) (*corev1.EventList, error) {
			return cli.CoreV1().Events(namespace).List(ctx, opts)
		}
	}
}

// WithClientSet sets the kubernetes client against
// the KubeClient instance
func WithClientSet(c *clientset.Clientset) KubeClientBuildOption {
	return func(k *KubeClient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeClientBuildOption {
	return func(k *KubeClient) {
		k.kubeConfigPath = path
	}
}

// NewKubeClient returns a new instance of KubeClient meant for
// cstor volume replica operations
func NewKubeClient(opts ...KubeClientBuildOption) *KubeClient {
	k := &KubeClient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// WithNamespace sets the kubernetes namespace against
// the provided namespace
func (k *KubeClient) WithNamespace(namespace string) *KubeClient {
	k.namespace = namespace
	return k
}

// WithKubeConfig sets the kubernetes config against
// the KubeClient instance
func (k *KubeClient) WithKubeConfig(config *rest.Config) *KubeClient {
	k.kubeConfig = config
	return k
}

func (k *KubeClient) getClientsetForPathOrDirect() (
	*clientset.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientsetOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *KubeClient) getClientsetOrCached() (*clientset.Clientset, error) {
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

func (k *KubeClient) getKubeConfigForPathOrDirect() (*rest.Config, error) {
	if k.kubeConfigPath != "" {
		return k.getKubeConfigForPath(k.kubeConfigPath)
	}
	return k.getKubeConfig()
}

// getKubeConfigOrCached returns either a new instance
// of kubernetes config or its cached copy
func (k *KubeClient) getKubeConfigOrCached() (*rest.Config, error) {
	if k.kubeConfig != nil {
		return k.kubeConfig, nil
	}

	kc, err := k.getKubeConfigForPathOrDirect()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get kube config")
	}
	k.kubeConfig = kc
	return k.kubeConfig, nil
}

// List returns a list of Event
// instances present in kubernetes cluster
func (k *KubeClient) List(ctx context.Context, opts metav1.ListOptions) (*corev1.EventList, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list Events")
	}
	return k.list(ctx, cli, k.namespace, opts)
}
