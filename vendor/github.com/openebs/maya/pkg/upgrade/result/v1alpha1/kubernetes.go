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

package v1alpha1

import (
	"encoding/json"
	"strings"
	"sync"
	"text/template"
	"time"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/upgrade/v1alpha1/clientset/internalclientset"
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	errors "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

// getClientsetFunc is a typed function that
// abstracts fetching internal clientset
type getClientsetFunc func() (cs *clientset.Clientset, err error)

// getClientsetForPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (cs *clientset.Clientset, err error)

// listFunc is a typed function that abstracts
// listing upgrade result instances
type listFunc func(cs *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.UpgradeResultList, error)

// getFunc is a typed function that abstracts
// getting upgrade result instances
type getFunc func(cs *clientset.Clientset, name string, namespace string, opts metav1.GetOptions) (*apis.UpgradeResult, error)

// createFunc is a typed function that abstracts
// creating upgrade result instances
type createFunc func(cs *clientset.Clientset, upgradeResultObj *apis.UpgradeResult,
	namespace string) (*apis.UpgradeResult, error)

// patchFunc is a typed function that abstracts
// patching upgrade result instances
type patchFunc func(cs *clientset.Clientset, name string, pt types.PatchType, patchObj []byte,
	namespace string) (*apis.UpgradeResult, error)

// updateFunc is a typed function that abstracts
// updating upgrade result instances
type updateFunc func(cs *clientset.Clientset, updateObj *apis.UpgradeResult,
	namespace string) (*apis.UpgradeResult, error)

// CoreClient holds all the properties that are required to
// execute a K8s API call. All of these properties can be set
// once in the lifetime of the application. In other words,
// a CoreClient instance should ideally be a singleton.
type CoreClient struct {
	// clientset refers to upgraderesult clientset that can
	// make kubernetes API calls
	clientset *clientset.Clientset

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFunc
	list                listFunc
	get                 getFunc
	create              createFunc
	patch               patchFunc
	update              updateFunc
	getClientsetForPath getClientsetForPathFn
}

// Kubeclient enables kubernetes API operations
// on upgraderesult instance
type Kubeclient struct {
	// CoreClient has all the core kubernetes client
	// related options
	*CoreClient

	// namespace to use during CRUD operations against
	// upgraderesult resource
	namespace string
}

// KubeclientBuildOption is a typed function that abstracts
// building an instance of Kubeclient
type KubeclientBuildOption func(*Kubeclient)

// NewKubeClient returns a new instance of Kubeclient
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{CoreClient: &CoreClient{}}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

var (
	coreClientInst *CoreClient
	once           sync.Once
)

// KubeClientInstanceOrDie returns the singleton instance of Kubeclient
//
// NOTE:
//  Here singleton points to CoreClient instance only since a Kubeclient
// instance needs to change at runtime based on namespace. CoreClient's
// clientset instance is the only field that is needed to be initialized
// to consider Kubeclient as a singleton.
//
// NOTE:
//  In order to keep this logic more caller code friendly, this function
// is not named as CoreClientInstanceOrDie.
//
// Usage:
//  Caller code will use syntax(-es) as shown below:
//
// ```go
// import (
//  uresult "github.com/openebs/maya/pkg/upgrade/result/v1alpha1"
// )
//
// uresult.KubeClientInstanceOrDie().WithNamespace("my_ns").Get(...)
// uresult.KubeClientInstanceOrDie().WithNamespace("my_ns").Create(...)
// uresult.KubeClientInstanceOrDie().WithNamespace("my_ns").Update(...)
// uresult.KubeClientInstanceOrDie().WithNamespace("my_ns").List(...)
// ```
func KubeClientInstanceOrDie(opts ...KubeclientBuildOption) *Kubeclient {
	once.Do(func() {
		k := NewKubeClient(opts...)
		_, err := k.getClientOrCached()
		if err != nil {
			klog.Fatalf("failed to initialise kubeclient instance: {%v}", err)
		}
		coreClientInst = k.CoreClient
	})

	return &Kubeclient{CoreClient: coreClientInst}
}

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (cs *clientset.Clientset, err error) {
			config, err := client.New().GetConfigForPathOrDirect()
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}

	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (clients *clientset.Clientset, err error) {
			config, err := client.New(client.WithKubeConfigPath(kubeConfigPath)).GetConfigForPathOrDirect()
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}

	if k.list == nil {
		k.list = func(cs *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.UpgradeResultList, error) {
			return cs.OpenebsV1alpha1().UpgradeResults(namespace).List(opts)
		}
	}

	if k.get == nil {
		k.get = func(cs *clientset.Clientset, name, namespace string, opts metav1.GetOptions) (*apis.UpgradeResult, error) {
			return cs.OpenebsV1alpha1().UpgradeResults(namespace).Get(name, opts)
		}
	}

	if k.create == nil {
		k.create = func(cs *clientset.Clientset, upgradeResultObj *apis.UpgradeResult,
			namespace string) (*apis.UpgradeResult, error) {
			return cs.OpenebsV1alpha1().
				UpgradeResults(namespace).
				Create(upgradeResultObj)
		}
	}

	if k.patch == nil {
		k.patch = func(cs *clientset.Clientset, name string,
			pt types.PatchType, patchObj []byte,
			namespace string) (*apis.UpgradeResult, error) {
			return cs.OpenebsV1alpha1().
				UpgradeResults(namespace).
				Patch(name, pt, patchObj)
		}
	}

	if k.update == nil {
		k.update = func(cs *clientset.Clientset,
			upgradeResultObj *apis.UpgradeResult,
			namespace string) (*apis.UpgradeResult, error) {
			return cs.OpenebsV1alpha1().
				UpgradeResults(namespace).
				Update(upgradeResultObj)
		}
	}

}

