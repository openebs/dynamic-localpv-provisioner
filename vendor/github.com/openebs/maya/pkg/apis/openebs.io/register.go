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

package openebsio

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// GroupName assigned to openebs.io
const (
	GroupName = "openebs.io"
)

// SchemeBuilderAdditions may be used to add all
// resources defined in the project to a Scheme
var SchemeBuilderAdditions runtime.SchemeBuilder

// AddToScheme adds the provided resource
// to Scheme
func AddToScheme(s *runtime.Scheme) error {
	return SchemeBuilderAdditions.AddToScheme(s)
}
