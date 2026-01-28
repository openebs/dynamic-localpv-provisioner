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

	hostpath "github.com/openebs/maya/pkg/hostpath/v1alpha1"
	"k8s.io/klog/v2"
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
	fullPath := filepath.Join(parentDir, volumeDir)
	if err := vm.executeCommand(ctx, "mkdir", "-m", fsMode, "-p", fullPath); err != nil {
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

	// Remove the directory
	fullPath := filepath.Join(parentDir, volumeDir)

	// check if path is xfs quota enabled and remove quota projid
	cleanupCmdsForPath := fmt.Sprintf(`
        d="%s"
        base="%s"
        # check fs type first
        fs=$(stat -f -c %%T $base 2>/dev/null)
        if [ "$fs" = "xfs" ]; then
			id=$(xfs_io -c stat $d 2>/dev/null | awk '/projid/{print $3}' | head -1)
			echo "projid=$id"
			if [ -n "$id" ] && [ "$id" != "0" ]; then
				# remove projid binding
				xfs_io -c "chproj -R 0" "$d" 2>/dev/null || true
				# remove quota limit
				xfs_quota -x -c "limit -p bsoft=0 bhard=0 $id" $base 2>/dev/null || true
			fi
        elif [ "$fs" = "ext2/ext3" ]; then
			ID=$(lsattr -pd $d/ | awk '{print $1}')
			if [ -n "$ID" ] && [ "$ID" != "0" ]; then
				setquota -P $ID 0 0 0 0 $base 2>/dev/null || true
			fi
		fi
		rm -rf $d
	`, fullPath, parentDir)

	if err := vm.executeCommand(ctx, "sh", "-c", cleanupCmdsForPath); err != nil {
		return fmt.Errorf("failed to delete directory: %v", err)
	}

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
	// Create a shell script to detect filesystem and apply quota
	// We need to find the actual XFS mount point since parentDir might be a bind mount

	// Extract numeric values for EXT4 setquota (it expects plain numbers in KB blocks, not with suffix)
	extSoftLimit := strings.TrimSuffix(softLimitGrace, "k")
	extHardLimit := strings.TrimSuffix(hardLimitGrace, "k")

	script := fmt.Sprintf(`
		set -e
		
		# Find the actual mount point for the path using host's mount info
		# The host's proc is mounted at /host/proc for container environments
		if [ -f /host/proc/1/mountinfo ]; then
			# Use host's mountinfo to find the real mount point
			MOUNT_POINT=$(findmnt -n -o TARGET --target %s --mountinfo /host/proc/1/mountinfo 2>/dev/null || findmnt -n -o TARGET --target %s 2>/dev/null || echo %s)
		else
			MOUNT_POINT=$(findmnt -n -o TARGET --target %s 2>/dev/null || echo %s)
		fi
		
		# Get filesystem type
		FS=$(stat -f -c %%T %s)
		
		if [[ "$FS" == "xfs" ]]; then
			# Get the next available project ID
			PID=$(xfs_quota -x -c 'report -h' "$MOUNT_POINT" 2>/dev/null | tail -2 | awk 'NR==1{print substr ($1,2)}+0' || echo "0")
			PID=$((PID + 1))
			# Set up project for the volume directory
			xfs_quota -x -c "project -s -p %s $PID" "$MOUNT_POINT"
			# Apply quota limits
			xfs_quota -x -c "limit -p bsoft=%s bhard=%s $PID" "$MOUNT_POINT"
		elif [[ "$FS" == "ext2/ext3" ]]; then
			PID=$(repquota -P "$MOUNT_POINT" 2>/dev/null | tail -3 | awk 'NR==1{print substr ($1,2)}+0' || echo "0")
			PID=$((PID + 1))
			chattr +P -p $PID %s
			# setquota -P expects block limits as plain numbers (in KB blocks)
			setquota -P $PID %s %s 0 0 "$MOUNT_POINT"
		else
			echo "Unsupported filesystem type: $FS"
			exit 1
		fi`,
		parentDir, parentDir, parentDir, // findmnt with host mountinfo, fallback, and default
		parentDir, parentDir, // findmnt without host mountinfo
		parentDir,                           // stat filesystem
		filepath.Join(parentDir, volumeDir), // project path
		softLimitGrace, hardLimitGrace,      // xfs limits (with 'k' suffix)
		filepath.Join(parentDir, volumeDir), // chattr path
		extSoftLimit, extHardLimit,          // ext quota limits (plain numbers in KB)
	)

	// Execute the quota script
	return vm.executeCommand(ctx, "sh", "-c", script)
}
