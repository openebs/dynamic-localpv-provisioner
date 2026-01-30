package tests

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ns "github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/namespace"
	"github.com/openebs/dynamic-localpv-provisioner/tests/disk"
)

const (
	namespacePrefix                 = "localpv-ns"
	storageClassLabelSelector       = "openebs.io/test-sc=true"
	LocalPVProvisionerLabelSelector = "openebs.io/component-name=openebs-localpv-provisioner"
	ndmLabelSelector                = "openebs.io/component-name=ndm"
	ndmOperatorLabelSelector        = "openebs.io/component-name=ndm-operator"
	openebsRootDir                  = "/var/openebs"
	hostpathDirNamePrefix           = "localpv-integration-test"
	loopHostpathDirName             = "loop-mountpoint"
	loopDiskImgDirName              = "loop-image"
)

var (
	kubeConfigPath   string
	openebsNamespace string
	namespaceObj     *corev1.Namespace
	hostpathDir      string
	loopHostpathDir  string
	loopDiskImgDir   string
	err              error
	physicalDisk     = disk.Disk{}
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test application deployment")
}

func init() {
	flag.StringVar(&kubeConfigPath, "kubeconfig", os.Getenv("KUBECONFIG"), "path to kubeconfig to invoke kubernetes API calls")
	flag.StringVar(&openebsNamespace, "openebs-namespace", "openebs", "kubernetes namespace where the OpenEBS components are present")
}

var ops *Operations

var _ = BeforeSuite(func() {

	ops = NewOperations(WithKubeConfigPath(kubeConfigPath))

	By("waiting for openebs-localpv-provisioner pod to come into running state")
	provPodCount := ops.GetPodRunningCountEventually(
		openebsNamespace,
		LocalPVProvisionerLabelSelector,
		1,
	)
	Expect(provPodCount).To(Equal(1))

	By("building a namespace")
	namespaceObj, err = ns.NewBuilder().
		WithGenerateName(namespacePrefix).
		APIObject()
	Expect(err).ShouldNot(HaveOccurred(), "while building namespace {%s}", namespaceObj.GenerateName)

	By("creating above namespace")
	namespaceObj, err = ops.NSClient.Create(namespaceObj)
	Expect(err).To(BeNil(), "while creating namespace with prefix {%s}", namespacePrefix)
	ops.NameSpace = namespaceObj.Name

	By("creating a directory for hostpath tests")
	err = os.MkdirAll(openebsRootDir, 0750)
	Expect(err).To(BeNil(), "when creating directory "+openebsRootDir)
	hostpathDir, err = os.MkdirTemp(openebsRootDir, hostpathDirNamePrefix+"-*")
	Expect(err).To(BeNil(), "when creating hostpath directory")
	loopHostpathDir = filepath.Join(hostpathDir, loopHostpathDirName)
	loopDiskImgDir = filepath.Join(hostpathDir, loopDiskImgDirName)

	By("preparing the loop device for hostpath Quota tests")
	physicalDisk, err = disk.PrepareDisk(loopDiskImgDir, loopHostpathDir)
	Expect(err).To(BeNil(), "while preparing disk {%+v}", physicalDisk)
})

var _ = AfterSuite(func() {

	By("deleting namespace")
	err = ops.NSClient.Delete(namespaceObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting namespace {%s}", namespaceObj.Name)

	By("deleting test StorageClasses")
	err = ops.SCClient.DeleteCollection(
		context.TODO(),
		metav1.ListOptions{
			LabelSelector: storageClassLabelSelector,
		},
		&metav1.DeleteOptions{},
	)

	By("destroying the created disk")
	err = physicalDisk.DestroyDisk(loopDiskImgDir, loopHostpathDir)
	Expect(err).To(BeNil(), "while destroying the disk {%+v}", physicalDisk)

	By("removing the hostpath directory")
	err = os.RemoveAll(hostpathDir)
	Expect(err).To(
		BeNil(),
		"when removing the hostpath directory at {%s}",
		hostpathDir,
	)
})
