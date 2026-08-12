package upgrade

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ns "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/namespace"
	"github.com/openebs/dynamic-localpv-provisioner/tests"
)

const (
	namespacePrefix = "localpv-upgrade-ns"

	// provisionerLabelSelector selects the localpv-provisioner Deployment and
	// its Pods, whichever chart version installed them.
	provisionerLabelSelector = "openebs.io/component-name=openebs-localpv-provisioner"

	// hostpathClassName is the StorageClass the chart itself ships. The upgrade
	// specs provision from it on purpose, so that the upgrade of the
	// StorageClass is exercised along with everything else.
	hostpathClassName = "openebs-hostpath"

	// imagePullPolicyEnv overrides how the upgraded release pulls the
	// provisioner image. It defaults to 'Never' because the image under test is
	// the one 'make provisioner-localpv-image' built into the cluster's runtime.
	imagePullPolicyEnv = "UPGRADE_IMAGE_PULL_POLICY"

	// namespaceTerminationTimeout bounds the best-effort wait for the test
	// namespace to go away after the suite.
	namespaceTerminationTimeout = 5 * time.Minute
)

var (
	kubeConfigPath    string
	openebsNamespace  string
	chartDir          string
	releaseName       string
	skipCleanup       bool
	uninstallExisting bool

	ops          *tests.Operations
	helm         *helmClient
	namespaceObj *corev1.Namespace
	err          error

	// releaseInstalled records that the release under test is one this suite
	// installed. Ginkgo runs AfterSuite even when BeforeSuite fails, so without
	// it a failure before the install would tear down somebody else's release.
	releaseInstalled bool

	// localChart is the chart under test, as read from the working tree.
	localChart chartMetadata
	// localValues holds the default values of the chart under test.
	localValues chartValues
	// fromVersion is the released chart version the upgrade starts from.
	fromVersion string
)

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test localpv-provisioner helm chart upgrade")
}

func init() {
	flag.StringVar(&kubeConfigPath, "kubeconfig", os.Getenv("KUBECONFIG"),
		"path to kubeconfig to invoke kubernetes API calls")
	flag.StringVar(&openebsNamespace, "openebs-namespace", envOrDefault("OPENEBS_NAMESPACE", "openebs"),
		"kubernetes namespace where the OpenEBS components are installed")
	flag.StringVar(&chartDir, "chart-dir", envOrDefault(chartDirEnv, defaultChartDir()),
		"path to the localpv-provisioner helm chart to upgrade to")
	flag.StringVar(&releaseName, "release-name", envOrDefault(releaseNameEnv, "localpv-provisioner"),
		"name of the helm release under test")
	flag.BoolVar(&skipCleanup, "skip-cleanup", envBool(skipCleanupEnv, false),
		"leave the helm release and the test namespace behind, to debug with")
	flag.BoolVar(&uninstallExisting, "uninstall-existing", envBool(uninstallExistingEnv, false),
		"uninstall a release which is already using the name under test, instead of refusing to run")
}

