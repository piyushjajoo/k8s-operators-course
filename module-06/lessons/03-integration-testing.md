# Lesson 6.3: Integration Testing

**Navigation:** [← Previous: Unit Testing with envtest](02-unit-testing-envtest.md) | [Module Overview](../README.md) | [Next: Debugging and Observability →](04-debugging-observability.md)

## Introduction

While unit tests verify logic in isolation, integration tests verify that your operator works correctly with a real Kubernetes cluster. Integration tests use actual clusters (like kind) to test end-to-end workflows and ensure everything works together.

## Integration Testing Flow

Here's how integration tests work:

```mermaid
sequenceDiagram
    participant Test
    participant Cluster as Test Cluster
    participant Operator
    participant Resources
    
    Test->>Cluster: Create Cluster
    Cluster-->>Test: Ready
    Test->>Cluster: Deploy Operator
    Cluster->>Operator: Start
    Test->>Cluster: Create CustomResource
    Cluster->>Operator: Trigger Reconcile
    Operator->>Cluster: Create Resources
    Test->>Cluster: Verify Resources
    Cluster-->>Test: Results
    Test->>Cluster: Cleanup
```

## Integration Test Structure

### Using Ginkgo for Integration Tests

```go
package integration

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
    "sigs.k8s.io/controller-runtime/pkg/envtest"
)

var _ = Describe("Database Operator Integration", func() {
    var (
        cluster *kind.Cluster
        client  client.Client
    )
    
    BeforeSuite(func() {
        // Create kind cluster
        cluster = kind.NewCluster("test-cluster")
        Expect(cluster.Create()).To(Succeed())
        
        // Get kubeconfig
        kubeconfig := cluster.KubeconfigPath()
        cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        Expect(err).NotTo(HaveOccurred())
        
        // Create client
        client, err = client.New(cfg, client.Options{})
        Expect(err).NotTo(HaveOccurred())
    })
    
    AfterSuite(func() {
        // Cleanup cluster
        Expect(cluster.Delete()).To(Succeed())
    })
    
    Context("Database lifecycle", func() {
        It("should create and manage a Database", func() {
            // Test implementation
        })
    })
})
```

## Testing End-to-End Workflows

### Example: Complete Database Lifecycle

```go
Describe("Database lifecycle", func() {
    It("should create, update, and delete a Database", func() {
        // Create Database
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
        
        // Wait for StatefulSet
        Eventually(func() error {
            ss := &appsv1.StatefulSet{}
            return k8sClient.Get(ctx, types.NamespacedName{
                Name:      "test-db",
                Namespace: "default",
            }, ss)
        }, timeout, interval).Should(Succeed())
        
        // Verify StatefulSet is ready
        Eventually(func() bool {
            ss := &appsv1.StatefulSet{}
            k8sClient.Get(ctx, types.NamespacedName{
                Name:      "test-db",
                Namespace: "default",
            }, ss)
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
            return errors.IsNotFound(err)
        }, timeout, interval).Should(BeTrue())
    })
})
```

## Testing Webhooks

### Example: Testing Validating Webhook

```go
Describe("Validating webhook", func() {
    It("should reject invalid Database", func() {
        db := &databasev1.Database{
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
```

## Testing with Eventually

Gomega's `Eventually` is perfect for integration tests:

```go
// Wait for resource to be created
Eventually(func() error {
    return k8sClient.Get(ctx, key, resource)
}, timeout, interval).Should(Succeed())

// Wait for condition
Eventually(func() bool {
    db := &databasev1.Database{}
    k8sClient.Get(ctx, key, db)
    return db.Status.Ready
}, timeout, interval).Should(BeTrue())

// Wait for resource count
Eventually(func() int {
    pods := &corev1.PodList{}
    k8sClient.List(ctx, pods, client.MatchingLabels{"app": "database"})
    return len(pods.Items)
}, timeout, interval).Should(Equal(3))
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install kind
        run: |
          go install sigs.k8s.io/kind@latest
      
      - name: Create cluster
        run: kind create cluster
      
      - name: Run integration tests
        run: |
          make test-integration
      
      - name: Cleanup
        if: always()
        run: kind delete cluster
```

## Key Takeaways

- **Integration tests** verify end-to-end workflows
- **Use real clusters** (kind) for integration tests
- **Ginkgo/Gomega** provide structure and assertions
- **Eventually** waits for async operations
- **Test complete workflows** (create, update, delete)
- **Test webhooks** with real API calls
- **Integrate with CI/CD** for automated testing
- **Clean up** resources after tests

## Understanding for Building Operators

When writing integration tests:
- Use real Kubernetes clusters
- Test complete workflows
- Use Eventually for async operations
- Test webhook behavior
- Clean up resources
- Integrate with CI/CD
- Test error scenarios
- Verify resource states

## Related Lab

- [Lab 6.3: Creating Integration Tests](../labs/lab-03-integration-testing.md) - Hands-on exercises for this lesson

## References

### Official Documentation
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [kind Documentation](https://kind.sigs.k8s.io/)

### Further Reading
- **Kubernetes Operators** by Jason Dobies and Joshua Wood - Chapter 10: Testing
- **Programming Kubernetes** by Michael Hausenblas and Stefan Schimanski - Chapter 10: Testing
- [BDD Testing with Ginkgo](https://onsi.github.io/ginkgo/)

### Related Topics
- [Integration Testing Patterns](https://martinfowler.com/articles/practical-test-pyramid.html)
- [Test Containers](https://www.testcontainers.org/)
- [CI/CD Testing](https://kubernetes.io/docs/concepts/architecture/test/)

## Next Steps

Now that you understand integration testing, let's learn about debugging and observability.

**Navigation:** [← Previous: Unit Testing with envtest](02-unit-testing-envtest.md) | [Module Overview](../README.md) | [Next: Debugging and Observability →](04-debugging-observability.md)

