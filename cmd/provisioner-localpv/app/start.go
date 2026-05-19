package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	analytics "github.com/openebs/google-analytics-4/usage"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
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

	// Analytics leader-election tuning. Drives API call rate for the analytics
	// Lease in node-deployment mode; no effect in helper-pod mode. Must satisfy
	// LeaseDuration > RenewDeadline > 1.2 × RetryPeriod — invalid combinations
	// disable analytics for this pod without affecting the provisioner.
	cmd.Flags().DurationVar(&AnalyticsLeaseDuration, "analytics-lease-duration", AnalyticsLeaseDuration,
		"Duration after which an unrenewed analytics Lease is considered expired and can be taken over.")
	cmd.Flags().DurationVar(&AnalyticsRenewDeadline, "analytics-renew-deadline", AnalyticsRenewDeadline,
		"Maximum time the analytics leader will keep retrying to renew the Lease before giving up leadership.")
	cmd.Flags().DurationVar(&AnalyticsRetryPeriod, "analytics-retry-period", AnalyticsRetryPeriod,
		"Interval between renewal attempts (leader) and acquisition attempts (followers). Drives steady-state API call rate.")

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

	threadiness := getWorkerThreads()
	log.Info("Provisioner concurrency configured", "workerThreads", threadiness)

	pc := pvController.NewProvisionController(
		ctx,
		kubeClient,
		provisionerName,
		provisioner,
		pvController.LeaderElection(leaderElection),
		pvController.Threadiness(threadiness),
	)

	if utils.GoogleAnalyticsEnabled(GoogleAnalyticsKey) {
		analytics.RegisterVersionGetter(version.GetVersionDetails)
		if nodeDeployment {
			go runAnalyticsLeaderElected(ctx, kubeClient, provisioner.namespace)
		} else {
			go runAnalyticsHelperPod(ctx, kubeClient, provisioner.namespace)
		}
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

// recoverAnalytics is deferred at every analytics goroutine and callback so
// a panic never crashes the provisioner. Logs one line; no stack trace.
func recoverAnalytics(where string) {
	if r := recover(); r != nil {
		klog.Errorf("analytics: recovered from panic in %s: %v", where, r)
	}
}

// runPingLoop is a panic-safe wrapper around PingCheckCtx for use as a goroutine entry.
func runPingLoop(ctx context.Context, category string, immediate bool) {
	defer recoverAnalytics(category)
	analytics.PingCheckCtx(ctx, DefaultCASType, category, immediate)
}

// startAnalyticsEmitters performs install bookkeeping and starts the ping
// and heartbeat goroutines. Used by both helper-pod and leader-elected paths.
func startAnalyticsEmitters(ctx context.Context, kubeClient kubernetes.Interface, namespace string) {
	if err := maybeEmitInstall(ctx, kubeClient, namespace); err != nil {
		klog.Warningf("analytics: install bookkeeping failed: %v", err)
	}
	go runPingLoop(ctx, Ping, false)
	go runPingLoop(ctx, Heartbeat, true)
}

// runAnalyticsHelperPod runs the analytics path for helper-pod (Deployment) mode.
// Single pod, no leader election; install is still CM-gated so a pod restart
// does not re-emit it.
func runAnalyticsHelperPod(ctx context.Context, kubeClient kubernetes.Interface, namespace string) {
	defer recoverAnalytics("helperPod")
	startAnalyticsEmitters(ctx, kubeClient, namespace)
}

// runAnalyticsLeaderElected runs the analytics path for node-deployment (DaemonSet)
// mode under a dedicated Lease so exactly one pod emits at a time. Blocks until
// ctx is cancelled. Uses NewLeaderElector + Run (not RunOrDie) so construction
// errors are recoverable rather than process-fatal.
func runAnalyticsLeaderElected(ctx context.Context, kubeClient kubernetes.Interface, namespace string) {
	defer recoverAnalytics("leaderElected")

	identity, err := analyticsLeaseIdentity()
	if err != nil {
		klog.Errorf("analytics: lease identity: %v", err)
		return
	}
	leaseName := getAnalyticsLeaseName()

	le, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock: &resourcelock.LeaseLock{
			LeaseMeta:  metav1.ObjectMeta{Name: leaseName, Namespace: namespace},
			Client:     kubeClient.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{Identity: identity},
		},
		ReleaseOnCancel: true,
		LeaseDuration:   AnalyticsLeaseDuration,
		RenewDeadline:   AnalyticsRenewDeadline,
		RetryPeriod:     AnalyticsRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(leaderCtx context.Context) {
				defer recoverAnalytics("OnStartedLeading")
				klog.V(2).Infof("analytics: acquired lease %q as %q", leaseName, identity)
				startAnalyticsEmitters(leaderCtx, kubeClient, namespace)
				<-leaderCtx.Done()
			},
			OnStoppedLeading: func() {
				klog.V(2).Infof("analytics: released lease %q (identity %q)", leaseName, identity)
			},
		},
	})
	if err != nil {
		klog.Errorf("analytics: leader elector: %v", err)
		return
	}
	le.Run(ctx)
}

