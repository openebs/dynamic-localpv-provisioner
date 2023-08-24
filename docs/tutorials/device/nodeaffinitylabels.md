# Use Node-Selector labels

Device LocalPV uses the Kubernetes Node label(s) (For example: `kubernetes.io/hostname=<node-name>`) to uniquely identify a node.

In some cases, this label (`hostname`) is not unique across all the nodes in the cluster. This was seen on clusters provisioned with the cloud provider [Bosh](https://bosh.io/docs/) across different fault domains.

A unique Node label (or set of labels) may be used instead of the above mentioned Kubernetes default label to uniquely identify a node. This label(s) may be set by you, the administrator.
This label(s) can be set when defining a StorageClass. One such sample StorageClass is given below...

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-hostpath
  annotations:
    openebs.io/cas-type: local
    cas.openebs.io/config: |
      - name: StorageType
        value: "device"
      - name: NodeAffinityLabels
        list:
          - "openebs.io/custom-node-unique-id"
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
```

**NOTE**: Using NodeAffinityLabels does not influence scheduling of the application Pod. You may use [allowedTopologies](./allowedtopologies.md) for that.
