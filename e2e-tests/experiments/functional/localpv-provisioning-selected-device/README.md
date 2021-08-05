## Experiment Metadata

| Type       | Description                                                  | Storage | Applications | K8s Platform |
| ---------- | ------------------------------------------------------------ | ------- | ------------ | ------------ |
| Functional | Ensure that local PV can be provisioned on selected block device | OpenEBS | Any          | Any          |

## Entry-Criteria

- K8s nodes should be ready.
- OpenEBS should be running.
- Unclaimed block device should be available

## Exit-Criteria

- Volume should be created on the labelled block device

## Procedure

- This functional test checks if the local pv can be provisioned on the selected block device. 

- This e2ebook accepts the parameters in form of job environmental variables and accordingly this test case can be run for two different test case type i.e. positive test case and negative test case.

For positive test case first we label the block devices and then provision the volume and verify the successful provisioning of the volume. For the negative test case first we provision the volume but claim for the persistent volume should be in pending state, and then we label the block device and verify the successful reconcilation of provisioning the volume is done. And later in both the cases we verify that block device is selected from only the list of tagged block devices.

1. Certain block device will be labelled with `openebs.io/block-device-tag=< tag-value >`

2. The `< tag-value >` can be passed to Local PV storage class via cas annotations. If the value is present, then Local PV device provisioner will set the following additional selector on the BDC:
  `openebs.io/block-device-tag=< tag-value >`

- The storage class spec will be built as follows:

  ```
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: openebs-device-mongo
    annotations:
      openebs.io/cas-type: local
      cas.openebs.io/config: |
        - name: StorageType
          value: "device"
        - name: BlockDeviceTag
          value: "<tag_value>"
  provisioner: openebs.io/local
  volumeBindingMode: WaitForFirstConsumer
  reclaimPolicy: Delete
  ```

- Upon using the above storage class, the PV should be provisioned on the tagged block device

- Finally, checking if the tagged BD alone is used by BDC being part of the volume.

## e2ebook Environment Variables

| Parameters    | Description                                                  |
| ------------- | ------------------------------------------------------------ |
| APP_NAMESPACE | Namespace where application and volume is deployed.          |
| PVC           | Name of PVC to be created                                    |
| OPERATOR_NS   | Namespace where OpenEBS is running                           |
| BD_TAG        | The label value to be used by the key `openebs.io/block-device-tag=< tag-value >` |
| TEST_CASE_TYPE| Run the test for `positive` or `negative` cases              |