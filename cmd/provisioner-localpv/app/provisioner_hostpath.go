package app

import (
	"context"

	"github.com/openebs/maya/pkg/alertlog"
	mconfig "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	pvController "sigs.k8s.io/sig-storage-lib-external-provisioner/v9/controller"

	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/persistentvolume"
)

const (
	EnableXfsQuota  string = "enableXfsQuota"
	EnableExt4Quota string = "enableExt4Quota"
	SoftLimitGrace  string = "softLimitGrace"
	HardLimitGrace  string = "hardLimitGrace"
)

// ProvisionHostPath is invoked by the Provisioner which expect HostPath PV
//
//	to be provisioned and a valid PV spec returned.
func (p *Provisioner) ProvisionHostPath(ctx context.Context, opts pvController.ProvisionOptions, volumeConfig *VolumeConfig) (*v1.PersistentVolume, pvController.ProvisioningState, error) {
	pvc := opts.PVC
	taints := GetTaints(opts.SelectedNode)
	name := opts.PVName
	stgType := volumeConfig.GetStorageType()
	saName := getOpenEBSServiceAccountName()

	// nodeAffinityLabels
	nodeAffinityLabels := make(map[string]string)

	nodeAffinityKeys := volumeConfig.GetNodeAffinityLabelKeys()
	if nodeAffinityKeys == nil {
		nodeAffinityLabels[k8sNodeLabelKeyHostname] = GetNodeLabelValue(opts.SelectedNode, k8sNodeLabelKeyHostname)
	} else {
		for _, nodeAffinityKey := range nodeAffinityKeys {
			nodeAffinityLabels[nodeAffinityKey] = GetNodeLabelValue(opts.SelectedNode, nodeAffinityKey)
		}
	}

	path, err := volumeConfig.GetPath()
	if err != nil {
		alertlog.Logger.Errorw("",
			"eventcode", "local.pv.provision.failure",
			"msg", "Failed to provision Local PV",
			"rname", opts.PVName,
			"reason", "Unable to get volume config",
			"storagetype", stgType,
		)
		return nil, pvController.ProvisioningFinished, err
	}

	imagePullSecrets := GetImagePullSecrets(getOpenEBSImagePullSecrets())

	hostNetwork := getHelperPodHostNetwork()

	klog.Infof("Creating volume %v at node with labels {%v}, path:%v,ImagePullSecrets:%v", name, nodeAffinityLabels, path, imagePullSecrets)

	//Before using the path for local PV, make sure it is created.
	initCmdsForPath := []string{"mkdir", "-m", "0777", "-p"}
	podOpts := &HelperPodOptions{
		cmdsForPath:        initCmdsForPath,
		name:               name,
		path:               path,
		nodeAffinityLabels: nodeAffinityLabels,
		serviceAccountName: saName,
		selectedNodeTaints: taints,
		imagePullSecrets:   imagePullSecrets,
		hostNetwork:        hostNetwork,
	}
	iErr := p.createInitPod(ctx, podOpts)
	if iErr != nil {
		klog.Infof("Initialize volume %v failed: %v", name, iErr)
		alertlog.Logger.Errorw("",
			"eventcode", "local.pv.provision.failure",
			"msg", "Failed to provision Local PV",
			"rname", opts.PVName,
			"reason", "Volume initialization failed",
			"storagetype", stgType,
		)
		return nil, pvController.ProvisioningFinished, iErr
	}

	if volumeConfig.IsXfsQuotaEnabled() {
		softLimitGrace := volumeConfig.getDataField(KeyXFSQuota, KeyQuotaSoftLimit)
		hardLimitGrace := volumeConfig.getDataField(KeyXFSQuota, KeyQuotaHardLimit)
		pvcStorage := opts.PVC.Spec.Resources.Requests.Storage().Value()

		podOpts := &HelperPodOptions{
			name:               name,
			path:               path,
			nodeAffinityLabels: nodeAffinityLabels,
			serviceAccountName: saName,
			selectedNodeTaints: taints,
			imagePullSecrets:   imagePullSecrets,
			softLimitGrace:     softLimitGrace,
			hardLimitGrace:     hardLimitGrace,
			pvcStorage:         pvcStorage,
			hostNetwork:        hostNetwork,
		}
		iErr := p.createQuotaPod(ctx, podOpts)
		if iErr != nil {
			klog.Infof("Applying quota failed: %v", iErr)
			alertlog.Logger.Errorw("",
				"eventcode", "local.pv.provision.failure",
				"msg", "Failed to provision Local PV",
				"rname", opts.PVName,
				"reason", "Quota enforcement failed",
				"storagetype", stgType,
			)
			return nil, pvController.ProvisioningFinished, iErr
		}
		alertlog.Logger.Infow("",
			"eventcode", "local.pv.quota.success",
			"msg", "Successfully applied quota",
			"rname", opts.PVName,
			"storagetype", stgType,
		)
	}

	if volumeConfig.IsExt4QuotaEnabled() {
		softLimitGrace := volumeConfig.getDataField(KeyEXT4Quota, KeyQuotaSoftLimit)
		hardLimitGrace := volumeConfig.getDataField(KeyEXT4Quota, KeyQuotaHardLimit)
		pvcStorage := opts.PVC.Spec.Resources.Requests.Storage().Value()

		podOpts := &HelperPodOptions{
			name:               name,
			path:               path,
			nodeAffinityLabels: nodeAffinityLabels,
			serviceAccountName: saName,
			selectedNodeTaints: taints,
			imagePullSecrets:   imagePullSecrets,
			softLimitGrace:     softLimitGrace,
			hardLimitGrace:     hardLimitGrace,
			pvcStorage:         pvcStorage,
			hostNetwork:        hostNetwork,
		}
		iErr := p.createQuotaPod(ctx, podOpts)
		if iErr != nil {
			klog.Infof("Applying quota failed: %v", iErr)
			alertlog.Logger.Errorw("",
				"eventcode", "local.pv.provision.failure",
				"msg", "Failed to provision Local PV",
				"rname", opts.PVName,
				"reason", "Quota enforcement failed",
				"storagetype", stgType,
			)
			return nil, pvController.ProvisioningFinished, iErr
		}
		alertlog.Logger.Infow("",
			"eventcode", "local.pv.quota.success",
			"msg", "Successfully applied quota",
			"rname", opts.PVName,
			"storagetype", stgType,
		)
	}

	// VolumeMode will always be specified as Filesystem for host path volume,
	// and the value passed in from the PVC spec will be ignored.
	fs := v1.PersistentVolumeFilesystem

	// It is possible that the HostPath doesn't already exist on the node.
	// Set the Local PV to create it.
	//hostPathType := v1.HostPathDirectoryOrCreate

	// TODO initialize the Labels and annotations
	// Use annotations to specify the context using which the PV was created.
	//volAnnotations := make(map[string]string)
	//volAnnotations[string(v1alpha1.CASTypeKey)] = casVolume.Spec.CasType
	//fstype := casVolume.Spec.FSType

	labels := make(map[string]string)
	labels[string(mconfig.CASTypeKey)] = "local-" + stgType
	//labels[string(v1alpha1.StorageClassKey)] = *className

	//TODO Change the following to a builder pattern
	pvObj, err := persistentvolume.NewBuilder().
		WithName(name).
		WithLabels(labels).
		WithReclaimPolicy(*opts.StorageClass.ReclaimPolicy).
		WithAccessModes(pvc.Spec.AccessModes).
		WithVolumeMode(fs).
		WithCapacityQty(pvc.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]).
		WithLocalHostDirectory(path).
		WithNodeAffinity(nodeAffinityLabels).
		Build()

	if err != nil {
		alertlog.Logger.Errorw("",
			"eventcode", "local.pv.provision.failure",
			"msg", "Failed to provision Local PV",
			"rname", opts.PVName,
			"reason", "failed to build persistent volume",
			"storagetype", stgType,
		)
		return nil, pvController.ProvisioningFinished, err
	}
	alertlog.Logger.Infow("",
		"eventcode", "local.pv.provision.success",
		"msg", "Successfully provisioned Local PV",
		"rname", opts.PVName,
		"storagetype", stgType,
	)
	return pvObj, pvController.ProvisioningFinished, nil
}

