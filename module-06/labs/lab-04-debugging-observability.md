# Lab 6.4: Adding Observability

**Related Lesson:** [Lesson 6.4: Debugging and Observability](../lessons/04-debugging-observability.md)  
**Navigation:** [← Previous Lab: Integration Testing](lab-03-integration-testing.md) | [Module Overview](../README.md)

## Objectives

- Understand existing structured logging
- Add custom Prometheus metrics
- Add Kubernetes event emission
- Set up debugging with Delve
- Verify all observability features work

## Prerequisites

- Completion of [Lab 6.3](lab-03-integration-testing.md)
- Database operator deployed to cluster
- Understanding of observability concepts

## Exercise 1: Verify Structured Logging

Kubebuilder already configures structured logging with zap. Let's verify it works.

### Task 1.1: Check Existing Logging Configuration

Your `cmd/main.go` already has logging configured:

```go
opts := zap.Options{
    Development: true,
}
opts.BindFlags(flag.CommandLine)
flag.Parse()

ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
```

### Task 1.2: Verify Logging in Controller

Your controller already uses structured logging. Check `internal/controller/database_controller.go`:

```go
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    
    // ... later in the code:
    logger.Info("Reconciling Database", "name", db.Name)
    logger.Info("STATE TRANSITION: Pending -> Provisioning", "database", db.Name)
}
```

### Task 1.3: Test Logging

```bash
# Deploy the operator (if not already deployed)
cd ~/postgres-operator
make deploy IMG=postgres-operator:latest

# Watch logs in real-time
kubectl logs -n postgres-operator-system -l control-plane=controller-manager -f

# In another terminal, create a Database to trigger reconciliation
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-logging
  namespace: default
spec:
  image: postgres:14
  databaseName: testdb
  username: testuser
  storage:
    size: 1Gi
EOF
```

**Expected output** (in the logs terminal):
```
INFO    Reconciling Database    {"controller": "database", "name": "test-logging"}
INFO    STATE TRANSITION: Pending -> Provisioning    {"database": "test-logging"}
INFO    Creating Secret    {"name": "test-logging-credentials"}
INFO    Creating StatefulSet    {"name": "test-logging"}
```

### Task 1.4: Cleanup

```bash
kubectl delete database test-logging
```

## Exercise 2: Add Prometheus Metrics

### Task 2.1: Add Metrics RBAC Binding

The Kubebuilder scaffolding creates a `metrics-reader` ClusterRole but doesn't bind it to anyone. We need to create the binding so the ServiceAccount can access its own metrics.

Create `config/rbac/metrics_reader_role_binding.yaml`:

```bash
cat > ~/postgres-operator/config/rbac/metrics_reader_role_binding.yaml << 'EOF'
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app.kubernetes.io/name: postgres-operator
    app.kubernetes.io/managed-by: kustomize
  name: metrics-reader-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: metrics-reader
subjects:
- kind: ServiceAccount
  name: controller-manager
  namespace: system
EOF
```

Update `config/rbac/kustomization.yaml` to include the new file:

```bash
cat > ~/postgres-operator/config/rbac/kustomization.yaml << 'EOF'
resources:
# All RBAC will be applied under this service account in
# the deployment namespace. You may comment out this resource
# if your manager will use a service account that exists at
# runtime. Be sure to update RoleBinding and ClusterRoleBinding
# subjects if changing service account names.
- service_account.yaml
- role.yaml
- role_binding.yaml
- leader_election_role.yaml
- leader_election_role_binding.yaml
# The following RBAC configurations are used to protect
# the metrics endpoint with authn/authz. These configurations
# ensure that only authorized users and service accounts
# can access the metrics endpoint. Comment the following
# permissions if you want to disable this protection.
# More info: https://book.kubebuilder.io/reference/metrics.html
- metrics_auth_role.yaml
- metrics_auth_role_binding.yaml
- metrics_reader_role.yaml
- metrics_reader_role_binding.yaml
# For each CRD, "Admin", "Editor" and "Viewer" roles are scaffolded by
# default, aiding admins in cluster management. Those roles are
# not used by the postgres-operator itself. You can comment the following lines
# if you do not want those helpers be installed with your Project.
- database_admin_role.yaml
- database_editor_role.yaml
- database_viewer_role.yaml
EOF
```

### Task 2.2: Create Metrics File

Create `internal/controller/metrics.go`:

```bash
cat > ~/postgres-operator/internal/controller/metrics.go << 'EOF'
/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

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
EOF
```

### Task 2.3: Update Controller to Use Metrics

Add metrics instrumentation to your `Reconcile` function. Update `internal/controller/database_controller.go`:

**Add import:**
```go
import (
    // ... existing imports ...
    "time"
)
```

**Update the Reconcile function** - add at the very beginning:

```go
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    start := time.Now()
    reconcileResult := "success"
    
    // Defer metrics recording
    defer func() {
        duration := time.Since(start).Seconds()
        ReconcileDuration.WithLabelValues(reconcileResult).Observe(duration)
        ReconcileTotal.WithLabelValues(reconcileResult).Inc()
    }()
    
    logger := log.FromContext(ctx)
    
    // ... rest of existing code ...
```

**Update error handling** - when returning errors, set the result:

```go
    // Example: in error returns, set reconcileResult before returning
    if err := r.Get(ctx, req.NamespacedName, db); err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil
        }
        reconcileResult = "error"  // Add this line
        return ctrl.Result{}, err
    }
```

**Add database info metric** - in the reconcile function after getting the database:

```go
    // Record database info metric
    DatabaseInfo.WithLabelValues(
        db.Name,
        db.Namespace, 
        db.Spec.Image,
        db.Status.Phase,
    ).Set(1)
```

### Task 2.4: Rebuild and Deploy

```bash
cd ~/postgres-operator

# Rebuild the operator (for docker)
make docker-build IMG=postgres-operator:latest

# Build with podman (note: image will be localhost/postgres-operator:latest)
make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman

# Load into kind (for docker)
kind load docker-image postgres-operator:latest

# For podman: Load image into kind (save to tarball, then load)
podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
rm /tmp/postgres-operator.tar

# For docker: Deploy operator
make deploy IMG=postgres-operator:latest

# For podman: Deploy operator - use localhost/ prefix to match the loaded image
make deploy IMG=localhost/postgres-operator:latest

# Redeploy
kubectl rollout restart deployment -n postgres-operator-system postgres-operator-controller-manager

# Wait for it to be ready
kubectl wait --for=condition=ready pod -l control-plane=controller-manager -n postgres-operator-system --timeout=60s
```

### Task 2.5: Test Metrics

The metrics endpoint uses HTTPS with authentication by default.

**Note**: The operator's RBAC includes a `metrics-reader-rolebinding` that grants the controller's ServiceAccount permission to read metrics. This was added to `config/rbac/metrics_reader_role_binding.yaml`.

```bash
# Create a test database first
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-metrics
  namespace: default
spec:
  image: postgres:14
  databaseName: metricsdb
  username: metricsuser
  storage:
    size: 1Gi
EOF

# Wait for it to be reconciled
sleep 30

# Get a token for the ServiceAccount
TOKEN=$(kubectl create token -n postgres-operator-system postgres-operator-controller-manager)

# Port forward to the metrics service
kubectl port-forward -n postgres-operator-system svc/postgres-operator-controller-manager-metrics-service 8443:8443 &
sleep 2

# Query metrics with the token
curl -k -H "Authorization: Bearer $TOKEN" https://localhost:8443/metrics 2>/dev/null | grep database_

# Stop port-forward
pkill -f "port-forward.*8443"
```

**Alternative: Disable secure metrics for local development**

If you prefer simpler access during development:

```bash
# Patch the deployment to disable secure metrics
kubectl patch deployment -n postgres-operator-system postgres-operator-controller-manager \
  --type='json' -p='[
    {"op": "replace", "path": "/spec/template/spec/containers/0/args", "value": [
      "--metrics-bind-address=:8080",
      "--leader-elect",
      "--health-probe-bind-address=:8081",
      "--metrics-secure=false"
    ]}
  ]'

# Wait for rollout
kubectl rollout status deployment -n postgres-operator-system postgres-operator-controller-manager

# Port forward and query metrics (no auth needed)
kubectl port-forward -n postgres-operator-system deployment/postgres-operator-controller-manager 8080:8080 &
sleep 2
curl http://localhost:8080/metrics 2>/dev/null | grep database_

# Stop port-forward
pkill -f "port-forward.*8080"
```

**Expected output:**
```
# HELP database_reconcile_total Total number of reconciliations per controller
# TYPE database_reconcile_total counter
database_reconcile_total{result="success"} 5
# HELP database_reconcile_duration_seconds Duration of reconciliations in seconds
# TYPE database_reconcile_duration_seconds histogram
database_reconcile_duration_seconds_bucket{result="success",le="0.005"} 2
...
# HELP database_info Information about Database resources
# TYPE database_info gauge
database_info{image="postgres:14",name="test-metrics",namespace="default",phase="Provisioning"} 1
```

### Task 2.6: Cleanup

```bash
# Stop port-forward
pkill -f "port-forward.*8443"

# Delete test database
kubectl delete database test-metrics
```

## Exercise 3: Add Kubernetes Events

### Task 3.1: Add Event Recorder to Controller

Update `internal/controller/database_controller.go`:

**Add import:**
```go
import (
    // ... existing imports ...
    "k8s.io/client-go/tools/record"
)
```

