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

// This file provides various examples w.r.t go's
// text/template package. This is done to help contributors
// into understanding the standard templating features used
// by maya
//
// Guides on go templating:
//
// Built-in funcs:
// - https://golang.org/src/text/template/funcs.go
//
// Custom template funcs:
// - https://github.com/Masterminds/sprig/tree/HEAD/docs
//
// Templating guides:
// - https://github.com/kubernetes/helm/tree/HEAD/docs/chart_template_guide
// - https://docs.ansible.com/ansible/latest/reference_appendices/YAMLSyntax.html
package templatefuncs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"text/template"

	"github.com/ghodss/yaml"
)

func TestAddTo(t *testing.T) {
	tests := map[string]struct {
		fields              string
		destination         map[string]interface{}
		value               string
		expectedValue       string
		expectedDestination map[string]interface{}
	}{
		//
		// start of test case
		//
		"101": {
			fields:        "k1.k2",
			destination:   map[string]interface{}{},
			value:         "hi",
			expectedValue: "hi",
			expectedDestination: map[string]interface{}{
				"k1": map[string]interface{}{
					"k2": "hi",
				},
			},
		},
		//
		// start of test case
		//
		"102": {
			fields: "k1.k2",
			destination: map[string]interface{}{
				"k1": map[string]interface{}{
					"k2": "hi",
				},
			},
			value:         "hello",
			expectedValue: "hello",
			expectedDestination: map[string]interface{}{
				"k1": map[string]interface{}{
					"k2": "hi, hello",
				},
			},
		},
		//
		// start of test case
		//
		"103": {
			fields: "k1.k2",
			destination: map[string]interface{}{
				"k1": map[string]interface{}{
					"k2": "hi",
				},
			},
			value:         "",
			expectedValue: "",
			expectedDestination: map[string]interface{}{
				"k1": map[string]interface{}{
					"k2": "hi",
				},
			},
		},
		//
		// start of test case
		//
		"104": {
			fields:              "k1.k2",
			destination:         map[string]interface{}{},
			value:               "",
			expectedValue:       "",
			expectedDestination: map[string]interface{}{},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			actual := addTo(mock.fields, mock.destination, mock.value)

			if actual != mock.expectedValue {
				t.Fatalf("addTo test failed: expected '%s' actual '%s'", mock.expectedValue, actual)
			}

			if !reflect.DeepEqual(mock.expectedDestination, mock.destination) {
				t.Fatalf("addTo test failed: expected destination '%#v' actual destination '%#v'", mock.expectedDestination, mock.destination)
			}
		})
	}
}

func TestToYaml(t *testing.T) {
	tests := map[string]struct {
		data           interface{}
		expected       string
		isErr          bool
		expectedErrMsg string
	}{
		//
		// start of test case
		//
		"101": {
			data: map[string]string{
				"co":      "k8s",
				"storage": "openebs",
			},
			expected: fmt.Sprintf("co: k8s \nstorage: openebs\n"),
		},
		//
		// start of test case
		//
		"102": {
			data: map[string]interface{}{
				"co": "k8s",
				"storage": map[string]string{
					"gen1": "jiva",
					"gen2": "cstor",
				},
			},
			expected: `
co: k8s
storage:
  gen1: jiva
  gen2: cstor`,
		},
		//
		// start of test case
		//
		"103": {
			data:           []interface{}{"co", "k8s", "storage", "gen1", "jiva", "gen2", "cstor"},
			isErr:          true,
			expectedErrMsg: "error unmarshaling JSON: json: cannot unmarshal array into Go value of type map[string]interface {}",
		},
		//
		// start of test case
		//
		"104": {
			data: nil,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			actual := fromYaml(ToYaml(mock.data))
			expected := fromYaml(mock.expected)

			if mock.isErr {
				actualErrMsg, _ := actual["Error"].(string)
				if mock.expectedErrMsg != actualErrMsg {
					t.Fatalf("toYaml test failed: expected error '%s' actual error '%s'", mock.expectedErrMsg, actualErrMsg)
				}
			}

			if !mock.isErr && !reflect.DeepEqual(expected, actual) {
				t.Fatalf("toYaml test failed: expected '%#v' actual '%#v'", expected, actual)
			}
		})
	}
}

