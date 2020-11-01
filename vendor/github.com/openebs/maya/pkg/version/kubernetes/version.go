/*
Copyright 2018 The OpenEBS Authors
Copyright 2018 The Kubernetes Authors

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

// NOTE:
// Some pieces of code was borrowed from:
// - k8s.io/apimachinery/pkg/version/helpers.go
package kubernetes

import (
	msg "github.com/openebs/maya/pkg/msg/v1alpha1"
	"github.com/pkg/errors"
	kubeval "k8s.io/apimachinery/pkg/util/validation"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type releaseType int

const (
	// Bigger the release type number, higher priority it is
	versionTypeAlpha releaseType = iota
	versionTypeBeta
	versionTypeGA
)

const (
	invalidVersionValue = "invalid"
)

// expression is the regular expression for kubernetes aware version
// strings
//
// NOTE:
// - ^v implies every kubernetes version starts with character 'v'
// - () evaluates the regular expression & stores it if match is a success
// - x? implies zero or one x prefer one; where x is the regular expression
// - ?: implies the expected regular expression going forward
const expression = "^v([\\d]+)(?:.)([\\d]+)(?:.)([\\d]+)(?:-(alpha|beta|gke|eks))?(?:.([\\d]+))?"

var regex = regexp.MustCompile(expression)

// A kubernetes aware version is looks like following:
// `vmajor.minor.patch.release.releaseNumber`
//
// NOTE:
// Below are a few valid kubernetes versions
// v1.0.0
// v1.11.1+a0ce1bf1
// v1.12.1-alpha
// v1.12.2-alpha.10
// v0.4.5-gke
// v0.4.6-gke.113
// v0.4.5-eks
// v0.4.6-eks.11
type version struct {
	orig          string      // original version received
	major         int         // version major value
	minor         int         // version minor value
	patch         int         // version patch value
	release       releaseType // represents if this is Alpha, Beta or GA release
	releaseNumber int         // extra digits that comes after the release type
	*msg.Msgs
}

func parse(v string) version {
	var err error
	ver := version{orig: v, Msgs: &msg.Msgs{}}
	submatches := regex.FindStringSubmatch(v)
	if len(submatches) != 6 {
		ver.AddError(errors.Errorf("failed to parse version '%s'", v))
		return ver
	}
	if ver.major, err = strconv.Atoi(submatches[1]); err != nil {
		ver.AddError(errors.Errorf("invalid major version found in '%s'", v))
		return ver
	}
	if ver.minor, err = strconv.Atoi(submatches[2]); err != nil {
		ver.AddError(errors.Errorf("invalid minor version found in '%s'", v))
		return ver
	}
	if ver.patch, err = strconv.Atoi(submatches[3]); err != nil {
		ver.AddError(errors.Errorf("invalid patch version found in '%s'", v))
		return ver
	}
	switch submatches[4] {
	case "alpha":
		ver.release = versionTypeAlpha
	case "beta":
		ver.release = versionTypeBeta
	default:
		// "eks", "gke", "dev", "", or any word other than "alpha" or "beta" are
		// categorized as "ga" i.e. general availability
		ver.release = versionTypeGA
	}
	// we ignore error due to release number
	ver.releaseNumber, _ = strconv.Atoi(submatches[5])
	return ver
}

// AsLabelValue sanitizes the provided version by making it eligible to be
// used as a kubernetes resource's label value
func AsLabelValue(v string) string {
	var (
		sanitized string
		errs      []string
	)
	errs = kubeval.IsValidLabelValue(v)
	if len(errs) == 0 {
		return v
	}
	submatches := regex.FindStringSubmatch(v)
	if len(submatches) != 6 {
		sanitized = invalidVersionValue
	} else {
		sanitized = submatches[0]
	}
	errs = kubeval.IsValidLabelValue(sanitized)
	if len(errs) == 0 {
		return sanitized
	}
	return invalidVersionValue
}

// Compare compares two kubernetes aware version strings
func Compare(v1, v2 string) int {
	if v1 == v2 {
		return 0
	}
	ver1 := parse(v1)
	ver2 := parse(v2)
	switch {
	case ver1.HasError() && ver2.HasError():
		return strings.Compare(v2, v1)
	case ver1.HasError() && !ver2.HasError():
		return -1
	case !ver1.HasError() && ver2.HasError():
		return 1
	}
	if ver1.major > ver2.major {
		return 1
	} else if ver1.major < ver2.major {
		return -1
	} else if ver1.minor > ver2.minor {
		return 1
	} else if ver1.minor < ver2.minor {
		return -1
	} else if ver1.patch > ver2.patch {
		return 1
	} else if ver1.patch < ver2.patch {
		return -1
	} else if int(ver1.release) > int(ver2.release) {
		return 1
	} else if int(ver1.release) < int(ver2.release) {
		return -1
	} else if ver1.releaseNumber > ver2.releaseNumber {
		return 1
	} else if ver1.releaseNumber < ver2.releaseNumber {
		return -1
	}
	return 0
}

// Equals returns true if both the provided kubernetes aware versions
// are equal
func Equals(v1, v2 string) bool {
	return Compare(v1, v2) == 0
}

// GreaterThan returns true if kubernetes aware version v1 is
// greater than kubernetes aware version v2
func GreaterThan(v1, v2 string) bool {
	return Compare(v1, v2) > 0
}

// GreaterThanOrEquals returns true if kubernetes aware version v1
// is either greater than or equal to kubernetes aware version v2
func GreaterThanOrEquals(v1, v2 string) bool {
	return GreaterThan(v1, v2) || Equals(v1, v2)
}

// LessThan returns true if kubernetes aware version v1 is less than
// kubernetes aware version v2
func LessThan(v1, v2 string) bool {
	return Compare(v1, v2) < 0
}

// LessThanOrEquals returns true if kubernetes aware version v1 is
// either less than or equal to kubernetes aware version v2
func LessThanOrEquals(v1, v2 string) bool {
	return LessThan(v1, v2) || Equals(v1, v2)
}

// TemplateFunctions exposes a few functions as go template functions
func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"kubeVersionCompare": Compare,
		"kubeVersionEq":      Equals,
		"kubeVersionGt":      GreaterThan,
		"kubeVersionGte":     GreaterThanOrEquals,
		"kubeVersionLt":      LessThan,
		"kubeVersionLte":     LessThanOrEquals,
	}
}
