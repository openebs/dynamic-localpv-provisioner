// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	errors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	//CStorNodeBase nodeBase for cstor volume
	CStorNodeBase string = "iqn.2016-09.com.openebs.cstor"
	// TargetPort is port for cstor volume
	TargetPort string = "3260"
)

// Builder is the builder object for CStorVolume
type Builder struct {
	cstorvolume *CStorVolume
	errs        []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{cstorvolume: &CStorVolume{object: &apis.CStorVolume{}}}
}

// BuilderFromAPI returns builder instance from core API
func BuilderFromAPI(cStorVolume *apis.CStorVolume) *Builder {
	return &Builder{cstorvolume: &CStorVolume{object: cStorVolume}}
}

// WithName sets the Name field of CStorVolume with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: missing name"),
		)
		return b
	}
	b.cstorvolume.object.Name = name
	return b
}

// WithGenerateName sets the GenerateName field of
// CStorVolume with provided value
func (b *Builder) WithGenerateName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: missing generateName"),
		)
		return b
	}

	b.cstorvolume.object.GenerateName = name
	return b
}

// WithNamespace sets the Namespace field of CStorVolume provided arguments
func (b *Builder) WithNamespace(namespace string) *Builder {
	if len(namespace) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: missing namespace"),
		)
		return b
	}
	b.cstorvolume.object.Namespace = namespace
	return b
}

// WithAnnotations merges existing annotations if any
// with the ones that are provided here
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: missing annotations"),
		)
		return b
	}

	if b.cstorvolume.object.Annotations == nil {
		return b.WithAnnotationsNew(annotations)
	}

	for key, value := range annotations {
		b.cstorvolume.object.Annotations[key] = value
	}
	return b
}

// WithAnnotationsNew resets existing annotations if any with
// ones that are provided here
func (b *Builder) WithAnnotationsNew(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: no new annotations"),
		)
		return b
	}

	// copy of original map
	newannotations := map[string]string{}
	for key, value := range annotations {
		newannotations[key] = value
	}

	// override
	b.cstorvolume.object.Annotations = newannotations
	return b
}

// WithOwnerRefernceNew sets ownerrefernce if any with
// ones that are provided here
func (b *Builder) WithOwnerRefernceNew(ownerRefernce []metav1.OwnerReference) *Builder {
	if len(ownerRefernce) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: no new ownerRefernce"),
		)
		return b
	}

	b.cstorvolume.object.OwnerReferences = ownerRefernce
	return b
}

// WithLabels merges existing labels if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: missing labels"),
		)
		return b
	}

	if b.cstorvolume.object.Labels == nil {
		return b.WithLabelsNew(labels)
	}

	for key, value := range labels {
		b.cstorvolume.object.Labels[key] = value
	}
	return b
}

// WithLabelsNew resets existing labels if any with
// ones that are provided here
func (b *Builder) WithLabelsNew(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: no new labels"),
		)
		return b
	}

	// copy of original map
	newlbls := map[string]string{}
	for key, value := range labels {
		newlbls[key] = value
	}

	// override
	b.cstorvolume.object.Labels = newlbls
	return b
}

// WithTargetIP sets the target IP address field of
// CStorVolume with provided arguments
func (b *Builder) WithTargetIP(targetip string) *Builder {
	if len(targetip) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: missing targetip"),
		)
		return b
	}
	b.cstorvolume.object.Spec.TargetIP = targetip
	return b
}

// WithCapacity sets the Capacity field of CStorVolume with provided arguments
func (b *Builder) WithCapacity(capacity string) *Builder {
	if len(capacity) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: missing capacity"),
		)
		return b
	}
	capacityQnt, err := resource.ParseQuantity(capacity)
	if err != nil {
		b.errs = append(
			b.errs,
			errors.Wrapf(
				err, "failed to build cstorvolume object: failed to parse capacity {%s}",
				capacity,
			),
		)
	}
	b.cstorvolume.object.Spec.Capacity = capacityQnt
	return b
}

