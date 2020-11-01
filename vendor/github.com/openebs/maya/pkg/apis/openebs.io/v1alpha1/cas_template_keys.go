/*
Copyright 2017 The OpenEBS Authors.

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

// TopLevelProperty represents the top level property that
// is a starting point to represent a hierarchical chain of
// properties.
//
// e.g.
// Config.prop1.subprop1 = val1
// Config.prop1.subprop2 = val2
// In above example Config is a top level object
//
// NOTE:
//  The value of any hierarchical chain of properties
// can be parsed via dot notation
type TopLevelProperty string

const (
	// CASTOptionsTLP is a top level property supported by CAS template engine.
	// CAS template specific options are placed here
	CASTOptionsTLP TopLevelProperty = "CAST"

	// ConfigTLP is a top level property supported by CAS template engine
	//
	// The policy specific properties are placed with ConfigTLP as the
	// top level property
	ConfigTLP TopLevelProperty = "Config"

	// VolumeTLP is a top level property supported by CAS template engine
	//
	// The properties provided by the caller are placed with VolumeTLP
	// as the top level property
	//
	// NOTE:
	//  CAS template engine cannot modify these properties. These are the
	// runtime properties that are provided as inputs to CAS template
	// engine.
	VolumeTLP TopLevelProperty = "Volume"

	// SnapshotTLP is a top level property supported by CAS template engine
	//
	// The properties provided by the caller are placed with SnapshotTLP
	// as the top level property
	//
	// NOTE:
	//  CAS template engine cannot modify these properties. These are the
	// runtime properties that are provided as inputs to CAS template
	// engine.
	SnapshotTLP TopLevelProperty = "Snapshot"

	// StoragePoolTLP is a top level property supported by CAS template engine
	//
	// The properties provided by the caller are placed with StoragePoolTLP
	// as the top level property
	//
	// NOTE:
	//  CAS template engine cannot modify these properties. These are the
	// runtime properties that are provided as inputs to CAS template
	// engine.
	StoragePoolTLP TopLevelProperty = "Storagepool"

	// TaskResultTLP is a top level property supported by CAS template engine
	//
	// The specific results after the execution of a task are placed with
	// TaskResultTLP as the top level property
	//
	// NOTE:
	//  This is typically used to feed inputs of a task's execution
	// result to **next task** before the later's execution
	TaskResultTLP TopLevelProperty = "TaskResult"

	// CurrentJSONResultTLP is a top level property supported by CAS template engine
	// The result of the current task's execution is stored in this top
	// level property.
	CurrentJSONResultTLP TopLevelProperty = "JsonResult"

	// ListItemsTLP is a top level property supported by CAS template engine
	//
	// Results of one or more tasks' execution can be saved in this property.
	//
	// Example:
	//  Below shows how specific properties of a list of items can be retrieved in
	// a go template. Below dot notation is for illustration purposes and only
	// reflects the way the specific property value was set.
	//
	// {{- .ListItems.volumes.default.mypv2.ip -}}
	// {{- .ListItems.volumes.default.mypv2.status -}}
	// {{- .ListItems.volumes.openebs.mypv.ip -}}
	// {{- .ListItems.volumes.openebs.mypv.status -}}
	ListItemsTLP TopLevelProperty = "ListItems"
)

// StoragePoolTLPProperty is used to define properties that comes
// after StoragePoolTLP
type StoragePoolTLPProperty string

const (
	// OwnerCTP indicates the owner of this pool; the one who
	// is executing this policy
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Storagepool.owner }}
	OwnerCTP StoragePoolTLPProperty = "owner"

	// BlockDeviceListCTP indicates the list of block devices
	BlockDeviceListCTP StoragePoolTLPProperty = "blockDeviceList"
	// BlockDeviceIDListCTP indicates the list of device id of disks
	// If device is is not present for a disk, it contians the device path.
	BlockDeviceIDListCTP StoragePoolTLPProperty = "blockDeviceIdList"
	// NodeNameCTP is the name of node where cstorpool will be ceated
	NodeNameCTP StoragePoolTLPProperty = "nodeName"
	// PoolTypeCTP is the type of cstorpool that will be ceated
	PoolTypeCTP StoragePoolTLPProperty = "poolType"
	// InitPhaseCTP is the init phase for the cstorpool CR that will be ceated
	InitPhaseCTP StoragePoolTLPProperty = "phase"
	// PoolCacheFileCTP is the cache file used in case of imporitng pool
	PoolCacheFileCTP StoragePoolTLPProperty = "poolCacheFile"
	// PoolROThresholdLimitCTP is pool read only threshold limit
	PoolROThresholdLimitCTP StoragePoolTLPProperty = "poolROThreshold"
)

// VolumeTLPProperty is used to define properties that comes
// after VolumeTLP
type VolumeTLPProperty string

const (
	// OwnerVTP indicates the owner of this volume; the one who
	// is executing this policy
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Volume.owner }}
	OwnerVTP VolumeTLPProperty = "owner"

	// RunNamespaceVTP is the namespace where this policy is
	// supposed to run
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Volume.runNamespace }}
	RunNamespaceVTP VolumeTLPProperty = "runNamespace"

	// CapacityVTP is the capacity of the volume
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Volume.capacity }}
	CapacityVTP VolumeTLPProperty = "capacity"

	// PersistentVolumeClaimVTP is the PVC of the volume
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Volume.pvc }}
	PersistentVolumeClaimVTP VolumeTLPProperty = "pvc"

	// IsRestoreVolumePropertyVTP indicates if the volume is being
	// created using a backedup volume
	//NOTE:
	// The Corresponding value will be accessed as
	// {{ .Volume.isRestoreVol }}
	IsRestoreVolumePropertyVTP VolumeTLPProperty = "isRestoreVol"

	// StorageClassVTP is the StorageClass of the volume
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Volume.storageclass }}
	StorageClassVTP VolumeTLPProperty = "storageclass"
)

// CloneTLPProperty is used to define properties for clone operations
type CloneTLPProperty string

const (
	// SnapshotNameVTP is the snapshot name
	SnapshotNameVTP CloneTLPProperty = "snapshotName"

	// SourceVolumeTargetIPVTP is source volume target IP
	SourceVolumeTargetIPVTP CloneTLPProperty = "sourceVolumeTargetIP"

	// IsCloneEnableVTP is a bool value for clone operations
	// for a volume
	IsCloneEnableVTP CloneTLPProperty = "isCloneEnable"

	// SourceVolumeVTP is the name of the source volume
	SourceVolumeVTP CloneTLPProperty = "sourceVolume"
)

// SnapshotTLPProperty is used to define properties for clone operations
type SnapshotTLPProperty string

const (
	// VolumeSTP is the snapshot name
	VolumeSTP SnapshotTLPProperty = "volumeName"
)

// PolicyTLPProperty is the name of the property that is found
// under PolicyTLP
type PolicyTLPProperty string

const (
	// EnabledPTP is the enabled property of the policy
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Policy.<PolicyName>.enabled }}
	EnabledPTP PolicyTLPProperty = "enabled"

	// ValuePTP is the value property of the policy
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Policy.<PolicyName>.value }}
	ValuePTP PolicyTLPProperty = "value"

	// DataPTP is the data property of the policy
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Policy.<PolicyName>.data }}
	DataPTP PolicyTLPProperty = "data"
)

const (
	// TaskIdentityPrefix is the prefix used for all TaskIdentity
	TaskIdentityPrefix string = "key"
)

// TaskTLPProperty is the name of the property that is found
// under TaskTLP
type TaskTLPProperty string

const (
	// APIVersionTTP is the apiVersion property of the task
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Task.<TaskIdentity>.apiVersion }}
	APIVersionTTP TaskTLPProperty = "apiVersion"

	// KindTTP is the kind property of the task
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Task.<TaskIdentity>.kind }}
	KindTTP TaskTLPProperty = "kind"
)

// TaskResultTLPProperty is the name of the property that is found
// under TaskResultTLP
type TaskResultTLPProperty string

const (
	// ObjectNameTRTP is the objectName property of the
	// TaskResultTLP
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .TaskResult.<TaskIdentity>.objectName }}
	ObjectNameTRTP TaskResultTLPProperty = "objectName"

	// AnnotationsTRTP is the annotations property of the
	// TaskResultTLP
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .TaskResult.<TaskIdentity>.annotations }}
	AnnotationsTRTP TaskResultTLPProperty = "annotations"

	// TaskResultVerifyErrTRTP is a property of TaskResultTLP
	//
	// First error found after **verification** checks done against the result of
	// the task's execution is stored in this property.
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .TaskResult.<TaskIdentity>.verifyErr }}
	TaskResultVerifyErrTRTP TaskResultTLPProperty = "verifyErr"

	// TaskResultNotFoundErrTRTP is a property of TaskResultTLP
	//
	// First error found after **not found** checks done against the result of
	// the task's execution is stored in this property.
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .TaskResult.<TaskIdentity>.notFoundErr }}
	TaskResultNotFoundErrTRTP TaskResultTLPProperty = "notFoundErr"

	// TaskResultVersionMismatchErrTRTP is a property of TaskResultTLP
	//
	// First error found after **version mismatch** checks done against the
	// result of the task's execution is stored in this property.
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .TaskResult.<TaskIdentity>.versionMismatchErr }}
	TaskResultVersionMismatchErrTRTP TaskResultTLPProperty = "versionMismatchErr"
)

// ListItemsTLPProperty is the name of the property that is found
// under ListItemsTLP
type ListItemsTLPProperty string

const (
	// CurrentRepeatResourceLITP is a property of ListItemsTLP
	//
	// It is the current repeat resource due to which a task is getting
	// executed is set here
	//
	// Example:
	// {{- .ListItems.currentRepeatResource -}}
	//
	// Above templating will give the current repeat resource name
	CurrentRepeatResourceLITP ListItemsTLPProperty = "currentRepeatResource"
)

// TaskConfigTLPProperty is used to define properties that comes after TaskConfigTLP
type TaskConfigTLPProperty string
