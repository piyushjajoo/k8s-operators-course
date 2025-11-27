# Module 1 Testing Guide

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

This document provides instructions for testing Module 1 content.

## Prerequisites

Before testing, ensure you have:
- Go 1.21+ installed
- kubectl installed and configured
- Docker or Podman running
- kind installed

## Quick Test

### 1. Test Setup Scripts

```bash
# Test development environment setup
./scripts/setup-dev-environment.sh

# Test kind cluster setup (creates a test cluster)
./scripts/setup-kind-cluster.sh
```

### 2. Test CRD Example

```bash
# Create the test cluster if not exists
./scripts/setup-kind-cluster.sh

# Apply the CRD from lab 04
cat <<EOF | kubectl apply -f -
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: websites.example.com
spec:
  group: example.com
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              url:
                type: string
                pattern: '^https?://'
              replicas:
                type: integer
                minimum: 1
                maximum: 10
            required:
            - url
            - replicas
          status:
            type: object
            properties:
              phase:
                type: string
                enum: [Pending, Running, Failed]
              readyReplicas:
                type: integer
  scope: Namespaced
  names:
    plural: websites
    singular: website
    kind: Website
    shortNames:
    - ws
EOF

# Verify CRD was created
kubectl get crd websites.example.com

# Create a custom resource
cat <<EOF | kubectl apply -f -
apiVersion: example.com/v1
kind: Website
metadata:
  name: test-website
spec:
  url: https://example.com
  replicas: 3
EOF

# Verify resource was created
kubectl get websites
kubectl get website test-website

# Test validation (should fail)
cat <<EOF | kubectl apply -f -
apiVersion: example.com/v1
kind: Website
metadata:
  name: invalid-website
spec:
  url: not-a-url
  replicas: 3
EOF

# Cleanup
kubectl delete website test-website
kubectl delete crd websites.example.com
```

### 3. Test Lab Exercises

Each lab can be run independently. Start with Lab 1.1 and work through sequentially.

```bash
# Lab 1.1: Control Plane
# Follow instructions in module-01/labs/lab-01-control-plane.md

# Lab 1.2: API Machinery
# Follow instructions in module-01/labs/lab-02-api-machinery.md

# Lab 1.3: Controller Pattern
# Follow instructions in module-01/labs/lab-03-controller-pattern.md

# Lab 1.4: Custom Resources
# Follow instructions in module-01/labs/lab-04-custom-resources.md
```

## Verification Checklist

- [ ] Setup scripts run without errors
- [ ] Kind cluster can be created
- [ ] CRD can be created and validated
- [ ] Custom resources can be created
- [ ] Validation rules work correctly
- [ ] All lab exercises complete successfully
- [ ] Mermaid diagrams render correctly (if viewing in markdown viewer)

## Common Issues

### Issue: kind cluster creation fails
**Solution**: Ensure Docker/Podman is running and you have sufficient resources.

### Issue: CRD validation not working
**Solution**: Ensure you're using Kubernetes 1.16+ which supports structural schema validation.

### Issue: kubectl proxy not working
**Solution**: Ensure no other process is using port 8001, or use a different port.

## Cleanup

After testing, clean up resources:

```bash
# Delete kind cluster
kind delete cluster --name k8s-operators-course

# Or use the cleanup script if available
```

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

