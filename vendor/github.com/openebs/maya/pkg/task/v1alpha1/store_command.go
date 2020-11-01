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
	"time"

	msg "github.com/openebs/maya/pkg/msg/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/pkg/errors"
)

var (
	// ErrorCanNotRunDueToFailedCondition is an error object that indicates that command could not be executed due to failed condition.
	ErrorCanNotRunDueToFailedCondition = errors.New("run condition failed: can not execute run command")
)

// StoreKey represents supported keys to store run command's execution results
type StoreKey string

const (
	// ResultStoreKey stores the run commands results
	ResultStoreKey StoreKey = "result"
	//DebugStoreKey stores the debug results
	DebugStoreKey StoreKey = "debug"
	//ErrorStoreKey stores the errors after the run command
	ErrorStoreKey StoreKey = "error"
	//RootCauseStoreKey stores the rootCause of the errors
	RootCauseStoreKey StoreKey = "rootCause"
)

// BucketStorageCondition abstracts saving of run command results and also
// abstracts the evaluation of whether the run command should get executed or
// skipped in the first place
type BucketStorageCondition interface {
	BucketStorage
	RunCondition
}

// BucketStorage abstracts storing one or more run command results into
// buckets
type BucketStorage interface {
	Bucket
	Storage
}

// Bucket abstracts categorising run command results into specific bucket
type Bucket interface {
	SetBucket(b string)
	IsBucketTaken(b string) bool
}

// Storage abstracts saving run command results
type Storage interface {
	Store(key StoreKey, data interface{})
}

// kvStore is used to store one or more run commands' execution results
type kvStore struct {
	bucket string                 // identifies one run command
	MStore map[string]interface{} `json:"mstore"` // database of all run commands' results categorized by buckets
}

// KVStore returns a new instance of kvStore
func KVStore(store map[string]interface{}) *kvStore {
	return &kvStore{MStore: store}
}

// String implements Stringer interface
func (kv kvStore) String() string {
	return msg.YamlString("kvstore", kv)
}

// GoString implements GoStringer interface
func (kv kvStore) GoString() string {
	return msg.YamlString("kvstore", kv)
}

// SetBucket sets the bucket to be considered while saving data into the store
func (kv *kvStore) SetBucket(b string) {
	kv.bucket = b
}

// IsBucketTaken flags if the given bucket is already in use
func (kv *kvStore) IsBucketTaken(bucket string) bool {
	for b := range kv.MStore {
		if b == bucket {
			return true
		}
	}
	return false
}

// storeRootCause saves the first ever error saved into this store as root cause
func (kv *kvStore) storeRootCause(data interface{}) {
	if data == nil {
		return
	}
	c := util.GetNestedField(kv.MStore, string(RootCauseStoreKey))
	if c != nil {
		return
	}
	util.SetNestedField(kv.MStore, data, string(RootCauseStoreKey))
}

// Store saves the given data along with the given key in this kv store
func (kv *kvStore) Store(key StoreKey, data interface{}) {
	if kv.MStore == nil {
		return
	}
	util.SetNestedField(kv.MStore, data, kv.bucket, string(key))
	if key == ErrorStoreKey {
		kv.storeRootCause(data)
	}
}

// WillRun evaluates if run command execution should proceed or be skipped based
// on the results of earlier run commands
//
// NOTE:
// This is an implementation of RunCondition
func (kv *kvStore) WillRun() (condition string, willrun bool) {
	condition = "errors with previous commands' execution(s)"
	for _, data := range kv.MStore {
		kv, ok := data.(map[string]interface{})
		if !ok {
			continue
		}
		if kv[string(ErrorStoreKey)] != nil {
			return
		}
	}
	willrun = true
	return
}

// storeCommand is a wrapper over RunCommand instance and provides extra
// features e.g. storing the run command results & evaluating if run command
// should be executed or skipped.
//
// NOTE:
// A storeCommand will mostly be preferred over a RunCommand instance
type storeCommand struct {
	storage   BucketStorage // store a run command's result, error, debug messages
	cond      RunCondition  // flag that determines if run command will execute or not
	id        string        // unique identification of run command
	cmd       *RunCommand   // current command to execute
	*msg.Msgs               // store and retrieve info, warns, errors, etc occurred during execution
}

// StoreCommand returns a new instance of storeCommand
func StoreCommand(s BucketStorageCondition) *storeCommand {
	return StoreCommandCondition(s, s)
}

// StoreCommandCondition returns a new instance of storeCommand
func StoreCommandCondition(s BucketStorage, c RunCondition) *storeCommand {
	return &storeCommand{storage: s, cond: c, Msgs: &msg.Msgs{}}
}

// reset will reset this instance to manage a new run command instance for
// later's execution
func (r *storeCommand) reset() {
	r.id = ""
	r.cmd = nil
	r.Msgs.Reset()
}

// setID sets the run command's identity
func (r *storeCommand) setID(id string) {
	r.storage.SetBucket(id)
	r.id = id
}

// setGenID sets the run command's identity based on a generated number
func (r *storeCommand) setGenID() {
	id := time.Now().Format("15040500000")
	r.setID(id)
}

// ID returns the id corresponding to the run command
func (r *storeCommand) ID() string {
	return r.id
}

// WillRun evaluates if the run command will get executed or skipped
func (r *storeCommand) WillRun() (condition string, willrun bool) {
	return r.cond.WillRun()
}

// Map sets the given run command against the given id
func (r *storeCommand) Map(id string, c *RunCommand) {
	// reset earlier settings if any
	r.reset()

	if len(id) == 0 {
		r.AddError(errors.Errorf("missing run command id: can not execute run command: '%s'", c.SelfInfo()))
		r.setGenID()
		return
	}
	if r.storage.IsBucketTaken(id) {
		r.AddError(errors.Errorf("duplicate id '%s': can not execute run command: '%s'", id, c.SelfInfo()))
		r.setGenID()
		return
	}
	r.setID(id)
	if c == nil {
		r.AddError(errors.Errorf("nil run command: can not execute run command with id '%s'", id))
		return
	}
	r.cmd = c
	return
}

// WithResult sets the result due to execution of run command
func (r *storeCommand) WithResult(v interface{}) *storeCommand {
	r.storage.Store(ResultStoreKey, v)
	return r
}

// WithDebug sets the debug information due to execution of run command
func (r *storeCommand) WithDebug(d msg.AllMsgs) *storeCommand {
	r.storage.Store(DebugStoreKey, d)
	return r
}

// WithError sets the error due to execution of run command
func (r *storeCommand) WithError(e error) *storeCommand {
	r.storage.Store(ErrorStoreKey, e)
	return r
}

// Stores sets various responses due to execution of run command
func (r *storeCommand) Store(res RunCommandResult) {
	r.WithResult(res.Result()).WithDebug(res.Debug()).WithError(res.Error())
}

// preRun sets options conditionally prior to execution of run command
func (r *storeCommand) preRun() {
	if r.cmd == nil {
		return
	}
	c, willrun := r.WillRun()
	if !willrun {
		r.cmd.Enable(Off).AddWarn(c).AddError(ErrorCanNotRunDueToFailedCondition)
	}
}

// postRun executes the post activities associated after the execution of
// run command
func (r *storeCommand) postRun(res RunCommandResult) {
	r.Store(res)
}

// Run executes the run command
func (r *storeCommand) Run() (res RunCommandResult) {
	if r.cmd == nil {
		res = NewRunCommandResult(nil, r.AllMsgs())
	} else {
		r.preRun()
		res = r.cmd.Run()
	}
	r.postRun(res)
	return
}
