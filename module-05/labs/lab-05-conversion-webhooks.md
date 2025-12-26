---
layout: default
title: "Lab 05.5: Conversion Webhooks"
nav_order: 14
parent: "Module 5: Webhooks & Admission Control"
grand_parent: Modules
mermaid: true
---

# Lab 5.5: Conversion Webhooks and API Versioning

**Related Lesson:** [Lesson 5.5: Conversion Webhooks](../lessons/05-conversion-webhooks.md)  
**Navigation:** [← Previous Lab: Webhook Deployment](lab-04-webhook-deployment.md) | [Module Overview](../README.md)

## Objectives

- Create a v2 API version for your Database resource
- Implement conversion functions between v1 and v2
- Configure conversion webhook in CRD
- Test bidirectional conversion
- Verify round-trip conversion preserves data

## Prerequisites

- Completion of [Lab 5.4](lab-04-webhook-deployment.md)
- Database operator with webhooks deployed
- Understanding of API versioning concepts

## Exercise 1: Create v2 API Version

### Task 1.1: Generate v2 API

```bash
# Navigate to your operator
cd ~/postgres-operator

# Create v2 API version
kubebuilder create api --group database --version v2 --kind Database
```

**Observe:**
- What files were created in `api/v2/`?
- What was modified in existing files?

### Task 1.2: Examine v2 API Structure

```bash
# Check v2 types
cat api/v2/database_types.go

# Check v2 group version info
cat api/v2/groupversion_info.go
```

**Note:** The generated v2 types will be identical to v1 initially. We'll modify them in the next task.

## Exercise 2: Evolve v2 API

Let's evolve the API by adding new features. This demonstrates a realistic API evolution scenario.

> **Important:** v1 remains the **storage version** and continues to be used throughout the course (modules 6, 7, 8) for consistency. Conversion webhooks allow both versions to coexist.

### Task 2.1: Update v2 DatabaseSpec

Edit `api/v2/database_types.go` to evolve the API. The v1 API already has `DatabaseName`, `Storage` as `StorageSpec`, etc. For v2, we'll add replication mode and backup configuration:

```go
package v2

import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DatabaseSpec defines the desired state of Database
type DatabaseSpec struct {
    Image        string                        `json:"image"`
    DatabaseName string                        `json:"databaseName"`
    Username     string                        `json:"username"`
    
    // Replication configuration (enhanced from v1)
    Replication *ReplicationConfig            `json:"replication,omitempty"`
    
    // Storage configuration (same structure as v1)
    Storage      StorageSpec                   `json:"storage"`
    
    // Resources (same as v1)
    Resources    corev1.ResourceRequirements   `json:"resources,omitempty"`
    
    // Backup configuration (new in v2)
    Backup       *BackupConfig                 `json:"backup,omitempty"`
}

// ReplicationConfig defines replication settings
type ReplicationConfig struct {
    // Replicas moved from top-level spec (v1 has Replicas at top level)
    Replicas *int32 `json:"replicas,omitempty"`
    
    // Mode specifies replication mode (new in v2)
    Mode string `json:"mode,omitempty"` // "async" or "sync"
}

// StorageSpec defines storage settings (same as v1)
type StorageSpec struct {
    Size        string `json:"size"`
    StorageClass string `json:"storageClass,omitempty"`
}

// BackupConfig defines backup settings (new in v2)
type BackupConfig struct {
    Enabled       bool   `json:"enabled"`
    Schedule      string `json:"schedule,omitempty"` // Cron expression
    RetentionDays int32  `json:"retentionDays,omitempty"`
}

// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
    // Phase indicates the current phase
    Phase DatabasePhase `json:"phase,omitempty"`
    
    // Ready indicates if the database is ready
    Ready bool `json:"ready,omitempty"`
}

// DatabasePhase represents the phase of a Database
type DatabasePhase string

const (
    DatabasePhasePending   DatabasePhase = "Pending"
    DatabasePhaseCreating  DatabasePhase = "Creating"
    DatabasePhaseRunning   DatabasePhase = "Running"
    DatabasePhaseFailed    DatabasePhase = "Failed"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Database is the Schema for the databases API
type Database struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   DatabaseSpec   `json:"spec,omitempty"`
    Status DatabaseStatus `json:"status,omitempty"`
}

// Hub marks this type as a conversion hub.
func (*Database) Hub() {}

// +kubebuilder:object:root=true

// DatabaseList contains a list of Database
type DatabaseList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []Database `json:"items"`
}

func init() {
    SchemeBuilder.Register(&Database{}, &DatabaseList{})
}
```

