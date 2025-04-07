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
	"time"

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
	"github.com/nubank/klaudete/internal/serde"
)

// ResourceGroupReconciler reconciles a ResourceGroup object
type ResourceGroupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resourcegroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resourcegroups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resourcegroups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ResourceGroup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (reconciler *ResourceGroupReconciler) Reconcile(ctx context.Context, resourceGroup *api.ResourceGroup) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if resourceGroup.Status.Phase == "" || len(resourceGroup.Status.Conditions) == 0 {
		resourceGroup.Status.Phase = api.ResourceGroupPhasePending
		resourceGroupWithCondition, err := reconciler.newResourceGroupCondition(ctx, resourceGroup, &metav1.Condition{
			Type:    string(api.ConditionTypePending),
			Status:  metav1.ConditionUnknown,
			Reason:  string(api.ConditionReasonPending),
			Message: "Starting reconciling...",
		})
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to update resource group status: %w", err)
		}
		resourceGroup = resourceGroupWithCondition
	}

	// TODO step 1 => initialize Resource expansion

	resourceGroup.Status.Phase = api.ResourceGroupPhaseInProgress
	resourceGroupWithCondition, err := reconciler.newResourceGroupCondition(ctx, resourceGroup, &metav1.Condition{
		Type:    string(api.ConditionTypePending),
		Status:  metav1.ConditionTrue,
		Reason:  string(api.ConditionReasonPending),
		Message: "Resource creation in progress...",
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to update resource group status: %w", err)
	}
	resourceGroup = resourceGroupWithCondition

	// 1.1 => traverse all resources to determine relationship between them

	resourceGroupGraph := dag.NewGraph[api.ResourceGroupResource]()

	for _, resource := range resourceGroup.Spec.Resources {
		// patches field should not be expanded
		patches := resource.Spec.Patches
		resource.Spec.Patches = nil

		resourceAsMap, err := serde.ToMap(resource)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to serialize resource spec to a map of properties: %w", err)
		}
		resourceProvisionerObjElement, err := resourceGroupGraph.NewElement(fmt.Sprintf("resources.%s", resource.Name), resourceAsMap)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to process resource properties: %w", err)
		}

		// restore patches content
		resource.Spec.Patches = patches

		resourceProvisionerObjElement.Ref = &resource
	}

	resourcesToBeProcessed, err := resourceGroupGraph.Sort()
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to generate a graph from resources: %w", err)
	}

	log.Info("generating resources...", "dag", fmt.Sprint(resourcesToBeProcessed))

	args, _ := dag.NewArgs()

	for _, candidate := range resourcesToBeProcessed {
		resourceToBeProcessed, err := resourceGroupGraph.Get(candidate)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("unable to find resource %s in graph: %w", candidate, err)
		}

		// check if the resource already exists
		resource, err := reconciler.getResource(ctx, resourceGroup, resourceToBeProcessed.Ref.Name)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure trying to find resource %s: %w", candidate, err)
		}
		if resource != nil {
			// check resource phase
			switch resource.Status.Phase {

			// still pending; just reconcile later
			case api.ResourceStatusPending, api.ResourceStatusProvisioningInProgress:
				return ctrl.Result{RequeueAfter: time.Duration(5) * time.Second}, nil

			// failed; cancel group processing
			case api.ResourceStatusFailed, api.ResourceStatusProvisioningFailed:
				resourceGroup.Status.Phase = api.ResourceGroupPhaseFailed
				_, err := reconciler.newResourceGroupCondition(ctx, resourceGroup, &metav1.Condition{
					Type:    string(api.ConditionTypePending),
					Status:  metav1.ConditionFalse,
					Reason:  string(api.ConditionReasonFailed),
					Message: fmt.Sprintf("ResourceGroup processing canceled; resource %s is failed.", resource.Name),
				})
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("failure to update resource group status: %w", err)
				}
				return ctrl.Result{}, nil

				// in other cases the resource is ready; go to the next
			}
		}

		log.Info("processing resource...", "resource", resourceToBeProcessed.Name)

		expandedResourceToBeProcessed, err := resourceToBeProcessed.Evaluate(args)

		expandedResource, err := serde.FromMap(expandedResourceToBeProcessed, &api.ResourceGroupResource{})
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to deserialize a map of properties to resource %s: %w", candidate, err)
		}

		// restore 'patches' content
		expandedResource.Spec.Patches = resourceToBeProcessed.Ref.Spec.Patches

		newResource, err := reconciler.newResource(ctx, resourceGroup, expandedResource)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("unable to create Resource %s: %w", expandedResource.Name, err)
		}

		newResourceAsMap, err := serde.ToMap(newResource)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to serialize properties from resource %s to map: %w", expandedResource.Name, err)
		}

		// update args with resource values
		args, err = args.WithArgs(dag.ResourcesArg(resourceToBeProcessed.Ref.Name, newResourceAsMap))
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update expression args map: %w", err)
		}
	}

	// TODO step 2 => resource group is ready

	resourceGroup.Status.Phase = api.ResourceGroupPhaseReady
	_, err = reconciler.newResourceGroupCondition(ctx, resourceGroup, &metav1.Condition{
		Type:    string(api.ConditionTypeReady),
		Status:  metav1.ConditionTrue,
		Reason:  string(api.ConditionReasonInSync),
		Message: "Resource group is done; all resources were created.",
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to update resource group status: %w", err)
	}

	return ctrl.Result{}, nil
}

