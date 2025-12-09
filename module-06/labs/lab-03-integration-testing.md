# Lab 6.3: Creating Integration Tests

**Related Lesson:** [Lesson 6.3: Integration Testing](../lessons/03-integration-testing.md)  
**Navigation:** [← Previous Lab: Unit Testing](lab-02-unit-testing-envtest.md) | [Module Overview](../README.md) | [Next Lab: Observability →](lab-04-debugging-observability.md)

## Objectives

- Set up integration test environment
- Write end-to-end tests
- Test complete workflows
- Integrate with CI/CD

## Prerequisites

- Completion of [Lab 6.2](lab-02-unit-testing-envtest.md)
- kind installed
- Understanding of integration testing

## Exercise 1: Set Up Integration Test Environment

### Task 1.1: Create Integration Test Directory

```bash
# Create integration test directory
mkdir -p test/integration
cd test/integration
```

### Task 1.2: Initialize Ginkgo Suite

```bash
# Initialize Ginkgo suite
ginkgo bootstrap
```

### Task 1.3: Create Suite Test

Create `test/integration/suite_test.go`:

```go
package integration_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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
```

## Exercise 2: Write End-to-End Test

### Task 2.1: Test Database Lifecycle

Create `test/integration/database_test.go`:

```go
package integration

import (
    "context"
    "time"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/utils/pointer"
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    databasev1 "github.com/example/postgres-operator/api/v1"
    appsv1 "k8s.io/api/apps/v1"
)

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
})
```

## Exercise 3: Test Webhooks

### Task 3.1: Test Validating Webhook

```go
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
```

## Exercise 4: Run Integration Tests

### Task 4.1: Run Tests Locally

```bash
# Ensure kind cluster is running
kind get clusters

# Deploy operator to cluster
make deploy

# Run integration tests
ginkgo -v ./test/integration
```

### Task 4.2: Run with Focus

```bash
# Run specific test
ginkgo -v -focus="Database lifecycle" ./test/integration
```

## Exercise 5: CI/CD Integration

### Task 5.1: Create GitHub Actions Workflow

Create `.github/workflows/integration-tests.yml`:

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
      
      - name: Install CRDs
        run: make install
      
      - name: Deploy operator
        run: make deploy
      
      - name: Wait for operator
        run: |
          kubectl wait --for=condition=ready pod -l control-plane=controller-manager --timeout=300s
      
      - name: Run integration tests
        run: |
          ginkgo -v ./test/integration
      
      - name: Cleanup
        if: always()
        run: kind delete cluster
```

## Cleanup

```bash
# Clean up test resources
kubectl delete databases --all

# Clean up cluster (if needed)
kind delete cluster
```

## Lab Summary

In this lab, you:
- Set up integration test environment
- Wrote end-to-end tests
- Tested complete workflows
- Tested webhooks
- Integrated with CI/CD

## Key Learnings

1. Integration tests use real clusters
2. Eventually waits for async operations
3. Test complete workflows
4. Webhooks can be tested with real API
5. CI/CD automates testing
6. Clean up resources after tests

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Integration Test Examples](../solutions/integration_test.go) - Complete integration test examples

## Next Steps

Now let's add observability and learn debugging techniques!

**Navigation:** [← Previous Lab: Unit Testing](lab-02-unit-testing-envtest.md) | [Related Lesson](../lessons/03-integration-testing.md) | [Next Lab: Observability →](lab-04-debugging-observability.md)

