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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ResourceDefinitionSpec defines the desired state of ResourceDefinition
type ResourceDefinitionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Generator ResourceGenerator          `json:"generator,omitempty"`
	Resource  ResourceDefinitionResource `json:"resource,omitempty"`
}

type ResourceDefinitionResource struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ResourceSpec `json:"spec"`
}

type Metadata struct {
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

// ResourceDefinitionStatus defines the observed state of ResourceDefinition
type ResourceDefinitionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Status     ResourceDefinitionStatusDescription `json:"status,omitempty"`
	Resources  ResourceDefinitionResourcesStatus   `json:"resources,omitempty"`
	Conditions []metav1.Condition                  `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

type ResourceDefinitionResourcesStatus []ResourceDefinitionResourceStatus

type ResourceDefinitionResourceStatus struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	UId        string `json:"uid"`
}

type ResourceDefinitionStatusDescription string

const (
	ResourceDefinitionStatusPending = ResourceDefinitionStatusDescription("Pending")
	ResourceDefinitionStatusFailed  = ResourceDefinitionStatusDescription("Failed")
	ResourceDefinitionStatusReady   = ResourceDefinitionStatusDescription("Ready")
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.status`

// ResourceDefinition is the Schema for the resourcedefinitions API
type ResourceDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceDefinitionSpec   `json:"spec,omitempty"`
	Status ResourceDefinitionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceDefinitionList contains a list of ResourceDefinition
type ResourceDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceDefinition{}, &ResourceDefinitionList{})
}
