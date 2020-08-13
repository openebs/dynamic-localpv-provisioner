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
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"

	mKube "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	pvController "sigs.k8s.io/sig-storage-lib-external-provisioner/controller"
)

var (
	cmdName         = "provisioner"
	provisionerName = "openebs.io/local"
	// LeaderElectionKey represents ENV for disable/enable leaderElection for
	// localpv provisioner
	LeaderElectionKey = "LEADER_ELECTION_ENABLED"
	usage             = fmt.Sprintf("%s", cmdName)
)

// StartProvisioner will start a new dynamic Host Path PV provisioner
func StartProvisioner() (*cobra.Command, error) {
	// Create a new command.
	cmd := &cobra.Command{
		Use:   usage,
		Short: "Dynamic Host Path PV Provisioner",
		Long: `Manage the Host Path PVs that includes: validating, creating,
			deleting and cleanup tasks. Host Path PVs are setup with
			node affinity`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(Start(cmd), util.Fatal)
		},
	}

	// add the default command line flags as global flags to cobra command
	// flagset
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Hack: Without the following line, the logs will be prefixed with Error
	_ = flag.CommandLine.Parse([]string{})

	return cmd, nil
}

// Start will initialize and run the dynamic provisioner daemon
func Start(cmd *cobra.Command) error {
	klog.Infof("Starting Provisioner...")

	// Dynamic Provisioner can run successfully if it can establish
	// connection to the Kubernetes Cluster. mKube helps with
	// establishing the connection either via InCluster or
	// OutOfCluster by using the following ENV variables:
	//   OPENEBS_IO_K8S_MASTER - Kubernetes master IP address
	//   OPENEBS_IO_KUBE_CONFIG - Path to the kubeConfig file.
	kubeClient, err := mKube.New().Clientset()
	if err != nil {
		return errors.Wrap(err, "unable to get k8s client")
	}

	serverVersion, err := kubeClient.Discovery().ServerVersion()
	if err != nil {
		return errors.Wrap(err, "Cannot start Provisioner: failed to get Kubernetes server version")
	}

	err = performPreupgradeTasks(kubeClient)
	if err != nil {
		return errors.Wrap(err, "failure in preupgrade tasks")
	}

	//Create a channel to receive shutdown signal to help
	// with graceful exit of the provisioner.
	stopCh := make(chan struct{})
	RegisterShutdownChannel(stopCh)

	//Create an instance of ProvisionerHandler to handle PV
	// create and delete events.
	provisioner, err := NewProvisioner(stopCh, kubeClient)
	if err != nil {
		return err
	}

	//Create an instance of the Dynamic Provisioner Controller
	// that has the reconciliation loops for PVC create and delete
	// events and invokes the Provisioner Handler.
	pc := pvController.NewProvisionController(
		kubeClient,
		provisionerName,
		provisioner,
		serverVersion.GitVersion,
		pvController.LeaderElection(isLeaderElectionEnabled()),
	)
	klog.V(4).Info("Provisioner started")
	//Run the provisioner till a shutdown signal is received.
	pc.Run(stopCh)
	klog.V(4).Info("Provisioner stopped")

	return nil
}

// isLeaderElectionEnabled returns true/false based on the ENV
// LEADER_ELECTION_ENABLED set via provisioner deployment.
// Defaults to true, means leaderElection enabled by default.
func isLeaderElectionEnabled() bool {
	leaderElection := os.Getenv(LeaderElectionKey)

	var leader bool
	switch strings.ToLower(leaderElection) {
	default:
		klog.Info("Leader election enabled for localpv-provisioner")
		leader = true
	case "y", "yes", "true":
		klog.Info("Leader election enabled for localpv-provisioner via leaderElectionKey")
		leader = true
	case "n", "no", "false":
		klog.Info("Leader election disabled for localpv-provisioner via leaderElectionKey")
		leader = false
	}
	return leader
}
