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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=blockDevice
// +k8s:openapi-gen=true

// BlockDevice is the Schema used to represent a BlockDevice CR
type BlockDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceSpec   `json:"spec,omitempty"`
	Status DeviceStatus `json:"status,omitempty"`
}

// DeviceSpec defines the properties and runtime status of a BlockDevice
type DeviceSpec struct {
	// NodeAttributes has the details of the node on which BD is attached
	NodeAttributes NodeAttribute `json:"nodeAttributes"`

	// Path contain devpath (e.g. /dev/sdb)
	Path string `json:"path"`

	// Capacity
	Capacity DeviceCapacity `json:"capacity"`

	// Details contain static attributes of BD like model,serial, and so forth
	Details DeviceDetails `json:"details"`

	// Reference to the BDC which has claimed this BD
	ClaimRef *v1.ObjectReference `json:"claimRef,omitempty"`

	// DevLinks contains soft links of a block device like
	// /dev/by-id/...
	// /dev/by-uuid/...
	DevLinks []DeviceDevLink `json:"devlinks"`

	// FileSystem contains mountpoint and filesystem type
	FileSystem FileSystemInfo `json:"filesystem,omitempty"`

	// Partitioned represents if BlockDevice has partions or not (YES/NO)
	// Currently always default to NO.
	// TODO @kmova to be implemented/deprecated
	Partitioned string `json:"partitioned"`

	// ParentDevice was intented to store the UUID of the parent
	// Block Device as is the case for partitioned block devices.
	//
	// For example:
	// /dev/sda is the parent for /dev/sdap1
	// TODO @kmova to be implemented/deprecated
	ParentDevice string `json:"parentDevice,omitempty"`

	// AggregateDevice was intended to store the hierachical
	// information in cases of LVM. However this is currently
	// not implemented and may need to be re-looked into for
	// better design.
	// TODO @kmova to be implemented/deprecated
	AggregateDevice string `json:"aggregateDevice,omitempty"`
}

// NodeAttribute defines the attributes of a node where
// the block device is attached.
//
// Note: Prior to introducing NodeAttributes, the BD would
// only support gathering hostname and add it as a label
// to the BD resource.
//
// In some use cases, the caller has access only to node name, not
// the hostname. node name and hostname are different in certain
// Kubernetes clusters.
//
// NodeAttributes is added to contain attributes that are not
// available on the labels like - node name, uuid, etc.
//
// The node attributes are helpful in querying for block devices
// based on node attributes.  Also, adding this in the spec allows for
// displaying in node name in the `kubectl get bd`
//
// TODO  @kmova @akhil
// Capture and add nodeUUID to BD, that can help in determining
// if the node was recreated with same node name.
type NodeAttribute struct {
	// NodeName is the name of the Kubernetes node resource on which the device is attached
	NodeName string `json:"nodeName"`
}

// DeviceCapacity defines the physical and logical size of the block device
type DeviceCapacity struct {
	// Storage is the blockdevice capacity in bytes
	Storage uint64 `json:"storage"`

	// PhysicalSectorSize is blockdevice physical-Sector size in bytes
	PhysicalSectorSize uint32 `json:"physicalSectorSize"`

	// LogicalSectorSize is blockdevice logical-sector size in bytes
	LogicalSectorSize uint32 `json:"logicalSectorSize"`
}

// DeviceDetails represent certain hardware/static attributes of the block device
type DeviceDetails struct {
	// DeviceType represents the type of device like
	// sparse, disk, partition, lvm, raid
	DeviceType string `json:"deviceType"`

	// DriveType is the type of backing drive, HDD/SSD
	DriveType string `json:"driveType"`

	// LogicalBlockSize is the logical block size in bytes
	// reported by /sys/class/block/sda/queue/logical_block_size
	LogicalBlockSize uint32 `json:"logicalBlockSize"`

	// PhysicalBlockSize is the physical block size in bytes
	// reported by /sys/class/block/sda/queue/physical_block_size
	PhysicalBlockSize uint32 `json:"physicalBlockSize"`

	// HardwareSectorSize is the hardware sector size in bytes
	HardwareSectorSize uint32 `json:"hardwareSectorSize"`

	// Model is model of disk
	Model string `json:"model"`

	// Compliance is standards/specifications version implmented by device firmware
	//  such as SPC-1, SPC-2, etc
	Compliance string `json:"compliance"`

	// Serial is serial number of disk
	Serial string `json:"serial"`

	// Vendor is vendor of disk
	Vendor string `json:"vendor"`

	// disk firmware revision
	FirmwareRevision string `json:"firmwareRevision"`
}

// FileSystemInfo defines the filesystem type and mountpoint of the device if it exists
type FileSystemInfo struct {
	//Type represents the FileSystem type of the block device
	Type string `json:"fsType,omitempty"`

	//MountPoint represents the mountpoint of the block device.
	Mountpoint string `json:"mountPoint,omitempty"`
}

// DeviceDevLink holds the maping between type and links like by-id type or by-path type link
type DeviceDevLink struct {
	// Kind is the type of link like by-id or by-path.
	Kind string `json:"kind,omitempty"`

	// Links are the soft links of Type type
	Links []string `json:"links,omitempty"`
}

// DeviceStatus defines the observed state of BlockDevice
type DeviceStatus struct {
	// claim state of the block device
	ClaimState DeviceClaimState `json:"claimState"`

	// State is the current state of the blockdevice (Active/Inactive)
	State BlockDeviceState `json:"state"`
}

// DeviceClaimState defines the observed state of BlockDevice
type DeviceClaimState string

const (
	// BlockDeviceUnclaimed represents that the block device is not bound to any BDC,
	// all cleanup jobs have been completed and is available for claiming.
	BlockDeviceUnclaimed DeviceClaimState = "Unclaimed"

	// BlockDeviceReleased represents that the block device is released from the BDC,
	// pending cleanup jobs
	BlockDeviceReleased DeviceClaimState = "Released"

	// BlockDeviceClaimed represents that the block device is bound to a BDC
	BlockDeviceClaimed DeviceClaimState = "Claimed"
)

// BlockDeviceState defines the observed state of the disk
type BlockDeviceState string

const (
	// BlockDeviceActive is the state for a block device that is connected to the node
	BlockDeviceActive BlockDeviceState = "Active"

	// BlockDeviceInactive is the state for a block device that is disconnected from a node
	BlockDeviceInactive BlockDeviceState = "Inactive"

	// BlockDeviceUnknown is the state for a block device whose state (attached/detached) cannot
	// be determined at this time.
	BlockDeviceUnknown BlockDeviceState = "Unknown"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=blockDeviceList
// +k8s:openapi-gen=true

// BlockDeviceList contains a list of BlockDevice
type BlockDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BlockDevice `json:"items"`
}
