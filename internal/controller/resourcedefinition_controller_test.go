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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	api "github.com/nubank/klaudete/api/v1alpha1"
	"github.com/nubank/klaudete/internal/generators"
	"github.com/nubank/klaudete/internal/serde"
)

var _ = Describe("ResourceDefinition Controller", func() {

	Describe("When reconciling a ResourceDefinition", func() {
		ctx := context.Background()

		resourceDefinitionName := "just-a-resource-definition"

		resourceDefinitionClusterName := types.NamespacedName{
			Name:      resourceDefinitionName,
			Namespace: "",
		}

		When("using a List generator", func() {
			listGeneratorSpec, err := generators.NewListGeneratorSpec("parameter", "one", "two")
			if err != nil {
				Fail(fmt.Sprintf("Failure to generate a list generator spec: %v", err))
			}

			BeforeEach(func() {
				resourceDefinition := &api.ResourceDefinition{}

				err := k8sClient.Get(ctx, resourceDefinitionClusterName, resourceDefinition)
				if err != nil && !errors.IsNotFound(err) {
					Fail(fmt.Sprintf("ResourceDefinition %s already exists.", resourceDefinitionName))
				}

				properties, err := serde.ToRaw(map[string]string{
					"name": "",
				})
				if err != nil {
					Fail(fmt.Sprintf("Failure to serialize properties map: %v", err))
				}

				resourceDefinition = &api.ResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: resourceDefinitionName,
					},
					Spec: api.ResourceDefinitionSpec{
						Generator: map[string]*runtime.RawExtension{
							string(generators.ListGeneratorType): listGeneratorSpec,
						},
						Resource: api.ResourceSpec{
							Name:            "just-a-pet-called-${generator.parameter}",
							Alias:           "just-a-pet-called-${generator.parameter}",
							Description:     "I'm just a pet, and my name is ${generator.parameter}",
							ResourceTypeRef: "random-pet",
							Properties:      properties,
							Connections: []api.ResourceConnection{
								api.ResourceConnection{
									Via: "belongs-to",
									Target: api.ResourceConnectionTarget{
										Ref: &api.ResourceConnectionTargetRef{
											ApiVersion: "klaudete.nubank.com.br/v1alpha1",
											Kind:       "Resource",
											Name:       "pet-owner",
										},
									},
								},
							},
							Provisioner: &api.ResourceProvisioner{
								Name: "pulumi",
								Ref: &api.ResourceProvisionerRef{
									ApiVersion: "klaudete.nubank.com.br/v1alpha1",
									Kind:       "PulumiProvisioner",
								},
							},
							Patches: api.ResourcePatches{
								api.ResourcePatch{
									From: "${provisioner.pulumi.status.outputs.name}",
									To:   "spec.properties.name",
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resourceDefinition)).To(Succeed())
			})

			It("should successfully reconcile the resource.", func() {
				resourceDefinitionReconciler := &ResourceDefinitionReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := reconcile.AsReconciler(k8sClient, resourceDefinitionReconciler).Reconcile(ctx, reconcile.Request{
					NamespacedName: resourceDefinitionClusterName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("we are expecting a dedicated namespace to ResourceDefinition objects", func() {
					namespace := &corev1.Namespace{}
					err := k8sClient.Get(ctx, types.NamespacedName{Name: resourceDefinitionName}, namespace)
					Expect(err).NotTo(HaveOccurred())
				})

				By("we are expecting to create one Resource for each generator element", func() {
					listGeneratorSpec, err := generators.UnmarshallSpec(listGeneratorSpec, &generators.ListGeneratorSpec{})
					Expect(err).NotTo(HaveOccurred())

					for _, value := range listGeneratorSpec.Values {
						namespacedName := types.NamespacedName{
							Name:      fmt.Sprintf("just-a-pet-called-%s", value),
							Namespace: resourceDefinitionName,
						}

						resource := &api.Resource{}
						err := k8sClient.Get(ctx, namespacedName, resource)
						Expect(err).NotTo(HaveOccurred())
					}
				})
			})

			AfterEach(func() {
				By("Cleanup the specific resource instance ResourceDefinition", func() {
					resourceDefinition := &api.ResourceDefinition{}
					err := k8sClient.Get(ctx, resourceDefinitionClusterName, resourceDefinition)
					Expect(err).NotTo(HaveOccurred())

					Expect(k8sClient.Delete(ctx, resourceDefinition)).To(Succeed())
				})

				By("Cleanup all generated Resources", func() {
					resourceList := &api.ResourceList{}
					err := k8sClient.List(ctx, resourceList, client.InNamespace(resourceDefinitionName))
					Expect(err).NotTo(HaveOccurred())

					for _, item := range resourceList.Items {
						err = k8sClient.Delete(ctx, &item)
						Expect(err).NotTo(HaveOccurred())
					}
				})
			})
		})

		When("using a Data generator", func() {
			dataGeneratorSpec, err := generators.NewDataGeneratorSpec("parameters",
				map[string]any{
					"value": "one",
				},
				map[string]any{
					"value": "two",
				},
			)
			if err != nil {
				Fail(fmt.Sprintf("Failure to generate a data generator spec: %v", err))
			}

			BeforeEach(func() {
				resourceDefinition := &api.ResourceDefinition{}

				err := k8sClient.Get(ctx, resourceDefinitionClusterName, resourceDefinition)
				if err != nil && !errors.IsNotFound(err) {
					Fail(fmt.Sprintf("ResourceDefinition %s already exists.", resourceDefinitionName))
				}

				properties, err := serde.ToRaw(map[string]string{
					"name": "",
				})
				if err != nil {
					Fail(fmt.Sprintf("Failure to serialize properties map: %v", err))
				}

				resourceDefinition = &api.ResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: resourceDefinitionName,
					},
					Spec: api.ResourceDefinitionSpec{
						Generator: map[string]*runtime.RawExtension{
							string(generators.DataGeneratorType): dataGeneratorSpec,
						},
						Resource: api.ResourceSpec{
							Name:            "just-a-pet-called-${generator.parameters.value}",
							Alias:           "just-a-pet-called-${generator.parameters.value}",
							Description:     "I'm just a pet, and my name is ${generator.parameters.value}",
							ResourceTypeRef: "random-pet",
							Properties:      properties,
							Connections: []api.ResourceConnection{
								api.ResourceConnection{
									Via: "belongs-to",
									Target: api.ResourceConnectionTarget{
										Ref: &api.ResourceConnectionTargetRef{
											ApiVersion: "klaudete.nubank.com.br/v1alpha1",
											Kind:       "Resource",
											Name:       "pet-owner",
										},
									},
								},
							},
							Provisioner: &api.ResourceProvisioner{
								Name: "pulumi",
								Ref: &api.ResourceProvisionerRef{
									ApiVersion: "klaudete.nubank.com.br/v1alpha1",
									Kind:       "PulumiProvisioner",
								},
							},
							Patches: api.ResourcePatches{
								api.ResourcePatch{
									From: "${provisioner.pulumi.status.outputs.name}",
									To:   "spec.properties.name",
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resourceDefinition)).To(Succeed())
			})

			It("should successfully reconcile the resource.", func() {
				resourceDefinitionReconciler := &ResourceDefinitionReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := reconcile.AsReconciler(k8sClient, resourceDefinitionReconciler).Reconcile(ctx, reconcile.Request{
					NamespacedName: resourceDefinitionClusterName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("we are expecting a dedicated namespace to ResourceDefinition objects", func() {
					namespace := &corev1.Namespace{}
					err := k8sClient.Get(ctx, types.NamespacedName{Name: resourceDefinitionName}, namespace)
					Expect(err).NotTo(HaveOccurred())
				})

				By("we are expecting to create one Resource for each generator element", func() {
					dataGeneratorSpec, err := generators.UnmarshallSpec(dataGeneratorSpec, &generators.DataGeneratorSpec{})
					Expect(err).NotTo(HaveOccurred())

					for _, obj := range dataGeneratorSpec.Values {
						namespacedName := types.NamespacedName{
							Name:      fmt.Sprintf("just-a-pet-called-%s", obj["value"]),
							Namespace: resourceDefinitionName,
						}

						resource := &api.Resource{}
						err := k8sClient.Get(ctx, namespacedName, resource)
						Expect(err).NotTo(HaveOccurred())
					}
				})
			})

			AfterEach(func() {
				By("Cleanup the specific resource instance ResourceDefinition", func() {
					resourceDefinition := &api.ResourceDefinition{}
					err := k8sClient.Get(ctx, resourceDefinitionClusterName, resourceDefinition)
					Expect(err).NotTo(HaveOccurred())

					Expect(k8sClient.Delete(ctx, resourceDefinition)).To(Succeed())
				})

				By("Cleanup all generated Resources", func() {
					resourceList := &api.ResourceList{}
					err := k8sClient.List(ctx, resourceList, client.InNamespace(resourceDefinitionName))
					Expect(err).NotTo(HaveOccurred())

					for _, item := range resourceList.Items {
						err = k8sClient.Delete(ctx, &item)
						Expect(err).NotTo(HaveOccurred())
					}
				})
			})
		})
	})
})
