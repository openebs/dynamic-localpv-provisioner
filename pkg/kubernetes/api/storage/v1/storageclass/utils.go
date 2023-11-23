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

package storageclass

import (
	"path/filepath"
	"regexp"
	"strings"

	mconfig "github.com/openebs/api/v3/pkg/apis/openebs.io/v1alpha1"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

func isValidPath(hostpath string) bool {
	// Is an abolute path
	if !filepath.IsAbs(hostpath) {
		return false
	}

	// IsNotRoot
	path := strings.TrimSuffix(string(hostpath), "/")
	parentDir, subDir := filepath.Split(path)
	parentDir = strings.TrimSuffix(parentDir, "/")
	subDir = strings.TrimSuffix(subDir, "/")
	if parentDir == "" || subDir == "" {
		return false
	}
	return true
}

func isCompatibleWithLocalPVcasType(s *storagev1.StorageClass) bool {
	if scCASTypeStr, ok := s.ObjectMeta.Annotations[string(mconfig.CASTypeKey)]; ok {
		if scCASTypeStr != localPVcasTypeValue && scCASTypeStr != "" {
			return false
		}
	}
	return true
}

// Used to check if the cas.openebs.io/config value string
// has valid parameters for hostpath or not
// e.g.
// Parameters like 'BlockDeviceSelectors', already existing 'StorageType'
// are incompatible.
func isCompatibleWithHostpath(s *storagev1.StorageClass) bool {
	if !isCompatibleWithLocalPVcasType(s) {
		return false
	}

	if scCASConfigStr, ok := s.ObjectMeta.Annotations[string(mconfig.CASConfigKey)]; ok {
		// Unmarshall to mconfig.Config
		scCASConfig, err := cast.UnMarshallToConfig(scCASConfigStr)
		if err != nil {
			return false
		}

		// Check for invalid CAS config parameters
		for _, config := range scCASConfig {
			switch strings.TrimSpace(config.Name) {
			case "NodeAffinityLabel":
				continue
			case "XFSQuota":
				if !isValidQuotaData(config.Data) {
					return false
				}
				continue
			case "EXT4Quota":
				if !isValidQuotaData(config.Data) {
					return false
				}
				continue
			default:
				return false
			}
		}
	}

	if len(s.Provisioner) > 0 && s.Provisioner != localPVprovisionerName {
		return false
	}

	return true
}

func isValidQuotaData(data map[string]string) bool {
	if data == nil {
		return true
	}
	if softLimit, ok := data[KeyQuotaSoftLimit]; ok {
		//Allows values of the following formats:
		// 123.456%, 123%, 123.%, .45%
		//Does not allow:
		// .%, %, ., 1234.45%
		if !regexp.MustCompile("^(^$|(^[0-9]{1,3}([.][0-9]*)?|[.][0-9]+)%)$").MatchString(softLimit) {
			return false
		}
	}
	if hardLimit, ok := data[KeyQuotaHardLimit]; ok {
		//Allows values of the following formats:
		// 123.456%, 123%, 123.%, .45%
		//Does not allow:
		// .%, %, ., 1234.45%
		if !regexp.MustCompile("^(^$|(^[0-9]{1,3}([.][0-9]*)?|[.][0-9]+)%)$").MatchString(hardLimit) {
			return false
		}
	}

	return true
}

func isCompatibleWithQuota(s *storagev1.StorageClass) bool {
	if !isCompatibleWithLocalPVcasType(s) {
		return false
	}

	if scCASConfigStr, ok := s.ObjectMeta.Annotations[string(mconfig.CASConfigKey)]; ok {
		// Unmarshall to mconfig.Config
		scCASConfig, err := cast.UnMarshallToConfig(scCASConfigStr)
		if err != nil {
			return false
		}

		// Check for invalid CAS config parameters
		for _, config := range scCASConfig {
			switch strings.TrimSpace(config.Name) {
			case "StorageType":
				if config.Value == "\"hostpath\"" || config.Value == "hostpath" {
					continue
				} else {
					return false
				}
			case "BasePath":
				if !isValidPath(config.Value) {
					return false
				}
				continue
			case "NodeAffinityLabel":
				continue
			default:
				return false
			}
		}
	}

	if len(s.Provisioner) > 0 && s.Provisioner != localPVprovisionerName {
		return false
	}

	return true
}

func isCompatibleWithNodeAffinityLabel(s *storagev1.StorageClass) bool {
	if !isCompatibleWithLocalPVcasType(s) {
		return false
	}

	if scCASConfigStr, ok := s.ObjectMeta.Annotations[string(mconfig.CASConfigKey)]; ok {
		// Unmarshall to mconfig.Config
		scCASConfig, err := cast.UnMarshallToConfig(scCASConfigStr)
		if err != nil {
			return false
		}

		// Check for invalid CAS config parameters
		for _, config := range scCASConfig {
			switch strings.TrimSpace(config.Name) {
			case "StorageType":
				if config.Value == "\"hostpath\"" || config.Value == "hostpath" || config.Value == "\"device\"" || config.Value == "device" {
					continue
				} else {
					return false
				}
			case "BasePath":
				if !isValidPath(config.Value) {
					return false
				}
				continue
			case "XFSQuota":
				if !isValidQuotaData(config.Data) {
					return false
				}
				continue
			case "EXT4Quota":
				if !isValidQuotaData(config.Data) {
					return false
				}
				continue
			default:
				return false
			}
		}
	}

	if len(s.Provisioner) > 0 && s.Provisioner != localPVprovisionerName {
		return false
	}

	return true
}

func isCompatibleWithDevice(s *storagev1.StorageClass) bool {
	if !isCompatibleWithLocalPVcasType(s) {
		return false
	}

	if scCASConfigStr, ok := s.ObjectMeta.Annotations[string(mconfig.CASConfigKey)]; ok {
		// Unmarshall to mconfig.Config
		scCASConfig, err := cast.UnMarshallToConfig(scCASConfigStr)
		if err != nil {
			return false
		}

		// Check for invalid CAS config parameters
		for _, config := range scCASConfig {
			switch strings.TrimSpace(config.Name) {
			case "BlockDeviceSelectors":
				continue
			case "FSType":
				continue
			default:
				return false
			}
		}
	}

	if len(s.Provisioner) > 0 && s.Provisioner != localPVprovisionerName {
		return false
	}

	return true
}

func isCompatibleWithFSType(s *storagev1.StorageClass) bool {
	if !isCompatibleWithLocalPVcasType(s) {
		return false
	}

	if scCASConfigStr, ok := s.ObjectMeta.Annotations[string(mconfig.CASConfigKey)]; ok {
		// Unmarshall to mconfig.Config
		scCASConfig, err := cast.UnMarshallToConfig(scCASConfigStr)
		if err != nil {
			return false
		}

		// Check for invalid CAS config parameters
		for _, config := range scCASConfig {
			switch strings.TrimSpace(config.Name) {
			case "StorageType":
				if config.Value == "\"device\"" || config.Value == "device" {
					continue
				} else {
					return false
				}
			case "BlockDeviceSelectors":
				continue
			default:
				return false
			}
		}
	}

	if len(s.Provisioner) > 0 && s.Provisioner != localPVprovisionerName {
		return false
	}

	return true
}

func isValidFilesystem(filesystem string) bool {
	switch filesystem {
	case "xfs":
		return true
	case "ext4":
		return true
	default:
		return false
	}
}

func isCompatibleWithBlockDeviceTag(s *storagev1.StorageClass) bool {
	if !isCompatibleWithLocalPVcasType(s) {
		return false
	}

	if scCASConfigStr, ok := s.ObjectMeta.Annotations[string(mconfig.CASConfigKey)]; ok {
		// Unmarshall to mconfig.Config
		scCASConfig, err := cast.UnMarshallToConfig(scCASConfigStr)
		if err != nil {
			return false
		}

		// Check for invalid CAS config parameters
		for _, config := range scCASConfig {
			switch strings.TrimSpace(config.Name) {
			case "StorageType":
				if config.Value == "\"device\"" || config.Value == "device" {
					continue
				} else {
					return false
				}
			case "FSType":
				continue
			default:
				return false
			}
		}
	}

	if len(s.Provisioner) > 0 && s.Provisioner != localPVprovisionerName {
		return false
	}

	return true
}

func writeOrAppendCASConfig(s *storagev1.StorageClass, config string) bool {
	if s.ObjectMeta.Annotations == nil {
		s.ObjectMeta.Annotations = map[string]string{
			string(mconfig.CASConfigKey): config,
		}
		return true
	}
	// Append to the existing CAS config
	if scCASConfig, ok := s.ObjectMeta.Annotations[string(mconfig.CASConfigKey)]; ok {
		s.ObjectMeta.Annotations[string(mconfig.CASConfigKey)] = scCASConfig + config
		return true
	}

	s.ObjectMeta.Annotations[string(mconfig.CASConfigKey)] = config
	return true
}

func appendAllowedTopologies(s *storagev1.StorageClass, allowedTopologies map[string][]string) {
	selectorTerm := corev1.TopologySelectorTerm{}
	for key, values := range allowedTopologies {
		selectorLabelRequirement := corev1.TopologySelectorLabelRequirement{
			Key:    key,
			Values: values,
		}
		selectorTerm.MatchLabelExpressions = append(selectorTerm.MatchLabelExpressions, selectorLabelRequirement)
	}

	s.AllowedTopologies = append(s.AllowedTopologies, selectorTerm)
}
