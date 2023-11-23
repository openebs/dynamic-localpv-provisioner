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
	"fmt"
	"strings"
	"testing"
	"text/template"

	msg "github.com/openebs/maya/pkg/msg/v1alpha1"
	cmd "github.com/openebs/maya/pkg/task/v1alpha1"
)

// TestIsValidTemplateFunctionName verifies if suggested template function names
// are:
//
// 1. accepted by go based templating
func TestIsValidTemplateFunctionName(t *testing.T) {
	tests := map[string]struct {
		name string
	}{
		// possible execution based template functions
		"'do' as function name":     {"do"},
		"'run' as function name":    {"run"},
		"'runlog' as function name": {"runlog"},
		// possible action based template functions
		"'delete' as function name": {"delete"},
		"'get' as function name":    {"get"},
		"'create' as function name": {"create"},
		"'lst' as function name":    {"lst"},
		"'patch' as function name":  {"patch"},
		"'update' as function name": {"update"},
		// possible domain based template functions
		"'jiva' as function name":       {"jiva"},
		"'cstor' as function name":      {"cstor"},
		"'volume' as function name":     {"volume"},
		"'pool' as function name":       {"pool"},
		"'withoption' as function name": {"withoption"},
		"'namespace' as function name":  {"namespace"},
		// possible selector based template functions
		"'select' as function name": {"select"},
		"'where' as function name":  {"where"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			f := template.FuncMap{mock.name: noop}
			tpl := template.New("is.func.name.valid").Funcs(f)
			tpl, err := tpl.Parse(fmt.Sprintf("{{- %s -}}", mock.name))
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			var values map[string]interface{}
			err = tpl.Execute(&buf, values)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}
		})
	}
}

// TestRunCommandTemplatingCombinations verifies the supported templating
// combinations made possible via `runtask command based templating functions`
func TestRunCommandTemplatingCombinations(t *testing.T) {
	tests := map[string]struct {
		templatefunc string
	}{
		// general jiva combinations
		"test 101": {"delete jiva volume | run"},
		"test 102": {"delete volume jiva | run"},
		"test 103": {"create jiva volume | run"},
		"test 104": {"create volume jiva | run"},
		"test 105": {"get jiva volume | run"},
		"test 106": {"get volume jiva | run"},
		"test 107": {"lst jiva volume | run"},
		"test 108": {"lst volume jiva | run"},
		// general cstor combinations
		"test 201": {"delete cstor volume | run"},
		"test 202": {"delete volume cstor | run"},
		"test 203": {"create cstor volume | run"},
		"test 204": {"create volume cstor | run"},
		"test 205": {"get cstor volume | run"},
		"test 206": {"get volume cstor | run"},
		"test 207": {"lst cstor volume | run"},
		"test 208": {"lst volume cstor | run"},
		// with withoption as input data combinations
		"test 301": {`delete jiva volume | withoption "url" "http://" | run`},
		"test 302": {`delete volume jiva | withoption "url" "http://" | run`},
		"test 303": {`delete cstor volume | withoption "url" "http://" | run`},
		"test 304": {`delete volume cstor | withoption "url" "http://" | run`},
		// with select path combinations
		"test 401": {`select "name" "namespace" | get jiva volume | run`},
		"test 402": {`select "all" | lst jiva volume | run`},
		"test 403": {`select "name" "namespace" | get cstor volume | run`},
		"test 404": {`select "all" | lst cstor volume | run`},
		// with runlog combinations
		"test 501": {`lst volume jiva | runlog "id.result" "id.extras" .Values`},
		"test 502": {`create volume jiva | runlog "id.result" "id.extras" .Values`},
		"test 503": {`get volume jiva | runlog "id.result" "id.extras" .Values`},
		"test 504": {`update volume jiva | runlog "id.result" "id.extras" .Values`},
		"test 505": {`patch volume jiva | runlog "id.result" "id.extras" .Values`},
		"test 506": {`delete volume jiva | runlog "id.result" "id.extras" .Values`},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			allfuncs := AllCustomFuncs()
			tpl := template.New(fmt.Sprintf("%s", mock.templatefunc)).Funcs(allfuncs)
			tpl, err := tpl.Parse(fmt.Sprintf("{{- %s -}}", mock.templatefunc))
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			values := map[string]interface{}{
				"Values": map[string]interface{}{},
			}
			err = tpl.Execute(&buf, values)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}
		})
	}
}

