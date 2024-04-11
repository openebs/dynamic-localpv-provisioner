# Quickstart

## Prerequisites

A Kubernetes cluster with Kubernetes v1.23 or above. 

For more platform-specific installation instructions, [click here](./installation/platforms/).

## Install using Helm chart
Install OpenEBS LocalPV Hostpath using the openebs helm chart. Sample command:
```console
#helm repo add openebs https://openebs.github.io/openebs
#helm repo update
helm install openebs openebs/openebs -n openebs --create-namespace
```
	
<details>
  <summary>Click here for configuration options.</summary>
  1. Install OpenEBS LocalPV Hostpath Provisioner with a custom hostpath directory. 
     This will change the `BasePath` value for the 'openebs-hostpath' StorageClass.
```console
helm install openebs openebs/openebs -n openebs --create-namespace \
	--set localpv-provisioner.hostpathClass.basePath=<custom-hostpath>
```
</details>

[Click here](https://github.com/openebs/openebs/tree/HEAD/charts) for detailed instructions on using the Helm chart.

You are ready to provision LocalPV volumes once the pods in 'openebs' namespace report RUNNING status.
```console
$ kubectl get pods -n openebs -l openebs.io/component-name=openebs-localpv-provisioner

NAME                                            READY   STATUS    RESTARTS   AGE
openebs-localpv-provisioner-6599766b76-kg5z9    1/1     Running   0          67s
```

## Provisioning LocalPV Hostpath Persistent Volume

You can provision LocalPV hostpath StorageType volumes dynamically using the default `openebs-hostpath` StorageClass.

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

For more advanced tutorials, visit [./tutorials/hostpath](./tutorials/hostpath).

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
       - 'date >> /mnt/data/date.txt; hostname >> /mnt/data/hostname.txt; sync; sleep 5; sync; tail -f /dev/null;'
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
