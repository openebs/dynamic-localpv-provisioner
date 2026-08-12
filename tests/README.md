# Local PV Provisioner BDD

Local PV Provisioner BDD tests are developed using ginkgo & gomega libraries.

## How to run the tests?

### Pre-requisites

- These tests are meant to be run in a single-node Kubernetes (v1.16+)
  cluster with one single available blockdevice with no filesystem on
  it (should not be mounted).

- Some of the tests require the 'xfsprogs' and 'quota' packages to run.
  For Ubuntu, you may need to install the quota_v2 kernel module. Install
  the 'linux-image-extra-virtual' package to install the kernel module.
  ```bash
  $ #For Ubuntu/Debian
  $ sudo apt-get update && sudo apt-get install -y xfsprogs quota linux-image-extra-virtual
  $ ##The kernel module package name may be different depending on the OS image
  $ ##E.g.: linux-modules-extra-`uname -r`
  $ #For CentOS/RHEL
  $ sudo yum install -y xfsprogs quota
  ```

- You will require the Ginkgo binary to be able to run the tests.
  Install the latest Ginkgo binary using the following command:
  ```bash
  $ go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo@v2.14.0
  ```

- Get your Kubernetes Cluster ready and make sure you can run 
  kubectl from your development machine. 
  Note down the path to the `kubeconfig` file used by kubectl 
  to access your cluster.  Example: /home/\<user\>/.kube/config

- Set the KUBECONFIG environment variable on your 
  development machine to point to the kubeconfig file. 
  Example: `export KUBECONFIG=$HOME/.kube/config`

  If you do not set this ENV, you will have to pass the file 
  to the Ginkgo CLI (see below)

- The tests should not be run in parallel as it may lead to
  unavailability of blockdevices for some of the tests.

- Install required OpenEBS LocalPV Provisioner components
  Example: `kubectl apply -f https://openebs.github.io/charts/openebs-operator-lite.yaml`

### Run tests

Run the tests by being in the localpv tests folder. 
>**Note:** The tests require privileges to create loop devices and to create
directories in the '/var' directory.
  
```bash
$ cd <repo-directory>/tests
$ sudo -E env "PATH=$PATH" ginkgo -v
```
In case the KUBECONFIG env is not configured, you can run:
```bash
$ sudo -E env "PATH=$PATH" ginkgo -v -kubeconfig=/path/to/kubeconfig
```

If your OpenEBS LocalPV components are in a different Kubernetes namespace than 'openebs', you may use the '-openebs-namespace' flag:
```bash
$ sudo -E env "PATH=$PATH" ginkgo -v -openebs-namespace=<your-namespace>
```

>**Tip:** Raising a pull request to this repo's 'develop' branch (or any one of the release branches) will automatically run the BDD tests in GitHub Actions. You can verify your code changes by moving to the 'Checks' tab in your pull request page, and checking the results of the 'integration-test' check.

## Upgrade tests

The upgrade tests live in the 'upgrade' folder and are a suite of their own.
They install the last released chart, upgrade it to the chart in the working
tree, and check that the release converges, that the volumes provisioned before
the upgrade survive it, and that the upgraded provisioner still provisions and
cleans up volumes.

The version the upgrade starts from is derived from the chart version in the
working tree, which always names the version the branch is working towards. The
newest release below it is therefore the release users would be upgrading from:

| Branch        | Chart version      | Upgrades from                             |
| ------------- | ------------------ | ----------------------------------------- |
| `develop`     | `x.y.0-develop`    | the last minor release, e.g. `4.5.1`      |
| `release/x.y` | `x.y.z-prerelease` | the last patch release of `x.y`, e.g. `4.5.1` |

### Pre-requisites

- Everything under [Pre-requisites](#pre-requisites) above, except the quota
  packages and the spare blockdevice, which these tests do not use.

- Helm v3.14 or above, for `helm get metadata` and
  `helm upgrade --reset-then-reuse-values`.

- No 'localpv-provisioner' helm release installed: the suite installs and
  uninstalls that release itself. It refuses to run when the name is already
  taken, rather than remove a release it did not create, so point it at another
  release with '-release-name', or opt in to removing the existing one with
  '-uninstall-existing' (which is what './ci/ci-test.sh upgrade -r' does).

- The image of the chart under test has to be available to the cluster's
  container runtime, as the upgraded release is deployed with an image pull
  policy of 'Never'. `make provisioner-localpv-image` builds it.

### Run the upgrade tests

From the repo directory:
```bash
$ make provisioner-localpv-image
$ make upgrade-test
```

Or directly, being in the upgrade tests folder:
```bash
$ cd <repo-directory>/tests/upgrade
$ sudo -E env "PATH=$PATH" ginkgo -v
```

The suite takes the same '-kubeconfig' and '-openebs-namespace' flags as the
tests above, plus the ones below. Ginkgo parses its own flags first and rejects
anything it does not recognise, so the suite's flags have to follow a '--':
```bash
$ sudo -E env "PATH=$PATH" ginkgo -v -- -skip-cleanup
```

| Flag             | Default                        | Description                                        |
| ---------------- | ------------------------------ | -------------------------------------------------- |
| `-chart-dir`     | `<repo-directory>/deploy/helm/charts` | The chart to upgrade to.                    |
| `-release-name`  | `localpv-provisioner`          | The helm release to install and upgrade.           |
| `-skip-cleanup`  | `false`                        | Leave the helm release and the test namespace behind, to debug with. |
| `-uninstall-existing` | `false`                   | Uninstall a release already using the name under test, instead of refusing to run. |

And these environment variables:

| Variable                     | Default                                                    | Description                                      |
| ---------------------------- | ---------------------------------------------------------- | ------------------------------------------------ |
| `UPGRADE_FROM_VERSION`       | resolved from the chart version, as described above         | The released chart version to upgrade from.      |
| `UPGRADE_CHART_REPO_URL`     | `https://openebs.github.io/dynamic-localpv-provisioner`      | Where the released charts are pulled from.       |
| `UPGRADE_CHART_REPO_NAME`    | `openebs-localpv`                                            | The local alias for that repository.             |
| `UPGRADE_CHART_NAME`         | `localpv-provisioner`                                        | The chart name in that repository.               |
| `UPGRADE_CHART_DIR`          | `<repo-directory>/deploy/helm/charts`                        | The chart to upgrade to, same as `-chart-dir`.   |
| `UPGRADE_RELEASE_NAME`       | `localpv-provisioner`                                        | The helm release, same as `-release-name`.       |
| `UPGRADE_IMAGE_PULL_POLICY`  | `Never`                                                      | Image pull policy of the upgraded release.       |
| `UPGRADE_SKIP_CLEANUP`       | `false`                                                      | Keep the release and namespace, same as `-skip-cleanup`. |
| `UPGRADE_UNINSTALL_EXISTING` | `false`                                                      | Remove a release already using the name, same as `-uninstall-existing`. |
| `OPENEBS_NAMESPACE`          | `openebs`                                                    | Namespace of the release, same as `-openebs-namespace`. |

The suite installs and uninstalls the release itself, so to keep everything
around after a run, use `-x`, which sets `UPGRADE_SKIP_CLEANUP` for the suite as
well as skipping the cleanup `ci-test.sh` does of its own:
```bash
$ ./ci/ci-test.sh upgrade -x
```

To upgrade from a specific release, instead of the one which would be resolved:
```bash
$ UPGRADE_FROM_VERSION=4.4.0 make upgrade-test
```

>**Tip:** In a pull request, these run as the first step of the
'integration-test' check, before the tests above. They go first because they
need a cluster with nothing installed on it yet.
