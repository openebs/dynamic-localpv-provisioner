# Dynamic LocalPV Provisioner - PVC Manager Architecture

This document describes the new PVC Manager architecture for the OpenEBS Dynamic LocalPV Provisioner.

## Architecture Overview

The new architecture replaces the helper pod creation mechanism with a DaemonSet-based PVC Manager service that runs on each node and provides HTTP API endpoints for volume operations.

### Previous Architecture (Helper Pods)
```
Provisioner -> Creates Helper Pod -> Pod performs mkdir/rm operations -> Pod terminates
```

### New Architecture (PVC Manager)
```
Provisioner -> HTTP Request -> PVC Manager (DaemonSet) -> Direct filesystem operations
```

## Components

### 1. PVC Manager Service
- **Location**: `cmd/pvc-manager/`
- **Purpose**: HTTP service running on each node via DaemonSet
- **Endpoints**:
  - `POST /api/v1/volumes/create` - Create volume directory
  - `POST /api/v1/volumes/delete` - Delete volume directory  
  - `POST /api/v1/volumes/quota` - Apply filesystem quota
  - `GET /api/v1/health` - Health check

### 2. PVC Manager Client
- **Location**: `cmd/provisioner-localpv/app/pvc_manager_client.go`
- **Purpose**: HTTP client for provisioner to communicate with PVC Manager

### 3. Provisioner Updates
- **Modified files**:
  - `cmd/provisioner-localpv/app/provisioner_hostpath.go`
  - `cmd/provisioner-localpv/app/helper_pvc_manager.go`
  - `cmd/provisioner-localpv/app/env.go`

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OPENEBS_IO_ENABLE_PVC_MANAGER` | `true` | Enable PVC Manager mode |
| `OPENEBS_IO_PVC_MANAGER_PORT` | `8080` | PVC Manager service port |

### Deployment

1. **Deploy PVC Manager DaemonSet**:
   ```bash
   kubectl apply -f deploy/kubectl/pvc-manager-rbac.yaml
   kubectl apply -f deploy/kubectl/pvc-manager-daemonset.yaml
   ```

2. **Update provisioner configuration to enable PVC Manager**:
   ```yaml
   env:
   - name: OPENEBS_IO_ENABLE_PVC_MANAGER
     value: "true"
   - name: OPENEBS_IO_PVC_MANAGER_PORT
     value: "8080"
   ```

## Benefits

### 1. Performance
- **Faster volume operations**: No pod creation/termination overhead
- **Reduced API server load**: Fewer ephemeral pod objects
- **Lower resource usage**: Single long-running process per node

### 2. Reliability
- **Persistent service**: Always available for volume operations
- **Better error handling**: Consistent HTTP API responses
- **Health monitoring**: Built-in health check endpoints

### 3. Scalability
- **Reduced scheduling overhead**: No pod scheduling delays
- **Better resource utilization**: Fixed resource consumption per node
- **Improved throughput**: Concurrent volume operations support

## Migration Path

### Backward Compatibility
The provisioner supports both modes simultaneously:
- Set `OPENEBS_IO_ENABLE_PVC_MANAGER=true` to use PVC Manager
- Set `OPENEBS_IO_ENABLE_PVC_MANAGER=false` to use helper pods (legacy)

### Migration Steps
1. Deploy PVC Manager DaemonSet
2. Verify PVC Manager is running on all nodes
3. Update provisioner with `OPENEBS_IO_ENABLE_PVC_MANAGER=true`
4. Test volume provisioning/deprovisioning
5. Monitor for any issues

## API Specification

### Create Volume Request
```json
{
  "name": "volume-name",
  "path": "/var/openebs/local/volume-path",
  "nodeAffinityLabels": {
    "kubernetes.io/hostname": "node-1"
  },
  "fsMode": "0777",
  "commands": ["mkdir", "-m", "0777", "-p"]
}
```

### Apply Quota Request
```json
{
  "name": "volume-name", 
  "path": "/var/openebs/local/volume-path",
  "softLimitGrace": "80%",
  "hardLimitGrace": "90%",
  "pvcStorage": 1073741824
}
```

### Response Format
```json
{
  "success": true,
  "message": "Volume created successfully"
}
```

## Building and Testing

### Build Commands
```bash
# Build PVC Manager
make pvc-manager

# Build PVC Manager image
make pvc-manager-image

# Build both provisioner and PVC Manager
make all
```

### Testing
```bash
# Test PVC Manager health
curl http://node-hostname:8080/api/v1/health

# Test volume creation (example)
curl -X POST http://node-hostname:8080/api/v1/volumes/create \
  -H "Content-Type: application/json" \
  -d '{"name":"test-vol","path":"/var/openebs/local/test-vol","nodeAffinityLabels":{"kubernetes.io/hostname":"node-1"},"commands":["mkdir","-p"]}'
```

## Security Considerations

### RBAC Permissions
The PVC Manager requires minimal RBAC permissions:
- Read access to nodes and PVs for validation
- Event creation for logging

### Network Security
- PVC Manager listens on host network for direct access
- Uses HTTP (consider HTTPS for production)
- Only accessible from within the cluster

### Filesystem Security
- Runs with privileged security context (required for filesystem operations)
- Validates paths to prevent directory traversal attacks
- Ensures volumes are created under controlled base paths

## Troubleshooting

### Common Issues

1. **PVC Manager not responding**
   - Check DaemonSet status: `kubectl get ds pvc-manager -n openebs`
   - Check pod logs: `kubectl logs -l app=pvc-manager -n openebs`

2. **Volume creation failures**
   - Check PVC Manager logs for filesystem errors
   - Verify node has sufficient disk space
   - Ensure base path exists and has correct permissions

3. **Network connectivity issues**
   - Verify PVC Manager port is accessible on nodes
   - Check firewall rules
   - Validate service configuration

### Debug Commands
```bash
# Check PVC Manager status
kubectl get pods -l app=pvc-manager -n openebs

# Get PVC Manager logs
kubectl logs -l app=pvc-manager -n openebs -f

# Test connectivity
kubectl exec -it <provisioner-pod> -- curl http://<node-ip>:8080/api/v1/health
```