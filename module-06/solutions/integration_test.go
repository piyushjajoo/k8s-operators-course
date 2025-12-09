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

// Solution: Integration Test Example from Module 6
// This demonstrates end-to-end testing with a real Kubernetes cluster
// Location: test/integration/integration_suite_test.go and test/integration/database_test.go

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	databasev1 "github.com/example/postgres-operator/api/v1"
)

var (
	k8sClient client.Client
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func() {
	By("setting up integration test environment")

	// CRITICAL: Register the Database type with the scheme
	// Without this, the client won't know how to serialize/deserialize Database objects!
	err := databasev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	cfg, err := config.GetConfig()
	Expect(err).NotTo(HaveOccurred())

	// Pass the scheme to the client so it knows about our custom types
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = Describe("Database Operator Integration", func() {
	var (
		ctx      context.Context
		cancel   context.CancelFunc
		timeout  = 5 * time.Minute
		interval = 2 * time.Second
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cancel()
	})

	Context("Database lifecycle", func() {
		var (
			dbName string
			key    types.NamespacedName
		)

		BeforeEach(func() {
			// Use unique name per test to avoid conflicts
			dbName = fmt.Sprintf("integration-test-%d", time.Now().UnixNano())
			key = types.NamespacedName{
				Name:      dbName,
				Namespace: "default",
			}
		})

		AfterEach(func() {
			// Cleanup: delete the Database if it exists
			db := &databasev1.Database{}
			if err := k8sClient.Get(ctx, key, db); err == nil {
				// Remove finalizer to allow deletion
				db.Finalizers = nil
				_ = k8sClient.Update(ctx, db)
				_ = k8sClient.Delete(ctx, db)
			}
		})

		It("should create, update, and delete a Database", func() {
			By("Creating a Database resource")
			db := &databasev1.Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dbName,
					Namespace: "default",
				},
				Spec: databasev1.DatabaseSpec{
					Image:        "postgres:14",
					Replicas:     ptr.To(int32(1)),
					DatabaseName: "mydb",
					Username:     "admin",
					Storage: databasev1.StorageSpec{
						Size: "1Gi",
					},
				},
			}
			Expect(k8sClient.Create(ctx, db)).To(Succeed())

			By("Waiting for StatefulSet to be created")
			Eventually(func() error {
				ss := &appsv1.StatefulSet{}
				return k8sClient.Get(ctx, key, ss)
			}, timeout, interval).Should(Succeed())

			By("Verifying StatefulSet has correct initial spec")
			ss := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(ctx, key, ss)).To(Succeed())
			Expect(*ss.Spec.Replicas).To(Equal(int32(1)))

			By("Updating Database replicas to 3")
			Expect(k8sClient.Get(ctx, key, db)).To(Succeed())
			db.Spec.Replicas = ptr.To(int32(3))
			Expect(k8sClient.Update(ctx, db)).To(Succeed())

			By("Waiting for StatefulSet replicas to be updated to 3")
			Eventually(func() int32 {
				ss := &appsv1.StatefulSet{}
				if err := k8sClient.Get(ctx, key, ss); err != nil {
					return 0
				}
				return *ss.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(3)))

			By("Deleting the Database")
			Expect(k8sClient.Delete(ctx, db)).To(Succeed())

			By("Verifying the Database is deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, db)
				return client.IgnoreNotFound(err) == nil
			}, timeout, interval).Should(BeTrue())
		})

		It("should create all child resources", func() {
			By("Creating a Database resource")
			db := &databasev1.Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dbName,
					Namespace: "default",
				},
				Spec: databasev1.DatabaseSpec{
					Image:        "postgres:14",
					Replicas:     ptr.To(int32(1)),
					DatabaseName: "mydb",
					Username:     "admin",
					Storage: databasev1.StorageSpec{
						Size: "1Gi",
					},
				},
			}
			Expect(k8sClient.Create(ctx, db)).To(Succeed())

			By("Verifying the StatefulSet was created")
			Eventually(func() error {
				ss := &appsv1.StatefulSet{}
				return k8sClient.Get(ctx, key, ss)
			}, timeout, interval).Should(Succeed())

			By("Verifying the Service was created")
			Eventually(func() error {
				svc := &corev1.Service{}
				return k8sClient.Get(ctx, key, svc)
			}, timeout, interval).Should(Succeed())

			By("Verifying the Secret was created")
			secretKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-credentials", dbName),
				Namespace: "default",
			}
			Eventually(func() error {
				secret := &corev1.Secret{}
				return k8sClient.Get(ctx, secretKey, secret)
			}, timeout, interval).Should(Succeed())
		})
	})

	// Only run if webhooks are deployed
	Context("Validating webhook", func() {
		It("should reject invalid Database", func() {
			db := &databasev1.Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-db",
					Namespace: "default",
				},
				Spec: databasev1.DatabaseSpec{
					Image:        "nginx:latest", // Invalid: not PostgreSQL
					DatabaseName: "mydb",
					Username:     "admin",
					Storage: databasev1.StorageSpec{
						Size: "10Gi",
					},
				},
			}

			err := k8sClient.Create(ctx, db)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must be a PostgreSQL image"))
		})

		It("should accept valid Database", func() {
			db := &databasev1.Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-db",
					Namespace: "default",
				},
				Spec: databasev1.DatabaseSpec{
					Image:        "postgres:14",
					DatabaseName: "mydb",
					Username:     "admin",
					Storage: databasev1.StorageSpec{
						Size: "10Gi",
					},
				},
			}

			Expect(k8sClient.Create(ctx, db)).To(Succeed())

			// Cleanup
			db.Finalizers = nil
			_ = k8sClient.Update(ctx, db)
			Expect(k8sClient.Delete(ctx, db)).To(Succeed())
		})
	})
})
