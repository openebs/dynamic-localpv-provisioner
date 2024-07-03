package storageclass

import (
	mconfig "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

const (
	localPVcasTypeValue = "local"

	//Provisioner Name
	localPVprovisionerName = "openebs.io/local"

	//The following are imported from mconfig at the moment
	// CASConfigKey = "cas.openebs.io/config"
	// CASTypeKey = "openebs.io/cas-type"

	//These are from 'app' package
	// cmd/provisioner-localpv/app/config.go
	KeyQuotaSoftLimit = "softLimitGrace"
	KeyQuotaHardLimit = "hardLimitGrace"
)

type StorageClassOption func(*storagev1.StorageClass) error

func NewStorageClass(opts ...StorageClassOption) (*storagev1.StorageClass, error) {
	s := &storagev1.StorageClass{}

	var err error
	for _, opt := range opts {
		err = opt(s)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to build StorageClass.")
		}
	}

	return s, nil
}

func WithName(name string) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if len(name) == 0 {
			return errors.New("Failed to set Name. Name is an empty string.")
		}

		s.ObjectMeta.Name = name
		return nil
	}
}

func WithGenerateName(generateName string) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if len(generateName) == 0 {
			return errors.New("Failed to set GenerateName. Name prefix is an empty string.")
		}

		s.ObjectMeta.GenerateName = generateName + "-"
		return nil
	}
}

func WithLabels(labels map[string]string) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if len(labels) == 0 {
			return errors.New("Failed to set Labels. " +
				"Input is invalid.")
		}

		if s.ObjectMeta.Labels == nil {
			s.ObjectMeta.Labels = map[string]string{}
		}
		for key, value := range labels {
			s.ObjectMeta.Labels[key] = value
		}

		return nil
	}
}

func WithAnnotations(annotations map[string]string) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if len(annotations) == 0 {
			return errors.New("Failed to set Annotations. " +
				"Input is invalid.")
		}

		if s.ObjectMeta.Annotations == nil {
			s.ObjectMeta.Annotations = map[string]string{}
		}
		for key, value := range annotations {
			s.ObjectMeta.Annotations[key] = value
		}

		return nil
	}
}

func WithParameters(parameters map[string]string) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if len(parameters) == 0 {
			return errors.New("Failed to set Parameters. " +
				"Input is invalid.")
		}

		if s.Parameters == nil {
			s.Parameters = map[string]string{}
		}
		for key, value := range parameters {
			s.Parameters[key] = value
		}

		return nil
	}
}

func WithLocalPV() StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if _, ok := s.ObjectMeta.Annotations[string(mconfig.CASTypeKey)]; ok {
			return errors.New("Annotation '" + string(mconfig.CASTypeKey) +
				"' is already set.")
		}
		if len(s.Provisioner) > 0 {
			return errors.New("Provisioner name is already set.")
		}

		// Set the cas-type annotation
		if s.ObjectMeta.Annotations == nil {
			s.ObjectMeta.Annotations = map[string]string{}
		}
		s.ObjectMeta.Annotations[string(mconfig.CASTypeKey)] = localPVcasTypeValue
		// Set the provisioner value for
		// openebs-localpv-provisioner PV controller
		s.Provisioner = localPVprovisionerName

		return nil
	}
}

func WithHostpath(hostpathDir string) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		// Check if the path is a valid one
		if !isValidPath(hostpathDir) {
			return errors.New("Invalid hostpath directory. Path" +
				" must be an absolute path and must be a " +
				"directory which is not directly under '/'.")
		}
		// Check for existing CAS config and Provisioner name
		// Check if the existing parameters are usable
		// with "hostpath" StorageType
		if !isCompatibleWithHostpath(s) {
			return errors.New("Failed to set StorageType and BasePath for Hostpath. " +
				"Invalid existing '" + string(mconfig.CASConfigKey) + "' annotation" +
				" parameters or Provisioner name.")
		}

		config := "- name: StorageType\n" +
			"  value: \"hostpath\"\n" +
			"- name: BasePath\n" +
			"  value: \"" + hostpathDir + "\"\n"

		ok := writeOrAppendCASConfig(s, config)
		if !ok {
			return errors.New("Failed to set StorageType and" +
				" BasePath parameters for Hostpath.")
		}
		return nil
	}
}

