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

package v1alpha1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	ndm "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

// BlockDeviceClaim encapsulates BlockDeviceClaim api object.
type BlockDeviceClaim struct {
	// actual block device claim object
	Object *ndm.BlockDeviceClaim

	// kubeconfig path
	configPath string
}

// BlockDeviceClaimList encapsulates BlockDeviceClaimList api object
type BlockDeviceClaimList struct {
	// list of blockdeviceclaims
	ObjectList *ndm.BlockDeviceClaimList
}

// Predicate defines an abstraction to determine conditional checks against the
// provided block device claim instance
type Predicate func(*BlockDeviceClaim) bool

// PredicateList holds the list of Predicates
type PredicateList []Predicate

// all returns true if all the predicates succeed against the provided block
// device instance.
func (l PredicateList) all(c *BlockDeviceClaim) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// HasAnnotation is predicate to filter out based on
// annotation in BDC instances
func HasAnnotation(key, value string) Predicate {
	return func(bdc *BlockDeviceClaim) bool {
		return bdc.HasAnnotation(key, value)
	}
}

// HasAnnotation return true if provided annotation
// key and value are present in the the provided BDCList
// instance
func (bdc *BlockDeviceClaim) HasAnnotation(key, value string) bool {
	val, ok := bdc.Object.GetAnnotations()[key]
	if ok {
		return val == value
	}
	return false
}

// HasAnnotationKey is predicate to filter out based on
// annotation key in BDC instances
func HasAnnotationKey(key string) Predicate {
	return func(bdc *BlockDeviceClaim) bool {
		return bdc.HasAnnotationKey(key)
	}
}

// HasAnnotationKey return true if provided annotation
// key is present in the the provided BDC instance.
func (bdc *BlockDeviceClaim) HasAnnotationKey(key string) bool {
	_, ok := bdc.Object.GetAnnotations()[key]
	return ok
}

// HasBD is predicate to filter out based on BD.
func HasBD(bdName string) Predicate {
	return func(bdc *BlockDeviceClaim) bool {
		return bdc.HasBD(bdName)
	}
}

// HasBD return true if provided BD belongs to the BDC instance.
func (bdc *BlockDeviceClaim) HasBD(bdName string) bool {
	return bdc.Object.Spec.BlockDeviceName == bdName
}

// HasLabel is predicate to filter out labeled
// BDC instances.
func HasLabel(key, value string) Predicate {
	return func(bdc *BlockDeviceClaim) bool {
		return bdc.HasLabel(key, value)
	}
}

// HasLabel returns true if provided label
// key and value are present in the provided BDC(BlockDeviceClaim)
// instance
func (bdc *BlockDeviceClaim) HasLabel(key, value string) bool {
	val, ok := bdc.Object.GetLabels()[key]
	if ok {
		return val == value
	}
	return false
}

// HasFinalizer is a predicate to filter out based on provided
// finalizer being present on the object.
func HasFinalizer(finalizer string) Predicate {
	return func(bdc *BlockDeviceClaim) bool {
		return bdc.HasFinalizer(finalizer)
	}
}

// HasFinalizer returns true if the provided finalizer is present on the object.
func (bdc *BlockDeviceClaim) HasFinalizer(finalizer string) bool {
	finalizersList := bdc.Object.GetFinalizers()
	return util.ContainsString(finalizersList, finalizer)
}

// AddFinalizer adds the given finalizer to the object.
func (bdc *BlockDeviceClaim) AddFinalizer(finalizer string) (*ndm.BlockDeviceClaim, error) {
	if bdc.HasFinalizer(finalizer) {
		klog.V(2).Infof("finalizer %s is already present on BDC %s", finalizer, bdc.Object.Name)
		return bdc.Object, nil
	}

	bdc.Object.Finalizers = append(bdc.Object.Finalizers, finalizer)

	bdcAPIObj, err := NewKubeClient(WithKubeConfigPath(bdc.configPath)).
		WithNamespace(bdc.Object.Namespace).
		Update(bdc.Object)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to update bdc %s while adding finalizer %s",
			bdc.Object.Name, finalizer)
	}

	klog.Infof("Finalizer %s added on blockdeviceclaim %s", finalizer, bdc.Object.Name)
	return bdcAPIObj, nil
}

