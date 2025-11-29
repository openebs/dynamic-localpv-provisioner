package app

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// createInitVolumeViaManager creates a volume via the PVC Manager HTTP API instead of using helper pods
func (p *Provisioner) createInitVolumeViaManager(ctx context.Context, pOpts *HelperPodOptions) error {
	klog.Infof("Creating init volume %s via PVC Manager", pOpts.name)

	// Get the Pod IP address from PVC Manager pod
	podIP, err := p.getPVCManagerPodIPFromLabels(pOpts.nodeAffinityLabels)
	if err != nil {
		return fmt.Errorf("failed to get PVC Manager Pod IP: %v", err)
	}

	// Create PVC Manager client using Pod IP address
	pvcManagerURL := GetPVCManagerURL(podIP)
	client := NewPVCManagerClient(pvcManagerURL)

	// Prepare the request
	req := &PVCManagerRequest{
		Name:               pOpts.name,
		Path:               pOpts.path,
		NodeAffinityLabels: pOpts.nodeAffinityLabels,
		FsMode:             "0777", // Default file permissions
		Commands:           pOpts.cmdsForPath,
	}

	// Send create volume request
	if err := client.CreateVolume(ctx, req); err != nil {
		return fmt.Errorf("failed to create volume via PVC Manager: %v", err)
	}

	klog.Infof("Successfully created init volume %s via PVC Manager", pOpts.name)
	return nil
}

// createQuotaViaManager applies quota via the PVC Manager HTTP API instead of using helper pods
func (p *Provisioner) createQuotaViaManager(ctx context.Context, pOpts *HelperPodOptions) error {
	klog.Infof("Applying quota for volume %s via PVC Manager", pOpts.name)

	// Get the Pod IP address from PVC Manager pod
	podIP, err := p.getPVCManagerPodIPFromLabels(pOpts.nodeAffinityLabels)
	if err != nil {
		return fmt.Errorf("failed to get PVC Manager Pod IP: %v", err)
	}

	// Create PVC Manager client using Pod IP address
	pvcManagerURL := GetPVCManagerURL(podIP)
	client := NewPVCManagerClient(pvcManagerURL)

	// Prepare the request
	req := &PVCManagerRequest{
		Name:               pOpts.name,
		Path:               pOpts.path,
		NodeAffinityLabels: pOpts.nodeAffinityLabels,
		SoftLimitGrace:     pOpts.softLimitGrace,
		HardLimitGrace:     pOpts.hardLimitGrace,
		PVCStorage:         pOpts.pvcStorage,
	}

	// Send apply quota request
	if err := client.ApplyQuota(ctx, req); err != nil {
		return fmt.Errorf("failed to apply quota via PVC Manager: %v", err)
	}

	klog.Infof("Successfully applied quota for volume %s via PVC Manager", pOpts.name)
	return nil
}

// createCleanupViaManager deletes a volume via the PVC Manager HTTP API instead of using helper pods
func (p *Provisioner) createCleanupViaManager(ctx context.Context, pOpts *HelperPodOptions) error {
	klog.Infof("Deleting volume %s via PVC Manager", pOpts.name)

	// Get the Pod IP address from PVC Manager pod
	podIP, err := p.getPVCManagerPodIPFromLabels(pOpts.nodeAffinityLabels)
	if err != nil {
		return fmt.Errorf("failed to get PVC Manager Pod IP: %v", err)
	}

	// Create PVC Manager client using Pod IP address
	pvcManagerURL := GetPVCManagerURL(podIP)
	client := NewPVCManagerClient(pvcManagerURL)

	// Prepare the request
	req := &PVCManagerRequest{
		Name:               pOpts.name,
		Path:               pOpts.path,
		NodeAffinityLabels: pOpts.nodeAffinityLabels,
		Commands:           pOpts.cmdsForPath,
	}

	// Send delete volume request
	if err := client.DeleteVolume(ctx, req); err != nil {
		return fmt.Errorf("failed to delete volume via PVC Manager: %v", err)
	}

	klog.Infof("Successfully deleted volume %s via PVC Manager", pOpts.name)
	return nil
}

// getPVCManagerPodIPFromLabels extracts the PVC Manager Pod IP from node affinity labels
func (p *Provisioner) getPVCManagerPodIPFromLabels(nodeAffinityLabels map[string]string) (string, error) {
	// Check if kubeClient is initialized
	if p.kubeClient == nil {
		return "", fmt.Errorf("kubeClient is not initialized")
	}

	// Get the node hostname from node affinity labels
	nodeHostname, err := p.getNodeHostnameFromLabels(nodeAffinityLabels)
	if err != nil {
		return "", fmt.Errorf("failed to get node hostname: %v", err)
	}

	// Find the PVC Manager pod running on this node
	pods, err := p.kubeClient.CoreV1().Pods("openebs").List(
		context.TODO(),
		metav1.ListOptions{
			LabelSelector: "app=pvc-manager",
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeHostname),
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to list PVC Manager pods: %v", err)
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no PVC Manager pod found on node %s", nodeHostname)
	}

	if len(pods.Items) > 1 {
		return "", fmt.Errorf("multiple PVC Manager pods found on node %s", nodeHostname)
	}

	pod := pods.Items[0]
	if pod.Status.PodIP == "" {
		return "", fmt.Errorf("PVC Manager pod on node %s has no Pod IP", nodeHostname)
	}

	return pod.Status.PodIP, nil
}

// getNodeIPFromLabels extracts the node IP address from node affinity labels
func (p *Provisioner) getNodeIPFromLabels(nodeAffinityLabels map[string]string) (string, error) {
	// Get the node object first
	nodeObject, err := p.GetNodeObjectFromLabels(nodeAffinityLabels)
	if err != nil {
		return "", fmt.Errorf("failed to get node object: %v", err)
	}

	// Extract the internal IP address from the node
	for _, address := range nodeObject.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			return address.Address, nil
		}
	}

	// Fallback to external IP if internal IP is not available
	for _, address := range nodeObject.Status.Addresses {
		if address.Type == corev1.NodeExternalIP {
			return address.Address, nil
		}
	}

	return "", fmt.Errorf("no IP address found for node")
}

// getNodeHostnameFromLabels extracts the node hostname from node affinity labels
func (p *Provisioner) getNodeHostnameFromLabels(nodeAffinityLabels map[string]string) (string, error) {
	// Try to get hostname from kubernetes.io/hostname label first
	if hostname, exists := nodeAffinityLabels[k8sNodeLabelKeyHostname]; exists {
		return hostname, nil
	}

	// If not found, try to get the node object and extract hostname
	nodeObject, err := p.GetNodeObjectFromLabels(nodeAffinityLabels)
	if err != nil {
		return "", fmt.Errorf("failed to get node object: %v", err)
	}

	hostname := GetNodeHostname(nodeObject)
	if hostname == "" {
		return "", fmt.Errorf("node hostname is empty")
	}

	return hostname, nil
}

// isPVCManagerEnabled checks if PVC Manager mode is enabled via environment variable
func isPVCManagerEnabled() bool {
	return getPVCManagerEnabled()
}

// GetPVCManagerPort returns the port for PVC Manager service
func GetPVCManagerPort() string {
	return getPVCManagerPort()
}
