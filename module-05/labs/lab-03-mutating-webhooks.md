---
layout: default
title: "Lab 05.3: Mutating Webhooks"
nav_order: 13
parent: "Module 5: Webhooks & Admission Control"
grand_parent: Modules
mermaid: true
---

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
    databaselog.Info("Defaulting for Database", "name", database.GetName(), "namespace", database.GetNamespace())

    // Set defaults based on namespace
    if database.Namespace == "production" {
        // Production defaults - ensure minimum 3 replicas
        // Note: We check < 3 instead of nil because CRD schema defaults may already set replicas=1
        if database.Spec.Replicas == nil || *database.Spec.Replicas < 3 {
            replicas := int32(3)
            database.Spec.Replicas = &replicas
        }
    }
    // For non-production, CRD schema default of 1 replica is fine

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

> **Note:** We check `< 3` instead of `nil` for replicas because CRD schema defaults (via `+kubebuilder:default=1`) are applied before webhooks run. This ensures production namespaces always get at least 3 replicas.

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

### Task 4.1: Enable MutatingWebhookConfiguration in Kustomization

Since we added a mutating webhook manually, we need to uncomment the MutatingWebhookConfiguration replacements in `config/default/kustomization.yaml` so cert-manager can inject the CA bundle:

```bash
cd ~/postgres-operator

# Uncomment the MutatingWebhookConfiguration section (around lines 188-217)
# Find the section that says "Uncomment the following block if you have a DefaultingWebhook"
# and uncomment it.

# Verify it's uncommented - should show MutatingWebhookConfiguration without # prefix
grep -A 5 "DefaultingWebhook" config/default/kustomization.yaml
```

### Task 4.2: Generate Manifests

```bash
# Generate manifests (includes new mutating webhook configuration)
make manifests

# Verify mutating webhook manifest was generated
grep "mutating" config/webhook/manifests.yaml
```

### Task 4.3: Undeploy and Clean Up Stale Webhooks

Since we added a new webhook, we need to fully redeploy. Also clean up any stale webhook configurations from previous deployments:

```bash
# Remove existing deployment
make undeploy

# Clean up any stale webhook configurations (from previous deployments without proper prefixes)
kubectl delete validatingwebhookconfiguration validating-webhook-configuration 2>/dev/null || true
kubectl delete mutatingwebhookconfiguration mutating-webhook-configuration 2>/dev/null || true

# Verify cleanup
kubectl get validatingwebhookconfigurations
kubectl get mutatingwebhookconfigurations
# Should only show cert-manager and ingress-nginx webhooks, not our old ones

# Wait for resources to be deleted
kubectl get all -n postgres-operator-system
# Should show "No resources found"
```

### Task 4.4: Rebuild and Deploy

```bash
# Rebuild the image
# For Docker:
make docker-build IMG=postgres-operator:latest

# For Podman:
make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman

# Load image into kind
# For Docker:
kind load docker-image postgres-operator:latest --name k8s-operators-course

# For Podman:
podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
rm /tmp/postgres-operator.tar

# Deploy
# For Docker:
make deploy IMG=postgres-operator:latest

# For Podman:
make deploy IMG=localhost/postgres-operator:latest
```

### Task 4.5: Wait for Certificates

cert-manager needs time to generate certificates and inject the CA bundle:

```bash
# Wait for certificate to be ready
kubectl get certificate -n postgres-operator-system -w

# Wait for pod to be ready
kubectl wait --for=condition=Ready pod -l control-plane=controller-manager \
  -n postgres-operator-system --timeout=120s

# Check operator logs - should see both webhooks registered
kubectl logs -n postgres-operator-system deployment/postgres-operator-controller-manager | grep -i webhook
```

### Task 4.6: Verify Both Webhooks are Registered

```bash
# Check both webhooks are configured
kubectl get validatingwebhookconfigurations
kubectl get mutatingwebhookconfigurations

# You should see both:
# - postgres-operator-validating-webhook-configuration
# - postgres-operator-mutating-webhook-configuration
```

### Task 4.7: Test Minimal Resource

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

### Task 4.8: Test Namespace-Based Defaults

Our webhook checks `replicas < 3` (not just `nil`) for production namespace, so it works even when CRD schema defaults have already set `replicas=1`.

```bash
# Clean up previous test resources
kubectl delete database --all --ignore-not-found
kubectl delete database --all -n production --ignore-not-found

# Create in default namespace (should stay at 1 replica - CRD default)
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: dev-db
spec:
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# Create in production namespace (should be bumped to 3 replicas)
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

# Check results
echo "=== Default namespace (should be 1 replica) ==="
kubectl get database dev-db -o jsonpath='Replicas: {.spec.replicas}'
echo

echo "=== Production namespace (should be 3 replicas) ==="
kubectl get database prod-db -n production -o jsonpath='Replicas: {.spec.replicas}'
echo
```

> **Key Learning:** CRD schema defaults (`+kubebuilder:default`) are applied before webhooks. To override them, check for the default value (e.g., `< 3`) instead of just checking for `nil`.

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
