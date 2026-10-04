package app

import (
	"fmt"
	"path/filepath"
	"strings"
)

// QuotaScriptConfig holds configuration for generating quota scripts.
//
// Path model:
//
//   - ParentDir is always the path as seen on the host (e.g. /var/openebs/local).
//   - When HostPathPrefix is set (typically "/host"), the script runs quota tools
//     via nsenter into the host mount namespace so XFS/EXT4 project quota resolves
//     the real host mount. Bind-mounted pod paths alone are not sufficient.
//   - CONTAINER_PARENT_PATH is used only for flock lockfiles, which must be
//     opened from the container mount namespace. It is ContainerParentDir
//     when set; otherwise HostPathPrefix+ParentDir, or ParentDir when the
//     prefix is empty.
type QuotaScriptConfig struct {
	// ParentDir is the base directory path as seen on the host
	// (e.g., "/var/openebs/local").
	ParentDir string
	// VolumeDir is the volume subdirectory name (e.g., "pvc-xxx")
	VolumeDir string
	// SoftLimitGrace is the soft quota limit with 'k' suffix (e.g., "1024k")
	SoftLimitGrace string
	// HardLimitGrace is the hard quota limit with 'k' suffix (e.g., "1024k")
	HardLimitGrace string
	// HostPathPrefix is the container path prefix used to reach the host
	// mount namespace: $HostPathPrefix/proc/1/ns/mnt.
	// DaemonSet mounts the host root at "/host". Helper pods mount only
	// host /proc at "/host/proc". Empty disables nsenter (paths are used
	// as-is in the current mount namespace).
	HostPathPrefix string
	// ContainerParentDir, when set, is the flock lockfile directory as seen
	// in the container (e.g. HelperPod "/data", which already bind-mounts
	// ParentDir). When empty, the path is derived from HostPathPrefix and
	// ParentDir.
	ContainerParentDir string
	// UseHostLock serializes the script via flock on a lockfile in the parent
	// directory. Set for HelperPod mode; NodeDeployment relies on an in-process
	// mutex instead.
	UseHostLock bool
}

// quotaPaths resolves host and container-visible paths from cfg.
// hostParent/hostVolume are used with nsenter; containerParent is for flock.
func quotaPaths(cfg QuotaScriptConfig) (hostParent, hostVolume, containerParent string) {
	hostParent = cfg.ParentDir
	hostVolume = filepath.Join(cfg.ParentDir, cfg.VolumeDir)
	switch {
	case cfg.ContainerParentDir != "":
		containerParent = cfg.ContainerParentDir
	case cfg.HostPathPrefix != "":
		// filepath.Join keeps the prefix when ParentDir is absolute:
		// Join("/host", "/var/x") is "/host/var/x", unlike Python's
		// os.path.join which would discard "/host".
		containerParent = filepath.Join(cfg.HostPathPrefix, cfg.ParentDir)
	default:
		// ParentDir is already the path visible in the container.
		containerParent = cfg.ParentDir
	}
	return hostParent, hostVolume, containerParent
}

// hostMountNSPath returns the path to the host mount namespace for nsenter,
// or "" when HostPathPrefix is unset.
func hostMountNSPath(cfg QuotaScriptConfig) string {
	if cfg.HostPathPrefix == "" {
		return ""
	}
	return filepath.Join(cfg.HostPathPrefix, "proc", "1", "ns", "mnt")
}

// shellQuote wraps value in POSIX single quotes, escaping any embedded
// single quotes with the standard break-and-rejoin technique ('…'"'"'…').
// The result is safe to interpolate into a shell script without risking
// command injection.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

