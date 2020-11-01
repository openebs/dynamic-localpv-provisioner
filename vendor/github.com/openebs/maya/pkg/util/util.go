/*
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

package util

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"k8s.io/klog"
)

var (
	// ErrMAPIADDRNotSet is the new error to display this error if MAPI_ADDR is not set.
	ErrMAPIADDRNotSet = errors.New("MAPI_ADDR environment variable not set")
	// ErrInternalServerError is the new error to raise if an error occurs while rendering the service
	ErrInternalServerError = errors.New("Internal Server Error")
	// ErrServerUnavailable is the new error to raise if the server is not available
	ErrServerUnavailable = errors.New("Server Unavailable")
	// ErrServerNotReachable is the new error to raise if the server is not reachable
	ErrServerNotReachable = errors.New("Server Not Reachable")
	// ErrPageNotFound is the new error to raise if the page is not found
	ErrPageNotFound = errors.New("Page Not Found")
)

// truthyValues maps a set of values which are considered as true
var truthyValues = map[string]bool{
	"1":       true,
	"YES":     true,
	"TRUE":    true,
	"OK":      true,
	"ENABLED": true,
	"ON":      true,
}

// CheckTruthy checks for truthiness of the passed argument.
func CheckTruthy(truth string) bool {
	return truthyValues[strings.ToUpper(truth)]
}

// falsyValues maps a set of values which are considered as false
var falsyValues = map[string]bool{
	"0":        true,
	"NO":       true,
	"FALSE":    true,
	"BLANK":    true,
	"DISABLED": true,
	"OFF":      true,
}

// CheckFalsy checks for non-truthiness of the passed argument.
func CheckFalsy(falsy string) bool {
	if len(falsy) == 0 {
		falsy = "blank"
	}
	return falsyValues[strings.ToUpper(falsy)]
}

// CheckErr to handle command errors
func CheckErr(err error, handleErr func(string)) {
	if err == nil {
		return
	}
	handleErr(err.Error())
}

// Fatal prints the message (if provided) and then exits. If V(2) or greater,
// klog.Fatal is invoked for extended information.
func Fatal(msg string) {
	if klog.V(2) {
		klog.FatalDepth(2, msg)
	}
	if len(msg) > 0 {
		// add newline if needed
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Fprint(os.Stderr, msg)
	}
	os.Exit(1)
}

// StringToInt32 converts a string type to corresponding
// *int32 type
func StringToInt32(val string) (*int32, error) {
	if len(val) == 0 {
		return nil, fmt.Errorf("Nil value to convert")
	}

	n, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return nil, err
	}
	n32 := int32(n)
	return &n32, nil
}

// StrToInt32 converts a string type to corresponding
// *int32 type
//
// NOTE:
//  This swallows the error if any
func StrToInt32(val string) *int32 {
	n32, _ := StringToInt32(val)
	return n32
}

// ContainsString returns true if the provided element is present in the
// provided array
func ContainsString(stringarr []string, element string) bool {
	for _, elem := range stringarr {
		if elem == element {
			return true
		}
	}
	return false
}

//TODO: Make better enhancement

// ListDiff returns list of string which are in listA but not in listB
func ListDiff(listA []string, listB []string) []string {
	outputList := []string{}
	mapListB := map[string]bool{}
	for _, str := range listB {
		mapListB[str] = true
	}
	for _, str := range listA {
		if !mapListB[str] {
			outputList = append(outputList, str)
		}
	}
	return outputList
}

// ListIntersection returns list of string which are in listA and listB
func ListIntersection(listA []string, listB []string) []string {
	outputList := []string{}
	mapListB := map[string]bool{}
	for _, str := range listB {
		mapListB[str] = true
	}
	for _, str := range listA {
		if mapListB[str] {
			outputList = append(outputList, str)
		}
	}
	return outputList
}

// ContainsKey returns true if the provided key is present in the provided map
func ContainsKey(mapOfObjs map[string]interface{}, key string) bool {
	for k := range mapOfObjs {
		if k == key {
			return true
		}
	}
	return false
}

// ContainKeys returns true if all the provided keys are present in the
// provided map
func ContainKeys(mapOfObjs map[string]interface{}, keys []string) bool {
	if len(keys) == 0 || len(mapOfObjs) == 0 {
		return false
	}

	allKeys := []string{}
	for k := range mapOfObjs {
		allKeys = append(allKeys, k)
	}

	for _, expectedKey := range keys {
		if !ContainsString(allKeys, expectedKey) {
			return false
		}
	}

	return true
}

// MergeMaps merges maps and returns the resulting map.
// map priority increases with order i.e. MergeMaps(m1,m2)
// will result in a map with overriding values from m2
func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// RemoveString removes all occurrences of a string from slice
// and returns a new updated slice
func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

// IsChangeInLists returns true if there is any difference in listA and listB
func IsChangeInLists(listA, listB []string) bool {
	listAMap := map[string]bool{}
	listBMap := map[string]bool{}
	for _, name := range listA {
		listAMap[name] = true
	}
	for _, name := range listB {
		listBMap[name] = true
	}
	for _, name := range listA {
		if !listBMap[name] {
			return true
		}
	}
	for _, name := range listB {
		if !listAMap[name] {
			return true
		}
	}
	return false
}

// RemoveItemFromSlice removes the string passed as argument from the slice
func RemoveItemFromSlice(slice []string, str string) []string {
	for index, value := range slice {
		if value == str {
			slice = append(slice[:index], slice[index+1:]...)
		}
	}
	return slice
}

// IsUniqueList returns true if values in list are not repeated else return
// false
func IsUniqueList(list []string) bool {
	listMap := map[string]bool{}

	for _, str := range list {
		if _, ok := listMap[str]; ok {
			return false
		}
		listMap[str] = true
	}
	return true
}

// IsCurrentLessThanNewVersion compares current and new version and returns true
// if currentversion is less `<` then new version (return true in case of equal version)
// TODO use version lib to properly handle versions https://github.com/hashicorp/go-version
func IsCurrentLessThanNewVersion(old, new string) bool {
	oldVersions := strings.Split(strings.Split(old, "-")[0], ".")
	newVersions := strings.Split(strings.Split(new, "-")[0], ".")
	for i := 0; i < len(oldVersions); i++ {
		oldVersion, _ := strconv.Atoi(oldVersions[i])
		newVersion, _ := strconv.Atoi(newVersions[i])
		if oldVersion > newVersion {
			return false
		}
	}
	return true
}
