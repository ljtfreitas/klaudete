package provisioning

import (
	"context"
	"fmt"

	api "github.com/nubank/klaudete/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nubank/klaudete/internal/dag"
	"github.com/nubank/klaudete/internal/exprs"
	"github.com/nubank/klaudete/internal/exprs/expr"
	"github.com/nubank/klaudete/internal/provisioning/functions"
	"github.com/nubank/klaudete/internal/serde"
)

func NewProvisioner(client client.Client, dynamicClient *dynamic.DynamicClient, scheme *runtime.Scheme, resourceProvisioner *api.ResourceProvisioner) *Provisioner {
	return &Provisioner{
		client:              client,
		dynamicClient:       dynamicClient,
		scheme:              scheme,
		resourceProvisioner: resourceProvisioner,
	}
}

type Provisioner struct {
	client              client.Client
	dynamicClient       *dynamic.DynamicClient
	scheme              *runtime.Scheme
	resourceProvisioner *api.ResourceProvisioner
}

func (provisioner *Provisioner) Run(ctx context.Context, resource *api.Resource) (*ProvisioningStatus, error) {
	log := log.FromContext(ctx)

	// generate a dag from provisioner objects, to create them in dependency order
	resourceProvisionerObjectGraph := dag.NewGraph[api.ResourceProvisionerObject]()
	for _, candidate := range resource.Spec.Provisioner.Resources {
		resourceProvisionerObjAsMap, err := serde.ToMap(candidate.Ref)
		if err != nil {
			return nil, fmt.Errorf("failure to serialize provisioner object spec to a map of properties: %w", err)
		}
		resourceProvisionerObjElement, err := resourceProvisionerObjectGraph.NewElement(fmt.Sprintf("resources.%s", candidate.Name), resourceProvisionerObjAsMap)
		if err != nil {
			return nil, fmt.Errorf("failure to expand resource provisioner: %w", err)
		}
		resourceProvisionerObjElement.Ref = &candidate
	}

	resourceProvisionerObjects, err := resourceProvisionerObjectGraph.Sort()
	if err != nil {
		return nil, fmt.Errorf("unable to generate a graph from provisioner objects: %w", err)
	}

	log.Info(fmt.Sprintf("generated provisioner objects dag: %s", resourceProvisionerObjects))

	// the resource itself can be used in expressions as a 'resource' variable (does it make sense?)
	resourceAsMap, err := serde.ToMap(resource)
	if err != nil {
		return nil, fmt.Errorf("failure to serialize resource spec to a map of properties: %w", err)
	}
	args, err := dag.NewArgs(dag.ResourceArg(resourceAsMap))
	if err != nil {
		return nil, fmt.Errorf("failure to initialize args list: %w", err)
	}

	managedResourceProvisioner := NewManagedResourceProvisioner(provisioner.client, provisioner.dynamicClient, provisioner.scheme)

	state := ProvisioningSuccessState
	managedResources := make([]*ManagedResource, 0)

	for _, resourceProvisionerObj := range resourceProvisionerObjects {
		resourceToBeProvisioned, err := resourceProvisionerObjectGraph.Get(resourceProvisionerObj)
		if err != nil {
			return nil, err
		}

		log.Info(fmt.Sprintf("provisioning resource: %+v", resourceToBeProvisioned))

		logWithResource := log.WithValues("resource", resourceToBeProvisioned.Name)

		logWithResource.Info(fmt.Sprintf("processing provisioner %s...", resourceToBeProvisioned.Name))

		expandedResourceToBeProvisioned, err := resourceToBeProvisioned.Evaluate(args)
		if err != nil {
			return nil, fmt.Errorf("failure to expand resource provisioner object: %w", err)
		}

		dynamicObj := &unstructured.Unstructured{Object: expandedResourceToBeProvisioned}

		managedResource, err := managedResourceProvisioner.Apply(ctx, resource, resourceToBeProvisioned.Ref, dynamicObj)
		if err != nil {
			return nil, err
		}

		// update args with provisioner values
		args, err = args.WithArgs(dag.ProvisionerObjectArg(resourceToBeProvisioned.Ref.Name, managedResource.Object))
		if err != nil {
			return nil, fmt.Errorf("failed to update expression args map: %w", err)
		}

		secretExprFunction := expr.Function("secret", functions.Secrets(ctx, provisioner.client))

		if resourceToBeProvisioned.Ref.FailedWhen != nil {
			expression, err := exprs.Parse(*resourceToBeProvisioned.Ref.FailedWhen)
			if err != nil {
				return nil, fmt.Errorf("unable to parse FailedWhen expression from resource %s: %w", resource.Name, err)
			}
			r, err := expression.Evaluate(args.ToMap(), secretExprFunction)
			if err != nil {
				return nil, fmt.Errorf("unable to evaluate the FailedWhen expression: %w", err)
			}
			failed := r.(bool)
			if failed {
				failedProvisioningStatus := &ProvisioningStatus{
					State:     ProvisioningFailedState,
					Resources: []*ManagedResource{managedResource},
				}
				return failedProvisioningStatus, nil
			}
		}
		if resourceToBeProvisioned.Ref.ReadyWhen != nil {
			expression, err := exprs.Parse(*resourceToBeProvisioned.Ref.ReadyWhen)
			if err != nil {
				return nil, fmt.Errorf("unable to parse the ReadyWhen expression from : %w", err)
			}
			r, err := expression.Evaluate(args.ToMap(), secretExprFunction)
			if err != nil {
				return nil, fmt.Errorf("unable to evaluate the ReadyWhen expression: %w", err)
			}
			ready := r.(bool)
			if !ready {
				state = ProvisioningRunningState
			}
			if ready && resourceToBeProvisioned.Ref.Outputs != nil {
				outputsExpr, err := exprs.Parse(*resourceToBeProvisioned.Ref.Outputs)
				if err != nil {
					return nil, fmt.Errorf("unable to parse the Outputs expression: %w", err)
				}
				r, err := outputsExpr.Evaluate(args.ToMap(), secretExprFunction)
				if err != nil {
					return nil, fmt.Errorf("unable to evaluate the Outputs expression: %w", err)
				}
				if r != nil {
					managedResource.Outputs = r.(map[string]any)
				}
			}
		}
		managedResources = append(managedResources, managedResource)
	}

	managedResourceStatus := &ProvisioningStatus{
		State:     state,
		Resources: managedResources,
	}

	return managedResourceStatus, nil
}
