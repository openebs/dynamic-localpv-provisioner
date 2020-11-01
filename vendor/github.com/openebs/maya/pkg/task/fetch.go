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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TaskSpecFetcher is the contract to fetch task
// specification that includes the task's meta specification
type TaskSpecFetcher interface {
	Fetch(taskName string) (runtask *v1alpha1.RunTask, err error)
}

// K8sTaskSpecFetcher deals with fetching a task specifications
// from K8s cluster
//
// NOTE:
//  A task is a K8s ConfigMap
type K8sTaskSpecFetcher struct {
	// k8sClient to make K8s API calls
	//
	// NOTE:
	//  This is also helpful for mocking during UT
	k8sClient *m_k8s_client.K8sClient
}

// Fetch returns specifications of a provided task in yaml
// string format
//
// NOTE:
//  This is an implementation of TaskSpecFetcher interface
func (f *K8sTaskSpecFetcher) Fetch(taskName string) (runtask *v1alpha1.RunTask, err error) {
	rtGetter := defaultRunTaskGetter(getRunTaskSpec{
		taskName:  taskName,
		k8sClient: f.k8sClient,
	})

	return rtGetter.Get()
}

// NewK8sTaskSpecFetcher returns a new instance of K8sTaskSpecFetcher based on the
// provided namespace.
//
// NOTE:
//  SearchNamespace refers to the K8s namespace where a task
// is expected to be found
func NewK8sTaskSpecFetcher(searchNamespace string) (*K8sTaskSpecFetcher, error) {
	kc, err := m_k8s_client.NewK8sClient(searchNamespace)
	if err != nil {
		return nil, err
	}

	return &K8sTaskSpecFetcher{
		k8sClient: kc,
	}, nil
}

// runTaskGetter abstracts fetching of runtask instance
type runTaskGetter interface {
	Get() (runtask *v1alpha1.RunTask, err error)
}

// getRunTaskSpec composes common properties required to get a run task
// instance
type getRunTaskSpec struct {
	taskName  string
	k8sClient *m_k8s_client.K8sClient
}

// runTaskGetterFn abstracts fetching of runtask instance based on the provided
// runtask specifications
type runTaskGetterFn func(getSpec getRunTaskSpec) (runtask *v1alpha1.RunTask, err error)

// getRunTaskFromConfigMap fetches runtask instance from a config map instance
func getRunTaskFromConfigMap(g getRunTaskSpec) (runtask *v1alpha1.RunTask, err error) {
	if len(strings.TrimSpace(g.taskName)) == 0 {
		err = fmt.Errorf("missing run task name: failed to get runtask from config map")
		return
	}

	if g.k8sClient == nil {
		err = fmt.Errorf("nil kubernetes client found: failed to get runtask '%s' from config map", g.taskName)
		return
	}

	cm, err := g.k8sClient.GetConfigMap(g.taskName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to get run task '%s' from config map", g.taskName))
		return
	}

	runtask = &v1alpha1.RunTask{}
	runtask.Name = g.taskName
	runtask.Spec.Meta = cm.Data["meta"]
	runtask.Spec.Task = cm.Data["task"]
	runtask.Spec.PostRun = cm.Data["post"]

	return
}

// getRunTaskFromCustomResource fetches runtask instance from its custom
// resource instance
func getRunTaskFromCustomResource(g getRunTaskSpec) (runtask *v1alpha1.RunTask, err error) {
	if len(strings.TrimSpace(g.taskName)) == 0 {
		err = fmt.Errorf("missing run task name: failed to get runtask from custom resource")
		return
	}

	if g.k8sClient == nil {
		err = fmt.Errorf("nil kubernetes client: failed to get runtask '%s' from custom resource", g.taskName)
		return
	}

	runtask, err = g.k8sClient.GetOEV1alpha1RunTask(g.taskName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to get runtask '%s' from custom resource", g.taskName))
	}

	return
}

// getRunTask enables fetching a run task instance based on various run task
// getter strategies
//
// NOTE:
//  This is an implementation of runTaskGetter
type getRunTask struct {
	// getRunTaskSpec is the specifications required to fetch runtask instance
	getRunTaskSpec
	// currentStrategy is the latest strategy to fetch runtask instance
	currentStrategy runTaskGetterFn
	// oldStrategies are the older strategies to fetch runtask instance
	oldStrategies []runTaskGetterFn
}

// Get returns an instance of runtask
func (g *getRunTask) Get() (runtask *v1alpha1.RunTask, err error) {
	var allStrategies []runTaskGetterFn
	if g.currentStrategy != nil {
		allStrategies = append(allStrategies, g.currentStrategy)
	}

	if len(g.oldStrategies) != 0 {
		allStrategies = append(allStrategies, g.oldStrategies...)
	}

	if len(allStrategies) == 0 {
		err = fmt.Errorf("no strategies to get runtask: failed to get runtask '%s'", g.taskName)
		return
	}

	for _, s := range allStrategies {
		runtask, err = s(g.getRunTaskSpec)
		if err == nil {
			return
		}

		err = errors.Wrap(err, fmt.Sprintf("failed to get runtask '%s'", g.taskName))
		klog.Warningf("%s", err)
	}

	// at this point, we have a real error we can not recover from
	err = fmt.Errorf("exhausted all strategies to get runtask: failed to get runtask '%s'", g.taskName)
	return
}

func defaultRunTaskGetter(getSpec getRunTaskSpec) *getRunTask {
	// current strategy
	current := getRunTaskFromCustomResource
	// older strategies
	old := []runTaskGetterFn{getRunTaskFromConfigMap}

	return &getRunTask{
		getRunTaskSpec:  getSpec,
		currentStrategy: current,
		oldStrategies:   old,
	}
}
