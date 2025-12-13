package config

import "github.com/tidwall/gjson"

// ResourceGetParent gets the resource parent name.
func ResourceGetParent(resources, virtuals []byte, resourceName string) string {
	path := resourceName + ".parentName"
	if gjson.GetBytes(resources, path).Exists() {
		return gjson.GetBytes(resources, path).String()
	}

	return gjson.GetBytes(virtuals, path).String()
}

// ResourceInheritsFrom checks if the resource name inherits from the target name.
func ResourceInheritsFrom(resources, virtuals []byte, resourceName, targetName string) bool {
	for parent := ResourceGetParent(resources, virtuals, resourceName); parent != ""; parent = ResourceGetParent(resources, virtuals, parent) {
		if parent == targetName {
			return true
		}
	}

	return false
}

// ResourceInheritsFromMap checks if the resource name inherits from the target name within a parent map.
func ResourceInheritsFromMap(parents map[string]string, resourceName, targetName string) bool {
	for {
		parent, ok := parents[resourceName]
		if !ok || parent == "" {
			return false
		}

		if parent == targetName {
			return true
		}

		resourceName = parent
	}
}
