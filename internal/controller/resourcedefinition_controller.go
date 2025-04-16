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
	log := log.FromContext(ctx)

	log.Info("Starting resource definition reconciling...")

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

	log.Info("processing resource...")

	// TODO step 1 => initialize Resource expansion

	// // 'patches' field shouldn't be expanded
	// patches := resourceDefinition.Spec.Resource.Spec.Patches
	// resourceDefinition.Spec.Resource.Spec.Patches = nil

	resourceAsMap, err := serde.ToMap(resourceDefinition.Spec.Resource)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to serialize spec.Resource to a map of properties: %w", err)
	}

	resourceToBeExpanded, err := dag.NewElement[api.ResourceDefinitionResource](resourceDefinition.Name, resourceAsMap, expr.Only("generator"))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to expand resource template: %w", err)
	}

	// the resource spec itself can be used in expressions as a 'resource' variable (does it make sense?)
	args, _ := dag.NewArgs(dag.ResourceArg(resourceAsMap))

	// TODO step 2 => check generator; if present, resolve to a list of values

	knowResources := make(api.ResourceDefinitionResourcesStatus, 0)

	generator := resourceDefinition.Spec.Generator
	if generator != nil {
		generatorType, generatorSpec, err := generator.Spec()
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to read generator spec: %w", err)
		}
		generatorList, err := generators.NewGeneratorList(ctx, generatorType, generators.GeneratorSpec(generatorSpec))
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to expand generators values: %w", err)
		}

		if generatorList != nil {

			// TODO step 2.1 => expand one resource to each generator's value

			for _, variable := range generatorList.Variables {
				args, err = args.WithArgs(dag.GeneratorArg(generatorList.Name, variable))
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("unable to initialize generator args: %w", err)
				}

				provisioner := resourceDefinition.Spec.Resource.Spec.Provisioner
				if provisioner != nil {
					resourceProvisionerAsMap, err := serde.ToMap(provisioner)
					if err != nil {
						return ctrl.Result{}, fmt.Errorf("failure to serialize spec.Resource.Provisioner field to a map of properties: %w", err)
					}
					resourceProvisionerElement, err := dag.NewElement[api.ResourceProvisioner]("provisioner", resourceProvisionerAsMap, expr.Only("generator"))
					if err != nil {
						return ctrl.Result{}, fmt.Errorf("failure to expand resource provisioner: %w", err)
					}
					expandedResourceProvisioner, err := resourceProvisionerElement.Evaluate(args)
					if err != nil {
						return ctrl.Result{}, fmt.Errorf("failure to expand resource provisioner: %w", err)
					}

					_, err = args.WithArgs(dag.ProvisionerArg(expandedResourceProvisioner))
					if err != nil {
						return ctrl.Result{}, fmt.Errorf("unable to initialize expression args: %w", err)
					}
				}

				expandedResourceProperties, err := resourceToBeExpanded.Evaluate(args)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("unable to expand resource properties: %w", err)
				}

				log.Info("expandedResourceProperties", "map", fmt.Sprint(expandedResourceProperties))

				expandedResource, err := serde.FromMap(expandedResourceProperties, &api.ResourceDefinitionResource{})
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("unable to serialize properties to Resource spec: %w", err)
				}

				resource, err := reconciler.newOrUpdateResource(ctx, resourceDefinition, expandedResource)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("unable to create Resource: %w", err)
				}

				knowResources = append(knowResources, api.ResourceDefinitionResourceStatus{
					ApiVersion: resource.APIVersion,
					Kind:       resource.Kind,
					Name:       resource.Name,
					UId:        string(resource.UID),
				})
			}

			err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resourceDefinition.Namespace, Name: resourceDefinition.Name}, resourceDefinition); err != nil {
					return fmt.Errorf("failure to refresh ResourceDefinition instance: %w", err)
				}
				resourceDefinition.Status.Status = api.ResourceDefinitionStatusReady
				resourceDefinition.Status.Resources = knowResources
				_, err = reconciler.newResourceDefinitionCondition(ctx, resourceDefinition, &metav1.Condition{
					Type:    string(api.ConditionTypeReady),
					Status:  metav1.ConditionTrue,
					Reason:  string(api.ConditionTypeReady),
					Message: "ResourceDefinition is done. Resources were expanded and created.",
				})
				return err
			})
			if err != nil {
				log.Error(err, "failure to update ResourceDefinition status to Ready. Rescheduling...")
				return ctrl.Result{RequeueAfter: time.Second * 5}, nil
			}

			return ctrl.Result{}, nil
		}
	}

	expandedResourceProperties, err := resourceToBeExpanded.Evaluate(args)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to expand resource properties: %w", err)
	}

	expandedResource, err := serde.FromMap(expandedResourceProperties, &api.ResourceDefinitionResource{})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to serialize properties to Resource spec: %w", err)
	}

	resource, err := reconciler.newOrUpdateResource(ctx, resourceDefinition, expandedResource)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to create Resource: %w", err)
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resourceDefinition.Namespace, Name: resourceDefinition.Name}, resourceDefinition); err != nil {
			return fmt.Errorf("failure to refresh ResourceDefinition instance: %w", err)
		}
		resourceDefinition.Status.Status = api.ResourceDefinitionStatusReady
		resourceDefinition.Status.Resources = api.ResourceDefinitionResourcesStatus{
			api.ResourceDefinitionResourceStatus{
				ApiVersion: resource.APIVersion,
				Kind:       resource.Kind,
				Name:       resource.Name,
				UId:        string(resource.UID),
			},
		}
		_, err = reconciler.newResourceDefinitionCondition(ctx, resourceDefinition, &metav1.Condition{
			Type:    string(api.ConditionTypeReady),
			Status:  metav1.ConditionTrue,
			Reason:  string(api.ConditionTypeReady),
			Message: "ResourceDefinition is done. Resources were expanded and created.",
		})
		return err
	})
	if err != nil {
		log.Error(err, "failure to update ResourceDefinition status to Ready. Rescheduling...")
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	return ctrl.Result{}, nil
}