### Task 2.2: Regenerate Code

```bash
# Generate deepcopy methods
make generate

# Generate manifests
make manifests
```

## Exercise 3: Implement Conversion Functions

### Task 3.1: Create Conversion File for v1

Create `api/v1/database_conversion.go`:

```go
package v1

import (
    "fmt"
    "sigs.k8s.io/controller-runtime/pkg/conversion"
    databasev2 "github.com/example/postgres-operator/api/v2"
)

// ConvertTo converts this Database to the Hub version (v2)
func (src *Database) ConvertTo(dstRaw conversion.Hub) error {
    dst, ok := dstRaw.(*databasev2.Database)
    if !ok {
        return fmt.Errorf("expected *v2.Database, got %T", dstRaw)
    }

    // Convert metadata
    dst.ObjectMeta = src.ObjectMeta

    // Convert spec: v1 → v2
    dst.Spec.Image = src.Spec.Image
    dst.Spec.DatabaseName = src.Spec.DatabaseName
    dst.Spec.Username = src.Spec.Username
    dst.Spec.Storage = src.Spec.Storage // Same structure
    dst.Spec.Resources = src.Spec.Resources
    
    // Convert replication: v1 has Replicas at top level, v2 has ReplicationConfig
    if src.Spec.Replicas != nil {
        if dst.Spec.Replication == nil {
            dst.Spec.Replication = &databasev2.ReplicationConfig{}
        }
        dst.Spec.Replication.Replicas = src.Spec.Replicas
        // Default mode if not set
        if dst.Spec.Replication.Mode == "" {
            dst.Spec.Replication.Mode = "async"
        }
    }
    
    // Backup config is new in v2, leave nil (no v1 equivalent)

    // Convert status
    dst.Status.Phase = src.Status.Phase
    dst.Status.Ready = src.Status.Ready
    dst.Status.Endpoint = src.Status.Endpoint
    dst.Status.SecretName = src.Status.SecretName
    dst.Status.Conditions = src.Status.Conditions

    return nil
}

// ConvertFrom converts from the Hub version (v2) to this version (v1)
func (dst *Database) ConvertFrom(srcRaw conversion.Hub) error {
    src, ok := srcRaw.(*databasev2.Database)
    if !ok {
        return fmt.Errorf("expected *v2.Database, got %T", srcRaw)
    }

    // Convert metadata
    dst.ObjectMeta = src.ObjectMeta

    // Convert spec: v2 → v1
    dst.Spec.Image = src.Spec.Image
    dst.Spec.DatabaseName = src.Spec.DatabaseName
    dst.Spec.Username = src.Spec.Username
    dst.Spec.Storage = src.Spec.Storage // Same structure
    dst.Spec.Resources = src.Spec.Resources
    
    // Convert replication: v2 has ReplicationConfig, v1 has Replicas at top level
    if src.Spec.Replication != nil {
        dst.Spec.Replicas = src.Spec.Replication.Replicas
    }
    // Note: Replication.Mode is lost in v1 conversion (acceptable - v1 doesn't support it)
    
    // Note: Backup config is lost in v1 conversion (v1 doesn't support backups)

    // Convert status
    dst.Status.Phase = src.Status.Phase
    dst.Status.Ready = src.Status.Ready
    dst.Status.Endpoint = src.Status.Endpoint
    dst.Status.SecretName = src.Status.SecretName
    dst.Status.Conditions = src.Status.Conditions

    return nil
}
```

