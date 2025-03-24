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
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	api "github.com/nubank/klaudete/api/v1alpha1"
	"github.com/nubank/klaudete/internal/dag"
	"github.com/nubank/klaudete/internal/generators"
)

// ResourceDefinitionReconciler reconciles a ResourceDefinition object
type ResourceDefinitionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resourcedefinitions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=klaudete.nubank.com.br,resources=resourcedefinitions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ResourceDefinition object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (reconciler *ResourceDefinitionReconciler) Reconcile(ctx context.Context, resourceDefinition *api.ResourceDefinition) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	if resourceDefinition.Status.Status == "" || len(resourceDefinition.Status.Conditions) == 0 {
		resourceDefinition.Status.Status = api.ResourceDefinitionStatusPending
		resourceDefinitionWithCondition, err := reconciler.newResourceDefinitionCondition(ctx, resourceDefinition, &metav1.Condition{
			Type:    string(api.ConditionTypePending),
			Status:  metav1.ConditionUnknown,
			Reason:  string(api.ConditionReasonPending),
			Message: "Starting reconciling...",
		})
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to update resource definition status: %w", err)
		}
		resourceDefinition = resourceDefinitionWithCondition
	}

	// TODO step 1 => resolve generators

	generatorList, err := generators.NewGeneratorList(ctx, resourceDefinition.Spec.Generator)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to expand generators values: %w", err)
	}

	// TODO step 2 => expand one resource to each generator value

	element, err := dag.NewElement[api.Resource](resourceDefinition.Name, resourceDefinition.Spec.Resource)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to expand resource template: %w", err)
	}

	args, err := dag.NewArgs()
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to initialize args list: %w", err)
	}

	if generatorList != nil {
		for _, variable := range generatorList.Variables {
			args, err = args.WithArgs(dag.GeneratorArg(generatorList.Name, variable))
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("unable to initialize generator args: %w", err)
			}

			resourceSpec := &api.ResourceSpec{}
			if err := json.Unmarshal(resourceDefinition.Spec.Resource.Raw, element); err != nil {
				return ctrl.Result{}, fmt.Errorf("failure to deserialize resource spec: %w", err)
			}
			if resourceSpec.Provisioner != nil {
				rawResourceProvisionerAsBytes, err := json.Marshal(resourceSpec.Provisioner)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("failure to serialize resource provisioner spec: %w", err)
				}
				resourceProvisionerElement, err := dag.NewElement[api.ResourceProvisioner](resourceSpec.Provisioner.Name, &runtime.RawExtension{Raw: rawResourceProvisionerAsBytes})
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("failure to expand resource provisioner: %w", err)
				}
				expandedResourceProvisioner, err := resourceProvisionerElement.Evaluate(args)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("failure to expand resource provisioner: %w", err)
				}

				args, err = args.WithArgs(dag.ProvisionerArg(resourceSpec.Provisioner.Name, expandedResourceProvisioner))
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("unable to initialize expression args: %w", err)
				}
			}

			expandedResource, err := element.Evaluate(args)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("unable to expand resource: %w", err)
			}
		}

		return ctrl.Result{}, nil
	}

	expandedResource, err := element.Evaluate(args)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to initialize expression args: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (reconciler *ResourceDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.ResourceDefinition{}).
		Complete(reconcile.AsReconciler(mgr.GetClient(), reconciler))
}

func (reconciler *ResourceDefinitionReconciler) newResourceDefinitionCondition(ctx context.Context, resourceDefinition *api.ResourceDefinition, newCondition *metav1.Condition) (*api.ResourceDefinition, error) {
	meta.SetStatusCondition(&resourceDefinition.Status.Conditions, *newCondition)
	if err := reconciler.Status().Update(ctx, resourceDefinition); err != nil {
		return nil, err
	}
	if err := reconciler.Get(ctx, types.NamespacedName{Name: resourceDefinition.Name}, resourceDefinition); err != nil {
		return nil, err
	}
	return resourceDefinition, nil
}