func TestNestedKeyMap(t *testing.T) {
	tests := map[string]struct {
		delimiters  string
		given       []string
		destination map[string]interface{}
		expected    map[string]interface{}
	}{
		//
		// start of test case
		//
		`101`: {
			delimiters:  "=",
			given:       []string{"co/co1=k8s", "co1=k8sNew", "co2=swarm", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"co/co1": "k8s",
				"co1":    "k8sNew",
				"co2":    "swarm",
				"co3":    "nomad",
			},
		},
		//
		// start of test case
		//
		`102`: {
			delimiters:  "=",
			given:       []string{"co/co1=k8s", "co2=swarm", "co2=swarm2", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"co/co1": "k8s",
				"co2":    "swarm, swarm2",
				"co3":    "nomad",
			},
		},
		//
		// start of test case
		//
		`103`: {
			delimiters:  "/ =",
			given:       []string{"co/co1=k8s", "co2=swarm", "co2=swarm2", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"co": map[string]interface{}{
					"co1": "k8s",
				},
				"co2": "swarm, swarm2",
				"co3": "nomad",
			},
		},
		//
		// start of test case
		//
		`104`: {
			delimiters:  "/ @ =",
			given:       []string{"cloud/co@co1=k8s", "co2=swarm", "co2=swarm2", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"cloud": map[string]interface{}{
					"co": map[string]interface{}{
						"co1": "k8s",
					},
				},
				"co2": "swarm, swarm2",
				"co3": "nomad",
			},
		},
		//
		// start of test case
		//
		`105`: {
			delimiters:  "/ / =",
			given:       []string{"cloud/co/co1=k8s", "co/co2=swarm", "co2=swarm2", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"cloud": map[string]interface{}{
					"co": map[string]interface{}{
						"co1": "k8s",
					},
				},
				"co": map[string]interface{}{
					"co2": "swarm",
				},
				"co2": "swarm2",
				"co3": "nomad",
			},
		},
		//
		// start of test case
		//
		`106`: {
			delimiters:  "=",
			given:       []string{" co/co1 =k8s", "co2 = swarm ", " co2= swarm2", "co3=nomad "},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"co/co1": "k8s",
				"co2":    "swarm, swarm2",
				"co3":    "nomad",
			},
		},
		//
		// start of test case
		//
		`107`: {
			delimiters:  "/ / =",
			given:       []string{"cloud/ /co1=k8s", "co/=swarm", "co=swarm2", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"cloud": map[string]interface{}{
					"co1": "k8s",
				},
				"co":  "swarm, swarm2",
				"co3": "nomad",
			},
		},
		//
		// start of test case
		//
		`108`: {
			delimiters:  "/ / =",
			given:       []string{"cloud/ /co1=k8s", "cloud/ /co1=", "co/=", "co=swarm2", "co3=nomad", "co4="},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"cloud": map[string]interface{}{
					"co1": "k8s",
				},
				"co":  "swarm2",
				"co3": "nomad",
				"co4": "",
			},
		},
		//
		// start of test case
		//
		`109`: {
			delimiters:  "/ / =",
			given:       []string{"cloud/ /co1=k8s", "cloud/ /co1=  ", "  =  ", " co3 = nomad ", "co4=  "},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"cloud": map[string]interface{}{
					"co1": "k8s",
				},
				"co3": "nomad",
				"co4": "",
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			result := nestedKeyMap(mock.delimiters, mock.destination, mock.given)

			if !reflect.DeepEqual(mock.expected, result) {
				t.Fatalf("test nested key map failed:\n\nexpected: '%#v' \n\nactual: '%#v'", mock.expected, result)
			}
		})
	}
}

