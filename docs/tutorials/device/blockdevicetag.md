# Use selected Blockdevices

Blockdevice resources may be labelled with several device attributes or with custom labels. StorageClasses with the BlockDeviceSelectors parameter set with BD labels, will only use volumes provisioned on Blockdevices with said labels.

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
      - name: BlockDeviceSelectors
        data:
          ndm.io/driveType: "SSD"
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
```

**Note**: This does not influence scheduling of the application Pod based on blockdevice availability. The [openebs/device-localpv](https://github.com/openebs/device-localpv) project supports scheduling schemes bases on storage availability.
