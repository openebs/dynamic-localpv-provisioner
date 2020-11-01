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
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	. "k8s.io/client-go/rest"
)

// HttpVerb is a typed constant that represents various http action
type HttpVerb string

const (
	DeleteAction HttpVerb = "DELETE"
	PostAction   HttpVerb = "POST"
	GetAction    HttpVerb = "GET"
	PutAction    HttpVerb = "PUT"
	PatchAction  HttpVerb = "PATCH"
)

// TODO
// expose REST via interface

// rest is a wrapper over restclient library
type Rest struct {
	client *RESTClient // rest client that does the actual API invocation
	base   string
	name   string
	verb   HttpVerb
	body   interface{}
}

// REST provides a new instance of rest client capable of invoking REST API
// calls
//
// NOTE:
// base url can be of form http://ipaddr:port with no trailing slash
func REST(base string) (r *Rest, err error) {
	c, err := RESTClientFor(&Config{
		Host: base,
		ContentConfig: ContentConfig{
			GroupVersion:         &v1.SchemeGroupVersion,
			NegotiatedSerializer: serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs},
		},
	})
	if err != nil {
		return
	}
	r = &Rest{base: base, client: c}
	return
}

func (r *Rest) WithName(n string) (u *Rest) {
	r.name = n
	return r
}

func (r *Rest) WithVerb(v HttpVerb) (u *Rest) {
	r.verb = v
	return r
}

func (r *Rest) WithBody(o interface{}) (u *Rest) {
	r.body = o
	return r
}

func (r *Rest) Do() (res interface{}, err error) {
	req := r.client.Verb(string(r.verb))
	if len(r.name) != 0 {
		req.Name(r.name)
	} else {
		req.RequestURI(r.base)
	}
	if r.body != nil {
		req.Body(r.body)
	}
	b, err := req.DoRaw()
	if err != nil {
		err = errors.Wrapf(err, "failed to invoke REST call %s %s %s", r.verb, r.base, r.name)
		return
	}
	err = json.Unmarshal(b, &res)
	if err != nil {
		err = errors.Wrapf(err, "failed to invoke REST call %s %s %s", r.verb, r.base, r.name)
	}
	return
}

// API performs a REST API call to a given named action
func API(verb HttpVerb, baseurl, name string) (b []byte, err error) {
	if len(name) == 0 {
		err = fmt.Errorf("empty resource name: failed to invoke REST API %s %s", verb, baseurl)
		return
	}
	r, err := REST(baseurl)
	if err != nil {
		err = errors.Wrapf(err, "failed to invoke REST API %s %s %s", verb, baseurl, name)
		return
	}
	req := r.client.Verb(string(verb)).Name(name)
	return doRaw(verb, req)
}

// URL performs a REST API call to a given url
func URL(verb HttpVerb, url string) (b []byte, err error) {
	if len(url) == 0 {
		err = fmt.Errorf("empty URL: failed to invoke REST API %s", verb)
		return
	}
	r, err := REST(url)
	if err != nil {
		err = errors.Wrapf(err, "failed to invoke REST API %s %s", verb, url)
		return
	}
	// override the formed URL with the one provided
	req := r.client.Verb(string(verb)).RequestURI(url)
	return doRaw(verb, req)
}

func doRaw(verb HttpVerb, req *Request) (b []byte, err error) {
	b, err = req.DoRaw()
	if err != nil {
		err = errors.Wrapf(err, "error invoking REST API %s %s", verb, req.URL())
		return
	}
	return
}
