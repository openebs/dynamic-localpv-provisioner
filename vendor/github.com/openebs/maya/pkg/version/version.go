/*
Copyright 2017 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package version

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// GitCommit that was compiled. This will be filled in by the compiler.
	GitCommit string

	// Version show the version number,fill in by the compiler
	Version string

	// VersionMeta is a pre-release marker for the version. If this is "" (empty string)
	// then it means that it is a final release. Otherwise, this is a pre-release
	// such as "dev" (in development), "beta", "rc1", etc.
	VersionMeta string
)

const (
	versionFile   string = "/src/github.com/openebs/maya/VERSION"
	buildMetaFile string = "/src/github.com/openebs/maya/BUILDMETA"

	// versionDelimiter is used as a delimiter to separate version info
	versionDelimiter string = "-"

	// versionChars consist of valid version characters
	versionChars string = ".0123456789"
)

// IsNotVersioned returns true if the given string does not have version as its
// suffix
func IsNotVersioned(given string) bool {
	return !IsVersioned(given)
}

// IsVersioned returns true if the given string has version as its name
// example: cstor-craete-pool-1.11.0-ce-rc2
func IsVersioned(given string) bool {
	a := strings.Split(given, versionDelimiter)
	if len(a) == 0 {
		return false
	}
	ver := ""
	for _, v := range a {
		if containsOnly(v, versionChars) {
			ver = v
			break
		}
	}
	return len(strings.Split(ver, ".")) == 3
}

// containsOnly returns true if provided string consists only of the provided
// set
func containsOnly(s string, set string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(set, r)
	}) == -1
}

// WithSuffix appends current version to the provided string
func WithSuffix(given string) (suffixed string) {
	return given + versionDelimiter + Current()
}

// WithSuffixLower appends current version to the provided string
func WithSuffixLower(given string) (suffixed string) {
	return given + versionDelimiter + strings.ToLower(Current())
}

// WithSuffixIf appends current version to the provided string if given predicate
// succeeds
func WithSuffixIf(given string, p func(string) bool) (suffixed string) {
	if p(given) {
		return WithSuffixLower(given)
	}
	return given
}

// WithSuffixesIf appends current version to the provided strings
func WithSuffixesIf(given []string, p func(string) bool) (suffixed []string) {
	for _, s := range given {
		if p(s) {
			suffixed = append(suffixed, WithSuffixLower(s))
		} else {
			suffixed = append(suffixed, s)
		}
	}
	return
}

// Current returns the current version of maya
func Current() string {
	return GetVersion()
}

// GetVersion returns the current version from the global Version variable.
// If Version is unset then from the VERSION file at the root of the repo.
func GetVersion() string {
	if Version != "" {
		return Version
	}
	path := filepath.Join(os.Getenv("GOPATH") + versionFile)
	vBytes, err := ioutil.ReadFile(path)
	if err != nil {
		// ignore error
		return ""
	}
	return strings.TrimSpace(string(vBytes))
}

// GetBuildMeta returns the build type from the global VersionMeta variable.
// If VersionMeta is unset then from the BUILDMETA file at the root of the repo.
func GetBuildMeta() string {
	if VersionMeta != "" {
		return "-" + VersionMeta
	}
	path := filepath.Join(os.Getenv("GOPATH") + buildMetaFile)
	vBytes, err := ioutil.ReadFile(path)
	if err != nil {
		// ignore error
		return ""
	}
	return "-" + strings.TrimSpace(string(vBytes))
}

// GetGitCommit returns the Git commit SHA-1 from the global GitCommit variable.
// If GitCommit is unset then by calling Git directly.
func GetGitCommit() string {
	if GitCommit != "" {
		return GitCommit
	}
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// ignore error
		return ""
	}
	return strings.TrimSpace(string(output))
}

func GetVersionDetails() string {
	return strings.Join([]string{GetVersion(), GetGitCommit()[0:7]}, "-")
}

// NewVersionCollector returns a collector which exports metrics
// about current version information.
// Note: program name should be similar to maya_exporter (with
// underscore not with dash)
func NewVersionCollector(program string) *prometheus.GaugeVec {
	buildInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Subsystem: program,
			Name:      "version",
			Help:      "A metric with a constant '1' value labeled by commit and version from which maya-exporter was built.",
		},
		[]string{"commit", "version", "metaversion"},
	)
	buildInfo.WithLabelValues(GitCommit, Version, VersionMeta).Set(1)
	return buildInfo
}
