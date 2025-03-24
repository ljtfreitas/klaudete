package generators

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

type GeneratorName string

var (
	knownGenerators = map[GeneratorName]Generator{
		ListGeneratorName: newListGenerator(),
		DataGeneratorName: newDataGenerator(),
	}
)

type GeneratorList struct {
	Name      string
	Variables Variables
}

func NewGeneratorList(ctx context.Context, spec map[string]*runtime.RawExtension) (*GeneratorList, error) {
	if len(spec) > 1 {
		return nil, fmt.Errorf("invalid spec; just one generator is allowed.")
	}
	for candidate, raw := range spec {
		generator, found := knownGenerators[GeneratorName(candidate)]
		if !found {
			return nil, fmt.Errorf("unsupported generator: %s", candidate)
		}
		name, variables, err := generator.Resolve(ctx, GeneratorSpec(raw))
		if err != nil {
			return nil, err
		}
		generatorList := &GeneratorList{
			Name:      name,
			Variables: variables,
		}
		return generatorList, nil
	}
	return nil, nil
}

type Generator interface {
	Resolve(ctx context.Context, spec GeneratorSpec) (string, Variables, error)
}

type GeneratorSpec *runtime.RawExtension

type Variables []Variable

type Variable any

func unmarshallSpec[T any](spec GeneratorSpec, target *T) (*T, error) {
	err := json.Unmarshal(spec.Raw, target)
	if err != nil {
		return nil, err
	}
	return target, nil
}
