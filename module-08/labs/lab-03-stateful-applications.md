# Lab 8.3: Managing Stateful Applications

**Related Lesson:** [Lesson 8.3: Stateful Application Management](../lessons/03-stateful-applications.md)  
**Navigation:** [← Previous Lab: Operator Composition](lab-02-operator-composition.md) | [Module Overview](../README.md) | [Next Lab: Final Project →](lab-04-final-project.md)

## Objectives

- Implement backup functionality
- Add restore capability
- Handle rolling updates
- Ensure data consistency

## Prerequisites

- Completion of [Lab 8.2](lab-02-operator-composition.md)
- Database operator with Backup controller deployed
- Understanding of StatefulSets

## Exercise 1: Implement Backup Functionality

In Lab 8.2, we created a basic Backup controller. Now we'll add the actual backup logic.

### Task 1.1: Create Backup Package

Create a new package for backup operations. Copy the complete implementation from the solutions file:

```bash
# Create the backup package directory
mkdir -p internal/backup

# Copy the complete backup implementation
cp path/to/solutions/backup.go internal/backup/backup.go
```

Or, if you prefer to type it yourself, copy from:
**[solutions/backup.go](../solutions/backup.go)**

The backup package includes:
- `PerformBackup()` - Executes pg_dump and saves to storage
- `saveToStorage()` - Uploads backup to S3/PVC
- `PerformScheduledBackup()` - Handles scheduled backups

> **Important:** The backup implementation uses `pg_dump` which requires PostgreSQL client tools to be installed in your operator container. You'll need to update your Dockerfile to include the `postgresql-client` package. See Task 1.2 below.

### Task 1.2: Update Dockerfile for PostgreSQL Client Tools

Since the backup functionality uses `pg_dump` and restore uses `psql`, you need to install PostgreSQL client tools in your operator container image.

Update your `Dockerfile` to include PostgreSQL client tools. Reference the example in [solutions/Dockerfile](../solutions/Dockerfile):

```dockerfile
# Runtime stage - use minimal Debian base instead of distroless
# (distroless doesn't have package manager for installing tools)
FROM debian:bookworm-slim

# Install PostgreSQL client tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    postgresql-client \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -r -u 65532 -g 65532 -m -s /bin/bash nonroot

WORKDIR /
COPY --from=builder /workspace/manager .

# Switch to non-root user
USER 65532:65532

ENTRYPOINT ["/manager"]
```

**Alternative:** If you want to keep using distroless for security, you can:
1. Use a sidecar container with PostgreSQL tools
2. Use Kubernetes Jobs with PostgreSQL client tools for backups
3. Use a separate backup operator that has the tools

**Key backup logic:**

```go
func PerformBackup(ctx context.Context, k8sClient client.Client, db *databasev1.Database) (string, error) {
    endpoint := db.Status.Endpoint
    if endpoint == "" {
        return "", fmt.Errorf("database endpoint not available")
    }

    // Get password from Secret
    secretName := db.Status.SecretName
    if secretName == "" {
        secretName = fmt.Sprintf("%s-credentials", db.Name)
    }

    secret := &corev1.Secret{}
    err := k8sClient.Get(ctx, client.ObjectKey{
        Name:      secretName,
        Namespace: db.Namespace,
    }, secret)
    if err != nil {
        return "", fmt.Errorf("failed to get secret: %w", err)
    }

    password := string(secret.Data["password"])

    // Create backup filename
    backupFile := fmt.Sprintf("/backups/%s-%s.sql",
        db.Name,
        time.Now().Format("20060102-150405"))

    // Perform pg_dump with password from Secret
    cmd := exec.CommandContext(ctx, "pg_dump",
        "-h", endpoint,
        "-U", db.Spec.Username,
        "-d", db.Spec.DatabaseName,
        "-f", backupFile)

    // Set password as environment variable
    cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", password))

    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("backup failed: %v, output: %s", err, string(output))
    }

    // Save to storage (S3, PVC, etc.)
    return saveToStorage(backupFile)
}
```

### Task 1.2: Integrate with Backup Controller

Update `internal/controller/backup_controller.go` to use the backup package.

The `performBackup` function in the Backup controller (from Lab 8.2) can call the backup package:

```go
import (
    // ... existing imports ...
    "github.com/example/postgres-operator/internal/backup"
)

func (r *BackupReconciler) performBackup(ctx context.Context, db *databasev1.Database, bkp *databasev1.Backup) (ctrl.Result, error) {
    // Update status to in progress
    bkp.Status.Phase = "InProgress"
    r.Status().Update(ctx, bkp)

    // Perform actual backup
    // Note: PerformBackup requires k8sClient to retrieve password from Secret
    location, err := backup.PerformBackup(ctx, r.Client, db)
    if err != nil {
        bkp.Status.Phase = "Failed"
        r.Status().Update(ctx, bkp)
        return ctrl.Result{}, err
    }

    // Update status to completed
    bkp.Status.Phase = "Completed"
    now := metav1.Now()
    bkp.Status.BackupTime = &now
    bkp.Status.BackupLocation = location

    return ctrl.Result{}, r.Status().Update(ctx, bkp)
}
```

