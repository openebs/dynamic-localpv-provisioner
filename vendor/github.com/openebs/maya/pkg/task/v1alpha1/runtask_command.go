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
	"errors"
	"fmt"
	"strings"

	jp "github.com/openebs/maya/pkg/jsonpath/v1alpha1"
	msg "github.com/openebs/maya/pkg/msg/v1alpha1"
)

const (
	// SkipExecutionMessage will skip run command execution
	SkipExecutionMessage string = "will skip run command execution"
)

// RunCommandAction determines the kind of action that gets executed by run task
// command
type RunCommandAction string

const (
	// DeleteCommandAction represents a run command as a delete action
	DeleteCommandAction RunCommandAction = "delete"
	// CreateCommandAction represents a run command as a create action
	CreateCommandAction RunCommandAction = "create"
	// PostCommandAction represents a run command as a post action
	PostCommandAction RunCommandAction = "post"
	// GetCommandAction represents a run command as a get action
	GetCommandAction RunCommandAction = "get"
	// ListCommandAction represents a run command as a list action
	ListCommandAction RunCommandAction = "list"
	// PatchCommandAction represents a run command as a patch action
	PatchCommandAction RunCommandAction = "patch"
	// UpdateCommandAction represents a run command as a update action
	UpdateCommandAction RunCommandAction = "update"
	// PutCommandAction represents a run command as a put action
	PutCommandAction RunCommandAction = "put"
)

// RunCommandCategory represents the category of the run command. It helps
// in determining the exact entity or feature this run command is targeting.
//
// NOTE:
//  A run command can have more than one categories to determine an entity
type RunCommandCategory string

const (
	// JivaCommandCategory categorises the run command as jiva based
	JivaCommandCategory RunCommandCategory = "jiva"
	// CstorCommandCategory categorises the run command as cstor based
	CstorCommandCategory RunCommandCategory = "cstor"
	// VolumeCommandCategory categorises the run command as volume based
	VolumeCommandCategory RunCommandCategory = "volume"
	// PoolCommandCategory categorises the run command as pool based
	PoolCommandCategory RunCommandCategory = "pool"
	// HttpCommandCategory categorises the run command as http based
	HttpCommandCategory RunCommandCategory = "http"
	// SnapshotCommandCategory categorises the run command as snapshot based
	SnapshotCommandCategory RunCommandCategory = "snapshot"
)

// RunCommandCategoryList represents a list of RunCommandCategory
type RunCommandCategoryList []RunCommandCategory

// String implements Stringer interface
func (l RunCommandCategoryList) String() string {
	return msg.YamlString("runcommandcategories", l)
}

// Contains returns true if this list has the given category
func (l RunCommandCategoryList) Contains(given RunCommandCategory) (no bool) {
	if len(l) == 0 {
		return
	}
	for _, category := range l {
		if category == given {
			return !no
		}
	}
	return
}

// IsJivaVolume returns true if this list has both jiva and volume as its
// category items
func (l RunCommandCategoryList) IsJivaVolume() (no bool) {
	if len(l) == 0 {
		return
	}
	if l.Contains(JivaCommandCategory) && l.Contains(VolumeCommandCategory) {
		return !no
	}
	return
}

// IsHttpReq returns true if this list points to a http based request
func (l RunCommandCategoryList) IsHttpReq() (no bool) {
	if len(l) == 0 {
		return
	}
	if l.Contains(HttpCommandCategory) {
		return !no
	}
	return
}

// IsCstorVolume returns true if this list has both cstor and volume as its
// category items
func (l RunCommandCategoryList) IsCstorVolume() (no bool) {
	if len(l) == 0 {
		return
	}
	if l.Contains(CstorCommandCategory) && l.Contains(VolumeCommandCategory) {
		return !no
	}
	return
}

// IsCstorSnapshot returns true if this list has both cstor and snapshot as its
// category items
func (l RunCommandCategoryList) IsCstorSnapshot() (no bool) {
	if l.Contains(CstorCommandCategory) && l.Contains(SnapshotCommandCategory) {
		return !no
	}
	return
}

// IsValid returns true if category list is valid
//
// TODO
// Move volume specific validations to volume command file
func (l RunCommandCategoryList) IsValid() (no bool) {
	if l.Contains(JivaCommandCategory) && l.Contains(CstorCommandCategory) {
		// a volume can be either cstor or jiva based; not both
		return
	}
	return !no
}

// IsEmpty returns true if no category is set
func (l RunCommandCategoryList) IsEmpty() (empty bool) {
	if len(l) == 0 {
		return true
	}
	return
}

