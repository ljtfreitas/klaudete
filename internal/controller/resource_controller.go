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
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	api "github.com/nubank/klaudete/api/v1alpha1"
)

// ResourceReconciler reconciles a Resource object
type ResourceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resources/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Resource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (reconciler *ResourceReconciler) Reconcile(ctx context.Context, resource *api.Resource) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	if resource.Status.Status == "" || len(resource.Status.Conditions) == 0 {
		resource.Status.Status = api.ResourceStatusPending
		resourceWithCondition, err := reconciler.newResourceCondition(ctx, resource, &metav1.Condition{
			Type:    string(api.ConditionTypePending),
			Status:  metav1.ConditionUnknown,
			Reason:  string(api.ConditionReasonPending),
			Message: "Starting reconciling...",
		})
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to update resource status: %w", err)
		}
		resource = resourceWithCondition
	}

	// TODO: step 1 => update inventory...

	// TODO: step 3 => provision resource

	resource.Status.Status = api.ResourceStatusProvisioningInProgress
	resource, err := reconciler.newResourceCondition(ctx, resource, &metav1.Condition{
		Type:    string(api.ConditionTypeInProgress),
		Status:  metav1.ConditionUnknown,
		Reason:  string(api.ConditionReasonInProgress),
		Message: "Initalizing provisioning from resource %s...",
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to update resource status: %w", err)
	}

	// TODO: step 3.1 => launch provisioning...

	resource.Status.Status = api.ResourceStatusReady
	resource, err = reconciler.newResourceCondition(ctx, resource, &metav1.Condition{
		Type:    string(api.ConditionTypeInSync),
		Status:  metav1.ConditionTrue,
		Reason:  string(api.ConditionReasonInSync),
		Message: "Provisioning was done. Resource is ready to be used.",
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to update resource status: %w", err)
	}

	reconciler.Recorder.Eventf(resource, "Normal", "Created", "Resource %s provisioned/in-sync.", resource.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (reconciler *ResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.Resource{}).
		Complete(reconcile.AsReconciler(mgr.GetClient(), reconciler))
}

func (reconciler *ResourceReconciler) newResourceCondition(ctx context.Context, resource *api.Resource, newCondition *metav1.Condition) (*api.Resource, error) {
	meta.SetStatusCondition(&resource.Status.Conditions, *newCondition)
	if err := reconciler.Status().Update(ctx, resource); err != nil {
		return nil, err
	}
	if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resource.Namespace, Name: resource.Name}, resource); err != nil {
		return nil, err
	}
	return resource, nil
}
