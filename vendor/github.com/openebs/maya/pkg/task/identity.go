/*
Copyright 2018 The OpenEBS Authors

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

package task

import (
	"fmt"

	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
)

// RunTaskKind represents type of runtask operation
type RunTaskKind string

// CommandKind is a runtask of type Command.
const (
	CommandKind RunTaskKind = "Command"
)

// MetaTaskIdentity will provide the required identity to a task
type MetaTaskIdentity struct {
	// Identity provides a unique identification to this
	// task.
	//
	// NOTE:
	//  Usage: There should not be two tasks with same identity
	// in a cas template engine run.
	Identity string `json:"id"`
	// Kind of the task
	Kind string `json:"kind"`
	// APIVersion of the task
	APIVersion string `json:"apiVersion"`
}

// taskIdentifier enables operations w.r.t a task's identity
type taskIdentifier struct {
	// identity identifies a task
	identity MetaTaskIdentity
}

func newTaskIdentifier(identity MetaTaskIdentity) (taskIdentifier, error) {
	if len(identity.Identity) == 0 {
		return taskIdentifier{}, fmt.Errorf("failed to create task identifier instance: task id is missing")
	}

	if len(identity.Kind) == 0 {
		return taskIdentifier{}, fmt.Errorf("failed to create task identifier instance: task kind is missing")
	}

	// apiversion is not mandatory for Command kind
	if len(identity.APIVersion) == 0 && string(identity.Kind) != string(CommandKind) {
		return taskIdentifier{}, fmt.Errorf("failed to create task identifier instance: task apiVersion is missing")
	}

	return taskIdentifier{
		identity: identity,
	}, nil
}

func (i taskIdentifier) isCommand() bool {
	return i.identity.Kind == string(CommandKind)
}

func (i taskIdentifier) isPod() bool {
	return i.identity.Kind == string(m_k8s_client.PodKK)
}

func (i taskIdentifier) isDeployment() bool {
	return i.identity.Kind == string(m_k8s_client.DeploymentKK)
}

func (i taskIdentifier) isReplicaSet() bool {
	return i.identity.Kind == string(m_k8s_client.ReplicaSetKK)
}

func (i taskIdentifier) isJob() bool {
	return i.identity.Kind == string(m_k8s_client.JobKK)
}

func (i taskIdentifier) isSTS() bool {
	return i.identity.Kind == string(m_k8s_client.STSKK)
}

func (i taskIdentifier) isService() bool {
	return i.identity.Kind == string(m_k8s_client.ServiceKK)
}

func (i taskIdentifier) isStoragePoolClaim() bool {
	return i.identity.Kind == string(m_k8s_client.StroagePoolClaimCRKK)
}

func (i taskIdentifier) isCStorPoolCluster() bool {
	return i.identity.Kind == string(m_k8s_client.CStorPoolClusterCRKK)
}

func (i taskIdentifier) isStoragePool() bool {
	return i.identity.Kind == string(m_k8s_client.StroagePoolCRKK)
}

func (i taskIdentifier) isUpgradeResult() bool {
	return i.identity.Kind == string(m_k8s_client.UpgradeResultCRKK)
}

func (i taskIdentifier) isConfigMap() bool {
	return i.identity.Kind == string(m_k8s_client.ConfigMapKK)
}

func (i taskIdentifier) isPVC() bool {
	return i.identity.Kind == string(m_k8s_client.PersistentVolumeClaimKK)
}

func (i taskIdentifier) isPV() bool {
	return i.identity.Kind == string(m_k8s_client.PersistentVolumeKK)
}

func (i taskIdentifier) isExtnV1B1() bool {
	return i.identity.APIVersion == string(m_k8s_client.ExtensionsV1Beta1KA)
}

func (i taskIdentifier) isBatchV1() bool {
	return i.identity.APIVersion == string(m_k8s_client.BatchV1KA)
}

func (i taskIdentifier) isAppsV1B1() bool {
	return i.identity.APIVersion == string(m_k8s_client.AppsV1B1KA)
}

func (i taskIdentifier) isAppsV1() bool {
	return i.identity.APIVersion == string(m_k8s_client.AppsV1KA)
}

func (i taskIdentifier) isCoreV1() bool {
	return i.identity.APIVersion == string(m_k8s_client.CoreV1KA)
}

func (i taskIdentifier) isOEV1alpha1() bool {
	return i.identity.APIVersion == string(m_k8s_client.OEV1alpha1KA)
}

func (i taskIdentifier) isBlockDevice() bool {
	return i.identity.Kind == string(m_k8s_client.BlockDeviceCRKK)
}

func (i taskIdentifier) isCstorPool() bool {
	return i.identity.Kind == string(m_k8s_client.CStorPoolCRKK)
}

func (i taskIdentifier) isCstorVolume() bool {
	return i.identity.Kind == string(m_k8s_client.CStorVolumeCRKK)
}

func (i taskIdentifier) isCstorVolumeReplica() bool {
	return i.identity.Kind == string(m_k8s_client.CStorVolumeReplicaCRKK)
}

func (i taskIdentifier) isStorageClass() bool {
	return i.identity.Kind == string(m_k8s_client.StorageClassKK)
}

func (i taskIdentifier) isVolumeSnapshotData() bool {
	return i.identity.Kind == string(m_k8s_client.VolumeSnapshotDataCRKK)
}

func (i taskIdentifier) isVolumeSnapshot() bool {
	return i.identity.Kind == string(m_k8s_client.VolumeSnapshotCRKK)
}

func (i taskIdentifier) isStorageV1() bool {
	return i.identity.APIVersion == string(m_k8s_client.StorageV1KA)
}

func (i taskIdentifier) isStorageV1SC() bool {
	return i.isStorageV1() && i.isStorageClass()
}

func (i taskIdentifier) isOEV1alpha1BlockDevice() bool {
	return i.isOEV1alpha1() && i.isBlockDevice()
}

func (i taskIdentifier) isOEV1alpha1CSP() bool {
	return i.isOEV1alpha1() && i.isCstorPool()
}

func (i taskIdentifier) isExtnV1B1Deploy() bool {
	return i.isExtnV1B1() && i.isDeployment()
}

func (i taskIdentifier) isExtnV1B1ReplicaSet() bool {
	return i.isExtnV1B1() && i.isReplicaSet()
}

func (i taskIdentifier) isBatchV1Job() bool {
	return i.isBatchV1() && i.isJob()
}

func (i taskIdentifier) isAppsV1STS() bool {
	return i.isAppsV1() && i.isSTS()
}

func (i taskIdentifier) isAppsV1B1Deploy() bool {
	return i.isAppsV1B1() && i.isDeployment()
}

func (i taskIdentifier) isAppsV1Deploy() bool {
	return i.isAppsV1() && i.isDeployment()
}

func (i taskIdentifier) isCoreV1Pod() bool {
	return i.isCoreV1() && i.isPod()
}

func (i taskIdentifier) isCoreV1Service() bool {
	return i.isCoreV1() && i.isService()
}

func (i taskIdentifier) isCoreV1PVC() bool {
	return i.isCoreV1() && i.isPVC()
}

func (i taskIdentifier) isCoreV1PV() bool {
	return i.isCoreV1() && i.isPV()
}

func (i taskIdentifier) isOEV1alpha1SPC() bool {
	return i.isOEV1alpha1() && i.isStoragePoolClaim()
}

func (i taskIdentifier) isOEV1alpha1CSPC() bool {
	return i.isOEV1alpha1() && i.isCStorPoolCluster()
}

func (i taskIdentifier) isOEV1alpha1SP() bool {
	return i.isOEV1alpha1() && i.isStoragePool()
}

func (i taskIdentifier) isOEV1alpha1UR() bool {
	return i.isOEV1alpha1() && i.isUpgradeResult()
}

func (i taskIdentifier) isOEV1alpha1CV() bool {
	return i.isOEV1alpha1() && i.isCstorVolume()
}

func (i taskIdentifier) isOEV1alpha1CVR() bool {
	return i.isOEV1alpha1() && i.isCstorVolumeReplica()
}
