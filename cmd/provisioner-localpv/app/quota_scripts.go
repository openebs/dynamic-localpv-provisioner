package app

import (
	_ "embed"
	"path/filepath"
	"strings"
)

// quotaScript is the body of quota.sh, embedded at compile time. The
// script handles both `apply` and `cleanup` modes and parses its own
// flags; see quota.sh for the full interface. We pass the script body to
// `sh -c` and let callers invoke modes uniformly through argv built by
// [QuotaScriptConfig.ApplyArgs] and [QuotaScriptConfig.CleanupArgs].
//
//go:embed quota.sh
var quotaScript string

// QuotaScriptConfig holds configuration for invoking the quota helper script.
type QuotaScriptConfig struct {
	// ParentDir is the base directory whose filesystem owns the quota tables.
	// For HelperPod: use "/data" (the mount point inside the container).
	// For NodeDeployment: use the actual host path (e.g., "/var/openebs/local").
	ParentDir string
	// VolumeDir is the volume subdirectory name (e.g., "pvc-xxx").
	VolumeDir string
	// SoftLimitGrace is the soft quota limit with 'k' suffix (e.g., "1024k").
	// Stripped to plain KB integer before being passed to the script.
	SoftLimitGrace string
	// HardLimitGrace is the hard quota limit with 'k' suffix (e.g., "1024k").
	HardLimitGrace string
	// HostPathPrefix is prepended to paths to access them from within the
	// container. For HelperPod: "" (ParentDir is already mounted at /data).
	// For NodeDeployment: "/host" (host root is mounted at /host).
	HostPathPrefix string
}

// resolvePaths returns the parent and volume paths with HostPathPrefix
// applied (used by NodeDeployment mode, which mounts the host at /host).
func (cfg QuotaScriptConfig) resolvePaths() (parentPath, volumePath string) {
	parentPath = cfg.ParentDir
	volumePath = filepath.Join(cfg.ParentDir, cfg.VolumeDir)
	if cfg.HostPathPrefix != "" {
		parentPath = filepath.Join(cfg.HostPathPrefix, parentPath)
		volumePath = filepath.Join(cfg.HostPathPrefix, volumePath)
	}
	return parentPath, volumePath
}

// ApplyArgs returns the argv to pass to exec for applying a quota. The
// returned slice begins with "sh" and includes the embedded script body,
// so the caller invokes it as e.g. `cmd.Args = cfg.ApplyArgs()`.
func (cfg QuotaScriptConfig) ApplyArgs() []string {
	parentPath, volumePath := cfg.resolvePaths()
	return []string{
		"sh", "-c", quotaScript, "quota.sh", "apply",
		"--parent", parentPath,
		"--volume", volumePath,
		"--soft-kb", strings.TrimSuffix(cfg.SoftLimitGrace, "k"),
		"--hard-kb", strings.TrimSuffix(cfg.HardLimitGrace, "k"),
	}
}

// CleanupArgs returns the argv to pass to exec for cleaning up a quota and
// removing the volume directory.
func (cfg QuotaScriptConfig) CleanupArgs() []string {
	parentPath, volumePath := cfg.resolvePaths()
	return []string{
		"sh", "-c", quotaScript, "quota.sh", "cleanup",
		"--parent", parentPath,
		"--volume", volumePath,
	}
}