**Update the struct:**
```go
// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
    client.Client
    Scheme   *runtime.Scheme
    Recorder record.EventRecorder
}
```

### Task 3.2: Update main.go to Provide Event Recorder

Update `cmd/main.go`:

```go
if err := (&controller.DatabaseReconciler{
    Client:   mgr.GetClient(),
    Scheme:   mgr.GetScheme(),
    Recorder: mgr.GetEventRecorderFor("database-controller"),
}).SetupWithManager(mgr); err != nil {
```

### Task 3.3: Emit Events in Controller

Add events at key points in your controller. Update `internal/controller/database_controller.go`:

**In `handleProvisioning` after creating StatefulSet:**
```go
func (r *DatabaseReconciler) handleProvisioning(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    // ... existing code ...
    
    if errors.IsNotFound(err) {
        logger.Info("Creating StatefulSet", "database", db.Name)
        if err := r.reconcileStatefulSet(ctx, db); err != nil {
            r.Recorder.Event(db, "Warning", "CreateFailed", "Failed to create StatefulSet: "+err.Error())
            return ctrl.Result{}, err
        }
        r.Recorder.Event(db, "Normal", "Created", "StatefulSet created successfully")
        return ctrl.Result{Requeue: true}, nil
    }
    
    // ... rest of existing code ...
}
```

**In `handleVerifying` when database becomes ready:**
```go
func (r *DatabaseReconciler) handleVerifying(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    // ... existing code ...
    
    logger.Info("Database is now READY!", "database", db.Name, "endpoint", db.Status.Endpoint)
    r.Recorder.Event(db, "Normal", "Ready", "Database is ready at "+db.Status.Endpoint)
    
    return ctrl.Result{}, r.Status().Update(ctx, db)
}
```

**In `handleDeletion`:**
```go
func (r *DatabaseReconciler) handleDeletion(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    // ... at the beginning ...
    r.Recorder.Event(db, "Normal", "Deleting", "Starting cleanup of database resources")
    
    // ... at the end before removing finalizer ...
    r.Recorder.Event(db, "Normal", "Deleted", "Cleanup completed successfully")
    
    // ... rest of code ...
}
```

### Task 3.4: Update Test Files (Important!)

Since we added `Recorder` to the struct, update `internal/controller/database_controller_test.go`:

```go
// In each test where you create DatabaseReconciler, add the Recorder field:
controllerReconciler := &DatabaseReconciler{
    Client:   k8sClient,
    Scheme:   k8sClient.Scheme(),
    Recorder: record.NewFakeRecorder(100),  // Add this line
}
```

**Add import:**
```go
import (
    // ... existing imports ...
    "k8s.io/client-go/tools/record"
)
```

### Task 3.5: Rebuild and Deploy

```
cd ~/postgres-operator

# Run tests first to make sure they pass
make test

# Rebuild the operator (for docker)
make docker-build IMG=postgres-operator:latest

# Build with podman (note: image will be localhost/postgres-operator:latest)
make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman

# Load into kind (for docker)
kind load docker-image postgres-operator:latest

# For podman: Load image into kind (save to tarball, then load)
podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
rm /tmp/postgres-operator.tar

# For docker: Deploy operator
make deploy IMG=postgres-operator:latest

# For podman: Deploy operator - use localhost/ prefix to match the loaded image
make deploy IMG=localhost/postgres-operator:latest

# Redeploy
kubectl rollout restart deployment -n postgres-operator-system postgres-operator-controller-manager

# Wait for it to be ready
kubectl wait --for=condition=ready pod -l control-plane=controller-manager -n postgres-operator-system --timeout=60s
```

### Task 3.6: Test Events

```bash
# Create a test database
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-events
  namespace: default
spec:
  image: postgres:14
  databaseName: eventsdb
  username: eventsuser
  storage:
    size: 1Gi
EOF

# Wait for reconciliation
sleep 30

# View events for the database
kubectl get events --field-selector involvedObject.name=test-events --sort-by='.lastTimestamp'

# Or view all recent events
kubectl get events -n default --sort-by='.lastTimestamp' | head -20
```

**Expected output:**
```
LAST SEEN   TYPE     REASON    OBJECT                  MESSAGE
30s         Normal   Created   database/test-events    StatefulSet created successfully
15s         Normal   Ready     database/test-events    Database is ready at test-events.default.svc.cluster.local:5432
```

### Task 3.7: Cleanup

```bash
kubectl delete database test-events
```

## Exercise 4: Set Up Delve Debugger

### Task 4.1: Install Delve

```bash
go install github.com/go-delve/delve/cmd/dlv@latest

# Verify installation
dlv version
```

### Task 4.2: Debug Locally (Without Cluster)