func (reconciler *ResourceDefinitionReconciler) newOrUpdateResource(ctx context.Context, resourceDefinition *api.ResourceDefinition, source *api.ResourceDefinitionResource) (*api.Resource, error) {
	name := source.GetName()
	if name == "" {
		name = source.Spec.Name
	}
	if name == "" {
		name = fmt.Sprintf("%s-resource", resourceDefinition.Name)
	}
	name = flect.Dasherize(name)

	namespace := source.Namespace
	if namespace == "" {
		namespace = resourceDefinition.Namespace
	}

	resource := &api.Resource{}
	if err := reconciler.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, resource); err != nil {
		if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("unable to read Resource: %w", err)
		}

		// Resource doesn't exist yet

		labels := source.Labels
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[api.Group+"/managedBy.group"] = resourceDefinition.GroupVersionKind().Group
		labels[api.Group+"/managedBy.version"] = resourceDefinition.GroupVersionKind().Version
		labels[api.Group+"/managedBy.kind"] = resourceDefinition.GroupVersionKind().Kind
		labels[api.Group+"/managedBy.name"] = resourceDefinition.Name
		labels[api.Group+"/managedBy.id"] = string(resourceDefinition.UID)

		resource = &api.Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Namespace:   namespace,
				Labels:      labels,
				Annotations: source.Annotations,
			},
			Spec: source.Spec,
		}
		if err := ctrl.SetControllerReference(resourceDefinition, resource, reconciler.Scheme); err != nil {
			return nil, fmt.Errorf("unable to set Resource's ownerReference: %w", err)
		}
		if err := reconciler.Create(ctx, resource); err != nil {
			return nil, fmt.Errorf("unable to create Resource: %w", err)
		}

		return resource, nil
	}

	// Resource already exists, update metadata and spec

	maps.Insert(resource.Labels, maps.All(source.Labels))
	maps.Insert(resource.Annotations, maps.All(source.Annotations))

	resource.Spec = source.Spec

	if err := reconciler.Update(ctx, resource); err != nil {
		return nil, fmt.Errorf("unable to update Resource: %w", err)
	}

	return resource, nil
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
	if err := reconciler.Get(ctx, types.NamespacedName{Namespace: resourceDefinition.Namespace, Name: resourceDefinition.Name}, resourceDefinition); err != nil {
		return nil, err
	}
	return resourceDefinition, nil
}
