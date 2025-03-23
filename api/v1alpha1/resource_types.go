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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ResourceSpec defines the desired state of Resource
type ResourceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name            string                `json:"name"`
	Alias           string                `json:"alias,omitempty"`
	Description     string                `json:"description,omitempty"`
	ResourceTypeRef string                `json:"resourceTypeRef"`
	Properties      *runtime.RawExtension `json:"properties,omitempty"`
	Provisioner     ResourceProvisioner   `json:"provisioner,omitempty"`
	Connections     ResourceConnections   `json:"connections,omitempty"`
}

type ResourcePropertyType string

const (
	ResourcePropertyStringType = ResourcePropertyType("stringValue")
)

type ResourceProperties map[string]any

type ResourceProvisioner struct {
	Ref ResourceProvisionerRef `json:"ref,omitempty"`
}

type ResourceProvisionerRef struct {
	Name string               `json:"name"`
	Spec runtime.RawExtension `json:"spec,omitempty"`
}

type ResourceConnections []ResourceConnection

type ResourceConnection struct {
	Via    string `json:"via"`
	Target string `json:"target"`
}

// ResourceStatus defines the observed state of Resource
type ResourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Status ResourceStatusDescription `json:"status"`
}

type ResourceStatusDescription string

const (
	ResourceStatusReady = ResourceStatusDescription("RESOURCE_STATUS_READY")
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Resource is the Schema for the resources API
type Resource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceSpec   `json:"spec,omitempty"`
	Status ResourceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceList contains a list of Resource
type ResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Resource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Resource{}, &ResourceList{})
}
