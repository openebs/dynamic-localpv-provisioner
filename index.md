# OpenEBS LocalPV Helm Repository

<img width="300" align="right" alt="OpenEBS Logo" src="https://raw.githubusercontent.com/cncf/artwork/master/projects/openebs/stacked/color/openebs-stacked-color.png" xmlns="http://www.w3.org/1999/html">

[Helm3](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

Once Helm is set up properly, add the repo as follows:

```bash
$ helm repo add openebs-localpv https://openebs.github.io/dynamic-localpv-provisioner
```

You can then run `helm search repo openebs-localpv` to see the charts.

#### Update OpenEBS LocalPV Repo

Once OpenEBS Localpv repository has been successfully fetched into the local system, it has to be updated to get the latest version. The LocalPV charts repo can be updated using the following command.

```bash
helm repo update
```

#### Install using Helm 3

- Assign openebs namespace to the current context:
```bash
kubectl config set-context <current_context_name> --namespace=openebs
```

- If namespace is not created, run the following command
```bash
helm install <your-relase-name> openebs-localpv/localpv-provisioner --create-namespace
```
- Else, if namespace is already created, run the following command
```bash
helm install <your-relase-name> openebs-localpv/localpv-provisioner
```

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Dependencies

By default this chart installs additional, dependent charts:

| Repository | Name | Version |
|------------|------|---------|
| https://openebs.github.io/node-disk-manager | openebs-ndm | 1.0.2 |


To disable the dependency during installation, set `openebsNDM.enabled` to `false`.

_See [helm dependency](https://helm.sh/docs/helm/helm_dependency/) for command documentation._

## Uninstall Chart

```console
# Helm
$ helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
# Helm
$ helm upgrade [RELEASE_NAME] [CHART] --install
```


## Configuration

The following table lists the configurable parameters of the OpenEBS LocalPV Provisioner chart and their default values.

| Parameter                                   | Description                                   | Default                                   |
| ------------------------------------------- | --------------------------------------------- | ----------------------------------------- |
| `release.version`                           | LocalPV Provisioner release version               | `2.4.0`                         |
| `analytics.enabled`                         | Enable sending stats to Google Analytics          | `true`                          |
| `analytics.pingInterval`                    | Duration(hours) between sending ping stat         | `24h`                           |
| `imagePullSecrets`                          | Provides image pull secrect                       | `""`                            |
| `localpv.enabled`                           | Enable LocalPV Provisioner                        | `true`                          |
| `localpv.image.registry`                    | Registry for LocalPV Provisioner image            | `""`                            |
| `localpv.image.repository`                  | Image repository for LocalPV Provisioner          | `openebs/localpv-provisioner`   |
| `localpv.image.pullPolicy`                  | Image pull policy for LocalPV Provisioner         | `IfNotPresent`                  |
| `localpv.image.tag`                         | Image tag for LocalPV Provisioner                 | `2.4.0`                         |
| `localpv.updateStrategy.type`               | Update strategy for LocalPV Provisioner           | `RollingUpdate`                 |
| `localpv.annotations`                       | Annotations for LocalPV Provisioner metadata      | `""`                            |
| `localpv.podAnnotations`                    | Annotations for LocalPV Provisioner pods metadata | `""`                            |
| `localpv.privileged`                        | Run LocalPV Provisioner with extra privileges     | `true`                          |
| `localpv.resources`                         | Resource and request and limit for containers     | `""`                            |
| `localpv.podLabels`                         | Appends labels to the pods                        | `""`                            |
| `localpv.nodeSelector`                      | Nodeselector for LocalPV Provisioner pods         | `""`                            |
| `localpv.tolerations`                       | LocalPV Provisioner pod toleration values         | `""`                            |
| `localpv.securityContext`                   | Seurity context for container                     | `""`                            |
| `localpv.healthCheck.initialDelaySeconds`   | Delay before liveness probe is initiated          | `30`                            |
| `localpv.healthCheck.periodSeconds`         | How often to perform the liveness probe           | `60`                            |
| `localpv.replicas`                          | No. of LocalPV Provisioner replica                | `1`                             |
| `localpv.enableLeaderElection`              | Enable leader election                            | `true`                          |
| `localpv.basePath`                          | BasePath for hostPath volumes on Nodes            | `"/var/openebs/local"`          |
| `localpv.affinity`                          | LocalPV Provisioner pod affinity                  | `{}`                            |
| `helperPod.image.registry`                  | Registry for helper image                         | `""`                            |
| `helperPod.image.repository`                | Image for helper pod                              | `"openebs/linux-utils"`         |
| `helperPod.image.pullPolicy`                | Pull policy for helper pod                        | `"IfNotPresent"`                |
| `helperPod.image.tag`                       | Image tag for helper image                        | `2.4.0`                         |
| `rbac.create`                               | Enable RBAC Resources                             | `true`                          |
| `rbac.pspEnabled`                           | Create pod security policy resources              | `false`                         |
| `openebsNDM.enabled`                        | Install openebs NDM dependency                    | `true`                          |