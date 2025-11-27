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

## Exercise 1: Implement Rate Limiting

### Task 1.1: Add Rate Limiter

Create `internal/ratelimiter/ratelimiter.go`:

```go
package ratelimiter

import (
    "sync"
    "time"
)

type RateLimiter struct {
    mu          sync.Mutex
    lastCall    time.Time
    minInterval time.Duration
}

func NewRateLimiter(minInterval time.Duration) *RateLimiter {
    return &RateLimiter{
        minInterval: minInterval,
    }
}

func (r *RateLimiter) Wait() {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    elapsed := time.Since(r.lastCall)
    if elapsed < r.minInterval {
        time.Sleep(r.minInterval - elapsed)
    }
    r.lastCall = time.Now()
}
```

### Task 1.2: Use in Controller

```go
import "github.com/example/postgres-operator/internal/ratelimiter"

type DatabaseReconciler struct {
    client.Client
    Scheme      *runtime.Scheme
    rateLimiter *ratelimiter.RateLimiter
}

func NewDatabaseReconciler(mgr ctrl.Manager) *DatabaseReconciler {
    return &DatabaseReconciler{
        Client:      mgr.GetClient(),
        Scheme:      mgr.GetScheme(),
        rateLimiter: ratelimiter.NewRateLimiter(100 * time.Millisecond),
    }
}

func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    r.rateLimiter.Wait()
    // ... reconciliation ...
}
```

## Exercise 2: Add Caching

### Task 2.1: Use Informer Cache

```go
import (
    "sigs.k8s.io/controller-runtime/pkg/cache"
)

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    // Use cache for faster lookups
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        WithOptions(controller.Options{
            CacheSyncTimeout: 2 * time.Minute,
        }).
        Complete(r)
}
```

### Task 2.2: Optimize Queries

```go
// Use field selectors instead of listing all
databases := &databasev1.DatabaseList{}
err := r.List(ctx, databases, client.MatchingFields{
    "spec.environment": "production",
})

// Use namespace selector
err := r.List(ctx, databases, client.InNamespace("production"))
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

## Exercise 4: Profile Performance

### Task 4.1: Add Performance Metrics

```go
var (
    reconcileDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "database_reconcile_duration_seconds",
            Help: "Duration of reconciliations",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"result"},
    )
    
    reconcileQueueDepth = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "database_reconcile_queue_depth",
            Help: "Number of items in reconcile queue",
        },
    )
)
```

### Task 4.2: Monitor Performance

```bash
# Port forward to metrics
kubectl port-forward -l control-plane=controller-manager 8080:8080

# Query metrics
curl http://localhost:8080/metrics | grep database_reconcile

# Check reconcile rate
watch 'curl -s http://localhost:8080/metrics | grep database_reconcile_total'
```

## Exercise 5: Load Testing

### Task 5.1: Create Many Resources

```bash
# Create multiple databases
for i in {1..100}; do
  kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-db-$i
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

### Task 5.2: Monitor Performance

```bash
# Watch operator metrics
watch 'kubectl top pods -l control-plane=controller-manager'

# Check reconcile rate
kubectl logs -l control-plane=controller-manager | grep -c "Reconciling"

# Check queue depth
curl -s http://localhost:8080/metrics | grep reconcile_queue_depth
```

## Cleanup

```bash
# Delete test resources
kubectl delete databases --all

# Remove rate limiter if needed
```

## Lab Summary

In this lab, you:
- Implemented rate limiting
- Added caching strategies
- Optimized reconciliation
- Added performance metrics
- Load tested operator

## Key Learnings

1. Rate limiting prevents API overload
2. Caching reduces API calls
3. Batch processing improves efficiency
4. Performance metrics help optimization
5. Load testing validates performance
6. Field selectors optimize queries
7. Monitoring is essential

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

