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

package main

import (
	"os"

	"github.com/openebs/maya/cmd/provisioner-localpv/app"
	mlogger "github.com/openebs/maya/pkg/logs"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// run starts the dynamic provisioner for Local PVs
func run() error {
	// Init logging
	mlogger.InitLogs()
	defer mlogger.FlushLogs()

	// Create & execute new command
	cmd, err := app.StartProvisioner()
	if err != nil {
		return err
	}

	return cmd.Execute()
}