func TestNotSupportedActionCommand(t *testing.T) {
	mockval := map[string]interface{}{"Values": map[string]interface{}{}}

	tests := map[string]struct {
		template string
		tvalues  map[string]interface{}
	}{
		// NOTE: If these combinations are supported in future then remove the
		// test case(s)
		"test 101": {`{{- create jiva volume | run -}}`, mockval},
		"test 102": {`{{- lst jiva volume | run -}}`, mockval},
		"test 103": {`{{- patch jiva volume | run -}}`, mockval},
		"test 104": {`{{- get jiva volume | run -}}`, mockval},
		"test 105": {`{{- update jiva volume | run -}}`, mockval},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			allfuncs := AllCustomFuncs()
			tpl := template.New(mock.template).Funcs(allfuncs)
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mock.tvalues)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			if len(buf.String()) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}

			if !strings.Contains(buf.String(), cmd.ErrorNotSupportedAction.Error()) {
				t.Fatalf("Test '%s' failed: expected error '%s' actual: '%s'", name, cmd.ErrorNotSupportedAction.Error(), buf.String())
			}
		})
	}
}

func TestNotSupportedCategoryCommand(t *testing.T) {
	mockval := map[string]interface{}{"Values": map[string]interface{}{}}

	tests := map[string]struct {
		template string
		tvalues  map[string]interface{}
	}{
		// NOTE: If these combinations are supported in future then remove the
		// test case(s)
		"test 101": {`{{- create cstor volume | run -}}`, mockval},
		"test 102": {`{{- lst cstor volume | run -}}`, mockval},
		"test 103": {`{{- patch cstor volume | run -}}`, mockval},
		"test 104": {`{{- get cstor volume | run -}}`, mockval},
		"test 105": {`{{- update cstor volume | run -}}`, mockval},
		"test 106": {`{{- delete cstor volume | run -}}`, mockval},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			allfuncs := AllCustomFuncs()
			tpl := template.New(mock.template).Funcs(allfuncs)
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mock.tvalues)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			if len(buf.String()) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}

			if !strings.Contains(buf.String(), cmd.ErrorNotSupportedCategory.Error()) {
				t.Fatalf("Test '%s' failed: expected error '%s' actual: '%s'", name, cmd.ErrorNotSupportedCategory.Error(), buf.String())
			}
		})
	}
}

func TestDeleteJivaVolumeCommand(t *testing.T) {
	mockval := map[string]interface{}{"Values": map[string]interface{}{}}

	tests := map[string]struct {
		template string
		tvalues  map[string]interface{}
	}{
		"test 101": {`{{- delete jiva volume | run -}}`, mockval},
		"test 102": {`{{- delete jiva volume | withoption "url" "" | run -}}`, mockval},
		"test 103": {`{{- delete jiva volume | withoption "url" "http://" | run -}}`, mockval},
		"test 104": {`{{- delete cstor volume | withoption "url" "http://0.0.0.0:1010/v1" | run -}}`, mockval},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			allfuncs := AllCustomFuncs()
			tpl := template.New(mock.template).Funcs(allfuncs)
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mock.tvalues)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			op := buf.String()
			if len(op) == 0 {
				t.Fatalf("Test '%s' failed: expected result: actual no result", name)
			}
		})
	}
}

