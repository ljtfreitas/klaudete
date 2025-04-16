package inventory

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/nubank/klaudete/internal/serde"
	inventoryv1alpha1 "github.com/nubank/nu-infra-inventory/api/gen/net/nuinfra/inventory/v1alpha1"
	clientv1alpha1 "github.com/nubank/nu-infra-inventory/sdk/pkg/client"
	"github.com/nubank/nu-infra-inventory/sdk/pkg/kcllang"
	"github.com/nubank/nu-infra-inventory/sdk/pkg/resource"
	"github.com/nubank/nu-infra-inventory/sdk/pkg/resourcetype"
)

type InventoryClient struct {
	client clientv1alpha1.Client
}

func GenerateNurnFrom(resourceType string, name string) string {
	prefix := fmt.Sprintf("nurn:nu:infra:%s", resourceType)
	return GenerateNurn(prefix, name)
}

func GenerateNurn(prefix string, name string) string {
	return fmt.Sprintf("%s/%s", prefix, name)
}

func NewInventoryClient() (*InventoryClient, error) {
	client, err := clientv1alpha1.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failure to create Inventory client: :%w", err)
	}
	inventory := &InventoryClient{
		client: client,
	}
	return inventory, nil
}

func (inventory *InventoryClient) UpsertResourceType(ctx context.Context, resourceType *inventoryv1alpha1.ResourceType) (*inventoryv1alpha1.ResourceType, error) {
	client := inventory.client
	wrapper, err := client.ResourceTypes().Get(ctx, resourceType.Name)
	if err != nil {
		// in case of errors assume that ResourceType does not exist and we need to create it
		wrapper := &resourcetype.Wrapper{ResourceType: resourceType}

		wrapper, err = client.ResourceTypes().Create(ctx, wrapper)
		if err != nil {
			return nil, fmt.Errorf("failed to create resource type: %w", err)
		}
		return wrapper.ResourceType, nil
	}

	// update resource type
	wrapper.Description = resourceType.Description
	if err := wrapper.Sync(ctx); err != nil {
		return nil, fmt.Errorf("failed to update resource type: %w", err)
	}
	return wrapper.ResourceType, nil
}

func (inventory *InventoryClient) GetResourceType(ctx context.Context, name string) (*inventoryv1alpha1.ResourceType, error) {
	client := inventory.client
	wrapper, err := client.ResourceTypes().Get(ctx, name)

	if err != nil {
		return nil, fmt.Errorf("failed to get resource type: %w", err)
	}
	return wrapper.ResourceType, nil
}

func (inventory *InventoryClient) CreateResource(ctx context.Context, newResource *inventoryv1alpha1.Resource) (*inventoryv1alpha1.Resource, error) {
	client := inventory.client

	wrapper := &resource.Wrapper{Resource: newResource}

	wrapper, err := client.Resources().Create(ctx, wrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	return wrapper.Resource, nil
}

func (inventory *InventoryClient) UpdateResource(ctx context.Context, resource *inventoryv1alpha1.Resource) (*inventoryv1alpha1.Resource, error) {
	client := inventory.client

	nurn := resource.Metadata.Nurn

	// Fetch the current Resource from the API
	currentWrapper, err := client.Resources().Get(ctx, nurn)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource %s from Inventory: %w", nurn, err)
	}

	currentWrapper.Metadata.Alias = resource.Metadata.Alias
	currentWrapper.Metadata.Description = resource.Metadata.Description
	currentWrapper.Metadata.Properties = resource.Metadata.Properties

	if err := currentWrapper.Sync(ctx); err != nil {
		return nil, fmt.Errorf("failed to update resource %s: %w", nurn, err)
	}

	return currentWrapper.Resource, nil
}

type ConnectionTarget struct {
	Via        string
	TargetNurn string
}

func (inventory *InventoryClient) Connect(ctx context.Context, sourceNurn string, target *ConnectionTarget) error {
	client := inventory.client

	sourceRes, err := client.Resources().Get(ctx, sourceNurn)
	if err != nil {
		return err
	}

	targetRes, err := client.Resources().Get(ctx, target.TargetNurn)
	if err != nil {
		return err
	}

	err = sourceRes.Connect(ctx).To(targetRes).Via(target.Via)
	if err != nil {
		return fmt.Errorf("failed to connection resource: %w", err)
	}
	return nil
}

func (inventory *InventoryClient) RunKcl(ctx context.Context, kcl string) (map[string]any, error) {
	input := strings.NewReader(kcl)
	output := bytes.Buffer{}

	err := kcllang.Run(kcllang.RunArgs{
		Client:       inventory.client,
		Input:        input,
		Output:       &output,
		OutputFormat: kcllang.OutputFormatJSON,
		Options:      kcllang.Options{},
	})
	if err != nil {
		return nil, err
	}

	outputAsMap, err := serde.FromJsonString(output.String(), &map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("unable to deserialize kcl output to map: %w", err)
	}

	return *outputAsMap, nil
}
