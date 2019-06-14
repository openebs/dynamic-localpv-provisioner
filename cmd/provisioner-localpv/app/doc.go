/*
Copyright 2019 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

/*

Package app contains OpenEBS Dynamic Local PV provisioner

Provisioner is created using the external storage provisioner library:
https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner

Local PVs are an extension to hostpath volumes, but are more secure.
https://kubernetes.io/docs/concepts/policy/pod-security-policy/#volumes-and-file-systems

Local PVs are great in cases like:
- The Stateful Workload can take care of replicating the data across
  nodes to handle cases like a complete node (and/or its storage) failure.
- For long running Stateful Workloads, the Backup/Recovery is provided
  by Operators/tools that can make use the Workload mounts and do not
  require the capabilities to be available in the underlying storage. Or
  if the hostpaths are created on external storage like EBS/GPD, administrator
  have tools that can periodically take snapshots/backups.

OpenEBS Local PVs extends the capabilities provided by the Kubernetes Local PV
by making use of the OpenEBS Node Storage Device Manager (NDM), the significant
differences include:
- Users need not pre-format and mount the devices in the node.
- Supports Dynamic Local PVs - where the devices can be used by CAS solutions
  and also by applications. CAS solutions typically directly access a device.
  OpenEBS Local PV ease the management of storage devices to be used between
  CAS solutions (direct access) and applications (via PV), by making use of
  BlockDeviceClaims supported by OpenEBS NDM.
- Supports using hostpath as well for provisioning a Local PV. In fact in some
  cases, the Kubernetes nodes may have limited number of storage devices
  attached to the node and hostpath based Local PVs offer efficient management
  of the storage available on the node.

Inspiration:
------------
OpenEBS Local PV has been inspired by the prior work done by the following
the Kubernetes projects:
- https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/tree/master/examples/hostpath-provisioner
- https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner
- https://github.com/rancher/local-path-provisioner

How it works:
-------------
Step 1: Multiple Storage Classes can be created by the Kubernetes Administrator,
to specify the required type of OpenEBS Local PV to be used by an application.
A simple StorageClass looks like:

---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: hostpath
  annotations:
    #Define a new OpenEBS CAS Type called `local`
    #which indicates that Data is stored
    #directly onto hostpath. The hostpath can be:
    #- device (as block or mounted path)
    #- hostpath (sub directory on OS or mounted path)
    openebs.io/cas-type: local
    cas.openebs.io/config: |
      #- name: StorageType
      #  value: "device"
      # (Default)
      - name: StorageType
        value: "hostpath"
      # If the StorageType is hostpath, then BasePath
      # specifies the location where the volume sub-directory
      # should be created.
      # (Default)
      - name: BasePath
        value: "/var/openebs/local"
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
reclaimPolicy: Delete
---

Step 2: The application developers will request for storage via PVC as follows:
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc-hp
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: hostpath
  resources:
    requests:
      storage: 2Gi
---

Step 3: A Local PV (type=hostpath) provisioned via the OpenEBS Dynamic
Local PV Provisioner looks like this:
---
apiVersion: v1
kind: PersistentVolume
metadata:
  annotations:
    pv.kubernetes.io/provisioned-by: openebs.io/local
  creationTimestamp: 2019-05-02T15:44:35Z
  finalizers:
  - kubernetes.io/pv-protection
  name: pvc-2fe08284-6cf1-11e9-be8b-42010a800155
  resourceVersion: "2062"
  selfLink: /api/v1/persistentvolumes/pvc-2fe08284-6cf1-11e9-be8b-42010a800155
  uid: 2fedaff8-6cf1-11e9-be8b-42010a800155
spec:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 2Gi
  claimRef:
    apiVersion: v1
    kind: PersistentVolumeClaim
    name: pvc-hp
    namespace: default
    resourceVersion: "2060"
    uid: 2fe08284-6cf1-11e9-be8b-42010a800155
  local:
    path: /var/openebs/local/pvc-2fe08284-6cf1-11e9-be8b-42010a800155
    fsType: ""
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - gke-kmova-helm-default-pool-6c1271a5-n8b0
  persistentVolumeReclaimPolicy: Delete
  storageClassName: hostpath
status:
  phase: Bound
---

Note that the location of the hostpaths on the node are abstracted from the
application developers and are under the Administrators control.

Implementation Details:
-----------------------
(a) The configuration of whether to select a complete storage device
    of a hostpath is determined by the StorageClass annotations, inline
    with other configuration options provided by OpenEBS.
(b) When using StorageType as device, the Local Provisioner will
    interact with the OpenEBS Node Storage Device Manager (NDM) to
    identify the device to be used. Each PVC of type StorageType=device
    will create a BDC and wait for the NDM to provide an appropriate BD.
    From the BD, the provisioner will extract the path details and create
    a Local PV. When using unformatted block devices, the administrator
    can specific the type of FS to be put on block devices using the
    FSType CAS Policy in StorageClass.
(c) The StorageClass can work either with `waitForConsumer`, in which
    case, the PV is created on the node where the Pod is scheduled or
    vice versa. (Note: In the initial version, only `waitForConsumer` is
    supported.)
(d) When using the hostpath, the administrator can select the location using:
    - BasePath: By default, the hostpath volumes will be created under
      `/var/openebs/local`. This default path can be changed by passing the
      "OPENEBS_IO_BASE_PATH" ENV variable to the Hostpath Provisioner Pod.
      It is also possible to specify a different location using the
      CAS Policy `BasePath` in the StorageClass.

    The hostpath used in the above configuration can be:
    - OS Disk  - possibly a folder dedicated to saving data on each node.
    - Additional Disks - mounted as ext4 or any other filesystem
    - External Storage - mounted as ext4 or any other filesystem
(e) The backup and restore via Velero Plugin has been verified to work for
    OpenEBS Local PV. Supported from OpenEBS 1.0 and higher.

Future Improvements and Limitations:
------------------------------------
- Ability to enforce capacity limits. The application can exceed it usage
  of capacity beyond what it requested.
- Ability to enforce provisioning limits based on the capacity available
  on a given node or the number of PVs already provisioned.
- Ability to use hostpaths and devices that can potentially support snapshots.
  Example: a hostpath backed by github, or by LVM or ZFS where capacity also
  can be enforced.
- Extend the capabilities of the Local PV provisioner to handle cases where
  underlying devices are moved to new node and needs changes to the node
  affinity.
- Move towards using a CSI based Hostpath provisioner.

*/
package app