// RunCommandData represents data provided to the run command before
// its execution i.e. input data
type RunCommandData interface{}

// RunCommandDataMap represents a map of input data required to execute
// run command
type RunCommandDataMap map[string]RunCommandData

// String implements Stringer interface
func (m RunCommandDataMap) String() string {
	return msg.YamlString("runcommanddatamap", m)
}

// RunCommandResult holds the result and execution info of a run command
type RunCommandResult struct {
	Res    interface{} `json:"result"`          // result of run command execution
	Err    error       `json:"error"`           // root cause of issue; error if any during run command execution
	Extras msg.AllMsgs `json:"debug,omitempty"` // debug details i.e. errors, warnings, information, etc during execution
}

// NewRunCommandResult returns a new RunCommandResult struct
func NewRunCommandResult(result interface{}, extras msg.AllMsgs) (r RunCommandResult) {
	return RunCommandResult{
		Res:    result,
		Err:    extras.Error(),
		Extras: extras,
	}
}

// String implements Stringer interface
func (r RunCommandResult) String() string {
	return msg.YamlString("runcommandresult", r)
}

// GoString implements GoStringer interface
func (r RunCommandResult) GoString() string {
	return msg.YamlString("runcommandresult", r)
}

// Error returns the error if any from the run command's result
func (r RunCommandResult) Error() error {
	return r.Err
}

// Result returns the expected output if any from the run command's result
func (r RunCommandResult) Result() interface{} {
	return r.Res
}

// Debug returns the debug info gathered during execution of run command's
// result
func (r RunCommandResult) Debug() msg.AllMsgs {
	return r.Extras
}

// SelectPathAliasDelimiter is used to delimit a select path from its alias
//
// e.g.
//
// ".metadata.namespace as namespace" implies
// - '.metadata.namespace' is the path
// - ' as ' is the delimiter
// - 'namespace' is the alias
type SelectPathAliasDelimiter string

const (
	// AsSelectDelimiter represents " as " as the delimiter
	AsSelectDelimiter SelectPathAliasDelimiter = " as "
)

// SelectPaths holds all the select paths specified in a run command
type SelectPaths []string

// String implements Stringer interface
func (s SelectPaths) String() (str string) {
	if len(s) > 0 {
		str = "select '" + strings.Join(s, "' '") + "'"
	}
	return
}

// aliasPaths transforms the select paths into a map of alias & corresponding
// path
func (s SelectPaths) aliasPaths() (ap map[string]string) {
	if len(s) == 0 {
		return
	}
	ap = map[string]string{}
	for idx, slt := range s {
		splits := strings.Split(slt, string(AsSelectDelimiter))
		if len(splits) == 2 {
			ap[splits[1]] = splits[0]
		} else {
			ap[fmt.Sprintf("s%d", idx)] = slt
		}
	}
	return
}

// QueryCommandResult queries the run command's result based on the select paths
func (s SelectPaths) QueryCommandResult(r RunCommandResult) (u RunCommandResult) {
	result := r.Result()
	if result == nil {
		msgs := r.Debug().ToMsgs().AddWarn(fmt.Sprintf("nil command result: can not query %s", s))
		return NewRunCommandResult(nil, msgs.AllMsgs())
	}
	// execute jsonpath query against the result
	j := jp.JSONPath(s.String()).WithTarget(result)
	sl := j.QueryAll(jp.SelectionList(s.aliasPaths()))
	// return a new result with selected path values and add additional debug info
	// due to jsonpath query
	u = NewRunCommandResult(sl.Values(), r.Debug().ToMsgs().Merge(j.Msgs).AllMsgs())
	return
}

// These represents error messages
var (
	ErrorNotSupportedCategory = errors.New("not supported category: invalid run command")
	ErrorNotSupportedAction   = errors.New("not supported action: invalid run command")
	ErrorInvalidCategory      = errors.New("invalid categories: invalid run command")
	ErrorEmptyCategory        = errors.New("missing categories: invalid run command")
)

// Interface abstracts execution run command
type Interface interface {
	IDMapper
	Runner
	RunCondition
}

// IDMapper abstracts mapping of a RunCommand instance against an id
type IDMapper interface {
	ID() string
	Map(id string, r *RunCommand)
}

// RunCondition abstracts evaluating the condition to run or skip executing a
// run command
type RunCondition interface {
	WillRun() (condition string, willrun bool)
}

// runAlways is an implementation of RunCondition that evaluates the condition
// to execute a run command to true. In other words, any run command will get
// executed if this instance is set as former's run condition.
type runAlways struct{}

// RunAlways returns a new instance of runAlways
func RunAlways() *runAlways {
	return &runAlways{}
}

