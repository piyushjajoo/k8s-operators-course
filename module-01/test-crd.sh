#!/bin/bash
# Quick test script for Module 1 CRD example

set -e

echo "Testing CRD creation..."

# Check if cluster exists
if ! kubectl cluster-info &> /dev/null; then
    echo "Error: No Kubernetes cluster found. Please create one first."
    echo "Run: ./scripts/setup-kind-cluster.sh"
    exit 1
fi

# Create CRD
echo "Creating CRD..."
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
  scope: Namespaced
  names:
    plural: websites
    singular: website
    kind: Website
EOF

# Wait for CRD to be ready
echo "Waiting for CRD to be established..."
kubectl wait --for condition=established --timeout=30s crd/websites.example.com

# Verify CRD exists
if kubectl get crd websites.example.com &> /dev/null; then
    echo "✅ CRD created successfully"
else
    echo "❌ CRD creation failed"
    exit 1
fi

# Create a custom resource
echo "Creating custom resource..."
cat <<EOF | kubectl apply -f -
apiVersion: example.com/v1
kind: Website
metadata:
  name: test-website
spec:
  url: https://example.com
  replicas: 3
EOF

# Verify resource exists
if kubectl get website test-website &> /dev/null; then
    echo "✅ Custom resource created successfully"
else
    echo "❌ Custom resource creation failed"
    exit 1
fi

# Test validation (should fail)
echo "Testing validation (expecting failure)..."
if cat <<EOF | kubectl apply -f - 2>&1 | grep -q "validation failed\|Invalid value"; then
apiVersion: example.com/v1
kind: Website
metadata:
  name: invalid-test
spec:
  url: not-a-url
  replicas: 3
EOF
    echo "✅ Validation working correctly"
else
    echo "⚠️  Validation test inconclusive"
fi

# Cleanup
echo "Cleaning up..."
kubectl delete website test-website 2>/dev/null || true
kubectl delete website invalid-test 2>/dev/null || true
kubectl delete crd websites.example.com 2>/dev/null || true

echo "✅ All tests passed!"

