package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"k8s.io/klog/v2"
)

// PVCManagerClient represents the HTTP client for communicating with PVC Manager
type PVCManagerClient struct {
	client  *http.Client
	baseURL string
}

// PVCManagerRequest represents a request to the PVC Manager
type PVCManagerRequest struct {
	Name               string            `json:"name"`
	Path               string            `json:"path"`
	NodeAffinityLabels map[string]string `json:"nodeAffinityLabels"`
	FsMode             string            `json:"fsMode,omitempty"`
	Commands           []string          `json:"commands"`
	SoftLimitGrace     string            `json:"softLimitGrace,omitempty"`
	HardLimitGrace     string            `json:"hardLimitGrace,omitempty"`
	PVCStorage         int64             `json:"pvcStorage,omitempty"`
}

// PVCManagerResponse represents a response from the PVC Manager
type PVCManagerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// NewPVCManagerClient creates a new PVC Manager client
func NewPVCManagerClient(baseURL string) *PVCManagerClient {
	return &PVCManagerClient{
		client: &http.Client{
			Timeout: 60 * time.Second, // Increased timeout for volume operations
		},
		baseURL: baseURL,
	}
}

// CreateVolume sends a request to create a volume via PVC Manager
func (c *PVCManagerClient) CreateVolume(ctx context.Context, req *PVCManagerRequest) error {
	url := fmt.Sprintf("%s/api/v1/volumes/create", c.baseURL)
	return c.sendRequest(ctx, "POST", url, req)
}

// DeleteVolume sends a request to delete a volume via PVC Manager
func (c *PVCManagerClient) DeleteVolume(ctx context.Context, req *PVCManagerRequest) error {
	url := fmt.Sprintf("%s/api/v1/volumes/delete", c.baseURL)
	return c.sendRequest(ctx, "POST", url, req)
}

// ApplyQuota sends a request to apply quota via PVC Manager
func (c *PVCManagerClient) ApplyQuota(ctx context.Context, req *PVCManagerRequest) error {
	url := fmt.Sprintf("%s/api/v1/volumes/quota", c.baseURL)
	return c.sendRequest(ctx, "POST", url, req)
}

// HealthCheck performs a health check on the PVC Manager
func (c *PVCManagerClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/health", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %v", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// sendRequest sends an HTTP request to the PVC Manager
func (c *PVCManagerClient) sendRequest(ctx context.Context, method, url string, reqData *PVCManagerRequest) error {
	// Marshal request data to JSON
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	klog.V(4).Infof("Sending %s request to %s", method, url)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// Check status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var response PVCManagerResponse
		if err := json.Unmarshal(body, &response); err != nil {
			klog.Warningf("Failed to parse response JSON: %v", err)
			return nil // Request succeeded even if we can't parse the response
		}

		if !response.Success {
			return fmt.Errorf("PVC Manager operation failed: %s", response.Error)
		}

		klog.V(4).Infof("PVC Manager operation succeeded: %s", response.Message)
		return nil
	}

	// Handle error response
	var errorResponse struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &errorResponse); err == nil {
		return fmt.Errorf("PVC Manager error (status %d): %s - %s", resp.StatusCode, errorResponse.Error, errorResponse.Message)
	}

	return fmt.Errorf("PVC Manager request failed with status %d: %s", resp.StatusCode, string(body))
}

// GetPVCManagerURL constructs the PVC Manager URL for a given Pod IP
func GetPVCManagerURL(podIPOrNodeIP string) string {
	// The PVC Manager will be running as a DaemonSet on each node
	// We can access it via the pod's IP address or node IP address on the configured port
	// Using Pod IP address is preferred for direct pod communication
	port := GetPVCManagerPort()
	return fmt.Sprintf("http://%s:%s", podIPOrNodeIP, port)
}
