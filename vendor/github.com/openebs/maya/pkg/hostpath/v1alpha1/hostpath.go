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
	"fmt"
	errors "github.com/pkg/errors"
	"path/filepath"
	"strings"
)

// HostPath is a wrapper over a given path string
// It provides build, validations and other common
// logic to be used by various feature specific callers.
type HostPath string

// Predicate defines an abstraction
// to determine conditional checks
// against the provided HostPath instance
type Predicate func(HostPath) bool

// isValidPath is a predicate that determines
// the hostpath is a well formed path
func isValidPath() Predicate {
	return func(hp HostPath) bool {
		//Validate that path is well formed.
		//_, err := filepath.Abs(string(hp))
		//if err != nil {
		//	return false
		//}
		//return true
		return filepath.IsAbs(string(hp))
	}
}

// IsNonRoot is a predicate that determines
// the hostpath is not directly under /
func IsNonRoot() Predicate {
	return func(hp HostPath) bool {
		//Validate that hostpath is not directly under "/"
		path := strings.TrimSuffix(string(hp), "/")
		parentDir, subDir := filepath.Split(path)
		parentDir = strings.TrimSuffix(parentDir, "/")
		subDir = strings.TrimSuffix(subDir, "/")
		if parentDir == "" || subDir == "" {
			// it covers the `/` case
			return false
		}
		return true
	}
}

// Builder provides utility functions
// on the HostPath to extract different information
type Builder struct {
	path   HostPath
	checks map[*Predicate]string
}

// NewBuilder builds a new Builder
func NewBuilder() *Builder {
	return &Builder{
		path:   "",
		checks: make(map[*Predicate]string),
	}
}

// WithPath initializes the Builder with the path on
// which it has to operate
func (b *Builder) WithPath(path string) *Builder {
	b.path = HostPath(path)
	return b
}

// WithPathJoin initializes the Builder with the path
// by joining the arguments
func (b *Builder) WithPathJoin(basePath, relPath string) *Builder {
	b.path = HostPath(filepath.Join(basePath, relPath))
	return b
}

// WithChecks adds the list of Predicates to
// the list of checks to be performed by the builder.
func (b *Builder) WithChecks(checks ...Predicate) *Builder {
	for _, check := range checks {
		b.WithCheck(check)
	}
	return b
}

// WithCheck add a single Predicate to the
// list of checks to be performed by the builder
func (b *Builder) WithCheck(check Predicate) *Builder {
	return b.WithCheckf(check, "")
}

// WithCheckf adds a single Predicate to the
// list of checks to be performed by the builder, along
// with the message to be returned when the check fails
func (b *Builder) WithCheckf(check Predicate, msg string, args ...interface{}) *Builder {
	b.checks[&check] = fmt.Sprintf(msg, args...)
	return b
}

// Validate runs through all the checks set on the
// builder and returns an error when any of the checks fails.
func (b *Builder) Validate() error {
	if b.path == "" {
		return errors.New("failed to validate: missing host path")
	}

	b = b.WithCheckf(isValidPath(), "invalid path %v", b.path)

	failedValidations := []string{}
	for check, msg := range b.checks {
		if !(*check)(b.path) {
			failedValidations = append(failedValidations, msg)
		}
	}

	if len(failedValidations) > 0 {
		return errors.Errorf("host path validation failed: %v", failedValidations)
	}
	return nil
}

// ValidateAndBuild is utility function that returns the
// path if all validation checks have passed
func (b *Builder) ValidateAndBuild() (string, error) {
	err := b.Validate()
	if err != nil {
		return "", err
	}
	return string(b.path), nil
}

// ExtractSubPath is utility function to split directory from path
func (b *Builder) ExtractSubPath() (string, string, error) {
	err := b.Validate()
	if err != nil {
		return "", "", err
	}
	parentDir, volumeDir := filepath.Split(string(b.path))
	parentDir = strings.TrimSuffix(parentDir, "/")
	volumeDir = strings.TrimSuffix(volumeDir, "/")
	return parentDir, volumeDir, nil
}
