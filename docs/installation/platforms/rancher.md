# Install Dynamic-LocalPV-Provisioner on Rancher v2/RKE

To use OpenEBS LocalPV Hostpath with an RKE/Rancher 2.x cluster, you will have to bind-mount the hostpath directories to the kubelet containers. You can do this by editing the kubelet configuration section of your RKE/Rancher 2.x cluster and adding in the `extra_binds` (see below).

**Note:** If you want to use a custom hostpath directory, then you will have to bind-mount the custom directory's absolute path. See below for an example with the default hostpath directory.

For an RKE cluster, you can add the `extra_binds` to your cluster.yml file and apply the changes using the `rke up` command.

For a Rancher 2.x cluster, you can edit your cluster's configuration options and add the `extra_binds` there.

```yaml
services:
  kubelet:
    extra_binds:
    #Default hostpath directory
    - /var/openebs/local:/var/openebs/local
```

For more information, please go through the official Rancher documentaion -- [RKE - Kubernetes Configuration Options](https://rancher.com/docs/rke/latest/en/config-options/services/services-extras/#extra-binds), [RKE - Installation](https://rancher.com/docs/rke/latest/en/installation/#deploying-kubernetes-with-rke).

After adding the `extra_binds` are added, proceed with installation as described in [the quickstart](https://github.com/openebs/dynamic-localpv-provisioner/blob/develop/docs/quickstart.md).
