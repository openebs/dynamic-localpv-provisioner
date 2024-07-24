package tests

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/storage/v1/storageclass"
)

var _ = Describe("TEST HOSTPATH XFS QUOTA LOCAL PV WITH UNSUPPORTED FILESYSTEM", func() {
	var (
		pvcObj      *corev1.PersistentVolumeClaim
		scObj       *storagev1.StorageClass
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity    = "5M"
		podName     = "busybox-hostpath"
		label       = "demo=hostpath-pod"
		createdPvc  *corev1.PersistentVolumeClaim

		pvcName       = "pvc-hp-xfs"
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
				sc.WithHostpath(loopHostpathDir),
				sc.WithXfsQuota("20%", "50%"),
				sc.WithVolumeBindingMode("WaitForFirstConsumer"),
				sc.WithReclaimPolicy("Delete"),
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
			Expect(scName).NotTo(BeEmpty(), "SC name should not be empty")
		})
	})

	When("pvc with storageclass "+scName+" is created", func() {
		It("should create a pvc", func() {
			By("Writing minix filesystem into the loop device")
			//umount
			errs := physicalDisk.Unmount()
			Expect(errs).To(BeEmpty(), "when unmounting loop device")
			//wipefs
			err := physicalDisk.Wipefs()
			Expect(err).To(BeNil(), "when wiping filesystem from loop device")
			//mkfs
			err = physicalDisk.CreateFilesystem("minix")
			Expect(err).To(BeNil(), "when writing filesystem to loop device")
			//mount
			err = physicalDisk.PrjquotaMount(loopHostpathDir)
			Expect(err).To(
				BeNil(),
				"when mounting loop device filesystem to mountpoint",
			)

			By("building a PVC with StorageClass " + scName)
			Expect(scName).NotTo(BeEmpty(), "SC name should not be empty")
			pvcObj, err = BuildPersistentVolumeClaim(namespaceObj.Name, pvcName, scName, capacity, accessModes)
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
			podObj, err = BuildPod(namespaceObj.Name, podName, pvcName, labelselector)
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pod {%s} in namespace {%s}",
				podName,
				namespaceObj.Name,
			)

			By("creating above pod with busybox image")
			_, err := ops.PodClient.WithNamespace(namespaceObj.Name).
				Create(context.TODO(), podObj)
			Expect(err).To(
				BeNil(),
				"while creating pod {%s} in namespace {%s}",
				podName,
				namespaceObj.Name,
			)

			By("verifying pod status as pending")
			//Giving the check till timeout to try to see if the Status changes to Running
			podPhase := ops.CheckPodStatusEventually(namespaceObj.Name, podName, corev1.PodRunning)
			Expect(podPhase).To(Equal(corev1.PodPending), "while verifying pod pending status")

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
			By("verifying that the VolumeName is empty for the PVC")
			pvName := ops.GetPVNameFromPVCName(namespaceObj.Name, pvcName)
			Expect(pvName).To(
				BeEmpty(),
				"while getting Spec.VolumeName from PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("deleting above PVC")
			err = ops.PVCClient.WithNamespace(namespaceObj.Name).Delete(context.TODO(), pvcName, &metav1.DeleteOptions{})
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

			By("removing the minix filesystem")
			//umount
			errs := physicalDisk.Unmount()
			Expect(errs).To(BeEmpty(), "when unmounting loop device")
			//wipefs
			err := physicalDisk.Wipefs()
			Expect(err).To(BeNil(), "when wiping filesystem from loop device")
		})
	})
})

