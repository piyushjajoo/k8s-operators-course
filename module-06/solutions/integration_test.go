// Solution: Integration Test Example from Module 6
// This demonstrates end-to-end testing with a real Kubernetes cluster

package integration

import (
    "context"
    "testing"
    "time"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/utils/pointer"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/client/config"
    
    databasev1 "github.com/example/postgres-operator/api/v1"
    appsv1 "k8s.io/api/apps/v1"
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
    
    cfg, err := config.GetConfig()
    Expect(err).NotTo(HaveOccurred())
    
    k8sClient, err = client.New(cfg, client.Options{})
    Expect(err).NotTo(HaveOccurred())
    Expect(k8sClient).NotTo(BeNil())
})

var _ = Describe("Database Operator Integration", func() {
    var (
        ctx    context.Context
        cancel context.CancelFunc
        timeout = 5 * time.Minute
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
                    Image:       "postgres:14",
                    Replicas:    pointer.Int32(1),
                    DatabaseName: "mydb",
                    Username:    "admin",
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
                k8sClient.Get(ctx, key, ss)
                return ss.Status.ReadyReplicas == *ss.Spec.Replicas
            }, timeout, interval).Should(BeTrue())
            
            // Update Database
            Expect(k8sClient.Get(ctx, key, db)).To(Succeed())
            db.Spec.Replicas = pointer.Int32(3)
            Expect(k8sClient.Update(ctx, db)).To(Succeed())
            
            // Wait for update
            Eventually(func() *int32 {
                ss := &appsv1.StatefulSet{}
                k8sClient.Get(ctx, key, ss)
                return ss.Spec.Replicas
            }, timeout, interval).Should(Equal(pointer.Int32(3)))
            
            // Delete Database
            Expect(k8sClient.Delete(ctx, db)).To(Succeed())
            
            // Verify cleanup
            Eventually(func() bool {
                err := k8sClient.Get(ctx, key, db)
                return client.IgnoreNotFound(err) == nil
            }, timeout, interval).Should(BeTrue())
        })
    })
    
    Context("Validating webhook", func() {
        It("should reject invalid Database", func() {
            db := &databasev1.Database{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "invalid-db",
                    Namespace: "default",
                },
                Spec: databasev1.DatabaseSpec{
                    Image: "nginx:latest", // Invalid: not PostgreSQL
                    DatabaseName: "mydb",
                    Username: "admin",
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
                    Image:       "postgres:14",
                    DatabaseName: "mydb",
                    Username:    "admin",
                    Storage: databasev1.StorageSpec{
                        Size: "10Gi",
                    },
                },
            }
            
            Expect(k8sClient.Create(ctx, db)).To(Succeed())
        })
    })
})

