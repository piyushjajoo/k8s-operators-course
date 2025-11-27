# Lab 1.1: Exploring the Control Plane

**Related Lesson:** [Lesson 1.1: Kubernetes Control Plane Review](../lessons/01-control-plane.md)  
**Navigation:** [Module Overview](../README.md) | [Next Lab: API Machinery →](lab-02-api-machinery.md)

## Objectives

- Explore Kubernetes control plane components
- Understand how components interact
- Observe controller behavior in real-time
- Trace API request flows

## Prerequisites

- Kind cluster running
- kubectl configured and working

## Exercise 1: Inspect Control Plane Components

### Task 1.1: View Control Plane Pods

```bash
# List all control plane components
kubectl get pods -n kube-system

# Get detailed information about API server
kubectl get pods -n kube-system -l component=kube-apiserver -o yaml | head -50

# Check controller manager
kubectl get pods -n kube-system -l component=kube-controller-manager

# Check scheduler
kubectl get pods -n kube-system -l component=kube-scheduler
```

**Expected Output**: You should see pods for API server, controller manager, scheduler, and etcd.

### Task 1.2: Explore API Server

```bash
# Get cluster information
kubectl cluster-info

# Get API server version
kubectl version --output=yaml

# Discover available API groups
kubectl api-versions | head -20

# Get API resources
kubectl api-resources | grep -E "NAME|deployments|pods|services"
```

**Questions to Answer:**
1. What version of Kubernetes is running?
2. How many API groups are available?
3. What API version are Deployments using?

## Exercise 2: Observe Controller Behavior

### Task 2.1: Create and Observe a Deployment

```bash
# Create a deployment
kubectl create deployment nginx --image=nginx:latest --replicas=3

# Immediately watch the deployment
kubectl get deployment nginx -w &
DEPLOY_PID=$!

# In another terminal (or wait a moment), watch ReplicaSets
kubectl get replicasets -w &
RS_PID=$!

# Watch pods
kubectl get pods -l app=nginx -w &
POD_PID=$!

# Wait 30 seconds to observe the creation flow
sleep 30

# Stop watching
kill $DEPLOY_PID $RS_PID $POD_PID 2>/dev/null
```

**Observations:**
1. What order were resources created?
2. How long did it take for all pods to be ready?
3. What status fields changed during creation?

### Task 2.2: Trace Resource Creation

```bash
# Get the deployment with all details
kubectl get deployment nginx -o yaml > /tmp/nginx-deployment.yaml

# Get the ReplicaSet
kubectl get replicasets -l app=nginx -o yaml > /tmp/nginx-rs.yaml

# Get one of the pods
kubectl get pods -l app=nginx -o yaml | head -100 > /tmp/nginx-pod.yaml

# Examine the owner references
grep -A 5 "ownerReferences" /tmp/nginx-pod.yaml
grep -A 5 "ownerReferences" /tmp/nginx-rs.yaml
```

**Questions:**
1. What is the relationship between Deployment, ReplicaSet, and Pod?
2. How are owner references used?

## Exercise 3: View Controller Logs

### Task 3.1: Controller Manager Logs

```bash
# View recent controller manager logs
kubectl logs -n kube-system -l component=kube-controller-manager --tail=50

# Filter for deployment-related logs
kubectl logs -n kube-system -l component=kube-controller-manager --tail=100 | grep -i deployment

# Watch logs in real-time
kubectl logs -n kube-system -l component=kube-controller-manager -f --tail=20
```

**In another terminal, trigger an action:**
```bash
# Scale the deployment
kubectl scale deployment nginx --replicas=5

# Watch the logs to see controller activity
```

### Task 3.2: Scheduler Logs

```bash
# View scheduler logs
kubectl logs -n kube-system -l component=kube-scheduler --tail=50

# Look for scheduling decisions
kubectl logs -n kube-system -l component=kube-scheduler --tail=100 | grep -i "scheduled"
```

## Exercise 4: Direct API Interaction

### Task 4.1: Use kubectl proxy

```bash
# Start kubectl proxy in background
kubectl proxy --port=8001 &
PROXY_PID=$!

# Wait for proxy to start
sleep 2

# Make direct API calls
curl http://localhost:8001/api/v1/namespaces

# Get pods via API
curl http://localhost:8001/api/v1/namespaces/default/pods | jq '.items[].metadata.name' | head -5

# Get the nginx deployment
curl http://localhost:8001/apis/apps/v1/namespaces/default/deployments/nginx | jq '.spec.replicas'

# Stop the proxy
kill $PROXY_PID
```

### Task 4.2: Create Resource via API

```bash
# Start proxy again
kubectl proxy --port=8001 &
PROXY_PID=$!
sleep 2

# Create a pod via direct API call
curl -X POST http://localhost:8001/api/v1/namespaces/default/pods \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
      "name": "api-created-pod",
      "namespace": "default"
    },
    "spec": {
      "containers": [{
        "name": "nginx",
        "image": "nginx:latest"
      }]
    }
  }'

# Verify it was created
kubectl get pod api-created-pod

# Clean up
kubectl delete pod api-created-pod
kill $PROXY_PID
```

## Exercise 5: Observe Reconciliation

### Task 5.1: Manual Pod Deletion

```bash
# Get a pod name
POD_NAME=$(kubectl get pods -l app=nginx -o jsonpath='{.items[0].metadata.name}')

# Delete the pod
kubectl delete pod $POD_NAME

# Immediately watch for recreation
kubectl get pods -l app=nginx -w

# The ReplicaSet controller should recreate it!
```

### Task 5.2: Change Desired State

```bash
# Scale down
kubectl scale deployment nginx --replicas=2

# Watch pods being terminated
kubectl get pods -l app=nginx -w

# Scale up
kubectl scale deployment nginx --replicas=4

# Watch pods being created
kubectl get pods -l app=nginx -w
```

## Cleanup

```bash
# Delete the deployment (this will cascade delete ReplicaSet and Pods)
kubectl delete deployment nginx
```

## Lab Summary

In this lab, you:
- Explored control plane components
- Observed controller behavior in real-time
- Traced API request flows
- Understood the reconciliation process
- Interacted with the Kubernetes API directly

## Key Learnings

1. Control plane components work together to manage the cluster
2. Controllers continuously watch and reconcile resources
3. The API Server is the central communication hub
4. Owner references maintain resource relationships
5. Reconciliation happens automatically when desired != actual state

**Navigation:** [← Module Overview](../README.md) | [Related Lesson](../lessons/01-control-plane.md) | [Next Lab: API Machinery →](lab-02-api-machinery.md)

