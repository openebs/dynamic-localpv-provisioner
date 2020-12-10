// Copyright © 2018-2020 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package persistentvolumeclaim

import (
	corev1 "k8s.io/api/core/v1"
)

func fakeAPIPVCList(pvcNames []string) *corev1.PersistentVolumeClaimList {
	if len(pvcNames) == 0 {
		return nil
	}
	list := &corev1.PersistentVolumeClaimList{}
	for _, name := range pvcNames {
		pvc := corev1.PersistentVolumeClaim{}
		pvc.SetName(name)
		list.Items = append(list.Items, pvc)
	}
	return list
}

func fakeAPIPVCListFromNameStatusMap(pvcs map[string]corev1.PersistentVolumeClaimPhase) *corev1.PersistentVolumeClaimList {
	if len(pvcs) == 0 {
		return nil
	}
	list := &corev1.PersistentVolumeClaimList{}
	for k, v := range pvcs {
		pvc := corev1.PersistentVolumeClaim{}
		pvc.SetName(k)
		pvc.Status.Phase = v
		list.Items = append(list.Items, pvc)
	}
	return list
}
