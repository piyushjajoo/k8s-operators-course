# Lab 7.1: Packaging Your Operator

**Related Lesson:** [Lesson 7.1: Packaging and Distribution](../lessons/01-packaging-distribution.md)  
**Navigation:** [Module Overview](../README.md) | [Next Lab: RBAC →](lab-02-rbac-security.md)

## Objectives

- Build container image for operator
- Create Helm chart for deployment
- Tag and version images properly
- Push to container registry

## Prerequisites

- Completion of [Module 6](../../module-06/README.md)
- Database operator ready
- Docker or Podman installed
- Access to container registry (or use kind for local)

## Exercise 1: Build Container Image

Kubebuilder already generated a production-ready Dockerfile when you scaffolded your project. Let's explore and use it.

### Task 1.1: Review the Kubebuilder-Generated Dockerfile

Kubebuilder creates a `Dockerfile` in your project root. Review it:

```bash
# Navigate to your operator project from module 3
cd ~/postgres-operator

# View the Dockerfile
cat Dockerfile
```

The generated Dockerfile should look like:

```dockerfile
# Build stage
FROM golang:1.24 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -a -o manager cmd/main.go

# Runtime stage
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
```

**Important:** The entire `internal/` directory is copied, which includes:
- `internal/controller/` - Your controller reconciliation logic
- `internal/webhook/` - Webhook handlers (created in Module 5)

### Task 1.2: Build Image Using Makefile

Kubebuilder provides Makefile targets for building images:

```bash
# Build the image using kubebuilder's make target
make docker-build IMG=postgres-operator:v0.1.0

# For kind, load image into the cluster
kind load docker-image postgres-operator:v0.1.0 --name k8s-operators-course

# Verify image is available in kind
docker exec -it k8s-operators-course-control-plane crictl images | grep postgres-operator
```

## Exercise 2: Deploy Using Kubebuilder's Kustomize (Recommended)

Kubebuilder uses Kustomize for deployment by default. This is the recommended approach.

### Task 2.1: Review Kustomize Configuration

```bash
# Explore the config directory structure
ls -la config/

# Key directories:
# config/crd/       - CRD definitions
# config/default/   - Main kustomization
# config/manager/   - Controller deployment
# config/rbac/      - RBAC rules
```

### Task 2.2: Deploy with Kustomize

```bash
# Install CRDs
make install

# Deploy the operator (builds and deploys)
make deploy IMG=postgres-operator:v0.1.0

# Verify deployment
kubectl get deployment -n postgres-operator-system
kubectl get pods -n postgres-operator-system
```

### Task 2.3: View Generated Manifests

```bash
# Preview what will be deployed
kustomize build config/default

# Or using make target
make build-installer IMG=postgres-operator:v0.1.0
```

## Exercise 3: Create Helm Chart from Kustomize

For wider distribution, you can generate a Helm chart from your Kustomize manifests. The chart must include **all operator components**:

- **CRDs** - Custom Resource Definitions (from `config/crd/`)
- **RBAC** - ServiceAccount, ClusterRole, ClusterRoleBinding (from `config/rbac/`)
- **Deployment** - Controller manager (from `config/manager/`)
- **Webhooks** - If created in Module 5 (from `config/webhook/`)

**Important:** A Helm chart with only the Deployment won't work! The operator needs all these components to function.

### Task 3.1: Add Helm Chart Make Target

Add these targets to your `Makefile`:

```makefile
# Helm chart configuration
CHART_NAME ?= postgres-operator
CHART_VERSION ?= 0.1.0
CHART_DIR ?= charts/$(CHART_NAME)

##@ Helm

.PHONY: helm-chart
helm-chart: manifests kustomize ## Generate Helm chart from Kustomize (includes CRDs, RBAC, Deployment, Webhooks)
	@echo "Generating Helm chart with ALL operator components..."
	@mkdir -p $(CHART_DIR)/templates
	@# Create Chart.yaml
	@printf '%s\n' \
		'apiVersion: v2' \
		'name: $(CHART_NAME)' \
		'description: A Helm chart for $(CHART_NAME) - includes CRDs, RBAC, and webhooks' \
		'type: application' \
		'version: $(CHART_VERSION)' \
		'appVersion: "$(VERSION)"' \
		> $(CHART_DIR)/Chart.yaml
	@# Create values.yaml
	@printf '%s\n' \
		'image:' \
		'  repository: $(IMAGE_TAG_BASE)' \
		'  tag: $(VERSION)' \
		'  pullPolicy: IfNotPresent' \
		'' \
		'replicaCount: 1' \
		'' \
		'resources:' \
		'  limits:' \
		'    cpu: 500m' \
		'    memory: 128Mi' \
		'  requests:' \
		'    cpu: 10m' \
		'    memory: 64Mi' \
		'' \
		'leaderElection:' \
		'  enabled: false' \
		'' \
		'namespace: $(CHART_NAME)-system' \
		> $(CHART_DIR)/values.yaml
	@# Generate ALL manifests from kustomize (CRDs, RBAC, Deployment, Webhooks)
	@cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	@$(KUSTOMIZE) build config/default > $(CHART_DIR)/templates/manifests.yaml
	@echo "Helm chart generated at $(CHART_DIR)"
	@echo "Contents include: CRDs, RBAC (ServiceAccount, ClusterRole, ClusterRoleBinding), Deployment, Webhooks"

.PHONY: helm-package
helm-package: helm-chart ## Package Helm chart
	@mkdir -p dist
	helm package $(CHART_DIR) -d dist/

.PHONY: helm-lint
helm-lint: helm-chart ## Lint Helm chart
	helm lint $(CHART_DIR)

.PHONY: helm-template
helm-template: helm-chart ## Render Helm templates locally
	helm template $(CHART_NAME) $(CHART_DIR)

.PHONY: helm-install
helm-install: helm-chart ## Install Helm chart to cluster
	helm upgrade --install $(CHART_NAME) $(CHART_DIR) \
		--namespace $(CHART_NAME)-system \
		--create-namespace

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall Helm chart
	helm uninstall $(CHART_NAME) --namespace $(CHART_NAME)-system
```

### Task 3.2: Generate and Verify the Helm Chart

```bash
# Generate Helm chart from Kustomize
make helm-chart IMG=postgres-operator:v0.1.0

# Verify the chart structure
ls -la charts/postgres-operator/
ls -la charts/postgres-operator/templates/

# IMPORTANT: Verify the generated manifests include all components
echo "=== Checking for CRDs ==="
grep -c "kind: CustomResourceDefinition" charts/postgres-operator/templates/manifests.yaml

echo "=== Checking for RBAC ==="
grep -c "kind: ServiceAccount" charts/postgres-operator/templates/manifests.yaml
grep -c "kind: ClusterRole" charts/postgres-operator/templates/manifests.yaml
grep -c "kind: ClusterRoleBinding" charts/postgres-operator/templates/manifests.yaml

echo "=== Checking for Deployment ==="
grep -c "kind: Deployment" charts/postgres-operator/templates/manifests.yaml

# Lint the chart
make helm-lint

# Preview ALL rendered templates
make helm-template | head -100
```

### Task 3.3: Package and Test the Chart

```bash
# Package the chart
make helm-package

# List packaged charts
ls -la dist/

# Test install (optional)
make helm-install

# Verify deployment
kubectl get pods -n postgres-operator-system

# Cleanup
make helm-uninstall
```

## Exercise 4: GitHub Actions for CI/CD

Automate chart publishing with GitHub Actions.

### Task 4.1: Create GitHub Actions Workflow

Create `.github/workflows/release.yaml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract version
        id: version
        run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT

      - name: Build and push Docker image
        run: |
          make docker-build docker-push IMG=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ steps.version.outputs.VERSION }}

  helm-release:
    runs-on: ubuntu-latest
    needs: build-and-push
    permissions:
      contents: write
      packages: write

    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Configure Git
        run: |
          git config user.name "$GITHUB_ACTOR"
          git config user.email "$GITHUB_ACTOR@users.noreply.github.com"

      - name: Install Helm
        uses: azure/setup-helm@v3
        with:
          version: v3.12.0

      - name: Extract version
        id: version
        run: echo "VERSION=${GITHUB_REF#refs/tags/v}" >> $GITHUB_OUTPUT

      - name: Generate Helm chart
        run: |
          make helm-chart \
            IMG=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.ref_name }} \
            CHART_VERSION=${{ steps.version.outputs.VERSION }} \
            VERSION=${{ steps.version.outputs.VERSION }}

      - name: Package Helm chart
        run: make helm-package

      - name: Push Helm chart to GHCR
        run: |
          helm push dist/*.tgz oci://${{ env.REGISTRY }}/${{ github.repository_owner }}/charts

      - name: Upload chart as release artifact
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*.tgz
```

### Task 4.2: Create Helm Chart Repository (Alternative)

For a traditional Helm repository using GitHub Pages, create `.github/workflows/helm-release.yaml`:

```yaml
name: Helm Chart Release

on:
  push:
    branches:
      - main
    paths:
      - 'charts/**'
      - '.github/workflows/helm-release.yaml'

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write

    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Configure Git
        run: |
          git config user.name "$GITHUB_ACTOR"
          git config user.email "$GITHUB_ACTOR@users.noreply.github.com"

      - name: Install Helm
        uses: azure/setup-helm@v3

      - name: Run chart-releaser
        uses: helm/chart-releaser-action@v1.6.0
        with:
          charts_dir: charts
        env:
          CR_TOKEN: "${{ secrets.GITHUB_TOKEN }}"
```

### Task 4.3: Add Repository Documentation

Create `charts/README.md`:

```markdown
# Database Operator Helm Chart

## Installation

### Using OCI Registry (GHCR)

```bash
helm install postgres-operator oci://ghcr.io/YOUR_USERNAME/charts/postgres-operator --version 0.1.0
```

### Using Helm Repository

```bash
# Add the repository
helm repo add postgres-operator https://YOUR_USERNAME.github.io/postgres-operator

# Update repositories
helm repo update

# Install
helm install postgres-operator postgres-operator/postgres-operator
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Image repository | `ghcr.io/YOUR_USERNAME/postgres-operator` |
| `image.tag` | Image tag | `v0.1.0` |
| `replicaCount` | Number of replicas | `1` |
| `leaderElection.enabled` | Enable leader election | `false` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `128Mi` |


### Task 4.4: Test the Workflow Locally (Optional)

```bash
# Create a test tag
git tag v0.1.0
git push origin v0.1.0

# Watch the Actions tab in GitHub for workflow execution
```

## Exercise 5: Version and Tag

### Task 5.1: Version Your Operator

Update the version in `Makefile`:

```makefile
# Image URL to use all building/pushing image targets
IMG ?= postgres-operator:v0.1.0
VERSION ?= 0.1.0
```

Or specify at build time:

```bash
# Build with specific version
make docker-build IMG=postgres-operator:v0.1.0

# Build for multiple architectures (if needed)
make docker-buildx IMG=postgres-operator:v0.1.0
```

### Task 5.2: Push to Registry

```bash
# Tag for your registry
docker tag postgres-operator:v0.1.0 ghcr.io/your-username/postgres-operator:v0.1.0

# Push (requires authentication)
docker push ghcr.io/your-username/postgres-operator:v0.1.0

# Or use make target
make docker-push IMG=ghcr.io/your-username/postgres-operator:v0.1.0
```

### Task 5.3: Deploy Specific Version

```bash
# Deploy with Kustomize
make deploy IMG=ghcr.io/your-username/postgres-operator:v0.1.0

# Or deploy with Helm
make helm-install IMG=ghcr.io/your-username/postgres-operator:v0.1.0

# Verify
kubectl get deployment -n postgres-operator-system -o yaml | grep image:
```

## Cleanup

```bash
# Undeploy the operator (Kustomize)
make undeploy

# Or undeploy with Helm
make helm-uninstall

# Uninstall CRDs
make uninstall

# Remove local images (optional)
docker rmi postgres-operator:v0.1.0

# Clean up generated charts
rm -rf charts/ dist/
```

## Lab Summary

In this lab, you:
- Reviewed kubebuilder's generated Dockerfile
- Built container images using `make docker-build`
- Deployed using kubebuilder's Kustomize configuration
- Created a make target to generate Helm charts from Kustomize
- Set up GitHub Actions for automated releases
- Tagged and versioned images properly

## Key Learnings

1. Kubebuilder generates a production-ready Dockerfile
2. Use `make docker-build` and `make docker-push` for images
3. Use `make deploy` for Kustomize-based deployment
4. `make helm-chart` generates Helm charts from Kustomize manifests
5. GitHub Actions automate image and chart publishing
6. OCI registries (like GHCR) can host both images AND Helm charts
7. Semantic versioning tracks operator releases

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Dockerfile](../solutions/Dockerfile) - Production-ready multi-stage Dockerfile
- [Helm Chart](../solutions/helm-chart/) - Complete Helm chart (Chart.yaml, values.yaml, templates)
- [GitHub Actions](../solutions/github-actions/) - CI/CD workflows for releases

## Next Steps

Now let's configure proper RBAC and security!

**Navigation:** [← Module Overview](../README.md) | [Related Lesson](../lessons/01-packaging-distribution.md) | [Next Lab: RBAC →](lab-02-rbac-security.md)

