# Lab 4.4: Multi-Phase Reconciliation

**Related Lesson:** [Lesson 4.4: Advanced Patterns](../lessons/04-advanced-patterns.md)  
**Navigation:** [← Previous Lab: Watching](lab-03-watching-indexing.md) | [Module Overview](../README.md)

## Objectives

- Implement multi-phase reconciliation
- Create a state machine for Database operator
- Handle external dependencies
- Ensure idempotency

## Prerequisites

- Completion of [Lab 4.3](lab-03-watching-indexing.md)
- Database operator with watches
- Understanding of advanced patterns

## Exercise 1: Implement State Machine

### Task 1.1: Define States

Add state constants:

```go
type DatabaseState string

const (
    StatePending     DatabaseState = "Pending"
    StateProvisioning DatabaseState = "Provisioning"
    StateConfiguring  DatabaseState = "Configuring"
    StateDeploying    DatabaseState = "Deploying"
    StateVerifying    DatabaseState = "Verifying"
    StateReady        DatabaseState = "Ready"
    StateFailed       DatabaseState = "Failed"
)
```

### Task 1.2: Implement State Machine

```go
func (r *DatabaseReconciler) reconcileWithStateMachine(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    currentState := DatabaseState(db.Status.Phase)
    if currentState == "" {
        currentState = StatePending
    }
    
    log := log.FromContext(ctx)
    log.Info("Reconciling", "state", currentState)
    
    switch currentState {
    case StatePending:
        return r.transitionToProvisioning(ctx, db)
    case StateProvisioning:
        return r.handleProvisioning(ctx, db)
    case StateConfiguring:
        return r.handleConfiguring(ctx, db)
    case StateDeploying:
        return r.handleDeploying(ctx, db)
    case StateVerifying:
        return r.handleVerifying(ctx, db)
    case StateReady:
        return r.handleReady(ctx, db)
    case StateFailed:
        return r.handleFailed(ctx, db)
    default:
        return ctrl.Result{}, nil
    }
}
```

## Exercise 2: Implement State Handlers

### Task 2.1: State Transition Functions

```go
func (r *DatabaseReconciler) transitionToProvisioning(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    db.Status.Phase = string(StateProvisioning)
    r.setCondition(db, "Progressing", metav1.ConditionTrue, "Provisioning", "Starting provisioning")
    return ctrl.Result{}, r.Status().Update(ctx, db)
}

func (r *DatabaseReconciler) handleProvisioning(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    // Check if StatefulSet exists
    statefulSet := &appsv1.StatefulSet{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet)
    
    if errors.IsNotFound(err) {
        // Create StatefulSet
        if err := r.reconcileStatefulSet(ctx, db); err != nil {
            return ctrl.Result{}, err
        }
        return ctrl.Result{Requeue: true}, nil
    }
    
    // StatefulSet exists, move to next phase
    db.Status.Phase = string(StateConfiguring)
    r.setCondition(db, "Progressing", metav1.ConditionTrue, "Configuring", "StatefulSet created, configuring")
    return ctrl.Result{}, r.Status().Update(ctx, db)
}

func (r *DatabaseReconciler) handleConfiguring(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    // Configure database (create users, databases, etc.)
    // For now, just move to next phase
    db.Status.Phase = string(StateDeploying)
    r.setCondition(db, "Progressing", metav1.ConditionTrue, "Deploying", "Configuration complete, deploying")
    return ctrl.Result{}, r.Status().Update(ctx, db)
}

func (r *DatabaseReconciler) handleDeploying(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    // Check if StatefulSet is ready
    statefulSet := &appsv1.StatefulSet{}
    if err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet); err != nil {
        return ctrl.Result{}, err
    }
    
    if statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas {
        db.Status.Phase = string(StateVerifying)
        r.setCondition(db, "Progressing", metav1.ConditionTrue, "Verifying", "Deployment complete, verifying")
        return ctrl.Result{}, r.Status().Update(ctx, db)
    }
    
    // Not ready yet
    return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

func (r *DatabaseReconciler) handleVerifying(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    // Verify database is working (connect, run test query, etc.)
    // For now, assume it's ready
    db.Status.Phase = string(StateReady)
    r.setCondition(db, "Ready", metav1.ConditionTrue, "AllChecksPassed", "Database is ready")
    r.setCondition(db, "Progressing", metav1.ConditionFalse, "ReconciliationComplete", "Reconciliation complete")
    return ctrl.Result{}, r.Status().Update(ctx, db)
}

func (r *DatabaseReconciler) handleReady(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    // Monitor and maintain ready state
    // Check if updates are needed
    return ctrl.Result{}, nil
}

func (r *DatabaseReconciler) handleFailed(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    // Handle failed state
    // Could retry or wait for manual intervention
    return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
}
```

