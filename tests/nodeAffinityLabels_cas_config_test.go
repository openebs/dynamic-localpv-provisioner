package tests

import (
	ctx "context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pvc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/persistentvolumeclaim"
	sc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/storage/v1/storageclass"
)

var _ = Describe("VOLUME PROVISIONING/DE-PROVISIONING WITH 'NodeAffinityLabels' CAS-CONFIG ON STORAGECLASS", func() {
	var (
		pvcNamePrefix           = "pvc-nod-aff-lab"
		scNamePrefix            = "sc-nod-aff-lab"
		deployNamePrefix        = "busybox-nod-aff-lab"
		deployment              *appsv1.Deployment
		scName                  string
		pvcName                 string
		pvcCapacity             = "2Gi"
		pvName                  string
		scNodeAffinityLabelKeys = []string{"kubernetes.io/hostname", "kubernetes.io/os", "kubernetes.io/arch"}
	)

	When("an application with a PVC which has custom NodeAffinityLabels cas-config on the StorageClass, is created", func() {
		It("should provision the volume", func() {
			By("creating the StorageClass with custom NodeAffinityLabels", func() {
				storageClass, err := sc.NewStorageClass(
					sc.WithGenerateName(scNamePrefix),
					sc.WithLabels(map[string]string{
						"openebs.io/test-sc": "true",
					}),
					sc.WithLocalPV(),
					sc.WithHostpath(hostpathDir),
					sc.WithNodeAffinityLabels(scNodeAffinityLabelKeys),
					sc.WithVolumeBindingMode("WaitForFirstConsumer"),
					sc.WithReclaimPolicy("Delete"),
				)
				Expect(err).To(
					BeNil(),
					"while building StorageClass with name prefix {%s}",
					scNamePrefix,
				)
				storageClass, err = ops.SCClient.Create(ctx.TODO(), storageClass)
				Expect(err).To(
					BeNil(),
					"while creating StorageClass with name prefix %s",
					scNamePrefix,
				)
				scName = storageClass.Name
			})
			By("creating the PVC with the StorageClass "+scName, func() {
				pvc, err := pvc.NewBuilder().
					WithGenerateName(pvcNamePrefix).
					WithNamespace(namespaceObj.Name).
					WithStorageClass(scName).
					WithAccessModes([]corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}).
					WithCapacity(pvcCapacity).
					Build()
				Expect(err).To(
					BeNil(),
					"while building PVC with name prefix %s in namespace %s",
					pvcNamePrefix,
					namespaceObj.Name,
				)
				pvc, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(ctx.TODO(), pvc)
				Expect(err).To(
					BeNil(),
					"while creating PVC with name prefix %s in namespace %s",
					pvcNamePrefix,
					namespaceObj.Name,
				)
				pvcName = pvc.Name
			})
			By("creating a bound PV", func() {
				deployment, err = ops.createDeploymentWhichConsumesHostpath(deployNamePrefix, namespaceObj.Name,
					pvcName)
				Expect(err).To(
					BeNil(),
					"while creating Deployment with name %s in namespace %s",
					deployment.Name,
					namespaceObj.Name,
				)
				Expect(ops.IsPVCBoundEventually(namespaceObj.Name, pvcName)).To(
					BeTrue(),
					"while checking if the PVC %s in namespace %s was bound to a PV",
					pvcName, namespaceObj.Name,
				)
				Expect(ops.GetPodRunningCountEventually(namespaceObj.Name, "app="+deployNamePrefix, 1)).To(BeNumerically("==",
					1))
				Expect(Eventually(func() string {
					pvName = ops.GetPVNameFromPVCName(namespaceObj.Name, pvcName)
					return pvName
				}).
					WithTimeout(time.Minute).
					WithPolling(time.Second).
					WithContext(ctx.TODO()).
					Should(Not(BeEmpty()))).To(BeTrue())
			})

			By("having the SC NodeAffinityLabels cas-config set correctly", func() {
				nodeAffinityLabelKeys, err := ops.GetNodeAffinityLabelKeysFromPv(pvName)
				Expect(err).To(BeNil(), "while getting NodeAffinityLabels from PV '%s'", pvName)
				Expect(isLabelSelectorsEqual(scNodeAffinityLabelKeys, nodeAffinityLabelKeys)).To(
					BeTrue(),
					"while checking if PV %s had the NodeAffinityLabels requested on the SC %s",
					scName,
				)
			})
		})
	})

	When("an application with a PV which has custom NodeAffinityLabels is deleted", func() {
		It("should de-provision the volume", func() {
			By("deleting the PV", func() {
				podList, err := ops.PodClient.List(ctx.TODO(), metav1.ListOptions{LabelSelector: "app=" + deployNamePrefix})
				Expect(err).To(BeNil(), "while listing Pods for busybox application deployment")
				Expect(len(podList.Items)).To(BeNumerically("==", 1))
				pod := &podList.Items[0]
				err = ops.DeployClient.WithNamespace(namespaceObj.Name).Delete(ctx.TODO(), deployment.Name, &metav1.DeleteOptions{})
				Expect(err).To(BeNil(), "while deleting busybox application deployment")
				Expect(ops.IsPodDeletedEventually(pod.Namespace, pod.Name)).To(
					BeTrue(),
					"while checking to see if the Pod %s in namespace %s for the busybox deployment is deleted",
					pod.Name, pod.Namespace,
				)

				ops.DeletePersistentVolumeClaim(pvcName, namespaceObj.Name)
				Expect(ops.IsPVCDeletedEventually(pvcName, namespaceObj.Name)).To(
					BeTrue(),
					"while checking if PVC %s in namespace %s is deleted",
					pvcName, namespaceObj.Namespace,
				)
				Expect(ops.IsPVDeletedEventually(pvName)).To(
					BeTrue(),
					"when checking to see if the underlying PV %s is deleted",
					pvName,
				)
			})
		})
	})
})
