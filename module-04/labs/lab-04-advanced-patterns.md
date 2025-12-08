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

### Task 1.0: Update the API Types (Important!)

Before implementing the state machine, you need to update the `Phase` field validation in your API types to allow the new states.

Edit `api/v1/database_types.go` and update the Phase field enum:

```go
// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
    // Phase is the current phase
    // +kubebuilder:validation:Enum=Pending;Provisioning;Configuring;Deploying;Verifying;Ready;Failed
    Phase string `json:"phase,omitempty"`
    
    // ... rest of status fields
}
```

Then regenerate and reinstall the CRD:

```bash
make manifests
make install
```

> **Important:** If you skip this step, you'll see validation errors like:
> `Database.database.example.com "test-db" is invalid: phase: Unsupported value: "Provisioning": supported values: "Pending", "Creating", "Ready", "Failed"`

### Task 1.1: Define States

Add state constants in `internal/controller/database_controller.go`:

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
    
    logger := log.FromContext(ctx)
    logger.Info("Reconciling", "state", currentState)
    
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

### Task 1.3: Update the Main Reconcile Function

**Important:** You must update your main `Reconcile` function to call the state machine instead of the direct reconciliation flow:

```go
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // Read Database resource
    db := &databasev1.Database{}
    if err := r.Get(ctx, req.NamespacedName, db); err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil
        }
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
        return r.handleDeletion(ctx, db)
    }

    logger.Info("Reconciling Database", "name", db.Name)

    // Use state machine for multi-phase reconciliation
    return r.reconcileWithStateMachine(ctx, db)
}
```

> **Note:** If you skip this step and keep the old Reconcile function that directly calls `reconcileStatefulSet`, `reconcileService`, and `updateStatus`, the state machine functions will never be called and you'll only see `Pending → Creating → Ready` transitions.

## Exercise 2: Implement State Handlers

### Task 2.1: State Transition Functions

