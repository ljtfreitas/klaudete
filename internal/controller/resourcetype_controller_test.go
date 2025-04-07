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

var _ = Describe("ResourceType Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceTypeName = "my-resource-type"

		ctx := context.Background()

		resourceTypeNamespacedName := types.NamespacedName{
			Name: resourceTypeName,
		}
		resourcetype := &api.ResourceType{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ResourceType", func() {
				err := k8sClient.Get(ctx, resourceTypeNamespacedName, resourcetype)
				if err != nil && errors.IsNotFound(err) {
					resource := &api.ResourceType{
						ObjectMeta: metav1.ObjectMeta{
							Name: resourceTypeName,
						},
						Spec: api.ResourceTypeSpec{
							Name:        "my-resource-type",
							Description: "I'm just a resource type.",
						},
					}
					Expect(k8sClient.Create(ctx, resource)).To(Succeed())
				}
			})
		})

		AfterEach(func() {
			resource := &api.ResourceType{}
			err := k8sClient.Get(ctx, resourceTypeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ResourceType", func() {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			})
		})
		It("...should successfully reconcile the resource", func() {
			controllerReconciler := &ResourceTypeReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := reconcile.AsReconciler(k8sClient, controllerReconciler).Reconcile(ctx, reconcile.Request{
				NamespacedName: resourceTypeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			resourceType := &api.ResourceType{}
			err = k8sClient.Get(ctx, resourceTypeNamespacedName, resourceType)
			Expect(err).NotTo(HaveOccurred())

			resourceTypeStatus := resourceType.Status.Status
			Expect(resourceTypeStatus).Should(Equal(api.ResourceTypeStatusInSync))
		})
	})
})
