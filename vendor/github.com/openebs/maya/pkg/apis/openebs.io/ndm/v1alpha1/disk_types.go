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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DiskSpec defines the desired state of Disk
type DiskSpec struct {
	Path             string         `json:"path"`                       //Path contain devpath (e.g. /dev/sdb)
	Capacity         DiskCapacity   `json:"capacity"`                   //Capacity
	Details          DiskDetails    `json:"details"`                    //Details contains static attributes (model, serial ..)
	DevLinks         []DiskDevLink  `json:"devlinks"`                   //DevLinks contains soft links of one disk
	FileSystem       FileSystemInfo `json:"fileSystem,omitempty"`       //Contains the data about filesystem on the disk
	PartitionDetails []Partition    `json:"partitionDetails,omitempty"` //Details of partitions in the disk (filesystem, partition type)
}

// DiskCapacity represents the capacity of the disk in bytes.
// Size of physical sector and logical sector is also defined.
type DiskCapacity struct {
	Storage            uint64 `json:"storage"`            // disk size in bytes
	PhysicalSectorSize uint32 `json:"physicalSectorSize"` // disk physical-Sector size in bytes
	LogicalSectorSize  uint32 `json:"logicalSectorSize"`  // disk logical-sector size in bytes
}

// DiskDetails defines the various physical attributes of disk like rotation rate,
// firmware revision, type of Disk etc.
type DiskDetails struct {
	RotationRate     uint16 `json:"rotationRate"`     // Disk rotation speed if disk is not SSD
	DriveType        string `json:"driveType"`        // DriveType represents the type of drive like SSD, HDD etc.,
	Model            string `json:"model"`            // Model is model of disk
	Compliance       string `json:"compliance"`       // Implemented standards/specifications version such as SPC-1, SPC-2, etc
	Serial           string `json:"serial"`           // Serial is serial no of disk
	Vendor           string `json:"vendor"`           // Vendor is vendor of disk
	FirmwareRevision string `json:"firmwareRevision"` // disk firmware revision
}

// DiskDevLink holds the maping between type and links like by-id type or by-path type link
type DiskDevLink struct {
	Kind  string   `json:"kind"`  // Kind is the type of link like by-id or by-path.
	Links []string `json:"links"` // Links are the soft links of Type type
}

// Partition represents the partition information of the disk
type Partition struct {
	PartitionType string         `json:"partitionType"`
	FileSystem    FileSystemInfo `json:"fileSystem,omitempty"`
}

// DiskStatus defines the observed state of Disk
type DiskStatus struct {
	State string `json:"state"` //current state of the disk (Active/Inactive)
}

// Temperature info provided by the disk. The fields will be filled only if it is possible
// to get valid temperature data from the disk.
type Temperature struct {
	CurrentTemperature int16 `json:"currentTemperature"`
	HighestTemperature int16 `json:"highestTemperature"`
	LowestTemperature  int16 `json:"lowestTemperature"`
}

// DiskStat represents the statistics data related to the Disk. The data represented by all
// these fields will be continuously changing. Currently its just updated once, when the CR
// is created.
type DiskStat struct {
	TempInfo              Temperature `json:"diskTemperature"`
	TotalBytesRead        uint64      `json:"totalBytesRead"`
	TotalBytesWritten     uint64      `json:"totalBytesWritten"`
	DeviceUtilizationRate float64     `json:"deviceUtilizationRate"`
	PercentEnduranceUsed  float64     `json:"percentEnduranceUsed"`
}

// DeviceInfo holds the UID of the Block Device which is backed by this Disk.
type DeviceInfo struct {
	DeviceUID string `json:"blockDeviceUID"` //Cross reference to BlockDevice CR backed by this disk
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=disk
// +k8s:openapi-gen=true

// Disk is the Schema for the disks API
type Disk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiskSpec   `json:"spec,omitempty"`
	Status DiskStatus `json:"status,omitempty"`
	Stats  DiskStat   `json:"stats,omitempty"`
	Device DeviceInfo `json:"deviceInfo"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=diskList
// +k8s:openapi-gen=true

// DiskList contains a list of Disk
type DiskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Disk `json:"items"`
}
