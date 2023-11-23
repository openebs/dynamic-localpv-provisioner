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
	"testing"

	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
)

func fakeGetClientSetOk() (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeListOk(cli *clientset.Clientset, opts metav1.ListOptions) (*corev1.NodeList, error) {
	return &corev1.NodeList{}, nil
}

func fakeListErr(cli *clientset.Clientset, opts metav1.ListOptions) (*corev1.NodeList, error) {
	return nil, errors.New("fake error")
}

func fakeGetClientSetNil() (clientset *clientset.Clientset, err error) {
	return nil, nil
}

func fakeGetClientSetErr() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("fake error")
}

func fakeGetClientSetForPathOk(fakeConfigPath string) (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeGetClientSetForPathErr(fakeConfigPath string) (cli *clientset.Clientset, err error) {
	return nil, errors.New("fake error")
}

func fakePatchFnOk(cli *clientset.Clientset, name string,
	pt types.PatchType, data []byte,
	subresources ...string) (*corev1.Node, error) {
	return &corev1.Node{}, nil
}

func fakePatchFnErr(cli *clientset.Clientset, name string,
	pt types.PatchType, data []byte,
	subresources ...string) (*corev1.Node, error) {
	return nil, errors.New("fake error")
}

func fakeUpdateFnOk(cli *clientset.Clientset,
	node *corev1.Node) (*corev1.Node, error) {
	return &corev1.Node{}, nil
}

func fakeUpdateFnErr(cli *clientset.Clientset,
	node *corev1.Node) (*corev1.Node, error) {
	return nil, errors.New("fake error")

}

func TestWithDefaultOptions(t *testing.T) {
	tests := map[string]struct {
		getClientSet getClientsetFn
		list         listFn
	}{
		"T1": {fakeGetClientSetErr, fakeListErr},
		"T2": {nil, nil},
		"T3": {fakeGetClientSetOk, nil},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := Kubeclient{
				getClientset: mock.getClientSet,
				list:         mock.list,
			}
			fc.withDefaults()
			if fc.getClientset == nil {
				t.Fatalf("test %q failed: expected getClientset not to be nil", name)
			}
			if fc.getClientsetForPath == nil {
				t.Fatalf("test %q failed: expected getClientset not to be nil", name)
			}
			if fc.list == nil {
				t.Fatalf("test %q failed: expected list not to be empty", name)
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
			fc := &Kubeclient{
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

func TestKubenetesNodeList(t *testing.T) {
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
			_, err := fc.List(metav1.ListOptions{})
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestNodePatch(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		patch               patchFn
		Name                string
		expectErr           bool
	}{
		"Patch Test 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", fakePatchFnOk, "alpha-1", true},
		"Patch Test 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", fakePatchFnOk, "alpha-2", true},
		"Patch Test 3": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakePatchFnOk, "beta-1", false},
		"Patch Test 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fp", fakePatchFnErr, "beta-2", true},
		"Patch Test 5": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fakepath", fakePatchFnOk, "", true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				patch:               mock.patch,
			}
			//fake data
			data, _ := json.Marshal(mock)
			_, err := k.Patch(mock.Name, types.MergePatchType, data)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestCordonViaPatch(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "alpha1",
		},
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
	}

	cases := []struct {
		name                  string
		getClientset          getClientsetFn
		patch                 patchFn
		expectedErr, isCordon bool
	}{
		{
			name:         "cordon node",
			getClientset: fakeGetClientSetOk,
			patch:        fakePatchFnOk,
			isCordon:     true,
			expectedErr:  false,
		},
		{
			name:         "cordon node with add new taint",
			getClientset: fakeGetClientSetOk,
			patch:        fakePatchFnOk,
			isCordon:     true,
			expectedErr:  false,
		},
		{
			name:         "uncordon node with no changes taints",
			getClientset: fakeGetClientSetOk,
			patch:        fakePatchFnOk,
			isCordon:     false,
			expectedErr:  false,
		},
		{
			name:         "uncordon node with client err",
			getClientset: fakeGetClientSetErr,
			patch:        fakePatchFnOk,
			isCordon:     false,
			expectedErr:  true,
		},
		{
			name:         "uncordon node with fake patch err",
			getClientset: fakeGetClientSetOk,
			patch:        fakePatchFnErr,
			isCordon:     false,
			expectedErr:  true,
		},
	}

	for _, mock := range cases {
		k := &Kubeclient{
			getClientset: mock.getClientset,
			patch:        mock.patch,
		}

		err := k.CordonViaPatch(node, mock.isCordon)
		if mock.expectedErr && err == nil {
			t.Fatalf("Test %q failed: expected error not to be nil", mock.name)
		}
		if !mock.expectedErr && err != nil {
			t.Fatalf("Test %q failed: expected error to be nil", mock.name)
		}
	}
}

