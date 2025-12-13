---
layout: default
title: "Lab 01.2: Api Machinery"
nav_order: 12
parent: "Module 1: Kubernetes Architecture"
grand_parent: Modules
mermaid: true
---

# Lab 1.2: Working with the Kubernetes API

**Related Lesson:** [Lesson 1.2: Kubernetes API Machinery](../lessons/02-api-machinery.md)  
**Navigation:** [← Previous Lab: Control Plane](lab-01-control-plane.md) | [Module Overview](../README.md) | [Next Lab: Controller Pattern →](lab-03-controller-pattern.md)

## Objectives

- Understand Kubernetes API structure
- Discover API groups and versions
- Make direct API calls
- Understand resource structure (spec vs status)
- Work with resource versions

## Prerequisites

- Kind cluster running
- kubectl configured
- `jq` installed (optional, for JSON parsing)

## Exercise 1: API Discovery

### Task 1.1: Explore API Versions

```bash
# List all API versions
kubectl api-versions

# Count API groups
kubectl api-versions | cut -d'/' -f1 | sort -u | wc -l

# List core API group resources
kubectl api-versions | grep "^v1$"

# List apps API group
kubectl api-versions | grep "^apps/"
```

**Questions:**
1. How many API groups are there?
2. What's the difference between `/api/v1` and `/apis/apps/v1`?

### Task 1.2: Discover API Resources

```bash
# List all API resources
kubectl api-resources

# List resources in core group
kubectl api-resources --api-group=""

# List resources in apps group
kubectl api-resources --api-group="apps"

# Get detailed information
kubectl api-resources -o wide | head -20
```

### Task 1.3: Explore API Group Details

```bash
# Get apps API group information
kubectl get --raw /apis/apps/v1 | jq '.'

# See what resources are available
kubectl get --raw /apis/apps/v1 | jq '.resources[].name'

# Get deployment API details
kubectl get --raw /apis/apps/v1 | jq '.resources[] | select(.name == "deployments")'
```

## Exercise 2: Resource Structure

### Task 2.1: Examine Resource Structure

```bash
# Create a simple deployment
kubectl create deployment test-api --image=nginx:latest

# Get the deployment in YAML format
kubectl get deployment test-api -o yaml > /tmp/deployment.yaml

# Examine the structure
cat /tmp/deployment.yaml

# Extract just the spec
kubectl get deployment test-api -o jsonpath='{.spec}' | jq '.'

# Extract just the status
kubectl get deployment test-api -o jsonpath='{.status}' | jq '.'
```

**Observations:**
1. What fields are in `spec`?
2. What fields are in `status`?
3. How do they differ?

### Task 2.2: Compare Spec vs Status

```bash
# Get desired replicas (from spec)
kubectl get deployment test-api -o jsonpath='{.spec.replicas}'
echo

# Get actual replicas (from status)
kubectl get deployment test-api -o jsonpath='{.status.replicas}'
echo

# Get ready replicas
kubectl get deployment test-api -o jsonpath='{.status.readyReplicas}'
echo

# Wait for deployment to be ready, then check again
kubectl wait --for=condition=available deployment/test-api --timeout=60s
kubectl get deployment test-api -o jsonpath='{.spec.replicas}' && echo " (desired)"
kubectl get deployment test-api -o jsonpath='{.status.replicas}' && echo " (actual)"
kubectl get deployment test-api -o jsonpath='{.status.readyReplicas}' && echo " (ready)"
```

## Exercise 3: Direct API Calls

### Task 3.1: Start kubectl Proxy

```bash
# Start proxy
kubectl proxy --port=8001 &
PROXY_PID=$!

# Wait for it to start
sleep 2

# Test connectivity
curl http://localhost:8001/api/v1
```

### Task 3.2: List Resources via API

```bash
# List namespaces
curl -s http://localhost:8001/api/v1/namespaces | jq '.items[].metadata.name'

# List pods in default namespace
curl -s http://localhost:8001/api/v1/namespaces/default/pods | jq '.items[].metadata.name'

# List deployments
curl -s http://localhost:8001/apis/apps/v1/namespaces/default/deployments | jq '.items[].metadata.name'
```

### Task 3.3: Get Specific Resource

```bash
# Get the test-api deployment
curl -s http://localhost:8001/apis/apps/v1/namespaces/default/deployments/test-api | jq '.metadata.name'
curl -s http://localhost:8001/apis/apps/v1/namespaces/default/deployments/test-api | jq '.spec.replicas'
curl -s http://localhost:8001/apis/apps/v1/namespaces/default/deployments/test-api | jq '.status'
```

