/*
Copyright 2018 The OpenEBS Authors

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
	cstor "github.com/openebs/maya/pkg/snapshot/cstor/v1alpha1"
	"github.com/pkg/errors"
)

// cstorSnapshotCreate represents a cstor snapshot create runtask command
//
// NOTE:
//  This is an implementation of CommandRunner
type cstorSnapshotCreate struct {
	*cstorSnapshotCommand
}

// Run creates cstor snapshot contents
func (c *cstorSnapshotCreate) Run() (r RunCommandResult) {
	err := c.validateOptions()
	if err != nil {
		return c.AddError(errors.Errorf("failed to create cstor snapshot: %s", err)).Result(nil)
	}
	ip, _ := c.Data["ip"].(string)

	// get snapshot operation struct
	snapOps := cstor.Cstor()
	snapOps.IP = ip
	snapOps.Snap = c.casSnapshot()

	// use the struct to call the Create method
	response, err := snapOps.Create()
	if err != nil {
		return c.AddError(err).Result(nil)
	}
	return c.Result(response)
}