func TestKeyMap(t *testing.T) {
	tests := map[string]struct {
		destinationFields string
		given             []string
		destination       map[string]interface{}
		expected          map[string]interface{}
	}{
		//
		// start of test case
		//
		`101`: {
			destinationFields: "test1",
			given:             []string{"co1=k8s,co2=swarm", "co3=nomad"},
			destination:       map[string]interface{}{},
			expected: map[string]interface{}{
				"test1": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`102`: {
			destinationFields: "test1",
			given:             []string{"co1=k8s,co2=swarm,pkey=one", "pkey=two,co3=nomad"},
			destination:       map[string]interface{}{},
			expected: map[string]interface{}{
				"test1": map[string]interface{}{
					"one": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
					},
					"two": map[string]interface{}{
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`103`: {
			destinationFields: "test2",
			given:             []string{"co1 = k8s,  co2= swarm  ", "   co3=nomad  "},
			destination:       map[string]interface{}{},
			expected: map[string]interface{}{
				"test2": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`104`: {
			destinationFields: "test3",
			given:             []string{"= k8s,  co2=  ", " = "},
			destination:       map[string]interface{}{},
			expected: map[string]interface{}{
				"test3": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co2": "",
					},
				},
			},
		},
		//
		// start of test case
		//
		`105`: {
			destinationFields: "test4",
			given:             []string{"pkey=,co1= k8s,  co2=,=nomad  ", " = "},
			destination:       map[string]interface{}{},
			expected: map[string]interface{}{
				"test4": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "",
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			result := keyMap(mock.destinationFields, mock.destination, mock.given)

			if !reflect.DeepEqual(mock.expected, result) {
				t.Fatalf("test key map failed:\n\nexpected: '%#v' \n\nactual: '%#v'", mock.expected, result)
			}
		})
	}
}

func TestSplitKeyMap(t *testing.T) {
	tests := map[string]struct {
		splitters   string
		destFields  string
		given       []string
		destination map[string]interface{}
		expected    map[string]interface{}
	}{
		//
		// start of test case
		//
		`101`: {
			splitters:   ":: =",
			destFields:  "test1",
			given:       []string{"co1=k8s::co2=swarm", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test1": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`102`: {
			splitters:   "-- =",
			destFields:  "test2",
			given:       []string{"co1=k8s--co2=swarm", "pkey=myCO--co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test2": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
					},
					"myCO": map[string]interface{}{
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`103`: {
			splitters:   "-- ::",
			destFields:  "test3",
			given:       []string{"co1::k8s--co2::swarm", "pkey::myCO--co3::nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test3": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
					},
					"myCO": map[string]interface{}{
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`104`: {
			splitters:   "",
			destFields:  "test4",
			given:       []string{"co1=k8s,co2=swarm", "pkey=myCO,co3=nomad", "co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test4": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
						"co4": "custom",
					},
					"myCO": map[string]interface{}{
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`105`: {
			splitters:   "",
			destFields:  "test5",
			given:       []string{"co1=k8s,co2=swarm", "co3=nomad", "co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test5": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
						"co3": "nomad",
						"co4": "custom",
					},
				},
			},
		},
		//
		// start of test case
		//
		`106`: {
			splitters:   "",
			destFields:  "test6",
			given:       []string{"co1=k8s,co2=swarm,co3=newK8s,co4=newSwarm", "co3=nomad", "co3=mySched,co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test6": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
						"co3": "newK8s, nomad, mySched",
						"co4": "newSwarm, custom",
					},
				},
			},
		},
		//
		// start of test case
		//
		`107`: {
			splitters:   "",
			destFields:  "test7",
			given:       []string{"pkey=alpha,co1=k8s,co2=swarm", "pkey=beta,co3=nomad", "pkey=delta,co3=mySched,co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test7": map[string]interface{}{
					"alpha": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
					},
					"beta": map[string]interface{}{
						"co3": "nomad",
					},
					"delta": map[string]interface{}{
						"co3": "mySched",
						"co4": "custom",
					},
				},
			},
		},
		//
		// start of test case
		//
		`108`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"pkey=alpha,co1=k8s,co2=swarm", "pkey=beta,co3=nomad", "pkey=delta,co3=mySched,co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"alpha": map[string]interface{}{
					"co1": "k8s",
					"co2": "swarm",
				},
				"beta": map[string]interface{}{
					"co3": "nomad",
				},
				"delta": map[string]interface{}{
					"co3": "mySched",
					"co4": "custom",
				},
			},
		},
		//
		// start of test case
		//
		`109`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"pkey=alpha,co1=k8s,co2=swarm", "co3=nomad", "co3=mySched,co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"alpha": map[string]interface{}{
					"co1": "k8s",
					"co2": "swarm",
				},
				"pkey": map[string]interface{}{
					"co3": "nomad, mySched",
					"co4": "custom",
				},
			},
		},
		//
		// start of test case
		//
		`110`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"co1=k8s,co2=swarm", "co3=nomad", "co3=mySched,co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"pkey": map[string]interface{}{
					"co1": "k8s",
					"co2": "swarm",
					"co3": "nomad, mySched",
					"co4": "custom",
				},
			},
		},
		//
		// start of test case
		//
		`111`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"co1=k8s,co2=swarm ,  co3=newK8s", "co3=nomad", "co3=mySched, co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"pkey": map[string]interface{}{
					"co1": "k8s",
					"co2": "swarm",
					"co3": "newK8s, nomad, mySched",
					"co4": "custom",
				},
			},
		},
		//
		// start of test case
		//
		`112`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"co1=,co2=swarm ,  co3= ", "co3=nomad", "co3=mySched, co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"pkey": map[string]interface{}{
					"co1": "",
					"co2": "swarm",
					"co3": "nomad, mySched",
					"co4": "custom",
				},
			},
		},
		//
		// start of test case
		//
		`113`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"co1=,=swarm ,  =newK8s ", "co3=nomad", "co3=mySched, co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"pkey": map[string]interface{}{
					"co1": "",
					"co3": "nomad, mySched",
					"co4": "custom",
				},
			},
		},
		//
		// start of test case
		//
		`114`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"", "  ", "  co3=mySched, co4=custom  "},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"pkey": map[string]interface{}{
					"co3": "mySched",
					"co4": "custom",
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			result := splitKeyMap(mock.splitters, mock.destFields, mock.destination, mock.given)

			if !reflect.DeepEqual(mock.expected, result) {
				t.Fatalf("test split key map failed:\n\nexpected: '%#v' \n\nactual: '%#v'", mock.expected, result)
			}
		})
	}
}

// MockJsonList is a struct used during unit testing that can be marshaled from
// go object to []byte & vice-versa
//
// NOTE:
//
//	The properties of this struct should be CapitalCased for marshal & un-marshal
//
// to work properly
type MockJsonList struct {
	ApiVersion string     `json:"apiVersion"`
	Items      []MockJson `json:"items"`
}

// MockJson is a struct used during unit testing that can be marshaled from go
// object to []byte & vice-versa
//
// NOTE:
//
//	The properties of this struct should be CapitalCased for marshal & un-marshal
//
// to work properly
type MockJson struct {
	Labels    map[string]string `json:"labels"`
	Owner     string            `json:"owner"`
	Kind      string            `json:"kind"`
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
}

func mockJsonMarshal(mock *MockJson) (op []byte) {
	op, err := json.Marshal(mock)
	if err != nil {
		fmt.Printf(err.Error())
	}
	return
}

func mockJsonListMarshal(mock *MockJsonList) (op []byte) {
	op, err := json.Marshal(mock)
	if err != nil {
		fmt.Printf(err.Error())
	}
	return
}

