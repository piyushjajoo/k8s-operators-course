# Lab 5.2: Building Validating Webhook

**Related Lesson:** [Lesson 5.2: Implementing Validating Webhooks](../lessons/02-validating-webhooks.md)  
**Navigation:** [← Previous Lab: Admission Control](lab-01-admission-control.md) | [Module Overview](../README.md) | [Next Lab: Mutating Webhooks →](lab-03-mutating-webhooks.md)

## Objectives

- Scaffold validating webhook with kubebuilder
- Implement custom validation logic
- Test with valid and invalid resources
- Provide meaningful error messages

## Prerequisites

- Completion of [Module 3](../module-03/README.md) or [Module 4](../module-04/README.md)
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
# Check API types file
cat api/v1/database_types.go | grep -A 20 "webhook"

# Check webhook markers
cat api/v1/database_types.go | grep "kubebuilder:webhook"
```

## Exercise 2: Implement Validation Logic

### Task 2.1: Add ValidateCreate

Edit `api/v1/database_webhook.go`:

```go
package v1

import (
    "fmt"
    "strings"
    
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/runtime/schema"
    ctrl "sigs.k8s.io/controller-runtime"
    logf "sigs.k8s.io/controller-runtime/pkg/log"
    "sigs.k8s.io/controller-runtime/pkg/webhook"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var databaseLog = logf.Log.WithName("database-resource")

func (r *Database) SetupWebhookWithManager(mgr ctrl.Manager) error {
    return ctrl.NewWebhookManager().
        For(r).
        Complete()
}

// +kubebuilder:webhook:path=/validate-database-example-com-v1-database,mutating=false,failurePolicy=fail,sideEffects=None,groups=database.example.com,resources=databases,verbs=create;update,versions=v1,name=vdatabase.kb.io

var _ webhook.Validator = &Database{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateCreate() (admission.Warnings, error) {
    databaseLog.Info("validate create", "name", r.Name)
    
    // Validate image is PostgreSQL
    if !strings.Contains(r.Spec.Image, "postgres") {
        return nil, fmt.Errorf("spec.image must be a PostgreSQL image, got %s", r.Spec.Image)
    }
    
    // Validate replicas and storage relationship
    if r.Spec.Replicas != nil && *r.Spec.Replicas > 5 {
        if r.Spec.Storage.Size == "10Gi" {
            return nil, fmt.Errorf("replicas > 5 requires storage >= 50Gi, got %s", r.Spec.Storage.Size)
        }
    }
    
    // Validate database name format
    if len(r.Spec.DatabaseName) > 63 {
        return nil, fmt.Errorf("spec.databaseName must be <= 63 characters, got %d", len(r.Spec.DatabaseName))
    }
    
    return nil, nil
}
```

### Task 2.2: Add ValidateUpdate

```go
// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
    databaseLog.Info("validate update", "name", r.Name)
    
    oldDB := old.(*Database)
    
    // Prevent reducing storage size
    oldSize := parseStorageSize(oldDB.Spec.Storage.Size)
    newSize := parseStorageSize(r.Spec.Storage.Size)
    
    if newSize < oldSize {
        return nil, fmt.Errorf("cannot reduce storage from %s to %s", oldDB.Spec.Storage.Size, r.Spec.Storage.Size)
    }
    
    // Prevent changing database name
    if oldDB.Spec.DatabaseName != r.Spec.DatabaseName {
        return nil, fmt.Errorf("cannot change spec.databaseName from %s to %s", oldDB.Spec.DatabaseName, r.Spec.DatabaseName)
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
        return 0 // Implement proper parsing
    }
    return 0
}
```

### Task 2.3: Add ValidateDelete (Optional)

```go
// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateDelete() (admission.Warnings, error) {
    databaseLog.Info("validate delete", "name", r.Name)
    
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

### Task 4.1: Generate Certificates

```bash
# Generate certificates for local development
make certs

# Install certificates
make install-cert
```

### Task 4.2: Run Operator with Webhook

```bash
# Run operator (webhook runs in same process)
make run
```

### Task 4.3: Test Valid Resource

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

### Task 4.4: Test Invalid Resources

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

### Task 4.5: Test Update Validation

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
func (r *Database) ValidateCreate() (admission.Warnings, error) {
    var errors []string
    
    // Validate image
    if !strings.Contains(r.Spec.Image, "postgres") {
        errors = append(errors, fmt.Sprintf("spec.image: must be a PostgreSQL image, got '%s'. Valid examples: postgres:14, postgres:13", r.Spec.Image))
    }
    
    // Validate storage
    if r.Spec.Replicas != nil && *r.Spec.Replicas > 5 {
        if r.Spec.Storage.Size == "10Gi" {
            errors = append(errors, fmt.Sprintf("spec.storage.size: when replicas > 5, storage must be >= 50Gi, got '%s'", r.Spec.Storage.Size))
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

1. Kubebuilder scaffolds webhooks easily
2. ValidateCreate, ValidateUpdate, ValidateDelete methods
3. Provide clear, actionable error messages
4. Test with both valid and invalid resources
5. Webhooks run after CRD schema validation
6. Error messages help users fix issues

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Validating Webhook](../solutions/validating-webhook.go) - Complete validating webhook implementation with custom validation logic

## Next Steps

Now let's build a mutating webhook for defaulting!

**Navigation:** [← Previous Lab: Admission Control](lab-01-admission-control.md) | [Related Lesson](../lessons/02-validating-webhooks.md) | [Next Lab: Mutating Webhooks →](lab-03-mutating-webhooks.md)

