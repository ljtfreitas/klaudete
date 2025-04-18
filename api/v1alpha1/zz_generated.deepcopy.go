//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Metadata) DeepCopyInto(out *Metadata) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Metadata.
func (in *Metadata) DeepCopy() *Metadata {
	if in == nil {
		return nil
	}
	out := new(Metadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Resource) DeepCopyInto(out *Resource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Resource.
func (in *Resource) DeepCopy() *Resource {
	if in == nil {
		return nil
	}
	out := new(Resource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Resource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceConnection) DeepCopyInto(out *ResourceConnection) {
	*out = *in
	in.Target.DeepCopyInto(&out.Target)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceConnection.
func (in *ResourceConnection) DeepCopy() *ResourceConnection {
	if in == nil {
		return nil
	}
	out := new(ResourceConnection)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceConnectionTarget) DeepCopyInto(out *ResourceConnectionTarget) {
	*out = *in
	if in.Ref != nil {
		in, out := &in.Ref, &out.Ref
		*out = new(ResourceConnectionTargetRef)
		**out = **in
	}
	if in.Nurn != nil {
		in, out := &in.Nurn, &out.Nurn
		*out = new(ResourceConnectionTargetNurn)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceConnectionTarget.
func (in *ResourceConnectionTarget) DeepCopy() *ResourceConnectionTarget {
	if in == nil {
		return nil
	}
	out := new(ResourceConnectionTarget)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceConnectionTargetNurn) DeepCopyInto(out *ResourceConnectionTargetNurn) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceConnectionTargetNurn.
func (in *ResourceConnectionTargetNurn) DeepCopy() *ResourceConnectionTargetNurn {
	if in == nil {
		return nil
	}
	out := new(ResourceConnectionTargetNurn)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceConnectionTargetRef) DeepCopyInto(out *ResourceConnectionTargetRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceConnectionTargetRef.
func (in *ResourceConnectionTargetRef) DeepCopy() *ResourceConnectionTargetRef {
	if in == nil {
		return nil
	}
	out := new(ResourceConnectionTargetRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in ResourceConnections) DeepCopyInto(out *ResourceConnections) {
	{
		in := &in
		*out = make(ResourceConnections, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceConnections.
func (in ResourceConnections) DeepCopy() ResourceConnections {
	if in == nil {
		return nil
	}
	out := new(ResourceConnections)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceDefinition) DeepCopyInto(out *ResourceDefinition) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceDefinition.
func (in *ResourceDefinition) DeepCopy() *ResourceDefinition {
	if in == nil {
		return nil
	}
	out := new(ResourceDefinition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceDefinition) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceDefinitionList) DeepCopyInto(out *ResourceDefinitionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ResourceDefinition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceDefinitionList.
func (in *ResourceDefinitionList) DeepCopy() *ResourceDefinitionList {
	if in == nil {
		return nil
	}
	out := new(ResourceDefinitionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceDefinitionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceDefinitionResource) DeepCopyInto(out *ResourceDefinitionResource) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceDefinitionResource.
func (in *ResourceDefinitionResource) DeepCopy() *ResourceDefinitionResource {
	if in == nil {
		return nil
	}
	out := new(ResourceDefinitionResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceDefinitionResourceStatus) DeepCopyInto(out *ResourceDefinitionResourceStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceDefinitionResourceStatus.
func (in *ResourceDefinitionResourceStatus) DeepCopy() *ResourceDefinitionResourceStatus {
	if in == nil {
		return nil
	}
	out := new(ResourceDefinitionResourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in ResourceDefinitionResourcesStatus) DeepCopyInto(out *ResourceDefinitionResourcesStatus) {
	{
		in := &in
		*out = make(ResourceDefinitionResourcesStatus, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceDefinitionResourcesStatus.
func (in ResourceDefinitionResourcesStatus) DeepCopy() ResourceDefinitionResourcesStatus {
	if in == nil {
		return nil
	}
	out := new(ResourceDefinitionResourcesStatus)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceDefinitionSpec) DeepCopyInto(out *ResourceDefinitionSpec) {
	*out = *in
	if in.Generator != nil {
		in, out := &in.Generator, &out.Generator
		*out = make(ResourceGenerator, len(*in))
		for key, val := range *in {
			var outVal *runtime.RawExtension
			if val == nil {
				(*out)[key] = nil
			} else {
				inVal := (*in)[key]
				in, out := &inVal, &outVal
				*out = new(runtime.RawExtension)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	in.Resource.DeepCopyInto(&out.Resource)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceDefinitionSpec.
func (in *ResourceDefinitionSpec) DeepCopy() *ResourceDefinitionSpec {
	if in == nil {
		return nil
	}
	out := new(ResourceDefinitionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceDefinitionStatus) DeepCopyInto(out *ResourceDefinitionStatus) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make(ResourceDefinitionResourcesStatus, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceDefinitionStatus.
func (in *ResourceDefinitionStatus) DeepCopy() *ResourceDefinitionStatus {
	if in == nil {
		return nil
	}
	out := new(ResourceDefinitionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in ResourceGenerator) DeepCopyInto(out *ResourceGenerator) {
	{
		in := &in
		*out = make(ResourceGenerator, len(*in))
		for key, val := range *in {
			var outVal *runtime.RawExtension
			if val == nil {
				(*out)[key] = nil
			} else {
				inVal := (*in)[key]
				in, out := &inVal, &outVal
				*out = new(runtime.RawExtension)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGenerator.
func (in ResourceGenerator) DeepCopy() ResourceGenerator {
	if in == nil {
		return nil
	}
	out := new(ResourceGenerator)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceGroup) DeepCopyInto(out *ResourceGroup) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGroup.
func (in *ResourceGroup) DeepCopy() *ResourceGroup {
	if in == nil {
		return nil
	}
	out := new(ResourceGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceGroup) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceGroupDefinition) DeepCopyInto(out *ResourceGroupDefinition) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGroupDefinition.
func (in *ResourceGroupDefinition) DeepCopy() *ResourceGroupDefinition {
	if in == nil {
		return nil
	}
	out := new(ResourceGroupDefinition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceGroupDefinition) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceGroupDefinitionGroup) DeepCopyInto(out *ResourceGroupDefinitionGroup) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]ResourceGroupResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGroupDefinitionGroup.
func (in *ResourceGroupDefinitionGroup) DeepCopy() *ResourceGroupDefinitionGroup {
	if in == nil {
		return nil
	}
	out := new(ResourceGroupDefinitionGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceGroupDefinitionList) DeepCopyInto(out *ResourceGroupDefinitionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ResourceGroupDefinition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGroupDefinitionList.
func (in *ResourceGroupDefinitionList) DeepCopy() *ResourceGroupDefinitionList {
	if in == nil {
		return nil
	}
	out := new(ResourceGroupDefinitionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceGroupDefinitionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceGroupDefinitionSpec) DeepCopyInto(out *ResourceGroupDefinitionSpec) {
	*out = *in
	if in.Generator != nil {
		in, out := &in.Generator, &out.Generator
		*out = make(ResourceGenerator, len(*in))
		for key, val := range *in {
			var outVal *runtime.RawExtension
			if val == nil {
				(*out)[key] = nil
			} else {
				inVal := (*in)[key]
				in, out := &inVal, &outVal
				*out = new(runtime.RawExtension)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	in.Group.DeepCopyInto(&out.Group)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGroupDefinitionSpec.
func (in *ResourceGroupDefinitionSpec) DeepCopy() *ResourceGroupDefinitionSpec {
	if in == nil {
		return nil
	}
	out := new(ResourceGroupDefinitionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceGroupDefinitionStatus) DeepCopyInto(out *ResourceGroupDefinitionStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGroupDefinitionStatus.
func (in *ResourceGroupDefinitionStatus) DeepCopy() *ResourceGroupDefinitionStatus {
	if in == nil {
		return nil
	}
	out := new(ResourceGroupDefinitionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceGroupList) DeepCopyInto(out *ResourceGroupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ResourceGroup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGroupList.
func (in *ResourceGroupList) DeepCopy() *ResourceGroupList {
	if in == nil {
		return nil
	}
	out := new(ResourceGroupList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceGroupList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceGroupResource) DeepCopyInto(out *ResourceGroupResource) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGroupResource.
func (in *ResourceGroupResource) DeepCopy() *ResourceGroupResource {
	if in == nil {
		return nil
	}
	out := new(ResourceGroupResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceGroupSpec) DeepCopyInto(out *ResourceGroupSpec) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]ResourceGroupResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGroupSpec.
func (in *ResourceGroupSpec) DeepCopy() *ResourceGroupSpec {
	if in == nil {
		return nil
	}
	out := new(ResourceGroupSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceGroupStatus) DeepCopyInto(out *ResourceGroupStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGroupStatus.
func (in *ResourceGroupStatus) DeepCopy() *ResourceGroupStatus {
	if in == nil {
		return nil
	}
	out := new(ResourceGroupStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceList) DeepCopyInto(out *ResourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Resource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceList.
func (in *ResourceList) DeepCopy() *ResourceList {
	if in == nil {
		return nil
	}
	out := new(ResourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourcePatch) DeepCopyInto(out *ResourcePatch) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourcePatch.
func (in *ResourcePatch) DeepCopy() *ResourcePatch {
	if in == nil {
		return nil
	}
	out := new(ResourcePatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in ResourcePatches) DeepCopyInto(out *ResourcePatches) {
	{
		in := &in
		*out = make(ResourcePatches, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourcePatches.
func (in ResourcePatches) DeepCopy() ResourcePatches {
	if in == nil {
		return nil
	}
	out := new(ResourcePatches)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceProvisioner) DeepCopyInto(out *ResourceProvisioner) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make(ResourceProvisionerObjects, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceProvisioner.
func (in *ResourceProvisioner) DeepCopy() *ResourceProvisioner {
	if in == nil {
		return nil
	}
	out := new(ResourceProvisioner)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceProvisionerObject) DeepCopyInto(out *ResourceProvisionerObject) {
	*out = *in
	if in.Ref != nil {
		in, out := &in.Ref, &out.Ref
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
	if in.ReadyWhen != nil {
		in, out := &in.ReadyWhen, &out.ReadyWhen
		*out = new(string)
		**out = **in
	}
	if in.FailedWhen != nil {
		in, out := &in.FailedWhen, &out.FailedWhen
		*out = new(string)
		**out = **in
	}
	if in.Outputs != nil {
		in, out := &in.Outputs, &out.Outputs
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceProvisionerObject.
func (in *ResourceProvisionerObject) DeepCopy() *ResourceProvisionerObject {
	if in == nil {
		return nil
	}
	out := new(ResourceProvisionerObject)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in ResourceProvisionerObjects) DeepCopyInto(out *ResourceProvisionerObjects) {
	{
		in := &in
		*out = make(ResourceProvisionerObjects, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceProvisionerObjects.
func (in ResourceProvisionerObjects) DeepCopy() ResourceProvisionerObjects {
	if in == nil {
		return nil
	}
	out := new(ResourceProvisionerObjects)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceSpec) DeepCopyInto(out *ResourceSpec) {
	*out = *in
	if in.Properties != nil {
		in, out := &in.Properties, &out.Properties
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
	if in.Connections != nil {
		in, out := &in.Connections, &out.Connections
		*out = make(ResourceConnections, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Provisioner != nil {
		in, out := &in.Provisioner, &out.Provisioner
		*out = new(ResourceProvisioner)
		(*in).DeepCopyInto(*out)
	}
	if in.Patches != nil {
		in, out := &in.Patches, &out.Patches
		*out = make(ResourcePatches, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceSpec.
func (in *ResourceSpec) DeepCopy() *ResourceSpec {
	if in == nil {
		return nil
	}
	out := new(ResourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceStatus) DeepCopyInto(out *ResourceStatus) {
	*out = *in
	in.AtProvisioner.DeepCopyInto(&out.AtProvisioner)
	if in.Inventory != nil {
		in, out := &in.Inventory, &out.Inventory
		*out = new(ResourceStatusInventory)
		(*in).DeepCopyInto(*out)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceStatus.
func (in *ResourceStatus) DeepCopy() *ResourceStatus {
	if in == nil {
		return nil
	}
	out := new(ResourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceStatusInventory) DeepCopyInto(out *ResourceStatusInventory) {
	*out = *in
	if in.Properties != nil {
		in, out := &in.Properties, &out.Properties
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceStatusInventory.
func (in *ResourceStatusInventory) DeepCopy() *ResourceStatusInventory {
	if in == nil {
		return nil
	}
	out := new(ResourceStatusInventory)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceStatusProvisioner) DeepCopyInto(out *ResourceStatusProvisioner) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]ResourceStatusProvisionerObject, len(*in))
		copy(*out, *in)
	}
	if in.Outputs != nil {
		in, out := &in.Outputs, &out.Outputs
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceStatusProvisioner.
func (in *ResourceStatusProvisioner) DeepCopy() *ResourceStatusProvisioner {
	if in == nil {
		return nil
	}
	out := new(ResourceStatusProvisioner)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceStatusProvisionerObject) DeepCopyInto(out *ResourceStatusProvisionerObject) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceStatusProvisionerObject.
func (in *ResourceStatusProvisionerObject) DeepCopy() *ResourceStatusProvisionerObject {
	if in == nil {
		return nil
	}
	out := new(ResourceStatusProvisionerObject)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceType) DeepCopyInto(out *ResourceType) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceType.
func (in *ResourceType) DeepCopy() *ResourceType {
	if in == nil {
		return nil
	}
	out := new(ResourceType)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceType) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceTypeList) DeepCopyInto(out *ResourceTypeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ResourceType, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceTypeList.
func (in *ResourceTypeList) DeepCopy() *ResourceTypeList {
	if in == nil {
		return nil
	}
	out := new(ResourceTypeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceTypeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceTypeSpec) DeepCopyInto(out *ResourceTypeSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceTypeSpec.
func (in *ResourceTypeSpec) DeepCopy() *ResourceTypeSpec {
	if in == nil {
		return nil
	}
	out := new(ResourceTypeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceTypeStatus) DeepCopyInto(out *ResourceTypeStatus) {
	*out = *in
	out.Inventory = in.Inventory
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceTypeStatus.
func (in *ResourceTypeStatus) DeepCopy() *ResourceTypeStatus {
	if in == nil {
		return nil
	}
	out := new(ResourceTypeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceTypeStatusInventory) DeepCopyInto(out *ResourceTypeStatusInventory) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceTypeStatusInventory.
func (in *ResourceTypeStatusInventory) DeepCopy() *ResourceTypeStatusInventory {
	if in == nil {
		return nil
	}
	out := new(ResourceTypeStatusInventory)
	in.DeepCopyInto(out)
	return out
}
