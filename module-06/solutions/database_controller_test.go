// Solution: Complete Unit Tests for Database Controller from Module 6
// This demonstrates comprehensive unit testing with envtest

package controller

import (
    "context"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/client-go/kubernetes/scheme"
    "k8s.io/utils/pointer"
    "sigs.k8s.io/controller-runtime/pkg/client"
    ctrl "sigs.k8s.io/controller-runtime"
    
    databasev1 "github.com/example/postgres-operator/api/v1"
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
)

var _ = Describe("DatabaseReconciler", func() {
    var (
        ctx    context.Context
        cancel context.CancelFunc
    )
    
    BeforeEach(func() {
        ctx, cancel = context.WithCancel(context.Background())
    })
    
    AfterEach(func() {
        cancel()
    })
    
    Context("When reconciling a Database", func() {
        It("should create a StatefulSet", func() {
            // Arrange
            db := &databasev1.Database{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-db",
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
            
            // Act
            reconciler := &DatabaseReconciler{
                Client: k8sClient,
                Scheme: scheme.Scheme,
            }
            _, err := reconciler.Reconcile(ctx, ctrl.Request{
                NamespacedName: types.NamespacedName{
                    Name:      "test-db",
                    Namespace: "default",
                },
            })
            
            // Assert
            Expect(err).NotTo(HaveOccurred())
            
            statefulSet := &appsv1.StatefulSet{}
            Expect(k8sClient.Get(ctx, types.NamespacedName{
                Name:      "test-db",
                Namespace: "default",
            }, statefulSet)).To(Succeed())
            
            Expect(statefulSet.Spec.Replicas).To(Equal(pointer.Int32(1)))
            Expect(statefulSet.Spec.Template.Spec.Containers[0].Image).To(Equal("postgres:14"))
        })
        
        It("should create a Service", func() {
            // Arrange
            db := &databasev1.Database{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-db",
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
            
            // Act
            reconciler := &DatabaseReconciler{
                Client: k8sClient,
                Scheme: scheme.Scheme,
            }
            _, err := reconciler.Reconcile(ctx, ctrl.Request{
                NamespacedName: types.NamespacedName{
                    Name:      "test-db",
                    Namespace: "default",
                },
            })
            
            // Assert
            Expect(err).NotTo(HaveOccurred())
            
            service := &corev1.Service{}
            Expect(k8sClient.Get(ctx, types.NamespacedName{
                Name:      "test-db",
                Namespace: "default",
            }, service)).To(Succeed())
            
            Expect(service.Spec.Ports[0].Port).To(Equal(int32(5432)))
        })
    })
    
    Context("When updating a Database", func() {
        It("should update the StatefulSet", func() {
            // Create initial Database
            db := &databasev1.Database{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-db",
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
            
            reconciler := &DatabaseReconciler{
                Client: k8sClient,
                Scheme: scheme.Scheme,
            }
            
            req := ctrl.Request{
                NamespacedName: types.NamespacedName{
                    Name:      "test-db",
                    Namespace: "default",
                },
            }
            
            // Reconcile to create StatefulSet
            reconciler.Reconcile(ctx, req)
            
            // Update Database
            Expect(k8sClient.Get(ctx, req.NamespacedName, db)).To(Succeed())
            db.Spec.Replicas = pointer.Int32(3)
            Expect(k8sClient.Update(ctx, db)).To(Succeed())
            
            // Reconcile again
            reconciler.Reconcile(ctx, req)
            
            // Verify StatefulSet updated
            statefulSet := &appsv1.StatefulSet{}
            Expect(k8sClient.Get(ctx, req.NamespacedName, statefulSet)).To(Succeed())
            Expect(statefulSet.Spec.Replicas).To(Equal(pointer.Int32(3)))
        })
    })
    
    Context("When Database is not found", func() {
        It("should not return an error", func() {
            reconciler := &DatabaseReconciler{
                Client: k8sClient,
                Scheme: scheme.Scheme,
            }
            
            req := ctrl.Request{
                NamespacedName: types.NamespacedName{
                    Name:      "non-existent",
                    Namespace: "default",
                },
            }
            
            result, err := reconciler.Reconcile(ctx, req)
            Expect(err).NotTo(HaveOccurred())
            Expect(result).To(Equal(ctrl.Result{}))
        })
    })
})

