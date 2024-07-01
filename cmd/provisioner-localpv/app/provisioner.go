/*
This file contains the volume creation and deletion handlers invoked by
the github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/controller.

The handler that are madatory to be implemented:

- Provision - is called by controller to perform custom validation on the PVC
  request and return a valid PV spec. The controller will create the PV object
  using the spec passed to it and bind it to the PVC.

- Delete - is called by controller to perform cleanup tasks on the PV before
  deleting it.

*/

package app

import (
	"context"
	"fmt"
	"strings"

	analytics "github.com/openebs/google-analytics-4/usage"
	"github.com/openebs/maya/pkg/alertlog"
	mconfig "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	pvController "sigs.k8s.io/sig-storage-lib-external-provisioner/v9/controller"
)

const (
	// Ping message
	Ping string = "ping"
	// Heartbeat message.
	Heartbeat string = "heartbeat"
	// DefaultCASType Event application name constant for volume event
	DefaultCASType string = "hostpath-localpv"
	// DefaultUnknownReplicaCount is the default replica count
	DefaultUnknownReplicaCount string = "1"
)

// NewProvisioner will create a new Provisioner object and initialize
//
//	it with global information used across PV create and delete operations.
func NewProvisioner(kubeClient *clientset.Clientset) (*Provisioner, error) {

	namespace := getOpenEBSNamespace() //menv.Get(menv.OpenEBSNamespace)
	if len(strings.TrimSpace(namespace)) == 0 {
		return nil, fmt.Errorf("Cannot start Provisioner: failed to get namespace")
	}

	p := &Provisioner{
		kubeClient:  kubeClient,
		namespace:   namespace,
		helperImage: getDefaultHelperImage(),
		defaultConfig: []mconfig.Config{
			{
				Name:  KeyPVBasePath,
				Value: getDefaultBasePath(),
			},
		},
	}
	p.getVolumeConfig = p.GetVolumeConfig

	return p, nil
}

// SupportsBlock will be used by controller to determine if block mode is
//
//	supported by the host path provisioner.
func (p *Provisioner) SupportsBlock(_ context.Context) bool {
	return true
}

// Provision is invoked by the PVC controller which expect the PV
//
//	to be provisioned and a valid PV spec returned.
func (p *Provisioner) Provision(ctx context.Context, opts pvController.ProvisionOptions) (*v1.PersistentVolume, pvController.ProvisioningState, error) {
	pvc := opts.PVC

	// validate pvc dataSource
	if err := validateVolumeSource(*pvc); err != nil {
		return nil, pvController.ProvisioningFinished, err
	}

	if pvc.Spec.Selector != nil {
		return nil, pvController.ProvisioningFinished, fmt.Errorf("claim.Spec.Selector is not supported")
	}
	for _, accessMode := range pvc.Spec.AccessModes {
		if accessMode != v1.ReadWriteOnce {
			return nil, pvController.ProvisioningFinished, fmt.Errorf("Only support ReadWriteOnce access mode")
		}
	}

	if opts.SelectedNode == nil {
		return nil, pvController.ProvisioningReschedule, fmt.Errorf("configuration error, no node was specified")
	}

	if GetNodeHostname(opts.SelectedNode) == "" {
		return nil, pvController.ProvisioningFinished, fmt.Errorf("configuration error, node{%v} hostname is empty", opts.SelectedNode.Name)
	}

	name := opts.PVName

	// Create a new Config instance for the PV by merging the
	// default configuration with configuration provided
	// via PVC and the associated StorageClass
	pvCASConfig, err := p.getVolumeConfig(ctx, name, pvc)
	if err != nil {
		return nil, pvController.ProvisioningFinished, err
	}

	//TODO: Determine if hostpath or device based Local PV should be created
	stgType := pvCASConfig.GetStorageType()
	size := resource.Quantity{}
	reqMap := pvc.Spec.Resources.Requests
	if reqMap != nil {
		size = pvc.Spec.Resources.Requests["storage"]
	}
	sendEventOrIgnore(pvc.Name, name, size.String(), stgType, analytics.VolumeProvision)

	// todo: Disable the localpv device provisioning for now. Revisit later to remove the code path.

	// EXCEPTION: Block VolumeMode
	if *opts.PVC.Spec.VolumeMode == v1.PersistentVolumeBlock && stgType != "device" {
		return nil, pvController.ProvisioningFinished, fmt.Errorf("PV with BlockMode is not supported with StorageType %v", stgType)
	}

	// StorageType: Hostpath
	if stgType == "hostpath" {
		return p.ProvisionHostPath(ctx, opts, pvCASConfig)
	}
	alertlog.Logger.Errorw("",
		"eventcode", "local.pv.provision.failure",
		"msg", "Failed to provision Local PV",
		"rname", opts.PVName,
		"reason", "StorageType not supported",
		"storagetype", stgType,
	)
	return nil, pvController.ProvisioningFinished, fmt.Errorf("PV with StorageType %v is not supported", stgType)
}

