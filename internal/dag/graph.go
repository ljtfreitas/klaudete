package dag

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/gobuffalo/flect"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	resourcesRe = regexp.MustCompile(`resources\.([^.]+)`)
)

type Graph[T any] struct {
	all map[string]*Element[T]
}

func (r Graph[T]) Get(name string) (*Element[T], error) {
	matches := resourcesRe.FindStringSubmatch(name)

	if len(matches) != 0 {
		name = matches[1]
	}

	resource, ok := r.all[name]
	if !ok {
		return nil, fmt.Errorf("resource %s is not registered", name)
	}
	return resource, nil
}

type Element[T any] struct {
	Name string
	Ref  *T

	properties   *Properties
	dependencies []string
}

func (r *Element[T]) NameAsKebabCase() string {
	return flect.Dasherize(r.Name)
}

func (r *Element[T]) Evaluate(args *Args) (ExpandedProperties, error) {
	newProperties := make(map[string]any)
	for name, property := range r.properties.all {
		expanded, err := property.Evaluate(args)
		if err != nil {
			return nil, err
		}
		newProperties[name] = expanded
	}

	return ExpandedProperties(newProperties), nil
}

type ExpandedProperties map[string]any

func NewGraph[T any]() *Graph[T] {
	return &Graph[T]{all: make(map[string]*Element[T])}
}

func (r *Graph[T]) NewElement(name string, properties *runtime.RawExtension) (*Element[T], error) {
	if _, ok := r.all[name]; ok {
		return nil, fmt.Errorf("resource '%s' is duplicated; check the spec", name)
	}

	element, err := NewElement[T](name, properties)
	if err != nil {
		return nil, err
	}
	r.all[name] = element

	return element, nil
}

func NewElement[T any](name string, properties *runtime.RawExtension) (*Element[T], error) {
	element := &Element[T]{Name: name}

	if properties != nil {
		rawProperties := make(map[string]any)
		if err := json.Unmarshal(properties.Raw, &rawProperties); err != nil {
			return nil, fmt.Errorf("unable to unmarshall properties: %w", err)
		}

		resourcePropertiesAsExpressions, err := newProperties(rawProperties)
		if err != nil {
			return nil, fmt.Errorf("unable to read resource properties from %s: %w", name, err)
		}
		element.properties = resourcePropertiesAsExpressions
		element.dependencies = resourcePropertiesAsExpressions.dependencies

	}
	return element, nil
}
