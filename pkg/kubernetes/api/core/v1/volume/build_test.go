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

package volume

import (
	"reflect"
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
			name: "vol1",
			builder: &Builder{volume: &Volume{
				object: &corev1.Volume{},
			}},
			expectErr: false,
		},
		"Test Builder without name": {
			name: "",
			builder: &Builder{volume: &Volume{
				object: &corev1.Volume{},
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

func TestBuildWithHostDirectory(t *testing.T) {
	tests := map[string]struct {
		path      string
		builder   *Builder
		expectErr bool
	}{
		"Test Builderwith hostpath": {
			path: "/var/openebs/local",
			builder: &Builder{volume: &Volume{
				object: &corev1.Volume{},
			}},
			expectErr: false,
		},
		"Test Builderwithout hostpath": {
			path: "",
			builder: &Builder{volume: &Volume{
				object: &corev1.Volume{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithHostDirectory(mock.path)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildHostPathVolume(t *testing.T) {

	tests := map[string]struct {
		name        string
		path        string
		expectedVol *corev1.Volume
		expectedErr bool
	}{
		"Hostpath Volume with correct details": {
			name: "PV1",
			path: "/var/openebs/local/PV1",
			expectedVol: &corev1.Volume{
				Name: "PV1",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/var/openebs/local/PV1",
					},
				},
			},
			expectedErr: false,
		},
		"Hostpath PV with error": {
			name:        "",
			path:        "",
			expectedVol: nil,
			expectedErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			volObj, err := NewBuilder().
				WithName(mock.name).
				WithHostDirectory(mock.path).
				Build()
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
			if !reflect.DeepEqual(volObj, mock.expectedVol) {
				t.Fatalf("Test %q failed: volume mismatch", name)
			}
		})
	}
}

func TestBuilerWithdHostPathandType(t *testing.T) {
	sampledirtype := corev1.HostPathDirectoryOrCreate
	tests := map[string]struct {
		path        string
		dirtype     *corev1.HostPathType
		expectedErr bool
	}{
		"Hostpath Volume with correct type": {
			path:        "/var/openebs/local/PV1",
			dirtype:     &sampledirtype,
			expectedErr: false,
		},
		"Hostpath Volume without type": {
			path:        "/var/openebs/local/PV1",
			dirtype:     nil,
			expectedErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().
				WithHostPathAndType(mock.path, mock.dirtype)
			if mock.expectedErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilerWithEmptyDir(t *testing.T) {
	tests := map[string]struct {
		path        string
		emptydir    *corev1.EmptyDirVolumeSource
		expectedErr bool
	}{
		"Volume with empty dir": {
			path:        "/var/openebs/local/PV1",
			emptydir:    &corev1.EmptyDirVolumeSource{},
			expectedErr: false,
		},
		"Volume without empty dir": {
			path:        "/var/openebs/local/PV1",
			emptydir:    nil,
			expectedErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().
				WithEmptyDir(mock.emptydir)
			if mock.expectedErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
