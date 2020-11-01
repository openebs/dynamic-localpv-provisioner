/*
Copyright 2019 The OpenEBS Authors

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

package debug

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

// NewClient create a client to connect to the injection server.
func NewClient(host string) *Client {
	c := &Client{
		BaseURL: &url.URL{
			Scheme: "http",
			Host:   host,
		},
		httpClient: &http.Client{},
	}
	return c
}

// GetInject gives the error injection object.
func (c *Client) GetInject() (*ErrorInjection, error) {
	rel := &url.URL{Path: "/inject"}
	u := c.BaseURL.ResolveReference(rel)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var users ErrorInjection
	err = json.Unmarshal(body, &users)
	return &users, err
}

// PostInject will post the error injection object.
func (c *Client) PostInject(EI *ErrorInjection) error {
	rel := &url.URL{Path: "/inject"}
	u := c.BaseURL.ResolveReference(rel)
	b, err := json.Marshal(EI)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