// maybeEmitInstall sends the install event once per Helm release, gated by a
// ConfigMap whose name is provided via OPENEBS_IO_ANALYTICS_STATE_CM. The chart
// creates the CM empty on helm install and removes it on helm uninstall, so a
// fresh install fires install again. When OPENEBS_IO_ANALYTICS_STATE_CM is
// unset or names a CM that does not exist in the cluster (typical for non-Helm
// deployments), install is emitted unconditionally — same behavior these
// deployments had before CM gating existed.
func maybeEmitInstall(ctx context.Context, kubeClient kubernetes.Interface, namespace string) error {
	cmName := getAnalyticsStateCMName()
	if cmName == "" {
		// Non-Helm deployment that hasn't set the env var. Fall back to
		// legacy emit-each-startup behavior rather than silently dropping
		// install telemetry for these deployers.
		emitInstall()
		return nil
	}

	cm, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, cmName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// Env var set but the CM is absent from the cluster — chart
		// updated without the analytics-state template, or someone
		// deleted it. Same fallback as no-env-var: emit without state
		// tracking.
		klog.V(2).Infof("analytics: state ConfigMap %q not found; emitting install without state tracking", cmName)
		emitInstall()
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "get analytics state configmap")
	}
	if _, done := cm.Data[AnalyticsInstalledAtKey]; done {
		return nil
	}

	// Send before marking: losing an install on Send failure is better than a
	// permanent loss on Patch failure after a successful Send.
	emitInstall()

	patch, err := json.Marshal(map[string]any{
		"data": map[string]string{AnalyticsInstalledAtKey: time.Now().UTC().Format(time.RFC3339)},
	})
	if err != nil {
		return errors.Wrap(err, "marshal install-at patch")
	}
	if _, err := kubeClient.CoreV1().ConfigMaps(namespace).Patch(
		ctx, cmName, types.MergePatchType, patch, metav1.PatchOptions{},
	); err != nil {
		return errors.Wrap(err, "patch analytics state configmap")
	}
	klog.V(2).Infof("analytics: install emitted; marked %q in %q", AnalyticsInstalledAtKey, cmName)
	return nil
}

// emitInstall sends a single install event to GA. Fire-and-forget; failures
// are logged inside the analytics library and do not propagate.
func emitInstall() {
	analytics.New().CommonBuild(DefaultCASType).InstallBuilder(true).Send()
}

// analyticsLeaseIdentity returns the leader-election identity for this pod.
func analyticsLeaseIdentity() (string, error) {
	if name := getPodName(); name != "" {
		return name, nil
	}
	h, err := os.Hostname()
	if err != nil {
		return "", errors.Wrap(err, "hostname")
	}
	return fmt.Sprintf("%s_%s", h, uuid.NewUUID()), nil
}
