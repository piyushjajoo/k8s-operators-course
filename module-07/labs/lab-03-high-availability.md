# Lab 7.3: Implementing High Availability

**Related Lesson:** [Lesson 7.3: High Availability](../lessons/03-high-availability.md)  
**Navigation:** [← Previous Lab: RBAC](lab-02-rbac-security.md) | [Module Overview](../README.md) | [Next Lab: Performance →](lab-04-performance-scalability.md)

## Objectives

- Enable leader election
- Deploy multiple replicas
- Configure resource limits
- Test failover scenarios
- Set up Pod Disruption Budget

## Prerequisites

- Completion of [Lab 7.2](lab-02-rbac-security.md)
- Operator ready for deployment
- Understanding of leader election

## Exercise 1: Enable Leader Election

### Task 1.1: Update main.go

Edit `main.go`:

```go
mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme:                  scheme,
    Metrics:                metricsserver.Options{BindAddress: metricsAddr},
    WebhookServer:          webhook.NewServer(webhook.Options{Port: 9443}),
    HealthProbeBindAddress: probeAddr,
    LeaderElection:         true,
    LeaderElectionID:       "database-operator-leader-election",
    LeaderElectionNamespace: "default",
    LeaseDuration:          &metav1.Duration{Duration: 15 * time.Second},
    RenewDeadline:          &metav1.Duration{Duration: 10 * time.Second},
    RetryPeriod:            &metav1.Duration{Duration: 2 * time.Second},
})
```

### Task 1.2: Verify Leader Election

```bash
# Deploy operator
make deploy

# Check for lease
kubectl get lease database-operator-leader-election

# Check logs for leader election
kubectl logs -l control-plane=controller-manager | grep -i leader
```

## Exercise 2: Deploy Multiple Replicas

### Task 2.1: Update Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: database-operator
spec:
  replicas: 3  # Multiple replicas
  selector:
    matchLabels:
      control-plane: controller-manager
  template:
    spec:
      containers:
      - name: manager
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
```

### Task 2.2: Deploy and Verify

```bash
# Apply deployment
kubectl apply -f config/manager/manager.yaml

# Check replicas
kubectl get deployment database-operator

# Check pods
kubectl get pods -l control-plane=controller-manager

# Verify only one is leader
kubectl logs -l control-plane=controller-manager | grep -i leader
```

## Exercise 3: Configure Resource Limits

### Task 3.1: Set Resource Requests and Limits

Update deployment:

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### Task 3.2: Monitor Resource Usage

```bash
# Check resource usage
kubectl top pods -l control-plane=controller-manager

# Watch resource usage
watch kubectl top pods -l control-plane=controller-manager
```

## Exercise 4: Test Failover

### Task 4.1: Identify Leader

```bash
# Find leader pod
kubectl get pods -l control-plane=controller-manager -o wide

# Check which pod holds the lease
kubectl get lease database-operator-leader-election -o yaml

# Check logs to identify leader
kubectl logs <pod-name> | grep -i "became leader"
```

### Task 4.2: Simulate Leader Failure

```bash
# Delete leader pod
kubectl delete pod <leader-pod-name>

# Watch failover
watch kubectl get pods -l control-plane=controller-manager

# Check new leader
kubectl get lease database-operator-leader-election -o yaml

# Verify reconciliation continues
kubectl logs -l control-plane=controller-manager | tail -20
```

## Exercise 5: Pod Disruption Budget

### Task 5.1: Create PDB

Create `config/manager/pdb.yaml`:

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: database-operator-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      control-plane: controller-manager
```

### Task 5.2: Test PDB

```bash
# Apply PDB
kubectl apply -f config/manager/pdb.yaml

# Try to drain node (should be blocked if it would violate PDB)
kubectl drain <node-name> --ignore-daemonsets

# Verify PDB is protecting pods
kubectl get pdb database-operator-pdb
```

## Cleanup

```bash
# Delete PDB
kubectl delete pdb database-operator-pdb

# Scale down
kubectl scale deployment database-operator --replicas=1
```

## Lab Summary

In this lab, you:
- Enabled leader election
- Deployed multiple replicas
- Configured resource limits
- Tested failover scenarios
- Set up Pod Disruption Budget

## Key Learnings

1. Leader election ensures only one active controller
2. Multiple replicas provide redundancy
3. Resource limits prevent exhaustion
4. Failover is automatic
5. PDB protects availability
6. Health checks ensure operator health

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Leader Election Configuration](../solutions/leader-election.go) - Complete leader election setup
- [HA Deployment](../solutions/ha-deployment.yaml) - HA deployment with PDB

## Next Steps

Now let's optimize performance!

**Navigation:** [← Previous Lab: RBAC](lab-02-rbac-security.md) | [Related Lesson](../lessons/03-high-availability.md) | [Next Lab: Performance →](lab-04-performance-scalability.md)

