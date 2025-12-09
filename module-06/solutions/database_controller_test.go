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

// Solution: Database Controller Tests from Module 6
// This file demonstrates comprehensive unit testing patterns for a state machine controller.
// Location: internal/controller/database_controller_test.go

package controller

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	databasev1 "github.com/example/postgres-operator/api/v1"
)

var _ = Describe("Database Controller", func() {
	// Original scaffolded test
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		database := &databasev1.Database{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Database")
			err := k8sClient.Get(ctx, typeNamespacedName, database)
			if err != nil && errors.IsNotFound(err) {
				resource := &databasev1.Database{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: databasev1.DatabaseSpec{
						Image:        "postgres:14",
						DatabaseName: "testdb",
						Username:     "testuser",
						Storage: databasev1.StorageSpec{
							Size: "1Gi",
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &databasev1.Database{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				// Remove finalizer to allow deletion
				resource.Finalizers = nil
				_ = k8sClient.Update(ctx, resource)
				By("Cleanup the specific resource instance Database")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &DatabaseReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	// Additional tests for state machine behavior
	Context("When reconciling a new Database", func() {
		var (
			resourceName       string
			typeNamespacedName types.NamespacedName
		)

		BeforeEach(func() {
			resourceName = fmt.Sprintf("test-db-%d", time.Now().UnixNano())
			typeNamespacedName = types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			resource := &databasev1.Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: databasev1.DatabaseSpec{
					Image:        "postgres:14",
					Replicas:     ptr.To(int32(1)),
					DatabaseName: "testdb",
					Username:     "testuser",
					Storage: databasev1.StorageSpec{
						Size: "1Gi",
					},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
		})

		AfterEach(func() {
			resource := &databasev1.Database{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				resource.Finalizers = nil
				_ = k8sClient.Update(ctx, resource)
				_ = k8sClient.Delete(ctx, resource)
			}
		})

		It("should transition from Pending to Provisioning", func() {
			By("Reconciling the created resource")
			controllerReconciler := &DatabaseReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			db := &databasev1.Database{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, db)).To(Succeed())
			Expect(db.Status.Phase).To(Equal("Provisioning"))
			Expect(db.Status.Ready).To(BeFalse())
		})

		It("should add finalizer on first reconcile", func() {
			reconciler := &DatabaseReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			db := &databasev1.Database{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, db)).To(Succeed())
			Expect(db.Finalizers).To(ContainElement("database.example.com/finalizer"))
		})
	})

	Context("When progressing through provisioning", func() {
		var (
			resourceName       string
			typeNamespacedName types.NamespacedName
		)

		BeforeEach(func() {
			resourceName = fmt.Sprintf("test-provision-%d", time.Now().UnixNano())
			typeNamespacedName = types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			resource := &databasev1.Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: databasev1.DatabaseSpec{
					Image:        "postgres:14",
					Replicas:     ptr.To(int32(1)),
					DatabaseName: "testdb",
					Username:     "testuser",
					Storage: databasev1.StorageSpec{
						Size: "1Gi",
					},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
		})

		AfterEach(func() {
			resource := &databasev1.Database{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				resource.Finalizers = nil
				_ = k8sClient.Update(ctx, resource)
				_ = k8sClient.Delete(ctx, resource)
			}
		})

		It("should create Secret and StatefulSet", func() {
			reconciler := &DatabaseReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			req := reconcile.Request{NamespacedName: typeNamespacedName}

			By("First reconcile: Pending -> Provisioning")
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			By("Second reconcile: Creates Secret and StatefulSet")
			_, err = reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			By("Verifying Secret was created")
			secret := &corev1.Secret{}
			secretName := fmt.Sprintf("%s-credentials", resourceName)
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      secretName,
				Namespace: "default",
			}, secret)).To(Succeed())
			Expect(secret.Data).To(HaveKey("username"))
			Expect(secret.Data).To(HaveKey("password"))

			By("Verifying StatefulSet was created")
			statefulSet := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, statefulSet)).To(Succeed())
			Expect(*statefulSet.Spec.Replicas).To(Equal(int32(1)))
			Expect(statefulSet.Spec.Template.Spec.Containers[0].Image).To(Equal("postgres:14"))
		})
	})

	Context("When Database is not found", func() {
		It("should not return an error", func() {
			reconciler := &DatabaseReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "non-existent-database",
					Namespace: "default",
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())
			Expect(result.RequeueAfter).To(Equal(time.Duration(0)))
		})
	})

	Context("When in Configuring phase", func() {
		var (
			resourceName       string
			typeNamespacedName types.NamespacedName
		)

		BeforeEach(func() {
			resourceName = fmt.Sprintf("test-service-%d", time.Now().UnixNano())
			typeNamespacedName = types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			resource := &databasev1.Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: databasev1.DatabaseSpec{
					Image:        "postgres:14",
					DatabaseName: "testdb",
					Username:     "testuser",
					Storage: databasev1.StorageSpec{
						Size: "1Gi",
					},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
		})

		AfterEach(func() {
			resource := &databasev1.Database{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				resource.Finalizers = nil
				_ = k8sClient.Update(ctx, resource)
				_ = k8sClient.Delete(ctx, resource)
			}
		})

		It("should create Service", func() {
			reconciler := &DatabaseReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			req := reconcile.Request{NamespacedName: typeNamespacedName}

			By("Progress through states to Configuring")
			// Pending -> Provisioning
			_, _ = reconciler.Reconcile(ctx, req)
			// Provisioning: creates Secret + StatefulSet, stays in Provisioning
			_, _ = reconciler.Reconcile(ctx, req)
			// Provisioning -> Configuring (StatefulSet exists)
			_, _ = reconciler.Reconcile(ctx, req)
			// Configuring: creates Service
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			By("Verifying Service was created")
			service := &corev1.Service{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, service)).To(Succeed())
			Expect(service.Spec.Ports[0].Port).To(Equal(int32(5432)))
		})
	})
})
