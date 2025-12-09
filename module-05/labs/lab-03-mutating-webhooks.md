# Lab 5.3: Building Mutating Webhook

**Related Lesson:** [Lesson 5.3: Implementing Mutating Webhooks](../lessons/03-mutating-webhooks.md)  
**Navigation:** [← Previous Lab: Validating Webhooks](lab-02-validating-webhooks.md) | [Module Overview](../README.md) | [Next Lab: Webhook Deployment →](lab-04-webhook-deployment.md)

## Objectives

- Add mutating webhook to existing validating webhook
- Implement defaulting logic
- Test mutation scenarios
- Ensure idempotency

## Prerequisites

- Completion of [Lab 5.2](lab-02-validating-webhooks.md)
- Database operator with validating webhook
- Understanding of defaulting patterns

## Exercise 1: Add Mutating Webhook

Since we already created a validating webhook in Lab 5.2, our webhook file already exists at `internal/webhook/v1/database_webhook.go`. We'll add the mutating (defaulting) logic to this file.

> **Note:** If you were starting fresh, you would run:
> ```bash
> kubebuilder create webhook --group database --version v1 --kind Database --defaulting
> ```
> But since we already have a webhook, we'll add the defaulter manually.

### Task 1.1: Understand the CustomDefaulter Interface

The new kubebuilder pattern uses `webhook.CustomDefaulter` interface:

```go
type CustomDefaulter interface {
    Default(ctx context.Context, obj runtime.Object) error
}
```

### Task 1.2: Add Defaulter to Webhook Setup

Edit `internal/webhook/v1/database_webhook.go` to update the webhook setup function:

```go
// SetupDatabaseWebhookWithManager registers the webhook for Database in the manager.
func SetupDatabaseWebhookWithManager(mgr ctrl.Manager) error {
    return ctrl.NewWebhookManagedBy(mgr).For(&databasev1.Database{}).
        WithValidator(&DatabaseCustomValidator{}).
        WithDefaulter(&DatabaseCustomDefaulter{}).
        Complete()
}
```

### Task 1.3: Add the Defaulter Struct and Marker

Add the following to `internal/webhook/v1/database_webhook.go`:

```go
// +kubebuilder:webhook:path=/mutate-database-example-com-v1-database,mutating=true,failurePolicy=fail,sideEffects=None,groups=database.example.com,resources=databases,verbs=create;update,versions=v1,name=mdatabase-v1.kb.io,admissionReviewVersions=v1

// DatabaseCustomDefaulter struct is responsible for setting default values on the Database resource.
type DatabaseCustomDefaulter struct {
    // Add fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &DatabaseCustomDefaulter{}
```

## Exercise 2: Implement Defaulting Logic

### Task 2.1: Add Default Method

Add the `Default` method to `internal/webhook/v1/database_webhook.go`:

```go
// Default implements webhook.CustomDefaulter so a webhook will be registered for the type Database.
func (d *DatabaseCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
    database, ok := obj.(*databasev1.Database)
    if !ok {
        return fmt.Errorf("expected a Database object but got %T", obj)
    }
    databaselog.Info("Defaulting for Database", "name", database.GetName())

    // Set default image if not specified
    if database.Spec.Image == "" {
        database.Spec.Image = "postgres:14"
    }

    // Set default replicas if not specified
    if database.Spec.Replicas == nil {
        replicas := int32(1)
        database.Spec.Replicas = &replicas
    }

    // Set default storage class if not specified
    if database.Spec.Storage.StorageClass == "" {
        database.Spec.Storage.StorageClass = "standard"
    }

    return nil
}
```

### Task 2.2: Add Context-Aware Defaults

Enhance the Default method with namespace-based defaults:

```go
func (d *DatabaseCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
    database, ok := obj.(*databasev1.Database)
    if !ok {
        return fmt.Errorf("expected a Database object but got %T", obj)
    }
    databaselog.Info("Defaulting for Database", "name", database.GetName())

    // Set defaults based on namespace
    if database.Namespace == "production" {
        // Production defaults
        if database.Spec.Image == "" {
            database.Spec.Image = "postgres:14"  // Stable version
        }
        if database.Spec.Replicas == nil {
            replicas := int32(3)  // More replicas
            database.Spec.Replicas = &replicas
        }
    } else {
        // Development defaults
        if database.Spec.Image == "" {
            database.Spec.Image = "postgres:latest"
        }
        if database.Spec.Replicas == nil {
            replicas := int32(1)
            database.Spec.Replicas = &replicas
        }
    }

    // Common defaults
    if database.Spec.Storage.StorageClass == "" {
        database.Spec.Storage.StorageClass = "standard"
    }

    // Add labels (idempotent)
    if database.Labels == nil {
        database.Labels = make(map[string]string)
    }
    if _, exists := database.Labels["managed-by"]; !exists {
        database.Labels["managed-by"] = "database-operator"
    }

    // Add annotations (idempotent)
    if database.Annotations == nil {
        database.Annotations = make(map[string]string)
    }
    if _, exists := database.Annotations["database.example.com/version"]; !exists {
        database.Annotations["database.example.com/version"] = "v1"
    }

    return nil
}
```

