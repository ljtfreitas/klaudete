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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/nubank/klaudete/api/v1alpha1"
	"github.com/nubank/klaudete/internal/provisioning"
	"github.com/nubank/klaudete/internal/serde"
)

var _ = Describe("Resource Controller", Ordered, func() {
	Context("When reconciling a resource", func() {
		resourceName := "pet-sample"

		ctx := context.Background()

		namespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		resource := &api.Resource{}

		BeforeAll(func() {
			By("Creating a ResourceType to be used", func() {
				resourceType := &api.ResourceType{
					ObjectMeta: metav1.ObjectMeta{
						Name: "random-pet",
					},
					Spec: api.ResourceTypeSpec{
						Name:        "random-pet",
						Description: "Just random pets",
					},
				}
				err := k8sClient.Create(ctx, resourceType)
				if err != nil && !errors.IsNotFound(err) {
					Fail(fmt.Sprintf("ResourceType random-pet already exists."))
				}
			})
		})

		BeforeEach(func() {
			By("creating a Resource object", func() {
				err := k8sClient.Get(ctx, namespacedName, resource)
				if err != nil && !errors.IsNotFound(err) {
					Fail(fmt.Sprintf("Resource %s already exists.", resourceName))
				}

				resourceProperties, err := serde.ToRaw(map[string]string{
					"name": "sample",
				})
				if err != nil {
					Fail(fmt.Sprintf("Failure to serialize properties map: %v", err))
				}

				configMapProvisionerSpec, err := serde.ToRaw(map[string]any{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]any{
						"name": "${resource.spec.name}-config",
					},
					"data": map[string]string{
						"name": "${resource.spec.properties.name}",
					},
				})
				if err != nil {
					Fail(fmt.Sprintf("Failure to serialize provisioner spec to map: %v", err))
				}

				anotherConfigMapProvisionerSpec, err := serde.ToRaw(map[string]any{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]any{
						"name": "${resource.spec.name}-another-config",
					},
					"data": map[string]string{
						"anotherName": "${provisioner.resources.configMap.data.name}",
					},
				})
				if err != nil {
					Fail(fmt.Sprintf("Failure to serialize properties map: %v", err))
				}

				resource := &api.Resource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      namespacedName.Name,
						Namespace: namespacedName.Namespace,
					},
					Spec: api.ResourceSpec{
						Name:            resourceName,
						Alias:           resourceName,
						Description:     "I'm just a pet",
						ResourceTypeRef: "random-pet",
						Properties:      resourceProperties,
						Connections: []api.ResourceConnection{
							api.ResourceConnection{
								Via: "belongs-to",
								Target: api.ResourceConnectionTarget{
									Ref: &api.ResourceConnectionTargetRef{
										ApiVersion: "klaudete.nubank.com.br/v1alpha1",
										Kind:       "Resource",
										Name:       "petOwner",
									},
								},
							},
						},
						Provisioner: &api.ResourceProvisioner{
							Resources: []api.ResourceProvisionerObject{
								api.ResourceProvisionerObject{
									Name:      "configMap",
									Ref:       configMapProvisionerSpec,
									ReadyWhen: ptr.To("${provisioner.resources.configMap.data != nil}"),
									Outputs:   ptr.To("${provisioner.resources.configMap.data}"),
								},
								api.ResourceProvisionerObject{
									Name:      "anotherConfigMap",
									Ref:       anotherConfigMapProvisionerSpec,
									ReadyWhen: ptr.To("${provisioner.resources.anotherConfigMap.data != nil}"),
									Outputs:   ptr.To("${provisioner.resources.anotherConfigMap.data}"),
								},
							},
						},
						Patches: []api.ResourcePatch{
							api.ResourcePatch{
								From: "${provisioner.resources.configMap.data.name}",
								To:   "status.properties.name",
							},
							api.ResourcePatch{
								From: "${provisioner.resources.anotherConfigMap.data.anotherName}",
								To:   "status.properties.anotherName",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			})
		})

		AfterEach(func() {
			resource := &api.Resource{}
			err := k8sClient.Get(ctx, namespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Resource", func() {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			})
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource", func() {
				resourceReconciler := &ResourceReconciler{
					Client:        k8sClient,
					Scheme:        k8sClient.Scheme(),
					Recorder:      record.NewFakeRecorder(1024),
					DynamicClient: dynamicK8sClient,
				}

				_, err := reconcile.AsReconciler(k8sClient, resourceReconciler).Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				resource := &api.Resource{}
				err = k8sClient.Get(ctx, namespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resourceStatus := resource.Status
				Expect(resourceStatus).ShouldNot(BeNil())
				Expect(resourceStatus.Phase).Should(Equal(api.ResourceStatusReady))

				atProvisioner := resource.Status.AtProvisioner
				Expect(atProvisioner.State).Should(Equal(string(provisioning.ProvisioningSuccessState)))

				atProvisionerOutputs, err := serde.FromBytes(atProvisioner.Outputs.Raw, &map[string]any{})
				Expect(err).NotTo(HaveOccurred())
				Expect(*atProvisionerOutputs).ShouldNot(BeEmpty())

				Expect(*atProvisionerOutputs).Should(HaveKey("name"))
				Expect(*atProvisionerOutputs).Should(HaveKey("anotherName"))

				configMap := corev1.ConfigMap{}
				err = k8sClient.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-config", resourceName), Namespace: "default"}, &configMap)
				Expect(err).NotTo(HaveOccurred())

				Expect(configMap.Data).ShouldNot(BeEmpty())
				Expect(configMap.Data).Should(HaveKeyWithValue("name", (*atProvisionerOutputs)["name"]))

				anotherConfigMap := corev1.ConfigMap{}
				err = k8sClient.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-another-config", resourceName), Namespace: "default"}, &anotherConfigMap)
				Expect(err).NotTo(HaveOccurred())

				Expect(anotherConfigMap.Data).ShouldNot(BeEmpty())
				Expect(anotherConfigMap.Data).Should(HaveKeyWithValue("anotherName", configMap.Data["name"]))

				rawStatusProperties := resourceStatus.Properties
				statusProperties, err := serde.FromBytes(rawStatusProperties.Raw, &map[string]any{})
				Expect(err).NotTo(HaveOccurred())

				Expect(*statusProperties).ShouldNot(BeEmpty())
				Expect(*statusProperties).Should(HaveKey("name"))
				Expect(*statusProperties).Should(HaveKey("anotherName"))

				Expect(*statusProperties).Should(HaveKeyWithValue("name", configMap.Data["name"]))
				Expect(*statusProperties).Should(HaveKeyWithValue("anotherName", anotherConfigMap.Data["anotherName"]))
			})
		})
	})
})
