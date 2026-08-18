package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pvController "sigs.k8s.io/sig-storage-lib-external-provisioner/v13/controller"

	mconfig "github.com/openebs/dynamic-localpv-provisioner/pkg/apis/openebs.io/v1alpha1"
)

// TestProvisionHostPathNodeDeploymentAppliesFsMode pins the regression this
// change fixes: in node-deployment mode the volume directory used to be created
// with a hardcoded 0777, ignoring the FilePermissions mode from the cas-config.
//
// The assertion is on the permission bits of the directory that actually gets
// created, so the test fails if any link in the chain breaks - the fsMode
// plumbed onto HelperPodOptions in ProvisionHostPath, the VolumeRequest built
// from it, or the mkdir mode used by LocalVolumeManager.
func TestProvisionHostPathNodeDeploymentAppliesFsMode(t *testing.T) {
	const (
		basePath = "/var/openebs/local"
		pvName   = "pvc-fsmode"
	)

	tests := map[string]struct {
		mode     string
		wantPerm os.FileMode
	}{
		"configured mode is honoured":              {mode: "0750", wantPerm: 0750},
		"restrictive mode is honoured":             {mode: "0700", wantPerm: 0700},
		"unset FilePermissions falls back to 0777": {mode: "", wantPerm: 0777},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Stand in for the host root that the node DaemonSet mounts at /host.
			hostRoot := t.TempDir()
			original := HostPathPrefix
			HostPathPrefix = hostRoot
			t.Cleanup(func() { HostPathPrefix = original })

			configData := map[string]interface{}{}
			if tc.mode != "" {
				configData[KeyFilePermissions] = map[string]RawLiteral{
					KeyFsMode: RawLiteral(tc.mode),
				}
			}

			volumeConfig := &VolumeConfig{
				pvName: pvName,
				options: map[string]interface{}{
					KeyPVBasePath: map[string]string{
						string(mconfig.ValuePTP): basePath,
					},
				},
				configData: configData,
			}

			p := &Provisioner{nodeDeployment: true}

			pv, state, err := p.ProvisionHostPath(
				context.Background(),
				provisionOptionsForFsMode(pvName),
				volumeConfig,
				nodeForFsMode(),
			)
			if err != nil {
				t.Fatalf("ProvisionHostPath() error = %v", err)
			}
			if pv == nil {
				t.Fatalf("ProvisionHostPath() returned a nil PV, state %v", state)
			}

			created := filepath.Join(hostRoot, basePath, pvName)
			info, err := os.Stat(created)
			if err != nil {
				t.Fatalf("expected the volume directory at %s: %v", created, err)
			}
			if got := info.Mode().Perm(); got != tc.wantPerm {
				t.Errorf("volume directory mode = %#o, want %#o", got, tc.wantPerm)
			}
		})
	}
}

func provisionOptionsForFsMode(pvName string) pvController.ProvisionOptions {
	return pvController.ProvisionOptions{
		PVName: pvName,
		PVC: &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: "claim", Namespace: "default"},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("64Mi"),
					},
				},
			},
		},
	}
}

func nodeForFsMode() *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node-one",
			Labels: map[string]string{k8sNodeLabelKeyHostname: "node-one"},
		},
	}
}
