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

package v1alpha1

// PoolMetricsList is used to store the collected metrics
type PoolMetricsList []PoolMetrics

// PoolMetrics stores the pool metrics proided by the maya-exporter
type PoolMetrics struct {
	Name   string          `json:"name"`
	Help   string          `json:"help"`
	Type   int             `json:"type"`
	Metric []MetricsFamily `json:"metric"`
}

// PoolStats ...
type PoolStats struct {
}

// ToMap converts metrics to map[string]MetricsFamily
func (pml PoolMetricsList) ToMap() map[string]MetricsFamily {
	newMetrics := make(map[string]MetricsFamily)
	for _, metric := range pml {
		if len(metric.Metric) == 0 {
			newMetrics[metric.Name] = MetricsFamily{}
		} else {
			newMetrics[metric.Name] = metric.Metric[0]
		}
	}
	return newMetrics
}

// GetValue returns the value of the key if the key is present in map[string]MetricsFamily.
func GetValue(key string, m map[string]MetricsFamily) float64 {
	if val, p := m[key]; p {
		return val.Gauge.Value
	}
	return -1
}
