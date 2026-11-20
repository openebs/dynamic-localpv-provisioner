package app

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	hostpath "github.com/openebs/dynamic-localpv-provisioner/pkg/hostpath/v1alpha1"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/container"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/pod"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/volume"
)

type podConfig struct {
	pOpts                         *HelperPodOptions
	parentDir, volumeDir, podName string
	taints                        []corev1.Taint
}

var (
	//CmdTimeoutCounts specifies the maximum number of seconds to wait
	//for a helper pod to complete. Configurable via OPENEBS_IO_HELPER_POD_TIMEOUT_SECS.
	CmdTimeoutCounts = getHelperPodTimeout()
)

// maxHelperPodMessageLen caps the length of a status message reported for a
// helper pod, e.g. the container log tail used as its termination message.
const maxHelperPodMessageLen = 512

// HelperPodOptions contains the options that
// will launch a Pod on a specific node (nodeHostname)
// to execute a command (cmdsForPath) on a given
// volume path (path)
type HelperPodOptions struct {
	//nodeAffinityLabels represents the labels of the node where pod should be launched.
	nodeAffinityLabels map[string]string

	//name is the name of the PV for which the pod is being launched
	name string

	//cmdsForPath represent either create (mkdir) or delete(rm)
	//commands that need to be executed on the volume path.
	cmdsForPath []string

	//path is the volume hostpath directory
	path string

	//fsMode is the file mode to apply to the volume directory,
	//derived from the FilePermissions config (defaults to 0777).
	fsMode string

	//serviceAccountName is the service account with which the pod should be launched
	serviceAccountName string

	selectedNodeTaints []corev1.Taint

	imagePullSecrets []corev1.LocalObjectReference

	//softLimitGrace is the soft limit of quota on the project
	softLimitGrace string

	//hardLimitGrace is the hard limit of quota on the project
	hardLimitGrace string

	//pvcStorage is the storage requested for pv
	pvcStorage int64

	//hostNetwork is the network type of helper Pod
	hostNetwork bool

	//imagePullPolicy for helper pod container
	imagePullPolicy corev1.PullPolicy
}

// validate checks that the required fields to launch
// helper pods are valid. helper pods are used to either
// create or delete a directory (path) on a given node hostname (nodeHostname).
// name refers to the volume being created or deleted.
func (pOpts *HelperPodOptions) validate() error {
	if pOpts.name == "" ||
		pOpts.path == "" ||
		pOpts.nodeAffinityLabels == nil ||
		len(pOpts.nodeAffinityLabels) == 0 ||
		pOpts.serviceAccountName == "" {
		return errors.Errorf("invalid empty name or hostpath or hostname or service account name")
	}
	return nil
}

// validateLimits check that the limits to setup qouta are valid
func (pOpts *HelperPodOptions) validateLimits() error {
	if pOpts.softLimitGrace == "0k" &&
		pOpts.hardLimitGrace == "0k" {
		// Hack: using convertToK() style converstion
		// TODO: Refactor this section of the code
		pvcStorageInK := math.Ceil(float64(pOpts.pvcStorage) / 1024)
		pvcStorageInKString := strconv.FormatFloat(pvcStorageInK, 'f', -1, 64) + "k"
		pOpts.softLimitGrace, pOpts.hardLimitGrace = pvcStorageInKString, pvcStorageInKString
		return nil
	}

	if pOpts.softLimitGrace == "0k" ||
		pOpts.hardLimitGrace == "0k" {
		return nil
	}

	if len(pOpts.softLimitGrace) > len(pOpts.hardLimitGrace) ||
		(len(pOpts.softLimitGrace) == len(pOpts.hardLimitGrace) &&
			pOpts.softLimitGrace > pOpts.hardLimitGrace) {
		return errors.Errorf("hard limit cannot be smaller than soft limit")
	}

	return nil
}

