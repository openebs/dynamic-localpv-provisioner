/*
Copyright 2020 The OpenEBS Authors

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

package persistentvolumeclaim

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuilderWithName(t *testing.T) {
	tests := map[string]struct {
		name      string
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with name": {
			name: "PVC1",
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: false,
		},
		"Test Builder without name": {
			name: "",
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithName(mock.name)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithNamespace(t *testing.T) {
	tests := map[string]struct {
		namespace string
		builder   *Builder
		expectErr bool
	}{
		"Test Builderwith namespae": {
			namespace: "jiva-ns",
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: false,
		},
		"Test Builderwithout namespace": {
			namespace: "",
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: false,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithNamespace(mock.namespace)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithAnnotations(t *testing.T) {
	tests := map[string]struct {
		annotations map[string]string
		builder     *Builder
		expectErr   bool
	}{
		"Test Builderwith annotations": {
			annotations: map[string]string{"persistent-volume": "PV", "application": "percona"},
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: false,
		},
		"Test Builderwithout annotations": {
			annotations: map[string]string{},
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithAnnotations(mock.annotations)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithLabels(t *testing.T) {
	tests := map[string]struct {
		labels    map[string]string
		builder   *Builder
		expectErr bool
	}{
		"Test Builderwith labels": {
			labels: map[string]string{"persistent-volume": "PV", "application": "percona"},
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: false,
		},
		"Test Builderwithout labels": {
			labels: map[string]string{},
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithLabels(mock.labels)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithLabelsNew(t *testing.T) {
	tests := map[string]struct {
		labels    map[string]string
		builder   *Builder
		expectErr bool
	}{
		"Test 1 - with labels": {
			labels: map[string]string{
				"persistent-volume": "PV",
				"application":       "percona",
			},
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: false,
		},

		"Test 2 - with empty labels": {
			labels: map[string]string{},
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: true,
		},

		"Test 3 - with nil labels": {
			labels: nil,
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: true,
		},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithLabelsNew(mock.labels)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithAccessModes(t *testing.T) {
	tests := map[string]struct {
		accessModes []corev1.PersistentVolumeAccessMode
		builder     *Builder
		expectErr   bool
	}{
		"Test Builderwith accessModes": {
			accessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce, corev1.ReadOnlyMany},
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: false,
		},
		"Test Builderwithout accessModes": {
			accessModes: []corev1.PersistentVolumeAccessMode{},
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: true,
		},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithAccessModes(mock.accessModes)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithStorageClass(t *testing.T) {
	tests := map[string]struct {
		scName    string
		builder   *Builder
		expectErr bool
	}{
		"Test Builderwith SC": {
			scName: "single-replica",
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: false,
		},
		"Test Builderwithout SC": {
			scName: "",
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: false,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithStorageClass(mock.scName)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithCapacity(t *testing.T) {
	tests := map[string]struct {
		capacity  string
		builder   *Builder
		expectErr bool
	}{
		"Test Builderwith capacity": {
			capacity: "5G",
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: false,
		},
		"Test Builderwithout capacity": {
			capacity: "",
			builder: &Builder{pvc: &PVC{
				object: &corev1.PersistentVolumeClaim{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithCapacity(mock.capacity)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	tests := map[string]struct {
		name        string
		capacity    string
		expectedPVC *corev1.PersistentVolumeClaim
		expectedErr bool
	}{
		"PVC with correct details": {
			name:     "PVC1",
			capacity: "10Ti",
			expectedPVC: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "PVC1"},
				Spec: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: fakeCapacity("10Ti"),
						},
					},
				},
			},
			expectedErr: false,
		},
		"PVC with error": {
			name:        "",
			capacity:    "500Gi",
			expectedPVC: nil,
			expectedErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			pvcObj, err := NewBuilder().WithName(mock.name).WithCapacity(mock.capacity).Build()
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
			if !reflect.DeepEqual(pvcObj, mock.expectedPVC) {
				t.Fatalf("Test %q failed: pvc mismatch", name)
			}
		})
	}
}

func fakeCapacity(capacity string) resource.Quantity {
	q, _ := resource.ParseQuantity(capacity)
	return q
}