### Task 3.2: Verify Conversion Functions

```bash
# Check that conversion file compiles
go build ./api/v1/...

# Check that v2 API compiles
go build ./api/v2/...
```

## Exercise 4: Configure Conversion Webhook

### Task 4.1: Update CRD for Conversion

The CRD should already have both versions after running `make manifests`. Verify:

```bash
# Check CRD has both versions
kubectl get crd databases.database.example.com -o yaml | grep -A 5 "versions:"
```

### Task 4.2: Set Storage Version

Edit `config/crd/patches/webhook_in_database.yaml` or manually patch the CRD to ensure:
- v1 is the storage version (`storage: true`)
- v2 is served but not storage (`storage: false`)
- Conversion strategy is set to Webhook

Check `config/crd/kustomization.yaml` and ensure webhook patches are included.

### Task 4.3: Verify Conversion Configuration

After running `make manifests`, check the generated CRD:

```bash
# Check conversion configuration
make manifests
cat config/crd/bases/database.example.com_databases.yaml | grep -A 20 "conversion:"
```

The conversion section should look like:

```yaml
conversion:
  strategy: Webhook
  webhook:
    clientConfig:
      service:
        namespace: postgres-operator-system
        name: postgres-operator-webhook-service
        path: /convert
    conversionReviewVersions:
    - v1
```

## Exercise 5: Register Conversion Webhook

### Task 5.1: Update main.go

Edit `cmd/main.go` to register the conversion webhook:

```go
package main

import (
    // ... existing imports ...
    "sigs.k8s.io/controller-runtime/pkg/webhook/conversion"
)

func main() {
    // ... existing code ...
    
    if err = (&databasev1.Database{}).SetupWebhookWithManager(mgr); err != nil {
        setupLog.Error(err, "unable to create webhook", "webhook", "Database")
        os.Exit(1)
    }
    
    // Setup conversion webhook
    if err = conversion.NewWebhookHandler(mgr.GetScheme()).SetupWebhookWithManager(mgr); err != nil {
        setupLog.Error(err, "unable to create conversion webhook")
        os.Exit(1)
    }
    
    // ... rest of main ...
}
```

### Task 5.2: Verify Webhook Registration

```bash
# Build the operator
make build

# Check for compilation errors
go vet ./...
```

## Exercise 6: Deploy and Test Conversion

### Task 6.1: Deploy Updated Operator

```bash
# Build and deploy
make docker-build IMG=postgres-operator:latest
kind load docker-image postgres-operator:latest --name k8s-operators-course
make deploy IMG=postgres-operator:latest

# Wait for deployment
kubectl wait --for=condition=Available deployment/postgres-operator-controller-manager \
  -n postgres-operator-system --timeout=120s
```

### Task 6.2: Verify CRD Versions

```bash
# Check CRD has both versions
kubectl get crd databases.database.example.com -o jsonpath='{.spec.versions[*].name}'

# Should show: v1 v2

# Check storage version
kubectl get crd databases.database.example.com -o jsonpath='{.spec.versions[?(@.storage==true)].name}'

# Should show: v1
```

### Task 6.3: Test v1 → v2 Conversion

```bash
# Create resource using v1 API (matches Module 3 structure)
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-conversion-v1
spec:
  image: postgres:14
  databaseName: mydatabase
  username: admin
  replicas: 3
  storage:
    size: 100Gi
    storageClass: standard
EOF

# Read using v2 API
kubectl get database.v2.database.example.com test-conversion-v1 -o yaml

# Verify conversion:
# - spec.databaseName should be "mydatabase"
# - spec.replication.replicas should be 3
# - spec.replication.mode should be "async" (default)
# - spec.storage.size should be "100Gi"
# - spec.storage.storageClass should be "standard"
```

### Task 6.4: Test v2 → v1 Conversion

