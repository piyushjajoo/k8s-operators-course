---
layout: default
title: "Lab 01.3: Controller Pattern"
nav_order: 13
parent: "Module 1: Kubernetes Architecture"
grand_parent: Modules
mermaid: true
---

# Lab 1.3: Observing Controllers in Action

**Related Lesson:** [Lesson 1.3: The Controller Pattern](../lessons/03-controller-pattern.md)  
**Navigation:** [← Previous Lab: API Machinery](lab-02-api-machinery.md) | [Module Overview](../README.md) | [Next Lab: Custom Resources →](lab-04-custom-resources.md)

## Objectives

- Observe controller reconciliation in real-time
- Understand the control loop pattern
- See declarative vs imperative behavior
- Test idempotency
- Understand watch mechanisms

## Prerequisites

- Kind cluster running
- kubectl configured

## Exercise 1: Observe Reconciliation Loop

### Task 1.1: Create and Watch Deployment

```bash
# Create a deployment
kubectl create deployment controller-demo --image=nginx:latest --replicas=2

# Watch deployment in one terminal
kubectl get deployment controller-demo -w

# In another terminal, watch ReplicaSet
kubectl get replicasets -w

# In another terminal, watch pods
kubectl get pods -l app=controller-demo -w
```

**Observations:**
1. What was created first?
2. How long until all resources were ready?
3. What status fields changed?

### Task 1.2: Trace the Reconciliation

```bash
# Get events to see the flow
kubectl get events --sort-by='.lastTimestamp' | grep controller-demo

# Get detailed deployment info
kubectl get deployment controller-demo -o yaml | grep -A 10 status:

# Check ReplicaSet owner reference
kubectl get replicasets -l app=controller-demo -o yaml | grep -A 10 ownerReferences

# Check Pod owner references
kubectl get pods -l app=controller-demo -o yaml | grep -A 10 ownerReferences
```

## Exercise 2: Test Reconciliation

### Task 2.1: Manual Pod Deletion

```bash
# Get a pod name
POD_NAME=$(kubectl get pods -l app=controller-demo -o jsonpath='{.items[0].metadata.name}')
echo "Deleting pod: $POD_NAME"

# Delete the pod
kubectl delete pod $POD_NAME

# Immediately watch for recreation
echo "Watching for pod recreation..."
kubectl get pods -l app=controller-demo -w
```

**Questions:**
1. How quickly was the pod recreated?
2. Which controller recreated it?
3. What does this tell you about the control loop?

### Task 2.2: Change Desired State

```bash
# Scale up
kubectl scale deployment controller-demo --replicas=5

# Watch pods being created
kubectl get pods -l app=controller-demo -w

# Scale down
kubectl scale deployment controller-demo --replicas=1

# Watch pods being terminated
kubectl get pods -l app=controller-demo -w
```

**Observations:**
1. How does the controller handle scaling up?
2. How does it handle scaling down?
3. What's the order of operations?

## Exercise 3: Declarative Behavior

### Task 3.1: Apply Same Resource Multiple Times

```bash
# Create a deployment manifest
cat <<EOF > /tmp/test-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: declarative-test
spec:
  replicas: 3
  selector:
    matchLabels:
      app: declarative
  template:
    metadata:
      labels:
        app: declarative
    spec:
      containers:
      - name: nginx
        image: nginx:latest
EOF

# Apply it
kubectl apply -f /tmp/test-deployment.yaml

# Wait for it to be ready
kubectl wait --for=condition=available deployment/declarative-test --timeout=60s

# Count pods
kubectl get pods -l app=declarative | wc -l

# Apply the SAME file again
kubectl apply -f /tmp/test-deployment.yaml

# Count pods again (should be the same!)
kubectl get pods -l app=declarative | wc -l
```

**Key Learning:** Applying the same resource multiple times is idempotent - it doesn't create duplicates.

### Task 3.2: Modify and Re-apply

```bash
# Modify the manifest (change image)
cat <<EOF > /tmp/test-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: declarative-test
spec:
  replicas: 3
  selector:
    matchLabels:
      app: declarative
  template:
    metadata:
      labels:
        app: declarative
    spec:
      containers:
      - name: nginx
        image: nginx:1.21  # Changed from latest
EOF

# Apply the modified version
kubectl apply -f /tmp/test-deployment.yaml

# Watch the rolling update
kubectl get pods -l app=declarative -w
```