func WithDevice() StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		// Check for existing CAS config and Provisioner name
		// Check if the existing parameters are usable
		// with "device" StorageType
		if !isCompatibleWithDevice(s) {
			return errors.New("Failed to set StorageType for Device. " +
				"Invalid existing '" + string(mconfig.CASConfigKey) +
				"' annotaion parameters or Provisioner name.")
		}

		config := "- name: StorageType\n" +
			"  value: \"device\"\n"

		ok := writeOrAppendCASConfig(s, config)
		if !ok {
			return errors.New("Failed to set StorageType parameter for Device.")
		}

		return nil
	}
}

func WithXfsQuota(softLimit, hardLimit string) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if !isCompatibleWithQuota(s) {
			return errors.New("Failed to set XFSQuota parameters. " +
				"Invalid existing '" + string(mconfig.CASConfigKey) + "' annotation" +
				" parameters or Provisioner name.")
		}

		// TODO: Refactor this code

		config := "- name: XFSQuota\n" +
			"  enabled: \"true\"\n"

		if len(softLimit) > 0 || len(hardLimit) > 0 {
			if !isValidQuotaData(map[string]string{
				KeyQuotaSoftLimit: softLimit,
				KeyQuotaHardLimit: hardLimit,
			}) {
				return errors.New("Failed to set XFSQuota parameters. " +
					"Invalid " + KeyQuotaSoftLimit + " and " +
					KeyQuotaHardLimit + " values")
			}

			config = config +
				"  data:\n" +
				"    " + KeyQuotaSoftLimit + ": \"" + softLimit + "\"\n" +
				"    " + KeyQuotaHardLimit + ": \"" + hardLimit + "\"\n"
		}

		ok := writeOrAppendCASConfig(s, config)
		if !ok {
			return errors.New("Failed to set XFSQuota parameters")
		}
		return nil
	}
}

func WithExt4Quota(softLimit, hardLimit string) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if !isCompatibleWithQuota(s) {
			return errors.New("Failed to set EXT4Quota parameters. " +
				"Invalid existing '" + string(mconfig.CASConfigKey) + "' annotation" +
				" parameters or Provisioner name.")
		}

		// TODO: Refactor this code

		config := "- name: EXT4Quota\n" +
			"  enabled: \"true\"\n"

		if len(softLimit) > 0 || len(hardLimit) > 0 {
			if !isValidQuotaData(map[string]string{
				KeyQuotaSoftLimit: softLimit,
				KeyQuotaHardLimit: hardLimit,
			}) {
				return errors.New("Failed to set EXT4Quota parameters. " +
					"Invalid " + KeyQuotaSoftLimit + " and " +
					KeyQuotaHardLimit + " values")
			}

			config = config +
				"  data:\n" +
				"    " + KeyQuotaSoftLimit + ": \"" + softLimit + "\"\n" +
				"    " + KeyQuotaHardLimit + ": \"" + hardLimit + "\"\n"
		}

		ok := writeOrAppendCASConfig(s, config)
		if !ok {
			return errors.New("Failed to set EXT4Quota parameters")
		}
		return nil
	}
}

func WithVolumeBindingMode(volBindingMode storagev1.VolumeBindingMode) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if len(volBindingMode) == 0 {
			volBindingMode = "WaitForFirstConsumer"
		}

		s.VolumeBindingMode = &volBindingMode
		return nil
	}
}

func WithReclaimPolicy(reclaimPolicy corev1.PersistentVolumeReclaimPolicy) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if len(reclaimPolicy) == 0 {
			reclaimPolicy = "Delete"
		}

		s.ReclaimPolicy = &reclaimPolicy
		return nil
	}
}

