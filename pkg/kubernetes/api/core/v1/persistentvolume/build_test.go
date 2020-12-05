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

package persistentvolume

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
			name: "PV1",
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
			}},
			expectErr: false,
		},
		"Test Builder without name": {
			name: "",
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
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

func TestBuildWithAnnotations(t *testing.T) {
	tests := map[string]struct {
		annotations map[string]string
		builder     *Builder
		expectErr   bool
	}{
		"Test Builderwith annotations": {
			annotations: map[string]string{"persistent-volume": "PV", "application": "percona"},
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
			}},
			expectErr: false,
		},
		"Test Builderwithout annotations": {
			annotations: map[string]string{},
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
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
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
			}},
			expectErr: false,
		},
		"Test Builderwithout labels": {
			labels: map[string]string{},
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
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

func TestBuildWithAccessModes(t *testing.T) {
	tests := map[string]struct {
		accessModes []corev1.PersistentVolumeAccessMode
		builder     *Builder
		expectErr   bool
	}{
		"Test Builderwith accessModes": {
			accessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce, corev1.ReadOnlyMany},
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
			}},
			expectErr: false,
		},
		"Test Builderwithout accessModes": {
			accessModes: []corev1.PersistentVolumeAccessMode{},
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
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

func TestBuildWithCapacity(t *testing.T) {
	tests := map[string]struct {
		capacity  string
		builder   *Builder
		expectErr bool
	}{
		"Test Builderwith capacity": {
			capacity: "5G",
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
			}},
			expectErr: false,
		},
		"Test Builderwithout capacity": {
			capacity: "",
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
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

func TestBuildWithLocalHostDirectory(t *testing.T) {
	tests := map[string]struct {
		path      string
		builder   *Builder
		expectErr bool
	}{
		"Test Builderwith hostpath": {
			path: "/var/openebs/local",
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
			}},
			expectErr: false,
		},
		"Test Builderwithout hostpath": {
			path: "",
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithLocalHostDirectory(mock.path)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithNodeAffinityHostname(t *testing.T) {
	tests := map[string]struct {
		nodeName  string
		builder   *Builder
		expectErr bool
	}{
		"Test Builderwith node name": {
			nodeName: "node1",
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
			}},
			expectErr: false,
		},
		"Test Builderwithout node name": {
			nodeName: "",
			builder: &Builder{pv: &PV{
				object: &corev1.PersistentVolume{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithNodeAffinityHostname(mock.nodeName)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildHostPath(t *testing.T) {
	fsType := ""

	tests := map[string]struct {
		name        string
		capacity    string
		path        string
		nodeName    string
		expectedPV  *corev1.PersistentVolume
		expectedErr bool
	}{
		"Hostpath PV with correct details": {
			name:     "PV1",
			capacity: "10Ti",
			path:     "/var/openebs/local/PV1",
			nodeName: "node1",
			expectedPV: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "PV1"},
				Spec: corev1.PersistentVolumeSpec{
					Capacity: corev1.ResourceList{
						corev1.ResourceStorage: fakeCapacity("10Ti"),
					},
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						Local: &corev1.LocalVolumeSource{
							Path:   "/var/openebs/local/PV1",
							FSType: &fsType,
						},
					},
					NodeAffinity: &corev1.VolumeNodeAffinity{
						Required: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      KeyNode,
											Operator: corev1.NodeSelectorOpIn,
											Values: []string{
												"node1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedErr: false,
		},
		"Hostpath PV with error": {
			name:        "",
			capacity:    "500Gi",
			path:        "",
			nodeName:    "500Gi",
			expectedPV:  nil,
			expectedErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			pvObj, err := NewBuilder().
				WithName(mock.name).
				WithCapacity(mock.capacity).
				WithLocalHostDirectory(mock.path).
				WithNodeAffinityHostname(mock.nodeName).
				Build()
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
			if !reflect.DeepEqual(pvObj, mock.expectedPV) {
				t.Fatalf("Test %q failed: pv mismatch", name)
			}
		})
	}
}

func fakeCapacity(capacity string) resource.Quantity {
	q, _ := resource.ParseQuantity(capacity)
	return q
}
