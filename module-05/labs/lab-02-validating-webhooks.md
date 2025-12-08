# Lab 5.2: Building Validating Webhook

**Related Lesson:** [Lesson 5.2: Implementing Validating Webhooks](../lessons/02-validating-webhooks.md)  
**Navigation:** [← Previous Lab: Admission Control](lab-01-admission-control.md) | [Module Overview](../README.md) | [Next Lab: Mutating Webhooks →](lab-03-mutating-webhooks.md)

## Objectives

- Scaffold validating webhook with kubebuilder
- Implement custom validation logic
- Test with valid and invalid resources
- Provide meaningful error messages

## Prerequisites

- Completion of [Module 3](../../module-03/README.md) or [Module 4](../../module-04/README.md)
- Database operator project
- Understanding of validation requirements

## Exercise 1: Scaffold Validating Webhook

### Task 1.1: Navigate to Your Operator

```bash
# Navigate to your Database operator
cd ~/postgres-operator
```

### Task 1.2: Create Validating Webhook

```bash
# Create validating webhook
kubebuilder create webhook \
  --group database \
  --version v1 \
  --kind Database \
  --programmatic-validation
```

**Observe:**
- What files were created?
- What was modified?

### Task 1.3: Examine Generated Code

```bash
# Check the generated webhook file
cat internal/webhook/v1/database_webhook.go

# Check webhook markers
grep "kubebuilder:webhook" internal/webhook/v1/database_webhook.go
```

**Observe the structure:**
- Webhook code is in `internal/webhook/v1/` directory
- Uses `DatabaseCustomValidator` struct
- Implements `webhook.CustomValidator` interface
- Methods take `context.Context` as first parameter

## Exercise 2: Implement Validation Logic

### Task 2.1: Add ValidateCreate

Edit `internal/webhook/v1/database_webhook.go`:

```go
package v1

import (
    "context"
    "fmt"
    "strings"

    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    logf "sigs.k8s.io/controller-runtime/pkg/log"
    "sigs.k8s.io/controller-runtime/pkg/webhook"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

    databasev1 "github.com/example/postgres-operator/api/v1"
)

var databaselog = logf.Log.WithName("database-resource")

// SetupDatabaseWebhookWithManager registers the webhook for Database in the manager.
func SetupDatabaseWebhookWithManager(mgr ctrl.Manager) error {
    return ctrl.NewWebhookManagedBy(mgr).For(&databasev1.Database{}).
        WithValidator(&DatabaseCustomValidator{}).
        Complete()
}

// +kubebuilder:webhook:path=/validate-database-example-com-v1-database,mutating=false,failurePolicy=fail,sideEffects=None,groups=database.example.com,resources=databases,verbs=create;update,versions=v1,name=vdatabase-v1.kb.io,admissionReviewVersions=v1

// DatabaseCustomValidator struct is responsible for validating the Database resource
// when it is created, updated, or deleted.
type DatabaseCustomValidator struct {
    // Add more fields as needed for validation
}

var _ webhook.CustomValidator = &DatabaseCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Database.
func (v *DatabaseCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
    database, ok := obj.(*databasev1.Database)
    if !ok {
        return nil, fmt.Errorf("expected a Database object but got %T", obj)
    }
    databaselog.Info("Validation for Database upon creation", "name", database.GetName())

    // Validate image is PostgreSQL
    if !strings.Contains(database.Spec.Image, "postgres") {
        return nil, fmt.Errorf("spec.image must be a PostgreSQL image, got %s", database.Spec.Image)
    }

    // Validate replicas and storage relationship
    if database.Spec.Replicas != nil && *database.Spec.Replicas > 5 {
        if database.Spec.Storage.Size == "10Gi" {
            return nil, fmt.Errorf("replicas > 5 requires storage >= 50Gi, got %s", database.Spec.Storage.Size)
        }
    }

    // Validate database name format
    if len(database.Spec.DatabaseName) > 63 {
        return nil, fmt.Errorf("spec.databaseName must be <= 63 characters, got %d", len(database.Spec.DatabaseName))
    }

    return nil, nil
}
```

### Task 2.2: Add ValidateUpdate

