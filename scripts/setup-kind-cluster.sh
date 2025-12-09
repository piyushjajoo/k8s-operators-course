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

# Install Prometheus Stack (for metrics and observability in Module 6-7)
echo "Installing Prometheus stack..."

# Check if Helm is installed
if ! command -v helm &> /dev/null; then
    echo "Warning: Helm is not installed. Skipping Prometheus installation."
    echo "Install Helm and run the prometheus install command from scripts/setup-kind-cluster.sh"
    PROMETHEUS_INSTALLED="no"
else
    # Add prometheus-community Helm repo
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo update

    # Install kube-prometheus-stack with settings optimized for the course:
    # - Minimal resources for kind
    # - ServiceMonitor discovery from ALL namespaces (not just release: prometheus labeled)
    # - PodMonitor discovery from ALL namespaces
    # 
    # Key settings for operator development:
    # - serviceMonitorSelectorNilUsesHelmValues=false: Empty selector means "select all"
    #   (instead of defaulting to only 'release: prometheus' labeled ServiceMonitors)
    # - podMonitorSelectorNilUsesHelmValues=false: Same for PodMonitors
    helm install prometheus prometheus-community/kube-prometheus-stack \
        --namespace monitoring \
        --create-namespace \
        --set prometheus.prometheusSpec.resources.requests.memory=256Mi \
        --set prometheus.prometheusSpec.resources.requests.cpu=100m \
        --set prometheus.prometheusSpec.resources.limits.memory=512Mi \
        --set prometheus.prometheusSpec.resources.limits.cpu=500m \
        --set grafana.resources.requests.memory=128Mi \
        --set grafana.resources.requests.cpu=50m \
        --set grafana.resources.limits.memory=256Mi \
        --set grafana.resources.limits.cpu=200m \
        --set alertmanager.enabled=false \
        --set nodeExporter.enabled=false \
        --set kubeStateMetrics.enabled=true \
        --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
        --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false \
        --wait \
        --timeout 300s

    # Label the monitoring namespace for network policy access
    kubectl label namespace monitoring metrics=enabled --overwrite
    
    PROMETHEUS_INSTALLED="yes"
    
    echo "Prometheus stack installed!"
    echo "  - Configured to discover ServiceMonitors from ALL namespaces"
    echo "  - Prometheus UI: kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090"
    echo "  - Grafana UI: kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80"
    echo "  - Grafana credentials: admin / prom-operator"
fi

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
if [ "$PROMETHEUS_INSTALLED" = "yes" ]; then
echo "  - Prometheus stack (for metrics - Module 6-7)"
echo "    - monitoring namespace labeled with: metrics=enabled"
fi
echo ""
echo "To use this cluster:"
echo "  kubectl cluster-info --context kind-${CLUSTER_NAME}"
echo ""
if [ "$PROMETHEUS_INSTALLED" = "yes" ]; then
echo "To access Prometheus:"
echo "  kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090"
echo ""
echo "To access Grafana (admin/prom-operator):"
echo "  kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80"
echo ""
fi
echo "To delete this cluster:"
echo "  kind delete cluster --name ${CLUSTER_NAME}"

# Cleanup temp file
rm -f /tmp/kind-config.yaml

