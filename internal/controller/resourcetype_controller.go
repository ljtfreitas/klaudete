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
	"strings"

	"github.com/gobuffalo/flect"
	kro "github.com/kro-run/kro/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	api "github.com/nubank/klaudete/api/v1alpha1"
)

// ResourceTypeReconciler reconciles a ResourceType object
type ResourceTypeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
	_ = log.FromContext(ctx)

	kroResourceGraphDefinition := kro.ResourceGraphDefinition{}
	if err := reconciler.Get(ctx, types.NamespacedName{Name: resourceType.Name}, &kroResourceGraphDefinition); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("failure to find Kro's ResourceGraphDefinition: %w", err)
		}

		resourceName := flect.Pascalize(newResourceName(resourceType.Spec.Name, resourceType.Name))
		resourceSpec, err := newResourceSpec()
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to generate a Kro custom schema: %w", err)
		}
		resources, err := newResources(resourceName)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to generate Kro resources: %w", err)
		}

		kroResourceGraphDefinition = kro.ResourceGraphDefinition{
			TypeMeta: metav1.TypeMeta{
				APIVersion: kro.GroupVersion.String(),
				Kind:       "ResourceGraphDefinition",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: resourceType.Name,
				Labels: map[string]string{
					api.Group + "/owned":             "true",
					api.Group + "/managedBy.group":   resourceType.GroupVersionKind().Group,
					api.Group + "/managedBy.version": resourceType.GroupVersionKind().Version,
					api.Group + "/managedBy.kind":    resourceType.GroupVersionKind().Kind,
					api.Group + "/managedBy.name":    resourceType.Name,
					api.Group + "/managedBy.id":      string(resourceType.UID),
				},
			},
			Spec: kro.ResourceGraphDefinitionSpec{
				Schema: &kro.Schema{
					Kind:       resourceName,
					Group:      api.GroupVersion.Group,
					APIVersion: api.GroupVersion.Version,
					Spec:       *resourceSpec,
				},
				Resources: resources,
			},
		}

		if err := ctrl.SetControllerReference(resourceType, &kroResourceGraphDefinition, reconciler.Scheme); err != nil {
			return ctrl.Result{}, fmt.Errorf("unable to set Kro's ResourceGraphDefinition ownerReference: %w", err)
		}

		if err := reconciler.Create(ctx, &kroResourceGraphDefinition); err != nil {
			return ctrl.Result{}, fmt.Errorf("failure to generate Kro's ResourceGraphDefinition: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourceTypeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.ResourceType{}).
		Complete(reconcile.AsReconciler(mgr.GetClient(), r))
}

func newResourceName(name *string, resourceTypeName string) string {
	if name != nil {
		return *name
	}
	return resourceTypeName
}

func newResourceSpec() (*runtime.RawExtension, error) {
	simpleSchema := map[string]any{
		"name":        "string | required=true",
		"description": "string | default={}",
		"alias":       "string | default={}",
		"properties":  "map[string]string | default={}",
		"provisioner": "map[string]string | default={}",
	}

	simpleSchemaAsBytes, err := json.Marshal(simpleSchema)
	if err != nil {
		return nil, err
	}

	rawExtension := &runtime.RawExtension{Raw: simpleSchemaAsBytes}

	return rawExtension, nil
}

func newResources(resourceTypeName string) ([]*kro.Resource, error) {
	resource := newResource(resourceTypeName)

	resourceAsBytes, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}

	kroResource := &kro.Resource{
		ID:          strings.ToLower(resourceTypeName),
		Template:    runtime.RawExtension{Raw: resourceAsBytes},
		ReadyWhen:   []string{},
		IncludeWhen: []string{},
	}
	return []*kro.Resource{kroResource}, nil
}

func newResource(resourceTypeName string) map[string]any {
	return map[string]any{
		"apiVersion": api.GroupVersion.String(),
		"kind":       "Resource",
		"metadata": map[string]any{
			"name": "${schema.spec.name}",
		},
		"spec": map[string]any{
			"name":            "${schema.spec.name}",
			"alias":           "${schema.spec.alias}",
			"description":     "${schema.spec.description}",
			"resourceTypeRef": resourceTypeName,
			"properties":      "${schema.spec.properties}",
			"provisioner":     "${schema.spec.provisioner}",
		},
	}
}
