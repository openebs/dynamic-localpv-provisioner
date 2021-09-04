# Local PV Provisioner BDD

Local PV Provisioner BDD tests are developed using ginkgo & gomega libraries.

## How to run the tests?

### Pre-requisites

- These tests are meant to be run in a single-node Kubernetes
  cluster with one single available blockdevice.

- Get your Kubernetes Cluster ready and make sure you can run 
  kubectl from your development machine. 
  Note down the path to the `kubeconfig` file used by kubectl 
  to access your cluster.  Example: /home/\<user\>/.kube/config

- (Optional) Set the KUBECONFIG environment variable on your 
  development machine to point to the kubeconfig file. 
  Example: `export KUBECONFIG=$HOME/.kube/config`

  If you do not set this ENV, you will have to pass the file 
  to the go test (or ginkgo) CLI

- Some of the tests require block devices (that are not mounted)
  to be available in the cluster.

- Install required OpenEBS LocalPV Provisioner components
  Example: `kubectl apply -f https://openebs.github.io/charts/openebs-operator-lite.yaml`

### Run tests

Run the tests by being in the localpv tests folder. 
  ```bash
  $ cd <repo-directory>/tests
  $ go test -ginkgo.v
  $ #Alternatively, you can run the test suite using the 'ginkgo' binary
  $ #ginkgo -v
 ```
  In case the KUBECONFIG env is not configured, you can run:
 ```bash
  $ go test -ginkgo.v -kubeconfig=/path/to/kubeconfig
  $ #Alternatively, you can run the test suite using the 'ginkgo' binary
  $ #ginkgo -v -- -kubeconfig=/path/to/kubeconfig
 ```

  If your OpenEBS LocalPV components are in a different Kubernetes namespace than 'openebs', you may use the '-openebs-namespace' flag:
 ```bash
  $ go test -ginkgo.v -openebs-namespace=<your-namespace>
  $ #Alternatively, you can run the test suite using the 'ginkgo' binary
  $ #ginkgo -v -- -openebs-namespace=<your-namespace>
 ```
