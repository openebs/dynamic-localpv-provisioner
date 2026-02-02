package main

import (
	"flag"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/openebs/dynamic-localpv-provisioner/cmd/provisioner-localpv/app"
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

	// Run new command
	return cmd.Execute()
}
