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
# Generate code and CRD manifests
make generate
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

### Task 2.1: Copy Complete Controller Implementation

The ClusterDatabase controller is similar to your existing Database controller, but with key differences for cluster-scoped resources. Rather than writing it from scratch, copy the complete implementation from the solutions file:

```bash
# Copy the complete controller implementation
cp path/to/solutions/clusterdatabase-controller.go internal/controller/clusterdatabase_controller.go
```

Or, if you prefer to type it yourself, copy the complete controller from:
**[solutions/clusterdatabase-controller.go](../solutions/clusterdatabase-controller.go)**

The complete controller includes:
- `Reconcile()` - Main reconciliation loop
- `validateNamespace()` - Validates target namespace exists
- `checkQuota()` - Checks resource quotas
- `reconcileSecret()` - Creates credentials Secret in target namespace
- `reconcileStatefulSet()` - Creates StatefulSet in target namespace
- `reconcileService()` - Creates Service in target namespace
- `updateStatus()` - Updates ClusterDatabase status

### Task 2.2: Understand Key Differences from Database Controller

Here are the key differences in the ClusterDatabase controller:

**1. Target Namespace Field:**
```go
// Database controller uses implicit namespace from the resource
namespace := db.Namespace

// ClusterDatabase controller uses explicit targetNamespace
namespace := db.Spec.TargetNamespace
```

**2. No OwnerReferences (use labels instead):**
```go
// Database controller can use OwnerReferences
ctrl.SetControllerReference(db, statefulSet, r.Scheme)

// ClusterDatabase controller CANNOT - use labels instead
statefulSet.Labels["clusterdatabase"] = db.Name
statefulSet.Labels["tenant"] = db.Spec.Tenant
```

**3. Namespace Validation:**
```go
// ClusterDatabase must validate target namespace exists
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
```

### Task 2.3: Verify Controller is Registered

Kubebuilder automatically registers the controller in `cmd/main.go`. Verify it looks like this:

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

### Task 2.4: Build and Verify

```bash
# Ensure the code compiles
make build

# If there are any compilation errors, verify you copied the complete
# controller from the solutions file
```

## Exercise 3: Handle Resource Quotas

The `checkQuota` function is already included in the solutions file you copied. Let's understand how it works and test it.

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

### Task 3.2: Understand Quota Checking

The `checkQuota` function in your controller (from solutions) works like this:

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

Key points:
- Checks if a ResourceQuota exists in the target namespace
- Counts all ClusterDatabases targeting that namespace (cluster-wide list, then filter)
- Returns error if quota would be exceeded

## Exercise 4: Test Multi-Tenant Scenarios

> **Prerequisites:** Ensure you have completed Exercise 2 (copied the complete controller from solutions) and your code compiles with `make build`.

### Task 4.1: Build and Deploy Operator to Kind Cluster

Since operators with webhooks (from earlier modules) require TLS certificates and in-cluster deployment, we'll deploy to the kind cluster:

```bash
# Verify code compiles
make build

# Generate code and manifests
make generate manifests

# Build the container image
make docker-build IMG=postgres-operator:latest

# Load image into kind cluster
kind load docker-image postgres-operator:latest --name k8s-operators-course
```

Before deploying, ensure `imagePullPolicy: IfNotPresent` is set in `config/manager/manager.yaml`:

```yaml
containers:
- name: manager
  image: controller:latest
  imagePullPolicy: IfNotPresent  # Add this line if not present
```

Now deploy:

```bash
# Deploy operator to cluster
make deploy IMG=postgres-operator:latest

# Verify operator is running
kubectl get pods -n postgres-operator-system

# Check logs
kubectl logs -n postgres-operator-system deployment/postgres-operator-controller-manager -f
```

> **Using Podman instead of Docker?**
> 
> ```bash
> # Build with podman
> make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman
> 
> # Load image into kind (save to tarball, then load)
> podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
> kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
> rm /tmp/postgres-operator.tar
> 
> # Deploy with localhost/ prefix
> make deploy IMG=localhost/postgres-operator:latest
> ```

> **Getting `ErrImagePull` or `ImagePullBackOff`?**
> 
> Ensure `imagePullPolicy: IfNotPresent` is set and the image name matches what's loaded in kind.

### Task 4.2: Create Tenant Namespaces

In a new terminal (or the same one after deployment):

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

> **Resources not being created?** Check the operator logs:
> ```bash
> kubectl logs -n postgres-operator-system deployment/postgres-operator-controller-manager
> ```

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

The solutions file already implements these patterns. This exercise explains the concepts so you understand what's happening.

### Task 5.1: Cluster-Scoped Owner Restrictions

Important: Cluster-scoped resources **cannot** use `OwnerReferences` to own namespace-scoped resources. The solutions file uses labels instead:

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

### Task 5.2: Cleanup with Finalizers

Since we can't use OwnerReferences for automatic garbage collection, the solutions file implements finalizers. Here's how they work:

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

### Task 5.3: Test Cleanup Behavior

```bash
# Ensure tenant-1 namespace exists (from earlier)
kubectl get namespace tenant-1 || kubectl create namespace tenant-1

# Create a ClusterDatabase
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: ClusterDatabase
metadata:
  name: test-cleanup
spec:
  targetNamespace: tenant-1
  tenant: tenant-1
  image: postgres:14
  replicas: 1
  databaseName: testdb
  username: admin
  storage:
    size: "5Gi"
EOF

# Verify resources were created
kubectl get statefulsets -n tenant-1

# Watch operator logs in another terminal to see cleanup happening
# kubectl logs -n postgres-operator-system deployment/postgres-operator-controller-manager -f

# Delete the ClusterDatabase
kubectl delete clusterdatabase test-cleanup

# Verify resources were cleaned up by the finalizer
kubectl get statefulsets -n tenant-1
# The StatefulSet should be deleted
```

## Cleanup

```bash
# Delete ClusterDatabases
kubectl delete clusterdatabases --all

# Delete test namespaces
kubectl delete namespace tenant-1 tenant-2

# (Optional) Undeploy operator
make undeploy
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
