# OpenEBS LocalPV Hostpath Modify XFS Quota

Follow these steps to change/remove XFS project quota enforcement from existing hostpath volumes:

## Step 1:

Make a note of the BasePath directory used for the hostpath volume. The default BasePath is '/var/openebs/local'. You can get the BasePath set on the StorageClass by executing the following command:
```console
$ kubectl describe sc <storageclass-name>
```

## Step 2:

Log in to the node where the volume exists. You can find the nodename (or any other identifying label selector) by describing the PV resource.
```console
$ kubectl get pvc --namespace demo

NAME              STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS       AGE
demo-vol-demo-0   Bound    pvc-0365904e-0add-45ec-9b4e-f4080929d6cd   2Gi        RWO            openebs-hostpath   21s


$ kubectl describe pv pvc-0365904e-0add-45ec-9b4e-f4080929d6cd

Name:              pvc-0365904e-0add-45ec-9b4e-f4080929d6cd
Labels:            openebs.io/cas-type=local-hostpath
Annotations:       pv.kubernetes.io/provisioned-by: openebs.io/local
Finalizers:        [kubernetes.io/pv-protection]
StorageClass:      openebs-hostpath
Status:            Bound
Claim:             demo/demo-vol-demo-0
Reclaim Policy:    Delete
Access Modes:      RWO
VolumeMode:        Filesystem
Capacity:          2Gi
Node Affinity:     
  Required Terms:  
    Term 0:        kubernetes.io/hostname in [storage-node-2]
Message:           
Source:
    Type:  LocalVolume (a persistent volume backed by local storage on a node)
    Path:  /var/openebs/local/pvc-0365904e-0add-45ec-9b4e-f4080929d6cd
Events:    <none>


$ kubectl get node -l 'kubernetes.io/hostname in (storage-node-2)'

NAME             STATUS   ROLES    AGE   VERSION
storage-node-2   Ready    worker   10m   v1.22.1
```

## Step 3:

You can change the soft and/or hard limit of an existing hostpath volume with XFS project quota enabled by following the steps below. If you'd like to remove the XFS project quota entirely, move on to the [Remove project](#remove-project) sub-section.

### Change limits

Execute the following commands on the node where the hostpath volume exists.

Make a note of the Project ID.
```command
$ sudo xfs_quota -x -c 'report -h' /var/openebs/local

Project quota on /var/openebs/local (/dev/nvme1n1)
                        Blocks              
Project ID   Used   Soft   Hard Warn/Grace   
---------- --------------------------------- 
#0              0      0      0  00 [------]
#1             1G   2.0G   2.0G  00 [------]
```

Modify the limits as desired using the following command. The values of bhard and bsoft must be in B/KB/MB/GB (not KiB/MiB/GiB). The sample command below sets the soft limit at 3G and the hard limit at 5G for project ID=1.
```command
$ sudo xfs_quota -x -c 'limit -p bsoft=3G bhard=5G 1' /var/openebs/local
```

Verify the changes:
```console
$ sudo xfs_quota -x -c 'report -h' /var/openebs/local

Project quota on /var/openebs/local (/dev/nvme1n1)
                        Blocks              
Project ID   Used   Soft   Hard Warn/Grace   
---------- --------------------------------- 
#0              0      0      0  00 [------]
#1             1G     3G     5G  00 [------]
```

### Remove project

Execute the following commands on the node where the hostpath volume exists.

Make a note of the Project ID.
```command
$ sudo xfs_quota -x -c 'report -h' /var/openebs/local

Project quota on /var/openebs/local (/dev/nvme1n1)
                        Blocks              
Project ID   Used   Soft   Hard Warn/Grace   
---------- --------------------------------- 
#0              0      0      0  00 [------]
#1             1G   2.0G   2.0G  00 [------]
```

Set the project limits to 0. The following sample command is for a project ID=1 at directory path '/var/openebs/local'.
```command
$ sudo xfs_quota -x -c 'limit -p bsoft=0 bhard=0 1' /var/openebs/local
```

Clear the directory tree from XFS project quota using the following command. The following sample command is for a project ID=1 at directory path '/var/openebs/local'.

```console
$ sudo xfs_quota 'project -C -p /var/openebs/local 1' /var/openebs/local
```

Verify the changes:
```console
$ sudo xfs_quota -x -c 'report -h' /var/openebs/local

Project quota on /var/openebs/local (/dev/nvme1n1)
                        Blocks              
Project ID   Used   Soft   Hard Warn/Grace   
---------- --------------------------------- 
#0             1G      0      0  00 [------]
```
