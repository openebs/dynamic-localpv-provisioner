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

	corev1 "k8s.io/api/core/v1"
)

func TestBuilderWithName(t *testing.T) {
	tests := map[string]struct {
		name      string
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with name": {
			name: "NS1",
			builder: &Builder{ns: &Namespace{
				object: &corev1.Namespace{},
			}},
			expectErr: false,
		},
		"Test Builder without name": {
			name: "",
			builder: &Builder{ns: &Namespace{
				object: &corev1.Namespace{},
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

func TestBuildWithBuilderFn(t *testing.T) {
	tests := map[string]struct {
		name        string
		expectedErr bool
	}{
		"Namespace with correct details": {
			name:        "NS1",
			expectedErr: false,
		},
		"PVC with error": {
			name:        "",
			expectedErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			_, err := NewBuilder().WithName(mock.name).Build()
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestAPIObjectWithBuilderFn(t *testing.T) {
	tests := map[string]struct {
		name        string
		expectedErr bool
	}{
		"Namespace with correct details": {
			name:        "NS1",
			expectedErr: false,
		},
		"PVC with error": {
			name:        "",
			expectedErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			_, err := NewBuilder().WithName(mock.name).APIObject()
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
