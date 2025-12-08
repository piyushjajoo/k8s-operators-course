#!/bin/bash

# Setup script for kind cluster
# Supports both Docker and Podman

set -e

CLUSTER_NAME="${CLUSTER_NAME:-k8s-operators-course}"
KUBERNETES_VERSION="${KUBERNETES_VERSION:-v1.32.0}"

# Detect container runtime
if command -v docker &> /dev/null && docker info &> /dev/null; then
    RUNTIME="docker"
elif command -v podman &> /dev/null; then
    RUNTIME="podman"
else
    echo "Error: Neither Docker nor Podman is available or running"
    exit 1
fi

echo "Using container runtime: $RUNTIME"
echo "Cluster name: $CLUSTER_NAME"
echo "Kubernetes version: $KUBERNETES_VERSION"

# Check if kind is installed
if ! command -v kind &> /dev/null; then
    echo "Error: kind is not installed"
    echo "Install it with: go install sigs.k8s.io/kind@latest"
    exit 1
fi

# Check if cluster already exists
if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    echo "Cluster ${CLUSTER_NAME} already exists"
    read -p "Do you want to delete it and create a new one? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Deleting existing cluster..."
        kind delete cluster --name "$CLUSTER_NAME"
    else
        echo "Using existing cluster"
        kubectl cluster-info --context "kind-${CLUSTER_NAME}"
        exit 0
    fi
fi

# Create kind cluster configuration
cat > /tmp/kind-config.yaml <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: ${CLUSTER_NAME}
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF

echo "Creating kind cluster..."
kind create cluster \
    --name "$CLUSTER_NAME" \
    --config /tmp/kind-config.yaml \
    --image "kindest/node:${KUBERNETES_VERSION}"

# Wait for cluster to be ready
echo "Waiting for cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=300s

# Install ingress-nginx
echo "Installing ingress-nginx..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
kubectl wait --namespace ingress-nginx \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=300s

# Install cert-manager (required for webhooks)
echo "Installing cert-manager..."
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

echo "Waiting for cert-manager to be ready..."
kubectl wait --namespace cert-manager \
    --for=condition=Available deployment/cert-manager \
    --timeout=120s
kubectl wait --namespace cert-manager \
    --for=condition=Available deployment/cert-manager-webhook \
    --timeout=120s
kubectl wait --namespace cert-manager \
    --for=condition=Available deployment/cert-manager-cainjector \
    --timeout=120s

# Set kubectl context
kubectl cluster-info --context "kind-${CLUSTER_NAME}"

echo ""
echo "✅ Cluster setup complete!"
echo ""
echo "Cluster name: ${CLUSTER_NAME}"
echo "Context: kind-${CLUSTER_NAME}"
echo ""
echo "Installed components:"
echo "  - ingress-nginx"
echo "  - cert-manager (for webhook TLS certificates)"
echo ""
echo "To use this cluster:"
echo "  kubectl cluster-info --context kind-${CLUSTER_NAME}"
echo ""
echo "To delete this cluster:"
echo "  kind delete cluster --name ${CLUSTER_NAME}"

# Cleanup temp file
rm -f /tmp/kind-config.yaml