## Exercise 3: Ensure Idempotency

### Task 3.1: Understand Idempotency

Mutations must be **idempotent** - applying them multiple times should have the same effect:

```go
// Idempotent: Only set if not already set
if database.Spec.Image == "" {
    database.Spec.Image = "postgres:14"
}
// If already set, doesn't change

// Idempotent: Check before adding to map
if _, exists := database.Labels["managed-by"]; !exists {
    database.Labels["managed-by"] = "database-operator"
}
// If already exists, doesn't add again
```

## Exercise 4: Deploy and Test Mutating Webhook

### Task 4.1: Generate Manifests

```bash
# Generate manifests (includes new mutating webhook)
make manifests
```

### Task 4.2: Rebuild and Deploy

```bash
# Rebuild the image (for docker)
make docker-build IMG=postgres-operator:latest

# Rebuild the image (for podman)
make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman

# For Docker:
kind load docker-image postgres-operator:latest --name k8s-operators-course

# For Podman:
podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
rm /tmp/postgres-operator.tar

# Redeploy for Docker
make deploy IMG=postgres-operator:latest

# Redeploy for Podman
make deploy IMG=localhost/postgres-operator:latest

# If you already have existing operator deployed from previous lab, restart the deployment
kubectl rollout restart deploy -n postgres-operator-system   postgres-operator-controller-manager

# in the operator logs you should see statements of registering validating and mutating webhooks
```

### Task 4.3: Verify Both Webhooks are Registered

```bash
# Check both webhooks are configured
kubectl get validatingwebhookconfigurations
kubectl get mutatingwebhookconfigurations
```

### Task 4.4: Test Minimal Resource

```bash
# Create resource with minimal spec (missing image, replicas)
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: minimal-db
spec:
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
  # image and replicas should be defaulted
EOF

# Check defaults were applied
echo "Image:"
kubectl get database minimal-db -o jsonpath='{.spec.image}'
echo
echo "Replicas:"
kubectl get database minimal-db -o jsonpath='{.spec.replicas}'
echo
echo "Storage Class:"
kubectl get database minimal-db -o jsonpath='{.spec.storage.storageClass}'
echo
echo "Managed-by label:"
kubectl get database minimal-db -o jsonpath='{.metadata.labels.managed-by}'
echo
```

### Task 4.5: Test Namespace-Based Defaults

```bash
# Create in production namespace
kubectl create namespace production --dry-run=client -o yaml | kubectl apply -f -

kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: prod-db
  namespace: production
spec:
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# Check production defaults (should be 3 replicas)
echo "Production replicas:"
kubectl get database prod-db -n production -o jsonpath='{.spec.replicas}'
echo

# Check default namespace (should be 1 replica from earlier test)
echo "Default namespace replicas:"
kubectl get database minimal-db -o jsonpath='{.spec.replicas}'
echo
```

## Exercise 5: Test Mutation Order

### Task 5.1: Verify Mutation Before Validation

```bash
# Create resource that would fail validation without defaults
# (missing image, but mutating webhook will set it to postgres:14)
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-order
spec:
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# Should succeed because:
# 1. Mutating webhook sets image to postgres:14
# 2. Validating webhook validates it's a postgres image
kubectl get database test-order -o jsonpath='{.spec.image}'
echo
```

## Cleanup

```bash
# Delete test resources
kubectl delete databases --all -A
kubectl delete namespace production --ignore-not-found
```

## Lab Summary

In this lab, you:
- Added mutating webhook to existing validating webhook
- Implemented defaulting logic using `CustomDefaulter` interface
- Added context-aware defaults
- Ensured idempotency
- Tested mutation scenarios
- Verified mutation order

## Key Learnings

1. Add mutating webhook to existing `internal/webhook/v1/database_webhook.go`
2. Use `webhook.CustomDefaulter` interface with separate struct
3. `Default` method receives `context.Context` and `runtime.Object`
4. Register defaulter with `.WithDefaulter(&DatabaseCustomDefaulter{})`
5. Defaults can be context-aware (namespace, etc.)
6. Mutations must be idempotent
7. Mutating webhooks run before validating webhooks

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Mutating Webhook](../solutions/mutating-webhook.go) - Complete mutating webhook implementation with defaulting logic

## Next Steps

Now let's learn about certificate management and deployment!

**Navigation:** [← Previous Lab: Validating Webhooks](lab-02-validating-webhooks.md) | [Related Lesson](../lessons/03-mutating-webhooks.md) | [Next Lab: Webhook Deployment →](lab-04-webhook-deployment.md)
