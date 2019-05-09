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
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	pvController "github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/controller"
	mconfig "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	mPV "github.com/openebs/maya/pkg/kubernetes/persistentvolume/v1alpha1"
	"k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	//KeyNode represents the key values used for specifying the Node Affinity
	// based on the hostname
	KeyNode = "kubernetes.io/hostname"
)

// NewProvisioner will create a new Provisioner object and initialize
//  it with global information used across PV create and delete operations.
func NewProvisioner(stopCh chan struct{}, kubeClient *clientset.Clientset) (*Provisioner, error) {

	namespace := getOpenEBSNamespace() //menv.Get(menv.OpenEBSNamespace)
	if len(strings.TrimSpace(namespace)) == 0 {
		return nil, fmt.Errorf("Cannot start Provisioner: failed to get namespace")
	}

	p := &Provisioner{
		stopCh: stopCh,

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
//  supported by the host path provisioner. Return false.
func (p *Provisioner) SupportsBlock() bool {
	return false
}

// Provision is invoked by the PVC controller which expect the PV
//  to be provisioned and a valid PV spec returned.
func (p *Provisioner) Provision(opts pvController.VolumeOptions) (*v1.PersistentVolume, error) {
	pvc := opts.PVC
	if pvc.Spec.Selector != nil {
		return nil, fmt.Errorf("claim.Spec.Selector is not supported")
	}
	for _, accessMode := range pvc.Spec.AccessModes {
		if accessMode != v1.ReadWriteOnce {
			return nil, fmt.Errorf("Only support ReadWriteOnce access mode")
		}
	}
	node := opts.SelectedNode
	if opts.SelectedNode == nil {
		return nil, fmt.Errorf("configuration error, no node was specified")
	}

	name := opts.PVName

	// Create a new Config instance for the PV by merging the
	// default configuration with configuration provided
	// via PVC and the associated StorageClass
	pvCASConfig, err := p.getVolumeConfig(name, pvc)
	if err != nil {
		return nil, err
	}

	//TODO: Determine if hostpath or device based Local PV should be created
	stgType := pvCASConfig.GetStorageType()
	if stgType != "hostpath" {
		return nil, fmt.Errorf("PV with StorageType %v is not supported", stgType)
	}

	path, err := pvCASConfig.GetPath()
	if err != nil {
		return nil, err
	}

	glog.Infof("Creating volume %v at %v:%v", name, node.Name, path)

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

	//labels := make(map[string]string)
	//labels[string(v1alpha1.CASTypeKey)] = casVolume.Spec.CasType
	//labels[string(v1alpha1.StorageClassKey)] = *className

	//TODO Change the following to a builder pattern
	pvObj, err := mPV.NewBuilder().
		WithName(name).
		WithReclaimPolicy(opts.PersistentVolumeReclaimPolicy).
		WithAccessModes(pvc.Spec.AccessModes).
		WithVolumeMode(fs).
		WithCapacityQty(pvc.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]).
		WithHostDirectory(path).
		WithNodeAffinity(node.Name).
		Build()

	if err != nil {
		return nil, err
	}

	return pvObj, nil

}

// Delete is invoked by the PVC controller to perform clean-up
//  activities before deleteing the PV object. If reclaim policy is
//  set to not-retain, then this function will create a helper pod
//  to delete the host path from the node.
func (p *Provisioner) Delete(pv *v1.PersistentVolume) (err error) {
	defer func() {
		err = errors.Wrapf(err, "failed to delete volume %v", pv.Name)
	}()

	//Determine the path and node of the Local PV.
	path, node, err := p.getPathAndNodeForPV(pv)
	if err != nil {
		return err
	}

	//Initiate clean up only when reclaim policy is not retain.
	if pv.Spec.PersistentVolumeReclaimPolicy != v1.PersistentVolumeReclaimRetain {
		glog.Infof("Deleting volume %v at %v:%v", pv.Name, node, path)
		cleanupCmdsForPath := []string{"rm", "-rf"}
		podOpts := &HelperPodOptions{
			cmdsForPath: cleanupCmdsForPath,
			name:        pv.Name,
			path:        path,
			nodeName:    node,
		}

		//if err := p.createCleanupPod(cleanupCmdsForPath, pv.Name, path, node); err != nil {
		if err := p.createCleanupPod(podOpts); err != nil {
			glog.Infof("clean up volume %v failed: %v", pv.Name, err)
			return err
		}
		return nil
	}
	glog.Infof("Retained volume %v", pv.Name)
	return nil
}
