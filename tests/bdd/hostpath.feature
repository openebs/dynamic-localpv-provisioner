Feature: TEST HOSTPATH LOCAL PV

  Scenario: Creating and Deleting StorageClass, PVC, and Deployment with Busybox
    Given a hostpath provisioner is running
    When a StorageClass is create
    Then it should create a StorageClass with the following attributes:
      | name                | sc-hp                      |
      | BasePath            | /path/to/hostpath          |
      | provisionerName     | openebs.io/local           |
      | volumeBindingMode   | WaitForFirstConsumer       |
      | reclaimPolicy       | Delete                     |

    When a PVC with StorageClass "sc-hp" is created
    Then it should create a PVC with the following attributes:
      | name               | pvc-hp                      |
      | storageClass       | sc-hp                       |
      | accessModes        | ReadWriteOnce               |
      | capacity           | 2Gi                         |

    When a deployment with a busybox image is created
    Then it should create a deployment and a running pod with the following attributes:
      | name               | busybox-hostpath                        |
      | image              | busybox                                 |
      | command            | ["sleep", "3600"]                       |
      | volumeMounts       | name: demo-vol1, mountPath: /mnt/store1 |
      | volumes            | name: demo-vol1, pvcName: pvc-hp        |

    When the deployment is deleted
    Then it should not have any deployment or pod remaining

    When the PVC with the created StorageClass is deleted
    Then it should delete the PVC
    And it should have the Provisioner delete the PV
    And it should verify the PVC is deleted
