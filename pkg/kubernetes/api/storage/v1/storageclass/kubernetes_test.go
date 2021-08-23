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
	"testing"

	errors "github.com/pkg/errors"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func fakeGetClientSetOk() (cli *kubernetes.Clientset, err error) {
	return &kubernetes.Clientset{}, nil
}

func fakeGetClientSetForPathOk(fakeConfigPath string) (cli *kubernetes.Clientset, err error) {
	return &kubernetes.Clientset{}, nil
}

func fakeGetClientSetForPathErr(fakeConfigPath string) (cli *kubernetes.Clientset, err error) {
	return nil, errors.New("fake error")
}

func fakeGetOk(ctx context.Context, cli *kubernetes.Clientset, name string, opts metav1.GetOptions) (*storagev1.StorageClass, error) {
	return &storagev1.StorageClass{}, nil
}

func fakeListOk(ctx context.Context, cli *kubernetes.Clientset, opts metav1.ListOptions) (*storagev1.StorageClassList, error) {
	return &storagev1.StorageClassList{}, nil
}

func fakeDeleteOk(ctx context.Context, cli *kubernetes.Clientset, name string, opts *metav1.DeleteOptions) error {
	return nil
}

func fakeListErr(ctx context.Context, cli *kubernetes.Clientset, opts metav1.ListOptions) (*storagev1.StorageClassList, error) {
	return &storagev1.StorageClassList{}, errors.New("some error")
}

func fakeGetErr(ctx context.Context, cli *kubernetes.Clientset, name string, opts metav1.GetOptions) (*storagev1.StorageClass, error) {
	return &storagev1.StorageClass{}, errors.New("some error")
}

func fakeDeleteErr(ctx context.Context, cli *kubernetes.Clientset, name string, opts *metav1.DeleteOptions) error {
	return errors.New("some error")
}

func fakeSetClientset(k *Kubeclient) {
	k.clientset = &kubernetes.Clientset{}
}

func fakeSetKubeConfigPath(k *Kubeclient) {
	k.kubeConfigPath = "fake-path"
}

func fakeSetNilClientset(k *Kubeclient) {
	k.clientset = nil
}

func fakeGetClientSetNil() (clientset *kubernetes.Clientset, err error) {
	return nil, nil
}

func fakeGetClientSetErr() (clientset *kubernetes.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeClientSet(k *Kubeclient) {}

func fakeCreateFnOk(ctx context.Context, cli *kubernetes.Clientset, sc *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	return &storagev1.StorageClass{}, nil
}

func fakeCreateFnErr(ctx context.Context, cli *kubernetes.Clientset, sc *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	return nil, errors.New("failed to create StorageClass")
}

func TestWithDefaultOptions(t *testing.T) {
	tests := map[string]struct {
		KubeClient *Kubeclient
	}{
		"When all are nil": {&Kubeclient{}},
		"When clientset is nil": {&Kubeclient{
			clientset:           nil,
			getClientset:        fakeGetClientSetOk,
			getClientsetForPath: fakeGetClientSetForPathOk,
			list:                fakeListOk,
			get:                 fakeGetOk,
			create:              fakeCreateFnOk,
			del:                 fakeDeleteOk,
			delCollection:       nil,
		}},
		"When listFn nil": {&Kubeclient{
			getClientset:        fakeGetClientSetOk,
			list:                nil,
			getClientsetForPath: fakeGetClientSetForPathErr,
			get:                 fakeGetOk,
			create:              fakeCreateFnOk,
			del:                 fakeDeleteOk,
			delCollection:       nil,
		}},
		"When getClientsetFn nil": {&Kubeclient{
			getClientset:        nil,
			list:                fakeListOk,
			get:                 fakeGetOk,
			getClientsetForPath: fakeGetClientSetForPathOk,
			create:              fakeCreateFnOk,
			del:                 fakeDeleteOk,
			delCollection:       nil,
		}},
		"When getFn and CreateFn are nil": {&Kubeclient{
			getClientset:        fakeGetClientSetOk,
			list:                fakeListOk,
			getClientsetForPath: fakeGetClientSetForPathErr,
			get:                 nil,
			create:              nil,
			del:                 fakeDeleteOk,
			delCollection:       nil,
		}},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			mock.KubeClient.withDefaults()
			if mock.KubeClient.getClientset == nil {
				t.Fatalf("test %q failed: expected getClientset not to be empty", name)
			}
			if mock.KubeClient.getClientsetForPath == nil {
				t.Fatalf("test %q failed: expected getClientset not to be nil", name)
			}
			if mock.KubeClient.list == nil {
				t.Fatalf("test %q failed: expected list not to be empty", name)
			}
			if mock.KubeClient.get == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
			}
			if mock.KubeClient.create == nil {
				t.Fatalf("test %q failed: expected create not to be empty", name)
			}
			if mock.KubeClient.del == nil {
				t.Fatalf("test %q failed: expected del not to be empty", name)
			}
			if mock.KubeClient.delCollection == nil {
				t.Fatalf("test %q failed: expected delCollection not to be empty", name)
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
			fc := &Kubeclient{
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
		Clientset             *kubernetes.Clientset
		expectKubeClientEmpty bool
	}{
		"Clientset is empty":     {nil, true},
		"Clientset is not empty": {&kubernetes.Clientset{}, false},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			h := WithClientSet(mock.Clientset)
			fake := &Kubeclient{}
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
		opts                   []KubeclientBuildOption
		expectClientSet        bool
		expectedKubeConfigPath bool
	}{
		"Positive 1": {[]KubeclientBuildOption{fakeSetClientset, fakeSetKubeConfigPath}, true, true},
		"Positive 2": {[]KubeclientBuildOption{fakeSetClientset, fakeClientSet}, true, false},
		"Positive 3": {[]KubeclientBuildOption{fakeSetClientset, fakeClientSet, fakeClientSet}, true, false},

		"Negative 1": {[]KubeclientBuildOption{fakeSetNilClientset, fakeSetKubeConfigPath}, false, true},
		"Negative 2": {[]KubeclientBuildOption{fakeSetNilClientset, fakeClientSet, fakeSetKubeConfigPath}, false, true},
		"Negative 3": {[]KubeclientBuildOption{fakeSetNilClientset, fakeClientSet, fakeClientSet}, false, false},
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
			if mock.expectedKubeConfigPath && c.kubeConfigPath == "" {
				t.Fatalf("test %q failed expected kubeConfigPath not to be empty", name)
			}
			if !mock.expectedKubeConfigPath && c.kubeConfigPath != "" {
				t.Fatalf("test %q failed expected kubeConfigPath to be empty", name)
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
			fc := &Kubeclient{
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

func TestKubernetesSCList(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		list                listFn
		expectedErr         bool
	}{
		// Positive tests
		"Positive 1": {fakeGetClientSetNil, fakeGetClientSetForPathOk, "fake-path", fakeListOk, false},
		"Positive 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeListOk, false},
		"Positive 3": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "fake-path", fakeListOk, false},
		"Positive 4": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "", fakeListOk, false},

		// Negative tests
		"Negative 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", fakeListOk, true},
		"Negative 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", fakeListOk, true},
		"Negative 3": {fakeGetClientSetErr, fakeGetClientSetForPathErr, "fake-path", fakeListOk, true},
		"Negative 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeListErr, true},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				list:                mock.list,
			}
			_, err := fc.List(context.TODO(), metav1.ListOptions{})
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubenetesGetSC(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		get                 getFn
		podName             string
		expectErr           bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", fakeGetOk, "pod-1", true},
		"Test 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", fakeGetOk, "pod-1", true},
		"Test 3": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeGetOk, "pod-2", false},
		"Test 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fp", fakeGetErr, "pod-3", true},
		"Test 5": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fakepath", fakeGetOk, "", true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
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

