/*
Copyright 2017 The OpenEBS Authors

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

package k8s

import (
	"fmt"

	"github.com/ghodss/yaml"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/template"
	api_apps_v1 "k8s.io/api/apps/v1"
	api_apps_v1beta1 "k8s.io/api/apps/v1beta1"
	api_batch_v1 "k8s.io/api/batch/v1"
	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
)

// STSYml helps generating kubernetes StatefulSet
// object
type STSYml struct {
	// YmlInBytes represents a kubernetes StatefulSet
	// in yaml format
	YmlInBytes []byte
}

// NewSTSYml returns a new instance of STSYml
func NewSTSYml(context, yml string, values map[string]interface{}) (*STSYml, error) {
	b, err := template.AsTemplatedBytes(context, yml, values)
	if err != nil {
		return nil, err
	}
	return &STSYml{
		YmlInBytes: b,
	}, nil
}

// AsAppsV1STS returns a apps/v1 StatefulSet instance
func (m *STSYml) AsAppsV1STS() (*api_apps_v1.StatefulSet, error) {
	if m.YmlInBytes == nil {
		return nil, fmt.Errorf("missing statefulset yaml")
	}
	sts := &api_apps_v1.StatefulSet{}
	err := yaml.Unmarshal(m.YmlInBytes, sts)
	if err != nil {
		return nil, err
	}
	return sts, nil
}

// JobYml helps generating kubernetes Job object
type JobYml struct {
	YmlInBytes []byte // YmlInBytes represents a K8s Job in yaml format
}

// NewJobYml returns a new instance of JobYml
func NewJobYml(context, yml string, values map[string]interface{}) (*JobYml, error) {
	b, err := template.AsTemplatedBytes(context, yml, values)
	if err != nil {
		return nil, err
	}
	return &JobYml{
		YmlInBytes: b,
	}, nil
}

// AsBatchV1Job returns a batch/v1 Job instance
func (m *JobYml) AsBatchV1Job() (*api_batch_v1.Job, error) {
	if m.YmlInBytes == nil {
		return nil, fmt.Errorf("Missing yaml")
	}
	job := &api_batch_v1.Job{}
	err := yaml.Unmarshal(m.YmlInBytes, job)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// DeploymentYml provides utility methods to generate K8s Deployment objects
type DeploymentYml struct {
	// YmlInBytes represents a K8s Deployment in
	// yaml format
	YmlInBytes []byte
}

// NewDeploymentYml returns a DeploymentYml instance.
func NewDeploymentYml(context, yml string, values map[string]interface{}) (*DeploymentYml, error) {
	b, err := template.AsTemplatedBytes(context, yml, values)
	if err != nil {
		return nil, err
	}

	return &DeploymentYml{
		YmlInBytes: b,
	}, nil
}

// AsExtnV1B1Deployment returns a extensions/v1beta1 Deployment instance
func (m *DeploymentYml) AsExtnV1B1Deployment() (*api_extn_v1beta1.Deployment, error) {
	if m.YmlInBytes == nil {
		return nil, fmt.Errorf("Missing yaml")
	}

	// unmarshall the byte into k8s Deployment object
	deploy := &api_extn_v1beta1.Deployment{}
	err := yaml.Unmarshal(m.YmlInBytes, deploy)
	if err != nil {
		return nil, err
	}

	return deploy, nil
}

// AsAppsV1B1Deployment returns a apps/v1 Deployment instance
func (m *DeploymentYml) AsAppsV1B1Deployment() (*api_apps_v1beta1.Deployment, error) {
	if m.YmlInBytes == nil {
		return nil, fmt.Errorf("Missing yaml")
	}

	// unmarshall the byte into k8s Deployment object
	deploy := &api_apps_v1beta1.Deployment{}
	err := yaml.Unmarshal(m.YmlInBytes, deploy)
	if err != nil {
		return nil, err
	}

	return deploy, nil
}

// AsAppsV1Deployment returns a apps/v1 Deployment instance
func (m *DeploymentYml) AsAppsV1Deployment() (*api_apps_v1.Deployment, error) {
	if m.YmlInBytes == nil {
		return nil, fmt.Errorf("Missing yaml")
	}

	// unmarshall the byte into k8s Deployment object
	deploy := &api_apps_v1.Deployment{}
	err := yaml.Unmarshal(m.YmlInBytes, deploy)
	if err != nil {
		return nil, err
	}

	return deploy, nil
}

// ServiceYml struct provides utility methods to generate K8s Service objects
type ServiceYml struct {
	// YmlInBytes represents a K8s Service in
	// yaml format
	YmlInBytes []byte
}

// NewServiceYml returns a ServiceYml instance.
func NewServiceYml(context, yml string, values map[string]interface{}) (*ServiceYml, error) {
	b, err := template.AsTemplatedBytes(context, yml, values)
	if err != nil {
		return nil, err
	}

	return &ServiceYml{
		YmlInBytes: b,
	}, nil
}

// AsCoreV1Service returns a v1 Service instance
func (m *ServiceYml) AsCoreV1Service() (*api_core_v1.Service, error) {
	if m.YmlInBytes == nil {
		return nil, fmt.Errorf("Missing yaml")
	}

	// unmarshall the byte into k8s Service object
	svc := &api_core_v1.Service{}
	err := yaml.Unmarshal(m.YmlInBytes, svc)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

//CStorPoolYml provides utility methods to generate K8s CStorPool objects
type CStorPoolYml struct {
	// YmlInBytes represents a CStorPool in
	// yaml format
	YmlInBytes []byte
}

//CStorVolumeYml provides utility methods to generate K8s CStorVolume objects
type CStorVolumeYml struct {
	// YmlInBytes represents a CStorVolume in
	// yaml format
	YmlInBytes []byte
}

//StoragePoolYml provides utility methods to generate K8s StoragePool objects
type StoragePoolYml struct {
	// YmlInBytes represents a StoragePool in
	// yaml format
	YmlInBytes []byte
}

// NewCStorPoolYml returns a CStorPoolYml instance.
func NewCStorPoolYml(context, yml string, values map[string]interface{}) (*CStorPoolYml, error) {
	b, err := template.AsTemplatedBytes(context, yml, values)
	if err != nil {
		return nil, err
	}

	return &CStorPoolYml{
		YmlInBytes: b,
	}, nil
}

// NewStoragePoolYml returns a StoragePoolYml instance.
func NewStoragePoolYml(context, yml string, values map[string]interface{}) (*StoragePoolYml, error) {
	b, err := template.AsTemplatedBytes(context, yml, values)
	if err != nil {
		return nil, err
	}

	return &StoragePoolYml{
		YmlInBytes: b,
	}, nil
}

// NewCStorVolumeYml returns a new CStorVolumeYml instance based on yml string and values.
func NewCStorVolumeYml(context, yml string, values map[string]interface{}) (*CStorVolumeYml, error) {
	b, err := template.AsTemplatedBytes(context, yml, values)
	if err != nil {
		return nil, err
	}

	return &CStorVolumeYml{
		YmlInBytes: b,
	}, nil
}

// AsCStorPoolYml returns a v1 CStorPool instance
func (m *CStorPoolYml) AsCStorPoolYml() (*v1alpha1.CStorPool, error) {
	if m.YmlInBytes == nil {
		return nil, fmt.Errorf("Missing yaml")
	}

	// unmarshall the byte into CStorVolume object
	cstorPool := &v1alpha1.CStorPool{}
	err := yaml.Unmarshal(m.YmlInBytes, cstorPool)
	if err != nil {
		return nil, err
	}

	return cstorPool, nil
}

// AsStoragePoolYml returns a v1 StoragePool instance
func (m *StoragePoolYml) AsStoragePoolYml() (*v1alpha1.StoragePool, error) {
	if m.YmlInBytes == nil {
		return nil, fmt.Errorf("Missing yaml")
	}

	// unmarshall the byte into StoragePool object
	storagePool := &v1alpha1.StoragePool{}
	err := yaml.Unmarshal(m.YmlInBytes, storagePool)
	if err != nil {
		return nil, err
	}

	return storagePool, nil
}

// AsCStorVolumeYml returns a v1 CStorVolume instance
func (m *CStorVolumeYml) AsCStorVolumeYml() (*v1alpha1.CStorVolume, error) {
	if m.YmlInBytes == nil {
		return nil, fmt.Errorf("Missing yaml")
	}

	// unmarshall the byte into CStorVolume object
	cstorVolume := &v1alpha1.CStorVolume{}
	err := yaml.Unmarshal(m.YmlInBytes, cstorVolume)
	if err != nil {
		return nil, err
	}

	return cstorVolume, nil
}

// CStorVolumeReplicaYml provides utility methods to generate K8s CStorVolumeReplica objects
type CStorVolumeReplicaYml struct {
	// YmlInBytes represents a CStorVolumeReplica in
	// yaml format
	YmlInBytes []byte
}

// NewCStorVolumeReplicaYml returns a CStorVolumeReplicaYml instance.
func NewCStorVolumeReplicaYml(context, yml string, values map[string]interface{}) (*CStorVolumeReplicaYml, error) {
	b, err := template.AsTemplatedBytes(context, yml, values)
	if err != nil {
		return nil, err
	}

	return &CStorVolumeReplicaYml{
		YmlInBytes: b,
	}, nil
}

// AsCStorVolumeReplicaYml returns a v1 Service instance
func (m *CStorVolumeReplicaYml) AsCStorVolumeReplicaYml() (*v1alpha1.CStorVolumeReplica, error) {
	if m.YmlInBytes == nil {
		return nil, fmt.Errorf("Missing yaml")
	}

	// unmarshall the byte into CStorVolumeReplica object
	cstorVolumeReplica := &v1alpha1.CStorVolumeReplica{}
	err := yaml.Unmarshal(m.YmlInBytes, cstorVolumeReplica)
	if err != nil {
		return nil, err
	}

	return cstorVolumeReplica, nil
}
