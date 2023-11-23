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

package v1alpha2

import (
	"testing"

	errors "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func fakeGetErr(cli dynamic.Interface, name, namespace string, opts *GetOption) (*unstructured.Unstructured, error) {
	return nil, errors.New("some error")
}

func fakeGetOk(cli dynamic.Interface, name, namespace string, opts *GetOption) (*unstructured.Unstructured, error) {
	return &unstructured.Unstructured{}, nil
}

func fakeCreateOk(cli dynamic.Interface, obj *unstructured.Unstructured, opts *CreateOption) (*unstructured.Unstructured, error) {
	return &unstructured.Unstructured{}, nil
}

func fakeCreateErr(cli dynamic.Interface, obj *unstructured.Unstructured, opts *CreateOption) (*unstructured.Unstructured, error) {
	return nil, errors.New("failed to create PVC")
}

func fakeDeleteErr(cli dynamic.Interface, obj *unstructured.Unstructured, opts *DeleteOption) error {
	return errors.New("some error")
}

func fakeDeleteOk(cli dynamic.Interface, obj *unstructured.Unstructured, opts *DeleteOption) error {
	return nil
}

func fakeGetClientSetOk() (dynamic.Interface, error) {
	return dynamic.NewForConfig(&rest.Config{})
}

func fakeGetClientSetNil() (dynamic.Interface, error) {
	return nil, nil
}

func fakeGetClientSetErr() (dynamic.Interface, error) {
	return nil, errors.New("fake-error")
}

func fakeGetClientSetForPathOk(fakeConfigPath string) (dynamic.Interface, error) {
	return dynamic.NewForConfig(&rest.Config{})
}

func fakeGetClientSetForPathErr(fakeConfigPath string) (dynamic.Interface, error) {
	return nil, errors.New("fake error")
}

func TestNewKubeClient(t *testing.T) {
	tests := map[string]struct {
		fakeKubeClientBuildOption KubeclientBuildOption
		expectedPath              bool
	}{
		"T1": {WithKubeConfigPath("fake-path"), true},
		//"T2": {nil, false},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			kubeclient := NewKubeClient(mock.fakeKubeClientBuildOption)
			if mock.expectedPath && kubeclient.kubeConfigPath == "" {
				t.Fatalf("test %q failed: expected kubeConfigPath not to be empty", name)
			}
			if !mock.expectedPath && kubeclient.kubeConfigPath != "" {
				t.Fatalf("test %q failed: expected kubeConfigPath to be empty", name)
			}
		})
	}
}

