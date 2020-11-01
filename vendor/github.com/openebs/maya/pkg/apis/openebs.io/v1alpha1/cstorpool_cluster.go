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

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PoolType is a label for the pool type of a cStor pool.
type PoolType string

// These are the valid pool types of cStor Pool.
const (
	// PoolStriped is the striped raid group.
	PoolStriped PoolType = "stripe"
	// PoolMirrored is the mirror raid group.
	PoolMirrored PoolType = "mirror"
	// PoolRaidz is the raidz raid group.
	PoolRaidz PoolType = "raidz"
	// PoolRaidz2 is the raidz2 raid group.
	PoolRaidz2 PoolType = "raidz2"
)

var (
	// SupportedPRaidType is a map holding the supported raid configurations
	// Value of the keys --
	// 1. In case of striped this is the minimum number of disk required.
	// 2. In all other cases this is the exact number of disks required.
	SupportedPRaidType = map[PoolType]int{
		PoolStriped:  1,
		PoolMirrored: 2,
		PoolRaidz:    3,
		PoolRaidz2:   6,
	}
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorpoolcluster

// CStorPoolCluster describes a CStorPoolCluster custom resource.
type CStorPoolCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorPoolClusterSpec   `json:"spec"`
	Status            CStorPoolClusterStatus `json:"status"`
	VersionDetails    VersionDetails         `json:"versionDetails"`
}

// CStorPoolClusterSpec is the spec for a CStorPoolClusterSpec resource
type CStorPoolClusterSpec struct {
	// Pools is the spec for pools for various nodes
	// where it should be created.
	Pools []PoolSpec `json:"pools"`
	// DefaultResources are the compute resources required by the cstor-pool
	// container.
	// If the resources at PoolConfig is not specified, this is written
	// to CSPI PoolConfig.
	DefaultResources *corev1.ResourceRequirements `json:"resources"`
	// AuxResources are the compute resources required by the cstor-pool pod
	// side car containers.
	DefaultAuxResources *corev1.ResourceRequirements `json:"auxResources"`
	// Tolerations, if specified, are the pool pod's tolerations
	// If tolerations at PoolConfig is empty, this is written to
	// CSPI PoolConfig.
	Tolerations []corev1.Toleration `json:"tolerations"`

	// DefaultPriorityClassName if specified applies to all the pool pods
	// in the pool spec if the priorityClass at the pool level is
	// not specified.
	DefaultPriorityClassName string `json:"priorityClassName"`
}

//PoolSpec is the spec for pool on node where it should be created.
type PoolSpec struct {
	// NodeSelector is the labels that will be used to select
	// a node for pool provisioning.
	// Required field
	NodeSelector map[string]string `json:"nodeSelector"`
	// RaidConfig is the raid group configuration for the given pool.
	RaidGroups []RaidGroup `json:"raidGroups"`
	// PoolConfig is the default pool config that applies to the
	// pool on node.
	PoolConfig PoolConfig `json:"poolConfig"`
}

// PoolConfig is the default pool config that applies to the
// pool on node.
type PoolConfig struct {
	// Cachefile is used for faster pool imports
	// optional -- if not specified or left empty cache file is not
	// used.
	CacheFile string `json:"cacheFile"`
	// DefaultRaidGroupType is the default raid type which applies
	// to all the pools if raid type is not specified there
	// Compulsory field if any raidGroup is not given Type
	DefaultRaidGroupType string `json:"defaultRaidGroupType"`

	// OverProvisioning to enable over provisioning
	// Optional -- defaults to false
	OverProvisioning bool `json:"overProvisioning"`
	// Compression to enable compression
	// Optional -- defaults to off
	// Possible values : lz, off
	Compression string `json:"compression"`
	// Resources are the compute resources required by the cstor-pool
	// container.
	Resources *corev1.ResourceRequirements `json:"resources"`
	// AuxResources are the compute resources required by the cstor-pool pod
	// side car containers.
	AuxResources *corev1.ResourceRequirements `json:"auxResources"`
	// Tolerations, if specified, the pool pod's tolerations.
	Tolerations []corev1.Toleration `json:"tolerations"`

	// PriorityClassName if specified applies to this pool pod
	// If left empty, DefaultPriorityClassName is applied.
	// (See CStorPoolClusterSpec.DefaultPriorityClassName)
	// If both are empty, not priority class is applied.
	PriorityClassName string `json:"priorityClassName"`
}

// RaidGroup contains the details of a raid group for the pool
type RaidGroup struct {
	// Type is the raid group type
	// Supported values are : stripe, mirror, raidz and raidz2

	// stripe -- stripe is a raid group which divides data into blocks and
	// spreads the data blocks across multiple block devices.

	// mirror -- mirror is a raid group which does redundancy
	// across multiple block devices.

	// raidz -- RAID-Z is a data/parity distribution scheme like RAID-5, but uses dynamic stripe width.
	// radiz2 -- TODO
	// Optional -- defaults to `defaultRaidGroupType` present in `PoolConfig`
	Type string `json:"type"`
	// IsWriteCache is to enable this group as a write cache.
	IsWriteCache bool `json:"isWriteCache"`
	// IsSpare is to declare this group as spare which will be
	// part of the pool that can be used if some block devices
	// fail.
	IsSpare bool `json:"isSpare"`
	// IsReadCache is to enable this group as read cache.
	IsReadCache bool `json:"isReadCache"`
	// BlockDevices contains a list of block devices that
	// constitute this raid group.
	BlockDevices []CStorPoolClusterBlockDevice `json:"blockDevices"`
}

// CStorPoolClusterBlockDevice contains the details of block devices that
// constitutes a raid group.
type CStorPoolClusterBlockDevice struct {
	// BlockDeviceName is the name of the block device.
	BlockDeviceName string `json:"blockDeviceName"`
	// Capacity is the capacity of the block device.
	// It is system generated
	Capacity string `json:"capacity"`
	// DevLink is the dev link for block devices
	DevLink string `json:"devLink"`
}

// CStorPoolClusterStatus is for handling status of pool.
type CStorPoolClusterStatus struct {
	Phase    string                `json:"phase"`
	Capacity CStorPoolCapacityAttr `json:"capacity"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorpoolclusters

// CStorPoolClusterList is a list of CStorPoolCluster resources
type CStorPoolClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorPoolCluster `json:"items"`
}
