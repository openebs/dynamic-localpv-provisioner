package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The tests below execute the embedded quota.sh under the exact argv the
// production code constructs (via ApplyArgs / CleanupArgs). To avoid
// requiring a quota-enabled filesystem (or root), every external command
// the script invokes (stat, xfs_quota, xfs_io, repquota, setquota, chattr,
// lsattr) is replaced with a stub that records its invocation to a log
// file and emits configurable synthetic stdout. The tests assert on the
// log to verify the script invoked the right tools with the right
// arguments.

// dispatcherScript is the body of a single shell script that every stub
// symlinks to. It records its invocation as "<name>: <args>" in ${STUB_LOG}
// and emits synthetic stdout for the commands whose output the quota
// script parses. Controlled via env vars:
//
//	STUB_FS      - filesystem type echoed by `stat -f -c %T`
//	STUB_PROJID  - project ID surfaced by `xfs_io -c stat` and `lsattr -pd`
const dispatcherScript = `#!/bin/sh
name=$(basename "$0")
printf '%s: %s\n' "$name" "$*" >> "${STUB_LOG}"
case "$name" in
    stat)   echo "${STUB_FS:-xfs}" ;;
    xfs_io) echo "fsxattr.projid = ${STUB_PROJID:-1}" ;;
    lsattr) echo "${STUB_PROJID:-1} ----P--------e----- /placeholder/" ;;
esac
exit 0
`

// stubbedCommands is the set of external commands quota.sh may invoke.
// Each gets a symlink to the dispatcher so its argv is captured.
var stubbedCommands = []string{
	"stat", "xfs_quota", "xfs_io", "repquota", "setquota",
	"chattr", "lsattr",
}

// setupQuotaStubs creates a directory containing stub executables and
// returns its path. The log file lives at <dir>/calls.log.
func setupQuotaStubs(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	dispatcher := filepath.Join(dir, "_dispatch")
	if err := os.WriteFile(dispatcher, []byte(dispatcherScript), 0o755); err != nil {
		t.Fatalf("write dispatcher: %v", err)
	}
	for _, name := range stubbedCommands {
		if err := os.Symlink(dispatcher, filepath.Join(dir, name)); err != nil {
			t.Fatalf("symlink %s: %v", name, err)
		}
	}
	return dir
}

// runQuotaArgs runs the given argv with stubDir at the front of PATH, then
// returns the contents of calls.log for substring assertions.
func runQuotaArgs(t *testing.T, stubDir string, argv []string, extraEnv map[string]string) string {
	t.Helper()
	logPath := filepath.Join(stubDir, "calls.log")
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Env = append(os.Environ(),
		"PATH="+stubDir+":"+os.Getenv("PATH"),
		"STUB_LOG="+logPath,
	)
	for k, v := range extraEnv {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("script failed: %v\nscript stdout/stderr:\n%s", err, out)
	}
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read calls.log: %v", err)
	}
	return strings.TrimSpace(string(logBytes))
}

// requireSh skips the test if /bin/sh isn't available (e.g., on Windows CI).
func requireSh(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
}

func TestQuotaScriptApply(t *testing.T) {
	requireSh(t)

	cases := map[string]struct {
		fsType    string
		wantCalls []string
	}{
		"xfs": {
			fsType: "xfs",
			wantCalls: []string{
				"xfs_quota: -x -c report -h",
				"xfs_quota: -x -c project -s -p",
				"xfs_quota: -x -c limit -p bsoft=1024k bhard=1024k 1",
			},
		},
		"ext4": {
			fsType: "ext2/ext3",
			wantCalls: []string{
				"repquota: -P",
				"chattr: +P -p 1",
				"setquota: -P 1 1024 1024 0 0",
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			stubDir := setupQuotaStubs(t)
			cfg := QuotaScriptConfig{
				ParentDir:      t.TempDir(),
				VolumeDir:      "pvc-abc",
				SoftLimitGrace: "1024k",
				HardLimitGrace: "1024k",
			}
			calls := runQuotaArgs(t, stubDir, cfg.ApplyArgs(), map[string]string{
				"STUB_FS": tc.fsType,
			})
			for _, want := range tc.wantCalls {
				if !strings.Contains(calls, want) {
					t.Errorf("expected call %q in calls.log:\n%s", want, calls)
				}
			}
		})
	}
}

func TestQuotaScriptCleanup(t *testing.T) {
	requireSh(t)

	cases := map[string]struct {
		fsType    string
		projid    string
		wantCalls []string
	}{
		"xfs": {
			fsType: "xfs",
			projid: "1",
			wantCalls: []string{
				"xfs_io: -c stat",
				"xfs_io: -c chproj -R 0",
				"xfs_quota: -x -c limit -p bsoft=0 bhard=0 1",
			},
		},
		"ext4": {
			fsType: "ext2/ext3",
			projid: "1",
			wantCalls: []string{
				"lsattr: -pd",
				"setquota: -P 1 0 0 0 0",
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			stubDir := setupQuotaStubs(t)
			cfg := QuotaScriptConfig{
				ParentDir: t.TempDir(),
				VolumeDir: "pvc-abc",
			}
			calls := runQuotaArgs(t, stubDir, cfg.CleanupArgs(), map[string]string{
				"STUB_FS":     tc.fsType,
				"STUB_PROJID": tc.projid,
			})
			for _, want := range tc.wantCalls {
				if !strings.Contains(calls, want) {
					t.Errorf("expected call %q in calls.log:\n%s", want, calls)
				}
			}
		})
	}
}

// TestResolvePaths covers the HostPathPrefix handling, which the
// execution-based tests above don't touch because they need writable
// paths under t.TempDir().
func TestResolvePaths(t *testing.T) {
	cases := map[string]struct {
		cfg        QuotaScriptConfig
		wantParent string
		wantVolume string
	}{
		"no prefix (HelperPod)": {
			cfg:        QuotaScriptConfig{ParentDir: "/data", VolumeDir: "pvc-abc"},
			wantParent: "/data",
			wantVolume: "/data/pvc-abc",
		},
		"with /host prefix (NodeDeployment)": {
			cfg: QuotaScriptConfig{
				ParentDir: "/var/openebs/local", VolumeDir: "pvc-abc",
				HostPathPrefix: "/host",
			},
			wantParent: "/host/var/openebs/local",
			wantVolume: "/host/var/openebs/local/pvc-abc",
		},
	}
	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			parent, volume := tc.cfg.resolvePaths()
			if parent != tc.wantParent {
				t.Errorf("parent: got %q, want %q", parent, tc.wantParent)
			}
			if volume != tc.wantVolume {
				t.Errorf("volume: got %q, want %q", volume, tc.wantVolume)
			}
		})
	}
}