func TestTemplatingWithMutatingTemplateValues(t *testing.T) {
	tests := map[string]struct {
		templateInYaml         string
		templateValues         map[string]interface{}
		expectedYaml           string
		expectedTemplateValues map[string]interface{}
	}{
		//
		// start of test scenario
		//
		`101`: {
			templateInYaml: `
{{- "I am not empty" | empty | verifyErr "empty string provided" | noop -}}
`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml:           ``,
			expectedTemplateValues: map[string]interface{}{},
		},
		//
		// start of test scenario
		//
		`102`: {
			templateInYaml: `
{{- list "I am not empty" | empty | verifyErr "empty list provided" | noop -}}
`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml:           ``,
			expectedTemplateValues: map[string]interface{}{},
		},
		//
		// start of test scenario
		//
		`103`: {
			templateInYaml: `
{{- dict "k1" "v1" "k2" "v2" | empty | verifyErr "empty dict provided" | noop -}}
`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml:           ``,
			expectedTemplateValues: map[string]interface{}{},
		},
		//
		// start of test scenario
		//
		`104`: {
			templateInYaml: `
{{- list | empty | verifyErr "empty list provided" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &VerifyError{"empty list provided"},
			},
		},
		//
		// start of test scenario
		//
		`105`: {
			templateInYaml: `
{{- dict | empty | verifyErr "empty dictionary provided" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &VerifyError{"empty dictionary provided"},
			},
		},
		//
		// start of test scenario
		//
		`106`: {
			templateInYaml: `
{{- "" | empty | verifyErr "empty string provided" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &VerifyError{"empty string provided"},
			},
		},
		//
		// start of test scenario
		//
		`107`: {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.items[*].kind}" | trim | saveAs "TaskResult.tskId.kind" .Values | noop -}}
{{- .Values.TaskResult.tskId.kind | splitList " " | isLen 2 | not | verifyErr "kind count is not two" | saveIf "err" .Values | noop -}}
value: {{ .Values.TaskResult.tskId.kind }}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						{
							Kind: "Pod",
						},
						{
							Kind: "Deployment",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
value: Pod Deployment`,
			expectedTemplateValues: map[string]interface{}{
				"err": nil,
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "Pod Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`108`: {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.items[*].kind}" | trim | saveAs "TaskResult.tskId.kind" .Values | noop -}}
{{- .Values.TaskResult.tskId.kind | splitList " " | isLen 2 | not | verifyErr "kind count is not two" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						{
							Kind: "",
						},
						{
							Kind: "",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &VerifyError{"kind count is not two"},
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`109`: {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.items[*].kind}" | trim | saveAs "TaskResult.tskId.kind" .Values | noop -}}
{{- .Values.TaskResult.tskId.kind | splitList " " | isLen 2 | not | verifyErr "kind count is not two" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						{
							Kind: "",
						},
						{
							Kind: "Deployment",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &VerifyError{"kind count is not two"},
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`110`: {
			templateInYaml: `
{{- list | notFoundErr "empty list provided" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &NotFoundError{"empty list provided"},
			},
		},
		//
		// start of test scenario
		//
		`111`: {
			templateInYaml: `
{{- dict | notFoundErr "empty dictionary provided" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &NotFoundError{"empty dictionary provided"},
			},
		},
		//
		// start of test scenario
		//
		`112`: {
			templateInYaml: `
{{- "" | notFoundErr "empty string provided" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &NotFoundError{"empty string provided"},
			},
		},
		//
		// start of test scenario
		//
		`113`: {
			templateInYaml: `
{{- "I am not empty" | notFoundErr "empty string provided" | saveIf "err" .Values | noop -}}
`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": nil,
			},
		},
		//
		// start of test scenario
		//
		`114`: {
			templateInYaml: `
{{- list "I am not empty" | notFoundErr "empty list provided" | saveIf "err" .Values | noop -}}
`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": nil,
			},
		},
		//
		// start of test scenario
		//
		`115`: {
			templateInYaml: `
{{- dict "k1" "v1" "k2" "v2" | notFoundErr "empty dict provided" | saveIf "err" .Values | noop -}}
`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": nil,
			},
		},
		//
		// start of test scenario
		//
		`116`: {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.kind}" | saveAs "TaskResult.tskId.kind" .Values | notFoundErr "kind is missing" | saveIf "err" .Values | noop -}}
`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonMarshal(&MockJson{
					Kind: "Pod",
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": nil,
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`117`: {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.kind}" | saveAs "TaskResult.tskId.kind" .Values | notFoundErr "kind is missing" | saveIf "err" .Values | noop -}}
`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonMarshal(&MockJson{
					Kind: "",
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &NotFoundError{"kind is missing"},
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`118`: {
			templateInYaml: `
{{- $kindArr := jsonpath .JsonDoc "{range .items[*]}{@.namespace}/{@.name}@kind={@.kind};{end}" | trim | splitList ";" -}}
{{- $kindArr | nestedKeyMap "/ @ =" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(nil),
				"Values":  map[string]interface{}{},
			},
			expectedYaml:           ``,
			expectedTemplateValues: map[string]interface{}{},
		},
		//
		// start of test scenario
		//
		`119`: {
			templateInYaml: `
{{- "default/mypod@app=jiva openebs/mypod@app=cstor" | splitList " " | nestedKeyMap "@ =" .Values | noop -}}
{{- "default/mypod@backend=true default/mypod@app=jiva2" | splitList " " | nestedKeyMap "@ =" .Values | noop -}}
{{- "litmus/mypod@backend=true" | splitList " " | nestedKeyMap "/ @ =" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"default/mypod": map[string]interface{}{
					"app":     "jiva, jiva2",
					"backend": "true",
				},
				"openebs/mypod": map[string]interface{}{
					"app": "cstor",
				},
				"litmus": map[string]interface{}{
					"mypod": map[string]interface{}{
						"backend": "true",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`120`: {
			templateInYaml: `
{{- $kindArr := jsonpath .JsonDoc "{range .items[*]}{@.namespace}/{@.name}@kind={@.kind};{end}" | trim | splitList ";" -}}
{{- $kindArr | nestedKeyMap "/ @ =" .Values | noop -}}
kindOne: {{ .Values.openebs.mypod.kind }}
kindTwo: {{ .Values.default.mydeploy.kind }}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						{
							Kind:      "Pod",
							Namespace: "openebs",
							Name:      "mypod",
						},
						{
							Kind:      "Deployment",
							Namespace: "default",
							Name:      "mydeploy",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
kindOne: Pod
kindTwo: Deployment`,
			expectedTemplateValues: map[string]interface{}{
				"openebs": map[string]interface{}{
					"mypod": map[string]interface{}{
						"kind": "Pod",
					},
				},
				"default": map[string]interface{}{
					"mydeploy": map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`121`: {
			templateInYaml: `
{{- $kindArr := jsonpath .JsonDoc "{range .items[*]}{@.namespace}/{@.name}@kind={@.kind};{end}" | trim | splitList ";" -}}
{{- $kindArr | nestedKeyMap "/ @ =" .Values | noop -}}
kindOne: {{ .Values.openebs.myapp.kind }}
kindTwo: {{ .Values.default.mydeploy.kind }}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						{
							Kind:      "Service",
							Namespace: "openebs",
							Name:      "myapp",
						},
						{
							Kind:      "Pod",
							Namespace: "openebs",
							Name:      "myapp",
						},
						{
							Kind:      "Deployment",
							Namespace: "default",
							Name:      "mydeploy",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
kindOne: Service, Pod
kindTwo: Deployment`,
			expectedTemplateValues: map[string]interface{}{
				"openebs": map[string]interface{}{
					"myapp": map[string]interface{}{
						"kind": "Service, Pod",
					},
				},
				"default": map[string]interface{}{
					"mydeploy": map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`122`: {
			templateInYaml: `
{{- $kindArr := jsonpath .JsonDoc "{range .items[*]}{@.namespace}/{@.name}@kind={@.kind};{end}" | trim | splitList ";" -}}
{{- $kindArr | nestedKeyMap "/ @ =" .Values | noop -}}
{{- .Values | toYaml -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						{
							Kind:      "Pod",
							Namespace: "openebs",
							Name:      "mypod",
						},
						{
							Kind:      "Deployment",
							Namespace: "default",
							Name:      "mydeploy",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
default:
  mydeploy:
    kind: Deployment
openebs:
  mypod:
    kind: Pod`,
			expectedTemplateValues: map[string]interface{}{
				"openebs": map[string]interface{}{
					"mypod": map[string]interface{}{
						"kind": "Pod",
					},
				},
				"default": map[string]interface{}{
					"mydeploy": map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`123`: {
			templateInYaml: `
{{- "pkey=openebs,stor1=jiva,stor2=cstor" | splitList " " | keyMap "scenario" .Values | noop -}}
{{- "co1=swarm,co2=k8s" | splitList " " | keyMap "scenario" .Values | noop -}}
{{- "pkey=openebs,stor2=mstor" | splitList " " | keyMap "scenario" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"scenario": map[string]interface{}{
					"openebs": map[string]interface{}{
						"stor1": "jiva",
						"stor2": "cstor, mstor",
					},
					"pkey": map[string]interface{}{
						"co1": "swarm",
						"co2": "k8s",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`124`: {
			templateInYaml: `
{{- $all := jsonpath .JsonDoc .Values.path | trim | splitList ";" -}}
{{- $all | keyMap "scenario" .Values | noop -}}
kind: MyList
apiVersion: v1alpha1
items:
# NOTE:
#   Below range and if blocks should end with }}
# If they end with -}} a new line is not formed and makes the yaml invalid
{{- range $pkey, $val := .Values.scenario }}
  - label: {{ $pkey }}
    name: {{ pluck "name" $val | first }}
    kind: {{ pluck "kind" $val | first }}
    owner: {{ pluck "owner" $val | first }}
    count: {{ pluck "kind" $val | first | default "" | splitList ", " | len }}
# NOTE:
#   Below end statements can end with -}} as there are no more yaml items
{{- end -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						{
							Labels: map[string]string{
								"openebs.io/pv":                 "pvc-1234-abc",
								"openebs.io/controller-service": "jiva-controller-svc",
							},
							Owner: "Admin",
							Kind:  "Service",
							Name:  "myservice",
						},
						{
							Labels: map[string]string{
								"openebs.io/pv":         "pvc-1234-abc",
								"openebs.io/controller": "jiva-controller",
							},
							Owner: "User",
							Kind:  "Pod",
							Name:  "mypod",
						},
						{
							Labels: map[string]string{
								"openebs.io/pv":         "pvc-1234-abc-def",
								"openebs.io/controller": "jiva-controller-deploy",
							},
							Owner: "User",
							Kind:  "Deployment",
							Name:  "mydeploy",
						},
					},
				}),
				"Values": map[string]interface{}{
					"path": `{range .items[*]}pkey={@.labels.openebs\.io/pv},name={@.name},kind={@.kind},owner={@.owner};{end}`,
				},
			},
			expectedYaml: `
kind: MyList
apiVersion: v1alpha1
items:
  - label: pvc-1234-abc
    name: myservice, mypod
    kind: Service, Pod
    owner: Admin, User
    count: 2
  - label: pvc-1234-abc-def
    name: mydeploy
    kind: Deployment
    owner: User
    count: 1`,
			expectedTemplateValues: map[string]interface{}{
				"path": `{range .items[*]}pkey={@.labels.openebs\.io/pv},name={@.name},kind={@.kind},owner={@.owner};{end}`,
				"scenario": map[string]interface{}{
					"pvc-1234-abc": map[string]interface{}{
						"kind":  "Service, Pod",
						"owner": "Admin, User",
						"name":  "myservice, mypod",
					},
					"pvc-1234-abc-def": map[string]interface{}{
						"owner": "User",
						"name":  "mydeploy",
						"kind":  "Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`125`: {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.apiVersion}" | saveAs "TaskResult.tskId.apiVersion" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"apiVersion": "v1beta1",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`126`: {
			templateInYaml: `
{{- $var := .Values.placeholder -}}
{{- jsonpath .JsonDoc "{.apiVersion}" | saveAs "TaskResult.tskId.myplace" .Values | noop -}}
{{- range $k, $v := .Values.TaskResult.tskId -}}
{{- if eq $k $var -}}
apiVersion: {{ $v }}
{{- end -}}
{{- end -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
				}),
				"Values": map[string]interface{}{
					"placeholder": "myplace",
				},
			},
			expectedYaml: `
apiVersion: v1beta1`,
			expectedTemplateValues: map[string]interface{}{
				"placeholder": "myplace",
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"myplace": "v1beta1",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		"127": {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.items[*].kind}" | trim | saveAs "TaskResult.tskId.kind" .Values | noop -}}
{{- .Values.TaskResult.tskId.kind | empty | saveAs "TaskResult.tskId.iskindempty" .Values | noop -}}
value: {{ .Values.TaskResult.tskId.kind }}
bool: {{ .Values.TaskResult.tskId.iskindempty }}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						{
							Kind:      "",
							Namespace: "openebs",
							Name:      "mypod",
						},
						{
							Kind:      "",
							Namespace: "openebs",
							Name:      "yourpod",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
value:
bool: true`,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind":        "",
						"iskindempty": true,
					},
				},
			},
		},
		//
		// start of test scenario
		//
		"128": {
			templateInYaml: `
{{- $noop := jsonpath .JsonDoc "{.items[*].kind}" | splitList " " | isLen 0 | saveAs "TaskResult.tskId.iskindlenzero" .Values -}}
show: {{ .Values.TaskResult.tskId.iskindlenzero }}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						{
							Kind:      "",
							Namespace: "openebs",
							Name:      "mypod",
						},
						{
							Kind:      "",
							Namespace: "openebs",
							Name:      "yourpod",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `show: false`,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"iskindlenzero": false,
					},
				},
			},
		},
		//
		// start of test scenario
		//
		"129": {
			templateInYaml: `
{{- $noop := jsonpath .JsonDoc "{.items[*].kind}" | splitList " " | saveAs "TaskResult.tskId.kinds" .Values -}}
show: {{ .Values.TaskResult.tskId.kinds }}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						{
							Kind:      "",
							Namespace: "openebs",
							Name:      "mypod",
						},
						{
							Kind:      "",
							Namespace: "openebs",
							Name:      "yourpod",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `show: [ ]`,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kinds": []string{"", ""},
					},
				},
			},
		},
		//
		// start of test scenario
		//
		"130": {
			templateInYaml: `
apiVersion: {{ jsonpath .JsonDoc "{.apiVersion}" | saveAs "TaskResult.tskId.apiVersion" .Values }}
kinds: {{ jsonpath .JsonDoc "{.items[*].kind}" | saveAs "TaskResult.tskId.kinds" .Values }}
count: {{ .Values.TaskResult.tskId.kinds | splitList " " | len }}
`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						{
							Kind:      "Pod",
							Namespace: "openebs",
							Name:      "mypod",
						},
						{
							Kind:      "Pod",
							Namespace: "openebs",
							Name:      "yourpod",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
apiVersion: v1beta1
kinds: Pod Pod
count: 2
`,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"apiVersion": "v1beta1",
						"kinds":      "Pod Pod",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		"131": {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.kind}" | saveAs "TaskResult.tskId.kind" .Values | noop -}}
{{- saveAs "TaskResult.tskId.kind" .Values "Deployment" | noop -}}
kind: {{ .Values.TaskResult.tskId.kind }}
`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonMarshal(&MockJson{
					Kind:      "Pod",
					Namespace: "openebs",
					Name:      "",
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
kind: Deployment
`,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		"132": {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.kind}" | saveIf "TaskResult.tskId.kind" .Values | noop -}}
{{- saveIf "TaskResult.tskId.kind" .Values "Deployment" | noop -}}
kind: {{ .Values.TaskResult.tskId.kind }}
`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonMarshal(&MockJson{
					Kind:      "Pod",
					Namespace: "openebs",
					Name:      "",
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
kind: Pod
`,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// augment the standard templating with sprig template functions
			tpl := template.New("TestTemplatingWithMutatingTemplateValues").Funcs(AllCustomFuncs())
			tpl, err := tpl.Parse(mock.templateInYaml)
			if err != nil {
				t.Fatalf("test failed: expected 'no template parse error': actual '%s'", err.Error())
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mock.templateValues)
			if err != nil {
				t.Fatalf("test failed: expected 'no template execution error': actual '%s'", err.Error())
			}

			// buffer that represents a YAML can be unmarshalled into a map of any objects
			var objActual map[string]interface{}
			err = yaml.Unmarshal(buf.Bytes(), &objActual)
			if err != nil {
				t.Fatalf("test failed: expected 'no error on un-marshalling templated bytes to any objects': actual '%s'", err.Error())
			}

			// unmarshall the expected yaml into a map of any objects
			var objExpected map[string]interface{}
			err = yaml.Unmarshal([]byte(mock.expectedYaml), &objExpected)
			if err != nil {
				t.Fatalf("test failed: expected 'no error on un-marshalling expected yaml bytes to any objects': actual '%s'", err.Error())
			}

			// compare expected vs. actual object
			ok := reflect.DeepEqual(objExpected, objActual)
			if !ok {
				t.Fatalf("test failed:\n\nexpected yaml: '%s' \n\nactual yaml: '%s'", mock.expectedYaml, buf.Bytes())
			}

			// compare the values as values can get modified at runtime
			if !reflect.DeepEqual(mock.expectedTemplateValues, mock.templateValues["Values"]) {
				t.Fatalf("test failed:\n\nexpected template values: '%#v' \n\nactual template values: '%#v'", mock.expectedTemplateValues, mock.templateValues["Values"])
			}
		})
	}
}

func TestDynamicTemplating(t *testing.T) {
	tests := map[string]struct {
		ymlTpl      string
		values      map[string]interface{}
		ymlExpected string
	}{
		//
		// start of test scenario
		//
		"101": {
			ymlTpl: `
{{- $ns := .defaultNamespace -}}
version: v1
kind: {{ .kind }}
label: {{ pluck $ns .labels | first }}
`,
			values: map[string]interface{}{
				"defaultNamespace": "openebs",
				"kind":             "pod",
				"labels": map[string]interface{}{
					"openebs": "cas-volume",
					"local":   "host-volume",
				},
			},
			ymlExpected: `
version: v1
kind: pod
label: cas-volume
`,
		},
		//
		// start of test scenario
		//
		"102": {
			ymlTpl: `
{{- $ns := .defaultNamespace -}}
{{- $nsLbl := pluck $ns .labels -}}
version: v1
kind: {{ .kind }}
label: {{ first $nsLbl }}
`,
			values: map[string]interface{}{
				"defaultNamespace": "openebs",
				"kind":             "pod",
				"labels": map[string]interface{}{
					"openebs": "cas-volume",
					"local":   "host-volume",
				},
			},
			ymlExpected: `
version: v1
kind: pod
label: cas-volume
`,
		},
		//
		// start of test scenario
		//
		"103": {
			ymlTpl: `
{{- $ns := .runNamespace -}}
{{- $results := pick .taskResult.taskid $ns -}}
version: v1
kind: {{ .kind }}
{{- range $k, $v := $results }}
{{- range $kk, $vv := $v }}
objectName: {{ if eq $kk "objectName" }}{{ $vv }}{{ end }}
{{- end }}
{{- end }}`,
			values: map[string]interface{}{
				"runNamespace": "default-ns",
				"kind":         "pod",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							"hostName":   "lenovo-laptop",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"hostName":   "k8s-minion-1",
						},
					},
				},
			},
			ymlExpected: `version: v1
kind: pod
objectName: my-replica-pod`,
		},
		//
		// start of test scenario
		//
		"104": {
			ymlTpl: `
{{- $ns := .runNamespace -}}
{{- $nsResults := pick .taskResult.taskid $ns -}}
version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      {{- range $k, $v := $nsResults }}
      {{- range $kk, $vv := $v }}
      {{- if eq $kk "nodeNames" }}
      {{- $nodeNames := $vv }}
      {{- if ne $nodeNames "" }}
      {{- $nodeNamesMap := $nodeNames | split " " }}
      {{- range $kkk, $vvv := $nodeNamesMap }}
      - {{ $vvv }}
      {{- end }}
      {{- end }}
      {{- end }}
      {{- end }}
      {{- end }}
`,
			values: map[string]interface{}{
				"runNamespace": "openebs-ns",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							"nodeNames":  "lenovo-laptop",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"nodeNames":  "k8s-minion-1 lenovo-laptop hp-laptop",
						},
					},
				},
			},
			ymlExpected: `version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      - k8s-minion-1
      - lenovo-laptop
      - hp-laptop`,
		},
		//
		// start of test scenario
		//
		"105": {
			ymlTpl: `{{- $ns := .runNamespace -}}
{{- $nsResults := pick .taskResult.taskid $ns -}}
{{- range $k, $v := $nsResults }}
{{- range $kk, $vv := $v }}
{{- if eq $kk "nodeNames" }}
version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      {{- if ne $vv "" }}
      - {{ splitList " " $vv | first }}
      {{- end }}
{{- end }}
{{- end }}
{{- end }}`,
			values: map[string]interface{}{
				"runNamespace": "openebs-ns",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							"nodeNames":  "lenovo-laptop",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"nodeNames":  "k8s-minion-1 lenovo-laptop hp-laptop",
						},
					},
				},
			},
			ymlExpected: `version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      - k8s-minion-1`,
		},
		//
		// start of test scenario
		//
		"106": {
			ymlTpl: `{{- $ns := .runNamespace -}}
{{- $nsResults := pick .taskResult.taskid $ns -}}
{{- range $k, $v := $nsResults }}
{{- range $kk, $vv := $v }}
{{- if eq $kk "nodeNames" }}
version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      {{- if ne $vv "" }}
      - {{ splitList " " $vv | first }}
      {{- end }}
{{- end }}
{{- end }}
{{- end }}`,
			values: map[string]interface{}{
				"runNamespace": "default-ns",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							// this is single value nodeNames
							"nodeNames": "lenovo-laptop",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"nodeNames":  "k8s-minion-1 lenovo-laptop hp-laptop",
						},
					},
				},
			},
			ymlExpected: `version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      - lenovo-laptop`,
		},
		//
		// start of test scenario
		//
		"107": {
			ymlTpl: `{{- $ns := .runNamespace -}}
{{- $nsResults := pick .taskResult.taskid $ns -}}
{{- range $k, $v := $nsResults }}
{{- range $kk, $vv := $v }}
{{- if eq $kk "nodeNames" }}
version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      {{- if ne $vv "" }}
      - {{ splitList " " $vv | first }}
      {{- end }}
{{- end }}
{{- end }}
{{- end }}`,
			values: map[string]interface{}{
				"runNamespace": "default-ns",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							// this is single value with dangling space
							"nodeNames": "lenovo-laptop ",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"nodeNames":  "k8s-minion-1 lenovo-laptop hp-laptop",
						},
					},
				},
			},
			ymlExpected: `version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      - lenovo-laptop`,
		},
		//
		// start of test scenario
		//
		"108": {
			ymlTpl: `{{- $ns := .runNamespace -}}
{{- $nsResults := pick .taskResult.taskid $ns -}}
{{- range $k, $v := $nsResults }}
{{- range $kk, $vv := $v }}
{{- if eq $kk "nodeNames" }}
version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeCount: {{ if ne $vv "" }}{{ splitList " " $vv | len }}{{ end }}
{{- end }}
{{- end }}
{{- end }}`,
			values: map[string]interface{}{
				"runNamespace": "openebs-ns",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							"nodeNames":  "lenovo-laptop ",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"nodeNames":  "k8s-minion-1 lenovo-laptop hp-laptop",
						},
					},
				},
			},
			ymlExpected: `version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeCount: 3`,
		},
		//
		// start of test scenario
		//
		"109": {
			ymlTpl: `{{- $ns := .runNamespace -}}
{{- $nsResults := pick .taskResult.taskid $ns -}}
{{- range $k, $v := $nsResults }}
{{- range $kk, $vv := $v }}
{{- if eq $kk "nodeNames" }}
version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodes: {{ if ne $vv "" }}{{ splitList " " $vv | join "," }}{{ end }}
{{- end }}
{{- end }}
{{- end }}`,
			values: map[string]interface{}{
				"runNamespace": "openebs-ns",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							"nodeNames":  "lenovo-laptop ",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"nodeNames":  "k8s-minion-1 lenovo-laptop hp-laptop",
						},
					},
				},
			},
			ymlExpected: `version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodes: k8s-minion-1,lenovo-laptop,hp-laptop`,
		},
		//
		// start of test scenario
		//
		"110": {
			ymlTpl: `{{- $count := "3" -}}
all:
{{- range $i, $e := until ($count|int) }}
- pvc-{{$e}}
{{- end -}}
`,
			ymlExpected: `all:
- pvc-0
- pvc-1
- pvc-2
`,
		},
		//
		// start of test scenario
		//
		"111": {
			ymlTpl: `{{- $count := "3" -}}
all:
{{- range $i, $e := untilStep 1 ($count|int) 1 }}
- pvc-{{$e}}
{{- end -}}
`,
			ymlExpected: `all:
- pvc-1
- pvc-2
`,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// augment the standard templating with sprig template functions
			tpl := template.New("testdynamicvartemplating").Funcs(AllCustomFuncs())
			tpl, err := tpl.Parse(mock.ymlTpl)
			if err != nil {
				t.Fatalf("failed to test dynamic templating: expected 'no instantiation error': actual '%s'", err.Error())
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mock.values)
			if err != nil {
				t.Fatalf("failed to test dynamic templating: expected 'no execution error': actual '%s'", err.Error())
			}

			// buffer that represents a YAML can be unmarshalled into a map of any objects
			var objActual map[string]interface{}
			err = yaml.Unmarshal(buf.Bytes(), &objActual)
			if err != nil {
				t.Fatalf("failed to test dynamic templating: expected 'no error w.r.t unmarshalling bytes to any objects': actual '%s'", err.Error())
			}

			// unmarshall the expected yaml into a map of any objects
			var objExpected map[string]interface{}
			err = yaml.Unmarshal([]byte(mock.ymlExpected), &objExpected)
			if err != nil {
				t.Fatalf("failed to test dynamic templating: expected 'no error w.r.t unmarshalling expected yaml to any objects': actual '%s'", err.Error())
			}

			// compare expected vs. actual object
			ok := reflect.DeepEqual(objExpected, objActual)
			if !ok {
				t.Fatalf("failed to test dynamic templating:\n\nexpected: '%s' \n\nactual: '%s'", mock.ymlExpected, buf.Bytes())
			}
		})
	}
}

