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

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/gobuffalo/flect"
	api "github.com/nubank/klaudete/api/v1alpha1"
	"github.com/nubank/klaudete/internal/dag"
	"github.com/nubank/klaudete/internal/exprs"
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
	_ = log.FromContext(ctx)

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

				_, err = reconciler.newResourceGroup(ctx, resourceGroupDefinition, expandedResourceGroupDefinitionGroup)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("unable to create ResourceGroup %s: %w", expandedResourceGroupDefinitionGroup.Name, err)
				}
			}

			resourceGroupDefinition.Status.Status = api.ResourceGroupDefinitionStatusReady
			_, err = reconciler.newResourceGroupDefinitionCondition(ctx, resourceGroupDefinition, &metav1.Condition{
				Type:    string(api.ConditionTypeReady),
				Status:  metav1.ConditionFalse,
				Reason:  string(api.ConditionTypeReady),
				Message: fmt.Sprintf("ResourceGroupDefinition is done. ResourceGroup %s expanded and created.", resourceGroupDefinition.Spec.Group.Name),
			})
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to update ResourceGroupDefinition's status: %w", err)
			}

			return ctrl.Result{}, nil
		}
	}

	expandedResourceGroupDefinitionGroup, err := expandResourceGroup(&dag.Args{}, &resourceGroupDefinition.Spec.Group)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to expand ResourceGroup definition: %w", err)
	}

	_, err = reconciler.newResourceGroup(ctx, resourceGroupDefinition, expandedResourceGroupDefinitionGroup)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to create ResourceGroup %s: %w", expandedResourceGroupDefinitionGroup.Name, err)
	}

	resourceGroupDefinition.Status.Status = api.ResourceGroupDefinitionStatusReady
	_, err = reconciler.newResourceGroupDefinitionCondition(ctx, resourceGroupDefinition, &metav1.Condition{
		Type:    string(api.ConditionTypeReady),
		Status:  metav1.ConditionFalse,
		Reason:  string(api.ConditionTypeReady),
		Message: fmt.Sprintf("ResourceGroupDefinition is done. ResourceGroup %s expanded and created.", resourceGroupDefinition.Spec.Group.Name),
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update ResourceGroupDefinition's status: %w", err)
	}

	return ctrl.Result{}, nil
}

func expandResourceGroup(args *dag.Args, resourceGroupDefinitionGroup *api.ResourceGroupDefinitionGroup) (*api.ResourceGroupDefinitionGroup, error) {
	newResourcesToGroup := make([]api.ResourceGroupResource, 0)

	for _, resource := range resourceGroupDefinitionGroup.Resources {
		// initialize Resource expansion

		// 'patches' field shouldn't be expanded
		patches := resource.Spec.Patches
		resource.Spec.Patches = nil

		resourceAsMap, err := serde.ToMap(resource)
		if err != nil {
			return nil, fmt.Errorf("failure to serialize spec.Resource to a map of properties: %w", err)
		}

		resourceToBeExpanded, err := dag.NewElement[api.ResourceGroupResource](resource.Name, resourceAsMap, expr.Exclude("provisioner"), expr.Exclude("resource"))
		if err != nil {
			return nil, fmt.Errorf("failure to process resource properties: %w", err)
		}

		expandedResource, err := resourceToBeExpanded.Evaluate(args)
		if err != nil {
			return nil, fmt.Errorf("failure to expand resource properties: %w", err)
		}

		resourceGroupResource, err := serde.FromMap(expandedResource, &api.ResourceGroupResource{})
		if err != nil {
			return nil, fmt.Errorf("failure to serialize a map of properties to ResourceGroupResource: %w", err)
		}

		// restore 'patches' content
		resourceGroupResource.Spec.Patches = patches
		newResourcesToGroup = append(newResourcesToGroup, *resourceGroupResource)

		// the resource spec itself can be used in expressions as a 'resource' variable (does it make sense?)
		args, _ = args.WithArgs(dag.ResourcesArg(resourceGroupResource.Name, expandedResource))
	}

	nameExpr, err := exprs.Parse(resourceGroupDefinitionGroup.Name)
	if err != nil {
		return nil, fmt.Errorf("failure to process ResourceGroup's name property: %w", err)
	}

	resourceGroupName, err := nameExpr.Evaluate(args.ToMap())
	if err != nil {
		return nil, fmt.Errorf("failure to evaluate ResourceGroup's name property: %w", err)
	}

	newResourceGroupDefinitionGroup := &api.ResourceGroupDefinitionGroup{
		Name:      resourceGroupName.(string),
		Resources: newResourcesToGroup,
	}

	return newResourceGroupDefinitionGroup, nil
}

func (reconciler *ResourceGroupDefinitionReconciler) newResourceGroup(ctx context.Context, resourceGroupDefinition *api.ResourceGroupDefinition, source *api.ResourceGroupDefinitionGroup) (*api.ResourceGroup, error) {
	name := flect.Dasherize(source.Name)

	newResourceGroup := &api.ResourceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: resourceGroupDefinition.Namespace,
			Labels: map[string]string{
				api.Group + "/managedBy.group":   resourceGroupDefinition.GroupVersionKind().Group,
				api.Group + "/managedBy.version": resourceGroupDefinition.GroupVersionKind().Version,
				api.Group + "/managedBy.kind":    resourceGroupDefinition.GroupVersionKind().Kind,
				api.Group + "/managedBy.name":    resourceGroupDefinition.Name,
				api.Group + "/managedBy.id":      string(resourceGroupDefinition.UID),
			},
		},
		Spec: api.ResourceGroupSpec{
			Resources: source.Resources,
		},
	}
	if err := ctrl.SetControllerReference(resourceGroupDefinition, newResourceGroup, reconciler.Scheme); err != nil {
		return nil, fmt.Errorf("unable to set ResourceGroups's ownerReference: %w", err)
	}
	if err := reconciler.Create(ctx, newResourceGroup); err != nil {
		return nil, fmt.Errorf("unable to create ResourceGroup %s: %w", source.Name, err)
	}
	return newResourceGroup, nil
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
