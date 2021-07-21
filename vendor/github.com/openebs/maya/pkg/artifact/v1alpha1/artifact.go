/*
Copyright 2018 The OpenEBS Authors

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

package v1alpha1

import (
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	"github.com/pkg/errors"
)

// Artifact represents a YAML compatible artifact that will be installed,
// applied, etc
type Artifact struct {
	// Doc represents the YAML compatible artifact
	Doc string
}

// ArtifactList represents a list of artifacts to install
type ArtifactList struct {
	Items []*Artifact
}

// ToUnstructuredList transforms this ArtifactList to corresponding list of
// unstructured instances
func (l ArtifactList) ToUnstructuredList() (ul k8s.UnstructedList, errs []error) {
	return l.UnstructuredListC(k8s.CreateUnstructuredFromYaml)
}

// UnstructuredListC transforms this ArtifactList to corresponding list of
// unstructured instances by making use of unstructured creator instance
func (l ArtifactList) UnstructuredListC(c k8s.UnstructuredCreator) (ul k8s.UnstructedList, errs []error) {
	for _, artifact := range l.Items {
		unstructured, err := c(artifact.Doc)
		if err != nil {
			errs = append(errs, errors.Wrap(err, "failed to transform artifact into unstructured instance"))
			continue
		}
		ul.Items = append(ul.Items, unstructured)
	}
	return
}
