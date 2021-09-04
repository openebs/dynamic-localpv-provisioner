# Install Dynamic-LocalPV-Provisioner on minikube

Follow the instructions below when installing dynamic-localpv-provisioner on minikube.

## Using no node driver

The node-disk-manager DaemonSet Pods require Kubernetes hostpath mounts for the directories '/run/udev', '/dev', '/proc', '/var/openebs' and '/var/openebs/sparse' from the host node. Running minikube with the 'none' driver allows the kubelet to mount the directories into the a mountpoint inside the NDM DaemonSet Pods.

Run minikube with the flag `--driver=none` to run minikube with no VM driver.
```bash
minikube start --driver=none
```
For more information on using the 'none' driver flag argument, [read the official minikube docs](https://minikube.sigs.k8s.io/docs/drivers/none/).

After minikube is started with no VM driver, proceed with installation as described in [the quickstart](https://github.com/openebs/dynamic-localpv-provisioner/blob/develop/docs/quickstart.md).
