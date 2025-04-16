package generators

import (
	"context"
	"fmt"

	"github.com/nubank/klaudete/internal/inventory"
)

const (
	InventoryGeneratorType = GeneratorType("inventory")
)

type InventoryGenerator struct {
	inventory *inventory.InventoryClient
}

func NewInventoryGenerator(inventory *inventory.InventoryClient) *InventoryGenerator {
	return &InventoryGenerator{
		inventory: inventory,
	}
}

func (g *InventoryGenerator) Resolve(ctx context.Context, spec GeneratorSpec) (string, Variables, error) {
	inventoryGeneratorSpec, err := UnmarshallSpec(spec, &InventoryGeneratorSpec{})
	if err != nil {
		return "", nil, err
	}
	output, err := g.inventory.RunKcl(ctx, inventoryGeneratorSpec.Kcl.Content)
	if err != nil {
		return "", nil, fmt.Errorf("failure to generate values from Inventory using a kcl program: %w", err)
	}

	valueToExport := output[inventoryGeneratorSpec.Kcl.Path]
	variables := make(Variables, 0)

	if arrayToExport, isArray := valueToExport.([]any); isArray {
		for _, element := range arrayToExport {
			variables = append(variables, Variable(element))
		}
	} else {
		variables = append(variables, Variable(valueToExport))
	}

	return inventoryGeneratorSpec.Name, variables, nil
}

type InventoryGeneratorSpec struct {
	Name string  `json:"name"`
	Kcl  KclSpec `json:"kcl"`
}

type KclSpec struct {
	Content string `json:"content"`
	Path    string `json:"path"`
}

func NewInventoryGeneratorSpec(name string, kcl string, path string) (GeneratorSpec, error) {
	return marshallSpec(InventoryGeneratorSpec{
		Name: name,
		Kcl: KclSpec{
			Content: kcl,
			Path:    path,
		},
	})
}
