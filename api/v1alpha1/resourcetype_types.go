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

// ResourceTypeSpec defines the desired state of ResourceType
type ResourceTypeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type ResourceTypeStatusDescription string

const (
	ResourceTypeStatusPending = ResourceTypeStatusDescription("Pending")
	ResourceTypeStatusInSync  = ResourceTypeStatusDescription("InSync")
)

// ResourceTypeStatus defines the observed state of ResourceType
type ResourceTypeStatus struct {
	Status     ResourceTypeStatusDescription `json:"status"`
	Inventory  ResourceTypeStatusInventory   `json:"inventory,omitempty"`
	Conditions []metav1.Condition            `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

type ResourceTypeStatusInventory struct {
	Id string `json:"id"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Description",type="string",JSONPath=`.spec.description`
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Inventory Id",type="string",JSONPath=`.status.inventory.id`

// ResourceType is the Schema for the resourcetypes API
type ResourceType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceTypeSpec   `json:"spec,omitempty"`
	Status ResourceTypeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceTypeList contains a list of ResourceType
type ResourceTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceType `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceType{}, &ResourceTypeList{})
}
