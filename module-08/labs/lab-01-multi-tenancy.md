# Lab 8.1: Building Multi-Tenant Operator

**Related Lesson:** [Lesson 8.1: Multi-Tenancy and Namespace Isolation](../lessons/01-multi-tenancy.md)  
**Navigation:** [Module Overview](../README.md) | [Next Lab: Operator Composition →](lab-02-operator-composition.md)

## Objectives

- Scaffold a new cluster-scoped API using kubebuilder
- Keep existing namespace-scoped Database controller
- Implement namespace isolation
- Handle resource quotas
- Test multi-tenant scenarios

## Prerequisites

- Completion of [Module 7](../../module-07/README.md)
- Database operator ready
- Understanding of namespaces and RBAC

## Overview

In this lab, you'll create a **new** cluster-scoped API called `ClusterDatabase` alongside your existing namespace-scoped `Database` API. This approach allows you to:

1. **Keep your existing Database controller** - No changes needed
2. **Learn cluster-scoped concepts** - With a dedicated API
3. **Compare both approaches** - Side by side in the same project

The key difference:
- `Database` (existing): Namespace-scoped, manages databases within a single namespace
- `ClusterDatabase` (new): Cluster-scoped, manages databases across any namespace

## Exercise 1: Scaffold Cluster-Scoped API with Kubebuilder

### Task 1.1: Create New API

Use kubebuilder to scaffold the new ClusterDatabase API:

```bash
# Navigate to your operator project
cd ~/postgres-operator

# Scaffold new API with cluster scope
kubebuilder create api \
  --group database \
  --version v1 \
  --kind ClusterDatabase \
  --resource --controller

# When prompted:
# Create Resource [y/n]: y
# Create Controller [y/n]: y
```

This creates:
- `api/v1/clusterdatabase_types.go` - API type definitions
- `internal/controller/clusterdatabase_controller.go` - Controller scaffold

### Task 1.2: Configure Cluster Scope

Edit `api/v1/clusterdatabase_types.go` to add the cluster scope marker:

```go
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Namespace",type="string",JSONPath=".spec.targetNamespace"
// +kubebuilder:printcolumn:name="Tenant",type="string",JSONPath=".spec.tenant"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ClusterDatabase is the Schema for the clusterdatabases API
// It is cluster-scoped and manages databases across namespaces
type ClusterDatabase struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   ClusterDatabaseSpec   `json:"spec,omitempty"`
    Status ClusterDatabaseStatus `json:"status,omitempty"`
}
```

The key marker is `// +kubebuilder:resource:scope=Cluster`.

### Task 1.3: Define ClusterDatabase Spec

Update the spec in `api/v1/clusterdatabase_types.go`:

```go
// ClusterDatabaseSpec defines the desired state of ClusterDatabase
type ClusterDatabaseSpec struct {
    // Image is the PostgreSQL image to use
    // +kubebuilder:validation:Required
    // +kubebuilder:default="postgres:14"
    Image string `json:"image"`

    // Replicas is the number of database replicas
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:validation:Maximum=10
    // +kubebuilder:default=1
    Replicas *int32 `json:"replicas,omitempty"`

    // Storage is the storage configuration
    Storage StorageSpec `json:"storage"`

    // Resources are the resource requirements
    Resources corev1.ResourceRequirements `json:"resources,omitempty"`

    // DatabaseName is the name of the database to create
    // +kubebuilder:validation:Required
    DatabaseName string `json:"databaseName"`

    // Username is the database user
    // +kubebuilder:validation:Required
    Username string `json:"username"`

    // TargetNamespace is where resources will be created
    // Required for cluster-scoped resources to know where to deploy
    // +kubebuilder:validation:Required
    TargetNamespace string `json:"targetNamespace"`

    // Tenant identifies which tenant owns this database
    // +optional
    Tenant string `json:"tenant,omitempty"`
}
```

Note: You can reuse the `StorageSpec` type from your existing Database API.

### Task 1.4: Define ClusterDatabase Status

