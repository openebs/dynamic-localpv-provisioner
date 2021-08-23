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

package tests

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	localpv_app "github.com/openebs/dynamic-localpv-provisioner/cmd/provisioner-localpv/app"
	deploy "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/apps/v1/deployment"
	container "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/container"
	pvc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/persistentvolumeclaim"
	pts "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/podtemplatespec"
	volume "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/volume"
	sc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/storage/v1/storageclass"
	blockdeviceclaim "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TEST HOSTDEVICE LOCAL PV", func() {
	var (
		scObj         *storagev1.StorageClass
		pvcObj        *corev1.PersistentVolumeClaim
		accessModes   = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity      = "2Gi"
		deployName    = "busybox-device"
		label         = "demo=hostdevice-deployment"
		scNamePrefix  = "sc-hd"
		scName        string
		bdcName       string
		pvcName       = "pvc-hd"
		deployObj     *appsv1.Deployment
		labelselector = map[string]string{
			"demo": "hostdevice-deployment",
		}
	)

	When("a StorageClass is created", func() {
		It("should create a StorageClass", func() {
			By("building a StorageClass")
			scObj, err = sc.NewStorageClass(
				sc.WithGenerateName(scNamePrefix),
				sc.WithLabels(map[string]string{
					"openebs.io/test-sc": "true",
				}),
				sc.WithLocalPV(),
				sc.WithDevice(),
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
		})
	})

	When("PVC with StorageClass "+scName+" is created", func() {
		It("should create a PVC ", func() {
			By("building a PVC")
			pvcObj, err = pvc.NewBuilder().
				WithName(pvcName).
				WithNamespace(namespaceObj.Name).
				WithStorageClass(scName).
				WithAccessModes(accessModes).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("creating above PVC")
			pvcObj, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(context.TODO(), pvcObj)
			Expect(err).To(
				BeNil(),
				"while creating PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)
		})
	})

	When("deployment with busybox image is created", func() {
		It("should create a deployment and a running pod", func() {

			By("building a deployment")
			deployObj, err = deploy.NewBuilder().
				WithName(deployName).
				WithNamespace(namespaceObj.Name).
				WithLabelsNew(labelselector).
				WithSelectorMatchLabelsNew(labelselector).
				WithPodTemplateSpecBuilder(
					pts.NewBuilder().
						WithLabelsNew(labelselector).
						WithContainerBuildersNew(
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
										corev1.VolumeMount{
											Name:      "demo-vol2",
											MountPath: "/mnt/store1",
										},
									},
								),
						).
						WithVolumeBuildersNew(
							volume.NewBuilder().
								WithName("demo-vol2").
								WithPVCSource(pvcName),
						),
				).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building delpoyment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("creating above deployment")
			_, err = ops.DeployClient.WithNamespace(namespaceObj.Name).
				Create(context.TODO(), deployObj)
			Expect(err).To(
				BeNil(),
				"while creating deployment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 1")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 1)
			Expect(podCount).To(Equal(1), "while verifying pod count")

		})
	})
	When("remove finalizer", func() {
		It("finalizer should come back after provisioner restart", func() {
			bdcName = "bdc-pvc-" + string(pvcObj.GetUID())
			bdcObj, err := ops.BDCClient.WithNamespace(openebsNamespace).Get(context.TODO(), bdcName,
				metav1.GetOptions{})
			Expect(err).To(BeNil())

			_, err = blockdeviceclaim.BuilderForAPIObject(bdcObj).WithConfigPath(ops.KubeConfigPath).
				BDC.RemoveFinalizer(localpv_app.LocalPVFinalizer)
			Expect(err).To(BeNil())

			podList, err := ops.PodClient.
				WithNamespace(openebsNamespace).
				List(context.TODO(), metav1.ListOptions{LabelSelector: LocalPVProvisionerLabelSelector})
			Expect(err).To(BeNil())
			err = ops.PodClient.WithNamespace(openebsNamespace).Delete(context.TODO(), podList.Items[0].Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())

			Expect(ops.IsFinalizerExistsOnBDC(bdcName, localpv_app.LocalPVFinalizer)).To(BeTrue())
		})
	})
	When("deployment is deleted", func() {
		It("should not have any deployment or running pod", func() {

			By("deleting above deployment")
			err = ops.DeployClient.WithNamespace(namespaceObj.Name).Delete(context.TODO(), deployName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting deployment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 0)
			Expect(podCount).To(Equal(0), "while verifying pod count")

		})
	})

	When("PVC with StorageClass "+scName+" is deleted ", func() {
		It("should delete the pvc", func() {
			By("getting the PV name")
			pvName := ops.GetPVNameFromPVCName(pvcName)

			By("deleting above pvc")
			err = ops.PVCClient.Delete(context.TODO(), pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting PVC {%s} in namespace {%s}",
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

			By("verifying BDC is deleted")
			status = ops.IsBDCDeletedEventually(bdcName, openebsNamespace)
			Expect(status).To(
				BeTrue(),
				"when checking status of BDC {%s}, which should have been deleted",
				bdcName,
			)
		})
	})
})