func TestWithDefaults(t *testing.T) {
	tests := map[string]struct {
		getClientSetFn      getClientsetFn
		getClientsetForPath getClientsetForPathFn
		getFn               GetFn
		createFn            CreateFn
		deleteFn            DeleteFn
	}{
		"T1": {fakeGetClientSetNil, nil, nil, nil, nil},
		"T2": {fakeGetClientSetErr, fakeGetClientSetForPathErr, fakeGetOk, fakeCreateOk, fakeDeleteOk},
		"T3": {fakeGetClientSetOk, fakeGetClientSetForPathOk, fakeGetOk, fakeCreateOk, fakeDeleteOk},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset: mock.getClientSetFn,
				create:       mock.createFn,
				delete:       mock.deleteFn,
				get:          mock.getFn,
			}
			withDefaults(fc)
			if fc.get == nil {
				t.Fatalf("test %q failed: expected fc.get not to be nil", name)
			}
			if fc.getClientsetForPath == nil {
				t.Fatalf("test %q failed: expected getClientsetForPath not to be nil", name)
			}
			if fc.create == nil {
				t.Fatalf("test %q failed: expected fc.create not to be nil", name)
			}
			if fc.delete == nil {
				t.Fatalf("test %q failed: expected fc.delete not to be nil", name)
			}
			if fc.getClientset == nil {
				t.Fatalf("test %q failed: expected fc.delete not to be nil", name)
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
		name, mock := name, mock
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

func TestGetClientSetPathOrDirect(t *testing.T) {
	tests := map[string]struct {
		kubeConfigPath      string
		getClientSetForPath getClientsetForPathFn
		getClientSet        getClientsetFn
		expectErr           bool
	}{
		"T1": {"fake-path", fakeGetClientSetForPathOk, fakeGetClientSetOk, false},
		"T2": {"", fakeGetClientSetForPathOk, fakeGetClientSetOk, false},
		"T3": {"fakepath", fakeGetClientSetForPathOk, fakeGetClientSetErr, false},
		"T4": {"", fakeGetClientSetForPathErr, fakeGetClientSetOk, false},
		"T5": {"fake-path", fakeGetClientSetForPathErr, fakeGetClientSetOk, true},
		"T6": {"", fakeGetClientSetForPathOk, fakeGetClientSetErr, true},
		"T7": {"fakepath", fakeGetClientSetForPathErr, fakeGetClientSetErr, true},
	}

	for name, mock := range tests {
		name, mock := name, mock
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

func TestKubernetesGet(t *testing.T) {
	tests := map[string]struct {
		getClientSetFn getClientsetFn
		get            GetFn
		name           string
		getOptionFn    GetOptionFn
		ExpectErr      bool
	}{
		"T1": {fakeGetClientSetNil, fakeGetOk, "fake-name", WithGetNamespace("fake-ns"), false},
		"T2": {fakeGetClientSetOk, fakeGetOk, "fake-name", WithGetNamespace("fake-ns"), false},
		// Negative casses
		"T5": {fakeGetClientSetErr, fakeGetOk, "fake-name", nil, true},
		"T6": {fakeGetClientSetNil, fakeGetErr, "fake-name", WithGetNamespace("fake-ns"), true},
		"T7": {fakeGetClientSetNil, fakeGetOk, "", WithGetOption(metav1.GetOptions{}), true},
		"T8": {fakeGetClientSetOk, fakeGetErr, "fake-name", WithGetNamespace("fake-ns"), true},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset: mock.getClientSetFn,
				get:          mock.get,
			}
			_, err := fc.Get(mock.name, mock.getOptionFn)
			if mock.ExpectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.ExpectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubernetesDelete(t *testing.T) {
	tests := map[string]struct {
		getClientSetFn getClientsetFn
		del            DeleteFn
		unstruct       *unstructured.Unstructured
		deleteOptionFn []DeleteOptionFn
		ExpectErr      bool
	}{
		"T1": {fakeGetClientSetOk, fakeDeleteOk, &unstructured.Unstructured{}, []DeleteOptionFn{}, false},
		"T2": {fakeGetClientSetOk, fakeDeleteOk, nil, []DeleteOptionFn{WithDeleteOption(&metav1.DeleteOptions{})}, false},
		// Negative casses
		"T5": {fakeGetClientSetErr, fakeDeleteOk, nil, nil, true},
		"T6": {fakeGetClientSetNil, fakeDeleteErr, nil, []DeleteOptionFn{}, true},
		"T7": {fakeGetClientSetOk, fakeDeleteErr, nil, nil, true},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset: mock.getClientSetFn,
				delete:       mock.del,
			}
			err := fc.Delete(mock.unstruct, mock.deleteOptionFn...)
			if mock.ExpectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.ExpectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubernetesCreate(t *testing.T) {
	tests := map[string]struct {
		getClientSetFn getClientsetFn
		create         CreateFn
		unstruct       *unstructured.Unstructured
		createOptionFn []CreateOptionFn
		ExpectErr      bool
	}{
		"T1": {fakeGetClientSetOk, fakeCreateOk, &unstructured.Unstructured{}, []CreateOptionFn{}, false},
		"T2": {fakeGetClientSetOk, fakeCreateOk, &unstructured.Unstructured{}, []CreateOptionFn{WithCreateOption(metav1.CreateOptions{})}, false},
		"T3": {fakeGetClientSetNil, fakeCreateOk, &unstructured.Unstructured{}, []CreateOptionFn{WithCreateOption(metav1.CreateOptions{}), WithCreateSubResources("fake-sub")}, false},
		// Negative casses
		"T5": {fakeGetClientSetErr, fakeCreateOk, nil, []CreateOptionFn{}, true},
		"T6": {fakeGetClientSetNil, fakeCreateErr, nil, []CreateOptionFn{WithCreateOption(metav1.CreateOptions{})}, true},
		"T7": {fakeGetClientSetOk, fakeCreateErr, nil, nil, true},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset: mock.getClientSetFn,
				create:       mock.create,
			}
			err := fc.Create(mock.unstruct, mock.createOptionFn...)
			if mock.ExpectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.ExpectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubernetesCreateAll(t *testing.T) {
	tests := map[string]struct {
		getClientSetFn getClientsetFn
		create         CreateFn
		unstruct       *unstructured.Unstructured
		ExpectErr      bool
	}{
		"T1": {fakeGetClientSetOk, fakeCreateOk, &unstructured.Unstructured{}, false},
		"T2": {fakeGetClientSetOk, fakeCreateOk, &unstructured.Unstructured{}, false},
		"T3": {fakeGetClientSetNil, fakeCreateOk, &unstructured.Unstructured{}, false},
		// Negative casses
		"T5": {fakeGetClientSetErr, fakeCreateOk, nil, true},
		"T6": {fakeGetClientSetNil, fakeCreateErr, nil, true},
		"T7": {fakeGetClientSetOk, fakeCreateErr, nil, true},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset: mock.getClientSetFn,
				create:       mock.create,
			}
			errs := fc.CreateAllOrNone(mock.unstruct)
			if mock.ExpectErr && len(errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.ExpectErr && len(errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubernetesDeleteAll(t *testing.T) {
	tests := map[string]struct {
		getClientSetFn getClientsetFn
		del            DeleteFn
		unstruct       *unstructured.Unstructured
		ExpectErr      bool
	}{
		"T1": {fakeGetClientSetOk, fakeDeleteOk, &unstructured.Unstructured{}, false},
		"T2": {fakeGetClientSetOk, fakeDeleteOk, nil, false},
		"T3": {fakeGetClientSetNil, fakeDeleteOk, nil, false},
		// Negative casses
		"T5": {fakeGetClientSetErr, fakeDeleteOk, nil, true},
		"T6": {fakeGetClientSetNil, fakeDeleteErr, nil, true},
		"T7": {fakeGetClientSetOk, fakeDeleteErr, nil, true},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset: mock.getClientSetFn,
				delete:       mock.del,
			}
			errs := fc.DeleteAll(mock.unstruct)
			if mock.ExpectErr && len(errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.ExpectErr && len(errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
