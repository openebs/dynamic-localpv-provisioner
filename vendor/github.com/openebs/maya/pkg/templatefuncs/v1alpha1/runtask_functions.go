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
	"encoding/json"
	"strconv"
	"strings"
	"text/template"

	cmd "github.com/openebs/maya/pkg/task/v1alpha1"
	"github.com/openebs/maya/pkg/util"
)

// delete returns a new instance of delete based runtask command
//
// Examples:
// ---------
// {{- delete jiva volume | run -}}
// {{- delete cstor volume | run -}}
func delete(ml ...cmd.RunCommandMiddleware) *cmd.RunCommand {
	return cmd.RunCommandMiddlewareList(ml).Update(cmd.Command().DeleteAction())
}

// get returns a new instance of get based runtask command
//
// Examples:
// ---------
// {{- get jiva volume | run -}}
// {{- get cstor volume | run -}}
// {{- get http | url $url | run -}}
func get(ml ...cmd.RunCommandMiddleware) *cmd.RunCommand {
	return cmd.RunCommandMiddlewareList(ml).Update(cmd.Command().GetAction())
}

// list returns a new instance of list based runtask command
//
// Examples:
// ---------
// {{- list jiva volume | run -}}
// {{- list cstor volume | run -}}
func list(ml ...cmd.RunCommandMiddleware) *cmd.RunCommand {
	return cmd.RunCommandMiddlewareList(ml).Update(cmd.Command().ListAction())
}

// create returns a new instance of create based runtask command
//
// Examples:
// ---------
// {{- create jiva volume | run -}}
// {{- create cstor volume | run -}}
func create(ml ...cmd.RunCommandMiddleware) *cmd.RunCommand {
	return cmd.RunCommandMiddlewareList(ml).Update(cmd.Command().CreateAction())
}

// post returns a new instance of post based runtask command
//
// Examples:
// ---------
// {{- post http | url $url -}}
func post(ml ...cmd.RunCommandMiddleware) *cmd.RunCommand {
	return cmd.RunCommandMiddlewareList(ml).Update(cmd.Command().PostAction())
}

// put returns a new instance of put based runtask command
//
// Examples:
// ---------
// {{- put http | url $url -}}
func put(ml ...cmd.RunCommandMiddleware) *cmd.RunCommand {
	return cmd.RunCommandMiddlewareList(ml).Update(cmd.Command().PutAction())
}

// patch returns a new instance of patch based runtask command
//
// Examples:
// ---------
// {{- patch jiva volume | run -}}
// {{- patch cstor volume | run -}}
func patch(ml ...cmd.RunCommandMiddleware) *cmd.RunCommand {
	return cmd.RunCommandMiddlewareList(ml).Update(cmd.Command().PatchAction())
}

// update returns a new instance of update based runtask command
//
// Examples:
// ---------
// {{- update jiva volume | run -}}
// {{- update cstor volume | run -}}
func update(ml ...cmd.RunCommandMiddleware) *cmd.RunCommand {
	return cmd.RunCommandMiddlewareList(ml).Update(cmd.Command().UpdateAction())
}

// http returns a new instance of http based runtask command
//
// Examples:
// ---------
// {{- post http | url $url | run -}}
// {{- get http | url $url | run -}}
func http() cmd.RunCommandMiddleware {
	return cmd.HttpCategory()
}

// jiva returns a new instance of jiva based runtask command
//
// Examples:
// ---------
// {{- get jiva volume | run -}}
// {{- list jiva volume | run -}}
func jiva() cmd.RunCommandMiddleware {
	return cmd.JivaCategory()
}

// cstor returns a new instance of cstor based runtask command
//
// Examples:
// ---------
// {{- get cstor volume | run -}}
// {{- list cstor volume | run -}}
func cstor() cmd.RunCommandMiddleware {
	return cmd.CstorCategory()
}

// volume returns a new instance of volume based runtask command
//
// Examples:
// ---------
// {{- get cstor volume | run -}}
// {{- list cstor volume | run -}}
// {{- update jiva volume | run -}}
// {{- create jiva volume | run -}}
func volume() cmd.RunCommandMiddleware {
	return cmd.VolumeCategory()
}

// snapshot returns a new instance of snapshot based runtask command
//
// Examples:
// {{- create cstor snapshot | run -}}
func snapshot() cmd.RunCommandMiddleware {
	return cmd.SnapshotCategory()
}

// slect returns the values of the specified paths post runtask command
// execution
//
// Examples:
// ---------
// {{- select "all" | get cstor volume | run -}}
// {{- select "name" | list cstor volume | run -}}
// {{- select "name" "namespace" | update jiva volume | run -}}
// {{- select ".metadata.name" | create jiva volume | run -}}
func slect(paths ...string) cmd.RunCommandMiddleware {
	if len(paths) == 0 {
		paths = append(paths, "all")
	}
	return cmd.Select(paths)
}