func TestRandomize(t *testing.T) {
	tests := map[string]struct {
		list []string
		want []string
	}{
		"SingleKey": {
			list: []string{
				"key1",
			},
			want: []string{"key1"},
		},
		"TwoKeys": {
			list: []string{
				"key1", "key2",
			},
			want: []string{"key1", "key2"},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			list := randomize(mock.list)
			if len(list) != len(mock.want) {
				t.Fatalf("failed to test randomize: expected '%v' %d: actual '%v' %d",
					mock.want, len(mock.want), list, len(list))
			}
		})
	}
}

func TestSplitListTrim(t *testing.T) {
	tests := map[string]struct {
		sep  string
		orig string
		want []string
	}{
		"No separator in string": {
			";",
			"pool1pool2",
			[]string{"pool1pool2"},
		},
		"Prefixed separator in string": {
			";",
			";pool1;pool2",
			[]string{"pool1", "pool2"},
		},
		"Suffixed separator in string": {
			";",
			"pool1;pool2;",
			[]string{"pool1", "pool2"},
		},
		"Multiple separators only": {
			";",
			";;;;",
			[]string{""},
		},
		"Multiple separators in between": {
			";",
			"p1;;;p2",
			[]string{"p1", "", "", "p2"},
		},
		"Prefix-Suffix separators in string": {
			";",
			";;p1;p2;p3;;;",
			[]string{"p1", "p2", "p3"},
		},
		"Single separator in string": {
			";",
			";",
			[]string{""},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			list := splitListTrim(mock.sep, mock.orig)
			ok := reflect.DeepEqual(list, mock.want)
			if !ok {
				t.Fatalf("failed to test splitListTrim: expected '%v': actual '%v'", mock.want, list)
			}
		})
	}
}