```go
// ClusterDatabaseStatus defines the observed state of ClusterDatabase
type ClusterDatabaseStatus struct {
    // Phase is the current phase
    // +kubebuilder:validation:Enum=Pending;Creating;Ready;Failed
    Phase string `json:"phase,omitempty"`

    // Ready indicates if the database is ready
    Ready bool `json:"ready,omitempty"`

    // Endpoint is the database endpoint
    Endpoint string `json:"endpoint,omitempty"`

    // SecretName is the name of the Secret containing credentials
    SecretName string `json:"secretName,omitempty"`

    // TargetNamespace shows where resources were created
    TargetNamespace string `json:"targetNamespace,omitempty"`

    // Conditions represent the latest observations
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}
```

### Task 1.5: Generate and Apply CRD

```bash
# Generate CRD manifests
make manifests

# Verify the CRD was generated with cluster scope
cat config/crd/bases/database.example.com_clusterdatabases.yaml | grep "scope:"
# Should output: scope: Cluster

# Install CRDs
make install

# Verify both CRDs exist
kubectl get crd | grep database.example.com
# Should show:
# clusterdatabases.database.example.com   (new, Cluster-scoped)
# databases.database.example.com          (existing, Namespaced)

# Check the scope
kubectl get crd clusterdatabases.database.example.com -o jsonpath='{.spec.scope}'
# Should output: Cluster
```

### Key Differences from Database:

| Aspect | Database (Namespaced) | ClusterDatabase (Cluster-Scoped) |
|--------|----------------------|----------------------------------|
| Scope marker | (none or `scope=Namespaced`) | `+kubebuilder:resource:scope=Cluster` |
| Namespace | Implicit from resource | Explicit `targetNamespace` field |
| Access | Within one namespace | Across all namespaces |
| Use case | Team-level resources | Platform-level management |

## Exercise 2: Implement ClusterDatabase Controller

### Task 2.1: Update Controller Logic

Edit `internal/controller/clusterdatabase_controller.go`:

```go
package controller

import (
    "context"
    "fmt"

    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"

    databasev1 "github.com/example/postgres-operator/api/v1"
)

type ClusterDatabaseReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=database.example.com,resources=clusterdatabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=clusterdatabases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=clusterdatabases/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=resourcequotas,verbs=get;list;watch

func (r *ClusterDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // Fetch ClusterDatabase (cluster-scoped, no namespace in request)
    db := &databasev1.ClusterDatabase{}
    if err := r.Get(ctx, req.NamespacedName, db); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    logger.Info("Reconciling ClusterDatabase",
        "name", db.Name,
        "targetNamespace", db.Spec.TargetNamespace,
        "tenant", db.Spec.Tenant)

    // Validate target namespace exists
    if err := r.validateNamespace(ctx, db.Spec.TargetNamespace); err != nil {
        return ctrl.Result{}, err
    }

    // Check quota for the target namespace
    if err := r.checkQuota(ctx, db.Spec.TargetNamespace); err != nil {
        logger.Error(err, "Quota exceeded")
        return ctrl.Result{}, err
    }

    // Reconcile resources in target namespace
    // (Similar to Database controller but creates in targetNamespace)
    if err := r.reconcileSecret(ctx, db); err != nil {
        return ctrl.Result{}, err
    }

    if err := r.reconcileStatefulSet(ctx, db); err != nil {
        return ctrl.Result{}, err
    }

    if err := r.reconcileService(ctx, db); err != nil {
        return ctrl.Result{}, err
    }

    // Update status
    return ctrl.Result{}, r.updateStatus(ctx, db)
}

func (r *ClusterDatabaseReconciler) validateNamespace(ctx context.Context, namespace string) error {
    ns := &corev1.Namespace{}
    if err := r.Get(ctx, client.ObjectKey{Name: namespace}, ns); err != nil {
        if errors.IsNotFound(err) {
            return fmt.Errorf("target namespace %s does not exist", namespace)
        }
        return err
    }
    return nil
}

func (r *ClusterDatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.ClusterDatabase{}).
        Complete(r)
}
```

### Task 2.2: Verify Controller is Registered

Kubebuilder automatically registers the controller in `cmd/main.go`. Verify:

```go
// This should already be added by kubebuilder
if err = (&controller.ClusterDatabaseReconciler{
    Client: mgr.GetClient(),
    Scheme: mgr.GetScheme(),
}).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "ClusterDatabase")
    os.Exit(1)
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
    # Limit ClusterDatabases targeting this namespace
    clusterdatabases.database.example.com: "5"
```

### Task 3.2: Implement Quota Checking

Add to your controller:

