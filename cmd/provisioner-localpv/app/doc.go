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

While the Kubernetes Local PVs are mainly recommended for cases where a complete
storage device should be assigned to a Pod. OpenEBS Dynamic Local PV provisioner
will help provisioning the Local PVs dynamically by integrating into the features
offered by OpenEBS Node Storage Device Manager, and also offers the
flexibility to either select a complete storage device or
a hostpath (or subpath) directory.

Infact in some cases, the Kubernetes nodes may have limited number of storage
devices attached to the node and hostpath based Local PVs offer efficient management
of the storage available on the node.

Inspiration:
------------
The implementation has been influenced by the prior work done by the Kubernetes community,
specifically the following:
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
      #  value: "storage-device"
      # (Default)
      - name: StorageType
        value: "hostpath"
      # If the StorageType is hostpath, then BasePath
      # specifies the location where the volume subdirectory
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
  hostPath:
    path: /var/openebs/local/pvc-2fe08284-6cf1-11e9-be8b-42010a800155
    type: DirectoryOrCreate
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
    identify the device to be used.
(c) The StorageClass can work either with `waitForConsumer`, in which
    case, the PV is created on the node where the Pod is scheduled or
    vice versa. (Note: In the initial version, only `waitForConsumer` is
    supported.)
(d) When using the hostpath, the administrator can select the location using:
    - BasePath: By default, the hostpath volumes will be created under `/var/openebs/local`.
      This default path can be changed by passing the "OPENEBS_IO_BASE_PATH" ENV
      variable to the Hostpath Provisioner Pod. It is also possible to specify
      a different location using the CAS Policy `BasePath` in the StorageClass.

    The location of the hostpaths used in the above configuration options can be:
    - OS Disk  - possibly a folder dedicated to saving data on each node.
    - Additional Disks - mounted as ext4 or any other filesystem
    - External Storage - mounted as ext4 or any other filesystem

Future Work:
------------
The current implementation provides basic support for using Local PVs. This will
be enhanced in the upcoming releases with the following features:
- Ability to use a git backed hostpath, so data can be backed up to a github/gitlab
- Ability to use hostpaths that are managed by NDM - that monitors for usage, helps
  with expanding the storage of a given hostpath. For example - use LVM or ZFS to
  create a host mount with attached disks. Where additional disks can be added or failed
  disks replaced without impacting the workloads running on their hostpaths.
- Integrate with Projects like Valero or Kasten that can handle backup and restor of data
  stored on the Hostpath PVs attached to a workload.
- Provide tools that can help with recovering from situations where PVs are tied to nodes
  that can never recover. For example, a Stateful Workload can be associated with a
  Hostpath PV on Node-z. Say Node-z  becomes inaccassible for reasons beyond the control
  like - a site/zone/rack disaster or the disks went up in flames. The PV will still be
  having the Node Affinity to Node-z, which will make the Workload to get stuck in pending
  state.
- Move towards using a CSI based Hostpath provisioner. Some of the features required to
  use the Hostpath PVs like the Volume Topology are not yet available in the CSI. As the
  CSI driver stabilizes, this can be moved into CSI.
- Provide an option to specify a white list of paths which can be used by the application
  developers. The white list can be tied into application developers namespace. For example:
  all /var/developer1/* for PVCs in namespace developer1, etc.
- Ability to enforce runtime capacity limits
- Ability to enforce provisioning limits based on the node capacity or the number of PVs
  already provisioned.

*/
package app
