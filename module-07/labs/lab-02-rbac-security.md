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

## Exercise 5: Network Policy for Operator Isolation

Network Policies restrict network traffic to/from your operator pods, providing defense in depth.

### Task 5.1: Create Network Policy Directory

```bash
cd ~/postgres-operator

# Create directory for network policy
mkdir -p config/network-policy
```

### Task 5.2: Create the Network Policy

Create the file `config/network-policy/network-policy.yaml` with the following content:

```yaml
# config/network-policy/network-policy.yaml
# Network Policy for postgres-operator controller manager
# Restricts network access to only what the operator needs
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: controller-manager
  namespace: system
  labels:
    app.kubernetes.io/name: postgres-operator
    app.kubernetes.io/component: network-policy
spec:
  # Apply to controller-manager pods
  podSelector:
    matchLabels:
      control-plane: controller-manager
  policyTypes:
  - Ingress
  - Egress
  
  # Ingress rules - what can connect TO the operator
  ingress:
  # Allow metrics scraping from any namespace (for Prometheus)
  - from:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 8080  # Metrics port
  # Allow health checks from within the cluster
  - from:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 8081  # Health probe port
  
  # Egress rules - what the operator can connect TO
  egress:
  # Allow DNS lookups (required for service discovery)
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: UDP
      port: 53
    - protocol: TCP
      port: 53
  # Allow Kubernetes API access (required for operator functionality)
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443   # Kubernetes API (HTTPS)
    - protocol: TCP
      port: 6443  # Kubernetes API (alternative port)
```

### Task 5.3: Create Kustomization for Network Policy

Create `config/network-policy/kustomization.yaml`:

```yaml
# config/network-policy/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- network-policy.yaml

# The namespace will be set by the parent kustomization
```

### Task 5.4: Add Network Policy to Default Kustomization

Edit `config/default/kustomization.yaml` to include the network policy:

```bash
# View current kustomization
cat config/default/kustomization.yaml
```

Add `../network-policy` to the resources list:

```yaml
# In config/default/kustomization.yaml, add to resources:
resources:
- ../crd
- ../rbac
- ../manager
- ../network-policy  # Add this line
# ... other resources
```

### Task 5.5: Verify Network Policy Generation

```bash
# Preview the generated manifests including network policy
kustomize build config/default | grep -A 50 "kind: NetworkPolicy"

# Or deploy and verify
make deploy IMG=postgres-operator:v0.1.0

# Check network policy was created
kubectl get networkpolicy -n postgres-operator-system
kubectl describe networkpolicy controller-manager -n postgres-operator-system
```

### Task 5.6: Test Network Policy (Optional)

To verify the network policy is working, you need a CNI that supports Network Policies (Calico, Cilium, etc.):

```bash
# Check if your cluster supports network policies
kubectl get pods -n kube-system | grep -E "(calico|cilium|weave)"

# If using kind, network policies are NOT enforced by default
# For testing, you can install Calico:
# kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.1/manifests/calico.yaml
```

**Note:** The default kind cluster does not enforce Network Policies. In production clusters with Calico/Cilium, the policy will block unauthorized traffic.

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
- Created Network Policies to isolate operator network traffic

## Key Learnings

1. RBAC is generated from markers via `make manifests`
2. Review `config/rbac/role.yaml` for generated permissions
3. Minimize markers to match actual controller needs
4. Kubebuilder includes security contexts by default
5. Scan images regularly with Trivy or similar tools
6. **Network Policies** provide defense in depth by restricting traffic
7. Operators typically only need egress to Kubernetes API (443) and DNS (53)
8. The distroless base image is already used by kubebuilder
9. Network Policies require a CNI that supports them (Calico, Cilium)

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [RBAC Configuration](../solutions/rbac.yaml) - Optimized RBAC with least privilege
- [Security Configuration](../solutions/security.yaml) - Security contexts, network policies

## Next Steps

Now let's implement high availability!

**Navigation:** [← Previous Lab: Packaging](lab-01-packaging-distribution.md) | [Related Lesson](../lessons/02-rbac-security.md) | [Next Lab: HA →](lab-03-high-availability.md)

