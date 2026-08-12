package upgrade

import (
	"context"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	deploy "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/apps/v1/deployment"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/container"
	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/pod"
	pts "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/podtemplatespec"
	k8svolume "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/volume"
	"github.com/openebs/dynamic-localpv-provisioner/tests"
)

const (
	// appMountPath is where the test applications mount their volume.
	appMountPath = "/mnt/store1"
	// appContainerName is the only container of the test applications.
	appContainerName = "busybox"
	// appVolumeName names the volume in the application's Pod spec.
	appVolumeName = "demo-vol1"
)

// runningProvisionerPods returns the localpv-provisioner Pods which are running.
func runningProvisionerPods() []corev1.Pod {
	return ops.GetPodList(
		openebsNamespace,
		provisionerLabelSelector,
		pod.PredicateList{pod.IsRunning()},
	).ToAPIList().Items
}

// provisionerImages returns the container images of the running
// localpv-provisioner Pods.
func provisionerImages() []string {
	images := []string{}
	for _, provisionerPod := range runningProvisionerPods() {
		for _, provisionerContainer := range provisionerPod.Spec.Containers {
			images = append(images, provisionerContainer.Image)
		}
	}
	return images
}

// provisionerDeployment returns the Deployment the chart installs the
// provisioner as.
func provisionerDeployment() *appsv1.Deployment {
	deployments, listErr := ops.DeployClient.
		WithNamespace(openebsNamespace).
		List(context.TODO(), &metav1.ListOptions{LabelSelector: provisionerLabelSelector})
	Expect(listErr).ShouldNot(
		HaveOccurred(),
		"while listing the localpv-provisioner Deployment in namespace {%s}",
		openebsNamespace,
	)
	Expect(deployments.Items).To(
		HaveLen(1),
		"while looking for exactly one localpv-provisioner Deployment in namespace {%s}",
		openebsNamespace,
	)

	return &deployments.Items[0]
}

// expectProvisionerOnVersion waits until the only running provisioner is the
// one the given chart version deploys, and until its Deployment has finished
// rolling out.
func expectProvisionerOnVersion(imageTag string) {
	deployment := provisionerDeployment()

	Eventually(func() bool {
		rollout, rolloutErr := ops.DeployClient.
			WithNamespace(openebsNamespace).
			RolloutStatus(context.TODO(), deployment.Name)
		if rolloutErr != nil {
			return false
		}
		if !rollout.IsRolledout {
			GinkgoWriter.Println("rollout of " + deployment.Name + " is not done: " + rollout.Message)
			return false
		}

		// A Pod of the previous revision reports Running until it is gone, so
		// keep polling until every running provisioner is on the new image.
		images := provisionerImages()
		if len(images) == 0 {
			return false
		}
		for _, image := range images {
			if !strings.HasSuffix(image, ":"+imageTag) {
				return false
			}
		}
		return true
	}, 300, 5).Should(
		BeTrue(),
		"while waiting for the localpv-provisioner Deployment {%s} to roll out the {%s} image",
		deployment.Name,
		imageTag,
	)
}

// createHostpathPVC creates a PVC against the StorageClass the chart ships.
func createHostpathPVC(name, capacity string) *corev1.PersistentVolumeClaim {
	pvcObj, buildErr := tests.BuildPersistentVolumeClaim(
		namespaceObj.Name,
		name,
		hostpathClassName,
		capacity,
		[]corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
	)
	Expect(buildErr).ShouldNot(
		HaveOccurred(),
		"while building PVC {%s} in namespace {%s}",
		name,
		namespaceObj.Name,
	)

	pvcObj, createErr := ops.PVCClient.WithNamespace(namespaceObj.Name).Create(context.TODO(), pvcObj)
	Expect(createErr).ShouldNot(
		HaveOccurred(),
		"while creating PVC {%s} in namespace {%s}",
		name,
		namespaceObj.Name,
	)

	return pvcObj
}

// boundPV returns the PersistentVolume bound to the given claim.
func boundPV(pvcName string) *corev1.PersistentVolume {
	Expect(ops.IsPVCBoundEventually(namespaceObj.Name, pvcName)).To(
		BeTrue(),
		"while waiting for PVC {%s} in namespace {%s} to be bound",
		pvcName,
		namespaceObj.Name,
	)

	pvName := ops.GetPVNameFromPVCName(namespaceObj.Name, pvcName)
	Expect(pvName).ToNot(BeEmpty(), "while reading the PV name of PVC {%s}", pvcName)

	pvObj, getErr := ops.PVClient.Get(context.TODO(), pvName, metav1.GetOptions{})
	Expect(getErr).ShouldNot(HaveOccurred(), "while fetching PV {%s}", pvName)

	return pvObj
}

// hostpathOf returns the node directory a hostpath PV is backed by.
func hostpathOf(pvObj *corev1.PersistentVolume) string {
	Expect(pvObj.Spec.Local).ToNot(
		BeNil(),
		"while reading the local volume source of PV {%s}",
		pvObj.Name,
	)
	return pvObj.Spec.Local.Path
}

// storageOf returns the storage capacity of a PV, as a string, which is a
// comparison the cached formatting of a resource.Quantity cannot get wrong.
func storageOf(pvObj *corev1.PersistentVolume) string {
	capacity := pvObj.Spec.Capacity[corev1.ResourceStorage]
	return capacity.String()
}

