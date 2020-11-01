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
	"fmt"
	"time"

	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/pkg/version"
	"github.com/pkg/errors"
	"k8s.io/klog"

	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstor/pool/v1alpha3"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// SPCFinalizer represents the finalizer on spc
	SPCFinalizer = "storagepoolclaim.openebs.io/finalizer"
)

// SupportedDiskTypes is a map containing the valid disk type
var SupportedDiskTypes = map[apis.CasPoolValString]bool{
	apis.TypeSparseCPV: true,
	apis.TypeDiskCPV:   true,
}

// DefaultDiskCount is a map containing the default block device count of various raid types.
var DefaultDiskCount = map[string]int{
	string(apis.PoolTypeMirroredCPV): int(apis.MirroredBlockDeviceCountCPV),
	string(apis.PoolTypeStripedCPV):  int(apis.StripedBlockDeviceCountCPV),
	string(apis.PoolTypeRaidzCPV):    int(apis.RaidzBlockDeviceCountCPV),
	string(apis.PoolTypeRaidz2CPV):   int(apis.Raidz2BlockDeviceCountCPV),
}

// SPC encapsulates StoragePoolClaim api object.
type SPC struct {
	// actual spc object
	Object *apis.StoragePoolClaim
}

// SPCList holds the list of StoragePoolClaim api
type SPCList struct {
	// list of storagepoolclaims
	ObjectList *apis.StoragePoolClaimList
}

// Builder is the builder object for SPC.
type Builder struct {
	Spc *SPC
}

// ListBuilder is the builder object for SPCList.
type ListBuilder struct {
	SpcList *SPCList
}

// Predicate defines an abstraction to determine conditional checks against the provided spc instance.
type Predicate func(*SPC) bool

type predicateList []Predicate

// all returns true if all the predicates succeed against the provided csp instance.
func (l predicateList) all(c *SPC) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// IsOverProvisioningEnabled returns OverProvisioning truth value.
// OverProvisioning field is deprecated and not honoured
func IsOverProvisioningEnabled() Predicate {
	return func(spc *SPC) bool {
		return spc.Object.Spec.PoolSpec.OverProvisioning
	}
}

// IsThickProvisioningEnabled returns ThickProvisioning truth value.
func IsThickProvisioningEnabled() Predicate {
	return func(spc *SPC) bool {
		return spc.Object.Spec.PoolSpec.ThickProvisioning
	}
}

// HasAnnotation returns true if provided annotation key and value are present in the provided spc instance.
func HasAnnotation(key, value string) Predicate {
	return func(c *SPC) bool {
		val, ok := c.Object.GetAnnotations()[key]
		if ok {
			return val == value
		}
		return false
	}
}

// HasFinalizer is a predicate to filter out based on provided
// finalizer being present on the object.
func HasFinalizer(finalizer string) Predicate {
	return func(spc *SPC) bool {
		return spc.HasFinalizer(finalizer)
	}
}

// HasFinalizer returns true if the provided finalizer is present on the object.
func (spc *SPC) HasFinalizer(finalizer string) bool {
	finalizersList := spc.Object.GetFinalizers()
	return util.ContainsString(finalizersList, finalizer)
}

// IsAutoProvisioning returns true if the spc is auto provisioning type
func (spc *SPC) IsAutoProvisioning() bool {
	return spc.Object.Spec.BlockDevices.BlockDeviceList == nil
}

// RemoveFinalizer removes the given finalizer from the object.
func (spc *SPC) RemoveFinalizer(finalizer string) error {
	if len(spc.Object.Finalizers) == 0 {
		klog.V(2).Infof("no finalizer present on SPC %s", spc.Object.Name)
		return nil
	}

	if !spc.HasFinalizer(finalizer) {
		klog.V(2).Infof("finalizer %s is already removed on SPC %s", finalizer, spc.Object.Name)
		return nil
	}

	spc.Object.Finalizers = util.RemoveString(spc.Object.Finalizers, finalizer)

	_, err := NewKubeClient().
		Update(spc.Object)
	if err != nil {
		return errors.Wrapf(err, "failed to remove finalizers from SPC %s", spc.Object.Name)
	}
	klog.Infof("Finalizer %s removed successfully from SPC %s", finalizer, spc.Object.Name)
	return nil
}

