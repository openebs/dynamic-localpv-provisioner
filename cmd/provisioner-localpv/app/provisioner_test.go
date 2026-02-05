package app

import (
	"context"
	"errors"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	pvController "sigs.k8s.io/sig-storage-lib-external-provisioner/v13/controller"
)

func TestProvision(t *testing.T) {
	// Create test nodes that will be fetched by the provisioner
	nodeWithHostname := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-one",
			Labels: map[string]string{
				k8sNodeLabelKeyHostname: "127.0.0.1",
			},
		},
	}
	nodeWithoutHostname := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-one",
		},
	}

	testsCases := map[string]struct {
		opts              pvController.ProvisionOptions
		errorMessage      string
		GetVolumeConfigFn GetVolumeConfigFn
		nodes             []runtime.Object // Nodes to add to fake client
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
				SelectedNodeName: "",
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
				SelectedNodeName: "node-one",
			},
			nodes:        []runtime.Object{nodeWithoutHostname},
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
				SelectedNodeName: "node-one",
			},
			nodes: []runtime.Object{nodeWithHostname},
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
				SelectedNodeName: "node-one",
			},
			nodes: []runtime.Object{nodeWithHostname},
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
				SelectedNodeName: "node-one",
			},
			nodes: []runtime.Object{nodeWithHostname},
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
				SelectedNodeName: "node-one",
			},
			nodes: []runtime.Object{nodeWithHostname},
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
				SelectedNodeName: "node-one",
			},
			nodes: []runtime.Object{nodeWithHostname},
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
				SelectedNodeName: "node-one",
			},
			nodes: []runtime.Object{nodeWithHostname},
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
			// Create fake client with test nodes
			var fakeClient kubernetes.Interface
			if len(tc.nodes) > 0 {
				fakeClient = fake.NewSimpleClientset(tc.nodes...)
			} else {
				fakeClient = fake.NewSimpleClientset()
			}

			p := Provisioner{
				kubeClient:      fakeClient,
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