func TestCreateCstorSnapshotCommand(t *testing.T) {
	mockval := map[string]interface{}{"Values": map[string]interface{}{}}

	tests := map[string]struct {
		template string
		tvalues  map[string]interface{}
	}{
		"test 101": {`{{- create cstor snapshot | run -}}`, mockval},
		"test 102": {`{{- create cstor snapshot | withoption "ip" "0.0.0.0" | withoption "volname" "vol1" | withoption "snapname" "" | run -}}`, mockval},
		"test 103": {`{{- create cstor snapshot | withoption "ip" "" | withoption "volname" "" | withoption "snapname" "snap1" | run -}}`, mockval},
		"test 104": {`{{- create cstor snapshot | withoption "ip" "0.0.0.0" | run -}}`, mockval},

		// Commenting this test case, as go test times out
		// TODO
		// Need to push this to Integration Test
		//
		//"test 105": {`{{- $ip := "0.0.0.0" -}}
		//			  {{- $volName := "vol1" -}}
		//			  {{- $snapName := "s1" -}}
		//   		  {{- $runCommandTemp := create cstor volume | withoption "ip" $ip | withoption "volname" $volName -}}
		//			  {{- $runCommandTemp := $runCommandTemp | withoption "snapname" $snapName -}}
		//			  {{- $runCommandTemp | run -}}`, mockval},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			allfuncs := AllCustomFuncs()
			tpl := template.New(mock.template).Funcs(allfuncs)
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mock.tvalues)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			op := buf.String()
			if len(op) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}
		})
	}
}

func TestDeleteJivaVolumeSaveAs(t *testing.T) {
	tests := map[string]struct {
		commandName string
		template    string
	}{
		"101": {"101", `{{- delete jiva volume | run | saveas "101" .Values -}}`},
		"102": {"102", `{{- delete jiva volume | withoption "url" "" | withoption "name" "" | run | saveas "102" .Values -}}`},
		"103": {"103", `{{- delete jiva volume | withoption "url" "http://" | withoption "name" "ab" | run | saveas "103" .Values -}}`},
		"104": {"104",
			`{{- $url := "http://0.0.0.0:1010/v1" -}}
		   {{- delete jiva volume | withoption "url" $url | withoption "name" "abcd" | run | saveas "104" .Values -}}`},
		"105": {"105",
			`{{- $url := "http://0.0.0.0:1010/v1/volumes" -}}
		  {{- delete jiva volume | withoption "url" $url | withoption "name" "abcde" | run | saveas "105" .Values -}}`},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			mockval := map[string]interface{}{"Values": map[string]interface{}{}}
			tpl := template.New(mock.template).Funcs(AllCustomFuncs())
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mockval)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			op := buf.String()
			if len(op) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}

			tvalues := mockval["Values"]
			if tvalues == nil {
				t.Fatalf("Test '%s' failed: nil template values post template execution", name)
			}

			response := tvalues.(map[string]interface{})[mock.commandName]
			if response == nil {
				t.Fatalf("Test '%s' failed: nil runtask command response post template execution", name)
			}

			cmdres := response.(map[string]interface{})
			cmdresult := cmdres["result"]
			if cmdresult != nil {
				t.Fatalf("Test '%s' failed: expected nil result: actual result %#v", name, cmdresult)
			}

			cmderr := cmdres["error"]
			if cmderr == nil {
				t.Fatalf("Test '%s' failed: expected not nil error: actual nil error", name)
			}

			cmddebug := cmdres["debug"]
			if cmddebug == nil {
				t.Fatalf("Test '%s' failed: expected not nil debug: actual nil debug", name)
			}
		})
	}
}

func TestDeleteJivaVolumeSaveAsVerifyError(t *testing.T) {
	tests := map[string]struct {
		commandName string
		template    string
	}{
		"t101": {"deljivavol",
			`{{- $url := "http://" -}}
		   {{- delete jiva volume | withoption "url" $url | withoption "name" "myvol" | run | saveas "deljivavol" .Values -}}
		   {{- $err := toString .Values.deljivavol.error -}}
		   {{- $err | empty | not | verifyErr $err | saveif "deljivavol.verifyerr" .Values | noop -}}`},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			mockval := map[string]interface{}{"Values": map[string]interface{}{}}
			tpl := template.New(mock.template).Funcs(AllCustomFuncs())
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mockval)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			op := buf.String()
			if len(op) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}

			tvalues := mockval["Values"]
			if tvalues == nil {
				t.Fatalf("Test '%s' failed: nil template values post template execution", name)
			}

			response := tvalues.(map[string]interface{})[mock.commandName]
			if response == nil {
				t.Fatalf("Test '%s' failed: nil runtask command response post template execution", name)
			}

			cmdres := response.(map[string]interface{})
			verifyerr := cmdres["verifyerr"]
			if verifyerr == nil {
				t.Fatalf("Test '%s' failed: expected not nil verifyerr: actual verifyerr %#v", name, verifyerr)
			}

			var verr error
			verr, ok := verifyerr.(*VerifyError)
			if !ok {
				t.Fatalf("Test '%s' failed: expected VerifyErr error: actual %#v", name, verr)
			}
		})
	}
}

