package tests

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	deploy "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/apps/v1/deployment"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/container"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/event"
	ns "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/namespace"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/node"
	pv "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/persistentvolume"
	pvc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/persistentvolumeclaim"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/pod"
	pts "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/podtemplatespec"
	svc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/service"
	k8svolume "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/volume"
	sc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/storage/v1/storageclass"
	kubeclient "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/client"
)

const maxRetry = 18

// Options holds the args used for exec'ing into the pod
type Options struct {
	podName   string
	container string
	namespace string
	cmd       []string
}

// NDM Path-Filter options
type PathFilterOption string

// Operations provides clients amd methods to perform operations
type Operations struct {
	KubeClient     *kubeclient.Client
	NodeClient     *v1alpha1.Kubeclient
	EventClient    *event.KubeClient
	PodClient      *pod.KubeClient
	PVCClient      *pvc.Kubeclient
	PVClient       *pv.Kubeclient
	SCClient       *sc.Kubeclient
	NSClient       *ns.Kubeclient
	SVCClient      *svc.Kubeclient
	DeployClient   *deploy.Kubeclient
	KubeConfigPath string
	NameSpace      string
	Config         interface{}
}

// OperationsOptions abstracts creating an
// instance of operations
type OperationsOptions func(*Operations)

// WithKubeConfigPath sets the kubeConfig path
// against operations instance
func WithKubeConfigPath(path string) OperationsOptions {
	return func(ops *Operations) {
		ops.KubeConfigPath = path
	}
}

// NewOperations returns a new instance of kubeclient meant for
// cstor volume replica operations
func NewOperations(opts ...OperationsOptions) *Operations {
	ops := &Operations{}
	for _, o := range opts {
		o(ops)
	}
	ops.withDefaults()
	return ops
}

// NewOptions returns the new instance of Options
func NewOptions() *Options {
	return new(Options)
}

// WithPodName fills the podName field in Options struct
func (o *Options) WithPodName(name string) *Options {
	o.podName = name
	return o
}

// WithNamespace fills the namespace field in Options struct
func (o *Options) WithNamespace(ns string) *Options {
	o.namespace = ns
	return o
}

// WithContainer fills the container field in Options struct
func (o *Options) WithContainer(container string) *Options {
	o.container = container
	return o
}

// WithCommand fills the cmd field in Options struct
func (o *Options) WithCommand(cmd ...string) *Options {
	o.cmd = cmd
	return o
}

