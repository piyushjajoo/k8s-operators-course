---
layout: default
title: "Lab 07.2: Rbac Security"
nav_order: 12
parent: "Module 7: Production Considerations"
grand_parent: Modules
mermaid: true
---

# Lab 7.2: Configuring RBAC

**Related Lesson:** [Lesson 7.2: RBAC and Security](../lessons/02-rbac-security.md)  
**Navigation:** [← Previous Lab: Packaging](lab-01-packaging-distribution.md) | [Module Overview](../README.md) | [Next Lab: HA →](lab-03-high-availability.md)

## Objectives

- Review and optimize RBAC permissions
- Configure service accounts
- Apply security best practices
- Scan images for vulnerabilities
- Configure Network Policies for operator isolation

## Prerequisites

- Completion of [Lab 7.1](lab-01-packaging-distribution.md)
- Operator with generated RBAC
- Understanding of RBAC concepts
- Kind cluster created with `scripts/setup-kind-cluster.sh` (includes Calico CNI and Prometheus)

## Exercise 1: Review Generated RBAC

Kubebuilder generates RBAC manifests automatically from markers in your controller code.

### Task 1.1: Review RBAC Markers in Controller

First, examine your controller's RBAC markers:

```bash
# Navigate to your operator project
cd ~/postgres-operator

# View RBAC markers in your controller
grep -n "// +kubebuilder:rbac" internal/controller/database_controller.go
```

You should see markers like:

```go
// +kubebuilder:rbac:groups=database.example.com,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=databases/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
```

### Task 1.2: Generate and Review RBAC Manifests

```bash
# Generate RBAC manifests from markers
make manifests

# View generated ClusterRole
cat config/rbac/role.yaml

# View generated ClusterRoleBinding
cat config/rbac/role_binding.yaml

# View ServiceAccount
cat config/rbac/service_account.yaml
```

### Task 1.3: Check for Overly Broad Permissions

```bash
# Look for wildcards that might indicate too broad permissions
grep -E "(verbs: \[\"\*\"\]|resources: \[\"\*\"\]|apiGroups: \[\"\*\"\])" config/rbac/role.yaml

# If any are found, review and restrict the corresponding markers
```

## Exercise 2: Optimize RBAC

### Task 2.1: Audit Required Permissions

Review what resources your controller actually accesses:

```bash
# Find all r.Get, r.Create, r.Update, r.Delete, r.List calls
grep -E "r\.(Get|Create|Update|Delete|List|Patch)" internal/controller/database_controller.go
```

### Task 2.2: Update RBAC Markers

Edit your controller to match only the permissions actually needed:

```go
// internal/controller/database_controller.go

// Only include markers for resources you actually use:
// +kubebuilder:rbac:groups=database.example.com,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=databases/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create

// Note: Removed 'delete' from secrets if your controller doesn't delete secrets
// Note: Consider if you need 'update;patch' for all resources
```

### Task 2.3: Regenerate RBAC

```bash
# Regenerate with optimized markers
make manifests

# Review the updated RBAC
cat config/rbac/role.yaml

# Compare rules - they should be minimal
```

## Exercise 3: Review Kubebuilder Security Configuration

Kubebuilder generates security configuration by default. Let's review and enhance it.

### Task 3.1: Review Generated ServiceAccount

```bash
# Kubebuilder creates ServiceAccount automatically
cat config/rbac/service_account.yaml
```

### Task 3.2: Review Deployment Security Context

Kubebuilder's generated deployment includes security contexts. Review them:

```bash
cat config/manager/manager.yaml
```

Look for the security settings:

```yaml
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: true
      containers:
      - name: manager
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
```

### Task 3.3: Enhance Security Context (Optional)

Add additional security hardening to `config/manager/manager.yaml`:

```yaml
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 65532
        fsGroup: 65532
        seccompProfile:
          type: RuntimeDefault
      containers:
      - name: manager
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
```

## Exercise 4: Security Scanning

### Task 4.1: Install Trivy

```bash
# Install Trivy
brew install trivy  # macOS
# or
sudo apt-get install trivy  # Linux
```

### Task 4.2: Scan Image

```bash
# Scan image
trivy image postgres-operator:latest

# Scan with JSON output
trivy image -f json -o scan-report.json postgres-operator:latest

# Fix high/critical vulnerabilities
```

## Exercise 5: Apply Security Hardening

### Task 5.1: Update Dockerfile

```dockerfile
# Use distroless base
FROM gcr.io/distroless/static:nonroot

# Run as non-root
USER 65532:65532
```

## Exercise 5: Enable Network Policies

