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
	"github.com/pkg/errors"
	extnv1beta1 "k8s.io/api/extensions/v1beta1"
)

// Predicate abstracts conditional logic w.r.t the deployment instance
//
// NOTE:
// predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*Deploy) bool

// Deploy is the wrapper over k8s deployment object
type Deploy struct {
	// kubernetes deployment instance
	object *extnv1beta1.Deployment
}

// Builder enables building an instance of
// deployment
type Builder struct {
	deployment *Deploy     // kubernetes deployment instance
	checks     []Predicate // predicate list for deploy
	errors     []error
}

// PredicateName type is wrapper over string.
// It is used to refer predicate and status msg.
type PredicateName string

const (
	// PredicateProgressDeadlineExceeded refer to
	// predicate IsProgressDeadlineExceeded.
	PredicateProgressDeadlineExceeded PredicateName = "ProgressDeadlineExceeded"
	// PredicateNotSpecSynced refer to predicate IsNotSpecSynced
	PredicateNotSpecSynced PredicateName = "NotSpecSynced"
	// PredicateOlderReplicaActive refer to predicate IsOlderReplicaActive
	PredicateOlderReplicaActive PredicateName = "OlderReplicaActive"
	// PredicateTerminationInProgress refer to predicate IsTerminationInProgress
	PredicateTerminationInProgress PredicateName = "TerminationInProgress"
	// PredicateUpdateInProgress refer to predicate IsUpdateInProgress.
	PredicateUpdateInProgress PredicateName = "UpdateInProgress"
)

// String implements the stringer interface
func (d *Deploy) String() string {
	return stringer.Yaml("deployment", d.object)
}

// GoString implements the goStringer interface
func (d *Deploy) GoString() string {
	return d.String()
}

// NewBuilder returns a new instance of builder meant for deployment
func NewBuilder() *Builder {
	return &Builder{
		deployment: &Deploy{
			object: &extnv1beta1.Deployment{},
		},
	}
}

// NewBuilderForAPIObject returns a new instance of builder
// for a given deployment object
func NewBuilderForAPIObject(deployment *extnv1beta1.Deployment) *Builder {
	b := NewBuilder()
	if deployment != nil {
		b.deployment.object = deployment
	} else {
		b.errors = append(b.errors,
			errors.New("nil deployment object given to get builder instance"))
	}
	return b
}

// Build returns a deployment instance
func (b *Builder) Build() (*Deploy, error) {
	err := b.validate()
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to build a deployment instance: %s",
			b.deployment.object)
	}
	return b.deployment, nil
}

func (b *Builder) validate() error {
	if len(b.errors) != 0 {
		return errors.Errorf("failed to validate: build errors were found: %v", b.errors)
	}
	return nil
}

// AddCheck adds the predicate as a condition to be validated
// against the deployment instance
func (b *Builder) AddCheck(p Predicate) *Builder {
	b.checks = append(b.checks, p)
	return b
}

// AddChecks adds the provided predicates as conditions to be
// validated against the deployment instance
func (b *Builder) AddChecks(p []Predicate) *Builder {
	for _, check := range p {
		b.AddCheck(check)
	}
	return b
}

// IsRollout range over rolloutChecks map and check status of each predicate
// also it generates status message from rolloutStatuses using predicate key
func (d *Deploy) IsRollout() (PredicateName, bool) {
	for pk, p := range RolloutChecks {
		if p(d) {
			return pk, false
		}
	}
	return "", true
}

// FailedRollout returns rollout status message for fail condition
func (d *Deploy) FailedRollout(name PredicateName) *RolloutOutput {
	return &RolloutOutput{
		Message:     RolloutStatuses[name](d),
		IsRolledout: false,
	}
}

// SuccessRollout returns rollout status message for success condition
func (d *Deploy) SuccessRollout() *RolloutOutput {
	return &RolloutOutput{
		Message:     "deployment successfully rolled out",
		IsRolledout: true,
	}
}

