package generators

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

type GeneratorType string

var (
	knownGenerators = map[GeneratorType]Generator{
		ListGeneratorType: newListGenerator(),
		DataGeneratorType: newDataGenerator(),
	}
)

type GeneratorList struct {
	Name      string
	Variables Variables
}

func NewGeneratorList(ctx context.Context, generatorType string, spec GeneratorSpec) (*GeneratorList, error) {
	generator, found := knownGenerators[GeneratorType(generatorType)]
	if !found {
		return nil, fmt.Errorf("unsupported generator: %s", generatorType)
	}
	name, variables, err := generator.Resolve(ctx, spec)
	if err != nil {
		return nil, err
	}
	generatorList := &GeneratorList{
		Name:      name,
		Variables: variables,
	}
	return generatorList, nil
}

type Generator interface {
	Resolve(ctx context.Context, spec GeneratorSpec) (string, Variables, error)
}

type GeneratorSpec *runtime.RawExtension

type Variables []Variable

type Variable any

func UnmarshallSpec[T any](spec GeneratorSpec, target *T) (*T, error) {
	err := json.Unmarshal(spec.Raw, target)
	if err != nil {
		return nil, err
	}
	return target, nil
}

func marshallSpec(spec any) (GeneratorSpec, error) {
	specAsBytes, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	return GeneratorSpec(&runtime.RawExtension{Raw: specAsBytes}), nil
}
