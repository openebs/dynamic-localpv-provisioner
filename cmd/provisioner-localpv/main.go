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
	"flag"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/openebs/dynamic-localpv-provisioner/cmd/provisioner-localpv/app"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/logger"
)

func init() {
	// Declare klog CLI flags
	klog.InitFlags(flag.CommandLine)
	// NOTE: As of klog/v2@v2.40.1 the --logtostderr=true option cannot be
	//       used alongside other klog flags to write logs in a file or
	//       directory.
	//       The --alsologtostderr=true option can be used alongside other
	//       klog flags to write logs to a file or directory. Disabling
	//       this flag will disable logging to stderr (while
	//	 --logtostderr=false is set).
	// Ref: https://github.com/kubernetes/klog/issues/60
	// User flags will be honored at Parse time.
	flag.CommandLine.Set("logtostderr", "false")
	flag.CommandLine.Set("alsologtostderr", "true")

	// Merge klog CLI flags to the global flagset
	// The pflag.Commandline FlagSet will be parsed
	// in the run() function in this package, before
	// initializing logging and cobra command execution.
	// Ref: github.com/spf13/pflag#supporting-go-flags-when-using-pflag
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// run starts the dynamic provisioner for Local PVs
func run() error {
	// Create new cobra command
	cmd, err := app.StartProvisioner()
	if err != nil {
		return err
	}

	// Merge all flags from the Cobra Command to the global FlagSet
	// and Parse them
	pflag.CommandLine.AddFlagSet(cmd.Flags())
	pflag.Parse()

	// NOTE: Logging must start after CLI flags have been parsed
	// Initialize logs and Flush logs on exit
	logger.InitLogging()
	defer logger.FinishLogging()

	// Run new command
	return cmd.Execute()
}
