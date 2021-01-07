/*
Copyright 2019-2020 The OpenEBS Authors

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

package pod

import (
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/container"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/volume"
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

const (
	// k8sNodeLabelKeyHostname is the label key used by Kubernetes
	// to store the hostname on the node resource.
	k8sNodeLabelKeyHostname = "kubernetes.io/hostname"
)

// Builder is the builder object for Pod
type Builder struct {
	pod  *Pod
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{pod: &Pod{object: &corev1.Pod{}}}
}

// WithTolerationsForTaints sets the Spec.Tolerations with provided taints.
func (b *Builder) WithTolerationsForTaints(taints ...corev1.Taint) *Builder {

	tolerations := []corev1.Toleration{}
	for i := range taints {
		var toleration corev1.Toleration
		toleration.Key = taints[i].Key
		toleration.Effect = taints[i].Effect
		if len(taints[i].Value) == 0 {
			toleration.Operator = corev1.TolerationOpExists
		} else {
			toleration.Value = taints[i].Value
			toleration.Operator = corev1.TolerationOpEqual
		}
		tolerations = append(tolerations, toleration)
	}

	b.pod.object.Spec.Tolerations = append(
		b.pod.object.Spec.Tolerations,
		tolerations...,
	)
	return b
}

// WithName sets the Name field of Pod with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing Pod name"),
		)
		return b
	}
	b.pod.object.Name = name
	return b
}

// WithNamespace sets the Namespace field of Pod with provided value.
func (b *Builder) WithNamespace(namespace string) *Builder {
	if len(namespace) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing namespace"),
		)
		return b
	}
	b.pod.object.Namespace = namespace
	return b
}

// WithContainerBuilder adds a container to this pod object.
//
// NOTE:
//   container details are present in the provided container
// builder object
func (b *Builder) WithContainerBuilder(
	containerBuilder *container.Builder,
) *Builder {
	containerObj, err := containerBuilder.Build()
	if err != nil {
		b.errs = append(b.errs, errors.Wrap(err, "failed to build pod"))
		return b
	}
	b.pod.object.Spec.Containers = append(
		b.pod.object.Spec.Containers,
		containerObj,
	)
	return b
}

// WithVolumeBuilder sets Volumes field of deployment.
func (b *Builder) WithVolumeBuilder(volumeBuilder *volume.Builder) *Builder {
	vol, err := volumeBuilder.Build()
	if err != nil {
		b.errs = append(b.errs, errors.Wrap(err, "failed to build deployment"))
		return b
	}
	b.pod.object.Spec.Volumes = append(
		b.pod.object.Spec.Volumes,
		*vol,
	)
	return b
}

// WithRestartPolicy sets the RestartPolicy field in Pod with provided arguments
func (b *Builder) WithRestartPolicy(
	restartPolicy corev1.RestartPolicy,
) *Builder {
	b.pod.object.Spec.RestartPolicy = restartPolicy
	return b
}

// WithNodeName sets the NodeName field of Pod with provided value.
func (b *Builder) WithNodeName(nodeName string) *Builder {
	if len(nodeName) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing Pod node name"),
		)
		return b
	}
	b.pod.object.Spec.NodeName = nodeName
	return b
}

// WithNodeSelectorHostnameNew sets the Pod NodeSelector to the provided hostname value
// This function replaces (resets) the NodeSelector to use only hostname selector
func (b *Builder) WithNodeSelectorHostnameNew(hostname string) *Builder {
	if len(hostname) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing Pod hostname"),
		)
		return b
	}

	b.pod.object.Spec.NodeSelector = map[string]string{
		k8sNodeLabelKeyHostname: hostname,
	}

	return b
}

// WithNodeAffinityNew sets the NodeAffinity field of Pod with provided node label and key
func (b *Builder) WithNodeAffinityNew(key, value string) *Builder {
	if len(key) == 0 || len(value) == 0 {
		b.errs = append(b.errs, errors.New("failed to build Pod object: missing node label key or value"))
		return b
	}
	nodeAffinity := &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      key,
							Operator: corev1.NodeSelectorOpIn,
							Values: []string{
								value,
							},
						},
					},
				},
			},
		},
	}
	b.pod.object.Spec.Affinity = &corev1.Affinity{
		NodeAffinity: nodeAffinity,
	}
	return b
}

// WithContainers sets the Containers field in Pod with provided arguments
func (b *Builder) WithContainers(containers []corev1.Container) *Builder {
	if len(containers) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing containers"),
		)
		return b
	}
	b.pod.object.Spec.Containers = containers
	return b
}

// WithContainer sets the Containers field in Pod with provided arguments
func (b *Builder) WithContainer(container corev1.Container) *Builder {
	return b.WithContainers([]corev1.Container{container})
}

// WithImagePullSecrets sets the pod image pull secrets
// + optional
// if the length is zero that no secret is needed to pull the image
func (b *Builder) WithImagePullSecrets(secrets []corev1.LocalObjectReference) *Builder {
	if len(secrets) == 0 {
		return b
	}
	b.pod.object.Spec.ImagePullSecrets = secrets
	return b
}

// WithVolumes sets the Volumes field in Pod with provided arguments
func (b *Builder) WithVolumes(volumes []corev1.Volume) *Builder {
	if len(volumes) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing volumes"),
		)
		return b
	}
	b.pod.object.Spec.Volumes = volumes
	return b
}

// WithVolume sets the Volumes field in Pod with provided arguments
func (b *Builder) WithVolume(volume corev1.Volume) *Builder {
	return b.WithVolumes([]corev1.Volume{volume})
}

// WithServiceAccountName sets the ServiceAccountName of Pod spec with
// the provided value
func (b *Builder) WithServiceAccountName(serviceAccountName string) *Builder {
	if len(serviceAccountName) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing Pod service account name"),
		)
		return b
	}
	b.pod.object.Spec.ServiceAccountName = serviceAccountName
	return b
}

// Build returns the Pod API instance
func (b *Builder) Build() (*corev1.Pod, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.pod.object, nil
}