// WithClientset sets the kubernetes clientset against
// the kubeclient instance
func WithClientset(c *clientset.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*clientset.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// WithNamespace sets namespace that should be used during
// kuberenets API calls against upgradeResult resource
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*clientset.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}
	cs, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil, err
	}
	k.clientset = cs
	return k.clientset, nil
}

// List returns a list of upgrade result
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*apis.UpgradeResultList, error) {
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cs, k.namespace, opts)
}

// Get returns an upgrade result instance from kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*apis.UpgradeResult, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get upgrade result: missing upgradeResult name")
	}
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cs, name, k.namespace, opts)
}

// CreateRaw creates an upgrade result instance
// and returns raw upgradeResult instance
func (k *Kubeclient) CreateRaw(upgradeResultObj *apis.UpgradeResult) ([]byte, error) {
	ur, err := k.Create(upgradeResultObj)
	if err != nil {
		return nil, err
	}
	return json.Marshal(ur)
}

// Create creates an upgrade result instance in kubernetes cluster
func (k *Kubeclient) Create(upgradeResultObj *apis.UpgradeResult) (*apis.UpgradeResult, error) {
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.create(cs, upgradeResultObj, k.namespace)
}

// Patch returns the patched upgrade result instance
func (k *Kubeclient) Patch(name string, pt types.PatchType,
	patchObj []byte) (*apis.UpgradeResult, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to patch upgrade result: missing upgradeResult name")
	}
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.patch(cs, name, pt, patchObj, k.namespace)
}

// Update returns the updated upgrade result instance
func (k *Kubeclient) Update(updateObj *apis.UpgradeResult) (*apis.UpgradeResult, error) {
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.update(cs, updateObj, k.namespace)
}

// UpgradeResultForTask enables update
// operation on upgrade result task instance
type UpgradeResultForTask struct {
	name      string
	namespace string
	task      *apis.UpgradeResultTask
}

// UpgradeResultForTaskOption defines the abstraction
// to build an update instance for upgrade result's task
type UpgradeResultForTaskOption func(*UpgradeResultForTask)

// WithTaskOwnerName sets the name of the upgrade
// result
func WithTaskOwnerName(name string) UpgradeResultForTaskOption {
	return func(u *UpgradeResultForTask) {
		u.name = name
	}
}