```bash
# Create resource using v2 API
kubectl apply -f - <<EOF
apiVersion: database.example.com/v2
kind: Database
metadata:
  name: test-conversion-v2
spec:
  image: postgres:14
  databaseName: mydatabase-v2
  username: admin
  replication:
    replicas: 5
    mode: sync
  storage:
    size: 200Gi
    storageClass: ssd
  backup:
    enabled: true
    schedule: "0 2 * * *"
    retentionDays: 30
EOF

# Read using v1 API
kubectl get database.v1.database.example.com test-conversion-v2 -o yaml

# Verify conversion:
# - spec.databaseName should be "mydatabase-v2"
# - spec.replicas should be 5
# - spec.storage.size should be "200Gi"
# - spec.storage.storageClass should be "ssd"
# - Note: backup config and replication.mode are lost (v1 doesn't support them)
```

### Task 6.5: Test Round-Trip Conversion

```bash
# Create v1 resource
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-roundtrip
spec:
  image: postgres:14
  databaseName: roundtrip-db
  username: admin
  replicas: 2
  storage:
    size: 50Gi
    storageClass: standard
EOF

# Read as v2
V2_RESOURCE=$(kubectl get database.v2.database.example.com test-roundtrip -o yaml)

# Read as v1 again
V1_RESOURCE=$(kubectl get database.v1.database.example.com test-roundtrip -o yaml)

# Compare - the v1 resource should match what we originally created
echo "$V1_RESOURCE" | grep -A 5 "spec:"
```

**Expected:** The v1 resource should have the same values as originally created, demonstrating lossless conversion.

## Exercise 7: Handle Edge Cases

### Task 7.1: Test Missing Fields

```bash
# Create v1 resource without replicas
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-missing-fields
spec:
  image: postgres:14
  databaseName: test-db
  username: admin
  storage:
    size: 100Gi
    storageClass: standard
EOF

# Read as v2 - replication should be nil or have defaults
kubectl get database.v2.database.example.com test-missing-fields -o yaml
```

### Task 7.2: Test v2 with Minimal Fields

```bash
# Create v2 resource with minimal fields
kubectl apply -f - <<EOF
apiVersion: database.example.com/v2
kind: Database
metadata:
  name: test-minimal-v2
spec:
  image: postgres:14
  databaseName: minimal-db
  username: admin
  storage:
    size: 100Gi
    storageClass: standard
EOF

# Read as v1
kubectl get database.v1.database.example.com test-minimal-v2 -o yaml
```

## Exercise 8: Verify Conversion Webhook Logs

### Task 8.1: Check Webhook Logs

```bash
# Check conversion webhook logs
kubectl logs -n postgres-operator-system \
  deployment/postgres-operator-controller-manager | grep -i convert
```

### Task 8.2: Test Conversion Failure Handling

If conversion fails, Kubernetes should handle it gracefully. Test by temporarily breaking conversion:

1. Modify conversion function to return an error
2. Redeploy
3. Try to read a resource in different version
4. Observe error handling

## Challenge: Add v3 API

As a challenge, try adding a v3 API version:

1. Create v3 API with further evolution
2. Set v2 as Hub, v3 as non-Hub
3. Implement conversion between v2 and v3
4. Test all conversion paths: v1 ↔ v2 ↔ v3

## Cleanup

```bash
# Delete test resources
kubectl delete database test-conversion-v1 test-conversion-v2 test-roundtrip \
  test-missing-fields test-minimal-v2

# Optional: Remove v2 API if not needed
# (This requires careful planning and migration)
```

## Key Takeaways

- **API versioning** enables safe API evolution
- **Conversion webhooks** handle version conversion automatically
- **Storage version** is what's stored in etcd
- **Conversion functions** must be bidirectional and lossless
- **Hub pattern** designates one version as the conversion hub
- **Round-trip conversion** should preserve all data
- **Test thoroughly** before deploying to production

## Next Steps

- Review the [solution](../solutions/conversion-webhook.go) for reference
- Experiment with more complex API evolutions
- Consider deprecation strategies for old versions
- Plan migration path for existing resources

**Navigation:** [← Previous Lab: Webhook Deployment](lab-04-webhook-deployment.md) | [Module Overview](../README.md)

