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
	// UseHostLock serializes the script via flock on a lockfile in the parent
	// directory. Set for HelperPod mode; NodeDeployment relies on an in-process
	// mutex instead.
	UseHostLock bool
}

// quotaPaths resolves the parent and volume paths from cfg, applying
// HostPathPrefix when set (DaemonSet mode mounts host root at /host).
func quotaPaths(cfg QuotaScriptConfig) (parent, volume string) {
	parent = cfg.ParentDir
	volume = filepath.Join(cfg.ParentDir, cfg.VolumeDir)
	if cfg.HostPathPrefix != "" {
		parent = filepath.Join(cfg.HostPathPrefix, cfg.ParentDir)
		volume = filepath.Join(cfg.HostPathPrefix, cfg.ParentDir, cfg.VolumeDir)
	}
	return parent, volume
}

// quotaLockPrologue returns a flock snippet on $PARENT_PATH/.openebs-quota.lock,
// or "" when disabled. fd 9 is util-linux's conventional lock fd; the kernel
// releases the lock when the shell exits.
func quotaLockPrologue(enabled bool) string {
	if !enabled {
		return ""
	}
	return `# Serialize quota operations on this host.
LOCKFILE="$PARENT_PATH/.openebs-quota.lock"
LOCK_WAIT_SECONDS=60
if ! command -v flock >/dev/null 2>&1; then
    echo "flock(1) not found in PATH; cannot serialize quota operations on $LOCKFILE" >&2
    exit 1
fi
# Pre-flight the lockfile path: exec 9>FILE aborts the shell on failure
# (busybox ash does not honour ||), so probe with touch first.
open_err=$(touch -- "$LOCKFILE" 2>&1) || {
    echo "could not open lockfile $LOCKFILE for writing${open_err:+: $open_err}" >&2
    exit 1
}
exec 9>"$LOCKFILE"
flock_err=$(flock -w "$LOCK_WAIT_SECONDS" 9 2>&1) || {
    echo "could not acquire flock on $LOCKFILE within ${LOCK_WAIT_SECONDS}s${flock_err:+: $flock_err}" >&2
    exit 1
}
`
}

// GenerateQuotaApplyScript generates a shell script to apply filesystem quota.
// It detects the filesystem type (XFS or EXT4) and applies a project quota
// with the configured limits.
func GenerateQuotaApplyScript(cfg QuotaScriptConfig) string {
	parent, volume := quotaPaths(cfg)
	// EXT4 setquota expects plain numbers (KB blocks), not the 'k' suffix.
	extSoft := strings.TrimSuffix(cfg.SoftLimitGrace, "k")
	extHard := strings.TrimSuffix(cfg.HardLimitGrace, "k")

	header := fmt.Sprintf(`set -e
PARENT_PATH="%s"
VOLUME_PATH="%s"
XFS_SOFT="%s"
XFS_HARD="%s"
EXT_SOFT="%s"
EXT_HARD="%s"
`, parent, volume, cfg.SoftLimitGrace, cfg.HardLimitGrace, extSoft, extHard)

	return header + quotaLockPrologue(cfg.UseHostLock) + applyScriptBody
}

// GenerateQuotaCleanupScript generates a shell script to remove filesystem
// quota and delete the volume directory.
func GenerateQuotaCleanupScript(cfg QuotaScriptConfig) string {
	parent, volume := quotaPaths(cfg)

	header := fmt.Sprintf(`set -e
PARENT_PATH="%s"
VOLUME_PATH="%s"
`, parent, volume)

	return header + quotaLockPrologue(cfg.UseHostLock) + cleanupScriptBody
}

// applyScriptBody is the static body of the quota-apply script. All inputs
// arrive via shell variables set by the generated header.
const applyScriptBody = `
FS=$(stat -f -c %T "$VOLUME_PATH")

if [[ "$FS" == "xfs" ]]; then
    # Allocate the next project ID and bind it to the volume.
    PID=$(xfs_quota -x -c 'report -h' "$PARENT_PATH" 2>/dev/null | tail -2 | awk 'NR==1{print substr ($1,2)}+0' || echo "0")
    PID=$((PID + 1))
    xfs_quota -x -c "project -s -p $VOLUME_PATH $PID" "$PARENT_PATH"
    xfs_quota -x -c "limit -p bsoft=$XFS_SOFT bhard=$XFS_HARD $PID" "$PARENT_PATH"
elif [[ "$FS" == "ext2/ext3" ]]; then
    PID=$(repquota -P "$PARENT_PATH" 2>/dev/null | tail -3 | awk 'NR==1{print substr ($1,2)}+0' || echo "0")
    PID=$((PID + 1))
    chattr +P -p $PID "$VOLUME_PATH"
    # setquota -P expects plain numbers (KB blocks).
    setquota -P $PID $EXT_SOFT $EXT_HARD 0 0 "$PARENT_PATH"
else
    echo "Unsupported filesystem type: $FS"
    rm -rf "$VOLUME_PATH"
    exit 1
fi`

// cleanupScriptBody is the static body of the quota-cleanup script. All
// inputs arrive via shell variables set by the generated header.
const cleanupScriptBody = `
FS=$(stat -f -c %T "$VOLUME_PATH" 2>/dev/null || echo "unknown")

if [[ "$FS" == "xfs" ]]; then
    ID=$(xfs_io -c stat "$VOLUME_PATH" 2>/dev/null | awk '/projid/{print $3}' | head -1)
    echo "projid=$ID"
    if [ -n "$ID" ] && [ "$ID" != "0" ]; then
        xfs_io -c "chproj -R 0" "$VOLUME_PATH" 2>/dev/null || true
        xfs_quota -x -c "limit -p bsoft=0 bhard=0 $ID" "$PARENT_PATH" 2>/dev/null || true
    fi
elif [[ "$FS" == "ext2/ext3" ]]; then
    ID=$(lsattr -pd "$VOLUME_PATH"/ 2>/dev/null | awk '{print $1}')
    if [ -n "$ID" ] && [ "$ID" != "0" ]; then
        setquota -P $ID 0 0 0 0 "$PARENT_PATH" 2>/dev/null || true
    fi
fi

rm -rf "$VOLUME_PATH"`
