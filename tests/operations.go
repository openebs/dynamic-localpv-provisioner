/*
Copyright 2019 The OpenEBS Authors

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

package tests

import (
	"bytes"
	"context"
	"fmt"
	//"sort"
	"strconv"
	"time"

	. "github.com/onsi/gomega"
	bd "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	kubeclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	node "github.com/openebs/maya/pkg/kubernetes/node/v1alpha1"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	templatefuncs "github.com/openebs/maya/pkg/templatefuncs/v1alpha1"
	unstruct "github.com/openebs/maya/pkg/unstruct/v1alpha2"
	result "github.com/openebs/maya/pkg/upgrade/result/v1alpha1"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	deploy "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/apps/v1/deployment"
	container "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/container"
	pv "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/persistentvolume"
	pvc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/persistentvolumeclaim"
	pod "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/pod"
	pts "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/podtemplatespec"
	k8svolume "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/volume"
	sc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/storage/v1/storageclass"
)

const (
	maxRetry = 30
)

type bdcExitStatus string

const (
	deleted bdcExitStatus = "deleted"
	pending bdcExitStatus = "pending"
	bound   bdcExitStatus = "bound"
	invalid bdcExitStatus = "invalid"
)

/*
type SortBDC struct {
	bdcList *apis.BlockDeviceClaimList
}

func (s SortBDC) Len() int {
	return len(s.bdcList.Items)
}

func (s SortBDC) Swap(i, j int) {
	s.bdcList.Items[i], s.bdcList.Items[j] = s.bdcList.Items[j], s.bdcList.Items[i]
}

func (s SortBDC) Less(i, j int) bool {
	return s.bdcList.Items[i].ObjectMeta.CreationTimestamp.Time.Before(s.bdcList.Items[j].ObjectMeta.CreationTimestamp.Time)
}
*/

// Options holds the args used for exec'ing into the pod
type Options struct {
	podName   string
	container string
	namespace string
	cmd       []string
}

// Operations provides clients amd methods to perform operations
type Operations struct {
	KubeClient     *kubeclient.Client
	NodeClient     *node.Kubeclient
	PodClient      *pod.KubeClient
	PVCClient      *pvc.Kubeclient
	PVClient       *pv.Kubeclient
	SCClient       *sc.Kubeclient
	NSClient       *ns.Kubeclient
	SVCClient      *svc.Kubeclient
	URClient       *result.Kubeclient
	UnstructClient *unstruct.Kubeclient
	DeployClient   *deploy.Kubeclient
	BDClient       *bd.Kubeclient
	BDCClient      *bdc.Kubeclient
	KubeConfigPath string
	NameSpace      string
	Config         interface{}
}

// SPCConfig provides config to create cstor pools
type SPCConfig struct {
	Name      string
	DiskType  string
	PoolType  string
	PoolCount int
	// OverProvisioning field is deprecated and not honoured
	IsOverProvisioning bool

	IsThickProvisioning bool
}

