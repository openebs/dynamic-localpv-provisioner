Feature: Volume Provisioning/De-provisioning with Additive and Conflicting CAS-configs on PVC and SC

  Scenario: Additive CAS-configs on PVC and SC
    When a StorageClass with is created with the following attributes:
      | name                | sc-additive-cas-config  |
      | BasePath            | /path/to/hostpath       |
      | provisionerName     | openebs.io/local        |
      | volumeBindingMode   | WaitForFirstConsumer    |
      | reclaimPolicy       | Delete                  |
    And a PVC "pvc-additive-cas-config" is created with the following attributes:
      | name               | pvc-additive-cas-config                  |
      | storageClass       | sc-hp                                    |
      | NodeAffinityLabels | "kubernetes.io/os", "kubernetes.io/arch" |
      | accessModes        | ReadWriteOnce                            |
      | capacity           | 2Gi                                      |
    And a Deployment is created with PVC "pvc-additive-cas-config"
    Then the Pod should be up and running
    And a bound PV should be created
    And the PVC NodeAffinityLabels CAS-configs should be set correctly on the PV

    When the application Deployment is deleted
    Then The Pod should be deleted

    When the PVC is deleted
    Then the PV should be deleted

  Scenario: Conflicting CAS-configs on PVC and SC
    When a StorageClass is created with the following attributes:
      | name                | sc-conflicting-cas-config |
      | BasePath            | /path/to/hostpath         |
      | NodeAffinityLabels  | "kubernetes.io/hostname"  |
      | provisionerName     | openebs.io/local          |
      | volumeBindingMode   | WaitForFirstConsumer      |
      | reclaimPolicy       | Delete                    |
    And a PVC "pvc-conflicting-cas-config" is created with the following attributes:
      | name               | pvc-conflicting-cas-config               |
      | storageClass       | sc-hp                                    |
      | NodeAffinityLabels | "kubernetes.io/os", "kubernetes.io/arch" |
      | accessModes        | ReadWriteOnce                            |
      | capacity           | 2Gi                                      |
    And a Deployment is created with PVC "pvc-conflicting-cas-config"
    Then a Pod should be up and running
    And a bound PV should be created
    And the SC NodeAffinityLabels CAS-config should be set correctly on the PV

    When the application Deployment deleted
    Then The Pod should be deleted

    When the PVC is deleted
    Then the PV should be deleted
