---
layout: default
title: "Lab 05.4: Webhook Deployment"
nav_order: 14
parent: "Module 5: Webhooks & Admission Control"
grand_parent: Modules
mermaid: true
---

# Lab 5.4: Webhook Deployment and Certificates

**Related Lesson:** [Lesson 5.4: Webhook Deployment and Certificates](../lessons/04-webhook-deployment.md)  
**Navigation:** [← Previous Lab: Mutating Webhooks](lab-03-mutating-webhooks.md) | [Module Overview](../README.md)

## Objectives

- Understand certificate requirements for webhooks
- Deploy operator with webhooks to cluster
- Configure cert-manager for certificate management
- Troubleshoot webhook issues

## Prerequisites

- Completion of [Lab 5.3](lab-03-mutating-webhooks.md)
- Database operator with webhooks
- Understanding of TLS and certificates
- kind cluster with cert-manager installed (from `scripts/setup-kind-cluster.sh`)

## Understanding Webhook Certificate Requirements

Webhooks require TLS certificates because the Kubernetes API server communicates with webhooks over HTTPS. The certificate must be trusted by the API server.

**Two approaches:**
1. **cert-manager (Recommended)** - Automatically manages certificates in the cluster
2. **Manual certificates** - For special cases only

> **Note:** The kubebuilder-generated project is already configured to work with cert-manager. The `config/default/kustomization.yaml` includes cert-manager resources.

## Exercise 1: Verify Cert-Manager Setup

### Task 1.1: Check Cert-Manager is Running

```bash
# If you used scripts/setup-kind-cluster.sh, cert-manager is already installed
kubectl get pods -n cert-manager

# Should show:
# cert-manager-xxx          Running
# cert-manager-cainjector-xxx   Running
# cert-manager-webhook-xxx      Running
```

### Task 1.2: Install Cert-Manager (if not installed)

```bash
# Only if cert-manager is not running
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

# Wait for cert-manager to be ready
kubectl wait --for=condition=Available deployment/cert-manager -n cert-manager --timeout=120s
kubectl wait --for=condition=Available deployment/cert-manager-webhook -n cert-manager --timeout=120s
kubectl wait --for=condition=Available deployment/cert-manager-cainjector -n cert-manager --timeout=120s
```

## Exercise 2: Examine Kubebuilder's Cert-Manager Configuration

### Task 2.1: Check Certificate Configuration

```bash
cd ~/postgres-operator

# Check cert-manager configuration
cat config/certmanager/certificate-*.yaml
```

This defines a Certificate resource that cert-manager will use to generate TLS certificates for the webhook.

### Task 2.2: Check Kustomization

```bash
# Check how cert-manager is integrated
cat config/default/kustomization.yaml
```

The `config/default/kustomization.yaml` should include `../certmanager` to enable cert-manager integration.

## Exercise 3: Deploy Operator with Webhooks

### Task 3.1: Build the Operator Image

```bash
cd ~/postgres-operator

# Build the container image
make docker-build IMG=postgres-operator:latest

# For Podman users:
# make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman
```

### Task 3.2: Load Image into Kind

```bash
# For Docker:
kind load docker-image postgres-operator:latest --name k8s-operators-course

# For Podman:
# podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
# kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
# rm /tmp/postgres-operator.tar
```

### Task 3.3: Deploy to Cluster

```bash
# Deploy operator with webhooks
make deploy IMG=postgres-operator:latest

# For Podman users:
# make deploy IMG=localhost/postgres-operator:latest
```

### Task 3.4: Verify Deployment

```bash
# Check deployment
kubectl get deployment -n postgres-operator-system

# Check pods
kubectl get pods -n postgres-operator-system

# Wait for pod to be ready
kubectl wait --for=condition=Ready pod -l control-plane=controller-manager -n postgres-operator-system --timeout=120s
```

## Exercise 4: Verify Webhook Configuration

### Task 4.1: Check Webhook Configurations

```bash
# Check validating webhook
kubectl get validatingwebhookconfigurations

# Check mutating webhook  
kubectl get mutatingwebhookconfigurations

# Get details
kubectl describe validatingwebhookconfiguration postgres-operator-validating-webhook-configuration
```

### Task 4.2: Check Certificates

