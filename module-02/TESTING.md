# Module 2 Testing Guide

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

This document provides instructions for testing Module 2 content.

## Prerequisites

Before testing, ensure you have:
- Completed [Module 1](../module-01/README.md)
- Go 1.21+ installed
- kubectl installed and configured
- Docker or Podman running
- kind installed

## Quick Test

### 1. Test Setup Scripts

```bash
# Test development environment setup
./scripts/setup-dev-environment.sh

# Test kind cluster setup
./scripts/setup-kind-cluster.sh
```

### 2. Test kubebuilder Installation

```bash
# Verify kubebuilder is installed
kubebuilder version

# Test kubebuilder init
mkdir -p /tmp/test-kb
cd /tmp/test-kb
kubebuilder init --domain test.com --repo github.com/test/test-operator
cd ~
rm -rf /tmp/test-kb
```

### 3. Test Hello World Operator

Follow the complete lab in [Lab 2.4](labs/lab-04-first-operator.md) to build and test the operator.

## Verification Checklist

- [ ] Setup scripts run without errors
- [ ] kubebuilder is installed and working
- [ ] Kind cluster can be created
- [ ] Can initialize kubebuilder project
- [ ] Can create API with kubebuilder
- [ ] Can generate code and manifests
- [ ] Can build and run Hello World operator
- [ ] Operator reconciles Custom Resources
- [ ] All lab exercises complete successfully
- [ ] Mermaid diagrams render correctly

## Common Issues

### Issue: kubebuilder not found
**Solution**: Ensure kubebuilder is in PATH or reinstall.

### Issue: Go module errors
**Solution**: Ensure `GO111MODULE=on` is set.

### Issue: kind cluster not accessible
**Solution**: Recreate cluster using setup script.

### Issue: Operator doesn't reconcile
**Solution**: 
- Check CRD is installed: `kubectl get crd`
- Check operator logs
- Verify RBAC permissions

## Testing the Hello World Operator

### Complete Test Flow

1. **Initialize project:**
   ```bash
   kubebuilder init --domain example.com --repo github.com/example/hello-world-operator
   ```

2. **Create API:**
   ```bash
   kubebuilder create api --group hello --version v1 --kind HelloWorld
   ```

3. **Implement types and controller** (see [Lesson 2.4](lessons/04-first-operator.md))

4. **Generate and install:**
   ```bash
   make generate manifests install
   ```

5. **Run operator:**
   ```bash
   make run
   ```

6. **Create Custom Resource:**
   ```bash
   kubectl apply -f - <<EOF
   apiVersion: hello.example.com/v1
   kind: HelloWorld
   metadata:
     name: test-hello
   spec:
     message: "Test message"
     count: 3
   EOF
   ```

7. **Verify reconciliation:**
   ```bash
   kubectl get helloworld test-hello
   kubectl get configmap test-hello-config
   ```

8. **Cleanup:**
   ```bash
   kubectl delete helloworld test-hello
   make uninstall
   ```

## Cleanup

After testing, clean up resources:

```bash
# Delete kind cluster (optional)
kind delete cluster --name k8s-operators-course

# Or keep it for Module 3
```

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

