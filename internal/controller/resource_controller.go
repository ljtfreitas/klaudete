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
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	api "github.com/nubank/klaudete/api/v1alpha1"
	"github.com/nubank/klaudete/internal/dag"
	"github.com/nubank/klaudete/internal/inventory"
	"github.com/nubank/klaudete/internal/patches"
	"github.com/nubank/klaudete/internal/provisioning"
	"github.com/nubank/klaudete/internal/serde"
	inventoryv1alpha1 "github.com/nubank/nu-infra-inventory/api/gen/net/nuinfra/inventory/v1alpha1"
)

// ResourceReconciler reconciles a Resource object
type ResourceReconciler struct {
	client.Client
	DynamicClient   *dynamic.DynamicClient
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	InventoryClient *inventory.InventoryClient
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

	log.Info("Starting resource reconciling...")

	if resource.Status.Phase == "" || len(resource.Status.Conditions) == 0 {
		resource.Status.Phase = api.ResourceStatusPending
		resourceWithCondition, err := reconciler.newResourceCondition(ctx, resource, &metav1.Condition{
			Type:    string(api.ConditionTypePending),
			Status:  metav1.ConditionUnknown,
			Reason:  string(api.ConditionReasonPending),
			Message: "Starting reconciling...",
		})
		if err != nil {
			log.Error(err, "failure to update resource status to Reconciling. Rescheduling...")
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}
		resource = resourceWithCondition
	}

	// check resource.resourceType; should be a valid ref
	resourceType := &api.ResourceType{}
	if err := reconciler.Get(ctx, types.NamespacedName{Name: resource.Spec.ResourceTypeRef}, resourceType); err != nil {
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
			log.Error(err, "failure to update resource status to Failure. Rescheduling...")
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}
		return ctrl.Result{}, nil
	}
	// resource type must be in sync with Inventory
	if resourceType.Status.Status != api.ResourceTypeStatusInSync {
		// try later...
		log.Info("ResourceType %s is not in-sync with inventory...", resourceType.Name)
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	// TODO: step 1 => update inventory...
	inventoryClient := reconciler.InventoryClient

	resourceTypeFromInventory, err := inventoryClient.GetResourceType(ctx, resourceType.Spec.Name)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to get resource type %s from Inventory: %w", resourceType.Spec.Name, err)
	}

	nurn := ""
	if resourceTypeFromInventory.GenerateNurn != nil {
		nurn = inventory.GenerateNurn(*resourceTypeFromInventory.GenerateNurn, resource.Spec.Name)
	} else {
		nurn = inventory.GenerateNurnFrom(resourceTypeFromInventory.Name, resource.Spec.Name)
	}

	resourcePropertiesAsMap, err := serde.ToMap(resource.Spec.Properties)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to serialize resource properties: %w", err)
	}

	resourceToInventoryProperties := map[string]*inventoryv1alpha1.PropertyValue{}
	for name, value := range resourcePropertiesAsMap {
		resourceToInventoryProperties[name] = &inventoryv1alpha1.PropertyValue{
			Value: &inventoryv1alpha1.PropertyValue_StringValue{
				StringValue: fmt.Sprint(value),
			},
		}
	}

	var resourceFromInventory *inventoryv1alpha1.Resource
	if resource.Status.Inventory == nil {
		// we suppose this resource does not exist in Inventory yet; try to insert
		resourceFromInventory, err = inventoryClient.CreateResource(ctx, &inventoryv1alpha1.Resource{
			Metadata: &inventoryv1alpha1.Metadata{
				Nurn:        nurn,
				Alias:       resource.Spec.Alias,
				Description: resource.Spec.Description,
				Properties:  resourceToInventoryProperties,
			},
			ResourceType: resourceTypeFromInventory,
		})
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to create resource type %s in Inventory: %w", resourceType.Name, err)
		}

		// update status
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resource.Namespace, Name: resource.Name}, resource); err != nil {
				return fmt.Errorf("failure to refresh resource instance: %w", err)
			}
			resource.Status.Inventory = &api.ResourceStatusInventory{
				Id:   resourceFromInventory.Id,
				Nurn: resourceFromInventory.Metadata.Nurn,
			}
			return reconciler.Status().Update(ctx, resource)
		})
		if err != nil {
			log.Error(err, fmt.Sprintf("failure to update resource %s with Inventory data. Rescheduling...", resourceType.Name))
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}
	} else {
		// resource already exists in inventory; load and update alias, description, and properties map
		resourceFromInventory, err = inventoryClient.UpdateResource(ctx, &inventoryv1alpha1.Resource{
			Metadata: &inventoryv1alpha1.Metadata{
				Nurn:        resource.Status.Inventory.Nurn,
				Alias:       resource.Spec.Alias,
				Description: resource.Spec.Description,
				Properties:  resourceToInventoryProperties,
			},
			ResourceType: resourceTypeFromInventory,
		})
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to update resource type %s in Inventory: %w", resourceType.Name, err)
		}
	}

	// here we know the resource exists in inventory, we can generate connections
	for _, connection := range resource.Spec.Connections {
		var targetNurn string

		target := connection.Target
		if targetRef := target.Ref; targetRef != nil {
			namespace := targetRef.Namespace
			if namespace == "" {
				namespace = resource.Namespace
			}
			targetResource := &api.Resource{}
			if reconciler.Get(ctx, types.NamespacedName{Namespace: namespace, Name: targetRef.Name}, targetResource); err != nil {
				log.Error(err, fmt.Sprintf("failure to get target resource %s to build a connnection with resource %s. Rescheduling...", targetRef.Name, resource.Name))
				return ctrl.Result{RequeueAfter: time.Second * 5}, nil
			}
			if targetResource.Status.Inventory == nil {
				log.Info(fmt.Sprintf("target resource %s isn't in-sync with Inventory. Rescheduling...", targetRef.Name))
				return ctrl.Result{RequeueAfter: time.Second * 5}, nil
			}
			targetNurn = targetResource.Status.Inventory.Nurn
		} else if target.Nurn != nil {
			targetNurn = target.Nurn.Value

		} else {
			// invalid connection; mark resource as failed and stop reconcilation
			log.Error(fmt.Errorf("invalid connection spec; target.Ref or target.Nurn are required"), "invalid connection spec")

			// just update status and stop reconciling; invalid spec
			resource.Status.Phase = api.ResourceStatusFailed
			_, err := reconciler.newResourceCondition(ctx, resource, &metav1.Condition{
				Type:    string(api.ConditionTypeFailure),
				Status:  metav1.ConditionFalse,
				Reason:  "InvalidResourceTypeRef",
				Message: "Invalid connection spec; target.Ref or target.Nurn are required",
			})
			if err != nil {
				log.Error(err, "failure to update resource status to Failure. Rescheduling...")
				return ctrl.Result{RequeueAfter: time.Second * 5}, nil
			}
			return ctrl.Result{}, nil
		}
		err = inventoryClient.Connect(ctx, resourceFromInventory.Metadata.Nurn, &inventory.ConnectionTarget{
			Via:        connection.Via,
			TargetNurn: targetNurn,
		})
		if err != nil {
			log.Error(err, fmt.Sprintf("failure to create connection with target resource %s. Rescheduling...", targetNurn))
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}
	}

	// TODO: step 2 => provision resource
	resourceProvisioner := resource.Spec.Provisioner
	if resourceProvisioner != nil {
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resource.Namespace, Name: resource.Name}, resource); err != nil {
				return fmt.Errorf("failure to refresh resource instance: %w", err)
			}
			resource.Status.Phase = api.ResourceStatusProvisioningInProgress
			resourceWithCondition, err := reconciler.newResourceCondition(ctx, resource, &metav1.Condition{
				Type:    string(api.ConditionTypeInProgress),
				Status:  metav1.ConditionUnknown,
				Reason:  string(api.ConditionReasonInProgress),
				Message: "Initalizing provisioning from resource...",
			})
			if resourceWithCondition != nil {
				resource = resourceWithCondition
			}
			return err
		})
		if err != nil {
			log.Error(err, "failure to update resource status to InProgress. Rescheduling...")
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}

		provisioner := provisioning.NewProvisioner(reconciler.Client, reconciler.DynamicClient, reconciler.Scheme, resourceProvisioner)

		log.Info("running provisioner...")

		provisioningStatus, err := provisioner.Run(ctx, resource)
		if err != nil {
			log.Error(err, "unable to launch provisioners and creating managed resources")

			resource.Status.Phase = api.ResourceStatusProvisioningFailed
			_, _ = reconciler.newResourceCondition(ctx, resource, &metav1.Condition{
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

		if provisioningStatus.Resources != nil {
			resources := make([]api.ResourceStatusProvisionerObject, 0)
			allOutputs := make(map[string]any)

			for _, r := range provisioningStatus.Resources {
				resources = append(resources, api.ResourceStatusProvisionerObject{
					ApiVersion: r.GetAPIVersion(),
					Kind:       r.Kind,
					Name:       r.GetName(),
					UId:        string(r.GetUID()),
				})

				maps.Insert(allOutputs, maps.All(r.Outputs))
			}

			log.Info("Outputs generated from provisioning:", "outputs", fmt.Sprint(allOutputs))

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

		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			resourceToBeUpdated := &api.Resource{}
			if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resource.Namespace, Name: resource.Name}, resourceToBeUpdated); err != nil {
				return fmt.Errorf("failure to refresh resource instance: %w", err)
			}

			resourceToBeUpdated.Status.Phase = api.ResourceStatusPhaseDescription(phase)
			resourceToBeUpdated.Status.AtProvisioner = resource.Status.AtProvisioner

			resourceWithProvisioner, err := reconciler.newResourceCondition(ctx, resourceToBeUpdated, condition)
			if resourceWithProvisioner != nil {
				resource = resourceWithProvisioner
			}
			return err
		})
		if err != nil {
			log.Error(err, "failure to update resource status with provisioning data. Rescheduling...")
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}

		if provisioningStatus.IsRunning() {
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
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

			newPatches := make(api.ResourcePatches, 0, len(resourcePatches))
			for _, patch := range resourcePatches {
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

				err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
					if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resource.Namespace, Name: resource.Name}, resource); err != nil {
						return fmt.Errorf("failure to refresh resource instance: %w", err)
					}
					resource.Status = *newResourceStatus
					return reconciler.Status().Update(ctx, resource)
				})
				if err != nil {
					log.Error(err, "failure to update resource status after calculate patches. Rescheduling...")
					return ctrl.Result{RequeueAfter: time.Second * 5}, nil
				}

				// update Inventory with new properties after patches
				if resource.Status.Inventory != nil && resource.Status.Inventory.Properties != nil {
					inventoryPropertiesAsMap, err := serde.ToMap(resource.Status.Inventory.Properties)
					if err != nil {
						return ctrl.Result{}, fmt.Errorf("failure to serialize properties to be sent to inventory: %w", err)
					}

					for name, value := range inventoryPropertiesAsMap {
						resourceToInventoryProperties[name] = &inventoryv1alpha1.PropertyValue{
							Value: &inventoryv1alpha1.PropertyValue_StringValue{
								StringValue: fmt.Sprint(value),
							},
						}
					}
					resourceFromInventory.Metadata.Properties = resourceToInventoryProperties
					resourceFromInventory, err = inventoryClient.UpdateResource(ctx, resourceFromInventory)
					if err != nil {
						return ctrl.Result{}, fmt.Errorf("failure to update resource type %s in Inventory: %w", resourceType.Name, err)
					}
				}
			}
		}
	}

	// TODO step 5: resource is done.
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resource.Namespace, Name: resource.Name}, resource); err != nil {
			return fmt.Errorf("failure to refresh resource instance: %w", err)
		}
		resource.Status.Phase = api.ResourceStatusReady
		_, err := reconciler.newResourceCondition(ctx, resource, &metav1.Condition{
			Type:    string(api.ConditionTypeInSync),
			Status:  metav1.ConditionTrue,
			Reason:  string(api.ConditionReasonInSync),
			Message: "Resource is ready to be used.",
		})
		return err
	})
	if err != nil {
		log.Error(err, "failure to update resource status to Ready. Rescheduling...")
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	reconciler.Recorder.Eventf(resource, "Normal", "Created", "Resource %s provisioned/in-sync.", resource.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (reconciler *ResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.Resource{}).
		WithEventFilter(checkObjectGenerationPredicate()).
		Complete(reconcile.AsReconciler(mgr.GetClient(), reconciler))
}

func (reconciler *ResourceReconciler) newResourceCondition(ctx context.Context, resource *api.Resource, newCondition *metav1.Condition) (*api.Resource, error) {
	meta.SetStatusCondition(&resource.Status.Conditions, *newCondition)
	if err := reconciler.Status().Update(ctx, resource); err != nil {
		return nil, err
	}
	r := &api.Resource{}
	if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resource.Namespace, Name: resource.Name}, r); err != nil {
		return nil, err
	}
	return r, nil
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
