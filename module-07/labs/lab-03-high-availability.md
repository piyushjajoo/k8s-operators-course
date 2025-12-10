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

Kubebuilder's generated `cmd/main.go` already supports leader election via the `--leader-elect` flag.

### Task 1.1: Review Leader Election Code

```bash
# Navigate to your operator project
cd ~/postgres-operator

# Review the leader election setup in main.go
grep -A 20 "LeaderElection" cmd/main.go
```

You should see code like:

```go
var enableLeaderElection bool
flag.BoolVar(&enableLeaderElection, "leader-elect", false,
    "Enable leader election for controller manager.")

mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    // ... other options ...
    LeaderElection:         enableLeaderElection,
    LeaderElectionID:       "your-operator-leader-election",
})
```

### Task 1.2: Enable Leader Election in Deployment

Update `config/manager/manager.yaml` to add the `--leader-elect` flag:

```yaml
spec:
  template:
    spec:
      containers:
      - name: manager
        args:
        - --leader-elect
        - --health-probe-bind-address=:8081
```

### Task 1.3: Deploy and Verify

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

# Check for lease object
kubectl get lease -n postgres-operator-system

# Check logs for leader election
kubectl logs -n postgres-operator-system -l control-plane=controller-manager | grep -i "leader"
```

## Exercise 2: Deploy Multiple Replicas

### Task 2.1: Update Deployment Replicas

Edit `config/manager/manager.yaml` to increase replicas:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
  namespace: system
spec:
  replicas: 3  # Change from 1 to 3
  selector:
    matchLabels:
      control-plane: controller-manager
  template:
    spec:
      containers:
      - name: manager
        args:
        - --leader-elect  # Required for HA
        - --health-probe-bind-address=:8081
        resources:
          limits:
            cpu: 500m
            memory: 128Mi
          requests:
            cpu: 10m
            memory: 64Mi
```

### Task 2.2: Deploy and Verify

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

# Check replicas
kubectl get deployment -n postgres-operator-system

# Check all pods are running
kubectl get pods -n postgres-operator-system -l control-plane=controller-manager

# Verify only one is leader (check logs)
for pod in $(kubectl get pods -n postgres-operator-system -l control-plane=controller-manager -o name); do
  echo "=== $pod ==="
  kubectl logs -n postgres-operator-system $pod | grep -i "leader" | tail -2
done
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
# List all pods
kubectl get pods -n postgres-operator-system -l control-plane=controller-manager

# Find the lease and identify the leader
kubectl get lease -n postgres-operator-system -o yaml

# The holderIdentity field shows which pod is the leader
# Look for the pod name in the holderIdentity

# Check logs to confirm leader
LEADER_POD=$(kubectl get lease -n postgres-operator-system -o jsonpath='{.items[0].spec.holderIdentity}' | cut -d'_' -f1)
echo "Leader pod: $LEADER_POD"
kubectl logs -n postgres-operator-system $LEADER_POD | grep -i "became leader"
```

### Task 4.2: Simulate Leader Failure

```bash
# Get the leader pod name
LEADER_POD=$(kubectl get lease -n postgres-operator-system -o jsonpath='{.items[0].spec.holderIdentity}' | cut -d'_' -f1)

# Delete the leader pod
kubectl delete pod -n postgres-operator-system $LEADER_POD

# Watch failover happen
watch kubectl get pods -n postgres-operator-system -l control-plane=controller-manager

# In another terminal, watch the lease
watch kubectl get lease -n postgres-operator-system -o jsonpath='{.items[0].spec.holderIdentity}'

# After a new leader is elected, verify reconciliation continues
kubectl logs -n postgres-operator-system -l control-plane=controller-manager --tail=20 | grep -i "reconcil"
```

## Exercise 5: Pod Disruption Budget

### Task 5.1: Create PDB

Create `config/manager/pdb.yaml`:

```bash
cat > config/manager/pdb.yaml << 'EOF'
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: controller-manager-pdb
  namespace: system
spec:
  minAvailable: 2
  selector:
    matchLabels:
      control-plane: controller-manager
EOF
```

Add to `config/manager/kustomization.yaml`:

```yaml
resources:
- manager.yaml
- pdb.yaml
```

### Task 5.2: Deploy and Test PDB

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

# Verify PDB is created
kubectl get pdb -n postgres-operator-system

# Check PDB status
kubectl describe pdb -n postgres-operator-system postgres-operator-controller-manager-pdb
```

**Important:** PDB only protects against **voluntary disruptions** (evictions), NOT direct `kubectl delete pod` commands!

### Task 5.3: Test PDB with Rollout Restart

The easiest way to test PDB is using `kubectl rollout restart`, which uses the eviction API internally:

```bash
# Check current PDB status - note ALLOWED DISRUPTIONS
kubectl get pdb -n postgres-operator-system

# Expected output:
# NAME                                        MIN AVAILABLE   MAX UNAVAILABLE   ALLOWED DISRUPTIONS   AGE
# postgres-operator-controller-manager-pdb   2               N/A               1                     5m

# Trigger a rolling restart (this respects PDB)
kubectl rollout restart deployment/postgres-operator-controller-manager -n postgres-operator-system

# Watch the rollout - PDB ensures at least 2 pods remain available
kubectl get pods -n postgres-operator-system -l control-plane=controller-manager -w

# In another terminal, watch PDB status during rollout
watch kubectl get pdb -n postgres-operator-system
```

