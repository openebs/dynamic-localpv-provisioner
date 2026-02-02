package app

import (
	"context"
	"fmt"
	"strings"

	analytics "github.com/openebs/google-analytics-4/usage"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	pvController "sigs.k8s.io/sig-storage-lib-external-provisioner/v13/controller"

	mconfig "github.com/openebs/dynamic-localpv-provisioner/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/utils"
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

	// GoogleAnalyticsKey This environment variable is set via env
	GoogleAnalyticsKey string = "OPENEBS_IO_ENABLE_ANALYTICS"
)

// NewProvisioner will create a new Provisioner object and initialize
//
//	it with global information used across PV create and delete operations.
func NewProvisioner(kubeClient kubernetes.Interface) (*Provisioner, error) {

	namespace := getOpenEBSNamespace()
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
// supported by the host path provisioner.
func (p *Provisioner) SupportsBlock(_ context.Context) bool {
	return false
}

// getSelectedNode fetches the Node object for the given node name.
func (p *Provisioner) getSelectedNode(ctx context.Context, nodeName string) (*v1.Node, error) {
	if nodeName == "" {
		return nil, fmt.Errorf("node name is empty")
	}
	return p.kubeClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
}

// Provision is invoked by the PVC controller which expect the PV
//
//	to be provisioned and a valid PV spec returned.
func (p *Provisioner) Provision(ctx context.Context, opts pvController.ProvisionOptions) (*v1.PersistentVolume, pvController.ProvisioningState, error) {
	// Get logger from context for contextual logging
	log := klog.FromContext(ctx)

	pvc := opts.PVC

	// validate pvc dataSource
	if err := validateVolumeSource(*pvc); err != nil {
		return nil, pvController.ProvisioningFinished, err
	}

	if pvc.Spec.Selector != nil && (len(pvc.Spec.Selector.MatchLabels) > 0 || len(pvc.Spec.Selector.MatchExpressions) > 0) {
		return nil, pvController.ProvisioningFinished, fmt.Errorf("claim.Spec.Selector is not supported")
	}
	for _, accessMode := range pvc.Spec.AccessModes {
		if accessMode != v1.ReadWriteOnce {
			return nil, pvController.ProvisioningFinished, fmt.Errorf("Only support ReadWriteOnce access mode")
		}
	}

	if opts.SelectedNodeName == "" {
		return nil, pvController.ProvisioningReschedule, fmt.Errorf("configuration error, no node was specified")
	}

	// Fetch the full Node object since we need labels and taints
	selectedNode, err := p.getSelectedNode(ctx, opts.SelectedNodeName)
	if err != nil {
		return nil, pvController.ProvisioningFinished, fmt.Errorf("failed to get node %s: %v", opts.SelectedNodeName, err)
	}

	if GetNodeHostname(selectedNode) == "" {
		return nil, pvController.ProvisioningFinished, fmt.Errorf("configuration error, node{%v} hostname is empty", selectedNode.Name)
	}

	// In node-deployment mode, only process PVCs scheduled to this node.
	// This prevents multiple provisioner instances from trying to provision the same PVC.
	if p.nodeDeployment {
		// Use the hostname label for consistency with Delete() and node affinity settings
		selectedNodeHostname := GetNodeHostname(selectedNode)
		if selectedNodeHostname == "" {
			// Fallback to node name if hostname label is not present
			selectedNodeHostname = selectedNode.Name
		}
		if selectedNodeHostname != p.nodeName {
			// This PVC is scheduled to a different node, skip it.
			// Return IgnoredError to tell the controller that this provisioner
			// is not responsible for this PVC. The controller will not treat
			// this as a failure and will let another provisioner handle it.
			log.V(4).Info("Skipping PVC: scheduled to different node",
				"pvc", klog.KObj(pvc),
				"scheduledNode", selectedNodeHostname,
				"thisNode", p.nodeName)
			return nil, pvController.ProvisioningFinished, &pvController.IgnoredError{
				Reason: fmt.Sprintf("PVC is scheduled to node %s, not this node %s", selectedNodeHostname, p.nodeName),
			}
		}
		log.Info("Processing PVC on local node",
			"pvc", klog.KObj(pvc),
			"node", p.nodeName)
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
	// VolumeMode is a pointer and can be nil (defaults to Filesystem)
	if opts.PVC.Spec.VolumeMode != nil && *opts.PVC.Spec.VolumeMode == v1.PersistentVolumeBlock && stgType != "device" {
		return nil, pvController.ProvisioningFinished, fmt.Errorf("PV with BlockMode is not supported with StorageType %v", stgType)
	}

	// StorageType: Hostpath
	if stgType == "hostpath" {
		return p.ProvisionHostPath(ctx, opts, pvCASConfig, selectedNode)
	}
	utils.Logger.Errorw("",
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
	// Get logger from context for contextual logging
	log := klog.FromContext(ctx)

	// In node-deployment mode, only process PVs that are on this node
	if p.nodeDeployment {
		// Get the node affinity from the PV
		if pv.Spec.NodeAffinity != nil && pv.Spec.NodeAffinity.Required != nil {
			isLocalNode := false
			var pvNodeName string
			for _, term := range pv.Spec.NodeAffinity.Required.NodeSelectorTerms {
				for _, expr := range term.MatchExpressions {
					if expr.Key == "kubernetes.io/hostname" && expr.Operator == v1.NodeSelectorOpIn {
						for _, value := range expr.Values {
							pvNodeName = value
							if value == p.nodeName {
								isLocalNode = true
								break
							}
						}
					}
				}
			}
			if !isLocalNode {
				// Return IgnoredError to tell the controller that this provisioner
				// is not responsible for this PV. The controller will not treat
				// this as a failure and will let another provisioner handle it.
				log.V(4).Info("Skipping PV deletion: belongs to different node",
					"pv", pv.Name,
					"pvNode", pvNodeName,
					"thisNode", p.nodeName)
				return &pvController.IgnoredError{
					Reason: fmt.Sprintf("PV belongs to node %s, not this node %s", pvNodeName, p.nodeName),
				}
			}
			log.Info("Processing PV deletion on local node",
				"pv", pv.Name,
				"node", p.nodeName)
		}
	}

	// Use defer for error wrapping only after node filtering
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "failed to delete volume %v", pv.Name)
		}
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
			utils.Logger.Errorw("",
				"eventcode", "local.pv.delete.failure",
				"msg", "Failed to delete Local PV",
				"rname", pv.Name,
				"reason", "failed to delete host path",
				"storagetype", pvType,
			)
		}
		return err
	}
	log.Info("Retained volume", "pv", pv.Name)
	utils.Logger.Infow("",
		"eventcode", "local.pv.delete.success",
		"msg", "Successfully deleted Local PV",
		"rname", pv.Name,
	)
	return nil
}

// sendEventOrIgnore sends anonymous local-pv provision/delete events
func sendEventOrIgnore(pvcName, pvName, capacity, stgType, method string) {
	if utils.GoogleAnalyticsEnabled(GoogleAnalyticsKey) {
		stgType = "local-" + stgType

		analytics.New().CommonBuild(stgType).ApplicationBuilder().
			SetVolumeName(pvName).
			SetVolumeClaimName(pvcName).
			SetReplicaCount(DefaultUnknownReplicaCount).
			SetCategory(method).
			SetVolumeCapacity(capacity).
			Send()
	}
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