// converToK converts the limits to kilobytes
func convertToK(limit string, pvcStorage int64) (string, error) {

	if len(limit) == 0 {
		return "0k", nil
	}

	valueRegex := regexp.MustCompile(`[\d]*[\.]?[\d]*`)
	valueString := valueRegex.FindString(limit)

	if limit != valueString+"%" {
		return "", errors.Errorf("invalid format for limit grace")
	}

	value, err := strconv.ParseFloat(valueString, 64)

	if err != nil {
		return "", errors.Errorf("invalid format, cannot parse")
	}
	if value > 100 {
		value = 100
	}

	value *= float64(pvcStorage)
	value /= 100
	value += float64(pvcStorage)
	value /= 1024

	value = math.Ceil(value)
	valueString = strconv.FormatFloat(value, 'f', -1, 64)
	valueString += "k"
	return valueString, nil
}

// createInitPod launches a helper(busybox) pod, to create the host path.
//
//	The local pv expect the hostpath to be already present before mounting
//	into pod. Validate that the local pv host path is not created under root.
func (p *Provisioner) createInitPod(ctx context.Context, pOpts *HelperPodOptions) error {
	var config podConfig
	config.pOpts, config.podName = pOpts, "init"
	//err := pOpts.validate()
	if err := pOpts.validate(); err != nil {
		return err
	}

	// Initialize HostPath builder and validate that
	// volume directory is not directly under root.
	// Extract the base path and the volume unique path.
	var vErr error
	config.parentDir, config.volumeDir, vErr = hostpath.NewBuilder().WithPath(pOpts.path).
		WithCheckf(hostpath.IsNonRoot(), "volume directory {%v} should not be under root directory", pOpts.path).
		ExtractSubPath()
	if vErr != nil {
		return vErr
	}

	//Pass on the taints, to create tolerations.
	config.taints = pOpts.selectedNodeTaints

	config.pOpts.cmdsForPath = append(config.pOpts.cmdsForPath, filepath.Join("/data/", config.volumeDir))

	_, err := p.launchPod(ctx, config)
	if err != nil && !k8serror.IsAlreadyExists(err) {
		return err
	}

	if err := p.exitPod(ctx, config.podName+"-"+config.pOpts.name); err != nil {
		return err
	}

	return nil
}

// createCleanupPod launches a helper(busybox) pod, to delete the host path.
//
//	This provisioner expects that the host paths are created using
//	an unique PV path - under a given BasePath. From the absolute path,
//	it extracts the base path and the PV path. The helper pod is then launched
//	by mounting the base path - and performing a delete on the unique PV path.
func (p *Provisioner) createCleanupPod(ctx context.Context, pOpts *HelperPodOptions) error {
	var config podConfig
	config.pOpts, config.podName = pOpts, "cleanup"
	//err := pOpts.validate()
	if err := pOpts.validate(); err != nil {
		return err
	}

	// Initialize HostPath builder and validate that
	// volume directory is not directly under root.
	// Extract the base path and the volume unique path.
	var vErr error
	config.parentDir, config.volumeDir, vErr = hostpath.NewBuilder().WithPath(pOpts.path).
		WithCheckf(hostpath.IsNonRoot(), "volume directory {%v} should not be under root directory", pOpts.path).
		ExtractSubPath()
	if vErr != nil {
		return vErr
	}

	config.taints = pOpts.selectedNodeTaints

	// Generate cleanup script using shared utility
	// Helper pod mounts parentDir at /data, so use /data as the parent path
	cleanupScript := GenerateQuotaCleanupScript(QuotaScriptConfig{
		ParentDir:      "/data",
		VolumeDir:      config.volumeDir,
		HostPathPrefix: "",   // No prefix needed, /data is the mount point
		UseHostLock:    true, // serialize against concurrent helper pods
	})

	config.pOpts.cmdsForPath = []string{"sh", "-c", cleanupScript}

	_, err := p.launchPod(ctx, config)
	if err != nil && !k8serror.IsAlreadyExists(err) {
		return err
	}

	if err := p.exitPod(ctx, config.podName+"-"+config.pOpts.name); err != nil {
		return err
	}
	return nil
}

