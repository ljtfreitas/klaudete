package patches

import (
	"strings"

	api "github.com/nubank/klaudete/api/v1alpha1"
)

func ApplyTo(patch api.ResourcePatch, properties map[string]any) map[string]any {
	propertyPath, propertyValue := apply(patch.To, patch.From, properties)
	properties[propertyPath] = propertyValue
	return properties
}

func apply(propertyPath string, value any, source map[string]any) (string, any) {
	name, path, found := strings.Cut(propertyPath, ".")

	if !found {
		return propertyPath, value
	}

	if property, found := source[name]; found {
		if property, ok := property.(map[string]any); ok {
			next, value := apply(path, value, property)
			property[next] = value
			return name, property
		}
	}

	next, value := apply(path, value, source)

	return name, map[string]any{
		next: value,
	}
}
