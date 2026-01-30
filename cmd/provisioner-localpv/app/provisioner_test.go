package app

import (
	"context"
	"errors"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pvController "sigs.k8s.io/sig-storage-lib-external-provisioner/v9/controller"
)

func TestProvision(t *testing.T) {
	testsCases := map[string]struct {
		opts              pvController.ProvisionOptions
		errorMessage      string
		GetVolumeConfigFn GetVolumeConfigFn
	}{
		"clone volume": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: &v1.TypedLocalObjectReference{
							Kind: PVCKind,
							Name: "source",
						},
					},
				},
			},
			errorMessage: "clone feature not supported by this provisioner",
		},
		"clone volume with no name": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-pvc",
					},
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: &v1.TypedLocalObjectReference{
							Kind: PVCKind,
						},
					},
				},
			},
			errorMessage: "dataSource name not found for PVC `my-pvc`",
		},
		"snapshot volume": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: &v1.TypedLocalObjectReference{
							APIGroup: func(name string) *string {
								return &name
							}(SnapshotAPIGroup),
							Kind: SnapshotKind,
							Name: "source",
						},
					},
				},
			},
			errorMessage: "datasource `VolumeSnapshot` of group `snapshot.storage.k8s.io` is not handled by the provisioner",
		},
		"snapshot volume with no name": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-pvc",
					},
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: &v1.TypedLocalObjectReference{
							APIGroup: func(name string) *string {
								return &name
							}(SnapshotAPIGroup),
							Kind: SnapshotKind,
						},
					},
				},
			},
			errorMessage: "dataSource name not found for PVC `my-pvc`",
		},
		"populator volume": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: &v1.TypedLocalObjectReference{
							APIGroup: func(name string) *string {
								return &name
							}("example.io"),
							Kind: "DemoPopulator",
							Name: "source",
						},
					},
				},
			},
			errorMessage: "datasource `DemoPopulator` of group `example.io` is not handled by the provisioner",
		},
		"populator volume with no name": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-pvc",
					},
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: &v1.TypedLocalObjectReference{
							APIGroup: func(name string) *string {
								return &name
							}("example.io"),
							Kind: "DemoPopulator",
						},
					},
				},
			},
			errorMessage: "dataSource name not found for PVC `my-pvc`",
		},
		"label selector match labels": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: nil,
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app.kubernetes.io/name": "my-app",
							},
						},
					},
				},
			},
			errorMessage: "claim.Spec.Selector is not supported",
		},
		"label selector match expressions": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: nil,
						Selector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "app.kubernetes.io/name",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"my-app"},
								},
							},
						},
					},
				},
			},
			errorMessage: "claim.Spec.Selector is not supported",
		},
		"ReadWriteMany access mode": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: nil,
						Selector:   &metav1.LabelSelector{},
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
							v1.ReadWriteMany,
						},
					},
				},
			},
			errorMessage: "Only support ReadWriteOnce access mode",
		},
		"empty selected node": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: nil,
						Selector: &metav1.LabelSelector{
							MatchLabels:      map[string]string{},
							MatchExpressions: []metav1.LabelSelectorRequirement{},
						},
					},
				},
			},
			errorMessage: "configuration error, no node was specified",
		},
		"hostname missing": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: nil,
						Selector: &metav1.LabelSelector{
							MatchLabels:      map[string]string{},
							MatchExpressions: []metav1.LabelSelectorRequirement{},
						},
					},
				},
				SelectedNode: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-one",
					},
				},
			},
			errorMessage: "configuration error, node{node-one} hostname is empty",
		},
		"get volume func returns error": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: nil,
						Selector: &metav1.LabelSelector{
							MatchLabels:      map[string]string{},
							MatchExpressions: []metav1.LabelSelectorRequirement{},
						},
					},
				},
				SelectedNode: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-one",
						Labels: map[string]string{
							k8sNodeLabelKeyHostname: "127.0.0.1",
						},
					},
				},
			},
			GetVolumeConfigFn: func(ctx context.Context, pvName string, pvc *v1.PersistentVolumeClaim) (*VolumeConfig, error) {
				return nil, errors.New("an error occured")
			},
			errorMessage: "an error occured",
		},
		"storage type not device when using volumemode Block": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: nil,
						VolumeMode: func(p v1.PersistentVolumeMode) *v1.PersistentVolumeMode {
							return &p
						}(v1.PersistentVolumeBlock),
					},
				},
				SelectedNode: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-one",
						Labels: map[string]string{
							k8sNodeLabelKeyHostname: "127.0.0.1",
						},
					},
				},
			},
			GetVolumeConfigFn: func(ctx context.Context, pvName string, pvc *v1.PersistentVolumeClaim) (*VolumeConfig, error) {
				return &VolumeConfig{
					options: map[string]interface{}{},
				}, nil
			},
			errorMessage: "PV with BlockMode is not supported with StorageType hostpath",
		},
		"invalid storage type": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: nil,
						VolumeMode: func(p v1.PersistentVolumeMode) *v1.PersistentVolumeMode {
							return &p
						}(v1.PersistentVolumeFilesystem),
					},
				},
				SelectedNode: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-one",
						Labels: map[string]string{
							k8sNodeLabelKeyHostname: "127.0.0.1",
						},
					},
				},
			},
			GetVolumeConfigFn: func(ctx context.Context, pvName string, pvc *v1.PersistentVolumeClaim) (*VolumeConfig, error) {
				return &VolumeConfig{
					options: map[string]interface{}{
						KeyPVStorageType: map[string]string{
							"value":   "foo",
							"enabled": "true",
						},
					},
				}, nil
			},
			errorMessage: "PV with StorageType foo is not supported",
		},
		"error in get path": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: nil,
						VolumeMode: func(p v1.PersistentVolumeMode) *v1.PersistentVolumeMode {
							return &p
						}(v1.PersistentVolumeFilesystem),
					},
				},
				SelectedNode: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-one",
						Labels: map[string]string{
							k8sNodeLabelKeyHostname: "127.0.0.1",
						},
					},
				},
			},
			GetVolumeConfigFn: func(ctx context.Context, pvName string, pvc *v1.PersistentVolumeClaim) (*VolumeConfig, error) {
				return &VolumeConfig{
					options: map[string]interface{}{},
				}, nil
			},
			errorMessage: "failed to get path: base path is empty",
		},
		"error when validating path": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: nil,
						VolumeMode: func(p v1.PersistentVolumeMode) *v1.PersistentVolumeMode {
							return &p
						}(v1.PersistentVolumeFilesystem),
					},
				},
				SelectedNode: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-one",
						Labels: map[string]string{
							k8sNodeLabelKeyHostname: "127.0.0.1",
						},
					},
				},
			},
			GetVolumeConfigFn: func(ctx context.Context, pvName string, pvc *v1.PersistentVolumeClaim) (*VolumeConfig, error) {
				return &VolumeConfig{
					pvName: "my-pv",
					options: map[string]interface{}{
						KeyPVBasePath: map[string]string{
							"value":   "/",
							"enabled": "true",
						},
					},
				}, nil
			},
			errorMessage: "host path validation failed: [path should not be a root directory: //my-pv]",
		},
		"error creating init pod": {
			opts: pvController.ProvisionOptions{
				PVC: &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						DataSource: nil,
						VolumeMode: func(p v1.PersistentVolumeMode) *v1.PersistentVolumeMode {
							return &p
						}(v1.PersistentVolumeFilesystem),
					},
				},
				SelectedNode: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-one",
						Labels: map[string]string{
							k8sNodeLabelKeyHostname: "127.0.0.1",
						},
					},
				},
			},
			GetVolumeConfigFn: func(ctx context.Context, pvName string, pvc *v1.PersistentVolumeClaim) (*VolumeConfig, error) {
				return &VolumeConfig{
					pvName: "my-pv",
					options: map[string]interface{}{
						KeyPVBasePath: map[string]string{
							"value":   "/data/",
							"enabled": "true",
						},
					},
				}, nil
			},
			errorMessage: "invalid empty name or hostpath or hostname or service account name",
		},
	}

	for name, tc := range testsCases {
		name := name
		tc := tc
		t.Run(name, func(t *testing.T) {
			p := Provisioner{
				getVolumeConfig: tc.GetVolumeConfigFn,
			}
			_, _, err := p.Provision(context.TODO(), tc.opts)

			if tc.errorMessage != "" {
				if err == nil {
					t.Fatalf("expected error '%s' but received none", tc.errorMessage)
				} else if err.Error() != tc.errorMessage {
					t.Fatalf("expected error '%s' but got '%s'", tc.errorMessage, err.Error())
				}
			} else if tc.errorMessage == "" && err != nil {
				t.Fatalf("expected no error but got %s", err)
			}
		})
	}
}

