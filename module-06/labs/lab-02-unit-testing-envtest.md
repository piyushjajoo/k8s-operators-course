# Lab 6.2: Writing Unit Tests

**Related Lesson:** [Lesson 6.2: Unit Testing with envtest](../lessons/02-unit-testing-envtest.md)  
**Navigation:** [← Previous Lab: Testing Fundamentals](lab-01-testing-fundamentals.md) | [Module Overview](../README.md) | [Next Lab: Integration Testing →](lab-03-integration-testing.md)

## Objectives

- Write unit tests for reconciliation logic
- Test resource creation and updates
- Test error cases
- Understand state machine testing patterns
- Achieve good test coverage

## Prerequisites

- Completion of [Lab 6.1](lab-01-testing-fundamentals.md)
- Test environment set up
- Database operator ready

## Understanding the Controller

Before writing tests, understand that the `DatabaseReconciler` uses a **state machine pattern** with phases:
- `Pending` → `Provisioning` → `Configuring` → `Deploying` → `Verifying` → `Ready`

Each `Reconcile()` call advances the state by one phase. This means multiple reconcile calls are needed to fully provision a database.

## Exercise 1: Test Basic Reconciliation

### Task 1.1: Test Initial State Transition

Update `internal/controller/database_controller_test.go` to add a new test Context. Note how we use unique resource names with `GenerateName` to avoid conflicts between tests:

```go
Context("When reconciling a new Database", func() {
    var (
        resourceName      string
        typeNamespacedName types.NamespacedName
    )

    BeforeEach(func() {
        // Generate unique name for each test
        resourceName = fmt.Sprintf("test-db-%d", time.Now().UnixNano())
        typeNamespacedName = types.NamespacedName{
            Name:      resourceName,
            Namespace: "default",
        }

        // Create the Database resource
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
        // Cleanup
        resource := &databasev1.Database{}
        err := k8sClient.Get(ctx, typeNamespacedName, resource)
        if err == nil {
            // Remove finalizer to allow deletion
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

        // First reconcile: Pending -> Provisioning
        _, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
            NamespacedName: typeNamespacedName,
        })
        Expect(err).NotTo(HaveOccurred())

        // Verify status was updated
        db := &databasev1.Database{}
        Expect(k8sClient.Get(ctx, typeNamespacedName, db)).To(Succeed())
        Expect(db.Status.Phase).To(Equal("Provisioning"))
        Expect(db.Status.Ready).To(BeFalse())
    })
})
```

**Required imports** (add to your import block):

```go
import (
    "context"
    "fmt"
    "time"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/utils/ptr"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"

    databasev1 "github.com/example/postgres-operator/api/v1"
)
```

## Exercise 2: Test Resource Creation Through State Machine

### Task 2.1: Test StatefulSet Creation

The StatefulSet is created during the `Provisioning` phase. Test this by running multiple reconcile calls:

```go
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
```

## Exercise 3: Test Error Cases

### Task 3.1: Test Missing Resource

```go
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
```

### Task 3.2: Test Finalizer Addition

```go
var _ = Describe("Database validation", func() {
    var (
        ctx               context.Context
        typeNamespacedName types.NamespacedName
    )

    BeforeEach(func() {
        ctx = context.Background()
        typeNamespacedName = types.NamespacedName{
            Name:      "test-database",
            Namespace: "default",
        }
        
        // Create the database resource
        resource := &databasev1.Database{
            ObjectMeta: metav1.ObjectMeta{
                Name:      typeNamespacedName.Name,
                Namespace: typeNamespacedName.Namespace,
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
```

## Exercise 4: Test Service Creation

### Task 4.1: Test Service Creation in Configuring Phase

```go
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
```

## Exercise 5: Test Coverage

### Task 5.1: Check Coverage

```bash
# Run tests with coverage
make test

# Or run with coverage profile
go test -coverprofile=coverage.out ./internal/controller/...

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
```

### Task 5.2: Improve Coverage

Add tests for:
- Deletion handling with finalizer cleanup
- Status condition updates
- Different replica counts
- Image changes

## Cleanup

The `AfterEach` blocks in each test Context handle cleanup automatically by:
1. Removing finalizers (to allow deletion)
2. Deleting the test Database resource

## Lab Summary

In this lab, you:
- Wrote unit tests following the Kubebuilder scaffolding pattern
- Tested state machine transitions
- Tested resource creation (Secret, StatefulSet, Service)
- Tested error cases (missing resources)
- Tested finalizer addition
- Checked test coverage

## Key Learnings

1. **State machine testing** - Controllers with phases need multiple reconcile calls
2. **Use unique resource names** - Avoid test conflicts with unique names per test
3. **Proper cleanup** - Remove finalizers before deletion in `AfterEach`
4. **Use `k8sClient.Scheme()`** - Not `scheme.Scheme` for reconciler initialization
5. **Use `reconcile.Request`** - The standard type for test requests
6. **Use `k8s.io/utils/ptr`** - For pointer helpers like `ptr.To(int32(1))`
7. **envtest provides real API** - Tests run against actual Kubernetes API server

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Test Suite Setup](../solutions/suite_test.go) - Complete test suite with envtest
- [Unit Test Examples](../solutions/database_controller_test.go) - Basic controller test structure

## Next Steps

Now let's create integration tests for end-to-end scenarios!

**Navigation:** [← Previous Lab: Testing Fundamentals](lab-01-testing-fundamentals.md) | [Related Lesson](../lessons/02-unit-testing-envtest.md) | [Next Lab: Integration Testing →](lab-03-integration-testing.md)

