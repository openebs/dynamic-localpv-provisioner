package v1alpha1

import (
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetHostName returns the hostname corresponding to
// the provided node name
func GetHostName(name string) (string, error) {
	return KubeClientInstanceOrDie().GetHostName(name, metav1.GetOptions{})
}

// GetHostNameOrNodeName returns the hostname corresponding
// to the provided node name or node name itself if hostname
// is not available
func GetHostNameOrNodeName(name string) (string, error) {
	return KubeClientInstanceOrDie().
		GetHostNameOrNodeName(name, metav1.GetOptions{})
}

// TemplateFunctions exposes a few functions as go template functions
func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"kubeNodeGetHostName":           GetHostName,
		"kubeNodeGetHostNameOrNodeName": GetHostNameOrNodeName,
	}
}
