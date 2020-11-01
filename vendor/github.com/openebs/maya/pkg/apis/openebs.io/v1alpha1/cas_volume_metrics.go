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

// VolumeMetricsList is used to store the collected metrics
// all the stats exposed by jiva stored into OpenEBSVolumeMetrics fields
type VolumeMetricsList []VolumeMetrics

// VolumeMetrics stores the volume metrics proided by the prometheus
type VolumeMetrics struct {
	Name   string          `json:"name"`
	Help   string          `json:"help"`
	Type   int             `json:"type"`
	Metric []MetricsFamily `json:"metric"`
}

// MetricsFamily is used store the prometheus metric members
type MetricsFamily struct {
	Label   []LabelItem `json:"label"`
	Counter Counter     `json:"counter"`
	Summary Summary     `json:"summary"`
	Gauge   Gauge       `json:"gauge"`
}

// LabelItem stores the labels provided by prometheus used for identifying the items
type LabelItem struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Counter stores the counters provided by the prometheus and can be used for counting events that happen (e.g. total number of requests) and query using rate()
type Counter struct {
	Value float64 `json:"value"`
}

// Summary stores the summary provided by prometheus which can be used for pre-calculated quantiles on client side, but be mindful of calculation cost and aggregation limitations
type Summary struct {
	SampleCount float64    `json:"sample_count"`
	SampleSum   float64    `json:"sample_sum"`
	Quantile    []Quantile `json:"quantile"`
}

// Quantile stores the quantile provided by the prometheus
type Quantile struct {
	Quantile float64 `json:"quantile"`
	Value    float64 `json:"value"`
}

// Gauge stores the gauge provided by the prometheus which can be used to instrument the current state of a metric
type Gauge struct {
	Value float64 `json:"value"`
}

// StatsJSON stores the statistics of an iSCSI volume.
type StatsJSON struct {
	IQN     string `json:"Iqn"`
	Volume  string `json:"Volume"`
	Portal  string `json:"Portal"`
	Size    string `json:"Size"`
	CASType string `json:"CASType"`

	ReadIOPS  int64 `json:"ReadIOPS"`
	WriteIOPS int64 `json:"WriteIOPS"`

	ReadThroughput  float64 `json:"ReadThroughput"`
	WriteThroughput float64 `json:"WriteThroughput"`

	ReadLatency  float64 `json:"ReadLatency"`
	WriteLatency float64 `json:"WriteLatency"`

	AvgReadBlockSize  int64 `json:"AvgReadBlockSize"`
	AvgWriteBlockSize int64 `json:"AvgWriteBlockSize"`

	SectorSize  float64 `json:"SectorSize"`
	ActualUsed  float64 `json:"ActualUsed"`
	LogicalSize float64 `json:"LogicalSize"`
}
