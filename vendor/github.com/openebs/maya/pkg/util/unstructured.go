/*
Copyright 2015 The Kubernetes Authors.
Copyright 2018 The OpenEBS Authors.

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

// Package util provides functions based on k8s.io/apimachinery/pkg/apis/meta/v1/unstructured
// They are copied here to make them exported.
//
// TODO
// Check if it makes sense to import the entire unstructured package of
// k8s.io/apimachinery/pkg/apis/meta/v1/unstructured versus. copying
//
// TODO
// Move to maya/pkg/unstructured/v1alpha1 as helpers.go
package util

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/json"
)

// GetNestedField returns a nested field from the provided map
func GetNestedField(obj map[string]interface{}, fields ...string) interface{} {
	var val interface{} = obj
	for _, field := range fields {
		if _, ok := val.(map[string]interface{}); !ok {
			return nil
		}
		val = val.(map[string]interface{})[field]
	}
	return val
}

// GetNestedFieldInto converts a nested field to requested type from the provided map
func GetNestedFieldInto(out interface{}, obj map[string]interface{}, fields ...string) error {
	objMap := GetNestedField(obj, fields...)
	if objMap == nil {
		// If field has no value, leave `out` as is.
		return nil
	}
	// Decode into the requested output type.
	data, err := json.Marshal(objMap)
	if err != nil {
		return fmt.Errorf("can't encode nested field %v: %v", strings.Join(fields, "."), err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("can't decode nested field %v into type %T: %v", strings.Join(fields, "."), out, err)
	}
	return nil
}

// GetNestedString returns a nested string from the provided map
func GetNestedString(obj map[string]interface{}, fields ...string) string {
	if obj == nil {
		return ""
	}
	if str, ok := GetNestedField(obj, fields...).(string); ok {
		return str
	}
	return ""
}

// GetNestedArray returns an nested array from the provided map
func GetNestedArray(obj map[string]interface{}, fields ...string) []interface{} {
	if arr, ok := GetNestedField(obj, fields...).([]interface{}); ok {
		return arr
	}
	return nil
}

// GetNestedInt64 returns an nested int64 from the provided map
func GetNestedInt64(obj map[string]interface{}, fields ...string) int64 {
	if str, ok := GetNestedField(obj, fields...).(int64); ok {
		return str
	}
	return 0
}

// GetNestedInt64Pointer returns a nested int64 pointer from the provided map
func GetNestedInt64Pointer(obj map[string]interface{}, fields ...string) *int64 {
	nested := GetNestedField(obj, fields...)
	switch n := nested.(type) {
	case int64:
		return &n
	case *int64:
		return n
	default:
		return nil
	}
}

// GetNestedSlice returns a nested slice from the provided map
func GetNestedSlice(obj map[string]interface{}, fields ...string) []string {
	if m, ok := GetNestedField(obj, fields...).([]interface{}); ok {
		strSlice := make([]string, 0, len(m))
		for _, v := range m {
			if str, ok := v.(string); ok {
				strSlice = append(strSlice, str)
			}
		}
		return strSlice
	}
	return nil
}

// GetNestedMap returns a nested map from the provided map
func GetNestedMap(obj map[string]interface{}, fields ...string) map[string]string {
	if m, ok := GetNestedField(obj, fields...).(map[string]interface{}); ok {
		strMap := make(map[string]string, len(m))
		for k, v := range m {
			if str, ok := v.(string); ok {
				strMap[k] = str
			}
		}
		return strMap
	}
	return nil
}

// SetNestedField sets a nested field into the provided map
func SetNestedField(obj map[string]interface{}, value interface{}, fields ...string) {
	if len(fields) == 0 || obj == nil {
		return
	}

	m := obj

	if len(fields) > 1 {
		for _, field := range fields[0 : len(fields)-1] {
			if _, ok := m[field].(map[string]interface{}); !ok {
				m[field] = make(map[string]interface{})
			}
			m = m[field].(map[string]interface{})
		}
	}
	m[fields[len(fields)-1]] = value
}

// DeleteNestedField deletes a nested field from the provided map
func DeleteNestedField(obj map[string]interface{}, fields ...string) {
	if len(fields) == 0 || obj == nil {
		return
	}

	m := obj
	if len(fields) > 1 {
		for _, field := range fields[0 : len(fields)-1] {
			if _, ok := m[field].(map[string]interface{}); !ok {
				m[field] = make(map[string]interface{})
			}
			m = m[field].(map[string]interface{})
		}
	}
	delete(m, fields[len(fields)-1])
}

// SetNestedSlice sets a nested slice from the provided map
func SetNestedSlice(obj map[string]interface{}, value []string, fields ...string) {
	m := make([]interface{}, 0, len(value))
	for _, v := range value {
		m = append(m, v)
	}
	SetNestedField(obj, m, fields...)
}

// SetNestedMap sets a nested map from the provided map
func SetNestedMap(obj map[string]interface{}, value map[string]string, fields ...string) {
	m := make(map[string]interface{}, len(value))
	for k, v := range value {
		m[k] = v
	}
	SetNestedField(obj, m, fields...)
}

// MergeMapOfStrings will merge the map from src to dest
func MergeMapOfStrings(dest map[string]string, src map[string]string) bool {
	// nil check as storing into a nil map panics
	if dest == nil {
		return false
	}

	for k, v := range src {
		dest[k] = v
	}

	return true
}

// MergeMapOfObjects will merge the map from src to dest. It will override
// existing keys of the destination
func MergeMapOfObjects(dest map[string]interface{}, src map[string]interface{}) bool {
	// nil check as storing into a nil map panics
	if dest == nil {
		return false
	}

	for k, v := range src {
		dest[k] = v
	}

	return true
}

// GetMapOfStrings gets the direct value from the passed obj & the field path
// The value returned should be expected of the form map[string]string
func GetMapOfStrings(obj map[string]interface{}, field string) map[string]string {
	if m, ok := obj[field].(map[string]string); ok {
		return m
	}
	return nil
}
