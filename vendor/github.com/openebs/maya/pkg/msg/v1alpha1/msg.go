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
	"fmt"
	"github.com/ghodss/yaml"
)

// YamlString returns the provided object as a yaml formatted string
func YamlString(ctx string, o interface{}) string {
	if o == nil {
		return ""
	}
	b, err := yaml.Marshal(o)
	if err != nil {
		return fmt.Sprintf("%s: failed to format '%s' as yaml string", err, ctx)
	}
	return fmt.Sprintf("\n%s", string(b))
}

type MsgType string

const (
	InfoMsg MsgType = "info"  // represents an information
	ErrMsg  MsgType = "error" // represents an error message
	WarnMsg MsgType = "warn"  // represents a warning message
	SkipMsg MsgType = "skip"  // represents a message about a skipped operation
)

type msg struct {
	Mtype MsgType `json:"type"`          // type of this message
	Desc  string  `json:"desc"`          // long description of this message
	Err   error   `json:"err,omitempty"` // if this message is an error
}

// String is an implementation of Stringer interface
func (m *msg) String() string {
	return YamlString("msg", m)
}

// GoString is an implementation of GoStringer interface
func (m *msg) GoString() string {
	return YamlString("msg", m)
}

// msgPredicate abstracts evaluation of a message condition
type msgPredicate func(given *msg) bool

func IsInfo(given *msg) (ok bool) {
	if given == nil {
		return
	}
	return given.Mtype == InfoMsg
}

func IsWarn(given *msg) (ok bool) {
	if given == nil {
		return
	}
	return given.Mtype == WarnMsg
}

func IsSkip(given *msg) (ok bool) {
	if given == nil {
		return
	}
	return given.Mtype == SkipMsg
}

func IsNotInfo(given *msg) (ok bool) {
	return !IsInfo(given)
}

func IsErr(given *msg) (ok bool) {
	if given == nil {
		return
	}
	return given.Mtype == ErrMsg
}

func IsNotErr(given *msg) (ok bool) {
	return !IsErr(given)
}

// Msgs represent a list of msg instance
type Msgs struct {
	Items []*msg `json:"items,omitempty"`
}

// String is an implementation of Stringer interface
func (m Msgs) String() string {
	return YamlString("msgs", m)
}

// GoString is an implementation of GoStringer interface
func (m Msgs) GoString() string {
	return YamlString("msgs", m)
}

func (m Msgs) Filter(p msgPredicate) (f Msgs) {
	for _, msg := range m.Items {
		if msg == nil {
			continue
		}
		if p(msg) {
			f.Items = append(f.Items, msg)
		}
	}
	return
}

func (m Msgs) Log(l func(string, ...interface{})) {
	for _, msg := range m.Items {
		if msg == nil {
			continue
		}
		l(msg.String())
	}
}

func (m Msgs) LogNonInfos(l func(string, ...interface{})) {
	m.Filter(IsNotInfo).Log(l)
}

func (m Msgs) LogNonErrors(l func(string, ...interface{})) {
	m.Filter(IsNotErr).Log(l)
}

func (m Msgs) LogErrors(l func(string, ...interface{})) {
	m.Filter(IsErr).Log(l)
}

func (m *Msgs) AddInfo(i string) (u *Msgs) {
	if len(i) == 0 {
		return m
	}
	m.Items = append(m.Items, &msg{Mtype: InfoMsg, Desc: i})
	return m
}

func (m *Msgs) AddWarn(w string) (u *Msgs) {
	if len(w) == 0 {
		return m
	}
	m.Items = append(m.Items, &msg{Mtype: WarnMsg, Desc: w})
	return m
}

func (m *Msgs) AddSkip(s string) (u *Msgs) {
	if len(s) == 0 {
		return m
	}
	m.Items = append(m.Items, &msg{Mtype: SkipMsg, Desc: s})
	return m
}

