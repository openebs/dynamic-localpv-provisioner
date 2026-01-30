package castemplate

import (
	"gopkg.in/yaml.v3"

	"github.com/openebs/dynamic-localpv-provisioner/pkg/apis/openebs.io/v1alpha1"
)

// UnMarshallToConfig un-marshals the provided
// cas template config in a yaml string format
// to a typed list of cas template config
func UnMarshallToConfig(config string) (configs []v1alpha1.Config, err error) {
	err = yaml.Unmarshal([]byte(config), &configs)
	return
}
