Feature: Volume Provisioning/De-provisioning with NodeAffinityLabels CAS-config on StorageClass

  Scenario: Volume provisioning/de-provisioning with custom NodeAffinityLabels CAS-config on StorageClass
    When a StorageClass is created with the following attributes:
      | name                | sc-nod-aff-lab                                                     |
      | BasePath            | /path/to/hostpath                                                  |
      | NodeAffinityLabels  | "kubernetes.io/hostname", "kubernetes.io/os", "kubernetes.io/arch" |
      | provisionerName     | openebs.io/local                                                   |
      | volumeBindingMode   | WaitForFirstConsumer                                               |
      | reclaimPolicy       | Delete                                                             |
    And a PVC "pvc-nod-aff-lab" is created with StorageClass "sc-nod-aff-lab"
    And a deployment with a busybox image is created with PVC "pvc-nod-aff-lab"
    Then a Pod should be up and running
    And a bound PV should be created
    And the SC NodeAffinityLabels CAS-config should be set correctly on the PV

    When the application Deployment is deleted
    Then The Pod should be deleted

    When the PVC is deleted
    Then the PV should be deleted
