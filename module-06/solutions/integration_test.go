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
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
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
		interval = 5 * time.Second
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cancel()
	})

	Context("Database lifecycle", func() {
		It("should create, update, and delete a Database", func() {
			// Create Database
			db := &databasev1.Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "integration-test-db",
					Namespace: "default",
				},
				Spec: databasev1.DatabaseSpec{
					Image:        "postgres:14",
					Replicas:     ptr.To(int32(1)),
					DatabaseName: "mydb",
					Username:     "admin",
					Storage: databasev1.StorageSpec{
						Size: "10Gi",
					},
				},
			}
			Expect(k8sClient.Create(ctx, db)).To(Succeed())

			key := types.NamespacedName{
				Name:      "integration-test-db",
				Namespace: "default",
			}

			// Wait for StatefulSet to be created
			Eventually(func() error {
				ss := &appsv1.StatefulSet{}
				return k8sClient.Get(ctx, key, ss)
			}, timeout, interval).Should(Succeed())

			// Verify StatefulSet is ready
			Eventually(func() bool {
				ss := &appsv1.StatefulSet{}
				if err := k8sClient.Get(ctx, key, ss); err != nil {
					return false
				}
				return ss.Status.ReadyReplicas == *ss.Spec.Replicas
			}, timeout, interval).Should(BeTrue())

			// Update Database
			Expect(k8sClient.Get(ctx, key, db)).To(Succeed())
			db.Spec.Replicas = ptr.To(int32(3))
			Expect(k8sClient.Update(ctx, db)).To(Succeed())

			// Wait for update
			Eventually(func() int32 {
				ss := &appsv1.StatefulSet{}
				if err := k8sClient.Get(ctx, key, ss); err != nil {
					return 0
				}
				return *ss.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(3)))

			// Delete Database
			Expect(k8sClient.Delete(ctx, db)).To(Succeed())

			// Verify cleanup
			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, db)
				return client.IgnoreNotFound(err) == nil
			}, timeout, interval).Should(BeTrue())
		})
	})

	// Note: Webhook tests only work when webhooks are deployed with cert-manager
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
			Expect(k8sClient.Delete(ctx, db)).To(Succeed())
		})
	})
})