func (reconciler *ResourceGroupReconciler) getResource(ctx context.Context, resourceGroup *api.ResourceGroup, resourceName string) (*api.Resource, error) {
	name := flect.Dasherize(resourceName)
	resource := &api.Resource{}
	err := reconciler.Client.Get(ctx, types.NamespacedName{Namespace: resourceGroup.Namespace, Name: name}, resource)
	return resource, client.IgnoreNotFound(err)
}

func (reconciler *ResourceGroupReconciler) newResource(ctx context.Context, resourceGroup *api.ResourceGroup, source *api.ResourceGroupResource) (*api.Resource, error) {
	name := flect.Dasherize(source.Spec.Name)

	newResource := &api.Resource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: resourceGroup.Namespace,
			Labels: map[string]string{
				api.Group + "/managedBy.group":   resourceGroup.GroupVersionKind().Group,
				api.Group + "/managedBy.version": resourceGroup.GroupVersionKind().Version,
				api.Group + "/managedBy.kind":    resourceGroup.GroupVersionKind().Kind,
				api.Group + "/managedBy.name":    resourceGroup.Name,
				api.Group + "/managedBy.id":      string(resourceGroup.UID),
			},
		},
		Spec: source.Spec,
	}
	if err := ctrl.SetControllerReference(resourceGroup, newResource, reconciler.Scheme); err != nil {
		return nil, fmt.Errorf("unable to set Resource's ownerReference: %w", err)
	}
	if err := reconciler.Create(ctx, newResource); err != nil {
		return nil, fmt.Errorf("unable to create Resource %s: %w", source.Name, err)
	}
	return newResource, nil
}

func (reconciler *ResourceGroupReconciler) newResourceGroupCondition(ctx context.Context, resourceGroup *api.ResourceGroup, newCondition *metav1.Condition) (*api.ResourceGroup, error) {
	meta.SetStatusCondition(&resourceGroup.Status.Conditions, *newCondition)
	if err := reconciler.Status().Update(ctx, resourceGroup); err != nil {
		return nil, err
	}
	if err := reconciler.Get(ctx, types.NamespacedName{Name: resourceGroup.Name, Namespace: resourceGroup.Namespace}, resourceGroup); err != nil {
		return nil, err
	}
	return resourceGroup, nil
}

// SetupWithManager sets up the controller with the Manager.
func (reconciler *ResourceGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.ResourceGroup{}).
		Complete(reconcile.AsReconciler(mgr.GetClient(), reconciler))
}