> **Note:** For this lab, the simplified backup in Lab 8.2's solution simulates the backup. The `backup.go` solution shows a more realistic implementation using `pg_dump`.

## Exercise 2: Implement Restore

### Task 2.1: Scaffold Restore API with Kubebuilder

Use kubebuilder to scaffold a new Restore API (same group as Database and Backup):

```bash
# Navigate to your operator project
cd ~/postgres-operator

# Scaffold the Restore API
kubebuilder create api \
  --group database \
  --version v1 \
  --kind Restore \
  --resource --controller

# When prompted:
# Create Resource [y/n]: y
# Create Controller [y/n]: y
```

This creates:
- `api/v1/restore_types.go` - API type definitions
- `internal/controller/restore_controller.go` - Controller scaffold

### Task 2.2: Define Restore Spec and Status

Edit `api/v1/restore_types.go` to define the Restore resource:

```go
package v1

import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RestoreSpec defines the desired state of Restore
type RestoreSpec struct {
    // DatabaseRef references the Database to restore to
    // +kubebuilder:validation:Required
    DatabaseRef corev1.LocalObjectReference `json:"databaseRef"`

    // BackupRef references the Backup to restore from
    // +kubebuilder:validation:Required
    BackupRef corev1.LocalObjectReference `json:"backupRef"`
}

// RestoreStatus defines the observed state of Restore
type RestoreStatus struct {
    // Phase is the current restore phase
    // +kubebuilder:validation:Enum=Pending;InProgress;Completed;Failed
    Phase string `json:"phase,omitempty"`

    // RestoreTime is when the restore completed
    RestoreTime *metav1.Time `json:"restoreTime,omitempty"`

    // Conditions represent the latest observations
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Database",type="string",JSONPath=".spec.databaseRef.name"
// +kubebuilder:printcolumn:name="Backup",type="string",JSONPath=".spec.backupRef.name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Restore is the Schema for the restores API
type Restore struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   RestoreSpec   `json:"spec,omitempty"`
    Status RestoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RestoreList contains a list of Restore
type RestoreList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []Restore `json:"items"`
}

func init() {
    SchemeBuilder.Register(&Restore{}, &RestoreList{})
}
```

### Task 2.3: Create Restore Package

Create the restore package with the actual restore logic:

```bash
# Create the restore package directory
mkdir -p internal/restore

# Copy the complete restore implementation
cp path/to/solutions/restore.go internal/restore/restore.go
```

Or copy from: **[solutions/restore.go](../solutions/restore.go)**

The restore package includes:
- `PerformRestore()` - Loads backup and restores to database
- `loadFromStorage()` - Downloads backup from S3/PVC
- `stopDatabase()` / `startDatabase()` - Graceful database operations

> **Note:** The restore implementation uses `psql` which also requires PostgreSQL client tools. If you haven't already updated your Dockerfile in Task 1.2, make sure to do so now.

**Key restore logic:**

```go
func PerformRestore(ctx context.Context, k8sClient client.Client, db *databasev1.Database, backupLocation string) error {
    // Load backup from storage
    backupData, err := loadFromStorage(backupLocation)
    if err != nil {
        return fmt.Errorf("failed to load backup: %v", err)
    }

    endpoint := db.Status.Endpoint
    if endpoint == "" {
        return fmt.Errorf("database endpoint not available")
    }

    // Get password from Secret
    secretName := db.Status.SecretName
    if secretName == "" {
        secretName = fmt.Sprintf("%s-credentials", db.Name)
    }

    secret := &corev1.Secret{}
    err = k8sClient.Get(ctx, client.ObjectKey{
        Name:      secretName,
        Namespace: db.Namespace,
    }, secret)
    if err != nil {
        return fmt.Errorf("failed to get secret: %w", err)
    }

    password := string(secret.Data["password"])

    // Perform restore using psql with password from Secret
    cmd := exec.CommandContext(ctx, "psql",
        "-h", endpoint,
        "-U", db.Spec.Username,
        "-d", db.Spec.DatabaseName)

    // Set password as environment variable
    cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", password))
    cmd.Stdin = bytes.NewReader(backupData)

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("restore failed: %v, output: %s", err, string(output))
    }

    return nil
}
```

### Task 2.4: Implement Restore Controller

Edit `internal/controller/restore_controller.go` to implement the reconciliation:

```go
// +kubebuilder:rbac:groups=database.example.com,resources=restores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=restores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=databases,verbs=get;list;watch
// +kubebuilder:rbac:groups=database.example.com,resources=backups,verbs=get;list;watch

func (r *RestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := ctrl.LoggerFrom(ctx)

    rst := &databasev1.Restore{}
    if err := r.Get(ctx, req.NamespacedName, rst); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Skip if already completed
    if rst.Status.Phase == "Completed" {
        return ctrl.Result{}, nil
    }

    // Get Database
    db := &databasev1.Database{}
    if err := r.Get(ctx, client.ObjectKey{
        Name:      rst.Spec.DatabaseRef.Name,
        Namespace: rst.Namespace,
    }, db); err != nil {
        rst.Status.Phase = "Pending"
        r.Status().Update(ctx, rst)
        return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
    }

    // Get Backup
    backup := &databasev1.Backup{}
    if err := r.Get(ctx, client.ObjectKey{
        Name:      rst.Spec.BackupRef.Name,
        Namespace: rst.Namespace,
    }, backup); err != nil {
        rst.Status.Phase = "Pending"
        r.Status().Update(ctx, rst)
        return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
    }

    // Check backup is completed
    if backup.Status.Phase != "Completed" {
        log.Info("Waiting for backup to complete", "backup", backup.Name)
        rst.Status.Phase = "Pending"
        r.Status().Update(ctx, rst)
        return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
    }

    // Perform restore
    return r.performRestore(ctx, db, backup, rst)
}
```

### Task 2.5: Generate and Install CRDs

```bash
# Generate code and manifests
make generate
make manifests

# Install CRDs
make install

# Verify the CRD was created
kubectl get crd restores.database.example.com
```

## Exercise 3: Handle Rolling Updates

Rolling updates allow you to update the database image without downtime. The Database controller already handles basic updates, but this exercise shows advanced patterns.

### Task 3.1: Review Rolling Update Logic

The complete rolling update implementation is in:
**[solutions/rolling-update.go](../solutions/rolling-update.go)**

The key functions are:
- `updateStatefulSet()` - Detects changes and updates StatefulSet
- `waitForRollingUpdate()` - Waits for all pods to be updated
- `createStatefulSet()` - Creates new StatefulSet if needed

Add these helper functions to `internal/controller/database_controller.go`:

```go
func (r *DatabaseReconciler) updateStatefulSet(ctx context.Context, db *databasev1.Database) error {
    statefulSet := &appsv1.StatefulSet{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet)

    if errors.IsNotFound(err) {
        return r.createStatefulSet(ctx, db)
    }
    if err != nil {
        return err
    }

    // Check if update needed
    desiredImage := db.Spec.Image
    currentImage := statefulSet.Spec.Template.Spec.Containers[0].Image

    if desiredImage != currentImage {
        // Update image
        statefulSet.Spec.Template.Spec.Containers[0].Image = desiredImage

        // Update StatefulSet (triggers rolling update)
        if err := r.Update(ctx, statefulSet); err != nil {
            return err
        }

        // Wait for update to complete
        return r.waitForRollingUpdate(ctx, statefulSet)
    }

    return nil
}

func (r *DatabaseReconciler) waitForRollingUpdate(ctx context.Context, ss *appsv1.StatefulSet) error {
    return wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
        err := r.Get(ctx, client.ObjectKeyFromObject(ss), ss)
        if err != nil {
            return false, err
        }

        // Check if all replicas are updated and ready
        return ss.Status.UpdatedReplicas == *ss.Spec.Replicas &&
            ss.Status.ReadyReplicas == *ss.Spec.Replicas, nil
    })
}
```

> **Note:** The existing Database controller from earlier modules already handles image updates. This exercise shows the explicit waiting pattern for more control.

## Exercise 4: Ensure Data Consistency

Data consistency is critical for stateful applications. This exercise shows patterns for verifying consistency.

### Task 4.1: Add Consistency Check Functions

Add these helper functions to `internal/controller/database_controller.go`:

```go
func (r *DatabaseReconciler) ensureDataConsistency(ctx context.Context, db *databasev1.Database) error {
    logger := log.FromContext(ctx)

    statefulSet := &appsv1.StatefulSet{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet)

    if err != nil {
        return err
    }

    // Check all replicas are ready
    if statefulSet.Status.ReadyReplicas != *statefulSet.Spec.Replicas {
        return fmt.Errorf("not all replicas ready: %d/%d",
            statefulSet.Status.ReadyReplicas, *statefulSet.Spec.Replicas)
    }

    logger.Info("All replicas ready, checking consistency",
        "replicas", statefulSet.Status.ReadyReplicas)

    // Perform consistency check
    return r.performConsistencyCheck(ctx, db)
}

func (r *DatabaseReconciler) performConsistencyCheck(ctx context.Context, db *databasev1.Database) error {
    // Application-specific consistency checks:
    // 1. Connect to primary and verify it's writable
    // 2. Connect to replicas and verify replication lag
    // 3. Run test queries to verify data integrity

    // Example: Check replication status (PostgreSQL)
    // SELECT * FROM pg_stat_replication;

    // For this lab, we'll just verify the database is accessible
    // In production, implement actual consistency checks

    return nil
}
```

