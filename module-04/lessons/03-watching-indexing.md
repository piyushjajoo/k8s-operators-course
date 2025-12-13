---
layout: default
title: "04.3 Watching Indexing"
nav_order: 3
parent: "Module 4: Advanced Reconciliation"
grand_parent: Modules
mermaid: true
---

# Lesson 4.3: Watching and Indexing

**Navigation:** [← Previous: Finalizers and Cleanup](02-finalizers-cleanup.md) | [Module Overview](../README.md) | [Next: Advanced Patterns →](04-advanced-patterns.md)

## Introduction

In [Module 3](../../module-03/README.md), you learned basic reconciliation. Now let's optimize controllers by watching dependent resources and using indexes for efficient lookups. This makes controllers more reactive and performant.

## Watching Dependent Resources

Controllers can watch resources they don't own:

```mermaid
graph TB
    CONTROLLER[Controller]
    
    CONTROLLER --> WATCH1[Watch CustomResource]
    CONTROLLER --> WATCH2[Watch StatefulSet]
    CONTROLLER --> WATCH3[Watch Service]
    CONTROLLER --> WATCH4[Watch Secret]
    
    WATCH1 --> EVENT1[Event: CustomResource changed]
    WATCH2 --> EVENT2[Event: StatefulSet changed]
    WATCH3 --> EVENT3[Event: Service changed]
    WATCH4 --> EVENT4[Event: Secret changed]
    
    EVENT1 --> RECONCILE[Reconcile]
    EVENT2 --> RECONCILE
    EVENT3 --> RECONCILE
    EVENT4 --> RECONCILE
    
    style CONTROLLER fill:#FFB6C1
    style RECONCILE fill:#90EE90
```

## Watch Setup Flow

Here's how watches are set up:

```mermaid
sequenceDiagram
    participant Controller
    participant Manager
    participant Informer
    participant API as API Server
    
    Controller->>Manager: SetupWithManager
    Manager->>Informer: Create Informer
    Informer->>API: Watch Resources
    API->>Informer: Event Stream
    Informer->>Controller: Enqueue Request
    Controller->>Controller: Reconcile
```

## Setting Up Watches

### Watch Owned Resources

```go
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Owns(&appsv1.StatefulSet{}).  // Watch owned StatefulSets
        Owns(&corev1.Service{}).      // Watch owned Services
        Complete(r)
}
```

When owned resources change, the owner is reconciled automatically.

### Watch Non-Owned Resources

```go
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Watches(
            &corev1.Secret{},
            handler.EnqueueRequestsFromMapFunc(r.findDatabasesForSecret),
        ).
        Complete(r)
}

// secretName returns the generated Secret name for a Database
func (r *DatabaseReconciler) secretName(db *databasev1.Database) string {
    return fmt.Sprintf("%s-credentials", db.Name)
}

func (r *DatabaseReconciler) findDatabasesForSecret(ctx context.Context, secret client.Object) []reconcile.Request {
    // Find all Databases that use this Secret
    // Secret name is derived from Database name: {db-name}-credentials
    databases := &databasev1.DatabaseList{}
    r.List(ctx, databases)
    
    var requests []reconcile.Request
    for _, db := range databases.Items {
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

## Indexes for Efficient Lookups

Indexes allow fast lookups without listing all resources:

```mermaid
graph TB
    INDEX[Index]
    
    INDEX --> FAST[Fast Lookup]
    INDEX --> EFFICIENT[Efficient Queries]
    INDEX --> SCALABLE[Scalable]
    
    FAST --> BY_OWNER[By Owner]
    FAST --> BY_LABEL[By Label]
    FAST --> BY_FIELD[By Field]
    
    style INDEX fill:#90EE90
    style FAST fill:#FFB6C1
```

## Setting Up Indexes

Indexes allow efficient lookups by field values without scanning all objects.

### Step 1: Define Index Function

```go
// Index function: extract the image field from Database objects
func indexDatabaseImage(obj client.Object) []string {
    db, ok := obj.(*databasev1.Database)
    if !ok {
        return nil
    }
    
    if db.Spec.Image != "" {
        return []string{db.Spec.Image}
    }
    return nil
}
```

### Step 2: Register Index

```go
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    // Create index for image field - find all Databases by PostgreSQL version
    if err := mgr.GetFieldIndexer().IndexField(
        context.Background(),
        &databasev1.Database{},
        "spec.image",
        indexDatabaseImage,
    ); err != nil {
        return err
    }
    
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Complete(r)
}
```

### Step 3: Use Index in Queries

```go
// Find all Databases using a specific PostgreSQL image
databases := &databasev1.DatabaseList{}
err := r.List(ctx, databases, client.MatchingFields{
    "spec.image": "postgres:14",
})
// This query is O(1) with index vs O(n) without
```

## Cross-Namespace Watching

Watch resources across namespaces:

```mermaid
graph TB
    CONTROLLER[Controller]
    
    CONTROLLER --> NS1[Namespace 1]
    CONTROLLER --> NS2[Namespace 2]
    CONTROLLER --> NS3[Namespace 3]
    
    NS1 --> RESOURCE1[Resource]
    NS2 --> RESOURCE2[Resource]
    NS3 --> RESOURCE3[Resource]
    
    RESOURCE1 --> RECONCILE[Reconcile]
    RESOURCE2 --> RECONCILE
    RESOURCE3 --> RECONCILE
    
    style CONTROLLER fill:#FFB6C1
    style RECONCILE fill:#90EE90
