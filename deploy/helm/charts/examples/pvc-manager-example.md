# Example: Enabling PVC Manager

This example shows how to install the LocalPV provisioner with PVC Manager enabled.

## Installation with PVC Manager

To install the LocalPV provisioner with PVC Manager enabled:

```bash
helm install openebs-localpv openebs-localpv/localpv-provisioner \
  --namespace openebs \
  --create-namespace \
  --set pvcManager.enabled=true
```

## Advanced Configuration

For advanced PVC Manager configuration:

```bash
helm install openebs-localpv openebs-localpv/localpv-provisioner \
  --namespace openebs \
  --create-namespace \
  --set pvcManager.enabled=true \
  --set pvcManager.logLevel=debug \
  --set pvcManager.resources.requests.cpu=150m \
  --set pvcManager.resources.requests.memory=96Mi
```

## Custom Values File

Create a `pvc-manager-values.yaml` file:

```yaml
pvcManager:
  enabled: true
  logLevel: info
  port: 8080
  resources:
    limits:
      cpu: 200m
      memory: 128Mi
    requests:
      cpu: 100m
      memory: 64Mi
  tolerations:
    - effect: NoSchedule
      operator: Exists
    - effect: NoExecute
      operator: Exists
  nodeSelector:
    kubernetes.io/os: linux

# Enable analytics (optional)
analytics:
  enabled: true
```

Then install:

```bash
helm install openebs-localpv openebs-localpv/localpv-provisioner \
  --namespace openebs \
  --create-namespace \
  -f pvc-manager-values.yaml
```

## Verification

After installation, verify that PVC Manager is running:

```bash
# Check DaemonSet status
kubectl get daemonset -n openebs -l app=pvc-manager

# Check PVC Manager pods
kubectl get pods -n openebs -l app=pvc-manager

# Check Service
kubectl get svc -n openebs -l app=pvc-manager

# Check logs
kubectl logs -n openebs -l app=pvc-manager
```

## PVC Manager vs Helper Pods

When PVC Manager is enabled (`pvcManager.enabled=true`), the provisioner will:
- Use HTTP requests to PVC Manager for volume operations
- Avoid creating helper pods for each volume operation
- Provide better performance for volume provisioning

When PVC Manager is disabled (`pvcManager.enabled=false`), the provisioner will:
- Fall back to the traditional helper pod approach
- Create a helper pod for each volume operation
- Maintain backward compatibility