// RolloutStatus returns rollout message of deployment instance
func (d *Deploy) RolloutStatus() (op *RolloutOutput, err error) {
	pk, ok := d.IsRollout()
	if ok {
		return d.SuccessRollout(), nil
	}
	return d.FailedRollout(pk), nil
}

// RolloutStatusRaw returns rollout message of deployment instance
// in byte format
func (d *Deploy) RolloutStatusRaw() (op []byte, err error) {
	message, err := d.RolloutStatus()
	if err != nil {
		return nil, err
	}
	return NewRollout(
		withOutputObject(message)).
		Raw()
}

// IsProgressDeadlineExceeded is used to check update is timed out or not.
// If `Progressing` condition's reason is `ProgressDeadlineExceeded` then
// it is not rolled out.
func IsProgressDeadlineExceeded() Predicate {
	return func(d *Deploy) bool {
		return d.IsProgressDeadlineExceeded()
	}
}

// IsProgressDeadlineExceeded is used to check update is timed out or not.
// If `Progressing` condition's reason is `ProgressDeadlineExceeded` then
// it is not rolled out.
func (d *Deploy) IsProgressDeadlineExceeded() bool {
	for _, cond := range d.object.Status.Conditions {
		if cond.Type == extnv1beta1.DeploymentProgressing &&
			cond.Reason == "ProgressDeadlineExceeded" {
			return true
		}
	}
	return false
}

// IsOlderReplicaActive check if older replica's are still active or not
// if Status.UpdatedReplicas < *Spec.Replicas then some of the replicas
// are updated and some of them are not.
func IsOlderReplicaActive() Predicate {
	return func(d *Deploy) bool {
		return d.IsOlderReplicaActive()
	}
}

// IsOlderReplicaActive check if older replica's are still active or not
// if Status.UpdatedReplicas < *Spec.Replicas then some of the replicas
// are updated and some of them are not.
func (d *Deploy) IsOlderReplicaActive() bool {
	return d.object.Spec.Replicas != nil && d.object.Status.UpdatedReplicas < *d.object.Spec.Replicas
}

// IsTerminationInProgress checks for older replicas are waiting to
// terminate or not. If Status.Replicas > Status.UpdatedReplicas then
// some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to
// come into running state then terminate.
func IsTerminationInProgress() Predicate {
	return func(d *Deploy) bool {
		return d.IsTerminationInProgress()
	}
}

// IsTerminationInProgress checks for older replicas are waiting to
// terminate or not. If Status.Replicas > Status.UpdatedReplicas then
// some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to
// come into running state then terminate.
func (d *Deploy) IsTerminationInProgress() bool {
	return d.object.Status.Replicas > d.object.Status.UpdatedReplicas
}

// IsUpdateInProgress Checks if all the replicas are updated or not. If
// Status.AvailableReplicas < Status.UpdatedReplicas then all the older
// replicas are not there but there are less number of availableReplicas
func IsUpdateInProgress() Predicate {
	return func(d *Deploy) bool {
		return d.IsUpdateInProgress()
	}
}

// IsUpdateInProgress Checks if all the replicas are updated or not. If
// Status.AvailableReplicas < Status.UpdatedReplicas then all the older
// replicas are not there but there are less number of availableReplicas
func (d *Deploy) IsUpdateInProgress() bool {
	return d.object.Status.AvailableReplicas < d.object.Status.UpdatedReplicas
}

// IsNotSyncSpec compare generation in status and spec and check if deployment
// spec is synced or not. If Generation <= Status.ObservedGeneration then
// deployment spec is not updated yet.
func IsNotSyncSpec() Predicate {
	return func(d *Deploy) bool {
		return d.IsNotSyncSpec()
	}
}

// IsNotSyncSpec compare generation in status and spec and check if deployment
// spec is synced or not. If Generation <= Status.ObservedGeneration then
// deployment spec is not updated yet.
func (d *Deploy) IsNotSyncSpec() bool {
	return d.object.Generation > d.object.Status.ObservedGeneration
}
