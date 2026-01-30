package app

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"k8s.io/klog/v2"

	hostpath "github.com/openebs/dynamic-localpv-provisioner/pkg/hostpath/v1alpha1"
)

// VolumeRequest represents a request for volume operations
type VolumeRequest struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	FsMode         string `json:"fsMode,omitempty"`
	SoftLimitGrace string `json:"softLimitGrace,omitempty"`
	HardLimitGrace string `json:"hardLimitGrace,omitempty"`
	PVCStorage     int64  `json:"pvcStorage,omitempty"`
}

// LocalVolumeManager handles volume operations on the local node
type LocalVolumeManager struct {
	// Add any necessary fields for volume management

	// race condition protection
	// mutex for thread-safe operations
	mu *sync.Mutex
}

const (
	// HostPathPrefix is the mount point where the host root filesystem is mounted
	// in the node DaemonSet. This allows the provisioner to access any path on the host.
	HostPathPrefix = "/host"
)

// NewLocalVolumeManager creates a new LocalVolumeManager instance
func NewLocalVolumeManager() *LocalVolumeManager {
	return &LocalVolumeManager{
		mu: &sync.Mutex{},
	}
}

// CreateVolume creates a new volume directory on the local node
func (vm *LocalVolumeManager) CreateVolume(ctx context.Context, req *VolumeRequest, enableQuota bool) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	klog.Infof("Creating volume %s at path %s", req.Name, req.Path)

	// Extract the base path and the volume unique path
	parentDir, volumeDir, err := ExtractPaths(req.Path)
	if err != nil {
		return fmt.Errorf("failed to extract paths: %v", err)
	}

	// Set default file permissions if not specified
	fsMode := req.FsMode
	if fsMode == "" {
		fsMode = "0777"
	}

	// Create the directory with specified permissions
	// Use HostPathPrefix to access the host filesystem
	fullPath := filepath.Join(parentDir, volumeDir)
	hostFullPath := filepath.Join(HostPathPrefix, fullPath)
	if err := vm.executeCommand(ctx, "mkdir", "-m", fsMode, "-p", hostFullPath); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	if enableQuota {
		// Apply quota if enabled
		if err := vm.ApplyQuota(ctx, req); err != nil {
			return fmt.Errorf("failed to apply quota: %v", err)
		}
	}

	klog.Infof("Successfully created volume %s at path %s", req.Name, fullPath)
	return nil
}

// DeleteVolume removes a volume directory from the local node
func (vm *LocalVolumeManager) DeleteVolume(ctx context.Context, req *VolumeRequest) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	klog.Infof("Deleting volume %s at path %s", req.Name, req.Path)

	// Extract the base path and the volume unique path
	parentDir, volumeDir, err := ExtractPaths(req.Path)
	if err != nil {
		return fmt.Errorf("failed to extract paths: %v", err)
	}

	// Generate cleanup script using shared utility
	// Node deployment mode accesses host filesystem via HostPathPrefix
	cleanupScript := GenerateQuotaCleanupScript(QuotaScriptConfig{
		ParentDir:      parentDir,
		VolumeDir:      volumeDir,
		HostPathPrefix: HostPathPrefix,
	})

	if err := vm.executeCommand(ctx, "sh", "-c", cleanupScript); err != nil {
		return fmt.Errorf("failed to delete directory: %v", err)
	}

	fullPath := filepath.Join(parentDir, volumeDir)
	klog.Infof("Successfully deleted volume %s at path %s", req.Name, fullPath)
	return nil
}

// ApplyQuota applies filesystem quota to a volume
func (vm *LocalVolumeManager) ApplyQuota(ctx context.Context, req *VolumeRequest) error {
	klog.Infof("Applying quota for volume %s at path %s", req.Name, req.Path)

	// Extract the base path and the volume unique path
	parentDir, volumeDir, err := ExtractPaths(req.Path)
	if err != nil {
		return fmt.Errorf("failed to extract paths: %v", err)
	}

	// Convert limits to kilobytes
	softLimitGrace, err := vm.convertToK(req.SoftLimitGrace, req.PVCStorage)
	if err != nil {
		return fmt.Errorf("failed to convert soft limit: %v", err)
	}

	hardLimitGrace, err := vm.convertToK(req.HardLimitGrace, req.PVCStorage)
	if err != nil {
		return fmt.Errorf("failed to convert hard limit: %v", err)
	}

	// Validate limits
	if err := vm.validateLimits(softLimitGrace, hardLimitGrace, req.PVCStorage); err != nil {
		return fmt.Errorf("invalid limits: %v", err)
	}

	// Apply quota based on filesystem type
	if err := vm.applyQuotaByFilesystem(ctx, parentDir, volumeDir, softLimitGrace, hardLimitGrace); err != nil {
		return fmt.Errorf("failed to apply quota: %v", err)
	}

	klog.Infof("Successfully applied quota for volume %s", req.Name)
	return nil
}

