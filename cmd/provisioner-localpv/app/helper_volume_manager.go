package app

import (
	"context"

	"k8s.io/klog/v2"
)

// newLocalVolumeRequest builds the VolumeRequest used for local volume creation
// from the helper pod options. The FsMode is taken from the configured
// FilePermissions mode (pOpts.fsMode); an empty value lets the volume manager
// fall back to its default.
func newLocalVolumeRequest(pOpts *HelperPodOptions) *VolumeRequest {
	return &VolumeRequest{
		Name:           pOpts.name,
		Path:           pOpts.path,
		FsMode:         pOpts.fsMode,
		SoftLimitGrace: pOpts.softLimitGrace,
		HardLimitGrace: pOpts.hardLimitGrace,
		PVCStorage:     pOpts.pvcStorage,
	}
}

// createVolumeLocally performs volume creation directly on the local node
func (p *Provisioner) createVolumeLocally(ctx context.Context, pOpts *HelperPodOptions, enableQuota bool) error {
	log := klog.FromContext(ctx)
	log.Info("Creating volume locally", "volume", pOpts.name)

	// Use the provisioner's volume manager so that its mutex serializes local
	// operations across concurrent requests
	vm := &p.localVolumeManager

	req := newLocalVolumeRequest(pOpts)

	return vm.CreateVolume(ctx, req, enableQuota)
}

// deleteVolumeLocally deletes volume directly on the local node
func (p *Provisioner) deleteVolumeLocally(ctx context.Context, pOpts *HelperPodOptions) error {
	log := klog.FromContext(ctx)
	log.Info("Deleting volume locally", "volume", pOpts.name)

	// Use the provisioner's volume manager so that its mutex serializes local
	// operations across concurrent requests
	vm := &p.localVolumeManager

	req := &VolumeRequest{
		Name: pOpts.name,
		Path: pOpts.path,
	}

	return vm.DeleteVolume(ctx, req)
}