// PVCConfig provides config to create PersistentVolumeClaim
type PVCConfig struct {
	Name        string
	Namespace   string
	SCName      string
	Capacity    string
	AccessModes []corev1.PersistentVolumeAccessMode
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
	if ops.URClient == nil {
		ops.URClient = result.NewKubeClient(result.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.UnstructClient == nil {
		ops.UnstructClient = unstruct.NewKubeClient(unstruct.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.DeployClient == nil {
		ops.DeployClient = deploy.NewKubeClient(deploy.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.BDClient == nil {
		ops.BDClient = bd.NewKubeClient(bd.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.NodeClient == nil {
		ops.NodeClient = node.NewKubeClient(node.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.BDCClient == nil {
		ops.BDCClient = bdc.NewKubeClient(bdc.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.SVCClient == nil {
		ops.SVCClient = svc.NewKubeClient(svc.WithKubeConfigPath(ops.KubeConfigPath))
	}
}

// CheckPodStatusEventually gives the phase of the pod eventually
func (ops *Operations) CheckPodStatusEventually(namespace, podName string, expectedPodPhase corev1.PodPhase) bool {
	for i := 0; i < maxRetry; i++ {
		pod, err := ops.PodClient.
			WithNamespace(namespace).
			Get(context.TODO(), podName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		if pod.Status.Phase == expectedPodPhase {
			return true
		}
		time.Sleep(5 * time.Second)
	}
	return false
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
	return node.
		NewListBuilder().
		WithAPIList(nodes).
		WithFilter(node.IsReady()).
		List().
		ToAPIList()
}

// IsPVCBound checks if the pvc is bound or not
func (ops *Operations) IsPVCBound(pvcName string) bool {
	volume, err := ops.PVCClient.
		Get(context.TODO(), pvcName, metav1.GetOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	return pvc.NewForAPIObject(volume).IsBound()
}

// IsPVCBoundEventually checks if the pvc is bound or not eventually
func (ops *Operations) IsPVCBoundEventually(pvcName string) bool {
	return Eventually(func() bool {
		volume, err := ops.PVCClient.
			Get(context.TODO(), pvcName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		return pvc.NewForAPIObject(volume).IsBound()
	},
		120, 10).
		Should(BeTrue())
}

// VerifyCapacity checks if the pvc capacity has been updated
func (ops *Operations) VerifyCapacity(pvcName, capacity string) bool {
	return Eventually(func() bool {
		volume, err := ops.PVCClient.
			Get(context.TODO(), pvcName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		actualCapacity := volume.Status.Capacity[corev1.ResourceStorage]
		desiredCapacity, _ := resource.ParseQuantity(capacity)
		return (desiredCapacity.Cmp(actualCapacity) == 0)
	},
		120, 10).
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
		150, 10).
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

// RestartPodEventually restarts the pod and return
func (ops *Operations) RestartPodEventually(podObj *corev1.Pod) error {
	status := ops.IsPodRunningEventually(podObj.Namespace, podObj.Name)
	if !status {
		return errors.Errorf(
			"while checking the status of pod {%s} in namespace {%s} before restarting",
			podObj.Name,
			podObj.Namespace,
		)
	}

	err := ops.PodClient.WithNamespace(podObj.Namespace).
		Delete(context.TODO(), podObj.Name, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err,
			"failed to delete pod {%s} in namespace {%s}",
			podObj.Name,
			podObj.Namespace,
		)
	}

	status = ops.IsPodDeletedEventually(podObj.Namespace, podObj.Name)
	if !status {
		return errors.Errorf(
			"while checking termination of pod {%s} in namespace {%s}",
			podObj.Name,
			podObj.Namespace,
		)
	}
	return nil
}

// IsPVCDeleted tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func (ops *Operations) IsPVCDeleted(pvcName, namespace string) bool {
	_, err := ops.PVCClient.WithNamespace(namespace).
		Get(context.TODO(), pvcName, metav1.GetOptions{})
	return isNotFound(err)
}

// IsPVCDeletedEventually tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func (ops *Operations) IsPVCDeletedEventually(pvcName, namespace string) bool {
	return Eventually(func() bool {
		_, err := ops.PVCClient.WithNamespace(namespace).
			Get(context.TODO(), pvcName, metav1.GetOptions{})
		return isNotFound(err)
	},
		120, 2).
		Should(BeTrue())
}

// IsPVDeleted tries to get the deleted pvc
// and returns true if PV is not found
// else returns false
func (ops *Operations) IsPVDeleted(pvName string) bool {
	_, err := ops.PVClient.
		Get(context.TODO(), pvName, metav1.GetOptions{})
	return isNotFound(err)
}

// IsPVDeletedEventually tries to get the deleted pvc
// and returns true if PV is not found
// else returns false
func (ops *Operations) IsPVDeletedEventually(pvName string) bool {
	return Eventually(func() bool {
		_, err := ops.PVClient.
			Get(context.TODO(), pvName, metav1.GetOptions{})
		return isNotFound(err)
	},
		120, 2).
		Should(BeTrue())
}

// IsPodDeletedEventually checks if the pod is deleted or not eventually
func (ops *Operations) IsPodDeletedEventually(namespace, podName string) bool {
	return Eventually(func() bool {
		_, err := ops.PodClient.
			WithNamespace(namespace).
			Get(context.TODO(), podName, metav1.GetOptions{})
		return isNotFound(err)
	},
		120, 10).
		Should(BeTrue())
}

// GetPVNameFromPVCName gives the pv name for the given pvc
func (ops *Operations) GetPVNameFromPVCName(pvcName string) string {
	p, err := ops.PVCClient.
		Get(context.TODO(), pvcName, metav1.GetOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	return p.Spec.VolumeName
}

// isNotFound returns true if the original
// cause of error was due to castemplate's
// not found error or kubernetes not found
// error
func isNotFound(err error) bool {
	switch err := errors.Cause(err).(type) {
	case *templatefuncs.NotFoundError:
		return true
	default:
		return k8serrors.IsNotFound(err)
	}
}

// IsBDCDeletedEventually tries to get the deleted BDC
// and returns true if BDC is not found
// else returns false
func (ops *Operations) IsBDCDeletedEventually(bdcName, namespace string) bool {
	return Eventually(func() bool {
		_, err := ops.BDCClient.WithNamespace(namespace).
			Get(context.TODO(), bdcName, metav1.GetOptions{})
		return isNotFound(err)
	},
		120, 2).
		Should(BeTrue())
}

// Sorts BDC in the descending order of their creation timestamp and
// returns the name of the BDC created at the latest timestamp
/*
func (ops *Operations) GetLatestCreatedBDCName(namespace string) string {
	bdcList, err := ops.BDCClient.WithNamespace(namespace).List(context.TODO(), metav1.ListOptions{})
	Expect(err).To(
		BeNil(),
		"when GET-ing BDC in namespace {%s}",
		ops.NameSpace,
	)

	sortableBDCList := SortBDC{
		bdcList: bdcList,
	}
	sort.Sort(sort.Reverse(sortableBDCList))
	return sortableBDCList.bdcList.Items[0].ObjectMeta.Name
}
*/
/*
func (ops *Operations) GetBDNameFromBDCName(bdcName, namespace string) string {
	bdcObj, err := ops.BDCClient.WithNamespace(namespace).
		Get(context.TODO(), bdcName, metav1.GetOptions{})
	Expect(err).To(
		BeNil(),
		"when trying to get BDC {%s}",
		bdcName,
	)
	Expect(bdcObj.Status.Phase).To(
		Equal(apis.BlockDeviceClaimStatusDone),
		"when trying to check if a BD is bound to BDC {%s}",
		bdcName,
	)
	return bdcObj.Spec.BlockDeviceName
}
*/
// GetBDCCountEventually gets BDC resource count based on provided list option.
func (ops *Operations) GetBDCCountEventually(listOptions metav1.ListOptions, expectedBDCCount int, namespace string) int {
	var bdcCount int
	for i := 0; i < maxRetry; i++ {
		bdcAPIList, err := ops.BDCClient.WithNamespace(namespace).List(context.TODO(), listOptions)
		Expect(err).To(BeNil())
		bdcCount = len(bdcAPIList.Items)
		if bdcCount == expectedBDCCount {
			return bdcCount
		}
		time.Sleep(5 * time.Second)
	}
	return bdcCount
}

// IsFinalizerExistsOnBDC returns true if the object with provided name contains the finalizer.
func (ops *Operations) IsFinalizerExistsOnBDC(bdcName, finalizer string) bool {
	for i := 0; i < maxRetry; i++ {
		bdcObj, err := ops.BDCClient.Get(context.TODO(), bdcName, metav1.GetOptions{})
		Expect(err).To(BeNil())
		for _, f := range bdcObj.Finalizers {
			if f == finalizer {
				return true
			}
		}
		time.Sleep(5 * time.Second)
	}
	return false
}

func (ops *Operations) GetBDCStatusAfterAge(bdcName string, namespace string, untilAge time.Duration) bdcExitStatus {
	bdcObj, err := ops.BDCClient.WithNamespace(namespace).Get(context.TODO(), bdcName, metav1.GetOptions{})
	Expect(err).To(
		BeNil(),
		"when geting BDC {%s} initally to calculate Age",
		bdcName,
	)
	initialCreationTimestamp := bdcObj.CreationTimestamp.Time
	untilTimestamp := initialCreationTimestamp.Add(untilAge)

	time.Sleep(time.Until(untilTimestamp))

	bdcObj, err = ops.BDCClient.WithNamespace(namespace).Get(context.TODO(), bdcName, metav1.GetOptions{})
	finalCreationTimestamp := bdcObj.CreationTimestamp.Time
	if isNotFound(err) || finalCreationTimestamp.After(initialCreationTimestamp) {
		return deleted
	} else if bdcObj.Status.Phase == "Pending" {
		return pending
	} else if bdcObj.Status.Phase == "Bound" {
		return bound
	} else {
		return invalid
	}
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

	err = exec.Stream(remotecommand.StreamOptions{
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

// VerifyUpgradeResultTasksIsNotFail checks whether all the tasks in upgraderesult
// have success
func (ops *Operations) VerifyUpgradeResultTasksIsNotFail(namespace, lselector string) bool {
	urList, err := ops.URClient.
		WithNamespace(namespace).
		List(metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	for _, task := range urList.Items[0].Tasks {
		if task.Status == "Fail" {
			fmt.Printf("task : %v\n", task)
			return false
		}
	}
	return true
}

// GetBDCCount gets BDC resource count based on provided label selector
func (ops *Operations) GetBDCCount(lSelector, namespace string) int {
	bdcList, err := ops.BDCClient.
		WithNamespace(namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: lSelector})
	Expect(err).ShouldNot(HaveOccurred())
	return len(bdcList.Items)
}

// BuildAndCreatePVC builds and creates PersistentVolumeClaim in cluster
func (ops *Operations) BuildAndCreatePVC() *corev1.PersistentVolumeClaim {
	pvcConfig := ops.Config.(*PVCConfig)
	pvcObj, err := pvc.NewBuilder().
		WithName(pvcConfig.Name).
		WithNamespace(pvcConfig.Namespace).
		WithStorageClass(pvcConfig.SCName).
		WithAccessModes(pvcConfig.AccessModes).
		WithCapacity(pvcConfig.Capacity).Build()
	Expect(err).ShouldNot(
		HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcConfig.Name,
		pvcConfig.Namespace,
	)
	pvcObj, err = ops.PVCClient.WithNamespace(pvcConfig.Namespace).Create(context.TODO(), pvcObj)
	Expect(err).To(
		BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcConfig.Name,
		pvcConfig.Namespace,
	)
	return pvcObj
}

// DeletePersistentVolumeClaim deletes PVC from cluster based on provided
// argument
func (ops *Operations) DeletePersistentVolumeClaim(name, namespace string) {
	err := ops.PVCClient.WithNamespace(namespace).Delete(context.TODO(), name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil())
}

//GetSVCClusterIP returns list of IP address of the services, having given label and namespace
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
