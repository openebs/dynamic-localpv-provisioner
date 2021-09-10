# Use a different FSType

Device StorageType defaults to 'ext4' Filesystem type when creating a volume. You may also opt for 'XFS' filesystem when creating the volume. Setting the FSType parameter to 'XFS' is necessary in case the Blockdevice which is selected for volume provisioning, already has an XFS filesystem.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-device
  annotations:
    openebs.io/cas-type: local
    cas.openebs.io/config: |
      - name: StorageType
        value: "device"
      - name: FSType
        value: "xfs"
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
```
