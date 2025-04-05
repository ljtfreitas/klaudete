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

	corev1 "k8s.io/api/core/v1"
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
	"github.com/nubank/klaudete/internal/generators"
	"github.com/nubank/klaudete/internal/serde"
	"k8s.io/apimachinery/pkg/api/errors"
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

	// TODO step 1 => generate a dedicated namespace

	namespace := &corev1.Namespace{}
	namespaceName := flect.Dasherize(resourceDefinition.Name)
	if err := reconciler.Get(ctx, types.NamespacedName{Name: namespaceName}, namespace); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("unable to fetch ResourceDefinition namespace: %w", err)
		}

		log.Info(fmt.Sprintf("there is no namespace to ResourceDefinition %s; trying to generate...", resourceDefinition.Name))

		namespace.Name = resourceDefinition.Name
		namespace.Labels = map[string]string{
			api.Group + "/managedBy.group":   resourceDefinition.GroupVersionKind().Group,
			api.Group + "/managedBy.version": resourceDefinition.GroupVersionKind().Version,
			api.Group + "/managedBy.kind":    resourceDefinition.GroupVersionKind().Kind,
			api.Group + "/managedBy.name":    resourceDefinition.Name,
			api.Group + "/managedBy.id":      string(resourceDefinition.UID),
		}
		if err := ctrl.SetControllerReference(resourceDefinition, namespace, reconciler.Scheme); err != nil {
			return ctrl.Result{}, fmt.Errorf("unable to set namespace's ownerReference: %w", err)
		}

		if err := reconciler.Create(ctx, namespace); err != nil {
			log.Error(err, fmt.Sprintf("unable to create namespace %s", namespace.Name), "namespace", namespace.Name)

			_, err = reconciler.newResourceDefinitionCondition(ctx, resourceDefinition, &metav1.Condition{
				Type:    string(api.ConditionTypeInSync),
				Status:  metav1.ConditionFalse,
				Reason:  "ResourceDefinitionNamespaceCreationFailed",
				Message: fmt.Sprintf("unable to create a namespace to ResourceDefinition %s", resourceDefinition.Name),
			})
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to update ResourceDefinition's status: %w", err)
			}

			return ctrl.Result{}, err
		}

		log.Info(fmt.Sprintf("a namespace was created to ResourceDefinition %s", resourceDefinition.Name))
	}

	// TODO step 2 => initialize Resource expansion

	// 'patches' field shouldn't be expanded
	patches := resourceDefinition.Spec.Resource.Spec.Patches
	resourceDefinition.Spec.Resource.Spec.Patches = nil

	resourceAsMap, err := serde.ToMap(resourceDefinition.Spec.Resource)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to serialize spec.Resource to a map of properties: %w", err)
	}

	element, err := dag.NewElement[api.Resource](resourceDefinition.Name, resourceAsMap)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to expand resource template: %w", err)
	}

	// the resource spec itself can be used in expressions as a 'resource' variable (does it make sense?)
	args, err := dag.NewArgs(dag.ResourceArg(resourceAsMap))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failure to initialize args list: %w", err)
	}

	// TODO step 3 => check generator; if present, resolve to a list of values

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

			// TODO step 4 => expand one resource to each generator's value

			for _, variable := range generatorList.Variables {
				args, err = args.WithArgs(dag.GeneratorArg(generatorList.Name, variable))
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("unable to initialize generator args: %w", err)
				}

				resourceSpec := resourceDefinition.Spec.Resource.Spec
				if resourceSpec.Provisioner != nil {
					resourceProvisionerAsMap, err := serde.ToMap(resourceSpec.Provisioner)
					if err != nil {
						return ctrl.Result{}, fmt.Errorf("failure to serialize spec.Resource.Provisioner field to a map of properties: %w", err)
					}
					resourceProvisionerElement, err := dag.NewElement[api.ResourceProvisioner]("provisioner", resourceProvisionerAsMap)
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

				expandedResourceProperties, err := element.Evaluate(args)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("unable to expand resource properties: %w", err)
				}
				expandedResource, err := serde.FromMap(expandedResourceProperties, &api.Resource{})
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("unable to serialize properties to Resource spec: %w", err)
				}

				// restore 'patches' content
				expandedResource.Spec.Patches = patches

				newResource := &api.Resource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      expandedResource.Spec.Name,
						Namespace: namespace.Name,
						Labels: map[string]string{
							api.Group + "/managedBy.group":   resourceDefinition.GroupVersionKind().Group,
							api.Group + "/managedBy.version": resourceDefinition.GroupVersionKind().Version,
							api.Group + "/managedBy.kind":    resourceDefinition.GroupVersionKind().Kind,
							api.Group + "/managedBy.name":    resourceDefinition.Name,
							api.Group + "/managedBy.id":      string(resourceDefinition.UID),
						},
					},
					Spec: expandedResource.Spec,
				}
				if err := ctrl.SetControllerReference(resourceDefinition, newResource, reconciler.Scheme); err != nil {
					return ctrl.Result{}, fmt.Errorf("unable to set Resource's ownerReference: %w", err)
				}
				if err := reconciler.Create(ctx, newResource); err != nil {
					return ctrl.Result{}, fmt.Errorf("unable to create Resource %s: %w", expandedResource.Name, err)
				}
			}

			return ctrl.Result{}, nil
		}
	}

	expandedResourceProperties, err := element.Evaluate(args)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to expand resource properties: %w", err)
	}

	expandedResource, err := serde.FromMap(expandedResourceProperties, &api.Resource{})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to serialize properties to Resource spec: %w", err)
	}

	// restore 'patches' content
	expandedResource.Spec.Patches = patches

	newResource := &api.Resource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      expandedResource.Spec.Name,
			Namespace: resourceDefinition.Name,
		},
		Spec: expandedResource.Spec,
	}
	if err := ctrl.SetControllerReference(resourceDefinition, newResource, reconciler.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to set Resource's ownerReference: %w", err)
	}
	if err := reconciler.Create(ctx, newResource); err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to create Resource %s: %w", expandedResource.Name, err)
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
