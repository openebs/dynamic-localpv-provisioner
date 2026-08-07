package version

import (
	"strings"
	"testing"
)

func TestGetVersionDetails(t *testing.T) {
	origVersion := Version
	origGitCommit := GitCommit
	defer func() {
		Version = origVersion
		GitCommit = origGitCommit
	}()

	testCases := map[string]struct {
		version     string
		gitCommit   string
		expectValue string
	}{
		"release version and full commit": {
			version:     "4.6.0",
			gitCommit:   "abcdef1234567890",
			expectValue: "hostpath-4.6.0-abcdef1",
		},
		"develop version and full commit": {
			version:     "4.6.0-develop",
			gitCommit:   "0123456789abcdef",
			expectValue: "hostpath-4.6.0-develop-0123456",
		},
	}

	for k, v := range testCases {
		v := v
		t.Run(k, func(t *testing.T) {
			Version = v.version
			GitCommit = v.gitCommit

			actualValue := GetVersionDetails()
			if actualValue != v.expectValue {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
			if strings.HasPrefix(actualValue, "zfs-") {
				t.Errorf("version details must not carry the zfs-localpv prefix, got %s", actualValue)
			}
		})
	}
}
