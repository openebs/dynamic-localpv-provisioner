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

package storageclass

import (
	"testing"

	mconfig "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	fakeCASTypeValue = "cstor"

	fakeProvisionerName = "openebs.io/nfsrwx"

	defaultHostpath = "/var/openebs/local"

	fakeCASConfigValue = "- name: StoragePoolClaim\n" +
		"  value: \"cstor-disk-pool\"\n" +
		"- name: ReplicaCount\n" +
		"  value: \"3\"\n"

	mockHostpathConfig = "- name: StorageType\n" +
		"  value: \"hostpath\"\n" +
		"- name: BasePath\n" +
		"  value: \"/var/openebs/local\"\n"

	mockDeviceConfig = "- name: StorageType\n" +
		"  value: \"device\"\n"

	mockNodeAffinityLabelConfig = "- name: NodeAffinityLabel\n" +
		"  list: \n" +
		"    - \"openebs.io/mock\"\n"

	mockBlockDeviceTagConfig = "- name: BlockDeviceTag\n" +
		"  value: \"openebs.io/mock-ssd\"\n"
)

func TestBuildWithName(t *testing.T) {
	tests := map[string]struct {
		name         string
		storageClass *storagev1.StorageClass
		expectErr    bool
	}{
		"Build with valid name": {
			name:         "demo-sc",
			storageClass: &storagev1.StorageClass{},
			expectErr:    false,
		},
		"Build with empty name": {
			name:         "",
			storageClass: &storagev1.StorageClass{},
			expectErr:    true,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			opt := WithName(mock.name)
			err := opt(mock.storageClass)

			if mock.expectErr && err == nil {
				t.Fatal("Test '" + name + "' failed: expected error to not be nil.")
			}
			if !mock.expectErr && err != nil {
				t.Fatal("Test '" + name + "' failed: expected error to be nil.")
			}
		})
	}
}

func TestBuildWithGenerateName(t *testing.T) {
	tests := map[string]struct {
		generateName string
		storageClass *storagev1.StorageClass
		expectErr    bool
	}{
		"Build with valid generateName": {
			generateName: "demo-sc-",
			storageClass: &storagev1.StorageClass{},
			expectErr:    false,
		},
		"Build with empty generateName": {
			generateName: "",
			storageClass: &storagev1.StorageClass{},
			expectErr:    true,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			opt := WithGenerateName(mock.generateName)
			err := opt(mock.storageClass)

			if mock.expectErr && err == nil {
				t.Fatal("Test '" + name + "' failed: expected error to not be nil.")
			}
			if !mock.expectErr && err != nil {
				t.Fatal("Test '" + name + "' failed: expected error to be nil.")
			}
		})
	}
}

func TestBuildWithLocalPV(t *testing.T) {
	tests := map[string]struct {
		storageClass *storagev1.StorageClass
		expectErr    bool
	}{
		"Build with " + string(mconfig.CASTypeKey) +
			" annotation and Provisioner name" +
			" not set": {
			storageClass: &storagev1.StorageClass{},
			expectErr:    false,
		},
		"Build with " + string(mconfig.CASTypeKey) +
			" annotation set": {
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): fakeCASTypeValue,
					},
				},
			},
			expectErr: true,
		},
		"Build with Provisioner name set": {
			storageClass: &storagev1.StorageClass{
				Provisioner: fakeProvisionerName,
			},
			expectErr: true,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			opt := WithLocalPV()
			err := opt(mock.storageClass)

			if mock.expectErr && err == nil {
				t.Fatal("Test '" + name + "' failed: expected error to not be nil.")
			}
			if !mock.expectErr && err != nil {
				t.Fatal("Test '" + name + "' failed: expected error to be nil.")
			}
		})
	}
}

