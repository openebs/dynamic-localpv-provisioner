# Quickstart

## Prerequisites

A Kubernetes cluster with Kubernetes v1.14 or above is required. 

<details>
  <summary>Click here if you are using RKE or Rancher 2.x.</summary>

  To use OpenEBS LocalPV Hostpath with an RKE/Rancher 2.x cluster, you will have to mount the hostpath directories to the kubelet containers. You can do this by editing the kubelet configuration section of your RKE/Rancher 2.x cluster and adding in the `extra_binds` (see below).

  **Note:** If you want to use a custom hostpath directory, then you will have to mount the custom directory's absolute path. See below for an example with the default hostpath directory.

  For an RKE cluster, you can add the `extra_binds` to your cluster.yml file and apply the changes using the `rke up` command.

  For a Rancher 2.x cluster, you can edit your cluster's configuration options and add the `extra_binds` there.

  ```yaml
  services:
    kubelet:
      extra_binds:
      # Default hostpath directory
      - /var/openebs/local:/var/openebs/local
  ```

  For more information, please go through the official Rancher documentaion -- [RKE - Kubernetes Configuration Options](https://rancher.com/docs/rke/latest/en/config-options/services/services-extras/#extra-binds), [RKE - Installation](https://rancher.com/docs/rke/latest/en/installation/#deploying-kubernetes-with-rke).
</details>

## Install using Helm chart
Install OpenEBS Dynamic LocalPV Provisioner using the localpv-provisioner helm chart. Sample command:
```console
# helm repo add openebs-localpv https://openebs.github.io/dynamic-localpv-provisioner
# helm repo update
helm install openebs-localpv openebs-localpv/localpv-provisioner -n openebs --create-namespace
```
	
<details>
  <summary>Click here for configuration options.</summary>

  1. Install OpenEBS Dynamic LocalPV Provisioner without NDM. 
     
     You may choose to exclude the NDM subchart from installation if...
     - you want to only use OpenEBS LocalPV Hostpath
     - you already have NDM installed. Check if NDM pods exist with the command `kubectl get pods -n openebs -l 'openebs.io/component-name in (ndm, ndm-operator)'`

```console
helm install openebs-localpv openebs-localpv/localpv-provisioner -n openebs --create-namespace \
	--set openebsNDM.enabled=false
```
  2. Install the OpenEBS Dynamic LocalPV Provisioner with a custom hostpath directory. 
     This will change the `BasePath` value for the 'openebs-hostpath' storageClass.
```console
helm install openebs-localpv openebs-localpv/localpv-provisioner -n openebs --create-namespace \
	--set hostpathClass.basePath=<custom-hostpath>
```
</details>

[Click here](https://openebs.github.io/dynamic-localpv-provisioner/) for detailed instructions on using the Helm chart.

## Install using operator YAML
Install the OpenEBS Dynamic LocalPV Provisioner using the following command:
```console
kubectl apply -f https://openebs.github.io/charts/openebs-operator-lite.yaml \
		-f https://openebs.github.io/charts/openebs-lite-sc.yaml
```
