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

package debug

import (
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// Client to connect to injection API server.
type Client struct {
	BaseURL    *url.URL
	httpClient *http.Client
}

const (
	// Inject is used to inject errors.
	Inject = "true"
	// Eject is used to eject errors.
	Eject = "false"
)

// EI is a global object which is used to decide whether the error is injected or not.
var EI = &ErrorInjection{}

// ErrorInjection schema to inject errors.
type ErrorInjection struct {
	CSPIError       CSPIErrorInjection       `json:"cspiError"`
	CSPCError       CSPCErrorInjection       `json:"cspcError"`
	DeploymentError DeploymentErrorInjection `json:"deploymentError"`
	ZFSError        ZFSErrorInjection        `json:"zfsError"`
	CVRError        CVRErrorInjection        `json:"cvrError"`
}

// CSPIErrorInjection is used to inject errors for CSPI related operations.
type CSPIErrorInjection struct {
	CRUDErrorInjection CRUDErrorInjection       `json:"crudErrorInjection"`
	ErrorPercentage    ErrorPercentageThreshold `json:"errorPercentage"`
}

// CSPCErrorInjection is used to inject errors for CSPC related operations.
type CSPCErrorInjection struct {
	CRUDErrorInjection CRUDErrorInjection       `json:"crudErrorInjection"`
	ErrorPercentage    ErrorPercentageThreshold `json:"errorPercentage"`
}

// DeploymentErrorInjection is used to inject errors for CSPC related operations.
type DeploymentErrorInjection struct {
	CRUDErrorInjection CRUDErrorInjection       `json:"crudErrorInjection"`
	ErrorPercentage    ErrorPercentageThreshold `json:"errorPercentage"`
}

// ZFSErrorInjection is used to inject errors for Volume Replica related
// operations.
type ZFSErrorInjection struct {
	CRUDErrorInjection CRUDErrorInjection       `json:"crudErrorInjection"`
	ErrorPercentage    ErrorPercentageThreshold `json:"errorPercentage"`
}

// CVRErrorInjection is used to inject errors in API calls for Volume Replica
// related operations
type CVRErrorInjection struct {
	CRUDErrorInjection CRUDErrorInjection       `json:"crudErrorInjection"`
	ErrorPercentage    ErrorPercentageThreshold `json:"errorPercentage"`
}

// CRUDErrorInjection is used to inject CRUD errors.
type CRUDErrorInjection struct {
	InjectDeleteCollectionError string `json:"injectDeleteCollectionError"`
	InjectDeleteError           string `json:"injectDeleteError"`
	InjectListError             string `json:"injectListError"`
	InjectGetError              string `json:"injectGetError"`
	InjectCreateError           string `json:"injectCreateError"`
	InjectUpdateError           string `json:"injectUpdateError"`
	InjectPatchError            string `json:"injectPatchError"`
}

// ErrorPercentageThreshold is the threshold value above which the error will not be injected.
type ErrorPercentageThreshold struct {
	Threshold int `json:"threshold"`
}

// GetRandomErrorPercentage returns an error percentage value.
// If the returned error percentage is greater then the error percentage threshold then
// error will not be injected.
func GetRandomErrorPercentage() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(101)
}

// NewErrorInjection returns a new ErrorInjection object.
func NewErrorInjection() *ErrorInjection {
	EI = &ErrorInjection{}
	return EI
}

// WithCSPIThreshold injects CSPI error depending on passed error threshold value.
func (ei *ErrorInjection) WithCSPIThreshold(threshold int) *ErrorInjection {
	ei.CSPIError.ErrorPercentage.Threshold = threshold
	if GetRandomErrorPercentage() > threshold {
		ei.WithCSPIDeleteCollectionError(Inject).
			WithCSPICreateError(Inject).
			WithCSPIDeleteError(Inject).
			WithCSPIGetError(Inject).
			WithCSPIListError(Inject).
			WithCSPIPatchError(Inject).
			WithCSPIUpdateError(Inject)
	}
	return ei
}

// WithCSPCThreshold injects CSPC error depending on passed error threshold value.
func (ei *ErrorInjection) WithCSPCThreshold(threshold int) *ErrorInjection {
	ei.CSPCError.ErrorPercentage.Threshold = threshold
	if GetRandomErrorPercentage() > threshold {
		ei.WithCSPCDeleteCollectionError(Inject).
			WithCSPCCreateError(Inject).
			WithCSPCDeleteError(Inject).
			WithCSPCGetError(Inject).
			WithCSPCListError(Inject).
			WithCSPCPatchError(Inject).
			WithCSPCUpdateError(Inject)
	}
	return ei
}

