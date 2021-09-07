/*
Copyright 2021 The OpenEBS Authors

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

package tests

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/container"
	pvc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/persistentvolumeclaim"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/pod"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/volume"
	sc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/storage/v1/storageclass"
	"github.com/openebs/dynamic-localpv-provisioner/tests/disk"
)

const (
	// DiskImageSize is the default file size(1GB) used while creating backing image
	DiskImageSize = 1073741824
)

var _ = Describe("TEST HOSTPATH XFS QUOTA LOCAL PV WITH NON-XFS FILESYSTEM", func() {
	var (
		pvcObj        *corev1.PersistentVolumeClaim
		scObj         *storagev1.StorageClass
		accessModes   = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity      = "5M"
		podName       = "busybox-hostpath"
		label         = "demo=hostpath-pod"
		createdPvc    *corev1.PersistentVolumeClaim
		pvcName       = "pvc-hp"
		scNamePrefix  = "sc-hp-xfs"
		scName        string
		podObj        *corev1.Pod
		labelselector = map[string]string{
			"demo": "hostpath-pod",
		}
	)

	When("StorageClass with valid xfs quota parameters is created", func() {
		It("should create a StorageClass", func() {
			By("building a StorageClass")
			scObj, err = sc.NewStorageClass(
				sc.WithGenerateName(scNamePrefix),
				sc.WithLabels(map[string]string{
					"openebs.io/test-sc": "true",
				}),
				sc.WithLocalPV(),
				sc.WithHostpath("/var/openebs/integration-test"),
				sc.WithVolumeBindingMode("WaitForFirstConsumer"),
				sc.WithReclaimPolicy("Delete"),
				sc.WithParameters(map[string]string{
					"enableXfsQuota": "true",
					"softLimitGrace": "20%",
					"hardLimitGrace": "50%",
				}),
			)
			Expect(err).To(
				BeNil(),
				"while building StorageClass with name prefix {%s}",
				scNamePrefix,
			)

			By("creating StorageClass API resource")
			scObj, err = ops.SCClient.Create(context.TODO(), scObj)
			Expect(err).To(
				BeNil(),
				"while creating StorageClass with name prefix {%s}",
				scNamePrefix,
			)
			scName = scObj.ObjectMeta.Name
		})
	})

	When("pvc with storageclass "+scName+" is created", func() {
		It("should create a pvc", func() {
			By("building a PVC with StorageClass " + scName)
			pvcObj, err = BuildPersistentVolumeClaim(pvcName, scName, capacity, accessModes)
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("creating above PVC with StorageClass " + scName)
			createdPvc, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(context.TODO(), pvcObj)
			Expect(err).To(
				BeNil(),
				"while creating PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)
		})
	})

	When("pod is created with pvc "+pvcName, func() {
		It("should be in pending state and so do pvc", func() {
			By("building a pod with busybox image")
			podObj, err = BuildPod(podName, pvcName, labelselector)
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pod {%s} in namespace {%s}",
				podName,
				namespaceObj.Name,
			)

			By("creating above pod with busybox image")
			createdPod, err := ops.PodClient.WithNamespace(namespaceObj.Name).
				Create(context.TODO(), podObj)
			Expect(err).To(
				BeNil(),
				"while creating pod {%s} in namespace {%s}",
				podName,
				namespaceObj.Name,
			)

			By("verifying pod status as pending")
			checkPhase := ops.CheckPodStatusEventually(createdPod, corev1.PodPending)
			Expect(checkPhase).To(Equal(true), "while verifying pod pending status")

			By("verifying the pvc phase as pending")
			Expect(createdPvc.Status.Phase).To(Equal(corev1.ClaimPending), "while verifying the pvc pending state")
		})
	})

	When("Pod consuming pvc "+pvcName+" is deleted", func() {
		It("should delete the pod", func() {
			By("deleting above pod")
			err = ops.PodClient.WithNamespace(namespaceObj.Name).Delete(context.TODO(), podName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pod {%s} in namespace {%s}",
				podName,
				namespaceObj.Name,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodCountEventually(namespaceObj.Name, label, nil, 0)
			Expect(podCount).To(Equal(0), "while verifying pod count")
		})
	})

	When("pvc with storageclass "+scName+" is deleted", func() {
		It("should delete the pvc", func() {
			By("deleting the PVC with StorageClass " + scName)
			pvName := ops.GetPVNameFromPVCName(pvcName)
			Expect(pvName).To(
				BeEmpty(),
				"while getting Spec.VolumeName from PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("deleting above PVC")
			err = ops.PVCClient.Delete(context.TODO(), pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("verifying PVC is deleted")
			status := ops.IsPVCDeletedEventually(pvcName, namespaceObj.Name)
			Expect(status).To(
				BeTrue(),
				"when checking status of deleted PVC {%s}",
				pvcName,
			)
		})
	})
})

var _ = Describe("TEST HOSTPATH XFS QUOTA LOCAL PV WITH XFS FILESYSTEM", func() {
	var (
		pvcObj        *corev1.PersistentVolumeClaim
		scObj         *storagev1.StorageClass
		accessModes   = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity      = "5M"
		podName       = "busybox-hostpath"
		label         = "demo=hostpath-pod"
		createdPod    *corev1.Pod
		pvcName       = "pvc-hp"
		scNamePrefix  = "sc-hp-xfs"
		scName        string
		podObj        *corev1.Pod
		labelselector = map[string]string{
			"demo": "hostpath-pod",
		}
	)

	When("StorageClass with valid xfs quota parameters is created", func() {
		It("should create a StorageClass", func() {
			By("building a StorageClass")
			scObj, err = sc.NewStorageClass(
				sc.WithGenerateName(scNamePrefix),
				sc.WithLabels(map[string]string{
					"openebs.io/test-sc": "true",
				}),
				sc.WithLocalPV(),
				sc.WithHostpath(xfsHostpathDir),
				sc.WithVolumeBindingMode("WaitForFirstConsumer"),
				sc.WithReclaimPolicy("Delete"),
				sc.WithParameters(map[string]string{
					"enableXfsQuota": "true",
					"softLimitGrace": "0%",
					"hardLimitGrace": "0%",
				}),
			)
			Expect(err).To(
				BeNil(),
				"while building StorageClass with name prefix {%s}",
				scNamePrefix,
			)

			By("creating StorageClass API resource")
			scObj, err = ops.SCClient.Create(context.TODO(), scObj)
			Expect(err).To(
				BeNil(),
				"while creating StorageClass with name prefix {%s}",
				scNamePrefix,
			)
			scName = scObj.ObjectMeta.Name
		})
	})

	When("pvc with storageclass "+scName+" is created", func() {
		It("should create a pvc", func() {
			By("building a PVC with StorageClass " + scName)
			pvcObj, err = BuildPersistentVolumeClaim(pvcName, scName, capacity, accessModes)
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("creating above PVC with StorageClass " + scName)
			_, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(context.TODO(), pvcObj)
			Expect(err).To(
				BeNil(),
				"while creating PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)
		})
	})

	When("pod is created with pvc "+pvcName, func() {
		It("should be up and running", func() {
			By("building a pod with busybox image")
			podObj, err = BuildPod(podName, pvcName, labelselector)
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pod {%s} in namespace {%s}",
				podName,
				namespaceObj.Name,
			)

			By("creating above pod with busybox image")
			createdPod, err = ops.PodClient.WithNamespace(namespaceObj.Name).
				Create(context.TODO(), podObj)
			Expect(err).To(
				BeNil(),
				"while creating pod {%s} in namespace {%s}",
				podName,
				namespaceObj.Name,
			)

			By("verifying pod count as 1")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 1)
			Expect(podCount).To(Equal(1), "while verifying pod count")
		})
	})

	When("writing data more than quota limit into the hostpath volume", func() {
		It("should not be able to write more than the enforced limit", func() {
			By("Verifying the quota applied on the volume works")
			execOption := NewOptions()
			// The command formed below is attempting to create a file named “test.txt” inside the attached volume,
			// its size increasing 1M in each iteration of the loop should show us the quota in action.
			option := execOption.WithPodName(createdPod.Name).
				WithContainer("busybox").
				WithNamespace(namespaceObj.Name).
				WithCommand([]string{"/bin/sh", "-c", fmt.Sprintf(`for i in $(seq 1 7); do dd if=/dev/zero of=/mnt/store1/test.txt bs=1M count=$i 2>/dev/null; if [ $? -ne 0 ]; then echo "filesize of $i Mb failed, quota reached!"; else echo "filesize of $i Mb was ok"; fi; done`)}...)
			stdOut, stdErr, err := ops.ExecPod(option)
			fmt.Printf("When running command to test enforced quota. stdOut: {%s}, stderr: {%s}, error: {%v}", stdOut, stdErr, err)
			Expect(stdErr).NotTo(BeNil(), "trying to write beyond the quota limit should not be allowed")
			Expect(stdOut).Should(ContainSubstring("filesize of 4 Mb was ok"), "trying to write beyond the quota limit should not be allowed")
			Expect(stdOut).Should(ContainSubstring("filesize of 5 Mb failed, quota reached!"), "trying to write beyond the quota limit should not be allowed")
		})
	})

	When("Pod consuming pvc "+pvcName+" is deleted", func() {
		It("should delete the pod", func() {
			By("deleting above pod")
			err = ops.PodClient.WithNamespace(namespaceObj.Name).Delete(context.TODO(), podName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pod {%s} in namespace {%s}",
				podName,
				namespaceObj.Name,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 0)
			Expect(podCount).To(Equal(0), "while verifying pod count")
		})
	})

	When("pvc with storageclass "+scName+" is deleted", func() {
		It("should delete the pvc", func() {
			By("deleting the PVC with StorageClass " + scName)
			pvName := ops.GetPVNameFromPVCName(pvcName)
			Expect(pvName).ToNot(
				BeEmpty(),
				"while getting Spec.VolumeName from "+
					"PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("deleting above PVC")
			err = ops.PVCClient.Delete(context.TODO(), pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("verifying PVC is deleted")
			status := ops.IsPVCDeletedEventually(pvcName, namespaceObj.Name)
			Expect(status).To(
				BeTrue(),
				"when checking status of deleted PVC {%s}",
				pvcName,
			)
		})
	})
})

// BuildPersistentVolumeClaim builds the PVC object
func BuildPersistentVolumeClaim(pvcName, scName, capacity string, accessModes []corev1.PersistentVolumeAccessMode) (*corev1.PersistentVolumeClaim, error) {
	return pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(namespaceObj.Name).
		WithStorageClass(scName).
		WithAccessModes(accessModes).
		WithCapacity(capacity).Build()
}

// BuildPod builds the pod object
func BuildPod(podName, pvcName string, labelselector map[string]string) (*corev1.Pod, error) {
	return pod.NewBuilder().
		WithName(podName).
		WithNamespace(namespaceObj.Name).
		WithLabels(labelselector).
		WithContainerBuilder(
			container.NewBuilder().
				WithName("busybox").
				WithImage("busybox").
				WithCommandNew(
					[]string{
						"/bin/sh",
					},
				).
				WithArgumentsNew(
					[]string{
						"-c",
						"sleep 3600",
					},
				).
				WithVolumeMountsNew(
					[]corev1.VolumeMount{
						{
							Name:      "demo-vol1",
							MountPath: "/mnt/store1",
						},
					},
				),
		).
		WithVolumeBuilder(
			volume.NewBuilder().
				WithName("demo-vol1").
				WithPVCSource(pvcName),
		).
		Build()
}

// PrepareDisk prepares the setup necessary for testing xfs hostpath quota
func PrepareDisk(fsType, hostPath string) (disk.Disk, error) {
	physicalDisk := disk.NewDisk(DiskImageSize)

	err := physicalDisk.CreateLoopDevice()
	if err != nil {
		return physicalDisk, errors.Wrapf(err, "while creating loop back device with disk %+v", physicalDisk)
	}

	// Make xfs fs on the created loopback device
	err = physicalDisk.CreateFileSystem(fsType)
	if err != nil {
		return physicalDisk, errors.Wrapf(err, "while formatting the disk {%+v} with xfs fs", physicalDisk)
	}

	err = disk.MkdirAll(hostPath)
	if err != nil {
		return physicalDisk, errors.Wrapf(err, "while making a new directory {%s}", hostPath)
	}

	// Mount the xfs formatted loopback device
	err = physicalDisk.Mount(hostPath)
	if err != nil {
		return physicalDisk, errors.Wrapf(err, "while mounting the disk with pquota option {%+v}", physicalDisk)
	}

	return physicalDisk, nil
}

// DestroyDisk performs performs the clean-up task after testing the features
func DestroyDisk(physicalDisk disk.Disk, hostPath string) error {
	// Unmount the disk
	err = physicalDisk.Unmount()
	if err != nil {
		return errors.Wrapf(err, "while unmounting the disk {%+v}", physicalDisk)
	}

	// Detach and delete the disk
	err = physicalDisk.DetachAndDeleteDisk()
	if err != nil {
		return errors.Wrapf(err, "while detaching and deleting the disk {%+v}", physicalDisk)
	}

	// Deleting the hostpath directory
	err = disk.RunCommandWithSudo("rm -rf " + hostPath)
	if err != nil {
		return errors.Wrapf(err, "while deleting the mountpoint directory {%s}", hostPath)
	}

	return nil
}
