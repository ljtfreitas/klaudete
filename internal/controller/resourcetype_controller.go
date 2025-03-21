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

	"github.com/gobuffalo/flect"
	kro "github.com/kro-run/kro/api/v1alpha1"
	kroSchema "github.com/kro-run/kro/pkg/simpleschema"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
func (r *ResourceTypeReconciler) Reconcile(ctx context.Context, resourceType *api.ResourceType) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	kroResourceGraphDefinition := kro.ResourceGraphDefinition{}
	if err := r.Get(ctx, types.NamespacedName{Name: resourceType.Name}, &kroResourceGraphDefinition); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("failure to find Kro's ResourceGraphDefinition from ResourceType %s: %w", resourceType.Name, err)
		}

		resourceName := flect.Pascalize(newResourceName(resourceType.Spec.Name, resourceType.Name))
		resourceSpec, err := newResourceSpec()
		if err != nil {
			fmt.Errorf("failure to generate a custom schema from ResourceType %s: %w", resourceType.Name, err)
		}
		resources := newResources()

		kroResourceGraphDefinition = kro.ResourceGraphDefinition{
			Spec: kro.ResourceGraphDefinitionSpec{
				Schema: &kro.Schema{
					Kind:       resourceName,
					APIVersion: resourceType.APIVersion,
					Spec:       *resourceSpec,
				},
				Resources: resources,
			},
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
	jsonSchemaProps := &extv1.JSONSchemaProps{
		Type: "object",
		Required: []string{
			"name",
		},
		Properties: map[string]extv1.JSONSchemaProps{
			"name": {
				Type: "string",
			},
			"description": {
				Type: "string",
			},
		},
	}

	simpleSchema, err := kroSchema.FromOpenAPISpec(jsonSchemaProps)
	if err != nil {
		return nil, err
	}

	simpleSchemaAsBytes, err := json.Marshal(simpleSchema)
	if err != nil {
		return nil, err
	}

	rawExtension := &runtime.RawExtension{Raw: simpleSchemaAsBytes}

	return rawExtension, nil
}

func newResources() []*kro.Resource {
	return make([]*kro.Resource, 0)
}