var _ = Describe("TEST HOSTPATH XFS QUOTA LOCAL PV WITH XFS FILESYSTEM", func() {
	var (
		pvcObj        *corev1.PersistentVolumeClaim
		scObj         *storagev1.StorageClass
		accessModes   = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacityInMib = "5"
		capacity      = capacityInMib + "M"
		podName       = "busybox-hostpath"
		label         = "demo=hostpath-pod"
		createdPod    *corev1.Pod
		pvcName       = "pvc-hp-xfs"
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
				sc.WithHostpath(loopHostpathDir),
				sc.WithXfsQuota("", ""),
				sc.WithVolumeBindingMode("WaitForFirstConsumer"),
				sc.WithReclaimPolicy("Delete"),
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
			Expect(scName).NotTo(BeEmpty(), "SC name should not be empty")
		})
	})

	When("pvc with storageclass "+scName+" is created", func() {
		It("should create a pvc", func() {
			By("Writing xfs filesystem into the loop device")
			//umount
			errs := physicalDisk.Unmount()
			Expect(errs).To(BeEmpty(), "when unmounting loop device")
			//wipefs
			err := physicalDisk.Wipefs()
			Expect(err).To(BeNil(), "when wiping filesystem from loop device")
			//mkfs
			err = physicalDisk.CreateFilesystem("xfs")
			Expect(err).To(BeNil(), "when writing filesystem to loop device")
			//mount
			err = physicalDisk.PrjquotaMount(loopHostpathDir)
			Expect(err).To(
				BeNil(),
				"when mounting loop device filesystem to mountpoint",
			)

			By("building a PVC with StorageClass " + scName)
			Expect(scName).NotTo(BeEmpty(), "SC name should not be empty")
			pvcObj, err = BuildPersistentVolumeClaim(namespaceObj.Name, pvcName, scName, capacity, accessModes)
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
			podObj, err = BuildPod(namespaceObj.Name, podName, pvcName, labelselector)
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
				WithCommand([]string{"/bin/sh", "-c", "dd if=/dev/zero of=/mnt/store1/test.txt bs=1M count=10 2>&1 || du -sm /mnt/store1 | cut -f -1 | tr -d '\n' 1>&2"}...)
			stdOut, stdErr, err := ops.ExecPod(option)
			fmt.Printf("When running command to test enforced quota. stdOut: {%s}, stderr: {%s}, error: {%v}", stdOut, stdErr, err)
			Expect(err).To(BeNil(), "while exec'ing into the pod and running command(s)")
			Expect(stdOut).NotTo(BeEmpty(), "trying to write till the quota limit should be allowed")
			Expect(stdErr).To(Equal(capacityInMib), "trying to write beyond the quota limit should not be allowed")
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
			By("getting the PV name from Bound PVC object spec")
			pvName := ops.GetPVNameFromPVCName(namespaceObj.Name, pvcName)
			Expect(pvName).ToNot(
				BeEmpty(),
				"while getting Spec.VolumeName from "+
					"PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("deleting above PVC")
			err = ops.PVCClient.WithNamespace(namespaceObj.Name).Delete(context.TODO(), pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("having the Provisioner delete the PV")
			status := ops.IsPVDeletedEventually(pvName)
			Expect(status).To(
				BeTrue(),
				"while waiting for the Provisioner to delete PV {%s}",
				pvName,
			)

			By("verifying PVC is deleted")
			status = ops.IsPVCDeletedEventually(pvcName, namespaceObj.Name)
			Expect(status).To(
				BeTrue(),
				"when checking status of deleted PVC {%s}",
				pvcName,
			)

			By("removing the xfs filesystem")
			//umount
			errs := physicalDisk.Unmount()
			Expect(errs).To(BeEmpty(), "when unmounting loop device")
			//wipefs
			err := physicalDisk.Wipefs()
			Expect(err).To(BeNil(), "when wiping filesystem from loop device")
		})
	})
})

