# Quickstart

## Prerequisites

A Kubernetes cluster with Kubernetes v1.14 or above is required.

## Install using Helm chart
Install OpenEBS Dynamic LocalPV Provisioner using the localpv-provisioner helm chart. Sample command:
```console
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

[Click here](https://openebs.github.io/dynamic-localpv-provisioner/) for detailed instructions.

## Install using operator YAML
Install the OpenEBS Dynamic LocalPV Provisioner using the following command:
```console
kubectl apply -f https://openebs.github.io/charts/openebs-operator-lite.yaml \
		-f https://openebs.github.io/charts/openebs-lite-sc.yaml
```
