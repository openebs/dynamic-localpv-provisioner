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

package v1alpha2

import (
	ndm "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
)

//TODO: While using these packages UnitTest
//must be written to corresponding function

// DefaultBlockDeviceCount is a map containing the
// default block device count of various raid types.
var DefaultBlockDeviceCount = map[string]int{
	string(apis.PoolTypeMirroredCPV): int(apis.MirroredBlockDeviceCountCPV),
	string(apis.PoolTypeStripedCPV):  int(apis.StripedBlockDeviceCountCPV),
	string(apis.PoolTypeRaidzCPV):    int(apis.RaidzBlockDeviceCountCPV),
	string(apis.PoolTypeRaidz2CPV):   int(apis.Raidz2BlockDeviceCountCPV),
}

// BlockDevice encapsulates BlockDevice api object.
type BlockDevice struct {
	// actual block device object
	Object *ndm.BlockDevice
}

// BlockDeviceList holds the list of BlockDevice api
type BlockDeviceList struct {
	// list of blockdevices
	ObjectList *ndm.BlockDeviceList
}

// Predicate defines an abstraction to determine conditional checks against the
// provided block device instance
type Predicate func(*BlockDevice) bool

// predicateList holds the list of Predicates
type predicateList []Predicate

// all returns true if all the predicates succeed against the provided block
// device instance.
func (l predicateList) all(c *BlockDevice) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// Validate defines an abstraction to determine conditional check against the
// provided block device
type Validate func(*BlockDevice) error

// HasAnnotation is predicate to filter out based on
// annotation in BDC instances
func HasAnnotation(key, value string) Predicate {
	return func(bd *BlockDevice) bool {
		return bd.HasAnnotation(key, value)
	}
}

// HasAnnotation return true if provided annotation
// key and value are present in the the provided block device List
// instance
func (bd *BlockDevice) HasAnnotation(key, value string) bool {
	val, ok := bd.Object.GetAnnotations()[key]
	if ok {
		return val == value
	}
	return false
}

// IsSparse filters the block device based on type of the disk
func IsSparse() Predicate {
	return func(bd *BlockDevice) bool {
		return bd.IsSparse()
	}
}

// IsSparse returns true if the block device is of sparse type
func (bd *BlockDevice) IsSparse() bool {
	return bd.Object.Spec.Details.DeviceType == string(apis.TypeBlockDeviceCPV)
}

// IsActive filters the block device based on the active status
func IsActive() Predicate {
	return func(bd *BlockDevice) bool {
		return bd.IsActive()
	}
}

// IsActive returns true if the block device is active.
func (bd *BlockDevice) IsActive() bool {
	return bd.Object.Status.State == ndm.BlockDeviceActive
}

// IsUnclaimed filters the block device based on unclaimed status
func IsUnclaimed() Predicate {
	return func(bd *BlockDevice) bool {
		return bd.IsUnclaimed()
	}
}

// IsUnclaimed returns true if the block device is unclaimed
func (bd *BlockDevice) IsUnclaimed() bool {
	return bd.Object.Status.ClaimState == ndm.BlockDeviceUnclaimed
}

// IsClaimed filters the block deive based on claimed status
func IsClaimed() Predicate {
	return func(bd *BlockDevice) bool {
		return bd.IsClaimed()
	}
}

// IsClaimed returns true if the block device is claimed
func (bd *BlockDevice) IsClaimed() bool {
	return bd.Object.Status.ClaimState == ndm.BlockDeviceClaimed
}

// HasFileSystem returns true if the block device has filesystem
func (bd *BlockDevice) HasFileSystem() bool {
	return bd.Object.Spec.FileSystem.Type != ""
}

// IsUsable filters the block device based on usage of disk
func IsUsable(usedBlockDevices map[string]int) Predicate {
	return func(bd *BlockDevice) bool {
		return bd.IsUsable(usedBlockDevices)
	}
}

// IsUsable returns true if this block device
// can be used for pool provisioning.
// The argument usedBlockDevice is a map containing
// key as block device cr name and value as integer.
// If the value of map is greater than 0 ,
// then this corresponding block device is not usable.
func (bd *BlockDevice) IsUsable(usedBD map[string]int) bool {
	return usedBD[bd.Object.Name] == 0
}

// IsUsableNode filters the block device based on usage of node
func IsUsableNode(usedNodes map[string]bool) Predicate {
	return func(bd *BlockDevice) bool {
		return bd.IsUsableNode(usedNodes)
	}
}

// IsUsableNode returns true if block device of this node can be used
// for pool provisioning. The argument usedNodes is a map containing
// key as node name and value as bool. If the value of map is greater
// than false, then this corresponding node is not usable.
func (bd *BlockDevice) IsUsableNode(usedNodes map[string]bool) bool {
	return !usedNodes[bd.GetNodeName()]
}

