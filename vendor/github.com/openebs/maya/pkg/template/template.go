/*
Copyright 2017 The OpenEBS Authors
Copyright 2016 The Kubernetes Authors
Copyright (C) 2013 Masterminds

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

package template

import (
	"bytes"
	"text/template"

	"github.com/openebs/maya/pkg/templatefuncs/v1alpha1"
	yaml "gopkg.in/yaml.v2"
)

// AsTemplatedBytes returns a byte slice
// based on the provided yaml & values
func AsTemplatedBytes(context string, yml string, values map[string]interface{}) ([]byte, error) {
	tpl := template.New(context + "YamlTpl")

	// Any maya yaml exposes below templating functions
	tpl.Funcs(templatefuncs.AllCustomFuncs())

	tpl, err := tpl.Parse(yml)
	if err != nil {
		return nil, err
	}

	// buf is an io.Writer implementation
	// as required by the template
	var buf bytes.Buffer

	// execute the parsed yaml against this instance
	// & write the result into the buffer
	err = tpl.Execute(&buf, values)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// AsMapOfObjects returns a map of objects based on the provided yaml & values
func AsMapOfObjects(yml string, values map[string]interface{}) (map[string]interface{}, error) {
	// templated & then unmarshall-ed version of this yaml
	b, err := AsTemplatedBytes("MapOfObjects", yml, values)
	if err != nil {
		return nil, err
	}

	// Any given YAML can be unmarshalled into a map of arbitrary objects
	var obj map[string]interface{}
	err = yaml.Unmarshal(b, &obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// AsMapOfStrings returns a map of strings based on the provided yaml & values
func AsMapOfStrings(context string, yml string, values map[string]interface{}) (map[string]string, error) {
	// templated & then unmarshall-ed version of this yaml
	b, err := AsTemplatedBytes(context+"MapOfStrings", yml, values)
	if err != nil {
		return nil, err
	}

	// Any given YAML can be unmarshalled into a map of strings
	var obj map[string]string
	err = yaml.Unmarshal(b, &obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