**What you should observe:**
- Pods are replaced one at a time (not all at once)
- `ALLOWED DISRUPTIONS` changes as pods are terminated/created
- At least 2 pods remain `Running` throughout the rollout

### Task 5.4: Understand How PDB Interacts with Rollouts

Let's understand the math behind PDB:

```bash
# Check current state
kubectl get pdb -n postgres-operator-system

# Formula: ALLOWED DISRUPTIONS = currentHealthy - minAvailable
# With 3 healthy pods and minAvailable=2: 3 - 2 = 1 disruption allowed
```

**Why `minAvailable=3` doesn't block rollouts:**
1. New pod (4th) is created and becomes healthy
2. Now `currentHealthy=4`, so `allowedDisruptions = 4 - 3 = 1`
3. One old pod can be terminated → rollout proceeds!

To truly block a rollout, you'd need `minAvailable > replicas`:

```bash
# Scale down to 2 replicas first
kubectl scale deployment/postgres-operator-controller-manager -n postgres-operator-system --replicas=2
kubectl wait --for=condition=available deployment/postgres-operator-controller-manager -n postgres-operator-system

# Set minAvailable=2 (equal to replicas)
kubectl patch pdb postgres-operator-controller-manager-pdb -n postgres-operator-system \
  --type='json' -p='[{"op": "replace", "path": "/spec/minAvailable", "value": 2}]'

# Check PDB - ALLOWED DISRUPTIONS should be 0
kubectl get pdb -n postgres-operator-system
# 2 healthy - 2 minimum = 0 allowed

# Try rollout restart - this WILL get stuck
kubectl rollout restart deployment/postgres-operator-controller-manager -n postgres-operator-system

# Watch - new pod created, but old pods can't be terminated
kubectl get pods -n postgres-operator-system -l control-plane=controller-manager -w

# You'll see 3 pods (2 old + 1 new), rollout stuck
# The deployment waits for allowedDisruptions > 0

# Check rollout status - it will be waiting
kubectl rollout status deployment/postgres-operator-controller-manager -n postgres-operator-system --timeout=30s
# Waiting for deployment "postgres-operator-controller-manager" rollout to finish: 1 old replicas are pending termination...

# Fix by reducing minAvailable
kubectl patch pdb postgres-operator-controller-manager-pdb -n postgres-operator-system \
  --type='json' -p='[{"op": "replace", "path": "/spec/minAvailable", "value": 1}]'

# Now rollout can complete
kubectl rollout status deployment/postgres-operator-controller-manager -n postgres-operator-system

# Restore to 3 replicas and minAvailable=2
kubectl scale deployment/postgres-operator-controller-manager -n postgres-operator-system --replicas=3
kubectl patch pdb postgres-operator-controller-manager-pdb -n postgres-operator-system \
  --type='json' -p='[{"op": "replace", "path": "/spec/minAvailable", "value": 2}]'
```

**Key insight:** PDB blocks disruptions when `minAvailable >= currentHealthy`. During rollouts, new pods increase `currentHealthy`, which increases `allowedDisruptions`.

### Understanding PDB Behavior

```bash
# Check current PDB status
kubectl get pdb -n postgres-operator-system

# The columns mean:
# MIN AVAILABLE: Minimum pods that must remain running
# ALLOWED DISRUPTIONS: How many pods can be evicted right now
#
# Formula: ALLOWED DISRUPTIONS = currentHealthy - minAvailable
# Example: 3 healthy - 2 minimum = 1 allowed disruption
```

**Why `kubectl delete pod` doesn't respect PDB:**
- `kubectl delete` is a **direct deletion**, not an eviction
- PDB only protects against the **Eviction API** used by:
  - `kubectl drain` (node maintenance)
  - `kubectl rollout restart` (deployment updates)
  - Cluster Autoscaler (scale down)
  - Kubernetes scheduler (pod preemption)
- In production, these tools use eviction, so PDB works as intended

## Cleanup

```bash
# Undeploy operator
make undeploy

# Or scale down for testing
kubectl scale deployment -n postgres-operator-system controller-manager --replicas=1
```

## Lab Summary

In this lab, you:
- Enabled leader election via `--leader-elect` flag
- Deployed multiple replicas by updating `config/manager/manager.yaml`
- Configured resource limits
- Tested failover by deleting leader pod
- Set up Pod Disruption Budget
- Tested PDB using the Eviction API

## Key Learnings

1. Leader election is enabled via command-line flag in kubebuilder
2. Increase replicas in `config/manager/manager.yaml` for HA
3. Use `make deploy` to apply all configurations
4. Failover is automatic - standby pods acquire the lease
5. **PDB only protects against voluntary disruptions** (evictions, not direct deletion)
6. Use `kubectl drain` or Eviction API to test PDB - NOT `kubectl delete pod`
7. Health checks are pre-configured by kubebuilder

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Leader Election Configuration](../solutions/leader-election.go) - Complete leader election setup
- [HA Deployment](../solutions/ha-deployment.yaml) - HA deployment with PDB

## Next Steps

Now let's optimize performance!

**Navigation:** [← Previous Lab: RBAC](lab-02-rbac-security.md) | [Related Lesson](../lessons/03-high-availability.md) | [Next Lab: Performance →](lab-04-performance-scalability.md)

