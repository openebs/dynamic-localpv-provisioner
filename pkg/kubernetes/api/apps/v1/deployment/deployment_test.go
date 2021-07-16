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
	"testing"

	pts "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/podtemplatespec"
	appsv1 "k8s.io/api/apps/v1"
)

func TestBuilderWithName(t *testing.T) {
	tests := map[string]struct {
		name      string
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with name": {
			name: "PVC1",
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: false,
		},
		"Test Builder without name": {
			name: "",
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithName(mock.name)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
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
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: false,
		},
		"Test Builder without namespace": {
			namespace: "",
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithNamespace(mock.namespace)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
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
			annotations: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: false,
		},
		"Test Builderwithout annotations": {
			annotations: map[string]string{},
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithAnnotations(mock.annotations)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
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
		"Test Builderwith annotations": {
			annotations: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: false,
		},
		"Test Builderwithout annotations": {
			annotations: map[string]string{},
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithAnnotationsNew(mock.annotations)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
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
			labels: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: false,
		},
		"Test Builderwithout labels": {
			labels: map[string]string{},
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithLabels(mock.labels)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
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
		"Test Builderwith labels": {
			labels: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: false,
		},
		"Test Builderwithout labels": {
			labels: map[string]string{},
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithLabelsNew(mock.labels)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithSelectorMatchLabels(t *testing.T) {
	tests := map[string]struct {
		matchLabels map[string]string
		builder     *Builder
		expectErr   bool
	}{
		"Test Builderwith matchLabels": {
			matchLabels: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: false,
		},
		"Test Builderwithout matchLabels": {
			matchLabels: map[string]string{},
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithSelectorMatchLabels(mock.matchLabels)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithSelectorMatchLabelsNew(t *testing.T) {
	tests := map[string]struct {
		matchLabels map[string]string
		builder     *Builder
		expectErr   bool
	}{
		"Test Builderwith matchLabels": {
			matchLabels: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: false,
		},
		"Test Builderwithout matchLabels": {
			matchLabels: map[string]string{},
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithSelectorMatchLabelsNew(mock.matchLabels)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithReplicas(t *testing.T) {
	samplereplicas := int32(3)
	invalidreplicas := int32(-1)
	tests := map[string]struct {
		replicas  *int32
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with replicas": {
			replicas: &samplereplicas,
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: false,
		},
		"Test Builder without replicas": {
			replicas: nil,
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: true,
		},
		"Test Builder with invalid replicas": {
			replicas: &invalidreplicas,
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithReplicas(mock.replicas)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithStrategyType(t *testing.T) {
	tests := map[string]struct {
		strategyType appsv1.DeploymentStrategyType
		builder      *Builder
		expectErr    bool
	}{
		"Test Builder with strategyType": {
			strategyType: appsv1.RecreateDeploymentStrategyType,
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: false,
		},
		"Test Builder without strategyType": {
			strategyType: appsv1.DeploymentStrategyType(""),
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithStrategyType(mock.strategyType)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithPodTemplateSpec(t *testing.T) {
	tests := map[string]struct {
		templateSpec *pts.Builder
		builder      *Builder
		expectErr    bool
	}{
		"Test Builder with templateSpec": {
			templateSpec: pts.NewBuilder(),
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: false,
		},
		"Test Builder without templateSpec": {
			templateSpec: nil,
			builder: &Builder{deployment: &Deploy{
				object: &appsv1.Deployment{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithPodTemplateSpecBuilder(mock.templateSpec)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
