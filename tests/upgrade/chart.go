package upgrade

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	// defaultChartRepoName is the local alias the released charts are added under.
	defaultChartRepoName = "openebs-localpv"
	// defaultChartRepoURL is where the released localpv-provisioner charts are published.
	defaultChartRepoURL = "https://openebs.github.io/dynamic-localpv-provisioner"
	// defaultChartName is the name of the chart in that repository.
	defaultChartName = "localpv-provisioner"

	// chartRepoNameEnv, chartRepoURLEnv and chartNameEnv override where the
	// chart to upgrade from is pulled from.
	chartRepoNameEnv = "UPGRADE_CHART_REPO_NAME"
	chartRepoURLEnv  = "UPGRADE_CHART_REPO_URL"
	chartNameEnv     = "UPGRADE_CHART_NAME"
	// upgradeFromVersionEnv pins the released version to upgrade from, instead
	// of resolving it from the chart under test.
	upgradeFromVersionEnv = "UPGRADE_FROM_VERSION"
	// chartDirEnv overrides the chart under test.
	chartDirEnv = "UPGRADE_CHART_DIR"
	// releaseNameEnv overrides the helm release the tests install and upgrade.
	releaseNameEnv = "UPGRADE_RELEASE_NAME"
	// skipCleanupEnv leaves the helm release and the test namespace behind.
	skipCleanupEnv = "UPGRADE_SKIP_CLEANUP"
	// uninstallExistingEnv opts in to removing a release which is already using
	// the name under test.
	uninstallExistingEnv = "UPGRADE_UNINSTALL_EXISTING"
)

// chartMetadata is the subset of Chart.yaml the upgrade test cares about.
type chartMetadata struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	AppVersion string `yaml:"appVersion"`
}

// chartValues is the subset of values.yaml the upgrade test cares about.
type chartValues struct {
	LocalPV struct {
		Image struct {
			Repository string `yaml:"repository"`
			Tag        string `yaml:"tag"`
		} `yaml:"image"`
	} `yaml:"localpv"`
}

// defaultChartDir points at the chart in the working tree. The tests run from
// their own package directory, which is two levels below the repo root.
func defaultChartDir() string {
	dir, err := filepath.Abs(filepath.Join("..", "..", "deploy", "helm", "charts"))
	if err != nil {
		return filepath.Join("..", "..", "deploy", "helm", "charts")
	}
	return dir
}

// envOrDefault returns the value of the environment variable, if it is set to
// something other than an empty string.
func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

// envBool returns the boolean value of the environment variable, falling back
// when it is unset or is not something strconv understands as a boolean.
func envBool(key string, fallback bool) bool {
	value := envOrDefault(key, "")
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

// readChartMetadata reads the Chart.yaml of the chart under test.
func readChartMetadata(dir string) (chartMetadata, error) {
	metadata := chartMetadata{}

	path := filepath.Join(dir, "Chart.yaml")
	contents, err := os.ReadFile(path)
	if err != nil {
		return metadata, errors.Wrapf(err, "while reading {%s}", path)
	}

	if err = yaml.Unmarshal(contents, &metadata); err != nil {
		return metadata, errors.Wrapf(err, "while parsing {%s}", path)
	}

	if metadata.Version == "" {
		return metadata, errors.Errorf("no chart version found in {%s}", path)
	}

	return metadata, nil
}

// readChartValues reads the values.yaml of the chart under test.
func readChartValues(dir string) (chartValues, error) {
	values := chartValues{}

	path := filepath.Join(dir, "values.yaml")
	contents, err := os.ReadFile(path)
	if err != nil {
		return values, errors.Wrapf(err, "while reading {%s}", path)
	}

	if err = yaml.Unmarshal(contents, &values); err != nil {
		return values, errors.Wrapf(err, "while parsing {%s}", path)
	}

	return values, nil
}

// provisionerImageTag is the image tag the chart under test deploys the
// provisioner with. The deployment template renders .Values.localpv.image.tag
// as it finds it, with no fallback of its own, so neither is there one here.
func (c chartValues) provisionerImageTag() string {
	return strings.TrimSpace(c.LocalPV.Image.Tag)
}

// addReleasedChartRepo registers the repository the released charts are pulled
// from. It is needed both to resolve the version to upgrade from and to
// install it, so it runs even when that version is pinned.
func addReleasedChartRepo(h *helmClient) error {
	repoName := envOrDefault(chartRepoNameEnv, defaultChartRepoName)
	repoURL := envOrDefault(chartRepoURLEnv, defaultChartRepoURL)

	if err := h.repoAdd(repoName, repoURL); err != nil {
		return err
	}

	return h.repoUpdate(repoName)
}

// resolveUpgradeFromVersion returns the released chart version an upgrade to
// the chart under test starts from.
//
// The chart version carried by a branch always has a pre-release suffix naming
// the version that branch is working towards:
//
//	develop      -> X.Y.0-develop     (X.Y being the next unreleased minor)
//	release/X.Y  -> X.Y.Z-prerelease  (Z being the next unreleased patch)
//
// Dropping the suffix therefore gives the version the branch will eventually
// publish, and the newest release below it is the one users would be upgrading
// from: the last minor release on develop, and the last patch release of X.Y
// on release/X.Y.
func resolveUpgradeFromVersion(h *helmClient, localVersion string) (string, error) {
	if pinned := envOrDefault(upgradeFromVersionEnv, ""); pinned != "" {
		return strings.TrimPrefix(pinned, "v"), nil
	}

	version, err := semver.NewVersion(localVersion)
	if err != nil {
		return "", errors.Wrapf(err, "while parsing the chart version {%s}", localVersion)
	}
	released := fmt.Sprintf("%d.%d.%d", version.Major(), version.Minor(), version.Patch())

	return h.searchLatestBelow(releasedChartRef(), released)
}

// releasedChartRef is the chart reference the upgrade starts from.
func releasedChartRef() string {
	return envOrDefault(chartRepoNameEnv, defaultChartRepoName) + "/" +
		envOrDefault(chartNameEnv, defaultChartName)
}
