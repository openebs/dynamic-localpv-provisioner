# Dynamic Kubernetes Local Persistent Volumes

<img width="300" align="right" alt="OpenEBS Logo" src="https://raw.githubusercontent.com/cncf/artwork/master/projects/openebs/stacked/color/openebs-stacked-color.png" xmlns="http://www.w3.org/1999/html">

<p align="justify">
<strong>OpenEBS Dynamic Local PV provisioner</strong> can be used to dynamically provision 
Kubernetes Local Volumes using different kinds of storage available on the Kubernetes nodes. 
<br>
<br>
</p>

Local Persistent Volumes are great for distributed cloud native data services that can handle resiliency and availability and expect low-latency access to the storage. Local Persistent Volumes can be provisioned using the hostpath, NVMe or PCIe based SSDs, Hard Disks or on top of other filesystems like ZFS, LVM. 

Some of the targetted applications are:
- Distributed SQL Databases like PostgreSQL
- Distributed No-SQL Databases like MongoDB, Cassandra
- Distributed Object Storages like MinIO (distributed mode)
- Distributed Streaming services like Apache Kakfa, 
- Distributed Logging and search services like ElasticSearch, Solr
- AI/ML workloads

## Overview 

Kubernetes Local persistent volumes allows users to access local storage through the
standard PVC interface in a simple and portable way.  The PV contains node
affinity information that the system uses to schedule pods to the correct
nodes.

OpenEBS Dynamic Local PVs extends the capabilities provided by the Kubernetes Local PV
by making use of the OpenEBS Node Storage Disk Manager (NDM), the significant
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

## Install

Please refer to our [Quickstart](https://github.com/openebs/dynamic-localpv-provisioner/tree/develop/docs/quickstart.md) and the [OpenEBS Documentation](http://docs.openebs.io/).

## Contributing

Head over to the [CONTRIBUTING.md](./CONTRIBUTING.md).

## Community, discussion, and support

Learn how to engage with the OpenEBS community on the [community page](https://github.com/openebs/openebs/tree/master/community).

You can reach the maintainers of this project at:

- [Kubernetes Slack](http://slack.k8s.io/) channels: 
      * [#openebs](https://kubernetes.slack.com/messages/openebs/)
      * [#openebs-dev](https://kubernetes.slack.com/messages/openebs-dev/)
- [Mailing List](https://lists.cncf.io/g/cncf-openebs-users)

### Code of conduct

Participation in the OpenEBS community is governed by the [CNCF Code of Conduct](CODE-OF-CONDUCT.md).

## Inspiration/Credit

OpenEBS Local PV has been inspired by the prior work done by the following the Kubernetes projects:
- https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/tree/master/examples/hostpath-provisioner
- https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner
- https://github.com/rancher/local-path-provisioner
