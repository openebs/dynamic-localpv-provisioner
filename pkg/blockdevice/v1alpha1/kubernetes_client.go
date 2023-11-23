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

package v1alpha1

import (
	"context"

	ndm "github.com/openebs/api/v3/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Get is Kubernetes client implementation to get block device
func (k *KubernetesClient) Get(name string, opts metav1.GetOptions) (*BlockDevice, error) {
	bd, err := k.Clientset.OpenebsV1alpha1().BlockDevices(k.Namespace).
		Get(context.TODO(), name, opts)

	return &BlockDevice{bd, nil}, err
}

// List is kubernetes client implementation to list block device
func (k *KubernetesClient) List(opts metav1.ListOptions) (*BlockDeviceList, error) {
	bdl, err := k.Clientset.OpenebsV1alpha1().BlockDevices(k.Namespace).
		List(context.TODO(), opts)
	return &BlockDeviceList{bdl, nil}, err
}

// Create is kubernetes client implementation to create block device
func (k *KubernetesClient) Create(bdObj *ndm.BlockDevice) (*BlockDevice, error) {
	bdObj, err := k.Clientset.OpenebsV1alpha1().BlockDevices(k.Namespace).
		Create(context.TODO(), bdObj, metav1.CreateOptions{})
	return &BlockDevice{bdObj, nil}, err
}
