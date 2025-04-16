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
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	api "github.com/nubank/klaudete/api/v1alpha1"
	"github.com/nubank/klaudete/internal/inventory"
	inventoryv1alpha1 "github.com/nubank/nu-infra-inventory/api/gen/net/nuinfra/inventory/v1alpha1"
)

// ResourceTypeReconciler reconciles a ResourceType object
type ResourceTypeReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	InventoryClient *inventory.InventoryClient
}

// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resourcetypes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resourcetypes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resourcetypes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ResourceType object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (reconciler *ResourceTypeReconciler) Reconcile(ctx context.Context, resourceType *api.ResourceType) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if resourceType.Status.Status == "" || len(resourceType.Status.Conditions) == 0 {
		resourceType.Status.Status = api.ResourceTypeStatusPending
		resourceTypeWithCondition, err := reconciler.newResourceTypeCondition(ctx, resourceType, &metav1.Condition{
			Type:    string(api.ConditionTypePending),
			Status:  metav1.ConditionUnknown,
			Reason:  string(api.ConditionReasonPending),
			Message: "Starting reconciling...",
		})
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to update resource type status: %w", err)
		}
		resourceType = resourceTypeWithCondition
	}

	// TODO: update inventory...
	resourceTypeFromInventory, err := reconciler.InventoryClient.UpsertResourceType(ctx, &inventoryv1alpha1.ResourceType{
		Name:        resourceType.Spec.Name,
		Description: resourceType.Spec.Description,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to upsert resource type %s in Inventory: %w", resourceType.Name, err)
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if reconciler.Get(ctx, types.NamespacedName{Namespace: resourceType.Namespace, Name: resourceType.Name}, resourceType); err != nil {
			return fmt.Errorf("failure to refresh ResourceType instance: %w", err)
		}
		resourceType.Status.Status = api.ResourceTypeStatusInSync
		resourceType.Status.Inventory = api.ResourceTypeStatusInventory{
			Id: resourceTypeFromInventory.Id,
		}
		_, err = reconciler.newResourceTypeCondition(ctx, resourceType, &metav1.Condition{
			Type:    string(api.ConditionTypeInSync),
			Status:  metav1.ConditionTrue,
			Reason:  string(api.ConditionReasonInSync),
			Message: "Reconciling done. In-Sync.",
		})
		return err
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to update resource type status: %w", err)
	}

	log.Info("ResourceType was successfuly created and synced with Inventory.")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (reconciler *ResourceTypeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.ResourceType{}).
		WithEventFilter(checkObjectGenerationPredicate()).
		Complete(reconcile.AsReconciler(mgr.GetClient(), reconciler))
}

func (reconciler *ResourceTypeReconciler) newResourceTypeCondition(ctx context.Context, resourceType *api.ResourceType, newCondition *metav1.Condition) (*api.ResourceType, error) {
	meta.SetStatusCondition(&resourceType.Status.Conditions, *newCondition)
	if err := reconciler.Status().Update(ctx, resourceType); err != nil {
		return nil, err
	}
	if err := reconciler.Get(ctx, types.NamespacedName{Name: resourceType.Name}, resourceType); err != nil {
		return nil, err
	}
	return resourceType, nil
}
