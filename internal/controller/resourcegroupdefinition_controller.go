/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"maps"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/gobuffalo/flect"
	api "github.com/nubank/klaudete/api/v1alpha1"
	"github.com/nubank/klaudete/internal/dag"
	"github.com/nubank/klaudete/internal/exprs/expr"
	"github.com/nubank/klaudete/internal/generators"
	"github.com/nubank/klaudete/internal/serde"
)

// ResourceGroupDefinitionReconciler reconciles a ResourceGroupDefinition object
type ResourceGroupDefinitionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resourcegroupdefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resourcegroupdefinitions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resourcegroupdefinitions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ResourceGroupDefinition object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (reconciler *ResourceGroupDefinitionReconciler) Reconcile(ctx context.Context, resourceGroupDefinition *api.ResourceGroupDefinition) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if resourceGroupDefinition.Status.Status == "" || len(resourceGroupDefinition.Status.Conditions) == 0 {
		resourceGroupDefinition.Status.Status = api.ResourceGroupDefinitionStatusPending
		resourceGroupDefinitionWithCondition, err := reconciler.newResourceGroupDefinitionCondition(ctx, resourceGroupDefinition, &metav1.Condition{
			Type:    string(api.ConditionTypePending),
			Status:  metav1.ConditionUnknown,
			Reason:  string(api.ConditionReasonPending),
			Message: "Starting reconciling...",
		})
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to update resource group definition status: %w", err)
		}
		resourceGroupDefinition = resourceGroupDefinitionWithCondition
	}

	// TODO step 1 => check/expand generator

	generator := resourceGroupDefinition.Spec.Generator
	if generator != nil {
		generatorType, generatorSpec, err := generator.Spec()
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to read generator spec: %w", err)
		}
		generatorList, err := generators.NewGeneratorList(ctx, generatorType, generators.GeneratorSpec(generatorSpec))
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to expand generators values: %w", err)
		}

		// TODO step 1.1 => expand one resource group to each generator's value
		if generatorList != nil {

			args, _ := dag.NewArgs()

			for _, variable := range generatorList.Variables {
				args, err = args.WithArgs(dag.GeneratorArg(generatorList.Name, variable))
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("unable to initialize generator args: %w", err)
				}

				expandedResourceGroupDefinitionGroup, err := expandResourceGroup(args, &resourceGroupDefinition.Spec.Group)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("failure to expand ResourceGroup definition: %w", err)
				}

				_, err = reconciler.newOrUpdateResourceGroup(ctx, resourceGroupDefinition, expandedResourceGroupDefinitionGroup)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("unable to create ResourceGroup %s: %w", expandedResourceGroupDefinitionGroup.Name, err)
				}
			}

			err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resourceGroupDefinition.Namespace, Name: resourceGroupDefinition.Name}, resourceGroupDefinition); err != nil {
					return fmt.Errorf("failure to refresh ResourceDefinition instance: %w", err)
				}
				resourceGroupDefinition.Status.Status = api.ResourceGroupDefinitionStatusReady
				_, err = reconciler.newResourceGroupDefinitionCondition(ctx, resourceGroupDefinition, &metav1.Condition{
					Type:    string(api.ConditionTypeReady),
					Status:  metav1.ConditionTrue,
					Reason:  string(api.ConditionTypeReady),
					Message: fmt.Sprintf("ResourceGroupDefinition is done. ResourceGroup %s expanded and created.", resourceGroupDefinition.Spec.Group.Name),
				})
				return err
			})
			if err != nil {
				log.Error(err, "failure to update ResourceGroupDefinition status to Ready. Rescheduling...")
				return ctrl.Result{RequeueAfter: time.Second * 5}, nil
			}

			return ctrl.Result{}, nil
		}
	}

	expandedResourceGroupDefinitionGroup, err := expandResourceGroup(&dag.Args{}, &resourceGroupDefinition.Spec.Group)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to expand ResourceGroup definition: %w", err)
	}

	_, err = reconciler.newOrUpdateResourceGroup(ctx, resourceGroupDefinition, expandedResourceGroupDefinitionGroup)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to create ResourceGroup %s: %w", expandedResourceGroupDefinitionGroup.Name, err)
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resourceGroupDefinition.Namespace, Name: resourceGroupDefinition.Name}, resourceGroupDefinition); err != nil {
			return fmt.Errorf("failure to refresh ResourceDefinition instance: %w", err)
		}
		resourceGroupDefinition.Status.Status = api.ResourceGroupDefinitionStatusReady
		_, err = reconciler.newResourceGroupDefinitionCondition(ctx, resourceGroupDefinition, &metav1.Condition{
			Type:    string(api.ConditionTypeReady),
			Status:  metav1.ConditionTrue,
			Reason:  string(api.ConditionTypeReady),
			Message: fmt.Sprintf("ResourceGroupDefinition is done. ResourceGroup %s expanded and created.", resourceGroupDefinition.Spec.Group.Name),
		})
		return err
	})
	if err != nil {
		log.Error(err, "failure to update ResourceGroupDefinition status to Ready. Rescheduling...")
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	return ctrl.Result{}, nil
}

