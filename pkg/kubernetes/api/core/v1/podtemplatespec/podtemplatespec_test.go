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

package podtemplatespec

import (
	"testing"

	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/container"
	corev1 "k8s.io/api/core/v1"
)

func TestBuilderWithName(t *testing.T) {
	tests := map[string]struct {
		name      string
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with name": {
			name: "PVC1",
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: false,
		},
		"Test Builder without name": {
			name: "",
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
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

func TestBuilderWithNamespace(t *testing.T) {
	tests := map[string]struct {
		namespace string
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with namespace": {
			namespace: "PVC1",
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: false,
		},
		"Test Builder without namespace": {
			namespace: "",
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: true,
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
		"Test Builder with annotations": {
			annotations: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: false,
		},
		"Test Builder without annotations": {
			annotations: map[string]string{},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
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

func TestBuildWithAnnotationsNew(t *testing.T) {
	tests := map[string]struct {
		annotations map[string]string
		builder     *Builder
		expectErr   bool
	}{
		"Test Builder with annotations": {
			annotations: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: false,
		},
		"Test Builder without annotations": {
			annotations: map[string]string{},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithAnnotationsNew(mock.annotations)
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
		"Test Builder with labels": {
			labels: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: false,
		},
		"Test Builder without labels": {
			labels: map[string]string{},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
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
		"Test Builder with labels": {
			labels: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: false,
		},
		"Test Builder without labels": {
			labels: map[string]string{},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
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

func TestBuildWithAffinity(t *testing.T) {
	tests := map[string]struct {
		affinity  *corev1.Affinity
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with affinity": {
			affinity: &corev1.Affinity{},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: false,
		},
		"Test Builder without affinity": {
			affinity: nil,
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithAffinity(mock.affinity)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithContainerBuilders(t *testing.T) {
	tests := map[string]struct {
		conBuilders []*container.Builder
		builder     *Builder
		expectErr   bool
	}{
		"Test Builder with containerBuilders": {
			conBuilders: []*container.Builder{
				container.NewBuilder(),
			},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: false,
		},
		"Test Builder without containerBuilders": {
			conBuilders: nil,
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithContainerBuilders(mock.conBuilders...)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithContainerBuildersNew(t *testing.T) {
	tests := map[string]struct {
		conBuilders []*container.Builder
		builder     *Builder
		expectErr   bool
	}{
		"Test Builder with containerBuilders": {
			conBuilders: []*container.Builder{
				container.NewBuilder(),
			},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: false,
		},
		"Test Builder without containerBuilders": {
			conBuilders: nil,
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithContainerBuildersNew(mock.conBuilders...)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithTolerations(t *testing.T) {
	tests := map[string]struct {
		tolerations []corev1.Toleration
		builder     *Builder
		expectErr   bool
	}{
		"Test Builder with tolerations": {
			tolerations: []corev1.Toleration{
				corev1.Toleration{},
			},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: false,
		},
		"Test Builder without tolerations": {
			tolerations: []corev1.Toleration{},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithTolerations(mock.tolerations...)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithTolerationsNew(t *testing.T) {
	tests := map[string]struct {
		tolerations []corev1.Toleration
		builder     *Builder
		expectErr   bool
	}{
		"Test Builder with tolerations": {
			tolerations: []corev1.Toleration{
				corev1.Toleration{},
			},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: false,
		},
		"Test Builder without tolerations": {
			tolerations: []corev1.Toleration{},
			builder: &Builder{podtemplatespec: &PodTemplateSpec{
				Object: &corev1.PodTemplateSpec{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithTolerationsNew(mock.tolerations...)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
