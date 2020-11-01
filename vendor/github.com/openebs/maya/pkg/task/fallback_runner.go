/*
Copyright 2018 The OpenEBS Authors

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
	"github.com/openebs/maya/pkg/client/k8s"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// RunOptions represents the various properties required during run
type RunOptions struct {
	TaskGroupRunner
	// taskFetcher enables fetching a runtask instance
	taskFetcher TaskSpecFetcher
	// values represent the template values to be used while executing cas
	// template
	values map[string]interface{}
}

// RunOptionsMiddleware abstracts updating the given RunOptions instance
type RunOptionsMiddleware func(given *RunOptions) (updated *RunOptions, err error)

// UpdateTaskRunner updates the task runner instance by executing the instance
// against the list of updaters
func UpdateTaskRunner(updaters []RunOptionsMiddleware) RunOptionsMiddleware {
	return func(given *RunOptions) (updated *RunOptions, err error) {
		for _, u := range updaters {
			given, err = u(given)
			if err != nil {
				return
			}
		}
		return given, nil
	}
}

// WithTaskFetcher updates the given RunOptions instance with task fetcher
// instance
func WithTaskFetcher(namespace string) RunOptionsMiddleware {
	return func(given *RunOptions) (updated *RunOptions, err error) {
		f, err := NewK8sTaskSpecFetcher(namespace)
		if err != nil {
			err = errors.Wrapf(err, "failed to update task runner with task fetcher")
			return
		}
		given.taskFetcher = f
		return given, nil
	}
}

// WithRunTaskList updates the given RunOptions instance with list of RunTasks
func WithRunTaskList(taskNames []string) RunOptionsMiddleware {
	return func(given *RunOptions) (updated *RunOptions, err error) {
		if given == nil {
			err = fmt.Errorf("nil task runner: failed to update task runner with task list")
			return
		}

		if given.taskFetcher == nil {
			err = fmt.Errorf("nil task fetcher: failed to update task runner with task list")
			return
		}

		for _, name := range taskNames {
			runtask, err := given.taskFetcher.Fetch(name)
			if err != nil {
				err = errors.Wrapf(err, "failed to update task runner with task list")
				return nil, err
			}

			err = given.AddRunTask(runtask)
			if err != nil {
				err = errors.Wrapf(err, "failed to update task runner with task list")
				return nil, err
			}
		}

		return given, nil
	}
}

// WithOutputTask updates the given RunOptions instance with output task
func WithOutputTask(taskName string) RunOptionsMiddleware {
	return func(given *RunOptions) (updated *RunOptions, err error) {
		if len(strings.TrimSpace(taskName)) == 0 {
			// nothing needs to be done
			return
		}
		if given == nil {
			err = fmt.Errorf("nil task runner: failed to update task runner with output task")
			return
		}
		if given.taskFetcher == nil {
			err = fmt.Errorf("nil task fetcher: failed to update task runner with output task")
			return
		}
		optask, err := given.taskFetcher.Fetch(taskName)
		if err != nil {
			err = errors.Wrapf(err, "failed to update task runner with output task")
			return
		}
		err = given.SetOutputTask(optask)
		if err != nil {
			err = errors.Wrapf(err, "failed to update task runner with output task")
			return
		}
		return given, nil
	}
}

// NewFallbackRunner returns a new instance of task group runner
func NewFallbackRunner(template string, values map[string]interface{}) (*RunOptions, error) {
	if len(strings.TrimSpace(template)) == 0 {
		return nil, fmt.Errorf("missing fallback template name: failed to create fallback runner")
	}

	k, err := k8s.NewK8sClient("")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create fallback runner")
	}

	cast, err := k.GetOEV1alpha1CAST(template, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create fallback runner")
	}

	options := &RunOptions{values: values}

	options, err = UpdateTaskRunner(
		[]RunOptionsMiddleware{
			WithTaskFetcher(cast.Spec.TaskNamespace),
			WithRunTaskList(cast.Spec.RunTasks.Tasks),
			WithOutputTask(cast.Spec.OutputTask),
		})(options)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create fallback runner")
	}

	return options, nil
}

// RunFallback executes the fallback tasks
func RunFallback(options *RunOptions) (output []byte, err error) {
	return options.Run(options.values)
}
