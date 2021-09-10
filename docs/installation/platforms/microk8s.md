# Install Dynamic-LocalPV-Provisioner on MicroK8s

The Dynamic-LocalPV-Provisioner may be installed into a MicroK8s cluster in ANY ONE of the following ways:

## Using the OpenEBS Addon

MicroK8s (v1.21 onwards) ships with an OpenEBS Addon which deploys LocalPV, cStor and Jiva storage engine control-plane components. Enable the addon using the following command:
```console
$ microk8s enable openebs
```

Once installation succeeds, you may verify the creation of the Dynamic-LocalPV-Provisioner components using the following commands:
```console
$ microk8s kubectl get pods -n openebs
$ microk8s kubectl get storageclass
```


## Using the OpenEBS Helm Chart

Using the helm chart directly let's you cuztomize your Dynamic-LocalPV-Provisioner deployment ([Helm chart README](https://github.com/openebs/charts/blob/develop/charts/openebs/README.md)). You will need to use the Helm3 MicroK8s Addon for this.

```console
$ microk8s enable helm3
```
Add the openebs helm chart repo
```console
$ microk8s helm3 repo add openebs https://openebs.github.io/charts
$ microk8s helm3 repo update
```

Install the helm chart.
```console
$ #Default installation command. This sets the default directories under '/var/snap/microk8s/common'
$ microk8s helm3 install openebs openebs/openebs -n openebs --create-namespace \
	--set localprovisioner.basePath="/var/snap/microk8s/common/var/openebs/local"
	--set ndm.sparse.path="/var/snap/microk8s/common/var/openebs/sparse"
	--set varDirectoryPath.baseDir="/var/snap/microk8s/common/var/openebs"
```

Once installation succeeds, you may verify the creation of the Dynamic-LocalPV-Provisioner components using the following commands:
```console
$ microk8s kubectl get pods -n openebs
$ microk8s kubectl get storageclass
```

## Using Operator YAML

You may install Dynamic-LocalPV-Provisioner using the openebs-operator-lite.yaml and openebs-lite-sc.yaml files as well. Use the following commands to install using the Operator YAMLs, while creating the default directories under '/var/snap/microk8s/common'

```console
$ #Apply openebs-operator-lite.yaml
$ curl -fSsL https://openebs.github.io/charts/openebs-operator-lite.yaml | sed 's|\(/var/openebs\)|/var/snap/microk8s/common\1|g' | kubectl apply -f -
$ #Apply openebs-lite-sc.yaml
$ curl -fSsL https://openebs.github.io/charts/openebs-lite-sc.yaml | sed 's|\(/var/openebs\)|/var/snap/microk8s/common\1|g' | kubectl apply -f -
```

Once installation succeeds, you may verify the creation of the Dynamic-LocalPV-Provisioner components using the following commands:
```console
$ microk8s kubectl get pods -n openebs
$ microk8s kubectl get storageclass
```

For instructions on using the StorageClasses and creating volumes, refer to the [quickstart](https://github.com/openebs/dynamic-localpv-provisioner/blob/develop/docs/quickstart.md).
