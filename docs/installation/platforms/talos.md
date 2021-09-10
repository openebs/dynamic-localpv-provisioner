# Install Dynamic-LocalPV-Provisioner on Talos

To use OpenEBS LocalPV Hostpath with an Talos cluster, you will have to bind-mount the hostpath directories to the kubelet containers. You can do this by editing the KubeletConfig section of your cluster machineconfig and adding in the `extraMounts` (see below).

**Note:** If you want to use a custom hostpath directory, then you will have to bind-mount the custom directory's absolute path. See below for an example with the default hostpath directory.

Visit the [Talos official documentation](https://www.talos.dev/docs) for instructions on editing machineconfig or using config patches.

```yaml
kubelet:
    extraMounts:
          #Default Hostpath directory
        - destination: /var/openebs/local
          type: bind
          source: /var/openebs/local
          options:
            - rbind
            - rshared
            - rw
```

After adding the `extraMounts` are added, proceed with installation as described in [the quickstart](https://github.com/openebs/dynamic-localpv-provisioner/blob/develop/docs/quickstart.md).
