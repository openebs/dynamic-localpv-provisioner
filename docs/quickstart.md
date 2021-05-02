# Quickstart

## Prerequisites

A Kubernetes cluster with Kubernetes v1.19 or above is required. 

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
  2. Install OpenEBS Dynamic LocalPV Provisioner for Hostpath volumes only
```console
helm install openebs-localpv openebs-localpv/localpv-provisioner -n openebs --create-namespace \
	--set openebsNDM.enabled=false \
	--set deviceClass.enabled=false
```
  3. Install OpenEBS Dynamic LocalPV Provisioner with a custom hostpath directory. 
     This will change the `BasePath` value for the 'openebs-hostpath' StorageClass.
```console
helm install openebs-localpv openebs-localpv/localpv-provisioner -n openebs --create-namespace \
	--set hostpathClass.basePath=<custom-hostpath>
```
</details>

[Click here](https://github.com/openebs/dynamic-localpv-provisioner/blob/master/deploy/helm/charts/README.md) for detailed instructions on using the Helm chart.

## Install using operator YAML
Install the OpenEBS Dynamic LocalPV Provisioner using the following command:
```console
kubectl apply -f https://openebs.github.io/charts/openebs-operator-lite.yaml -f https://openebs.github.io/charts/openebs-lite-sc.yaml
```

You are ready to provision LocalPV volumes once the pods in 'openebs' namespace report RUNNING status.
```console
$ kubectl get pods -n openebs

NAME                                           READY   STATUS    RESTARTS   AGE
openebs-localpv-provisioner-5696c4f884-mvfvz   1/1     Running   0          7s
openebs-ndm-ctn5d                              1/1     Running   0          8s
openebs-ndm-lpf86                              1/1     Running   0          8s
openebs-ndm-operator-6b86bbc48-7lf7r           1/1     Running   0          8s
openebs-ndm-pqr2v                              1/1     Running   0          8s
```

## Provisioning LocalPV Hostpath Persistent Volume

You can provision LocalPV Hostpath volumes dynamically using the default `openebs-hostpath` StorageClass.

<details>
  <summary>Click here if you want to configure your own custom StorageClass.</summary>

  ```yaml
  # This is a custom StorageClass template
  # Uncomment config options as desired
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: custom-hostpath
    annotations:
      ## Use this annotation to set this StorageClass by default
      # storageclass.kubernetes.io/is-default-class: true
      openebs.io/cas-type: local
      cas.openebs.io/config: |
        - name: StorageType
          value: hostpath
      # - name: BasePath     # Use this to set a custom 
      #   value: /mnt/data   # hostpath directory
      # - name: NodeAffinityLabel                   # Use this to set a custom 
      #   value: "openebs.io/custom-node-unique-id" # label for node selection
  provisioner: openebs.io/local
  reclaimPolicy: Delete
  ## It is necessary to have volumeBindingMode as WaitForFirstConsumer
  volumeBindingMode: WaitForFirstConsumer
  ```
</details>

Create a PVC with the StorageClass.
```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: localpv-vol
spec:
  ## Change this name if you are using a custom StorageClass
  storageClassName: openebs-hostpath
  accessModes: ["ReadWriteOnce"]
  resources:
    requests:
      ## Set capacity here
      storage: 5Gi
```
The PVC will be in 'Pending' state until the volume is mounted.
```console
$ kubectl get pvc

NAME          STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS       AGE
localpv-vol   Pending                                      openebs-hostpath   21s
```


## Provisioning LocalPV Device Persistent Volume

You must have NDM installed to be able to use LocalPV Device. Use the following command to check if NDM pods are present:
```console
$ kubectl -n openebs get pods  -l 'openebs.io/component-name in (ndm, ndm-operator)'

NAME                                    READY   STATUS    RESTARTS   AGE
openebs-ndm-gctb7                       1/1     Running   0          6d7h
openebs-ndm-sfczv                       1/1     Running   0          6d7h
openebs-ndm-vgdnv                       1/1     Running   0          6d6h
openebs-ndm-operator-86b6dd687d-4lmpl   1/1     Running   0          6d7h
```

You can provision LocalPV Hostpath volumes dynamically using the default `openebs-device` StorageClass.

<details>
  <summary>Click here if you want to configure your own custom StorageClass.</summary>

  ```yaml
  # This is a custom StorageClass template
  # Uncomment config options as desired
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: custom-device
    annotations:
      ## Use this annotation to set this StorageClass by default
      # storageclass.kubernetes.io/is-default-class: true
      openebs.io/cas-type: local
      cas.openebs.io/config: |
        - name: StorageType
          value: device
      # - name: FSType  # Use this to set the filesystem
      #   value: xfs    # type. Default is ext4.
      # - name: BlockDeviceTag  # Only blockdevices with the label 
      #   value: "mongo"        # openebs.io/block-device-tag=mongo will be used
  provisioner: openebs.io/local
  reclaimPolicy: Delete
  ## It is necessary to have volumeBindingMode as WaitForFirstConsumer
  volumeBindingMode: WaitForFirstConsumer
  ## Match labels in allowedTopologies to select nodes for volume provisioning
  # allowedTopologies:
  # - matchLabelExpressions:
  #   - key: kubernetes.io/hostname
  #     values:
  #     - worker-1
  #     - worker-2
  ```
</details>

Create a PVC with the StorageClass.
```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: localpv-vol
spec:
  ## Change this name if you are using a custom StorageClass
  storageClassName: openebs-device
  accessModes: ["ReadWriteOnce"]
  ## You can also provision a raw block volume
  # volumeMode: Block
  volumeMode: Filesystem
  resources:
    requests:
      ## Set capacity here
      storage: 5Gi
```
The PVC will be in 'Pending' state until the volume is mounted/attached.
```console
$ kubectl get pvc

NAME          STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS     AGE
localpv-vol   Pending                                      openebs-device   21s
```
**Note**: LocalPV Device requires an Active and Unclaimed BlockDevice to be present on the node for provisioning to succeed. You may use allowedTopologies in your StorageClass to specify which nodes are usable.

## Mount the volume

Mount the volume to the application pod container. The PVC status will change to 'Bound' when the volume is mounted to a container. A sample BusyBox Pod template is given below.
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox
spec:
  containers:
  - command:
       - sh
       - -c
       - 'date >> /mnt/openebs-csi/date.txt; hostname >> /mnt/openebs-csi/hostname.txt; sync; sleep 5; sync; tail -f /dev/null;'
    image: busybox
    name: busybox
    volumeMounts:
    - mountPath: /mnt/data
      name: demo-vol
  volumes:
  - name: demo-vol
    persistentVolumeClaim:
      claimName: localpv-vol
```


Visit the official [OpenEBS documentation](https://docs.openebs.io) for more information.

Connect with the OpenEBS maintainers at the [Kubernetes Slack workspace](https://kubernetes.slack.com/messages/openebs). Visit [openebs.io/community](https://openebs.io/community) for details.
