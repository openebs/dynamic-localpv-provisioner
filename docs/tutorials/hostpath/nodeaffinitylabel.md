# Use a unique Node-Selector label

Hostpath LocalPV uses the Kubernetes Node label `kubernetes.io/hostname=<node-name>` to uniquely identifly a node.

In some cases, this label (`hostname`) is not unique across all the nodes in the cluster. This was seen on clusters provisioned with [Bosh](https://bosh.io/docs/) across different fault domains.

A unique Node label may be used instead of the above mentioned Kubernetes default label to uniquely identify a node. This label may be set by you, the administrator.
This label can be set when defining a StorageClass. One such sample StorageClass is given below...

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-hostpath
  annotations:
    openebs.io/cas-type: local
    cas.openebs.io/config: |
      - name: StorageType
      - value: "hostpath"
      - name: NodeAffinityLabel
        value: "openebs.io/custom-node-unique-id"
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
```