### Task 4.2: Integrate Consistency Check

You can call `ensureDataConsistency` in the `handleVerifying` state of the Database controller, or after rolling updates complete.

Example integration in `handleVerifying`:

```go
func (r *DatabaseReconciler) handleVerifying(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // Verify database consistency
    if err := r.ensureDataConsistency(ctx, db); err != nil {
        logger.Info("Consistency check failed, retrying", "error", err.Error())
        return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
    }

    // All checks passed, transition to Ready
    db.Status.Phase = "Ready"
    db.Status.Ready = true
    return ctrl.Result{}, r.Status().Update(ctx, db)
}
```

## Exercise 5: Test Backup and Restore

### Task 5.1: Build and Deploy Operator

Build and deploy the operator with the new Restore controller:

```bash
# Generate code and manifests
make generate
make manifests

# Build the container image
make docker-build IMG=postgres-operator:latest

# Load image into kind cluster
kind load docker-image postgres-operator:latest --name k8s-operators-course

# Deploy operator to cluster
make deploy IMG=postgres-operator:latest

# If redeploying, restart the deployment
kubectl rollout restart deploy -n postgres-operator-system postgres-operator-controller-manager
kubectl rollout status deploy -n postgres-operator-system postgres-operator-controller-manager
```

### Task 5.2: Create Database and Backup

```bash
# Create Database
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-db
spec:
  image: postgres:14
  replicas: 1
  databaseName: testdb
  username: admin
  storage:
    size: "1Gi"
EOF

# Wait for Database to be ready
kubectl wait --for=jsonpath='{.status.phase}'=Ready database/test-db --timeout=120s

# Create Backup
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Backup
metadata:
  name: test-backup
spec:
  databaseRef:
    name: test-db
EOF

# Check Backup status
kubectl get backup test-backup
```

### Task 5.3: Test Restore

```bash
# Wait for Backup to complete
kubectl wait --for=jsonpath='{.status.phase}'=Completed backup/test-backup --timeout=60s

# Create Restore
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Restore
metadata:
  name: test-restore
spec:
  databaseRef:
    name: test-db
  backupRef:
    name: test-backup
EOF

# Check Restore status
kubectl get restore test-restore

# Verify restore completed
kubectl get restore test-restore -o yaml
```

### Task 5.4: Verify All Resources

```bash
# Check all resources
kubectl get databases
kubectl get backups
kubectl get restores

# Check operator logs
kubectl logs -n postgres-operator-system -l control-plane=controller-manager | grep -i restore
```

## Cleanup

```bash
# Delete test resources
kubectl delete restore test-restore
kubectl delete backup test-backup
kubectl delete database test-db
```

## Lab Summary

In this lab, you:
- Created backup package with `pg_dump` integration
- Scaffolded Restore API using kubebuilder
- Implemented Restore controller
- Added rolling update handling patterns
- Implemented data consistency checks

## Key Learnings

1. **Use kubebuilder to scaffold new APIs** - `kubebuilder create api` for Restore
2. **Separate concerns with packages** - `internal/backup/` and `internal/restore/`
3. **Rolling updates are handled by StatefulSet** - Controller just updates spec
4. **Wait for updates to complete** - Use `wait.PollImmediate` pattern
5. **Data consistency is application-specific** - Implement checks for your database
6. **Coordinate multiple resources** - Restore depends on both Database and Backup

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Backup Implementation](../solutions/backup.go) - Complete backup functionality with `pg_dump`
- [Restore Implementation](../solutions/restore.go) - Complete restore functionality with `psql`
- [Rolling Update](../solutions/rolling-update.go) - Rolling update handling with wait logic

### Using the Solutions

```bash
# Copy backup package
mkdir -p internal/backup
cp path/to/solutions/backup.go internal/backup/

# Copy restore package  
mkdir -p internal/restore
cp path/to/solutions/restore.go internal/restore/

# Reference rolling-update.go for Database controller enhancements
```

## Next Steps

Now let's build the final project!

**Navigation:** [← Previous Lab: Operator Composition](lab-02-operator-composition.md) | [Related Lesson](../lessons/03-stateful-applications.md) | [Next Lab: Final Project →](lab-04-final-project.md)

