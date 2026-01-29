package app

import (
	"fmt"
	"path/filepath"
	"strings"
)

// QuotaScriptConfig holds configuration for generating quota scripts
type QuotaScriptConfig struct {
	// ParentDir is the base directory path where volumes are created
	// For HelperPod: use "/data" (the mount point inside the container)
	// For DaemonSet: use the actual host path (e.g., "/var/openebs/local")
	ParentDir string
	// VolumeDir is the volume subdirectory name (e.g., "pvc-xxx")
	VolumeDir string
	// SoftLimitGrace is the soft quota limit with 'k' suffix (e.g., "1024k")
	SoftLimitGrace string
	// HardLimitGrace is the hard quota limit with 'k' suffix (e.g., "1024k")
	HardLimitGrace string
	// HostPathPrefix is prepended to paths to access them from within the container
	// For HelperPod: use "" (parentDir is already mounted at /data)
	// For DaemonSet: use "/host" (host root is mounted at /host)
	HostPathPrefix string
}

// GenerateQuotaApplyScript generates a shell script to apply filesystem quota
// This script:
// 1. Detects filesystem type (XFS or EXT4)
// 2. Applies project quota with the specified limits
func GenerateQuotaApplyScript(cfg QuotaScriptConfig) string {
	// Extract numeric values for EXT4 setquota (it expects plain numbers in KB blocks, not with suffix)
	extSoftLimit := strings.TrimSuffix(cfg.SoftLimitGrace, "k")
	extHardLimit := strings.TrimSuffix(cfg.HardLimitGrace, "k")

	// Build paths with optional prefix
	parentPath := cfg.ParentDir
	volumePath := filepath.Join(cfg.ParentDir, cfg.VolumeDir)
	if cfg.HostPathPrefix != "" {
		parentPath = filepath.Join(cfg.HostPathPrefix, cfg.ParentDir)
		volumePath = filepath.Join(cfg.HostPathPrefix, cfg.ParentDir, cfg.VolumeDir)
	}

	return fmt.Sprintf(`set -e

# Path to parent directory (mount point for quota commands)
PARENT_PATH="%s"
# Path to volume directory
VOLUME_PATH="%s"

# Get filesystem type
FS=$(stat -f -c %%T "$VOLUME_PATH")

if [[ "$FS" == "xfs" ]]; then
    # Get the next available project ID
    PID=$(xfs_quota -x -c 'report -h' "$PARENT_PATH" 2>/dev/null | tail -2 | awk 'NR==1{print substr ($1,2)}+0' || echo "0")
    PID=$((PID + 1))
    # Set up project for the volume directory
    xfs_quota -x -c "project -s -p $VOLUME_PATH $PID" "$PARENT_PATH"
    # Apply quota limits
    xfs_quota -x -c "limit -p bsoft=%s bhard=%s $PID" "$PARENT_PATH"
elif [[ "$FS" == "ext2/ext3" ]]; then
    PID=$(repquota -P "$PARENT_PATH" 2>/dev/null | tail -3 | awk 'NR==1{print substr ($1,2)}+0' || echo "0")
    PID=$((PID + 1))
    chattr +P -p $PID "$VOLUME_PATH"
    # setquota -P expects block limits as plain numbers (in KB blocks)
    setquota -P $PID %s %s 0 0 "$PARENT_PATH"
else
    echo "Unsupported filesystem type: $FS"
    rm -rf "$VOLUME_PATH"
    exit 1
fi`,
		parentPath, volumePath,
		cfg.SoftLimitGrace, cfg.HardLimitGrace, // xfs limits (with 'k' suffix)
		extSoftLimit, extHardLimit, // ext quota limits (plain numbers in KB)
	)
}

// GenerateQuotaCleanupScript generates a shell script to remove filesystem quota and delete volume
// This script:
// 1. Detects filesystem type (XFS or EXT4)
// 2. Removes project quota association
// 3. Deletes the volume directory
func GenerateQuotaCleanupScript(cfg QuotaScriptConfig) string {
	// Build paths with optional prefix
	parentPath := cfg.ParentDir
	volumePath := filepath.Join(cfg.ParentDir, cfg.VolumeDir)
	if cfg.HostPathPrefix != "" {
		parentPath = filepath.Join(cfg.HostPathPrefix, cfg.ParentDir)
		volumePath = filepath.Join(cfg.HostPathPrefix, cfg.ParentDir, cfg.VolumeDir)
	}

	return fmt.Sprintf(`set -e

# Path to parent directory (mount point for quota commands)
PARENT_PATH="%s"
# Path to volume directory
VOLUME_PATH="%s"

# Get filesystem type
FS=$(stat -f -c %%T "$VOLUME_PATH" 2>/dev/null || echo "unknown")

if [[ "$FS" == "xfs" ]]; then
    ID=$(xfs_io -c stat "$VOLUME_PATH" 2>/dev/null | awk '/projid/{print $3}' | head -1)
    echo "projid=$ID"
    if [ -n "$ID" ] && [ "$ID" != "0" ]; then
        # Remove projid binding
        xfs_io -c "chproj -R 0" "$VOLUME_PATH" 2>/dev/null || true
        # Remove quota limit
        xfs_quota -x -c "limit -p bsoft=0 bhard=0 $ID" "$PARENT_PATH" 2>/dev/null || true
    fi
elif [[ "$FS" == "ext2/ext3" ]]; then
    ID=$(lsattr -pd "$VOLUME_PATH"/ 2>/dev/null | awk '{print $1}')
    if [ -n "$ID" ] && [ "$ID" != "0" ]; then
        setquota -P $ID 0 0 0 0 "$PARENT_PATH" 2>/dev/null || true
    fi
fi

rm -rf "$VOLUME_PATH"`,
		parentPath, volumePath,
	)
}
