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

package v1alpha1

import (
	"strings"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/task"
	"github.com/openebs/maya/pkg/util"
	errors "github.com/pkg/errors"
)

// UnMarshallToConfig un-marshals the provided
// cas template config in a yaml string format
// to a typed list of cas template config
func UnMarshallToConfig(config string) (configs []v1alpha1.Config, err error) {
	err = yaml.Unmarshal([]byte(config), &configs)
	return
}

// MergeConfig will merge configuration fields
// from lowPriority that are not present in
// highPriority configuration and return the
// resulting config
func MergeConfig(highPriority, lowPriority []v1alpha1.Config) (final []v1alpha1.Config) {
	var book []string
	for _, h := range highPriority {
		final = append(final, h)
		book = append(book, strings.TrimSpace(h.Name))
	}
	for _, l := range lowPriority {
		// include only if the config was not present
		// earlier in high priority configuration
		if !util.ContainsString(book, strings.TrimSpace(l.Name)) {
			final = append(final, l)
		}
	}
	return
}

// ConfigToMap transforms CAS template config type
// to a nested map
func ConfigToMap(all []v1alpha1.Config) (m map[string]interface{}, err error) {
	var configName string
	m = map[string]interface{}{}
	for _, config := range all {
		configName = strings.TrimSpace(config.Name)
		if len(configName) == 0 {
			err = errors.Errorf("failed to transform cas config to map: missing config name: %s", config)
			return nil, err
		}
		confHierarchy := map[string]interface{}{
			configName: map[string]string{
				string(v1alpha1.EnabledPTP): config.Enabled,
				string(v1alpha1.ValuePTP):   config.Value,
			},
		}
		isMerged := util.MergeMapOfObjects(m, confHierarchy)
		if !isMerged {
			err = errors.Errorf("failed to transform cas config to map: failed to merge: %s", config)
			return nil, err
		}
	}
	return
}

// Interface abstracts various operations
// exposed by cas template engine
type Interface interface {
	Configurer
	Runner
}

// Configurer abstracts configuring
// a cas template object
type Configurer interface {
	SetConfig(values map[string]interface{})
	SetValues(key string, values map[string]interface{})
}

// Runner abstracts execution of
// cas template engine
type Runner interface {
	Run() (output []byte, err error)
}

// engine implements various cas template
// related operations
type engine struct {
	cast            *v1alpha1.CASTemplate  // refers to a CASTemplate
	values          map[string]interface{} // values used during go templating
	key             string                 // a path to store template values
	taskSpecFetcher task.TaskSpecFetcher   // to fetch runtask specification
	taskGroupRunner *task.TaskGroupRunner  // runs the tasks in the sequence specified in cas template
}

// initValues provides engine specific
// initialization values to be used during
// template execution
func initValues() (v map[string]interface{}) {
	return map[string]interface{}{
		// configTLP is set to nil, this would
		// be reset to cas template defaults
		// if the action methods i.e. create/read/list/delete
		// do not set this
		string(v1alpha1.ConfigTLP): nil,

		// list items is set as a top level property
		string(v1alpha1.ListItemsTLP): map[string]interface{}{},

		// task result is set as a top level property
		string(v1alpha1.TaskResultTLP): map[string]interface{}{},
	}
}

// Engine returns a new instance of engine
func Engine(cast *v1alpha1.CASTemplate, key string, values map[string]interface{}) (e *engine, err error) {
	if cast == nil {
		err = errors.Errorf("failed to instantiate cas template engine: nil cas template provided")
		return
	}

	f, err := task.NewK8sTaskSpecFetcher(cast.Spec.TaskNamespace)
	if err != nil {
		return
	}

	r := task.NewTaskGroupRunner()
	if r == nil {
		err = errors.Errorf("failed to instantiate cas template engine: nil task group runner: %s", cast)
		return
	}

	ev := initValues()

	// add/override with provided values iif key is not empty
	key = strings.TrimSpace(key)
	if len(key) != 0 {
		ev[key] = values
	}

	e = &engine{
		cast:            cast,
		key:             key,
		values:          ev,
		taskSpecFetcher: f,
		taskGroupRunner: r,
	}
	return
}

