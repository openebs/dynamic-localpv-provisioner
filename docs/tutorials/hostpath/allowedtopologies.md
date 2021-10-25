# Scheduling based on Node label selector

The ['Allowed Topologies'](https://kubernetes.io/docs/concepts/storage/storage-classes/#allowed-topologies) feature allows you select the Nodes where the application Pods may be scheduled based on Node labels.

The nodes which are preferred for scheduling may be labelled using a unique label and key. Multiple such labels and keys per label may be specified. All of the selection criteria is AND-ed.

The following is a sample StorageClass which allows scheduling on nodes with the labels `kubernetes.io/hostname=worker-2`, `kubernetes.io/hostname=worker-3` and `kubernetes.io/hostname=worker-5`.

```yaml
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
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
allowedTopologies:
- matchLabelExpressions:
  - key: kubernetes.io/hostname
    values:
    - worker-2
    - worker-3
    - worker-5
```
