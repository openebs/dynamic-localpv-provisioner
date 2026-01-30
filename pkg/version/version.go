package version

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
)

var (
	// GitCommit that was compiled; filled in by
	// the compiler.
	GitCommit string

	// Version is the version of this repo; filled
	// in by the compiler
	Version string
)

const (
	versionFile string = "/src/github.com/openebs/zfs-localpv/VERSION"
)

// Get returns current version from global
// Version variable. If Version is unset then
// from VERSION file at the root of this repo.
func Get() string {
	if Version != "" {
		return Version
	}

	path := filepath.Join(os.Getenv("GOPATH") + versionFile)
	vBytes, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		klog.Errorf("failed to get version: %s", err.Error())
		return ""
	}

	return strings.TrimSpace(string(vBytes))
}

// GetGitCommit returns Git commit SHA-1 from
// global GitCommit variable. If GitCommit is
// unset this calls Git directly.
func GetGitCommit() string {
	if GitCommit != "" {
		return GitCommit
	}

	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		klog.Errorf("failed to get git commit: %s", err.Error())
		return ""
	}

	return strings.TrimSpace(string(output))
}

// GetVersionDetails return version info from git commit
func GetVersionDetails() string {
	return "zfs-" + strings.Join([]string{Get(), GetGitCommit()[0:7]}, "-")
}