func TestBuildWithHostpath(t *testing.T) {
	tests := map[string]struct {
		hostpathDir  string
		storageClass *storagev1.StorageClass
		expectErr    bool
	}{
		"Build with valid hostpath directory and empty annotations, Provisioner name": {
			hostpathDir:  "/mnt/data",
			storageClass: &storagev1.StorageClass{},
			expectErr:    false,
		},
		"Build with invalid hostpath directory (under root)": {
			hostpathDir:  "/data",
			storageClass: &storagev1.StorageClass{},
			expectErr:    true,
		},
		"Build with invalid hostpath directory (relative path)": {
			hostpathDir:  "../data",
			storageClass: &storagev1.StorageClass{},
			expectErr:    true,
		},
		"Build with empty hostpath directory": {
			hostpathDir:  "",
			storageClass: &storagev1.StorageClass{},
			expectErr:    true,
		},
		"Build with compatible '" + string(mconfig.CASConfigKey) + "' annotation": {
			hostpathDir: defaultHostpath,
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): mockNodeAffinityLabelConfig,
					},
				},
			},
			expectErr: false,
		},
		"Build with incompatible '" + string(mconfig.CASConfigKey) + "' annotation": {
			hostpathDir: defaultHostpath,
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): mockBlockDeviceTagConfig,
					},
				},
			},
			expectErr: true,
		},
		"Build with empty '" + string(mconfig.CASConfigKey) + "' annotation": {
			hostpathDir: defaultHostpath,
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): "",
					},
				},
			},
			expectErr: false,
		},
		"Build with valid '" + string(mconfig.CASTypeKey) + "' annotation": {
			hostpathDir: defaultHostpath,
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): localPVcasTypeValue,
					},
				},
			},
			expectErr: false,
		},
		"Build with invalid '" + string(mconfig.CASTypeKey) + "' annotation": {
			hostpathDir: defaultHostpath,
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): fakeCASTypeValue,
					},
				},
			},
			expectErr: true,
		},
		"Build with empty '" + string(mconfig.CASTypeKey) + "' annotation": {
			hostpathDir: defaultHostpath,
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): "",
					},
				},
			},
			expectErr: false,
		},
		"Build with valid Provisioner name": {
			hostpathDir: defaultHostpath,
			storageClass: &storagev1.StorageClass{
				Provisioner: localPVprovisionerName,
			},
			expectErr: false,
		},
		"Build with invalid Provisioner name": {
			hostpathDir: defaultHostpath,
			storageClass: &storagev1.StorageClass{
				Provisioner: fakeProvisionerName,
			},
			expectErr: true,
		},
		"Build with empty Provisioner name": {
			hostpathDir: defaultHostpath,
			storageClass: &storagev1.StorageClass{
				Provisioner: "",
			},
			expectErr: false,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			opt := WithHostpath(mock.hostpathDir)
			err := opt(mock.storageClass)

			if mock.expectErr && err == nil {
				t.Fatal("Test '" + name + "' failed: expected error to not be nil.")
			}
			if !mock.expectErr && err != nil {
				t.Fatal("Test '" + name + "' failed: expected error to be nil.")
			}
		})
	}
}