// AddFinalizer adds the given finalizer to the object.
func (spc *SPC) AddFinalizer(finalizer string) (*apis.StoragePoolClaim, error) {
	if spc.HasFinalizer(finalizer) {
		klog.V(2).Infof("finalizer %s is already present on SPC %s", finalizer, spc.Object.Name)
		return spc.Object, nil
	}

	spc.Object.Finalizers = append(spc.Object.Finalizers, finalizer)

	spcAPIObj, err := NewKubeClient().
		Update(spc.Object)

	if err != nil {
		return nil, errors.Wrap(err, "failed to update SPC while adding finalizers")
	}
	klog.Infof("Finalizer %s added on storagepoolclaim %s", finalizer, spc.Object.Name)
	return spcAPIObj, nil
}

func (spc *SPC) addSPCFinalizerOnAssociatedBDCs() error {
	namespace := env.Get(env.OpenEBSNamespace)

	bdcList, err := bdc.NewKubeClient().WithNamespace(namespace).List(
		metav1.ListOptions{
			LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spc.Object.Name,
		})

	if err != nil {
		return errors.Wrapf(err, "failed to get bdclist for %s to add SPC finalizer", spc.Object.Name)
	}

	for _, bdcObj := range bdcList.Items {
		bdcObj := bdcObj
		if !bdcObj.DeletionTimestamp.IsZero() {
			continue
		}
		_, err := bdc.BuilderForAPIObject(&bdcObj).BDC.AddFinalizer(SPCFinalizer)
		if err != nil {
			return errors.Wrap(err, "failed to add SPC finalizer on BDC")
		}
	}
	return nil
}

// Filter will filter the csp instances if all the predicates succeed against that spc.
func (l *SPCList) Filter(p ...Predicate) *SPCList {
	var plist predicateList
	plist = append(plist, p...)
	if len(plist) == 0 {
		return l
	}

	filtered := NewListBuilder().List()
	for _, spcAPI := range l.ObjectList.Items {
		spcAPI := spcAPI // pin it
		SPC := BuilderForAPIObject(&spcAPI).Spc
		if plist.all(SPC) {
			filtered.ObjectList.Items = append(filtered.ObjectList.Items, *SPC.Object)
		}
	}
	return filtered
}

// NewBuilder returns an empty instance of the Builder object.
func NewBuilder() *Builder {
	return &Builder{
		Spc: &SPC{&apis.StoragePoolClaim{}},
	}
}

// BuilderForObject returns an instance of the Builder object based on spc object
func BuilderForObject(SPC *SPC) *Builder {
	return &Builder{
		Spc: SPC,
	}
}

// BuilderForAPIObject returns an instance of the Builder object based on spc api object.
func BuilderForAPIObject(spc *apis.StoragePoolClaim) *Builder {
	return &Builder{
		Spc: &SPC{spc},
	}
}

// WithName sets the Name field of spc with provided argument value.
func (sb *Builder) WithName(name string) *Builder {
	sb.Spc.Object.Name = name
	sb.Spc.Object.Spec.Name = name
	return sb
}

// WithGenerateName appends a random string after the name
func (sb *Builder) WithGenerateName(name string) *Builder {
	name = name + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
	return sb.WithName(name)
}

// WithDiskType sets the Type field of spc with provided argument value.
func (sb *Builder) WithDiskType(diskType string) *Builder {
	sb.Spc.Object.Spec.Type = diskType
	return sb
}

// WithPoolType sets the poolType field of spc with provided argument value.
func (sb *Builder) WithPoolType(poolType string) *Builder {
	sb.Spc.Object.Spec.PoolSpec.PoolType = poolType
	return sb
}

