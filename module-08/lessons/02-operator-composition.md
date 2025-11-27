# Lesson 8.2: Operator Composition

**Navigation:** [← Previous: Multi-Tenancy](01-multi-tenancy.md) | [Module Overview](../README.md) | [Next: Stateful Applications →](03-stateful-applications.md)

## Introduction

Real-world applications often require multiple operators working together. This lesson covers operator composition patterns, dependency management, coordination strategies, and how to build operators that work well with others.

## Operator Composition Patterns

### Pattern 1: Independent Operators

```mermaid
graph TB
    APP[Application]
    
    APP --> OP1[Operator 1]
    APP --> OP2[Operator 2]
    APP --> OP3[Operator 3]
    
    OP1 --> RESOURCE1[Resource 1]
    OP2 --> RESOURCE2[Resource 2]
    OP3 --> RESOURCE3[Resource 3]
    
    style APP fill:#90EE90
```

**Characteristics:**
- Operators work independently
- No direct dependencies
- Each manages its own resources

### Pattern 2: Dependent Operators

```mermaid
graph TB
    OP1[Operator 1] --> OP2[Operator 2]
    OP2 --> OP3[Operator 3]
    
    OP1 --> RESOURCE1[Resource 1]
    OP2 --> RESOURCE2[Resource 2]
    OP3 --> RESOURCE3[Resource 3]
    
    style OP1 fill:#90EE90
    style OP2 fill:#FFE4B5
    style OP3 fill:#FFB6C1
```

**Characteristics:**
- Operators depend on each other
- Order matters
- Coordination needed

## Dependency Management

### Dependency Flow

```mermaid
sequenceDiagram
    participant User
    participant OP1 as Operator 1
    participant OP2 as Operator 2
    participant K8s as Kubernetes
    
    User->>K8s: Create Resource 1
    K8s->>OP1: Reconcile Resource 1
    OP1->>K8s: Create Resource 2
    K8s->>OP2: Reconcile Resource 2
    OP2->>K8s: Create Final Resources
    K8s-->>User: Complete
    
    Note over OP1,OP2: Operators coordinate<br/>through resources
```

### Managing Dependencies

```go
// Operator 1 creates resource that Operator 2 watches
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Create Database
    db := &databasev1.Database{...}
    r.Create(ctx, db)
    
    // Create Backup resource (watched by Backup Operator)
    backup := &backupv1.Backup{
        ObjectMeta: metav1.ObjectMeta{
            Name:      db.Name + "-backup",
            Namespace: db.Namespace,
        },
        Spec: backupv1.BackupSpec{
            DatabaseRef: db.Name,
        },
    }
    r.Create(ctx, backup)
    
    // Backup Operator will reconcile backup
}
```

## Coordination Strategies

### Strategy 1: Resource References

```go
// Database references Backup
type DatabaseSpec struct {
    BackupRef *corev1.LocalObjectReference `json:"backupRef,omitempty"`
}

// Operator checks if backup exists
func (r *DatabaseReconciler) checkBackup(ctx context.Context, db *databasev1.Database) error {
    if db.Spec.BackupRef == nil {
        return nil
    }
    
    backup := &backupv1.Backup{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Spec.BackupRef.Name,
        Namespace: db.Namespace,
    }, backup)
    
    if errors.IsNotFound(err) {
        return fmt.Errorf("backup %s not found", db.Spec.BackupRef.Name)
    }
    
    // Wait for backup to be ready
    if backup.Status.Phase != "Ready" {
        return fmt.Errorf("backup not ready")
    }
    
    return nil
}
```

### Strategy 2: Status Conditions

```go
// Operator 1 sets condition
meta.SetStatusCondition(&db.Status.Conditions, metav1.Condition{
    Type:    "BackupReady",
    Status:  metav1.ConditionTrue,
    Reason:  "BackupCompleted",
    Message: "Backup is ready",
})

// Operator 2 checks condition
backupReady := meta.FindStatusCondition(db.Status.Conditions, "BackupReady")
if backupReady == nil || backupReady.Status != metav1.ConditionTrue {
    // Wait for backup
    return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}
```

### Strategy 3: Events

```go
// Operator 1 emits event
r.Recorder.Event(db, "Normal", "BackupCreated", "Backup created successfully")

// Operator 2 watches for events
// Can react to events from other operators
```

## Composite Operator Pattern

### Composite Operator Flow

```mermaid
graph TB
    COMPOSITE[Composite Operator]
    
    COMPOSITE --> COMPONENT1[Component 1]
    COMPOSITE --> COMPONENT2[Component 2]
    COMPONENT1 --> COMPONENT3[Component 3]
    
    COMPONENT1 --> RESOURCE1[Resource 1]
    COMPONENT2 --> RESOURCE2[Resource 2]
    COMPONENT3 --> RESOURCE3[Resource 3]
    
    style COMPOSITE fill:#90EE90
```

### Example: Database with Backup

```go
type DatabaseReconciler struct {
    client.Client
    Scheme *runtime.Scheme
    BackupReconciler *BackupReconciler
}

func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Reconcile Database
    if err := r.reconcileDatabase(ctx, req); err != nil {
        return ctrl.Result{}, err
    }
    
    // Reconcile Backup (component)
    if err := r.BackupReconciler.Reconcile(ctx, req); err != nil {
        return ctrl.Result{}, err
    }
    
    return ctrl.Result{}, nil
}
```

## Key Takeaways

- **Operator composition** enables complex applications
- **Independent operators** work separately
- **Dependent operators** require coordination
- **Resource references** link operators
- **Status conditions** coordinate state
- **Events** enable communication
- **Composite operators** combine multiple components

## Understanding for Building Operators

When composing operators:
- Design for independence when possible
- Use resource references for dependencies
- Coordinate through status conditions
- Emit events for coordination
- Handle dependency failures gracefully
- Document dependencies clearly

## Related Lab

- [Lab 8.2: Composing Operators](../labs/lab-02-operator-composition.md) - Hands-on exercises for this lesson

## Next Steps

Now that you understand operator composition, let's learn about managing stateful applications.

**Navigation:** [← Previous: Multi-Tenancy](01-multi-tenancy.md) | [Module Overview](../README.md) | [Next: Stateful Applications →](03-stateful-applications.md)

