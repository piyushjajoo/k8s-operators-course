# Module 5 Solutions

This directory contains complete, working solutions for Module 5 labs.

## Files

- [**validating-webhook.go**](https://github.com/piyushjajoo/k8s-operators-course/blob/main/module-05/solutions/validating-webhook.go): Complete validating webhook implementation
- [**mutating-webhook.go**](https://github.com/piyushjajoo/k8s-operators-course/blob/main/module-05/solutions/mutating-webhook.go): Complete mutating webhook implementation

## Usage

These solutions can be used as:
- Reference when implementing your own webhooks
- Starting point if you get stuck
- Examples of best practices

## Integration

To use these solutions in your operator:

1. **For validating/mutating webhooks:** Copy the webhook code to `internal/webhook/v1/database_webhook.go`
2. Ensure your API types match the structure
3. Run `make generate` and `make manifests`

## Testing Webhooks

Webhooks require TLS certificates and must be reachable by the Kubernetes API server. Unlike controllers, webhooks cannot be easily tested with `make run`.

### Option 1: Deploy to Cluster (Recommended for webhook testing)

```bash
# Ensure cert-manager is installed (handles TLS certificates)
kubectl get pods -n cert-manager

# Build and load image into kind
make docker-build IMG=postgres-operator:latest
kind load docker-image postgres-operator:latest --name k8s-operators-course

# Deploy operator with webhooks
make deploy IMG=postgres-operator:latest
```

### Option 2: Local Development (Controller logic only)

```bash
# For testing controller/reconciliation logic (webhooks won't be invoked)
make install && make run
```

### Podman Users

```bash
# Build with podman
make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman

# Load into kind via tarball
podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
rm /tmp/postgres-operator.tar

# Deploy with localhost prefix
make deploy IMG=localhost/postgres-operator:latest
```

## Notes

- **Validating/Mutating webhooks:** Code goes in `internal/webhook/v1/` directory
- Uses `webhook.CustomValidator` and `webhook.CustomDefaulter` interfaces for admission webhooks
- Methods receive `context.Context` as first parameter
- `ValidateUpdate` receives both old and new objects as `runtime.Object`
- Error messages are clear and actionable
- Mutations are idempotent
- Validation covers common scenarios

## Important: CRD Schema Defaults vs Webhook Defaults

CRD schema defaults (via `+kubebuilder:default` markers) are applied **before** mutating webhooks run. This means:

- If your CRD has `+kubebuilder:default=1` for replicas, `Spec.Replicas` will be `1` (not `nil`) when your webhook runs
- To override CRD defaults in webhooks, check for the default value instead of `nil`

Example:
```go
// Instead of checking nil (won't work if CRD has default):
if database.Spec.Replicas == nil {
    replicas := int32(3)
    database.Spec.Replicas = &replicas
}

// Check for the value you want to override:
if database.Spec.Replicas == nil || *database.Spec.Replicas < 3 {
    replicas := int32(3)
    database.Spec.Replicas = &replicas
}
```

**Best Practice:** Use CRD schema defaults for simple static defaults, and webhooks for context-aware defaults (e.g., based on namespace).