// WithOverProvisioning sets the OverProvisioning field of spc with provided argument value.
func (sb *Builder) WithOverProvisioning(val bool) *Builder {
	sb.Spc.Object.Spec.PoolSpec.OverProvisioning = val
	return sb
}

// WithThickProvisioning sets the ThickProvisioning field of spc with provided argument value.
func (sb *Builder) WithThickProvisioning(val bool) *Builder {
	sb.Spc.Object.Spec.PoolSpec.ThickProvisioning = val
	return sb
}

// WithPool sets the poolType field of spc with provided argument value.
func (sb *Builder) WithPool(poolType string) *Builder {
	sb.Spc.Object.Spec.PoolSpec.PoolType = poolType
	return sb
}

// WithMaxPool sets the maxpool field of spc with provided argument value.
func (sb *Builder) WithMaxPool(val int) *Builder {
	maxPool := newInt(val)
	sb.Spc.Object.Spec.MaxPools = maxPool
	return sb
}

// newInt returns a pointer to the int value.
func newInt(val int) *int {
	newVal := val
	return &newVal
}

// Build returns the SPC object built by this builder.
func (sb *Builder) Build() *SPC {
	return sb.Spc
}

// NewListBuilder returns a new instance of ListBuilder object.
func NewListBuilder() *ListBuilder {
	return &ListBuilder{SpcList: &SPCList{ObjectList: &apis.StoragePoolClaimList{}}}
}

// WithUIDs builds a list of StoragePoolClaims based on the provided pool UIDs
func (b *ListBuilder) WithUIDs(poolUIDs ...string) *ListBuilder {
	for _, uid := range poolUIDs {
		obj := &SPC{&apis.StoragePoolClaim{}}
		obj.Object.SetUID(types.UID(uid))
		b.SpcList.ObjectList.Items = append(b.SpcList.ObjectList.Items, *obj.Object)
	}
	return b
}

// WithList builds the list based on the provided *SPCList instances.
func (b *ListBuilder) WithList(pools *SPCList) *ListBuilder {
	if pools == nil {
		return b
	}
	b.SpcList.ObjectList.Items = append(b.SpcList.ObjectList.Items, pools.ObjectList.Items...)
	return b
}

// WithAPIList builds the list based on the provided *apis.CStorPoolList.
func (b *ListBuilder) WithAPIList(pools *apis.StoragePoolClaimList) *ListBuilder {
	if pools == nil {
		return b
	}
	for _, pool := range pools.Items {
		pool := pool //pin it
		b.SpcList.ObjectList.Items = append(b.SpcList.ObjectList.Items, pool)
	}
	return b
}

// List returns the list of csp instances that were built by this builder.
func (b *ListBuilder) List() *SPCList {
	return b.SpcList
}

// Len returns the length og SPCList.
func (l *SPCList) Len() int {
	return len(l.ObjectList.Items)
}

// IsEmpty returns false if the SPCList is empty.
func (l *SPCList) IsEmpty() bool {
	return len(l.ObjectList.Items) == 0
}

// GetPoolUIDs retuns the UIDs of the pools available in the list.
func (l *SPCList) GetPoolUIDs() []string {
	uids := []string{}
	for _, pool := range l.ObjectList.Items {
		uids = append(uids, string(pool.GetUID()))
	}
	return uids
}

// EstimateSPCVersion returns the csp version if any csp is present for the spc or
// returns the maya version as the new csp created will be of maya version
func (spc *SPC) EstimateSPCVersion() (string, error) {

	cspList, err := csp.KubeClient().List(
		metav1.ListOptions{
			LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spc.Object.Name,
		})
	if err != nil {
		return "", errors.Wrapf(
			err,
			"failed to get the csp list related to spc : %s",
			spc.Object.Name,
		)
	}
	if len(cspList.Items) == 0 {
		return version.Current(), nil
	}
	return cspList.Items[0].Labels[string(apis.OpenEBSVersionKey)], nil

}
