package stringer

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Yaml returns the provided object
// as a yaml formatted string
func Yaml(ctx string, obj interface{}) string {
	if obj == nil {
		return fmt.Sprintf("\n%s {nil}", ctx)
	}

	str, ok := obj.(string)
	if ok {
		return fmt.Sprintf("\n%s {%s}", ctx, str)
	}

	b, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Sprintf("\n%s {nil}", ctx)
	}

	return fmt.Sprintf("\n%s {%s}", ctx, string(b))
}

// JSONIndent returns the provided object
// as a json indent string
func JSONIndent(ctx string, obj interface{}) string {
	if obj == nil {
		return fmt.Sprintf("\n%s {nil}", ctx)
	}

	str, ok := obj.(string)
	if ok {
		return fmt.Sprintf("\n%s {%s}", ctx, str)
	}

	b, err := json.MarshalIndent(obj, "", ".")
	if err != nil {
		return fmt.Sprintf("\n%s {nil}", ctx)
	}

	return fmt.Sprintf("\n%s %s", ctx, string(b))
}