func WithAllowedTopologies(allowedTopologies map[string][]string) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if len(allowedTopologies) == 0 {
			return errors.New("Failed to set AllowedTopologies. " +
				"Input is invalid.")
		}

		appendAllowedTopologies(s, allowedTopologies)
		return nil
	}
}

func WithNodeAffinityLabels(nodeLabelKeys []string) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if len(nodeLabelKeys) == 0 {
			return errors.New("Failed to set NodeLabelKey. " +
				"Input is invalid.")
		}

		// Check if the existing parameters and Provisioner name
		// are usable with NodeAffnityLabel.
		if !isCompatibleWithNodeAffinityLabel(s) {
			return errors.New("Failed to set NodeAffinityLabel. " +
				"Invalid existing '" + string(mconfig.CASConfigKey) +
				"' annotaion parameters or Provisioner name.")
		}

		// labelKeys stores all the node label keys in a yaml list format
		labelKeys := ""

		for _, nodeLabelKey := range nodeLabelKeys {
			if len(nodeLabelKey) != 0 {
				labelKeys = labelKeys + "    - \"" + nodeLabelKey + "\"\n"
			}
		}

		if labelKeys == "" {
			return errors.New("Failed to set NodeLabelKey. " +
				"Input is invalid.")
		}

		config := "- name: NodeAffinityLabels\n" +
			"  list:\n" +
			labelKeys

		ok := writeOrAppendCASConfig(s, config)
		if !ok {
			return errors.New("Failed to set NodeAffinityLabel" +
				" parameter for Hostpath.")
		}
		return nil
	}
}

func WithFSType(filesystem string) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if !isValidFilesystem(filesystem) {
			return errors.New("Filesystem is invalid. " +
				"Accepted values are \"ext4\" and \"xfs\".")
		}

		// Check if the existing parameters and
		// Provisioner name are usable with FSType.
		// FSType is only compatible with
		// Device StorageType.
		if !isCompatibleWithFSType(s) {
			return errors.New("Failed to set FSType. " +
				"Invalid existing '" + string(mconfig.CASConfigKey) +
				"' annotation parameters or Provisioner name")
		}

		config := "- name: FSType\n" +
			"  value: \"" + filesystem + "\"\n"

		ok := writeOrAppendCASConfig(s, config)
		if !ok {
			return errors.New("Failed to set FSType" +
				" parameter for Device.")
		}
		return nil
	}
}

func WithBlockDeviceSelectors(bdSelectors map[string]string) StorageClassOption {
	return func(s *storagev1.StorageClass) error {
		if len(bdSelectors) == 0 {
			return errors.New("Failed to set BlockDeviceSelectors. " +
				"Input is invalid.")
		}

		// Check if the existing parameters and Provisioner name
		// are usable with BlockDeviceSelectors.
		// BlockDeviceSelectors is only compatible with
		// Device StorageType.
		if !isCompatibleWithBlockDeviceTag(s) {
			return errors.New("Failed to set BlockDeviceTag. " +
				"Invalid existing '" + string(mconfig.CASConfigKey) +
				"' annotaion parameters or Provisioner name.")
		}

		// bdSelectorMap
		bdSelectorMap := ""

		for key, value := range bdSelectors {
			if len(value) != 0 {
				bdSelectorMap = bdSelectorMap + "    \"" + key + "\": \"" + value + "\"\n"
			}
		}

		if bdSelectorMap == "" {
			return errors.New("Failed to set BlockDeviceSelectors. " +
				"Input is invalid.")
		}

		config := "- name: BlockDeviceSelectors\n" +
			"  data:\n" +
			bdSelectorMap

		ok := writeOrAppendCASConfig(s, config)
		if !ok {
			return errors.New("Failed to set BlockDeviceTag" +
				" parameter for Device.")
		}
		return nil
	}
}
