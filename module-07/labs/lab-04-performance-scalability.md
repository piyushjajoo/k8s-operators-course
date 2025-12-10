# Lab 7.4: Optimizing Performance

**Related Lesson:** [Lesson 7.4: Performance and Scalability](../lessons/04-performance-scalability.md)  
**Navigation:** [← Previous Lab: HA](lab-03-high-availability.md) | [Module Overview](../README.md)

## Objectives

- Implement rate limiting
- Add caching strategies
- Optimize reconciliation
- Profile and optimize performance

## Prerequisites

- Completion of [Lab 7.3](lab-03-high-availability.md)
- Operator with HA setup
- Understanding of performance concepts

## Exercise 1: Configure Controller Rate Limiting

Controller-runtime (used by kubebuilder) has built-in rate limiting. Let's configure it.

### Task 1.1: Configure MaxConcurrentReconciles

Update your controller's `SetupWithManager` in `internal/controller/database_controller.go`:

```go
import (
    "time"
    
    "k8s.io/client-go/util/workqueue"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/controller"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        Owns(&appsv1.StatefulSet{}).
        Owns(&corev1.Service{}).
        Owns(&corev1.Secret{}).
        WithOptions(controller.Options{
            // Limit concurrent reconciliations
            MaxConcurrentReconciles: 2,
            // Custom rate limiter for requeue (typed for controller-runtime v0.19+)
            RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[reconcile.Request](
                time.Millisecond*5,    // Base delay
                time.Second*1000,      // Max delay
            ),
        }).
        Complete(r)
}
```

### Task 1.2: Add Rate Limiting for External API Calls (Optional)

If your operator calls external APIs, add rate limiting:

```go
import (
    "golang.org/x/time/rate"
)

type DatabaseReconciler struct {
    client.Client
    Scheme     *runtime.Scheme
    APILimiter *rate.Limiter  // For external API calls
}

// In cmd/main.go when creating the reconciler:
if err = (&controller.DatabaseReconciler{
    Client:     mgr.GetClient(),
    Scheme:     mgr.GetScheme(),
    APILimiter: rate.NewLimiter(rate.Limit(10), 1), // 10 req/sec
}).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "Database")
    os.Exit(1)
}

// In Reconcile, use before external calls:
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Wait for rate limiter before external API calls
    if err := r.APILimiter.Wait(ctx); err != nil {
        return ctrl.Result{}, err
    }
    // ... reconciliation with external API calls ...
}
```

## Exercise 2: Add Field Indexing for Fast Lookups

Controller-runtime provides automatic caching. You can add custom indexes for fast lookups.

### Task 2.1: Create Field Indexer

Add indexing in `cmd/main.go` before starting the manager:

```go
// In cmd/main.go, after creating manager but before SetupWithManager

// Index databases by environment for fast filtering
if err := mgr.GetFieldIndexer().IndexField(
    context.Background(),
    &databasev1.Database{},
    "spec.environment",
    func(obj client.Object) []string {
        db := obj.(*databasev1.Database)
        if db.Spec.Environment == "" {
            return nil
        }
        return []string{db.Spec.Environment}
    },
); err != nil {
    setupLog.Error(err, "unable to create field index")
    os.Exit(1)
}
```

### Task 2.2: Use Indexes in Controller

```go
// In your reconciler, use MatchingFields for indexed queries
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Fast lookup using index
    prodDatabases := &databasev1.DatabaseList{}
    if err := r.List(ctx, prodDatabases, client.MatchingFields{
        "spec.environment": "production",
    }); err != nil {
        return ctrl.Result{}, err
    }
    
    // Use namespace selector for namespace-scoped queries
    nsDatabases := &databasev1.DatabaseList{}
    if err := r.List(ctx, nsDatabases, client.InNamespace(req.Namespace)); err != nil {
        return ctrl.Result{}, err
    }
    
    // ... rest of reconciliation
}
```

## Exercise 3: Optimize Reconciliation

### Task 3.1: Batch Operations

```go
func (r *DatabaseReconciler) reconcileBatch(ctx context.Context, databases []databasev1.Database) error {
    // Group by operation
    var toCreate, toUpdate []databasev1.Database
    
    for _, db := range databases {
        if db.Status.Phase == "" {
            toCreate = append(toCreate, db)
        } else {
            toUpdate = append(toUpdate, db)
        }
    }
    
    // Batch create
    for _, db := range toCreate {
        if err := r.reconcileDatabase(ctx, &db); err != nil {
            return err
        }
    }
    
    // Batch update
    for _, db := range toUpdate {
        if err := r.reconcileDatabase(ctx, &db); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Exercise 4: Monitor Performance with Built-in Metrics

Controller-runtime automatically exposes metrics. Let's explore and add custom ones.

### Task 4.1: Access Built-in Metrics

```bash
# For Docker: Build and Deploy the operator with network policies enabled
make docker-build IMG=postgres-operator:latest
kind load docker-image postgres-operator:latest --name k8s-operators-course
make deploy IMG=postgres-operator:latest

