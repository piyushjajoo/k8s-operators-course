# Lab 8.1: Building Multi-Tenant Operator

**Related Lesson:** [Lesson 8.1: Multi-Tenancy and Namespace Isolation](../lessons/01-multi-tenancy.md)  
**Navigation:** [Module Overview](../README.md) | [Next Lab: Operator Composition →](lab-02-operator-composition.md)

## Objectives

- Create cluster-scoped CRD
- Implement namespace isolation
- Handle resource quotas
- Test multi-tenant scenarios

## Prerequisites

- Completion of [Module 7](../module-07/README.md)
- Database operator ready
- Understanding of namespaces and RBAC

## Exercise 1: Create Cluster-Scoped CRD

### Task 1.1: Update CRD Scope

Edit `config/crd/bases/database.example.com_databases.yaml`:

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: databases.database.example.com
spec:
  group: database.example.com
  versions:
  - name: v1
    served: true
    storage: true
  scope: Cluster  # Change from Namespaced to Cluster
  names:
    plural: databases
    singular: database
    kind: Database
```

### Task 1.2: Update API Types

Update your API types to remove namespace requirement:

```go
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

type Database struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec              DatabaseSpec   `json:"spec,omitempty"`
    Status            DatabaseStatus `json:"status,omitempty"`
}
```

### Task 1.3: Regenerate and Apply

```bash
# Regenerate CRD
make manifests

# Apply CRD
kubectl apply -f config/crd/bases/database.example.com_databases.yaml

# Verify scope
kubectl get crd databases.database.example.com -o jsonpath='{.spec.scope}'
# Should output: Cluster
```

## Exercise 2: Implement Namespace Isolation

### Task 2.1: Add Tenant Label

Update your Database spec to include tenant information:

```go
type DatabaseSpec struct {
    // ... existing fields ...
    Tenant string `json:"tenant,omitempty"`
}
```

### Task 2.2: Filter by Tenant

Update controller to filter by tenant:

```go
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    db := &databasev1.Database{}
    if err := r.Get(ctx, req.NamespacedName, db); err != nil {
        return ctrl.Result{}, err
    }
    
    // Get tenant from spec or namespace
    tenant := db.Spec.Tenant
    if tenant == "" {
        // Extract from namespace if namespaced
        tenant = req.Namespace
    }
    
    // Apply tenant-specific logic
    return r.reconcileForTenant(ctx, db, tenant)
}
```

## Exercise 3: Handle Resource Quotas

### Task 3.1: Create Resource Quota

Create `config/samples/quota.yaml`:

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: database-quota
  namespace: tenant-1
spec:
  hard:
    databases.database.example.com: "5"
```

### Task 3.2: Check Quota in Controller

```go
func (r *DatabaseReconciler) checkQuota(ctx context.Context, namespace string) error {
    quota := &corev1.ResourceQuota{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      "database-quota",
        Namespace: namespace,
    }, quota)
    
    if errors.IsNotFound(err) {
        // No quota, proceed
        return nil
    }
    
    // Count existing databases
    databases := &databasev1.DatabaseList{}
    r.List(ctx, databases, client.InNamespace(namespace))
    
    used := int64(len(databases.Items))
    hard := quota.Spec.Hard["databases.database.example.com"]
    
    if hard.Value() <= used {
        return fmt.Errorf("quota exceeded: %d/%d databases", used, hard.Value())
    }
    
    return nil
}
```

## Exercise 4: Test Multi-Tenant Scenarios

### Task 4.1: Create Multiple Tenants

```bash
# Create namespaces for tenants
kubectl create namespace tenant-1
kubectl create namespace tenant-2

# Create databases in different tenants
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: db-tenant-1
spec:
  tenant: tenant-1
  image: postgres:14
  replicas: 1
EOF

kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: db-tenant-2
spec:
  tenant: tenant-2
  image: postgres:14
  replicas: 1
EOF
```

### Task 4.2: Verify Isolation

```bash
# List databases for tenant-1
kubectl get databases -l tenant=tenant-1

# Verify resources are isolated
kubectl get statefulsets -n tenant-1
kubectl get statefulsets -n tenant-2
```

## Cleanup

```bash
# Delete test resources
kubectl delete databases --all
kubectl delete namespace tenant-1 tenant-2
```

## Lab Summary

In this lab, you:
- Created cluster-scoped CRD
- Implemented namespace isolation
- Added resource quota handling
- Tested multi-tenant scenarios

## Key Learnings

1. Cluster-scoped operators manage all namespaces
2. Namespace isolation provides tenant separation
3. Resource quotas limit tenant usage
4. Labels enable tenant filtering
5. Multi-tenancy requires careful design
6. RBAC enforces tenant boundaries

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Cluster-Scoped CRD](../solutions/cluster-scoped-crd.yaml) - Cluster-scoped CRD example
- [Multi-Tenant Controller](../solutions/multi-tenant-controller.go) - Multi-tenant controller implementation

## Next Steps

Now let's learn about operator composition!

**Navigation:** [← Module Overview](../README.md) | [Related Lesson](../lessons/01-multi-tenancy.md) | [Next Lab: Operator Composition →](lab-02-operator-composition.md)

