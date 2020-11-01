/*
Copyright 2017 The OpenEBS Authors

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

package task

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	cstorpool "github.com/openebs/maya/pkg/cstor/pool/v1alpha2"
	cstorvolume "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	m_k8s "github.com/openebs/maya/pkg/k8s"
	deploy_appsv1 "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	deploy_extnv1beta1 "github.com/openebs/maya/pkg/kubernetes/deployment/extnv1beta1/v1alpha1"
	patch "github.com/openebs/maya/pkg/kubernetes/patch/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	podexec "github.com/openebs/maya/pkg/kubernetes/podexec/v1alpha1"
	replicaset "github.com/openebs/maya/pkg/kubernetes/replicaset/v1alpha1"
	service "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	snapshot "github.com/openebs/maya/pkg/kubernetes/snapshot/v1alpha1"
	snapshotdata "github.com/openebs/maya/pkg/kubernetes/snapshotdata/v1alpha1"
	storagepool "github.com/openebs/maya/pkg/storagepool/v1alpha1"
	"github.com/openebs/maya/pkg/template"
	templatefuncs "github.com/openebs/maya/pkg/templatefuncs/v1alpha1"
	upgraderesult "github.com/openebs/maya/pkg/upgrade/result/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	errors "github.com/pkg/errors"
	api_apps_v1 "k8s.io/api/apps/v1"
	api_apps_v1beta1 "k8s.io/api/apps/v1beta1"
	api_batch_v1 "k8s.io/api/batch/v1"
	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

var (
	// ErrorUnSupportedTask is used to throw error
	// for the tasks which are not supported by
	// the executor instance(s)
	ErrorUnSupportedTask error = errors.New("task not supported")
)

// Executor provides the contract to execute
// RunTasks
type Executor interface {
	Execute() (err error)
}

// OutputExecutor provides the contract to
// generate content from a RunTask's
// specifications
//
// NOTE:
//  The output format is specified in the
// RunTask itself
type OutputExecutor interface {
	Output() (output []byte, err error)
}

type executor struct {
	// Values is applied against the
	// task's specification (~ a go template)
	Values map[string]interface{}

	// MetaExec is used to execute meta
	// operations on this task
	MetaExec *MetaExecutor

	// Runtask defines a task & operations
	// associated with it
	Runtask *v1alpha1.RunTask
}

// newExecutor returns a new instance of
// executor
func newExecutor(rt *v1alpha1.RunTask, values map[string]interface{}) (*executor, error) {
	mte, err := NewMetaExecutor(rt.Spec.Meta, values)
	if err != nil {
		return nil,
			errors.Wrapf(err, "failed to init task executor: failed to init meta executor: %s %s", rt, stringer.Yaml("template values", values))
	}

	return &executor{
		Values:   values,
		MetaExec: mte,
		Runtask:  rt,
	}, nil
}

// String is the Stringer implementation
// of executor
func (m *executor) String() string {
	return stringer.Yaml("task executor", m)
}

// GoString is the GoStringer implementation
// of executor
func (m *executor) GoString() string {
	return stringer.Yaml("task executor", m)
}

// getTaskIdentity gets the task identity
func (m *executor) getTaskIdentity() string {
	return m.MetaExec.getIdentity()
}

// getTaskObjectName gets the task's object name
func (m *executor) getTaskObjectName() string {
	return m.MetaExec.getObjectName()
}

// getTaskRunNamespace gets the namespace where
// RunTask should get executed
func (m *executor) getTaskRunNamespace() string {
	return m.MetaExec.getRunNamespace()
}

// getK8sClient gets the kubernetes client to execute this task
func (m *executor) getK8sClient() *m_k8s_client.K8sClient {
	return m.MetaExec.getK8sClient()
}

// Output returns the result of templating a
// RunTask meant for templating only purpose
//
// NOTE:
//  This implements OutputExecutor interface
func (m *executor) Output() ([]byte, error) {
	output, err := template.AsTemplatedBytes(
		"output",
		m.Runtask.Spec.Task,
		m.Values,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate output: %s", m)
	}
	return output, nil
}

// getNotFoundError fetches NotFound error if any; post
// the execution of this runtask. This is extracted from
// the updated template values
//
// NOTE:
//  Logic to determine NotFound error is specified at
// Post property which is executed after the task's
// execution.
//
// NOTE:
//  In case of NotFound error, template values
// is set with NotFound error against below
// nested key
//
//  .TaskResult.<taskID>.notFoundErr
func (m *executor) getNotFoundError() interface{} {
	return util.GetNestedField(
		m.Values,
		string(v1alpha1.TaskResultTLP),
		m.getTaskIdentity(),
		string(v1alpha1.TaskResultNotFoundErrTRTP),
	)
}

// getVersionMismatchError fetches VersionMismatch error
// if any; post the execution of this runtask
//
// NOTE:
//  Logic to determine VersionMismatch error is specified at
// Post property which is executed after the task's execution.
//
// NOTE:
//  In case of VersionMismatch error, template values
// is set with VersionMismatch error against below nested
// key
//
//  .TaskResult.<taskID>.versionMismatchErr
func (m *executor) getTaskResultVersionMismatchError() interface{} {
	return util.GetNestedField(
		m.Values,
		string(v1alpha1.TaskResultTLP),
		m.getTaskIdentity(),
		string(v1alpha1.TaskResultVersionMismatchErrTRTP),
	)
}

// getVerifyError fetches the verification error if any;
// post the execution of this runtask
//
// NOTE:
//  Logic to determine Verify error is specified at Post
// property which is executed after the task's execution.
//
// NOTE:
//  In case of Verify error, template values is
// set with Verify error against below nested key
//
//  .TaskResult.<taskID>.verifyErr
func (m *executor) getVerifyError() interface{} {
	return util.GetNestedField(
		m.Values,
		string(v1alpha1.TaskResultTLP),
		m.getTaskIdentity(),
		string(v1alpha1.TaskResultVerifyErrTRTP),
	)
}

// resetVerifyError resets verification error if any;
// post the execution of this runtask
//
// NOTE:
//  If a runtask results in Verify error, its execution
// can be retried by reseting Verify error
//
// NOTE:
//  Reset here implies setting the verification error's
// placeholder value to nil
//
//  Below property is reset with 'nil':
//  .TaskResult.<taskID>.verifyErr
func (m *executor) resetTaskResultVerifyError() {
	util.SetNestedField(
		m.Values,
		nil,
		string(v1alpha1.TaskResultTLP),
		m.getTaskIdentity(),
		string(v1alpha1.TaskResultVerifyErrTRTP),
	)
}

// repeatWith repeats execution of the task based on
// repeatWith property of meta task specifications.
//
// NOTE:
//  With this property RunTask can be executed repeatedly
// based on the resource names set against the repeatWith
// property.
//
// NOTE:
//  Each task execution depends on the current resource
// index
func (m *executor) repeatWith() (err error) {
	rptExec := m.MetaExec.getRepeatExecutor()
	if !rptExec.isRepeat() {
		// no need to repeat if this task
		// is not meant to be repeated;
		// so execute once & return
		err = m.retryOnVerificationError()
		return
	}

	// execute the task based on each repeat
	repeats := rptExec.len()
	var (
		rptMetaExec *MetaExecutor
		current     string
	)

	for idx := 0; idx < repeats; idx++ {
		// fetch a new repeat meta task instance
		rptMetaExec, err = m.MetaExec.asRepeatInstance(idx)
		if err != nil {
			// stop repetition on unhandled runtime errors
			// & return
			return
		}
		// mutate the original meta task executor
		// to this repeater instance
		m.MetaExec = rptMetaExec

		// set the currently active repeat item
		current, err = m.MetaExec.repeater.getItem(idx)
		if err != nil {
			// stop repetition on unhandled runtime error
			// & return
			return
		}

		util.SetNestedField(
			m.Values,
			current,
			string(v1alpha1.ListItemsTLP),
			string(v1alpha1.CurrentRepeatResourceLITP),
		)

		// execute the task function finally
		err = m.retryOnVerificationError()
		if err != nil {
			// stop repetition on unhandled runtime error
			// & return
			return
		}
	}

	return
}

// retryOnVerificationError retries execution of the task
// if the task execution resulted into verification error.
// The number of retry attempts & interval between each
// attempt is specified in the task's meta specification.
func (m *executor) retryOnVerificationError() (err error) {
	retryAttempts, interval := m.MetaExec.getRetry()

	// original invocation as well as all retry attempts
	// i == 0 implies original task execute invocation
	// i > 0 implies a retry operation
	for i := 0; i <= retryAttempts; i++ {
		// first reset the previous verify error if any
		m.resetTaskResultVerifyError()

		// execute the task function
		err = m.ExecuteIt()
		if err != nil {
			// break this retry execution loop
			// if there were any runtime errors
			return
		}

		// check for VerifyError if any
		//
		// NOTE:
		//  VerifyError is a handled runtime error
		// which is set via templating
		//
		// NOTE:
		//  retry is done only if VerifyError is
		// set during post task execution
		verifyErr := m.getVerifyError()
		if verifyErr == nil {
			// no need to retry if task execution was a
			// success i.e. there was no verification error
			// found with the task result
			return
		}

		// current verify error
		err, _ = verifyErr.(*templatefuncs.VerifyError)

		if i != retryAttempts {
			klog.Warningf(
				"verify error was found for runtask {%s}: error {%s}: will retry task execution-'%d'",
				m.getTaskIdentity(),
				err,
				i+1,
			)

			// will retry after the specified interval
			time.Sleep(interval)
		}
	}

	// return after exhausting the original invocation
	// and all retries; verification error of the final
	// attempt will be returned here
	return
}

// Execute executes a runtask by following the
// directives specified in the runtask's meta
// specifications
func (m *executor) Execute() (err error) {
	if m.MetaExec.isDisabled() {
		// do nothing if runtask is disabled
		return
	}
	return m.repeatWith()
}

// postExecuteIt executes a go template against
// the provided template values. This is run
// after executing a task.
//
// NOTE:
//  This go template is a set of template functions
// that queries specified properties from the result
// of task's execution & stores them at placeholders
// within the **template values**. These stored values
// can later be queried by subsequent runtasks.
func (m *executor) postExecuteIt() (err error) {
	if m.Runtask == nil || len(m.Runtask.Spec.PostRun) == 0 {
		// do nothing if post specs is empty
		return
	}

	// post runtask operation
	_, err = template.AsTemplatedBytes(
		"PostRun",
		m.Runtask.Spec.PostRun,
		m.Values,
	)
	if err != nil {
		// return any un-handled runtime error
		return
	}

	// verMismatchErr is a handled runtime error i.e.
	// is set in template values. This needs to be
	// extracted and thrown as VersionMismatchError
	verMismatchErr := m.getTaskResultVersionMismatchError()
	if verMismatchErr != nil {
		klog.Warningf(
			"version mismatch error at runtask {%s}: error {%s}",
			m.getTaskIdentity(),
			verMismatchErr,
		)
		err, _ = verMismatchErr.(*templatefuncs.VersionMismatchError)
		return
	}

	// notFoundErr is a handled runtime error i.e. is
	// set is in template values. This needs to be
	// extracted and thrown as NotFoundError
	notFoundErr := m.getNotFoundError()
	if notFoundErr != nil {
		klog.Warningf(
			"notfound error at runtask {%s}: error {%s}",
			m.getTaskIdentity(),
			notFoundErr,
		)
		err, _ = notFoundErr.(*templatefuncs.NotFoundError)
		return
	}

	return nil
}

// ExecuteIt will execute the runtask based on
// its meta & task specifications
func (m *executor) ExecuteIt() (err error) {
	if m.getK8sClient() == nil {
		return errors.Errorf("failed to execute task: nil k8s client: verify if namespace is set: %s", m)
	}

	// kind as command is a special case of task execution
	if m.MetaExec.isCommand() {
		return m.postExecuteIt()
	}

	if m.MetaExec.isRolloutstatus() {
		err = m.rolloutStatus()
	} else if m.MetaExec.isPutExtnV1B1Deploy() {
		err = m.putExtnV1B1Deploy()
	} else if m.MetaExec.isPutAppsV1B1Deploy() {
		err = m.putAppsV1B1Deploy()
	} else if m.MetaExec.isPutAppsV1Deploy() {
		err = m.putAppsV1Deploy()
	} else if m.MetaExec.isPatchExtnV1B1Deploy() {
		err = m.patchExtnV1B1Deploy()
	} else if m.MetaExec.isPatchAppsV1Deploy() {
		err = m.patchAppsV1Deploy()
	} else if m.MetaExec.isPatchAppsV1B1Deploy() {
		err = m.patchAppsV1B1Deploy()
	} else if m.MetaExec.isPatchOEV1alpha1SPC() {
		err = m.patchOEV1alpha1SPC()
	} else if m.MetaExec.isPatchOEV1alpha1CSPC() {
		err = m.patchOEV1alpha1CSPC()
	} else if m.MetaExec.isPutCoreV1Service() {
		err = m.putCoreV1Service()
	} else if m.MetaExec.isPatchV1alpha1VolumeSnapshotData() {
		err = m.patchV1alpha1VolumeSnapshotData()
	} else if m.MetaExec.isPatchCoreV1Service() {
		err = m.patchCoreV1Service()
	} else if m.MetaExec.isDeleteExtnV1B1Deploy() {
		err = m.deleteExtnV1B1Deployment()
	} else if m.MetaExec.isDeleteExtnV1B1ReplicaSet() {
		err = m.deleteExtnV1B1ReplicaSet()
	} else if m.MetaExec.isGetExtnV1B1Deploy() {
		err = m.getExtnV1B1Deployment()
	} else if m.MetaExec.isGetExtnV1B1ReplicaSet() {
		err = m.getExtnV1B1ReplicaSet()
	} else if m.MetaExec.isGetCoreV1Pod() {
		err = m.getCoreV1Pod()
	} else if m.MetaExec.isDeleteAppsV1B1Deploy() {
		err = m.deleteAppsV1B1Deployment()
	} else if m.MetaExec.isDeleteAppsV1Deploy() {
		err = m.deleteAppsV1Deployment()
	} else if m.MetaExec.isDeleteCoreV1Service() {
		err = m.deleteCoreV1Service()
	} else if m.MetaExec.isGetOEV1alpha1BlockDevice() {
		err = m.getOEV1alpha1BlockDevice()
	} else if m.MetaExec.isGetV1alpha1VolumeSnapshotData() {
		err = m.getV1alpha1VolumeSnapshotData()
	} else if m.MetaExec.isGetOEV1alpha1SPC() {
		err = m.getOEV1alpha1SPC()
	} else if m.MetaExec.isGetOEV1alpha1CSPC() {
		err = m.getOEV1alpha1CSPC()
	} else if m.MetaExec.isGetOEV1alpha1SP() {
		err = m.getOEV1alpha1SP()
	} else if m.MetaExec.isGetOEV1alpha1CSP() {
		err = m.getOEV1alpha1CSP()
	} else if m.MetaExec.isGetOEV1alpha1UR() {
		err = m.getOEV1alpha1UR()
	} else if m.MetaExec.isGetCoreV1PVC() {
		err = m.getCoreV1PVC()
	} else if m.MetaExec.isGetCoreV1Service() {
		err = m.getCoreV1Service()
	} else if m.MetaExec.isGetOEV1alpha1CSV() {
		err = m.getOEV1alpha1CSV()
	} else if m.MetaExec.isPutOEV1alpha1CSP() {
		err = m.putCStorPool()
	} else if m.MetaExec.isPutOEV1alpha1SP() {
		err = m.putStoragePool()
	} else if m.MetaExec.isPutOEV1alpha1CSV() {
		err = m.putCStorVolume()
	} else if m.MetaExec.isPutOEV1alpha1CVR() {
		err = m.putCStorVolumeReplica()
	} else if m.MetaExec.isPutOEV1alpha1UR() {
		err = m.putUpgradeResult()
	} else if m.MetaExec.isDeleteOEV1alpha1SP() {
		err = m.deleteOEV1alpha1SP()
	} else if m.MetaExec.isDeleteOEV1alpha1CSP() {
		err = m.deleteOEV1alpha1CSP()
	} else if m.MetaExec.isDeleteOEV1alpha1CSV() {
		err = m.deleteOEV1alpha1CSV()
	} else if m.MetaExec.isDeleteOEV1alpha1CVR() {
		err = m.deleteOEV1alpha1CVR()
	} else if m.MetaExec.isPatchOEV1alpha1CSV() {
		err = m.patchOEV1alpha1CSV()
	} else if m.MetaExec.isPatchOEV1alpha1CVR() {
		err = m.patchOEV1alpha1CVR()
	} else if m.MetaExec.isPatchOEV1alpha1UR() {
		err = m.patchUpgradeResult()
	} else if m.MetaExec.isPatchOEV1alpha1SP() {
		err = m.patchStoragePool()
	} else if m.MetaExec.isPatchOEV1alpha1CSP() {
		err = m.patchCstorPool()
	} else if m.MetaExec.isList() {
		err = m.listK8sResources()
	} else if m.MetaExec.isGetStorageV1SC() {
		err = m.getStorageV1SC()
	} else if m.MetaExec.isGetCoreV1PV() {
		err = m.getCoreV1PV()
	} else if m.MetaExec.isDeleteBatchV1Job() {
		err = m.deleteBatchV1Job()
	} else if m.MetaExec.isGetBatchV1Job() {
		err = m.getBatchV1Job()
	} else if m.MetaExec.isPutBatchV1Job() {
		err = m.putBatchV1Job()
	} else if m.MetaExec.isPutAppsV1STS() {
		err = m.putAppsV1STS()
	} else if m.MetaExec.isDeleteAppsV1STS() {
		err = m.deleteAppsV1STS()
	} else if m.MetaExec.isExecCoreV1Pod() {
		err = m.execCoreV1Pod()
	} else if m.MetaExec.isGetAppsV1Deploy() {
		err = m.getAppsV1Deployment()
	} else {
		err = ErrorUnSupportedTask
	}

	if err != nil {
		return errors.Wrapf(err, "failed to execute task: %s", m)
	}

	// run the post operations after a runtask is executed
	return m.postExecuteIt()
}

// asRollbackInstance will provide the rollback
// instance associated to this task's instance
func (m *executor) asRollbackInstance(objectName string) (*executor, error) {
	mte, willRollback, err := m.MetaExec.asRollbackInstance(objectName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build rollback executor for object {%s}: %v", objectName, m.MetaExec)
	}

	if !willRollback {
		// no need of rollback
		return nil, nil
	}

	// Only the meta info is required for a rollback. In
	// other words no need of task yaml template & values
	return &executor{
		MetaExec: mte,
	}, nil
}

// asBatchV1Job generates a K8s Job object
// out of the embedded yaml
func (m *executor) asBatchV1Job() (*api_batch_v1.Job, error) {
	j, err := m_k8s.NewJobYml("BatchV1Job", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build job")
	}
	return j.AsBatchV1Job()
}

// asAppsV1STS generates a kubernetes StatefulSet api
// instance from the yaml string specification
func (m *executor) asAppsV1STS() (*api_apps_v1.StatefulSet, error) {
	s, err := m_k8s.NewSTSYml("AppsV1StatefulSet", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build statefulset")
	}
	return s.AsAppsV1STS()
}

// asAppsV1B1Deploy generates a K8s Deployment object
// out of the embedded yaml
func (m *executor) asAppsV1B1Deploy() (*api_apps_v1beta1.Deployment, error) {
	d, err := m_k8s.NewDeploymentYml("AppsV1B1Deploy", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build deployment")
	}

	return d.AsAppsV1B1Deployment()
}

func (m *executor) asAppsV1Deploy() (*api_apps_v1.Deployment, error) {
	d, err := m_k8s.NewDeploymentYml("AppsV1Deploy", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build deployment")
	}

	return d.AsAppsV1Deployment()
}

// asExtnV1B1Deploy generates a K8s Deployment object
// out of the embedded yaml
func (m *executor) asExtnV1B1Deploy() (*api_extn_v1beta1.Deployment, error) {
	d, err := m_k8s.NewDeploymentYml("ExtnV1B11Deploy", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build deployment")
	}

	return d.AsExtnV1B1Deployment()
}

// asCStorPool generates a CstorPool object
// out of the embedded yaml
func (m *executor) asCStorPool() (*v1alpha1.CStorPool, error) {
	d, err := m_k8s.NewCStorPoolYml("CStorPool", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build cstorpool")
	}

	return d.AsCStorPoolYml()
}

// asStoragePool generates a StoragePool object
// out of the embedded yaml
func (m *executor) asStoragePool() (*v1alpha1.StoragePool, error) {
	d, err := m_k8s.NewStoragePoolYml("StoragePool", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build storagepool")
	}

	return d.AsStoragePoolYml()
}

// asCStorVolume generates a CstorVolume object
// out of the embedded yaml
func (m *executor) asCStorVolume() (*v1alpha1.CStorVolume, error) {
	d, err := m_k8s.NewCStorVolumeYml("CstorVolume", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build cstorvolume")
	}

	return d.AsCStorVolumeYml()
}

// asCstorVolumeReplica generates a CStorVolumeReplica object
// out of the embedded yaml
func (m *executor) asCstorVolumeReplica() (*v1alpha1.CStorVolumeReplica, error) {
	d, err := m_k8s.NewCStorVolumeReplicaYml("CstorVolumeReplica", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build cstorvolumereplica")
	}

	return d.AsCStorVolumeReplicaYml()
}

// asCoreV1Svc generates a K8s Service object
// out of the embedded yaml
func (m *executor) asCoreV1Svc() (*api_core_v1.Service, error) {
	s, err := m_k8s.NewServiceYml("CoreV1Svc", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build service")
	}

	return s.AsCoreV1Service()
}

// putBatchV1Job will create a Job object
func (m *executor) putBatchV1Job() error {
	j, err := m.asBatchV1Job()
	if err != nil {

		return errors.Wrap(err, "failed to create job")
	}

	job, err := m.getK8sClient().CreateBatchV1JobAsRaw(j)
	if err != nil {
		return errors.Wrap(err, "failed to create job")
	}

	util.SetNestedField(m.Values, job, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// putAppsV1STS will create a new StatefulSet
// object in the cluster and store the response
// in a json format
func (m *executor) putAppsV1STS() error {
	j, err := m.asAppsV1STS()
	if err != nil {
		return errors.Wrap(err, "failed to create statefulset")
	}

	sts, err := m.getK8sClient().CreateAppsV1STSAsRaw(j)
	if err != nil {
		return errors.Wrap(err, "failed to create statefulset")
	}

	util.SetNestedField(m.Values, sts, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// putAppsV1B1Deploy will create (i.e. apply to a kubernetes cluster) a Deployment
// object. The Deployment specs is configured in the RunTask.
func (m *executor) putAppsV1B1Deploy() error {
	d, err := m.asAppsV1B1Deploy()
	if err != nil {
		return errors.Wrap(err, "failed to create deployment")
	}

	deploy, err := m.getK8sClient().CreateAppsV1B1DeploymentAsRaw(d)
	if err != nil {
		return errors.Wrap(err, "failed to create deployment")
	}

	util.SetNestedField(m.Values, deploy, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// putAppsV1Deploy will create (i.e. apply to a kubernetes cluster) a Deployment
// object. The Deployment specs is configured in the RunTask.
func (m *executor) putAppsV1Deploy() error {
	d, err := m.asAppsV1Deploy()
	if err != nil {
		return errors.Wrap(err, "failed to create deployment")
	}

	deploy, err := m.getK8sClient().CreateAppsV1DeploymentAsRaw(d)
	if err != nil {
		return errors.Wrap(err, "failed to create deployment")
	}

	util.SetNestedField(m.Values, deploy, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// putExtnV1B1Deploy will create (i.e. apply to kubernetes cluster) a Deployment
// whose specifications are defined in the RunTask
func (m *executor) putExtnV1B1Deploy() error {
	d, err := m.asExtnV1B1Deploy()
	if err != nil {
		return errors.Wrap(err, "failed to create deployment")
	}

	deploy, err := m.getK8sClient().CreateExtnV1B1DeploymentAsRaw(d)
	if err != nil {
		return errors.Wrap(err, "failed to create deployment")
	}

	util.SetNestedField(m.Values, deploy, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// patchSPC will patch a SPC object in a kubernetes cluster.
// The patch specifications as configured in the RunTask
func (m *executor) patchOEV1alpha1SPC() error {
	patch, err := asTaskPatch("patchSPC", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return errors.Wrap(err, "failed to patch storagepoolclaim")
	}

	pe, err := newTaskPatchExecutor(patch)
	if err != nil {
		return errors.Wrap(err, "failed to patch storagepoolclaim")
	}

	raw, err := pe.toJson()
	if err != nil {
		return errors.Wrap(err, "failed to patch storagepoolclaim")
	}

	// patch storagepoolclaim
	spc, err := m.getK8sClient().PatchOEV1alpha1SPCAsRaw(m.getTaskObjectName(), pe.patchType(), raw)
	if err != nil {
		return errors.Wrap(err, "failed to patch storagepoolclaim")
	}

	util.SetNestedField(m.Values, spc, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// patchOEV1alpha1CSPC will patch a CSPC object in a kubernetes cluster.
// The patch specifications as configured in the RunTask
func (m *executor) patchOEV1alpha1CSPC() (err error) {
	patch, err := asTaskPatch("patchSPC", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return errors.Wrap(err, "failed to patch cspc object")
	}

	pe, err := newTaskPatchExecutor(patch)
	if err != nil {
		return errors.Wrap(err, "failed to patch cspc object")
	}

	raw, err := pe.toJson()
	if err != nil {
		return errors.Wrap(err, "failed to patch cspc object")
	}

	// patch the CSPC
	cspc, err := m.getK8sClient().PatchOEV1alpha1CSPCAsRaw(m.getTaskObjectName(), pe.patchType(), raw)
	if err != nil {
		return errors.Wrap(err, "failed to patch cspc object")
	}

	util.SetNestedField(m.Values, cspc, string(v1alpha1.CurrentJSONResultTLP))
	return
}

func (m *executor) patchV1alpha1VolumeSnapshotData() (err error) {
	patch, err := asTaskPatch("patchVolumeSnapshotData", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return errors.Wrap(err, "failed to patch VolumeSnapshotData object")
	}

	pe, err := newTaskPatchExecutor(patch)
	if err != nil {
		return errors.Wrap(err, "failed to patch VolumeSnapshotData object")
	}

	raw, err := pe.toJson()
	if err != nil {
		return errors.Wrap(err, "failed to patch VolumeSnapshotData object")
	}

	// patch the VolumeSnapshotData
	vsd, err := snapshotdata.NewKubeClient().Patch(
		m.getTaskObjectName(),
		pe.patchType(),
		raw,
	)
	if err != nil {
		return errors.Wrap(err, "failed to patch VolumeSnapshotData")
	}

	util.SetNestedField(m.Values, vsd, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// patchOEV1alpha1CSV will patch a CStorVolume as defined in the task
func (m *executor) patchOEV1alpha1CSV() error {
	patch, err := asTaskPatch("patchCSV", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return errors.Wrap(err, "failed to patch cstorvolume")
	}

	pe, err := newTaskPatchExecutor(patch)
	if err != nil {
		return errors.Wrap(err, "failed to patch cstorvolume")
	}

	raw, err := pe.toJson()
	if err != nil {
		return errors.Wrap(err, "failed to patch cstorvolume")
	}

	// patch the cstorvolume
	csv, err := m.getK8sClient().PatchOEV1alpha1CSV(
		m.getTaskObjectName(),
		m.getTaskRunNamespace(),
		pe.patchType(),
		raw,
	)
	if err != nil {
		return errors.Wrap(err, "failed to patch cstorvolume")
	}

	util.SetNestedField(m.Values, csv, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// patchOEV1alpha1CVR will patch a CStorVolumeReplica as defined in the task
func (m *executor) patchOEV1alpha1CVR() error {
	patch, err := asTaskPatch("patchCVR", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return errors.Wrap(err, "failed to patch cstorvolumereplica")
	}

	pe, err := newTaskPatchExecutor(patch)
	if err != nil {
		return errors.Wrap(err, "failed to patch cstorvolumereplica")
	}

	raw, err := pe.toJson()
	if err != nil {
		return errors.Wrap(err, "failed to patch cstorvolumereplica")
	}

	// patch cstorvolumereplica
	cvr, err := m.getK8sClient().PatchOEV1alpha1CVR(
		m.getTaskObjectName(),
		m.getTaskRunNamespace(),
		pe.patchType(),
		raw,
	)
	if err != nil {
		return errors.Wrap(err, "failed to patch cstorvolumereplica")
	}

	util.SetNestedField(m.Values, cvr, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// patchUpgradeResult will patch an UpgradeResult
// as defined in the task
func (m *executor) patchUpgradeResult() error {
	// build a runtask patch instance
	patch, err := patch.
		BuilderForRuntask("UpgradeResult", m.Runtask.Spec.Task, m.Values).
		AddCheckf(patch.IsValidType(), "IsValidType").
		Build()
	if err != nil {
		return errors.Wrap(err, "failed to patch upgraderesult")
	}

	// patch Upgrade Result
	p, err := upgraderesult.
		NewKubeClient().
		WithNamespace(m.getTaskRunNamespace()).
		Patch(m.getTaskObjectName(), patch.Type, patch.Object)
	if err != nil {
		return errors.Wrap(err, "failed to patch upgraderesult")
	}

	util.SetNestedField(m.Values, p, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// patchStoragePool will patch an StoragePool as defined in the task
func (m *executor) patchStoragePool() (err error) {
	// build a runtask patch instance
	patch, err := patch.
		BuilderForRuntask("StoragePool", m.Runtask.Spec.Task, m.Values).
		AddCheckf(patch.IsValidType(), "patch type is not valid").
		Build()
	if err != nil {
		return errors.Wrap(err, "failed to patch storage pool")
	}

	p, err := storagepool.
		NewKubeClient().
		Patch(m.getTaskObjectName(), patch.Type, patch.Object)
	if err != nil {
		return errors.Wrap(err, "failed to patch storage pool")
	}
	util.SetNestedField(m.Values, p, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// patchCstorPool will patch an CstorPool as defined in the task
func (m *executor) patchCstorPool() (err error) {
	patch, err := patch.
		BuilderForRuntask("CstorPool", m.Runtask.Spec.Task, m.Values).
		AddCheckf(patch.IsValidType(), "patch type is not valid").
		Build()
	if err != nil {
		return errors.Wrap(err, "failed to patch cstorpool")
	}

	p, err := cstorpool.
		NewKubeClient().
		Patch(m.getTaskObjectName(), patch.Type, patch.Object)
	if err != nil {
		return errors.Wrap(err, "failed to patch cstorpool")
	}
	util.SetNestedField(m.Values, p, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// patchAppsV1B1Deploy will patch a Deployment object where patch specifications
// are configured in the RunTask
func (m *executor) patchAppsV1Deploy() error {
	patch, err := asTaskPatch("AppsV1DeployPatch", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return errors.Wrap(err, "failed to patch deployment")
	}

	pe, err := newTaskPatchExecutor(patch)
	if err != nil {
		return errors.Wrap(err, "failed to patch deployment")
	}

	raw, err := pe.toJson()
	if err != nil {
		return errors.Wrap(err, "failed to patch deployment")
	}

	// patch the deployment
	deploy, err := m.getK8sClient().PatchAppsV1DeploymentAsRaw(
		m.getTaskObjectName(),
		pe.patchType(),
		raw,
	)
	if err != nil {
		return errors.Wrap(err, "failed to patch deployment")
	}

	util.SetNestedField(m.Values, deploy, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// patchAppsV1B1Deploy will patch a Deployment object in a kubernetes cluster.
// The patch specifications as configured in the RunTask
func (m *executor) patchAppsV1B1Deploy() (err error) {
	err = errors.Errorf("patchAppsV1B1Deploy is not implemented")
	return
}

// patchExtnV1B1Deploy will patch a Deployment
// object where patch specifications are
// configured in the RunTask
func (m *executor) patchExtnV1B1Deploy() error {
	patch, err := asTaskPatch("ExtnV1B1DeployPatch", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return errors.Wrap(err, "failed to patch deployment")
	}

	pe, err := newTaskPatchExecutor(patch)
	if err != nil {
		return errors.Wrap(err, "failed to patch deployment")
	}

	raw, err := pe.toJson()
	if err != nil {
		return errors.Wrap(err, "failed to patch deployment")
	}

	// patch the deployment
	deploy, err := m.getK8sClient().PatchExtnV1B1DeploymentAsRaw(
		m.getTaskObjectName(),
		pe.patchType(),
		raw,
	)
	if err != nil {
		return errors.Wrap(err, "failed to patch deployment")
	}

	util.SetNestedField(m.Values, deploy, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// patchCoreV1Service will patch a Service
// where patch specifications are configured
// in the RunTask
func (m *executor) patchCoreV1Service() error {
	patch, err := asTaskPatch("CoreV1ServicePatch", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return errors.Wrap(err, "failed to patch service")
	}

	pe, err := newTaskPatchExecutor(patch)
	if err != nil {
		return errors.Wrap(err, "failed to patch service")
	}

	raw, err := pe.toJson()
	if err != nil {
		return errors.Wrap(err, "failed to patch service")
	}

	// patch service
	service, err := m.getK8sClient().PatchCoreV1ServiceAsRaw(
		m.getTaskObjectName(),
		pe.patchType(),
		raw,
	)
	if err != nil {
		return errors.Wrap(err, "failed to patch service")
	}

	util.SetNestedField(m.Values, service, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// deleteAppsV1B1Deployment will delete one or
// more Deployments as specified in the RunTask
func (m *executor) deleteAppsV1B1Deployment() error {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err := m.getK8sClient().DeleteAppsV1B1Deployment(strings.TrimSpace(name))
		if err != nil {
			return errors.Wrapf(err, "failed to delete deployment {%s}", name)
		}
	}

	return nil
}

// deleteAppsV1Deployment will delete one or
// more Deployments as specified in the RunTask
func (m *executor) deleteAppsV1Deployment() error {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err := m.getK8sClient().DeleteAppsV1Deployment(strings.TrimSpace(name))
		if err != nil {
			return errors.Wrapf(err, "failed to delete deployment {%s}", name)
		}
	}

	return nil
}

// deleteOEV1alpha1CVR will delete one or more
// CStorVolumeReplica as specified in
// the RunTask
func (m *executor) deleteOEV1alpha1CVR() error {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err := m.getK8sClient().DeleteOEV1alpha1CVR(name)
		if err != nil {
			return errors.Wrapf(err, "failed to delete cstorvolumereplica {%s}", name)
		}
	}

	return nil
}

// deleteExtnV1B1Deployment will delete one or
// more Deployments as specified in the RunTask
func (m *executor) deleteExtnV1B1Deployment() error {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err := m.getK8sClient().DeleteExtnV1B1Deployment(strings.TrimSpace(name))
		if err != nil {
			return errors.Wrapf(err, "failed to delete deployment {%s}", name)
		}
	}

	return nil
}

// getExtnV1B1ReplicaSet will get the Replicaset
// as specified in the RunTask
func (m *executor) getExtnV1B1ReplicaSet() error {
	rs, err := replicaset.
		KubeClient(replicaset.WithNamespace(m.getTaskRunNamespace())).
		GetRaw(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get replicaset")
	}

	util.SetNestedField(m.Values, rs, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// deleteExtnV1B1ReplicaSet will delete one or
// more ReplicaSets as specified in the RunTask
func (m *executor) deleteExtnV1B1ReplicaSet() error {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")
	client := replicaset.KubeClient(
		replicaset.WithNamespace(m.getTaskRunNamespace()))

	for _, name := range objectNames {
		err := client.Delete(strings.TrimSpace(name))
		if err != nil {
			return errors.Wrapf(err, "failed to delete replicaset {%s}", name)
		}
	}

	return nil
}

// listExtnV1B1ReplicaSet lists the replica sets
// based on the provided list options
func (m *executor) listExtnV1B1ReplicaSet(opt metav1.ListOptions) ([]byte, error) {
	return replicaset.
		KubeClient(replicaset.WithNamespace(m.getTaskRunNamespace())).
		ListRaw(opt)
}

func (m *executor) listV1alpha1VolumeSnapshotData(opt metav1.ListOptions) ([]byte, error) {
	return snapshotdata.
		NewKubeClient().
		ListRaw(opt)
}

func (m *executor) listV1alpha1VolumeSnapshot(opt metav1.ListOptions) ([]byte, error) {
	return snapshot.
		NewKubeClient().
		WithNamespace(m.getTaskRunNamespace()).
		ListRaw(opt)
}

// putCoreV1Service will create a Service whose
// specs are configured in the RunTask
func (m *executor) putCoreV1Service() error {
	s, err := m.asCoreV1Svc()
	if err != nil {
		return errors.Wrap(err, "failed to create service")
	}

	svc, err := m.getK8sClient().CreateCoreV1ServiceAsRaw(s)
	if err != nil {
		return errors.Wrap(err, "failed to create service")
	}

	util.SetNestedField(m.Values, svc, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// deleteCoreV1Service will delete one or more
// services as specified in the RunTask
func (m *executor) deleteCoreV1Service() error {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err := m.getK8sClient().DeleteCoreV1Service(strings.TrimSpace(name))
		if err != nil {
			return errors.Wrapf(err, "failed to delete service {%s}", name)
		}
	}

	return nil
}

// getOEV1alpha1Disk() will get the Disk
// as specified in the RunTask
func (m *executor) getOEV1alpha1BlockDevice() error {
	disk, err := m.getK8sClient().GetOEV1alpha1BlockDeviceAsRaw(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get disk")
	}

	util.SetNestedField(m.Values, disk, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

func (m *executor) getV1alpha1VolumeSnapshotData() error {
	//	snapshotData, err := m.getK8sClient().
	snapshotData, err := snapshotdata.NewKubeClient().GetRaw(m.getTaskObjectName(), metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get volume snapshot data")
	}

	util.SetNestedField(m.Values, snapshotData, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// getOEV1alpha1SPC() will get the StoragePoolClaim
// as specified in the RunTask
func (m *executor) getOEV1alpha1SPC() error {
	spc, err := m.getK8sClient().GetOEV1alpha1SPCAsRaw(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get storagepoolclaim")
	}

	util.SetNestedField(m.Values, spc, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// getOEV1alpha1CSPC() will get the CStorPoolCluster as specified in the RunTask
func (m *executor) getOEV1alpha1CSPC() (err error) {
	cspc, err := m.getK8sClient().GetOEV1alpha1CSPCAsRaw(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get cstor pool cluster")
	}

	util.SetNestedField(m.Values, cspc, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getOEV1alpha1SP will get the StoragePool as specified in the RunTask
func (m *executor) getOEV1alpha1SP() (err error) {
	sp, err := m.getK8sClient().GetOEV1alpha1SPAsRaw(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get storagepool")
	}

	util.SetNestedField(m.Values, sp, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

func (m *executor) getOEV1alpha1CSP() error {
	csp, err := m.getK8sClient().GetOEV1alpha1CSPAsRaw(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get cstorstoragepool")
	}

	util.SetNestedField(m.Values, csp, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// getOEV1alpha1UR will get the UpgradeResult
// as specified in the RunTask
func (m *executor) getOEV1alpha1UR() error {
	uresult, err := upgraderesult.
		NewKubeClient().
		WithNamespace(m.getTaskRunNamespace()).
		Get(m.getTaskObjectName(), metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get upgraderesult")
	}
	ur, err := json.Marshal(uresult)
	if err != nil {
		return errors.Wrap(err, "failed to get upgraderesult")
	}

	util.SetNestedField(m.Values, ur, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// getOEV1alpha1CSV will get the CstorVolume as specified in the RunTask
func (m *executor) getOEV1alpha1CSV() error {
	csv, err := cstorvolume.NewKubeclient(
		cstorvolume.WithNamespace(m.getTaskRunNamespace())).
		GetRaw(m.getTaskObjectName(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	util.SetNestedField(m.Values, csv, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// getCoreV1Service will get the Service as specified in the RunTask
func (m *executor) getCoreV1Service() error {
	svc, err := service.NewKubeClient(
		service.WithNamespace(m.getTaskRunNamespace())).
		GetRaw(m.getTaskObjectName(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	util.SetNestedField(m.Values, svc, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// getExtnV1B1Deployment will get the Deployment as specified in the RunTask
func (m *executor) getExtnV1B1Deployment() (err error) {
	dclient := deploy_extnv1beta1.KubeClient(
		deploy_extnv1beta1.WithNamespace(m.getTaskRunNamespace()),
		deploy_extnv1beta1.WithClientset(m.getK8sClient().GetKCS()))
	d, err := dclient.GetRaw(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get deployment")
	}

	util.SetNestedField(m.Values, d, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// extnV1B1DeploymentRollOutStatus generates rollout status for a given deployment from deployment object
func (m *executor) extnV1B1DeploymentRollOutStatus() (err error) {
	dclient := deploy_extnv1beta1.KubeClient(
		deploy_extnv1beta1.WithNamespace(m.getTaskRunNamespace()),
		deploy_extnv1beta1.WithClientset(m.getK8sClient().GetKCS()))
	res, err := dclient.RolloutStatusf(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get deployment rollout status")
	}

	util.SetNestedField(m.Values, res, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// appsV1DeploymentRollOutStatus generates rollout status for a given deployment from deployment object
func (m *executor) appsV1DeploymentRollOutStatus() (err error) {
	dclient := deploy_appsv1.NewKubeClient(
		deploy_appsv1.WithNamespace(m.getTaskRunNamespace()),
		deploy_appsv1.WithClientset(m.getK8sClient().GetKCS()))
	res, err := dclient.RolloutStatusf(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get deployment rollout status")
	}

	util.SetNestedField(m.Values, res, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getAppsV1Deployment will get the Deployment as specified in the RunTask
func (m *executor) getAppsV1Deployment() (err error) {
	dclient := deploy_appsv1.NewKubeClient(
		deploy_appsv1.WithNamespace(m.getTaskRunNamespace()),
		deploy_appsv1.WithClientset(m.getK8sClient().GetKCS()))
	d, err := dclient.GetRaw(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get deployment")
	}
	util.SetNestedField(m.Values, d, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getCoreV1PVC will get the PVC as specified in the RunTask
func (m *executor) getCoreV1PVC() (err error) {
	pvc, err := m.getK8sClient().GetCoreV1PVCAsRaw(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get pvc")
	}

	util.SetNestedField(m.Values, pvc, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getCoreV1PV will get the PersistentVolume as specified in the RunTask
func (m *executor) getCoreV1PV() (err error) {
	pv, err := m.getK8sClient().GetCoreV1PersistentVolumeAsRaw(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get pv")
	}

	util.SetNestedField(m.Values, pv, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getBatchV1Job will get the Job as specified in the RunTask
func (m *executor) getBatchV1Job() (err error) {
	job, err := m.getK8sClient().GetBatchV1JobAsRaw(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get job")
	}
	util.SetNestedField(m.Values, job, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getCoreV1Pod will get the Pod as specified in the RunTask
func (m *executor) getCoreV1Pod() (err error) {
	podClient := pod.NewKubeClient().WithNamespace(m.getTaskRunNamespace())

	pod, err := podClient.GetRaw(m.getTaskObjectName(), metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get pod")
	}

	util.SetNestedField(m.Values, pod, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// deleteBatchV1Job will delete one or more Jobs specified in the RunTask
func (m *executor) deleteBatchV1Job() (err error) {
	jobs := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")
	for _, name := range jobs {
		err = m.getK8sClient().DeleteBatchV1Job(strings.TrimSpace(name))
		if err != nil {
			return errors.Wrap(err, "failed to delete job")
		}
	}
	return
}

// deleteAppsV1STS will delete one or more StatefulSets
func (m *executor) deleteAppsV1STS() (err error) {
	stss := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")
	for _, name := range stss {
		err = m.getK8sClient().DeleteAppsV1STS(strings.TrimSpace(name))
		if err != nil {
			return errors.Wrap(err, "failed to delete statefulset")
		}
	}
	return
}

// getStorageV1SC will get the StorageClass as specified in the RunTask
func (m *executor) getStorageV1SC() (err error) {
	sc, err := m.getK8sClient().GetStorageV1SCAsRaw(m.getTaskObjectName())
	if err != nil {
		return errors.Wrap(err, "failed to get storageclass")
	}

	util.SetNestedField(m.Values, sc, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// putStoragePool will create a CStorPool as defined in the task
func (m *executor) putStoragePool() (err error) {
	c, err := m.asStoragePool()
	if err != nil {
		return errors.Wrap(err, "failed to create storage pool")
	}

	storagePool, err := m.getK8sClient().CreateOEV1alpha1SPAsRaw(c)
	if err != nil {
		return errors.Wrap(err, "failed to create storage pool")
	}

	util.SetNestedField(m.Values, storagePool, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// putCStorVolume will create a CStorPool as defined in the task
func (m *executor) putCStorPool() (err error) {
	c, err := m.asCStorPool()
	if err != nil {
		return errors.Wrap(err, "failed to create cstor pool")
	}

	cstorPool, err := m.getK8sClient().CreateOEV1alpha1CSPAsRaw(c)
	if err != nil {
		return errors.Wrap(err, "failed to create cstor pool")
	}

	util.SetNestedField(m.Values, cstorPool, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// putCStorVolume will create a CStorVolume as defined in the task
func (m *executor) putCStorVolume() (err error) {
	c, err := m.asCStorVolume()
	if err != nil {
		return errors.Wrap(err, "failed to create cstor volume")
	}

	cstorVolume, err := m.getK8sClient().CreateOEV1alpha1CVAsRaw(c)
	if err != nil {
		return errors.Wrap(err, "failed to create cstor volume")
	}

	util.SetNestedField(m.Values, cstorVolume, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// putCStorVolumeReplica will create a CStorVolumeReplica as defined in the task
func (m *executor) putCStorVolumeReplica() (err error) {
	d, err := m.asCstorVolumeReplica()
	if err != nil {
		return errors.Wrap(err, "failed to create cstor volume replica")
	}

	cstorVolumeReplica, err := m.getK8sClient().CreateOEV1alpha1CVRAsRaw(d)
	if err != nil {
		return errors.Wrap(err, "failed to create cstor volume replica")
	}

	util.SetNestedField(m.Values, cstorVolumeReplica, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// putUpgradeResult will create an upgrade result as defined in the task
func (m *executor) putUpgradeResult() (err error) {
	raw, err := template.AsTemplatedBytes("UpgradeResult", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return
	}
	uresult, err := upgraderesult.
		BuilderForYAMLObject(raw).
		Build()
	if err != nil {
		return errors.Wrap(err, "failed to create upgrade result")
	}
	uraw, err := upgraderesult.
		NewKubeClient().
		WithNamespace(m.getTaskRunNamespace()).
		CreateRaw(uresult)
	if err != nil {
		return errors.Wrap(err, "failed to create upgrade result")
	}
	util.SetNestedField(m.Values, uraw, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// deleteOEV1alpha1SP will delete one or more StoragePool as specified in
// the RunTask
func (m *executor) deleteOEV1alpha1SP() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err = m.getK8sClient().DeleteOEV1alpha1SP(name)
		if err != nil {
			return errors.Wrap(err, "failed to delete storage pool")
		}
	}

	return
}

// deleteOEV1alpha1CSP will delete one or more CStorPool as specified in
// the RunTask
func (m *executor) deleteOEV1alpha1CSP() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err = m.getK8sClient().DeleteOEV1alpha1CSP(name)
		if err != nil {
			return errors.Wrap(err, "failed to delete cstor pool")
		}
	}

	return
}

// deleteOEV1alpha1CSV will delete one or more CStorVolume as specified in
// the RunTask
func (m *executor) deleteOEV1alpha1CSV() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err = m.getK8sClient().DeleteOEV1alpha1CSV(name)
		if err != nil {
			return errors.Wrap(err, "failed to delete cstor volume")
		}
	}

	return
}

// execCoreV1Pod runs given command remotely in given container of given pod
// and post stdout and and stderr in JsonResult. You can get it using -
// {{- jsonpath .JsonResult "{.stdout}" | trim | saveAs "XXX" .TaskResult | noop -}}
func (m *executor) execCoreV1Pod() error {
	raw, err := template.AsTemplatedBytes("execCoreV1Pod", m.Runtask.Spec.Task, m.Values)
	if err != nil {
		return errors.Wrap(err, "failed to run templating on pod exec object")
	}
	podexecopts, err := podexec.BuilderForYAMLObject(raw).AsAPIPodExec()
	if err != nil {
		return errors.Wrap(err, "failed to build pod exec options")
	}
	execRaw, err := pod.NewKubeClient().
		WithNamespace(m.getTaskRunNamespace()).
		ExecRaw(m.getTaskObjectName(), podexecopts)
	if err != nil {
		return errors.Wrap(err, "failed to run pod exec")
	}

	util.SetNestedField(m.Values, execRaw, string(v1alpha1.CurrentJSONResultTLP))
	return nil
}

// rolloutStatus generates rollout status of a given resource form it's object details
func (m *executor) rolloutStatus() (err error) {
	if m.MetaExec.isRolloutstatusExtnV1B1Deploy() {
		err = m.extnV1B1DeploymentRollOutStatus()
	} else if m.MetaExec.isRolloutstatusAppsV1Deploy() {
		err = m.appsV1DeploymentRollOutStatus()
	} else {
		err = errors.Errorf("meta task not supported: task details: %+v", m.MetaExec.getTaskIdentity())
	}
	if err != nil {
		err = errors.Wrap(err, "failed to get rollout status ")
	}
	return
}

// listK8sResources will list resources as specified in the RunTask
func (m *executor) listK8sResources() (err error) {
	opts, err := m.MetaExec.getListOptions()
	if err != nil {
		err = errors.Wrap(err, "failed to list k8s resources")
		return
	}

	var op []byte
	kc := m.getK8sClient()

	if m.MetaExec.isListCoreV1Pod() {
		op, err = kc.ListCoreV1PodAsRaw(opts)
	} else if m.MetaExec.isListCoreV1Service() {
		op, err = kc.ListCoreV1ServiceAsRaw(opts)
	} else if m.MetaExec.isListExtnV1B1Deploy() {
		op, err = kc.ListExtnV1B1DeploymentAsRaw(opts)
	} else if m.MetaExec.isListExtnV1B1ReplicaSet() {
		op, err = m.listExtnV1B1ReplicaSet(opts)
	} else if m.MetaExec.isListV1alpha1VolumeSnapshotData() {
		op, err = m.listV1alpha1VolumeSnapshotData(opts)
	} else if m.MetaExec.isListV1alpha1VolumeSnapshot() {
		op, err = m.listV1alpha1VolumeSnapshot(opts)
	} else if m.MetaExec.isListAppsV1B1Deploy() {
		op, err = kc.ListAppsV1B1DeploymentAsRaw(opts)
	} else if m.MetaExec.isListAppsV1Deploy() {
		op, err = kc.ListAppsV1DeploymentAsRaw(opts)
	} else if m.MetaExec.isListCoreV1PVC() {
		op, err = kc.ListCoreV1PVCAsRaw(opts)
	} else if m.MetaExec.isListCoreV1PV() {
		op, err = kc.ListCoreV1PVAsRaw(opts)
	} else if m.MetaExec.isListOEV1alpha1BlockDevice() {
		op, err = kc.ListOEV1alpha1BlockDeviceRaw(opts)
	} else if m.MetaExec.isListOEV1alpha1SP() {
		op, err = kc.ListOEV1alpha1SPRaw(opts)
	} else if m.MetaExec.isListOEV1alpha1CSP() {
		op, err = kc.ListOEV1alpha1CSPRaw(opts)
	} else if m.MetaExec.isListOEV1alpha1CVR() {
		op, err = kc.ListOEV1alpha1CVRRaw(opts)
	} else if m.MetaExec.isListOEV1alpha1CV() {
		op, err = kc.ListOEV1alpha1CVRaw(opts)
	} else if m.MetaExec.isListOEV1alpha1UR() {
		op, err = m.listOEV1alpha1URRaw(opts)
	} else {
		err = errors.Errorf("meta task not supported: task details '%+v'", m.MetaExec.getTaskIdentity())
	}

	if err != nil {
		err = errors.Wrap(err, "failed to list k8s resources")
		return
	}

	// set the json doc result
	util.SetNestedField(m.Values, op, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// listOEV1alpha1URRaw fetches a list of UpgradeResults as per the
// provided options
func (m *executor) listOEV1alpha1URRaw(opts metav1.ListOptions) (result []byte, err error) {
	urList, err := upgraderesult.NewKubeClient().
		WithNamespace(m.getTaskRunNamespace()).
		List(opts)
	if err != nil {
		err = errors.Wrap(err, "failed to list upgraderesult")
		return
	}
	result, err = json.Marshal(urList)
	if err != nil {
		err = errors.Wrapf(err, "failed to list upgraderesult: %s", urList.String())
	}
	return
}
