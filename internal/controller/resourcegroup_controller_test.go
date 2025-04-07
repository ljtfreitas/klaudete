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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/nubank/klaudete/api/v1alpha1"
)

var _ = Describe("ResourceGroup Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceGroupName = "sample-resource-group"

		ctx := context.Background()

		resourceGroupNamespacedName := types.NamespacedName{
			Name:      resourceGroupName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		resourceGroup := &api.ResourceGroup{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ResourceGroup", func() {
				err := k8sClient.Get(ctx, resourceGroupNamespacedName, resourceGroup)
				if err != nil && errors.IsNotFound(err) {
					resourceGroup := &api.ResourceGroup{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      resourceGroupName,
						},
						Spec: api.ResourceGroupSpec{
							Resources: []api.ResourceGroupResource{
								api.ResourceGroupResource{
									Name: "just-a-resource",
									Spec: api.ResourceSpec{
										Name:        "just-a-resource",
										Alias:       "just-a-resource",
										Description: "I'm just a resource",
									},
								},
								api.ResourceGroupResource{
									Name: "just-another-resource",
									Spec: api.ResourceSpec{
										Name:        "just-another-resource",
										Alias:       "just-another-resource",
										Description: "I'm just a resource",
										Connections: []api.ResourceConnection{
											api.ResourceConnection{
												Via: "belongs-to",
												Target: api.ResourceConnectionTarget{
													Ref: &api.ResourceConnectionTargetRef{
														ApiVersion: "klaudete.nubank.com.br",
														Kind:       "Resource",
														Name:       `${resources["just-a-resource"].metadata.name}`,
													},
												},
											},
										},
									},
								},
							},
						},
					}
					Expect(k8sClient.Create(ctx, resourceGroup)).To(Succeed())
				}
			})
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &api.ResourceGroup{}
			err := k8sClient.Get(ctx, resourceGroupNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ResourceGroup", func() {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			})
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource", func() {
				resourceGroupReconciler := &ResourceGroupReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := reconcile.AsReconciler(k8sClient, resourceGroupReconciler).Reconcile(ctx, reconcile.Request{
					NamespacedName: resourceGroupNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
				// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
				// Example: If you expect a certain status condition after reconciliation, verify it here.
			})
		})
	})
})
