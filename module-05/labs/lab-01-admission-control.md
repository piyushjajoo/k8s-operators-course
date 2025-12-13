---
layout: default
title: "Lab 05.1: Admission Control"
nav_order: 11
parent: "Module 5: Webhooks & Admission Control"
grand_parent: Modules
permalink: /module-05/labs/admission-control/
mermaid: true
---

# Lab 5.1: Exploring Admission Control

**Related Lesson:** [Lesson 5.1: Kubernetes Admission Control](../lessons/01-admission-control.md)  
**Navigation:** [Module Overview](../README.md) | [Next Lab: Validating Webhooks →](lab-02-validating-webhooks.md)

## Objectives

- Explore existing admission controllers
- Understand webhook configuration
- Test webhook endpoints
- Understand admission control flow

## Prerequisites

- Completion of [Module 4](../../module-04/README.md)
- Kind cluster running
- Understanding of admission control concepts

## Exercise 1: Explore Built-in Admission Controllers

### Task 1.1: List Admission Controllers

```bash
# Check API server admission plugins enabled
# Note: Modern Kubernetes clusters have many default admission controllers enabled
# automatically (NamespaceLifecycle, LimitRanger, ServiceAccount, ResourceQuota,
# MutatingAdmissionWebhook, ValidatingAdmissionWebhook, etc.)

# For kind cluster, check API server args for admission plugins
kubectl get pod -n kube-system -l component=kube-apiserver -o yaml | grep -A 10 "admission"

# Alternative: Check the kube-apiserver manifest directly (kind-specific)
docker exec kind-control-plane cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep -i admission

# If no --enable-admission-plugins flag is shown, the cluster uses the default set
# See: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#which-plugins-are-enabled-by-default
```

### Task 1.2: Test ResourceQuota Admission

```bash
# Create a namespace with quota
kubectl create namespace quota-test
kubectl create quota test-quota --namespace=quota-test --hard=cpu=1,memory=1Gi

# Try to create pod that exceeds quota
kubectl run test-pod --image=nginx:latest --namespace=quota-test --overrides='{"spec":{"containers":[{"name":"test-pod","image":"nginx:latest","resources":{"requests":{"cpu":"2","memory":"2Gi"}}}]}}'

# Should be rejected by ResourceQuota admission controller
```

## Exercise 2: Explore Webhook Configurations

### Task 2.1: List Webhook Configurations

```bash
# List validating webhook configurations
kubectl get validatingwebhookconfigurations

# List mutating webhook configurations
kubectl get mutatingwebhookconfigurations

# Get details
kubectl get validatingwebhookconfiguration <name> -o yaml
```

### Task 2.2: Examine Webhook Structure

If you have any webhooks installed (e.g., from cert-manager):

```bash
# Get webhook configuration
kubectl get validatingwebhookconfiguration -o yaml | head -50

# Examine:
# - Rules (when webhook is called)
# - Client config (how to reach webhook)
# - Failure policy
```

## Exercise 3: Understand Webhook Rules

### Task 3.1: Analyze Rule Structure

Create a sample webhook configuration to understand structure:

```bash
# Create sample webhook config (won't work without service, but shows structure)
cat <<EOF | kubectl apply -f -
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: example-webhook
webhooks:
- name: example.example.com
  rules:
  - apiGroups: ["apps"]
    apiVersions: ["v1"]
    resources: ["deployments"]
    operations: ["CREATE", "UPDATE"]
  clientConfig:
    service:
      name: example-webhook-service
      namespace: default
      path: "/validate"
  admissionReviewVersions: ["v1"]
  sideEffects: None
  failurePolicy: Fail
EOF

# Examine it
kubectl get validatingwebhookconfiguration example-webhook -o yaml

# Delete it (it won't work anyway)
kubectl delete validatingwebhookconfiguration example-webhook
```

## Exercise 4: Test Admission Flow

### Task 4.1: Create Resource and Trace Flow

```bash
# Create a pod with verbose output
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: nginx
    image: nginx:latest
EOF

# Watch events to see admission process
kubectl get events --sort-by='.lastTimestamp' | tail -20
```

### Task 4.2: Test Validation Failure

```bash
# Try to create invalid resource
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: invalid-pod
spec:
  containers:
  - name: nginx
    image: nginx:latest
    resources:
      requests:
        cpu: "invalid"  # Invalid value
EOF

# Observe validation error
# This is caught by schema validation, not webhook
```

## Exercise 5: Understand Mutating vs Validating

### Task 5.1: Observe Built-in Mutations

```bash
# Create a pod without namespace
cat <<EOF > /tmp/pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-mutation
spec:
  containers:
  - name: nginx
    image: nginx:latest
EOF

# Apply it
kubectl apply -f /tmp/pod.yaml

# Check what was added (mutations)
kubectl get pod test-mutation -o yaml | grep -A 10 "metadata:"

# Built-in admission controllers may add:
# - Default service account
# - Security context defaults
# - etc.
```

## Exercise 6: Webhook Service Requirements

### Task 6.1: Understand Service Requirements

For a webhook to work, you need:

1. **Service** - To route requests to webhook pods
2. **Certificate** - For TLS connection
3. **Webhook Configuration** - To register webhook
4. **Webhook Handler** - To process requests

```bash
# Check if any webhook services exist
kubectl get services -A | grep webhook

# Check webhook pods
kubectl get pods -A | grep webhook
```

## Cleanup

```bash
# Clean up test resources
kubectl delete pod test-pod test-mutation invalid-pod 2>/dev/null || true
kubectl delete namespace quota-test 2>/dev/null || true
rm -f /tmp/pod.yaml
```

## Lab Summary

In this lab, you:
- Explored built-in admission controllers
- Examined webhook configurations
- Understood webhook rules and structure
- Traced admission flow
- Observed mutations
- Understood webhook requirements

## Key Learnings

1. Admission control intercepts API requests
2. Mutating webhooks run before validating
3. Webhook configurations define when webhooks are called
4. Webhooks need services and certificates
5. Built-in admission controllers provide basic functionality
6. Custom webhooks extend validation/mutation

## Next Steps

Now let's build your own validating webhook!

**Navigation:** [← Module Overview](../README.md) | [Related Lesson](../lessons/01-admission-control.md) | [Next Lab: Validating Webhooks →](lab-02-validating-webhooks.md)
