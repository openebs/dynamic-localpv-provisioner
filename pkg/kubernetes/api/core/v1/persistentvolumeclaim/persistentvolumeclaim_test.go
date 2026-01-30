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
