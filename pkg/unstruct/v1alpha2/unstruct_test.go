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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	fakeK8sResource = `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: icstcee
  name: icstcee
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: icstcee
  type: LoadBalancer
`
	fakeInvalidK8sResource = `
apiVersion: v1
kind: Service
metadata
  labels
    app: icstcee
  name: icstcee
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: icstcee
  type: LoadBalancer
`
)

func fakePodObject() *unstructured.Unstructured {
	n := &unstructured.Unstructured{}
	n.SetKind("Pod")
	n.SetName("fake pod")
	return n
}

func fakeDeploymentObject() *unstructured.Unstructured {
	n := &unstructured.Unstructured{}
	n.SetKind("Deployment")
	n.SetName("fake deployment")
	return n
}

func fakeServiceObject() *unstructured.Unstructured {
	n := &unstructured.Unstructured{}
	n.SetKind("Service")
	n.SetName("fake service")
	return n
}

func fakeNamespaceObject() *unstructured.Unstructured {
	n := &unstructured.Unstructured{}
	n.SetKind("Namespace")
	n.SetName("fake namespace")
	return n
}

func fakeUnstructObject(name string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetName(name)
	return u
}

func fakeK8sResourceList(resource string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += resource + "---"
	}
	return result
}

func TestBuilderForYAML(t *testing.T) {
	tests := map[string]struct {
		resourceYAML, expectedName string
		expectError                bool
	}{
		"Test 1": {fakeK8sResource, "icstcee", false},
		"Test 2": {fakeInvalidK8sResource, "", true},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := BuilderForYaml(mock.resourceYAML)
			if mock.expectError && len(b.errs) == 0 {
				t.Fatalf("Test %s failed, expected err but got 0", name)
			} else if b.unstruct.Object.GetName() != mock.expectedName {
				t.Fatalf("Test %s failed, expected %v but got %v", name, mock.expectedName, b.unstruct.Object.GetName())
			}
		})
	}
}

func TestBuilderForObject(t *testing.T) {
	tests := map[string]struct {
		resourceName, expectedName string
	}{
		"Test 1": {"icstcee", "icstcee"},
		"Test 2": {"icstcee1", "icstcee1"},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			mockObj := fakeUnstructObject(mock.resourceName)
			b := BuilderForObject(mockObj)
			if b.unstruct.Object.GetName() != mock.expectedName {
				t.Fatalf("Test %s failed, expected %v but got %v", name, mock.expectedName, b.unstruct.Object.GetName())
			}
		})
	}
}

func TestBuilderForYamlBuild(t *testing.T) {
	tests := map[string]struct {
		resourceYAML, expectedName string
		expectError                bool
	}{
		"Test 1": {fakeK8sResource, "icstcee", false},
		"Test 2": {fakeInvalidK8sResource, "", true},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b, err := BuilderForYaml(mock.resourceYAML).Build()
			if mock.expectError && err == nil {
				t.Fatalf("Test %s failed, expected err but got nil", name)
			} else if b != nil && b.Object.GetName() != mock.expectedName {
				t.Fatalf("Test %s failed, expected %v but got %v", name, mock.expectedName, b.Object.GetName())
			}
		})
	}
}

func TestListBuilderForYamls(t *testing.T) {
	tests := map[string]struct {
		resourceYAML          string
		expectedResourceCount int
		expectErr             bool
	}{
		"Test 1": {fakeK8sResourceList(fakeK8sResource, 1), 1, false},
		"Test 2": {fakeK8sResourceList(fakeK8sResource, 2), 2, false},
		"Test 3": {fakeK8sResourceList(fakeInvalidK8sResource, 1), 0, true},
		"Test 4": {fakeK8sResourceList(fakeInvalidK8sResource, 2), 0, true},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			lb := ListBuilderForYamls(mock.resourceYAML)
			if mock.expectErr && len(lb.errs) == 0 {
				t.Fatalf("Test %s failed, expected err but got nil", name)
			} else if len(lb.list.Items) != mock.expectedResourceCount {
				t.Fatalf("Test %s failed, expected resource count %v but got %v", name, mock.expectedResourceCount, len(lb.list.Items))
			}
		})
	}
}

func TestListUnstructBuilderForObjects(t *testing.T) {
	tests := map[string]struct {
		availableResources    []*unstructured.Unstructured
		expectedResourceCount int
		expectErr             bool
	}{
		"Test 1": {[]*unstructured.Unstructured{fakePodObject()}, 1, false},
		"Test 2": {[]*unstructured.Unstructured{fakePodObject(), fakeDeploymentObject()}, 2, false},
		"Test 3": {[]*unstructured.Unstructured{fakePodObject(), fakeDeploymentObject(), fakeServiceObject()}, 3, false},
		"Test 4": {[]*unstructured.Unstructured{fakePodObject(), fakeDeploymentObject(), fakeServiceObject(), fakeNamespaceObject()}, 4, false},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			lb := ListBuilderForObjects(mock.availableResources...)
			if mock.expectErr && len(lb.errs) == 0 {
				t.Fatalf("Test %s failed, expected err but got nil", name)
			} else if len(lb.list.Items) != mock.expectedResourceCount {
				t.Fatalf("Test %s failed, expected resource count %v but got %v", name, mock.expectedResourceCount, len(lb.list.Items))
			}
		})
	}
}
