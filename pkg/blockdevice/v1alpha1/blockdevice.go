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
	"context"
	"strings"

	apis "github.com/openebs/api/v3/pkg/apis/openebs.io/v1alpha1"
	bdc_v1alpha1 "github.com/openebs/dynamic-localpv-provisioner/pkg/blockdeviceclaim/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	apisclientset "github.com/openebs/api/v3/pkg/client/clientset/versioned"
)

// TODO: Update the file with latest pattern
const (
	// StorageNodePredicateKey is the key for StorageNodePredicate function.
	FilterInactive    = "filterInactive"
	FilterNonInactive = "filterNonInactive"
	//FilterNonPartitions    = "filterNonPartitions"
	FilterNonFSType         = "filterNonFSType"
	FilterSparseDevices     = "filterSparseDevices"
	FilterNonSparseDevices  = "filterNonSparseDevices"
	InActiveStatus          = "Inactive"
	FilterNonRelesedDevices = "filterNonRelesedDevices"
	FilterNotAllowedBDTag   = "filterNotAllowedBDTag"

	// BlockDeviceTagLabelKey is the key to fetch tag of a block
	// device.
	// For more info : https://github.com/openebs/node-disk-manager/pull/400
	BlockDeviceTagLabelKey = "openebs.io/block-device-tag"
)

var bdFilterOptions FilterOptions

// KubernetesClient is the kubernetes client which will implement block device actions/behaviours
type KubernetesClient struct {
	// Kubeclientset is a standard kubernetes clientset
	Kubeclientset kubernetes.Interface

	// Clientset is a apis custom resource package generated for apis custom API group
	Clientset apisclientset.Interface

	//Namespace is namespace where blockdevice is available
	Namespace string
}

type errs []error

// SpcObjectClient is the kubernetes client perform block devie operations in
// case of manual provisioning
type SpcObjectClient struct {
	*KubernetesClient
	Spc *apis.StoragePoolClaim
}

// BlockDevice is a wrapper over BlockDevice api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers
type BlockDevice struct {
	*apis.BlockDevice
	errs
}

// BlockDeviceList is a wrapper over BlockDeviceList api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers
type BlockDeviceList struct {
	*apis.BlockDeviceList
	errs
}

// BuildOptionFunc is the typed function to build BlockDevice object.
type BuildOptionFunc func(*BlockDevice)

// predicate is the typed predicate function to validate BlockDevice object.
type predicate func(*BlockDevice) (message string, ok bool)

// filterOptionFunc is the typed function to filter BlockDevice objects.
type filterOptionFunc func(original *BlockDeviceList) *BlockDeviceList

// BlockDeviceInterface abstracts operations on BlockDevice entity.
// Different orchestrators may need to implement this interface.
type BlockDeviceInterface interface {
	Get(name string, opts metav1.GetOptions) (*BlockDevice, error)
	List(opts metav1.ListOptions) (*BlockDeviceList, error)
	Create(*apis.BlockDevice) (*BlockDevice, error)
}

// checkPredicatesFuncs is an array of check predicate functions.
var checkPredicatesFuncs = [...]predicate{
	checkName,
}

// filterPredicatesFuncMap is an array of filter predicate functions
// filter predicates should be tunable by client.
var filterOptionFuncMap = map[string]filterOptionFunc{
	FilterInactive:    filterInactive,
	FilterNonInactive: filterNonInactive,
	//FilterNonPartitions:    filterNonPartitions,
	FilterNonFSType:         filterNonFSType,
	FilterSparseDevices:     filterSparseDevices,
	FilterNonSparseDevices:  filterNonSparseDevices,
	FilterNonRelesedDevices: filterNonRelesedDevices,
	FilterNotAllowedBDTag:   filterNotAllowedBDTag,
}

// predicateFailedError returns the predicate error which is provided to this function as an argument
func predicateFailedError(message string) error {
	return errors.Errorf("predicatefailed: %s", message)
}