func TestSkipOnError(t *testing.T) {
	tests := map[string]struct {
		commandToTest string
		template      string
		resultCount   int
		isErr         bool
		isSkip        bool
		isInfo        bool
	}{
		"t101": {"delJivaVol",
			`{{- $store :=  storeAt .Values -}}
       {{- $runner := storeRunner $store -}}
		   {{- $url := "http://" -}}
		   {{- delete jiva volume | withoption "url" $url | withoption "name" "myvol" | runas "delJivaVol" $runner -}}`,
			1, true, false, true},
		"t102": {"delCstorVol",
			`{{- $store :=  storeAt .Values -}}
       {{- $runner := storeRunner $store -}}
		   {{- $url := "http://" -}}
		   {{- delete jiva volume | withoption "url" $url | withoption "name" "myvol1" | runas "delJivaVol" $runner -}}
		   {{- delete cstor volume | withoption "url" $url | withoption "name" "myvol2" | runas "delCstorVol" $runner -}}`,
			2, true, true, true},
		"t103": {"createCstorVol",
			`{{- $store :=  storeAt .Values -}}
       {{- $runner := storeRunner $store -}}
		   {{- $url := "http://" -}}
		   {{- delete jiva volume | withoption "url" $url | withoption "name" "myvol1" | runas "delJivaVol" $runner -}}
		   {{- delete cstor volume | withoption "name" "vol2" | runas "delCvol2" $runner -}}
		   {{- get cstor volume | withoption "name" "vol3" | runas "getCvol3" $runner -}}
		   {{- get cstor volume | withoption "name" "vol4" | runas "getCvol4" $runner -}}
		   {{- lst cstor volume | runas "listCstorVol" $runner -}}
		   {{- create cstor volume | withoption "url" $url | withoption "name" "vol5" | runas "createCstorVol" $runner -}}`,
			6, true, true, true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			mockval := map[string]interface{}{"Values": map[string]interface{}{}}
			tpl := template.New(mock.template).Funcs(AllCustomFuncs())
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mockval)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			op := buf.String()
			if len(op) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}

			tvalues := mockval["Values"]
			if tvalues == nil {
				t.Fatalf("Test '%s' failed: nil template values post template execution", name)
			}

			results := tvalues.(map[string]interface{})
			mockResultCount := mock.resultCount

			if results["rootCause"] != nil {
				mockResultCount = mockResultCount + 1
			}

			if len(results) != mockResultCount {
				t.Fatalf("Test '%s' failed: expected run command results count '%d': actual '%d'", name, mock.resultCount, len(results))
			}

			response := results[mock.commandToTest]
			if response == nil {
				t.Fatalf("Test '%s' failed: nil runtask command response post template execution", name)
			}

			cmdres := response.(map[string]interface{})
			err = cmdres["error"].(error)
			if mock.isErr && err == nil {
				t.Fatalf("Test '%s' failed: expected not nil err: actual %#v", name, cmdres)
			}

			debug := cmdres["debug"].(msg.AllMsgs)
			if mock.isErr && len(debug[msg.ErrMsg].Items) == 0 {
				t.Fatalf("Test '%s' failed: expected error messages: actual %s", name, debug)
			}
			if mock.isSkip && len(debug[msg.SkipMsg].Items) == 0 {
				t.Fatalf("Test '%s' failed: expected skip messages: actual %s", name, debug)
			}
			if mock.isInfo && len(debug[msg.InfoMsg].Items) == 0 {
				t.Fatalf("Test '%s' failed: expected info messages: actual %s", name, debug)
			}
		})
	}
}

