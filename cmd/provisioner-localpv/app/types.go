/*
Copyright 2019 The OpenEBS Authors.

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

package app

import (
	"context"

	mconfig "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	SnapshotKind     string = "VolumeSnapshot"
	PVCKind          string = "PersistentVolumeClaim"
	SnapshotAPIGroup string = "snapshot.storage.k8s.io"
)

// Provisioner struct has the configuration and utilities required
// across the different work-flows.
type Provisioner struct {
	kubeClient  *clientset.Clientset
	namespace   string
	helperImage string
	// defaultConfig is the default configurations
	// provided from ENV or Code
	defaultConfig []mconfig.Config
	// getVolumeConfig is a reference to a function
	getVolumeConfig GetVolumeConfigFn
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