```go
func (r *DatabaseReconciler) transitionToProvisioning(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    logger.Info("STATE TRANSITION: Pending -> Provisioning", "database", db.Name)

    db.Status.Phase = string(StateProvisioning)
    db.Status.Ready = false
    r.setCondition(db, "Progressing", metav1.ConditionTrue, "Provisioning", "Starting provisioning")
    if err := r.Status().Update(ctx, db); err != nil {
        return ctrl.Result{}, err
    }

    logger.Info("Waiting 15 seconds before next reconciliation (for visualization)", "currentPhase", db.Status.Phase)
    // Delay to visualize state transition (remove in production)
    return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}

func (r *DatabaseReconciler) handleProvisioning(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    logger.Info("Handling Provisioning phase", "database", db.Name)

    // Ensure Secret exists first (StatefulSet needs it for credentials)
    if err := r.reconcileSecret(ctx, db); err != nil {
        return ctrl.Result{}, err
    }

    // Check if StatefulSet exists
    statefulSet := &appsv1.StatefulSet{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet)
    
    if errors.IsNotFound(err) {
        // Create StatefulSet
        logger.Info("Creating StatefulSet", "database", db.Name)
        if err := r.reconcileStatefulSet(ctx, db); err != nil {
            return ctrl.Result{}, err
        }
        return ctrl.Result{Requeue: true}, nil
    }
    
    // StatefulSet exists, move to next phase
    logger.Info("STATE TRANSITION: Provisioning -> Configuring", "database", db.Name)
    db.Status.Phase = string(StateConfiguring)
    r.setCondition(db, "Progressing", metav1.ConditionTrue, "Configuring", "StatefulSet created, configuring")
    if err := r.Status().Update(ctx, db); err != nil {
        return ctrl.Result{}, err
    }

    logger.Info("Waiting 15 seconds before next reconciliation (for visualization)", "currentPhase", db.Status.Phase)
    // Delay to visualize state transition (remove in production)
    return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}

func (r *DatabaseReconciler) handleConfiguring(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    logger.Info("Handling Configuring phase", "database", db.Name)

    // Ensure Service exists
    logger.Info("Creating Service", "database", db.Name)
    if err := r.reconcileService(ctx, db); err != nil {
        return ctrl.Result{}, err
    }

    // Configure database (create users, databases, etc.)
    // For now, just move to next phase
    logger.Info("STATE TRANSITION: Configuring -> Deploying", "database", db.Name)
    db.Status.Phase = string(StateDeploying)
    r.setCondition(db, "Progressing", metav1.ConditionTrue, "Deploying", "Configuration complete, deploying")
    if err := r.Status().Update(ctx, db); err != nil {
        return ctrl.Result{}, err
    }

    logger.Info("Waiting 15 seconds before next reconciliation (for visualization)", "currentPhase", db.Status.Phase)
    // Delay to visualize state transition (remove in production)
    return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}

func (r *DatabaseReconciler) handleDeploying(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    logger.Info("Handling Deploying phase", "database", db.Name)

    // Check if StatefulSet is ready
    statefulSet := &appsv1.StatefulSet{}
    if err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet); err != nil {
        return ctrl.Result{}, err
    }
    
    if statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas {
        logger.Info("STATE TRANSITION: Deploying -> Verifying", "database", db.Name)
        db.Status.Phase = string(StateVerifying)
        r.setCondition(db, "Progressing", metav1.ConditionTrue, "Verifying", "Deployment complete, verifying")
        if err := r.Status().Update(ctx, db); err != nil {
            return ctrl.Result{}, err
        }

        logger.Info("Waiting 15 seconds before next reconciliation (for visualization)", "currentPhase", db.Status.Phase)
        // Delay to visualize state transition (remove in production)
        return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
    }
    
    // Not ready yet
    logger.Info("Waiting for StatefulSet replicas to be ready",
        "database", db.Name,
        "readyReplicas", statefulSet.Status.ReadyReplicas,
        "desiredReplicas", *statefulSet.Spec.Replicas)
    return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

func (r *DatabaseReconciler) handleVerifying(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    logger.Info("Handling Verifying phase", "database", db.Name)

    // Verify database is working (connect, run test query, etc.)
    // For now, assume it's ready
    logger.Info("STATE TRANSITION: Verifying -> Ready", "database", db.Name)
    db.Status.Phase = string(StateReady)
    db.Status.Ready = true
    db.Status.SecretName = r.secretName(db)
    db.Status.Endpoint = fmt.Sprintf("%s.%s.svc.cluster.local:5432", db.Name, db.Namespace)
    r.setCondition(db, "Ready", metav1.ConditionTrue, "AllChecksPassed", "Database is ready")
    r.setCondition(db, "Progressing", metav1.ConditionFalse, "ReconciliationComplete", "Reconciliation complete")

    logger.Info("Database is now READY!", "database", db.Name, "endpoint", db.Status.Endpoint)
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

> **Note:** The `RequeueAfter: 15 * time.Second` delays and `logger.Info()` calls are added to help visualize state transitions during development. Watch both the operator logs and the Database status to see each transition. In production, you would remove these delays and reduce logging verbosity.

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

# Watch phase transitions
watch -n 1 'kubectl get database test-db -o jsonpath="{.status.phase}"'
```

### Task 3.3: Observe State Transitions

Open **two terminals** to observe the state machine in action:

**Terminal 1 - Watch the operator logs:**
```bash
# The operator logs will show STATE TRANSITION messages like:
# STATE TRANSITION: Pending -> Provisioning
# STATE TRANSITION: Provisioning -> Configuring
# etc.

# If running with `make run`, logs appear in that terminal
# Look for lines containing "STATE TRANSITION" and "Waiting 15 seconds"
```

**Terminal 2 - Watch the Database status:**
```bash
# Watch phase transitions (updates every second)
watch -n 1 'kubectl get database test-db -o jsonpath="{.status.phase}"'

# Or watch the full status including conditions
kubectl get database test-db -o jsonpath='{.status.conditions}' | jq '.'
```