```go
func (r *ClusterDatabaseReconciler) checkQuota(ctx context.Context, namespace string) error {
    quota := &corev1.ResourceQuota{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      "database-quota",
        Namespace: namespace,
    }, quota)

    if errors.IsNotFound(err) {
        // No quota, proceed
        return nil
    }
    if err != nil {
        return err
    }

    // Count ClusterDatabases targeting this namespace
    databases := &databasev1.ClusterDatabaseList{}
    if err := r.List(ctx, databases); err != nil {
        return err
    }

    var count int64
    for _, db := range databases.Items {
        if db.Spec.TargetNamespace == namespace {
            count++
        }
    }

    hard, exists := quota.Spec.Hard["clusterdatabases.database.example.com"]
    if !exists {
        return nil
    }

    if hard.Value() <= count {
        return fmt.Errorf("quota exceeded: %d/%d clusterdatabases", count, hard.Value())
    }

    return nil
}
```

## Exercise 4: Test Multi-Tenant Scenarios

### Task 4.1: Run the Operator

```bash
# Generate and install CRDs
make manifests install

# Run the operator locally
make run
```

### Task 4.2: Create Tenant Namespaces

In a new terminal:

```bash
# Create namespaces for tenants
kubectl create namespace tenant-1
kubectl create namespace tenant-2

# Label namespaces for tenant identification
kubectl label namespace tenant-1 tenant=tenant-1
kubectl label namespace tenant-2 tenant=tenant-2
```

### Task 4.3: Create ClusterDatabases for Different Tenants

```bash
# Create ClusterDatabase for tenant-1
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: ClusterDatabase
metadata:
  name: cdb-tenant-1-prod
spec:
  targetNamespace: tenant-1
  tenant: tenant-1
  image: postgres:14
  replicas: 1
  databaseName: proddb
  username: admin
  storage:
    size: "10Gi"
EOF

# Create ClusterDatabase for tenant-2
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: ClusterDatabase
metadata:
  name: cdb-tenant-2-prod
spec:
  targetNamespace: tenant-2
  tenant: tenant-2
  image: postgres:14
  replicas: 1
  databaseName: proddb
  username: admin
  storage:
    size: "10Gi"
EOF
```

### Task 4.4: Verify Isolation

```bash
# List all ClusterDatabases (cluster-wide view)
kubectl get clusterdatabases

# Output shows all databases with their target namespaces:
# NAME               PHASE   NAMESPACE   TENANT     READY   AGE
# cdb-tenant-1-prod  Ready   tenant-1    tenant-1   true    1m
# cdb-tenant-2-prod  Ready   tenant-2    tenant-2   true    1m

# Verify resources are created in correct namespaces
kubectl get statefulsets -n tenant-1
kubectl get statefulsets -n tenant-2

# Filter by tenant using jsonpath
kubectl get clusterdatabases -o jsonpath='{range .items[?(@.spec.tenant=="tenant-1")]}{.metadata.name}{"\n"}{end}'
```

### Task 4.5: Compare with Namespace-Scoped Database

```bash
# You can still use the namespace-scoped Database in parallel
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: local-db
  namespace: tenant-1
spec:
  image: postgres:14
  replicas: 1
  databaseName: localdb
  username: user
  storage:
    size: "5Gi"
EOF

# List both types
kubectl get databases -n tenant-1    # Shows namespace-scoped
kubectl get clusterdatabases          # Shows cluster-scoped
```

## Exercise 5: Understanding Ownership Limitations

### Task 5.1: Cluster-Scoped Owner Restrictions

Important: Cluster-scoped resources **cannot** use `OwnerReferences` to own namespace-scoped resources. Instead, use labels:

```go
// In ClusterDatabaseReconciler
func (r *ClusterDatabaseReconciler) reconcileStatefulSet(ctx context.Context, db *databasev1.ClusterDatabase) error {
    statefulSet := &appsv1.StatefulSet{
        ObjectMeta: metav1.ObjectMeta{
            Name:      db.Name,
            Namespace: db.Spec.TargetNamespace,
            Labels: map[string]string{
                // Use labels to track ownership instead of OwnerReferences
                "app.kubernetes.io/managed-by": "clusterdatabase-controller",
                "clusterdatabase":              db.Name,
                "tenant":                       db.Spec.Tenant,
            },
        },
        // ... spec
    }

    // Note: Cannot use ctrl.SetControllerReference() here
    // because cluster-scoped -> namespaced ownership is not allowed

    return r.Create(ctx, statefulSet)
}
```