func TestKubernetesDeletePVC(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		podName             string
		delete              deleteFn
		expectErr           bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", "pod-1", fakeDeleteOk, true},
		"Test 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fake-path2", "pod-2", fakeDeleteOk, false},
		"Test 3": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", "pod-3", fakeDeleteErr, true},
		"Test 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fakepath", "", fakeDeleteOk, true},
		"Test 5": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path2", "pod1", fakeDeleteOk, true},
		"Test 6": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path2", "pod1", fakeDeleteErr, true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
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

func TestKubernetesSCCreate(t *testing.T) {
	tests := map[string]struct {
		getClientSet        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		create              createFn
		sc                  *storagev1.StorageClass
		expectErr           bool
	}{
		"Test 1": {
			getClientSet:        fakeGetClientSetErr,
			getClientSetForPath: fakeGetClientSetForPathErr,
			kubeConfigPath:      "",
			create:              fakeCreateFnOk,
			sc:                  &storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "SC-1"}},
			expectErr:           true,
		},
		"Test 2": {
			getClientSet:        fakeGetClientSetOk,
			getClientSetForPath: fakeGetClientSetForPathOk,
			kubeConfigPath:      "",
			create:              fakeCreateFnErr,
			sc:                  &storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "SC-2"}},
			expectErr:           true,
		},
		"Test 3": {
			getClientSet:        fakeGetClientSetOk,
			getClientSetForPath: fakeGetClientSetForPathOk,
			kubeConfigPath:      "fake-path",
			create:              fakeCreateFnErr,
			sc:                  nil,
			expectErr:           true,
		},
		"Test 4": {
			getClientSet:        fakeGetClientSetErr,
			getClientSetForPath: fakeGetClientSetForPathOk,
			kubeConfigPath:      "fake-path",
			create:              fakeCreateFnOk,
			sc:                  nil,
			expectErr:           true,
		},
		"Test 5": {
			getClientSet:        fakeGetClientSetOk,
			getClientSetForPath: fakeGetClientSetForPathErr,
			kubeConfigPath:      "fake-path",
			create:              fakeCreateFnOk,
			sc:                  nil,
			expectErr:           true,
		},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientSet,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				create:              mock.create,
			}
			_, err := fc.Create(context.TODO(), mock.sc)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
