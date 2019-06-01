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

package app

import (
	//"fmt"
	//pvController "github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/controller"
	//mconfig "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	//"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"os"
	//"reflect"
	//"testing"
)

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
