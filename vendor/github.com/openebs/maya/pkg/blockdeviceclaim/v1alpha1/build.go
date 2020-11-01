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
	ndm "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// StoragePoolKind holds the value of StoragePoolClaim
	StoragePoolKind = "StoragePoolClaim"

	// StoragePoolKindCSPC holds the value of CStorPoolCluster
	StoragePoolKindCSPC = "CStorPoolCluster"

	// APIVersion holds the value of OpenEBS version
	APIVersion = "openebs.io/v1alpha1"

	// bdTagKey defines the label selector key
	// used for grouping block devices using a tag.
	bdTagKey = "openebs.io/block-device-tag"
)

// Builder is the builder object for BlockDeviceClaim
type Builder struct {
	BDC  *BlockDeviceClaim
	errs []error
}

// NewBuilder returns an empty instance of the Builder object
func NewBuilder() *Builder {
	return &Builder{
		BDC:  &BlockDeviceClaim{&ndm.BlockDeviceClaim{}, ""},
		errs: []error{},
	}
}

// BuilderForObject returns an instance of the Builder object based on block
// device object
func BuilderForObject(BlockDeviceClaim *BlockDeviceClaim) *Builder {
	return &Builder{
		BDC:  BlockDeviceClaim,
		errs: []error{},
	}
}

// BuilderForAPIObject returns an instance of the Builder object based on block
// device claim api object.
func BuilderForAPIObject(bdc *ndm.BlockDeviceClaim) *Builder {
	return &Builder{
		BDC:  &BlockDeviceClaim{bdc, ""},
		errs: []error{},
	}
}

// WithConfigPath sets the path for k8s config
func (b *Builder) WithConfigPath(configpath string) *Builder {
	b.BDC.configPath = configpath
	return b
}

// Build returns the BlockDeviceClaim instance
func (b *Builder) Build() (*BlockDeviceClaim, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.BDC, nil
}

// WithName sets the Name field of BDC with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: missing BDC name"),
		)
		return b
	}
	b.BDC.Object.Name = name
	return b
}

// WithNamespace sets the Namespace field of BDC provided arguments
func (b *Builder) WithNamespace(namespace string) *Builder {
	if len(namespace) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: missing BDC namespace"),
		)
		return b
	}
	b.BDC.Object.Namespace = namespace
	return b
}

// WithAnnotationsNew sets the Annotations field of BDC with provided arguments
func (b *Builder) WithAnnotationsNew(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: missing annotations"),
		)
		return b
	}
	b.BDC.Object.Annotations = make(map[string]string)
	for key, value := range annotations {
		b.BDC.Object.Annotations[key] = value
	}
	return b
}

// WithAnnotations appends or overwrites existing Annotations
// values of BDC with provided arguments
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: missing annotations"),
		)
		return b
	}
	if b.BDC.Object.Annotations == nil {
		return b.WithAnnotationsNew(annotations)
	}
	for key, value := range annotations {
		b.BDC.Object.Annotations[key] = value
	}
	return b
}

// WithLabelsNew sets the Labels field of BDC with provided arguments
func (b *Builder) WithLabelsNew(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: missing labels"),
		)
		return b
	}
	b.BDC.Object.Labels = make(map[string]string)
	for key, value := range labels {
		b.BDC.Object.Labels[key] = value
	}
	return b
}

// WithLabels appends or overwrites existing Labels
// values of BDC with provided arguments
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: missing labels"),
		)
		return b
	}
	if b.BDC.Object.Labels == nil {
		return b.WithLabelsNew(labels)
	}
	for key, value := range labels {
		b.BDC.Object.Labels[key] = value
	}
	return b
}

// WithBlockDeviceName sets the BlockDeviceName field of BDC provided arguments
func (b *Builder) WithBlockDeviceName(bdName string) *Builder {
	if len(bdName) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: missing BlockDevice name"),
		)
		return b
	}
	b.BDC.Object.Spec.BlockDeviceName = bdName
	return b
}

// WithDeviceType sets the DeviceType field of BDC provided arguments
func (b *Builder) WithDeviceType(dType string) *Builder {
	if len(dType) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: missing device type"),
		)
		return b
	}
	b.BDC.Object.Spec.DeviceType = dType
	return b
}