// IsBelongToNode returns true if the block device belongs to the provided node.
func IsBelongToNode(nodeName string) Predicate {
	return func(bd *BlockDevice) bool {
		return bd.IsBelongToNode(nodeName)
	}
}

// IsBelongToNode returns true if the block device belongs to the provided node.
func (bd *BlockDevice) IsBelongToNode(nodeName string) bool {
	return bd.GetNodeName() == nodeName
}

// IsValidPoolTopology returns true if the block device count
// is multiples of default block device count of various raid types
func IsValidPoolTopology(poolType string, bdCount int) bool {
	return DefaultBlockDeviceCount[poolType]%bdCount == 0
}

// GetNodeName returns the node name to which the block device is attached
func (bd *BlockDevice) GetNodeName() string {
	return bd.Object.Spec.NodeAttributes.NodeName
}

// CheckIfBDIsActive validates the block device based on status
func CheckIfBDIsActive() Validate {
	return func(bd *BlockDevice) error {
		return bd.CheckIfBDIsActive()
	}
}

// CheckIfBDIsActive returns error only when block device presents in other than
// active state or else return nil
func (bd *BlockDevice) CheckIfBDIsActive() error {
	if !bd.IsActive() {
		return errors.Errorf(
			"block device is in not in active state",
		)
	}
	return nil
}

// CheckIfBDBelongsToNode validates the block device based on nodeName provided
// via argument
func CheckIfBDBelongsToNode(nodeName string) Validate {
	return func(bd *BlockDevice) error {
		return bd.CheckIfBDBelongsToNode(nodeName)
	}
}

// CheckIfBDBelongsToNode returns error only when the block device node name
// doesn't matched with provided node name
func (bd *BlockDevice) CheckIfBDBelongsToNode(nodeName string) error {
	if !bd.IsBelongToNode(nodeName) {
		return errors.Errorf(
			"block device doesn't belongs to node %s",
			bd.Object.Spec.NodeAttributes.NodeName,
		)
	}
	return nil
}

// CheckIfBDIsNonFsType validates the block device based on filesystem type
func CheckIfBDIsNonFsType() Validate {
	return func(bd *BlockDevice) error {
		return bd.CheckIfBDIsNonFsType()
	}
}

// CheckIfBDIsNonFsType return error only when the block device has filesystem
func (bd *BlockDevice) CheckIfBDIsNonFsType() error {
	if bd.HasFileSystem() {
		return errors.Errorf("block device has file system {%s}",
			bd.Object.Spec.FileSystem.Type,
		)
	}
	return nil
}

// ValidateBlockDevice validates the block device based on the arguments
// provided and returns error if validation fails
func (bd *BlockDevice) ValidateBlockDevice(v ...Validate) error {
	var err error
	for _, validate := range v {
		if err = validate(bd); err != nil {
			return errors.Wrapf(
				err,
				"block device %s validation failed",
				bd.Object.Name,
			)
		}
	}
	return nil
}

// Filter will filter the block device instances if all the predicates succeed
// against that block device.
func (l *BlockDeviceList) Filter(p ...Predicate) *BlockDeviceList {
	var plist predicateList
	plist = append(plist, p...)
	if len(plist) == 0 {
		return l
	}

	filtered := NewListBuilder().List()
	for _, bdAPI := range l.ObjectList.Items {
		bdAPI := bdAPI // pin it
		BlockDevice := BuilderForAPIObject(&bdAPI).BlockDevice
		if plist.all(BlockDevice) {
			filtered.ObjectList.Items = append(
				filtered.ObjectList.Items,
				*BlockDevice.Object)
		}
	}
	return filtered
}

// GetDeviceID returns the device link of the block device.
// If device link is not found it returns device path.
// For a cstor pool creation -- this link or path is used.
// For convenience, we call it as device ID.
// Hence, device ID can either be a  device link or device path
// depending on what was available in block device cr.
func (bd *BlockDevice) GetDeviceID() string {
	deviceID := bd.GetLink()
	if deviceID != "" {
		return deviceID
	}
	return bd.GetPath()
}

// GetLink returns the link of the block device
// if present else return empty string
func (bd *BlockDevice) GetLink() string {
	if len(bd.Object.Spec.DevLinks) != 0 &&
		len(bd.Object.Spec.DevLinks[0].Links) != 0 {
		return bd.Object.Spec.DevLinks[0].Links[0]
	}
	return ""
}

// GetPath returns path of the block device
func (bd *BlockDevice) GetPath() string {
	return bd.Object.Spec.Path
}

// Len returns the length og BlockDeviceList.
func (l *BlockDeviceList) Len() int {
	return len(l.ObjectList.Items)
}
