// Copyright © 2018-2020 The OpenEBS Authors
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

package pod

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func fakeGetClientSetOk() (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeListFnOk(ctx context.Context, cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
	return &corev1.PodList{}, nil
}

func fakeListFnErr(ctx context.Context, cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
	return &corev1.PodList{}, errors.New("some error")
}

func fakeDeleteFnOk(ctx context.Context, cli *clientset.Clientset, namespace, name string, opts *metav1.DeleteOptions) error {
	return nil
}

func fakeDeleteFnErr(ctx context.Context, cli *clientset.Clientset, namespace, name string, opts *metav1.DeleteOptions) error {
	return errors.New("some error while delete")
}

func fakeGetFnOk(ctx context.Context, cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*corev1.Pod, error) {
	return &corev1.Pod{}, nil
}

func fakeGetErrfn(ctx context.Context, cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*corev1.Pod, error) {
	return &corev1.Pod{}, errors.New("Not found")
}

func fakeSetClientset(k *KubeClient) {
	k.clientset = &clientset.Clientset{}
}

func fakeSetNilClientset(k *KubeClient) {
	k.clientset = nil
}

func fakeGetClientSetNil() (clientset *clientset.Clientset, err error) {
	return nil, nil
}

func fakeGetClientSetErr() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeClientSet(k *KubeClient) {}

func fakeGetClientSetForPathOk(fakeConfigPath string) (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeGetClientSetForPathErr(fakeConfigPath string) (cli *clientset.Clientset, err error) {
	return nil, errors.New("fake error")
}

func fakeDeleteCollectionOk(ctx context.Context, cli *clientset.Clientset, namespace string, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
	return nil
}

func fakeDeleteCollectionErr(ctx context.Context, cli *clientset.Clientset, namespace string, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
	return errors.New("fake error")
}

func TestWithDefaultOptions(t *testing.T) {
	tests := map[string]struct {
		kubeClient *KubeClient
	}{
		"T1": {&KubeClient{}},
		"T2": {&KubeClient{
			clientset:    nil,
			getClientset: fakeGetClientSetOk,
			list:         fakeListFnOk,
			get:          fakeGetFnOk,
			del:          fakeDeleteFnOk,
		}},
		"T3": {&KubeClient{
			getClientset: fakeGetClientSetOk,
			list:         nil,
			get:          fakeGetFnOk,
			del:          fakeDeleteFnOk,
		}},
		"T4": {&KubeClient{
			getClientset: nil,
			list:         fakeListFnOk,
			get:          fakeGetFnOk,
			del:          fakeDeleteFnOk,
		}},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			mock.kubeClient.withDefaults()
			if mock.kubeClient.get == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
			}
			if mock.kubeClient.list == nil {
				t.Fatalf("test %q failed: expected list not to be empty", name)
			}
			if mock.kubeClient.del == nil {
				t.Fatalf("test %q failed: expected delete not to be empty", name)
			}
			if mock.kubeClient.getClientset == nil {
				t.Fatalf("test %q failed: expected get clientset not to be empty", name)
			}
		})
	}
}

func TestWithDefaultsForClientSetPath(t *testing.T) {
	tests := map[string]struct {
		getClientSetForPath getClientsetForPathFn
	}{
		"T1": {nil},
		"T2": {fakeGetClientSetForPathOk},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &KubeClient{
				getClientsetForPath: mock.getClientSetForPath,
			}
			fc.withDefaults()
			if fc.getClientsetForPath == nil {
				t.Fatalf("test %q failed: expected getClientsetForPath not to be nil", name)
			}
		})
	}
}

func TestGetClientSetForPathOrDirect(t *testing.T) {
	tests := map[string]struct {
		getClientSet        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		isErr               bool
	}{
		// Positive tests
		"Positive 1": {fakeGetClientSetNil, fakeGetClientSetForPathOk, "fake-path", false},
		"Positive 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", false},
		"Positive 3": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "fake-path", false},
		"Positive 4": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "", false},

		// Negative tests
		"Negative 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", true},
		"Negative 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", true},
		"Negative 3": {fakeGetClientSetErr, fakeGetClientSetForPathErr, "fake-path", true},
		"Negative 4": {fakeGetClientSetErr, fakeGetClientSetForPathErr, "", true},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &KubeClient{
				getClientset:        mock.getClientSet,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
			}
			_, err := fc.getClientsetForPathOrDirect()
			if mock.isErr && err == nil {
				t.Fatalf("test %q failed : expected error not to be nil but got %v", name, err)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test %q failed : expected error be nil but got %v", name, err)
			}
		})
	}
}

func TestWithClientsetBuildOption(t *testing.T) {
	tests := map[string]struct {
		Clientset             *clientset.Clientset
		expectKubeClientEmpty bool
	}{
		"Clientset is empty":     {nil, true},
		"Clientset is not empty": {&clientset.Clientset{}, false},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			h := WithClientSet(mock.Clientset)
			fake := &KubeClient{}
			h(fake)
			if mock.expectKubeClientEmpty && fake.clientset != nil {
				t.Fatalf("test %q failed expected fake.clientset to be empty", name)
			}
			if !mock.expectKubeClientEmpty && fake.clientset == nil {
				t.Fatalf("test %q failed expected fake.clientset not to be empty", name)
			}
		})
	}
}

