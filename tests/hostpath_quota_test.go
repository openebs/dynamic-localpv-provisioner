package tests

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

var _ = Describe("TEST HOSTPATH LOCAL PV", func() {
	var (
		pvcObj         *corev1.PersistentVolumeClaim
		scObj          *storagev1.StorageClass
		accessModes    = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity       = "2M"
		xfsHostpathDir = "/var/openebs/integration-test/xfs/"
		podName        = "busybox-hostpath"
		label          = "demo=hostpath-pod"
		pvcName        = "pvc-hp"
		scNamePrefix   = "sc-hp-xfs"
		scName         string
		podObj         *corev1.Pod
		labelselector  = map[string]string{
			"demo": "hostpath-pod",
		}
	)

	physicalDisk := disk.NewDisk(DiskImageSize)
	err := physicalDisk.AttachDisk()
	Expect(err).To(
		BeNil(),
		"while creating loopback device with disk {%+v}",
		physicalDisk,
	)

	// Make xfs fs on the created loopback device
	err = physicalDisk.CreateFileSystem("xfs")
	Expect(err).To(
		BeNil(),
		"while formatting the disk {%+v} with xfs fs",
		physicalDisk,
	)

	err = disk.MkdirAll(xfsHostpathDir)
	Expect(err).To(
		BeNil(),
		"while making a new directory {%s}",
		xfsHostpathDir,
	)

	// Mount the xfs formatted loopback device
	err = physicalDisk.Mount(xfsHostpathDir)
	Expect(err).To(
		BeNil(),
		"while mounting the disk with pquota option {%+v}",
		physicalDisk,
	)

	When("hostpath is not having xfs filesystem", func() {
		When("pod consuming pvc with a valid quota storageclass is created", func() {
			It("should not be up and running", func() {
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

				By("building a PVC with StorageClass " + scName)
				pvcObj, err = BuildPersistentVolumeClaim(pvcName, scName, capacity, accessModes)
				Expect(err).ShouldNot(
					HaveOccurred(),
					"while building PVC {%s} in namespace {%s}",
					pvcName,
					namespaceObj.Name,
				)

				By("creating above PVC with StorageClass " + scName)
				createdPvc, err := ops.PVCClient.WithNamespace(namespaceObj.Name).Create(context.TODO(), pvcObj)
				Expect(err).To(
					BeNil(),
					"while creating PVC {%s} in namespace {%s}",
					pvcName,
					namespaceObj.Name,
				)

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

				By("verifying pod status as not running")
				podPhase := ops.GetPodStatusEventually(createdPod)
				Expect(podPhase).ToNot(Equal(corev1.PodRunning), "while verifying pod running status")

				By("verifying the above created pvc phase")
				Expect(createdPvc.Status.Phase).To(Equal(corev1.ClaimPending), "while verifying the pvc state")
			})
		})

		When("Pod consuming pvc with a valid quota storageclass is deleted along with pvc and storageclass", func() {
			It("should delete pod, pvc and storageclass", func() {
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

				By("deleting the PVC with StorageClass " + scName)
				pvName := ops.GetPVNameFromPVCName(pvcName)
				Expect(pvName).To(
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

				By("deleting the storageClass " + scName)
				err = ops.SCClient.Delete(context.TODO(), scName, &metav1.DeleteOptions{})
				Expect(err).To(
					BeNil(),
					"while deleting storageclass {%s}",
					scName,
				)
			})
		})
	})

	When("hostpath is having xfs filesystem", func() {
		When("pod consuming pvc with a valid quota storageclass is created", func() {
			It("should be up and running", func() {
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

				By("building a pod with busybox image")
				podObj, err = BuildPod(podName, pvcName, labelselector)
				Expect(err).ShouldNot(
					HaveOccurred(),
					"while building pod {%s} in namespace {%s}",
					podName,
					namespaceObj.Name,
				)

				By("creating above pod with busybox image")
				_, err = ops.PodClient.WithNamespace(namespaceObj.Name).
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

		When("pod consuming pvc with a valid quota storageclass is deleted along with pvc and storageclass", func() {
			It("should delete pod, pvc and storageclass", func() {
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

				By("deleting the storageClass " + scName)
				err = ops.SCClient.Delete(context.TODO(), scName, &metav1.DeleteOptions{})
				Expect(err).To(
					BeNil(),
					"while deleting storageclass {%s}",
					scName,
				)
			})
		})
	})

	// Unmount the disk
	err = physicalDisk.Unmount()
	Expect(err).To(
		BeNil(),
		"while unmounting the disk {%+v}",
		physicalDisk,
	)

	// Detach and delete the disk
	err = physicalDisk.DetachAndDeleteDisk()
	Expect(err).To(
		BeNil(),
		"while detaching and deleting the disk {%+v}",
		physicalDisk,
	)
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
						"sleep",
						"3600",
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