func TestBuildWithXfsQuota(t *testing.T) {
	tests := map[string]struct {
		softLimit    string
		hardLimit    string
		storageClass *storagev1.StorageClass
		expectErr    bool
	}{
		"Build with sane limits": {
			softLimit:    "75%",
			hardLimit:    "80%",
			storageClass: &storagev1.StorageClass{},
			expectErr:    false,
		},
		"Build with empty softLimit": {
			softLimit:    "",
			hardLimit:    "80%",
			storageClass: &storagev1.StorageClass{},
			expectErr:    false,
		},
		"Build with empty hardLimit": {
			softLimit:    "80%",
			hardLimit:    "",
			storageClass: &storagev1.StorageClass{},
			expectErr:    false,
		},
		"Build with compatible '" + string(mconfig.CASConfigKey) + "' annotation": {
			softLimit: "75%",
			hardLimit: "80%",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): mockNodeAffinityLabelConfig,
					},
				},
			},
			expectErr: false,
		},
		"Build with incompatible '" + string(mconfig.CASConfigKey) + "' annotation": {
			softLimit: "75%",
			hardLimit: "80%",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): mockBlockDeviceTagConfig,
					},
				},
			},
			expectErr: true,
		},
		"Build with empty '" + string(mconfig.CASConfigKey) + "' annotation": {
			softLimit: "75%",
			hardLimit: "80%",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): "",
					},
				},
			},
			expectErr: false,
		},
		"Build with valid '" + string(mconfig.CASTypeKey) + "' annotation": {
			softLimit: "75%",
			hardLimit: "80%",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): localPVcasTypeValue,
					},
				},
			},
			expectErr: false,
		},
		"Build with invalid '" + string(mconfig.CASTypeKey) + "' annotation": {
			softLimit: "75%",
			hardLimit: "80%",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): fakeCASTypeValue,
					},
				},
			},
			expectErr: true,
		},
		"Build with empty '" + string(mconfig.CASTypeKey) + "' annotation": {
			softLimit: "75%",
			hardLimit: "80%",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): "",
					},
				},
			},
			expectErr: false,
		},
		"Build with valid Provisioner name": {
			softLimit: "75%",
			hardLimit: "80%",
			storageClass: &storagev1.StorageClass{
				Provisioner: localPVprovisionerName,
			},
			expectErr: false,
		},
		"Build with invalid Provisioner name": {
			softLimit: "75%",
			hardLimit: "80%",
			storageClass: &storagev1.StorageClass{
				Provisioner: fakeProvisionerName,
			},
			expectErr: true,
		},
		"Build with empty Provisioner name": {
			softLimit: "75%",
			hardLimit: "80%",
			storageClass: &storagev1.StorageClass{
				Provisioner: "",
			},
			expectErr: false,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			opt := WithXfsQuota(mock.softLimit, mock.hardLimit)
			err := opt(mock.storageClass)

			if mock.expectErr && err == nil {
				t.Fatal("Test '" + name + "' failed: expected error to not be nil.")
			}
			if !mock.expectErr && err != nil {
				t.Fatal("Test '" + name + "' failed: expected error to be nil.")
			}
		})
	}
}

func TestBuildWithDevice(t *testing.T) {
	tests := map[string]struct {
		storageClass *storagev1.StorageClass
		expectErr    bool
	}{
		"Build with empty StorageClass": {
			storageClass: &storagev1.StorageClass{},
			expectErr:    false,
		},
		"Build with compatible '" + string(mconfig.CASConfigKey) + "' annotation": {
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): mockBlockDeviceTagConfig,
					},
				},
			},
			expectErr: false,
		},
		"Build with incompatible '" + string(mconfig.CASConfigKey) + "' annotation": {
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): mockNodeAffinityLabelConfig,
					},
				},
			},
			expectErr: true,
		},
		"Build with empty '" + string(mconfig.CASConfigKey) + "' annotation": {
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): "",
					},
				},
			},
			expectErr: false,
		},
		"Build with valid '" + string(mconfig.CASTypeKey) + "' annotation": {
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): localPVcasTypeValue,
					},
				},
			},
			expectErr: false,
		},
		"Build with invalid '" + string(mconfig.CASTypeKey) + "' annotation": {
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): fakeCASTypeValue,
					},
				},
			},
			expectErr: true,
		},
		"Build with empty '" + string(mconfig.CASTypeKey) + "' annotation": {
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): "",
					},
				},
			},
			expectErr: false,
		},
		"Build with valid Provisioner name": {
			storageClass: &storagev1.StorageClass{
				Provisioner: localPVprovisionerName,
			},
			expectErr: false,
		},
		"Build with invalid Provisioner name": {
			storageClass: &storagev1.StorageClass{
				Provisioner: fakeProvisionerName,
			},
			expectErr: true,
		},
		"Build with empty Provisioner name": {
			storageClass: &storagev1.StorageClass{
				Provisioner: "",
			},
			expectErr: false,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			opt := WithDevice()
			err := opt(mock.storageClass)

			if mock.expectErr && err == nil {
				t.Fatal("Test '" + name + "' failed: expected error to not be nil.")
			}
			if !mock.expectErr && err != nil {
				t.Fatal("Test '" + name + "' failed: expected error to be nil.")
			}
		})
	}
}

