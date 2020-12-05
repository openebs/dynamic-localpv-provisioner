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

package container

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

// fakeAlwaysTrue is a concrete implementation of container Predicate
func fakeAlwaysTrue(d *container) (string, bool) {
	return "fakeAlwaysTrue", true
}

// fakeAlwaysFalse is a concrete implementation of container Predicate
func fakeAlwaysFalse(d *container) (string, bool) {
	return "fakeAlwaysFalse", false
}

func TestPredicateFailedError(t *testing.T) {
	tests := map[string]struct {
		predicateMessage string
		expectedErr      string
	}{
		"always true":  {"fakeAlwaysTrue", "predicatefailed: fakeAlwaysTrue"},
		"always false": {"fakeAlwaysFalse", "predicatefailed: fakeAlwaysFalse"},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			e := predicateFailedError(mock.predicateMessage)
			if e.Error() != mock.expectedErr {
				t.Fatalf(
					"test '%s' failed: expected '%s': actual '%s'",
					name,
					mock.expectedErr,
					e.Error(),
				)
			}
		})
	}
}

func TestNewWithName(t *testing.T) {
	n := "con1"
	c := New(WithName(n))
	if c.Name != n {
		t.Fatalf("test failed: expected name '%s': actual '%s'", n, c.Name)
	}
}

func TestNewWithImage(t *testing.T) {
	i := "openebs.io/m-apiserver:1.0.0"
	c := New(WithImage(i))
	if c.Image != i {
		t.Fatalf("test failed: expected image '%s': actual '%s'", i, c.Image)
	}
}

func TestBuilderBuild(t *testing.T) {
	_, err := NewBuilder().Build()
	if err != nil {
		t.Fatalf("test failed: expected no err: actual '%+v'", err)
	}
}

func TestBuilderValidation(t *testing.T) {
	tests := map[string]struct {
		checks  []Predicate
		isError bool
	}{
		"always true":  {[]Predicate{fakeAlwaysTrue}, false},
		"always false": {[]Predicate{fakeAlwaysFalse}, true},
		"true & false": {[]Predicate{fakeAlwaysTrue, fakeAlwaysFalse}, true},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			_, err := NewBuilder().AddChecks(mock.checks).Build()
			if mock.isError && err == nil {
				t.Fatalf(
					"test '%s' failed: expected error: actual no error",
					name,
				)
			}
			if !mock.isError && err != nil {
				t.Fatalf(
					"test '%s' failed: expected no error: actual error '%+v'",
					name,
					err,
				)
			}
		})
	}
}

func TestBuilderAddChecks(t *testing.T) {
	tests := map[string]struct {
		checks        []Predicate
		expectedCount int
	}{
		"zero": {[]Predicate{}, 0},
		"one":  {[]Predicate{fakeAlwaysTrue}, 1},
		"two":  {[]Predicate{fakeAlwaysTrue, fakeAlwaysFalse}, 2},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().AddChecks(mock.checks)
			if len(b.checks) != mock.expectedCount {
				t.Fatalf(
					"test '%s' failed: expected no of checks '%d': actual '%d'",
					name,
					mock.expectedCount,
					len(b.checks),
				)
			}
		})
	}
}

func TestBuilderWithName(t *testing.T) {
	tests := map[string]struct {
		name         string
		expectedName string
	}{
		"t1": {"nginx", "nginx"},
		"t2": {"maya", "maya"},
		"t3": {"ndm", "ndm"},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			c, _ := NewBuilder().WithName(mock.name).Build()
			if c.Name != mock.expectedName {
				t.Fatalf(
					"test '%s' failed: expected name '%s': actual '%s'",
					name,
					mock.expectedName,
					c.Name,
				)
			}
		})
	}
}

func TestBuilderWithImage(t *testing.T) {
	tests := map[string]struct {
		image         string
		expectedImage string
	}{
		"t1": {"nginx:1.0.0", "nginx:1.0.0"},
		"t2": {"openebs.io/maya:1.0", "openebs.io/maya:1.0"},
		"t3": {"openebs.io/ndm:1.0", "openebs.io/ndm:1.0"},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			c, _ := NewBuilder().WithImage(mock.image).Build()
			if c.Image != mock.expectedImage {
				t.Fatalf(
					"test '%s' failed: expected image '%s': actual '%s'",
					name,
					mock.expectedImage,
					c.Image,
				)
			}
		})
	}
}