// RemoveFinalizer removes the given finalizer from the object.
func (bdc *BlockDeviceClaim) RemoveFinalizer(
	finalizer string) (*ndm.BlockDeviceClaim, error) {
	if len(bdc.Object.Finalizers) == 0 {
		klog.V(2).Infof("no finalizer present on BDC %s", bdc.Object.Name)
		return bdc.Object, nil
	}

	if !bdc.HasFinalizer(finalizer) {
		klog.V(2).Infof("finalizer %s is already removed on BDC %s", finalizer, bdc.Object.Name)
		return bdc.Object, nil
	}

	bdc.Object.Finalizers = util.RemoveString(bdc.Object.Finalizers, finalizer)

	newBDC, err := NewKubeClient(WithKubeConfigPath(bdc.configPath)).
		WithNamespace(bdc.Object.Namespace).
		Update(bdc.Object)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update object while removing finalizer")
	}
	klog.Infof("Finalizer %s removed successfully from BDC %s", finalizer, bdc.Object.Name)
	return newBDC, nil
}

// IsStatus is predicate to filter out BDC instances based on argument provided
func IsStatus(status string) Predicate {
	return func(bdc *BlockDeviceClaim) bool {
		return bdc.IsStatus(status)
	}
}

// IsStatus returns true if the status on
// block device claim matches with provided status.
func (bdc *BlockDeviceClaim) IsStatus(status string) bool {
	return string(bdc.Object.Status.Phase) == status
}

// GetSpecHostName return hostName from spec of blockdeviceclaim
func (bdc *BlockDeviceClaim) GetSpecHostName() string {
	return bdc.Object.Spec.HostName
}

// GetNodeAtributesHostName return hostName from blockdeviceclaim attribute hostName
func (bdc *BlockDeviceClaim) GetNodeAtributesHostName() string {
	return bdc.Object.Spec.BlockDeviceNodeAttributes.HostName
}

// GetHostName return hostName from blockdeviceclaim
func (bdc *BlockDeviceClaim) GetHostName() string {
	hostName := bdc.GetNodeAtributesHostName()
	if hostName == "" {
		return bdc.GetSpecHostName()
	}
	return hostName
}

// Len returns the length og BlockDeviceClaimList.
func (bdcl *BlockDeviceClaimList) Len() int {
	return len(bdcl.ObjectList.Items)
}

// GetBlockDeviceNamesByNode returns map of node name and corresponding blockdevices to that
// node from blockdeviceclaim list
func (bdcl *BlockDeviceClaimList) GetBlockDeviceNamesByNode() map[string][]string {
	newNodeBDList := make(map[string][]string)
	if bdcl == nil {
		return newNodeBDList
	}
	for _, bdc := range bdcl.ObjectList.Items {
		bdc := bdc
		bdcObj := BlockDeviceClaim{Object: &bdc}
		hostName := bdcObj.GetHostName()
		newNodeBDList[hostName] = append(
			newNodeBDList[hostName],
			bdcObj.Object.Spec.BlockDeviceName,
		)
	}
	return newNodeBDList
}

// GetBlockDeviceClaimFromBDName return block device claim if claim exists for
// provided blockdevice name in claim list else return error
func (bdcl *BlockDeviceClaimList) GetBlockDeviceClaimFromBDName(
	bdName string) (*apis.BlockDeviceClaim, error) {
	for _, bdc := range bdcl.ObjectList.Items {
		// pin it
		bdc := bdc
		if bdc.Spec.BlockDeviceName == bdName {
			return &bdc, nil
		}
	}
	return nil, errors.Errorf("claim doesn't exist for blockdevice %s", bdName)
}