var _ = BeforeSuite(func() {

	ops = tests.NewOperations(tests.WithKubeConfigPath(kubeConfigPath))

	By("reading the chart under test at " + chartDir)
	localChart, err = readChartMetadata(chartDir)
	Expect(err).To(BeNil(), "while reading the chart metadata at {%s}", chartDir)
	localValues, err = readChartValues(chartDir)
	Expect(err).To(BeNil(), "while reading the chart values at {%s}", chartDir)
	Expect(localValues.provisionerImageTag()).ToNot(
		BeEmpty(),
		"while reading localpv.image.tag from the values of the chart at {%s}",
		chartDir,
	)

	By("setting up a helm client")
	helm, err = newHelmClient(openebsNamespace, releaseName, kubeConfigPath)
	Expect(err).To(BeNil(), "while setting up the helm client")
	DeferCleanup(func() {
		Expect(helm.cleanup()).To(Succeed(), "while removing the temporary helm home")
	})

	By("adding the released chart repository at " + envOrDefault(chartRepoURLEnv, defaultChartRepoURL))
	Expect(addReleasedChartRepo(helm)).To(Succeed(), "while adding the released chart repository")

	By("resolving the released chart version to upgrade from")
	fromVersion, err = resolveUpgradeFromVersion(helm, localChart.Version)
	Expect(err).To(
		BeNil(),
		"while resolving the released version to upgrade to v%s from",
		localChart.Version,
	)
	upgradePath := fmt.Sprintf("v%s -> v%s", fromVersion, localChart.Version)
	AddReportEntry("upgrade path", upgradePath)
	GinkgoWriter.Println("upgrading the localpv-provisioner chart " + upgradePath)

	By("checking that the '" + releaseName + "' release name is free")
	nameInUse, listErr := helm.installed()
	Expect(listErr).To(
		BeNil(),
		"while looking for an existing helm release {%s} in namespace {%s}",
		releaseName,
		openebsNamespace,
	)

	// The suite installs and uninstalls this release itself, so a release it did
	// not create is not its to remove: on a cluster which is not disposable that
	// would take out a running provisioner and whatever depends on it.
	if nameInUse && !uninstallExisting {
		Fail(fmt.Sprintf(
			"the helm release %q already exists in namespace %q. These tests install and "+
				"uninstall that release themselves, and will not remove one they did not "+
				"create. Uninstall it first, point the tests at another release with "+
				"-release-name, or pass -uninstall-existing (./ci/ci-test.sh upgrade -r) to "+
				"have it removed.",
			releaseName,
			openebsNamespace,
		))
	}

	if nameInUse {
		By("uninstalling the pre-existing '" + releaseName + "' helm release")
		Expect(helm.uninstall()).To(
			Succeed(),
			"while uninstalling the pre-existing helm release {%s} in namespace {%s}",
			releaseName,
			openebsNamespace,
		)
	}

	By("installing the released localpv-provisioner chart v" + fromVersion)
	installErr := helm.install(releasedChartRef(), fromVersion, "--set", "analytics.enabled=false")
	// Anything helm managed to create is this suite's to remove, even when the
	// install itself then failed.
	releaseInstalled = true
	Expect(installErr).To(Succeed(), "while installing the released chart v%s", fromVersion)

	By("waiting for the v" + fromVersion + " localpv-provisioner Pod to be running")
	provPodCount := ops.GetPodRunningCountEventually(
		openebsNamespace,
		provisionerLabelSelector,
		1,
	)
	Expect(provPodCount).To(Equal(1))

	By("verifying that the running provisioner is the released one")
	releasedMetadata, metadataErr := helm.metadata()
	Expect(metadataErr).To(BeNil(), "while reading the released release's metadata")
	Expect(releasedMetadata.Version).To(
		Equal(fromVersion),
		"while verifying that the release is on the chart version asked for",
	)
	// The image tag follows the chart's appVersion, which is not always the
	// chart version: chart-only patches have shipped with the two apart, as
	// v3.4.1 did with an appVersion of 3.4.0.
	Expect(provisionerImages()).To(
		HaveEach(HaveSuffix(":"+releasedMetadata.AppVersion)),
		"while verifying that the provisioner Pod runs the released v%s image",
		releasedMetadata.AppVersion,
	)

	By("building a namespace")
	namespaceObj, err = ns.NewBuilder().
		WithGenerateName(namespacePrefix).
		APIObject()
	Expect(err).ShouldNot(HaveOccurred(), "while building namespace {%s}", namespacePrefix)

	By("creating above namespace")
	namespaceObj, err = ops.NSClient.Create(namespaceObj)
	Expect(err).To(BeNil(), "while creating namespace with prefix {%s}", namespacePrefix)
	ops.NameSpace = namespaceObj.Name
})

var _ = AfterSuite(func() {

	if ops == nil {
		return
	}

	// The suite owns the release and the namespace, so 'ci-test.sh upgrade -x'
	// only leaves them behind if the request reaches here too.
	if skipCleanup {
		left := "the '" + releaseName + "' helm release"
		if namespaceObj != nil {
			left += " and namespace " + namespaceObj.Name
		}
		GinkgoWriter.Println("skipping cleanup, leaving " + left + " behind")
		return
	}

	if namespaceObj != nil {
		By("deleting namespace " + namespaceObj.Name)
		err = ops.NSClient.Delete(namespaceObj.Name, &metav1.DeleteOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			Expect(err).To(BeNil(), "while deleting namespace {%s}", namespaceObj.Name)
		}

		// The volumes left in the namespace are cleaned up by the provisioner,
		// which is uninstalled right after this. Give it a chance to finish,
		// but do not fail the suite over it: a failed spec is the more useful
		// signal, and ci-test.sh sweeps up whatever is left behind.
		By("waiting for namespace " + namespaceObj.Name + " to be gone")
		if !namespaceGone(namespaceObj.Name, namespaceTerminationTimeout) {
			GinkgoWriter.Println("namespace " + namespaceObj.Name + " is still terminating")
		}
	}

	// Only ever remove the release this suite installed. AfterSuite runs even
	// when BeforeSuite failed, so a pre-existing release which the suite refused
	// to touch must survive this.
	if helm != nil && releaseInstalled {
		By("uninstalling the '" + releaseName + "' helm release")
		Expect(helm.uninstall()).To(
			Succeed(),
			"while uninstalling the helm release {%s} in namespace {%s}",
			releaseName,
			openebsNamespace,
		)
	}
})

// namespaceGone polls until the namespace has been fully removed, and reports
// whether it got there within the timeout.
func namespaceGone(name string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if _, getErr := ops.NSClient.Get(name, metav1.GetOptions{}); k8serrors.IsNotFound(getErr) {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(5 * time.Second)
	}
}
