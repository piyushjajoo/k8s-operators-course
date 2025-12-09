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
make deploy IMG=postgres-operator:v0.1.0

# For Podman: Deploy operator - use localhost/ prefix to match the loaded image
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

For Prometheus to scrape metrics and webhooks to work, label the appropriate namespaces:

```bash
# Label namespace where Prometheus runs (to allow metrics scraping)
kubectl label namespace monitoring metrics=enabled

# Label namespaces where you'll create Database CRs (for webhook access)
kubectl label namespace default webhook=enabled

# Verify labels
kubectl get namespaces --show-labels | grep -E "(metrics|webhook)"
```

### Task 5.6: Test Network Policy Enforcement (Optional)

Network Policies require a CNI that supports them (Calico, Cilium, etc.):

```bash
# Check if your cluster has a CNI that supports network policies
kubectl get pods -n kube-system | grep -E "(calico|cilium|weave)"

# Test: Try to access metrics from an unlabeled namespace (should fail with CNI)
kubectl run test-curl --rm -it --image=curlimages/curl --restart=Never -- \
  curl -k https://postgres-operator-controller-manager-metrics-service.postgres-operator-system:8443/metrics

# Now label the namespace and try again (should work)
kubectl label namespace default metrics=enabled
```

**Note:** The default kind cluster does NOT enforce Network Policies. In production clusters with Calico/Cilium, unlabeled namespaces will be blocked.

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

## Key Learnings

1. RBAC is generated from markers via `make manifests`
2. Review `config/rbac/role.yaml` for generated permissions
3. Minimize markers to match actual controller needs
4. Kubebuilder includes security contexts by default
5. Scan images regularly with Trivy or similar tools
6. **Kubebuilder generates Network Policies** in `config/network-policy/`
7. Enable network policies by uncommenting `../network-policy` in kustomization
8. Label namespaces with `metrics: enabled` or `webhook: enabled` to allow access
9. The distroless base image is already used by kubebuilder
10. Network Policies require a CNI that supports them (Calico, Cilium)

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [RBAC Configuration](../solutions/rbac.yaml) - Optimized RBAC with least privilege
- [Security Configuration](../solutions/security.yaml) - Security contexts, network policies

## Next Steps

Now let's implement high availability!

**Navigation:** [← Previous Lab: Packaging](lab-01-packaging-distribution.md) | [Related Lesson](../lessons/02-rbac-security.md) | [Next Lab: HA →](lab-03-high-availability.md)