```bash
cd ~/postgres-operator

# Start the operator with Delve (won't connect to cluster without kubeconfig)
dlv debug ./cmd/main.go -- --metrics-bind-address=:8080 --health-probe-bind-address=:8081

# In Delve console:
(dlv) break internal/controller/database_controller.go:64
(dlv) continue
# The breakpoint will hit when Reconcile is called
(dlv) print req
(dlv) next
(dlv) step
(dlv) quit
```

### Task 4.3: Debug with Running Cluster

```bash
# Run operator locally (outside cluster) for debugging
cd ~/postgres-operator

# Make sure webhooks are disabled for local run
export ENABLE_WEBHOOKS=false

# Start with Delve
dlv debug ./cmd/main.go -- --metrics-bind-address=:8080 --health-probe-bind-address=:8081

# Set breakpoints and debug
(dlv) break internal/controller/database_controller.go:64
(dlv) continue

# In another terminal, create a Database to trigger the breakpoint
kubectl apply -f config/samples/database_v1_database.yaml
```

### Task 4.4: VS Code Debugging (Alternative)

Create `.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Operator",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/main.go",
            "args": [
                "--metrics-bind-address=:8080",
                "--health-probe-bind-address=:8081"
            ],
            "env": {
                "ENABLE_WEBHOOKS": "false"
            }
        }
    ]
}
```

## Exercise 5: Full Observability Verification

### Task 5.1: Deploy and Create Test Resource

```bash
cd ~/postgres-operator

# Make sure latest version is deployed
make docker-build IMG=postgres-operator:latest
kind load docker-image postgres-operator:latest
make deploy IMG=postgres-operator:latest

# Wait for deployment
kubectl wait --for=condition=ready pod -l control-plane=controller-manager -n postgres-operator-system --timeout=120s

# Create test database
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: observability-test
  namespace: default
spec:
  image: postgres:14
  databaseName: obsdb
  username: obsuser
  storage:
    size: 1Gi
EOF
```

### Task 5.2: Verify All Observability Features

```bash
echo "=== 1. Checking Logs ==="
kubectl logs -n postgres-operator-system -l control-plane=controller-manager --tail=50 | grep -E "(Reconciling|STATE TRANSITION|Creating|Ready)"

echo ""
echo "=== 2. Checking Events ==="
kubectl get events --field-selector involvedObject.name=observability-test --sort-by='.lastTimestamp'

echo ""
echo "=== 3. Checking Database Status ==="
kubectl get database observability-test -o jsonpath='{.status}' | jq .

echo ""
echo "=== 4. Checking Metrics ==="
# Note: This assumes you've disabled secure metrics (Option A from Exercise 2)
# If secure metrics are enabled, you'll need to use a token
kubectl port-forward -n postgres-operator-system deployment/postgres-operator-controller-manager 8080:8080 &
sleep 2
curl -s http://localhost:8080/metrics 2>/dev/null | grep -E "^database_" | head -20 || echo "Metrics not available (secure metrics may be enabled)"
pkill -f "port-forward.*8080" 2>/dev/null
```

### Task 5.3: Cleanup

```bash
kubectl delete database observability-test
```

## Lab Summary

In this lab, you:
- Verified existing structured logging works
- Added custom Prometheus metrics for reconciliation tracking
- Added Kubernetes event emission for user visibility
- Learned to use Delve debugger for troubleshooting
- Verified all observability features work together

## Key Learnings

1. **Structured logging** - Already configured by Kubebuilder; use `log.FromContext(ctx)` with key-value pairs
2. **Custom metrics** - Register with `metrics.Registry.MustRegister()` in an `init()` function
3. **Event Recorder** - Add to reconciler struct, get from manager with `mgr.GetEventRecorderFor()`
4. **Events are user-facing** - Use `Normal` for success, `Warning` for errors
5. **Update tests** - When adding fields to reconciler struct, update test files too
6. **Delve debugging** - Use `ENABLE_WEBHOOKS=false` for local debugging
7. **Secure metrics** - Modern Kubebuilder uses HTTPS with auth on port 8443; use ServiceAccount token to access
8. **Disable secure metrics for dev** - Add `--metrics-secure=false` flag for easier local testing

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Metrics RBAC Binding](../solutions/metrics_reader_role_binding.yaml) - ClusterRoleBinding for metrics access
- [RBAC Kustomization](../solutions/rbac_kustomization.yaml) - Updated kustomization with metrics binding
- [Metrics Implementation](../solutions/metrics.go) - Custom Prometheus metrics
- [Observability Examples](../solutions/observability.go) - Logging and events patterns

## Congratulations!

You've completed Module 6! You now understand:
- Testing fundamentals and strategies
- Unit testing with envtest
- Integration testing with real clusters
- Debugging and observability

In Module 7, you'll learn about production deployment and best practices!

**Navigation:** [← Previous Lab: Integration Testing](lab-03-integration-testing.md) | [Related Lesson](../lessons/04-debugging-observability.md) | [Module Overview](../README.md)