// WillRun returns true always
func (r *runAlways) WillRun() (condition string, willrun bool) {
	return "execute the run command always", true
}

// Runner abstracts execution of command
type Runner interface {
	Run() (r RunCommandResult)
}

// RunPredicate abstracts evaluation of executing or skipping execution
// of a runner instance
type RunPredicate func() bool

// On enables a runner instance
func On() bool {
	return true
}

// Off disables a runner instance
func Off() bool {
	return false
}

// RunCommand represent a run command
type RunCommand struct {
	ID        string                 // uniquely identifies a run command
	WillRun   bool                   // flags if this run command should get executed or not
	Action    RunCommandAction       // represents the run command's action
	Category  RunCommandCategoryList // classification of run command
	Data      RunCommandDataMap      // input data required to execute run command
	Selects   SelectPaths            // paths whose values will be retrieved after run command execution
	*msg.Msgs                        // store and retrieve info, warns, errors, etc occurred during execution
}

// SelfInfo returns this instance of RunCommand as a string format
func (c *RunCommand) SelfInfo() (me string) {
	if c == nil {
		return
	}
	var selects, categories, data string
	if len(c.Selects) > 0 {
		selects = c.Selects.String() + " "
	}
	for _, c := range c.Category {
		categories = categories + " " + string(c)
	}
	for n, d := range c.Data {
		data = data + fmt.Sprintf(" --%s=%s", n, d)
	}
	willrun := fmt.Sprintf(" --willrun=%t", c.WillRun)
	me = fmt.Sprintf("%s%s%s%s%s", selects, c.Action, categories, data, willrun)
	return
}

// Command returns a new instance of RunCommand
func Command() *RunCommand {
	return &RunCommand{Msgs: &msg.Msgs{}, WillRun: true}
}

// Enable enables or disables execution of RunCommand instance based on the
// outcome of given predicate
func (c *RunCommand) Enable(p RunPredicate) (u *RunCommand) {
	c.WillRun = p()
	return c
}

// IsRun flags if this run command will get executed or not
func (c *RunCommand) IsRun() bool {
	return c.WillRun
}

// AddError updates RunCommand instance with given error
func (c *RunCommand) AddError(err error) (u *RunCommand) {
	c.Msgs.AddError(err)
	return c
}

// CreateAction updates RunCommand instance with create action
func (c *RunCommand) CreateAction() (u *RunCommand) {
	c.Action = CreateCommandAction
	return c
}

// PostAction updates RunCommand instance with post action
func (c *RunCommand) PostAction() (u *RunCommand) {
	c.Action = PostCommandAction
	return c
}

// PutAction updates RunCommand instance with put action
func (c *RunCommand) PutAction() (u *RunCommand) {
	c.Action = PutCommandAction
	return c
}

// DeleteAction updates RunCommand instance with delete action
func (c *RunCommand) DeleteAction() (u *RunCommand) {
	c.Action = DeleteCommandAction
	return c
}

// GetAction updates RunCommand instance with get action
func (c *RunCommand) GetAction() (u *RunCommand) {
	c.Action = GetCommandAction
	return c
}

// ListAction updates RunCommand instance with list action
func (c *RunCommand) ListAction() (u *RunCommand) {
	c.Action = ListCommandAction
	return c
}

// UpdateAction updates RunCommand instance with update action
func (c *RunCommand) UpdateAction() (u *RunCommand) {
	c.Action = UpdateCommandAction
	return c
}

// PatchAction updates RunCommand instance with patch action
func (c *RunCommand) PatchAction() (u *RunCommand) {
	c.Action = PatchCommandAction
	return c
}

// WithCategory updates the given RunCommand instance with provided category
func WithCategory(given *RunCommand, category RunCommandCategory) (updated *RunCommand) {
	given.Category = append(given.Category, category)
	return given
}

// WithAction updates the given RunCommand instance with provided action
func WithAction(given *RunCommand, action RunCommandAction) (updated *RunCommand) {
	given.Action = action
	return given
}

// WithData updates the given RunCommand instance with provided input data
func WithData(given *RunCommand, name string, d RunCommandData) (updated *RunCommand) {
	if given.Data == nil {
		given.Data = map[string]RunCommandData{}
	}
	if d == nil {
		given.AddWarn(fmt.Sprintf("nil value provided for '%s': run command may fail", name))
	}
	given.Data[name] = d
	return given
}

// WithSelect updates the given RunCommand instance with provided select paths
func WithSelect(given *RunCommand, paths []string) (updated *RunCommand) {
	if len(paths) == 0 {
		return given
	}
	given.Selects = append(given.Selects, paths...)
	return given
}

