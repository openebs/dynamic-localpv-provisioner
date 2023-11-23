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

package v1alpha1

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/api/v3/pkg/apis/openebs.io/v1alpha1"
	ndm "github.com/openebs/api/v3/pkg/apis/openebs.io/v1alpha1"
)

func TestBuilderWithName(t *testing.T) {
	tests := map[string]struct {
		name      string
		expectErr bool
	}{
		"Test Builder with name": {
			name:      "BDC1",
			expectErr: false,
		},
		"Test Builder without name": {
			name:      "",
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().WithName(mock.name)
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
		expectErr bool
	}{
		"Test Builderwith namespae": {
			namespace: "jiva-ns",
			expectErr: false,
		},
		"Test Builderwithout namespace": {
			namespace: "",
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().WithNamespace(mock.namespace)
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
		expectErr   bool
	}{
		"Test Builderwith annotations": {
			annotations: map[string]string{"persistent-volume": "PV", "application": "percona"},
			expectErr:   false,
		},
		"Test Builderwithout annotations": {
			annotations: map[string]string{},
			expectErr:   true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().WithAnnotations(mock.annotations)
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
		expectErr bool
	}{
		"Test Builderwith labels": {
			labels:    map[string]string{"persistent-volume": "PV", "application": "percona"},
			expectErr: false,
		},
		"Test Builderwithout labels": {
			labels:    map[string]string{},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().WithLabels(mock.labels)
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
			labels: map[string]string{"blockdeviceclaim": "BDC", "application": "percona"},
			builder: &Builder{BDC: &BlockDeviceClaim{
				Object: &apis.BlockDeviceClaim{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"openebs.io/storage-pool-claim": "cstor-pool"},
					},
				},
			}},
			expectErr: false,
		},
		"Test Builderwithout labels": {
			labels: map[string]string{},
			builder: &Builder{BDC: &BlockDeviceClaim{
				Object: &apis.BlockDeviceClaim{},
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

func TestBuildWithCapacity(t *testing.T) {
	tests := map[string]struct {
		capacity  string
		expectErr bool
	}{
		"Test Builderwith capacity": {
			capacity:  "5G",
			expectErr: false,
		},
		"Test Builderwithout capacity": {
			capacity:  "",
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().WithCapacity(mock.capacity)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithBlockDeviceTag(t *testing.T) {
	tests := map[string]struct {
		tag       string
		expectErr bool
	}{
		"Test Builder with tag": {
			tag:       "test",
			expectErr: false,
		},
		"Test Builder without tag": {
			tag:       "",
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().WithBlockDeviceTag(mock.tag)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilder_WithSelector(t *testing.T) {
	tests := map[string]struct {
		labelSelector   map[string]string
		ExpectedBuilder *Builder
		expectErr       bool
	}{
		"Test Builder with empty labelSelector map": {
			labelSelector: map[string]string{},
			ExpectedBuilder: &Builder{
				BDC: &BlockDeviceClaim{
					Object: &apis.BlockDeviceClaim{
						Spec: apis.DeviceClaimSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{},
							},
						},
					},
				},
			},
			expectErr: true,
		},
		"Test Builder with labelSelector map key having empty value": {
			labelSelector: map[string]string{
				"testKey": "",
			},
			ExpectedBuilder: &Builder{
				BDC: &BlockDeviceClaim{
					Object: &apis.BlockDeviceClaim{
						Spec: apis.DeviceClaimSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{},
							},
						},
					},
				},
			},
			expectErr: false,
		},
		"Test Builder with non-empty labelSelector map": {
			labelSelector: map[string]string{
				"testKey": "testValue",
			},
			ExpectedBuilder: &Builder{
				BDC: &BlockDeviceClaim{
					Object: &apis.BlockDeviceClaim{
						Spec: apis.DeviceClaimSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"testKey": "testValue",
								},
							},
						},
					},
				},
			},
			expectErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().WithSelector(tt.labelSelector)
			if tt.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !tt.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
			if !tt.expectErr {
				if !reflect.DeepEqual(b.BDC.Object.Spec.Selector.MatchLabels, tt.ExpectedBuilder.BDC.Object.Spec.Selector.MatchLabels) {
					t.Fatalf("WithSelector() = %v, want %v", b.BDC.Object.Spec.Selector.MatchLabels, tt.ExpectedBuilder.BDC.Object.Spec.Selector.MatchLabels)
				}
			}
		})
	}
}

func TestBuild(t *testing.T) {
	tests := map[string]struct {
		name        string
		capacity    string
		tagValue    string
		expectedBDC *apis.BlockDeviceClaim
		expectedErr bool
	}{
		"BDC with correct details": {
			name:     "BDC1",
			capacity: "10Ti",
			tagValue: "",
			expectedBDC: &apis.BlockDeviceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "BDC1"},
				Spec: apis.DeviceClaimSpec{
					Resources: apis.DeviceClaimResources{
						Requests: corev1.ResourceList{
							corev1.ResourceName(ndm.ResourceStorage): fakeCapacity("10Ti"),
						},
					},
				},
			},
			expectedErr: false,
		},
		"BDC with correct details, including device pool": {
			name:     "BDC1",
			capacity: "10Ti",
			tagValue: "test",
			expectedBDC: &apis.BlockDeviceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "BDC1"},
				Spec: apis.DeviceClaimSpec{
					Resources: apis.DeviceClaimResources{
						Requests: corev1.ResourceList{
							corev1.ResourceName(ndm.ResourceStorage): fakeCapacity("10Ti"),
						},
					},
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							bdTagKey: "test",
						},
					},
				},
			},
			expectedErr: false,
		},
		"BDC with error": {
			name:        "",
			capacity:    "500Gi",
			tagValue:    "test",
			expectedBDC: nil,
			expectedErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			bdcObjBuilder := NewBuilder().
				WithName(mock.name).
				WithCapacity(mock.capacity)

			if len(mock.tagValue) > 0 {
				bdcObjBuilder.WithBlockDeviceTag(mock.tagValue)
			}

			bdcObj, err := bdcObjBuilder.Build()

			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
			if err == nil && !reflect.DeepEqual(bdcObj.Object, mock.expectedBDC) {
				t.Fatalf("Test %q failed: bdc mismatch", name)
			}
		})
	}
}

func fakeCapacity(capacity string) resource.Quantity {
	q, _ := resource.ParseQuantity(capacity)
	return q
}
