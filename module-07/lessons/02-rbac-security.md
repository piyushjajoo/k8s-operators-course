# Lesson 7.2: RBAC and Security

**Navigation:** [← Previous: Packaging and Distribution](01-packaging-distribution.md) | [Module Overview](../README.md) | [Next: High Availability →](03-high-availability.md)

## Introduction

Operators need permissions to manage resources, but they should follow the **principle of least privilege** - only requesting the minimum permissions needed. This lesson covers RBAC (Role-Based Access Control) configuration and security best practices for operators.

## Theory: RBAC and Security

Security is critical for production operators - they have **significant permissions** in your cluster.

### Why Security Matters

**Attack Surface:**
- Operators run with elevated permissions
- Compromised operator = compromised cluster
- Security breaches can be catastrophic
- Compliance requirements

**Principle of Least Privilege:**
- Grant minimum permissions needed
- Reduce attack surface
- Limit blast radius
- Follow security best practices

**Defense in Depth:**
- Multiple security layers
- RBAC for authorization
- Network policies for isolation
- Security contexts for containers

### RBAC Components

**Service Account:**
- Identity for operator pod
- Used for authentication
- Tied to RBAC permissions

**Role/ClusterRole:**
- Defines permissions
- Role: Namespace-scoped
- ClusterRole: Cluster-scoped

**RoleBinding/ClusterRoleBinding:**
- Binds role to service account
- Grants permissions
- RoleBinding: Namespace-scoped
- ClusterRoleBinding: Cluster-scoped

### Security Best Practices

**Image Security:**
- Use distroless images
- Scan for vulnerabilities
- Keep images updated
- Minimal base images

**Container Security:**
- Run as non-root
- Read-only root filesystem
- Drop all capabilities
- Use security contexts

**Network Security:**
- Network policies
- Limit network access
- Isolate operator traffic
- Encrypt communication

Understanding security helps you build secure, production-ready operators.

## RBAC Architecture

Here's how RBAC works for operators:

```mermaid
graph TB
    OPERATOR[Operator Pod]
    
    OPERATOR --> SA[Service Account]
    SA --> ROLE[Role/RoleBinding]
    SA --> CLUSTERROLE[ClusterRole/ClusterRoleBinding]
    
    ROLE --> PERMISSIONS[Permissions]
    CLUSTERROLE --> PERMISSIONS
    
    PERMISSIONS --> API[Kubernetes API]
    
    style OPERATOR fill:#FFB6C1
    style PERMISSIONS fill:#90EE90
```

## RBAC Components

### Service Account

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: database-operator
  namespace: default
```

### Role vs ClusterRole

```mermaid
graph TB
    RBAC[RBAC]
    
    RBAC --> ROLE[Role: Namespaced]
    RBAC --> CLUSTERROLE[ClusterRole: Cluster-wide]
    
    ROLE --> NAMESPACE[Single Namespace]
    CLUSTERROLE --> ALL[All Namespaces]
    
    style ROLE fill:#90EE90
    style CLUSTERROLE fill:#FFB6C1
```

**Role**: Permissions within a namespace  
**ClusterRole**: Permissions across all namespaces

## Kubebuilder RBAC Markers

Kubebuilder generates RBAC from markers:

```go
//+kubebuilder:rbac:groups=database.example.com,resources=databases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=database.example.com,resources=databases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=database.example.com,resources=databases/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
```

### RBAC Marker Format

```mermaid
graph LR
    MARKER[RBAC Marker] --> PARTS[Parts]
    
    PARTS --> GROUPS[groups]
    PARTS --> RESOURCES[resources]
    PARTS --> VERBS[verbs]
    PARTS --> NAMESPACE[namespace]
    
    style MARKER fill:#90EE90
```

## Principle of Least Privilege

```mermaid
flowchart TD
    START[Determine Needs] --> MINIMUM[Minimum Permissions]
    MINIMUM --> REVIEW[Review Generated RBAC]
    REVIEW --> REMOVE[Remove Unnecessary]
    REMOVE --> TEST[Test Functionality]
    TEST --> VERIFY[Verify Works]
    VERIFY --> DEPLOY[Deploy]
    
    style MINIMUM fill:#90EE90
```

**Best Practices:**
- Only request permissions you need
- Use specific verbs (not `*`)
- Use specific resources (not `*`)
- Review generated RBAC
- Test with minimal permissions

## Security Best Practices

### Practice 1: Use Distroless Images

```dockerfile
FROM gcr.io/distroless/static:nonroot
# No shell, no package manager, minimal attack surface
```

### Practice 2: Run as Non-Root

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 65532
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
```

### Practice 3: Read-Only Root Filesystem

```yaml
securityContext:
  readOnlyRootFilesystem: true
volumeMounts:
- name: tmp
  mountPath: /tmp
volumes:
- name: tmp
  emptyDir: {}
```

### Practice 4: Network Policies

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

## Security Scanning

### Scanning Flow

```mermaid
sequenceDiagram
    participant Dev
    participant Build as Build Process
    participant Scanner as Security Scanner
    participant Registry as Registry
    
    Dev->>Build: Build Image
    Build->>Scanner: Scan Image
    Scanner->>Scanner: Check Vulnerabilities
    Scanner-->>Dev: Report Issues
    Dev->>Build: Fix Issues
    Build->>Registry: Push Image
    
    Note over Scanner: Tools: Trivy,<br/>Grype, Snyk
```

## Key Takeaways

- **RBAC** controls operator permissions
- **Service Accounts** identify the operator
- **Roles** are namespaced, **ClusterRoles** are cluster-wide
- **Kubebuilder markers** generate RBAC automatically
- **Principle of least privilege** minimizes risk
- **Security best practices** harden operators
- **Security scanning** finds vulnerabilities
- **Review and minimize** generated RBAC

## Understanding for Building Operators

When configuring RBAC and security:
- Use kubebuilder markers for RBAC
- Review and minimize permissions
- Use distroless images
- Run as non-root
- Use read-only root filesystem
- Apply network policies
- Scan images for vulnerabilities
- Follow least privilege principle

## Related Lab

- [Lab 7.2: Configuring RBAC](../labs/lab-02-rbac-security.md) - Hands-on exercises for this lesson

## References

### Official Documentation
- [RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Service Accounts](https://kubernetes.io/docs/concepts/security/service-accounts/)
- [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)

### Further Reading
- **Kubernetes Security** by Andrew Martin and Michael Hausenblas - Security best practices
- **Kubernetes Operators** by Jason Dobies and Joshua Wood - Chapter 13: Security
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)

### Related Topics
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Security Contexts](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)
- [Image Scanning Tools](https://kubernetes.io/docs/concepts/security/image-scanning/)

## Next Steps

Now that you understand RBAC and security, let's learn about high availability.

**Navigation:** [← Previous: Packaging and Distribution](01-packaging-distribution.md) | [Module Overview](../README.md) | [Next: High Availability →](03-high-availability.md)

