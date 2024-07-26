Feature: Volume Provisioning/De-provisioning with NodeAffinityLabels CAS-config on StorageClass

  Scenario: Volume provisioning/de-provisioning with custom NodeAffinityLabels CAS-config on StorageClass
    When a StorageClass with custom NodeAffinityLabels is created
    Then it should create a StorageClass with the following attributes:
      | name                | sc-nod-aff-lab                                                     |
      | BasePath            | /path/to/hostpath                                                  |
      | NodeAffinityLabels  | "kubernetes.io/hostname", "kubernetes.io/os", "kubernetes.io/arch" |
      | provisionerName     | openebs.io/local                                                   |
      | volumeBindingMode   | WaitForFirstConsumer                                               |
      | reclaimPolicy       | Delete                                                             |

    When a PVC "pvc-nod-aff-lab" is created with StorageClass "sc-nod-aff-lab"
    Then the PVC should be created successfully

    When a deployment with a busybox image is created with PVC "pvc-nod-aff-lab"
    Then the deployment should be created
    And a Pod should be up and running
    And a bound PV should be created
    And the SC NodeAffinityLabels CAS-config should be set correctly

    When the application Deployment deleted
    And the PVC is deleted
    Then The Pod should be deleted
    And the PV should be deleted
