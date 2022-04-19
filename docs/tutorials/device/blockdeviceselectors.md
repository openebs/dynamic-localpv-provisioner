# Use selected Blockdevices

Blockdevice resources may be labelled with several device attributes or with custom labels. StorageClasses with the BlockDeviceSelectors parameter set with BD labels, will only use volumes provisioned on Blockdevices with ALL of said labels.

Sample kubectl command to label Blockdevices:
```console
$ kubectl -n openebs label blockdevice <blockdevice-name> openebs.io/block-device-tag=mongo
```

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
          openebs.io/block-device-tag: <tag-value>
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
```

## NDM metaconfigs

You may enable `metaconfigs` in NDM's `openebs-ndm-config` ConfigMap in the `openebs` namespace, and use the node and device labels specified there. To enable metaconfigs, edit the NDM Configmap.

>**NOTE:** If you are using the helm chart, the name and the namespace of the ConfigMap object may change based on the release-name/fullname.

```console
$ kubectl -n openebs edit cm openebs-ndm-config
```
For more information, [click here](https://github.com/openebs/node-disk-manager/pull/618).

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
          ndm.io/fsType: "ext4"
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
```


**NOTE**: Using BlockDeviceSelectors does not influence scheduling of the application Pod based on Blockdevice availability. The [openebs/device-localpv](https://github.com/openebs/device-localpv) project supports scheduling schemes bases on storage availability.