func expandResourceGroup(args *dag.Args, resourceGroupDefinitionGroup *api.ResourceGroupDefinitionGroup) (*api.ResourceGroupDefinitionGroup, error) {
	resourceGroupDefinitionGroupAsMap, err := serde.ToMap(resourceGroupDefinitionGroup)
	if err != nil {
		return nil, fmt.Errorf("failure to serialize spec.Group to a map of properties: %w", err)
	}

	resourceGroupDefinitionToBeExpanded, err := dag.NewElement[api.ResourceGroupResource](resourceGroupDefinitionGroup.Name, resourceGroupDefinitionGroupAsMap, expr.Only("generator"))
	if err != nil {
		return nil, fmt.Errorf("failure to process resource group properties: %w", err)
	}

	expandedResourceGroupDefinition, err := resourceGroupDefinitionToBeExpanded.Evaluate(args)
	if err != nil {
		return nil, fmt.Errorf("failure to expand resource group properties: %w", err)
	}

	newResourceGroupDefinitionGroup, err := serde.FromMap(expandedResourceGroupDefinition, &api.ResourceGroupDefinitionGroup{})
	if err != nil {
		return nil, fmt.Errorf("failure to serialize a map of properties to ResourceGroupResource: %w", err)
	}

	// newResourcesToGroup := make([]api.ResourceGroupResource, 0)

	// for _, resourceGroup := range resourceGroupDefinitionGroup.Resources {
	// 	// initialize Resource expansion

	// 	resourceGroupAsMap, err := serde.ToMap(resourceGroup)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failure to serialize spec.Resource to a map of properties: %w", err)
	// 	}

	// 	resourceToBeExpanded, err := dag.NewElement[api.ResourceGroupResource](resourceGroup.Name, resourceGroupAsMap, expr.Only("generator"))
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failure to process resource properties: %w", err)
	// 	}

	// 	expandedResource, err := resourceToBeExpanded.Evaluate(args)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failure to expand resource properties: %w", err)
	// 	}

	// 	resourceGroupResource, err := serde.FromMap(expandedResource, &api.ResourceGroupResource{})
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failure to serialize a map of properties to ResourceGroupResource: %w", err)
	// 	}

	// 	newResourcesToGroup = append(newResourcesToGroup, *resourceGroupResource)

	// 	// the resource spec itself can be used in expressions as a 'resource' variable (does it make sense?)
	// 	args, _ = args.WithArgs(dag.ResourcesArg(resourceGroupResource.Name, expandedResource))
	// }

	// nameExpr, err := exprs.Parse(resourceGroupDefinitionGroup.Name)
	// if err != nil {
	// 	return nil, fmt.Errorf("failure to process ResourceGroup's name property: %w", err)
	// }

	// resourceGroupName, err := nameExpr.Evaluate(args.ToMap())
	// if err != nil {
	// 	return nil, fmt.Errorf("failure to evaluate ResourceGroup's name property: %w", err)
	// }

	// newResourceGroupDefinitionGroup := &api.ResourceGroupDefinitionGroup{
	// 	Name:      resourceGroupName.(string),
	// 	Resources: newResourcesToGroup,
	// }

	return newResourceGroupDefinitionGroup, nil
}

