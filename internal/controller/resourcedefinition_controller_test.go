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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	api "github.com/nubank/klaudete/api/v1alpha1"
	"github.com/nubank/klaudete/internal/generators"
	"github.com/nubank/klaudete/internal/serde"
)

var _ = Describe("ResourceDefinition Controller", Ordered, func() {

	Describe("When reconciling a ResourceDefinition", func() {
		ctx := context.Background()

		resourceDefinitionName := "just-a-resource-definition"

		resourceDefinitionNamespacedName := types.NamespacedName{
			Name:      resourceDefinitionName,
			Namespace: "default",
		}

		When("...using a List generator", func() {
			listGeneratorSpec, err := generators.NewListGeneratorSpec("parameter", "one", "two")
			if err != nil {
				Fail(fmt.Sprintf("Failure to generate a list generator spec: %v", err))
			}

			configMapProvisionerSpec, err := serde.ToRaw(map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name": "${resource.spec.name}-config",
				},
				"data": map[string]string{
					"name":              "${resource.spec.properties.name}",
					"belongsToResource": "${resource.metadata.name}-dev-${generator.parameter}",
				},
			})
			if err != nil {
				Fail(fmt.Sprintf("Failure to serialize provisioner spec to map: %v", err))
			}

			BeforeEach(func() {
				resourceDefinition := &api.ResourceDefinition{}

				err := k8sClient.Get(ctx, resourceDefinitionNamespacedName, resourceDefinition)
				if err != nil && !errors.IsNotFound(err) {
					Fail(fmt.Sprintf("ResourceDefinition %s already exists.", resourceDefinitionName))
				}

				properties, err := serde.ToRaw(map[string]string{
					"name": "just-a-name",
				})
				if err != nil {
					Fail(fmt.Sprintf("Failure to serialize properties map: %v", err))
				}

				resourceDefinition = &api.ResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceDefinitionNamespacedName.Name,
						Namespace: resourceDefinitionNamespacedName.Namespace,
					},
					Spec: api.ResourceDefinitionSpec{
						Generator: map[string]*runtime.RawExtension{
							string(generators.ListGeneratorType): listGeneratorSpec,
						},
						Resource: api.ResourceDefinitionResource{
							ObjectMeta: metav1.ObjectMeta{
								Name: "just-a-simple-pet-called-${generator.parameter}",
							},
							Spec: api.ResourceSpec{
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
												Name: "pet-owner",
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
									},
								},
								Patches: api.ResourcePatches{
									api.ResourcePatch{
										From: "${provisioner.resources.configMap.data.name}",
										To:   "status.properties.name",
									},
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resourceDefinition)).To(Succeed())
			})

			It("...should successfully reconcile the resource.", func() {
				resourceDefinitionReconciler := &ResourceDefinitionReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := reconcile.AsReconciler(k8sClient, resourceDefinitionReconciler).Reconcile(ctx, reconcile.Request{
					NamespacedName: resourceDefinitionNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("we are expecting to create one Resource for each generator element", func() {
					listGeneratorSpec, err := generators.UnmarshallSpec(listGeneratorSpec, &generators.ListGeneratorSpec{})
					Expect(err).NotTo(HaveOccurred())

					for _, value := range listGeneratorSpec.Values {
						namespacedName := types.NamespacedName{
							Name:      fmt.Sprintf("just-a-simple-pet-called-%s", value),
							Namespace: resourceDefinitionNamespacedName.Namespace,
						}

						resource := &api.Resource{}
						err := k8sClient.Get(ctx, namespacedName, resource)
						Expect(err).NotTo(HaveOccurred())

						for _, pr := range resource.Spec.Provisioner.Resources {
							m := make(map[string]any)
							json.Unmarshal(pr.Ref.Raw, &m)
							fmt.Fprintf(GinkgoWriter, "Resource To Check: %s", m)
						}
					}
				})
			})

			AfterEach(func() {
				By("Cleanup the specific resource instance ResourceDefinition", func() {
					resourceDefinition := &api.ResourceDefinition{}
					err := k8sClient.Get(ctx, resourceDefinitionNamespacedName, resourceDefinition)
					Expect(err).NotTo(HaveOccurred())

					Expect(k8sClient.Delete(ctx, resourceDefinition)).To(Succeed())
				})

				By("Cleanup all generated Resources", func() {
					resourceList := &api.ResourceList{}
					err := k8sClient.List(ctx, resourceList, client.InNamespace(resourceDefinitionNamespacedName.Namespace))
					Expect(err).NotTo(HaveOccurred())

					for _, item := range resourceList.Items {
						err = k8sClient.Delete(ctx, &item)
						Expect(err).NotTo(HaveOccurred())
					}
				})
			})
		})

		When("...using a Data generator", func() {
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

				err := k8sClient.Get(ctx, resourceDefinitionNamespacedName, resourceDefinition)
				if err != nil && !errors.IsNotFound(err) {
					Fail(fmt.Sprintf("ResourceDefinition %s already exists.", resourceDefinitionName))
				}

				properties, err := serde.ToRaw(map[string]string{
					"name": "just-a-name",
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
						"name":              "${resource.spec.properties.name}",
						"belongsToResource": "${resource.metadata.name}-dev-${generator.parameters.value}",
					},
				})
				if err != nil {
					Fail(fmt.Sprintf("Failure to serialize provisioner spec to map: %v", err))
				}

				resourceDefinition = &api.ResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceDefinitionNamespacedName.Name,
						Namespace: resourceDefinitionNamespacedName.Namespace,
					},
					Spec: api.ResourceDefinitionSpec{
						Generator: map[string]*runtime.RawExtension{
							string(generators.DataGeneratorType): dataGeneratorSpec,
						},
						Resource: api.ResourceDefinitionResource{
							ObjectMeta: metav1.ObjectMeta{
								Name: "just-a-pet-${generator.parameters.value}",
							},
							Spec: api.ResourceSpec{
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
												Name: "pet-owner",
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
									},
								},
								Patches: api.ResourcePatches{
									api.ResourcePatch{
										From: "${provisioner.resources.configMap.data.name}",
										To:   "status.properties.name",
									},
								},
							},
						},
					},
				}

				Expect(k8sClient.Create(ctx, resourceDefinition)).To(Succeed())
			})

			It("...should successfully reconcile the resource.", func() {
				resourceDefinitionReconciler := &ResourceDefinitionReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := reconcile.AsReconciler(k8sClient, resourceDefinitionReconciler).Reconcile(ctx, reconcile.Request{
					NamespacedName: resourceDefinitionNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("we are expecting to create one Resource for each generator element", func() {
					dataGeneratorSpec, err := generators.UnmarshallSpec(dataGeneratorSpec, &generators.DataGeneratorSpec{})
					Expect(err).NotTo(HaveOccurred())

					for _, obj := range dataGeneratorSpec.Values {
						namespacedName := types.NamespacedName{
							Name:      fmt.Sprintf("just-a-pet-%s", obj["value"]),
							Namespace: resourceDefinitionNamespacedName.Namespace,
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
					err := k8sClient.Get(ctx, resourceDefinitionNamespacedName, resourceDefinition)
					Expect(err).NotTo(HaveOccurred())

					Expect(k8sClient.Delete(ctx, resourceDefinition)).To(Succeed())
				})

				By("Cleanup all generated Resources", func() {
					resourceList := &api.ResourceList{}
					err := k8sClient.List(ctx, resourceList, client.InNamespace(resourceDefinitionNamespacedName.Namespace))
					Expect(err).NotTo(HaveOccurred())

					for _, item := range resourceList.Items {
						err = k8sClient.Delete(ctx, &item)
						Expect(err).NotTo(HaveOccurred())
					}
				})
			})
		})

		When("...using an Inventory generator", func() {
			kcl := `import kcl_plugin.inventory

_environments = inventory.list_resources("environment")

environments = [{id: e.id, nurn: e.metadata.nurn, alias: e.metadata.alias} for e in _environments]
`
			inventoryGeneratorSpec, err := generators.NewInventoryGeneratorSpec("environments", kcl, "environments")
			if err != nil {
				Fail(fmt.Sprintf("Failure to generate a data generator spec: %v", err))
			}

			BeforeEach(func() {
				generators.Register(generators.InventoryGeneratorType, generators.NewInventoryGenerator(inventoryClient))
			})

			BeforeEach(func() {
				resourceDefinition := &api.ResourceDefinition{}

				err := k8sClient.Get(ctx, resourceDefinitionNamespacedName, resourceDefinition)
				if err != nil && !errors.IsNotFound(err) {
					Fail(fmt.Sprintf("ResourceDefinition %s already exists.", resourceDefinitionName))
				}

				properties, err := serde.ToRaw(map[string]string{
					"name": "just-a-name",
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
						"name":              "${resource.spec.properties.name}",
						"belongsToResource": "${resource.metadata.name}-dev-${generator.environments.alias}",
					},
				})
				if err != nil {
					Fail(fmt.Sprintf("Failure to serialize provisioner spec to map: %v", err))
				}

				resourceDefinition = &api.ResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceDefinitionNamespacedName.Name,
						Namespace: resourceDefinitionNamespacedName.Namespace,
					},
					Spec: api.ResourceDefinitionSpec{
						Generator: map[string]*runtime.RawExtension{
							string(generators.InventoryGeneratorType): inventoryGeneratorSpec,
						},
						Resource: api.ResourceDefinitionResource{
							ObjectMeta: metav1.ObjectMeta{
								Name: "just-a-pet-${generator.environments.alias}",
							},
							Spec: api.ResourceSpec{
								Name:            "just-a-pet-from-${generator.environments.alias}",
								Alias:           "just-a-pet-from-${generator.environments.alias}",
								Description:     "I'm just a pet, and my environment is ${generator.environments.alias}",
								ResourceTypeRef: "random-pet",
								Properties:      properties,
								Connections: []api.ResourceConnection{
									api.ResourceConnection{
										Via: "belongs-to",
										Target: api.ResourceConnectionTarget{
											Nurn: &api.ResourceConnectionTargetNurn{
												Value: "${generator.environments.nurn}",
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
									},
								},
								Patches: api.ResourcePatches{
									api.ResourcePatch{
										From: "${provisioner.resources.configMap.data.name}",
										To:   "status.properties.name",
									},
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resourceDefinition)).To(Succeed())
			})

			It("...should successfully reconcile the resource.", func() {
				resourceDefinitionReconciler := &ResourceDefinitionReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := reconcile.AsReconciler(k8sClient, resourceDefinitionReconciler).Reconcile(ctx, reconcile.Request{
					NamespacedName: resourceDefinitionNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("we are expecting to create one Resource for each generator element", func() {
					By("We don't know how many are...so we try to find Resources by labels", func() {
						resourceList := api.ResourceList{}
						err := k8sClient.List(ctx, &resourceList, client.HasLabels([]string{
							api.Group + "/managedBy.group",
							api.Group + "/managedBy.version",
							api.Group + "/managedBy.kind",
							api.Group + "/managedBy.name",
						}))
						Expect(err).NotTo(HaveOccurred())
						Expect(resourceList.Items).Should(Not(BeEmpty()))

						for _, resource := range resourceList.Items {
							fmt.Fprintf(GinkgoWriter, "Resource is: %s\n", resource.Name)
						}
					})

				})
			})

			AfterEach(func() {
				By("Cleanup the specific resource instance ResourceDefinition", func() {
					resourceDefinition := &api.ResourceDefinition{}
					err := k8sClient.Get(ctx, resourceDefinitionNamespacedName, resourceDefinition)
					Expect(err).NotTo(HaveOccurred())

					Expect(k8sClient.Delete(ctx, resourceDefinition)).To(Succeed())
				})

				By("Cleanup all generated Resources", func() {
					resourceList := api.ResourceList{}
					err := k8sClient.List(ctx, &resourceList, client.HasLabels([]string{
						api.Group + "/managedBy.group",
						api.Group + "/managedBy.version",
						api.Group + "/managedBy.kind",
						api.Group + "/managedBy.name",
					}))
					Expect(err).NotTo(HaveOccurred())

					for _, resource := range resourceList.Items {
						err = k8sClient.Delete(ctx, &resource)
						Expect(err).NotTo(HaveOccurred())
					}
				})
			})
		})

		When("...don't using generator at all", func() {
			BeforeEach(func() {
				resourceDefinition := &api.ResourceDefinition{}

				err := k8sClient.Get(ctx, resourceDefinitionNamespacedName, resourceDefinition)
				if err != nil && !errors.IsNotFound(err) {
					Fail(fmt.Sprintf("ResourceDefinition %s already exists.", resourceDefinitionName))
				}

				properties, err := serde.ToRaw(map[string]string{
					"name": "just-a-name",
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

				resourceDefinition = &api.ResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceDefinitionNamespacedName.Name,
						Namespace: resourceDefinitionNamespacedName.Namespace,
					},
					Spec: api.ResourceDefinitionSpec{
						Resource: api.ResourceDefinitionResource{
							ObjectMeta: metav1.ObjectMeta{
								Name: "just-a-pet-with-no-name",
							},
							Spec: api.ResourceSpec{
								Name:            "just-a-pet-called-no-name",
								Alias:           "just-a-pet-called-no-name",
								Description:     "I'm just a pet, and my name is no-name",
								ResourceTypeRef: "random-pet",
								Properties:      properties,
								Connections: []api.ResourceConnection{
									api.ResourceConnection{
										Via: "belongs-to",
										Target: api.ResourceConnectionTarget{
											Ref: &api.ResourceConnectionTargetRef{
												Name: "pet-owner",
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
									},
								},
								Patches: api.ResourcePatches{
									api.ResourcePatch{
										From: "${provisioner.resources.configMap.data.name}",
										To:   "status.properties.name",
									},
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resourceDefinition)).To(Succeed())
			})

			It("...should successfully reconcile the resource.", func() {
				resourceDefinitionReconciler := &ResourceDefinitionReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := reconcile.AsReconciler(k8sClient, resourceDefinitionReconciler).Reconcile(ctx, reconcile.Request{
					NamespacedName: resourceDefinitionNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("we are expecting to create just one Resource", func() {
					namespacedName := types.NamespacedName{
						Name:      "just-a-pet-with-no-name",
						Namespace: resourceDefinitionNamespacedName.Namespace,
					}

					resource := &api.Resource{}
					err := k8sClient.Get(ctx, namespacedName, resource)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			AfterEach(func() {
				By("Cleanup the specific resource instance ResourceDefinition", func() {
					resourceDefinition := &api.ResourceDefinition{}
					err := k8sClient.Get(ctx, resourceDefinitionNamespacedName, resourceDefinition)
					Expect(err).NotTo(HaveOccurred())

					Expect(k8sClient.Delete(ctx, resourceDefinition)).To(Succeed())
				})

				By("Cleanup all generated Resources", func() {
					resourceList := &api.ResourceList{}
					err := k8sClient.List(ctx, resourceList, client.InNamespace(resourceDefinitionNamespacedName.Namespace))
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