// createQuotaPod launches a helper(busybox) pod, to apply the quota.
//
//	The local pv expect the hostpath to be already present before mounting
//	into pod. Validate that the local pv host path is not created under root.
func (p *Provisioner) createQuotaPod(ctx context.Context, pOpts *HelperPodOptions) error {
	var config podConfig
	config.pOpts, config.podName = pOpts, "quota"
	//err := pOpts.validate()
	if err := pOpts.validate(); err != nil {
		return err
	}

	// Initialize HostPath builder and validate that
	// volume directory is not directly under root.
	// Extract the base path and the volume unique path.
	var vErr error
	config.parentDir, config.volumeDir, vErr = hostpath.NewBuilder().WithPath(pOpts.path).
		WithCheckf(hostpath.IsNonRoot(), "volume directory {%v} should not be under root directory", pOpts.path).
		ExtractSubPath()
	if vErr != nil {
		return vErr
	}

	//Pass on the taints, to create tolerations.
	config.taints = pOpts.selectedNodeTaints

	var lErr error
	config.pOpts.softLimitGrace, lErr = convertToK(config.pOpts.softLimitGrace, config.pOpts.pvcStorage)
	if lErr != nil {
		return lErr
	}
	config.pOpts.hardLimitGrace, lErr = convertToK(config.pOpts.hardLimitGrace, config.pOpts.pvcStorage)
	if lErr != nil {
		return lErr
	}

	if err := pOpts.validateLimits(); err != nil {
		return err
	}

	// Generate quota script using shared utility
	// Helper pod mounts parentDir at /data, so use /data as the parent path
	quotaScript := GenerateQuotaApplyScript(QuotaScriptConfig{
		ParentDir:      "/data",
		VolumeDir:      config.volumeDir,
		SoftLimitGrace: config.pOpts.softLimitGrace,
		HardLimitGrace: config.pOpts.hardLimitGrace,
		HostPathPrefix: "",   // No prefix needed, /data is the mount point
		UseHostLock:    true, // serialize project-ID allocation
	})

	config.pOpts.cmdsForPath = []string{"sh", "-c", quotaScript}

	_, err := p.launchPod(ctx, config)
	if err != nil && !k8serror.IsAlreadyExists(err) {
		return err
	}

	if err := p.exitPod(ctx, config.podName+"-"+config.pOpts.name); err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) launchPod(ctx context.Context, config podConfig) (*corev1.Pod, error) {
	// the helper pod need to be launched in privileged mode. This is because in CoreOS
	// nodes, pods without privileged access cannot write to the host directory.
	// Helper pods need to create and delete directories on the host.
	privileged := true

	helperPod, err := pod.NewBuilder().
		WithName(config.podName + "-" + config.pOpts.name).
		WithRestartPolicy(corev1.RestartPolicyNever).
		//WithNodeSelectorHostnameNew(config.pOpts.nodeHostname).
		WithNodeAffinityNew(config.pOpts.nodeAffinityLabels).
		WithServiceAccountName(config.pOpts.serviceAccountName).
		WithTolerationsForTaints(config.taints...).
		WithContainerBuilder(
			container.NewBuilder().
				WithName("local-path-" + config.podName).
				WithImage(p.helperImage).
				WithCommandNew(config.pOpts.cmdsForPath).
				WithVolumeMountsNew([]corev1.VolumeMount{
					{
						Name:      "data",
						ReadOnly:  false,
						MountPath: "/data/",
					},
				}).
				WithImagePullPolicy(config.pOpts.imagePullPolicy).
				// Use the tail of the container logs as the termination message
				// when the helper command fails, so that its error output is
				// available after the helper pod has been deleted.
				WithTerminationMessagePolicy(corev1.TerminationMessageFallbackToLogsOnError).
				WithPrivilegedSecurityContext(&privileged),
		).
		WithImagePullSecrets(config.pOpts.imagePullSecrets).
		WithVolumeBuilder(
			volume.NewBuilder().
				WithName("data").
				WithHostDirectory(config.parentDir),
		).
		WithHostNetwork(config.pOpts.hostNetwork).
		Build()

	if err != nil {
		return nil, err
	}

	var hPod *corev1.Pod

	//Launch the helper pod.
	hPod, err = p.kubeClient.CoreV1().Pods(p.namespace).Create(ctx, helperPod, metav1.CreateOptions{})
	return hPod, err
}