func TestCordonViaUpdate(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "alpha1",
		},
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
	}

	cases := []struct {
		name                  string
		getClientset          getClientsetFn
		update                updateFn
		expectedErr, isCordon bool
	}{
		{
			name:         "cordon node with no changes taints",
			getClientset: fakeGetClientSetOk,
			update:       fakeUpdateFnOk,
			isCordon:     true,
			expectedErr:  false,
		},
		{
			name:         "cordon node with add new taint",
			getClientset: fakeGetClientSetOk,
			update:       fakeUpdateFnOk,
			isCordon:     true,
			expectedErr:  false,
		},
		{
			name:         "uncordon node with no changes taints",
			getClientset: fakeGetClientSetOk,
			update:       fakeUpdateFnOk,
			isCordon:     false,
			expectedErr:  false,
		},
		{
			name:         "uncordon node with client err",
			getClientset: fakeGetClientSetErr,
			update:       fakeUpdateFnOk,
			isCordon:     false,
			expectedErr:  true,
		},
		{
			name:         "uncordon node with fake update err",
			getClientset: fakeGetClientSetOk,
			update:       fakeUpdateFnErr,
			isCordon:     false,
			expectedErr:  true,
		},
	}

	for _, mock := range cases {
		k := &Kubeclient{
			getClientset: mock.getClientset,
			update:       mock.update,
		}

		err := k.CordonViaUpdate(node, mock.isCordon)
		if mock.expectedErr && err == nil {
			t.Fatalf("Test %q failed: expected error not to be nil", mock.name)
		}
		if !mock.expectedErr && err != nil {
			t.Fatalf("Test %q failed: expected error to be nil", mock.name)
		}

	}
}

func TestCordonWithTaintsPatch(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "alpha1",
		},
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
	}

	cases := []struct {
		name                  string
		getClientset          getClientsetFn
		patch                 patchFn
		taintsToAdd           []corev1.Taint
		expectedTaints        []corev1.Taint
		expectedErr, isCordon bool
	}{
		{
			name:           "cordon node with no changes taints",
			getClientset:   fakeGetClientSetOk,
			patch:          fakePatchFnOk,
			taintsToAdd:    []corev1.Taint{},
			expectedTaints: node.Spec.Taints,
			isCordon:       true,
			expectedErr:    false,
		},
		{
			name:         "cordon node with add new taint",
			getClientset: fakeGetClientSetOk,
			patch:        fakePatchFnOk,
			taintsToAdd: []corev1.Taint{
				{
					Key:    "foo_1",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
			expectedTaints: append([]corev1.Taint{{Key: "foo_1",
				Effect: corev1.TaintEffectNoExecute}},
				node.Spec.Taints...),
			isCordon:    true,
			expectedErr: false,
		},
		{
			name:           "uncordon node with no changes taints",
			getClientset:   fakeGetClientSetOk,
			patch:          fakePatchFnOk,
			taintsToAdd:    []corev1.Taint{},
			expectedTaints: node.Spec.Taints,
			isCordon:       false,
			expectedErr:    false,
		},
		{
			name:           "uncordon node with client err",
			getClientset:   fakeGetClientSetErr,
			patch:          fakePatchFnOk,
			taintsToAdd:    []corev1.Taint{},
			expectedTaints: node.Spec.Taints,
			isCordon:       false,
			expectedErr:    true,
		},
		{
			name:           "uncordon node with fake patch err",
			getClientset:   fakeGetClientSetOk,
			patch:          fakePatchFnErr,
			taintsToAdd:    []corev1.Taint{},
			expectedTaints: node.Spec.Taints,
			isCordon:       false,
			expectedErr:    true,
		},
	}

	for _, mock := range cases {
		k := &Kubeclient{
			getClientset: mock.getClientset,
			patch:        mock.patch,
		}

		b := NewBuilder().WithAPINode(node).WithTaints(mock.taintsToAdd)
		err := k.CordonViaPatch(b.Node.object, mock.isCordon)
		if mock.expectedErr && err == nil {
			t.Fatalf("Test %q failed: expected error not to be nil", mock.name)
		}
		if !mock.expectedErr && err != nil {
			t.Fatalf("Test %q failed: expected error to be nil", mock.name)
		}
	}
}

func fakeOptionGetNonNil(k *Kubeclient) {
	k.get = func(cli *kubernetes.Clientset,
		name string,
		opts metav1.GetOptions,
	) (*corev1.Node, error) {
		return &corev1.Node{}, nil
	}
}

func fakeOptionGetErr(k *Kubeclient) {
	k.get = func(cli *kubernetes.Clientset,
		name string,
		opts metav1.GetOptions,
	) (*corev1.Node, error) {
		return nil, errors.New("Woo!!!")
	}
}

func fakeOptionGetValid(k *Kubeclient) {
	k.get = func(cli *kubernetes.Clientset,
		name string,
		opts metav1.GetOptions,
	) (*corev1.Node, error) {
		return &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					labelKeyHostName: "apple",
				},
			},
		}, nil
	}
}

