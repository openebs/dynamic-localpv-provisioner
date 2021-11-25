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
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	ProbeConfigs  []ProbeConfig  `yaml:"probeconfigs,omitempty"`  // ProbeConfigs contains configs of Probes
	FilterConfigs []FilterConfig `yaml:"filterconfigs,omitempty"` // FilterConfigs contains configs of Filters
	TagConfigs    []TagConfig    `yaml:"tagconfigs,omitempty"`    // TagConfigs contains configs for tags
}

// ProbeConfig contains configs of Probe
type ProbeConfig struct {
	Key   string `yaml:"key"`   // Key is key for each Probe
	Name  string `yaml:"name"`  // Name is name of Probe
	State string `yaml:"state"` // State is state of Probe
}

// FilterConfig contains configs of Filter
type FilterConfig struct {
	Key     string `yaml:"key"`               // Key is key for each Filter
	Name    string `yaml:"name"`              // Name is name of Filer
	State   string `yaml:"state"`             // State is state of Filter
	Include string `yaml:"include,omitempty"` // Include contains , separated values which we want to include for filter
	Exclude string `yaml:"exclude,omitempty"` // Exclude contains , separated values which we want to exclude for filter
}

type TagConfig struct {
	Name    string `yaml:"name,omitempty"`
	Type    string `yaml:"type,omitempty"`
	Pattern string `yaml:"pattern,omitempty"`
	TagName string `yaml:"tag,omitempty"`
}

// Type of list -- include, exclude
type ListType string

const (
	Include ListType = "include"
	Exclude ListType = "exclude"
)

func NewConfigFromAPIConfigMap(ndmConfigMap *corev1.ConfigMap) (*Config, error) {
	if ndmConfigMap == nil {
		return nil, errors.New("NDM ConfigMap is 'nil'")
	}

	var c Config

	ndmConfigString := ndmConfigMap.Data["node-disk-manager.config"]
	err := yaml.Unmarshal([]byte(ndmConfigString), &c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal NDM config")
	}

	return &c, nil
}

func (c *Config) AppendToPathFilter(listtype ListType, diskPath string) error {
	if c == nil {
		return errors.New("Config is nil")
	}

	var pathFilterIndex int
	foundFlag := 0
	for index, filterConfig := range c.FilterConfigs {
		if filterConfig.Key == "path-filter" {
			pathFilterIndex = index
			foundFlag++
			break
		}
	}
	if foundFlag == 0 {
		return errors.New("No filterconfig with 'key: path-filter' found")
	}

	separator := ""
	switch listtype {
	case Include:
		if len(c.FilterConfigs[pathFilterIndex].Include) > 0 {
			separator = ","
		}
		c.FilterConfigs[pathFilterIndex].Include = c.FilterConfigs[pathFilterIndex].Include + separator + diskPath
		return nil
	case Exclude:
		if len(c.FilterConfigs[pathFilterIndex].Exclude) > 0 {
			separator = ","
		}
		c.FilterConfigs[pathFilterIndex].Exclude = c.FilterConfigs[pathFilterIndex].Exclude + separator + diskPath
		return nil
	default:
		return errors.New("invalid filterconfig path-filter list name")
	}
}

func (c *Config) RemoveFromPathFilter(listtype ListType, diskPath string) error {
	if c == nil {
		return errors.New("Config is nil")
	}

	var pathFilterIndex int
	foundFlag := 0
	for index, filterConfig := range c.FilterConfigs {
		if filterConfig.Key == "path-filter" {
			pathFilterIndex = index
			foundFlag++
			break
		}
	}
	if foundFlag == 0 {
		return errors.New("No filterconfig with 'key: path-filter' found")
	}

	switch listtype {
	case Include:
		finalList := strings.ReplaceAll(c.FilterConfigs[pathFilterIndex].Include, ","+diskPath, "")
		finalList = strings.ReplaceAll(finalList, diskPath, "")
		c.FilterConfigs[pathFilterIndex].Include = finalList
		return nil
	case Exclude:
		finalList := strings.ReplaceAll(c.FilterConfigs[pathFilterIndex].Exclude, ","+diskPath, "")
		finalList = strings.ReplaceAll(finalList, diskPath, "")
		c.FilterConfigs[pathFilterIndex].Exclude = finalList
		return nil
	default:
		return errors.New("invalid filterconfig path-filter list name")
	}
}

func (c *Config) GetConfigYaml() (string, error) {
	if c == nil {
		return "", errors.New("Config is nil")
	}

	yml, err := yaml.Marshal(c)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal NDM Config to YAML")
	}

	return string(yml), nil
}
