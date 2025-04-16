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

// ResourceGroupDefinitionSpec defines the desired state of ResourceGroupDefinition
type ResourceGroupDefinitionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Generator ResourceGenerator            `json:"generator,omitempty"`
	Group     ResourceGroupDefinitionGroup `json:"group,omitempty"`
}

type ResourceGroupDefinitionGroup struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Name              string                  `json:"name,omitempty"`
	Resources         []ResourceGroupResource `json:"resources,omitempty"`
}

type ResourceGroupDefinitionStatusDescription string

const (
	ResourceGroupDefinitionStatusPending = ResourceDefinitionStatusDescription("Pending")
	ResourceGroupDefinitionStatusFailed  = ResourceDefinitionStatusDescription("Failed")
	ResourceGroupDefinitionStatusReady   = ResourceDefinitionStatusDescription("Ready")
)

// ResourceGroupDefinitionStatus defines the observed state of ResourceGroupDefinition
type ResourceGroupDefinitionStatus struct {
	Status     ResourceDefinitionStatusDescription `json:"status,omitempty"`
	Conditions []metav1.Condition                  `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"

// ResourceGroupDefinition is the Schema for the resourcegroupdefinitions API
type ResourceGroupDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceGroupDefinitionSpec   `json:"spec"`
	Status ResourceGroupDefinitionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceGroupDefinitionList contains a list of ResourceGroupDefinition
type ResourceGroupDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceGroupDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceGroupDefinition{}, &ResourceGroupDefinitionList{})
}
