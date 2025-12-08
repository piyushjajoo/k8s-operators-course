#!/bin/bash

# Development environment setup script
# Installs required tools for the course

set -e

echo "🚀 Setting up development environment for Kubernetes Operators Course"
echo ""

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed"
    echo "Please install Go 1.24+ from https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go version: $GO_VERSION"

# Check kubectl installation
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed"
    echo "Installing kubectl..."
    # macOS
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install kubectl
    else
        echo "Please install kubectl manually: https://kubernetes.io/docs/tasks/tools/"
        exit 1
    fi
fi

KUBECTL_VERSION=$(kubectl version --client 2>/dev/null | grep "Client Version" | awk '{print $3}' || kubectl version --client 2>/dev/null | head -1)
echo "✅ kubectl version: $KUBECTL_VERSION"

# Check Docker/Podman
if command -v docker &> /dev/null && docker info &> /dev/null; then
    DOCKER_VERSION=$(docker --version)
    echo "✅ Docker: $DOCKER_VERSION"
elif command -v podman &> /dev/null; then
    PODMAN_VERSION=$(podman --version)
    echo "✅ Podman: $PODMAN_VERSION"
else
    echo "❌ Neither Docker nor Podman is available or running"
    echo "Please install and start Docker or Podman"
    exit 1
fi

# Check kind installation
if ! command -v kind &> /dev/null; then
    echo "⚠️  kind is not installed"
    echo "Installing kind..."
    go install sigs.k8s.io/kind@latest
    if [ -d "$HOME/go/bin" ] && [[ ":$PATH:" != *":$HOME/go/bin:"* ]]; then
        export PATH="$HOME/go/bin:$PATH"
    fi
fi

KIND_VERSION=$(kind --version 2>/dev/null || echo "unknown")
echo "✅ kind version: $KIND_VERSION"

# Check kubebuilder installation
if ! command -v kubebuilder &> /dev/null; then
    echo "⚠️  kubebuilder is not installed"
    echo "Installing kubebuilder..."
    
    # Detect OS
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    if [ "$ARCH" = "x86_64" ]; then
        ARCH="amd64"
    fi
    
    KUBEBUILDER_VERSION="4.7.1"
    KUBEBUILDER_URL="https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/kubebuilder_${OS}_${ARCH}"
    
    curl -L "$KUBEBUILDER_URL" -o /tmp/kubebuilder
    chmod +x /tmp/kubebuilder
    sudo mv /tmp/kubebuilder /usr/local/bin/kubebuilder
    
    # Install kustomize (required by kubebuilder)
    if ! command -v kustomize &> /dev/null; then
        go install sigs.k8s.io/kustomize/kustomize/v5@latest
    fi
fi

KUBEBUILDER_VERSION=$(kubebuilder version 2>/dev/null | head -n1 || echo "unknown")
echo "✅ kubebuilder version: $KUBEBUILDER_VERSION"

# Verify controller-gen
if ! command -v controller-gen &> /dev/null; then
    echo "⚠️  controller-gen is not installed"
    echo "Installing controller-gen..."
    go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
fi

echo ""
echo "✅ Development environment setup complete!"
echo ""
echo "Next steps:"
echo "  1. Run: ./scripts/setup-kind-cluster.sh"
echo "  2. Start with Module 1 lessons"
echo ""

