# Lab 7.2: Configuring RBAC

**Related Lesson:** [Lesson 7.2: RBAC and Security](../lessons/02-rbac-security.md)  
**Navigation:** [← Previous Lab: Packaging](lab-01-packaging-distribution.md) | [Module Overview](../README.md) | [Next Lab: HA →](lab-03-high-availability.md)

## Objectives

- Review and optimize RBAC permissions
- Configure service accounts
- Apply security best practices
- Scan images for vulnerabilities

## Prerequisites

- Completion of [Lab 7.1](lab-01-packaging-distribution.md)
- Operator with generated RBAC
- Understanding of RBAC concepts

## Exercise 1: Review Generated RBAC

### Task 1.1: Generate RBAC Manifests

```bash
# Generate RBAC manifests
make manifests

# Check generated RBAC
cat config/rbac/role.yaml
cat config/rbac/role_binding.yaml
```

### Task 1.2: Review Permissions

```bash
# List all permissions
kubectl get role database-operator -o yaml | grep -A 100 "rules:"

# Check for overly broad permissions
# Look for:
# - verbs: ["*"]
# - resources: ["*"]
# - apiGroups: ["*"]
```

## Exercise 2: Optimize RBAC

### Task 2.1: Minimize Permissions

Review your controller code and remove unnecessary RBAC markers:

```go
// Remove if not needed
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// Keep only what you use
// +kubebuilder:rbac:groups=database.example.com,resources=databases,verbs=get;list;watch;create;update;patch;delete
```

### Task 2.2: Regenerate RBAC

```bash
# Regenerate with optimized markers
make manifests

# Review new RBAC
cat config/rbac/role.yaml
```

## Exercise 3: Configure Service Account

### Task 3.1: Create Service Account

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: database-operator
  namespace: default
```

### Task 3.2: Update Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      serviceAccountName: database-operator
      securityContext:
        runAsNonRoot: true
        runAsUser: 65532
        allowPrivilegeEscalation: false
        capabilities:
          drop:
          - ALL
      containers:
      - name: manager
        securityContext:
          readOnlyRootFilesystem: true
          allowPrivilegeEscalation: false
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
trivy image database-operator:latest

# Scan with JSON output
trivy image -f json -o scan-report.json database-operator:latest

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

### Task 5.2: Add Network Policy

Create `config/security/network-policy.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: database-operator
spec:
  podSelector:
    matchLabels:
      app: database-operator
  policyTypes:
  - Ingress
  - Egress
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443  # Kubernetes API
```

## Cleanup

```bash
# Remove test resources
kubectl delete networkpolicy database-operator
```

## Lab Summary

In this lab, you:
- Reviewed generated RBAC
- Optimized permissions
- Configured service accounts
- Scanned images for vulnerabilities
- Applied security hardening

## Key Learnings

1. Review generated RBAC carefully
2. Minimize permissions to least privilege
3. Use service accounts for identity
4. Scan images for vulnerabilities
5. Apply security best practices
6. Use distroless images
7. Configure security contexts

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [RBAC Configuration](../solutions/rbac.yaml) - Optimized RBAC with least privilege
- [Security Configuration](../solutions/security.yaml) - Security contexts, network policies

## Next Steps

Now let's implement high availability!

**Navigation:** [← Previous Lab: Packaging](lab-01-packaging-distribution.md) | [Related Lesson](../lessons/02-rbac-security.md) | [Next Lab: HA →](lab-03-high-availability.md)

