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

package v1alpha1

import (
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetHostName returns the hostname corresponding to
// the provided node name
func GetHostName(name string) (string, error) {
	return KubeClientInstanceOrDie().GetHostName(name, metav1.GetOptions{})
}

// GetHostNameOrNodeName returns the hostname corresponding
// to the provided node name or node name itself if hostname
// is not available
func GetHostNameOrNodeName(name string) (string, error) {
	return KubeClientInstanceOrDie().
		GetHostNameOrNodeName(name, metav1.GetOptions{})
}

// TemplateFunctions exposes a few functions as go template functions
func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"kubeNodeGetHostName":           GetHostName,
		"kubeNodeGetHostNameOrNodeName": GetHostNameOrNodeName,
	}
}
