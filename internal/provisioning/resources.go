package provisioning

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"github.com/gobuffalo/flect"
	api "github.com/nubank/klaudete/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type ProvisioningStateDescription string

const (
	ProvisioningRunningState = ProvisioningStateDescription("Running")
	ProvisioningFailedState  = ProvisioningStateDescription("Failed")
	ProvisioningSuccessState = ProvisioningStateDescription("Success")
)

type ProvisioningStatus struct {
	State     ProvisioningStateDescription
	Resources []*ManagedResource
}

func (p *ProvisioningStatus) IsRunning() bool {
	return p.State == ProvisioningRunningState
}

type ManagedResource struct {
	schema.GroupVersionKind
	*unstructured.Unstructured

	Name    string
	Outputs map[string]any
}

type ManagedResourceProvisioner struct {
	client.Client

	dynamicClient *dynamic.DynamicClient
	log           logr.Logger
	scheme        *runtime.Scheme
}

func NewManagedResourceProvisioner(client client.Client, dynamicClient *dynamic.DynamicClient, scheme *runtime.Scheme) *ManagedResourceProvisioner {
	return &ManagedResourceProvisioner{
		Client:        client,
		dynamicClient: dynamicClient,
		scheme:        scheme,
	}
}

func (provisioner *ManagedResourceProvisioner) Apply(ctx context.Context, resource *api.Resource, resourceToBeProvisioned *api.ResourceProvisionerObject, obj *unstructured.Unstructured) (*ManagedResource, error) {
	gvk := obj.GroupVersionKind()

	kindAsPlural := strings.ToLower(flect.Pluralize(gvk.Kind))

	objGvWithResource := gvk.GroupVersion().WithResource(kindAsPlural)

	name := obj.GetName()
	if name == "" {
		name = resource.Name
	}
	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = resource.Namespace
	}

	dynamicResourceClient := provisioner.dynamicClient.
		Resource(objGvWithResource).
		Namespace(namespace)

	o, err := dynamicResourceClient.Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}

		resourceGvk, err := apiutil.GVKForObject(resource, provisioner.scheme)
		if err != nil {
			return nil, err
		}

		obj.SetName(name)
		obj.SetNamespace(namespace)

		labels := obj.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[api.Group+"/managedBy.group"] = resourceGvk.Group
		labels[api.Group+"/managedBy.version"] = resourceGvk.Version
		labels[api.Group+"/managedBy.kind"] = resourceGvk.Kind
		labels[api.Group+"/managedBy.name"] = resource.Name
		labels[api.Group+"/managedBy.id"] = string(resource.UID)
		obj.SetLabels(labels)

		obj.SetOwnerReferences([]metav1.OwnerReference{
			{
				APIVersion:         resourceGvk.GroupVersion().String(),
				Kind:               resourceGvk.Kind,
				Name:               resource.Name,
				UID:                resource.UID,
				BlockOwnerDeletion: ptr.To(true),
				Controller:         ptr.To(true),
			},
		})

		obj, err = dynamicResourceClient.Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	} else {
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			o, err = dynamicResourceClient.Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if newSpec, exists, err := unstructured.NestedMap(obj.Object, "spec"); exists && err == nil {
				if spec, exists, err := unstructured.NestedMap(o.Object, "spec"); exists && err == nil {
					for name, _ := range spec {
						if _, ok := newSpec[name]; !ok {
							delete(spec, name)
						}
					}
					for name, newValue := range newSpec {
						spec[name] = newValue
					}
					err = unstructured.SetNestedMap(o.Object, spec, "spec")
					o, err = dynamicResourceClient.Update(ctx, o, metav1.UpdateOptions{})
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// time.Sleep(time.Second * 5)

	// refresh
	obj, err = dynamicResourceClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	managedResource := &ManagedResource{
		Name:             resourceToBeProvisioned.Name,
		GroupVersionKind: obj.GroupVersionKind(),
		Unstructured:     obj,
	}

	return managedResource, nil
}
