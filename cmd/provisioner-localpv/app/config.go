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
	//"fmt"
	//"path/filepath"
	"strings"

	"github.com/golang/glog"
	mconfig "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
	hostpath "github.com/openebs/maya/pkg/hostpath/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	//"github.com/pkg/errors"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//storagev1 "k8s.io/api/storage/v1"
)

const (
	//KeyPVStorageType defines if the PV should be backed
	// a hostpath ( sub directory or a storage device)
	KeyPVStorageType = "StorageType"
	//KeyPVBasePath defines base directory for hostpath volumes
	// can be configured via the StorageClass annotations.
	KeyPVBasePath = "BasePath"
	//KeyPVRelativePath defines the alternate folder name under the BasePath
	// By default, the pv name will be used as the folder name.
	// KeyPVBasePath can be useful for providing the same underlying folder
	// name for all replicas in a Statefulset.
	// Will be a property of the PVC annotations.
	KeyPVRelativePath = "RelativePath"
	//KeyPVAbsolutePath specifies a complete hostpath instead of
	// auto-generating using BasePath and RelativePath. This option
	// is specified with PVC and is useful for granting shared access
	// to underlying hostpaths across multiple pods.
	KeyPVAbsolutePath = "AbsolutePath"
)

const (
	// Some of the PVCs launched with older helm charts, still
	// refer to the StorageClass via beta annotations.
	betaStorageClassAnnotation = "volume.beta.kubernetes.io/storage-class"
)

//GetVolumeConfig creates a new VolumeConfig struct by
// parsing and merging the configuration provided in the PVC
// annotation - cas.openebs.io/config with the
// default configuration of the provisioner.
func (p *Provisioner) GetVolumeConfig(pvName string, pvc *v1.PersistentVolumeClaim) (*VolumeConfig, error) {

	pvConfig := p.defaultConfig

	//Fetch the SC
	scName := GetStorageClassName(pvc)
	sc, err := p.kubeClient.StorageV1().StorageClasses().Get(*scName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get storageclass: missing sc name {%v}", scName)
	}

	// extract and merge the cas config from storageclass
	scCASConfigStr := sc.ObjectMeta.Annotations[string(mconfig.CASConfigKey)]
	glog.Infof("SC %v has config:%v", *scName, scCASConfigStr)
	if len(strings.TrimSpace(scCASConfigStr)) != 0 {
		scCASConfig, err := cast.UnMarshallToConfig(scCASConfigStr)
		if err == nil {
			pvConfig = cast.MergeConfig(scCASConfig, pvConfig)
		} else {
			return nil, errors.Wrapf(err, "failed to get config: invalid sc config {%v}", scCASConfigStr)
		}
	}

	//TODO : extract and merge the cas volume config from pvc
	//This block can be added once validation checks are added
	// as to the type of config that can be passed via PVC
	//pvcCASConfigStr := pvc.ObjectMeta.Annotations[string(mconfig.CASConfigKey)]
	//if len(strings.TrimSpace(pvcCASConfigStr)) != 0 {
	//	pvcCASConfig, err := cast.UnMarshallToConfig(pvcCASConfigStr)
	//	if err == nil {
	//		pvConfig = cast.MergeConfig(pvcCASConfig, pvConfig)
	//	}
	//}

	pvConfigMap, err := cast.ConfigToMap(pvConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read volume config: pvc {%v}", pvc.ObjectMeta.Name)
	}

	c := &VolumeConfig{
		pvName:  pvName,
		pvcName: pvc.ObjectMeta.Name,
		scName:  *scName,
		options: pvConfigMap,
	}
	return c, nil
}

//GetStorageType returns the StorageType value configured
// in StorageClass. Default is hostpath
func (c *VolumeConfig) GetStorageType() string {
	stgType := c.getValue(KeyPVStorageType)
	if len(strings.TrimSpace(stgType)) == 0 {
		return "hostpath"
	}
	return stgType
}

//GetPath returns a valid PV path based on the configuration
// or an error. The Path is constructed using the following rules:
// If AbsolutePath is specified return it. (Future)
// If PVPath is specified, suffix it with BasePath and return it. (Future)
// If neither of above are specified, suffix the PVName to BasePath
//  and return it
// Also before returning the path, validate that path is safe
//  and matches the filters specified in StorageClass.
func (c *VolumeConfig) GetPath() (string, error) {
	//This feature need to be supported with some more
	// security checks are in place, so that rouge pods
	// don't get access to node directories.
	//absolutePath := c.getValue(KeyPVAbsolutePath)
	//if len(strings.TrimSpace(absolutePath)) != 0 {
	//	return c.validatePath(absolutePath)
	//}

	basePath := c.getValue(KeyPVBasePath)
	if strings.TrimSpace(basePath) == "" {
		return "", errors.Errorf("failed to get path: base path is empty")
	}

	//This feature need to be supported after the
	// security checks are in place.
	//pvRelPath := c.getValue(KeyPVRelativePath)
	//if len(strings.TrimSpace(pvRelPath)) == 0 {
	//	pvRelPath = c.pvName
	//}

	pvRelPath := c.pvName
	//path := filepath.Join(basePath, pvRelPath)

	return hostpath.NewBuilder().
		WithPathJoin(basePath, pvRelPath).
		WithCheckf(hostpath.IsNonRoot(), "path should not be a root directory: %s/%s", basePath, pvRelPath).
		ValidateAndBuild()
}

//getValue is a utility function to extract the value
// of the `key` from the ConfigMap object - which is
// map[string]interface{map[string][string]}
// Example:
// {
//     key1: {
//             value: value1
//             enabled: true
//           }
// }
// In the above example, if `key1` is passed as input,
//   `value1` will be returned.
func (c *VolumeConfig) getValue(key string) string {
	if configObj, ok := util.GetNestedField(c.options, key).(map[string]string); ok {
		if val, p := configObj[string(mconfig.ValuePTP)]; p {
			return val
		}
	}
	return ""
}

// GetStorageClassName extracts the StorageClass name from PVC
func GetStorageClassName(pvc *v1.PersistentVolumeClaim) *string {
	// Use beta annotation first
	class, found := pvc.Annotations[betaStorageClassAnnotation]
	if found {
		return &class
	}
	return pvc.Spec.StorageClassName
}

// GetLocalPVType extracts the Local PV Type from PV
func GetLocalPVType(pv *v1.PersistentVolume) string {
	casType, found := pv.Labels[string(mconfig.CASTypeKey)]
	if found {
		return casType
	}
	return ""
}