func TestBuildWithVolumeBindingMode(t *testing.T) {
	tests := map[string]struct {
		volBindMode  storagev1.VolumeBindingMode
		storageClass *storagev1.StorageClass
		checkFn      func(*storagev1.StorageClass) error
		expectErr    bool
	}{
		"Build with empty/default VolumeBindingMode (WaitForFIrstConsumer)": {
			volBindMode:  "",
			storageClass: &storagev1.StorageClass{},
			checkFn: func(s *storagev1.StorageClass) error {
				if *s.VolumeBindingMode == storagev1.VolumeBindingWaitForFirstConsumer {
					return nil
				}
				return errors.New("Failed to set default " +
					"VolumeBindingMode as WaitForFIrstConsumer")
			},
			expectErr: false,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			opt := WithVolumeBindingMode(mock.volBindMode)
			_ = opt(mock.storageClass)
			err := mock.checkFn(mock.storageClass)

			if mock.expectErr && err == nil {
				t.Fatal("Test '" + name + "' failed: expected error to not be nil.")
			}
			if !mock.expectErr && err != nil {
				t.Fatal("Test '" + name + "' failed: expected error to be nil.")
			}
		})
	}
}

func TestBuildWithReclaimPolicy(t *testing.T) {
	tests := map[string]struct {
		rPolicy      corev1.PersistentVolumeReclaimPolicy
		storageClass *storagev1.StorageClass
		checkFn      func(*storagev1.StorageClass) error
		expectErr    bool
	}{
		"Build with empty/default ReclaimPolicy (Delete)": {
			rPolicy:      "",
			storageClass: &storagev1.StorageClass{},
			checkFn: func(s *storagev1.StorageClass) error {
				if *s.ReclaimPolicy == corev1.PersistentVolumeReclaimDelete {
					return nil
				}
				return errors.New("Failed to set default " +
					"ReclaimPolicy as Delete")
			},
			expectErr: false,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			opt := WithReclaimPolicy(mock.rPolicy)
			_ = opt(mock.storageClass)
			err := mock.checkFn(mock.storageClass)

			if mock.expectErr && err == nil {
				t.Fatal("Test '" + name + "' failed: expected error to not be nil.")
			}
			if !mock.expectErr && err != nil {
				t.Fatal("Test '" + name + "' failed: expected error to be nil.")
			}
		})
	}
}

func TestBuildWithAllowedTopologies(t *testing.T) {
	tests := map[string]struct {
		nodeSelector map[string][]string
		storageClass *storagev1.StorageClass
		expectErr    bool
	}{
		"Build with valid NodeSelector": {
			nodeSelector: map[string][]string{
				"kubernetes.io/hostname": {"alpha", "bravo", "charlie"},
			},
			storageClass: &storagev1.StorageClass{},
			expectErr:    false,
		},
		"Build with invalid/empty NodeSelector": {
			nodeSelector: map[string][]string{},
			storageClass: &storagev1.StorageClass{},
			expectErr:    true,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			opt := WithAllowedTopologies(mock.nodeSelector)
			err := opt(mock.storageClass)

			if mock.expectErr && err == nil {
				t.Fatal("Test '" + name + "' failed: expected error to not be nil.")
			}
			if !mock.expectErr && err != nil {
				t.Fatal("Test '" + name + "' failed: expected error to be nil.")
			}
		})
	}
}

