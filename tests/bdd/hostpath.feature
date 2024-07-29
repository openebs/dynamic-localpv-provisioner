Feature: TEST HOSTPATH LOCAL PV

  Scenario: Creating and Deleting StorageClass, PVC, and Deployment with Busybox
    Given a hostpath provisioner is running
    When a StorageClass is created with the following attributes:
      | name                | sc-hp                      |
      | BasePath            | /path/to/hostpath          |
      | provisionerName     | openebs.io/local           |
      | volumeBindingMode   | WaitForFirstConsumer       |
      | reclaimPolicy       | Delete                     |
    And a PVC is created with the following attributes:
      | name               | pvc-hp                      |
      | storageClass       | sc-hp                       |
      | accessModes        | ReadWriteOnce               |
      | capacity           | 2Gi                         |
    And a deployment with a busybox image is created with the following attributes:
      | name               | busybox-hostpath                        |
      | image              | busybox                                 |
      | command            | ["sleep", "3600"]                       |
      | volumeMounts       | name: demo-vol1, mountPath: /mnt/store1 |
      | volumes            | name: demo-vol1, pvcName: pvc-hp        |
    Then the Pod should be in Running state
    And a bound PV should be created

    When the deployment is deleted
    Then the deployment should not have any deployment or pod remaining

    When the PVC is deleted
    Then the PVC should be deleted successfully
    Then the PV should be deleted