// New is a constructor returns a new instance of block device
func New(opts ...BuildOptionFunc) *BlockDevice {
	r := &BlockDevice{BlockDevice: &apis.BlockDevice{}}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Build returns the final instance of BlockDevice
func (bd *BlockDevice) Build() (*apis.BlockDevice, []error) {
	bd.validate()
	if len(bd.errs) > 0 {
		return nil, bd.errs
	}
	return bd.BlockDevice, nil
}

// validate validates the block device object against the check predicates.
func (bd *BlockDevice) validate() {
	for _, c := range checkPredicatesFuncs {
		msg, ok := c(bd)
		if !ok {
			bd.errs = append(bd.errs, predicateFailedError(msg))
		}
	}
}

// WithName method fills the name field of BlockDevice object.
func (bd *BlockDevice) WithName(name string) *BlockDevice {
	WithName(name)(bd)
	return bd
}

// WithName function is used by WithName method as a util.
// Ideas is to give flexibility for building object by using dot operator as well as passing
// build predicated to the New constructor.
func WithName(name string) BuildOptionFunc {
	return func(bd *BlockDevice) {
		bd.BlockDevice.Name = name
	}
}

// WithState method fills the name field of BlockDevice object.
func (bd *BlockDevice) WithState(state string) *BlockDevice {
	WithState(state)(bd)
	return bd
}

// WithState function is used by WithState method as a util.
// Ideas is to give flexibility for building object by using dot operator as well as passing
// build predicated to the New constructor.
func WithState(state string) BuildOptionFunc {
	return func(bd *BlockDevice) {
		bd.BlockDevice.Status.State = apis.BlockDeviceState(state)
	}
}

// checkName validate the name field of BlockDevice object.
func checkName(db *BlockDevice) (string, bool) {
	if db.BlockDevice.Name == "" {
		//TODO: Think about having some good organization in putting error messages.
		return "blockDevice name field on the object may not be empty", false
	}
	return "", true
}

type FilterOptions struct {
	AllowedBDTags map[string]bool
}

// Filter adds filters on which the blockdevice has to be filtered
func (bdl *BlockDeviceList) Filter(filterOps *FilterOptions, predicateKeys ...string) *BlockDeviceList {
	// Initialize filtered block device list
	filteredBlockDeviceList := &BlockDeviceList{
		BlockDeviceList: &apis.BlockDeviceList{},
		errs:            nil,
	}
	errMsg, ok := bdl.Hasitems()
	if !ok {
		filteredBlockDeviceList.errs = append(filteredBlockDeviceList.errs, errors.New(errMsg))
		return filteredBlockDeviceList
	}
	filteredBlockDeviceList = bdl
	for _, key := range predicateKeys {
		if key == FilterNotAllowedBDTag {
			bdFilterOptions.AllowedBDTags = filterOps.AllowedBDTags
		}
		filteredBlockDeviceList = filterOptionFuncMap[key](filteredBlockDeviceList)
	}
	return filteredBlockDeviceList
}

// GetUsableBlockDevices returns custom blockdevice list(wrapper over api list)
// which contains list of blockdevice objects usable for provisioning
func (bdl *BlockDeviceList) GetUsableBlockDevices(spcName, namespace string) (*BlockDeviceList, error) {
	filteredBlockDeviceList := &BlockDeviceList{
		BlockDeviceList: &apis.BlockDeviceList{},
		errs:            nil,
	}
	bdcClient := bdc_v1alpha1.NewKubeClient().WithNamespace(namespace)
	for _, bdObj := range bdl.Items {
		if bdObj.Status.ClaimState == apis.BlockDeviceClaimed {
			bdcName := bdObj.Spec.ClaimRef.Name
			bdcObj, err := bdcClient.Get(context.TODO(), bdcName, metav1.GetOptions{})
			if err != nil {
				return nil, errors.Wrapf(err,
					"failed to get blockdeviceclaim %s details of blockdevice %s",
					bdcName,
					bdObj.Name,
				)
			}

			if bdcObj.Labels[string(apis.StoragePoolClaimCPK)] == spcName {
				filteredBlockDeviceList.Items = append(filteredBlockDeviceList.Items, bdObj)
			}
		} else if bdObj.Status.ClaimState == apis.BlockDeviceUnclaimed {
			filteredBlockDeviceList.Items = append(filteredBlockDeviceList.Items, bdObj)
		}
	}
	return filteredBlockDeviceList, nil
}

// filterInactive filter and give out all the inactive block device
func filterInactive(originalList *BlockDeviceList) *BlockDeviceList {
	filteredList := &BlockDeviceList{
		BlockDeviceList: &apis.BlockDeviceList{},
		errs:            nil,
	}
	for _, device := range originalList.Items {
		if device.Status.State == InActiveStatus {
			filteredList.Items = append(filteredList.Items, device)
		}
	}
	return filteredList
}

// filterNonInactive give out all the block device except inactive block devices
func filterNonInactive(orignialList *BlockDeviceList) *BlockDeviceList {
	filteredList := &BlockDeviceList{
		BlockDeviceList: &apis.BlockDeviceList{},
		errs:            nil,
	}
	for _, device := range orignialList.Items {
		if !(device.Status.State == InActiveStatus) {
			filteredList.Items = append(filteredList.Items, device)
		}
	}
	return filteredList
}

func filterNonPartitions(originalList *BlockDeviceList) *BlockDeviceList {
	filteredList := &BlockDeviceList{
		BlockDeviceList: &apis.BlockDeviceList{},
		errs:            nil,
	}
	for _, device := range originalList.Items {
		if strings.EqualFold(device.Spec.Partitioned, "No") {
			filteredList.Items = append(filteredList.Items, device)
		}
	}
	return filteredList
}

func filterNonFSType(originalList *BlockDeviceList) *BlockDeviceList {
	filteredList := &BlockDeviceList{
		BlockDeviceList: &apis.BlockDeviceList{},
		errs:            nil,
	}
	for _, device := range originalList.Items {
		if device.Spec.FileSystem.Type == "" {
			filteredList.Items = append(filteredList.Items, device)
		}
	}
	return filteredList
}

func filterSparseDevices(originalList *BlockDeviceList) *BlockDeviceList {
	filteredList := &BlockDeviceList{
		BlockDeviceList: &apis.BlockDeviceList{},
		errs:            nil,
	}
	for _, device := range originalList.Items {
		if device.Spec.Details.DeviceType == string(apis.TypeSparseCPV) {
			filteredList.Items = append(filteredList.Items, device)
		}
	}
	return filteredList
}

func filterNonSparseDevices(originalList *BlockDeviceList) *BlockDeviceList {
	filteredList := &BlockDeviceList{
		BlockDeviceList: &apis.BlockDeviceList{},
		errs:            nil,
	}
	for _, device := range originalList.Items {
		if !(device.Spec.Details.DeviceType == string(apis.TypeSparseCPV)) {
			filteredList.Items = append(filteredList.Items, device)
		}
	}
	return filteredList
}

func filterNonRelesedDevices(originalList *BlockDeviceList) *BlockDeviceList {
	filteredList := &BlockDeviceList{
		BlockDeviceList: &apis.BlockDeviceList{},
		errs:            nil,
	}
	for _, device := range originalList.Items {
		if !(device.Status.ClaimState == apis.BlockDeviceReleased) {
			filteredList.Items = append(filteredList.Items, device)
		}
	}
	return filteredList
}

func filterNotAllowedBDTag(originalList *BlockDeviceList) *BlockDeviceList {
	filteredList := &BlockDeviceList{
		BlockDeviceList: &apis.BlockDeviceList{},
		errs:            nil,
	}
	for _, device := range originalList.Items {
		value, ok := device.Labels[BlockDeviceTagLabelKey]
		bdTag := strings.TrimSpace(value)
		if ok {
			if bdTag == "" || !bdFilterOptions.AllowedBDTags[bdTag] {
				continue
			}
			filteredList.Items = append(filteredList.Items, device)
		} else {
			filteredList.Items = append(filteredList.Items, device)
		}
	}
	return filteredList
}

// Hasitems checks whether the BlockDeviceList contains BlockDevices
func (bdl *BlockDeviceList) Hasitems() (string, bool) {
	if bdl == nil || bdl.BlockDeviceList == nil || bdl.Items == nil {
		return "No item found in blockdevice list", false
	}
	return "", true
}

// IsClaimed returns true if block device is claimed
func (bd *BlockDevice) IsClaimed() bool {
	return bd.Status.ClaimState == apis.BlockDeviceClaimed
}

// GetDeviceID returns the device link of the block device.
// If device link is not found it returns device path.
// For a cstor pool creation -- this link or path is used.
// For convenience, we call it as device ID.
// Hence, device ID can either be a  device link or device path
// depending on what was available in block device cr.
func (bd *BlockDevice) GetDeviceID() string {
	deviceID := bd.GetLink()
	if deviceID != "" {
		return deviceID
	}
	return bd.GetPath()
}

// GetLink returns the link of the block device
// if present else return empty string
func (bd *BlockDevice) GetLink() string {
	if len(bd.Spec.DevLinks) != 0 &&
		len(bd.Spec.DevLinks[0].Links) != 0 {
		return bd.Spec.DevLinks[0].Links[0]
	}
	return ""
}

// GetPath returns path of the block device
func (bd *BlockDevice) GetPath() string {
	return bd.Spec.Path
}

// GetBlockDevice returns the block device object present in the block device list
func (bdl *BlockDeviceList) GetBlockDevice(bdcName string) *apis.BlockDevice {
	for _, bdcObj := range bdl.Items {
		bdcObj := bdcObj
		if bdcObj.Name == bdcName {
			return &bdcObj
		}
	}
	return nil
}
