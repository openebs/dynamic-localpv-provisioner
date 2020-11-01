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
)

// UpgradeConfig represents configuration for a job or executor or some other
// component. This contains Resources, Castemplate and RuntimeConfigs
// Mentioned Resources will be updated using mention castemplate
type UpgradeConfig struct {
	// CASTemplate contain castemplate name which task executor
	// will use to upgrade single unit of resources.
	CASTemplate string `json:"casTemplate"`
	// Data is used to provide some runtime configurations to
	// castemplate engine. Task executor will directly copy these
	// configurations to castemplate engine.
	Data []DataItem `json:"data"`
	// Resources contains list of resources which we are going to upgrade
	Resources []ResourceDetails `json:"resources"`
}

// String implements Stringer interface
func (uc UpgradeConfig) String() string {
	return stringer.Yaml("upgrade config", uc)
}

// GoString implements GoStringer interface
func (uc UpgradeConfig) GoString() string {
	return uc.String()
}

// DataItem holds a runtime configuration for executor
type DataItem struct {
	// Name of the configuration
	Name string `json:"name"`
	// Value represents any specific value that is applicable
	// to this configuration
	Value string `json:"value"`
	// Entries represents an arbitrary map of key value pairs
	Entries map[string]string `json:"entries"`
}
