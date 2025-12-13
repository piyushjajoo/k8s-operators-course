---
layout: default
title: "Lab 02.3: Dev Environment"
nav_order: 13
parent: "Module 2: Introduction to Operators"
grand_parent: Modules
mermaid: true
---

# Lab 2.3: Setting Up Your Environment

**Related Lesson:** [Lesson 2.3: Development Environment Setup](../lessons/03-dev-environment.md)  
**Navigation:** [← Previous Lab: Kubebuilder Fundamentals](lab-02-kubebuilder-fundamentals.md) | [Module Overview](../README.md) | [Next Lab: First Operator →](lab-04-first-operator.md)

## Objectives

- Verify all required tools are installed
- Set up complete development environment
- Create and verify kind cluster
- Test the complete setup

## Prerequisites

- Completion of [Module 1](../module-01/README.md) setup
- Basic understanding of the tools needed

## Exercise 1: Verify Prerequisites

### Task 1.1: Check Go Installation

```bash
# Check Go version (need 1.21+)
go version

# Verify Go is working
go env

# Check GOPATH and GOROOT
echo $GOPATH
echo $GOROOT
```

**Expected:** Go 1.21 or higher

### Task 1.2: Check kubectl

```bash
# Check kubectl version
kubectl version --client

# Verify kubectl is working
kubectl cluster-info
```

**Note:** If no cluster is configured, that's okay - we'll create one.

### Task 1.3: Check Docker/Podman

```bash
# Check Docker
docker --version
docker info

# OR check Podman
podman --version
podman info
```

**Expected:** Docker or Podman running

## Exercise 2: Install Missing Tools

### Task 2.1: Use Setup Script

```bash
# Run the setup script
./scripts/setup-dev-environment.sh
```

**Observe:**
- Which tools are already installed?
- Which tools need installation?
- What gets installed automatically?

### Task 2.2: Manual Verification

After running the script, verify each tool:

```bash
# Go
go version

# kubectl
kubectl version --client

# kubebuilder
kubebuilder version

# kind
kind version

# Docker/Podman
docker --version  # or podman --version
```

## Exercise 3: Set Up Kind Cluster

### Task 3.1: Use Setup Script

```bash
# Run kind cluster setup
./scripts/setup-kind-cluster.sh
```

**Observe:**
- Cluster creation process
- What gets installed?
- How long does it take?

### Task 3.2: Verify Cluster

```bash
# Check cluster info
kubectl cluster-info

# List nodes
kubectl get nodes

# Check cluster context
kubectl config current-context

# Should show: kind-k8s-operators-course
```

### Task 3.3: Test Cluster

```bash
# Create a test pod
kubectl run test-pod --image=nginx:latest

# Wait for it to be ready
kubectl wait --for=condition=ready pod/test-pod --timeout=60s

# Verify it's running
kubectl get pods

# Clean up
kubectl delete pod test-pod
```

## Exercise 4: Verify kubebuilder

### Task 4.1: Check Installation

```bash
# Check kubebuilder version
kubebuilder version

# Verify it's in PATH
which kubebuilder

# Test kubebuilder commands
kubebuilder --help
```

### Task 4.2: Test kubebuilder Init

```bash
# Create a test directory
mkdir -p /tmp/env-test
cd /tmp/env-test

# Test kubebuilder init
kubebuilder init --domain test.com --repo github.com/test/env-test

# Verify project was created
ls -la

# Check main.go exists
test -f main.go && echo "✅ main.go exists" || echo "❌ main.go missing"

# Clean up
cd ~
rm -rf /tmp/env-test
```

## Exercise 5: Complete Environment Checklist

### Task 5.1: Run Verification Checklist

Check each item:

```bash
# Go 1.21+
go version | grep -q "go1.2[1-9]\|go1.[3-9]" && echo "✅ Go version OK" || echo "❌ Go version too old"

# kubectl
kubectl version --client > /dev/null 2>&1 && echo "✅ kubectl OK" || echo "❌ kubectl missing"

# kubebuilder
kubebuilder version > /dev/null 2>&1 && echo "✅ kubebuilder OK" || echo "❌ kubebuilder missing"

# kind
kind version > /dev/null 2>&1 && echo "✅ kind OK" || echo "❌ kind missing"

# Docker/Podman
(docker info > /dev/null 2>&1 || podman info > /dev/null 2>&1) && echo "✅ Container runtime OK" || echo "❌ Container runtime missing"

# Kind cluster
kubectl cluster-info --context kind-k8s-operators-course > /dev/null 2>&1 && echo "✅ Kind cluster OK" || echo "❌ Kind cluster missing"
```

### Task 5.2: Fix Any Issues

If any checks fail:
- Review error messages
- Re-run setup scripts
- Check documentation
- Ask for help if needed

## Exercise 6: Test Development Workflow

### Task 6.1: Create Test Project

```bash
# Create test project
mkdir -p /tmp/workflow-test
cd /tmp/workflow-test

# Initialize project
kubebuilder init --domain test.com --repo github.com/test/workflow-test
```

### Task 6.2: Generate and Install

```bash
# Generate code
make generate

# Generate manifests
make manifests

# Install CRD (this will fail if no cluster, that's OK for testing)
make install || echo "No cluster, skipping install"
```

### Task 6.3: Verify Workflow

```bash
# Check generated files
ls -la config/crd/bases/
ls -la config/rbac/

# Check Makefile targets
make help
```

### Task 6.4: Cleanup

```bash
cd ~
rm -rf /tmp/workflow-test
```

## Exercise 7: IDE Setup (Optional)

### Task 7.1: VS Code Setup

If using VS Code:

```bash
# Install Go extension
code --install-extension golang.go

# Install Kubernetes extension
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
```

### Task 7.2: GoLand Setup

If using GoLand:
- GoLand has built-in Go and Kubernetes support
- No additional setup needed

## Environment Verification Summary

Your environment should have:

- ✅ Go 1.21+
- ✅ kubectl
- ✅ kubebuilder
- ✅ kind
- ✅ Docker or Podman
- ✅ Kind cluster running
- ✅ kubectl context set to kind cluster

## Troubleshooting

### Issue: kubebuilder not found
```bash
# Add to PATH
export PATH=$PATH:/usr/local/bin
# Or reinstall
```

### Issue: kind cluster not accessible
```bash
# Recreate cluster
kind delete cluster --name k8s-operators-course
./scripts/setup-kind-cluster.sh
```

### Issue: Go module errors
```bash
# Enable Go modules
export GO111MODULE=on
```

## Lab Summary

In this lab, you:
- Verified all required tools
- Set up complete development environment
- Created and verified kind cluster
- Tested the development workflow
- Verified everything works together

## Key Learnings

1. Complete environment includes: Go, kubebuilder, kubectl, kind, Docker/Podman
2. Setup scripts automate installation
3. Kind cluster provides local Kubernetes
4. All tools must work together
5. Verification is important before starting development

## Next Steps

Your environment is ready! Now let's build your first operator.

**Navigation:** [← Previous Lab: Kubebuilder Fundamentals](lab-02-kubebuilder-fundamentals.md) | [Related Lesson](../lessons/03-dev-environment.md) | [Next Lab: First Operator →](lab-04-first-operator.md)