func TestBuildWithNodeAffinityLabel(t *testing.T) {
	tests := map[string]struct {
		nodeAffinityLabel []string
		storageClass      *storagev1.StorageClass
		expectErr         bool
	}{
		"Build with valid NodeAffinityLabel": {
			nodeAffinityLabel: []string{"openebs.io/my-fav-node"},
			storageClass:      &storagev1.StorageClass{},
			expectErr:         false,
		},
		"Build with invalid/empty NodeAffinityLabel": {
			nodeAffinityLabel: []string{},
			storageClass:      &storagev1.StorageClass{},
			expectErr:         true,
		},
		"Build with valid '" + string(mconfig.CASTypeKey) + "' annotation": {
			nodeAffinityLabel: []string{"openebs.io/my-fav-node"},
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): localPVcasTypeValue,
					},
				},
			},
			expectErr: false,
		},
		"Build with invalid '" + string(mconfig.CASTypeKey) + "' annotation": {
			nodeAffinityLabel: []string{"openebs.io/my-fav-node"},
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): fakeCASTypeValue,
					},
				},
			},
			expectErr: true,
		},
		"Build with empty '" + string(mconfig.CASTypeKey) + "' annotation": {
			nodeAffinityLabel: []string{"openebs.io/my-fav-node"},
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): "",
					},
				},
			},
			expectErr: false,
		},
		"Build with valid '" + string(mconfig.CASConfigKey) + "' annotation": {
			nodeAffinityLabel: []string{"openebs.io/my-fav-node"},
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): mockHostpathConfig,
					},
				},
			},
			expectErr: false,
		},
		"Build with invalid '" + string(mconfig.CASConfigKey) + "' annotation": {
			nodeAffinityLabel: []string{"openebs.io/my-fav-node"},
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): fakeCASConfigValue,
					},
				},
			},
			expectErr: true,
		},
		"Build with empty '" + string(mconfig.CASConfigKey) + "' annotation": {
			nodeAffinityLabel: []string{"openebs.io/my-fav-node"},
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): "",
					},
				},
			},
			expectErr: false,
		},
		"Build with valid Provisioner name": {
			nodeAffinityLabel: []string{"openebs.io/my-fav-node"},
			storageClass: &storagev1.StorageClass{
				Provisioner: localPVprovisionerName,
			},
			expectErr: false,
		},
		"Build with invalid Provisioner name": {
			nodeAffinityLabel: []string{"openebs.io/my-fav-node"},
			storageClass: &storagev1.StorageClass{
				Provisioner: fakeProvisionerName,
			},
			expectErr: true,
		},
		"Build with empty Provisioner name": {
			nodeAffinityLabel: []string{"openebs.io/my-fav-node"},
			storageClass: &storagev1.StorageClass{
				Provisioner: "",
			},
			expectErr: false,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			opt := WithNodeAffinityLabels(mock.nodeAffinityLabel)
			err := opt(mock.storageClass)

			if mock.expectErr && err == nil {
				t.Fatal("Test '" + name + "' failed: expected error to not be nil.")
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test '"+name+"' failed: expected error to be nil. Error: {%v}", err)
			}
		})
	}
}

func TestBuildWithFSType(t *testing.T) {
	tests := map[string]struct {
		fstype       string
		storageClass *storagev1.StorageClass
		expectErr    bool
	}{
		"Build with valid 'xfs' FSType": {
			fstype:       "xfs",
			storageClass: &storagev1.StorageClass{},
			expectErr:    false,
		},
		"Build with valid 'ext4' FSType": {
			fstype:       "ext4",
			storageClass: &storagev1.StorageClass{},
			expectErr:    false,
		},
		"Build with empty FSType": {
			fstype:       "",
			storageClass: &storagev1.StorageClass{},
			expectErr:    true,
		},
		"Build with valid '" + string(mconfig.CASTypeKey) + "' annotation": {
			fstype: "xfs",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): localPVcasTypeValue,
					},
				},
			},
			expectErr: false,
		},
		"Build with invalid '" + string(mconfig.CASTypeKey) + "' annotation": {
			fstype: "xfs",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): fakeCASTypeValue,
					},
				},
			},
			expectErr: true,
		},
		"Build with empty '" + string(mconfig.CASTypeKey) + "' annotation": {
			fstype: "xfs",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): "",
					},
				},
			},
			expectErr: false,
		},
		"Build with valid '" + string(mconfig.CASConfigKey) + "' annotation": {
			fstype: "xfs",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): mockDeviceConfig,
					},
				},
			},
			expectErr: false,
		},
		"Build with invalid '" + string(mconfig.CASConfigKey) + "' annotation": {
			fstype: "xfs",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): fakeCASConfigValue,
					},
				},
			},
			expectErr: true,
		},
		"Build with empty '" + string(mconfig.CASConfigKey) + "' annotation": {
			fstype: "xfs",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): "",
					},
				},
			},
			expectErr: false,
		},
		"Build with valid Provisioner name": {
			fstype: "xfs",
			storageClass: &storagev1.StorageClass{
				Provisioner: localPVprovisionerName,
			},
			expectErr: false,
		},
		"Build with invalid Provisioner name": {
			fstype: "xfs",
			storageClass: &storagev1.StorageClass{
				Provisioner: fakeProvisionerName,
			},
			expectErr: true,
		},
		"Build with empty Provisioner name": {
			fstype: "xfs",
			storageClass: &storagev1.StorageClass{
				Provisioner: "",
			},
			expectErr: false,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			opt := WithFSType(mock.fstype)
			err := opt(mock.storageClass)

			if mock.expectErr && err == nil {
				t.Fatal("Test '" + name + "' failed: expected error to not be nil.")
			}
			if !mock.expectErr && err != nil {
				t.Fatal("Test '" + name + "' failed: expected error to be nil.")
			}
		})
	}
}

