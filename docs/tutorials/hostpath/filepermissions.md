# File permission tuning

Hostpath LocalPV will by default create folder with the following rights: `0777`. In some usecases, these rights are too wide and should be reduced.
As an important point, when using hostpath the underlying PV will be a localpath whichs allows kubelet to chown the folder based on the [fsGroup](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#configure-volume-permission-and-ownership-change-policy-for-pods))

We allow to set file permissions using:

```yaml
  #This is a custom StorageClass template
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: custom-hostpath
    annotations:
      openebs.io/cas-type: local
      cas.openebs.io/config: |
        - name: StorageType
          value: "hostpath"
        - name: BasePath
          value: "/var/openebs/local"
        - name: FilePermissions
          data:
            mode: "0770"
  provisioner: openebs.io/local
  reclaimPolicy: Delete
  #It is necessary to have volumeBindingMode as WaitForFirstConsumer
  volumeBindingMode: WaitForFirstConsumer
```

With such configuration the folder will be crated with `0770` rights for all the PVC using this storage class.

The same configuration is available at PVC level to have a more fined grained configuration capability (the Storage class configuration will always win against PVC one):

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: localpv-vol
  annotations:
    cas.openebs.io/config: |
      - name: FilePermissions
        data:
          mode: "0770"
spec:
  #Change this name if you are using a custom StorageClass
  storageClassName: openebs-hostpath
  accessModes: ["ReadWriteOnce"]
  resources:
    requests:
      #Set capacity here
      storage: 5Gi
```