var _ = Describe("TEST HOSTDEVICE LOCAL PV WITH VOLUMEMODE AS BLOCK", func() {
	var (
		scObj         *storagev1.StorageClass
		pvcObj        *corev1.PersistentVolumeClaim
		accessModes   = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity      = "2Gi"
		deployName    = "busybox-device"
		label         = "demo=hostdevice-deployment"
		scNamePrefix  = "sc-hd-block"
		scName        string
		pvcName       = "pvc-hd-block"
		deployObj     *appsv1.Deployment
		labelselector = map[string]string{
			"demo": "hostdevice-deployment",
		}
	)

	When("a StorageClass is created", func() {
		It("should create a StorageClass", func() {
			By("building a StorageClass")
			scObj, err = sc.NewStorageClass(
				sc.WithGenerateName(scNamePrefix),
				sc.WithLabels(map[string]string{
					"openebs.io/test-sc": "true",
				}),
				sc.WithLocalPV(),
				sc.WithDevice(),
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
		})
	})

	When("PVC with StorageClass "+scName+", and volumeMode as Block, is created", func() {
		It("should create a PVC ", func() {
			var (
				blockVolumeMode = corev1.PersistentVolumeBlock
			)

			By("building a PVC")
			pvcObj, err = pvc.NewBuilder().
				WithName(pvcName).
				WithNamespace(namespaceObj.Name).
				WithStorageClass(scName).
				WithAccessModes(accessModes).
				WithVolumeMode(blockVolumeMode).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("creating above PVC")
			pvcObj, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(context.TODO(), pvcObj)
			Expect(err).To(
				BeNil(),
				"while creating PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)
		})
	})

	When("deployment with busybox image is created", func() {
		It("should create a deployment and a running pod", func() {

			By("building a deployment")
			deployObj, err = deploy.NewBuilder().
				WithName(deployName).
				WithNamespace(namespaceObj.Name).
				WithLabelsNew(labelselector).
				WithSelectorMatchLabelsNew(labelselector).
				WithPodTemplateSpecBuilder(
					pts.NewBuilder().
						WithLabelsNew(labelselector).
						WithContainerBuildersNew(
							container.NewBuilder().
								WithName("busybox").
								WithImage("busybox").
								WithCommandNew(
									[]string{
										"sleep",
										"3600",
									},
								).
								WithVolumeDevices(
									[]corev1.VolumeDevice{
										corev1.VolumeDevice{
											Name:       "demo-block-vol1",
											DevicePath: "/dev/sdc",
										},
									},
								),
						).
						WithVolumeBuildersNew(
							volume.NewBuilder().
								WithName("demo-block-vol1").
								WithPVCSource(pvcName),
						),
				).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building delpoyment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("creating above deployment")
			_, err = ops.DeployClient.WithNamespace(namespaceObj.Name).
				Create(context.TODO(), deployObj)
			Expect(err).To(
				BeNil(),
				"while creating deployment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 1")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 1)
			Expect(podCount).To(Equal(1), "while verifying pod count")

		})
	})

	When("remove finalizer", func() {
		It("finalizer should come back after provisioner restart", func() {
			bdcName := "bdc-pvc-" + string(pvcObj.GetUID())
			bdcObj, err := ops.BDCClient.WithNamespace(openebsNamespace).Get(context.TODO(), bdcName,
				metav1.GetOptions{})
			Expect(err).To(BeNil())

			_, err = blockdeviceclaim.BuilderForAPIObject(bdcObj).WithConfigPath(ops.KubeConfigPath).
				BDC.RemoveFinalizer(localpv_app.LocalPVFinalizer)
			Expect(err).To(BeNil())

			podList, err := ops.PodClient.
				WithNamespace(openebsNamespace).
				List(context.TODO(), metav1.ListOptions{LabelSelector: LocalPVProvisionerLabelSelector})
			Expect(err).To(BeNil())
			err = ops.PodClient.WithNamespace(openebsNamespace).Delete(context.TODO(), podList.Items[0].Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())

			Expect(ops.IsFinalizerExistsOnBDC(bdcName, localpv_app.LocalPVFinalizer)).To(BeTrue())
		})
	})
	When("deployment is deleted", func() {
		It("should not have any deployment or running pod", func() {

			By("deleting above deployment")
			err = ops.DeployClient.WithNamespace(namespaceObj.Name).Delete(context.TODO(), deployName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting deployment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 0)
			Expect(podCount).To(Equal(0), "while verifying pod count")

		})
	})

	When("PVC with StorageClass "+scName+" is deleted ", func() {
		It("should delete the PVC", func() {
			By("getting the PV name and the BDC name")
			bdcName := "bdc-pvc-" + string(pvcObj.GetUID())
			pvName := ops.GetPVNameFromPVCName(pvcName)

			By("deleting above pvc")
			err = ops.PVCClient.Delete(context.TODO(), pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting PVC {%s} in namespace {%s}",
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

			By("verifying BDC is deleted")
			status = ops.IsBDCDeletedEventually(bdcName, openebsNamespace)
			Expect(status).To(
				BeTrue(),
				"when checking status of BDC {%s}, which should have been deleted",
				bdcName,
			)

		})
	})

})
var _ = Describe("[-ve] TEST HOSTDEVICE LOCAL PV", func() {
	var (
		scObj         *storagev1.StorageClass
		pvcObj        *corev1.PersistentVolumeClaim
		accessModes   = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity      = "2Gi"
		deployName    = "busybox-device"
		label         = "demo=hostdevice-deployment"
		scNamePrefix  = "sc-hd"
		scName        string
		pvcName       = "pvc-hd"
		deployObj     *appsv1.Deployment
		labelselector = map[string]string{
			"demo": "hostdevice-deployment",
		}
		//bdcTimeoutDuration    = 60
		existingPVCObj        *corev1.PersistentVolumeClaim
		existingDeployName    = "existing-busybox-device"
		existinglabel         = "demo=existing-hostdevice-deployment"
		existingPVCName       = "existing-pvc-hd"
		existingDeployObj     *appsv1.Deployment
		existingLabelselector = map[string]string{
			"demo": "existing-hostdevice-deployment",
		}
	)

	When("a StorageClass is created", func() {
		It("should create a StorageClass", func() {
			By("building a StorageClass")
			scObj, err = sc.NewStorageClass(
				sc.WithGenerateName(scNamePrefix),
				sc.WithLabels(map[string]string{
					"openebs.io/test-sc": "true",
				}),
				sc.WithLocalPV(),
				sc.WithDevice(),
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
		})
	})

	When("existing PVC with StorageClass "+scName+" is created", func() {
		It("should create a PVC", func() {

			By("building a pvc")
			existingPVCObj, err = pvc.NewBuilder().
				WithName(existingPVCName).
				WithNamespace(namespaceObj.Name).
				WithStorageClass(scName).
				WithAccessModes(accessModes).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building PVC {%s} in namespace {%s}",
				existingPVCName,
				namespaceObj.Name,
			)

			By("creating above PVC")
			existingPVCObj, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(context.TODO(), existingPVCObj)
			Expect(err).To(
				BeNil(),
				"while creating PVC {%s} in namespace {%s}",
				existingPVCName,
				namespaceObj.Name,
			)
		})
	})

	When("existing deployment with busybox image is created", func() {
		It("should create a deployment and a running pod", func() {

			By("building a deployment")
			existingDeployObj, err = deploy.NewBuilder().
				WithName(existingDeployName).
				WithNamespace(namespaceObj.Name).
				WithLabelsNew(existingLabelselector).
				WithSelectorMatchLabelsNew(existingLabelselector).
				WithPodTemplateSpecBuilder(
					pts.NewBuilder().
						WithLabelsNew(existingLabelselector).
						WithContainerBuildersNew(
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
										corev1.VolumeMount{
											Name:      "demo-vol3",
											MountPath: "/mnt/store1",
										},
									},
								),
						).
						WithVolumeBuildersNew(
							volume.NewBuilder().
								WithName("demo-vol3").
								WithPVCSource(existingPVCName),
						),
				).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building deployment {%s} in namespace {%s}",
				existingDeployName,
				namespaceObj.Name,
			)

			By("creating above deployment")
			_, err = ops.DeployClient.WithNamespace(namespaceObj.Name).
				Create(context.TODO(), existingDeployObj)
			Expect(err).To(
				BeNil(),
				"while creating deployment {%s} in namespace {%s}",
				existingDeployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 1")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, existinglabel, 1)
			Expect(podCount).To(Equal(1), "while verifying pod count")

		})
	})

	When("another PVC with StorageClass "+scName+" is created", func() {
		It("should create a PVC ", func() {

			By("building a PVC")
			pvcObj, err = pvc.NewBuilder().
				WithName(pvcName).
				WithNamespace(namespaceObj.Name).
				WithStorageClass(scName).
				WithAccessModes(accessModes).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("creating above pvc")
			pvcObj, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(context.TODO(), pvcObj)
			Expect(err).To(
				BeNil(),
				"while creating PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)
		})
	})

	When("another deployment with busybox image and above PVC is created", func() {
		It("should not create a deployment and a running pod", func() {

			By("building a deployment")
			deployObj, err = deploy.NewBuilder().
				WithName(deployName).
				WithNamespace(namespaceObj.Name).
				WithLabelsNew(labelselector).
				WithSelectorMatchLabelsNew(labelselector).
				WithPodTemplateSpecBuilder(
					pts.NewBuilder().
						WithLabelsNew(labelselector).
						WithContainerBuildersNew(
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
										corev1.VolumeMount{
											Name:      "demo-vol2",
											MountPath: "/mnt/store1",
										},
									},
								),
						).
						WithVolumeBuildersNew(
							volume.NewBuilder().
								WithName("demo-vol2").
								WithPVCSource(pvcName),
						),
				).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building delpoyment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("creating above deployment")
			_, err = ops.DeployClient.WithNamespace(namespaceObj.Name).
				Create(context.TODO(), deployObj)
			Expect(err).To(
				BeNil(),
				"while creating deployment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			/*
				By("checking if BDC gets deleted")
				staleBDCName := "bdc-pvc-" + string(pvcObj.GetUID())
				exitStatus := ops.GetBDCStatusAfterAge(staleBDCName, openebsNamespace, time.Duration(bdcTimeoutDuration+1)*time.Second)
				Expect(exitStatus).To(
					Equal(deleted),
					"while checking if the stale BDC {%s} got deleted",
					staleBDCName,
				)
			*/

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 0)
			Expect(podCount).To(Equal(0), "while verifying pod count")
		})
	})

	When("above deployment is deleted", func() {
		It("should not have any deployment or running pod", func() {

			By("deleting above deployment")
			err = ops.DeployClient.WithNamespace(namespaceObj.Name).Delete(context.TODO(), deployName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting deployment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 0)
			Expect(podCount).To(Equal(0), "while verifying pod count")

		})
	})

	When("above PVC with StorageClass "+scName+" is deleted ", func() {
		It("should delete the PVC", func() {

			By("deleting above PVC")
			err = ops.PVCClient.Delete(context.TODO(), pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting PVC {%s} in namespace {%s}",
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

	When("existing deployment is deleted", func() {
		It("should not have any deployment or running pod", func() {

			By("deleting above deployment")
			err = ops.DeployClient.WithNamespace(namespaceObj.Name).
				Delete(context.TODO(), existingDeployName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting deployment {%s} in namespace {%s}",
				existingDeployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, existinglabel, 1)
			Expect(podCount).To(Equal(1), "while verifying pod count")

		})
	})

	When("existing PVC with storageclass "+scName+" is deleted ", func() {
		It("should delete the PVC", func() {
			By("getting the PV name and the BDC name")
			bdcName := "bdc-pvc-" + string(existingPVCObj.GetUID())
			pvName := ops.GetPVNameFromPVCName(existingPVCName)

			By("deleting above PVC")
			err = ops.PVCClient.Delete(context.TODO(), existingPVCName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting PVC {%s} in namespace {%s}",
				existingPVCName,
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
			status = ops.IsPVCDeletedEventually(existingPVCName, namespaceObj.Name)
			Expect(status).To(
				BeTrue(),
				"when checking status of deleted PVC {%s}",
				existingPVCName,
			)

			By("verifying BDC is deleted")
			status = ops.IsBDCDeletedEventually(bdcName, openebsNamespace)
			Expect(status).To(
				BeTrue(),
				"when checking status of BDC {%s}, which should have been deleted",
				bdcName,
			)

		})
	})
})

var _ = Describe("[-ve] TEST HOSTDEVICE LOCAL PV WITH VOLUMEMODE AS BLOCK ", func() {
	var (
		scObj *storagev1.StorageClass
		//pvcObj          *corev1.PersistentVolumeClaim
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity    = "2Gi"
		//deployName      = "busybox-device"
		//label           = "demo=hostdevice-deployment"
		//pvcName         = "pvc-hd-block"
		//deployObj       *appsv1.Deployment
		blockVolumeMode = corev1.PersistentVolumeBlock
		//labelselector   = map[string]string{
		//	"demo": "hostdevice-deployment",
		//}
		scNamePrefix          = "sc-hd-block"
		scName                string
		existingPVCObj        *corev1.PersistentVolumeClaim
		existingDeployName    = "existing-busybox-device"
		existinglabel         = "demo=existing-hostdevice-deployment"
		existingPVCName       = "existing-pvc-hd-block"
		existingDeployObj     *appsv1.Deployment
		existingLabelselector = map[string]string{
			"demo": "existing-hostdevice-deployment",
		}
	)

	When("a StorageClass is created", func() {
		It("should create a StorageClass", func() {
			By("building a StorageClass")
			scObj, err = sc.NewStorageClass(
				sc.WithGenerateName(scNamePrefix),
				sc.WithLabels(map[string]string{
					"openebs.io/test-sc": "true",
				}),
				sc.WithLocalPV(),
				sc.WithDevice(),
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
		})
	})

	When("existing PVC with StorageClass "+scName+" is created", func() {
		It("should create a PVC", func() {

			By("building a PVC")
			existingPVCObj, err = pvc.NewBuilder().
				WithName(existingPVCName).
				WithNamespace(namespaceObj.Name).
				WithStorageClass(scName).
				WithAccessModes(accessModes).
				WithVolumeMode(blockVolumeMode).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pvc {%s} in namespace {%s}",
				existingPVCName,
				namespaceObj.Name,
			)

			By("creating above PVC")
			existingPVCObj, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(context.TODO(), existingPVCObj)
			Expect(err).To(
				BeNil(),
				"while creating PVC {%s} in namespace {%s}",
				existingPVCName,
				namespaceObj.Name,
			)
		})
	})
	When("existing deployment with busybox image is created", func() {
		It("should create a deployment but should be unable to get a running pod, with PVC volumeMode set to Block,but added as volumeMount in Deployment", func() {

			By("building a deployment, with volume Mount for a Block volumeMode PVC")
			existingDeployObj, err = deploy.NewBuilder().
				WithName(existingDeployName).
				WithNamespace(namespaceObj.Name).
				WithLabelsNew(existingLabelselector).
				WithSelectorMatchLabelsNew(existingLabelselector).
				WithPodTemplateSpecBuilder(
					pts.NewBuilder().
						WithLabelsNew(existingLabelselector).
						WithContainerBuildersNew(
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
										corev1.VolumeMount{
											Name:      "demo-block-vol2",
											MountPath: "/mnt/store1",
										},
									},
								),
						).
						WithVolumeBuildersNew(
							volume.NewBuilder().
								WithName("demo-block-vol2").
								WithPVCSource(existingPVCName),
						),
				).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building deployment {%s} in namespace {%s}",
				existingDeployName,
				namespaceObj.Name,
			)

			By("creating above deployment")
			_, err = ops.DeployClient.WithNamespace(namespaceObj.Name).
				Create(context.TODO(), existingDeployObj)
			Expect(err).To(
				BeNil(),
				"while creating deployment {%s} in namespace {%s}",
				existingDeployName,
				namespaceObj.Name,
			)

			By("verifying PVC status as bound")
			status := ops.IsPVCBoundEventually(existingPVCName)
			Expect(status).To(Equal(true), "while checking status equal to bound")

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, existinglabel, 1)
			Expect(podCount).To(Equal(0), "while verifying pod count")

		})
	})
	When("above deployment is deleted", func() {
		It("should not have any deployment or running pod", func() {

			By("deleting above deployment")
			err = ops.DeployClient.WithNamespace(namespaceObj.Name).Delete(context.TODO(), existingDeployName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting deployment {%s} in namespace {%s}",
				existingDeployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, existinglabel, 0)
			Expect(podCount).To(Equal(0), "while verifying pod count")

		})
	})

	When("existing PVC with storageclass "+scName+" is deleted ", func() {
		It("should delete the PVC", func() {

			By("deleting above PVC")
			err = ops.PVCClient.Delete(context.TODO(), existingPVCName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting PVC {%s} in namespace {%s}",
				existingPVCName,
				namespaceObj.Name,
			)

			By("verifying PVC is deleted")
			status := ops.IsPVCDeletedEventually(existingPVCName, namespaceObj.Name)
			Expect(status).To(
				BeTrue(),
				"when checking status of deleted PVC {%s}",
				existingPVCName,
			)

		})
	})
})