func TestBuildWithBlockDeviceTag(t *testing.T) {
	tests := map[string]struct {
		bdtag        string
		storageClass *storagev1.StorageClass
		expectErr    bool
	}{
		"Build with valid BlockDeviceTag": {
			bdtag:        "openebs.io/my-fav-disk",
			storageClass: &storagev1.StorageClass{},
			expectErr:    false,
		},
		"Build with empty BlockDeviceTag": {
			bdtag:        "",
			storageClass: &storagev1.StorageClass{},
			expectErr:    true,
		},
		"Build with valid '" + string(mconfig.CASTypeKey) + "' annotation": {
			bdtag: "openebs.io/my-fav-disk",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): localPVcasTypeValue,
					},
				},
			},
			expectErr: false,
		},
		"Build with invalid '" + string(mconfig.CASTypeKey) + "' annotation": {
			bdtag: "openebs.io/my-fav-disk",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): fakeCASTypeValue,
					},
				},
			},
			expectErr: true,
		},
		"Build with empty '" + string(mconfig.CASTypeKey) + "' annotation": {
			bdtag: "openebs.io/my-fav-disk",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASTypeKey): "",
					},
				},
			},
			expectErr: false,
		},
		"Build with valid '" + string(mconfig.CASConfigKey) + "' annotation": {
			bdtag: "openebs.io/my-fav-disk",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): mockDeviceConfig,
					},
				},
			},
			expectErr: false,
		},
		"Build with invalid '" + string(mconfig.CASConfigKey) + "' annotation": {
			bdtag: "openebs.io/my-fav-disk",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): fakeCASConfigValue,
					},
				},
			},
			expectErr: true,
		},
		"Build with empty '" + string(mconfig.CASConfigKey) + "' annotation": {
			bdtag: "openebs.io/my-fav-disk",
			storageClass: &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						string(mconfig.CASConfigKey): "",
					},
				},
			},
			expectErr: false,
		},
		"Build with valid Provisioner name": {
			bdtag: "openebs.io/my-fav-disk",
			storageClass: &storagev1.StorageClass{
				Provisioner: localPVprovisionerName,
			},
			expectErr: false,
		},
		"Build with invalid Provisioner name": {
			bdtag: "openebs.io/my-fav-disk",
			storageClass: &storagev1.StorageClass{
				Provisioner: fakeProvisionerName,
			},
			expectErr: true,
		},
		"Build with empty Provisioner name": {
			bdtag: "openebs.io/my-fav-disk",
			storageClass: &storagev1.StorageClass{
				Provisioner: "",
			},
			expectErr: false,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			opt := WithBlockDeviceTag(mock.bdtag)
			err := opt(mock.storageClass)

			if mock.expectErr && err == nil {
				t.Fatal("Test '" + name + "' failed: expected error to not be nil.")
			}
			if !mock.expectErr && err != nil {
				t.Fatal("Test '" + name + "' failed: expected error to be nil.")
			}
		})
	}
}