// WithNodeBase sets the NodeBase field of CStorVolume with provided arguments
func (b *Builder) WithNodeBase(nodebase string) *Builder {
	if len(nodebase) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: missing nodebase"),
		)
		return b
	}
	b.cstorvolume.object.Spec.NodeBase = nodebase
	return b
}

// WithCStorIQN sets the iqn field of CStorVolume with provided arguments
func (b *Builder) WithCStorIQN(name string) *Builder {
	iqn := CStorNodeBase + ":" + name
	b.cstorvolume.object.Spec.Iqn = iqn
	return b
}

// WithIQN sets the IQN field of CStorVolume with provided arguments
func (b *Builder) WithIQN(iqn string) *Builder {
	if len(iqn) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: missing iqn"),
		)
		return b
	}
	b.cstorvolume.object.Spec.Iqn = iqn
	return b
}

// WithTargetPort sets the TargetPort field of
// CStorVolume with provided arguments
func (b *Builder) WithTargetPort(targetport string) *Builder {
	if len(targetport) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: missing targetport"),
		)
		return b
	}
	b.cstorvolume.object.Spec.TargetPort = targetport
	return b
}

// WithTargetPortal sets the TargetPortal field of
// CStorVolume with provided arguments
func (b *Builder) WithTargetPortal(targetportal string) *Builder {
	if len(targetportal) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: missing targetportal"),
		)
		return b
	}
	b.cstorvolume.object.Spec.TargetPortal = targetportal
	return b
}

// WithReplicationFactor sets the ReplicationFactor field of
// CStorVolume with provided arguments
func (b *Builder) WithReplicationFactor(replicationfactor int) *Builder {
	if replicationfactor <= 0 {
		b.errs = append(
			b.errs,
			errors.Errorf(
				"failed to build cstorvolume object: invalid replicationfactor {%d}",
				replicationfactor,
			),
		)
		return b
	}
	b.cstorvolume.object.Spec.ReplicationFactor = replicationfactor
	return b
}

// WithDesiredReplicationFactor sets the DesiredReplicationFactor field of
// CStorVolume with provided arguments
func (b *Builder) WithDesiredReplicationFactor(desiredRF int) *Builder {
	if desiredRF <= 0 {
		b.errs = append(
			b.errs,
			errors.Errorf(
				"failed to build cstorvolume object: invalid desiredreplicationfactor {%d}",
				desiredRF,
			),
		)
		return b
	}
	b.cstorvolume.object.Spec.DesiredReplicationFactor = desiredRF
	return b
}

// WithConsistencyFactor sets the ConsistencyFactor field of
// CStorVolume with provided arguments
func (b *Builder) WithConsistencyFactor(consistencyfactor int) *Builder {
	if consistencyfactor <= 0 {
		b.errs = append(
			b.errs,
			errors.Errorf(
				"failed to build cstorvolume object: invalid consistencyfactor {%d}",
				consistencyfactor,
			),
		)
		return b
	}
	b.cstorvolume.object.Spec.ConsistencyFactor = consistencyfactor
	return b
}

// WithNewVersion sets the current and desired version field of
// CStorVolume with provided arguments
func (b *Builder) WithNewVersion(version string) *Builder {
	if version == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build cstorvolume object: version can't be empty",
			),
		)
		return b
	}
	b.cstorvolume.object.VersionDetails.Status.Current = version
	b.cstorvolume.object.VersionDetails.Desired = version
	return b
}

// WithDependentsUpgraded sets the field to true for new volume
func (b *Builder) WithDependentsUpgraded() *Builder {
	b.cstorvolume.object.VersionDetails.Status.DependentsUpgraded = true
	return b
}

// Build returns the CStorVolume API instance
func (b *Builder) Build() (*apis.CStorVolume, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.cstorvolume.object, nil
}
