package generators

const (
	DataGeneratorType = GeneratorType("data")
)

type DataGenerator struct {
	ListGenerator
}

func newDataGenerator() *DataGenerator {
	return &DataGenerator{}
}

type DataGeneratorSpec struct {
	Name   string           `json:"name"`
	Values []map[string]any `json:"values"`
}

func NewDataGeneratorSpec(name string, values ...map[string]any) (GeneratorSpec, error) {
	return marshallSpec(DataGeneratorSpec{
		Name:   name,
		Values: values,
	})
}