// setDefaults sets cas template default
// config values
func (c *engine) setDefaults() (err error) {
	m, err := ConfigToMap(c.cast.Spec.Defaults)
	if err != nil {
		return errors.Wrapf(err, "failed to set default config: %s", c.cast)
	}
	c.values[string(v1alpha1.ConfigTLP)] = m
	return
}

// setDefaultsIfEmptyConfig sets cas template
// default config values if config property
// is not set
func (c *engine) setDefaultsIfEmptyConfig() (err error) {
	if c.values[string(v1alpha1.ConfigTLP)] != nil {
		return
	}
	return c.setDefaults()
}

// setLabels sets cas template's labels as
// template values
func (c *engine) setLabels() {
	c.values[string(v1alpha1.CASTOptionsTLP)] = c.cast.Labels
}

// SetConfig sets cas template config as
// template values
func (c *engine) SetConfig(v map[string]interface{}) {
	c.values[string(v1alpha1.ConfigTLP)] = v
}

// SetValues sets template values to be used
// during cas template execution
//
// NOTE:
//  This is not the same thing as cas config.
// One way to understand this is to enable
// engine to accept run time values, etc that
// may be required to execute cas template.
func (c *engine) SetValues(k string, v map[string]interface{}) {
	if len(k) == 0 {
		return
	}
	c.key = k
	util.SetNestedField(c.values, v, c.key)
}

// prepareTasksForExec prepares the taskGroupRunner
// instance with the info needed to run the tasks
func (c *engine) prepareTasksForExec() error {
	// prepare the tasks mentioned in cas template
	for _, taskName := range c.cast.Spec.RunTasks.Tasks {
		// fetch runtask from task name
		runtask, err := c.taskSpecFetcher.Fetch(taskName)
		if err != nil {
			return errors.Wrapf(err, "failed to prepare task '%s': %s", taskName, c.cast)
		}

		// prepare the task group runner by adding the runtask
		err = c.taskGroupRunner.AddRunTask(runtask)
		if err != nil {
			return errors.Wrapf(err, "failed to prepare task: %s", runtask)
		}
	}
	return nil
}

// prepareOutputTask prepares the taskGroupRunner instance with the
// yaml template which becomes the output of the runner. This is
// invoked after successful execution of all the runtasks.
func (c *engine) prepareOutputTask() (err error) {
	opTaskName := c.cast.Spec.OutputTask
	if len(opTaskName) == 0 {
		// no output task was specified;
		// nothing to do
		return
	}
	runtask, err := c.taskSpecFetcher.Fetch(opTaskName)
	if err != nil {
		err = errors.Wrapf(err, "failed to prepare output task '%s': %s", opTaskName, c.cast)
		return
	}
	return c.taskGroupRunner.SetOutputTask(runtask)
}

// prepareFallback prepares the taskGroupRunner instance with the
// fallback template which is used in case of specific errors e.g. version
// mismatch error
func (c *engine) prepareFallback() {
	f := c.cast.Spec.Fallback
	c.taskGroupRunner.SetFallback(f)
}

// Run executes the cas engine based on the tasks
// specified in the cas template
func (c *engine) Run() (output []byte, err error) {
	c.setLabels()

	err = c.setDefaultsIfEmptyConfig()
	if err != nil {
		err = errors.Wrap(err, "failed to run cas template engine")
		return
	}

	err = c.prepareTasksForExec()
	if err != nil {
		err = errors.Wrap(err, "failed to run cas template engine")
		return
	}

	err = c.prepareOutputTask()
	if err != nil {
		err = errors.Wrap(err, "failed to run cas template engine")
		return
	}

	c.prepareFallback()
	return c.taskGroupRunner.Run(c.values)
}