func (reconciler *ResourceGroupDefinitionReconciler) newOrUpdateResourceGroup(ctx context.Context, resourceGroupDefinition *api.ResourceGroupDefinition, source *api.ResourceGroupDefinitionGroup) (*api.ResourceGroup, error) {
	name := source.ObjectMeta.GetName()
	if name == "" {
		name = source.Name
	}
	if name == "" {
		name = fmt.Sprintf("%s-resource-group", resourceGroupDefinition.Name)
	}
	name = flect.Dasherize(name)

	namespace := source.ObjectMeta.GetNamespace()
	if namespace == "" {
		namespace = resourceGroupDefinition.Namespace
	}

	resourceGroup := &api.ResourceGroup{}
	if err := reconciler.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, resourceGroup); err != nil {
		if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("unable to read ResourceGroup: %w", err)
		}

		labels := source.Labels
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[api.Group+"/managedBy.group"] = resourceGroupDefinition.GroupVersionKind().Group
		labels[api.Group+"/managedBy.version"] = resourceGroupDefinition.GroupVersionKind().Version
		labels[api.Group+"/managedBy.kind"] = resourceGroupDefinition.GroupVersionKind().Kind
		labels[api.Group+"/managedBy.name"] = resourceGroupDefinition.Name
		labels[api.Group+"/managedBy.id"] = string(resourceGroupDefinition.UID)

		// ResourceGroup does not exist yet
		resourceGroup = &api.ResourceGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Namespace:   namespace,
				Labels:      labels,
				Annotations: source.Annotations,
			},
			Spec: api.ResourceGroupSpec{
				Resources: source.Resources,
			},
		}
		if err := ctrl.SetControllerReference(resourceGroupDefinition, resourceGroup, reconciler.Scheme); err != nil {
			return nil, fmt.Errorf("unable to set ResourceGroups's ownerReference: %w", err)
		}
		if err := reconciler.Create(ctx, resourceGroup); err != nil {
			return nil, fmt.Errorf("unable to create ResourceGroup %s: %w", source.Name, err)
		}

		return resourceGroup, nil
	}

	// Resource already exists, update metadata and spec

	maps.Insert(resourceGroup.Labels, maps.All(source.Labels))
	maps.Insert(resourceGroup.Annotations, maps.All(source.Annotations))

	resourceGroup.Spec.Resources = source.Resources

	if err := reconciler.Update(ctx, resourceGroup); err != nil {
		return nil, fmt.Errorf("unable to update Resource: %w", err)
	}

	return resourceGroup, nil
}

func (reconciler *ResourceGroupDefinitionReconciler) newResourceGroupDefinitionCondition(ctx context.Context, resourceGroupDefinition *api.ResourceGroupDefinition, newCondition *metav1.Condition) (*api.ResourceGroupDefinition, error) {
	meta.SetStatusCondition(&resourceGroupDefinition.Status.Conditions, *newCondition)
	if err := reconciler.Status().Update(ctx, resourceGroupDefinition); err != nil {
		return nil, err
	}
	if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resourceGroupDefinition.Namespace, Name: resourceGroupDefinition.Name}, resourceGroupDefinition); err != nil {
		return nil, err
	}
	return resourceGroupDefinition, nil
}

// SetupWithManager sets up the controller with the Manager.
func (reconciler *ResourceGroupDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.ResourceGroupDefinition{}).
		Complete(reconcile.AsReconciler(mgr.GetClient(), reconciler))
}