// WithDeploymentThreshold injects Deployment error depending on passed error threshold value.
func (ei *ErrorInjection) WithDeploymentThreshold(threshold int) *ErrorInjection {
	ei.DeploymentError.ErrorPercentage.Threshold = threshold
	if GetRandomErrorPercentage() > threshold {
		ei.WithDeploymentDeleteCollectionError(Inject).
			WithDeploymentCreateError(Inject).
			WithDeploymentDeleteError(Inject).
			WithDeploymentGetError(Inject).
			WithDeploymentListError(Inject).
			WithDeploymentPatchError(Inject).
			WithDeploymentUpdateError(Inject)
	}
	return ei
}

// WithZFSThreshold injects ZFS errors depending on passed error threshold value
func (ei *ErrorInjection) WithZFSThreshold(threshold int) *ErrorInjection {
	ei.ZFSError.ErrorPercentage.Threshold = threshold
	if GetRandomErrorPercentage() > threshold {
		ei.WithZFSCreateError(Inject).
			WithZFSDeleteError(Inject).
			WithZFSGetError(Inject)
	}
	return ei
}

// WithCVRThreshold injects CVR error depending on passed error threshold value
func (ei *ErrorInjection) WithCVRThreshold(threshold int) *ErrorInjection {
	ei.CVRError.ErrorPercentage.Threshold = threshold
	if GetRandomErrorPercentage() > threshold {
		ei.WithCVRCreateError(Inject).
			WithCVRDeleteError(Inject).
			WithCVRGetError(Inject).
			WithCVRUpdateError(Inject)
	}
	return ei
}

// WithCSPCDeleteCollectionError injects/ejects  CSPC delete collection error.
func (ei *ErrorInjection) WithCSPCDeleteCollectionError(ejectOrInject string) *ErrorInjection {
	ei.CSPCError.CRUDErrorInjection.InjectDeleteCollectionError = ejectOrInject
	return ei
}

// WithCSPCDeleteError injects/ejects  CSPC delete error.
func (ei *ErrorInjection) WithCSPCDeleteError(ejectOrInject string) *ErrorInjection {
	ei.CSPCError.CRUDErrorInjection.InjectDeleteError = ejectOrInject
	return ei
}

// WithCSPCListError injects/ejects  CSPC list error.
func (ei *ErrorInjection) WithCSPCListError(ejectOrInject string) *ErrorInjection {
	ei.CSPCError.CRUDErrorInjection.InjectListError = ejectOrInject
	return ei
}

// WithCSPCGetError injects/ejects  CSPC get error.
func (ei *ErrorInjection) WithCSPCGetError(ejectOrInject string) *ErrorInjection {
	ei.CSPCError.CRUDErrorInjection.InjectGetError = ejectOrInject
	return ei
}

// WithCSPCCreateError injects/ejects  CSPC create error.
func (ei *ErrorInjection) WithCSPCCreateError(ejectOrInject string) *ErrorInjection {
	ei.CSPCError.CRUDErrorInjection.InjectCreateError = ejectOrInject
	return ei
}

// WithCSPCUpdateError injects/ejects  CSPC update error.
func (ei *ErrorInjection) WithCSPCUpdateError(ejectOrInject string) *ErrorInjection {
	ei.CSPCError.CRUDErrorInjection.InjectUpdateError = ejectOrInject
	return ei
}

// WithCSPCPatchError injects/ejects  CSPC patch error.
func (ei *ErrorInjection) WithCSPCPatchError(ejectOrInject string) *ErrorInjection {
	ei.CSPCError.CRUDErrorInjection.InjectPatchError = ejectOrInject
	return ei
}

// WithCSPIDeleteCollectionError injects/ejects  CSPI delete collection error.
func (ei *ErrorInjection) WithCSPIDeleteCollectionError(ejectOrInject string) *ErrorInjection {
	ei.CSPIError.CRUDErrorInjection.InjectDeleteCollectionError = ejectOrInject
	return ei
}

// WithCSPIDeleteError injects/ejects  CSPI delete error.
func (ei *ErrorInjection) WithCSPIDeleteError(ejectOrInject string) *ErrorInjection {
	ei.CSPIError.CRUDErrorInjection.InjectDeleteError = ejectOrInject
	return ei
}

// WithCSPIListError injects/ejects  CSPI list error.
func (ei *ErrorInjection) WithCSPIListError(ejectOrInject string) *ErrorInjection {
	ei.CSPIError.CRUDErrorInjection.InjectListError = ejectOrInject
	return ei
}

// WithCSPIGetError injects/ejects  CSPI get error.
func (ei *ErrorInjection) WithCSPIGetError(ejectOrInject string) *ErrorInjection {
	ei.CSPIError.CRUDErrorInjection.InjectGetError = ejectOrInject
	return ei
}

