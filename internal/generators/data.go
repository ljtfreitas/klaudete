package generators

import "context"

const (
	DataGeneratorName = GeneratorName("data")
)

type DataGenerator struct{}

func newDataGenerator() *DataGenerator {
	return &DataGenerator{}
}

func (g *DataGenerator) Resolve(_ context.Context, spec GeneratorSpec) (string, Variables, error) {
	dataGeneratorSpec, err := unmarshallSpec(spec, &DataGeneratorSpec{})
	if err != nil {
		return "", nil, err
	}

	variables := make(Variables, 0, len(dataGeneratorSpec.Values))
	for _, value := range dataGeneratorSpec.Values {
		variables = append(variables, Variable(value))
	}

	return dataGeneratorSpec.Name, variables, nil
}

type DataGeneratorSpec struct {
	Name   string           `json:"name"`
	Values []map[string]any `json:"values"`
}