### Task 3.4: Create Resource via API

```bash
# Create a pod via API
curl -X POST http://localhost:8001/api/v1/namespaces/default/pods \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
      "name": "api-pod",
      "namespace": "default"
    },
    "spec": {
      "containers": [{
        "name": "nginx",
        "image": "nginx:latest"
      }]
    }
  }' | jq '.metadata.name'

# Verify it was created
kubectl get pod api-pod
```

### Task 3.5: Update Resource via API

```bash
# Get current resource version
RV=$(curl -s http://localhost:8001/api/v1/namespaces/default/pods/api-pod | jq -r '.metadata.resourceVersion')
echo "Current resourceVersion: $RV"

# Add a label
curl -X PATCH http://localhost:8001/api/v1/namespaces/default/pods/api-pod \
  -H "Content-Type: application/merge-patch+json" \
  -d '{
    "metadata": {
      "labels": {
        "created-by": "api"
      }
    }
  }' | jq '.metadata.labels'

# Verify the label
kubectl get pod api-pod --show-labels
```

## Exercise 4: Resource Versions

### Task 4.1: Understand Resource Version

```bash
# Get resource version
kubectl get pod api-pod -o jsonpath='{.metadata.resourceVersion}'
echo

# Make a change
kubectl label pod api-pod test=value

# Get resource version again (should be different)
kubectl get pod api-pod -o jsonpath='{.metadata.resourceVersion}'
echo
```

### Task 4.2: Optimistic Concurrency

```bash
# Get current resource version
RV1=$(kubectl get pod api-pod -o jsonpath='{.metadata.resourceVersion}')

# Try to update with old resource version (should fail or conflict)
curl -X PATCH http://localhost:8001/api/v1/namespaces/default/pods/api-pod \
  -H "Content-Type: application/merge-patch+json" \
  -d "{
    \"metadata\": {
      \"resourceVersion\": \"$RV1\",
      \"labels\": {
        \"test2\": \"value2\"
      }
    }
  }" | jq '.'

# Get new resource version
RV2=$(kubectl get pod api-pod -o jsonpath='{.metadata.resourceVersion}')

# Update with correct resource version
curl -X PATCH http://localhost:8001/api/v1/namespaces/default/pods/api-pod \
  -H "Content-Type: application/merge-patch+json" \
  -d "{
    \"metadata\": {
      \"resourceVersion\": \"$RV2\",
      \"labels\": {
        \"test2\": \"value2\"
      }
    }
  }" | jq '.metadata.labels'
```

## Exercise 5: Subresources

### Task 5.1: Status Subresource

```bash
# Get status subresource
curl -s http://localhost:8001/api/v1/namespaces/default/pods/api-pod/status | jq '.phase'

# Get scale subresource (for deployments)
curl -s http://localhost:8001/apis/apps/v1/namespaces/default/deployments/test-api/scale | jq '.'
```

### Task 5.2: Scale via Subresource

```bash
# Get current scale
curl -s http://localhost:8001/apis/apps/v1/namespaces/default/deployments/test-api/scale | jq '.spec.replicas'

# Update scale
curl -X PATCH http://localhost:8001/apis/apps/v1/namespaces/default/deployments/test-api/scale \
  -H "Content-Type: application/merge-patch+json" \
  -d '{
    "spec": {
      "replicas": 3
    }
  }' | jq '.spec.replicas'

# Verify
kubectl get deployment test-api
```

## Cleanup

```bash
# Stop proxy
kill $PROXY_PID 2>/dev/null

# Delete resources
kubectl delete deployment test-api
kubectl delete pod api-pod
```

## Lab Summary

In this lab, you:
- Discovered API groups and versions
- Explored resource structure (spec vs status)
- Made direct API calls using kubectl proxy
- Understood resource versions and optimistic concurrency
- Worked with subresources (status, scale)

## Key Learnings

1. Kubernetes API is RESTful and organized into groups
2. Resources have consistent structure: apiVersion, kind, metadata, spec, status
3. Spec describes desired state, status describes actual state
4. Resource versions enable optimistic concurrency control
5. Subresources provide additional functionality (status, scale, exec, etc.)

**Navigation:** [← Previous Lab: Control Plane](lab-01-control-plane.md) | [Related Lesson](../lessons/02-api-machinery.md) | [Next Lab: Controller Pattern →](lab-03-controller-pattern.md)
