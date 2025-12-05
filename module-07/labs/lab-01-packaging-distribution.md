# Lab 7.1: Packaging Your Operator

**Related Lesson:** [Lesson 7.1: Packaging and Distribution](../lessons/01-packaging-distribution.md)  
**Navigation:** [Module Overview](../README.md) | [Next Lab: RBAC →](lab-02-rbac-security.md)

## Objectives

- Build container image for operator
- Create Helm chart for deployment
- Tag and version images properly
- Push to container registry

## Prerequisites

- Completion of [Module 6](../module-06/README.md)
- Database operator ready
- Docker or Podman installed
- Access to container registry (or use kind for local)

## Exercise 1: Build Container Image

### Task 1.1: Create Dockerfile

Create `Dockerfile` in your operator root:

```dockerfile
# Build stage
FROM golang:1.21 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# Cache deps before building and copying source
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Runtime stage
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
```

### Task 1.2: Build Image

```bash
# Build image
docker build -t database-operator:latest .

# For kind, load image
kind load docker-image database-operator:latest --name k8s-operators-course

# Tag with version
docker tag database-operator:latest database-operator:v0.1.0
```

## Exercise 2: Create Helm Chart

### Task 2.1: Initialize Helm Chart

```bash
# Create Helm chart
helm create database-operator

# Remove default templates (we'll create our own)
rm -rf database-operator/templates/*
```

### Task 2.2: Create Chart.yaml

Edit `database-operator/Chart.yaml`:

```yaml
apiVersion: v2
name: database-operator
description: A Helm chart for Database Operator
type: application
version: 0.1.0
appVersion: "0.1.0"
```

### Task 2.3: Create values.yaml

Edit `database-operator/values.yaml`:

```yaml
image:
  repository: database-operator
  pullPolicy: IfNotPresent
  tag: "latest"

replicaCount: 1

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi

leaderElection:
  enabled: true
  resourceName: database-operator-leader-election
```

### Task 2.4: Create Deployment Template

Create `database-operator/templates/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "database-operator.fullname" . }}
  labels:
    {{- include "database-operator.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "database-operator.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "database-operator.selectorLabels" . | nindent 8 }}
    spec:
      serviceAccountName: {{ include "database-operator.serviceAccountName" . }}
      containers:
      - name: manager
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        resources:
          {{- toYaml .Values.resources | nindent 10 }}
        args:
        - --leader-elect={{ .Values.leaderElection.enabled }}
```

## Exercise 3: Package and Test

### Task 3.1: Package Helm Chart

```bash
# Package chart
helm package database-operator

# Verify package
helm show chart database-operator-0.1.0.tgz
```

### Task 3.2: Install with Helm

```bash
# Install operator
helm install database-operator ./database-operator

# Verify deployment
kubectl get deployment database-operator

# Check logs
kubectl logs -l app=database-operator
```

## Exercise 4: Version and Tag

### Task 4.1: Tag Image with Version

```bash
# Tag with semantic version
docker tag database-operator:latest database-operator:v0.1.0
docker tag database-operator:latest database-operator:v0.1.0-amd64

# List tags
docker images database-operator
```

### Task 4.2: Update Chart Version

```bash
# Update Chart.yaml version
# Update values.yaml image tag
# Package new version
helm package database-operator
```

## Cleanup

```bash
# Uninstall Helm release
helm uninstall database-operator

# Remove image (optional)
docker rmi database-operator:latest
```

## Lab Summary

In this lab, you:
- Created Dockerfile for operator
- Built container image
- Created Helm chart
- Packaged and deployed with Helm
- Tagged images with versions

## Key Learnings

1. Multi-stage builds create smaller images
2. Distroless images improve security
3. Helm charts simplify deployment
4. Semantic versioning tracks changes
5. Proper tagging enables rollbacks
6. Container images enable distribution

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Dockerfile](../solutions/Dockerfile) - Production-ready multi-stage Dockerfile
- [Helm Chart](../solutions/helm-chart/) - Complete Helm chart (Chart.yaml, values.yaml, templates)

## Next Steps

Now let's configure proper RBAC and security!

**Navigation:** [← Module Overview](../README.md) | [Related Lesson](../lessons/01-packaging-distribution.md) | [Next Lab: RBAC →](lab-02-rbac-security.md)

