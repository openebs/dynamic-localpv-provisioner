#!/bin/bash

# Test script for PVC Manager architecture
# This script validates that the PVC Manager service is working correctly

set -e

echo "=== OpenEBS LocalPV PVC Manager Test Suite ==="
echo

# Configuration
NAMESPACE="openebs"
PVC_MANAGER_PORT="8080"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl not found. Please install kubectl."
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        log_error "curl not found. Please install curl."
        exit 1
    fi
    
    log_info "Dependencies check passed"
}

check_namespace() {
    log_info "Checking OpenEBS namespace..."
    
    if ! kubectl get namespace $NAMESPACE &> /dev/null; then
        log_error "OpenEBS namespace '$NAMESPACE' not found"
        exit 1
    fi
    
    log_info "OpenEBS namespace exists"
}

check_pvc_manager_daemonset() {
    log_info "Checking PVC Manager DaemonSet..."
    
    # Check if DaemonSet exists
    if ! kubectl get daemonset pvc-manager -n $NAMESPACE &> /dev/null; then
        log_error "PVC Manager DaemonSet not found"
        exit 1
    fi
    
    # Check DaemonSet status
    DESIRED=$(kubectl get daemonset pvc-manager -n $NAMESPACE -o jsonpath='{.status.desiredNumberScheduled}')
    READY=$(kubectl get daemonset pvc-manager -n $NAMESPACE -o jsonpath='{.status.numberReady}')
    
    if [ "$DESIRED" != "$READY" ]; then
        log_error "PVC Manager DaemonSet not ready. Desired: $DESIRED, Ready: $READY"
        kubectl get pods -l app=pvc-manager -n $NAMESPACE
        exit 1
    fi
    
    log_info "PVC Manager DaemonSet is ready ($READY/$DESIRED pods)"
}

check_pvc_manager_health() {
    log_info "Checking PVC Manager health endpoints..."
    
    # Get list of PVC Manager pods with their IPs
    PODS=$(kubectl get pods -l app=pvc-manager -n $NAMESPACE -o jsonpath='{range .items[*]}{.metadata.name}{","}{.status.podIP}{"\n"}{end}')
    
    if [ -z "$PODS" ]; then
        log_error "No PVC Manager pods found"
        exit 1
    fi
    
    local success_count=0
    local total_count=0
    
    echo "$PODS" | while IFS=',' read -r pod_name pod_ip; do
        [ -z "$pod_name" ] && continue
        total_count=$((total_count + 1))
        log_info "Testing PVC Manager health for pod: $pod_name"
        
        if [ -z "$pod_ip" ]; then
            log_warn "Could not get IP for pod $pod_name"
            continue
        fi
        
        # Test health endpoint using pod IP
        if curl -s -f "http://$pod_ip:$PVC_MANAGER_PORT/api/v1/health" > /dev/null; then
            log_info "✓ Health check passed for pod $pod_name ($pod_ip)"
            success_count=$((success_count + 1))
        else
            log_warn "✗ Health check failed for pod $pod_name ($pod_ip)"
        fi
    done
    
    # Alternative: Test via service if available
    if kubectl get service pvc-manager -n $NAMESPACE &> /dev/null; then
        log_info "Testing PVC Manager service endpoint..."
        if kubectl exec -n $NAMESPACE deployment/openebs-localpv-provisioner -- curl -s -f "http://pvc-manager.$NAMESPACE.svc.cluster.local:$PVC_MANAGER_PORT/api/v1/health" > /dev/null 2>&1; then
            log_info "✓ Service health check passed"
        else
            log_warn "✗ Service health check failed"
        fi
    fi
    
    log_info "PVC Manager health checks completed"
}

check_provisioner() {
    log_info "Checking LocalPV Provisioner..."
    
    # Check if provisioner deployment exists
    if ! kubectl get deployment openebs-localpv-provisioner -n $NAMESPACE &> /dev/null; then
        log_error "LocalPV Provisioner deployment not found"
        exit 1
    fi
    
    # Check provisioner status
    REPLICAS=$(kubectl get deployment openebs-localpv-provisioner -n $NAMESPACE -o jsonpath='{.status.replicas}')
    READY_REPLICAS=$(kubectl get deployment openebs-localpv-provisioner -n $NAMESPACE -o jsonpath='{.status.readyReplicas}')
    
    if [ "$REPLICAS" != "$READY_REPLICAS" ]; then
        log_error "LocalPV Provisioner not ready. Replicas: $REPLICAS, Ready: $READY_REPLICAS"
        exit 1
    fi
    
    log_info "LocalPV Provisioner is ready"
}

test_pvc_creation() {
    log_info "Testing PVC creation with PVC Manager..."
    
    # Create test PVC
    local test_pvc_name="test-pvc-manager-$(date +%s)"
    
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: $test_pvc_name
  namespace: default
spec:
  storageClassName: openebs-hostpath-pvc-manager
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
EOF

    # Wait for PVC to be bound
    log_info "Waiting for PVC to be bound..."
    local timeout=60
    local elapsed=0
    
    while [ $elapsed -lt $timeout ]; do
        local status=$(kubectl get pvc $test_pvc_name -n default -o jsonpath='{.status.phase}' 2>/dev/null || echo "")
        
        if [ "$status" = "Bound" ]; then
            log_info "✓ PVC $test_pvc_name successfully bound"
            break
        elif [ "$status" = "Pending" ]; then
            log_info "PVC $test_pvc_name is pending..."
            sleep 5
            elapsed=$((elapsed + 5))
        else
            log_error "PVC $test_pvc_name has unexpected status: $status"
            kubectl describe pvc $test_pvc_name -n default
            cleanup_test_pvc $test_pvc_name
            exit 1
        fi
    done
    
    if [ $elapsed -ge $timeout ]; then
        log_error "Timeout waiting for PVC to be bound"
        kubectl describe pvc $test_pvc_name -n default
        cleanup_test_pvc $test_pvc_name
        exit 1
    fi
    
    # Cleanup test PVC
    cleanup_test_pvc $test_pvc_name
    log_info "PVC creation test passed"
}

cleanup_test_pvc() {
    local pvc_name=$1
    log_info "Cleaning up test PVC: $pvc_name"
    kubectl delete pvc $pvc_name -n default --ignore-not-found=true
    
    # Wait for PVC to be deleted
    local timeout=30
    local elapsed=0
    
    while kubectl get pvc $pvc_name -n default &> /dev/null && [ $elapsed -lt $timeout ]; do
        sleep 2
        elapsed=$((elapsed + 2))
    done
}

show_logs() {
    log_info "Showing recent PVC Manager logs..."
    kubectl logs -l app=pvc-manager -n $NAMESPACE --tail=10 --prefix=true || true
    
    echo
    log_info "Showing recent Provisioner logs..."
    kubectl logs -l name=openebs-localpv-provisioner -n $NAMESPACE --tail=10 --prefix=true || true
}

main() {
    echo "Starting PVC Manager validation tests..."
    echo
    
    check_dependencies
    check_namespace
    check_pvc_manager_daemonset
    check_pvc_manager_health
    check_provisioner
    test_pvc_creation
    
    echo
    log_info "=== All tests passed! PVC Manager is working correctly ==="
    echo
    
    if [ "${1:-}" = "--show-logs" ]; then
        show_logs
    fi
}

# Handle cleanup on script exit
cleanup() {
    if [ -n "${test_pvc_name:-}" ]; then
        cleanup_test_pvc $test_pvc_name
    fi
}

trap cleanup EXIT

# Run tests
main "$@"