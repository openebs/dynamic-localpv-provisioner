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
	corev1 "k8s.io/api/core/v1"
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

// Default ping cadence; mirrors the constants previously provided by
// github.com/openebs/google-analytics-4/usage so removing that loop does not
// change observed event frequency. OPENEBS_IO_ANALYTICS_PING_INTERVAL still
// controls the interval at runtime.
const (
	analyticsPingPeriodEnv = "OPENEBS_IO_ANALYTICS_PING_INTERVAL"
	defaultPingPeriod      = 24 * time.Hour
	minimumPingPeriod      = 1 * time.Hour
)

// getAnalyticsPingPeriod returns the configured ping interval, falling back
// to the default when unset or below the minimum. Matches the validation
// previously done inside analytics.PingCheckCtx.
func getAnalyticsPingPeriod() time.Duration {
	raw := os.Getenv(analyticsPingPeriodEnv)
	if raw == "" {
		return defaultPingPeriod
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d < minimumPingPeriod {
		return defaultPingPeriod
	}
	return d
}

// analyticsEmitter bundles the dependencies shared by every analytics
// helper (kube client, target namespace and the persistence CM name) so
// individual methods do not have to thread them through long parameter
// lists. An empty cmName means "no persistence" — callers fall back to
// the legacy emit-each-startup behavior for non-Helm deployments.
type analyticsEmitter struct {
	kubeClient kubernetes.Interface
	namespace  string
	cmName     string
}

func newAnalyticsEmitter(kubeClient kubernetes.Interface, namespace string) *analyticsEmitter {
	return &analyticsEmitter{
		kubeClient: kubeClient,
		namespace:  namespace,
		cmName:     getAnalyticsStateCMName(),
	}
}

// start performs install bookkeeping and launches the ping and heartbeat
// goroutines. Used by both helper-pod and leader-elected paths.
//
// Ping/heartbeat cadence is persisted in the state ConfigMap (data keys
// AnalyticsLastPingTSKey and AnalyticsLastHeartbeatTSKey) so a pod restart
// — or a leadership transition in DaemonSet mode — picks up where the
// previous emitter left off instead of immediately re-emitting.
func (a *analyticsEmitter) start(ctx context.Context) {
	if err := a.maybeEmitInstall(ctx); err != nil {
		klog.Warningf("analytics: install bookkeeping failed: %v", err)
	}
	// Ping: legacy behavior was to wait one period before the first emit.
	// Heartbeat: legacy behavior was to emit immediately and then every
	// period. Persisted timestamps in the CM override these defaults when
	// they are recent enough.
	go a.runChannel(ctx, Ping, AnalyticsLastPingTSKey, false)
	go a.runChannel(ctx, Heartbeat, AnalyticsLastHeartbeatTSKey, true)
}

// runAnalyticsHelperPod runs the analytics path for helper-pod (Deployment) mode.
// Single pod, no leader election; install is still CM-gated so a pod restart
// does not re-emit it.
func runAnalyticsHelperPod(ctx context.Context, kubeClient kubernetes.Interface, namespace string) {
	defer recoverAnalytics("helperPod")
	newAnalyticsEmitter(kubeClient, namespace).start(ctx)
}

// runAnalyticsLeaderElected runs the analytics path for node-deployment (DaemonSet)
// mode under a dedicated Lease so exactly one pod emits at a time. Blocks until
// ctx is cancelled. Uses NewLeaderElector + Run (not RunOrDie) so construction
// errors are recoverable rather than process-fatal.
//
// The Lease itself is no longer pre-created by the Helm chart; client-go's
// LeaseLock creates it on first acquisition, matching the same app-managed
// model used for the analytics state ConfigMap.
func runAnalyticsLeaderElected(ctx context.Context, kubeClient kubernetes.Interface, namespace string) {
	defer recoverAnalytics("leaderElected")

	identity, err := analyticsLeaseIdentity()
	if err != nil {
		klog.Errorf("analytics: lease identity: %v", err)
		return
	}
	leaseName := getAnalyticsLeaseName()
	emitter := newAnalyticsEmitter(kubeClient, namespace)

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
				emitter.start(leaderCtx)
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

// ensureStateCM returns the existing analytics state ConfigMap, creating
// an empty one if it is not present. Returns (nil, nil) when cmName is
// empty (non-Helm deployments that did not set the env var) so callers
// can fall back to no-persistence behavior.
//
// A 409 AlreadyExists on create is treated as a successful read race: a
// concurrent emitter (or a previous run of this process) created the CM
// between our Get and Create, and we just re-fetch it.
func (a *analyticsEmitter) ensureStateCM(ctx context.Context) (*corev1.ConfigMap, error) {
	if a.cmName == "" {
		return nil, nil
	}
	cms := a.kubeClient.CoreV1().ConfigMaps(a.namespace)
	cm, err := cms.Get(ctx, a.cmName, metav1.GetOptions{})
	if err == nil {
		return cm, nil
	}
	if !apierrors.IsNotFound(err) {
		return nil, errors.Wrap(err, "get analytics state configmap")
	}
	created, err := cms.Create(ctx, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: a.cmName, Namespace: a.namespace},
		Data:       map[string]string{},
	}, metav1.CreateOptions{})
	if err == nil {
		klog.V(2).Infof("analytics: created state ConfigMap %q", a.cmName)
		return created, nil
	}
	if apierrors.IsAlreadyExists(err) {
		cm, getErr := cms.Get(ctx, a.cmName, metav1.GetOptions{})
		if getErr != nil {
			return nil, errors.Wrap(getErr, "re-get analytics state configmap after AlreadyExists race")
		}
		return cm, nil
	}
	return nil, errors.Wrap(err, "create analytics state configmap")
}