```go
// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Database.
func (v *DatabaseCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
    database, ok := newObj.(*databasev1.Database)
    if !ok {
        return nil, fmt.Errorf("expected a Database object for the newObj but got %T", newObj)
    }
    oldDB, ok := oldObj.(*databasev1.Database)
    if !ok {
        return nil, fmt.Errorf("expected a Database object for the oldObj but got %T", oldObj)
    }
    databaselog.Info("Validation for Database upon update", "name", database.GetName())

    // Prevent reducing storage size
    oldSize := parseStorageSize(oldDB.Spec.Storage.Size)
    newSize := parseStorageSize(database.Spec.Storage.Size)

    if newSize < oldSize {
        return nil, fmt.Errorf("cannot reduce storage from %s to %s", oldDB.Spec.Storage.Size, database.Spec.Storage.Size)
    }

    // Prevent changing database name
    if oldDB.Spec.DatabaseName != database.Spec.DatabaseName {
        return nil, fmt.Errorf("cannot change spec.databaseName from %s to %s", oldDB.Spec.DatabaseName, database.Spec.DatabaseName)
    }

    return nil, nil
}

// Helper function
func parseStorageSize(size string) int64 {
    // Simple parser for "10Gi" format
    // In production, use proper parsing
    if strings.HasSuffix(size, "Gi") {
        num := strings.TrimSuffix(size, "Gi")
        // Parse and convert to bytes (simplified)
        _ = num // Implement proper parsing
        return 0
    }
    return 0
}
```

### Task 2.3: Add ValidateDelete (Optional)

```go
// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Database.
func (v *DatabaseCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
    database, ok := obj.(*databasev1.Database)
    if !ok {
        return nil, fmt.Errorf("expected a Database object but got %T", obj)
    }
    databaselog.Info("Validation for Database upon deletion", "name", database.GetName())

    // Add any deletion validation logic
    // For example, prevent deletion if database has important data

    return nil, nil
}
```

## Exercise 3: Generate Manifests

### Task 3.1: Generate Webhook Manifests

```bash
# Generate manifests
make manifests

# Check webhook configuration was generated
ls -la config/webhook/

# Examine webhook configuration
cat config/webhook/manifests.yaml
```

### Task 3.2: Verify Webhook Configuration

```bash
# Check the configuration
cat config/webhook/manifests.yaml | grep -A 20 "ValidatingWebhookConfiguration"
```

## Exercise 4: Test Validating Webhook

### Understanding Webhook Testing

Unlike controller logic, webhooks cannot be easily tested with `make run` because:
- Webhooks require TLS certificates
- The Kubernetes API server (inside the cluster) needs to reach the webhook endpoint
- When running locally, the API server cannot call back to your localhost

**Two approaches for development:**

| Approach | Command | Webhooks Work? | Use When |
|----------|---------|----------------|----------|
| Local development | `make install && make run` | ❌ No | Testing controller/reconciliation logic |
| In-cluster deployment | `make deploy` | ✅ Yes | Testing webhook validation |

> **Note:** If you used the course's `scripts/setup-kind-cluster.sh` script to create your cluster, cert-manager is already installed. Verify with: `kubectl get pods -n cert-manager`

### Task 4.1: Ensure Cert-Manager is Installed

If cert-manager is not installed:

```bash
# Install cert-manager in your cluster
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

# Wait for cert-manager to be ready
kubectl wait --for=condition=Available deployment/cert-manager -n cert-manager --timeout=120s
kubectl wait --for=condition=Available deployment/cert-manager-webhook -n cert-manager --timeout=120s
kubectl wait --for=condition=Available deployment/cert-manager-cainjector -n cert-manager --timeout=120s
```

### Task 4.2: Deploy Operator to Cluster

Since webhooks need to run inside the cluster, we need to build and deploy:

```bash
# Build the container image
make docker-build IMG=postgres-operator:latest

# Load image into kind cluster
kind load docker-image postgres-operator:latest --name k8s-operators-course

# Deploy operator with webhooks to cluster
make deploy IMG=postgres-operator:latest
```