// quotaLockPrologue returns a flock snippet on
// $CONTAINER_PARENT_PATH/.openebs-quota.lock, or "" when disabled.
// fd 9 is util-linux's conventional lock fd; the kernel releases the lock
// when the shell exits. The lockfile path must be container-visible (not
// an unmapped host path), so it uses CONTAINER_PARENT_PATH.
func quotaLockPrologue(enabled bool) string {
	if !enabled {
		return ""
	}
	return `# Serialize quota operations on this host.
LOCKFILE="$CONTAINER_PARENT_PATH/.openebs-quota.lock"
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

// hostExecPrologue defines host_exec(), which runs commands in the host
// mount namespace when HOST_MOUNT_NS is set. XFS/EXT4 project quota is
// resolved against the mount table of the calling process; bind-mounts
// under /host or /data in the pod namespace are not equivalent.
func hostExecPrologue(hostMountNS string) string {
	if hostMountNS == "" {
		return `# No host mount namespace configured; run tools in the current namespace.
host_exec() { "$@"; }
`
	}
	return fmt.Sprintf(`HOST_MOUNT_NS=%s
if [ ! -e "$HOST_MOUNT_NS" ]; then
    echo "host mount namespace not found at $HOST_MOUNT_NS; cannot run filesystem quota tools against the host mount" >&2
    exit 1
fi
if ! command -v nsenter >/dev/null 2>&1; then
    echo "nsenter(1) not found in PATH; required to enter host mount namespace for quota operations" >&2
    exit 1
fi
# Run a command in the host mount namespace so project-quota tools resolve
# the real host XFS/EXT4 mount (pod bind-mount paths are not sufficient).
host_exec() { nsenter --mount="$HOST_MOUNT_NS" -- "$@"; }
`, shellQuote(hostMountNS))
}

// GenerateQuotaApplyScript generates a shell script to apply filesystem quota.
// It detects the filesystem type (XFS or EXT4) and applies a project quota
// with the configured limits.
func GenerateQuotaApplyScript(cfg QuotaScriptConfig) string {
	hostParent, hostVolume, containerParent := quotaPaths(cfg)
	// EXT4 setquota expects plain numbers (KB blocks), not the 'k' suffix.
	extSoft := strings.TrimSuffix(cfg.SoftLimitGrace, "k")
	extHard := strings.TrimSuffix(cfg.HardLimitGrace, "k")

	header := fmt.Sprintf(`set -e
PARENT_PATH=%s
VOLUME_PATH=%s
CONTAINER_PARENT_PATH=%s
XFS_SOFT=%s
XFS_HARD=%s
EXT_SOFT=%s
EXT_HARD=%s
`,
		shellQuote(hostParent),
		shellQuote(hostVolume),
		shellQuote(containerParent),
		shellQuote(cfg.SoftLimitGrace),
		shellQuote(cfg.HardLimitGrace),
		shellQuote(extSoft),
		shellQuote(extHard),
	)

	return header + hostExecPrologue(hostMountNSPath(cfg)) + quotaLockPrologue(cfg.UseHostLock) + applyScriptBody
}

// GenerateQuotaCleanupScript generates a shell script to remove filesystem
// quota and delete the volume directory.
func GenerateQuotaCleanupScript(cfg QuotaScriptConfig) string {
	hostParent, hostVolume, containerParent := quotaPaths(cfg)

	header := fmt.Sprintf(`set -e
PARENT_PATH=%s
VOLUME_PATH=%s
CONTAINER_PARENT_PATH=%s
`,
		shellQuote(hostParent),
		shellQuote(hostVolume),
		shellQuote(containerParent),
	)

	return header + hostExecPrologue(hostMountNSPath(cfg)) + quotaLockPrologue(cfg.UseHostLock) + cleanupScriptBody
}

// applyScriptBody is the static body of the quota-apply script. All inputs
// arrive via shell variables set by the generated header. Filesystem tools
// run through host_exec so they see the host mount table.
const applyScriptBody = `
FS=$(host_exec stat -f -c %T "$VOLUME_PATH")

if [[ "$FS" == "xfs" ]]; then
    # Allocate the next project ID and bind it to the volume.
    PID=$(host_exec xfs_quota -x -c 'report -h' "$PARENT_PATH" 2>/dev/null | tail -2 | awk 'NR==1{print substr ($1,2)}+0' || echo "0")
    PID=$((PID + 1))
    host_exec xfs_quota -x -c "project -s -p $VOLUME_PATH $PID" "$PARENT_PATH"
    host_exec xfs_quota -x -c "limit -p bsoft=$XFS_SOFT bhard=$XFS_HARD $PID" "$PARENT_PATH"
elif [[ "$FS" == "ext2/ext3" ]]; then
    PID=$(host_exec repquota -P "$PARENT_PATH" 2>/dev/null | tail -3 | awk 'NR==1{print substr ($1,2)}+0' || echo "0")
    PID=$((PID + 1))
    host_exec chattr +P -p $PID "$VOLUME_PATH"
    # setquota -P expects plain numbers (KB blocks).
    host_exec setquota -P $PID $EXT_SOFT $EXT_HARD 0 0 "$PARENT_PATH"
else
    echo "Unsupported filesystem type: $FS"
    host_exec rm -rf "$VOLUME_PATH"
    exit 1
fi`

// cleanupScriptBody is the static body of the quota-cleanup script. All
// inputs arrive via shell variables set by the generated header. Filesystem
// tools run through host_exec so they see the host mount table.
const cleanupScriptBody = `
FS=$(host_exec stat -f -c %T "$VOLUME_PATH" 2>/dev/null || echo "unknown")

if [[ "$FS" == "xfs" ]]; then
    ID=$(host_exec xfs_io -c stat "$VOLUME_PATH" 2>/dev/null | awk '/projid/{print $3}' | head -1)
    echo "projid=$ID"
    if [ -n "$ID" ] && [ "$ID" != "0" ]; then
        # -D (not -R): skip special files such as pipes; xfs_io hangs on them.
        host_exec xfs_io -c "chproj -D 0" "$VOLUME_PATH" 2>/dev/null || true
        host_exec xfs_quota -x -c "limit -p bsoft=0 bhard=0 $ID" "$PARENT_PATH" 2>/dev/null || true
    fi
elif [[ "$FS" == "ext2/ext3" ]]; then
    ID=$(host_exec lsattr -pd "$VOLUME_PATH"/ 2>/dev/null | awk '{print $1}')
    if [ -n "$ID" ] && [ "$ID" != "0" ]; then
        host_exec setquota -P $ID 0 0 0 0 "$PARENT_PATH" 2>/dev/null || true
    fi
fi

host_exec rm -rf "$VOLUME_PATH"`