Kubebuilder already generates Network Policies for your operator! Let's review and enable them.

### Task 5.1: Review Generated Network Policies

Kubebuilder creates network policies in `config/network-policy/`:

```bash
cd ~/postgres-operator

# List the generated network policy files
ls -la config/network-policy/

# Review the metrics traffic policy
cat config/network-policy/allow-metrics-traffic.yaml

# Review the webhook traffic policy  
cat config/network-policy/allow-webhook-traffic.yaml
```

**What Kubebuilder generates:**

1. **`allow-metrics-traffic.yaml`** - Controls access to metrics endpoint:
   - Only allows ingress from namespaces labeled `metrics: enabled`
   - Restricts to port 8443 (HTTPS metrics)

2. **`allow-webhook-traffic.yaml`** - Controls access to webhook server:
   - Only allows ingress from namespaces labeled `webhook: enabled`
   - Restricts to port 443 (webhook HTTPS)

### Task 5.2: Understand the Network Policies

Review the metrics policy:

```yaml
# config/network-policy/allow-metrics-traffic.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-metrics-traffic
  namespace: system
spec:
  podSelector:
    matchLabels:
      control-plane: controller-manager
      app.kubernetes.io/name: postgres-operator
  policyTypes:
    - Ingress
  ingress:
    - from:
      - namespaceSelector:
          matchLabels:
            metrics: enabled  # Only from namespaces with this label
      ports:
        - port: 8443
          protocol: TCP
```

**Key points:**
- Applies to pods with `control-plane: controller-manager` label
- Only allows ingress (incoming traffic)
- Requires source namespace to have `metrics: enabled` label
- This means Prometheus must run in a labeled namespace to scrape metrics

### Task 5.3: Enable Network Policies in Kustomization

Network policies are commented out by default. Enable them:

```bash
# View the current kustomization
cat config/default/kustomization.yaml | grep -A 5 "NETWORK POLICY"
```

You'll see:
```yaml
# [NETWORK POLICY] Protect the /metrics endpoint and Webhook Server with NetworkPolicy.
#- ../network-policy
```

Edit `config/default/kustomization.yaml` and uncomment the network-policy line:

```yaml
# [NETWORK POLICY] Protect the /metrics endpoint and Webhook Server with NetworkPolicy.
- ../network-policy
```

### Task 5.4: Deploy with Network Policies

```bash
# For Docker: Deploy the operator with network policies enabled
kind load docker-image postgres-operator:latest --name k8s-operators-course
make deploy IMG=postgres-operator:latest

# For Podman: Deploy operator - use localhost/ prefix to match the loaded image
podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
rm /tmp/postgres-operator.tar
make deploy IMG=localhost/postgres-operator:latest

# Verify network policies were created
kubectl get networkpolicy -n postgres-operator-system

# View the network policies
kubectl describe networkpolicy -n postgres-operator-system
```

Expected output:
```
NAME                    POD-SELECTOR                                                    AGE
allow-metrics-traffic   app.kubernetes.io/name=postgres-operator,control-plane=...     10s
allow-webhook-traffic   app.kubernetes.io/name=postgres-operator,control-plane=...     10s
```

### Task 5.5: Label Namespaces for Access

For Prometheus to scrape metrics and webhooks to work, label the appropriate namespaces.

**Note:** If you used the course setup script (`scripts/setup-kind-cluster.sh`), the `monitoring` namespace is already labeled with `metrics=enabled`.

```bash
# Check if monitoring namespace already has the label
kubectl get namespace monitoring --show-labels

# If not labeled, add it (the setup script does this automatically)
kubectl label namespace monitoring metrics=enabled --overwrite

# Label namespaces where you'll create Database CRs (for webhook access)
kubectl label namespace default webhook=enabled

# Verify labels
kubectl get namespaces --show-labels | grep -E "(metrics|webhook)"
```

### Task 5.6: Enable ServiceMonitor for Prometheus

Kubebuilder generates a `ServiceMonitor` in `config/prometheus/` that tells Prometheus how to scrape your operator's metrics. It's disabled by default.

#### Step 1: Review the Generated ServiceMonitor

```bash
# View the kubebuilder-generated ServiceMonitor
cat config/prometheus/monitor.yaml
```

You'll see:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: controller-manager-metrics-monitor
spec:
  endpoints:
    - path: /metrics
      port: https
      scheme: https
      bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
      tlsConfig:
        insecureSkipVerify: true
  selector:
    matchLabels:
      control-plane: controller-manager
