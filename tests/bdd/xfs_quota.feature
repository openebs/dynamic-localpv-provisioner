Feature: Hostpath XFS Quota Local PV

  Scenario: HostPath XFS Quota Local PV with Unsupported Filesystem
    Given a sparse file "disk.img"
    And a loop device is created on top of disk.img

    When a StorageClass is created with the following attributes:
      | name                | sc-hp-xfs                  |
      | BasePath            | /path/to/hostpath          |
      | XFSQuotaEnabled     | "true"                     |
      | softLimit           | "20%"                      |
      | hardLimit           | "50%"                      |
      | provisionerName     | openebs.io/local           |
      | volumeBindingMode   | WaitForFirstConsumer       |
      | reclaimPolicy       | Delete                     |
    And a minix filesystem is written into the loop device
    And the minix filesystem is mounted with project quota enabled
    And a PVC "pvc-hp-xfs" is created with the StorageClass "sc-hp-xfs"
    And a Pod is created with PVC "pvc-hp-xfs"
    Then the Pod should be in pending state
    And the PVC should be in pending state

    When the Pod "busybox-hostpath" is deleted
    Then the Pod should be deleted successfully

    When the PVC "pvc-hp-xfs" is deleted
    Then the PVC should be deleted successfully

  Scenario: HostPath XFS Quota Local PV with XFS Filesystem
    Given a sparse file "disk.img"
    And a loop device is created on top of disk.img

    When a StorageClass is created with the following attributes:
      | name                | sc-hp-xfs                  |
      | BasePath            | /path/to/hostpath          |
      | XFSQuotaEnabled     | "true"                     |
      | provisionerName     | openebs.io/local           |
      | volumeBindingMode   | WaitForFirstConsumer       |
      | reclaimPolicy       | Delete                     |
    And the loop device is formatted with XFS filesystem
    And the xfs filesysten is mounted with project quota enabled
    And a PVC "pvc-hp-xfs" is created with the StorageClass "sc-hp-xfs"
    And a Pod is created with PVC "pvc-hp-xfs"
    Then the Pod should be up and running

    When data is written more than the quota limit into the hostpath volume
    Then the container process should not be able to write more than the enforced limit

    When the Pod consuming PVC "pvc-hp-xfs" is deleted
    Then the Pod should be deleted successfully

    When the PVC "pvc-hp-xfs" is deleted
    Then the PVC should be deleted successfully
    And the Provisioner should delete the PV