func TestGetHttp(t *testing.T) {
	tests := map[string]struct {
		commandToTest string
		template      string
		resultCount   int
		isErr         bool
		isSkip        bool
		isInfo        bool
	}{
		// test scenario with one get http request
		"t101": {"step1",
			`{{- $store :=  storeAt .Values -}}
       {{- $runner := storeRunner $store -}}
		   {{- $url := "http://" -}}
		   {{- get http | withoption "url" $url | withoption "name" "myvol" | runas "step1" $runner -}}`,
			1, true, false, true},
		// test scenario with more than one get http requests
		"t102": {"step2",
			`{{- $store :=  storeAt .Values -}}
       {{- $runner := storeRunner $store -}}
		   {{- $url := "http://" -}}
		   {{- get http | withoption "url" $url | withoption "name" "myvol1" | runas "step1" $runner -}}
		   {{- get http | withoption "url" $url | withoption "name" "myvol2" | runas "step2" $runner -}}`,
			2, true, true, true},
		// test scenario with a series of get http requests
		"t103": {"step3",
			`{{- $store :=  storeAt .Values -}}
       {{- $runner := storeRunner $store -}}
		   {{- $url := "http://" -}}
		   {{- get http | withoption "url" $url | withoption "name" "vol1" | runas "step1" $runner -}}
		   {{- get http | withoption "url" $url | withoption "name" "vol2" | runas "step2" $runner -}}
 		   {{- get http | withoption "url" $url | withoption "name" "vol3" | runas "step3" $runner -}}
 		   {{- get http | withoption "url" $url | withoption "name" "vol4" | runas "step4" $runner -}}
 		   {{- get http | withoption "url" $url | withoption "name" "vol4" | runas "step5" $runner -}}`,
			5, true, true, true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			mockval := map[string]interface{}{"Values": map[string]interface{}{}}
			tpl := template.New(mock.template).Funcs(AllCustomFuncs())
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mockval)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			op := buf.String()
			if len(op) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}

			tvalues := mockval["Values"]
			if tvalues == nil {
				t.Fatalf("Test '%s' failed: nil template values post template execution", name)
			}

			results := tvalues.(map[string]interface{})
			mockResultCount := mock.resultCount

			if results["rootCause"] != nil {
				mockResultCount = mockResultCount + 1
			}

			if len(results) != mockResultCount {
				t.Fatalf("Test '%s' failed: expected run command results count '%d': actual '%d'", name, mock.resultCount, len(results))
			}

			response := results[mock.commandToTest]
			if response == nil {
				t.Fatalf("Test '%s' failed: nil runtask command response post template execution", name)
			}

			cmdres := response.(map[string]interface{})
			err = cmdres["error"].(error)
			if mock.isErr && err == nil {
				t.Fatalf("Test '%s' failed: expected not nil err: actual %#v", name, cmdres)
			}

			debug := cmdres["debug"].(msg.AllMsgs)
			if mock.isErr && len(debug[msg.ErrMsg].Items) == 0 {
				t.Fatalf("Test '%s' failed: expected error messages: actual %s", name, debug)
			}
			if mock.isSkip && len(debug[msg.SkipMsg].Items) == 0 {
				t.Fatalf("Test '%s' failed: expected skip messages: actual %s", name, debug)
			}
			if mock.isInfo && len(debug[msg.InfoMsg].Items) == 0 {
				t.Fatalf("Test '%s' failed: expected info messages: actual %s", name, debug)
			}
		})
	}
}

