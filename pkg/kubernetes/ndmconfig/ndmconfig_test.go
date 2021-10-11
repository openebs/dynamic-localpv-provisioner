/*
Copyright 2021 The OpenEBS Authors

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

package ndmconfig

import (
	"testing"
)

func TestAppendToPathFilter(t *testing.T) {
	tests := map[string]struct {
		availableConfig     *Config
		lisType             ListType
		diskPath            string
		expectedIncludeList string
		expectedExcludeList string
	}{
		"append to non-empty include list": {
			availableConfig:     &Config{FilterConfigs: []FilterConfig{{"path-filter", "path filter", "true", "/dev/loop1,/dev/loop2", ""}}},
			lisType:             Include,
			diskPath:            "/dev/loop9000",
			expectedIncludeList: "/dev/loop1,/dev/loop2,/dev/loop9000",
			expectedExcludeList: "",
		},
		"append to empty include list": {
			availableConfig:     &Config{FilterConfigs: []FilterConfig{{"path-filter", "path filter", "true", "", ""}}},
			lisType:             Include,
			diskPath:            "/dev/loop9000",
			expectedIncludeList: "/dev/loop9000",
			expectedExcludeList: "",
		},
		"append to non-empty exclude list": {
			availableConfig:     &Config{FilterConfigs: []FilterConfig{{"path-filter", "path filter", "true", "", "/dev/fd0,/dev/sr0,/dev/ram,/dev/dm-,/dev/md,/dev/rbd,/dev/zd"}}},
			lisType:             Exclude,
			diskPath:            "/dev/loop9000",
			expectedIncludeList: "",
			expectedExcludeList: "/dev/fd0,/dev/sr0,/dev/ram,/dev/dm-,/dev/md,/dev/rbd,/dev/zd,/dev/loop9000",
		},
		"append to empty exclude list": {
			availableConfig:     &Config{FilterConfigs: []FilterConfig{{"path-filter", "path filter", "true", "", ""}}},
			lisType:             Exclude,
			diskPath:            "/dev/loop9000",
			expectedIncludeList: "",
			expectedExcludeList: "/dev/loop9000",
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			err := test.availableConfig.AppendToPathFilter(test.lisType, test.diskPath)

			if err != nil ||
				test.availableConfig.FilterConfigs[0].Include != test.expectedIncludeList ||
				test.availableConfig.FilterConfigs[0].Exclude != test.expectedExcludeList {
				t.Fatalf(
					"Test %v failed, Expected 'include: %v' but got 'include: %v', "+
						"Expected 'exclude: %v' but got 'exclude: %v'",
					name,
					test.expectedIncludeList, test.availableConfig.FilterConfigs[0].Include,
					test.expectedExcludeList, test.availableConfig.FilterConfigs[0].Exclude,
				)
			}
		})
	}
}

func TestRemoveFromPathFilter(t *testing.T) {
	tests := map[string]struct {
		availableConfig     *Config
		lisType             ListType
		diskPath            string
		expectedIncludeList string
		expectedExcludeList string
	}{
		"remove from non-empty include list": {
			availableConfig:     &Config{FilterConfigs: []FilterConfig{{"path-filter", "path filter", "true", "/dev/loop1,/dev/loop2,/dev/loop9000", ""}}},
			lisType:             Include,
			diskPath:            "/dev/loop9000",
			expectedIncludeList: "/dev/loop1,/dev/loop2",
			expectedExcludeList: "",
		},
		"remove from otherwise-empty include list": {
			availableConfig:     &Config{FilterConfigs: []FilterConfig{{"path-filter", "path filter", "true", "/dev/loop9000", ""}}},
			lisType:             Include,
			diskPath:            "/dev/loop9000",
			expectedIncludeList: "",
			expectedExcludeList: "",
		},
		"remove from non-empty exclude list": {
			availableConfig:     &Config{FilterConfigs: []FilterConfig{{"path-filter", "path filter", "true", "", "/dev/fd0,/dev/sr0,/dev/ram,/dev/dm-,/dev/md,/dev/rbd,/dev/zd,/dev/loop9000"}}},
			lisType:             Exclude,
			diskPath:            "/dev/loop9000",
			expectedIncludeList: "",
			expectedExcludeList: "/dev/fd0,/dev/sr0,/dev/ram,/dev/dm-,/dev/md,/dev/rbd,/dev/zd",
		},
		"remove from otherwise-empty exclude list": {
			availableConfig:     &Config{FilterConfigs: []FilterConfig{{"path-filter", "path filter", "true", "", "/dev/loop9000"}}},
			lisType:             Exclude,
			diskPath:            "/dev/loop9000",
			expectedIncludeList: "",
			expectedExcludeList: "",
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			err := test.availableConfig.RemoveFromPathFilter(test.lisType, test.diskPath)

			if err != nil ||
				test.availableConfig.FilterConfigs[0].Include != test.expectedIncludeList ||
				test.availableConfig.FilterConfigs[0].Exclude != test.expectedExcludeList {
				t.Fatalf(
					"Test %v failed, Expected 'include: %v' but got 'include: %v', "+
						"Expected 'exclude: %v' but got 'exclude: %v'",
					name,
					test.expectedIncludeList, test.availableConfig.FilterConfigs[0].Include,
					test.expectedExcludeList, test.availableConfig.FilterConfigs[0].Exclude,
				)
			}
		})
	}
}