// WithTaskOwnerNamespace sets namespace where upgrade
// result is present
func WithTaskOwnerNamespace(namespace string) UpgradeResultForTaskOption {
	return func(u *UpgradeResultForTask) {
		u.namespace = namespace
	}
}

// WithTaskName sets the name of the
// task to be updated
func WithTaskName(name string) UpgradeResultForTaskOption {
	return func(u *UpgradeResultForTask) {
		u.task.Name = name
	}
}

// WithTaskStatus sets the current status
// of the task i.e. whether it has successfully
// completed or not
func WithTaskStatus(status string) UpgradeResultForTaskOption {
	return func(u *UpgradeResultForTask) {
		u.task.Status = status
	}
}

// WithTaskMessage sets the message for a
// particular task i.e. the message about its
// successful completion or failure
func WithTaskMessage(message string) UpgradeResultForTaskOption {
	return func(u *UpgradeResultForTask) {
		u.task.Message = message
	}
}

// WithTaskStartTime sets the time when the
// task started to execute
func WithTaskStartTime(startTime time.Time) UpgradeResultForTaskOption {
	return func(u *UpgradeResultForTask) {
		u.task.StartTime = &metav1.Time{startTime}
	}
}

// WithTaskEndTime sets the time when the
// task finished execution
func WithTaskEndTime(endTime time.Time) UpgradeResultForTaskOption {
	return func(u *UpgradeResultForTask) {
		u.task.EndTime = &metav1.Time{endTime}
	}
}

// WithTaskRetries sets the no of times that
// a runtask has retried executing a particular task
func WithTaskRetries(retries int) UpgradeResultForTaskOption {
	return func(u *UpgradeResultForTask) {
		u.task.Retries = retries
	}
}

// NewUpgradeResultForTask returns a new instance of updateUpgradeResult
// meant for updating an upgrade result instance
func NewUpgradeResultForTask(opts ...UpgradeResultForTaskOption) *UpgradeResultForTask {
	u := &UpgradeResultForTask{
		task: &apis.UpgradeResultTask{},
	}
	for _, o := range opts {
		o := o
		o(u)
	}
	return u
}

// UpdateTasks is a template function exposed for
// updating an upgrade result instance
func UpdateTasks(opts ...UpgradeResultForTaskOption) error {
	new := NewUpgradeResultForTask(opts...)
	if new.name == "" {
		return errors.New("failed to update upgrade result tasks: missing upgrade result name")
	}
	// First get the desired upgrade result instance
	k := NewKubeClient()
	k.namespace = new.namespace
	existing, err := k.Get(new.name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to update upgrade result tasks: upgrade result name {%s} namespace {%s}",
			new.name,
			new.namespace,
		)
	}
	// Iterate over the upgrade result tasks to check if the
	// desired task to be updated exists or not,
	// if exists then update the task instance with the given values.
	for i, existingTask := range existing.Tasks {
		i := i
		existingTask := existingTask
		if existingTask.Name == new.task.Name {
			existingTask = *new.task
		}
		existing.Tasks[i] = existingTask
	}
	// Update the upgrade result instance with
	// the provided values
	_, err = k.Update(existing)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to update upgrade result tasks: upgrade result name {%s} namespace {%s}",
			existing.Name,
			existing.Namespace,
		)
	}
	return nil
}

// TemplateFunctions exposes a few functions as
// go template functions to be used for upgrade result
func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"upgradeResultUpdateTasks":            UpdateTasks,
		"upgradeResultWithTaskOwnerName":      WithTaskOwnerName,
		"upgradeResultWithTaskOwnerNamespace": WithTaskOwnerNamespace,
		"upgradeResultWithTaskName":           WithTaskName,
		"upgradeResultWithTaskStatus":         WithTaskStatus,
		"upgradeResultWithTaskMessage":        WithTaskMessage,
		"upgradeResultWithTaskStartTime":      WithTaskStartTime,
		"upgradeResultWithTaskEndTime":        WithTaskEndTime,
		"upgradeResultWithTaskRetries":        WithTaskRetries,
	}
}
