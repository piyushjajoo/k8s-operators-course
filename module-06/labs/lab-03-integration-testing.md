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

Create `test/integration/integration_suite_test.go`:

**Important**: The client needs to know about your custom `Database` type. You must register it with the scheme!

```go
package integration_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
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

	// Register the Database type with the scheme
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
```

**Why is scheme registration needed?**
- The Kubernetes client uses the scheme to convert Go types to/from JSON/YAML
- Built-in types (Pod, Service, etc.) are already registered
- Custom Resource types like `Database` must be explicitly registered

## Exercise 2: Write End-to-End Test

### Task 2.1: Test Database Lifecycle

Create `test/integration/database_test.go`:

**Note**: The package must match the suite file (`integration_test`).

```go
package integration_test

import (
    "context"
	"fmt"
    "time"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    databasev1 "github.com/example/postgres-operator/api/v1"
)

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
})
```

**Key features:**
- Uses unique resource names to avoid test conflicts
- Proper cleanup in `AfterEach` (removes finalizers before deletion)
- Tests full lifecycle: create → update → delete
- Tests scaling (replicas 1 → 3)
- Tests child resource creation (StatefulSet, Service, Secret)

## Exercise 3: Test Webhooks (Optional)

**Note**: Webhook tests require webhooks to be deployed and configured with cert-manager. If you haven't set up webhooks, skip this exercise.

### Task 3.1: Test Validating Webhook

Add to `test/integration/database_test.go` (inside the main Describe block):

```go
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
			Expect(k8sClient.Delete(ctx, db)).To(Succeed())
    })
})
```

**Note**: If webhooks aren't deployed, the "reject invalid" test will fail because the validation only happens in the webhook. You can skip webhook tests by using:

```bash
ginkgo -v -skip="webhook" ./test/integration
```

## Exercise 4: Run Integration Tests

### Task 4.1: Run Tests Locally

```bash
# Ensure kind cluster is running, if not use ./scripts/setup-kind-cluster.sh
kind get clusters

# Deploy
# For Docker:
make deploy IMG=postgres-operator:latest

# For Podman:
make deploy IMG=localhost/postgres-operator:latest

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

env:
  IMG: postgres-operator:ci

jobs:
  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Install kind
        run: |
          go install sigs.k8s.io/kind@latest
      
      - name: Install ginkgo
        run: |
          go install github.com/onsi/ginkgo/v2/ginkgo@latest
      
      - name: Create cluster
        run: kind create cluster --image kindest/node:v1.32.0 --wait 60s
      
      - name: Build Docker image
        run: make docker-build IMG=${{ env.IMG }}
      
      - name: Load image into kind
        run: kind load docker-image ${{ env.IMG }}
      
      - name: Deploy operator (includes cert-manager)
        run: make deploy IMG=${{ env.IMG }}
      
      - name: Wait for cert-manager
        run: |
          kubectl wait --for=condition=Available deployment/cert-manager -n cert-manager --timeout=120s
          kubectl wait --for=condition=Available deployment/cert-manager-webhook -n cert-manager --timeout=120s
      
      - name: Wait for operator
        run: |
          kubectl wait --for=condition=ready pod -l control-plane=controller-manager -n postgres-operator-system --timeout=120s
      
      - name: Run integration tests
        run: |
          ginkgo -v ./test/integration
      
      - name: Debug on failure
        if: failure()
        run: |
          echo "=== Pods in all namespaces ==="
          kubectl get pods -A
          echo "=== Operator logs ==="
          kubectl logs -n postgres-operator-system -l control-plane=controller-manager --tail=100 || true
          echo "=== Events ==="
          kubectl get events -n postgres-operator-system --sort-by='.lastTimestamp' || true
      
      - name: Cleanup
        if: always()
        run: kind delete cluster
```

**Key points:**
- **Build image first** - `make docker-build` creates the container image
- **Load into kind** - `kind load docker-image` makes the image available to the cluster
- **Include namespace** - `-n postgres-operator-system` in kubectl wait
- **Wait for cert-manager** - Cert-manager must be ready before the operator can start (webhooks need TLS certs)
- **Debug on failure** - Logs help diagnose issues

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

1. **Register custom types with scheme** - The k8s client must know about your CRD types via `databasev1.AddToScheme(scheme.Scheme)`
2. **Pass scheme to client** - Use `client.Options{Scheme: scheme.Scheme}` when creating the client
3. **Integration tests use real clusters** - Tests run against actual Kubernetes API (kind, minikube, etc.)
4. **Eventually waits for async operations** - Controllers are async; use `Eventually` for assertions
5. **Test complete workflows** - Create → Update → Delete lifecycle
6. **Webhooks require deployment** - Webhook tests only work when webhooks are deployed with cert-manager
7. **CI/CD automates testing** - Use GitHub Actions or similar for automated testing
8. **Clean up resources after tests** - Delete created resources to avoid test pollution

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Integration Test Examples](../solutions/integration_test.go) - Complete integration test examples

## Next Steps

Now let's add observability and learn debugging techniques!

**Navigation:** [← Previous Lab: Unit Testing](lab-02-unit-testing-envtest.md) | [Related Lesson](../lessons/03-integration-testing.md) | [Next Lab: Observability →](lab-04-debugging-observability.md)

