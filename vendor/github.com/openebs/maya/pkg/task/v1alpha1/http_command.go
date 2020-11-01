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

import (
	"encoding/json"

	rest "github.com/openebs/maya/pkg/client/http/v1alpha1"
	"github.com/pkg/errors"
)

// TODO
// Make use of interfaces exposed from pkg/client/http/v1alpha1 than using the
// direct structures

// httpCommand represents a http API invocation command
//
// NOTE:
//  This is an implementation of Runner
type httpCommand struct {
	*RunCommand
	url         string
	rest        *rest.Rest
	name        string
	verb        rest.HttpVerb
	body        interface{}
	isUnmarshal bool
}

// HttpCommand returns a new instance of httpCommand
func HttpCommand(c *RunCommand) *httpCommand {
	return &httpCommand{RunCommand: c}
}

// withURL sets either the base url or url with full path
func (c *httpCommand) withURL(u string) *httpCommand {
	c.url = u
	return c
}

// withREST sets the rest instance capable to perform REST invocation
func (c *httpCommand) withREST(r *rest.Rest) *httpCommand {
	c.rest = r
	return c
}

// withName sets the resource name to be used during REST invocation
func (c *httpCommand) withName() *httpCommand {
	n, _ := c.Data["name"].(string)
	c.name = n
	return c
}

// withVerb sets the resource name to be used during REST invocation
func (c *httpCommand) withVerb(v rest.HttpVerb) *httpCommand {
	c.verb = v
	return c
}

// withBody sets the body of the REST request
func (c *httpCommand) withBody() *httpCommand {
	b, _ := c.Data["body"]
	c.body = b
	return c
}

// withUnmarshal is used to do unmarshal of response
func (c *httpCommand) withIsUnmarshal() *httpCommand {
	unmarshalKey, isPresent := c.Data["unmarshal"].(bool)
	if isPresent {
		c.isUnmarshal = unmarshalKey
		return c
	}
	c.isUnmarshal = true
	return c
}

// do invokes the REST call
func (c *httpCommand) do() (r RunCommandResult) {
	res, err := c.rest.WithName(c.name).WithVerb(c.verb).WithBody(c.body).Do()
	if err != nil {
		return c.AddError(err).Result(nil)
	}
	return c.Result(res)
}

// instance returns specific http api command implementation based on the
// command's action
func (c *httpCommand) instance() (r Runner) {
	switch c.Action {
	case DeleteCommandAction:
		r = &httpDelete{c}
	case GetCommandAction:
		r = &httpGet{c}
	case PostCommandAction:
		r = &httpPost{c}
	case PutCommandAction:
		r = &httpPut{c}
	case PatchCommandAction:
		r = &httpPatch{c}
	default:
		r = &notSupportedActionCommand{c.RunCommand}
	}
	return
}

// invokeURL invokes http call using the provided verb
func (c *httpCommand) invokeURL(verb rest.HttpVerb) (b []byte, err error) {
	return rest.URL(verb, c.url)
}

// invokeAPI invokes http call using the provided verb and resource name
func (c *httpCommand) invokeAPI(verb rest.HttpVerb, name string) (b []byte, err error) {
	return rest.API(verb, c.url, name)
}

// invoke invokes http call using the provided http verb
func (c *httpCommand) invoke(verb rest.HttpVerb) (r RunCommandResult) {
	var (
		b   []byte
		err error
		res interface{}
	)
	name, _ := c.Data["name"].(string)
	if len(name) != 0 {
		b, err = c.invokeAPI(verb, name)
	} else {
		b, err = c.invokeURL(verb)
	}
	if err != nil {
		return c.AddError(err).Result(nil)
	}

	if c.isUnmarshal {
		err = json.Unmarshal(b, &res)
		if err != nil {
			return c.AddError(errors.Wrap(err, "failed to invoke http command")).Result(nil)
		}
		return c.Result(res)
	}

	return c.Result(b)
}

// Run executes various jiva volume related operations
func (c *httpCommand) Run() (r RunCommandResult) {
	url, _ := c.Data["url"].(string)
	if len(url) == 0 {
		return c.AddError(errors.New("missing url: failed to invoke http command")).Result(nil)
	}
	return c.withURL(url).withIsUnmarshal().instance().Run()
}

// httpDelete represents a delete http invocation command
//
// NOTE:
//  This is an implementation of Runner
type httpDelete struct {
	*httpCommand
}

// Run invokes delete http call
func (d *httpDelete) Run() (r RunCommandResult) {
	return d.invoke(rest.DeleteAction)
}

// httpGet represents a GET http invocation command
//
// NOTE:
//  This is an implementation of CommandRunner
type httpGet struct {
	*httpCommand
}

// Run invokes GET http call
func (g *httpGet) Run() (r RunCommandResult) {
	return g.invoke(rest.GetAction)
}

// httpPost represents a POST http invocation command
//
// NOTE:
//  This is an implementation of CommandRunner
type httpPost struct {
	*httpCommand
}

// Run invokes POST http call
func (p *httpPost) Run() (r RunCommandResult) {
	robj, err := rest.REST(p.url)
	if err != nil {
		return p.AddError(errors.Wrap(err, "failed to invoke http post command")).Result(nil)
	}
	return p.withREST(robj).withName().withVerb(rest.PostAction).withBody().do()
}

// httpPut represents a PUT http invocation command
//
// NOTE:
//  This is an implementation of CommandRunner
type httpPut struct {
	*httpCommand
}

// Run invokes PUT http call
func (g *httpPut) Run() (r RunCommandResult) {
	return g.invoke(rest.PutAction)
}

// httpPatch represents a PATCH http invocation command
//
// NOTE:
//  This is an implementation of CommandRunner
type httpPatch struct {
	*httpCommand
}

// Run invokes PATCH http call
func (g *httpPatch) Run() (r RunCommandResult) {
	return g.invoke(rest.PatchAction)
}
