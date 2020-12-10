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

package persistentvolume

import (
	corev1 "k8s.io/api/core/v1"
)

func fakeAPIPVList(pvNames []string) *corev1.PersistentVolumeList {
	if len(pvNames) == 0 {
		return nil
	}
	list := &corev1.PersistentVolumeList{}
	for _, name := range pvNames {
		pv := corev1.PersistentVolume{}
		pv.SetName(name)
		list.Items = append(list.Items, pv)
	}
	return list
}