// GetNodeObjectFromLabels returns the Node Object with matching label key and value
func (p *Provisioner) GetNodeObjectFromLabels(nodeLabels map[string]string) (*v1.Node, error) {
	labelSelector := metav1.LabelSelector{MatchLabels: nodeLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}
	nodeList, err := p.kubeClient.CoreV1().Nodes().List(context.TODO(), listOptions)
	if err != nil || len(nodeList.Items) == 0 {
		// After the PV is created and node affinity is set
		// based on kubernetes.io/hostname label, either:
		// - hostname label changed on the node or
		// - the node is deleted from the cluster.
		return nil, errors.Errorf("Unable to get the Node with the Node Labels {%v}", nodeLabels)
	}
	if len(nodeList.Items) != 1 {
		// After the PV is created and node affinity is set
		// on a custom affinity label, there may be a transitory state
		// with two nodes matching (old and new) label.
		return nil, errors.Errorf("Unable to determine the Node. Found multiple nodes matching the labels {%v}", nodeLabels)
	}
	return &nodeList.Items[0], nil

}

// DeleteHostPath is invoked by the PVC controller to perform clean-up
//
//	activities before deleteing the PV object. If reclaim policy is
//	set to not-retain, then this function will create a helper pod
//	to delete the host path from the node.
func (p *Provisioner) DeleteHostPath(ctx context.Context, pv *v1.PersistentVolume) (err error) {
	defer func() {
		err = errors.Wrapf(err, "failed to delete volume %v", pv.Name)
	}()

	saName := getOpenEBSServiceAccountName()
	//Determine the path and node of the Local PV.
	pvObj := persistentvolume.NewForAPIObject(pv)
	path := pvObj.GetPath()
	if path == "" {
		return errors.Errorf("no HostPath set")
	}

	nodeAffinityLabels := pvObj.GetAffinitedNodeLabels()
	if len(nodeAffinityLabels) == 0 {
		return errors.Errorf("cannot find affinited node details")
	}
	alertlog.Logger.Infof("Get the Node Object with label {%v}", nodeAffinityLabels)

	//Get the node Object once again to get updated Taints.
	nodeObject, err := p.GetNodeObjectFromLabels(nodeAffinityLabels)
	if err != nil {
		return err
	}
	taints := GetTaints(nodeObject)

	imagePullSecrets := GetImagePullSecrets(getOpenEBSImagePullSecrets())

	hostNetwork := getHelperPodHostNetwork()

	//Initiate clean up only when reclaim policy is not retain.
	klog.Infof("Deleting volume %v at %v:%v", pv.Name, GetNodeHostname(nodeObject), path)
	cleanupCmdsForPath := []string{"rm", "-rf"}
	podOpts := &HelperPodOptions{
		cmdsForPath:        cleanupCmdsForPath,
		name:               pv.Name,
		path:               path,
		nodeAffinityLabels: nodeAffinityLabels,
		serviceAccountName: saName,
		selectedNodeTaints: taints,
		imagePullSecrets:   imagePullSecrets,
		hostNetwork:        hostNetwork,
	}

	if err := p.createCleanupPod(ctx, podOpts); err != nil {
		return errors.Wrapf(err, "clean up volume %v failed", pv.Name)
	}
	return nil
}