### Task 5.2: Implement Cleanup with Finalizers

Since we can't use OwnerReferences for automatic garbage collection, use finalizers:

```go
const clusterDatabaseFinalizer = "database.example.com/clusterdatabase-finalizer"

func (r *ClusterDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    db := &databasev1.ClusterDatabase{}
    if err := r.Get(ctx, req.NamespacedName, db); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Handle deletion
    if !db.DeletionTimestamp.IsZero() {
        if controllerutil.ContainsFinalizer(db, clusterDatabaseFinalizer) {
            // Clean up managed resources
            if err := r.cleanupManagedResources(ctx, db); err != nil {
                return ctrl.Result{}, err
            }
            controllerutil.RemoveFinalizer(db, clusterDatabaseFinalizer)
            return ctrl.Result{}, r.Update(ctx, db)
        }
        return ctrl.Result{}, nil
    }

    // Add finalizer if not present
    if !controllerutil.ContainsFinalizer(db, clusterDatabaseFinalizer) {
        controllerutil.AddFinalizer(db, clusterDatabaseFinalizer)
        return ctrl.Result{}, r.Update(ctx, db)
    }

    // ... rest of reconciliation
}

func (r *ClusterDatabaseReconciler) cleanupManagedResources(ctx context.Context, db *databasev1.ClusterDatabase) error {
    // Delete resources by label selector
    labelSelector := client.MatchingLabels{
        "clusterdatabase": db.Name,
    }

    // Delete StatefulSet
    if err := r.DeleteAllOf(ctx, &appsv1.StatefulSet{},
        client.InNamespace(db.Spec.TargetNamespace), labelSelector); err != nil {
        return err
    }

    // Delete Service
    if err := r.DeleteAllOf(ctx, &corev1.Service{},
        client.InNamespace(db.Spec.TargetNamespace), labelSelector); err != nil {
        return err
    }

    // Delete Secret
    return r.DeleteAllOf(ctx, &corev1.Secret{},
        client.InNamespace(db.Spec.TargetNamespace), labelSelector)
}
```

## Cleanup

```bash
# Delete ClusterDatabases
kubectl delete clusterdatabases --all

# Delete test namespaces
kubectl delete namespace tenant-1 tenant-2
```

## Lab Summary

In this lab, you:
- Scaffolded a new cluster-scoped API using kubebuilder
- Kept the existing namespace-scoped Database controller
- Implemented namespace isolation with `targetNamespace`
- Added resource quota handling
- Tested multi-tenant scenarios
- Learned about cluster-scoped ownership limitations

## Key Learnings

1. **Use kubebuilder to scaffold new APIs** - `kubebuilder create api` handles boilerplate
2. **Use `+kubebuilder:resource:scope=Cluster` marker** - Makes the CRD cluster-scoped
3. **Cluster-scoped resources need explicit namespace fields** - Use `targetNamespace`
4. **Cannot use OwnerReferences across scopes** - Use labels and finalizers instead
5. **Both controllers can coexist** - Each manages its own resource type
6. **`make manifests` generates CRDs** - No need to write CRD YAML manually

## Comparison: Database vs ClusterDatabase

| Feature | Database | ClusterDatabase |
|---------|----------|-----------------|
| Scope | Namespaced | Cluster |
| Namespace | Implicit | Explicit (`targetNamespace`) |
| OwnerReferences | Yes | No (use labels) |
| Cleanup | Automatic (GC) | Manual (finalizers) |
| RBAC | Per namespace | Cluster-wide |
| Use case | Team resources | Platform management |

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [ClusterDatabase Types](../solutions/clusterdatabase-types.go) - Complete API type definitions
- [ClusterDatabase Controller](../solutions/clusterdatabase-controller.go) - Complete controller implementation
- [Multi-Tenant Controller](../solutions/multi-tenant-controller.go) - Multi-tenant patterns example

## Next Steps

Now let's learn about operator composition!

**Navigation:** [← Module Overview](../README.md) | [Related Lesson](../lessons/01-multi-tenancy.md) | [Next Lab: Operator Composition →](lab-02-operator-composition.md)
