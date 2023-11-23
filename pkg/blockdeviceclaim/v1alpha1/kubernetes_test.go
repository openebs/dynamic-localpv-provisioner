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
	"context"
	"encoding/json"
	"testing"

	errors "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/api/v3/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/api/v3/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/types"
)

func fakeGetClientsetOk() (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeGetClientsetForPathOk(fakeConfigPath string) (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeGetClientsetForPathErr(fakeConfigPath string) (cli *clientset.Clientset, err error) {
	return nil, errors.New("fake error")
}

func fakeGetFnOk(ctx context.Context, cli *clientset.Clientset, name, namespace string, opts metav1.GetOptions) (*apis.BlockDeviceClaim, error) {
	return &apis.BlockDeviceClaim{}, nil
}

func fakeListFnOk(ctx context.Context, cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.BlockDeviceClaimList, error) {
	return &apis.BlockDeviceClaimList{}, nil
}

func fakeDeleteFnOk(ctx context.Context, cli *clientset.Clientset, name, namespace string, opts *metav1.DeleteOptions) error {
	return nil
}

func fakeListFnErr(ctx context.Context, cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.BlockDeviceClaimList, error) {
	return &apis.BlockDeviceClaimList{}, errors.New("some error")
}

func fakeGetFnErr(ctx context.Context, cli *clientset.Clientset, name, namespace string, opts metav1.GetOptions) (*apis.BlockDeviceClaim, error) {
	return &apis.BlockDeviceClaim{}, errors.New("some error")
}

func fakeDeleteFnErr(ctx context.Context, cli *clientset.Clientset, name, namespace string, opts *metav1.DeleteOptions) error {
	return errors.New("some error")
}

func fakeGetClientsetErr() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeCreateFnOk(ctx context.Context, cli *clientset.Clientset, namespace string, bdc *apis.BlockDeviceClaim) (*apis.BlockDeviceClaim, error) {
	return &apis.BlockDeviceClaim{}, nil
}

func fakeCreateErr(ctx context.Context, cli *clientset.Clientset, namespace string, bdc *apis.BlockDeviceClaim) (*apis.BlockDeviceClaim, error) {
	return nil, errors.New("failed to create BDC")
}

func fakePatchFnOk(ctx context.Context, cli *clientset.Clientset, namespace, name string, pt types.PatchType, data []byte, subresources ...string) (*apis.BlockDeviceClaim, error) {
	return &apis.BlockDeviceClaim{}, nil
}

func fakePatchFnErr(ctx context.Context, cli *clientset.Clientset, namespace, name string, pt types.PatchType, data []byte, subresources ...string) (*apis.BlockDeviceClaim, error) {
	return nil, errors.New("fake error")
}

func TestWithDefaultOptionsForClientset(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
	}{
		"TestCase1":               {nil, nil},
		"When clientset is error": {fakeGetClientsetErr, fakeGetClientsetForPathErr},
		"When clientset is ok":    {fakeGetClientsetOk, fakeGetClientsetForPathOk},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
			}
			fc.WithDefaults()
			if fc.getClientset == nil {
				t.Fatalf("test %q failed: expected getClientset not to be empty", name)
			}
			if fc.getClientsetForPath == nil {
				t.Fatalf("test %q failed: expected getClientset not to be nil", name)
			}
		})
	}
}

func TestWithDefaultOptionsForCRUDOPS(t *testing.T) {
	tests := map[string]struct {
		list          listFn
		get           getFn
		create        createFn
		del           deleteFn
		delCollection deleteCollectionFn
	}{
		"TestCase1":               {nil, nil, nil, nil, nil},
		"When clientset is error": {fakeListFnErr, fakeGetFnErr, fakeCreateErr, fakeDeleteFnErr, nil},
		"When clientset is ok":    {fakeListFnOk, fakeGetFnOk, fakeCreateFnOk, fakeDeleteFnOk, nil},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				list:          mock.list,
				get:           mock.get,
				create:        mock.create,
				del:           mock.del,
				delCollection: mock.delCollection,
			}
			fc.WithDefaults()
			if fc.list == nil {
				t.Fatalf("test %q failed: expected list not to be empty", name)
			}
			if fc.get == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
			}
			if fc.create == nil {
				t.Fatalf("test %q failed: expected create not to be empty", name)
			}
			if fc.del == nil {
				t.Fatalf("test %q failed: expected del not to be empty", name)
			}
			if fc.delCollection == nil {
				t.Fatalf("test %q failed: expected delCollection not to be empty", name)
			}
		})
	}
}

func TestGetClientsetForPathOrDirect(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		isErr               bool
	}{
		// Positive tests
		"Positive 1": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "", false},
		"Positive 2": {fakeGetClientsetErr, fakeGetClientsetForPathOk, "fake-path", false},
		"Positive 3": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "", false},

		// Negative tests
		"Negative 1": {fakeGetClientsetErr, fakeGetClientsetForPathOk, "", true},
		"Negative 2": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "fake-path", true},
		"Negative 3": {fakeGetClientsetErr, fakeGetClientsetForPathErr, "fake-path", true},
		"Negative 4": {fakeGetClientsetErr, fakeGetClientsetForPathErr, "", true},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
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
			h := WithKubeClient(mock.Clientset)
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

