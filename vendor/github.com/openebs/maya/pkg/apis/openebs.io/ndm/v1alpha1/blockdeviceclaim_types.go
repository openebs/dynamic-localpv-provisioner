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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=blockDeviceClaim
// +k8s:openapi-gen=true

// BlockDeviceClaim is the Schema for the BlockDeviceClaim CR
type BlockDeviceClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceClaimSpec   `json:"spec,omitempty"`
	Status DeviceClaimStatus `json:"status,omitempty"`
}

// DeviceClaimSpec defines the request details for a BlockDevice
type DeviceClaimSpec struct {

	// Selector is used to find block devices to be considered for claiming
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Resources will help with placing claims on Capacity, IOPS
	Resources DeviceClaimResources `json:"resources"`

	// DeviceType represents the type of drive like SSD, HDD etc.,
	DeviceType string `json:"deviceType"`

	// HostName from where blockdevice has to be claimed.
	HostName string `json:"hostName"`

	// Details of the device to be claimed
	Details DeviceClaimDetails `json:"deviceClaimDetails,omitempty"`

	// BlockDeviceName is the reference to the block-device backing this claim
	BlockDeviceName string `json:"blockDeviceName,omitempty"`

	// BlockDeviceNodeAttributes is the attributes on the node from which a BD should
	// be selected for this claim. It can include nodename, failure domain etc.
	BlockDeviceNodeAttributes BlockDeviceNodeAttributes `json:"blockDeviceNodeAttributes,omitempty"`
}

// DeviceClaimStatus defines the observed state of BlockDeviceClaim
type DeviceClaimStatus struct {
	Phase DeviceClaimPhase `json:"phase"`
}

// DeviceClaimPhase is a typed string for phase field of BlockDeviceClaim.
type DeviceClaimPhase string

// BlockDeviceClaim CR, when created pass through phases before it got some Devices Assigned.
// Given below table, have all phases which BlockDeviceClaim CR can go before it is marked done.
const (
	// BlockDeviceClaimStatusEmpty represents that the BlockDeviceClaim was just created.
	BlockDeviceClaimStatusEmpty DeviceClaimPhase = ""

	// BlockDeviceClaimStatusPending represents BlockDeviceClaim has not been assigned devices yet. Rather
	// search is going on for matching devices.
	BlockDeviceClaimStatusPending DeviceClaimPhase = "Pending"

	// BlockDeviceClaimStatusInvalidCapacity represents BlockDeviceClaim has invalid capacity request i.e. 0/-1
	BlockDeviceClaimStatusInvalidCapacity DeviceClaimPhase = "Invalid Capacity Request"

	// BlockDeviceClaimStatusDone represents BlockDeviceClaim has been assigned backing blockdevice and ready for use.
	BlockDeviceClaimStatusDone DeviceClaimPhase = "Bound"
)

// DeviceClaimResources defines the request by the claim, eg, Capacity, IOPS
type DeviceClaimResources struct {
	// Requests describes the minimum resources required. eg: if storage resource of 10G is
	// requested minimum capacity of 10G should be available
	Requests corev1.ResourceList `json:"requests"`
}

const (
	// ResourceStorage defines the storage required as v1.Quantity
	ResourceStorage corev1.ResourceName = "storage"
)

// DeviceClaimDetails defines the details of the block device that should be claimed
type DeviceClaimDetails struct {
	// BlockVolumeMode represents whether to claim a device in Block mode or Filesystem mode.
	// These are use cases of BlockVolumeMode:
	// 1) Not specified: VolumeMode check will not be effective
	// 2) VolumeModeBlock: BD should not have any filesystem or mountpoint
	// 3) VolumeModeFileSystem: BD should have a filesystem and mountpoint. If DeviceFormat is
	//    specified then the format should match with the FSType in BD
	BlockVolumeMode BlockDeviceVolumeMode `json:"blockVolumeMode,omitempty"`

	//Format of the device required, eg:ext4, xfs
	DeviceFormat string `json:"formatType,omitempty"`

	//AllowPartition represents whether to claim a full block device or a device that is a partition
	AllowPartition bool `json:"allowPartition,omitempty"`
}

// BlockDeviceVolumeMode specifies the type in which the BlockDevice can be used
type BlockDeviceVolumeMode string

const (
	// VolumeModeBlock specifies that the block device needs to be used as a raw block
	VolumeModeBlock BlockDeviceVolumeMode = "Block"

	// VolumeModeFileSystem specifies that block device will be used with a filesystem
	// already existing
	VolumeModeFileSystem BlockDeviceVolumeMode = "FileSystem"
)

// BlockDeviceNodeAttributes contains the attributes of the node from which the BD should
// be selected for claiming. A BDC can specify one or more attributes. When multiple values
// are specified, the NDM Operator will claim a Block Device that matches all
// the requested attributes.
type BlockDeviceNodeAttributes struct {
	// NodeName represents the name of the Kubernetes node resource
	// where the BD should be present
	NodeName string `json:"nodeName,omitempty"`

	// HostName represents the hostname of the Kubernetes node resource
	// where the BD should be present
	HostName string `json:"hostName,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=blockDeviceClaimList
// +k8s:openapi-gen=true

// BlockDeviceClaimList contains a list of BlockDeviceClaim
type BlockDeviceClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BlockDeviceClaim `json:"items"`
}
