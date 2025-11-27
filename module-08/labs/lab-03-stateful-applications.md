# Lab 8.3: Managing Stateful Applications

**Related Lesson:** [Lesson 8.3: Stateful Application Management](../lessons/03-stateful-applications.md)  
**Navigation:** [← Previous Lab: Operator Composition](lab-02-operator-composition.md) | [Module Overview](../README.md) | [Next Lab: Final Project →](lab-04-final-project.md)

## Objectives

- Implement backup functionality
- Add restore capability
- Handle rolling updates
- Ensure data consistency

## Prerequisites

- Completion of [Lab 8.2](lab-02-operator-composition.md)
- Database operator with backup operator
- Understanding of StatefulSets

## Exercise 1: Implement Backup Functionality

### Task 1.1: Add Backup Method

Create `internal/backup/backup.go`:

```go
package backup

import (
    "context"
    "fmt"
    "time"
)

func PerformBackup(ctx context.Context, db *databasev1.Database) (string, error) {
    // Connect to database
    conn, err := connectToDatabase(db)
    if err != nil {
        return "", err
    }
    defer conn.Close()
    
    // Create backup
    backupFile := fmt.Sprintf("/backups/%s-%s.sql", db.Name, time.Now().Format("20060102-150405"))
    
    // Perform pg_dump or equivalent
    cmd := exec.CommandContext(ctx, "pg_dump", "-h", db.Status.Endpoint, "-U", db.Spec.Username, db.Spec.DatabaseName)
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    
    // Save to storage
    if err := saveToStorage(backupFile, output); err != nil {
        return "", err
    }
    
    return backupFile, nil
}
```

### Task 1.2: Integrate with Controller

```go
func (r *BackupReconciler) performBackup(ctx context.Context, db *databasev1.Database, backup *backupv1.Backup) (ctrl.Result, error) {
    // Perform backup
    location, err := backup.PerformBackup(ctx, db)
    if err != nil {
        backup.Status.Phase = "Failed"
        r.Status().Update(ctx, backup)
        return ctrl.Result{}, err
    }
    
    // Update status
    backup.Status.Phase = "Completed"
    backup.Status.BackupTime = metav1.Now()
    backup.Status.BackupLocation = location
    
    return ctrl.Result{}, r.Status().Update(ctx, backup)
}
```

## Exercise 2: Implement Restore

### Task 2.1: Add Restore Method

```go
func PerformRestore(ctx context.Context, db *databasev1.Database, backupLocation string) error {
    // Connect to database
    conn, err := connectToDatabase(db)
    if err != nil {
        return err
    }
    defer conn.Close()
    
    // Load backup from storage
    backupData, err := loadFromStorage(backupLocation)
    if err != nil {
        return err
    }
    
    // Restore database
    cmd := exec.CommandContext(ctx, "psql", "-h", db.Status.Endpoint, "-U", db.Spec.Username, db.Spec.DatabaseName)
    cmd.Stdin = bytes.NewReader(backupData)
    
    if err := cmd.Run(); err != nil {
        return err
    }
    
    return nil
}
```

### Task 2.2: Create Restore CRD

```go
type RestoreSpec struct {
    DatabaseRef corev1.LocalObjectReference `json:"databaseRef"`
    BackupRef   corev1.LocalObjectReference  `json:"backupRef"`
}

type RestoreStatus struct {
    Phase       string    `json:"phase,omitempty"`
    RestoreTime time.Time `json:"restoreTime,omitempty"`
}
```

## Exercise 3: Handle Rolling Updates

### Task 3.1: Update StatefulSet Safely

```go
func (r *DatabaseReconciler) updateStatefulSet(ctx context.Context, db *databasev1.Database) error {
    statefulSet := &appsv1.StatefulSet{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet)
    
    if err != nil {
        return err
    }
    
    // Check if update needed
    desiredImage := db.Spec.Image
    currentImage := statefulSet.Spec.Template.Spec.Containers[0].Image
    
    if desiredImage != currentImage {
        // Update image
        statefulSet.Spec.Template.Spec.Containers[0].Image = desiredImage
        
        // Update StatefulSet (will trigger rolling update)
        if err := r.Update(ctx, statefulSet); err != nil {
            return err
        }
        
        // Wait for update to complete
        return r.waitForRollingUpdate(ctx, statefulSet)
    }
    
    return nil
}

func (r *DatabaseReconciler) waitForRollingUpdate(ctx context.Context, ss *appsv1.StatefulSet) error {
    return wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
        err := r.Get(ctx, client.ObjectKeyFromObject(ss), ss)
        if err != nil {
            return false, err
        }
        
        return ss.Status.UpdatedReplicas == *ss.Spec.Replicas, nil
    })
}
```

## Exercise 4: Ensure Data Consistency

### Task 4.1: Verify Consistency

```go
func (r *DatabaseReconciler) ensureDataConsistency(ctx context.Context, db *databasev1.Database) error {
    statefulSet := &appsv1.StatefulSet{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet)
    
    if err != nil {
        return err
    }
    
    // Check all replicas are ready
    if statefulSet.Status.ReadyReplicas != *statefulSet.Spec.Replicas {
        return fmt.Errorf("not all replicas ready: %d/%d", 
            statefulSet.Status.ReadyReplicas, *statefulSet.Spec.Replicas)
    }
    
    // Perform consistency check
    return r.performConsistencyCheck(ctx, db)
}

func (r *DatabaseReconciler) performConsistencyCheck(ctx context.Context, db *databasev1.Database) error {
    // Connect to each replica and verify data consistency
    // This is application-specific
    return nil
}
```

## Exercise 5: Test Backup and Restore

### Task 5.1: Create Backup

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
EOF

# Create Backup
kubectl apply -f - <<EOF
apiVersion: backup.example.com/v1
kind: Backup
metadata:
  name: test-backup
spec:
  databaseRef:
    name: test-db
EOF
```

### Task 5.2: Test Restore

```bash
# Create Restore
kubectl apply -f - <<EOF
apiVersion: restore.example.com/v1
kind: Restore
metadata:
  name: test-restore
spec:
  databaseRef:
    name: test-db
  backupRef:
    name: test-backup
EOF

# Verify restore
kubectl get restore test-restore -o yaml
```

## Cleanup

```bash
# Delete test resources
kubectl delete databases --all
kubectl delete backups --all
kubectl delete restores --all
```

## Lab Summary

In this lab, you:
- Implemented backup functionality
- Added restore capability
- Handled rolling updates
- Ensured data consistency

## Key Learnings

1. Backups protect data from loss
2. Restores recover from backups
3. Rolling updates update without downtime
4. Data consistency ensures correctness
5. StatefulSets provide ordered pods
6. Persistent volumes maintain data

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Backup Implementation](../solutions/backup.go) - Complete backup functionality
- [Restore Implementation](../solutions/restore.go) - Complete restore functionality
- [Rolling Update](../solutions/rolling-update.go) - Rolling update handling

## Next Steps

Now let's build the final project!

**Navigation:** [← Previous Lab: Operator Composition](lab-02-operator-composition.md) | [Related Lesson](../lessons/03-stateful-applications.md) | [Next Lab: Final Project →](lab-04-final-project.md)

