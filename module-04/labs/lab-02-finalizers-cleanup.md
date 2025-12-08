# Lab 4.2: Implementing Finalizers

**Related Lesson:** [Lesson 4.2: Finalizers and Cleanup](../lessons/02-finalizers-cleanup.md)  
**Navigation:** [← Previous Lab: Conditions](lab-01-conditions-status.md) | [Module Overview](../README.md) | [Next Lab: Watching →](lab-03-watching-indexing.md)

## Objectives

- Add finalizers to Database operator
- Implement cleanup logic
- Handle graceful deletion
- Test cleanup scenarios

## Prerequisites

- Completion of [Lab 4.1](lab-01-conditions-status.md)
- Database operator with conditions
- Understanding of finalizers

## Exercise 1: Add Finalizer on Creation

### Task 1.1: Add Finalizer Logic

Modify `Reconcile` function to add finalizer:

```go
import (
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
    finalizerName := "database.example.com/finalizer"
)

func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    
    db := &databasev1.Database{}
    if err := r.Get(ctx, req.NamespacedName, db); err != nil {
        return ctrl.Result{}, err
    }
    
    // Add finalizer if not present
    if !controllerutil.ContainsFinalizer(db, finalizerName) {
        controllerutil.AddFinalizer(db, finalizerName)
        if err := r.Update(ctx, db); err != nil {
            return ctrl.Result{}, err
        }
        logger.Info("Added finalizer", "name", db.Name)
    }
    
    // Check if resource is being deleted
    if !db.DeletionTimestamp.IsZero() {
        // Resource is being deleted
        return r.handleDeletion(ctx, db)
    }
    
    // Normal reconciliation
    // ... existing reconciliation logic ...
}
```

## Exercise 2: Implement Cleanup Logic

### Task 2.1: Create Cleanup Function

Add cleanup function:

```go
func (r *DatabaseReconciler) handleDeletion(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    
    // Check if finalizer exists
    if !controllerutil.ContainsFinalizer(db, finalizerName) {
        return ctrl.Result{}, nil
    }
    
    logger.Info("Handling deletion", "name", db.Name)
    
    // Perform cleanup operations
    if err := r.cleanupExternalResources(ctx, db); err != nil {
        logger.Error(err, "Failed to cleanup external resources")
        r.setCondition(db, "Ready", metav1.ConditionFalse, "CleanupFailed", err.Error())
        r.Status().Update(ctx, db)
        // Retry after delay
        return ctrl.Result{RequeueAfter: 10 * time.Second}, err
    }
    
    // Cleanup successful, remove finalizer
    controllerutil.RemoveFinalizer(db, finalizerName)
    if err := r.Update(ctx, db); err != nil {
        return ctrl.Result{}, err
    }
    
    logger.Info("Finalizer removed, resource will be deleted")
    return ctrl.Result{}, nil
}
```

### Task 2.2: Implement Cleanup

```go
func (r *DatabaseReconciler) cleanupExternalResources(ctx context.Context, db *databasev1.Database) error {
    logger := log.FromContext(ctx)
    
    // Delete StatefulSet if it exists
    statefulSet := &appsv1.StatefulSet{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet)
    
    if err == nil {
        // StatefulSet exists, delete it
        log.Info("Deleting StatefulSet", "name", statefulSet.Name)
        if err := r.Delete(ctx, statefulSet); err != nil && !errors.IsNotFound(err) {
            return fmt.Errorf("failed to delete StatefulSet: %w", err)
        }
        // Requeue to wait for deletion to complete
        return fmt.Errorf("waiting for StatefulSet to be deleted")
    } else if !errors.IsNotFound(err) {
        // Some other error occurred
        return fmt.Errorf("failed to get StatefulSet: %w", err)
    }
    
    // StatefulSet is gone, now cleanup Service
    service := &corev1.Service{}
    err = r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, service)
    
    if err == nil {
        log.Info("Deleting Service", "name", service.Name)
        if err := r.Delete(ctx, service); err != nil && !errors.IsNotFound(err) {
            return fmt.Errorf("failed to delete Service: %w", err)
        }
        return fmt.Errorf("waiting for Service to be deleted")
    } else if !errors.IsNotFound(err) {
        return fmt.Errorf("failed to get Service: %w", err)
    }
    
    // Cleanup Secret
    secret := &corev1.Secret{}
    err = r.Get(ctx, client.ObjectKey{
        Name:      r.secretName(db),
        Namespace: db.Namespace,
    }, secret)
    
    if err == nil {
        log.Info("Deleting Secret", "name", secret.Name)
        if err := r.Delete(ctx, secret); err != nil && !errors.IsNotFound(err) {
            return fmt.Errorf("failed to delete Secret: %w", err)
        }
        return fmt.Errorf("waiting for Secret to be deleted")
    } else if !errors.IsNotFound(err) {
        return fmt.Errorf("failed to get Secret: %w", err)
    }
    
    // Example: Delete backup in external system
    // if err := r.deleteBackup(ctx, db); err != nil {
    //     return err
    // }
    
    logger.Info("Cleanup completed")
    return nil
}
```