func (m *Msgs) AddError(e error) (u *Msgs) {
	if e == nil {
		return m
	}
	m.Items = append(m.Items, &msg{Mtype: ErrMsg, Desc: e.Error(), Err: e})
	return m
}

func (m *Msgs) Merge(s *Msgs) (u *Msgs) {
	if s == nil {
		return m
	}
	m.Items = append(m.Items, s.Items...)
	return m
}

// Reset clears the list of messages
func (m *Msgs) Reset() (u *Msgs) {
	m.Items = nil
	return m
}

func (m Msgs) Infos() (f Msgs) {
	return m.Filter(IsInfo)
}

func (m Msgs) NonInfos() (f Msgs) {
	return m.Filter(IsNotInfo)
}

func (m Msgs) Errors() (f Msgs) {
	return m.Filter(IsErr)
}

func (m Msgs) HasError() bool {
	return len(m.Errors().Items) != 0
}

func (m Msgs) NonErrors() (f Msgs) {
	return m.Filter(IsNotErr)
}

func (m Msgs) Skips() (f Msgs) {
	return m.Filter(IsSkip)
}

func (m Msgs) Warns() (f Msgs) {
	return m.Filter(IsWarn)
}

func (m Msgs) HasWarn() bool {
	return len(m.Filter(IsWarn).Items) != 0
}

// AllMsgs holds messages categorized per message type
type AllMsgs map[MsgType]Msgs

// String is an implementation of Stringer interface
func (a AllMsgs) String() string {
	return YamlString("allmsgs", a)
}

// GoString is an implementation of GoStringer interface
func (a AllMsgs) GoString() string {
	return YamlString("allmsgs", a)
}

// Error returns the first error that was recorded
func (a AllMsgs) Error() (err error) {
	if !a.HasError() {
		return
	}
	e := a[ErrMsg].Items[0]
	if e == nil {
		return
	}
	return e.Err
}

func (a AllMsgs) HasError() (iserr bool) {
	errs := a[ErrMsg]
	if len(errs.Items) == 0 {
		return
	}
	return true
}

func (a AllMsgs) HasWarn() (iswarn bool) {
	warns := a[WarnMsg]
	if len(warns.Items) == 0 {
		return
	}
	return true
}

func (a AllMsgs) HasSkip() (isskip bool) {
	skips := a[SkipMsg]
	if len(skips.Items) == 0 {
		return
	}
	return true
}

func (a AllMsgs) HasInfo() (isinfo bool) {
	infos := a[InfoMsg]
	if len(infos.Items) == 0 {
		return
	}
	return true
}

func (a AllMsgs) IsEmpty() (isempty bool) {
	warns := a[WarnMsg]
	infos := a[InfoMsg]
	errs := a[ErrMsg]
	skips := a[SkipMsg]

	if len(warns.Items) == 0 && len(errs.Items) == 0 && len(infos.Items) == 0 && len(skips.Items) == 0 {
		return true
	}
	return
}

func (a AllMsgs) ToMsgs() (m *Msgs) {
	m = &Msgs{}
	if len(a) == 0 {
		return
	}
	// grab the errors
	errors := a[ErrMsg].Items
	if len(errors) != 0 {
		m.Items = append(m.Items, errors...)
	}
	// grab the warns
	warns := a[WarnMsg].Items
	if len(warns) != 0 {
		m.Items = append(m.Items, warns...)
	}
	// grab the infos
	infos := a[InfoMsg].Items
	if len(infos) != 0 {
		m.Items = append(m.Items, infos...)
	}
	// grab the skips
	skips := a[SkipMsg].Items
	if len(skips) != 0 {
		m.Items = append(m.Items, skips...)
	}
	return
}

func (m Msgs) AllMsgs() (all AllMsgs) {
	return map[MsgType]Msgs{
		InfoMsg: m.Infos(),
		ErrMsg:  m.Errors(),
		WarnMsg: m.Warns(),
		SkipMsg: m.Skips(),
	}
}
