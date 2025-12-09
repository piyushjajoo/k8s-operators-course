# Lab 6.2: Writing Unit Tests

**Related Lesson:** [Lesson 6.2: Unit Testing with envtest](../lessons/02-unit-testing-envtest.md)  
**Navigation:** [← Previous Lab: Testing Fundamentals](lab-01-testing-fundamentals.md) | [Module Overview](../README.md) | [Next Lab: Integration Testing →](lab-03-integration-testing.md)

## Objectives

- Write unit tests for reconciliation logic
- Test resource creation and updates
- Test error cases
- Use table-driven tests
- Achieve good test coverage

## Prerequisites

- Completion of [Lab 6.1](lab-01-testing-fundamentals.md)
- Test environment set up
- Database operator ready

## Exercise 1: Test Resource Creation

### Task 1.1: Test StatefulSet Creation

Add to `internal/controller/database_controller_test.go`:

```go
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
})
```

## Exercise 2: Test Resource Updates

### Task 2.1: Test StatefulSet Update

```go
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
```

## Exercise 3: Test Error Cases

### Task 3.1: Test Missing Resource

```go
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
```

## Exercise 4: Table-Driven Tests

### Task 4.1: Test Multiple Scenarios

```go
var _ = Describe("Database validation", func() {
    tests := []struct {
        name    string
        db      *databasev1.Database
        wantErr bool
    }{
        {
            name: "valid database",
            db: &databasev1.Database{
                Spec: databasev1.DatabaseSpec{
                    Image:       "postgres:14",
                    DatabaseName: "mydb",
                    Username:    "admin",
                    Storage: databasev1.StorageSpec{
                        Size: "10Gi",
                    },
                },
            },
            wantErr: false,
        },
        {
            name: "missing image",
            db: &databasev1.Database{
                Spec: databasev1.DatabaseSpec{
                    DatabaseName: "mydb",
                    Username:    "admin",
                    Storage: databasev1.StorageSpec{
                        Size: "10Gi",
                    },
                },
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        It(tt.name, func() {
            err := k8sClient.Create(ctx, tt.db)
            if tt.wantErr {
                Expect(err).To(HaveOccurred())
            } else {
                Expect(err).NotTo(HaveOccurred())
            }
        })
    }
})
```

## Exercise 5: Test Coverage

### Task 5.1: Check Coverage

```bash
# Run tests with coverage
go test -coverprofile=coverage.out ./internal/controller/...

# View coverage
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Task 5.2: Improve Coverage

Add tests for:
- Service creation
- Status updates
- Error handling
- Edge cases

## Cleanup

```bash
# Tests should clean up automatically
# Verify no resources left
```

## Lab Summary

In this lab, you:
- Wrote unit tests for reconciliation
- Tested resource creation
- Tested resource updates
- Tested error cases
- Used table-driven tests
- Checked test coverage

## Key Learnings

1. Unit tests verify reconciliation logic
2. Test both success and error cases
3. Table-driven tests organize multiple scenarios
4. Coverage helps identify gaps
5. envtest provides real Kubernetes API
6. Gomega provides rich assertions

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Test Suite Setup](../solutions/suite_test.go) - Complete test suite with envtest
- [Unit Test Examples](../solutions/database_controller_test.go) - Complete unit test examples

## Next Steps

Now let's create integration tests for end-to-end scenarios!

**Navigation:** [← Previous Lab: Testing Fundamentals](lab-01-testing-fundamentals.md) | [Related Lesson](../lessons/02-unit-testing-envtest.md) | [Next Lab: Integration Testing →](lab-03-integration-testing.md)

