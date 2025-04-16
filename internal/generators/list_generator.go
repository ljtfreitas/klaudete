package generators

import "context"

const (
	ListGeneratorType = GeneratorType("list")
)

type ListGenerator struct{}

func newListGenerator() *ListGenerator {
	return &ListGenerator{}
}

func (g *ListGenerator) Resolve(_ context.Context, spec GeneratorSpec) (string, Variables, error) {
	listGeneratorSpec, err := UnmarshallSpec(spec, &ListGeneratorSpec{})
	if err != nil {
		return "", nil, err
	}

	variables := make(Variables, 0, len(listGeneratorSpec.Values))
	for _, value := range listGeneratorSpec.Values {
		variables = append(variables, Variable(value))
	}

	return listGeneratorSpec.Name, variables, nil
}

type ListGeneratorSpec struct {
	Name   string `json:"name"`
	Values []any  `json:"values"`
}

func NewListGeneratorSpec(name string, values ...any) (GeneratorSpec, error) {
	return marshallSpec(ListGeneratorSpec{
		Name:   name,
		Values: values,
	})
}
