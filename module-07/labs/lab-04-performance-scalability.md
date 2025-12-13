---
layout: default
title: "Lab 07.4: Performance Scalability"
nav_order: 14
parent: "Module 7: Production Considerations"
grand_parent: Modules
mermaid: true
---

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

Controller-runtime automatically exposes metrics. Your postgres-operator already has custom metrics configured!

### Task 4.1: Review Existing Metrics Code

Your operator already has metrics in `internal/controller/metrics.go`:

```bash
cd ~/postgres-operator

# Review the metrics file
cat internal/controller/metrics.go
```

You should see these custom metrics already defined:

```go
var (
    // ReconcileTotal counts the total number of reconciliations
    ReconcileTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "database_reconcile_total",
            Help: "Total number of reconciliations per controller",
        },
        []string{"result"}, // success, error, requeue
    )

    // ReconcileDuration measures the duration of reconciliations
    ReconcileDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "database_reconcile_duration_seconds",
            Help:    "Duration of reconciliations in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"result"},
    )

    // DatabasesTotal tracks the current number of Database resources
    DatabasesTotal = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "database_resources_total",
            Help: "Current number of Database resources by phase",
        },
        []string{"phase"},
    )

    // DatabaseInfo provides information about each database
    DatabaseInfo = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "database_info",
            Help: "Information about Database resources",
        },
        []string{"name", "namespace", "image", "phase"},
    )
)

func init() {
    // Register custom metrics with the global registry
    metrics.Registry.MustRegister(
        ReconcileTotal,
        ReconcileDuration,
        DatabasesTotal,
        DatabaseInfo,
    )
}
```

### Task 4.2: Review Metrics Usage in Controller

Check how metrics are used in the Reconcile function:

```bash
# See how metrics are recorded in the reconcile loop
grep -A 10 "Defer metrics" internal/controller/database_controller.go
```

You should see:

```go
// Defer metrics recording
defer func() {
    duration := time.Since(start).Seconds()
    ReconcileDuration.WithLabelValues(reconcileResult).Observe(duration)
    ReconcileTotal.WithLabelValues(reconcileResult).Inc()
}()
```

And database info metrics being set:

```go
DatabaseInfo.WithLabelValues(
    db.Name,
    db.Namespace,
    db.Spec.Image,
    db.Status.Phase,
).Set(1)
```

### Task 4.3: Access Metrics Endpoint

The metrics endpoint requires authentication with a bearer token. We'll use a ServiceAccount token to authenticate.

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

# Restart operator if already deployed
kubectl rollout restart deploy -n postgres-operator-system postgres-operator-controller-manager

# Port forward to metrics endpoint (using HTTPS on port 8443)
kubectl port-forward -n postgres-operator-system \
  svc/postgres-operator-controller-manager-metrics-service 8443:8443 &

# Get a token for authentication (use the controller-manager service account)
TOKEN=$(kubectl create token postgres-operator-controller-manager -n postgres-operator-system)

# create database to generate some metrics
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: valid-db
spec:
  image: postgres:14
  replicas: 1
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# View custom database metrics (using -k for self-signed cert, -H for auth header)
curl -sk -H "Authorization: Bearer $TOKEN" https://localhost:8443/metrics | grep database_

# View reconciliation metrics
curl -sk -H "Authorization: Bearer $TOKEN" https://localhost:8443/metrics | grep database_reconcile

# View controller-runtime built-in metrics
curl -sk -H "Authorization: Bearer $TOKEN" https://localhost:8443/metrics | grep controller_runtime

# View all metrics
curl -sk -H "Authorization: Bearer $TOKEN" https://localhost:8443/metrics | head -100

# Stop port-forward
pkill -f "port-forward.*8443"
```

**Note:** The metrics endpoint uses Kubernetes RBAC for authorization. The ServiceAccount must have the `metrics-reader` ClusterRole (configured in Lab 7.2).

### Task 4.4: View Metrics in Prometheus

If you have Prometheus set up (from Lab 7.2), view metrics there:

```bash
# Port forward to Prometheus
kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090 &

# Open http://localhost:9090 and query:
# - database_reconcile_total
# - database_reconcile_duration_seconds
# - database_resources_total
# - database_info

# Stop port-forward when done
pkill -f "port-forward.*9090"
```

**Example Prometheus queries:**

| Query | Description |
|-------|-------------|
| `database_reconcile_total` | Total reconciliations by result |
| `rate(database_reconcile_total[5m])` | Reconciliations per second |
| `database_reconcile_duration_seconds_bucket` | Reconciliation latency histogram |
| `histogram_quantile(0.99, rate(database_reconcile_duration_seconds_bucket[5m]))` | p99 latency |
| `database_resources_total` | Current databases by phase |
| `database_info` | Info about each database |

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

**Note:** `kubectl top` requires metrics-server to be installed. The course setup script (`scripts/setup-kind-cluster.sh`) installs it automatically. If you get "Metrics API not available", install it manually:

```bash
# Install metrics-server (if not already installed)
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Patch for kind (disable TLS verification for kubelet)
kubectl patch deployment metrics-server -n kube-system --type='json' -p='[
  {"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"},
  {"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-preferred-address-types=InternalIP"}
]'

# Wait for it to be ready
kubectl rollout status deployment/metrics-server -n kube-system
```

Now monitor performance:

```bash
# Watch operator resource usage (requires metrics-server)
watch kubectl top pods -n postgres-operator-system -l control-plane=controller-manager

# Alternative: Check resource requests/limits if metrics-server not available
kubectl get pods -n postgres-operator-system -l control-plane=controller-manager -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[0].resources}{"\n"}{end}'

# In another terminal, watch reconciliation metrics
# Port forward to metrics endpoint (using HTTPS on port 8443)
kubectl port-forward -n postgres-operator-system \
  svc/postgres-operator-controller-manager-metrics-service 8443:8443 &

# Get a token for authentication (use the controller-manager service account)
TOKEN=$(kubectl create token postgres-operator-controller-manager -n postgres-operator-system)

while true; do
  curl -sk -H "Authorization: Bearer $TOKEN" https://localhost:8443/metrics 2>/dev/null | grep database_reconcile_total
  sleep 5
done

# Check queue length
curl -sk -H "Authorization: Bearer $TOKEN" https://localhost:8443/metrics | grep workqueue

# Check controller logs for reconciliation activity
kubectl logs -n postgres-operator-system -l control-plane=controller-manager --tail=20 -f
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
4. Built-in metrics are available at `:8443/metrics`
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
