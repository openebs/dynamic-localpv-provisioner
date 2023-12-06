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
  $ go install github.com/onsi/ginkgo/ginkgo@v1.16.4
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