// WithHostName sets the hostName field of BDC provided arguments
func (b *Builder) WithHostName(hName string) *Builder {
	if len(hName) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: missing host name"),
		)
		return b
	}
	b.BDC.Object.Spec.BlockDeviceNodeAttributes.HostName = hName
	return b
}

// WithNodeName sets the node name field of BDC provided arguments
func (b *Builder) WithNodeName(nName string) *Builder {
	if len(nName) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: missing node name"),
		)
		return b
	}
	b.BDC.Object.Spec.BlockDeviceNodeAttributes.NodeName = nName
	return b
}

// WithCapacity sets the Capacity field in BDC with provided arguments
func (b *Builder) WithCapacity(capacity string) *Builder {
	resCapacity, err := resource.ParseQuantity(capacity)
	if err != nil {
		b.errs = append(
			b.errs,
			errors.Wrapf(
				err, "failed to build BDC object: failed to parse capacity {%s}",
				capacity,
			),
		)
		return b
	}
	resourceList := corev1.ResourceList{
		corev1.ResourceName(ndm.ResourceStorage): resCapacity,
	}
	b.BDC.Object.Spec.Resources.Requests = resourceList
	return b
}

// WithOwnerReference sets the OwnerReference field in BDC with required
//fields
func (b *Builder) WithOwnerReference(spc *apis.StoragePoolClaim) *Builder {
	if spc == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: spc object is nil"),
		)
		return b
	}
	trueVal := true
	reference := metav1.OwnerReference{
		APIVersion:         APIVersion,
		Kind:               StoragePoolKind,
		UID:                spc.ObjectMeta.UID,
		Name:               spc.ObjectMeta.Name,
		BlockOwnerDeletion: &trueVal,
		Controller:         &trueVal,
	}
	b.BDC.Object.OwnerReferences = append(b.BDC.Object.OwnerReferences, reference)
	return b
}

// WithCSPCOwnerReference sets the OwnerReference field in BDC with required
//fields
func (b *Builder) WithCSPCOwnerReference(cspc *apis.CStorPoolCluster) *Builder {
	if cspc == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: cspc object is nil"),
		)
		return b
	}
	trueVal := true
	reference := metav1.OwnerReference{
		APIVersion:         APIVersion,
		Kind:               StoragePoolKindCSPC,
		UID:                cspc.ObjectMeta.UID,
		Name:               cspc.ObjectMeta.Name,
		BlockOwnerDeletion: &trueVal,
		Controller:         &trueVal,
	}
	b.BDC.Object.OwnerReferences = append(b.BDC.Object.OwnerReferences, reference)
	return b
}

// WithFinalizer sets the finalizer field in the BDC
func (b *Builder) WithFinalizer(finalizers ...string) *Builder {
	if len(finalizers) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: missing finalizer"),
		)
		return b
	}
	b.BDC.Object.Finalizers = append(b.BDC.Object.Finalizers, finalizers...)
	return b
}

// WithBlockVolumeMode sets the volumeMode as volumeModeBlock,
// if persistentVolumeMode is set to "Block"
func (b *Builder) WithBlockVolumeMode(mode corev1.PersistentVolumeMode) *Builder {
	if len(mode) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: missing PersistentVolumeMode"),
		)
	}
	if mode == corev1.PersistentVolumeBlock {
		b.BDC.Object.Spec.Details.BlockVolumeMode = ndm.VolumeModeBlock
	}

	return b
}

// WithBlockDeviceTag appends (or creates) the BDC Label Selector
// by setting the provided value to the fixed key
// openebs.io/block-device-tag
// This will enable the NDM to pick only devices that
// match the node (hostname) and block device tag value.
func (b *Builder) WithBlockDeviceTag(bdTagValue string) *Builder {
	if len(bdTagValue) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build BDC object: missing block device tag value"),
		)
		return b
	}

	if b.BDC.Object.Spec.Selector == nil {
		b.BDC.Object.Spec.Selector = &metav1.LabelSelector{}
	}
	if b.BDC.Object.Spec.Selector.MatchLabels == nil {
		b.BDC.Object.Spec.Selector.MatchLabels = map[string]string{}
	}

	b.BDC.Object.Spec.Selector.MatchLabels[bdTagKey] = bdTagValue
	return b
}