**Observations:**
1. What happened when you changed the image?
2. How did Kubernetes handle the update?
3. This is declarative - you described what you want, Kubernetes figured out how to achieve it.

## Exercise 4: Controller Logs

### Task 4.1: View Controller Manager Logs

```bash
# View recent logs
kubectl logs -n kube-system -l component=kube-controller-manager --tail=50

# Filter for our deployment
kubectl logs -n kube-system -l component=kube-controller-manager --tail=100 | grep declarative-test
```

### Task 4.2: Watch Logs During Action

```bash
# Start watching logs in background
kubectl logs -n kube-system -l component=kube-controller-manager -f --tail=20 > /tmp/controller.log &
LOG_PID=$!

# Trigger an action
kubectl scale deployment declarative-test --replicas=5

# Wait a moment
sleep 5

# Check the logs
cat /tmp/controller.log | tail -20

# Stop log watching
kill $LOG_PID
```

## Exercise 5: Test Idempotency

### Task 5.1: Multiple Applies

```bash
# Apply the same deployment 5 times
for i in {1..5}; do
  echo "Apply #$i"
  kubectl apply -f /tmp/test-deployment.yaml
  sleep 2
done

# Check how many deployments exist
kubectl get deployments declarative-test

# Check how many ReplicaSets exist
kubectl get replicasets -l app=declarative

# Check how many pods exist
kubectl get pods -l app=declarative
```

**Expected Result:** Only one deployment, one ReplicaSet, and the correct number of pods.

### Task 5.2: Verify Idempotency

```bash
# Get current state
kubectl get deployment declarative-test -o yaml > /tmp/before.yaml

# Apply again
kubectl apply -f /tmp/test-deployment.yaml

# Get state after
kubectl get deployment declarative-test -o yaml > /tmp/after.yaml

# Compare (they should be identical or very similar)
diff /tmp/before.yaml /tmp/after.yaml
```

## Exercise 6: Watch Mechanism

### Task 6.1: Use kubectl watch

```bash
# Watch deployments
kubectl get deployments -w

# In another terminal, make changes
kubectl scale deployment declarative-test --replicas=2
kubectl scale deployment declarative-test --replicas=4
```

**Observations:**
1. How quickly do you see updates?
2. What information is shown in the watch output?

### Task 6.2: Observe Event Stream

```bash
# Watch events
kubectl get events -w --sort-by='.lastTimestamp'

# In another terminal, trigger actions
kubectl scale deployment declarative-test --replicas=1
kubectl label deployment declarative-test env=test
```

## Exercise 7: Status Updates

### Task 7.1: Monitor Status Changes

```bash
# Watch status fields
watch -n 1 'kubectl get deployment declarative-test -o jsonpath="{.status.conditions[?(@.type==\"Available\")].status}"'

# In another terminal, scale
kubectl scale deployment declarative-test --replicas=0
kubectl scale deployment declarative-test --replicas=3
```

### Task 7.2: Compare Spec vs Status

```bash
# Get desired vs actual
echo "Desired replicas: $(kubectl get deployment declarative-test -o jsonpath='{.spec.replicas}')"
echo "Actual replicas: $(kubectl get deployment declarative-test -o jsonpath='{.status.replicas}')"
echo "Ready replicas: $(kubectl get deployment declarative-test -o jsonpath='{.status.readyReplicas}')"

# The controller continuously works to make actual match desired
```

## Cleanup

```bash
# Delete deployments
kubectl delete deployment controller-demo
kubectl delete deployment declarative-test

# Clean up temp files
rm -f /tmp/test-deployment.yaml /tmp/before.yaml /tmp/after.yaml /tmp/controller.log
```

## Lab Summary

In this lab, you:
- Observed controller reconciliation in real-time
- Tested declarative behavior
- Verified idempotency
- Understood the control loop pattern
- Monitored status updates

## Key Learnings

1. Controllers continuously reconcile desired vs actual state
2. Reconciliation happens automatically when state changes
3. Kubernetes uses a declarative model - describe what you want
4. Operations are idempotent - safe to repeat
5. Status fields reflect actual state, updated by controllers
6. Watch mechanisms provide real-time updates

**Navigation:** [← Previous Lab: API Machinery](lab-02-api-machinery.md) | [Related Lesson](../lessons/03-controller-pattern.md) | [Next Lab: Custom Resources →](lab-04-custom-resources.md)
