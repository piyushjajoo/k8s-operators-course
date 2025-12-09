# Lab 6.1: Setting Up Testing Environment

**Related Lesson:** [Lesson 6.1: Testing Fundamentals](../lessons/01-testing-fundamentals.md)  
**Navigation:** [Module Overview](../README.md) | [Next Lab: Unit Testing →](lab-02-unit-testing-envtest.md)

## Objectives

- Set up testing tools and dependencies
- Understand testing structure
- Create test scaffolding
- Prepare for writing tests

## Prerequisites

- Completion of [Module 5](../../module-05/README.md)
- Database operator from Module 3/4/5
- Go 1.24+ installed
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

### Task 2.1: Navigate to Your Operator

```bash
# Navigate to your operator
cd ~/postgres-operator
```

When you run `kubebuilder create api` with `--resource --controller`, Kubebuilder automatically generates test scaffolding files in `internal/controller/`:
- `suite_test.go` - Test suite setup with envtest
- `<resource>_controller_test.go` - Basic controller test

### Task 2.2: Examine the Generated Suite Test File

The generated `internal/controller/suite_test.go` follows this structure:

```go
package controller

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "k8s.io/client-go/kubernetes/scheme"
    "k8s.io/client-go/rest"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/envtest"
    logf "sigs.k8s.io/controller-runtime/pkg/log"
    "sigs.k8s.io/controller-runtime/pkg/log/zap"

    databasev1 "github.com/example/postgres-operator/api/v1"
    // +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
    ctx       context.Context
    cancel    context.CancelFunc
    testEnv   *envtest.Environment
    cfg       *rest.Config
    k8sClient client.Client
)

func TestControllers(t *testing.T) {
    RegisterFailHandler(Fail)

    RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
    logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

    ctx, cancel = context.WithCancel(context.TODO())

    var err error
    err = databasev1.AddToScheme(scheme.Scheme)
    Expect(err).NotTo(HaveOccurred())

    // +kubebuilder:scaffold:scheme

    By("bootstrapping test environment")
    testEnv = &envtest.Environment{
        CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
        ErrorIfCRDPathMissing: true,
    }

    // Retrieve the first found binary directory to allow running tests from IDEs
    if getFirstFoundEnvTestBinaryDir() != "" {
        testEnv.BinaryAssetsDirectory = getFirstFoundEnvTestBinaryDir()
    }

    // cfg is defined in this file globally.
    cfg, err = testEnv.Start()
    Expect(err).NotTo(HaveOccurred())
    Expect(cfg).NotTo(BeNil())

    k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
    Expect(err).NotTo(HaveOccurred())
    Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
    By("tearing down the test environment")
    cancel()
    err := testEnv.Stop()
    Expect(err).NotTo(HaveOccurred())
})

// getFirstFoundEnvTestBinaryDir locates the first binary in the specified path.
// ENVTEST-based tests depend on specific binaries, usually located in paths set by
// controller-runtime. When running tests directly (e.g., via an IDE) without using
// Makefile targets, the 'BinaryAssetsDirectory' must be explicitly configured.
//
// This function streamlines the process by finding the required binaries, similar to
// setting the 'KUBEBUILDER_ASSETS' environment variable. To ensure the binaries are
// properly set up, run 'make setup-envtest' beforehand.
func getFirstFoundEnvTestBinaryDir() string {
    basePath := filepath.Join("..", "..", "bin", "k8s")
    entries, err := os.ReadDir(basePath)
    if err != nil {
        logf.Log.Error(err, "Failed to read directory", "path", basePath)
        return ""
    }
    for _, entry := range entries {
        if entry.IsDir() {
            return filepath.Join(basePath, entry.Name())
        }
    }
    return ""
}
```

**Key features of the generated suite:**
- **Package-level context**: `ctx` and `cancel` are available to all tests
- **IDE support**: `getFirstFoundEnvTestBinaryDir()` locates envtest binaries for IDE execution
- **Logging**: Configured with zap logger writing to GinkgoWriter
- **Scaffold markers**: `// +kubebuilder:scaffold:imports` and `// +kubebuilder:scaffold:scheme` for future API additions

## Exercise 3: Examine the Generated Controller Test

### Task 3.1: Understand the Scaffolded Test Structure

The generated `internal/controller/database_controller_test.go` follows this structure:

