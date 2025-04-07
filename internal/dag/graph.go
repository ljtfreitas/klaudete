package dag

import (
	"fmt"
	"maps"
	"regexp"

	"github.com/dominikbraun/graph"
	"github.com/gobuffalo/flect"
	"github.com/nubank/klaudete/internal/exprs/expr"
)

var (
	resourcesRe = regexp.MustCompile(`resources\.([^.]+)`)
)

type Graph[T any] struct {
	all map[string]*Element[T]
}

func (r Graph[T]) Get(name string) (*Element[T], error) {
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

func (properties ExpandedProperties) With(name string, value any) ExpandedProperties {
	properties[name] = value
	return properties
}

func NewGraph[T any]() *Graph[T] {
	return &Graph[T]{all: make(map[string]*Element[T])}
}

func (r *Graph[T]) NewElement(name string, properties map[string]any, opts ...expr.ExprOption) (*Element[T], error) {
	if _, ok := r.all[name]; ok {
		return nil, fmt.Errorf("resource '%s' is duplicated; check the spec", name)
	}

	element, err := NewElement[T](name, properties, opts...)
	if err != nil {
		return nil, err
	}
	r.all[name] = element

	return element, nil
}

func (r *Graph[T]) Sort() ([]string, error) {
	dag := graph.New(graph.StringHash, graph.Directed(), graph.PreventCycles())

	vertexNameFn := func(name string) string {
		return name
	}

	for name := range maps.Keys(r.all) {
		err := dag.AddVertex(vertexNameFn(name))
		if err != nil {
			return nil, fmt.Errorf("addVertex, name %s: %w", name, err)
		}
	}

	for name, resource := range r.all {
		for _, dependency := range resource.dependencies {
			err := dag.AddEdge(dependency, vertexNameFn(name))
			if err != nil {
				return nil, fmt.Errorf("failure to put graph in topologycal order; unable to add a new edge: name is %s, dependency is %s: %w", name, dependency, err)
			}
		}
	}

	return graph.StableTopologicalSort(dag, func(a, b string) bool {
		return a < b
	})
}

func NewElement[T any](name string, properties map[string]any, opts ...expr.ExprOption) (*Element[T], error) {
	element := &Element[T]{Name: name, properties: &Properties{}}

	if properties != nil && len(properties) != 0 {
		resourcePropertiesAsExpressions, err := newProperties(properties, opts...)
		if err != nil {
			return nil, fmt.Errorf("unable to read resource properties from %s: %w", name, err)
		}
		element.properties = resourcePropertiesAsExpressions
		element.dependencies = resourcePropertiesAsExpressions.dependencies

	}
	return element, nil
}