// WithCSPICreateError injects/ejects  CSPI create error.
func (ei *ErrorInjection) WithCSPICreateError(ejectOrInject string) *ErrorInjection {
	ei.CSPIError.CRUDErrorInjection.InjectCreateError = ejectOrInject
	return ei
}

// WithCSPIUpdateError injects/ejects  CSPI update error.
func (ei *ErrorInjection) WithCSPIUpdateError(ejectOrInject string) *ErrorInjection {
	ei.CSPIError.CRUDErrorInjection.InjectUpdateError = ejectOrInject
	return ei
}

// WithCSPIPatchError injects/ejects  CSPI patch error.
func (ei *ErrorInjection) WithCSPIPatchError(ejectOrInject string) *ErrorInjection {
	ei.CSPIError.CRUDErrorInjection.InjectPatchError = ejectOrInject
	return ei
}

// WithDeploymentDeleteCollectionError injects/ejects  Deployment delete collection error.
func (ei *ErrorInjection) WithDeploymentDeleteCollectionError(ejectOrInject string) *ErrorInjection {
	ei.DeploymentError.CRUDErrorInjection.InjectDeleteCollectionError = ejectOrInject
	return ei
}

// WithDeploymentDeleteError injects/ejects  Deployment delete error.
func (ei *ErrorInjection) WithDeploymentDeleteError(ejectOrInject string) *ErrorInjection {
	ei.DeploymentError.CRUDErrorInjection.InjectDeleteError = ejectOrInject
	return ei
}

// WithDeploymentListError injects/ejects  Deployment delete error.
func (ei *ErrorInjection) WithDeploymentListError(ejectOrInject string) *ErrorInjection {
	ei.DeploymentError.CRUDErrorInjection.InjectListError = ejectOrInject
	return ei
}

// WithDeploymentGetError injects/ejects  Deployment get error.
func (ei *ErrorInjection) WithDeploymentGetError(ejectOrInject string) *ErrorInjection {
	ei.DeploymentError.CRUDErrorInjection.InjectGetError = ejectOrInject
	return ei
}

// WithDeploymentCreateError injects/ejects  Deployment create error.
func (ei *ErrorInjection) WithDeploymentCreateError(ejectOrInject string) *ErrorInjection {
	ei.DeploymentError.CRUDErrorInjection.InjectCreateError = ejectOrInject
	return ei
}

// WithDeploymentUpdateError injects/ejects  Deployment update error.
func (ei *ErrorInjection) WithDeploymentUpdateError(ejectOrInject string) *ErrorInjection {
	ei.DeploymentError.CRUDErrorInjection.InjectUpdateError = ejectOrInject
	return ei
}

// WithDeploymentPatchError injects/ejects  Deployment patch error.
func (ei *ErrorInjection) WithDeploymentPatchError(ejectOrInject string) *ErrorInjection {
	ei.DeploymentError.CRUDErrorInjection.InjectPatchError = ejectOrInject
	return ei
}

// WithZFSGetError injects/ejects ZFS get error.
func (ei *ErrorInjection) WithZFSGetError(ejectOrInject string) *ErrorInjection {
	ei.ZFSError.CRUDErrorInjection.InjectGetError = ejectOrInject
	return ei
}

// WithZFSCreateError injects/ejects ZFS create error.
func (ei *ErrorInjection) WithZFSCreateError(ejectOrInject string) *ErrorInjection {
	ei.ZFSError.CRUDErrorInjection.InjectCreateError = ejectOrInject
	return ei
}

// WithZFSDeleteError injects/ejects ZFS delete error.
func (ei *ErrorInjection) WithZFSDeleteError(ejectOrInject string) *ErrorInjection {
	ei.ZFSError.CRUDErrorInjection.InjectDeleteError = ejectOrInject
	return ei
}

// WithCVRCreateError injects/ejects CVR create error.
func (ei *ErrorInjection) WithCVRCreateError(ejectOrInject string) *ErrorInjection {
	ei.CVRError.CRUDErrorInjection.InjectCreateError = ejectOrInject
	return ei
}

// WithCVRGetError injects/ejects CVR get error.
func (ei *ErrorInjection) WithCVRGetError(ejectOrInject string) *ErrorInjection {
	ei.CVRError.CRUDErrorInjection.InjectGetError = ejectOrInject
	return ei
}

// WithCVRDeleteError injects/ejects CVR delete error.
func (ei *ErrorInjection) WithCVRDeleteError(ejectOrInject string) *ErrorInjection {
	ei.CVRError.CRUDErrorInjection.InjectDeleteError = ejectOrInject
	return ei
}

// WithCVRUpdateError injects/ejects CVR update error.
func (ei *ErrorInjection) WithCVRUpdateError(ejectOrInject string) *ErrorInjection {
	ei.CVRError.CRUDErrorInjection.InjectUpdateError = ejectOrInject
	return ei
}
