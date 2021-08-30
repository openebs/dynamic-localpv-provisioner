# OpenEBS LocalPV Hostpath Enable Quota

### Prerequisites

1. A Kubernetes cluster with Kubernetes v1.16 or above is required
2. All the nodes must have xfs utils installed
3. The base path used by the provisioner should be mounted with XFS project quotas enabled

### Setup for testing
Create a disk file which will be formatted as xfs fs, mounted as loopback device and exposed as the Basepath
```console
sudo dd if=/dev/zero of=xfs.32M bs=1 count=0 seek=32M
sudo mkfs -t xfs -q xfs.32M
sudo mount -o loop,rw xfs.32M -o pquota /var/openebs/local
```

### Check Basepath
Verify that filesystem is xfs
```console
stat -f -c %T /var/openebs/local
```
Filesystem of Basepath will be shown 
```console
xfs
```

Verify that project quotas are enabled
```console
xfs_quota -x -c state | grep 'Project quota state on /var/openebs/local' -A 2
```
Both Accounting and Enforcement should be ON
```console
Project quota state on /var/openebs/local (/dev/loop16)
  Accounting: ON
  Enforcement: ON
```

### Installing
Install the OpenEBS Dynamic LocalPV Provisioner using the following command:
```console
kubectl apply -f https://openebs.github.io/charts/openebs-operator-lite.yaml
```
Verify that pods in openebs namespace are running
```console
$ kubectl get pods -n openebs

NAME                                           READY   STATUS    RESTARTS       AGE
openebs-localpv-provisioner-6ddbd95d4d-htp7g   1/1     Running       0          7m12s
openebs-ndm-operator-849d89cb87-djct8          1/1     Running       0          7m12s
openebs-ndm-zd8lt                              1/1     Running       0          7m12s
```

### Deployment

#### 1. Create a Storage Class
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-hostpath-xfs
  annotations:
    openebs.io/cas-type: local
    cas.openebs.io/config: |
      #hostpath type will create a PV by 
      # creating a sub-directory under the
      # BASEPATH provided below.
      - name: StorageType
        value: "hostpath"
      #Specify the location (directory) where
      # where PV(volume) data will be saved. 
      # A sub-directory with pv-name will be 
      # created. When the volume is deleted, 
      # the PV sub-directory will be deleted.
      #Default value is /var/openebs/local
      - name: BasePath
        value: "/var/openebs/local/"
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
reclaimPolicy: Delete
parameters:
  enableXfsQuota: 'true'
  softLimitGrace: 20%
  hardLimitGrace: 40%
```
`softLimitGrace` and `hardLimitGrace` with PV Storage Request will decide the soft limit and hard limit to be set.
The size of a limit will be => size of PV Storage request * ( 1 + LimitGrace% )
Anyone of hard limit or soft limit can also be used
[Click here](https://man7.org/linux/man-pages/man8/xfs_quota.8.html#QUOTA_OVERVIEW) for detailed instructions about soft and hard limits.

#### 2. Create a PVC with storage class
```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: local-hostpath-pvc
spec:
  storageClassName: openebs-hostpath-xfs
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5G
```
The PVC will be in 'Pending' state until the volume is mounted.
```console
$ kubectl get pvc

NAME                  STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS          AGE
local-hostpath-pvc   Pending                                      openebs-hostpath-xfs   21s
```

#### 3. Mount the Volume
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
      claimName: local-hostpath-pvc
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
Verify that quota is applied successfully
```console
master@node$ sudo xfs_quota -x -c 'report -h' /var/openebs/local/  
Project quota on /var/openebs/local (/dev/loop16)
                        Blocks              
Project ID   Used   Soft   Hard Warn/Grace   
---------- --------------------------------- 
#0              0      0      0  00 [------]
#1              0   5.7G   6.7G  00 [------]
```
#### Limitation
* Resize of quota is not supported