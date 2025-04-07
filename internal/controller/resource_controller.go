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
	"maps"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	api "github.com/nubank/klaudete/api/v1alpha1"
	"github.com/nubank/klaudete/internal/dag"
	"github.com/nubank/klaudete/internal/patches"
	"github.com/nubank/klaudete/internal/provisioning"
	"github.com/nubank/klaudete/internal/serde"
)

// ResourceReconciler reconciles a Resource object
type ResourceReconciler struct {
	client.Client
	DynamicClient *dynamic.DynamicClient
	Scheme        *runtime.Scheme
	Recorder      record.EventRecorder
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
	log := log.FromContext(ctx)

	resource.Status.Phase = api.ResourceStatusPending
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

	// check resource.resourceType; should be a valid ref
	if err := reconciler.Get(ctx, types.NamespacedName{Name: resource.Spec.ResourceTypeRef}, &api.ResourceType{}); err != nil {
		log.Error(fmt.Errorf("unable to find resourceType %s: %w", resource.Spec.ResourceTypeRef, err), "unable to find resourceType")

		// just update status and stop reconciling; invalid spec
		resource.Status.Phase = api.ResourceStatusFailed
		_, err := reconciler.newResourceCondition(ctx, resource, &metav1.Condition{
			Type:    string(api.ConditionTypeFailure),
			Status:  metav1.ConditionFalse,
			Reason:  "InvalidResourceTypeRef",
			Message: fmt.Sprintf("Unable to find resourceType %s", resource.Spec.ResourceTypeRef),
		})
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to update resource status: %w", err)
		}
		return ctrl.Result{}, nil
	}

	// TODO: step 1 => update inventory...

	// TODO: step 2 => provision resource
	resource.Status.Phase = api.ResourceStatusProvisioningInProgress
	resource, err = reconciler.newResourceCondition(ctx, resource, &metav1.Condition{
		Type:    string(api.ConditionTypeInProgress),
		Status:  metav1.ConditionUnknown,
		Reason:  string(api.ConditionReasonInProgress),
		Message: "Initalizing provisioning from resource...",
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to update resource status: %w", err)
	}

	provisioner := provisioning.NewProvisioner(reconciler.Client, reconciler.DynamicClient, reconciler.Scheme, resource.Spec.Provisioner)

	log.Info("running provisioner...")

	provisioningStatus, err := provisioner.Run(ctx, resource)
	if err != nil {
		log.Error(err, "unable to launch provisioners and creating managed resources")

		resource.Status.Phase = api.ResourceStatusProvisioningFailed
		_, err := reconciler.newResourceCondition(ctx, resource, &metav1.Condition{
			Type:    string(api.ConditionTypeFailure),
			Status:  metav1.ConditionFalse,
			Reason:  string(api.ConditionReasonFailed),
			Message: "failed to run provisioner",
		})

		return ctrl.Result{}, err
	}

	log.Info(fmt.Sprintf("current state from provisioning is [%s]", provisioningStatus.State))

	// TODO: step 3 => update resource status with provisioner data

	phase, condition := statusToCondition(provisioningStatus, resource)

	resource.Status.Phase = api.ResourceStatusPhaseDescription(phase)

	if provisioningStatus.Resources != nil {
		resources := make([]api.ResourceStatusProvisionerObject, 0)
		allOutputs := make(map[string]any)

		for _, r := range provisioningStatus.Resources {
			resources = append(resources, api.ResourceStatusProvisionerObject{
				Group:   r.Group,
				Version: r.Version,
				Kind:    r.Kind,
				Name:    r.GetName(),
			})

			maps.Insert(allOutputs, maps.All(r.Outputs))
		}

		rawOutputs, err := json.Marshal(allOutputs)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to serialize provisioner resource outputs: %w", err)
		}

		resource.Status.AtProvisioner = api.ResourceStatusProvisioner{
			State:     string(provisioningStatus.State),
			Resources: resources,
			Outputs:   &runtime.RawExtension{Raw: rawOutputs},
		}
	}

	_, err = reconciler.newResourceCondition(ctx, resource, condition)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to update resource status: %w", err)
	}

	if provisioningStatus.IsRunning() {
		return ctrl.Result{RequeueAfter: time.Duration(5) * time.Second}, nil
	}

	// TODO step 4: run patches

	resourcePatches := resource.Spec.Patches
	if resourcePatches != nil && len(resourcePatches) != 0 {
		// inject resource as a variable
		resourceAsMap, err := serde.ToMap(resource)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to serialize resource to a map of properties: %w", err)
		}

		args, err := dag.NewArgs(dag.ResourceArg(resourceAsMap))
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to initialize patches args: %w", err)
		}

		// inject provisioner data as a variable
		for _, managedResource := range provisioningStatus.Resources {
			args, err = args.WithArgs(dag.ProvisionerObjectArg(managedResource.Name, managedResource.Object))
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failure to initialize patch args: %w", err)
			}
		}

		newPatches := make(api.ResourcePatches, 0, len(resource.Spec.Patches))
		for _, patch := range resource.Spec.Patches {
			patchAsMap, err := serde.ToMap(patch)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failure to serialize patch to map: %w", err)
			}
			patchToBeExpanded, err := dag.NewElement[api.ResourcePatches]("patch", patchAsMap)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failure to generate a dag.Element to a resource.Patch instance: %w", err)
			}
			expandedPatch, err := patchToBeExpanded.Evaluate(args)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failure to expand patches: %w", err)
			}
			newPatch, err := serde.FromMap(expandedPatch, &api.ResourcePatch{})
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failure to deserialize patch properties: %w", err)
			}
			newPatches = append(newPatches, *newPatch)
		}

		if len(newPatches) != 0 {
			resourceStatusAsMap, err := serde.ToMap(resource.Status)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failure to serialize resource status: %w", err)
			}
			newResourceStatusAsMap := map[string]any{
				"status": resourceStatusAsMap,
			}

			for _, patch := range newPatches {
				newResourceStatusAsMap = patches.ApplyTo(patch, newResourceStatusAsMap)
			}

			newResourceStatus, err := serde.FromMap(newResourceStatusAsMap["status"].(map[string]any), &api.ResourceStatus{})
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failure to deserialize resource status: %w", err)
			}

			resource.Status = *newResourceStatus
			if err := reconciler.Status().Update(ctx, resource); err != nil {
				return ctrl.Result{}, fmt.Errorf("failure to update resource status: %w", err)
			}
			if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resource.Namespace, Name: resource.Name}, resource); err != nil {
				return ctrl.Result{}, fmt.Errorf("failure to refresh resource instance: %w", err)
			}
		}
	}

	// TODO step 5: provisioning is done.

	resource.Status.Phase = api.ResourceStatusReady
	resource, err = reconciler.newResourceCondition(ctx, resource, &metav1.Condition{
		Type:    string(api.ConditionTypeInSync),
		Status:  metav1.ConditionTrue,
		Reason:  string(api.ConditionReasonInSync),
		Message: "Provisioning is done. Resource is ready to be used.",
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

func statusToCondition(status *provisioning.ProvisioningStatus, resource *api.Resource) (api.ResourceStatusPhaseDescription, *metav1.Condition) {
	switch status.State {
	case provisioning.ProvisioningSuccessState:
		return api.ResourceStatusReady, &metav1.Condition{
			Type:    string(api.ConditionTypeReady),
			Status:  metav1.ConditionTrue,
			Reason:  string(api.ConditionReasonDone),
			Message: fmt.Sprintf("Provisioning from Resource %s was successfully done.", resource.Name),
		}
	case provisioning.ProvisioningFailedState:
		return api.ResourceStatusProvisioningFailed, &metav1.Condition{
			Type:    string(api.ConditionTypeFailure),
			Status:  metav1.ConditionFalse,
			Reason:  string(api.ConditionReasonFailed),
			Message: fmt.Sprintf("Provisioning from Resource %s failed", resource.Name),
		}
	default:
		return api.ResourceStatusProvisioningInProgress, &metav1.Condition{
			Type:    string(api.ConditionTypeReady),
			Status:  metav1.ConditionUnknown,
			Reason:  string(api.ConditionReasonInProgress),
			Message: fmt.Sprintf("Provisioning from Resource %s is running...", resource.Name),
		}
	}
}
