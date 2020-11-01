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
	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=upgrade

// UpgradeResult contains the desired specifications of an
// upgrade result
type UpgradeResult struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Config UpgradeResultConfig `json:"config"`
	// Tasks are the runtasks that needs to be
	// executed to perform this upgrade
	Tasks  []UpgradeResultTask `json:"tasks"`
	Status UpgradeResultStatus `json:"status"`
}

// UpgradeResultConfig represents the config of UpgradeResult i.e.
// It contains resource details of single unit of upgrade and
// all runtime configuration.
type UpgradeResultConfig struct {
	ResourceDetails
	// data is used to provide some runtime configurations to
	// castemplate engine. Task executor will directly copy these
	// configurations to castemplate engine.
	Data []DataItem `json:"data"`
}

// UpgradeResultStatus represents the current state of UpgradeResult
type UpgradeResultStatus struct {
	// DesiredCount is the total no of resources that
	// needs to be upgraded to a desired version
	DesiredCount int `json:"desiredCount"`
	// ActualCount represents the no of resources
	// that has been successfully upgraded
	ActualCount int `json:"actualCount"`
	// FailedCount represents the no of resources
	// that has failed to upgrade
	FailedCount int `json:"failedCount"`
	// Resource is the resource that needs to
	// be upgraded
	Resource UpgradeResource `json:"resource"`
}

// UpgradeResource represents a resource that needs to
// be upgraded to a desired version
type UpgradeResource struct {
	ResourceDetails
	// PreState represents the state of the resource
	// before upgrade
	PreState ResourceState `json:"preState"`
	// PostState represents the state of the resource
	// after upgrade
	PostState ResourceState `json:"postState"`
	// SubResources are the resources related to
	// this resource which needs to be upgraded
	SubResources []UpgradeSubResource `json:"subResources"`
}

// UpgradeResultTask represents details of a task(runtask) required
// to be executed for upgrading a particular resource
type UpgradeResultTask struct {
	// Name of the task
	Name string `json:"name"`
	// Status is the status of the task which
	// could be successful or failed
	Status string `json:"status"`
	// LastTransitionTime is the last time the status
	// transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is a human readable message
	// indicating details about the task
	Message string `json:"message"`
	// LastError is the last error occurred
	// while executing this task
	LastError string `json:"lastError"`
	// StartTime of the task
	StartTime *metav1.Time `json:"startTime"`
	// EndTime of the task
	EndTime *metav1.Time `json:"endTime"`
	// Retries is the no of times this task
	// has tried to execute
	Retries int `json:"retries"`
}

// UpgradeSubResource represents the details of
// a subresource which needs to be upgraded
type UpgradeSubResource struct {
	ResourceDetails
	// PreState represents the state of the
	// subresource before upgrade
	PreState ResourceState `json:"preState"`
	// PostState represents the state of the
	// subresource after upgrade
	PostState ResourceState `json:"postState"`
}

// ResourceDetails represents the basic details
// of a particular resource
type ResourceDetails struct {
	// Name of the resource
	Name string `json:"name"`
	// Kind is the type of resource i.e.
	// cvr, deployment, ..
	Kind string `json:"kind"`
	// APIVersion of the resource
	APIVersion string `json:"apiVersion"`
	// Namespace of the resource
	Namespace string `json:"namespace"`
	// Generation of resource represents last successful Generation
	// observed by resource controller (ie. - deployment controller).
	// Every time we patched a resource it will assign a new Generation.
	// This is helpful at the time of roll back.
	Generation string `json:"generation"`
}

// String implements Stringer interface
func (rd ResourceDetails) String() string {
	return stringer.Yaml("resource details", rd)
}

// GoString implements GoStringer interface
func (rd ResourceDetails) GoString() string {
	return rd.String()
}

// ResourceState represents the state of a resource
type ResourceState struct {
	// Status of the resource
	Status string `json:"status"`
	// LastTransitionTime is the last time the status
	// transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is a human readable message indicating details about the transition.
	Message string `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=upgrades

// UpgradeResultList is a list of UpgradeResults
type UpgradeResultList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []UpgradeResult `json:"items"`
}

// String implements Stringer interface
func (urList UpgradeResultList) String() string {
	return stringer.Yaml("upgraderesult list", urList)
}

// GoString implements GoStringer interface
func (urList UpgradeResultList) GoString() string {
	return urList.String()
}
