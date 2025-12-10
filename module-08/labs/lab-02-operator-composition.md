# Lab 8.2: Composing Operators

**Related Lesson:** [Lesson 8.2: Operator Composition](../lessons/02-operator-composition.md)  
**Navigation:** [← Previous Lab: Multi-Tenancy](lab-01-multi-tenancy.md) | [Module Overview](../README.md) | [Next Lab: Stateful Applications →](lab-03-stateful-applications.md)

## Objectives

- Create dependent operators
- Implement operator coordination
- Use resource references
- Test operator composition

## Prerequisites

- Completion of [Lab 8.1](lab-01-multi-tenancy.md)
- Database operator ready
- Understanding of operator dependencies

## Exercise 1: Create Backup Operator

### Task 1.1: Scaffold Backup API with Kubebuilder

Use kubebuilder to scaffold the new Backup API. Since Backup is related to Database, we use the same `database` group:

```bash
# Navigate to your operator project
cd ~/postgres-operator

# Scaffold the Backup API (same group as Database)
kubebuilder create api \
  --group database \
  --version v1 \
  --kind Backup \
  --resource --controller

# When prompted:
# Create Resource [y/n]: y
# Create Controller [y/n]: y
```

> **Note:** We use `--group database` (same as Database) because both resources are part of the same operator. Using a different group would require enabling multi-group layout. See [kubebuilder multi-group docs](https://kubebuilder.io/migration/multi-group.html) if you need separate groups.

This creates:
- `api/v1/backup_types.go` - API type definitions
- `internal/controller/backup_controller.go` - Controller scaffold

### Task 1.2: Define Backup Spec and Status

Edit the generated `api/v1/backup_types.go` to add the spec and status fields:

```go
package v1

import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupSpec defines the desired state of Backup
type BackupSpec struct {
    // DatabaseRef references the Database to backup
    // +kubebuilder:validation:Required
    DatabaseRef corev1.LocalObjectReference `json:"databaseRef"`

    // Schedule is the cron schedule for automated backups (optional)
    // +optional
    Schedule string `json:"schedule,omitempty"`

    // Retention is the number of backups to retain
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:default=5
    // +optional
    Retention int `json:"retention,omitempty"`
}

// BackupStatus defines the observed state of Backup
type BackupStatus struct {
    // Phase is the current backup phase
    // +kubebuilder:validation:Enum=Pending;InProgress;Completed;Failed
    Phase string `json:"phase,omitempty"`

    // BackupTime is when the backup was created
    BackupTime *metav1.Time `json:"backupTime,omitempty"`

    // BackupLocation is where the backup is stored
    BackupLocation string `json:"backupLocation,omitempty"`

    // Conditions represent the latest observations
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Database",type="string",JSONPath=".spec.databaseRef.name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Backup is the Schema for the backups API
type Backup struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   BackupSpec   `json:"spec,omitempty"`
    Status BackupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BackupList contains a list of Backup
type BackupList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []Backup `json:"items"`
}

func init() {
    SchemeBuilder.Register(&Backup{}, &BackupList{})
}
```

### Task 1.3: Generate and Install CRD

```bash
# Generate code and CRD manifests
make generate
make manifests

# Install CRDs
make install

# Verify the CRD was created (same group as Database)
kubectl get crd backups.database.example.com
```

### Task 1.4: Implement Backup Controller

The Backup controller needs several functions to work properly. Rather than writing it from scratch, copy the complete implementation from the solutions file:

```bash
# Copy the complete controller implementation
cp path/to/solutions/backup-operator.go internal/controller/backup_controller.go
```

Or, if you prefer to type it yourself, copy the complete controller from:
**[solutions/backup-operator.go](../solutions/backup-operator.go)**

The complete controller includes:
- `Reconcile()` - Main reconciliation loop (shown below)
- `performBackup()` - Updates status and triggers backup
- `createBackup()` - Performs the actual backup operation
- `SetupWithManager()` - Registers controller with manager

**Key reconciliation logic:**

```go
func (r *BackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    backup := &databasev1.Backup{}
    if err := r.Get(ctx, req.NamespacedName, backup); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }
    
    // Get Database
    db := &databasev1.Database{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      backup.Spec.DatabaseRef.Name,
        Namespace: backup.Namespace,
    }, db)
    
    if errors.IsNotFound(err) {
        // Database not found, wait
        return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
    }
    if err != nil {
        return ctrl.Result{}, err
    }
    
    // Check if database is ready
    if db.Status.Phase != "Ready" {
        return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
    }
    
    // Perform backup
    return r.performBackup(ctx, db, backup)
}
```

### Task 1.5: Build and Verify

```bash
# Ensure the code compiles
make build

# If there are any compilation errors, verify you copied the complete
# controller from the solutions file
```

## Exercise 2: Coordinate Operators

### Task 2.1: Add Backup Reference to Database

Update Database spec:

```go
type DatabaseSpec struct {
    // ... existing fields ...
    BackupRef *corev1.LocalObjectReference `json:"backupRef,omitempty"`
}
```

### Task 2.2: Check Backup Status

Update Database controller:

```go
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    db := &databasev1.Database{}
    if err := r.Get(ctx, req.NamespacedName, db); err != nil {
        return ctrl.Result{}, err
    }
    
    // Check if backup is required
    if db.Spec.BackupRef != nil {
        backup := &databasev1.Backup{}
        err := r.Get(ctx, client.ObjectKey{
            Name:      db.Spec.BackupRef.Name,
            Namespace: db.Namespace,
        }, backup)
        
        if err != nil {
            return ctrl.Result{}, err
        }
        
        // Wait for backup to be ready
        if backup.Status.Phase != "Completed" {
            return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
        }
    }
    
    // Continue with database reconciliation
    return r.reconcileDatabase(ctx, db)
}
```

## Exercise 3: Use Status Conditions

### Task 3.1: Set Condition in Backup

```go
func (r *BackupReconciler) performBackup(ctx context.Context, db *databasev1.Database, backup *databasev1.Backup) (ctrl.Result, error) {
    // Perform backup...
    
    // Set condition
    meta.SetStatusCondition(&backup.Status.Conditions, metav1.Condition{
        Type:    "BackupReady",
        Status:  metav1.ConditionTrue,
        Reason:  "BackupCompleted",
        Message: "Backup completed successfully",
    })
    
    backup.Status.Phase = "Completed"
    return ctrl.Result{}, r.Status().Update(ctx, backup)
}
```

### Task 3.2: Check Condition in Database

```go
func (r *DatabaseReconciler) checkBackupCondition(ctx context.Context, db *databasev1.Database) error {
    if db.Spec.BackupRef == nil {
        return nil
    }
    
    backup := &databasev1.Backup{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Spec.BackupRef.Name,
        Namespace: db.Namespace,
    }, backup)
    
    if err != nil {
        return err
    }
    
    condition := meta.FindStatusCondition(backup.Status.Conditions, "BackupReady")
    if condition == nil || condition.Status != metav1.ConditionTrue {
        return fmt.Errorf("backup not ready")
    }
    
    return nil
}
```

## Exercise 4: Test Operator Composition

### Task 4.1: Create Database with Backup

```bash
# Create Database
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: my-database
spec:
  image: postgres:14
  replicas: 1
  backupRef:
    name: my-database-backup
EOF

# Create Backup
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Backup
metadata:
  name: my-database-backup
spec:
  databaseRef:
    name: my-database
  schedule: "0 2 * * *"
EOF
```

### Task 4.2: Verify Coordination

```bash
# Check Database status
kubectl get database my-database -o yaml

# Check Backup status
kubectl get backup my-database-backup -o yaml

# Verify operators coordinate
kubectl logs -l control-plane=controller-manager | grep -i backup
```

## Cleanup

```bash
# Delete test resources
kubectl delete databases --all
kubectl delete backups --all
```

## Lab Summary

In this lab, you:
- Scaffolded a new Backup API using kubebuilder
- Implemented backup operator with coordination logic
- Used resource references between operators
- Tested operator composition

## Key Learnings

1. **Use kubebuilder to scaffold new APIs** - `kubebuilder create api` handles boilerplate
2. **Operators can depend on each other** - Backup depends on Database
3. **Resource references link operators** - `DatabaseRef` connects Backup to Database
4. **Status conditions coordinate state** - `BackupReady` condition for cross-operator checks
5. **Dependency management is important** - Wait for dependencies before proceeding
6. **Composition enables complex applications** - Multiple operators working together

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Backup Types](../solutions/backup_types.go) - Complete API type definitions
- [Backup Operator](../solutions/backup-operator.go) - Complete backup controller
- [Operator Coordination](../solutions/operator-coordination.go) - Coordination examples

## Next Steps

Now let's learn about managing stateful applications!

**Navigation:** [← Previous Lab: Multi-Tenancy](lab-01-multi-tenancy.md) | [Related Lesson](../lessons/02-operator-composition.md) | [Next Lab: Stateful Applications →](lab-03-stateful-applications.md)

