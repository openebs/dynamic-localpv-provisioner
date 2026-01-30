// Provides functions based on k8s.io/apimachinery/pkg/apis/meta/v1/unstructured
// They are copied here to make them exported.
//
// TODO
// Check if it makes sense to import the entire unstructured package of
// k8s.io/apimachinery/pkg/apis/meta/v1/unstructured versus. copying
//
// TODO
// Move to maya/pkg/unstructured/v1alpha1 as helpers.go

package utils

// GetNestedField returns a nested field from the provided map
func GetNestedField(obj map[string]interface{}, fields ...string) interface{} {
	var val interface{} = obj
	for _, field := range fields {
		if _, ok := val.(map[string]interface{}); !ok {
			return nil
		}
		val = val.(map[string]interface{})[field]
	}
	return val
}

// MergeMapOfObjects will merge the map from src to dest. It will override
// existing keys of the destination
func MergeMapOfObjects(dest map[string]interface{}, src map[string]interface{}) bool {
	// nil check as storing into a nil map panics
	if dest == nil {
		return false
	}

	for k, v := range src {
		dest[k] = v
	}

	return true
}

// ContainsString returns true if the provided element is present in the
// provided array
func ContainsString(stringarr []string, element string) bool {
	for _, elem := range stringarr {
		if elem == element {
			return true
		}
	}
	return false
}
