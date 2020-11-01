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

package v1alpha1

// jivaVolumeCommand represents a jiva volume command
//
// NOTE:
//  This is an implementation of Runner
type jivaVolumeCommand struct {
	*RunCommand
}

// instance returns specific jiva volume command implementation based
// on the command's action
func (c *jivaVolumeCommand) instance() (r Runner) {
	switch c.Action {
	case DeleteCommandAction:
		r = &jivaVolumeDelete{c}
	default:
		r = &notSupportedActionCommand{c.RunCommand}
	}
	return
}

// Run executes various jiva volume related operations
func (c *jivaVolumeCommand) Run() (r RunCommandResult) {
	return c.instance().Run()
}
