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
)

// RepeatWithResource provides properties that influence task execution's
// repetition behaviour
type RepeatWithResource struct {
	// resources is a list of resources that will drive the repetition logic
	// of task execution
	//
	// NOTE:
	//  This is typically a set of items which does not belong to MetaTask
	// category. Any random list of items that influences the repetition logic
	// should be set here.
	Resources []string `json:"resources"`
	// metas is a list of meta task info that will drive the repetition logic
	// of task execution
	Metas []MetaTaskProps `json:"metas"`
}

// repeatExecutor exposes operations with respect to repeat resources
type repeatExecutor struct {
	repeatResource RepeatWithResource
}

// newRepeatExecutor returns a new instance of repeatExecutor
func newRepeatExecutor(repeat RepeatWithResource) (repeatExecutor, error) {
	if len(repeat.Resources) > 0 && len(repeat.Metas) > 0 {
		return repeatExecutor{}, fmt.Errorf("failed to create repeat executor instance: either 'resources' or 'metas' can be used with 'repeatWith': '%+v'", repeat)
	}

	return repeatExecutor{
		repeatResource: RepeatWithResource{
			Resources: repeat.Resources,
			Metas:     repeat.Metas,
		},
	}, nil
}

// isRepeat flags if there is any requirement to repeat the task execution
func (r repeatExecutor) isRepeat() bool {
	return len(r.repeatResource.Resources) > 0 || len(r.repeatResource.Metas) > 0
}

// isMetaRepeat flags if there is any requirement to repeat the task execution
// based on multiple meta tasks
func (r repeatExecutor) isMetaRepeat() bool {
	return len(r.repeatResource.Metas) > 0
}

// getResources returns the list of resources based on which a task will get
// executed multiple times, each time depending on exactly one resource.
func (r repeatExecutor) getResources() []string {
	return r.repeatResource.Resources
}

// getMetas returns the list of meta tasks based on which a task will get
// executed multiple times, each time depending on exactly one meta task.
func (r repeatExecutor) getMetas() []MetaTaskProps {
	return r.repeatResource.Metas
}

// len returns the count of repeats i.e. either length of resources or length
// of metas
func (r repeatExecutor) len() int {
	if r.isMetaRepeat() {
		return len(r.getMetas())
	}

	return len(r.getResources())
}

// getResource returns the repeat item based on the provided index
func (r repeatExecutor) getResource(index int) (string, error) {
	count := len(r.repeatResource.Resources)
	if index >= count {
		return "", fmt.Errorf("failed to fetch repeat resource: invalid index '%d' w.r.t count '%d'", index, count)
	}

	return r.repeatResource.Resources[index], nil
}

// getMeta returns the repeat meta task item based on the provided index
func (r repeatExecutor) getMeta(index int) (MetaTaskProps, error) {
	count := len(r.repeatResource.Metas)
	if index >= count {
		return MetaTaskProps{}, fmt.Errorf("failed to fetch repeat meta task: invalid index '%d' w.r.t count '%d'", index, count)
	}

	return r.repeatResource.Metas[index], nil
}

// getMetaAsString returns the repeat meta task item based on the provided index
func (r repeatExecutor) getMetaAsString(index int) (string, error) {
	m, err := r.getMeta(index)
	if err != nil {
		return "", err
	}

	return m.toString(), nil
}

// getItem returns the repeat item based on the provided index
func (r repeatExecutor) getItem(index int) (string, error) {
	if r.isMetaRepeat() {
		return r.getMetaAsString(index)
	}

	return r.getResource(index)
}
