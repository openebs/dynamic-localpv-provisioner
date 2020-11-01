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

import (
	ndm "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CasPoolKey is the key for the CasPool.
type CasPoolKey string

// CasPoolValString represents the string value for a CasPoolKey.
type CasPoolValString string

// CasPoolValInt represents the integer value for a CasPoolKey
type CasPoolValInt int

const (
	// HostNameCPK is the kubernetes host name label
	HostNameCPK CasPoolKey = "kubernetes.io/hostname"
	// StoragePoolClaimCPK is the storage pool claim label
	StoragePoolClaimCPK CasPoolKey = "openebs.io/storage-pool-claim"
	// CStorPoolClusterCPK is the CStorPoolcluster label
	CStorPoolClusterCPK CasPoolKey = "openebs.io/cstor-pool-cluster"
	// CStorPoolInstanceCPK is the CStorPoolInstance label
	CStorPoolInstanceCPK CasPoolKey = "openebs.io/cstor-pool-instance"
	// NdmDiskTypeCPK is the node-disk-manager disk type e.g. 'sparse' or 'disk'
	NdmDiskTypeCPK CasPoolKey = "ndm.io/disk-type"
	// NdmBlockDeviceTypeCPK is the node-disk-manager blockdevice type e.g. // 'blockdevice'
	NdmBlockDeviceTypeCPK CasPoolKey = "ndm.io/blockdevice-type"
	// PoolTypeMirroredCPV is a key for mirrored for pool
	PoolTypeMirroredCPV CasPoolValString = "mirrored"
	// PoolTypeStripedCPV is a key for striped for pool
	PoolTypeStripedCPV CasPoolValString = "striped"
	// PoolTypeRaidzCPV is a key for raidz for pool
	PoolTypeRaidzCPV CasPoolValString = "raidz"
	// PoolTypeRaidz2CPV is a key for raidz for pool
	PoolTypeRaidz2CPV CasPoolValString = "raidz2"
	// TypeSparseCPV is a key for sparse disk pool
	TypeSparseCPV CasPoolValString = "sparse"
	// TypeDiskCPV is a key for physical,iscsi,virtual etc disk pool
	TypeDiskCPV CasPoolValString = "disk"
	// TypeBlockDeviceCPV is a key for physical,iscsi,virtual etc disk pool
	TypeBlockDeviceCPV CasPoolValString = "blockdevice"
	// StripedBlockDeviceCountCPV is the count for striped type pool
	StripedBlockDeviceCountCPV CasPoolValInt = 1
	// MirroredBlockDeviceCountCPV is the count for mirrored type pool
	MirroredBlockDeviceCountCPV CasPoolValInt = 2
	// RaidzBlockDeviceCountCPV is the count for raidz type pool
	RaidzBlockDeviceCountCPV CasPoolValInt = 3
	// Raidz2BlockDeviceCountCPV is the count for raidz2 type pool
	Raidz2BlockDeviceCountCPV CasPoolValInt = 6
	// PersistentVolumeCPK is a key on the various resources to identify it belongs
	// to particular volume
	PersistentVolumeCPK CasPoolKey = "openebs.io/persistent-volume"
)

const (
	// PoolProtectionFinalizer will make sure that pool will be destroyed
	// on the disks before removing the CSP from etcd
	PoolProtectionFinalizer = "openebs.io/pool-protection"
)

// CasPool is a type which will be utilised by CAS engine to perform
// storagepool related operation.
// TODO: Restrucutre CasPool struct.
type CasPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// StoragePoolClaim is the name of the storagepoolclaim object
	StoragePoolClaim string

	// CasCreateTemplate is the cas template that will be used for storagepool create
	// operation
	CasCreateTemplate string

	// CasDeleteTemplate is the cas template that will be used for storagepool delete
	// operation
	CasDeleteTemplate string

	// Namespace can be passed via storagepoolclaim as labels to decide on the
	// execution of namespaced resources with respect to storagepool
	Namespace string

	// BlockDeviceList is the list of block devices over which a storagepool will be provisioned
	BlockDeviceList []BlockDeviceGroup

	// PoolType is the type of pool to be provisioned e.g. striped or mirrored
	PoolType string

	// PoolCacheFile is cache file which used at the time of importing a pool
	PoolCacheFile string

	// MaxPool is the maximum number of pool that should be provisioned
	MaxPools int

	// MinPool is the minimum number of pool that should be provisioned
	MinPools int

	// Type is the CasPool type e.g. sparse or openebs-cstor
	Type string

	// NodeName is the node where cstor pool will be created
	NodeName string

	// reSync will decide whether the event is a reconciliation event
	ReSync bool

	// PendingPoolCount is the number of pools that will be tried for creation as a part of reconciliation.
	PendingPoolCount int

	DeviceID           []string
	APIBlockDeviceList ndm.BlockDeviceList

	// PoolROThreshold is threshold limit for pool read only mode
	PoolROThreshold int
}
