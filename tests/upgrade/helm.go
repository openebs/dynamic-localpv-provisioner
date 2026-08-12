package upgrade

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

const (
	// helmMinVersion is the oldest helm release which understands both
	// 'helm get metadata' and 'helm upgrade --reset-then-reuse-values'.
	helmMinVersion = "3.14.0"

	// helmTimeout bounds the --wait of an install/upgrade/uninstall.
	helmTimeout = "5m"
)

// helmClient shells out to the helm binary, against a fixed release and
// namespace.
//
// The helm Go SDK is deliberately not used: it would pull the better part of
// the Kubernetes ecosystem into go.mod for the handful of commands the upgrade
// test needs, and the CLI is what the chart's users (and ci/ci-test.sh) run
// anyway.
type helmClient struct {
	bin        string
	namespace  string
	release    string
	kubeConfig string
	// home is a throwaway HELM_*_HOME root, so that the test neither reads nor
	// writes the chart repository list of whoever is running it.
	home string
}

// helmMetadata is the subset of 'helm get metadata -o json' the test asserts on.
type helmMetadata struct {
	Name       string `json:"name"`
	Chart      string `json:"chart"`
	Version    string `json:"version"`
	AppVersion string `json:"appVersion"`
	Namespace  string `json:"namespace"`
	Revision   int    `json:"revision"`
	Status     string `json:"status"`
}

// helmSearchResult is one entry of 'helm search repo -o json'.
type helmSearchResult struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	AppVersion string `json:"app_version"`
}

func newHelmClient(namespace, release, kubeConfig string) (*helmClient, error) {
	bin, err := exec.LookPath("helm")
	if err != nil {
		return nil, errors.Wrap(err, "while looking up the helm binary in PATH")
	}

	home, err := os.MkdirTemp("", "localpv-upgrade-helm-*")
	if err != nil {
		return nil, errors.Wrap(err, "while creating a temporary helm home")
	}

	h := &helmClient{
		bin:        bin,
		namespace:  namespace,
		release:    release,
		kubeConfig: kubeConfig,
		home:       home,
	}

	if err = h.verifyVersion(); err != nil {
		_ = h.cleanup()
		return nil, err
	}

	return h, nil
}

// cleanup removes the throwaway helm home.
func (h *helmClient) cleanup() error {
	if h == nil || h.home == "" {
		return nil
	}
	return os.RemoveAll(h.home)
}

// verifyVersion fails early, and with an actionable message, on a helm which
// is too old for the flags this test relies on.
func (h *helmClient) verifyVersion() error {
	out, err := h.run("version", "--template", "{{.Version}}")
	if err != nil {
		return err
	}

	found, err := semver.NewVersion(strings.TrimSpace(out))
	if err != nil {
		return errors.Wrapf(err, "while parsing the helm version {%s}", out)
	}

	minimum := semver.MustParse(helmMinVersion)
	if found.LessThan(minimum) {
		return errors.Errorf(
			"helm v%s or above is required by the upgrade tests, found v%s",
			minimum, found,
		)
	}

	return nil
}

// run executes helm with the given arguments and returns its stdout.
func (h *helmClient) run(args ...string) (string, error) {
	if h.kubeConfig != "" {
		args = append(args, "--kubeconfig", h.kubeConfig)
	}

	cmd := exec.Command(h.bin, args...)
	cmd.Env = append(os.Environ(),
		"HELM_CACHE_HOME="+filepath.Join(h.home, "cache"),
		"HELM_CONFIG_HOME="+filepath.Join(h.home, "config"),
		"HELM_DATA_HOME="+filepath.Join(h.home, "data"),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stdout.String(), errors.Wrapf(
			err,
			"while running {helm %s}, stderr {%s}",
			strings.Join(args, " "),
			strings.TrimSpace(stderr.String()),
		)
	}

	return stdout.String(), nil
}

// repoAdd registers the chart repository the released charts are pulled from.
func (h *helmClient) repoAdd(name, url string) error {
	_, err := h.run("repo", "add", name, url, "--force-update")
	return err
}

// repoUpdate refreshes the index of the given chart repository.
func (h *helmClient) repoUpdate(name string) error {
	_, err := h.run("repo", "update", name)
	return err
}

// searchLatestBelow returns the highest version of chartRef below the given
// version. Pre-releases are excluded: a helm constraint without a pre-release
// of its own never matches one, which is exactly what is wanted here as the
// repository also hosts the '-develop' and '-prerelease' charts.
func (h *helmClient) searchLatestBelow(chartRef, version string) (string, error) {
	out, err := h.run("search", "repo", chartRef, "--version", "<"+version, "-o", "json")
	if err != nil {
		return "", err
	}

	var results []helmSearchResult
	if err = json.Unmarshal([]byte(out), &results); err != nil {
		return "", errors.Wrapf(err, "while parsing the helm search output {%s}", out)
	}

	// 'helm search repo' matches on substrings, so pick the chart which was
	// actually asked for rather than whatever sorted first.
	for _, result := range results {
		if result.Name == chartRef {
			return result.Version, nil
		}
	}

	return "", errors.Errorf("no released {%s} chart found below v%s", chartRef, version)
}

// install installs chartRef as the release under test and waits for it to be
// ready. An empty version installs whatever the chart reference resolves to.
func (h *helmClient) install(chartRef, version string, args ...string) error {
	cmd := []string{
		"install", h.release, chartRef,
		"--namespace", h.namespace,
		"--create-namespace",
		"--wait",
		"--timeout", helmTimeout,
	}
	if version != "" {
		cmd = append(cmd, "--version", version)
	}

	_, err := h.run(append(cmd, args...)...)
	return err
}

// upgrade upgrades the release under test to chartRef and waits for the
// rollout to settle.
func (h *helmClient) upgrade(chartRef string, args ...string) error {
	cmd := []string{
		"upgrade", h.release, chartRef,
		"--namespace", h.namespace,
		"--wait",
		"--timeout", helmTimeout,
	}

	_, err := h.run(append(cmd, args...)...)
	return err
}

// installed reports whether a release with the name under test already exists
// in the namespace. It looks at every state, not just the deployed ones,
// because any of them makes the name unavailable to a fresh install.
func (h *helmClient) installed() (bool, error) {
	out, err := h.run("list", "--namespace", h.namespace, "--all", "-o", "json")
	if err != nil {
		return false, err
	}

	var releases []struct {
		Name string `json:"name"`
	}
	if err = json.Unmarshal([]byte(out), &releases); err != nil {
		return false, errors.Wrapf(err, "while parsing the helm list output {%s}", out)
	}

	for _, release := range releases {
		if release.Name == h.release {
			return true, nil
		}
	}

	return false, nil
}

// uninstall removes the release under test, if it is installed.
func (h *helmClient) uninstall() error {
	_, err := h.run(
		"uninstall", h.release,
		"--namespace", h.namespace,
		"--ignore-not-found",
		"--wait",
		"--timeout", helmTimeout,
	)
	return err
}

// metadata returns the release metadata, which is where the installed chart
// version is read back from.
func (h *helmClient) metadata() (helmMetadata, error) {
	metadata := helmMetadata{}

	out, err := h.run("get", "metadata", h.release, "--namespace", h.namespace, "-o", "json")
	if err != nil {
		return metadata, err
	}

	if err = json.Unmarshal([]byte(out), &metadata); err != nil {
		return metadata, errors.Wrapf(err, "while parsing the helm metadata {%s}", out)
	}

	return metadata, nil
}
