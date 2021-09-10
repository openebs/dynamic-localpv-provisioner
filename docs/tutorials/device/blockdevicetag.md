# Use selected Blockdevices

Blockdevice resources may be labelled with the label key `openebs.io/block-device-tag` and a unique value. StorageClasses with the BlockDeviceTag parameter set to the label's unique value, will only use volumes provisioned on Blockdevices with said label.

Sample kubectl command to label Blockdevices:
```console
$ kubectl -n openebs label blockdevice <blockdevice-name> openebs.io/block-device-tag=mongo
```

Sample StorageClass:
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-device
  annotations:
    openebs.io/cas-type: local
    cas.openebs.io/config: |
      - name: StorageType
        value: device
      - name: BlockDeviceTag
        value: "mongo"
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
```

**Note**: This does not influence scheduling of the application Pod based on blockdevice availability. The [openebs/device-localpv](https://github.com/openebs/device-localpv) project supports scheduling schemes bases on storage availability.