func fakeDefaultConfigParser(path string, pvc *v1.PersistentVolumeClaim) (*VolumeConfig, error) {
	c := &VolumeConfig{
		pvName:  "pvName",
		pvcName: "pvcName",
		scName:  "scName",
		options: map[string]interface{}{
			KeyPVBasePath: map[string]string{
				"enabled": "true",
				"value":   "/var/openebs/local",
			},
		},
	}
	return c, nil
}

func fakeValidConfigParser(path string, pvc *v1.PersistentVolumeClaim) (*VolumeConfig, error) {
	c := &VolumeConfig{
		pvName:  "pvName",
		pvcName: "pvcName",
		scName:  "scName",
		options: map[string]interface{}{
			KeyPVBasePath: map[string]string{
				"enabled": "true",
				"value":   "/custom",
			},
		},
	}
	return c, nil
}

//func fakeInvalidConfigParser(path string, pvc *v1.PersistentVolumeClaim) (*VolumeConfig, error) {
//	return nil, fmt.Errorf("failed to read configuration for pvc %v", path)
//}

/*
//func (p *Provisioner) Provision(opts pvController.VolumeOptions) (*v1.PersistentVolume, error) {
func TestProvision(t *testing.T) {
	testCases := map[string]struct {
		pvOpts          pvController.VolumeOptions
		getVolumeConfig GetVolumeConfigFn
		expectValue     string
		expectError     bool
	}{
		"Default Base Path": {
			pvOpts: pvController.VolumeOptions{
				PVName: "pvName",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvcName",
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						Selector: nil,
					},
				},
				SelectedNode: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "selectednode",
					},
				},
			},
			getVolumeConfig: fakeDefaultConfigParser,
			expectValue:     "/var/openebs/local/pvName",
			expectError:     false,
		},
		"Custom Base Path": {
			pvOpts: pvController.VolumeOptions{
				PVName: "pvName",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvcName",
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						Selector: nil,
					},
				},
				SelectedNode: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "selectednode",
					},
				},
			},
			getVolumeConfig: fakeValidConfigParser,
			expectValue:     "/custom/pvName",
			expectError:     false,
		},
		"Selected Node is missing": {
			pvOpts: pvController.VolumeOptions{
				PVName: "pvName",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvcName",
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						Selector: nil,
					},
				},
				//SelectedNode: &v1.Node{
				//	ObjectMeta: metav1.ObjectMeta{
				//		Name: "selectednode",
				//	},
				//},
			},
			getVolumeConfig: fakeValidConfigParser,
			expectValue:     "/test/pvName",
			expectError:     true,
		},
	}

	for k, v := range testCases {
		v := v
		t.Run(k, func(t *testing.T) {
			p := &Provisioner{}
			p.getVolumeConfig = v.getVolumeConfig
			//p, _ := NewProvisioner(nil, nil)
			pv, err := p.Provision(v.pvOpts)

			if v.expectError && err != nil {
				//t.Errorf("expected to error, but got %v", pv)
				return
			}

			if v.expectError && err == nil {
				t.Errorf("expected to error, but got pv %v", pv)
				return
			}
			if !v.expectError && err != nil {
				t.Errorf("expected not to get pv, but got %v", err)
				return
			}
			if err == nil && pv == nil {
				t.Errorf("expected pv, but got nil")
				return
			}
			if err == nil && pv.Spec.Local == nil {
				t.Errorf("expected pv.Spec.HostPath, but got nil %v", pv)
				return
			}

			actualValue := pv.Spec.PersistentVolumeSource.Local.Path
			if !v.expectError && !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
		})
	}
}
*/
