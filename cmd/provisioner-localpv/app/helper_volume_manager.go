package app

import (
	"context"

	"k8s.io/klog/v2"
)

// createVolumeLocally performs volume creation directly on the local node
func (p *Provisioner) createVolumeLocally(ctx context.Context, pOpts *HelperPodOptions, enableQuota bool) error {
	klog.Infof("Creating volume %s locally", pOpts.name)

	// Create a temporary volume manager to perform local operations
	vm := NewLocalVolumeManager()

	req := &VolumeRequest{
		Name:           pOpts.name,
		Path:           pOpts.path,
		FsMode:         "0777", // Default file permissions
		SoftLimitGrace: pOpts.softLimitGrace,
		HardLimitGrace: pOpts.hardLimitGrace,
		PVCStorage:     pOpts.pvcStorage,
	}

	return vm.CreateVolume(ctx, req, enableQuota)
}

// deleteVolumeLocally deletes volume directly on the local node
func (p *Provisioner) deleteVolumeLocally(ctx context.Context, pOpts *HelperPodOptions) error {
	klog.Infof("Deleting volume %s locally", pOpts.name)

	// Create a temporary volume manager to perform local operations
	vm := NewLocalVolumeManager()

	req := &VolumeRequest{
		Name: pOpts.name,
		Path: pOpts.path,
	}

	return vm.DeleteVolume(ctx, req)
}