func TestPostHttp(t *testing.T) {
	tests := map[string]struct {
		commandToTest string
		template      string
		resultCount   int
		isErr         bool
		isSkip        bool
		isInfo        bool
	}{
		// test scenario with one post http request
		"t101": {"step1",
			`{{- $store :=  storeAt .Values -}}
       {{- $runner := storeRunner $store -}}
		   {{- $url := "http://" -}}
		   {{- post http | withoption "url" $url | withoption "name" "myvol" | runas "step1" $runner -}}`,
			1, true, false, true},
		// test scenario with more than one post http requests
		"t102": {"step2",
			`{{- $store :=  storeAt .Values -}}
       {{- $runner := storeRunner $store -}}
		   {{- $url := "http://" -}}
		   {{- post http | withoption "url" $url | withoption "name" "myvol1" | runas "step1" $runner -}}
		   {{- post http | withoption "url" $url | withoption "body" "myvol2" | runas "step2" $runner -}}`,
			2, true, true, true},
		// test scenario with a series of post http requests
		"t103": {"step3",
			`{{- $store :=  storeAt .Values -}}
       {{- $runner := storeRunner $store -}}
		   {{- $url := "http://" -}}
		   {{- post http | withoption "url" $url | withoption "name" "vol1" | runas "step1" $runner -}}
		   {{- post http | withoption "url" $url | withoption "name" "vol2" | runas "step2" $runner -}}
 		   {{- post http | withoption "url" $url | withoption "body" "vol3" | runas "step3" $runner -}}
 		   {{- post http | withoption "url" $url | withoption "body" "vol4" | runas "step4" $runner -}}
 		   {{- post http | withoption "url" $url | withoption "name" "vol4" | runas "step5" $runner -}}`,
			5, true, true, true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			mockval := map[string]interface{}{"Values": map[string]interface{}{}}
			tpl := template.New(mock.template).Funcs(AllCustomFuncs())
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mockval)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			op := buf.String()
			if len(op) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}

			tvalues := mockval["Values"]
			if tvalues == nil {
				t.Fatalf("Test '%s' failed: nil template values post template execution", name)
			}

			results := tvalues.(map[string]interface{})
			mockResultCount := mock.resultCount

			if results["rootCause"] != nil {
				mockResultCount = mockResultCount + 1
			}

			if len(results) != mockResultCount {
				t.Fatalf("Test '%s' failed: expected run command results count '%d': actual '%d'", name, mock.resultCount, len(results))
			}

			response := results[mock.commandToTest]
			if response == nil {
				t.Fatalf("Test '%s' failed: nil runtask command response post template execution", name)
			}

			cmdres := response.(map[string]interface{})
			err = cmdres["error"].(error)
			if mock.isErr && err == nil {
				t.Fatalf("Test '%s' failed: expected not nil err: actual %#v", name, cmdres)
			}

			debug := cmdres["debug"].(msg.AllMsgs)
			if mock.isErr && len(debug[msg.ErrMsg].Items) == 0 {
				t.Fatalf("Test '%s' failed: expected error messages: actual %s", name, debug)
			}
			if mock.isSkip && len(debug[msg.SkipMsg].Items) == 0 {
				t.Fatalf("Test '%s' failed: expected skip messages: actual %s", name, debug)
			}
			if mock.isInfo && len(debug[msg.InfoMsg].Items) == 0 {
				t.Fatalf("Test '%s' failed: expected info messages: actual %s", name, debug)
			}
		})
	}
}

