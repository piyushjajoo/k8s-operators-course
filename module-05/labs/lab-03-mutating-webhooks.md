# Lab 5.3: Building Mutating Webhook

**Related Lesson:** [Lesson 5.3: Implementing Mutating Webhooks](../lessons/03-mutating-webhooks.md)  
**Navigation:** [← Previous Lab: Validating Webhooks](lab-02-validating-webhooks.md) | [Module Overview](../README.md) | [Next Lab: Webhook Deployment →](lab-04-webhook-deployment.md)

## Objectives

- Scaffold mutating webhook
- Implement defaulting logic
- Test mutation scenarios
- Ensure idempotency

## Prerequisites

- Completion of [Lab 5.2](lab-02-validating-webhooks.md)
- Database operator with validating webhook
- Understanding of defaulting patterns

## Exercise 1: Scaffold Mutating Webhook

### Task 1.1: Create Mutating Webhook

```bash
# Navigate to your operator
cd ~/postgres-operator

# Create mutating webhook
kubebuilder create webhook \
  --group database \
  --version v1 \
  --kind Database \
  --defaulting
```

**Observe:**
- What was added to database_types.go?
- What webhook marker was created?

### Task 1.2: Examine Generated Code

```bash
# Check for Default method
cat api/v1/database_webhook.go | grep -A 10 "Default"

# Check webhook marker
cat api/v1/database_webhook.go | grep "kubebuilder:webhook.*mutating"
```

## Exercise 2: Implement Defaulting Logic

### Task 2.1: Add Default Method

Edit `api/v1/database_webhook.go`:

```go
//+kubebuilder:webhook:path=/mutate-database-example-com-v1-database,mutating=true,failurePolicy=fail,sideEffects=None,groups=database.example.com,resources=databases,verbs=create;update,versions=v1,name=mdatabase.kb.io

var _ webhook.Defaulter = &Database{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Database) Default() {
    databaseLog.Info("default", "name", r.Name)
    
    // Set default image if not specified
    if r.Spec.Image == "" {
        r.Spec.Image = "postgres:14"
    }
    
    // Set default replicas if not specified
    if r.Spec.Replicas == nil {
        replicas := int32(1)
        r.Spec.Replicas = &replicas
    }
    
    // Set default storage class if not specified
    if r.Spec.Storage.StorageClass == "" {
        r.Spec.Storage.StorageClass = "standard"
    }
}
```

### Task 2.2: Add Context-Aware Defaults

```go
func (r *Database) Default() {
    databaseLog.Info("default", "name", r.Name)
    
    // Set defaults based on namespace
    if r.Namespace == "production" {
        // Production defaults
        if r.Spec.Image == "" {
            r.Spec.Image = "postgres:14"  // Stable version
        }
        if r.Spec.Replicas == nil {
            replicas := int32(3)  // More replicas
            r.Spec.Replicas = &replicas
        }
    } else {
        // Development defaults
        if r.Spec.Image == "" {
            r.Spec.Image = "postgres:latest"
        }
        if r.Spec.Replicas == nil {
            replicas := int32(1)
            r.Spec.Replicas = &replicas
        }
    }
    
    // Common defaults
    if r.Spec.Storage.StorageClass == "" {
        r.Spec.Storage.StorageClass = "standard"
    }
    
    // Add labels
    if r.Labels == nil {
        r.Labels = make(map[string]string)
    }
    r.Labels["managed-by"] = "database-operator"
    
    // Add annotations
    if r.Annotations == nil {
        r.Annotations = make(map[string]string)
    }
    r.Annotations["database.example.com/version"] = "v1"
}
```

## Exercise 3: Ensure Idempotency

### Task 3.1: Make Defaults Idempotent

```go
func (r *Database) Default() {
    // Idempotent: Only set if not already set
    if r.Spec.Image == "" {
        r.Spec.Image = "postgres:14"
    }
    // If already set, don't change it
    
    // Idempotent: Check before adding to slice
    if r.Labels == nil {
        r.Labels = make(map[string]string)
    }
    if _, exists := r.Labels["managed-by"]; !exists {
        r.Labels["managed-by"] = "database-operator"
    }
    // If already exists, don't add again
}
```

### Task 3.2: Test Idempotency

```bash
# Create resource
kubectl apply -f database.yaml

# Get the resource (should have defaults)
kubectl get database test-db -o yaml

# Apply again (should be idempotent)
kubectl apply -f database.yaml

# Check - should be the same
kubectl get database test-db -o yaml
```

## Exercise 4: Test Mutating Webhook

### Task 4.1: Generate and Install

```bash
# Generate manifests
make manifests

# Generate certificates
make certs
make install-cert

# Run operator
make run
```

### Task 4.2: Test Minimal Resource

```bash
# Create resource with minimal spec
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
kubectl get database minimal-db -o jsonpath='{.spec.image}'
echo
kubectl get database minimal-db -o jsonpath='{.spec.replicas}'
echo
kubectl get database minimal-db -o jsonpath='{.spec.storage.storageClass}'
echo
kubectl get database minimal-db -o jsonpath='{.metadata.labels.managed-by}'
echo
```

### Task 4.3: Test Namespace-Based Defaults

```bash
# Create in production namespace
kubectl create namespace production
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

# Check production defaults
kubectl get database prod-db -n production -o jsonpath='{.spec.replicas}'
echo  # Should be 3

# Create in default namespace
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

# Check dev defaults
kubectl get database dev-db -o jsonpath='{.spec.replicas}'
echo  # Should be 1
```

## Exercise 5: Test Mutation Order

### Task 5.1: Verify Mutation Before Validation

```bash
# Create resource that would fail validation without defaults
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-order
spec:
  # Missing image - should be defaulted by mutating webhook
  # Then validated by validating webhook
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# Should succeed because:
# 1. Mutating webhook sets image to postgres:14
# 2. Validating webhook validates it's a postgres image
```

## Cleanup

```bash
# Delete test resources
kubectl delete databases --all -A
kubectl delete namespace production 2>/dev/null || true
```

## Lab Summary

In this lab, you:
- Scaffolded mutating webhook
- Implemented defaulting logic
- Added context-aware defaults
- Ensured idempotency
- Tested mutation scenarios
- Verified mutation order

## Key Learnings

1. Mutating webhooks modify resources before validation
2. Default() method sets defaults
3. Defaults can be context-aware (namespace, etc.)
4. Mutations must be idempotent
5. Mutating webhooks run before validating webhooks
6. Kubebuilder handles patching automatically

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Mutating Webhook](../solutions/mutating-webhook.go) - Complete mutating webhook implementation with defaulting logic

## Next Steps

Now let's learn about certificate management and deployment!

**Navigation:** [← Previous Lab: Validating Webhooks](lab-02-validating-webhooks.md) | [Related Lesson](../lessons/03-mutating-webhooks.md) | [Next Lab: Webhook Deployment →](lab-04-webhook-deployment.md)

