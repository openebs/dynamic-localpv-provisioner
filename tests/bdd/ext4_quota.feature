Feature: Hostpath EXT4 Quota Local PV

  Scenario: HostPath EXT4 Quota Local PV with Unsupported Filesystem
    Given a sparse file "disk.img"
    And a loop device is created on top of disk.img

    When a StorageClass is created with the following attributes:
      | name                | sc-hp-ext4                 |
      | BasePath            | /path/to/hostpath          |
      | EXT4QuotaEnabled    | "true"                     |
      | softLimit           | "20%"                      |
      | hardLimit           | "50%"                      |
      | provisionerName     | openebs.io/local           |
      | volumeBindingMode   | WaitForFirstConsumer       |
      | reclaimPolicy       | Delete                     |
    And a minix filesystem is written into the loop device
    And the minix filesystem is mounted with project quota enabled
    And a PVC "pvc-hp-ext4" is created with the StorageClass "sc-hp-ext4"
    And a Pod is created with PVC "pvc-hp-ext4"
    Then the Pod should be in pending state
    And the PVC should be in pending state

    When the Pod "busybox-hostpath" is deleted
    Then the Pod should be deleted successfully

    When the PVC "pvc-hp-ext4" is deleted
    Then the PVC should be deleted successfully

  Scenario: HostPath EXT4 Quota Local PV with EXT4 Filesystem
    Given a sparse file "disk.img"
    And a loop device is created on top of disk.img

    When a StorageClass with valid EXT4 quota parameters is created
    Then it should create a StorageClass with the following attributes:
      | name                | sc-hp-ext4                 |
      | BasePath            | /path/to/hostpath          |
      | EXT4QuotaEnabled    | "true"                     |
      | provisionerName     | openebs.io/local           |
      | volumeBindingMode   | WaitForFirstConsumer       |
      | reclaimPolicy       | Delete                     |

    When the loop device is formatted with EXT4 filesystem
    And the ext4 filesysten is mounted with project quota enabled
    And a PVC "pvc-hp-ext4" is created with the StorageClass "sc-hp-ext4"
    And a Pod is created with PVC "pvc-hp-ext4"
    Then the Pod should be up and running

    When data is written more than the quota limit into the hostpath volume
    Then the container process should not be able to write more than the enforced limit

    When the Pod consuming PVC "pvc-hp-ext4" is deleted
    Then the Pod should be deleted successfully

    When the PVC "pvc-hp-ext4" is deleted
    Then the PVC should be deleted successfully
    And the Provisioner should delete the PV