// extractPaths extracts parent directory and volume directory from the full path
func ExtractPaths(fullPath string) (parentDir, volumeDir string, err error) {
	// Use hostpath builder to validate and extract paths
	return hostpath.NewBuilder().WithPath(fullPath).
		WithCheckf(hostpath.IsNonRoot(), "volume directory {%v} should not be under root directory", fullPath).
		ExtractSubPath()
}

// executeCommand executes a system command with timeout
func (vm *LocalVolumeManager) executeCommand(ctx context.Context, name string, args ...string) error {
	// Create command context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, name, args...)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				// Check for specific error conditions
				switch status.ExitStatus() {
				case 127: // Command not found
					return fmt.Errorf("command not found: %s - please ensure required quota tools are installed", name)
				case 1: // General error
					if strings.Contains(outputStr, "Unsupported filesystem type") {
						return fmt.Errorf("unsupported filesystem type - please ensure the filesystem supports quotas and is properly mounted with quota options")
					}
					return fmt.Errorf("command failed with exit code %d: %s", status.ExitStatus(), outputStr)
				default:
					return fmt.Errorf("command failed with exit code %d: %s", status.ExitStatus(), outputStr)
				}
			}
		}
		return fmt.Errorf("command failed: %v, output: %s", err, outputStr)
	}

	return nil
}

// convertToK converts the limits to kilobytes
func (vm *LocalVolumeManager) convertToK(limit string, pvcStorage int64) (string, error) {
	if len(limit) == 0 {
		return "0k", nil
	}

	valueRegex := regexp.MustCompile(`[\d]*[\.]?[\d]*`)
	valueString := valueRegex.FindString(limit)

	if limit != valueString+"%" {
		return "", fmt.Errorf("invalid format for limit grace")
	}

	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return "", fmt.Errorf("invalid format, cannot parse")
	}

	if value > 100 {
		value = 100
	}

	value *= float64(pvcStorage)
	value /= 100
	value += float64(pvcStorage)
	value /= 1024

	value = math.Ceil(value)
	valueString = strconv.FormatFloat(value, 'f', -1, 64)
	valueString += "k"
	return valueString, nil
}

// validateLimits validates quota limits
func (vm *LocalVolumeManager) validateLimits(softLimitGrace, hardLimitGrace string, pvcStorage int64) error {
	if softLimitGrace == "0k" && hardLimitGrace == "0k" {
		// Use PVC storage as both limits
		pvcStorageInK := math.Ceil(float64(pvcStorage) / 1024)
		pvcStorageInKString := strconv.FormatFloat(pvcStorageInK, 'f', -1, 64) + "k"
		softLimitGrace = pvcStorageInKString
		hardLimitGrace = pvcStorageInKString
		return nil
	}

	if softLimitGrace == "0k" || hardLimitGrace == "0k" {
		return nil
	}

	if len(softLimitGrace) > len(hardLimitGrace) ||
		(len(softLimitGrace) == len(hardLimitGrace) && softLimitGrace > hardLimitGrace) {
		return fmt.Errorf("hard limit cannot be smaller than soft limit")
	}

	return nil
}

// applyQuotaByFilesystem applies quota based on the filesystem type
func (vm *LocalVolumeManager) applyQuotaByFilesystem(ctx context.Context, parentDir, volumeDir, softLimitGrace, hardLimitGrace string) error {
	// Generate quota script using shared utility
	// Node deployment mode accesses host filesystem via HostPathPrefix
	script := GenerateQuotaApplyScript(QuotaScriptConfig{
		ParentDir:      parentDir,
		VolumeDir:      volumeDir,
		SoftLimitGrace: softLimitGrace,
		HardLimitGrace: hardLimitGrace,
		HostPathPrefix: HostPathPrefix,
	})

	// Execute the quota script
	return vm.executeCommand(ctx, "sh", "-c", script)
}
