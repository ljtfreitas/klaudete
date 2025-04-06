package dag

import (
	"fmt"

	"github.com/nubank/klaudete/internal/exprs"
	"github.com/nubank/klaudete/internal/exprs/expr"
	"k8s.io/apimachinery/pkg/util/sets"
)

type Properties struct {
	all          map[string]Property
	dependencies []string
}

type Property interface {
	Name() string
	Dependencies() []string
	Evaluate(*Args, ...any) (any, error)
}

func newProperties(source map[string]any, opts ...expr.ExprOption) (*Properties, error) {
	propertiesWithExpressions := make(map[string]Property)
	dependencies := sets.NewString()

	for name, value := range source {
		elementWithExpressions, err := readProperty(name, value, opts...)
		if err != nil {
			return nil, fmt.Errorf("unable to read properties from field %s: %w", name, err)
		}

		propertiesWithExpressions[name] = elementWithExpressions

		dependencies = dependencies.Insert(elementWithExpressions.Dependencies()...)
	}

	properties := &Properties{
		all:          propertiesWithExpressions,
		dependencies: dependencies.List(),
	}

	return properties, nil
}

func readProperty(name string, value any, opts ...expr.ExprOption) (Property, error) {
	switch value := value.(type) {
	case map[string]any:
		return readObjectProperty(name, value, opts...)
	case []any:
		return readArrayProperty(name, value, opts...)
	default:
		e, err := exprs.Parse(value, opts...)
		if err != nil {
			return nil, err
		}
		expressionProperty := &ExpressionProperty{
			name:         name,
			expression:   e,
			dependencies: e.Dependencies(),
		}
		return expressionProperty, nil
	}
}

func readObjectProperty(name string, value map[string]any, opts ...expr.ExprOption) (Property, error) {
	properties := make(map[string]Property)
	dependencies := make([]string, 0)
	for propertyName, element := range value {
		newElement, err := readProperty(fmt.Sprintf("%s.%s", name, propertyName), element, opts...)
		if err != nil {
			return nil, err
		}
		properties[propertyName] = newElement
		dependencies = append(dependencies, newElement.Dependencies()...)
	}
	objectProperty := &ObjectProperty{
		name:         name,
		properties:   properties,
		dependencies: dependencies,
	}
	return objectProperty, nil
}

func readArrayProperty(name string, value []any, opts ...expr.ExprOption) (Property, error) {
	values := make([]Property, 0)
	dependencies := make([]string, 0)
	for i, element := range value {
		newElement, err := readProperty(fmt.Sprintf("%s[%d]", name, i), element, opts...)
		if err != nil {
			return nil, err
		}
		values = append(values, newElement)
		dependencies = append(dependencies, newElement.Dependencies()...)
	}
	arrayProperty := &ArrayProperty{
		name:         name,
		properties:   values,
		dependencies: dependencies,
	}
	return arrayProperty, nil
}

type ExpressionProperty struct {
	name         string
	expression   exprs.Expression
	dependencies []string
}

func (p ExpressionProperty) Name() string {
	return p.name
}

func (p ExpressionProperty) Dependencies() []string {
	return p.dependencies
}

func (p ExpressionProperty) Evaluate(args *Args, options ...any) (any, error) {
	return p.expression.Evaluate(args.all, options...)
}

type ObjectProperty struct {
	name         string
	properties   map[string]Property
	dependencies []string
}

func (p ObjectProperty) Name() string {
	return p.name
}

func (p ObjectProperty) Dependencies() []string {
	return p.dependencies
}

func (p ObjectProperty) Evaluate(args *Args, options ...any) (any, error) {
	newMap := make(map[string]any)
	for name, property := range p.properties {
		newValue, err := property.Evaluate(args)
		if err != nil {
			return nil, err
		}

		newMap[name] = newValue
	}
	return newMap, nil
}

type ArrayProperty struct {
	name         string
	properties   []Property
	dependencies []string
}

func (p ArrayProperty) Name() string {
	return p.name
}

func (p ArrayProperty) Dependencies() []string {
	return p.dependencies
}

func (p ArrayProperty) Evaluate(args *Args, options ...any) (any, error) {
	newArray := make([]any, 0, len(p.properties))
	for _, property := range p.properties {
		newValue, err := property.Evaluate(args)
		if err != nil {
			return nil, err
		}
		newArray = append(newArray, newValue)
	}
	return newArray, nil
}