```

#### Step 2: Enable the ServiceMonitor

Edit `config/default/kustomization.yaml` and uncomment the prometheus line:

```bash
# Find the PROMETHEUS section
grep -n "PROMETHEUS" config/default/kustomization.yaml
```

Uncomment `- ../prometheus`:
```yaml
# [PROMETHEUS] To enable prometheus monitor, uncomment all sections with 'PROMETHEUS'.
- ../prometheus  # <-- Uncomment this line
```

#### Step 3: Grant Prometheus RBAC Access to Metrics

**Important:** The operator's metrics endpoint requires authentication AND authorization. By default, only the controller-manager ServiceAccount has access. We need to grant Prometheus access too.

Create `config/rbac/metrics_reader_prometheus_binding.yaml`:

```yaml
# config/rbac/metrics_reader_prometheus_binding.yaml
# Grant Prometheus ServiceAccount permission to read metrics
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app.kubernetes.io/name: postgres-operator
    app.kubernetes.io/managed-by: kustomize
  name: metrics-reader-prometheus
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: metrics-reader
subjects:
- kind: ServiceAccount
  name: prometheus-kube-prometheus-prometheus
  namespace: monitoring
```

Add it to `config/rbac/kustomization.yaml`:

```yaml
resources:
- service_account.yaml
- role.yaml
- role_binding.yaml
- leader_election_role.yaml
- leader_election_role_binding.yaml
- metrics_auth_role.yaml
- metrics_auth_role_binding.yaml
- metrics_reader_role.yaml
- metrics_reader_role_binding.yaml
- metrics_reader_prometheus_binding.yaml  # <-- Add this line
# ... rest of file
```

#### Step 4: Redeploy with ServiceMonitor and RBAC

```bash
# Regenerate manifests to include the new RBAC binding
make manifests

# For Docker: Redeploy to include the ServiceMonitor
make deploy IMG=postgres-operator:latest

# For Podman: Redeploy to include the ServiceMonitor
make deploy IMG=localhost/postgres-operator:latest

# Verify the ServiceMonitor was created
kubectl get servicemonitor -n postgres-operator-system

# Verify Prometheus RBAC binding was created
kubectl get clusterrolebinding | grep metrics-reader
```

Expected output:
```
NAME                                              AGE
postgres-operator-metrics-reader-prometheus       10s
postgres-operator-metrics-reader-rolebinding      10s
```

**Note:** The course setup script (`scripts/setup-kind-cluster.sh`) configures Prometheus to discover ServiceMonitors from all namespaces without requiring specific labels. If you're using a different Prometheus installation, you may need to add `release: prometheus` label to your ServiceMonitor.

### Task 5.7: Verify Prometheus Can Scrape Metrics

Now let's verify Prometheus is scraping your operator's metrics.

#### Step 1: Start Port-Forward to Prometheus

```bash
# Start port-forward in background
kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090 &

# Note the PID for later cleanup
PF_PID=$!
echo "Port-forward PID: $PF_PID"
```

#### Step 2: Open Prometheus UI

Open your browser and go to: **http://localhost:9090**

#### Step 3: Check if Your Operator Target is Being Scraped

1. In Prometheus UI, click **Status** → **Targets** in the top menu
2. Look for a target with `serviceMonitor/postgres-operator-system/` in the name
3. The **State** should show `UP` (green)

If the target doesn't appear or shows `DOWN`, check:
- Is the ServiceMonitor deployed? (`kubectl get servicemonitor -n postgres-operator-system`)
- Is the monitoring namespace labeled? (`kubectl get ns monitoring --show-labels`)
- Are network policies blocking access?
- Is Prometheus configured to discover all ServiceMonitors?

```bash
# Check if Prometheus discovers all ServiceMonitors (should be empty selector)
kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
# Empty {} means it discovers all ServiceMonitors

# If it shows 'release: prometheus', upgrade Prometheus with the course setup settings
# or add 'release: prometheus' label to your ServiceMonitor
```

#### Step 4: Query Operator Metrics

1. Go back to the main Prometheus page (click **Prometheus** logo or **Graph**)
2. In the **Expression** input box, type one of these queries:
   
   ```promql
   controller_runtime_reconcile_total
   ```
   
3. Click the **Execute** button (or press Enter)
4. Click the **Graph** tab to see a time-series visualization

**Common operator metrics to explore:**

| Metric | Description |
|--------|-------------|
| `controller_runtime_reconcile_total` | Total reconciliations by controller and result |
| `controller_runtime_reconcile_errors_total` | Total reconciliation errors |
| `controller_runtime_reconcile_time_seconds` | Time spent in reconciliation |
| `workqueue_depth` | Current depth of the work queue |
| `workqueue_adds_total` | Total items added to the queue |

#### Step 5: Example Queries to Try

Paste these into the Prometheus Expression box:

```promql
# Reconciliation rate per second (last 5 minutes)
rate(controller_runtime_reconcile_total[5m])