func TestGetClientOrCached(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		expectErr           bool
	}{
		// Positive tests
		"Positive 1": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "", false},
		"Positive 2": {fakeGetClientsetErr, fakeGetClientsetForPathOk, "fake-path", false},
		"Positive 3": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "", false},

		// Negative tests
		"Negative 1": {fakeGetClientsetErr, fakeGetClientsetForPathOk, "", true},
		"Negative 2": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "fake-path", true},
		"Negative 3": {fakeGetClientsetErr, fakeGetClientsetForPathErr, "fake-path", true},
		"Negative 4": {fakeGetClientsetErr, fakeGetClientsetForPathErr, "", true},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
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

func TestBlockDeviceClaimList(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		list                listFn
		expectedErr         bool
	}{
		// Positive tests
		"Positive 1": {nil, fakeGetClientsetForPathOk, "fake-path", fakeListFnOk, false},
		"Positive 2": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "", fakeListFnOk, false},
		"Positive 3": {fakeGetClientsetErr, fakeGetClientsetForPathOk, "fake-path", fakeListFnOk, false},
		"Positive 4": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "", fakeListFnOk, false},

		// Negative tests
		"Negative 1": {fakeGetClientsetErr, fakeGetClientsetForPathOk, "", fakeListFnOk, true},
		"Negative 2": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "fake-path", fakeListFnOk, true},
		"Negative 3": {fakeGetClientsetErr, fakeGetClientsetForPathErr, "fake-path", fakeListFnOk, true},
		"Negative 4": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "", fakeListFnErr, true},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
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

func TestBlockDeviceClaimGet(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		get                 getFn
		bdName              string
		expectErr           bool
	}{
		"Test 1": {fakeGetClientsetErr, fakeGetClientsetForPathOk, "", fakeGetFnOk, "bd-1", true},
		"Test 2": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "fake-path", fakeGetFnOk, "bd-1", true},
		"Test 3": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "", fakeGetFnOk, "bd-2", false},
		"Test 4": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "fp", fakeGetFnErr, "bd-3", true},
		"Test 5": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "fakepath", fakeGetFnOk, "", true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				namespace:           "default",
				get:                 mock.get,
			}
			_, err := k.Get(context.TODO(), mock.bdName, metav1.GetOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBlockDeviceClaimDelete(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		bdName              string
		delete              deleteFn
		expectErr           bool
	}{
		"Test 1": {fakeGetClientsetErr, fakeGetClientsetForPathOk, "", "bd-1", fakeDeleteFnOk, true},
		"Test 2": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "fake-path2", "bd-2", fakeDeleteFnOk, false},
		"Test 3": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "", "bd-3", fakeDeleteFnErr, true},
		"Test 4": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "fakepath", "", fakeDeleteFnOk, true},
		"Test 5": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "fake-path2", "bd1", fakeDeleteFnOk, true},
		"Test 6": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "fake-path2", "bd1", fakeDeleteFnErr, true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				namespace:           "",
				del:                 mock.delete,
			}
			err := k.Delete(context.TODO(), mock.bdName, &metav1.DeleteOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBlockDeviceClaimCreate(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		create              createFn
		bdc                 *apis.BlockDeviceClaim
		expectErr           bool
	}{
		"Test 1": {
			getClientset:        fakeGetClientsetErr,
			getClientsetForPath: fakeGetClientsetForPathErr,
			kubeConfigPath:      "",
			create:              fakeCreateFnOk,
			bdc:                 &apis.BlockDeviceClaim{ObjectMeta: metav1.ObjectMeta{Name: "BDC-1"}},
			expectErr:           true,
		},
		"Test 2": {
			getClientset:        fakeGetClientsetOk,
			getClientsetForPath: fakeGetClientsetForPathOk,
			kubeConfigPath:      "",
			create:              fakeCreateErr,
			bdc:                 &apis.BlockDeviceClaim{ObjectMeta: metav1.ObjectMeta{Name: "BDC-2"}},
			expectErr:           true,
		},
		"Test 3": {
			getClientset:        fakeGetClientsetOk,
			getClientsetForPath: fakeGetClientsetForPathOk,
			kubeConfigPath:      "fake-path",
			create:              fakeCreateErr,
			bdc:                 nil,
			expectErr:           true,
		},
		"Test 4": {
			getClientset:        fakeGetClientsetErr,
			getClientsetForPath: fakeGetClientsetForPathOk,
			kubeConfigPath:      "fake-path",
			create:              fakeCreateFnOk,
			bdc:                 nil,
			expectErr:           true,
		},
		"Test 5": {
			getClientset:        fakeGetClientsetOk,
			getClientsetForPath: fakeGetClientsetForPathErr,
			kubeConfigPath:      "fake-path",
			create:              fakeCreateFnOk,
			bdc:                 nil,
			expectErr:           true,
		},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				create:              mock.create,
			}
			_, err := fc.Create(context.TODO(), mock.bdc)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBlockDeviceClaimPatch(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		patch               patchFn
		bdName              string
		expectErr           bool
	}{
		"Test 1": {fakeGetClientsetErr, fakeGetClientsetForPathOk, "", fakePatchFnOk, "vsd-1", true},
		"Test 2": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "fake-path", fakePatchFnOk, "vsd-1", true},
		"Test 3": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "", fakePatchFnOk, "vsd-2", false},
		"Test 4": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "fp", fakePatchFnErr, "vsd-3", true},
		"Test 5": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "fakepath", fakePatchFnOk, "", true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				patch:               mock.patch,
			}
			//fake data
			data, _ := json.Marshal(mock)
			_, err := k.Patch(context.TODO(), mock.bdName, types.MergePatchType, data)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
