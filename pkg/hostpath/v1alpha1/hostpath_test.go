package v1alpha1

import (
	"strings"
	"testing"
)

func TestValidateWithCheckf(t *testing.T) {
	tests := map[string]struct {
		path        string
		check       Predicate
		msg         string
		expectError bool
	}{
		"verify empty path": {
			"",
			IsNonRoot(),
			"missing host path",
			true,
		},
		"verify non root with msg": {
			"/pv",
			IsNonRoot(),
			"root directory should not be used",
			true,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().
				WithPath(mock.path).
				WithCheckf(mock.check, "%s", mock.msg)
			err := b.Validate()
			if (err != nil) != mock.expectError {
				t.Fatalf("test %s failed, expected error: %t but got: %v",
					name, mock.expectError, err)
			}
			if (err != nil) && !strings.Contains(err.Error(), mock.msg) {
				t.Fatalf("test %s failed, expected error to include: %v but got: %v",
					name, mock.msg, err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		path        string
		checks      []Predicate
		expectError bool
	}{
		"predicate returns true": {
			"/var/openebs/pv",
			[]Predicate{IsNonRoot()},
			false,
		},
		"predicate returns false": {
			"/pv",
			[]Predicate{IsNonRoot()},
			true,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().
				WithPath(mock.path).
				WithChecks(mock.checks...)
			err := b.Validate()
			if (err != nil) != mock.expectError {
				t.Fatalf("test %s failed, expected error: %t but got: %v",
					name, mock.expectError, err)
			}
		})
	}
}

func TestValidateAndBuild(t *testing.T) {
	tests := map[string]struct {
		basePath    string
		relPath     string
		expectError bool
		expectPath  string
	}{
		"verify empty path": {
			"",
			"",
			true,
			"",
		},
		"verify empty rel path": {
			"/pv",
			"",
			true,
			"",
		},
		"verify empty base path": {
			"",
			"/pv",
			true,
			"",
		},
		"verify incomplete path": {
			"abc",
			"/pv",
			true,
			"",
		},
		"verify valid path": {
			"/abc",
			"/def",
			false,
			"/abc/def",
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().
				WithPathJoin(mock.basePath, mock.relPath).
				WithCheckf(IsNonRoot(), "root directory is not allowed")
			path, err := b.ValidateAndBuild()
			if (err != nil) != mock.expectError {
				t.Fatalf("test %s failed, expected error: %t but got: %v",
					name, mock.expectError, err)
			}
			if (err == nil) && strings.Compare(path, mock.expectPath) != 0 {
				t.Fatalf("test %s failed, expected error to include: %v but got: %v",
					name, mock.expectPath, path)
			}
		})
	}
}
