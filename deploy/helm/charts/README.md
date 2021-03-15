#  OpenEBS LocalPV Provisioner

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Chart Lint and Test](https://github.com/openebs/dynamic-localpv-provisioner/workflows/Chart%20Lint%20and%20Test/badge.svg)
![Release Charts](https://github.com/openebs/dynamic-localpv-provisioner/workflows/Release%20Charts/badge.svg?branch=develop)

A Helm chart for openebs dynamic localpv provisioner. This chart bootstraps OpenEBS Dynamic LocalPV provisioner deployment on a [Kubernetes](http://kubernetes.io) cluster using the  [Helm](https://helm.sh) package manager.


**Homepage:** <http://www.openebs.io/>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| akhilerm | akhil.mohan@mayadata.io |  |
| kiranmova | kiran.mova@mayadata.io |  |
| prateekpandey14 | prateek.pandey@mayadata.io |  |


## Get Repo Info

```console
helm repo add openebs-localpv https://openebs.github.io/dynamic-localpv-provisioner
helm repo update
```

_See [helm repo](https://helm.sh/docs/helm/helm_repo/) for command documentation._

## Install Chart

Please visit the [link](https://openebs.github.io/dynamic-localpv-provisioner/) for install instructions via helm3.

```console
# Helm
$ helm install [RELEASE_NAME] openebs-localpv/localpv-provisioner
```

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Dependencies

By default this chart installs additional, dependent charts:

| Repository | Name | Version |
|------------|------|---------|
| https://openebs.github.io/node-disk-manager | openebs-ndm | 1.3.0 |


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
| `release.version`                           | LocalPV Provisioner release version               | `2.7.0`                         |
| `analytics.enabled`                         | Enable sending stats to Google Analytics          | `true`                          |
| `analytics.pingInterval`                    | Duration(hours) between sending ping stat         | `24h`                           |
| `imagePullSecrets`                          | Provides image pull secrect                       | `""`                            |
| `localpv.enabled`                           | Enable LocalPV Provisioner                        | `true`                          |
| `localpv.image.registry`                    | Registry for LocalPV Provisioner image            | `""`                            |
| `localpv.image.repository`                  | Image repository for LocalPV Provisioner          | `openebs/localpv-provisioner`   |
| `localpv.image.pullPolicy`                  | Image pull policy for LocalPV Provisioner         | `IfNotPresent`                  |
| `localpv.image.tag`                         | Image tag for LocalPV Provisioner                 | `2.7.0`                         |
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
| `helperPod.image.tag`                       | Image tag for helper image                        | `2.7.0`                         |
| `rbac.create`                               | Enable RBAC Resources                             | `true`                          |
| `rbac.pspEnabled`                           | Create pod security policy resources              | `false`                         |
| `openebsNDM.enabled`                        | Install openebs NDM dependency                    | `true`                          |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
helm install <release-name> -f values.yaml --namespace openebs localpv-provisioner
```

> **Tip**: You can use the default [values.yaml](values.yaml)
