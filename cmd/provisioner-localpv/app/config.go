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
	"strconv"
	"strings"

	mconfig "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
	hostpath "github.com/openebs/maya/pkg/hostpath/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	//KeyPVStorageType defines if the PV should be backed
	// a hostpath ( sub directory or a storage device)
	KeyPVStorageType = "StorageType"

	//KeyPVBasePath defines base directory for hostpath volumes
	// can be configured via the StorageClass annotations.
	KeyPVBasePath = "BasePath"

	//KeyPVFSType defines filesystem type to be used with devices
	// and can be configured via the StorageClass annotations.
	KeyPVFSType = "FSType"

	// NOTE: This key should not be used as it is deprecated.
	//        Instead use "KeyBlockDeviceSelectors" key
	KeyBDTag = "BlockDeviceTag"

	//KeyBlockDeviceSelectors defines the value for the Block Device selectors
	//during bdc to bd claim configured via the StorageClass annotations.
	// NOTE: This key should not be used as it is deprecated.
	KeyNodeAffinityLabel = "NodeAffinityLabel"

	//KeyNodeAffinityLabels defines the label keys that should be
	//used in the nodeAffinitySpec.
	//
	//Example: Local PV device StorageClass for selecting devices
	//of SSD type and no filesystem present on it will be as follows
	//
	// kind: StorageClass
	// metadata:
	//   name: local-device
	//   annotations:
	//     openebs.io/cas-type: local
	//     cas.openebs.io/config: |
	//       - name: StorageType
	//         value: "device"
	//       - name: NodeAffinityLabels
	//         list:
	//           - "openebs.io/node-affinity-value-1"
	//           - "openebs.io/node-affinity-value-2"
	KeyNodeAffinityLabels = "NodeAffinityLabels"

	//KeyBlockDeviceSelectors defines the value for the Block Device selectors
	//during bdc to bd claim configured via the StorageClass annotations.
	//
	//Example: Local PV device StorageClass for selecting devices
	//of SSD type and no filesystem present on it will be as follows
	//
	// kind: StorageClass
	// metadata:
	//   name: local-device
	//   annotations:
	//     openebs.io/cas-type: local
	//     cas.openebs.io/config: |
	//       - name: StorageType
	//         value: "device"
	//       - name: BlockDeviceSelectors
	//         data:
	//           ndm.io/driveType: "SSD"
	//           ndm.io/fsType: "none"
	// provisioner: openebs.io/local
	// volumeBindingMode: WaitForFirstConsumer
	// reclaimPolicy: Delete
	//
	KeyBlockDeviceSelectors = "BlockDeviceSelectors"

	//KeyPVRelativePath defines the alternate folder name under the BasePath
	// By default, the pv name will be used as the folder name.
	// KeyPVBasePath can be useful for providing the same underlying folder
	// name for all replicas in a Statefulset.
	// Will be a property of the PVC annotations.
	//KeyPVRelativePath = "RelativePath"
	//KeyPVAbsolutePath specifies a complete hostpath instead of
	// auto-generating using BasePath and RelativePath. This option
	// is specified with PVC and is useful for granting shared access
	// to underlying hostpaths across multiple pods.
	//KeyPVAbsolutePath = "AbsolutePath"

	//KeyXFSQuota enables/sets parameters for XFS Quota.
	// Example StorageClass snippet:
	//    - name: XFSQuota
	//      enabled: true
	//      data:
	//        softLimitGrace: "80%"
	//        hardLimitGrace: "85%"
	KeyXFSQuota = "XFSQuota"

	//KeyEXT4Quota enables/sets parameters for EXT4 Quota.
	// Example StorageClass snippet:
	//    - name: EXT4Quota
	//      enabled: true
	//      data:
	//        softLimitGrace: "80%"
	//        hardLimitGrace: "85%"
	KeyEXT4Quota = "EXT4Quota"

	KeyQuotaSoftLimit = "softLimitGrace"
	KeyQuotaHardLimit = "hardLimitGrace"
)

const (
	// Some of the PVCs launched with older helm charts, still
	// refer to the StorageClass via beta annotations.
	betaStorageClassAnnotation = "volume.beta.kubernetes.io/storage-class"

	// k8sNodeLabelKeyHostname is the label key used by Kubernetes
	// to store the hostname on the node resource.
	k8sNodeLabelKeyHostname = "kubernetes.io/hostname"
)

