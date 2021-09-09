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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	deploy "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/apps/v1/deployment"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/container"
	pvc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/persistentvolumeclaim"
	pts "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/podtemplatespec"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/volume"
	sc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/storage/v1/storageclass"
)

var _ = Describe("TEST HOSTPATH LOCAL PV", func() {
	var (
		pvcObj        *corev1.PersistentVolumeClaim
		scObj         *storagev1.StorageClass
		accessModes   = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity      = "2Gi"
		deployName    = "busybox-hostpath"
		label         = "demo=hostpath-deployment"
		pvcName       = "pvc-hp"
		scNamePrefix  = "sc-hp"
		scName        string
		deployObj     *appsv1.Deployment
		labelselector = map[string]string{
			"demo": "hostpath-deployment",
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
				sc.WithHostpath("/var/openebs/integration-test"),
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
				"while building PVC {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("creating above PVC")
			_, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(context.TODO(), pvcObj)
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
											Name:      "demo-vol1",
											MountPath: "/mnt/store1",
										},
									},
								),
						).
						WithVolumeBuilders(
							volume.NewBuilder().
								WithName("demo-vol1").
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
			By("getting the PV name from Bound PVC object spec")
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
		})
	})
})
