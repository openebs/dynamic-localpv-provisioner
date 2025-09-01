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
	"syscall"
	"time"

	hostpath "github.com/openebs/maya/pkg/hostpath/v1alpha1"
	"k8s.io/klog/v2"
)

// VolumeManager handles volume operations on the local node
type VolumeManager struct {
	// Add any necessary fields for volume management
}

// NewVolumeManager creates a new VolumeManager instance
func NewVolumeManager() *VolumeManager {
	return &VolumeManager{}
}

// CreateVolume creates a new volume directory on the local node
func (vm *VolumeManager) CreateVolume(ctx context.Context, req *VolumeRequest) error {
	klog.Infof("Creating volume %s at path %s", req.Name, req.Path)

	// Extract the base path and the volume unique path
	parentDir, volumeDir, err := vm.extractPaths(req.Path)
	if err != nil {
		return fmt.Errorf("failed to extract paths: %v", err)
	}

	// Set default file permissions if not specified
	fsMode := req.FsMode
	if fsMode == "" {
		fsMode = "0777"
	}

	// Create the directory with specified permissions
	fullPath := filepath.Join(parentDir, volumeDir)
	if err := vm.executeCommand(ctx, "mkdir", "-m", fsMode, "-p", fullPath); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	klog.Infof("Successfully created volume %s at path %s", req.Name, fullPath)
	return nil
}

// DeleteVolume removes a volume directory from the local node
func (vm *VolumeManager) DeleteVolume(ctx context.Context, req *VolumeRequest) error {
	klog.Infof("Deleting volume %s at path %s", req.Name, req.Path)

	// Extract the base path and the volume unique path
	parentDir, volumeDir, err := vm.extractPaths(req.Path)
	if err != nil {
		return fmt.Errorf("failed to extract paths: %v", err)
	}

	// Remove the directory
	fullPath := filepath.Join(parentDir, volumeDir)
	if err := vm.executeCommand(ctx, "rm", "-rf", fullPath); err != nil {
		return fmt.Errorf("failed to delete directory: %v", err)
	}

	klog.Infof("Successfully deleted volume %s at path %s", req.Name, fullPath)
	return nil
}

// ApplyQuota applies filesystem quota to a volume
func (vm *VolumeManager) ApplyQuota(ctx context.Context, req *VolumeRequest) error {
	klog.Infof("Applying quota for volume %s at path %s", req.Name, req.Path)

	// Extract the base path and the volume unique path
	parentDir, volumeDir, err := vm.extractPaths(req.Path)
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
func (vm *VolumeManager) extractPaths(fullPath string) (parentDir, volumeDir string, err error) {
	// Use hostpath builder to validate and extract paths
	return hostpath.NewBuilder().WithPath(fullPath).
		WithCheckf(hostpath.IsNonRoot(), "volume directory {%v} should not be under root directory", fullPath).
		ExtractSubPath()
}

// executeCommand executes a system command with timeout
func (vm *VolumeManager) executeCommand(ctx context.Context, name string, args ...string) error {
	// Create command context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, name, args...)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("command failed with exit code %d: %s", status.ExitStatus(), string(output))
			}
		}
		return fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}

	return nil
}

// convertToK converts the limits to kilobytes
func (vm *VolumeManager) convertToK(limit string, pvcStorage int64) (string, error) {
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
func (vm *VolumeManager) validateLimits(softLimitGrace, hardLimitGrace string, pvcStorage int64) error {
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
func (vm *VolumeManager) applyQuotaByFilesystem(ctx context.Context, parentDir, volumeDir, softLimitGrace, hardLimitGrace string) error {
	// Create a shell script to detect filesystem and apply quota
	script := fmt.Sprintf(`
		FS=$(stat -f -c %%T %s)
		if [[ "$FS" == "xfs" ]]; then
			PID=$(xfs_quota -x -c 'report -h' %s | tail -2 | awk 'NR==1{print substr ($1,2)}+0')
			PID=$((PID + 1))
			xfs_quota -x -c "project -s -p %s $PID" %s
			xfs_quota -x -c "limit -p bsoft=%s bhard=%s $PID" %s
		elif [[ "$FS" == "ext2/ext3" ]]; then
			PID=$(repquota -P %s | tail -3 | awk 'NR==1{print substr ($1,2)}+0')
			PID=$((PID + 1))
			chattr +P -p $PID %s
			setquota -P $PID %s %s 0 0 %s
		else
			echo "Unsupported filesystem type: $FS"
			exit 1
		fi`,
		parentDir,                           // stat filesystem
		parentDir,                           // xfs_quota report
		filepath.Join(parentDir, volumeDir), // project path
		parentDir,                           // project base
		softLimitGrace, hardLimitGrace,      // xfs limits
		parentDir,                                                        // xfs base
		parentDir,                                                        // repquota
		filepath.Join(parentDir, volumeDir),                              // chattr path
		strings.ToUpper(softLimitGrace), strings.ToUpper(hardLimitGrace), // ext quota limits
		parentDir, // setquota base
	)

	// Execute the quota script
	return vm.executeCommand(ctx, "sh", "-c", script)
}
