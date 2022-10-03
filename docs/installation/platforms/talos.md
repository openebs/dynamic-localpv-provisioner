# Install Dynamic-LocalPV-Provisioner on Talos

To use OpenEBS LocalPV Hostpath with a Talos cluster, you will have to bind-mount the hostpath directories to the kubelet containers. You can do this by editing the KubeletConfig section of your cluster machineconfig and adding in the `extraMounts` (see below).

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

If you are using the default Talos security policy you will also have to add privileged security labels on the `openebs` namespace to allow it to use `hostPath` volumes. Eg:

```yaml
apiVersion: v1
kind: Namespace
metadata:
    name: openebs
    labels:
        pod-security.kubernetes.io/audit: privileged
        pod-security.kubernetes.io/enforce: privileged
        pod-security.kubernetes.io/warn: privileged
```

Caution: When using local storage on Talos, you must remember to pass the `--preserve` argument when running `talosctl upgrade` to avoid host paths getting wiped out during the upgrade (as noted in [Talos Local Storage documentation](https://www.talos.dev/v1.2/kubernetes-guides/configuration/replicated-local-storage-with-openebs-jiva/)).

After adding the required configuration, proceed with installation as described in [the quickstart](https://github.com/openebs/dynamic-localpv-provisioner/blob/develop/docs/quickstart.md).
