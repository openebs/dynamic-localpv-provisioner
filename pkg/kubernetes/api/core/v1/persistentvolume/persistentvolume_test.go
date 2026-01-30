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
