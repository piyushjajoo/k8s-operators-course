---
layout: default
title: "Lab 04.3: Watching Indexing"
nav_order: 13
parent: "Module 4: Advanced Reconciliation"
grand_parent: Modules
mermaid: true
---

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

We already have modifed `SetupWithManager` to watch owned resources:

```go
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Owns(&appsv1.StatefulSet{}).  // Watch owned StatefulSets
        Owns(&corev1.Service{}).      // Watch owned Services
        Owns(&corev1.Secret{}).       // Watch owned Secrets
        Complete(r)
}
```

### Task 1.2: Test Watch Behavior

```bash
# Install and run operator
make install
make run

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

# Manually delete StatefulSet
kubectl delete statefulset test-db

# Watch operator logs - should detect and recreate

# Validate the deleted statefulset appears
kubectl get statefulset test-db

# Delete the database
kubectl delete database test-db
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
        Owns(&corev1.Service{}).
        // deliberately removing Owns(&corev1.Secret{}). to demonstrate non-owned resources
        Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findDatabasesForSecret),
		).
        Complete(r)
}

func (r *DatabaseReconciler) findDatabasesForSecret(ctx context.Context, secret client.Object) []reconcile.Request {
	databases := &databasev1.DatabaseList{}
	r.List(context.Background(), databases)

	var requests []reconcile.Request
	for _, db := range databases.Items {
		// If Database references this Secret
		if r.secretName(&db) == secret.GetName() &&
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
# Install and run operator
make install
make run

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

# Update the Secret
kubectl patch secret test-db-credentials --type merge -p '{"data":{"password":"newpassword"}}'

# Watch operator logs - should reconcile Database
```

## Exercise 3: Create Indexes

Indexes allow efficient lookups of resources by field values without scanning all objects.

### Task 3.1: Set Up Index

Add an index for the `image` field to quickly find all Databases using a specific PostgreSQL version:

```go
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    // Create index for image field
    if err := mgr.GetFieldIndexer().IndexField(
        context.Background(),
        &databasev1.Database{},
        "spec.image",
        func(obj client.Object) []string {
            db, ok := obj.(*databasev1.Database)
            if !ok {
                return nil
            }
            if db.Spec.Image != "" {
                return []string{db.Spec.Image}
            }
            return nil
        },
    ); err != nil {
        return err
    }
    
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Owns(&appsv1.StatefulSet{}).
        Owns(&corev1.Service{}).
        Watches(
            &corev1.Secret{},
            handler.EnqueueRequestsFromMapFunc(r.findDatabasesForSecret),
        ).
        Complete(r)
}
```

### Task 3.2: Use Index in Query

Use the index to efficiently find all Databases using a specific image:

```go
// findDatabasesByImage finds all Databases using a specific PostgreSQL image
func (r *DatabaseReconciler) findDatabasesByImage(ctx context.Context, image string) ([]databasev1.Database, error) {
    databases := &databasev1.DatabaseList{}
    err := r.List(ctx, databases, client.MatchingFields{
        "spec.image": image,
    })
    
    if err != nil {
        return nil, err
    }
    
    return databases.Items, nil
}
```

### Task 3.3: Test Index Usage

```bash
# Install and run operator
make install
make run

# Create databases with different images
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: db-postgres14
spec:
  image: postgres:14
  replicas: 1
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
---
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: db-postgres15
spec:
  image: postgres:15
  replicas: 1
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
---
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: db-postgres14-2
spec:
  image: postgres:14
  replicas: 1
  databaseName: testdb
  username: admin
  storage:
    size: 5Gi
EOF

# The index allows efficient lookup - finding all postgres:14 databases
# doesn't require scanning every Database object
```

> **Note:** Indexes are particularly useful when you have many resources and need to find subsets quickly. Without an index, `List()` with field matching would need to scan all objects.

## Exercise 4: Event Predicates

### Task 4.1: Add Predicates

Filter events to only reconcile on important changes. 

> **Important:** When filtering StatefulSet updates, you must include **both** spec changes (Generation) AND status changes (ReadyReplicas). Otherwise, the Database will never become Ready because status updates will be filtered out!

```go
import (
    "sigs.k8s.io/controller-runtime/pkg/builder"
    "sigs.k8s.io/controller-runtime/pkg/predicate"
    "sigs.k8s.io/controller-runtime/pkg/event"
)

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Owns(&appsv1.StatefulSet{}, builder.WithPredicates(predicate.Funcs{
            UpdateFunc: func(e event.UpdateEvent) bool {
                oldSS := e.ObjectOld.(*appsv1.StatefulSet)
                newSS := e.ObjectNew.(*appsv1.StatefulSet)
                // Reconcile on spec changes (Generation) OR status changes (ReadyReplicas)
                // Without checking ReadyReplicas, Database status would never update to Ready!
                return oldSS.Generation != newSS.Generation ||
                    oldSS.Status.ReadyReplicas != newSS.Status.ReadyReplicas
            },
            CreateFunc: func(e event.CreateEvent) bool {
                return true
            },
            DeleteFunc: func(e event.DeleteEvent) bool {
                return true
            },
        })).
        Owns(&corev1.Service{}).
        Watches(
            &corev1.Secret{},
            handler.EnqueueRequestsFromMapFunc(r.findDatabasesForSecret),
        ).
        Complete(r)
}
```

## Exercise 5: Test Watch Performance

### Task 5.1: Create Multiple Resources

```bash
# Install and run operator
make install
make run

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

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Watch Setup](../solutions/watch-setup.go) - Examples of setting up watches for owned and non-owned resources

## Next Steps

Now let's implement advanced patterns like multi-phase reconciliation!

**Navigation:** [← Previous Lab: Finalizers](lab-02-finalizers-cleanup.md) | [Related Lesson](../lessons/03-watching-indexing.md) | [Next Lab: Advanced Patterns →](lab-04-advanced-patterns.md)