func TestPutHttp(t *testing.T) {
	tests := map[string]struct {
		commandToTest string
		template      string
		resultCount   int
		isErr         bool
		isSkip        bool
		isInfo        bool
	}{
		// test scenario with one put http request
		"t101": {"step1",
			`{{- $store :=  storeAt .Values -}}
       {{- $runner := storeRunner $store -}}
		   {{- $url := "http://" -}}
		   {{- put http | withoption "url" $url | withoption "name" "myvol" | runas "step1" $runner -}}`,
			1, true, false, true},
		// test scenario with more than one put http requests
		"t102": {"step2",
			`{{- $store :=  storeAt .Values -}}
       {{- $runner := storeRunner $store -}}
		   {{- $url := "http://" -}}
		   {{- put http | withoption "url" $url | withoption "name" "myvol1" | runas "step1" $runner -}}
		   {{- put http | withoption "url" $url | withoption "name" "myvol2" | runas "step2" $runner -}}`,
			2, true, true, true},
		// test scenario with a series of put http requests
		"t103": {"step3",
			`{{- $store :=  storeAt .Values -}}
       {{- $runner := storeRunner $store -}}
		   {{- $url := "http://" -}}
		   {{- put http | withoption "url" $url | withoption "name" "vol1" | runas "step1" $runner -}}
		   {{- put http | withoption "url" $url | withoption "name" "vol2" | runas "step2" $runner -}}
 		   {{- put http | withoption "url" $url | withoption "name" "vol3" | runas "step3" $runner -}}
 		   {{- put http | withoption "url" $url | withoption "name" "vol4" | runas "step4" $runner -}}
 		   {{- put http | withoption "url" $url | withoption "name" "vol4" | runas "step5" $runner -}}`,
			5, true, true, true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			mockval := map[string]interface{}{"Values": map[string]interface{}{}}
			tpl := template.New(mock.template).Funcs(AllCustomFuncs())
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mockval)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			op := buf.String()
			if len(op) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}

			tvalues := mockval["Values"]
			if tvalues == nil {
				t.Fatalf("Test '%s' failed: nil template values post template execution", name)
			}

			results := tvalues.(map[string]interface{})
			mockResultCount := mock.resultCount

			if results["rootCause"] != nil {
				mockResultCount = mockResultCount + 1
			}

			if len(results) != mockResultCount {
				t.Fatalf("Test '%s' failed: expected run command results count '%d': actual '%d'", name, mock.resultCount, len(results))
			}

			response := results[mock.commandToTest]
			if response == nil {
				t.Fatalf("Test '%s' failed: nil runtask command response post template execution", name)
			}

			cmdres := response.(map[string]interface{})
			err = cmdres["error"].(error)
			if mock.isErr && err == nil {
				t.Fatalf("Test '%s' failed: expected not nil err: actual %#v", name, cmdres)
			}

			debug := cmdres["debug"].(msg.AllMsgs)
			if mock.isErr && len(debug[msg.ErrMsg].Items) == 0 {
				t.Fatalf("Test '%s' failed: expected error messages: actual %s", name, debug)
			}
			if mock.isSkip && len(debug[msg.SkipMsg].Items) == 0 {
				t.Fatalf("Test '%s' failed: expected skip messages: actual %s", name, debug)
			}
			if mock.isInfo && len(debug[msg.InfoMsg].Items) == 0 {
				t.Fatalf("Test '%s' failed: expected info messages: actual %s", name, debug)
			}
		})
	}
}

func TestCreateCstorSnapshotSaveAsVerifyError(t *testing.T) {
	tests := map[string]struct {
		template string
	}{
		// NOTE:
		//  Name of the test case should equal to the key used by saveas & saveif
		"t101": {`{{- $ip := "0.0.0.0" -}}
					  {{- $volName := "vol1" -}}
					  {{- $snapName := "s1" -}}
					  {{- $runCmd := create cstor snapshot | withoption "ip" $ip | withoption "volname" $volName -}}
					  {{- $runCmd | withoption "snapname" $snapName | run | saveas "t101" .Values -}}
					  {{- $err := .Values.t101.error | default "" | toString -}}
					  {{- $err | empty | not | verifyErr $err | saveIf "t101.verifyerr" .Values | noop -}}`,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			mockval := map[string]interface{}{"Values": map[string]interface{}{}}
			tpl := template.New(mock.template).Funcs(AllCustomFuncs())
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mockval)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			op := buf.String()
			if len(op) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}

			tvalues := mockval["Values"]
			if tvalues == nil {
				t.Fatalf("Test '%s' failed: nil template values post template execution", name)
			}

			response := tvalues.(map[string]interface{})[name]
			if response == nil {
				t.Fatalf("Test '%s' failed: nil runtask command response post template execution", name)
			}

			cmdres := response.(map[string]interface{})
			verifyerr := cmdres["verifyerr"]
			if verifyerr == nil {
				t.Fatalf("Test '%s' failed: expected not nil verifyerr: actual verifyerr %#v", name, verifyerr)
			}

			var verr error
			verr, ok := verifyerr.(*VerifyError)
			if !ok {
				t.Fatalf("Test '%s' failed: expected VerifyErr error: actual %#v", name, verr)
			}
		})
	}
}