// withDefaults sets the default options
// of operations instance
func (ops *Operations) withDefaults() {
	if ops.KubeClient == nil {
		ops.KubeClient = kubeclient.New(kubeclient.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.NSClient == nil {
		ops.NSClient = ns.NewKubeClient(ns.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.EventClient == nil {
		ops.EventClient = event.NewKubeClient(event.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.PodClient == nil {
		ops.PodClient = pod.NewKubeClient(pod.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.PVCClient == nil {
		ops.PVCClient = pvc.NewKubeClient(pvc.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.PVClient == nil {
		ops.PVClient = pv.NewKubeClient(pv.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.SCClient == nil {
		ops.SCClient = sc.NewKubeClient(sc.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.DeployClient == nil {
		ops.DeployClient = deploy.NewKubeClient(deploy.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.NodeClient == nil {
		ops.NodeClient = v1alpha1.NewKubeClient(v1alpha1.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.SVCClient == nil {
		ops.SVCClient = svc.NewKubeClient(svc.WithKubeConfigPath(ops.KubeConfigPath))
	}
}

// CheckPodStatusEventually gives the phase of the pod eventually
func (ops *Operations) CheckPodStatusEventually(namespace, podName string, expectedPodPhase corev1.PodPhase) corev1.PodPhase {
	var pod *corev1.Pod
	var err error
	for i := 0; i < maxRetry; i++ {
		pod, err = ops.PodClient.
			WithNamespace(namespace).
			Get(context.TODO(), podName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		if pod.Status.Phase == expectedPodPhase {
			return pod.Status.Phase
		}
		time.Sleep(5 * time.Second)
	}
	return pod.Status.Phase
}

// GetPodRunningCountEventually gives the number of pods running eventually
func (ops *Operations) GetPodRunningCountEventually(namespace, lselector string, expectedPodCount int) int {
	var podCount int
	for i := 0; i < maxRetry; i++ {
		podCount = ops.GetPodRunningCount(namespace, lselector)
		if podCount == expectedPodCount {
			return podCount
		}
		time.Sleep(5 * time.Second)
	}
	return podCount
}

// GetPodRunningCount gives number of pods running currently
func (ops *Operations) GetPodRunningCount(namespace, lselector string) int {
	pods, err := ops.PodClient.
		WithNamespace(namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	return pod.
		ListBuilderForAPIList(pods).
		WithFilter(pod.IsRunning()).
		List().
		Len()
}

// GetPodCount gives number of current pods
func (ops *Operations) GetPodCount(namespace, lselector string) int {
	pods, err := ops.PodClient.
		WithNamespace(namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	return pod.
		ListBuilderForAPIList(pods).
		List().
		Len()
}

// GetReadyNodes gives cstorvolumereplica healthy count currently based on selecter
func (ops *Operations) GetReadyNodes() *corev1.NodeList {
	nodes, err := ops.NodeClient.
		List(metav1.ListOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	return v1alpha1.NewListBuilder().
		WithAPIList(nodes).
		WithFilter(v1alpha1.IsReady()).
		List().
		ToAPIList()
}

// IsPVCBound checks if the pvc is bound or not
func (ops *Operations) IsPVCBound(namespace, pvcName string) bool {
	volume, err := ops.PVCClient.WithNamespace(namespace).
		Get(context.TODO(), pvcName, metav1.GetOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	return pvc.NewForAPIObject(volume).IsBound()
}

// IsPVCBoundEventually checks if the pvc is bound or not eventually
func (ops *Operations) IsPVCBoundEventually(namespace string, pvcName string) bool {
	return Eventually(func() bool {
		volume, err := ops.PVCClient.WithNamespace(namespace).
			Get(context.TODO(), pvcName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		return pvc.NewForAPIObject(volume).IsBound()
	},
		90, 5).
		Should(BeTrue())
}

// VerifyCapacity checks if the pvc capacity has been updated
func (ops *Operations) VerifyCapacity(namespace, pvcName, capacity string) bool {
	return Eventually(func() bool {
		volume, err := ops.PVCClient.WithNamespace(namespace).
			Get(context.TODO(), pvcName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		actualCapacity := volume.Status.Capacity[corev1.ResourceStorage]
		desiredCapacity, _ := resource.ParseQuantity(capacity)
		return (desiredCapacity.Cmp(actualCapacity) == 0)
	},
		90, 5).
		Should(BeTrue())
}

// PodDeleteCollection deletes all the pods in a namespace matched the given
// labelselector
func (ops *Operations) PodDeleteCollection(ns string, lopts metav1.ListOptions) error {
	deletePolicy := metav1.DeletePropagationForeground
	dopts := &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	return ops.PodClient.WithNamespace(ns).DeleteCollection(context.TODO(), lopts, dopts)
}

// IsPodRunningEventually return true if the pod comes to running state
func (ops *Operations) IsPodRunningEventually(namespace, podName string) bool {
	return Eventually(func() bool {
		p, err := ops.PodClient.
			WithNamespace(namespace).
			Get(context.TODO(), podName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		return pod.NewForAPIObject(p).
			IsRunning()
	},
		90, 5).
		Should(BeTrue())
}

// ExecuteCMDEventually executes the command on pod container
// and returns stdout
func (ops *Operations) ExecuteCMDEventually(
	podObj *corev1.Pod,
	containerName,
	cmd string,
	expectStdout bool,
) string {
	var err error
	output := &pod.ExecOutput{}
	podName := podObj.Name
	namespace := podObj.Namespace
	status := ops.IsPodRunningEventually(namespace, podName)
	Expect(status).To(Equal(true),
		"while checking the status of pod {%s} in namespace {%s}",
		podName,
		namespace,
	)
	for i := 0; i < maxRetry; i++ {
		output, err = ops.PodClient.WithNamespace(namespace).
			Exec(
				podName,
				&corev1.PodExecOptions{
					Command: []string{
						"/bin/sh",
						"-c",
						cmd,
					},
					Container: containerName,
					Stdin:     false,
					Stdout:    true,
					Stderr:    true,
				},
			)
		Expect(err).ShouldNot(
			HaveOccurred(),
			"failed to execute command {%s} on pod {%s} namespace {%s}",
			cmd,
			podName,
			namespace,
		)
		// If caller pass expectStdout as false return from here
		if !expectStdout {
			return ""
		}
		if output.Stdout != "" {
			return output.Stdout
		}
		time.Sleep(5 * time.Second)
	}
	err = errors.Errorf(
		"failed to execute cmd %s on pod %s",
		cmd,
		podName,
	)
	Expect(err).To(BeNil(),
		"failed to execute cmd {%s} on pod {%s} in namespace {%s} stdout {%s}",
		cmd,
		podName,
		namespace,
		output.Stdout,
	)
	return ""
}

// IsPVCDeleted tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func (ops *Operations) IsPVCDeleted(pvcName, namespace string) bool {
	_, err := ops.PVCClient.WithNamespace(namespace).
		Get(context.TODO(), pvcName, metav1.GetOptions{})
	return k8serrors.IsNotFound(err)
}

// IsPVCDeletedEventually tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func (ops *Operations) IsPVCDeletedEventually(pvcName, namespace string) bool {
	return Eventually(func() bool {
		_, err := ops.PVCClient.WithNamespace(namespace).
			Get(context.TODO(), pvcName, metav1.GetOptions{})
		return k8serrors.IsNotFound(err)
	},
		90, 5).
		Should(BeTrue())
}

// IsPVDeleted tries to get the deleted pvc
// and returns true if PV is not found
// else returns false
func (ops *Operations) IsPVDeleted(pvName string) bool {
	_, err := ops.PVClient.
		Get(context.TODO(), pvName, metav1.GetOptions{})
	return k8serrors.IsNotFound(err)
}

// IsPVDeletedEventually tries to get the deleted pvc
// and returns true if PV is not found
// else returns false
func (ops *Operations) IsPVDeletedEventually(pvName string) bool {
	return Eventually(func() bool {
		_, err := ops.PVClient.
			Get(context.TODO(), pvName, metav1.GetOptions{})
		return k8serrors.IsNotFound(err)
	},
		90, 5).
		Should(BeTrue())
}

// IsPodDeletedEventually checks if the pod is deleted or not eventually
func (ops *Operations) IsPodDeletedEventually(namespace, podName string) bool {
	return Eventually(func() bool {
		_, err := ops.PodClient.
			WithNamespace(namespace).
			Get(context.TODO(), podName, metav1.GetOptions{})
		return k8serrors.IsNotFound(err)
	},
		90, 5).
		Should(BeTrue())
}

// GetPVNameFromPVCName gives the pv name for the given pvc
func (ops *Operations) GetPVNameFromPVCName(namespace, pvcName string) string {
	p, err := ops.PVCClient.WithNamespace(namespace).Get(context.TODO(), pvcName, metav1.GetOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	return p.Spec.VolumeName
}

// GetNodeAffinityLabelKeysFromPv returns the label keys for NodeSelector MatchExpressions in a PV.
func (ops *Operations) GetNodeAffinityLabelKeysFromPv(pvName string) ([]string, error) {
	pv, err := ops.PVClient.Get(context.TODO(), pvName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var nodeAffinityLabelKeys []string

	for _, selectorTerm := range pv.Spec.NodeAffinity.Required.NodeSelectorTerms {
		for _, matchExpression := range selectorTerm.MatchExpressions {
			nodeAffinityLabelKeys = append(nodeAffinityLabelKeys, matchExpression.Key)
		}
	}

	return nodeAffinityLabelKeys, nil
}

func (ops *Operations) GetNdmConfigMap(
	clientset *kubernetes.Clientset,
	namespace string,
	ndmConfigLabelSelector string,
) (*corev1.ConfigMap, error) {
	configMapList, err := clientset.CoreV1().ConfigMaps(namespace).
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: ndmConfigLabelSelector,
		})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list ConfigMaps in {%s} namespace", namespace)
	}

	cmLength := len(configMapList.Items)
	if cmLength != 1 {
		return nil, errors.Errorf("expected 1 ConfigMap with LabelSelector {%s} in namespace {%s}, but got {%v}",
			ndmConfigLabelSelector, namespace, cmLength)
	}

	return &(configMapList.Items[0]), nil
}

// ExecPod executes arbitrary command inside the pod
func (ops *Operations) ExecPod(opts *Options) (string, string, error) {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
		err     error
	)
	config, err := ops.KubeClient.GetConfigForPathOrDirect()
	if err != nil {
		return "", "", errors.Errorf("error while getting config for exec'ing into pod: %v", err)
	}

	cset, err := ops.KubeClient.Clientset()
	if err != nil {
		return "", "", errors.Errorf("while getting clientset for exec'ing into pod: %v", err)
	}
	req := cset.
		CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Name(opts.podName).
		Namespace(opts.namespace).
		SubResource("exec").
		Param("container", opts.container).
		VersionedParams(&corev1.PodExecOptions{
			Container: opts.container,
			Command:   opts.cmd,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("error while creating Executor: %v", err)
	}

	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdout: &execOut,
		Stderr: &execErr,
		Tty:    false,
	})
	if err != nil {
		return execOut.String(), execErr.String(), errors.Errorf("error in Stream: %v", err)
	}

	return execOut.String(), execErr.String(), nil
}

// GetPodCompletedCountEventually gives the number of pods running eventually
func (ops *Operations) GetPodCompletedCountEventually(namespace, lselector string, expectedPodCount int) int {
	var podCount int
	for i := 0; i < maxRetry; i++ {
		podCount = ops.GetPodCompletedCount(namespace, lselector)
		if podCount == expectedPodCount {
			return podCount
		}
		time.Sleep(5 * time.Second)
	}
	return podCount
}

// GetPodCompletedCount gives number of pods running currently
func (ops *Operations) GetPodCompletedCount(namespace, lselector string) int {
	pods, err := ops.PodClient.
		WithNamespace(namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	return pod.
		ListBuilderForAPIList(pods).
		WithFilter(pod.IsCompleted()).
		List().
		Len()
}

// GetPodList gives list of running pods for given namespace + label
func (ops *Operations) GetPodList(namespace, lselector string, predicateList pod.PredicateList) *pod.PodList {
	pods, err := ops.PodClient.
		WithNamespace(namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	return pod.
		ListBuilderForAPIList(pods).
		WithFilter(predicateList...).
		List()
}

// GetPodCountEventually returns the no.of pods exists with specified labelselector
func (ops *Operations) GetPodCountEventually(
	namespace, lselector string,
	predicateList pod.PredicateList, expectedCount int) int {
	var podCount int
	for i := 0; i < maxRetry; i++ {
		podList := ops.GetPodList(namespace, lselector, predicateList)
		podCount = podList.Len()
		if podCount == expectedCount {
			return podCount
		}
		time.Sleep(5 * time.Second)
	}
	return podCount
}

// DeletePersistentVolumeClaim deletes PVC from cluster based on provided
// argument
func (ops *Operations) DeletePersistentVolumeClaim(name, namespace string) {
	err := ops.PVCClient.WithNamespace(namespace).Delete(context.TODO(), name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil())
}

// GetSVCClusterIP returns list of IP address of the services, having given label and namespace
func (ops *Operations) GetSVCClusterIP(ns, lselector string) ([]string, error) {
	addr := []string{}
	svclist, err := ops.SVCClient.
		WithNamespace(ns).
		List(
			metav1.ListOptions{
				LabelSelector: lselector,
			},
		)
	if err != nil {
		return addr, errors.Errorf("failed to get service err=%v", err)
	}

	if len(svclist.Items) == 0 {
		return addr, errors.Errorf("no service with label=%s in ns=%s", lselector, ns)
	}

	for _, s := range svclist.Items {
		if len(s.Spec.ClusterIP) != 0 {
			addr = append(addr, s.Spec.ClusterIP+":"+strconv.FormatInt(int64(s.Spec.Ports[0].Port), 10))
		}
	}

	return addr, nil
}

func (ops *Operations) BuildAndDeployBusyBoxPod(
	appName, pvcName, namespace string,
	labels map[string]string) (*appsv1.Deployment, error) {
	var err error
	appDeployment, err := deploy.NewBuilder().
		WithName(appName).
		WithNamespace(namespace).
		WithLabelsNew(labels).
		WithSelectorMatchLabelsNew(labels).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(labels).
				WithContainerBuilders(
					container.NewBuilder().
						WithImage("busybox").
						WithName("busybox").
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithCommandNew(
							[]string{
								"sh",
								"-c",
								"date > /mnt/cstore1/date.txt; sync; sleep 5; sync; tail -f /dev/null;",
							},
						).
						WithVolumeMountsNew(
							[]corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "datavol1",
									MountPath: "/mnt/cstore1",
								},
							},
						),
				).
				WithTerminationGracePeriodSeconds(5).
				WithVolumeBuilders(
					k8svolume.NewBuilder().
						WithName("datavol1").
						WithPVCSource(pvcName),
				),
		).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build busybox: %s deployment", appName)
	}

	appDeployment, err = ops.DeployClient.WithNamespace(namespace).Create(context.TODO(), appDeployment)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create busybox %s deployment in namespace %s", appName, namespace)
	}
	return appDeployment, nil
}

// BuildPersistentVolumeClaim builds the PVC object
func BuildPersistentVolumeClaim(namespace, pvcName, scName, capacity string, accessModes []corev1.PersistentVolumeAccessMode) (*corev1.PersistentVolumeClaim, error) {
	return pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(namespace).
		WithStorageClass(scName).
		WithAccessModes(accessModes).
		WithCapacity(capacity).Build()
}

// BuildPod builds the pod object
func BuildPod(namespace, podName, pvcName string, labelselector map[string]string) (*corev1.Pod, error) {
	return pod.NewBuilder().
		WithName(podName).
		WithNamespace(namespace).
		WithLabels(labelselector).
		WithContainerBuilder(
			container.NewBuilder().
				WithName("busybox").
				WithImage("busybox").
				WithCommandNew(
					[]string{
						"/bin/sh",
					},
				).
				WithArgumentsNew(
					[]string{
						"-c",
						"sleep 3600",
					},
				).
				WithVolumeMountsNew(
					[]corev1.VolumeMount{
						{
							Name:      "demo-vol1",
							MountPath: "/mnt/store1",
						},
					},
				),
		).
		WithVolumeBuilder(
			k8svolume.NewBuilder().
				WithName("demo-vol1").
				WithPVCSource(pvcName),
		).
		WithTerminationGracePeriodSeconds(5).
		Build()
}

// createDeploymentWhichConsumesHostpath creates a single-replica Deployment whose Pod consumes hostpath PVC.
func (ops *Operations) createDeploymentWhichConsumesHostpath(namePrefix, namespace, pvcName string) (*appsv1.Deployment, error) {
	labelSelector := map[string]string{
		"app":  namePrefix,
		"role": "test",
	}
	deployment, err := deploy.NewBuilder().
		WithGenerateName(namePrefix).
		WithNamespace(namespace).
		WithLabelsNew(labelSelector).
		WithSelectorMatchLabelsNew(labelSelector).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(labelSelector).
				WithContainerBuildersNew(
					container.NewBuilder().
						WithName("busybox").
						WithImage("busybox").
						WithCommandNew(
							[]string{
								"sleep",
								"3600",
							},
						).
						WithVolumeMountsNew(
							[]corev1.VolumeMount{
								{
									Name:      "demo-vol1",
									MountPath: "/mnt/store1",
								},
							},
						),
				).
				WithTerminationGracePeriodSeconds(5).
				WithVolumeBuilders(
					k8svolume.NewBuilder().
						WithName("demo-vol1").
						WithPVCSource(pvcName),
				),
		).
		Build()
	if err != nil {
		return nil, err
	}

	return ops.DeployClient.WithNamespace(namespace).Create(context.TODO(), deployment)
}

// isLabelSelectorsEqual compares two arrays of label selector keys.
func isLabelSelectorsEqual(request, result []string) bool {
	if len(request) != len(result) {
		return false
	}

	ch := make(chan struct{}, 2)
	collectFrequency := func(labelKeys []string, freq *map[string]int) {
		for _, elem := range labelKeys {
			(*freq)[elem]++
		}

		ch <- struct{}{}
	}

	// Maps to hold the frequency of strings in the string slices.
	freqRequest := make(map[string]int)
	freqResult := make(map[string]int)

	go collectFrequency(request, &freqRequest)
	go collectFrequency(result, &freqResult)

	for i := 0; i < 2; i++ {
		<-ch
	}

	// Compare frequencies
	for key, countRequest := range freqRequest {
		countResult, ok := freqResult[key]
		if !ok || countRequest != countResult {
			return false
		}
	}

	return true
}