// isHostpathVisible reports whether the suite is running on the node which
// hosts the volume, and can therefore make assertions about the directory
// itself. It is not the case when the tests are pointed at a remote cluster.
func isHostpathVisible(hostpath string) bool {
	_, statErr := os.Stat(hostpath)
	if statErr != nil {
		GinkgoWriter.Printf(
			"not asserting on the hostpath directory %s: %v\n", hostpath, statErr,
		)
	}
	return statErr == nil
}

// application is a single replica Deployment which mounts one PVC, used to
// keep a volume in use across the upgrade.
type application struct {
	name    string
	pvcName string
	labels  map[string]string
}

func newApplication(name, pvcName string) *application {
	return &application{
		name:    name,
		pvcName: pvcName,
		labels: map[string]string{
			"app": name,
			// ci-test.sh sweeps up leftovers by this label.
			"role": "test",
		},
	}
}

// selector is the label selector matching the application's Pods.
func (a *application) selector() string {
	return "app=" + a.name
}

// deploy creates the application and waits for its Pod to run, which is also
// what triggers the provisioning of its WaitForFirstConsumer volume.
func (a *application) deploy() {
	deployObj, buildErr := deploy.NewBuilder().
		WithName(a.name).
		WithNamespace(namespaceObj.Name).
		WithLabelsNew(a.labels).
		WithSelectorMatchLabelsNew(a.labels).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(a.labels).
				WithContainerBuildersNew(
					container.NewBuilder().
						WithName(appContainerName).
						WithImage("busybox").
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithCommandNew([]string{"/bin/sh"}).
						WithArgumentsNew([]string{"-c", "sleep 3600"}).
						WithVolumeMountsNew([]corev1.VolumeMount{
							{
								Name:      appVolumeName,
								MountPath: appMountPath,
							},
						}),
				).
				WithTerminationGracePeriodSeconds(5).
				WithVolumeBuilders(
					k8svolume.NewBuilder().
						WithName(appVolumeName).
						WithPVCSource(a.pvcName),
				),
		).
		Build()
	Expect(buildErr).ShouldNot(HaveOccurred(), "while building Deployment {%s}", a.name)

	_, createErr := ops.DeployClient.
		WithNamespace(namespaceObj.Name).
		Create(context.TODO(), deployObj)
	Expect(createErr).ShouldNot(HaveOccurred(), "while creating Deployment {%s}", a.name)

	a.waitForPod()
}

// waitForPod waits until exactly one of the application's Pods is running, and
// returns it.
func (a *application) waitForPod() *corev1.Pod {
	Expect(ops.GetPodRunningCountEventually(namespaceObj.Name, a.selector(), 1)).To(
		Equal(1),
		"while waiting for the Pod of Deployment {%s} to be running",
		a.name,
	)

	return a.pod()
}

// pod returns the application's running Pod.
func (a *application) pod() *corev1.Pod {
	pods := ops.GetPodList(
		namespaceObj.Name,
		a.selector(),
		pod.PredicateList{pod.IsRunning()},
	).ToAPIList().Items
	Expect(pods).To(HaveLen(1), "while looking for the Pod of Deployment {%s}", a.name)

	return &pods[0]
}

// restart replaces the application's Pod, which remounts its volume. It is how
// the upgraded provisioner's volumes are proven to still be usable.
func (a *application) restart() {
	previous := a.pod().Name

	Expect(ops.PodDeleteCollection(
		namespaceObj.Name,
		metav1.ListOptions{LabelSelector: a.selector()},
	)).To(Succeed(), "while deleting the Pod of Deployment {%s}", a.name)

	Eventually(func() bool {
		pods := ops.GetPodList(
			namespaceObj.Name,
			a.selector(),
			pod.PredicateList{pod.IsRunning()},
		).ToAPIList().Items
		return len(pods) == 1 && pods[0].Name != previous
	}, 300, 5).Should(
		BeTrue(),
		"while waiting for the Pod of Deployment {%s} to be recreated",
		a.name,
	)
}

// write writes contents into a file on the application's volume.
func (a *application) write(fileName, contents string) {
	path := fmt.Sprintf("%s/%s", appMountPath, fileName)
	ops.ExecuteCMDEventually(
		a.pod(),
		appContainerName,
		fmt.Sprintf("echo '%s' > %s && sync", contents, path),
		false,
	)
}

// read returns the contents of a file on the application's volume.
func (a *application) read(fileName string) string {
	path := fmt.Sprintf("%s/%s", appMountPath, fileName)
	return strings.TrimSpace(ops.ExecuteCMDEventually(
		a.pod(),
		appContainerName,
		"cat "+path,
		true,
	))
}

// delete removes the application, leaving its PVC behind.
func (a *application) delete() {
	Expect(ops.DeployClient.
		WithNamespace(namespaceObj.Name).
		Delete(context.TODO(), a.name, &metav1.DeleteOptions{}),
	).To(Succeed(), "while deleting Deployment {%s}", a.name)

	Expect(ops.GetPodRunningCountEventually(namespaceObj.Name, a.selector(), 0)).To(
		Equal(0),
		"while waiting for the Pods of Deployment {%s} to be gone",
		a.name,
	)
}
