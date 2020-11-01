/*
Copyright 2018 The OpenEBS Authors.

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
	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=castemplate

// CASTemplate describes a Container Attached Storage template that is used
// to provision a CAS volume
type CASTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CASTemplateSpec `json:"spec"`
}

// CASTemplateSpec is the specifications for a CASTemplate resource
type CASTemplateSpec struct {
	// Defaults are a list of default configurations that may be applied
	// during execution of CAS template
	Defaults []Config `json:"defaultConfig"`
	// TaskNamespace is the namespace where the tasks are expected to be found
	TaskNamespace string `json:"taskNamespace"`
	// RunTasks refers to a set of tasks to be run
	RunTasks RunTasks `json:"run"`
	// OutputTask is the task that has the CAS template result's output
	// format
	OutputTask string `json:"output"`
	// Fallback is the CASTemplate to fallback to in-case of specific failures
	// e.g. VersionMismatchError, etc
	Fallback string `json:"fallback"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=castemplates

// CASTemplateList is a list of CASTemplate resources
type CASTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CASTemplate `json:"items"`
}

// Config holds a configuration element
//
// For example, it can represent a config property of a CAS volume
type Config struct {
	// Name of the config
	Name string `json:"name"`
	// Enabled flags if this config is enabled or disabled;
	// true indicates enabled while false indicates disabled
	Enabled string `json:"enabled"`
	// Value represents any specific value that is applicable
	// to this config
	Value string `json:"value"`
	// Data represents an arbitrary map of key value pairs
	Data map[string]string `json:"data"`
}

// RunTasks contains fields to run a set of
// runtasks referred to by their names
type RunTasks struct {
	// Items is a set of order-ed tasks referred to by their names
	Tasks []string `json:"tasks"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=runtask

// RunTask forms the specifications that deal with running a CAS template
// engine based task
type RunTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RunTaskSpec `json:"spec"`
}

// RunTaskSpec is the specifications of a RunTask resource
type RunTaskSpec struct {
	// Meta is the meta specifications to run this task
	Meta string `json:"meta"`
	// Task is the resource to be operated via this task
	Task string `json:"task"`
	// PostRun is a set of go template functions that is run
	// against the result of this task's execution. In other words, this
	// is run post the task execution.
	PostRun string `json:"post"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=runtasks

// RunTaskList is a list of RunTask resources
type RunTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RunTask `json:"items"`
}

// String implements Stringer interface
func (c Config) String() string {
	return stringer.Yaml("cas config", c)
}

// GoString implements GoStringer interface
func (c Config) GoString() string {
	return c.String()
}

// String implements Stringer interface
func (r *RunTask) String() string {
	return stringer.Yaml("runtask", r)
}

// GoString implements GoStringer interface
func (r *RunTask) GoString() string {
	return r.String()
}

// String implements Stringer interface
func (c *CASTemplate) String() string {
	return stringer.Yaml("castemplate", c)
}

// GoString implements GoStringer interface
func (c *CASTemplate) GoString() string {
	return c.String()
}