var _ = Describe("TEST HOSTPATH EXT4 QUOTA LOCAL PV WITH UNSUPPORTED FILESYSTEM", func() {
	var (
		pvcObj      *corev1.PersistentVolumeClaim
		scObj       *storagev1.StorageClass
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity    = "5M"
		podName     = "busybox-hostpath"
		label       = "demo=hostpath-pod"
		createdPvc  *corev1.PersistentVolumeClaim

		pvcName       = "pvc-hp-ext4"
		scNamePrefix  = "sc-hp-ext4"
		scName        string
		podObj        *corev1.Pod
		labelselector = map[string]string{
			"demo": "hostpath-pod",
		}
	)

	When("StorageClass with valid ext4 quota parameters is created", func() {
		It("should create a StorageClass", func() {
			By("building a StorageClass")
			scObj, err = sc.NewStorageClass(
				sc.WithGenerateName(scNamePrefix),
				sc.WithLabels(map[string]string{
					"openebs.io/test-sc": "true",
				}),
				sc.WithLocalPV(),
				sc.WithHostpath(loopHostpathDir),
				sc.WithExt4Quota("20%", "50%"),
				sc.WithVolumeBindingMode("WaitForFirstConsumer"),
				sc.WithReclaimPolicy("Delete"),
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
			Expect(scName).NotTo(BeEmpty(), "SC name should not be empty")
		})
	})

	When("pvc with storageclass "+scName+" is created", func() {
		It("should create a pvc", func() {
			By("Writing minix filesystem into the loop device")
			//umount
			errs := physicalDisk.Unmount()
			Expect(errs).To(BeEmpty(), "when unmounting loop device")
			//wipefs
			err := physicalDisk.Wipefs()
			Expect(err).To(BeNil(), "when wiping filesystem from loop device")
			//mkfs
			err = physicalDisk.CreateFilesystem("minix")
			Expect(err).To(BeNil(), "when writing filesystem to loop device")
			//mount
			err = physicalDisk.PrjquotaMount(loopHostpathDir)
			Expect(err).To(
				BeNil(),
				"when mounting loop device filesystem to mountpoint",
			)

			By("building a PVC with StorageClass " + scName)
			Expect(scName).NotTo(BeEmpty(), "SC name should not be empty")
			pvcObj, err = BuildPersistentVolumeClaim(namespaceObj.Name, pvcName, scName, capacity, accessModes)
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
			podObj, err = BuildPod(namespaceObj.Name, podName, pvcName, labelselector)
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pod {%s} in namespace {%s}",
				podName,
				namespaceObj.Name,
			)

			By("creating above pod with busybox image")
			_, err := ops.PodClient.WithNamespace(namespaceObj.Name).
				Create(context.TODO(), podObj)
			Expect(err).To(
				BeNil(),
				"while creating pod {%s} in namespace {%s}",
				podName,
				namespaceObj.Name,
			)

			By("verifying pod status as pending")
			//Giving the check till timeout to try to see if the Status changes to Running
			podPhase := ops.CheckPodStatusEventually(namespaceObj.Name, podName, corev1.PodRunning)
			Expect(podPhase).To(Equal(corev1.PodPending), "while verifying pod pending status")

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
			By("verifying that the VolumeName is empty for the PVC")
			pvName := ops.GetPVNameFromPVCName(namespaceObj.Name, pvcName)
			Expect(pvName).To(
				BeEmpty(),
				"while getting Spec.VolumeName from PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("deleting above PVC")
			err = ops.PVCClient.WithNamespace(namespaceObj.Name).Delete(context.TODO(), pvcName, &metav1.DeleteOptions{})
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

			By("removing the minix filesystem")
			//umount
			errs := physicalDisk.Unmount()
			Expect(errs).To(BeEmpty(), "when unmounting loop device")
			//wipefs
			err := physicalDisk.Wipefs()
			Expect(err).To(BeNil(), "when wiping filesystem from loop device")
		})
	})
})

var _ = Describe("TEST HOSTPATH EXT4 QUOTA LOCAL PV WITH EXT4 FILESYSTEM", func() {
	var (
		pvcObj        *corev1.PersistentVolumeClaim
		scObj         *storagev1.StorageClass
		accessModes   = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacityInMib = "5"
		capacity      = capacityInMib + "M"
		podName       = "busybox-hostpath"
		label         = "demo=hostpath-pod"
		createdPod    *corev1.Pod
		pvcName       = "pvc-hp-ext4"
		scNamePrefix  = "sc-hp-ext4"
		scName        string
		podObj        *corev1.Pod
		labelselector = map[string]string{
			"demo": "hostpath-pod",
		}
	)

	When("StorageClass with valid ext4 quota parameters is created", func() {
		It("should create a StorageClass", func() {
			By("building a StorageClass")
			scObj, err = sc.NewStorageClass(
				sc.WithGenerateName(scNamePrefix),
				sc.WithLabels(map[string]string{
					"openebs.io/test-sc": "true",
				}),
				sc.WithLocalPV(),
				sc.WithHostpath(loopHostpathDir),
				sc.WithExt4Quota("", ""),
				sc.WithVolumeBindingMode("WaitForFirstConsumer"),
				sc.WithReclaimPolicy("Delete"),
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
			Expect(scName).NotTo(BeEmpty(), "SC name should not be empty")
		})
	})

	When("pvc with storageclass "+scName+" is created", func() {
		It("should create a pvc", func() {
			By("Writing ext4 filesystem into the loop device")
			//umount
			errs := physicalDisk.Unmount()
			Expect(errs).To(BeEmpty(), "when unmounting loop device")
			//wipefs
			err := physicalDisk.Wipefs()
			Expect(err).To(BeNil(), "when wiping filesystem from loop device")
			//mkfs
			err = physicalDisk.CreateFilesystem("ext4")
			Expect(err).To(BeNil(), "when writing filesystem to loop device")
			//mount
			err = physicalDisk.PrjquotaMount(loopHostpathDir)
			Expect(err).To(
				BeNil(),
				"when mounting loop device filesystem to mountpoint",
			)

			By("building a PVC with StorageClass " + scName)
			Expect(scName).NotTo(BeEmpty(), "SC name should not be empty")
			pvcObj, err = BuildPersistentVolumeClaim(namespaceObj.Name, pvcName, scName, capacity, accessModes)
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
			podObj, err = BuildPod(namespaceObj.Name, podName, pvcName, labelselector)
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
				WithCommand([]string{"/bin/sh", "-c", "dd if=/dev/zero of=/mnt/store1/test.txt bs=1M count=10 2>&1 || du -sm /mnt/store1 | cut -f -1 | tr -d '\n' 1>&2"}...)
			stdOut, stdErr, err := ops.ExecPod(option)
			fmt.Printf("When running command to test enforced quota. stdOut: {%s}, stderr: {%s}, error: {%v}", stdOut, stdErr, err)
			Expect(err).To(BeNil(), "while exec'ing into the pod and running command(s)")
			Expect(stdOut).NotTo(BeEmpty(), "trying to write till the quota limit should be allowed")
			Expect(stdErr).To(Equal(capacityInMib), "trying to write beyond the quota limit should not be allowed")
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
			By("getting the PV name from Bound PVC object spec")
			pvName := ops.GetPVNameFromPVCName(namespaceObj.Name, pvcName)
			Expect(pvName).ToNot(
				BeEmpty(),
				"while getting Spec.VolumeName from "+
					"PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("deleting above PVC")
			err = ops.PVCClient.WithNamespace(namespaceObj.Name).Delete(context.TODO(), pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("having the Provisioner delete the PV")
			status := ops.IsPVDeletedEventually(pvName)
			Expect(status).To(
				BeTrue(),
				"while waiting for the Provisioner to delete PV {%s}",
				pvName,
			)

			By("verifying PVC is deleted")
			status = ops.IsPVCDeletedEventually(pvcName, namespaceObj.Name)
			Expect(status).To(
				BeTrue(),
				"when checking status of deleted PVC {%s}",
				pvcName,
			)

			By("removing the ext4 filesystem")
			//umount
			errs := physicalDisk.Unmount()
			Expect(errs).To(BeEmpty(), "when unmounting loop device")
			//wipefs
			err := physicalDisk.Wipefs()
			Expect(err).To(BeNil(), "when wiping filesystem from loop device")
		})
	})
})