```bash
# Check certificate was created by cert-manager
kubectl get certificate -n postgres-operator-system

# Check certificate status
kubectl describe certificate -n postgres-operator-system

# Check the secret containing TLS certs
kubectl get secret -n postgres-operator-system | grep tls
```

### Task 4.3: Check Webhook Service

```bash
# Check webhook service
kubectl get service -n postgres-operator-system

# Check service endpoints
kubectl get endpoints -n postgres-operator-system
```

## Exercise 5: Test Webhooks

### Task 5.1: Test Mutating Webhook (Defaults)

```bash
# Create resource with minimal spec
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: webhook-test
spec:
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# Check if defaults were applied
echo "Image (should be defaulted):"
kubectl get database webhook-test -o jsonpath='{.spec.image}'
echo

echo "Replicas (should be defaulted):"
kubectl get database webhook-test -o jsonpath='{.spec.replicas}'
echo

echo "Labels (should include managed-by):"
kubectl get database webhook-test -o jsonpath='{.metadata.labels}'
echo
```

### Task 5.2: Test Validating Webhook (Rejection)

```bash
# Try to create invalid resource
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: invalid-test
spec:
  image: nginx:latest  # Invalid - not a postgres image
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# Should be rejected with validation error
```

### Task 5.3: Test Update Validation

```bash
# Try to reduce storage (should fail)
kubectl patch database webhook-test --type merge -p '{"spec":{"storage":{"size":"5Gi"}}}'

# Should be rejected
```

## Exercise 6: Troubleshoot Webhook Issues

### Task 6.1: Check Operator Logs

```bash
# Get operator logs
kubectl logs -n postgres-operator-system deployment/postgres-operator-controller-manager

# Look for webhook-related messages
kubectl logs -n postgres-operator-system deployment/postgres-operator-controller-manager | grep -i webhook
```

### Task 6.2: Check Certificate Status

```bash
# Check certificate status
kubectl get certificate -n postgres-operator-system -o wide

# Describe for details
kubectl describe certificate -n postgres-operator-system

# Check cert-manager logs if certificate not ready
kubectl logs -n cert-manager deployment/cert-manager
```

### Task 6.3: Common Issues and Fixes

**Issue: Certificate not ready**
```bash
# Check cert-manager is running
kubectl get pods -n cert-manager

# Check certificate events
kubectl describe certificate -n postgres-operator-system
```

**Issue: Webhook connection refused**
```bash
# Check service endpoints
kubectl get endpoints -n postgres-operator-system

# Check pod is running
kubectl get pods -n postgres-operator-system
```

**Issue: CA bundle mismatch**
```bash
# Check CA bundle in webhook config
kubectl get validatingwebhookconfiguration -o jsonpath='{.items[0].webhooks[0].clientConfig.caBundle}' | base64 -d | openssl x509 -text -noout | head -20
```

## Cleanup

```bash
# Delete test resources
kubectl delete databases --all

# Undeploy operator
make undeploy
```

## Lab Summary

In this lab, you:
- Verified cert-manager setup
- Examined kubebuilder's cert-manager integration
- Deployed operator with webhooks to cluster
- Verified webhook configuration and certificates
- Tested webhook functionality
- Learned troubleshooting techniques

## Key Learnings

1. Webhooks require TLS certificates - API server must trust them
2. cert-manager automatically manages certificates in the cluster
3. Kubebuilder projects are pre-configured for cert-manager
4. Webhooks cannot easily work with `make run` (requires in-cluster deployment)
5. Use `kubectl describe` and logs for troubleshooting
6. Certificate issues are common - check cert-manager status first

## Solutions

This lab focuses on deployment and certificates. For webhook implementation, refer to:
- [Validating Webhook](../solutions/validating-webhook.go) - From Lab 5.2
- [Mutating Webhook](../solutions/mutating-webhook.go) - From Lab 5.3

## Congratulations!

You've completed Module 5! You now understand:
- Admission control and webhooks
- Validating webhooks for custom validation
- Mutating webhooks for defaulting
- Certificate management with cert-manager
- Webhook deployment and troubleshooting

In Module 6, you'll learn about testing and debugging operators!

**Navigation:** [← Previous Lab: Mutating Webhooks](lab-03-mutating-webhooks.md) | [Module Overview](../README.md) | [Next: Module 6 →](../../module-06/README.md)
