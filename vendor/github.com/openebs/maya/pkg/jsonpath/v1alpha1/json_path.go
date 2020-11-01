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
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/openebs/maya/pkg/msg/v1alpha1"
	ft "k8s.io/client-go/third_party/forked/golang/template"
	jp "k8s.io/client-go/util/jsonpath"
	"reflect"
)

// selection is a utility struct that wraps over json path and values obtained
// after querying this path
type selection struct {
	Alias  string   `json:"alias"`  // alias name tagged against the json path
	Path   string   `json:"path"`   // json path against which jsonpath query will get executed
	Values []string `json:"values"` // resulting values due to after querying
	*Msgs
}

// Selection returns a new instance of selection
func Selection(alias, path string) *selection {
	return &selection{
		Alias: alias,
		Path:  path,
		Msgs:  &Msgs{},
	}
}

// SetValues populates the selection instance with values; typically obtained
// after querying the select path
func (s *selection) SetValues(rvals [][]reflect.Value) {
	if len(rvals) == 0 {
		s.AddWarn(fmt.Sprintf("no value(s) found for %s %s", s.Alias, s.Path))
		s.Values = append(s.Values, "<no value>")
		return
	}
	for _, rvs := range rvals {
		for _, rv := range rvs {
			pv, ok := ft.PrintableValue(rv)
			if !ok {
				s.AddWarn(fmt.Sprintf("can not print type %s: failed to query %s %s", rv.Type(), s.Alias, s.Path))
				pv = "<not printable>"
			}
			var buffer bytes.Buffer
			fmt.Fprint(&buffer, pv)
			s.Values = append(s.Values, buffer.String())
		}
	}
	return
}

// Value returns the first value that was obtained after querying the select
// path
func (s *selection) Value() (val string) {
	if len(s.Values) == 0 {
		return
	}
	val = s.Values[0]
	return
}

// selectionList represents a list of selection instances
type selectionList []*selection

// SelectionList returns a new list of selection instances based on given
// select aliases & paths
func SelectionList(aliasPaths map[string]string) (sl selectionList) {
	for alias, path := range aliasPaths {
		sl = append(sl, Selection(alias, path))
	}
	return
}

// String is an implementation of Stringer interface
func (l selectionList) String() string {
	return YamlString("selectionlist", l)
}

// Values maps the values (obtained after querying the path) against the path's
// alias
func (l selectionList) Values() (vals map[string]interface{}) {
	if len(l) == 0 {
		return
	}
	vals = map[string]interface{}{}
	var v interface{}
	for _, s := range l {
		v = nil
		if len(s.Values) > 1 {
			v = s.Values
		} else if len(s.Values) == 1 {
			v = s.Values[0]
		}
		vals[s.Alias] = v
	}
	return
}

// ValuesByAlias returns the values (obtained after querying the path) of
// corresponding path's alias
func (l selectionList) ValuesByAlias(alias string) (vals []string) {
	for _, s := range l {
		if s.Alias == alias {
			return s.Values
		}
	}
	return
}

// ValueByAlias returns the first value / only value (obtained after querying
// the path) of corresponding path's alias
func (l selectionList) ValueByAlias(alias string) (value string) {
	vals := l.ValuesByAlias(alias)
	if len(vals) == 0 {
		return
	}
	return vals[0]
}

// ValuesByPath returns the values (obtained after querying the path) of
// corresponding path
func (l selectionList) ValuesByPath(path string) (vals []string) {
	for _, s := range l {
		if s.Path == path {
			return s.Values
		}
	}
	return
}

// ValueByPath returns the first value / only value (obtained after querying
// the path) of corresponding path
func (l selectionList) ValueByPath(path string) (value string) {
	vals := l.ValuesByPath(path)
	if len(vals) == 0 {
		return
	}
	return vals[0]
}

// jsonpath is a wrapper over jsonpath library
type jsonpath struct {
	name    string        // name given to the json querying
	target  interface{}   // target to be queried against
	jpath   *jp.JSONPath  // instance that understands querying json doc
	selects selectionList // selective queries to be done against json doc
	*Msgs
}

// JSONPath returns a new jsonpath instance
func JSONPath(name string) (j *jsonpath) {
	return &jsonpath{
		name:  name,
		jpath: jp.New(name).AllowMissingKeys(true),
		Msgs:  &Msgs{},
	}
}

// WithTarget sets the target to be queried against
func (j *jsonpath) WithTarget(target interface{}) (u *jsonpath) {
	j.target = target
	return j
}

// WithTarget sets the raw target to be queried against
func (j *jsonpath) WithTargetAsRaw(target []byte) (u *jsonpath) {
	var t interface{}
	err := json.Unmarshal(target, &t)
	if err != nil {
		j.AddError(fmt.Errorf("failed to build target for jsonpath %s: error - %s", j.name, err.Error()))
		return j
	}
	j.target = t
	return j
}

// Values executes jsonpath query by parsing the target against the provided
// path
func (j *jsonpath) Values(path string) (vals [][]reflect.Value, err error) {
	err = j.jpath.Parse(path)
	if err != nil {
		return
	}
	return j.jpath.FindResults(j.target)
}

// Query executes jsonpath query for given select path against the target
func (j *jsonpath) Query(s *selection) (u *selection) {
	vals, err := j.Values(s.Path)
	if err != nil {
		j.AddError(fmt.Errorf("failed to query %s %s: error - %s", s.Alias, s.Path, err.Error()))
	}
	s.SetValues(vals)
	j.Msgs.Merge(s.Msgs)
	return s
}

// Query executes jsonpath query for each given select path against the target
func (j *jsonpath) QueryAll(selects selectionList) (l selectionList) {
	for _, s := range selects {
		l = append(l, j.Query(s))
	}
	return
}
