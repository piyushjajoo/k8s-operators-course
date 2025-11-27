# Lesson 7.3: High Availability

**Navigation:** [← Previous: RBAC and Security](02-rbac-security.md) | [Module Overview](../README.md) | [Next: Performance and Scalability →](04-performance-scalability.md)

## Introduction

Production operators need to be highly available - they should continue operating even if individual pods fail. This lesson covers leader election, multiple replicas, failover handling, and resource management for high availability.

## High Availability Architecture

Here's how HA works for operators:

```mermaid
graph TB
    OPERATOR[Operator Deployment]
    
    OPERATOR --> REPLICA1[Replica 1]
    OPERATOR --> REPLICA2[Replica 2]
    OPERATOR --> REPLICA3[Replica 3]
    
    REPLICA1 --> LEADER[Leader Election]
    REPLICA2 --> LEADER
    REPLICA3 --> LEADER
    
    LEADER --> ACTIVE[Active Controller]
    LEADER --> STANDBY[Standby Controllers]
    
    style LEADER fill:#90EE90
    style ACTIVE fill:#FFB6C1
```

## Leader Election

### How Leader Election Works

```mermaid
sequenceDiagram
    participant R1 as Replica 1
    participant R2 as Replica 2
    participant R3 as Replica 3
    participant API as API Server
    
    R1->>API: Acquire Lease
    API-->>R1: Lease Acquired
    R1->>R1: Become Leader
    
    R2->>API: Try Acquire Lease
    API-->>R2: Lease Already Held
    R2->>R2: Standby Mode
    
    R3->>API: Try Acquire Lease
    API-->>R3: Lease Already Held
    R3->>R3: Standby Mode
    
    Note over R1: Leader runs controller
    
    R1->>API: Renew Lease
    API-->>R1: Renewed
    
    Note over R1: If leader fails...
    R2->>API: Acquire Lease
    API-->>R2: Lease Acquired
    R2->>R2: Become Leader
```

### Leader Election Configuration

```go
mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme:                  scheme,
    LeaderElection:          true,
    LeaderElectionID:        "database-operator-leader-election",
    LeaderElectionNamespace: "default",
    LeaseDuration:           &metav1.Duration{Duration: 15 * time.Second},
    RenewDeadline:           &metav1.Duration{Duration: 10 * time.Second},
    RetryPeriod:             &metav1.Duration{Duration: 2 * time.Second},
})
```

## Multiple Replicas

### Deployment Configuration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: database-operator
spec:
  replicas: 3
  selector:
    matchLabels:
      app: database-operator
  template:
    metadata:
      labels:
        app: database-operator
    spec:
      containers:
      - name: manager
        image: database-operator:latest
```

### Replica Coordination

```mermaid
graph TB
    REPLICAS[3 Replicas]
    
    REPLICAS --> LEADER[1 Leader]
    REPLICAS --> STANDBY[2 Standby]
    
    LEADER --> RECONCILE[Reconciles Resources]
    STANDBY --> WAIT[Wait for Leadership]
    
    LEADER --> FAIL[Leader Fails]
    FAIL --> ELECT[New Leader Elected]
    ELECT --> RECONCILE
    
    style LEADER fill:#90EE90
    style STANDBY fill:#FFE4B5
```

## Failover Process

### Failover Flow

```mermaid
flowchart TD
    NORMAL[Leader Running] --> FAILURE[Leader Fails]
    FAILURE --> DETECT[Lease Expires]
    DETECT --> ELECT[New Leader Elected]
    ELECT --> RESUME[Resume Reconciliation]
    RESUME --> NORMAL
    
    style FAILURE fill:#FFB6C1
    style ELECT fill:#90EE90
```

### Handling Failover

```go
// Leader election handles failover automatically
// When leader fails:
// 1. Lease expires (after LeaseDuration)
// 2. Another replica acquires lease
// 3. New leader starts reconciling
// 4. No reconciliation is lost (idempotent operations)
```

## Resource Management

### Resource Limits

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### Resource Sizing

```mermaid
graph LR
    RESOURCES[Resources]
    
    RESOURCES --> SMALL[Small Operator<br/>100m CPU, 128Mi]
    RESOURCES --> MEDIUM[Medium Operator<br/>500m CPU, 512Mi]
    RESOURCES --> LARGE[Large Operator<br/>1000m CPU, 1Gi]
    
    style SMALL fill:#90EE90
    style MEDIUM fill:#FFE4B5
    style LARGE fill:#FFB6C1
```

## Pod Disruption Budget

### PDB Configuration

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: database-operator-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: database-operator
```

### PDB Protection

```mermaid
graph TB
    PDB[Pod Disruption Budget]
    
    PDB --> MIN[Min Available: 2]
    PDB --> PROTECT[Protects Replicas]
    
    PROTECT --> DRAIN[Prevents Drain]
    PROTECT --> DELETE[Prevents Delete]
    
    style PDB fill:#90EE90
```

## Health Checks

### Liveness and Readiness

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8081
  initialDelaySeconds: 15
  periodSeconds: 20

readinessProbe:
  httpGet:
    path: /readyz
    port: 8081
  initialDelaySeconds: 5
  periodSeconds: 10
```

## Key Takeaways

- **Leader election** ensures only one active controller
- **Multiple replicas** provide redundancy
- **Failover** is automatic with leader election
- **Resource limits** prevent resource exhaustion
- **Pod Disruption Budgets** protect availability
- **Health checks** ensure operator health
- **Idempotent operations** handle failover gracefully

## Understanding for Building Operators

When implementing high availability:
- Enable leader election
- Deploy multiple replicas
- Set appropriate resource limits
- Configure Pod Disruption Budgets
- Add health checks
- Ensure operations are idempotent
- Test failover scenarios

## Related Lab

- [Lab 7.3: Implementing HA](../labs/lab-03-high-availability.md) - Hands-on exercises for this lesson

## Next Steps

Now that you understand high availability, let's learn about performance optimization.

**Navigation:** [← Previous: RBAC and Security](02-rbac-security.md) | [Module Overview](../README.md) | [Next: Performance and Scalability →](04-performance-scalability.md)

