# Lab 6.4: Adding Observability

**Related Lesson:** [Lesson 6.4: Debugging and Observability](../lessons/04-debugging-observability.md)  
**Navigation:** [← Previous Lab: Integration Testing](lab-03-integration-testing.md) | [Module Overview](../README.md)

## Objectives

- Add structured logging
- Expose Prometheus metrics
- Emit Kubernetes events
- Set up debugging with Delve
- Add observability to operator

## Prerequisites

- Completion of [Lab 6.3](lab-03-integration-testing.md)
- Database operator ready
- Understanding of observability

## Exercise 1: Add Structured Logging

### Task 1.1: Update Logging Configuration

Update `main.go`:

```go
import (
    "sigs.k8s.io/controller-runtime/pkg/log/zap"
    ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
    // Use structured logging
    ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
    
    // ... rest of main
}
```

### Task 1.2: Add Structured Logs to Controller

Update `controllers/database_controller.go`:

```go
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)
    
    log.Info("Reconciling Database",
        "name", req.Name,
        "namespace", req.Namespace,
    )
    
    db := &databasev1.Database{}
    if err := r.Get(ctx, req.NamespacedName, db); err != nil {
        if errors.IsNotFound(err) {
            log.Info("Database not found, ignoring",
                "name", req.Name,
                "namespace", req.Namespace,
            )
            return ctrl.Result{}, nil
        }
        log.Error(err, "Failed to get Database",
            "name", req.Name,
            "namespace", req.Namespace,
        )
        return ctrl.Result{}, err
    }
    
    log.Info("Database found",
        "name", db.Name,
        "generation", db.Generation,
        "replicas", db.Spec.Replicas,
    )
    
    // ... reconciliation logic
    
    log.Info("Reconciliation complete",
        "name", db.Name,
        "phase", db.Status.Phase,
    )
    
    return ctrl.Result{}, nil
}
```

## Exercise 2: Add Prometheus Metrics

### Task 2.1: Define Metrics

Create `controllers/metrics.go`:

```go
package controller

import (
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
    reconcileTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "database_reconcile_total",
            Help: "Total number of reconciliations",
        },
        []string{"result"}, // success, error
    )
    
    reconcileDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "database_reconcile_duration_seconds",
            Help: "Duration of reconciliations",
            Buckets: prometheus.DefBuckets,
        },
        []string{"result"},
    )
)

func init() {
    metrics.Registry.MustRegister(reconcileTotal, reconcileDuration)
}
```

### Task 2.2: Use Metrics in Controller

Update `Reconcile` function:

```go
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    start := time.Now()
    var err error
    result := "success"
    
    defer func() {
        duration := time.Since(start).Seconds()
        if err != nil {
            result = "error"
        }
        reconcileDuration.WithLabelValues(result).Observe(duration)
        reconcileTotal.WithLabelValues(result).Inc()
    }()
    
    // ... reconciliation logic
    
    return ctrl.Result{}, err
}
```

## Exercise 3: Emit Kubernetes Events

### Task 3.1: Add Event Recorder

Update controller struct:

```go
type DatabaseReconciler struct {
    client.Client
    Scheme   *runtime.Scheme
    Recorder record.EventRecorder
}
```

### Task 3.2: Emit Events

```go
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ... get Database
    
    // Emit event on success
    r.Recorder.Event(db, "Normal", "Reconciled", "Database reconciled successfully")
    
    // Emit event on error
    if err != nil {
        r.Recorder.Event(db, "Warning", "ReconcileFailed", err.Error())
    }
    
    // ... rest of reconciliation
}
```

## Exercise 4: Set Up Delve Debugger

### Task 4.1: Debug Locally

```bash
# Start operator with Delve
dlv debug ./cmd/manager/main.go

# In Delve:
# (dlv) break controllers/database_controller.go:50
# (dlv) continue
# (dlv) print db
# (dlv) step
# (dlv) next
```

### Task 4.2: Debug Running Operator

```bash
# Attach to running process
dlv attach <pid>

# Or attach to container
kubectl exec -it <pod> -- dlv attach 1
```

## Exercise 5: Verify Observability

### Task 5.1: Check Logs

```bash
# View operator logs
kubectl logs -l control-plane=controller-manager -f

# Filter logs
kubectl logs -l control-plane=controller-manager | grep "Reconciling"
```

### Task 5.2: Check Metrics

```bash
# Port forward to metrics endpoint
kubectl port-forward -l control-plane=controller-manager 8080:8080

# Query metrics
curl http://localhost:8080/metrics | grep database_reconcile
```

### Task 5.3: Check Events

```bash
# View events
kubectl get events --sort-by='.lastTimestamp'

# Filter by resource
kubectl get events --field-selector involvedObject.name=test-db
```

## Cleanup

```bash
# Clean up test resources
kubectl delete databases --all
```

## Lab Summary

In this lab, you:
- Added structured logging
- Exposed Prometheus metrics
- Emitted Kubernetes events
- Set up Delve debugger
- Verified observability

## Key Learnings

1. Structured logging provides context
2. Metrics expose operational data
3. Events communicate state changes
4. Delve enables debugging
5. Observability is essential for production
6. Multiple observability tools work together

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Observability Examples](../solutions/) - Complete observability examples

## Congratulations!

You've completed Module 6! You now understand:
- Testing fundamentals and strategies
- Unit testing with envtest
- Integration testing with real clusters
- Debugging and observability

In Module 7, you'll learn about production deployment and best practices!

**Navigation:** [← Previous Lab: Integration Testing](lab-03-integration-testing.md) | [Related Lesson](../lessons/04-debugging-observability.md) | [Module Overview](../README.md)