// Delete is invoked by the PVC controller to perform clean-up
//
//	activities before deleteing the PV object. If reclaim policy is
//	set to not-retain, then this function will create a helper pod
//	to delete the host path from the node.
func (p *Provisioner) Delete(ctx context.Context, pv *v1.PersistentVolume) (err error) {
	defer func() {
		err = errors.Wrapf(err, "failed to delete volume %v", pv.Name)
	}()
	//Initiate clean up only when reclaim policy is not retain.
	if pv.Spec.PersistentVolumeReclaimPolicy != v1.PersistentVolumeReclaimRetain {
		//TODO: Determine the type of PV
		pvType := GetLocalPVType(pv)
		size := resource.Quantity{}
		reqMap := pv.Spec.Capacity
		if reqMap != nil {
			size = pv.Spec.Capacity["storage"]
		}

		pvcName := ""
		if pv.Spec.ClaimRef != nil {
			pvcName = pv.Spec.ClaimRef.Name
		}
		sendEventOrIgnore(pvcName, pv.Name, size.String(), pvType, analytics.VolumeDeprovision)
		// todo: Disable the localpv device deprovisioning for now. Revisit later to remove the code path.

		err = p.DeleteHostPath(ctx, pv)
		if err != nil {
			alertlog.Logger.Errorw("",
				"eventcode", "local.pv.delete.failure",
				"msg", "Failed to delete Local PV",
				"rname", pv.Name,
				"reason", "failed to delete host path",
				"storagetype", pvType,
			)
		}
		return err
	}
	klog.Infof("Retained volume %v", pv.Name)
	alertlog.Logger.Infow("",
		"eventcode", "local.pv.delete.success",
		"msg", "Successfully deleted Local PV",
		"rname", pv.Name,
	)
	return nil
}

// sendEventOrIgnore sends anonymous local-pv provision/delete events
func sendEventOrIgnore(pvcName, pvName, capacity, stgType, method string) {
	stgType = "local-" + stgType

	analytics.New().CommonBuild(stgType).ApplicationBuilder().
		SetVolumeName(pvName).
		SetVolumeClaimName(pvcName).
		SetReplicaCount(DefaultUnknownReplicaCount).
		SetCategory(method).
		SetVolumeCapacity(capacity).
		Send()
}

// validateVolumeSource validates datasource field of the pvc.
// - clone - not handled by this provisioner
// - snapshot - not handled by this provisioner
// - volume populator - not handled by this provisioner
func validateVolumeSource(pvc v1.PersistentVolumeClaim) error {
	if pvc.Spec.DataSource != nil {
		// PVC.Spec.DataSource.Name is the name of the VolumeSnapshot or PVC or populator
		if pvc.Spec.DataSource.Name == "" {
			return fmt.Errorf("dataSource name not found for PVC `%s`", pvc.Name)
		}
		switch pvc.Spec.DataSource.Kind {

		// DataSource is snapshot
		case SnapshotKind:
			if *(pvc.Spec.DataSource.APIGroup) != SnapshotAPIGroup {
				return fmt.Errorf("snapshot feature not supported by this provisioner")
			}
			return fmt.Errorf("datasource `%s` of group `%s` is not handled by the provisioner",
				pvc.Spec.DataSource.Kind, *pvc.Spec.DataSource.APIGroup)

		// DataSource is pvc
		case PVCKind:
			return fmt.Errorf("clone feature not supported by this provisioner")

		// Custom DataSource (volume populator)
		default:
			return fmt.Errorf("datasource `%s` of group `%s` is not handled by the provisioner",
				pvc.Spec.DataSource.Kind, *pvc.Spec.DataSource.APIGroup)
		}
	}
	return nil
}
