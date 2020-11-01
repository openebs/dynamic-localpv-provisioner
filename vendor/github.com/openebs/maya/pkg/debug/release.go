// +build !debug

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

// LogBuildDetails is no-op production alternative for the same function in present in debug build.
func LogBuildDetails() {}

// StartInjectionServer is no-op production alternative for the same function in present in debug build.
func StartInjectionServer() {}

// IsCSPCDeleteCollectionErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPCDeleteCollectionErrorInjected() bool {
	return false
}

// IsCSPCDeleteErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPCDeleteErrorInjected() bool {
	return false
}

// IsCSPCListErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPCListErrorInjected() bool {
	return false
}

// IsCSPCGetErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPCGetErrorInjected() bool {
	return false
}

// IsCSPCCreateErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPCCreateErrorInjected() bool {
	return false
}

// IsCSPCUpdateErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPCUpdateErrorInjected() bool {
	return false
}

// IsCSPCPatchErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPCPatchErrorInjected() bool {
	return false
}

// IsCSPIDeleteCollectionErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPIDeleteCollectionErrorInjected() bool {
	return false
}

// IsCSPIDeleteErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPIDeleteErrorInjected() bool {
	return false
}

// IsCSPIListErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPIListErrorInjected() bool {
	return false
}

// IsCSPIGetErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPIGetErrorInjected() bool {
	return false
}

// IsCSPICreateErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPICreateErrorInjected() bool {
	return false
}

// IsCSPIUpdateErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPIUpdateErrorInjected() bool {
	return false
}

// IsCSPIPatchErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsCSPIPatchErrorInjected() bool {
	return false
}

// IsDeploymentDeleteCollectionErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsDeploymentDeleteCollectionErrorInjected() bool {
	return false
}

// IsDeploymentDeleteErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsDeploymentDeleteErrorInjected() bool {
	return false
}

// IsDeploymentListErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsDeploymentListErrorInjected() bool {
	return false
}

// IsDeploymentGetErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsDeploymentGetErrorInjected() bool {
	return false
}

// IsDeploymentCreateErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsDeploymentCreateErrorInjected() bool {
	return false
}

// IsDeploymentUpdateErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsDeploymentUpdateErrorInjected() bool {
	return false
}

// IsDeploymentPatchErrorInjected is production alternative for the same function in present in debug build.
func (ei *ErrorInjection) IsDeploymentPatchErrorInjected() bool {
	return false
}

// IsZFSGetErrorInjected is production alternative for the same function that
// exists in debug build
func (ei *ErrorInjection) IsZFSGetErrorInjected() bool {
	return false
}

// IsZFSCreateErrorInjected is production alternative for the same function that
// exists in debug build
func (ei *ErrorInjection) IsZFSCreateErrorInjected() bool {
	return false
}

// IsZFSDeleteErrorInjected is production alternative for the same function that
// exists in debug build
func (ei *ErrorInjection) IsZFSDeleteErrorInjected() bool {
	return false
}

// IsCVRCreateErrorInjected is production alternative for the same function that
// exists in debug build
func (ei *ErrorInjection) IsCVRCreateErrorInjected() bool {
	return false
}

// IsCVRDeleteErrorInjected is production alternative for the same function that
// exists in debug build
func (ei *ErrorInjection) IsCVRDeleteErrorInjected() bool {
	return false
}

// IsCVRGetErrorInjected is production alternative for the same function that
// exists in debug build
func (ei *ErrorInjection) IsCVRGetErrorInjected() bool {
	return false
}

// IsCVRUpdateErrorInjected is production alternative for the same function that
// exists in debug build
func (ei *ErrorInjection) IsCVRUpdateErrorInjected() bool {
	return false
}