```

### Cluster-Scoped Watching

```go
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Watches(
            &corev1.Namespace{},
            handler.EnqueueRequestsFromMapFunc(r.findDatabasesForNamespace),
        ).
        Complete(r)
}

func (r *DatabaseReconciler) findDatabasesForNamespace(ctx context.Context, namespace client.Object) []reconcile.Request {
    // Reconcile all Databases in this namespace
    databases := &databasev1.DatabaseList{}
    r.List(ctx, databases, client.InNamespace(namespace.GetName()))
    
    var requests []reconcile.Request
    for _, db := range databases.Items {
        requests = append(requests, reconcile.Request{
            NamespacedName: types.NamespacedName{
                Name:      db.Name,
                Namespace: db.Namespace,
            },
        })
    }
    return requests
}
```

## Event Handling

Handle different event types with predicates to filter which events trigger reconciliation:

> **Important:** When filtering StatefulSet updates, include both spec changes (Generation) AND status changes (ReadyReplicas). Otherwise your controller won't react to pods becoming ready!

```go
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Owns(&appsv1.StatefulSet{}, builder.WithPredicates(predicate.Funcs{
            UpdateFunc: func(e event.UpdateEvent) bool {
                oldSS := e.ObjectOld.(*appsv1.StatefulSet)
                newSS := e.ObjectNew.(*appsv1.StatefulSet)
                // Reconcile on spec changes OR status changes
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
        Complete(r)
}
```

## Performance Considerations

### When to Use Indexes

```mermaid
flowchart TD
    QUERY[Need to Query] --> MANY{Many Resources?}
    MANY -->|Yes| INDEX[Use Index]
    MANY -->|No| LIST[Use List]
    
    QUERY --> FREQUENT{Frequent Query?}
    FREQUENT -->|Yes| INDEX
    FREQUENT -->|No| LIST
    
    style INDEX fill:#90EE90
```

**Use indexes when:**
- Querying many resources frequently
- Need fast lookups
- Resources scale to hundreds/thousands

**Use List when:**
- Few resources
- Infrequent queries
- Simple filtering

## Key Takeaways

- **Watch owned resources** with `Owns()`
- **Watch non-owned resources** with `Watches()`
- **Indexes** enable fast lookups
- **Cross-namespace watching** for cluster-scoped controllers
- **Event predicates** filter which events trigger reconciliation
- **Performance** improves with proper watching and indexing

## Understanding for Building Operators

When setting up watches:
- Watch resources that affect your Custom Resource
- Use indexes for frequent queries
- Filter events with predicates
- Watch across namespaces if needed
- Balance performance with complexity

## Related Lab

- [Lab 4.3: Setting Up Watches and Indexes](../labs/lab-03-watching-indexing.md) - Hands-on exercises for this lesson

## References

### Official Documentation
- [Informers](https://github.com/kubernetes/client-go/blob/master/tools/cache/shared_informer.go)
- [Field Selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/)
- [Indexers](https://pkg.go.dev/k8s.io/client-go/tools/cache#Indexer)

### Further Reading
- **Programming Kubernetes** by Michael Hausenblas and Stefan Schimanski - Chapter 4: Working with Client Libraries
- **Kubernetes Operators** by Jason Dobies and Joshua Wood - Chapter 7: Advanced Patterns
- [client-go Informers](https://github.com/kubernetes/client-go/tree/master/examples/workqueue)

### Related Topics
- [Informer Pattern](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md)
- [Workqueue Pattern](https://github.com/kubernetes/client-go/blob/master/util/workqueue/)
- [Controller Performance](https://kubernetes.io/docs/concepts/architecture/controller/#controller-performance)

## Next Steps

Now that you understand watching and indexing, let's learn advanced patterns like multi-phase reconciliation and state machines.

**Navigation:** [← Previous: Finalizers and Cleanup](02-finalizers-cleanup.md) | [Module Overview](../README.md) | [Next: Advanced Patterns →](04-advanced-patterns.md)