```go
package controller

import (
    "context"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/types"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    databasev1 "github.com/example/postgres-operator/api/v1"
)

var _ = Describe("Database Controller", func() {
    Context("When reconciling a resource", func() {
        const resourceName = "test-resource"

        ctx := context.Background()

        typeNamespacedName := types.NamespacedName{
            Name:      resourceName,
            Namespace: "default", // TODO(user):Modify as needed
        }
        database := &databasev1.Database{}

        BeforeEach(func() {
            By("creating the custom resource for the Kind Database")
            err := k8sClient.Get(ctx, typeNamespacedName, database)
            if err != nil && errors.IsNotFound(err) {
                resource := &databasev1.Database{
                    ObjectMeta: metav1.ObjectMeta{
                        Name:      resourceName,
                        Namespace: "default",
                    },
                    // TODO(user): Specify other spec details if needed.
                }
                Expect(k8sClient.Create(ctx, resource)).To(Succeed())
            }
        })

        AfterEach(func() {
            // TODO(user): Cleanup logic after each test, like removing the resource instance.
            resource := &databasev1.Database{}
            err := k8sClient.Get(ctx, typeNamespacedName, resource)
            Expect(err).NotTo(HaveOccurred())

            By("Cleanup the specific resource instance Database")
            Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
        })
        
        It("should successfully reconcile the resource", func() {
            By("Reconciling the created resource")
            controllerReconciler := &DatabaseReconciler{
                Client: k8sClient,
                Scheme: k8sClient.Scheme(),
            }

            _, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
                NamespacedName: typeNamespacedName,
            })
            Expect(err).NotTo(HaveOccurred())
            // TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
            // Example: If you expect a certain status condition after reconciliation, verify it here.
        })
    })
})
```

**Key features of the generated test:**
- **Resource setup/cleanup**: `BeforeEach` creates the resource, `AfterEach` deletes it
- **Direct reconciler invocation**: Creates `DatabaseReconciler` and calls `Reconcile()` directly
- **NamespacedName pattern**: Uses `types.NamespacedName` for resource identification
- **TODO markers**: Indicates where to customize for your specific controller
- **Uses package-level variables**: Accesses `k8sClient` from `suite_test.go`

## Exercise 4: Run Tests

### Task 4.1: Run Tests

```bash
# Setup envtest binaries first
make setup-envtest

# Run all tests using make (recommended)
make test

# Or run tests directly with go test
go test ./internal/controller/...

# Run with Ginkgo (verbose)
ginkgo -v ./internal/controller/...

# Run specific test
ginkgo -v -focus="Database Controller" ./internal/controller/...
```

### Task 4.2: Check Test Coverage

```bash
# Run with coverage
go test -cover ./internal/controller/...

# Generate coverage report
go test -coverprofile=coverage.out ./internal/controller/...
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
- Examined Kubebuilder-generated test scaffolding structure
- Understood the test suite setup with envtest
- Examined the controller test pattern
- Ran tests and checked coverage

## Key Learnings

1. **Kubebuilder generates test scaffolding** - When you create an API with `--controller`, test files are auto-generated
2. **Ginkgo provides BDD-style test structure** - Describe/Context/It blocks organize tests
3. **envtest provides lightweight Kubernetes API** - No full cluster needed for controller tests
4. **Suite setup in BeforeSuite/AfterSuite** - Environment initialized once per test suite
5. **Package-level variables** - `ctx`, `k8sClient`, `cfg` are shared across tests
6. **IDE support built-in** - `getFirstFoundEnvTestBinaryDir()` enables running tests from IDEs
7. **Direct reconciler invocation** - Tests call `Reconcile()` directly for deterministic results
8. **Scaffold markers** - `// +kubebuilder:scaffold:*` comments allow future API additions

## Solutions

The test suite setup from this lab matches the Kubebuilder-generated scaffolding:
- [Test Suite Setup](../solutions/suite_test.go) - Complete test suite with envtest configuration
- [Controller Test](../solutions/database_controller_test.go) - Basic controller test structure

## Next Steps

Now let's write comprehensive unit tests for your operator!

**Navigation:** [← Module Overview](../README.md) | [Related Lesson](../lessons/01-testing-fundamentals.md) | [Next Lab: Unit Testing →](lab-02-unit-testing-envtest.md)