func TestKubeClientBuildOption(t *testing.T) {
	tests := map[string]struct {
		opts            []KubeClientBuildOption
		expectClientSet bool
	}{
		"Positive 1": {[]KubeClientBuildOption{fakeSetClientset, WithKubeConfigPath("fake-path")}, true},
		"Positive 2": {[]KubeClientBuildOption{fakeSetClientset, fakeClientSet}, true},
		"Positive 3": {[]KubeClientBuildOption{fakeSetClientset, fakeClientSet, WithKubeConfigPath("fake-path")}, true},

		"Negative 1": {[]KubeClientBuildOption{fakeSetNilClientset, WithKubeConfigPath("fake-path")}, false},
		"Negative 2": {[]KubeClientBuildOption{fakeSetNilClientset, fakeClientSet}, false},
		"Negative 3": {[]KubeClientBuildOption{fakeSetNilClientset, fakeClientSet, WithKubeConfigPath("fake-path")}, false},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			c := NewKubeClient(mock.opts...)
			if !mock.expectClientSet && c.clientset != nil {
				t.Fatalf("test %q failed expected fake.clientset to be empty", name)
			}
			if mock.expectClientSet && c.clientset == nil {
				t.Fatalf("test %q failed expected fake.clientset not to be empty", name)
			}
		})
	}
}

func TestGetClientOrCached(t *testing.T) {
	tests := map[string]struct {
		getClientSet        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		expectErr           bool
	}{
		// Positive tests
		"Positive 1": {fakeGetClientSetNil, fakeGetClientSetForPathOk, "fake-path", false},
		"Positive 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", false},
		"Positive 3": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "fake-path", false},
		"Positive 4": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "", false},

		// Negative tests
		"Negative 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", true},
		"Negative 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", true},
		"Negative 3": {fakeGetClientSetErr, fakeGetClientSetForPathErr, "fake-path", true},
		"Negative 4": {fakeGetClientSetErr, fakeGetClientSetForPathErr, "", true},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &KubeClient{
				getClientset:        mock.getClientSet,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
			}
			_, err := fc.getClientsetOrCached()
			if mock.expectErr && err == nil {
				t.Fatalf("test %q failed : expected error not to be nil but got %v", name, err)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("test %q failed : expected error be nil but got %v", name, err)
			}
		})
	}
}

func TestKubernetesPodList(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		list                listFn
		expectErr           bool
	}{
		// Positive tests
		"Positive 1": {fakeGetClientSetNil, fakeGetClientSetForPathOk, "fake-path", fakeListFnOk, false},
		"Positive 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeListFnOk, false},
		"Positive 3": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "fake-path", fakeListFnOk, false},
		"Positive 4": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "", fakeListFnOk, false},

		// Negative tests
		"Negative 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", fakeListFnOk, true},
		"Negative 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", fakeListFnOk, true},
		"Negative 3": {fakeGetClientSetErr, fakeGetClientSetForPathErr, "fake-path", fakeListFnOk, true},
		"Negative 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeListFnErr, true},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &KubeClient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				list:                mock.list,
			}
			_, err := fc.List(context.TODO(), metav1.ListOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubernetesDeletePod(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		podName             string
		delete              deleteFn
		expectErr           bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", "pod-1", fakeDeleteFnOk, true},
		"Test 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fake-path2", "pod-2", fakeDeleteFnOk, false},
		"Test 3": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", "pod-3", fakeDeleteFnErr, true},
		"Test 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fakepath", "", fakeDeleteFnOk, true},
		"Test 5": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path2", "pod1", fakeDeleteFnOk, true},
		"Test 6": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path2", "pod1", fakeDeleteFnErr, true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &KubeClient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				namespace:           "",
				del:                 mock.delete,
			}
			err := k.Delete(context.TODO(), mock.podName, &metav1.DeleteOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubernetesDeleteCollection(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		listOpts            string
		deleteCollection    deleteCollectionFn
		expectErr           bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", "selector=selector1", fakeDeleteCollectionOk, true},
		"Test 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fake-path2", "selector=selector1", fakeDeleteCollectionOk, false},
		"Test 3": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", "selector=selector1", fakeDeleteCollectionErr, true},
		"Test 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fakepath", "selector=selector1", fakeDeleteCollectionErr, true},
		"Test 5": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path2", "selector=selector1", fakeDeleteCollectionOk, true},
		"Test 6": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path2", "selector=selector1", fakeDeleteCollectionErr, true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &KubeClient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				namespace:           "",
				delCollection:       mock.deleteCollection,
			}
			err := k.DeleteCollection(
				context.TODO(),
				metav1.ListOptions{LabelSelector: mock.listOpts},
				&metav1.DeleteOptions{},
			)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubernetesGetPod(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		get                 getFn
		podName             string
		expectErr           bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", fakeGetFnOk, "pod-1", true},
		"Test 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", fakeGetFnOk, "pod-1", true},
		"Test 3": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeGetFnOk, "pod-2", false},
		"Test 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fp", fakeGetErrfn, "pod-3", true},
		"Test 5": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fakepath", fakeGetFnOk, "", true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &KubeClient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				namespace:           "",
				get:                 mock.get,
			}
			_, err := k.Get(context.TODO(), mock.podName, metav1.GetOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestWithBuildOption(t *testing.T) {
	tests := map[string]struct {
		namespace      string
		kubeConfigPath string
	}{
		"Test 1": {"", ""},
		"Test 2": {"namespace 1", ""},
		"Test 3": {"namespace 2", "fake-path"},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := NewKubeClient(WithKubeConfigPath(mock.kubeConfigPath)).WithNamespace(mock.namespace)
			if k.namespace != mock.namespace {
				t.Fatalf("Test %q failed: expected %v got %v", name, mock.namespace, k.namespace)
			}
			if k.kubeConfigPath != mock.kubeConfigPath {
				t.Fatalf("Test %q failed: expected %v got %v", name, mock.namespace, k.namespace)
			}
		})
	}
}
