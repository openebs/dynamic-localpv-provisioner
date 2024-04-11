# OpenEBS LocalPV Hostpath Enable XFS Quota

### Prerequisites

1. The BasePath used by the provisioner should have XFS filesystem
2. All of the nodes used for hostpath storage must have [the 'xfsprogs' package installed](./prerequisites.md#install-the-xfsprogs-package).
3. The BasePath used by the provisioner [should be mounted with XFS project quotas enabled](./prerequisites.md#mount-filesystem-using-pquota-mount-option).

>**Note:** You may [use a loop device with XFS filesystem](./use-xfs-fs-with-loop-device.md) for XFS Quota. With a loop device based setup, you don't have to have an XFS root filesystem or external disks with XFS.

### Install the OpenEBS Dynamic LocalPV Provisioner
Install the OpenEBS LocalPV Hostpath Provisioner using the following given below. For more installation options, refer to [the quickstart guide](../../../quickstart.md).
```console
kubectl apply -f https://openebs.github.io/charts/openebs-operator-lite.yaml
```

Verify that pods in openebs namespace are running
```console
$ kubectl get pods -n openebs

NAME                                           READY   STATUS    RESTARTS       AGE
openebs-localpv-provisioner-6ddbd95d4d-htp7g   1/1     Running       0          7m12s
```

### Create StorageClass

Create a hostpath StorageClass with the XFSQuota config option.
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-hostpath-xfs
  annotations:
    openebs.io/cas-type: local
    cas.openebs.io/config: |
      - name: StorageType
        value: "hostpath"
      - name: BasePath
        value: "/var/openebs/local/"
      - name: XFSQuota
        enabled: "true"
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
reclaimPolicy: Delete
```
<details>
  <summary>Click here if you want to configure advanced XFSQuota options.</summary>
  
  ```yaml
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: openebs-hostpath-xfs
    annotations:
      openebs.io/cas-type: local
      cas.openebs.io/config: |
        - name: StorageType
          value: "hostpath"
        - name: BasePath
          value: "/var/openebs/local/"
        - name: XFSQuota
          enabled: "true"
          data:
            softLimitGrace: "0%"
            hardLimitGrace: "0%"
  provisioner: openebs.io/local
  volumeBindingMode: WaitForFirstConsumer
  reclaimPolicy: Delete
  ```
  
  `softLimitGrace` and `hardLimitGrace` with PV Storage Request will decide the soft limit and hard limit to be set beyond the storage capacity of the PV.
  
  The size of a limit will be as follows:<br>
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Size of PV storage request * ( 1 + LimitGrace% )

  Setting no value defaults to --> softLimitGrace: "0%" / hardLimitGrace: "0%"<br>
  This limits capacity to the value specified in the PV storage request.<br>
  For a PV with 100Gi capcity, this sets the soft and hard limits at 100Gi.<br>

  For a PV with 100Gi capacity, and values --> softLimitGrace: "90%" / hardLimitGrace: "100%"<br>
  This sets the soft limit at 190Gi and the hard limit at 200Gi.
  
  Anyone one of hardLimitGrace or softLimitGrace can also be used.<br>
  [Click here](https://man7.org/linux/man-pages/man8/xfs_quota.8.html#QUOTA_OVERVIEW) for detailed instructions about soft and hard limits.

</details><br>

### Create a PVC

Create a PVC using the StorageClass's name.
```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: local-hostpath-xfs
spec:
  storageClassName: openebs-hostpath-xfs
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
```
The PVC will be in 'Pending' state until the volume is mounted.
```console
$ kubectl get pvc

NAME                  STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS           AGE
local-hostpath-xfs    Pending                                      openebs-hostpath-xfs   21s
```

### Mount the Volume
Mount the volume to the application pod container. The PVC status will change to 'Bound' when the volume is mounted to a container and quota is applied. A sample BusyBox Pod template is given below.
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox
spec:
  volumes:
  - name: local-storage
    persistentVolumeClaim:
      claimName: local-hostpath-xfs
  containers:
  - name: busybox
    image: busybox
    command:
       - sh
       - -c
       - 'while true; do echo "`date` [`hostname`] Hello from OpenEBS Local PV." >> /mnt/store/greet.txt; sleep $(($RANDOM % 5 + 300)); done'
    volumeMounts:
    - mountPath: /mnt/store
      name: local-storage
```

Verify that the project quota is applied successfully.
```console
$ sudo xfs_quota -x -c 'report -h' /var/openebs/local/  

Project quota on /var/openebs/local (/dev/loop16)
                        Blocks              
Project ID   Used   Soft   Hard Warn/Grace   
---------- --------------------------------- 
#0              0      0      0  00 [------]
#1              0   5.7G   6.7G  00 [------]
```
### Limitation
* Resize of quota is not supported.