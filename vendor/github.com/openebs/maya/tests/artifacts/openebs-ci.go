// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package artifacts

import (
	"errors"
	"io/ioutil"
	"strings"

	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/openebs/maya/pkg/artifact/v1alpha1"
)

// ArtifactSource holds the path to fetch artifacts
type ArtifactSource string
type Artifact string

const (
	OpenEBSArtifacts  ArtifactSource = "../artifacts/openebs-ci.yaml"
	CStorPVCArtifacts ArtifactSource = "../artifacts/cstor-pvc.yaml"
	JivaPVCArtifacts  ArtifactSource = "../artifacts/jiva-pvc.yaml"
	SingleReplicaSC   ArtifactSource = "../artifacts/storageclass-1r.yaml"
	CVRArtifact       ArtifactSource = "../artifacts/cvr-schema.yaml"
	CRArtifact        ArtifactSource = "../artifacts/cr-schema.yaml"
)

// PodName holds the name of the pod
type PodName string

const (
	// MayaAPIServerPodName is the name of the maya api server pod
	MayaAPIServerPodName PodName = "maya-apiserver"
)

// Namespace holds the name of the namespace
type Namespace string

const (
	// OpenebsNamespace is the name of the openebs namespace
	OpenebsNamespace Namespace = "openebs"
)

// LabelSelector holds the label got openebs components
type LabelSelector string

const (
	MayaAPIServerLabelSelector             LabelSelector = "name=maya-apiserver"
	OpenEBSProvisionerLabelSelector        LabelSelector = "name=openebs-provisioner"
	OpenEBSLocalPVProvisionerLabelSelector LabelSelector = "openebs.io/component-name=openebs-localpv-provisioner"
	OpenEBSSnapshotOperatorLabelSelector   LabelSelector = "name=openebs-snapshot-operator"
	OpenEBSAdmissionServerLabelSelector    LabelSelector = "app=admission-webhook"
	OpenEBSNDMLabelSelector                LabelSelector = "openebs.io/component-name=openebs-ndm"
	OpenEBSNDMOperatorLabelSelector        LabelSelector = "openebs.io/component-name=ndm-operator"
	OpenEBSCStorPoolLabelSelector          LabelSelector = "app=cstor-pool"
)

func parseK8sYaml(yamls string) (k8s.UnstructedList, []error) {
	sepYamlfiles := strings.Split(yamls, "---")
	artifacts := v1alpha1.ArtifactList{}
	for _, f := range sepYamlfiles {
		if f == "\n" || f == "" {
			// ignore empty cases
			continue
		}
		f = strings.TrimSpace(f)
		artifacts.Items = append(artifacts.Items, &v1alpha1.Artifact{Doc: f})
	}
	return artifacts.ToUnstructuredList()
}

// parseK8sYamlFromFile parses the kubernetes yaml and returns the objects in a UnstructuredList
func parseK8sYamlFromFile(filename string) (k8s.UnstructedList, []error) {
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return k8s.UnstructedList{}, []error{err}
	}
	fileAsString := string(fileBytes[:])
	return parseK8sYaml(fileAsString)
}

// GetArtifactsListUnstructuredFromFile returns the unstructured list of openebs components
func GetArtifactsListUnstructuredFromFile(a ArtifactSource) ([]*unstructured.Unstructured, []error) {
	ulist, err := parseK8sYamlFromFile(string(a))
	if err != nil {
		return nil, err
	}
	nList := ulist.MapAllIfAny([]k8s.UnstructuredMiddleware{})
	return nList.Items, err
}

// GetArtifactUnstructuredFromFile returns the unstructured list of openebs components
func GetArtifactUnstructuredFromFile(a ArtifactSource) (*unstructured.Unstructured, error) {
	ulist, err := parseK8sYamlFromFile(string(a))
	if len(err) != 0 {
		return nil, err[0]
	}
	if len(ulist.Items) != 1 {
		return nil, errors.New("more than one artifacts found")
	}
	nList := ulist.MapAllIfAny([]k8s.UnstructuredMiddleware{})
	return nList.Items[0], nil
}

// GetArtifactsListUnstructured returns the unstructured list of openebs components
func GetArtifactsListUnstructured(a Artifact) ([]*unstructured.Unstructured, []error) {
	ulist, err := parseK8sYaml(strings.TrimSpace(string(a)))
	if err != nil {
		return nil, err
	}
	nList := ulist.MapAllIfAny([]k8s.UnstructuredMiddleware{})
	return nList.Items, err
}

// GetArtifactUnstructured returns the unstructured list of openebs components
func GetArtifactUnstructured(a Artifact) (*unstructured.Unstructured, error) {
	ulist, err := parseK8sYaml(string(a))
	if len(err) != 0 {
		return nil, err[0]
	}
	if len(ulist.Items) != 1 {
		return nil, errors.New("more than one artifacts found")
	}
	nList := ulist.MapAllIfAny([]k8s.UnstructuredMiddleware{})
	return nList.Items[0], nil
}
