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

package templatefuncs

import (
	"bytes"
	"encoding/json"

	"k8s.io/client-go/util/jsonpath"
)

// QueryType is a typed string to indicate the type of query to use to extract
// data from a taskresult
type QueryType string

const (
	// JsonQT represents the json query to extract data, mostly from a json doc
	JsonQT QueryType = "jsonpath"
	// GoQT represents the go template function used to extract data, mostly
	// from a yaml doc
	GoQT QueryType = "go-template"
)

// Query represents a templating interface which provides Execute
// method that is implemented by specific templating implementations
type Query interface {
	Query() (string, error)
}

// JsonQuery contains parameters to describe a Query.
type JsonQuery struct {
	// name given to this json query operation
	name string
	// jsondoc is the json doc against which json path will be run
	jsondoc []byte
	// path represents the json path used to query the json doc
	path string
}

// NewJsonQuery takes json query operation name, json doc, json path and
// returns a struct of type JsonQuery.
func NewJsonQuery(name string, jsondoc []byte, path string) *JsonQuery {
	return &JsonQuery{
		name:    name,
		jsondoc: jsondoc,
		path:    path,
	}
}

// Query will run jsonpath against the json document and return the value at the
// specified path
func (m *JsonQuery) Query() (output string, err error) {
	// get a new jsonpath instance
	j := jsonpath.New(m.name)
	j.AllowMissingKeys(true)

	// set the parse path i.e. jsonpath
	err = j.Parse(m.path)
	if err != nil {
		return
	}

	var values interface{}
	err = json.Unmarshal(m.jsondoc, &values)
	if err != nil {
		return
	}

	buf := new(bytes.Buffer)
	err = j.Execute(buf, values)
	if err != nil {
		return
	}

	output = buf.String()
	return
}
