package tests

import (
	ctx "context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mayav1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	deploy "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/apps/v1/deployment"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/container"
	pvc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/persistentvolumeclaim"
	pts "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/podtemplatespec"
	k8svolume "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/volume"
	sc "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/storage/v1/storageclass"
)

// createDeploymentWhichConsumesHostpath creates a single-replica Deployment whose Pod consumes hostpath PVC.
func createDeploymentWhichConsumesHostpath(namePrefix, namespace, pvcName string) (*appsv1.Deployment, error) {
	labelSelector := map[string]string{
		"app": namePrefix,
	}
	deployment, err := deploy.NewBuilder().
		WithGenerateName(namePrefix).
		WithNamespace(namespace).
		WithLabelsNew(labelSelector).
		WithSelectorMatchLabelsNew(labelSelector).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(labelSelector).
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
								{
									Name:      "demo-vol1",
									MountPath: "/mnt/store1",
								},
							},
						),
				).
				WithVolumeBuilders(
					k8svolume.NewBuilder().
						WithName("demo-vol1").
						WithPVCSource(pvcName),
				),
		).
		Build()
	if err != nil {
		return nil, err
	}

	return ops.DeployClient.WithNamespace(namespaceObj.Name).Create(context.TODO(), deployment)
}

// isLabelSelectorsEqual compares two arrays of label selector keys.
func isLabelSelectorsEqual(request, result []string) bool {
	if len(request) != len(result) {
		return false
	}

	ch := make(chan struct{}, 2)
	collectFrequency := func(labelKeys []string, freq *map[string]int) {
		for _, elem := range labelKeys {
			(*freq)[elem]++
		}

		ch <- struct{}{}
	}

	// Maps to hold the frequency of strings in the string slices.
	freqRequest := make(map[string]int)
	freqResult := make(map[string]int)

	go collectFrequency(request, &freqRequest)
	go collectFrequency(result, &freqResult)

	for i := 0; i < 2; i++ {
		select {
		case <-ch:
			continue
		}
	}

	// Compare frequencies
	for key, countRequest := range freqRequest {
		countResult, ok := freqResult[key]
		if !ok || countRequest != countResult {
			return false
		}
	}

	return true
}

var _ = Describe("VOLUME PROVISIONING/DE-PROVISIONING WITH ADDITIVE CAS-CONFIGS ON PVC AND SC", func() {
	var (
		pvcNamePrefix            = "pvc-additive-cas-config"
		scNamePrefix             = "sc-additive-cas-config"
		deployNamePrefix         = "busybox-additive-cas-config"
		deployment               *appsv1.Deployment
		scName                   string
		pvcName                  string
		pvcCapacity              = "2Gi"
		pvName                   string
		pvcNodeAffinityLabelKeys = []string{"kubernetes.io/os", "kubernetes.io/hostname"}
	)

	When("an application with a PVC which has cas-config which does not have conflicts with the cas-config on the"+
		" StorageClass, is created", func() {
		It("should provision the volume", func() {
			By("creating the StorageClass with cas-config", func() {
				storageClass, err := sc.NewStorageClass(
					sc.WithGenerateName("sc-additive-cas-config"),
					sc.WithLabels(map[string]string{
						"openebs.io/test-sc": "true",
					}),
					sc.WithLocalPV(),

					// This is the StorageClass config in question here.
					sc.WithHostpath(hostpathDir),

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
			By("creating the PVC with additive cas-config", func() {
				pvcCasConfig := []mayav1alpha1.Config{
					{
						Name: "NodeAffinityLabels", // This is the config that needs to not be for the same config key name.
						List: pvcNodeAffinityLabelKeys,
					},
				}
				pvcCasConfigStr, err := yaml.Marshal(pvcCasConfig)
				Expect(err).To(BeNil(), "while marshaling cas-config")
				pvc, err := pvc.NewBuilder().
					WithGenerateName(pvcNamePrefix).
					WithNamespace(namespaceObj.Name).
					WithAnnotations(map[string]string{
						string(mayav1alpha1.CASConfigKey): string(pvcCasConfigStr),
					}).
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
				deployment, err = createDeploymentWhichConsumesHostpath(deployNamePrefix, namespaceObj.Name,
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

			By("having the PVC non-conflicting cas-config set correctly", func() {
				nodeAffinityLabelKeys, err := ops.GetNodeAffinityLabelKeysFromPv(pvName)
				Expect(err).To(BeNil(), "while getting NodeAffinityLabels from PV '%s'", pvName)
				Expect(isLabelSelectorsEqual(pvcNodeAffinityLabelKeys, nodeAffinityLabelKeys)).To(
					BeTrue(),
					"while checking if PV %s had the NodeAffinityLabels requested on the PVC %s on namespace %s",
					pvName, pvcName, namespaceObj.Name,
				)
			})
		})
	})

	When("an application with a PVC which has cas-config which does not have conflicts with the cas-config on the"+
		" StorageClass, is deleted", func() {
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

var _ = Describe("VOLUME PROVISIONING/DE-PROVISIONING WITH CONFLICTING CAS-CONFIGS ON PVC AND SC", func() {
	var (
		pvcNamePrefix           = "pvc-conflicting-cas-config"
		scNamePrefix            = "sc-conflicting-cas-config"
		deployNamePrefix        = "busybox-conflicting-cas-config"
		deployment              *appsv1.Deployment
		scName                  string
		pvcName                 string
		pvcCapacity             = "2Gi"
		pvName                  string
		scNodeAffinityLabelKeys = []string{"kubernetes.io/hostname"}
	)

	When("an application with a PVC which has cas-config which has conflicts with the cas-config on the"+
		" StorageClass, is created", func() {
		It("should provision the volume", func() {
			By("creating the StorageClass with cas-config", func() {
				storageClass, err := sc.NewStorageClass(
					sc.WithGenerateName("sc-conflicting-cas-config"),
					sc.WithLabels(map[string]string{
						"openebs.io/test-sc": "true",
					}),
					sc.WithLocalPV(),
					sc.WithHostpath(hostpathDir),
					// This is the config in question.
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
			By("creating the PVC with the same cas-config key, but a different value", func() {
				pvcNodeAffinityLabelKeys := []string{"kubernetes.io/os", "kubernetes.io/arch"}
				pvcCasConfig := []mayav1alpha1.Config{
					{
						Name: "NodeAffinityLabels", // This is the config that needs to not be for the same config key name.
						List: pvcNodeAffinityLabelKeys,
					},
				}
				pvcCasConfigStr, err := yaml.Marshal(pvcCasConfig)
				Expect(err).To(BeNil(), "while marshaling cas-config")
				pvc, err := pvc.NewBuilder().
					WithGenerateName(pvcNamePrefix).
					WithNamespace(namespaceObj.Name).
					WithAnnotations(map[string]string{
						string(mayav1alpha1.CASConfigKey): string(pvcCasConfigStr),
					}).
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
				deployment, err = createDeploymentWhichConsumesHostpath(deployNamePrefix, namespaceObj.Name,
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

			By("having the SC cas-config set correctly", func() {
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

	When("an application with a PVC which has cas-config which has conflicts with the cas-config on the"+
		" StorageClass, is deleted", func() {
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