func (c *RunCommand) String() string {
	return msg.YamlString("runcommand", c)
}

// Result is the name of method on RunCommand
func (c *RunCommand) Result(result interface{}) (r RunCommandResult) {
	return NewRunCommandResult(result, c.AllMsgs())
}

// instance fetches the specific run command implementation instance based
// on command categories
func (c *RunCommand) instance() (r Runner) {
	if c.Category.IsJivaVolume() {
		r = &jivaVolumeCommand{c}
	} else if c.Category.IsHttpReq() {
		r = HttpCommand(c)
	} else if c.Category.IsCstorSnapshot() {
		r = &cstorSnapshotCommand{c}
	} else {
		r = &notSupportedCategoryCommand{c}
	}
	return
}

// preRun evaluates conditions and sets options prior to execution of run
// command
func (c *RunCommand) preRun() {
	if c.Category.IsEmpty() {
		c.Enable(Off).AddError(ErrorEmptyCategory)
	}
	if !c.Category.IsValid() {
		c.Enable(Off).AddError(ErrorInvalidCategory)
	}
	if !c.IsRun() {
		c.AddSkip(SkipExecutionMessage)
	}
}

// postRun invokes operations after executing the run command
func (c *RunCommand) postRun(r RunCommandResult) (u RunCommandResult) {
	if len(c.Selects) == 0 {
		return r
	}
	u = c.Selects.QueryCommandResult(r)
	return
}

// Run finds the specific run command implementation and executes the same
func (c *RunCommand) Run() (r RunCommandResult) {
	// prior to run
	c.preRun()

	// run
	c.AddInfo(c.SelfInfo())
	if !c.IsRun() {
		// no need of post run
		return c.Result(nil)
	}
	r = c.instance().Run()

	// post run
	r = c.postRun(r)
	return
}

// RunCommandMiddleware abstracts updating the given RunCommand instance
type RunCommandMiddleware func(given *RunCommand) (updated *RunCommand)

// JivaCategory updates RunCommand instance with jiva as the run command's
// category
func JivaCategory() RunCommandMiddleware {
	return func(given *RunCommand) (updated *RunCommand) {
		return WithCategory(given, JivaCommandCategory)
	}
}

// HttpCategory updates RunCommand instance with http as the run command's
// category
func HttpCategory() RunCommandMiddleware {
	return func(given *RunCommand) (updated *RunCommand) {
		return WithCategory(given, HttpCommandCategory)
	}
}

// CstorCategory updates RunCommand instance with cstor as the run command's
// category
func CstorCategory() RunCommandMiddleware {
	return func(given *RunCommand) (updated *RunCommand) {
		return WithCategory(given, CstorCommandCategory)
	}
}

// VolumeCategory updates RunCommand instance with volume as the run
// command's category
func VolumeCategory() RunCommandMiddleware {
	return func(given *RunCommand) (updated *RunCommand) {
		return WithCategory(given, VolumeCommandCategory)
	}
}

// SnapshotCategory updates RunCommand instance with snapshot as the runtask
// command's category
func SnapshotCategory() RunCommandMiddleware {
	return func(given *RunCommand) (updated *RunCommand) {
		return WithCategory(given, SnapshotCommandCategory)
	}
}

// Select updates the RunCommand instance with paths whose values will be
// extracted after execution of run command
func Select(paths []string) RunCommandMiddleware {
	return func(given *RunCommand) (updated *RunCommand) {
		return WithSelect(given, paths)
	}
}

// RunCommandMiddlewareList represents a list of RunCommandMiddleware
type RunCommandMiddlewareList []RunCommandMiddleware

// Update updates the given RunCommand instance through all the middlewares
func (l RunCommandMiddlewareList) Update(given *RunCommand) (updated *RunCommand) {
	if len(l) == 0 || given == nil {
		return given
	}
	for _, middleware := range l {
		given = middleware(given)
	}
	return given
}

// notSupportedCategoryCommand is a CommandRunner implementation for
// un-supported run command category
type notSupportedCategoryCommand struct {
	*RunCommand
}

func (c *notSupportedCategoryCommand) Run() (r RunCommandResult) {
	c.AddError(ErrorNotSupportedCategory)
	return NewRunCommandResult(nil, c.AllMsgs())
}

// notSupportedActionCommand is a CommandRunner implementation for
// un-supported run command action
type notSupportedActionCommand struct {
	*RunCommand
}

func (c *notSupportedActionCommand) Run() (r RunCommandResult) {
	c.AddError(ErrorNotSupportedAction)
	return NewRunCommandResult(nil, c.AllMsgs())
}