// maybeEmitInstall sends the install event once per analytics-state
// ConfigMap, gated by the AnalyticsInstallTSKey data key. The provisioner
// creates the ConfigMap itself on first run; reinstalls into a namespace
// that still holds the previous ConfigMap will not re-fire install unless
// the ConfigMap is deleted manually (matches rawfile-localpv).
//
// When cmName is unset (typical for non-Helm deployments) install is
// emitted unconditionally on every startup — same behavior these
// deployments had before CM gating existed.
func (a *analyticsEmitter) maybeEmitInstall(ctx context.Context) error {
	if a.cmName == "" {
		// Non-Helm deployment that hasn't set the env var. Fall back to
		// legacy emit-each-startup behavior rather than silently dropping
		// install telemetry for these deployers.
		emitInstall()
		return nil
	}

	cm, err := a.ensureStateCM(ctx)
	if err != nil {
		// CM ops failed; emit unconditionally so a transient API error
		// doesn't drop install telemetry, and let the caller log.
		emitInstall()
		return err
	}
	if _, done := cm.Data[AnalyticsInstallTSKey]; done {
		return nil
	}

	// Send before marking: losing an install on Send failure is better than a
	// permanent loss on Patch failure after a successful Send.
	emitInstall()

	if err := a.patchState(ctx, AnalyticsInstallTSKey, time.Now().UTC()); err != nil {
		return errors.Wrap(err, "record install timestamp")
	}
	klog.V(2).Infof("analytics: install emitted; marked %q in %q", AnalyticsInstallTSKey, a.cmName)
	return nil
}

// patchState writes a single RFC3339 timestamp into the state CM under
// dataKey using a merge patch so it never clobbers unrelated keys written
// by other goroutines. Must only be called when a.cmName is non-empty.
func (a *analyticsEmitter) patchState(ctx context.Context, dataKey string, ts time.Time) error {
	patch, err := json.Marshal(map[string]any{
		"data": map[string]string{dataKey: ts.UTC().Format(time.RFC3339)},
	})
	if err != nil {
		return errors.Wrap(err, "marshal analytics state patch")
	}
	if _, err := a.kubeClient.CoreV1().ConfigMaps(a.namespace).Patch(
		ctx, a.cmName, types.MergePatchType, patch, metav1.PatchOptions{},
	); err != nil {
		return errors.Wrap(err, "patch analytics state configmap")
	}
	return nil
}

// runChannel drives one analytics event channel (ping or heartbeat) off a
// CM-persisted timestamp so cadence survives restarts and leadership
// transitions.
//
// On entry it computes the wait until the next emit from the last persisted
// timestamp and the configured period; if the previous emit is already older
// than the period, the next emit fires immediately. When the CM is missing or
// the data key is absent, the immediate parameter controls first-emit
// behavior — true for heartbeat (preserves the immediate=true behavior the
// library used) and false for ping (preserves immediate=false).
//
// Each successful send is followed by a best-effort patch to the CM; patch
// failures are logged but do not stop the loop, so a transient API error
// cannot silently halt analytics until the next restart.
func (a *analyticsEmitter) runChannel(ctx context.Context, category, dataKey string, immediate bool) {
	defer recoverAnalytics(category)

	period := getAnalyticsPingPeriod()
	timer := time.NewTimer(a.initialWait(ctx, dataKey, period, immediate))
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		emitPing(category)
		if a.cmName != "" {
			if err := a.patchState(ctx, dataKey, time.Now()); err != nil {
				klog.Warningf("analytics: %s: persist %q: %v", category, dataKey, err)
			}
		}
		timer.Reset(period)
	}
}

// initialWait computes how long to wait before the first emit of a
// channel. Reads the last-emit timestamp from the CM (if available) and
// returns period minus elapsed, clamped to >= 0. Falls back to the legacy
// immediate/period default when no timestamp is available.
func (a *analyticsEmitter) initialWait(ctx context.Context, dataKey string, period time.Duration, immediate bool) time.Duration {
	fallback := period
	if immediate {
		fallback = 0
	}
	if a.cmName == "" {
		return fallback
	}
	cm, err := a.kubeClient.CoreV1().ConfigMaps(a.namespace).Get(ctx, a.cmName, metav1.GetOptions{})
	if err != nil {
		// Don't block startup on a transient API error: behave as if the
		// CM were absent. ensureStateCM already logged the create path;
		// here we just fall back.
		return fallback
	}
	raw, ok := cm.Data[dataKey]
	if !ok || raw == "" {
		return fallback
	}
	last, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		klog.Warningf("analytics: ignoring unparseable %q=%q in %q: %v", dataKey, raw, a.cmName, err)
		return fallback
	}
	remaining := period - time.Since(last)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// emitPing sends a single ping/heartbeat event under the given category. It
// mirrors the payload that analytics.PingCheckCtx used to build internally
// (CommonBuild + InstallBuilder(true) + SetCategory), so the analytics
// backend continues to see the same event shape.
func emitPing(category string) {
	analytics.New().CommonBuild(DefaultCASType).InstallBuilder(true).SetCategory(category).Send()
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