//GetVolumeConfig creates a new VolumeConfig struct by
// parsing and merging the configuration provided in the PVC/SC
// annotation - cas.openebs.io/config with the
// default configuration of the provisioner.
func (p *Provisioner) GetVolumeConfig(ctx context.Context, pvName string, pvc *corev1.PersistentVolumeClaim) (*VolumeConfig, error) {

	pvConfig := p.defaultConfig

	//Fetch the SC
	scName := GetStorageClassName(pvc)
	sc, err := p.kubeClient.StorageV1().StorageClasses().Get(ctx, *scName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get storageclass: missing sc name {%v}", scName)
	}

	// extract and merge the cas config from storageclass
	scCASConfigStr := sc.ObjectMeta.Annotations[string(mconfig.CASConfigKey)]
	klog.V(4).Infof("SC %v has config:%v", *scName, scCASConfigStr)
	if len(strings.TrimSpace(scCASConfigStr)) != 0 {
		scCASConfig, err := cast.UnMarshallToConfig(scCASConfigStr)
		if err == nil {
			pvConfig = cast.MergeConfig(scCASConfig, pvConfig)
		} else {
			return nil, errors.Wrapf(err, "failed to get config: invalid sc config {%v}", scCASConfigStr)
		}
	}

	// extract and merge the cas config from persistentvolumeclaim
	pvcCASConfigStr := pvc.Annotations[string(mconfig.CASConfigKey)]
	klog.V(4).Infof("PVC %v has config:%v", pvc.Name, pvcCASConfigStr)
	if len(strings.TrimSpace(pvcCASConfigStr)) != 0 {
		pvcCASConfig, err := cast.UnMarshallToConfig(pvcCASConfigStr)
		if err == nil {
			pvConfig = cast.MergeConfig(pvcCASConfig, pvConfig)
		} else {
			return nil, errors.Wrapf(err, "failed to get config: invalid pvc config {%v}", pvcCASConfigStr)
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

	dataPvConfigMap, err := dataConfigToMap(pvConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read volume config: pvc {%v}", pvc.ObjectMeta.Name)
	}

	listPvConfigMap, err := listConfigToMap(pvConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read volume config: pvc {%v}", pvc.ObjectMeta.Name)
	}

	c := &VolumeConfig{
		pvName:     pvName,
		pvcName:    pvc.ObjectMeta.Name,
		scName:     *scName,
		options:    pvConfigMap,
		configData: dataPvConfigMap,
		configList: listPvConfigMap,
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

//GetBlockDeviceSelectors returns the BlockDeviceSelectors data configured
// in StorageClass. Default is nil
func (c *VolumeConfig) GetBlockDeviceSelectors() map[string]string {
	blockDeviceSelector := c.getData(KeyBlockDeviceSelectors)
	return blockDeviceSelector
}

// NOTE: This function should not be used, as KeyBDTag has been deprecated.
//       GetBlockDeviceSelectors() is the right function to use.
func (c *VolumeConfig) GetBDTagValue() string {
	bdTagValue := c.getValue(KeyBDTag)
	if len(strings.TrimSpace(bdTagValue)) == 0 {
		return ""
	}
	return bdTagValue
}

//GetFSType returns the FSType value configured
// in StorageClass. Default is "", auto-determined
// by Local PV
func (c *VolumeConfig) GetFSType() string {
	fsType := c.getValue(KeyPVFSType)
	if len(strings.TrimSpace(fsType)) == 0 {
		return ""
	}
	return fsType
}

// NOTE: This function should not be used, as NodeAffinityLabel has been deprecated.
// GetNodeAffinityLabelKeys() is the right function to use.
func (c *VolumeConfig) GetNodeAffinityLabelKey() string {
	nodeAffinityLabelKey := c.getValue(KeyNodeAffinityLabel)
	if len(strings.TrimSpace(nodeAffinityLabelKey)) == 0 {
		return ""
	}
	return nodeAffinityLabelKey
}

//GetNodeAffinityLabelKey returns the custom node affinity
//label keys as configured in StorageClass.
//
//Default is nil.
func (c *VolumeConfig) GetNodeAffinityLabelKeys() []string {
	nodeAffinityLabelKeys := c.getList(KeyNodeAffinityLabels)
	if nodeAffinityLabelKeys == nil {
		return nil
	}
	return nodeAffinityLabelKeys
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

func (c *VolumeConfig) IsXfsQuotaEnabled() bool {
	xfsQuotaEnabled := c.getEnabled(KeyXFSQuota)
	xfsQuotaEnabled = strings.TrimSpace(xfsQuotaEnabled)

	enableXfsQuotaBool, err := strconv.ParseBool(xfsQuotaEnabled)
	//Default case
	// this means that we have hit either of the two cases below:
	//     i. The value was something other than a straightforward
	//        true or false
	//    ii. The value was empty
	if err != nil {
		return false
	}

	return enableXfsQuotaBool
}

func (c *VolumeConfig) IsExt4QuotaEnabled() bool {
	ext4QuotaEnabled := c.getEnabled(KeyEXT4Quota)
	ext4QuotaEnabled = strings.TrimSpace(ext4QuotaEnabled)

	enableExt4QuotaBool, err := strconv.ParseBool(ext4QuotaEnabled)
	//Default case
	// this means that we have hit either of the two cases below:
	//     i. The value was something other than a straightforward
	//        true or false
	//    ii. The value was empty
	if err != nil {
		return false
	}

	return enableExt4QuotaBool
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

//Similar to getValue() above. Returns value of the
// 'Enabled' parameter.
func (c *VolumeConfig) getEnabled(key string) string {
	if configObj, ok := util.GetNestedField(c.options, key).(map[string]string); ok {
		if val, p := configObj[string(mconfig.EnabledPTP)]; p {
			return val
		}
	}
	return ""
}

//This is similar to getValue() and getEnabled().
// This gets the value for a specific
// 'Data' parameter key-value pair.
func (c *VolumeConfig) getDataField(key string, dataKey string) string {
	if configData, ok := util.GetNestedField(c.configData, key).(map[string]string); ok {
		if val, p := configData[dataKey]; p {
			return val
		}
	}
	//Default case
	return ""
}

//This is similar to getValue() and getEnabled().
// This returns the value of the `Data` parameter
func (c *VolumeConfig) getData(key string) map[string]string {
	if configData, ok := util.GetNestedField(c.configData, key).(map[string]string); ok {
		return configData
	}
	//Default case
	return nil
}

// This gets the list of values for the 'List' parameter.
func (c *VolumeConfig) getList(key string) []string {
	if listValues, ok := util.GetNestedField(c.configList, key).([]string); ok {
		return listValues
	}
	//Default case
	return nil
}

// GetStorageClassName extracts the StorageClass name from PVC
func GetStorageClassName(pvc *corev1.PersistentVolumeClaim) *string {
	// Use beta annotation first
	class, found := pvc.Annotations[betaStorageClassAnnotation]
	if found {
		return &class
	}
	return pvc.Spec.StorageClassName
}

// GetLocalPVType extracts the Local PV Type from PV
func GetLocalPVType(pv *corev1.PersistentVolume) string {
	casType, found := pv.Labels[string(mconfig.CASTypeKey)]
	if found {
		return casType
	}
	return ""
}

// GetNodeHostname extracts the Hostname from the labels on the Node
// If hostname label `kubernetes.io/hostname` is not present
// an empty string is returned.
func GetNodeHostname(n *corev1.Node) string {
	hostname, found := n.Labels[k8sNodeLabelKeyHostname]
	if !found {
		return ""
	}
	return hostname
}

// GetNodeLabelValue extracts the value from the given label on the Node
// If specificed label is not present an empty string is returned.
func GetNodeLabelValue(n *corev1.Node, labelKey string) string {
	labelValue, found := n.Labels[labelKey]
	if !found {
		return ""
	}
	return labelValue
}

// GetTaints extracts the Taints from the Spec on the node
// If Taints are empty, it just returns empty structure of corev1.Taints
func GetTaints(n *corev1.Node) []corev1.Taint {
	return n.Spec.Taints
}

// GetImagePullSecrets  parse image pull secrets from env
// transform  string to corev1.LocalObjectReference
// multiple secrets are separated by commas
func GetImagePullSecrets(s string) []corev1.LocalObjectReference {
	s = strings.TrimSpace(s)
	list := make([]corev1.LocalObjectReference, 0)
	if len(s) == 0 {
		return list
	}
	arr := strings.Split(s, ",")
	for _, item := range arr {
		if len(item) > 0 {
			l := corev1.LocalObjectReference{Name: strings.TrimSpace(item)}
			list = append(list, l)
		}
	}
	return list
}

func dataConfigToMap(pvConfig []mconfig.Config) (map[string]interface{}, error) {
	m := map[string]interface{}{}

	for _, configObj := range pvConfig {
		//No Data Parameter
		if configObj.Data == nil {
			continue
		}

		configName := strings.TrimSpace(configObj.Name)
		confHierarchy := map[string]interface{}{
			configName: configObj.Data,
		}
		isMerged := util.MergeMapOfObjects(m, confHierarchy)
		if !isMerged {
			return nil, errors.Errorf("failed to transform cas config 'Data' for configName '%s' to map: failed to merge: %s", configName, configObj)
		}
	}

	return m, nil
}

func listConfigToMap(pvConfig []mconfig.Config) (map[string]interface{}, error) {
	m := map[string]interface{}{}

	for _, configObj := range pvConfig {
		//No List Parameter
		if len(configObj.List) == 0 {
			continue
		}

		configName := strings.TrimSpace(configObj.Name)
		confHierarchy := map[string]interface{}{
			configName: configObj.List,
		}
		isMerged := util.MergeMapOfObjects(m, confHierarchy)
		if !isMerged {
			return nil, errors.Errorf("failed to transform cas config 'List' for configName '%s' to map: failed to merge: %s", configName, configObj)
		}
	}

	return m, nil
}
