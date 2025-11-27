# Lab 6.1: Setting Up Testing Environment

**Related Lesson:** [Lesson 6.1: Testing Fundamentals](../lessons/01-testing-fundamentals.md)  
**Navigation:** [Module Overview](../README.md) | [Next Lab: Unit Testing →](lab-02-unit-testing-envtest.md)

## Objectives

- Set up testing tools and dependencies
- Understand testing structure
- Create test scaffolding
- Prepare for writing tests

## Prerequisites

- Completion of [Module 5](../module-05/README.md)
- Database operator from Module 3/4/5
- Go 1.21+ installed
- Understanding of Go testing

## Exercise 1: Install Testing Tools

### Task 1.1: Install Ginkgo and Gomega

```bash
# Install Ginkgo
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Install Gomega
go get github.com/onsi/gomega/...

# Verify installation
ginkgo version
```

### Task 1.2: Install envtest Tools

```bash
# Install setup-envtest
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

# Download envtest binaries
setup-envtest use

# Verify
setup-envtest list
```

### Task 1.3: Install Delve Debugger

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Verify installation
dlv version
```

## Exercise 2: Set Up Test Structure

### Task 2.1: Create Test Directory

```bash
# Navigate to your operator
cd ~/postgres-operator

# Create test directory structure
mkdir -p controllers/suite_test
```

### Task 2.2: Initialize Ginkgo Suite

```bash
# Initialize Ginkgo suite
cd controllers
ginkgo bootstrap
```

### Task 2.3: Create Suite Test File

Create `controllers/suite_test/suite_test.go`:

```go
package suite_test

import (
    "path/filepath"
    "testing"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
    "k8s.io/client-go/kubernetes/scheme"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/envtest"
    
    databasev1 "github.com/example/postgres-operator/api/v1"
)

var (
    k8sClient client.Client
    testEnv   *envtest.Environment
)

func TestControllers(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
    By("bootstrapping test environment")
    testEnv = &envtest.Environment{
        CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
        ErrorIfCRDPathMissing: true,
    }
    
    cfg, err := testEnv.Start()
    Expect(err).NotTo(HaveOccurred())
    Expect(cfg).NotTo(BeNil())
    
    err = databasev1.AddToScheme(scheme.Scheme)
    Expect(err).NotTo(HaveOccurred())
    
    k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
    Expect(err).NotTo(HaveOccurred())
    Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
    By("tearing down the test environment")
    err := testEnv.Stop()
    Expect(err).NotTo(HaveOccurred())
})
```

## Exercise 3: Create First Test

### Task 3.1: Create Database Controller Test

Create `controllers/database_controller_test.go`:

```go
package controller

import (
    "context"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/utils/pointer"
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    databasev1 "github.com/example/postgres-operator/api/v1"
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
    
    Context("When creating a Database", func() {
        It("should be created successfully", func() {
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
            
            // Verify it was created
            created := &databasev1.Database{}
            Expect(k8sClient.Get(ctx, client.ObjectKey{
                Name:      "test-db",
                Namespace: "default",
            }, created)).To(Succeed())
            
            Expect(created.Spec.Image).To(Equal("postgres:14"))
        })
    })
})
```

## Exercise 4: Run Tests

### Task 4.1: Run Tests

```bash
# Run all tests
go test ./controllers/...

# Run with Ginkgo
ginkgo -v ./controllers

# Run specific test
ginkgo -v -focus="DatabaseReconciler" ./controllers
```

### Task 4.2: Check Test Coverage

```bash
# Run with coverage
go test -cover ./controllers/...

# Generate coverage report
go test -coverprofile=coverage.out ./controllers/...
go tool cover -html=coverage.out
```

## Exercise 5: Verify Setup

### Task 5.1: Verify All Tools

```bash
# Check Ginkgo
ginkgo version

# Check envtest
setup-envtest list

# Check Delve
dlv version

# Check Go
go version
```

## Cleanup

```bash
# Clean up test resources (if any)
# Tests should clean up automatically
```

## Lab Summary

In this lab, you:
- Installed testing tools (Ginkgo, Gomega, envtest, Delve)
- Set up test structure
- Created test suite
- Created first test
- Ran tests and checked coverage

## Key Learnings

1. Ginkgo provides BDD-style test structure
2. envtest provides lightweight Kubernetes API
3. Test suite setup in BeforeSuite/AfterSuite
4. Tests use Gomega for assertions
5. Coverage helps identify untested code
6. Proper test structure is important

## Next Steps

Now let's write comprehensive unit tests for your operator!

**Navigation:** [← Module Overview](../README.md) | [Related Lesson](../lessons/01-testing-fundamentals.md) | [Next Lab: Unit Testing →](lab-02-unit-testing-envtest.md)

