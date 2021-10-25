# Quickstart

## Prerequisites

A Kubernetes cluster with Kubernetes v1.16 or above. 

For more platform-specific installation instructions, [click here](https://github.com/openebs/dynamic-localpv-provisioner/tree/develop/docs/installation/platforms/).

## Install using Helm chart
Install OpenEBS Dynamic LocalPV Provisioner using the openebs helm chart. Sample command:
```console
#helm repo add openebs https://openebs.github.io/charts
#helm repo update
helm install openebs openebs/openebs -n openebs --create-namespace \
	--set legacy.enabled=false
```
	
<details>
  <summary>Click here for configuration options.</summary>

  1. Install OpenEBS Dynamic LocalPV Provisioner without NDM. 
     
     You may choose to exclude the NDM subchart from installation if...
     - you want to only use OpenEBS LocalPV Hostpath
     - you already have NDM installed. Check if NDM pods exist with the command `kubectl get pods -n openebs -l 'openebs.io/component-name in (ndm, ndm-operator)'`

```console
helm install openebs openebs/openebs -n openebs --create-namespace \
	--set legacy.enabled=false \
	--set ndm.enabled=false \
	--set ndmOperator.enabled=false
```
  2. Install OpenEBS Dynamic LocalPV Provisioner for Hostpath volumes only
```console
helm install openebs openebs/openebs -n openebs --create-namespace \
	--set legacy.enabled=false \
	--set ndm.enabled=false \
	--set ndmOperator.enabled=false \
	--set localprovisioner.enableDeviceClass=false
```
  3. Install OpenEBS Dynamic LocalPV Provisioner with a custom hostpath directory. 
     This will change the `BasePath` value for the 'openebs-hostpath' StorageClass.
```console
helm install openebs openebs/openebs -n openebs --create-namespace \
	--set legacy.enabled=false \
	--set localprovisioner.basePath=<custom-hostpath>
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
  #This is a custom StorageClass template
  # Uncomment config options as desired
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: custom-hostpath
    annotations:
      #Use this annotation to set this StorageClass by default
      # storageclass.kubernetes.io/is-default-class: true
      openebs.io/cas-type: local
      cas.openebs.io/config: |
        - name: StorageType
          value: "hostpath"
       #Use this to set a custom
       # hostpath directory
       #- name: BasePath
       #  value: "/var/openebs/local"
  provisioner: openebs.io/local
  reclaimPolicy: Delete
  #It is necessary to have volumeBindingMode as WaitForFirstConsumer
  volumeBindingMode: WaitForFirstConsumer
  #Match labels in allowedTopologies to select nodes for volume provisioning
  # allowedTopologies:
  # - matchLabelExpressions:
  #   - key: kubernetes.io/hostname
  #     values:
  #     - worker-1
  #     - worker-2
  ```
</details><br>
For more advanced tutorials, visit [tutorials/hostpath](./tutorials/hostpath).

Create a PVC with the StorageClass.
```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: localpv-vol
spec:
  #Change this name if you are using a custom StorageClass
  storageClassName: openebs-hostpath
  accessModes: ["ReadWriteOnce"]
  resources:
    requests:
      #Set capacity here
      storage: 5Gi
```
The PVC will be in 'Pending' state until the volume is mounted.
```console
$ kubectl get pvc

NAME          STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS       AGE
localpv-vol   Pending                                      openebs-hostpath   21s
```
**Note**: The NodeAffinityLabel parameter does not influence where the application Pod will be scheduled. The NodeAffinityLabel parameter is to be used in cases where the value of the 'kubernetes.io/hostname' node label may change due to auto-scaling or similar behavior for the same node. In such cases, the administrator may choose to set a unique label which persists across node reboots and replacements.

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
  #This is a custom StorageClass template
  # Uncomment config options as desired
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: custom-device
    annotations:
      #Use this annotation to set this StorageClass by default
      # storageclass.kubernetes.io/is-default-class: true
      openebs.io/cas-type: local
      cas.openebs.io/config: |
        - name: StorageType
          value: "device"
       #Use this to set the filesystem
       # type. Default is 'ext4'.
       #- name: FSType
       #  value: "xfs"
       #Only blockdevices with the label
       # openebs.io/block-device-tag=mongo
       # will be used
       #- name: BlockDeviceTag
       #  value: "mongo"
  provisioner: openebs.io/local
  reclaimPolicy: Delete
  #It is necessary to have volumeBindingMode as WaitForFirstConsumer
  volumeBindingMode: WaitForFirstConsumer
  #Match labels in allowedTopologies to select nodes for volume provisioning
  # allowedTopologies:
  # - matchLabelExpressions:
  #   - key: kubernetes.io/hostname
  #     values:
  #     - worker-1
  #     - worker-2
  ```
</details><br>
For more advanced tutorials, visit [tutorials/device](./tutorials/device).

Create a PVC with the StorageClass.
```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: localpv-vol
spec:
  #Change this name if you are using a custom StorageClass
  storageClassName: openebs-device
  accessModes: ["ReadWriteOnce"]
  #You can also provision a raw block volume
  # volumeMode: Block
  volumeMode: Filesystem
  resources:
    requests:
      #Set capacity here
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


Visit the official [OpenEBS documentation](https://openebs.io/docs/) for more information.

Connect with the OpenEBS maintainers at the [Kubernetes Slack workspace](https://kubernetes.slack.com/messages/openebs). Visit [openebs.io/community](https://openebs.io/community) for details.