func fakeOptionClientsetNonNil(k *Kubeclient) {
	k.clientset = &clientset.Clientset{}
}

func TestGetHostNameOrNodeName(t *testing.T) {
	tests := map[string]struct {
		getfn          KubeClientBuildOption
		nodename       string
		expectErr      bool
		expectHostname string
	}{
		"mising labels": {
			getfn:          fakeOptionGetNonNil,
			nodename:       "lenovo-laptop",
			expectErr:      false,
			expectHostname: "lenovo-laptop",
		},
		"error during get": {
			getfn:     fakeOptionGetErr,
			expectErr: true,
		},
		"valid get": {
			getfn:          fakeOptionGetValid,
			nodename:       "lenovo-laptop",
			expectErr:      false,
			expectHostname: "apple",
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			name := name // pin it
			mock := mock // pin it
			gotHostName, err := NewKubeClient(mock.getfn, fakeOptionClientsetNonNil).
				GetHostNameOrNodeName(mock.nodename, metav1.GetOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("test %q failed: expected error got nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("test %q failed: expected no error got err: %+v", name, err)
			}
			if !mock.expectErr && mock.expectHostname != gotHostName {
				t.Fatalf(
					"test %q failed: expected hostname %q got %q",
					name,
					mock.expectHostname,
					gotHostName,
				)
			}
		})
	}
}

func TestGetHostName(t *testing.T) {
	tests := map[string]struct {
		getfn          KubeClientBuildOption
		nodename       string
		expectErr      bool
		expectHostname string
	}{
		"mising labels": {
			getfn:     fakeOptionGetNonNil,
			expectErr: true,
		},
		"error during get": {
			getfn:     fakeOptionGetErr,
			expectErr: true,
		},
		"valid get": {
			getfn:          fakeOptionGetValid,
			nodename:       "lenovo-laptop",
			expectErr:      false,
			expectHostname: "apple",
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			name := name // pin it
			mock := mock // pin it
			gotHostName, err := NewKubeClient(mock.getfn, fakeOptionClientsetNonNil).
				GetHostName(mock.nodename, metav1.GetOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("test %q failed: expected error got nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("test %q failed: expected no error got err: %+v", name, err)
			}
			if !mock.expectErr && mock.expectHostname != gotHostName {
				t.Fatalf(
					"test %q failed: expected hostname %q got %q",
					name,
					mock.expectHostname,
					gotHostName,
				)
			}
		})
	}
}
