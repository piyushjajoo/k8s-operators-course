# Lab 5.4: Certificate Management

**Related Lesson:** [Lesson 5.4: Webhook Deployment and Certificates](../lessons/04-webhook-deployment.md)  
**Navigation:** [← Previous Lab: Mutating Webhooks](lab-03-mutating-webhooks.md) | [Module Overview](../README.md)

## Objectives

- Set up certificate management
- Configure webhook service
- Test webhooks locally
- Understand certificate rotation

## Prerequisites

- Completion of [Lab 5.3](lab-03-mutating-webhooks.md)
- Database operator with webhooks
- Understanding of TLS and certificates

## Exercise 1: Local Development Certificates

### Task 1.1: Generate Certificates with Kubebuilder

```bash
# Navigate to your operator
cd ~/postgres-operator

# Generate certificates
make certs

# Check what was created
ls -la config/certmanager/

# Install certificates
make install-cert
```

### Task 1.2: Examine Certificate Setup

```bash
# Check certificate manifests
cat config/certmanager/kustomization.yaml

# Check certificate resources
cat config/certmanager/certificates/*.yaml
```

## Exercise 2: Set Up cert-manager (Optional for Production)

### Task 2.1: Install cert-manager

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Wait for cert-manager to be ready
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager --timeout=300s
```

### Task 2.2: Create Issuer

```bash
# Create self-signed issuer
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: selfsigned-issuer
  namespace: default
spec:
  selfSigned: {}
EOF

# Verify issuer
kubectl get issuer selfsigned-issuer
```

### Task 2.3: Create Certificate

```bash
# Create certificate
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: database-webhook-cert
  namespace: default
spec:
  secretName: database-webhook-cert
  issuerRef:
    name: selfsigned-issuer
    kind: Issuer
  dnsNames:
  - database-webhook-service.default.svc
  - database-webhook-service.default.svc.cluster.local
EOF

# Wait for certificate
kubectl wait --for=condition=ready certificate/database-webhook-cert --timeout=60s

# Check certificate
kubectl get certificate database-webhook-cert
kubectl get secret database-webhook-cert
```

## Exercise 3: Configure Webhook Service

### Task 3.1: Check Service Configuration

```bash
# Check if service exists
kubectl get service database-webhook-service

# If not, check generated manifests
cat config/webhook/manifests.yaml | grep -A 20 "kind: Service"
```

### Task 3.2: Verify Service Endpoints

```bash
# Create service if needed (kubebuilder should generate it)
# Check service endpoints
kubectl get endpoints database-webhook-service

# Should point to operator pods
```

## Exercise 4: Test Webhook Locally

### Task 4.1: Run Operator with Webhooks

```bash
# Generate certs
make certs
make install-cert

# Run operator
make run
```

### Task 4.2: Test Webhook Connectivity

```bash
# In another terminal, test webhook
# Create a Database resource
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

# Check if defaults were applied (mutating webhook)
kubectl get database webhook-test -o jsonpath='{.spec.image}'
echo

# Check if validation worked (validating webhook)
# Try invalid resource
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: invalid-test
spec:
  image: nginx:latest  # Invalid
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# Should be rejected
```

## Exercise 5: Deploy to Cluster

### Task 5.1: Build and Deploy

```bash
# Build image
make docker-build IMG=database-operator:latest

# For kind, load image
kind load docker-image database-operator:latest --name k8s-operators-course

# Deploy
make deploy IMG=database-operator:latest
```

### Task 5.2: Verify Deployment

```bash
# Check deployment
kubectl get deployment database-controller-manager

# Check pods
kubectl get pods -l control-plane=controller-manager

# Check webhook service
kubectl get service database-webhook-service

# Check webhook configurations
kubectl get validatingwebhookconfiguration
kubectl get mutatingwebhookconfiguration
```

## Exercise 6: Troubleshoot Webhook Issues

### Task 6.1: Check Certificate Status

```bash
# Check certificate
kubectl get certificate database-webhook-cert

# Check certificate events
kubectl describe certificate database-webhook-cert

# Check secret
kubectl get secret database-webhook-cert -o yaml
```

### Task 6.2: Check Webhook Configuration

```bash
# Get webhook configuration
kubectl get validatingwebhookconfiguration -o yaml

# Check CA bundle
kubectl get validatingwebhookconfiguration -o jsonpath='{.items[0].webhooks[0].clientConfig.caBundle}' | base64 -d
```

### Task 6.3: Check Webhook Logs

```bash
# Check operator logs
kubectl logs -l control-plane=controller-manager

# Look for webhook-related errors
kubectl logs -l control-plane=controller-manager | grep -i webhook
```

## Cleanup

```bash
# Delete test resources
kubectl delete databases --all

# Uninstall operator
make undeploy

# Remove cert-manager (if installed)
kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

## Lab Summary

In this lab, you:
- Set up certificate management
- Configured webhook service
- Tested webhooks locally
- Deployed webhooks to cluster
- Troubleshot webhook issues

## Key Learnings

1. Webhooks require TLS certificates
2. Kubebuilder simplifies local certificate generation
3. cert-manager handles production certificates
4. Webhook service routes to operator pods
5. Certificates must match CA bundle in webhook config
6. Webhook connectivity issues need troubleshooting

## Solutions

This lab focuses on deployment and certificates. For webhook implementation, refer to:
- [Validating Webhook](../solutions/validating-webhook.go) - From Lab 5.2
- [Mutating Webhook](../solutions/mutating-webhook.go) - From Lab 5.3

## Congratulations!

You've completed Module 5! You now understand:
- Admission control and webhooks
- Validating webhooks for custom validation
- Mutating webhooks for defaulting
- Certificate management and deployment

In Module 6, you'll learn about testing and debugging operators!

**Navigation:** [← Previous Lab: Mutating Webhooks](lab-03-mutating-webhooks.md) | [Related Lesson](../lessons/04-webhook-deployment.md) | [Module Overview](../README.md)