## Exercise 3: Test State Machine

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
kubectl apply -f database.yaml

# Watch phase transitions
watch -n 1 'kubectl get database test-db -o jsonpath="{.status.phase}"'
```

### Task 3.3: Observe State Transitions

```bash
# Watch conditions to see state progression
kubectl get database test-db -o jsonpath='{.status.conditions}' | jq '.'

# Should see transitions:
# Pending -> Provisioning -> Configuring -> Deploying -> Verifying -> Ready
```

## Exercise 4: Handle External Dependencies

### Task 4.1: Add External Dependency Check

```go
func (r *DatabaseReconciler) checkExternalDependency(ctx context.Context, db *databasev1.Database) error {
    // Simulate external dependency check
    // In real operator, this would check external API, service, etc.
    
    // For demo, always return success
    return nil
}

func (r *DatabaseReconciler) handleProvisioning(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    // Check external dependency before proceeding
    if err := r.checkExternalDependency(ctx, db); err != nil {
        r.setCondition(db, "Ready", metav1.ConditionFalse, "ExternalDependencyUnavailable", err.Error())
        r.Status().Update(ctx, db)
        // Retry after delay
        return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
    }
    
    // Proceed with provisioning
    // ...
}
```

## Exercise 5: Ensure Idempotency

### Task 5.1: Make All Operations Idempotent

```go
func (r *DatabaseReconciler) ensureStatefulSet(ctx context.Context, db *databasev1.Database) error {
    statefulSet := &appsv1.StatefulSet{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet)
    
    desiredStatefulSet := r.buildStatefulSet(db)
    
    if errors.IsNotFound(err) {
        // Create
        if err := ctrl.SetControllerReference(db, desiredStatefulSet, r.Scheme); err != nil {
            return err
        }
        return r.Create(ctx, desiredStatefulSet)
    } else if err != nil {
        return err
    }
    
    // Update if needed (idempotent)
    if !reflect.DeepEqual(statefulSet.Spec, desiredStatefulSet.Spec) {
        statefulSet.Spec = desiredStatefulSet.Spec
        return r.Update(ctx, statefulSet)
    }
    
    // Already in desired state
    return nil
}
```

## Cleanup

```bash
# Delete test resources
kubectl delete databases --all
```

## Lab Summary

In this lab, you:
- Implemented multi-phase reconciliation
- Created a state machine
- Handled external dependencies
- Ensured idempotency
- Tested state transitions

## Key Learnings

1. Multi-phase reconciliation handles complex deployments
2. State machines provide structured transitions
3. External dependencies need availability checks
4. All operations must be idempotent
5. State transitions should be clear and observable
6. Error handling is crucial in state machines

## Congratulations!

You've completed Module 4! You now understand:
- Status management with conditions
- Finalizers for cleanup
- Watching and indexing
- Advanced reconciliation patterns

In Module 5, you'll learn about webhooks for validation and mutation!

**Navigation:** [← Previous Lab: Watching](lab-03-watching-indexing.md) | [Related Lesson](../lessons/04-advanced-patterns.md) | [Module Overview](../README.md)