> **Important:** The cleanup function must **explicitly delete** child resources. While owner references enable automatic garbage collection when a parent is deleted, finalizers prevent the parent from being deleted until cleanup completes. This creates a deadlock if you only wait for resources to disappear - you must actively delete them.

## Exercise 3: Test Finalizers

### Task 3.1: Install and Run

```bash
# Install CRD
make install

# Run operator
make run
```

### Task 3.2: Create Database

```bash
# Create Database
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-db
spec:
  image: postgres:14
  replicas: 1
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# Verify finalizer was added
kubectl get database test-db -o jsonpath='{.metadata.finalizers}'
```

### Task 3.3: Delete Database

```bash
# Delete Database
kubectl delete database test-db

# Check deletion timestamp
kubectl get database test-db -o jsonpath='{.metadata.deletionTimestamp}'

# Resource should still exist (has finalizer)
kubectl get database test-db

# Watch operator logs - should see cleanup
```

### Task 3.4: Verify Cleanup

```bash
# Watch finalizer removal
watch -n 1 'kubectl get database test-db -o jsonpath="{.metadata.finalizers}"'

# After cleanup, resource should be deleted
kubectl get database test-db
```

## Exercise 4: Test Cleanup Failure

### Task 4.1: Simulate Cleanup Failure

Temporarily modify cleanup to always fail:

```go
func (r *DatabaseReconciler) cleanupExternalResources(ctx context.Context, db *databasev1.Database) error {
    return fmt.Errorf("simulated cleanup failure")
}
```

### Task 4.2: Test Behavior

```bash
# Create and delete Database
kubectl apply -f database.yaml
kubectl delete database test-db

# Resource should remain (cleanup failing)
kubectl get database test-db

# Check conditions
kubectl get database test-db -o jsonpath='{.status.conditions}'
```

## Exercise 5: Test Idempotent Cleanup

### Task 5.1: Make Cleanup Idempotent

```go
func (r *DatabaseReconciler) cleanupExternalResources(ctx context.Context, db *databasev1.Database) error {
    // Check if already cleaned up
    if r.isAlreadyCleanedUp(ctx, db) {
        return nil  // Already cleaned up, idempotent
    }
    
    // Perform cleanup
    return r.performCleanup(ctx, db)
}
```

## Cleanup

```bash
# Delete any remaining resources
kubectl delete databases --all
```

## Lab Summary

In this lab, you:
- Added finalizers to Database operator
- Implemented cleanup logic
- Handled graceful deletion
- Tested cleanup scenarios
- Made cleanup idempotent

## Key Learnings

1. Finalizers prevent deletion until cleanup is complete
2. Add finalizer early in reconciliation
3. Check DeletionTimestamp to detect deletion
4. **Explicitly delete child resources** - don't rely on owner reference cascade during finalizer cleanup (this causes a deadlock)
5. Perform cleanup before removing finalizer
6. Make cleanup idempotent
7. Handle cleanup failures gracefully

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Finalizer Handler](../solutions/finalizer-handler.go) - Complete finalizer implementation with cleanup logic

## Next Steps

Now let's set up watches and indexes for efficient controllers!

**Navigation:** [← Previous Lab: Conditions](lab-01-conditions-status.md) | [Related Lesson](../lessons/02-finalizers-cleanup.md) | [Next Lab: Watching →](lab-03-watching-indexing.md)