func TestBuilderWithCommandNew(t *testing.T) {
	tests := map[string]struct {
		cmd         []string
		expectedCmd []string
	}{
		"t1": {
			[]string{"kubectl", "get", "po"},
			[]string{"kubectl", "get", "po"},
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			c, _ := NewBuilder().WithCommandNew(mock.cmd).Build()
			if !reflect.DeepEqual(c.Command, mock.expectedCmd) {
				t.Fatalf(
					"test '%s' failed: expected command '%q': actual '%q'",
					name,
					mock.expectedCmd,
					c.Command,
				)
			}
		})
	}
}

func TestBuilderWithArgumentsNew(t *testing.T) {
	tests := map[string]struct {
		args         []string
		expectedArgs []string
	}{
		"t1": {
			[]string{"-jsonpath", "metadata.name"},
			[]string{"-jsonpath", "metadata.name"},
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			c, _ := NewBuilder().WithArgumentsNew(mock.args).Build()
			if !reflect.DeepEqual(c.Args, mock.expectedArgs) {
				t.Fatalf(
					"test '%s' failed: expected arguments '%q': actual '%q'",
					name,
					mock.expectedArgs,
					c.Args,
				)
			}
		})
	}
}

func TestBuilderWithPrivilegedSecurityContext(t *testing.T) {
	tests := map[string]struct {
		secCont   bool
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with templateSpec": {
			secCont: true,
			builder: &Builder{con: &container{
				corev1.Container{},
			}},
			expectErr: false,
		},
		"Test Builder without templateSpec": {
			secCont: false,
			builder: &Builder{con: &container{
				corev1.Container{},
			}},
			expectErr: false,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithPrivilegedSecurityContext(&mock.secCont)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithEnvsNew(t *testing.T) {
	tests := map[string]struct {
		envList   []corev1.EnvVar
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with envList": {
			envList: []corev1.EnvVar{
				corev1.EnvVar{},
			},
			builder: &Builder{con: &container{
				corev1.Container{},
			}},
			expectErr: false,
		},
		"Test Builder without envList": {
			envList: nil,
			builder: &Builder{con: &container{
				corev1.Container{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithEnvsNew(mock.envList)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithPortsNew(t *testing.T) {
	tests := map[string]struct {
		portList  []corev1.ContainerPort
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with portList": {
			portList: []corev1.ContainerPort{
				corev1.ContainerPort{},
			},
			builder: &Builder{con: &container{
				corev1.Container{},
			}},
			expectErr: false,
		},
		"Test Builder without portList": {
			portList: nil,
			builder: &Builder{con: &container{
				corev1.Container{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithPortsNew(mock.portList)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithResources(t *testing.T) {
	tests := map[string]struct {
		requirements *corev1.ResourceRequirements
		builder      *Builder
		expectErr    bool
	}{
		"Test Builder with resource requirements": {
			requirements: &corev1.ResourceRequirements{},
			builder: &Builder{con: &container{
				corev1.Container{},
			}},
			expectErr: false,
		},
		"Test Builder without resource requirements": {
			requirements: nil,
			builder: &Builder{con: &container{
				corev1.Container{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithResources(mock.requirements)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithVolumeMountsNew(t *testing.T) {
	tests := map[string]struct {
		mounts    []corev1.VolumeMount
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with volume mounts": {
			mounts: []corev1.VolumeMount{
				corev1.VolumeMount{},
			},
			builder: &Builder{con: &container{
				corev1.Container{},
			}},
			expectErr: false,
		},
		"Test Builder without volume mounts": {
			mounts: nil,
			builder: &Builder{con: &container{
				corev1.Container{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithVolumeMountsNew(mock.mounts)
			if mock.expectErr && len(b.errors) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errors) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
