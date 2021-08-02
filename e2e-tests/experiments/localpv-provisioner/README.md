## About this experiment

This experiment deploys the openebs-operator provisioner in openebs namespace which includes localpv provisioner,openebs-ndm deamonset, openebs-ndm operator.

## Entry-Criteria

- K8s cluster should be in healthy state including all the nodes in ready state.
- If we don't want to use this experiment to deploy openebs operator lite, we can directly apply the operator file as mentioned below.

```
kubectl apply -f https://raw.githubusercontent.com/openebs/charts/gh-pages/<release-version>/openebs-operator-lite.yaml
```

## Exit-Criteria

- LocalPV and NDM components should be deployed successfully and all the pods including LocalPV provisioner, openebs-ndm and NDM operator are in running state.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of deploying openebs operator, clone openens/dynamic-localpv-provisioner[https://github.com/openebs/dynamic-localpv-provisioner] repo and then first apply rbac and crds for e2e-framework.
```
kubectl apply -f dynamic-localpv-provisioner/e2e-tests/hack/rbac.yaml
kubectl apply -f dynamic-localpv-provisioner/e2e-tests/hack/crds.yaml
```
then update the needed test specific values in run_e2e_test.yml file and create the kubernetes job.
```
kubectl create -f run_e2e_test.yml
```
All the env variables description is provided with the comments in the same file.