# For Podman: Build and Deploy operator - use localhost/ prefix to match the loaded image
make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman
podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
rm /tmp/postgres-operator.tar
make deploy IMG=localhost/postgres-operator:latest

# Port forward to metrics endpoint
kubectl port-forward -n postgres-operator-system \
  $(kubectl get pods -n postgres-operator-system -l control-plane=controller-manager -o name | head -1) \
  8080:8080

# View all metrics
curl -s http://localhost:8080/metrics | head -50

# View controller-runtime reconciliation metrics
curl -s http://localhost:8080/metrics | grep controller_runtime_reconcile
```

### Task 4.2: Add Custom Metrics

Add custom metrics in `internal/controller/database_controller.go`:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
    databasesTotal = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "database_operator_databases_total",
            Help: "Total number of Database resources by phase",
        },
        []string{"phase"},
    )
    
    reconcileDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "database_operator_reconcile_duration_seconds",
            Help:    "Duration of reconciliations",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"result"},
    )
)

func init() {
    // Register custom metrics with controller-runtime
    metrics.Registry.MustRegister(databasesTotal, reconcileDuration)
}
```

### Task 4.3: Use Metrics in Reconcile

```go
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    start := time.Now()
    result := "success"
    
    defer func() {
        reconcileDuration.WithLabelValues(result).Observe(time.Since(start).Seconds())
    }()
    
    // ... reconciliation logic ...
    
    if err != nil {
        result = "error"
        return ctrl.Result{}, err
    }
    
    return ctrl.Result{}, nil
}
```

## Exercise 5: Load Testing

### Task 5.1: Create Many Resources

```bash
# Create multiple databases for load testing
for i in {1..50}; do
  kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-db-$i
  namespace: default
spec:
  image: postgres:14
  replicas: 1
  databaseName: db$i
  username: admin
  storage:
    size: 1Gi
EOF
done

echo "Created 50 test databases"
```

### Task 5.2: Monitor Performance Under Load

```bash
# Watch operator resource usage
watch kubectl top pods -n postgres-operator-system -l control-plane=controller-manager

# In another terminal, watch reconciliation metrics
while true; do
  curl -s http://localhost:8080/metrics 2>/dev/null | grep controller_runtime_reconcile_total
  sleep 5
done

# Check controller logs for reconciliation activity
kubectl logs -n postgres-operator-system -l control-plane=controller-manager --tail=20 -f

# Check queue length
curl -s http://localhost:8080/metrics | grep workqueue
```

### Task 5.3: Verify All Resources Are Reconciled

```bash
# Check status of all databases
kubectl get databases -o custom-columns=NAME:.metadata.name,PHASE:.status.phase,READY:.status.ready

# Count databases in each phase
kubectl get databases -o jsonpath='{range .items[*]}{.status.phase}{"\n"}{end}' | sort | uniq -c
```

## Cleanup

```bash
# Delete test resources
kubectl delete databases --all

# Undeploy operator
make undeploy
```

## Lab Summary

In this lab, you:
- Configured controller-runtime rate limiting
- Added field indexing for fast lookups
- Optimized reconciliation with MaxConcurrentReconciles
- Added custom performance metrics
- Load tested the operator with many resources

## Key Learnings

1. Controller-runtime has built-in rate limiting via RateLimiter option
2. MaxConcurrentReconciles controls parallelism
3. Field indexes enable fast filtered queries
4. Built-in metrics are available at `:8080/metrics`
5. Custom metrics use prometheus client with metrics.Registry
6. Load testing validates operator performance at scale
7. `client.MatchingFields{}` leverages indexes for fast lookups

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Rate Limiter](../solutions/ratelimiter.go) - Complete rate limiting implementation
- [Performance Optimizations](../solutions/performance.go) - Batch processing, caching, metrics

## Congratulations!

You've completed Module 7! You now understand:
- Packaging and distribution
- RBAC and security
- High availability
- Performance optimization

In Module 8, you'll learn about advanced topics and real-world patterns!

**Navigation:** [← Previous Lab: HA](lab-03-high-availability.md) | [Related Lesson](../lessons/04-performance-scalability.md) | [Module Overview](../README.md)