> **Using Podman instead of Docker?**
> 
> The Makefile uses `CONTAINER_TOOL` variable (defaults to `docker`). To use Podman:
> ```bash
> # Build with podman
> make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman
> 
> # Load image into kind - Option 1: Use KIND_EXPERIMENTAL_PROVIDER
> KIND_EXPERIMENTAL_PROVIDER=podman kind load docker-image postgres-operator:latest --name k8s-operators-course
> 
> # Load image into kind - Option 2: Save and load via tarball
> podman save postgres-operator:latest -o /tmp/postgres-operator.tar
> kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
> rm /tmp/postgres-operator.tar
> ```

> **Tip:** For day-to-day controller development, you can still use `make install && make run`. Only deploy to cluster when you need to test webhook behavior.

### Task 4.3: Verify Webhook is Registered

```bash
# Check webhook configuration was created
kubectl get validatingwebhookconfigurations

# Check operator pods are running
kubectl get pods -n postgres-operator-system

# Check logs if needed
kubectl logs -n postgres-operator-system deployment/postgres-operator-controller-manager
```

### Task 4.4: Test Valid Resource

```bash
# Create valid Database
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: valid-db
spec:
  image: postgres:14
  replicas: 1
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# Should succeed
kubectl get database valid-db
```

### Task 4.5: Test Invalid Resources

```bash
# Test invalid image
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: invalid-image
spec:
  image: nginx:latest  # Not PostgreSQL
  replicas: 1
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# Should fail with validation error

# Test invalid storage for replicas
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: invalid-storage
spec:
  image: postgres:14
  replicas: 10  # Too many replicas
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi  # Too small
EOF

# Should fail with validation error
```

### Task 4.6: Test Update Validation

```bash
# Create database
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: update-test
spec:
  image: postgres:14
  replicas: 1
  databaseName: mydb
  username: admin
  storage:
    size: 20Gi
EOF

# Try to reduce storage
kubectl patch database update-test --type merge -p '{"spec":{"storage":{"size":"10Gi"}}}'

# Should fail with validation error

# Try to change database name
kubectl patch database update-test --type merge -p '{"spec":{"databaseName":"newdb"}}'

# Should fail with validation error
```

## Exercise 5: Improve Error Messages

### Task 5.1: Add Context to Errors

Enhance error messages:

```go
func (v *DatabaseCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
    database, ok := obj.(*databasev1.Database)
    if !ok {
        return nil, fmt.Errorf("expected a Database object but got %T", obj)
    }
    databaselog.Info("Validation for Database upon creation", "name", database.GetName())

    var errors []string

    // Validate image
    if !strings.Contains(database.Spec.Image, "postgres") {
        errors = append(errors, fmt.Sprintf("spec.image: must be a PostgreSQL image, got '%s'. Valid examples: postgres:14, postgres:13", database.Spec.Image))
    }

    // Validate storage
    if database.Spec.Replicas != nil && *database.Spec.Replicas > 5 {
        if database.Spec.Storage.Size == "10Gi" {
            errors = append(errors, fmt.Sprintf("spec.storage.size: when replicas > 5, storage must be >= 50Gi, got '%s'", database.Spec.Storage.Size))
        }
    }

    if len(errors) > 0 {
        return nil, fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
    }

    return nil, nil
}
```

## Cleanup

```bash
# Delete test resources
kubectl delete databases --all

# Stop operator (Ctrl+C)
```

## Lab Summary

In this lab, you:
- Scaffolded validating webhook with kubebuilder
- Implemented custom validation logic
- Tested with valid and invalid resources
- Improved error messages
- Tested update validation

## Key Learnings

1. Kubebuilder scaffolds webhooks easily in `internal/webhook/v1/`
2. Uses `DatabaseCustomValidator` struct implementing `webhook.CustomValidator`
3. Methods receive `context.Context` as first parameter
4. `ValidateUpdate` receives both old and new objects as `runtime.Object`
5. Type-assert `runtime.Object` to your actual resource type
6. Provide clear, actionable error messages
7. Test with both valid and invalid resources
8. Webhooks run after CRD schema validation
9. Error messages help users fix issues

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Validating Webhook](../solutions/validating-webhook.go) - Complete validating webhook implementation with custom validation logic

## Next Steps

Now let's build a mutating webhook for defaulting!

**Navigation:** [← Previous Lab: Admission Control](lab-01-admission-control.md) | [Related Lesson](../lessons/02-validating-webhooks.md) | [Next Lab: Mutating Webhooks →](lab-03-mutating-webhooks.md)