func (p *Provisioner) exitPod(ctx context.Context, hPodName string) error {
	log := klog.FromContext(ctx)
	defer func() {
		e := p.kubeClient.CoreV1().Pods(p.namespace).Delete(ctx, hPodName, metav1.DeleteOptions{})
		if e != nil {
			log.Error(e, "Unable to delete the helper pod")
		}
	}()

	//Wait for the helper pod to complete it job and exit
	var checkPod *corev1.Pod
	for i := 0; i < CmdTimeoutCounts; i++ {
		var err error
		checkPod, err = p.kubeClient.CoreV1().Pods(p.namespace).Get(ctx, hPodName, metav1.GetOptions{})
		if err != nil {
			return err
		} else if checkPod.Status.Phase == corev1.PodSucceeded {
			return nil
		} else {
			// Currently we use `RestartPolicyNever`, if this changes we may need a different logic here, ex: x many restarts.
			if checkPod.Spec.RestartPolicy == corev1.RestartPolicyNever && checkPod.Status.Phase == corev1.PodFailed {
				return errors.Errorf("pod %v has failed: %s", checkPod.Name, describeHelperPodStatus(checkPod))
			}
		}
		time.Sleep(1 * time.Second)
	}
	if checkPod == nil {
		return errors.Errorf("create process timeout after %v seconds", CmdTimeoutCounts)
	}
	return errors.Errorf("create process timeout after %v seconds, pod %v status: %s",
		CmdTimeoutCounts, checkPod.Name, describeHelperPodStatus(checkPod))
}

// describeHelperPodStatus summarizes why a helper pod has not succeeded, i.e.
// the pod phase, the reason it could not be scheduled and the state of its
// containers. The helper pod is deleted once it exits or times out, so this
// is the only place where the cause of a failure is visible to the user.
func describeHelperPodStatus(hPod *corev1.Pod) string {
	details := []string{fmt.Sprintf("phase %v", hPod.Status.Phase)}
	if hPod.Status.Reason != "" || hPod.Status.Message != "" {
		details = append(details, formatReason("pod", hPod.Status.Reason, hPod.Status.Message))
	}
	for _, condition := range hPod.Status.Conditions {
		if condition.Type == corev1.PodScheduled && condition.Status == corev1.ConditionFalse {
			details = append(details, formatReason("not scheduled", condition.Reason, condition.Message))
		}
	}
	for _, status := range hPod.Status.ContainerStatuses {
		if waiting := status.State.Waiting; waiting != nil {
			details = append(details, formatReason(
				fmt.Sprintf("container %v waiting", status.Name), waiting.Reason, waiting.Message))
		}
		if terminated := status.State.Terminated; terminated != nil {
			details = append(details, formatReason(
				fmt.Sprintf("container %v terminated with exit code %d", status.Name, terminated.ExitCode),
				terminated.Reason, terminated.Message))
		}
	}
	return strings.Join(details, "; ")
}

// formatReason appends the non-empty reason and message to the given prefix.
// The message is collapsed into a single line and capped, as it may be the
// tail of the helper command's output and ends up in the PVC/PV events.
func formatReason(prefix, reason, message string) string {
	if reason != "" {
		prefix += ": " + reason
	}
	if message = strings.Join(strings.Fields(message), " "); message != "" {
		if len(message) > maxHelperPodMessageLen {
			cut := maxHelperPodMessageLen
			// Do not split a multi-byte character.
			for cut > 0 && !utf8.RuneStart(message[cut]) {
				cut--
			}
			message = message[:cut] + "..."
		}
		prefix += ": " + message
	}
	return prefix
}
