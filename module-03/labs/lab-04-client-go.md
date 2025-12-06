# Lab 3.4: Advanced Client Operations

**Related Lesson:** [Lesson 3.4: Working with Client-Go](../lessons/04-client-go.md)  
**Navigation:** [← Previous Lab: Reconciliation Logic](lab-03-reconciliation-logic.md) | [Module Overview](../README.md)

## Objectives

- Use advanced client operations
- Implement watches for dependent resources
- Use strategic merge patches
- Handle conflicts with retries

## Prerequisites

- Completion of [Lab 3.3](lab-03-reconciliation-logic.md)
- PostgreSQL operator from previous lab
- Understanding of client operations

## Exercise 1: List Resources with Filters

### Task 1.1: List by Namespace

Add a function to list all databases in a namespace:

```go
func (r *DatabaseReconciler) listDatabasesInNamespace(ctx context.Context, namespace string) (*databasev1.DatabaseList, error) {
    list := &databasev1.DatabaseList{}
    err := r.List(ctx, list, client.InNamespace(namespace))
    return list, err
}
```

### Task 1.2: List by Labels

```go
func (r *DatabaseReconciler) listDatabasesByLabel(ctx context.Context, labels map[string]string) (*databasev1.DatabaseList, error) {
    list := &databasev1.DatabaseList{}
    err := r.List(ctx, list, client.MatchingLabels(labels))
    return list, err
}
```

## Exercise 2: Implement Strategic Merge Patch

### Task 2.1: Patch StatefulSet Replicas

Instead of full update, use patch:

```go
func (r *DatabaseReconciler) patchStatefulSetReplicas(ctx context.Context, statefulSet *appsv1.StatefulSet, replicas int32) error {
    patch := client.MergeFrom(statefulSet.DeepCopy())
    statefulSet.Spec.Replicas = &replicas
    return r.Patch(ctx, statefulSet, patch)
}
```

### Task 2.2: Use in Reconciliation

Update your reconcileStatefulSet to use patch when only replicas change:

```go
// If only replicas changed, use patch
if statefulSet.Spec.Replicas != desiredStatefulSet.Spec.Replicas {
    return r.patchStatefulSetReplicas(ctx, statefulSet, *desiredStatefulSet.Spec.Replicas)
}
```

## Exercise 3: Handle Conflicts

### Task 3.1: Implement Retry Logic

Add a helper function for conflict retries:

```go
func (r *DatabaseReconciler) updateWithRetry(ctx context.Context, obj client.Object, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := r.Update(ctx, obj)
        if err == nil {
            return nil
        }
        
        if !errors.IsConflict(err) {
            return err
        }
        
        // Conflict - get fresh version and retry
        key := client.ObjectKeyFromObject(obj)
        if err := r.Get(ctx, key, obj); err != nil {
            return err
        }
        
        time.Sleep(100 * time.Millisecond)
    }
    return fmt.Errorf("max retries exceeded")
}
```

### Task 3.2: Use in Reconciliation

```go
// Use retry logic for updates
if err := r.updateWithRetry(ctx, statefulSet, 3); err != nil {
    return err
}
```

## Exercise 4: Watch Dependent Resources

### Task 4.1: Set Up Watch

Modify SetupWithManager to watch StatefulSets:

```go
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Owns(&appsv1.StatefulSet{}).  // Watch owned StatefulSets
        Owns(&corev1.Service{}).      // Watch owned Services
        Owns(&corev1.Secret{}).       // Watch owner Secrets
        Complete(r)
}
```

### Task 4.2: Handle Watch Events

When StatefulSet changes, Database will be reconciled automatically!

## Exercise 5: Field Selectors

### Task 5.1: Find Databases by Owner

```go
func (r *DatabaseReconciler) findDatabasesByOwner(ctx context.Context, ownerName string) (*databasev1.DatabaseList, error) {
    list := &databasev1.DatabaseList{}
    err := r.List(ctx, list, client.MatchingFields{
        ".metadata.ownerReferences[0].name": ownerName,
    })
    return list, err
}
```

## Exercise 6: Test Advanced Operations

### Task 6.1: Test Patch

```bash
# Create database
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: my-database
spec:
  image: postgres:14
  replicas: 1
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# Update replicas using patch (simulate)
kubectl patch database my-database --type merge -p '{"spec":{"replicas":2}}'

# Watch operator logs to see patch in action

# Validate 2 replicas are available
kubectl get database my-database -o jsonpath='{.spec.replicas}'
kubectl get statefulset my-database
```

### Task 6.2: Test Conflict Handling

```bash
# Quickly update multiple times to trigger conflicts
kubectl patch database my-database --type merge -p '{"spec":{"replicas":3}}'
kubectl patch database my-database --type merge -p '{"spec":{"replicas":4}}'
kubectl patch database my-database --type merge -p '{"spec":{"replicas":5}}'

# Observe how operator handles conflicts

# Validate 5 replicas are eventually available
kubectl get database my-database -o jsonpath='{.spec.replicas}'
kubectl get statefulset my-database
```

### Task 6.3: Test Watch

```bash
# Manually delete StatefulSet
kubectl delete statefulset my-database

# Watch operator logs - should detect and recreate
kubectl get statefulset my-database
```

## Cleanup

```bash
# Delete test resources
kubectl delete databases --all

# Validate that all the resources are gone
kubectl get databases my-database
kubectl get statefulset my-database
kubectl get service my-database
kubectl get secret my-database-credentials
```

## Lab Summary

In this lab, you:
- Used advanced list operations with filters
- Implemented strategic merge patches
- Added conflict retry logic
- Set up watches for dependent resources
- Used field selectors
- Tested all operations

## Key Learnings

1. List operations can be filtered efficiently
2. Patches are better for partial updates
3. Conflicts need retry logic
4. Watches enable reactive reconciliation
5. Field selectors provide powerful queries
6. Advanced operations improve operator efficiency

## Congratulations!

You've completed Module 3! You now understand:
- Controller-runtime architecture
- API design principles
- Reconciliation logic
- Advanced client operations

In Module 4, you'll learn advanced patterns like conditions, finalizers, and multi-phase reconciliation.

**Navigation:** [← Previous Lab: Reconciliation Logic](lab-03-reconciliation-logic.md) | [Related Lesson](../lessons/04-client-go.md) | [Module Overview](../README.md)

