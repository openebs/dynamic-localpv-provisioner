package app

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	analytics "github.com/openebs/google-analytics-4/usage"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	pvController "sigs.k8s.io/sig-storage-lib-external-provisioner/v13/controller"

	mKube "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/client"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/logger"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/utils"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/version"
)

var (
	cmdName         = "provisioner"
	provisionerName = "openebs.io/local"
	// LeaderElectionKey represents ENV for disable/enable leaderElection for
	// localpv provisioner
	LeaderElectionKey = "LEADER_ELECTION_ENABLED"
	usage             = cmdName
)

// StartProvisioner will start a new dynamic Host Path PV provisioner
func StartProvisioner() (*cobra.Command, error) {
	var (
		nodeDeployment bool
		logFlushFreq   time.Duration
	)

	// Create a new command.
	cmd := &cobra.Command{
		Use:   usage,
		Short: "Dynamic Host Path PV Provisioner",
		Long: `Manage the Host Path PVs that includes: validating, creating,
			deleting and cleanup tasks. Host Path PVs are setup with
			node affinity`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ctx = logger.InitLogging(ctx, logFlushFreq)
			defer logger.FinishLogging()

			logger.CheckErr(Start(ctx, nodeDeployment), logger.Fatal)
		},
	}

	// Add node deployment flag
	cmd.Flags().BoolVar(&nodeDeployment, "node-deployment", false, "Enables deploying the provisioner together with a CSI driver on nodes to manage node-local volumes")
	// Log flush frequency flag
	cmd.Flags().DurationVar(&logFlushFreq, "log-flush-frequency", logger.DefaultFlushInterval, "Delay between log flushes")
	return cmd, nil
}

// Start will initialize and run the dynamic provisioner daemon
func Start(ctx context.Context, nodeDeployment bool) error {
	log := klog.FromContext(ctx)
	log.Info("Starting Provisioner...")

	// Setup signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	//Create an instance of ProvisionerHandler to handle PV
	// create and delete events.
	provisioner, err := NewProvisioner(kubeClient)
	if err != nil {
		return err
	}

	// Set node deployment mode in provisioner
	provisioner.nodeDeployment = nodeDeployment

	// In node deployment mode, set the current node name for filtering
	if nodeDeployment {
		provisioner.nodeName = getNodeName()
		if provisioner.nodeName == "" {
			return errors.New("NODE_NAME environment variable is required in node-deployment mode")
		}
		log.Info("Node deployment mode enabled", "node", provisioner.nodeName)
	}

	//Create an instance of the Dynamic Provisioner Controller
	// that has the reconciliation loops for PVC create and delete
	// events and invokes the Provisioner Handler.
	leaderElection := isLeaderElectionEnabled(ctx, nodeDeployment)

	pc := pvController.NewProvisionController(
		ctx,
		kubeClient,
		provisionerName,
		provisioner,
		pvController.LeaderElection(leaderElection),
	)

	if utils.GoogleAnalyticsEnabled(GoogleAnalyticsKey) {
		analytics.RegisterVersionGetter(version.GetVersionDetails)
		analytics.New().CommonBuild(DefaultCASType).InstallBuilder(true).Send()
		go analytics.PingCheck(DefaultCASType, Ping, false)
		go analytics.PingCheck(DefaultCASType, Heartbeat, true)
	}

	log.V(4).Info("Provisioner started")
	//Run the provisioner till a shutdown signal is received.
	pc.Run(ctx)
	log.V(4).Info("Provisioner stopped")

	return nil
}

// isLeaderElectionEnabled returns true/false based on the ENV
// LEADER_ELECTION_ENABLED set via provisioner deployment.
// Defaults to true, means leaderElection enabled by default.
func isLeaderElectionEnabled(ctx context.Context, nodeDeployment bool) bool {
	log := klog.FromContext(ctx)

	// In node deployment mode, disable leader election as each node will have its own provisioner
	if nodeDeployment {
		log.Info("Leader election disabled for node deployment mode")
		return false
	}

	leaderElection := os.Getenv(LeaderElectionKey)

	var leader bool
	switch strings.ToLower(leaderElection) {
	default:
		log.Info("Leader election enabled for localpv-provisioner")
		leader = true
	case "y", "yes", "true":
		log.Info("Leader election enabled for localpv-provisioner via leaderElectionKey")
		leader = true
	case "n", "no", "false":
		log.Info("Leader election disabled for localpv-provisioner via leaderElectionKey")
		leader = false
	}
	return leader
}
