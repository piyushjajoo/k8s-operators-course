# Module 6 Solutions

This directory contains complete, working solutions for Module 6 labs.

## Files

### Testing (Labs 1-3)
- **suite_test.go**: Complete test suite setup with envtest (for `internal/controller/`)
- **database_controller_test.go**: Complete unit test examples
- **integration_test.go**: Complete integration test examples (combines suite + tests)

### Observability (Lab 4)
- **metrics.go**: Custom Prometheus metrics (copy to `internal/controller/metrics.go`)
- **observability.go**: Patterns for structured logging and event emission
- **metrics_reader_role_binding.yaml**: RBAC binding for metrics access (copy to `config/rbac/`)
- **rbac_kustomization.yaml**: Updated kustomization including metrics binding

## Usage

These solutions can be used as:
- Reference when writing your own tests
- Starting point if you get stuck
- Examples of testing best practices

## Integration

To use these solutions in your operator:

### 1. For unit tests (envtest)

Copy the suite and test files to your controller directory:

```bash
# Copy suite_test.go to internal/controller/suite_test.go
cp suite_test.go ~/postgres-operator/internal/controller/suite_test.go

# Copy test examples to internal/controller/database_controller_test.go
cp database_controller_test.go ~/postgres-operator/internal/controller/database_controller_test.go

# Run tests
cd ~/postgres-operator
make test
# Or: ginkgo -v ./internal/controller/...
```

### 2. For integration tests (real cluster)

Integration tests require:
- A running Kubernetes cluster (kind, minikube, etc.)
- Your operator deployed to the cluster
- **CRD types registered with the scheme** (critical!)

```bash
# Create integration test directory
mkdir -p ~/postgres-operator/test/integration

# Copy integration_test.go (contains both suite setup and tests)
# Split into two files for your project:

# 1. Create integration_suite_test.go with BeforeSuite (scheme registration)
# 2. Create database_test.go with the Describe blocks

# Ensure operator is deployed
make deploy IMG=<your-image>

# Run integration tests
ginkgo -v ./test/integration

# Skip webhook tests if webhooks aren't deployed
ginkgo -v -skip="webhook" ./test/integration
```

### 3. For observability

```bash
# Step 1: Add RBAC for metrics access
cp metrics_reader_role_binding.yaml ~/postgres-operator/config/rbac/metrics_reader_role_binding.yaml

# Step 2: Update config/rbac/kustomization.yaml to include the new file
# Add this line after 'metrics_reader_role.yaml':
#   - metrics_reader_role_binding.yaml
# (See rbac_kustomization.yaml for the complete file)

# Step 3: Copy metrics code to internal/controller/metrics.go
cp metrics.go ~/postgres-operator/internal/controller/metrics.go

# Step 4: Add event recorder to your controller struct (see observability.go)
# Step 5: Update Reconcile function with metrics and events

# Step 6: Redeploy the operator
cd ~/postgres-operator
make deploy IMG=<your-image>
```

## Key Points

### Scheme Registration (Integration Tests)

Integration tests **must** register custom types with the scheme:

```go
// In BeforeSuite
err := databasev1.AddToScheme(scheme.Scheme)
Expect(err).NotTo(HaveOccurred())

// Pass scheme to client
k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
```

Without this, you'll get: `no kind is registered for the type v1.Database`

### Use k8sClient.Scheme() in Unit Tests

When creating a reconciler in unit tests, use:

```go
reconciler := &DatabaseReconciler{
    Client: k8sClient,
    Scheme: k8sClient.Scheme(),  // NOT scheme.Scheme
}
```

### Pointer Helpers

Use `k8s.io/utils/ptr` for pointer helpers:

```go
import "k8s.io/utils/ptr"

Replicas: ptr.To(int32(1))
```

## Testing Commands

```bash
# Run unit tests (envtest)
cd ~/postgres-operator
make test

# Run unit tests with Ginkgo directly
ginkgo -v ./internal/controller/...

# Run integration tests (requires deployed operator)
ginkgo -v ./test/integration

# Skip webhook tests
ginkgo -v -skip="webhook" ./test/integration

# Run with coverage
go test -coverprofile=coverage.out ./internal/controller/...
go tool cover -html=coverage.out -o coverage.html

# Check metrics (after deploying operator)
kubectl port-forward -n postgres-operator-system deployment/postgres-operator-controller-manager 8080:8080
curl http://localhost:8080/metrics | grep database_
```

## Notes

- These are complete, working examples
- They follow best practices from the lessons
- Tests use Ginkgo/Gomega for BDD-style structure
- Unit tests use envtest for lightweight Kubernetes API
- Integration tests run against real clusters
- Metrics use Prometheus client library
- Logging uses structured logging (zap)
- Events use Kubernetes event recorder
