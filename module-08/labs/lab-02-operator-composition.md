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

### Task 1.1: Create Backup CRD

Create `api/v1/backup_types.go`:

```go
package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type Backup struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec              BackupSpec   `json:"spec,omitempty"`
    Status            BackupStatus `json:"status,omitempty"`
}

type BackupSpec struct {
    DatabaseRef corev1.LocalObjectReference `json:"databaseRef"`
    Schedule    string                      `json:"schedule,omitempty"`
    Retention   int                         `json:"retention,omitempty"`
}

type BackupStatus struct {
    Phase          string    `json:"phase,omitempty"`
    BackupTime     time.Time `json:"backupTime,omitempty"`
    BackupLocation string    `json:"backupLocation,omitempty"`
}
```

### Task 1.2: Create Backup Controller

Create `internal/controller/backup_controller.go`:

```go
func (r *BackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    backup := &backupv1.Backup{}
    if err := r.Get(ctx, req.NamespacedName, backup); err != nil {
        return ctrl.Result{}, err
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
    
    // Perform backup
    return r.performBackup(ctx, db, backup)
}
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
        backup := &backupv1.Backup{}
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
func (r *BackupReconciler) performBackup(ctx context.Context, db *databasev1.Database, backup *backupv1.Backup) (ctrl.Result, error) {
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
    
    backup := &backupv1.Backup{}
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
apiVersion: backup.example.com/v1
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
- Created backup operator
- Implemented operator coordination
- Used resource references
- Tested operator composition

## Key Learnings

1. Operators can depend on each other
2. Resource references link operators
3. Status conditions coordinate state
4. Operators coordinate through resources
5. Dependency management is important
6. Composition enables complex applications

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Backup Operator](../solutions/backup-operator.go) - Complete backup operator
- [Operator Coordination](../solutions/operator-coordination.go) - Coordination examples

## Next Steps

Now let's learn about managing stateful applications!

**Navigation:** [← Previous Lab: Multi-Tenancy](lab-01-multi-tenancy.md) | [Related Lesson](../lessons/02-operator-composition.md) | [Next Lab: Stateful Applications →](lab-03-stateful-applications.md)

