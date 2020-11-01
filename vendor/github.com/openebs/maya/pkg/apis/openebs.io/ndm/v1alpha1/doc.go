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

// The API files in this package are APIs used by NDM and is copied from
// https://github.com/openebs/node-disk-manager/tree/master/pkg/apis/openebs/v1alpha1.
//
// All *_types.go from the above directory is copied into this package and code is
// generated. For all resource and resourceList, it should be added to register.go
// file so that the APIs are added to the scheme. The build tags also need to be modified
// for each resource type, so that listers and informers are auto-generated.

// +k8s:deepcopy-gen=package,register

// Package v1alpha1 is the v1alpha1 version of the API
// +groupName=openebs.io
package v1alpha1
