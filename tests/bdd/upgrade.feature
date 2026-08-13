Feature: UPGRADE LOCALPV PROVISIONER

  The chart version in the working tree always names the version the branch is
  working towards, so the newest release below it is the one users would be
  upgrading from:
    | branch      | chart version    | upgrade from                        |
    | develop     | x.y.0-develop    | the last minor release, e.g. 4.5.1  |
    | release/x.y | x.y.z-prerelease | the last x.y patch, e.g. 4.5.1      |

  Background:
    Given a single-node Kubernetes cluster
    And the image of the localpv-provisioner chart under test is loaded into the cluster
    And no localpv-provisioner helm release is installed
    And the last released localpv-provisioner chart is installed with the following attributes:
      | release          | localpv-provisioner |
      | namespace        | openebs             |
      | analytics        | disabled            |
    Then the localpv-provisioner Pod should be running the released image

  Scenario: Upgrading to the localpv-provisioner chart in the working tree
    Given a PVC is created with the following attributes:
      | name               | pvc-upgrade-from                         |
      | storageClass       | openebs-hostpath                         |
      | accessModes        | ReadWriteOnce                            |
      | capacity           | 1Gi                                      |
    And a deployment with a busybox image is created with the following attributes:
      | name               | busybox-upgrade-from                     |
      | volumeMounts       | name: demo-vol1, mountPath: /mnt/store1  |
      | volumes            | name: demo-vol1, pvcName: pvc-upgrade-from |
    Then the Pod should be in Running state
    And a bound PV should be created
    And the application should be able to write to, and read from, the volume

    When the helm release is upgraded to the chart in the working tree with '--reset-then-reuse-values'
    Then the helm release metadata should report the chart version under test
    And the helm release status should be 'deployed'
    And the localpv-provisioner Deployment should be rolled out on the image of the chart under test
    And no localpv-provisioner Pod should be left running the released image

    When the upgrade is done
    Then the pre-upgrade PV should be unchanged, and still bound to its PVC
    And the data written before the upgrade should still be readable after the application Pod is recreated

    When a PVC is created with the following attributes:
      | name               | pvc-upgrade-to                           |
      | storageClass       | openebs-hostpath                         |
      | accessModes        | ReadWriteOnce                            |
      | capacity           | 1Gi                                      |
    And a deployment with a busybox image is created with the following attributes:
      | name               | busybox-upgrade-to                       |
      | volumeMounts       | name: demo-vol1, mountPath: /mnt/store1  |
      | volumes            | name: demo-vol1, pvcName: pvc-upgrade-to |
    Then a new bound PV should be created by the upgraded provisioner

    When the deployment consuming the pre-upgrade PVC is deleted
    And the pre-upgrade PVC is deleted
    Then the PVC should be deleted successfully
    And the PV should be deleted by the upgraded provisioner
    And the volume directory should be removed from the node
