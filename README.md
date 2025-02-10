# OpenEBS Dynamic LocalPV Provisioner

[![CNCF Status](https://img.shields.io/badge/cncf%20status-sandbox-blue.svg)](https://www.cncf.io/projects/openebs/)
[![LICENSE](https://img.shields.io/github/license/openebs/openebs.svg)](./LICENSE)
[![Slack](https://img.shields.io/badge/chat-slack-ff1493.svg?style=flat-square)](https://kubernetes.slack.com/messages/openebs)
[![Community Meetings](https://img.shields.io/badge/Community-Meetings-blue)](https://us05web.zoom.us/j/87535654586?pwd=CigbXigJPn38USc6Vuzt7qSVFoO79X.1)
[![Go Report Card](https://goreportcard.com/badge/github.com/openebs/dynamic-localpv-provisioner)](https://goreportcard.com/report/github.com/openebs/dynamic-localpv-provisioner)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B162%2Fgithub.com%2Fopenebs%2Fdynamic-localpv-provisioner.svg?type=shield&issueType=license)](https://app.fossa.com/projects/custom%2B162%2Fgithub.com%2Fopenebs%2Fdynamic-localpv-provisioner?ref=badge_shield&issueType=license)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/9666/badge)](https://www.bestpractices.dev/projects/9666)
[![CLOMonitor](https://img.shields.io/endpoint?url=https://clomonitor.io/api/projects/cncf/openebs/badge)](https://clomonitor.io/projects/cncf/openebs)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/openebs)](https://artifacthub.io/packages/helm/openebs/openebs)

## Overview

OpenEBS Dynamic LocalPV Provisioner is an open‐source Kubernetes component that automates the dynamic provisioning of local persistent volumes. It converts local storage available on Kubernetes nodes, such as hostPath directories into persistent volumes accessible via PVCs. The provisioner automatically assigns node affinity metadata, ensuring that pods run on the node hosting the storage. The tool simplifies local storage management by automating volume creation, binding, and cleanup processes for hostpaths. It overcomes challenges of static provisioning by dynamically allocating storage on demand. Overall, the provisioner offers a robust, scalable solution for managing local persistent hostpath volumes in Kubernetes.

## Why OpenEBS Dynamic LocalPV Provisioner?

- <b>Dynamic Provisioning</b>: Automatically creates persistent hostpath volumes on demand from local node storage, reducing manual configuration. 
- <b>Seamless Kubernetes Integration</b>: Uses node affinity to ensure pods are scheduled on the node where the volume is located, maintaining data consistency.  
- <b>Customizable Storage Behavior</b>: Offers flexible configuration through StorageClasses, supporting hostpath features like quota enforcement etc.
- <b>Optimized for High Performance</b>: Ideal for high-performance, low-latency, stateful applications like replicated databases which need local storage. 

## Architecture

```mermaid

graph TD
    %% Define Styles with Black Text
    style kubelet fill:#ffcc00,stroke:#d4a017,stroke-width:2px,color:#000
    style Provisioner fill:#66ccff,stroke:#3388cc,stroke-width:2px,color:#000
    style HelperPod fill:#99ff99,stroke:#44aa44,stroke-width:2px,color:#000
    style App fill:#ff9999,stroke:#cc6666,stroke-width:2px,color:#000
    style Hostpath fill:#ffdd99,stroke:#d4a017,stroke-width:2px,color:#000
    style PVC fill:#ffdd99,stroke:#d4a017,stroke-width:2px,color:#000
    style PV fill:#d9b3ff,stroke:#9955cc,stroke-width:2px,color:#000

    subgraph "Kubernetes Cluster"
        subgraph "Kubelet"
            kubelet["Kubelet"]
        end
        subgraph "OpenEBS LocalPV"
            Provisioner["LocalPV Provisioner"]
            HelperPod["Helper Pod (Creates/Cleanup Directory)"]
        end
        subgraph "Worker Node"
            App["Application Pod"]
            PVC["Persistent Volume Claim"]
            PV["Persistent Volume"]
            Hostpath["[User-defined path on host]"]
        end
    end

    %% Storage Flow
    App -->|Requests Storage| PVC
    PVC -->|Binds to| PV
    PV -->|Mounted on| Hostpath
    kubelet -->|Mounts Path| Hostpath

    %% Provisioning Flow
    Provisioner -->|Launches| HelperPod
    Provisioner -->|Watches PVC Requests| Provisioner
    Provisioner -->|Creates PV| PV
    HelperPod -->|Creates Directory| Hostpath
    PV -->|Bound to| PVC

```

Please check [here](./design/hostpath_localpv_provisioner.md) for complete design and architecture.

## Kubernetes Compatibility Matrix

|          | Kubernetes <= 1.18 | Kubernetes >=1.19 |
|----------|--------------------|-------------------|
| `v4.0.x` | ✕                  | ✓                 | 
| `v4.1.x` | ✕                  | ✓                 |
| `v4.2.x` | ✕                  | ✓                 |
| `HEAD`   | ✕                  | ✓                 |

## Documents

- [Prerequisites](./docs/quickstart.md#prerequisites)
- [Quickstart](./docs/quickstart.md#quickstart)
- [Developer Setup](./docs/developer.md)
- [Testing](./docs/developer-setup.md#testing)
- [Contibuting Guidelines](./CONTRIBUTING.md)
- [Governance](./GOVERNANCE.md)
- [Changelog](./CHANGELOG.md)
- [Release Process](./RELEASE.md)

## Features

- [x] Access Modes
    - [x] ReadWriteOnce
    - ~~ReadOnlyMany~~
    - ~~ReadWriteMany~~
- [x] Volume modes
    - [x] `Filesystem` mode
    - [ ] `Block` mode
- [x] [Volume Resize(Via Quotas)](./docs/tutorials/hostpath/xfs_quota/)

## Inspiration/Credit

OpenEBS Local PV has been inspired by the prior work done by the following the Kubernetes projects:

- <https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/tree/HEAD/examples/hostpath-provisioner>
- <https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner>
- <https://github.com/rancher/local-path-provisioner>

## Dev Activity dashboard

![Alt](https://repobeats.axiom.co/api/embed/d7011cfd2ca7cfc49674fc2b64c587adf20862e2.svg "Repobeats analytics image")

## License Compliance

[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B162%2Fgithub.com%2Fopenebs%2Fdynamic-localpv-provisioner.svg?type=large&issueType=license)](https://app.fossa.com/projects/custom%2B162%2Fgithub.com%2Fopenebs%2Fdynamic-localpv-provisioner?ref=badge_large&issueType=license)

## OpenEBS is a [CNCF Sandbox Project](https://www.cncf.io/projects/openebs)

![OpenEBS is a CNCF Sandbox Project](https://github.com/cncf/artwork/blob/main/other/cncf/horizontal/color/cncf-color.png)