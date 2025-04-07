package generators

import (
	"context"
	"fmt"

	"github.com/nubank/klaudete/internal/generators/inventory"
	clientv1alpha1 "github.com/nubank/nu-infra-inventory/sdk/pkg/client"
)

const (
	InventoryGeneratorType = GeneratorType("inventory")
)

type InventoryGenerator struct {
	client clientv1alpha1.Client
}

func NewInventoryGenerator(client clientv1alpha1.Client) *InventoryGenerator {
	return &InventoryGenerator{
		client: client,
	}
}

func (g *InventoryGenerator) Resolve(ctx context.Context, spec GeneratorSpec) (string, Variables, error) {
	inventoryGeneratorSpec, err := UnmarshallSpec(spec, &InventoryGeneratorSpec{})
	if err != nil {
		return "", nil, err
	}
	output, err := inventory.RunKcl(ctx, g.client, inventoryGeneratorSpec.Kcl.Content)
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
