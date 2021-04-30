# OpenEBS LocalPV Helm Repository

<img width="300" align="right" alt="OpenEBS Logo" src="https://raw.githubusercontent.com/cncf/artwork/master/projects/openebs/stacked/color/openebs-stacked-color.png" xmlns="http://www.w3.org/1999/html">

[Helm3](https://helm.sh) must be installed to use the charts.
Please refer to the official Helm [documentation](https://helm.sh/docs/) to get started.

Once Helm is set up properly, add the repo as follows:

```bash
helm repo add openebs-localpv https://openebs.github.io/dynamic-localpv-provisioner
```

You can then run `helm search repo openebs-localpv` to see the charts.

#### Update OpenEBS LocalPV Repo

Once OpenEBS LocalPV repository has been successfully fetched into the local system, it has to be updated to get the latest version. The LocalPV charts repo can be updated using the following command:

```bash
helm repo update
```

#### Install using Helm 3

- Run the following command to install the OpenEBS Dynamic LocalPV Provisioner helm chart:
```bash
helm install [RELEASE_NAME] openebs-localpv/localpv-provisioner --namespace [NAMESPACE]
```


_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Dependencies

By default this chart installs additional, dependent charts:

| Repository | Name |
|------------|------|
| https://openebs.github.io/node-disk-manager | openebs-ndm |


To disable the dependency during installation, set `openebsNDM.enabled` to `false`.

_See [helm dependency](https://helm.sh/docs/helm/helm_dependency/) for command documentation._

## Uninstall Chart

```console
# Helm
helm uninstall [RELEASE_NAME] --namespace [NAMESPACE]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
# Helm
helm upgrade [RELEASE_NAME] [CHART] --install --namespace [NAMESPACE]
```


## Configuration

Refer to the OpenEBS Dynamic LocalPV Provisioner Helm chart [README.md file](https://github.com/openebs/dynamic-localpv-provisioner/blob/develop/deploy/helm/charts/README.md) for detailed configuration options.

[Click here](https://github.com/openebs/dynamic-localpv-provisioner/blob/develop/docs/quickstart.md) for the Quickstart guide.
