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

	"github.com/ghodss/yaml"
	api_core_v1 "k8s.io/api/core/v1"
)

// PodExec represents the details of pod exec options
type PodExec struct {
	object       *api_core_v1.PodExecOptions
	ignoreErrors bool
	errs         []error
}

// AsAPIPodExec validate and returns PodExecOptions object pointer and error
// depending on ignoreErrors opt and errors
func (p *PodExec) AsAPIPodExec() (*api_core_v1.PodExecOptions, error) {
	err := p.Validate()
	if err != nil && !p.ignoreErrors {
		return nil, err
	}
	return p.object, nil
}

// BuilderForYAMLObject returns a new instance
// of Builder for a given template object
func BuilderForYAMLObject(object []byte) *PodExec {
	p := &PodExec{}
	exec := &api_core_v1.PodExecOptions{}
	err := yaml.Unmarshal(object, exec)
	if err != nil {
		p.errs = append(p.errs, err)
		return p
	}
	p.object = exec
	return p
}

// Validate validates PodExecOptions it mainly checks for container name is present or not and
// commands are present or not.
func (p *PodExec) Validate() error {
	if len(p.errs) != 0 {
		return fmt.Errorf("validation failed: %v", p.errs)
	}
	if len(p.object.Command) == 0 {
		return fmt.Errorf("validation failed: command not provided")
	}
	if p.object.Container == "" {
		return fmt.Errorf("validation failed: container name not provided")
	}
	return nil
}

// BuildOption represents the various build options
// against PodExec operation
type BuildOption func(*PodExec)

// IgnoreErrors is a buildOption that is used ignore errors
func IgnoreErrors() BuildOption {
	return func(p *PodExec) {
		p.ignoreErrors = true
	}
}

// Apply applies all build options in PodExec
func (p *PodExec) Apply(opts ...BuildOption) *PodExec {
	for _, o := range opts {
		o(p)
	}
	return p
}
