package upgrade

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("UPGRADE LOCALPV PROVISIONER", Ordered, func() {
	const (
		capacity = "1Gi"

		markerFileName = "before-upgrade.txt"
		markerContents = "written before the upgrade"

		preUpgradePVCName = "pvc-upgrade-from"
		preUpgradeAppName = "busybox-upgrade-from"

		postUpgradePVCName = "pvc-upgrade-to"
		postUpgradeAppName = "busybox-upgrade-to"
	)

	var (
		preUpgradeApp  *application
		postUpgradeApp *application

		// State carried from the pre-upgrade specs into the post-upgrade ones.
		preUpgradePV    *corev1.PersistentVolume
		preUpgradePath  string
		hostpathVisible bool
	)

	BeforeAll(func() {
		preUpgradeApp = newApplication(preUpgradeAppName, preUpgradePVCName)
		postUpgradeApp = newApplication(postUpgradeAppName, postUpgradePVCName)
	})

	When("a volume is provisioned by the released chart", func() {
		It("should bind a PVC created against the chart's hostpath StorageClass", func() {
			By("creating a PVC against the " + hostpathClassName + " StorageClass")
			createHostpathPVC(preUpgradePVCName, capacity)

			By("deploying an application which consumes the PVC")
			preUpgradeApp.deploy()

			By("verifying that the PVC is bound")
			preUpgradePV = boundPV(preUpgradePVCName)
			preUpgradePath = hostpathOf(preUpgradePV)
			hostpathVisible = isHostpathVisible(preUpgradePath)
		})

		It("should let the application write to the volume", func() {
			By("writing to a file on the volume")
			preUpgradeApp.write(markerFileName, markerContents)

			By("reading the file back")
			Expect(preUpgradeApp.read(markerFileName)).To(
				Equal(markerContents),
				"while reading {%s} back from the volume of PVC {%s}",
				markerFileName,
				preUpgradePVCName,
			)
		})
	})

	When("the release is upgraded to the chart under test", func() {
		It("should complete the helm upgrade", func() {
			By("upgrading the '" + releaseName + "' release to the chart at " + chartDir)
			Expect(helm.upgrade(
				chartDir,
				"--reset-then-reuse-values",
				"--set", "localpv.image.pullPolicy="+envOrDefault(imagePullPolicyEnv, "Never"),
			)).To(
				Succeed(),
				"while upgrading the helm release {%s} from v%s to v%s",
				releaseName,
				fromVersion,
				localChart.Version,
			)
		})

		It("should report the chart under test in the release metadata", func() {
			metadata, metadataErr := helm.metadata()
			Expect(metadataErr).To(BeNil(), "while reading the helm release metadata")

			Expect(metadata.Version).To(
				Equal(localChart.Version),
				"while verifying that the release is on the chart version under test",
			)
			Expect(metadata.AppVersion).To(
				Equal(localChart.AppVersion),
				"while verifying that the release is on the app version under test",
			)
			Expect(metadata.Status).To(Equal("deployed"))
		})

		It("should roll out the provisioner from the chart under test", func() {
			// expectProvisionerOnVersion requires every running provisioner to be
			// on this tag, and at least one to be running, so it already rules
			// out any Pod left behind on the released image.
			imageTag := localValues.provisionerImageTag()

			By("waiting for the provisioner to be rolled out on the " + imageTag + " image")
			expectProvisionerOnVersion(imageTag)
		})
	})

	When("the upgrade is done", func() {
		It("should leave the pre-upgrade PersistentVolume untouched", func() {
			pvObj, getErr := ops.PVClient.Get(context.TODO(), preUpgradePV.Name, metav1.GetOptions{})
			Expect(getErr).ShouldNot(
				HaveOccurred(),
				"while fetching the pre-upgrade PV {%s}",
				preUpgradePV.Name,
			)

			Expect(pvObj.UID).To(Equal(preUpgradePV.UID), "the pre-upgrade PV was replaced")
			Expect(hostpathOf(pvObj)).To(Equal(preUpgradePath))
			Expect(storageOf(pvObj)).To(Equal(storageOf(preUpgradePV)))
			Expect(pvObj.Spec.NodeAffinity).To(Equal(preUpgradePV.Spec.NodeAffinity))
			Expect(pvObj.Spec.PersistentVolumeReclaimPolicy).To(
				Equal(preUpgradePV.Spec.PersistentVolumeReclaimPolicy),
			)
			Expect(pvObj.Status.Phase).To(Equal(corev1.VolumeBound))

			Expect(ops.IsPVCBound(namespaceObj.Name, preUpgradePVCName)).To(
				BeTrue(),
				"while verifying that PVC {%s} is still bound",
				preUpgradePVCName,
			)
		})

		It("should still serve the data written before the upgrade", func() {
			By("restarting the application, so that it remounts the volume")
			preUpgradeApp.restart()

			By("reading the file written before the upgrade")
			Expect(preUpgradeApp.read(markerFileName)).To(
				Equal(markerContents),
				"while reading {%s} back from the volume of PVC {%s} after the upgrade",
				markerFileName,
				preUpgradePVCName,
			)
		})

		It("should provision a new volume", func() {
			By("creating a PVC against the " + hostpathClassName + " StorageClass")
			createHostpathPVC(postUpgradePVCName, capacity)

			By("deploying an application which consumes the PVC")
			postUpgradeApp.deploy()

			By("verifying that a new PV was provisioned")
			postUpgradePV := boundPV(postUpgradePVCName)
			Expect(postUpgradePV.Name).ToNot(Equal(preUpgradePV.Name))

			By("writing to, and reading from, the new volume")
			postUpgradeApp.write(markerFileName, markerContents)
			Expect(postUpgradeApp.read(markerFileName)).To(Equal(markerContents))
		})

		It("should clean up a pre-upgrade volume when its PVC is deleted", func() {
			By("deleting the application which consumes the pre-upgrade PVC")
			preUpgradeApp.delete()

			By("deleting the pre-upgrade PVC")
			ops.DeletePersistentVolumeClaim(preUpgradePVCName, namespaceObj.Name)
			Expect(ops.IsPVCDeletedEventually(preUpgradePVCName, namespaceObj.Name)).To(
				BeTrue(),
				"while waiting for PVC {%s} to be deleted",
				preUpgradePVCName,
			)

			By("verifying that the upgraded provisioner deleted the PV")
			Eventually(func() bool {
				return ops.IsPVDeleted(preUpgradePV.Name)
			}, 300, 5).Should(
				BeTrue(),
				"while waiting for PV {%s} to be deleted",
				preUpgradePV.Name,
			)
		})

		It("should remove the volume directory of a deleted pre-upgrade volume", func() {
			if !hostpathVisible {
				Skip("the volume directory is not visible from where the tests run")
			}

			Eventually(func() bool {
				_, statErr := os.Stat(preUpgradePath)
				return os.IsNotExist(statErr)
			}, 300, 5).Should(
				BeTrue(),
				"while waiting for the hostpath directory {%s} to be cleaned up",
				preUpgradePath,
			)
		})
	})
})
