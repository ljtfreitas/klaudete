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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/nubank/klaudete/api/v1alpha1"
	"github.com/nubank/klaudete/internal/generators"
	"github.com/nubank/klaudete/internal/serde"
)

var _ = Describe("ResourceGroupDefinition Controller", Ordered, func() {
	Describe("When reconciling a ResourceGroupDefinition...", func() {
		const resourceGroupDefinitionName = "resource-group-definition"

		ctx := context.Background()

		resourceGroupDefinitionNamespacedName := types.NamespacedName{
			Name:      resourceGroupDefinitionName,
			Namespace: "default",
		}

		resourceProperties, err := serde.ToRaw(map[string]string{
			"name": "just-a-name",
		})
		if err != nil {
			Fail(fmt.Sprintf("Failure to serialize a resource properties map: %v", err))
		}

		When("...using a List generator", func() {
			listGeneratorSpec, err := generators.NewListGeneratorSpec("parameter", "one", "two")
			if err != nil {
				Fail(fmt.Sprintf("Failure to generate a list generator spec: %v", err))
			}

			BeforeEach(func() {
				err := k8sClient.Get(ctx, resourceGroupDefinitionNamespacedName, &api.ResourceGroupDefinition{})
				if err != nil && errors.IsNotFound(err) {

					resourceGroupDefinition := &api.ResourceGroupDefinition{
						ObjectMeta: metav1.ObjectMeta{
							Name:      resourceGroupDefinitionName,
							Namespace: "default",
						},
						Spec: api.ResourceGroupDefinitionSpec{
							Generator: map[string]*runtime.RawExtension{
								string(generators.ListGeneratorType): listGeneratorSpec,
							},
							Group: api.ResourceGroupDefinitionGroup{
								Name: "just-a-resource-group-${generator.parameter}",
								Resources: []api.ResourceGroupResource{
									api.ResourceGroupResource{
										Name: "just-a-resource-${generator.parameter}",
										Spec: api.ResourceSpec{
											Name:            "i-am-just-a-resource-called-${generator.parameter}",
											Alias:           "i-am-just-a-resource-called-${generator.parameter}",
											Description:     "I am Just a Resource Called ${generator.parameter}",
											ResourceTypeRef: "my-resource-type",
											Properties:      resourceProperties,
										},
									},
								},
							},
						},
					}
					Expect(k8sClient.Create(ctx, resourceGroupDefinition)).To(Succeed())
				}
			})

			It("...should successfully reconcile the resource", func() {
				By("Reconciling the created resource", func() {
					resourceGroupDefinitionReconciler := &ResourceGroupDefinitionReconciler{
						Client: k8sClient,
						Scheme: k8sClient.Scheme(),
					}

					_, err := reconcile.AsReconciler(k8sClient, resourceGroupDefinitionReconciler).Reconcile(ctx, reconcile.Request{
						NamespacedName: resourceGroupDefinitionNamespacedName,
					})
					Expect(err).NotTo(HaveOccurred())

					By("we are expecting to create one ResourceGroup for each generator element", func() {
						listGeneratorSpec, err := generators.UnmarshallSpec(listGeneratorSpec, &generators.ListGeneratorSpec{})
						Expect(err).NotTo(HaveOccurred())

						for _, value := range listGeneratorSpec.Values {
							namespacedName := types.NamespacedName{
								Name:      fmt.Sprintf("just-a-resource-group-%s", value),
								Namespace: "default",
							}

							resource := &api.ResourceGroup{}
							err := k8sClient.Get(ctx, namespacedName, resource)
							Expect(err).NotTo(HaveOccurred())
						}
					})
				})
			})
			AfterEach(func() {
				By("Cleanup the specific resource instance ResourceGroupDefinition", func() {
					resourceGroupDefinition := &api.ResourceGroupDefinition{}
					err := k8sClient.Get(ctx, resourceGroupDefinitionNamespacedName, resourceGroupDefinition)
					Expect(err).NotTo(HaveOccurred())

					Expect(k8sClient.Delete(ctx, resourceGroupDefinition)).To(Succeed())
				})

				By("Cleanup all generated ResourceGroups", func() {
					resourceGroupList := &api.ResourceGroupList{}
					err := k8sClient.List(ctx, resourceGroupList, client.InNamespace(resourceGroupDefinitionNamespacedName.Namespace))
					Expect(err).NotTo(HaveOccurred())

					for _, resourceGroup := range resourceGroupList.Items {
						err = k8sClient.Delete(ctx, &resourceGroup)
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
				err := k8sClient.Get(ctx, resourceGroupDefinitionNamespacedName, &api.ResourceGroupDefinition{})
				if err != nil && errors.IsNotFound(err) {

					resourceGroupDefinition := &api.ResourceGroupDefinition{
						ObjectMeta: metav1.ObjectMeta{
							Name:      resourceGroupDefinitionName,
							Namespace: "default",
						},
						Spec: api.ResourceGroupDefinitionSpec{
							Generator: map[string]*runtime.RawExtension{
								string(generators.DataGeneratorType): dataGeneratorSpec,
							},
							Group: api.ResourceGroupDefinitionGroup{
								Name: "just-a-resource-group-${generator.parameters.value}",
								Resources: []api.ResourceGroupResource{
									api.ResourceGroupResource{
										Name: "just-a-resource-${generator.parameters.value}",
										Spec: api.ResourceSpec{
											Name:            "i-am-just-a-resource-called-${generator.parameters.value}",
											Alias:           "i-am-just-a-resource-called-${generator.parameters.value}",
											Description:     "I am Just a Resource Called ${generator.parameters.value}",
											ResourceTypeRef: "my-resource-type",
											Properties:      resourceProperties,
										},
									},
								},
							},
						},
					}
					Expect(k8sClient.Create(ctx, resourceGroupDefinition)).To(Succeed())
				}
			})

			It("...should successfully reconcile the resource", func() {
				By("Reconciling the created resource", func() {
					resourceGroupDefinitionReconciler := &ResourceGroupDefinitionReconciler{
						Client: k8sClient,
						Scheme: k8sClient.Scheme(),
					}

					_, err := reconcile.AsReconciler(k8sClient, resourceGroupDefinitionReconciler).Reconcile(ctx, reconcile.Request{
						NamespacedName: resourceGroupDefinitionNamespacedName,
					})
					Expect(err).NotTo(HaveOccurred())

					By("we are expecting to create one ResourceGroup for each generator element", func() {
						dataGeneratorSpec, err := generators.UnmarshallSpec(dataGeneratorSpec, &generators.DataGeneratorSpec{})
						Expect(err).NotTo(HaveOccurred())

						for _, obj := range dataGeneratorSpec.Values {
							namespacedName := types.NamespacedName{
								Name:      fmt.Sprintf("just-a-resource-group-%s", obj["value"]),
								Namespace: resourceGroupDefinitionNamespacedName.Namespace,
							}

							resourceGroup := &api.ResourceGroup{}
							err := k8sClient.Get(ctx, namespacedName, resourceGroup)
							Expect(err).NotTo(HaveOccurred())
						}
					})
				})
			})

			AfterEach(func() {
				By("Cleanup the specific resource instance ResourceGroupDefinition", func() {
					resourceGroupDefinition := &api.ResourceGroupDefinition{}
					err := k8sClient.Get(ctx, resourceGroupDefinitionNamespacedName, resourceGroupDefinition)
					Expect(err).NotTo(HaveOccurred())

					Expect(k8sClient.Delete(ctx, resourceGroupDefinition)).To(Succeed())
				})

				By("Cleanup all generated ResourceGroups", func() {
					resourceGroupList := &api.ResourceGroupList{}
					err := k8sClient.List(ctx, resourceGroupList, client.InNamespace(resourceGroupDefinitionNamespacedName.Namespace))
					Expect(err).NotTo(HaveOccurred())

					for _, resourceGroup := range resourceGroupList.Items {
						err = k8sClient.Delete(ctx, &resourceGroup)
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
				resourceGroupDefinition := &api.ResourceGroupDefinition{}

				err := k8sClient.Get(ctx, resourceGroupDefinitionNamespacedName, resourceGroupDefinition)
				if err != nil && !errors.IsNotFound(err) {
					Fail(fmt.Sprintf("ResourceGroupDefinition %s already exists.", resourceGroupDefinitionName))
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

				resourceGroupDefinition = &api.ResourceGroupDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceGroupDefinitionNamespacedName.Name,
						Namespace: resourceGroupDefinitionNamespacedName.Namespace,
					},
					Spec: api.ResourceGroupDefinitionSpec{
						Generator: map[string]*runtime.RawExtension{
							string(generators.InventoryGeneratorType): inventoryGeneratorSpec,
						},
						Group: api.ResourceGroupDefinitionGroup{
							Name: "just-a-resource-group-${generator.environments.alias}",
							Resources: []api.ResourceGroupResource{
								api.ResourceGroupResource{
									Name: "just-a-resource-${generator.environments.alias}",
									Spec: api.ResourceSpec{
										Name:            "i-am-just-a-resource-called-${generator.environments.alias}",
										Alias:           "i-am-just-a-resource-called-${generator.environments.alias}",
										Description:     "I am Just a Resource Called ${generator.environments.alias}",
										ResourceTypeRef: "my-resource-type",
										Properties:      resourceProperties,
										Connections: []api.ResourceConnection{
											api.ResourceConnection{
												Via: "belongs-to",
												Target: api.ResourceConnectionTarget{
													Nurn: ptr.To("${generator.environments.nurn}"),
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
									},
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resourceGroupDefinition)).To(Succeed())
			})

			It("...should successfully reconcile the resource.", func() {
				resourceGroupDefinitionReconciler := &ResourceGroupDefinitionReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := reconcile.AsReconciler(k8sClient, resourceGroupDefinitionReconciler).Reconcile(ctx, reconcile.Request{
					NamespacedName: resourceGroupDefinitionNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("we are expecting to create one ResourceGroup for each generator element", func() {
					By("we don't know how many are...so we try to find ResourceGroup by labels", func() {
						resourceGroupList := api.ResourceGroupList{}
						err := k8sClient.List(ctx, &resourceGroupList, client.HasLabels([]string{
							api.Group + "/managedBy.group",
							api.Group + "/managedBy.version",
							api.Group + "/managedBy.kind",
							api.Group + "/managedBy.name",
						}))
						Expect(err).NotTo(HaveOccurred())
						Expect(resourceGroupList.Items).Should(Not(BeEmpty()))

						for _, resource := range resourceGroupList.Items {
							fmt.Fprintf(GinkgoWriter, "ResourceGroup is: %s\n", resource.Name)
						}
					})

				})
			})

			AfterEach(func() {
				By("Cleanup the specific resource instance ResourceGroupDefinition", func() {
					resourceGroupDefinition := &api.ResourceGroupDefinition{}
					err := k8sClient.Get(ctx, resourceGroupDefinitionNamespacedName, resourceGroupDefinition)
					Expect(err).NotTo(HaveOccurred())

					Expect(k8sClient.Delete(ctx, resourceGroupDefinition)).To(Succeed())
				})

				By("Cleanup all generated ResourceGroups", func() {
					resourceGroupList := api.ResourceGroupList{}
					err := k8sClient.List(ctx, &resourceGroupList, client.HasLabels([]string{
						api.Group + "/managedBy.group",
						api.Group + "/managedBy.version",
						api.Group + "/managedBy.kind",
						api.Group + "/managedBy.name",
					}))
					Expect(err).NotTo(HaveOccurred())

					for _, resourceGroup := range resourceGroupList.Items {
						err = k8sClient.Delete(ctx, &resourceGroup)
						Expect(err).NotTo(HaveOccurred())
					}
				})
			})
		})
	})
})