// toJsonObj marshals and returns the json representation of value interface
//
// {{- "{'name':'openebs'}" | toJsonObj -}}
func toJsonObj(value interface{}) (b []byte) {
	b, _ = json.Marshal(value)
	return
}

// withOption sets the provided <key,value> pair as an input data to run command
//
// {{- delete jiva volume | withOption "url" $url | withOption "name" "myvol" | run -}}
func withOption(key string, value interface{}, given *cmd.RunCommand) (updated *cmd.RunCommand) {
	return cmd.WithData(given, key, value)
}

// run executes the runtask command
//
// Examples:
// ---------
// {{- get jiva volume | run -}}
func run(given *cmd.RunCommand) cmd.RunCommandResult {
	return given.Run()
}

// runAlways will always execute the provided command
//
// Examples:
// ---------
// {{- $cond := runAlways .TaskResult -}}
// {{- $store :=  storeAt .TaskResult -}}
// {{- $runner := storeRunnerCond $store $cond -}}
func runAlways() cmd.RunCondition {
	return cmd.RunAlways()
}

// storeAt sets the storage to save the command's execution result(s)
//
// Examples:
// ---------
// {{- $store :=  storeAt .TaskResult -}}
// {{- $runner := storeRunner $store -}}
func storeAt(kv map[string]interface{}) cmd.BucketStorageCondition {
	return cmd.KVStore(kv)
}

// storeRunner provides a utility that helps executing a run command as well as
// saving the execution result(s)
//
// Examples:
// ---------
// {{- $store :=  storeAt .TaskResult -}}
// {{- $runner := storeRunner $store -}}
func storeRunner(store cmd.BucketStorageCondition) cmd.Interface {
	return cmd.StoreCommand(store)
}

// storeRunnerCond provides a utility that helps executing a run command as well
// as saving the execution result(s)
//
// Examples:
// ---------
// {{- $cond := runAlways .TaskResult -}}
// {{- $store :=  storeAt .TaskResult -}}
// {{- $runner := storeRunnerCond $store $cond -}}
func storeRunnerCond(store cmd.BucketStorage, cond cmd.RunCondition) cmd.Interface {
	return cmd.StoreCommandCondition(store, cond)
}

// runas executes the runtask command by mapping it against a provided id
//
// Examples:
// ---------
// {{- $store :=  storeAt .TaskResult -}}
// {{- $runner := storeRunner $store -}}
// {{- get jiva volume | runas "getJivaVol" $runner -}}
// {{- get cstor volume | runas "getCstorVol" $runner -}}
func runas(id string, runner cmd.Interface, given *cmd.RunCommand) cmd.RunCommandResult {
	runner.Map(id, given)
	return runner.Run()
}

// runlog executes the runtask command and stores the result as well as other
// information at the provided paths
//
// Examples:
// ---------
// {{- get jiva volume | runlog "getj.result" "getj.extras" .Values -}}
//
// NOTES:
// -----------
// '.Values' is of type map[string]interface{} & is provided to the
// template before template execution
//
// Once above template gets executed '.Values' will have .Values.getj.result
// and .Values.getj.extras set with result and extras due to runtask command
// execution
func runlog(resultpath, debugpath string, store map[string]interface{}, given *cmd.RunCommand) (res cmd.RunCommandResult) {
	res = run(given)

	resultpath = strings.TrimPrefix(resultpath, ".")
	debugpath = strings.TrimPrefix(debugpath, ".")

	// result store
	resultlocator := strings.Split(resultpath, ".")
	util.SetNestedField(store, res.Result(), resultlocator...)

	// debug store
	debuglocator := strings.Split(debugpath, ".")
	util.SetNestedField(store, res.Debug(), debuglocator...)
	return
}

func mkNumberedSlice(n string) []string {
	slice := []string{}
	x, _ := strconv.Atoi(n)
	for i := 1; i <= x; i++ {
		slice = append(slice, strconv.Itoa(i))
	}
	return slice
}

// removeFirstElement removes the first element from a space
// separated string
func removeFirstElement(s string) string {
	i := strings.Index(s, " ") + 1
	return s[i:]
}

// runCommandFuncs returns the set of runtask command based template functions
func runCommandFuncs() template.FuncMap {
	return template.FuncMap{
		"delete":             delete,
		"get":                get,
		"lst":                list,
		"create":             create,
		"post":               post,
		"patch":              patch,
		"update":             update,
		"put":                put,
		"jiva":               jiva,
		"cstor":              cstor,
		"volume":             volume,
		"http":               http,
		"withOption":         withOption,
		"withoption":         withOption,
		"run":                run,
		"runlog":             runlog,
		"select":             slect,
		"snapshot":           snapshot,
		"storeAt":            storeAt,
		"storeRunner":        storeRunner,
		"storeRunnerCond":    storeRunnerCond,
		"runas":              runas,
		"runAlways":          runAlways,
		"toJsonObj":          toJsonObj,
		"mkNumberedSlice":    mkNumberedSlice,
		"removeFirstElement": removeFirstElement,
	}
}
