# Lab 4.3: Setting Up Watches and Indexes

**Related Lesson:** [Lesson 4.3: Watching and Indexing](../lessons/03-watching-indexing.md)  
**Navigation:** [← Previous Lab: Finalizers](lab-02-finalizers-cleanup.md) | [Module Overview](../README.md) | [Next Lab: Advanced Patterns →](lab-04-advanced-patterns.md)

## Objectives

- Set up watches for dependent resources
- Create indexes for efficient lookups
- Handle watch events
- Test watch behavior

## Prerequisites

- Completion of [Lab 4.2](lab-02-finalizers-cleanup.md)
- Database operator with finalizers
- Understanding of watching patterns

## Exercise 1: Watch Owned Resources

### Task 1.1: Update SetupWithManager

Modify `SetupWithManager` to watch owned resources:

```go
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Owns(&appsv1.StatefulSet{}).  // Watch owned StatefulSets
        Owns(&corev1.Service{}).      // Watch owned Services
        Complete(r)
}
```

### Task 1.2: Test Watch Behavior

```bash
# Install and run operator
make install
make run

# Create Database
kubectl apply -f database.yaml

# Manually delete StatefulSet
kubectl delete statefulset test-db

# Watch operator logs - should detect and recreate
```

## Exercise 2: Watch Non-Owned Resources

### Task 2.1: Watch Secrets

Add watch for Secrets that Databases reference:

```go
import (
    "sigs.k8s.io/controller-runtime/pkg/handler"
    "sigs.k8s.io/controller-runtime/pkg/source"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
    "k8s.io/apimachinery/pkg/types"
)

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Owns(&appsv1.StatefulSet{}).
        Watches(
            &source.Kind{Type: &corev1.Secret{}},
            handler.EnqueueRequestsFromMapFunc(r.findDatabasesForSecret),
        ).
        Complete(r)
}

func (r *DatabaseReconciler) findDatabasesForSecret(secret client.Object) []reconcile.Request {
    databases := &databasev1.DatabaseList{}
    r.List(context.Background(), databases)
    
    var requests []reconcile.Request
    for _, db := range databases.Items {
        // If Database references this Secret
        if db.Spec.SecretName == secret.GetName() && 
           db.Namespace == secret.GetNamespace() {
            requests = append(requests, reconcile.Request{
                NamespacedName: types.NamespacedName{
                    Name:      db.Name,
                    Namespace: db.Namespace,
                },
            })
        }
    }
    return requests
}
```

### Task 2.2: Test Secret Watch

```bash
# Create Database that references a Secret
kubectl apply -f database-with-secret.yaml

# Update the Secret
kubectl patch secret db-secret --type merge -p '{"data":{"password":"newpassword"}}'

# Watch operator logs - should reconcile Database
```

## Exercise 3: Create Indexes

### Task 3.1: Set Up Index

Add index for Secret name:

```go
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    // Create index for Secret name
    if err := mgr.GetFieldIndexer().IndexField(
        context.Background(),
        &databasev1.Database{},
        "spec.secretName",
        func(obj client.Object) []string {
            db, ok := obj.(*databasev1.Database)
            if !ok {
                return nil
            }
            if db.Spec.SecretName != "" {
                return []string{db.Spec.SecretName}
            }
            return nil
        },
    ); err != nil {
        return err
    }
    
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Complete(r)
}
```

### Task 3.2: Use Index in Query

```go
// Find all Databases that use a specific Secret
func (r *DatabaseReconciler) findDatabasesForSecret(secret client.Object) []reconcile.Request {
    databases := &databasev1.DatabaseList{}
    err := r.List(context.Background(), databases, client.MatchingFields{
        "spec.secretName": secret.GetName(),
    })
    
    if err != nil {
        return nil
    }
    
    var requests []reconcile.Request
    for _, db := range databases.Items {
        if db.Namespace == secret.GetNamespace() {
            requests = append(requests, reconcile.Request{
                NamespacedName: types.NamespacedName{
                    Name:      db.Name,
                    Namespace: db.Namespace,
                },
            })
        }
    }
    return requests
}
```

## Exercise 4: Event Predicates

### Task 4.1: Add Predicates

Filter events to only reconcile on important changes:

```go
import (
    "sigs.k8s.io/controller-runtime/pkg/predicate"
    "sigs.k8s.io/controller-runtime/pkg/event"
)

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Owns(&appsv1.StatefulSet{}, builder.WithPredicates(predicate.Funcs{
            UpdateFunc: func(e event.UpdateEvent) bool {
                // Only reconcile on spec changes
                oldSS := e.ObjectOld.(*appsv1.StatefulSet)
                newSS := e.ObjectNew.(*appsv1.StatefulSet)
                return oldSS.Generation != newSS.Generation
            },
            CreateFunc: func(e event.CreateEvent) bool {
                return true
            },
            DeleteFunc: func(e event.DeleteEvent) bool {
                return true
            },
        })).
        Complete(r)
}
```

## Exercise 5: Test Watch Performance

### Task 5.1: Create Multiple Resources

```bash
# Create multiple Databases
for i in {1..10}; do
  kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: db-$i
spec:
  image: postgres:14
  replicas: 1
  databaseName: db$i
  username: admin
  storage:
    size: 10Gi
EOF
done
```

### Task 5.2: Observe Watch Behavior

```bash
# Watch operator logs
# Should see efficient reconciliation

# Update one Database
kubectl patch database db-5 --type merge -p '{"spec":{"replicas":2}}'

# Only db-5 should be reconciled
```

## Cleanup

```bash
# Delete all test resources
kubectl delete databases --all
```

## Lab Summary

In this lab, you:
- Set up watches for owned resources
- Watched non-owned resources
- Created indexes for efficient lookups
- Added event predicates
- Tested watch performance

## Key Learnings

1. Watch owned resources with `Owns()`
2. Watch non-owned resources with `Watches()`
3. Indexes enable fast lookups
4. Event predicates filter events
5. Watches make controllers reactive
6. Proper watching improves performance

## Next Steps

Now let's implement advanced patterns like multi-phase reconciliation!

**Navigation:** [← Previous Lab: Finalizers](lab-02-finalizers-cleanup.md) | [Related Lesson](../lessons/03-watching-indexing.md) | [Next Lab: Advanced Patterns →](lab-04-advanced-patterns.md)

