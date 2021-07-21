# Local PV Provisioner BDD

Local PV Provisioner BDD tests are developed using ginkgo & gomega libraries.

## How to run the tests?

### Pre-requisites

- Install Ginkgo and Gomega on your development machine. 
  ```
  $ go get github.com/onsi/ginkgo/ginkgo
  $ go get github.com/onsi/gomega/...
  ```
- Get your Kubernetes Cluster ready and make sure you can run 
  kubectl from your development machine. 
  Note down the path to the `kubeconfig` file used by kubectl 
  to access your cluster.  Example: /home/<user>/.kube/config

- (Optional) Set the KUBECONFIG environment variable on your 
  development machine to point to the kubeconfig file. 
  Example: KUBECONFIG=/home/<user>/.kube/config

  If you do not set this ENV, you will have to pass the file 
  to the ginkgo CLI

- Some of the tests require block devices (that are not mounted)
  to be available in the cluster.

- Install required OpenEBS components. 
  Example: `kubectl apply -f https://openebs.github.io/charts/openebs-operator-lite.yaml -f https://openebs.github.io/charts/openebs-lite-sc.yaml`

### Run tests

- Run the tests by being in the localpv tests folder. 
  `$ cd $GOPATH/src/github.com/openebs/maya/tests/localpv/`
  `$ ginkgo -v --`
 
  In case the KUBECONFIG env is not configured, you can run:
  `$ ginkgo -v -- -kubeconfig=/path/to/kubeconfig`

