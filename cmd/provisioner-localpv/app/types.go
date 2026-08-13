package app

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	mconfig "github.com/openebs/dynamic-localpv-provisioner/pkg/apis/openebs.io/v1alpha1"
)

const (
	SnapshotKind     string = "VolumeSnapshot"
	PVCKind          string = "PersistentVolumeClaim"
	SnapshotAPIGroup string = "snapshot.storage.k8s.io"
)

// Provisioner struct has the configuration and utilities required
// across the different work-flows.
type Provisioner struct {
	kubeClient     kubernetes.Interface
	namespace      string
	helperImage    string
	nodeDeployment bool
	nodeName       string // Current node name (used in node-deployment mode)
	// allowInsecurePvcBasePathOverride when true permits BasePath values
	// supplied via PVC annotations, restoring the legacy behaviour. By
	// default PVC-supplied BasePath is rejected to prevent namespace
	// tenants from overriding the cluster-scoped StorageClass BasePath.
	allowInsecurePvcBasePathOverride bool
	// defaultConfig is the default configurations
	// provided from ENV or Code
	defaultConfig []mconfig.Config
	// getVolumeConfig is a reference to a function
	getVolumeConfig GetVolumeConfigFn
	// localVolumeManager performs volume operations directly on the node in
	// node-deployment mode. It is shared by every request so that its mutex
	// serializes those operations across concurrent provisioning requests.
	localVolumeManager *LocalVolumeManager
}

// VolumeConfig struct contains the merged configuration of the PVC
// and the associated SC. The configuration is derived from the
// annotation `cas.openebs.io/config`. The configuration will be
// in the following json format:
//
//	{
//	  Key1:{
//		enabled: true
//		value: "string value"
//	  },
//	  Key2:{
//		enabled: true
//		value: "string value"
//	  },
//	}
type VolumeConfig struct {
	pvName     string
	pvcName    string
	scName     string
	options    map[string]interface{}
	configData map[string]interface{}
	configList map[string]interface{}
}

// GetVolumeConfigFn allows to plugin a custom function
//
//	and makes it easy to unit test provisioner
type GetVolumeConfigFn func(ctx context.Context, pvName string, pvc *v1.PersistentVolumeClaim) (*VolumeConfig, error)
