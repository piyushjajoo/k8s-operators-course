# GitHub Actions Solutions for Module 7

This directory contains GitHub Actions workflows for automating operator releases.

## Workflows

### `release.yaml` - Main Release Workflow

Triggers on version tags (e.g., `v0.1.0`) and:

1. **Builds and pushes Docker image** to GitHub Container Registry (GHCR)
   - Multi-architecture support (amd64, arm64)
   - Semantic version tags
   - SHA-based tags for traceability

2. **Generates and publishes Helm chart** to GHCR OCI registry
   - Uses `make helm-chart` target
   - Pushes to `oci://ghcr.io/<owner>/charts`
   - Attaches chart as release artifact

3. **Creates installer manifest**
   - Single-file install via `kubectl apply -f`
   - Attached to GitHub release

### `helm-gh-pages.yaml` - GitHub Pages Helm Repository

Alternative approach using GitHub Pages:

- Triggers on changes to `charts/` directory
- Uses helm/chart-releaser-action
- Publishes to GitHub Pages-based Helm repository

## Usage

### Option 1: OCI Registry (Recommended)

Copy `release.yaml` to `.github/workflows/release.yaml`:

```bash
mkdir -p .github/workflows
cp release.yaml .github/workflows/
```

Install charts from OCI:

```bash
helm install my-operator oci://ghcr.io/YOUR_USERNAME/charts/postgres-operator --version 0.1.0
```

### Option 2: GitHub Pages Helm Repository

Copy `helm-gh-pages.yaml` to `.github/workflows/`:

```bash
cp helm-gh-pages.yaml .github/workflows/
```

Add Helm repository:

```bash
helm repo add postgres-operator https://YOUR_USERNAME.github.io/postgres-operator
helm repo update
helm install my-operator postgres-operator/postgres-operator
```

## Required Makefile Targets

Ensure your Makefile has these targets:

```makefile
CHART_NAME ?= postgres-operator
CHART_VERSION ?= 0.1.0
CHART_DIR ?= charts/$(CHART_NAME)

.PHONY: helm-chart
helm-chart: manifests kustomize
	@mkdir -p $(CHART_DIR)/templates
	@cat > $(CHART_DIR)/Chart.yaml <<EOF
apiVersion: v2
name: $(CHART_NAME)
description: A Helm chart for $(CHART_NAME)
type: application
version: $(CHART_VERSION)
appVersion: "$(VERSION)"
EOF
	@cat > $(CHART_DIR)/values.yaml <<EOF
image:
  repository: $(IMAGE_TAG_BASE)
  tag: $(VERSION)
  pullPolicy: IfNotPresent
replicaCount: 1
EOF
	@cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	@$(KUSTOMIZE) build config/default > $(CHART_DIR)/templates/manifests.yaml

.PHONY: helm-package
helm-package: helm-chart
	@mkdir -p dist
	helm package $(CHART_DIR) -d dist/

.PHONY: helm-lint
helm-lint: helm-chart
	helm lint $(CHART_DIR)
```

## Creating a Release

```bash
# Ensure you're on main branch
git checkout main

# Tag the release
git tag v0.1.0

# Push tag to trigger workflow
git push origin v0.1.0
```

## Permissions

The workflows require these repository permissions:

- `contents: write` - Create releases, push to gh-pages
- `packages: write` - Push to GHCR

These are configured in the workflow files using `permissions:` blocks.

## Secrets

No additional secrets are required when using `GITHUB_TOKEN`. The workflows use the automatic `secrets.GITHUB_TOKEN` which has appropriate permissions for GHCR.

