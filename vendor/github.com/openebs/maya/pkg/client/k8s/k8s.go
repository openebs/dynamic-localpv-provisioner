/*
Copyright 2017 The OpenEBS Authors

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

package k8s

import (
	"encoding/json"

	openebs "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	ndm "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset"
	errors "github.com/pkg/errors"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	api_ndm_v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	api_oe_v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"

	api_apps_v1 "k8s.io/api/apps/v1"
	api_apps_v1beta1 "k8s.io/api/apps/v1beta1"
	api_batch_v1 "k8s.io/api/batch/v1"
	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
	api_storage_v1 "k8s.io/api/storage/v1"

	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	typed_oe_v1alpha1 "github.com/openebs/maya/pkg/client/generated/clientset/versioned/typed/openebs.io/v1alpha1"
	typed_ndm_v1alpha1 "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset/typed/ndm/v1alpha1"

	typed_apps_v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typed_apps_v1beta1 "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	typed_core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	typed_ext_v1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	typed_storage_v1 "k8s.io/client-go/kubernetes/typed/storage/v1"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sKind represents the Kinds understood by Kubernetes
type K8sKind string

const (
	// STSKK refers to Kubernetes StatefulSet kind
	STSKK K8sKind = "StatefulSet"

	// JobKK is a Kubernetes Job kind
	JobKK K8sKind = "Job"

	// StorageClassKK is a K8s StorageClass Kind
	StorageClassKK K8sKind = "StorageClass"

	// PodKK is a K8s Pod Kind
	PodKK K8sKind = "Pod"

	// DeploymentKK is a K8s Deployment Kind
	DeploymentKK K8sKind = "Deployment"

	// ReplicaSetKK is a K8s ReplicaSet Kind
	ReplicaSetKK K8sKind = "ReplicaSet"

	// ConfigMapKK is a K8s ConfigMap Kind
	ConfigMapKK K8sKind = "ConfigMap"

	// ServiceKK is a K8s Service Kind
	ServiceKK K8sKind = "Service"

	// CRDKK is a K8s CustomResourceDefinition Kind
	CRDKK K8sKind = "CustomResourceDefinition"

	// StroagePoolCRKK is a K8s CR of kind StoragePool
	StroagePoolCRKK K8sKind = "StoragePool"

	// StroagePoolClaimCRKK is a K8s CR of kind StoragePool
	StroagePoolClaimCRKK K8sKind = "StoragePoolClaim"

	// CStorPoolClusterCRKK is a K8s CR of kind CStorPoolCluster
	CStorPoolClusterCRKK K8sKind = "CStorPoolCluster"

	// PersistentVolumeKK is K8s PersistentVolume Kind
	PersistentVolumeKK K8sKind = "PersistentVolume"

	// PersistentVolumeClaimKK is a K8s PersistentVolumeClaim Kind
	PersistentVolumeClaimKK K8sKind = "PersistentVolumeClaim"

	// CstorPoolCRKK is a K8s CR of kind CStorPool
	CStorPoolCRKK K8sKind = "CStorPool"

	// DiskCRKK is a K8s CR of kind Disk
	DiskCRKK K8sKind = "Disk"

	// BlockDeviceCRKK is a K8s CR of kind BlockDevice
	BlockDeviceCRKK K8sKind = "BlockDevice"

	// CstorVolumeCRKK is a K8s CR of kind CStorVolume
	CStorVolumeCRKK K8sKind = "CStorVolume"

	// CstorVolumeReplicaCRKK is a K8s CR of kind CStorVolumeReplica
	CStorVolumeReplicaCRKK K8sKind = "CStorVolumeReplica"

	// UpgradeResultCRKK is a K8s CR of kind UpgradeResult
	UpgradeResultCRKK K8sKind = "UpgradeResult"

	// VolumeSnapshotDataCRKK is a K8s CR of kind VolumeSnapshotData
	VolumeSnapshotDataCRKK K8sKind = "VolumeSnapshotData"

	// VolumeSnapshotCRKK is a K8s CR of kind VolumeSnapshotData
	VolumeSnapshotCRKK K8sKind = "VolumeSnapshot"
)

// K8sAPIVersion represents valid kubernetes api version of a native or custom
// resource
type K8sAPIVersion string

const (
	// ExtensionsV1Beta1KA is the extensions/v1beta API
	ExtensionsV1Beta1KA K8sAPIVersion = "extensions/v1beta1"

	// AppsV1KA refers to kubernetes API version
	// apps/v1
	AppsV1KA K8sAPIVersion = "apps/v1"

	// AppsV1B1KA is the apps/v1beta1 API
	AppsV1B1KA K8sAPIVersion = "apps/v1beta1"

	// CoreV1KA is the v1 API
	CoreV1KA K8sAPIVersion = "v1"

	// OEV1alpha1KA is the openebs.io/v1alpha1 API
	OEV1alpha1KA K8sAPIVersion = "openebs.io/v1alpha1"

	// StorageV1KA is the storage.k8s.io/v1 API
	StorageV1KA K8sAPIVersion = "storage.k8s.io/v1"

	BatchV1KA K8sAPIVersion = "batch/v1"
)

// K8sClient provides the necessary utility to operate over
// various K8s Kind objects
type K8sClient struct {
	// ns refers to K8s namespace where the operation
	// will be performed
	ns string

	// cs refers to the Clientset capable of communicating
	// within the current K8s cluster
	cs *kubernetes.Clientset

	// oecs refers to the Clientset capable of communicating
	// within the current K8s cluster for OpenEBS objects
	oecs *openebs.Clientset

	// ndmcs refers to the Clientset capable of communicating
	// within the current K8s cluster for NDM objects
	ndmcs *ndm.Clientset

	// PV refers to a K8s PersistentVolume object
	PV *api_core_v1.PersistentVolume

	// PVC refers to a K8s PersistentVolumeClaim object
	// NOTE: This property enables unit testing
	PVC *api_core_v1.PersistentVolumeClaim

	// Pod refers to a K8s Pod object
	// NOTE: This property enables unit testing
	Pod *api_core_v1.Pod

	// Service refers to a K8s Service object
	// NOTE: This property enables unit testing
	Service *api_core_v1.Service

	// ConfigMap refers to a K8s Service object
	// NOTE: This property enables unit testing
	ConfigMap *api_core_v1.ConfigMap

	// Deployment refers to a K8s Deployment object
	// NOTE: This property enables unit testing
	Deployment *api_extn_v1beta1.Deployment

	// StorageClass refers to a K8s StorageClass object
	// NOTE: This property is useful to mock
	// during unit testing
	StorageClass *api_storage_v1.StorageClass

	// BlockDevice refers to a K8s BlockDevice CRD object
	// NOTE: This property is useful to mock
	// during unit testing
	BlockDevice *api_ndm_v1alpha1.BlockDevice

	// StoragePoolClaim refers to a K8s StoragePoolClaim CRD object
	// NOTE: This property is useful to mock
	// during unit testing
	StoragePoolClaim *api_oe_v1alpha1.StoragePoolClaim

	// CStorPoolCluster refers to a K8s CStorPoolCluster CRD object
	// NOTE: This property is useful to mock
	// during unit testing
	CStorPoolCluster *api_oe_v1alpha1.CStorPoolCluster

	// StoragePool refers to a K8s StoragePool CRD object
	// NOTE: This property is useful to mock
	// during unit testing
	StoragePool *api_oe_v1alpha1.StoragePool

	// CStorPool refers to a K8s CStorPool CRD object
	// NOTE: This property is useful to mock
	// during unit testing
	CStorPool *api_oe_v1alpha1.CStorPool

	// CStorVolume refers to a K8s CStorVolume CRD object
	// NOTE: This property is useful to mock
	// during unit testing
	CStorVolume *api_oe_v1alpha1.CStorVolume

	// CStorVolumeReplica refers to a K8s CStorVolumeReplica CRD object
	// NOTE: This property is useful to mock
	// during unit testing
	CStorVolumeReplica *api_oe_v1alpha1.CStorVolumeReplica

	// CASTemplate refers to a K8s CASTemplate custom resource
	// NOTE: This property is useful to mock
	// during unit testing
	CASTemplate *api_oe_v1alpha1.CASTemplate

	// various cert related to connecting to K8s API
	caCert     string
	caPath     string
	clientCert string
	clientKey  string
	insecure   bool
}

// NewK8sClient creates a new K8sClient
func NewK8sClient(ns string) (*K8sClient, error) {
	// get the appropriate clientset
	cs, err := getInClusterCS()
	if err != nil {
		return nil, err
	}

	// get the appropriate openebs clientset
	oecs, err := getInClusterOECS()
	if err != nil {
		return nil, err
	}

	// get the appropriate ndm clientset
	ndmcs, err := getInClusterNDMCS()
	if err != nil {
		return nil, err
	}

	return &K8sClient{
		ns:    ns,
		cs:    cs,
		oecs:  oecs,
		ndmcs: ndmcs,
	}, nil
}

// GetOECS is a getter method for fetching openebs clientset as
// the openebs clientset is not exported.
func (k *K8sClient) GetOECS() *openebs.Clientset {
	return k.oecs
}

// GetNDMCS is a getter method for fetching ndm clientset as
// the ndm clientset is not exported.
func (k *K8sClient) GetNDMCS() *ndm.Clientset {
	return k.ndmcs
}

// GetKCS is a getter method for fetching kubernetes clientset as
// the kubernetes clientset is not exported.
func (k *K8sClient) GetKCS() *kubernetes.Clientset {
	return k.cs
}

// scOps is a utility function that provides a instance capable of
// executing various K8s StorageClass related operations
func (k *K8sClient) storageV1SCOps() typed_storage_v1.StorageClassInterface {
	return k.cs.StorageV1().StorageClasses()
}

// GetStorageV1SC fetches the K8s StorageClass specs based on
// the provided name
func (k *K8sClient) GetStorageV1SC(name string, opts mach_apis_meta_v1.GetOptions) (*api_storage_v1.StorageClass, error) {
	if k.StorageClass != nil {
		return k.StorageClass, nil
	}

	scops := k.storageV1SCOps()
	return scops.Get(name, opts)
}

// GetStorageV1SCAsRaw returns a StorageClass instance
func (k *K8sClient) GetStorageV1SCAsRaw(name string) (result []byte, err error) {
	result, err = k.cs.StorageV1().RESTClient().
		Get().
		Resource("storageclasses").
		Name(name).
		VersionedParams(&mach_apis_meta_v1.GetOptions{}, scheme.ParameterCodec).
		DoRaw()

	return
}

// GetBatchV1JobAsRaw returns a Job instance
func (k *K8sClient) GetBatchV1JobAsRaw(name string) (result []byte, err error) {
	return k.cs.BatchV1().RESTClient().
		Get().
		Resource("jobs").
		Namespace(k.ns).
		Name(name).
		VersionedParams(&mach_apis_meta_v1.GetOptions{}, scheme.ParameterCodec).
		DoRaw()
}

// oeV1alpha1SPCOps is a utility function that provides a instance capable of
// executing various OpenEBS StoragePoolClaim related operations
func (k *K8sClient) oeV1alpha1SPCOps() typed_oe_v1alpha1.StoragePoolClaimInterface {
	return k.oecs.OpenebsV1alpha1().StoragePoolClaims()
}

// oeV1alpha1CSPCOps is a utility function that provides a instance capable of
// executing various OpenEBS CStorPoolCluster related operations
func (k *K8sClient) oeV1alpha1CSPCOps() typed_oe_v1alpha1.CStorPoolClusterInterface {
	return k.oecs.OpenebsV1alpha1().CStorPoolClusters(k.ns)
}

// oeV1alpha1SPOps is a utility function that provides a instance capable of
// executing various OpenEBS StoragePool related operations
func (k *K8sClient) oeV1alpha1SPOps() typed_oe_v1alpha1.StoragePoolInterface {
	return k.oecs.OpenebsV1alpha1().StoragePools()
}

// ndmV1alpha1BlockDeviceOps is a utility function that provides a instance capable of
// executing various OpenEBS BlockDevice related operations
func (k *K8sClient) ndmV1alpha1BlockDeviceOps() typed_ndm_v1alpha1.BlockDeviceInterface {
	return k.ndmcs.OpenebsV1alpha1().BlockDevices(k.ns)
}

// oeV1alpha1CSPOps is a utility function that provides a instance capable of
// executing various OpenEBS CStorPool related operations
func (k *K8sClient) oeV1alpha1CSPOps() typed_oe_v1alpha1.CStorPoolInterface {
	return k.oecs.OpenebsV1alpha1().CStorPools()
}

// GetOEV1alpha1CSP fetches the OpenEBS CStorPool specs based on
// the provided name
func (k *K8sClient) GetOEV1alpha1CSP(name string) (*api_oe_v1alpha1.CStorPool, error) {
	if k.CStorPool != nil {
		return k.CStorPool, nil
	}

	cspOps := k.oeV1alpha1CSPOps()
	return cspOps.Get(name, mach_apis_meta_v1.GetOptions{})
}

// GetOEV1alpha1BlockDevice fetches the disk specs based on
// the provided name
func (k *K8sClient) GetOEV1alpha1BlockDevice(name string) (*api_ndm_v1alpha1.BlockDevice, error) {
	if k.BlockDevice != nil {
		return k.BlockDevice, nil
	}

	diskOps := k.ndmV1alpha1BlockDeviceOps()
	return diskOps.Get(name, mach_apis_meta_v1.GetOptions{})
}

// GetOEV1alpha1SPC fetches the OpenEBS StoragePoolClaim specs based on
// the provided name
func (k *K8sClient) GetOEV1alpha1SPC(name string) (*api_oe_v1alpha1.StoragePoolClaim, error) {
	if k.StoragePoolClaim != nil {
		return k.StoragePoolClaim, nil
	}

	spcOps := k.oeV1alpha1SPCOps()
	return spcOps.Get(name, mach_apis_meta_v1.GetOptions{})
}

// GetOEV1alpha1CSPC fetches the OpenEBS CStorPoolCluster specs based on
// the provided name
func (k *K8sClient) GetOEV1alpha1CSPC(name string) (*api_oe_v1alpha1.CStorPoolCluster, error) {
	if k.CStorPoolCluster != nil {
		return k.CStorPoolCluster, nil
	}

	cspcOps := k.oeV1alpha1CSPCOps()
	return cspcOps.Get(name, mach_apis_meta_v1.GetOptions{})
}

// GetOEV1alpha1SP fetches the OpenEBS StoragePool specs based on
// the provided name
func (k *K8sClient) GetOEV1alpha1SP(name string) (*api_oe_v1alpha1.StoragePool, error) {
	if k.StoragePool != nil {
		return k.StoragePool, nil
	}

	spOps := k.oeV1alpha1SPOps()
	return spOps.Get(name, mach_apis_meta_v1.GetOptions{})
}

// CreateOEV1alpha1CSP creates a CStorPool
func (k *K8sClient) CreateOEV1alpha1CSP(csp *api_oe_v1alpha1.CStorPool) (*api_oe_v1alpha1.CStorPool, error) {
	cspops := k.oeV1alpha1CSPOps()
	return cspops.Create(csp)
}

// CreateOEV1alpha1SP creates a StoragePool
func (k *K8sClient) CreateOEV1alpha1SP(sp *api_oe_v1alpha1.StoragePool) (*api_oe_v1alpha1.StoragePool, error) {
	spops := k.oeV1alpha1SPOps()
	return spops.Create(sp)
}

// CreateOEV1alpha1CV creates a CStorVolume
func (k *K8sClient) CreateOEV1alpha1CV(cv *api_oe_v1alpha1.CStorVolume) (*api_oe_v1alpha1.CStorVolume, error) {
	cvops := k.oeV1alpha1CVOps()
	return cvops.Create(cv)
}

// oeV1alpha1CVOps is a utility function that provides a instance capable of
// executing various OpenEBS CStorVolume related operations
func (k *K8sClient) oeV1alpha1CVOps() typed_oe_v1alpha1.CStorVolumeInterface {
	return k.oecs.OpenebsV1alpha1().CStorVolumes(k.ns)
}

// GetOEV1alpha1CV fetches the OpenEBS CStorVolume specs based on
// the provided name
func (k *K8sClient) GetOEV1alpha1CV(name string) (*api_oe_v1alpha1.CStorVolume, error) {
	if k.CStorVolume != nil {
		return k.CStorVolume, nil
	}

	cvOps := k.oeV1alpha1CVOps()
	return cvOps.Get(name, mach_apis_meta_v1.GetOptions{})
}

// CreateOEV1alpha1CSPAsRaw creates a CStorVolume
func (k *K8sClient) CreateOEV1alpha1CSPAsRaw(v *api_oe_v1alpha1.CStorPool) (result []byte, err error) {
	csp, err := k.CreateOEV1alpha1CSP(v)
	if err != nil {
		return
	}
	return json.Marshal(csp)
}

// CreateOEV1alpha1SPAsRaw creates a StoragePool
func (k *K8sClient) CreateOEV1alpha1SPAsRaw(v *api_oe_v1alpha1.StoragePool) (result []byte, err error) {
	sp, err := k.CreateOEV1alpha1SP(v)
	if err != nil {
		return
	}
	return json.Marshal(sp)
}

// CreateOEV1alpha1CVAsRaw creates a CStorVolume
func (k *K8sClient) CreateOEV1alpha1CVAsRaw(v *api_oe_v1alpha1.CStorVolume) (result []byte, err error) {
	csv, err := k.CreateOEV1alpha1CV(v)
	if err != nil {
		return
	}

	return json.Marshal(csv)
}

// CreateOEV1alpha1CVR creates a CStorVolumeReplica
func (k *K8sClient) CreateOEV1alpha1CVR(cvr *api_oe_v1alpha1.CStorVolumeReplica) (*api_oe_v1alpha1.CStorVolumeReplica, error) {
	cvrops := k.oeV1alpha1CVROps()
	return cvrops.Create(cvr)
}

// oeV1alpha1CVROps is a utility function that provides a instance capable of
// executing various OpenEBS CStorVolumeReplica related operations
func (k *K8sClient) oeV1alpha1CVROps() typed_oe_v1alpha1.CStorVolumeReplicaInterface {
	return k.oecs.OpenebsV1alpha1().CStorVolumeReplicas(k.ns)
}

// GetOEV1alpha1CVR fetches the OpenEBS CStorVolumeReplica specs based on
// the provided name
func (k *K8sClient) GetOEV1alpha1CVR(name string) (*api_oe_v1alpha1.CStorVolumeReplica, error) {
	if k.CStorVolume != nil {
		return k.CStorVolumeReplica, nil
	}

	cvrOps := k.oeV1alpha1CVROps()
	return cvrOps.Get(name, mach_apis_meta_v1.GetOptions{})
}

// CreateOEV1alpha1CVRAsRaw creates a CStorVolumeReplica
func (k *K8sClient) CreateOEV1alpha1CVRAsRaw(vr *api_oe_v1alpha1.CStorVolumeReplica) (result []byte, err error) {
	csvr, err := k.CreateOEV1alpha1CVR(vr)
	if err != nil {
		return
	}

	return json.Marshal(csvr)
}

// oeV1alpha1CASTOps is a utility function that provides a instance capable of
// executing various OpenEBS CASTemplate related operations
func (k *K8sClient) oeV1alpha1CASTOps() typed_oe_v1alpha1.CASTemplateInterface {
	return k.oecs.OpenebsV1alpha1().CASTemplates()
}

// GetOEV1alpha1CAST fetches the OpenEBS CASTemplate specs based on
// the provided name
func (k *K8sClient) GetOEV1alpha1CAST(name string, opts mach_apis_meta_v1.GetOptions) (*api_oe_v1alpha1.CASTemplate, error) {
	if k.CASTemplate != nil {
		return k.CASTemplate, nil
	}

	castOps := k.oeV1alpha1CASTOps()
	return castOps.Get(name, opts)
}

// oeV1alpha1RunTaskOps is a utility function that provides a instance capable
// of executing operations on RunTask custom resource
func (k *K8sClient) oeV1alpha1RunTaskOps() typed_oe_v1alpha1.RunTaskInterface {
	return k.oecs.OpenebsV1alpha1().RunTasks(k.ns)
}

// GetOEV1alpha1RunTask fetches the OpenEBS CASTemplate specs based on
// the provided name
func (k *K8sClient) GetOEV1alpha1RunTask(name string, opts mach_apis_meta_v1.GetOptions) (*api_oe_v1alpha1.RunTask, error) {
	rtOps := k.oeV1alpha1RunTaskOps()
	return rtOps.Get(name, opts)
}

// cmOps is a utility function that provides a instance capable of
// executing various K8s ConfigMap related operations.
func (k *K8sClient) cmOps() typed_core_v1.ConfigMapInterface {
	return k.cs.CoreV1().ConfigMaps(k.ns)
}

// GetConfigMap fetches the K8s ConfigMap with the provided name
func (k *K8sClient) GetConfigMap(name string, opts mach_apis_meta_v1.GetOptions) (*api_core_v1.ConfigMap, error) {
	if k.ConfigMap != nil {
		return k.ConfigMap, nil
	}

	cops := k.cmOps()
	return cops.Get(name, opts)
}

// coreV1PVCOps is a utility function that provides a instance capable of
// executing various K8s PVC related operations.
func (k *K8sClient) coreV1PVCOps() typed_core_v1.PersistentVolumeClaimInterface {
	return k.cs.CoreV1().PersistentVolumeClaims(k.ns)
}

// GetPVC fetches the K8s PVC with the provided name
func (k *K8sClient) GetPVC(name string, opts mach_apis_meta_v1.GetOptions) (*api_core_v1.PersistentVolumeClaim, error) {
	if k.PVC != nil {
		return k.PVC, nil
	}

	pops := k.coreV1PVCOps()
	return pops.Get(name, opts)
}

// coreV1PVOps is a utility function that provides an instance capable of
// executing various K8s PV related operations.
func (k *K8sClient) coreV1PVOps() typed_core_v1.PersistentVolumeInterface {
	return k.cs.CoreV1().PersistentVolumes()
}

// GetPV fetches the K8s PV with the provided name
func (k *K8sClient) GetPV(name string, opts mach_apis_meta_v1.GetOptions) (*api_core_v1.PersistentVolume, error) {
	if k.PV != nil {
		return k.PV, nil
	}

	pops := k.coreV1PVOps()
	return pops.Get(name, opts)
}

// GetCoreV1PersistentVolumeAsRaw fetches the K8s PersistentVolume with the
// provided name
func (k *K8sClient) GetCoreV1PersistentVolumeAsRaw(name string) (result []byte, err error) {
	result, err = k.cs.CoreV1().RESTClient().
		Get().
		Resource("persistentvolumes").
		Name(name).
		VersionedParams(&mach_apis_meta_v1.GetOptions{}, scheme.ParameterCodec).
		DoRaw()

	return
}

// GetCoreV1PVCAsRaw fetches the K8s PVC with the provided name
func (k *K8sClient) GetCoreV1PVCAsRaw(name string) (result []byte, err error) {
	result, err = k.cs.CoreV1().RESTClient().
		Get().
		Namespace(k.ns).
		Resource("persistentvolumeclaims").
		Name(name).
		VersionedParams(&mach_apis_meta_v1.GetOptions{}, scheme.ParameterCodec).
		DoRaw()

	return
}

// GetExtnV1B1DeploymentAsRaw fetches the K8s Deployment with the provided name
func (k *K8sClient) GetExtnV1B1DeploymentAsRaw(name string) (result []byte, err error) {
	result, err = k.cs.ExtensionsV1beta1().RESTClient().
		Get().
		Namespace(k.ns).
		Resource("deployments").
		Name(name).
		VersionedParams(&mach_apis_meta_v1.GetOptions{}, scheme.ParameterCodec).
		DoRaw()

	return
}

// GetAppsV1B1DeploymentAsRaw fetches the K8s Deployment with the provided name
func (k *K8sClient) GetAppsV1B1DeploymentAsRaw(name string) (result []byte, err error) {
	result, err = k.cs.AppsV1beta1().RESTClient().
		Get().
		Namespace(k.ns).
		Resource("deployments").
		Name(name).
		VersionedParams(&mach_apis_meta_v1.GetOptions{}, scheme.ParameterCodec).
		DoRaw()

	return
}

// GetOEV1alpha1BlockDeviceAsRaw fetches the OpenEBS Disk with the provided name
func (k *K8sClient) GetOEV1alpha1BlockDeviceAsRaw(name string) (result []byte, err error) {
	bd, err := k.GetOEV1alpha1BlockDevice(name)
	if err != nil {
		return
	}

	return json.Marshal(bd)

	// TODO
	//  A better way needs to be determined to get or use raw bytes of a resource.
	// These lines will be removed or refactor-ed once we conclude on this better
	// approach.
	//
	//result, err = k.oecs.OpenebsV1alpha1().RESTClient().
	//	Get().
	//	Namespace(k.ns).
	//	Resource("storagepools").
	//	Name(name).
	//	VersionedParams(&mach_apis_meta_v1.GetOptions{}, scheme.ParameterCodec).
	//	DoRaw()

	//return
}

// GetOEV1alpha1SPCAsRaw fetches the OpenEBS SPC with the provided name
func (k *K8sClient) GetOEV1alpha1SPCAsRaw(name string) (result []byte, err error) {
	spc, err := k.GetOEV1alpha1SPC(name)
	if err != nil {
		return
	}

	return json.Marshal(spc)

	// TODO
	//  A better way needs to be determined to get or use raw bytes of a resource.
	// These lines will be removed or refactor-ed once we conclude on this better
	// approach.
	//
	//result, err = k.oecs.OpenebsV1alpha1().RESTClient().
	//	Get().
	//	Namespace(k.ns).
	//	Resource("storagepools").
	//	Name(name).
	//	VersionedParams(&mach_apis_meta_v1.GetOptions{}, scheme.ParameterCodec).
	//	DoRaw()

	//return
}

// GetOEV1alpha1CSPCAsRaw fetches the OpenEBS CSPC with the provided name
func (k *K8sClient) GetOEV1alpha1CSPCAsRaw(name string) (result []byte, err error) {
	cspc, err := k.GetOEV1alpha1CSPC(name)
	if err != nil {
		return
	}

	return json.Marshal(cspc)
}

// GetOEV1alpha1SPAsRaw fetches the OpenEBS SP with the provided name
func (k *K8sClient) GetOEV1alpha1SPAsRaw(name string) (result []byte, err error) {
	sp, err := k.GetOEV1alpha1SP(name)
	if err != nil {
		return
	}

	return json.Marshal(sp)

	// TODO
	//  A better way needs to be determined to get or use raw bytes of a resource.
	// These lines will be removed or refactor-ed once we conclude on this better
	// approach.
	//
	//result, err = k.oecs.OpenebsV1alpha1().RESTClient().
	//	Get().
	//	Namespace(k.ns).
	//	Resource("storagepools").
	//	Name(name).
	//	VersionedParams(&mach_apis_meta_v1.GetOptions{}, scheme.ParameterCodec).
	//	DoRaw()

	//return
}

// GetOEV1alpha1CSPAsRaw fetches the OpenEBS CSP with the provided name
func (k *K8sClient) GetOEV1alpha1CSPAsRaw(name string) (result []byte, err error) {
	csp, err := k.GetOEV1alpha1CSP(name)
	if err != nil {
		return
	}

	return json.Marshal(csp)
}

// podOps is a utility function that provides a instance capable of
// executing various K8s pod related operations.
func (k *K8sClient) podOps() typed_core_v1.PodInterface {
	return k.cs.CoreV1().Pods(k.ns)
}

// GetPod fetches the K8s Pod with the provided name
func (k *K8sClient) GetPod(name string, opts mach_apis_meta_v1.GetOptions) (*api_core_v1.Pod, error) {
	if k.Pod != nil {
		return k.Pod, nil
	}

	pops := k.podOps()
	return pops.Get(name, opts)
}

// GetPods fetches the K8s Pods
func (k *K8sClient) GetPods() ([]api_core_v1.Pod, error) {
	podLists, err := k.cs.CoreV1().Pods(k.ns).List(mach_apis_meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return podLists.Items, nil
}

// ListCoreV1PVCAsRaw fetches a list of K8s PVCs with the provided options
func (k *K8sClient) ListCoreV1PVCAsRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	result, err = k.cs.CoreV1().RESTClient().Get().
		Namespace(k.ns).
		Resource("persistentvolumeclaims").
		VersionedParams(&opts, scheme.ParameterCodec).
		DoRaw()
	err = errors.WithStack(err)
	return
}

// ListCoreV1PVAsRaw fetches a list of K8s PVs with the provided options
func (k *K8sClient) ListCoreV1PVAsRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	result, err = k.cs.CoreV1().RESTClient().Get().
		Namespace(k.ns).
		Resource("persistentvolumes").
		VersionedParams(&opts, scheme.ParameterCodec).
		DoRaw()
	err = errors.WithStack(err)
	return
}

// ListCoreV1PodAsRaw fetches a list of K8s Pods as per the provided options
func (k *K8sClient) ListCoreV1PodAsRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	result, err = k.cs.CoreV1().RESTClient().Get().
		Namespace(k.ns).
		Resource("pods").
		VersionedParams(&opts, scheme.ParameterCodec).
		DoRaw()
	err = errors.WithStack(err)
	return
}

// ListCoreV1ServiceAsRaw fetches a list of K8s Services as per the provided options
func (k *K8sClient) ListCoreV1ServiceAsRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	result, err = k.cs.CoreV1().RESTClient().Get().
		Namespace(k.ns).
		Resource("services").
		VersionedParams(&opts, scheme.ParameterCodec).
		DoRaw()
	err = errors.WithStack(err)
	return
}

// ListExtnV1B1DeploymentAsRaw fetches a list of K8s Deployments as per the
// provided options
func (k *K8sClient) ListExtnV1B1DeploymentAsRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	result, err = k.cs.ExtensionsV1beta1().RESTClient().Get().
		Namespace(k.ns).
		Resource("deployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		DoRaw()
	err = errors.WithStack(err)
	return
}

// ListAppsV1B1DeploymentAsRaw fetches a list of K8s Deployments as per the
// provided options
func (k *K8sClient) ListAppsV1B1DeploymentAsRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	result, err = k.cs.AppsV1beta1().RESTClient().Get().
		Namespace(k.ns).
		Resource("deployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		DoRaw()
	err = errors.WithStack(err)
	return
}

// ListAppsV1DeploymentAsRaw fetches a list of K8s apps/v1 Deployments as per the
// provided options
func (k *K8sClient) ListAppsV1DeploymentAsRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	result, err = k.cs.AppsV1().RESTClient().Get().
		Namespace(k.ns).
		Resource("deployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		DoRaw()
	err = errors.WithStack(err)
	return
}

// ListOEV1alpha1BlockDeviceRaw fetches a list of BlockDevices as per the
// provided options
func (k *K8sClient) ListOEV1alpha1BlockDeviceRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	bdOps := k.ndmV1alpha1BlockDeviceOps()
	bdList, err := bdOps.List(opts)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	result, err = json.Marshal(bdList)
	err = errors.WithStack(err)
	return
}

// ListOEV1alpha1SPRaw fetches a list of StoragePool as per the
// provided options
func (k *K8sClient) ListOEV1alpha1SPRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	spOps := k.oeV1alpha1SPOps()
	spList, err := spOps.List(opts)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	result, err = json.Marshal(spList)
	err = errors.WithStack(err)
	return
}

// ListOEV1alpha1CSPRaw fetches a list of CStorPool as per the
// provided options
func (k *K8sClient) ListOEV1alpha1CSPRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	cspOps := k.oeV1alpha1CSPOps()
	cspList, err := cspOps.List(opts)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	result, err = json.Marshal(cspList)
	err = errors.WithStack(err)
	return
}

// ListOEV1alpha1CVRRaw fetches a list of CStorVolumeReplica as per the
// provided options
func (k *K8sClient) ListOEV1alpha1CVRRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	cvrOps := k.oeV1alpha1CVROps()
	cvrList, err := cvrOps.List(opts)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	result, err = json.Marshal(cvrList)
	err = errors.WithStack(err)
	return
}

// ListOEV1alpha1CVRaw fetches a list of CStorVolume as per the
// provided options
func (k *K8sClient) ListOEV1alpha1CVRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	cvOps := k.oeV1alpha1CVOps()
	cvrList, err := cvOps.List(opts)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	result, err = json.Marshal(cvrList)
	err = errors.WithStack(err)
	return
}

// serviceOps is a utility function that provides a instance capable of
// executing various k8s service related operations.
func (k *K8sClient) serviceOps() typed_core_v1.ServiceInterface {
	return k.cs.CoreV1().Services(k.ns)
}

// GetService fetches the K8s Service with the provided name
func (k *K8sClient) GetService(name string, opts mach_apis_meta_v1.GetOptions) (*api_core_v1.Service, error) {
	if k.Service != nil {
		return k.Service, nil
	}

	sops := k.serviceOps()
	return sops.Get(name, opts)
}

// coreV1ServiceOps is a utility function that provides a instance capable of
// executing various k8s service related operations.
func (k *K8sClient) coreV1ServiceOps() typed_core_v1.ServiceInterface {
	return k.cs.CoreV1().Services(k.ns)
}

// CreateCoreV1Service creates a K8s Service
func (k *K8sClient) CreateCoreV1Service(svc *api_core_v1.Service) (*api_core_v1.Service, error) {
	sops := k.coreV1ServiceOps()
	return sops.Create(svc)
}

// DeleteCoreV1Service deletes a K8s Service
func (k *K8sClient) DeleteCoreV1Service(name string) error {
	sops := k.coreV1ServiceOps()
	deletePropagation := mach_apis_meta_v1.DeletePropagationForeground
	return sops.Delete(name, &mach_apis_meta_v1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	})
}

// DeleteBatchV1Job deletes a K8s job
func (k *K8sClient) DeleteBatchV1Job(name string) error {
	deletePropagation := mach_apis_meta_v1.DeletePropagationForeground
	return k.cs.BatchV1().Jobs(k.ns).Delete(
		name,
		&mach_apis_meta_v1.DeleteOptions{PropagationPolicy: &deletePropagation})
}

// DeleteAppsV1STS deletes a kubernetes StatefulSet
// object
func (k *K8sClient) DeleteAppsV1STS(name string) error {
	deletePropagation := mach_apis_meta_v1.DeletePropagationForeground
	return k.cs.AppsV1().StatefulSets(k.ns).Delete(
		name,
		&mach_apis_meta_v1.DeleteOptions{PropagationPolicy: &deletePropagation})
}

// TODO deprecate
//
// deploymentOps is a utility function that provides a instance capable of
// executing various k8s Deployment related operations.
func (k *K8sClient) deploymentOps() typed_ext_v1beta1.DeploymentInterface {
	return k.cs.ExtensionsV1beta1().Deployments(k.ns)
}

// extnV1B1DeploymentOps is a utility function that provides a instance capable of
// executing various k8s Deployment related operations.
func (k *K8sClient) extnV1B1DeploymentOps() typed_ext_v1beta1.DeploymentInterface {
	return k.cs.ExtensionsV1beta1().Deployments(k.ns)
}

// extnV1B1ReplicaSetOps is a utility function that provides an instance capable of
// executing various k8s ReplicaSet related operations.
func (k *K8sClient) extnV1B1ReplicaSetOps() typed_ext_v1beta1.ReplicaSetInterface {
	return k.cs.ExtensionsV1beta1().ReplicaSets(k.ns)
}

// GetDeployment fetches the K8s Deployment with the provided name
func (k *K8sClient) GetDeployment(name string, opts mach_apis_meta_v1.GetOptions) (*api_extn_v1beta1.Deployment, error) {
	if k.Deployment != nil {
		return k.Deployment, nil
	}

	dops := k.deploymentOps()
	return dops.Get(name, opts)
}

// CreateExtnV1B1Deployment creates a K8s Deployment
func (k *K8sClient) CreateExtnV1B1Deployment(d *api_extn_v1beta1.Deployment) (*api_extn_v1beta1.Deployment, error) {
	dops := k.extnV1B1DeploymentOps()
	return dops.Create(d)
}

// CreateExtnV1B1DeploymentAsRaw creates a K8s Deployment
func (k *K8sClient) CreateExtnV1B1DeploymentAsRaw(d *api_extn_v1beta1.Deployment) (result []byte, err error) {
	deploy, err := k.CreateExtnV1B1Deployment(d)
	if err != nil {
		return
	}

	return json.Marshal(deploy)

	// TODO
	//  A better way needs to be determined to get or use raw bytes of a resource.
	// These lines will be removed or refactor-ed once we conclude on this better
	// approach.
	//
	//result, err = k.cs.ExtensionsV1beta1().RESTClient().
	//	Put().
	//	Namespace(k.ns).
	//	Resource("deployments").
	//	Body(d).
	//	DoRaw()

	//return
}

// CreateAppsV1B1DeploymentAsRaw creates a K8s Deployment
func (k *K8sClient) CreateAppsV1B1DeploymentAsRaw(d *api_apps_v1beta1.Deployment) (result []byte, err error) {
	deploy, err := k.CreateAppsV1B1Deployment(d)
	if err != nil {
		return
	}

	return json.Marshal(deploy)

	// TODO
	//  A better way needs to be determined to get or use raw bytes of a resource.
	// These lines will be removed or refactor-ed once we conclude on this better
	// approach.
	//
	//result, err = k.cs.AppsV1beta1().RESTClient().
	//	Put().
	//	Namespace(k.ns).
	//	Resource("deployments").
	//	Body(d).
	//	DoRaw()

	//return
}

// CreateAppsV1DeploymentAsRaw creates a K8s Deployment based on apps/v1
// apiVersion
func (k *K8sClient) CreateAppsV1DeploymentAsRaw(d *api_apps_v1.Deployment) (result []byte, err error) {
	deploy, err := k.CreateAppsV1Deployment(d)
	if err != nil {
		return
	}

	return json.Marshal(deploy)
}

// CreateCoreV1ServiceAsRaw creates a K8s Service
func (k *K8sClient) CreateCoreV1ServiceAsRaw(s *api_core_v1.Service) (result []byte, err error) {
	svc, err := k.CreateCoreV1Service(s)
	if err != nil {
		return
	}

	return json.Marshal(svc)

	// TODO
	//  A better way needs to be determined to get or use raw bytes of a resource.
	// These lines will be removed or refactor-ed once we conclude on this better
	// approach.
	//
	//result, err = k.cs.CoreV1().RESTClient().
	//	Put().
	//	Namespace(k.ns).
	//	Resource("services").
	//	Body(s).
	//	DoRaw()

	//return
}

// PatchExtnV1B1Deployment patches the K8s Deployment with the provided patches
func (k *K8sClient) PatchExtnV1B1Deployment(name string, patchType types.PatchType, patches []byte) (*api_extn_v1beta1.Deployment, error) {
	dops := k.extnV1B1DeploymentOps()
	return dops.Patch(name, patchType, patches)
}

// PatchOEV1alpha1SPCAsRaw patches the SPC object with the provided patches
func (k *K8sClient) PatchOEV1alpha1SPCAsRaw(name string, patchType types.PatchType, patches []byte) (result *api_oe_v1alpha1.StoragePoolClaim, err error) {
	result, err = k.oecs.OpenebsV1alpha1().StoragePoolClaims().Patch(name, patchType, patches)
	return
}

// PatchOEV1alpha1CSPCAsRaw patches the CSPC object with the provided patches
func (k *K8sClient) PatchOEV1alpha1CSPCAsRaw(name string, patchType types.PatchType, patches []byte) (result *api_oe_v1alpha1.CStorPoolCluster, err error) {
	result, err = k.oecs.OpenebsV1alpha1().CStorPoolClusters(k.ns).Patch(name, patchType, patches)
	return
}

// PatchOEV1alpha1CSV patches the CSV object with the provided patches
func (k *K8sClient) PatchOEV1alpha1CSV(name, namespace string, patchType types.PatchType, patches []byte) (result *api_oe_v1alpha1.CStorVolume, err error) {
	result, err = k.oecs.OpenebsV1alpha1().CStorVolumes(namespace).Patch(name, patchType, patches)
	return
}

// PatchOEV1alpha1CVR patches the CVR object with the provided patches
func (k *K8sClient) PatchOEV1alpha1CVR(name, namespace string, patchType types.PatchType, patches []byte) (result *api_oe_v1alpha1.CStorVolumeReplica, err error) {
	result, err = k.oecs.OpenebsV1alpha1().CStorVolumeReplicas(namespace).Patch(name, patchType, patches)
	return
}

// PatchExtnV1B1DeploymentAsRaw patches the K8s Deployment with the provided patches
func (k *K8sClient) PatchExtnV1B1DeploymentAsRaw(name string, patchType types.PatchType, patches []byte) (result []byte, err error) {
	result, err = k.cs.ExtensionsV1beta1().RESTClient().Patch(patchType).
		Namespace(k.ns).
		Resource("deployments").
		Name(name).
		Body(patches).
		DoRaw()

	return
}

// PatchAppsV1DeploymentAsRaw patches the K8s Deployment with the provided patches
func (k *K8sClient) PatchAppsV1DeploymentAsRaw(name string, patchType types.PatchType, patches []byte) (result []byte, err error) {
	result, err = k.cs.AppsV1().RESTClient().Patch(patchType).
		Namespace(k.ns).
		Resource("deployments").
		Name(name).
		Body(patches).
		DoRaw()

	return
}

// PatchCoreV1ServiceAsRaw patches the K8s Service with the provided patches
func (k *K8sClient) PatchCoreV1ServiceAsRaw(name string, patchType types.PatchType, patches []byte) (result []byte, err error) {
	result, err = k.cs.CoreV1().RESTClient().Patch(patchType).
		Namespace(k.ns).
		Resource("services").
		Name(name).
		Body(patches).
		DoRaw()

	return
}

// DeleteExtnV1B1Deployment deletes the K8s Deployment with the provided name
func (k *K8sClient) DeleteExtnV1B1Deployment(name string) error {
	dops := k.extnV1B1DeploymentOps()
	// ensure all the dependants are deleted
	deletePropagation := mach_apis_meta_v1.DeletePropagationForeground
	return dops.Delete(name, &mach_apis_meta_v1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	})
}

// CreateBatchV1JobAsRaw creates a kubernetes Job
func (k *K8sClient) CreateBatchV1JobAsRaw(j *api_batch_v1.Job) ([]byte, error) {
	job, err := k.cs.BatchV1().Jobs(k.ns).Create(j)
	if err != nil {
		return nil, err
	}
	return json.Marshal(job)
}

// CreateAppsV1STSAsRaw creates a kubernetes StatefulSet
func (k *K8sClient) CreateAppsV1STSAsRaw(sts *api_apps_v1.StatefulSet) ([]byte, error) {
	s, err := k.cs.AppsV1().StatefulSets(k.ns).Create(sts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(s)
}

// appsV1B1DeploymentOps is a utility function that provides a instance capable of
// executing various k8s Deployment related operations.
func (k *K8sClient) appsV1B1DeploymentOps() typed_apps_v1beta1.DeploymentInterface {
	return k.cs.AppsV1beta1().Deployments(k.ns)
}

func (k *K8sClient) appsV1DeploymentOps() typed_apps_v1.DeploymentInterface {
	return k.cs.AppsV1().Deployments(k.ns)
}

// GetAppsV1B1Deployment fetches the K8s Deployment with the provided name
func (k *K8sClient) GetAppsV1B1Deployment(name string, opts mach_apis_meta_v1.GetOptions) (*api_apps_v1beta1.Deployment, error) {
	dops := k.appsV1B1DeploymentOps()
	return dops.Get(name, opts)
}

// CreateAppsV1B1Deployment creates the K8s Deployment with the provided name
func (k *K8sClient) CreateAppsV1B1Deployment(d *api_apps_v1beta1.Deployment) (*api_apps_v1beta1.Deployment, error) {
	dops := k.appsV1B1DeploymentOps()
	return dops.Create(d)
}

// CreateAppsV1Deployment creates the K8s Deployment with the provided name
func (k *K8sClient) CreateAppsV1Deployment(d *api_apps_v1.Deployment) (*api_apps_v1.Deployment, error) {
	dops := k.appsV1DeploymentOps()
	return dops.Create(d)
}

// DeleteAppsV1B1Deployment deletes the K8s Deployment with the provided name
func (k *K8sClient) DeleteAppsV1B1Deployment(name string) error {
	dops := k.appsV1B1DeploymentOps()
	// ensure all the dependants are deleted
	deletePropagation := mach_apis_meta_v1.DeletePropagationForeground
	return dops.Delete(name, &mach_apis_meta_v1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	})
}

// DeleteAppsV1Deployment deletes the K8s Deployment with the provided name
func (k *K8sClient) DeleteAppsV1Deployment(name string) error {
	dops := k.appsV1DeploymentOps()
	// ensure all the dependants are deleted
	deletePropagation := mach_apis_meta_v1.DeletePropagationForeground
	return dops.Delete(name, &mach_apis_meta_v1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	})
}

// DeleteOEV1alpha1SP deletes the StoragePool with the provided name
func (k *K8sClient) DeleteOEV1alpha1SP(name string) error {
	spops := k.oeV1alpha1SPOps()
	// ensure all the dependants are deleted
	deletePropagation := mach_apis_meta_v1.DeletePropagationForeground
	return spops.Delete(name, &mach_apis_meta_v1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	})
}

// DeleteOEV1alpha1CSP deletes the CStorPool with the provided name
func (k *K8sClient) DeleteOEV1alpha1CSP(name string) error {
	cspops := k.oeV1alpha1CSPOps()
	// ensure all the dependants are deleted
	deletePropagation := mach_apis_meta_v1.DeletePropagationForeground
	return cspops.Delete(name, &mach_apis_meta_v1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	})
}

// DeleteOEV1alpha1CSV deletes the CStorVolume with the provided name
func (k *K8sClient) DeleteOEV1alpha1CSV(name string) error {
	cvops := k.oeV1alpha1CVOps()
	// ensure all the dependants are deleted
	deletePropagation := mach_apis_meta_v1.DeletePropagationForeground
	return cvops.Delete(name, &mach_apis_meta_v1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	})
}

// DeleteOEV1alpha1CVR deletes the CStorVolumeReplica with the provided name
func (k *K8sClient) DeleteOEV1alpha1CVR(name string) error {
	cvrops := k.oeV1alpha1CVROps()
	// ensure all the dependants are deleted
	deletePropagation := mach_apis_meta_v1.DeletePropagationForeground
	return cvrops.Delete(name, &mach_apis_meta_v1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	})
}

func getK8sConfig() (config *rest.Config, err error) {
	k8sMaster := env.Get(env.KubeMaster)
	kubeConfig := env.Get(env.KubeConfig)

	if len(k8sMaster) != 0 || len(kubeConfig) != 0 {
		// creates the config from k8sMaster or kubeConfig
		return clientcmd.BuildConfigFromFlags(k8sMaster, kubeConfig)
	}

	// creates the in-cluster config making use of the Pod's ENV & secrets
	return rest.InClusterConfig()
}

// getInClusterCS is used to initialize and return a new http client capable
// of invoking K8s APIs within the cluster
func getInClusterCS() (clientset *kubernetes.Clientset, err error) {
	config, err := getK8sConfig()
	if err != nil {
		return nil, err
	}

	// creates the in-cluster kubernetes clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// getInClusterOECS is used to initialize and return a new http client capable
// of invoking OpenEBS CRD APIs within the cluster
func getInClusterOECS() (clientset *openebs.Clientset, err error) {
	config, err := getK8sConfig()
	if err != nil {
		return nil, err
	}

	// creates the in-cluster openebs clientset
	clientset, err = openebs.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// getInClusterNDMCS is used to initialize and return a new http client capable
// of invoking NDM CRD APIs within the cluster
func getInClusterNDMCS() (clientset *ndm.Clientset, err error) {
	config, err := getK8sConfig()
	if err != nil {
		return nil, err
	}

	// creates the in-cluster openebs clientset
	clientset, err = ndm.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