**Expected state progression (each phase visible for ~15 seconds):**
```
Pending -> Provisioning -> Configuring -> Deploying -> Verifying -> Ready
```

**Expected log output:**
```
INFO    STATE TRANSITION: Pending -> Provisioning    {"database": "test-db"}
INFO    Waiting 15 seconds before next reconciliation (for visualization)    {"currentPhase": "Provisioning"}
INFO    Handling Provisioning phase    {"database": "test-db"}
INFO    STATE TRANSITION: Provisioning -> Configuring    {"database": "test-db"}
...
INFO    Database is now READY!    {"database": "test-db", "endpoint": "test-db.default.svc.cluster.local:5432"}
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

### Task 5.1: Review Idempotent Operations

Your existing `reconcileStatefulSet` function already follows the idempotent pattern. Let's review how it works:

```go
func (r *DatabaseReconciler) reconcileStatefulSet(ctx context.Context, db *databasev1.Database) error {
    logger := log.FromContext(ctx)

    // Step 1: Get current state
    statefulSet := &appsv1.StatefulSet{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet)

    // Step 2: Build desired state
    desiredStatefulSet := r.buildStatefulSet(db)

    // Step 3: Create if not exists (idempotent - won't fail if already exists)
    if errors.IsNotFound(err) {
        if err := ctrl.SetControllerReference(db, desiredStatefulSet, r.Scheme); err != nil {
            return err
        }
        logger.Info("Creating StatefulSet", "name", desiredStatefulSet.Name)
        return r.Create(ctx, desiredStatefulSet)
    } else if err != nil {
        return err
    }

    // Step 4: Update only if different (idempotent - won't update if already correct)
    if statefulSet.Spec.Replicas != desiredStatefulSet.Spec.Replicas {
        return r.patchStatefulSetReplicas(ctx, statefulSet, *desiredStatefulSet.Spec.Replicas)
    }

    if statefulSet.Spec.Template.Spec.Containers[0].Image != desiredStatefulSet.Spec.Template.Spec.Containers[0].Image {
        statefulSet.Spec = desiredStatefulSet.Spec
        logger.Info("Updating StatefulSet", "name", statefulSet.Name)
        return r.updateWithRetry(ctx, statefulSet, 3)
    }

    // Step 5: Already in desired state - do nothing (idempotent)
    return nil
}
```

### Key Idempotency Principles

1. **Check before create**: Always check if resource exists before creating
2. **Compare before update**: Only update if actual state differs from desired state
3. **Use patches when possible**: `patchStatefulSetReplicas` is more targeted than full updates
4. **Handle conflicts**: `updateWithRetry` handles concurrent modification conflicts
5. **No side effects on no-op**: If state is already correct, function returns immediately

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

## Solutions

This lab combines concepts from previous labs. Refer to:
- [State Machine Controller](../solutions/state-machine-controller.go) - **Complete state machine implementation**
- [Condition Helpers](../solutions/conditions-helpers.go) - For status management
- [Finalizer Handler](../solutions/finalizer-handler.go) - For cleanup patterns
- [Watch Setup](../solutions/watch-setup.go) - For watching patterns

> **Note:** The `state-machine-controller.go` file contains the complete implementation including the updated `Reconcile` function that calls the state machine. Make sure your main `Reconcile` function calls `reconcileWithStateMachine` instead of directly reconciling resources.

## Congratulations!

You've completed Module 4! You now understand:
- Status management with conditions
- Finalizers for cleanup
- Watching and indexing
- Advanced reconciliation patterns

In Module 5, you'll learn about webhooks for validation and mutation!

**Navigation:** [← Previous Lab: Watching](lab-03-watching-indexing.md) | [Related Lesson](../lessons/04-advanced-patterns.md) | [Module Overview](../README.md)

