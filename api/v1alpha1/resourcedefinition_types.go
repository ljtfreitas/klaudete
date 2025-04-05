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
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ResourceDefinitionSpec defines the desired state of ResourceDefinition
type ResourceDefinitionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Generator ResourceDefinitionGenerator `json:"generator,omitempty"`
	Resource  Resource                    `json:"resource,omitempty"`
}

type ResourceDefinitionGenerator map[string]*runtime.RawExtension

func (generator ResourceDefinitionGenerator) Spec() (string, *runtime.RawExtension, error) {
	if generator == nil || len(generator) == 0 {
		return "", nil, nil
	}
	if len(generator) > 1 {
		return "", nil, errors.New("just one generator is allowed.")
	}
	for generatorType, generatorSpec := range generator {
		return generatorType, generatorSpec, nil
	}
	return "", nil, nil
}

// ResourceDefinitionStatus defines the observed state of ResourceDefinition
type ResourceDefinitionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Status     ResourceDefinitionStatusDescription `json:"status,omitempty"`
	Conditions []metav1.Condition                  `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

type ResourceDefinitionStatusDescription string

const (
	ResourceDefinitionStatusPending                = ResourceDefinitionStatusDescription("Pending")
	ResourceDefinitionStatusFailed                 = ResourceDefinitionStatusDescription("Failed")
	ResourceDefinitionStatusProvisioningInProgress = ResourceDefinitionStatusDescription("ProvisioningInProgress")
	ResourceDefinitionStatusReady                  = ResourceDefinitionStatusDescription("Ready")
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

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
