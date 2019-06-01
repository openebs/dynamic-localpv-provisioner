/*
Copyright 2019 The OpenEBS Authors.

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

package app

import (
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
)

// HelperPodOptions contains the options that
// will be extracted from the persistent volume
type HelperPodOptions struct {
	cmdsForPath []string
	name        string
	path        string
	nodeName    string
}

// validate checks that the required fields to launch
// helper pods are valid. helper pods are used to either
// create or delete a directory (path) on a given node (nodeName).
// name refers to the volume being created or deleted.
func (pOpts *HelperPodOptions) validate() error {
	if pOpts.name == "" || pOpts.path == "" || pOpts.nodeName == "" {
		return errors.Errorf("invalid empty name or path or node")
	}
	return nil
}