# Error rate
rate(controller_runtime_reconcile_errors_total[5m])

# 99th percentile reconciliation latency
histogram_quantile(0.99, rate(controller_runtime_reconcile_time_seconds_bucket[5m]))
```

#### Step 6: Cleanup Port-Forward

```bash
# Stop the port-forward
pkill -f "port-forward.*9090"
```

### Task 5.8: Test Network Policy Enforcement

The course setup script installs **Calico CNI**, which enforces Network Policies. Let's verify it's working.

#### Step 1: Verify Calico is Running

```bash
# Check Calico pods are running
kubectl get pods -n kube-system | grep calico

# Expected output:
# calico-kube-controllers-xxx   1/1     Running
# calico-node-xxx               1/1     Running
```

#### Step 2: Test Access from Unlabeled Namespace

```bash
# Create a test namespace WITHOUT the metrics=enabled label
kubectl create namespace test-netpol

# Try to access the operator metrics from the unlabeled namespace
# This should FAIL (timeout or connection refused) because of the NetworkPolicy
kubectl run test-curl -n test-netpol --rm -it --image=curlimages/curl --restart=Never -- \
  curl -k --connect-timeout 5 https://postgres-operator-controller-manager-metrics-service.postgres-operator-system:8443/metrics

# Expected: curl: (28) Connection timed out or similar error
```

#### Step 3: Test Access from Labeled Namespace

```bash
# Label the test namespace to allow metrics access
kubectl label namespace test-netpol metrics=enabled

# Try again - this should SUCCEED
kubectl run test-curl2 -n test-netpol --rm -it --image=curlimages/curl --restart=Never -- \
  curl -k --connect-timeout 5 https://postgres-operator-controller-manager-metrics-service.postgres-operator-system:8443/metrics

# Expected: Metrics output (or 401 Unauthorized if auth is required, but connection succeeds)
```

#### Step 4: Cleanup

```bash
kubectl delete namespace test-netpol
```

**Key takeaway:** Network Policies enforce that only pods in namespaces with `metrics=enabled` label can access the metrics endpoint. This is defense in depth!

## Cleanup

```bash
# Undeploy operator (this also removes the network policy)
make undeploy

# If you need to remove network policy separately
kubectl delete networkpolicy controller-manager -n postgres-operator-system

# Uninstall CRDs
make uninstall
```

## Lab Summary

In this lab, you:
- Reviewed RBAC markers in kubebuilder controllers
- Generated and reviewed RBAC manifests with `make manifests`
- Optimized RBAC permissions using least privilege
- Reviewed kubebuilder's security configurations
- Scanned images for vulnerabilities with Trivy
- Enhanced security hardening with security contexts
- Enabled kubebuilder-generated Network Policies
- Labeled namespaces to allow metrics and webhook traffic
- Enabled ServiceMonitor for Prometheus scraping
- Verified metrics collection in Prometheus UI

## Key Learnings

1. RBAC is generated from markers via `make manifests`
2. Review `config/rbac/role.yaml` for generated permissions
3. Minimize markers to match actual controller needs
4. Kubebuilder includes security contexts by default
5. Scan images regularly with Trivy or similar tools
6. **Kubebuilder generates Network Policies** in `config/network-policy/`
7. Enable network policies by uncommenting `../network-policy` in kustomization
8. Label namespaces with `metrics: enabled` or `webhook: enabled` to allow access
9. **Kubebuilder generates ServiceMonitor** in `config/prometheus/` - enable it!
10. **Grant Prometheus RBAC access** to the metrics endpoint (Step 3 above)
11. Use Prometheus UI **Status → Targets** to verify scraping is working
12. The course setup script configures Prometheus to discover all ServiceMonitors
13. The distroless base image is already used by kubebuilder
14. Network Policies require a CNI that supports them (Calico)

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [RBAC Configuration](../solutions/rbac.yaml) - Optimized RBAC with least privilege
- [Security Configuration](../solutions/security.yaml) - Security contexts, network policies

## Next Steps

Now let's implement high availability!

**Navigation:** [← Previous Lab: Packaging](lab-01-packaging-distribution.md) | [Related Lesson](../lessons/02-rbac-security.md) | [Next Lab: HA →](lab-03-high-availability.md)
