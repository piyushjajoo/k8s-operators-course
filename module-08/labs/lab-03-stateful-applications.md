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

> **Important Security Note:** In [Module 7](../../module-07/labs/lab-01-packaging-distribution.md), we recommended using distroless images for maximum security. However, distroless images don't include package managers, making it impossible to install PostgreSQL client tools (`pg_dump`, `psql`) directly.

**For this course (learning purposes):** We'll take a pragmatic shortcut and use a minimal Debian base image to include PostgreSQL client tools. This makes the lab simpler and allows you to focus on learning backup/restore functionality.

**For production:** See the production-ready alternatives below that maintain security while providing the necessary tools.

#### Option A: Minimal Debian Base (Course Shortcut)

For this course, update your `Dockerfile` to use a minimal Debian base image. Reference the example in [solutions/Dockerfile](../solutions/Dockerfile):

```dockerfile
# Runtime stage - use minimal Debian base instead of distroless
# NOTE: This is a shortcut for learning purposes
# For production, see Option B below
FROM debian:bookworm-slim

# Install PostgreSQL client tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    postgresql-client \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user and group
# Using UID 65532 to match distroless images (suppress warning with --no-log-init)
RUN groupadd -r -g 65532 nonroot && \
    useradd -r -u 65532 -g nonroot --no-log-init -m -s /bin/bash nonroot

WORKDIR /
COPY --from=builder /workspace/manager .

# Switch to non-root user
USER 65532:65532

ENTRYPOINT ["/manager"]
```

#### Option B: Production-Ready Approaches (Recommended)

For production environments, maintain security by using distroless images and one of these patterns:

**1. Sidecar Container Pattern**

Keep your operator using distroless, and use a sidecar container with PostgreSQL tools:

```yaml
# In your operator Deployment
spec:
  template:
    spec:
      containers:
      - name: manager
        image: your-operator:distroless
        # ... operator config ...
      - name: postgres-client
        image: postgres:14-alpine
        command: ["sleep", "infinity"]
        # Mount shared volume for backup files
        volumeMounts:
        - name: backup-storage
          mountPath: /backups
```

Then modify your backup code to exec into the sidecar container:

```go
// Execute pg_dump in sidecar container
cmd := exec.CommandContext(ctx, "kubectl", "exec", "-i", podName, 
    "-c", "postgres-client", "--", "pg_dump", ...)
```

**2. Kubernetes Jobs Pattern**

Create Kubernetes Jobs with PostgreSQL client tools for each backup:

```go
job := &batchv1.Job{
    Spec: batchv1.JobSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{{
                    Name:  "backup",
                    Image: "postgres:14-alpine",
                    Command: []string{"pg_dump", ...},
                }},
            },
        },
    },
}
```

**3. Separate Backup Operator**

Create a dedicated backup operator that includes PostgreSQL tools, keeping your main operator distroless.

**4. Init Container Pattern**

Use an init container to prepare backup tools, then use a shared volume.

> **For this course:** We'll use Option A (Debian base) to keep things simple. In production, choose one of the Option B approaches based on your security requirements and operational preferences.

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

### Task 1.3: Integrate with Backup Controller

Update `internal/controller/backup_controller.go` to use the backup package. The Backup controller from Lab 8.2 has a `createBackup` method that currently simulates the backup. Replace it with a call to the actual backup package.

**Current state (from Lab 8.2):**

The `createBackup` method in your Backup controller currently looks like this:

```go
func (r *BackupReconciler) createBackup(ctx context.Context, db *databasev1.Database, backup *databasev1.Backup) (string, error) {
    // Actual backup implementation would:
    // 1. Connect to database
    // 2. Create backup (pg_dump, mysqldump, etc.)
    // 3. Store backup in storage (S3, PVC, etc.)
    // 4. Return backup location

    backupLocation := fmt.Sprintf("s3://backups/%s/%s-%s.sql",
        db.Namespace,
        db.Name,
        time.Now().Format("20060102-150405"))

    // Simulate backup creation
    // In real implementation, this would actually perform the backup

    return backupLocation, nil
}
```

**Updated state:**

Replace the `createBackup` method to use the backup package:

```go
import (
    // ... existing imports ...
    backupPkg "github.com/example/postgres-operator/internal/backup"
)

// Replace the createBackup method to use the backup package
func (r *BackupReconciler) createBackup(ctx context.Context, db *databasev1.Database, backup *databasev1.Backup) (string, error) {
    // Use the backup package to perform actual backup
    // Note: PerformBackup requires k8sClient to retrieve password from Secret
    // Note: We use 'backupPkg' alias to avoid conflict with 'backup' variable name
    backupLocation, err := backupPkg.PerformBackup(ctx, r.Client, db)
    if err != nil {
        return "", fmt.Errorf("failed to perform backup: %w", err)
    }

    return backupLocation, nil
}
```

> **Important:** We use the package alias `backupPkg` because the function parameter `backup *databasev1.Backup` would shadow the package name `backup`. Without the alias, Go would try to call `backup.PerformBackup()` on the Backup resource variable instead of the backup package, causing a compile error: `backup.PerformBackup undefined`.

**How it works:**

The `performBackup` method in your controller already handles status updates correctly. It calls `createBackup` and updates the Backup status based on the result:

```go
func (r *BackupReconciler) performBackup(ctx context.Context, db *databasev1.Database, backup *databasev1.Backup) (ctrl.Result, error) {
    // Update status to in progress
    backup.Status.Phase = "InProgress"
    // ... status updates ...

    // Perform actual backup (calls createBackup)
    backupLocation, err := r.createBackup(ctx, db, backup)
    if err != nil {
        // Handle error and update status
        return ctrl.Result{}, err
    }

    // Update status to completed
    backup.Status.BackupLocation = backupLocation
    // ... more status updates ...
}
```

With this change, `createBackup` will now perform the actual backup using `pg_dump` instead of simulating it.

**Controller structure:**
- `performBackup()` - Handles status updates, error handling, and calls `createBackup()`
- `createBackup()` - Performs the actual backup work (now calls `backupPkg.PerformBackup()`)

### Task 1.4: Build, Deploy, and Test Backup Functionality

Now let's build and test the backup functionality:

```bash
# Generate code and manifests
make generate
make manifests

# Ensure code compiles
make build

# Build the container image (with PostgreSQL client tools)
make docker-build IMG=postgres-operator:latest

# Load image into kind cluster
kind load docker-image postgres-operator:latest --name k8s-operators-course

# Deploy operator to cluster
make deploy IMG=postgres-operator:latest

# rollout restart the deployment just in case you are using existing kind cluster with operator deployed
kubectl rollout restart deploy -n postgres-operator-system postgres-operator-controller-manager
kubectl rollout status deploy -n postgres-operator-system   postgres-operator-controller-manager

# Verify operator is running
kubectl get pods -n postgres-operator-system

# Check logs
kubectl logs -n postgres-operator-system -l control-plane=controller-manager -f
```

> **Using Podman instead of Docker?**
> 
> ```bash
> # Build with podman
> make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman
> 
> # Load image into kind (save to tarball, then load)
> podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
> kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
> rm /tmp/postgres-operator.tar
> 
> # Deploy with localhost/ prefix
> make deploy IMG=localhost/postgres-operator:latest
> ```

> **Getting `ErrImagePull` or `ImagePullBackOff`?**
> 
> Ensure `imagePullPolicy: IfNotPresent` is set in `config/manager/manager.yaml` and the image name matches what's loaded in kind.

**Test the backup functionality:**

```bash
# Create a Database
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

# Create a Backup
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Backup
metadata:
  name: test-backup
spec:
  databaseRef:
    name: test-db
EOF

# Watch Backup status
kubectl get backup test-backup -w

# Check Backup status details
kubectl get backup test-backup -o yaml

# Verify backup location is set
kubectl get backup test-backup -o jsonpath='{.status.backupLocation}'
```

**Verify backup was performed:**

```bash
# Check operator logs for backup activity
kubectl logs -n postgres-operator-system -l control-plane=controller-manager | grep -i backup

# Check Backup conditions
kubectl get backup test-backup -o jsonpath='{.status.conditions}'
```

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

Edit `internal/controller/restore_controller.go` to implement the complete reconciliation logic.

Copy the complete restore controller implementation from: **[solutions/restore-controller.go](../solutions/restore-controller.go)**

The restore controller:
- Waits for Database to be ready
- Waits for Backup to be completed
- Calls `restorePkg.PerformRestore()` to perform the actual restore
- Updates Restore status with phases (Pending → InProgress → Completed/Failed)
- Sets conditions for observability

**Key implementation details:**

```go
func (r *RestoreReconciler) performRestore(ctx context.Context, db *databasev1.Database, backup *databasev1.Backup, rst *databasev1.Restore) (ctrl.Result, error) {
    // Update status to in progress
    rst.Status.Phase = "InProgress"
    // ... status updates ...

    // Get backup location from Backup status
    if backup.Status.BackupLocation == "" {
        // Handle error
    }

    // Perform actual restore using restore package
    // Note: PerformRestore requires k8sClient to retrieve password from Secret
    err := restorePkg.PerformRestore(ctx, r.Client, db, backup.Status.BackupLocation)
    if err != nil {
        // Handle error and update status
    }

    // Update status to completed
    rst.Status.Phase = "Completed"
    rst.Status.RestoreTime = &metav1.Now()
    // ... more status updates ...
}
```

**Controller structure:**
- `Reconcile()` - Main reconciliation loop, validates prerequisites
- `performRestore()` - Handles status updates, error handling, and calls `restorePkg.PerformRestore()`

### Task 2.5: Register Restore Controller

Ensure the Restore controller is registered in `cmd/main.go`:

```go
if err = (&controller.RestoreReconciler{
    Client: mgr.GetClient(),
    Scheme: mgr.GetScheme(),
}).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "Restore")
    os.Exit(1)
}
```

### Task 2.6: Generate and Install CRDs

```bash
# Generate code and manifests
make generate
make manifests

# Install CRDs
make install

# Verify the CRD was created
kubectl get crd restores.database.example.com
```

### Task 2.7: Build, Deploy, and Test Restore Functionality

Now let's build and test the restore functionality:

```bash
# Generate code and manifests
make generate
make manifests

# Ensure code compiles
make build

# Build the container image
make docker-build IMG=postgres-operator:latest

# Load image into kind cluster
kind load docker-image postgres-operator:latest --name k8s-operators-course

# Deploy operator to cluster
make deploy IMG=postgres-operator:latest

# rollout restart the deployment just in case you are using existing kind cluster with operator deployed
kubectl rollout restart deploy -n postgres-operator-system postgres-operator-controller-manager
kubectl rollout status deploy -n postgres-operator-system   postgres-operator-controller-manager

# Verify operator is running
kubectl get pods -n postgres-operator-system

# Check logs
kubectl logs -n postgres-operator-system -l control-plane=controller-manager -f
```

> **Using Podman instead of Docker?**
> 
> ```bash
> # Build with podman
> make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman
> 
> # Load image into kind (save to tarball, then load)
> podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
> kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
> rm /tmp/postgres-operator.tar
> 
> # Deploy with localhost/ prefix
> make deploy IMG=localhost/postgres-operator:latest
> ```

> **Getting `ErrImagePull` or `ImagePullBackOff`?**
> 
> Ensure `imagePullPolicy: IfNotPresent` is set in `config/manager/manager.yaml` and the image name matches what's loaded in kind.

**Test the restore functionality:**

```bash
# Ensure you have a Database and completed Backup from Task 1.4
# If not, create them first:

# Create a Database
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: restore-test-db
spec:
  image: postgres:14
  replicas: 1
  databaseName: restoredb
  username: admin
  storage:
    size: "1Gi"
EOF

# Wait for Database to be ready
kubectl wait --for=jsonpath='{.status.phase}'=Ready database/restore-test-db --timeout=120s

# Create a Backup
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Backup
metadata:
  name: restore-test-backup
spec:
  databaseRef:
    name: restore-test-db
EOF

# Wait for Backup to complete
kubectl wait --for=jsonpath='{.status.phase}'=Completed backup/restore-test-backup --timeout=300s

# Verify backup location exists
kubectl get backup restore-test-backup -o jsonpath='{.status.backupLocation}'
echo

# Now create a Restore
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Restore
metadata:
  name: test-restore
spec:
  databaseRef:
    name: restore-test-db
  backupRef:
    name: restore-test-backup
EOF

# Watch Restore status
kubectl get restore test-restore -w

# Check Restore status details
kubectl get restore test-restore -o yaml

# Verify restore completed successfully
kubectl get restore test-restore -o jsonpath='{.status.phase}'
echo
```

**Verify restore was performed:**

```bash
# Check operator logs for restore activity
kubectl logs -n postgres-operator-system -l control-plane=controller-manager | grep -i restore

# Check Restore conditions
kubectl get restore test-restore -o jsonpath='{.status.conditions}'
echo

# Verify restore time is set
kubectl get restore test-restore -o jsonpath='{.status.restoreTime}'
echo
```

**Test error scenarios:**

```bash
# Test with non-existent database
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Restore
metadata:
  name: test-restore-fail-db
spec:
  databaseRef:
    name: non-existent-db
  backupRef:
    name: restore-test-backup
EOF

# Watch status - should stay in Pending
kubectl get restore test-restore-fail-db -w

# Test with non-existent backup
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Restore
metadata:
  name: test-restore-fail-backup
spec:
  databaseRef:
    name: restore-test-db
  backupRef:
    name: non-existent-backup
EOF

# Watch status - should stay in Pending
kubectl get restore test-restore-fail-backup -w
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

**Step 1: Add the helper functions**

Copy the complete implementation from `solutions/rolling-update.go` to `internal/controller/database_controller.go`. The functions handle:
- Detecting image changes
- Updating the StatefulSet to trigger rolling updates
- Waiting for all replicas to be updated and ready
- Handling replica count changes

**Step 2: Integrate into reconciliation logic**

To use these functions in your Database controller, call `updateStatefulSet()` from your reconciliation logic. For example, in your `Reconcile()` method or state machine handler:

```go
func (r *DatabaseReconciler) handleProvisioning(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    
    // ... other provisioning logic ...
    
    // Update StatefulSet (handles rolling updates if image/replicas changed)
    if err := r.updateStatefulSet(ctx, db); err != nil {
        logger.Error(err, "Failed to update StatefulSet")
        return ctrl.Result{}, err
    }
    
    // Check if StatefulSet is ready
    statefulSet := &appsv1.StatefulSet{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet)
    
    if err != nil {
        return ctrl.Result{}, err
    }
    
    // Transition to Ready if all replicas are ready
    if statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas {
        db.Status.Phase = "Ready"
        return ctrl.Result{}, r.Status().Update(ctx, db)
    }
    
    return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}
```

**Alternative: Call from main reconcile loop**

If you prefer, you can call `updateStatefulSet()` directly from your main `Reconcile()` method after ensuring the database is in a ready state:

```go
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ... get database ...
    
    // Handle rolling updates
    if err := r.updateStatefulSet(ctx, db); err != nil {
        return ctrl.Result{}, err
    }
    
    // ... rest of reconciliation ...
}
```

**How it works:**

1. `updateStatefulSet()` checks if the StatefulSet exists, creates it if not
2. Compares desired image/replicas with current StatefulSet spec
3. If different, updates the StatefulSet (triggers Kubernetes rolling update)
4. Calls `waitForRollingUpdate()` to wait for all pods to be updated and ready
5. Returns when the rolling update completes or times out

> **Note:** The existing Database controller from earlier modules already handles image updates. This exercise shows the explicit waiting pattern for more control. The `waitForRollingUpdate()` function uses `wait.PollImmediate()` to poll the StatefulSet status until all replicas are updated and ready, with a 5-minute timeout.

### Task 3.2: Test Rolling Updates

Build and deploy the updated operator:

```bash
# Ensure code compiles
make build

# Build the container image
make docker-build IMG=postgres-operator:latest

# Load image into kind cluster
kind load docker-image postgres-operator:latest --name k8s-operators-course

# Deploy operator to cluster
make deploy IMG=postgres-operator:latest

# Restart deployment if redeploying
kubectl rollout restart deploy -n postgres-operator-system postgres-operator-controller-manager
kubectl rollout status deploy -n postgres-operator-system postgres-operator-controller-manager
```

> **Using Podman instead of Docker?**
> 
> ```bash
> # Build with podman
> make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman
> 
> # Load image into kind (save to tarball, then load)
> podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
> kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
> rm /tmp/postgres-operator.tar
> 
> # Deploy with localhost/ prefix
> make deploy IMG=localhost/postgres-operator:latest
> ```

> **Getting `ErrImagePull` or `ImagePullBackOff`?**
> 
> Ensure `imagePullPolicy: IfNotPresent` is set in `config/manager/manager.yaml` and the image name matches what's loaded in kind.

**Test rolling update:**

```bash
# Create a Database with initial image
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: rolling-update-test
spec:
  image: postgres:14
  replicas: 2
  databaseName: testdb
  username: admin
  storage:
    size: "1Gi"
EOF

# Wait for Database to be ready
kubectl wait --for=jsonpath='{.status.phase}'=Ready database/rolling-update-test --timeout=120s

# Verify initial StatefulSet pods
kubectl get pods -l app=database,name=rolling-update-test

# Check current image version
kubectl get statefulset rolling-update-test -o jsonpath='{.spec.template.spec.containers[0].image}'
echo

# Update to new image version
kubectl patch database rolling-update-test --type=merge -p '{"spec":{"image":"postgres:15"}}'

# Watch StatefulSet update
kubectl get statefulset rolling-update-test -w

# Watch pods during rolling update
kubectl get pods -l app=database,name=rolling-update-test -w

# Verify all pods are updated
kubectl get statefulset rolling-update-test -o jsonpath='{.status.updatedReplicas}/{.spec.replicas}'
echo

# Verify new image version
kubectl get statefulset rolling-update-test -o jsonpath='{.spec.template.spec.containers[0].image}'
echo

# Check operator logs for rolling update activity
kubectl logs -n postgres-operator-system -l control-plane=controller-manager | grep -i "rolling\|update"
```

**Verify rolling update completed:**

```bash
# Check StatefulSet status
kubectl get statefulset rolling-update-test -o yaml | grep -A 10 status:

# Verify all replicas are ready
kubectl get statefulset rolling-update-test -o jsonpath='{.status.readyReplicas}/{.spec.replicas}'
echo

# Verify pods are running with new image
kubectl get pods -l app=database,name=rolling-update-test -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[0].image}{"\n"}{end}'
```

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

### Task 4.3: Test Data Consistency Checks

Build and deploy the updated operator:

```bash
# Ensure code compiles
make build

# Build the container image
make docker-build IMG=postgres-operator:latest

# Load image into kind cluster
kind load docker-image postgres-operator:latest --name k8s-operators-course

# Deploy operator to cluster
make deploy IMG=postgres-operator:latest

# Restart deployment if redeploying
kubectl rollout restart deploy -n postgres-operator-system postgres-operator-controller-manager
kubectl rollout status deploy -n postgres-operator-system postgres-operator-controller-manager
```

> **Using Podman instead of Docker?**
> 
> ```bash
> # Build with podman
> make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman
> 
> # Load image into kind (save to tarball, then load)
> podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
> kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
> rm /tmp/postgres-operator.tar
> 
> # Deploy with localhost/ prefix
> make deploy IMG=localhost/postgres-operator:latest
> ```

> **Getting `ErrImagePull` or `ImagePullBackOff`?**
> 
> Ensure `imagePullPolicy: IfNotPresent` is set in `config/manager/manager.yaml` and the image name matches what's loaded in kind.

**Test consistency checks:**

```bash
# Create a Database with multiple replicas
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: consistency-test
spec:
  image: postgres:14
  replicas: 3
  databaseName: testdb
  username: admin
  storage:
    size: "1Gi"
EOF

# Watch Database status transitions
kubectl get database consistency-test -w

# Wait for Database to reach Verifying phase
kubectl wait --for=jsonpath='{.status.phase}'=Verifying database/consistency-test --timeout=120s || true

# Check Database status
kubectl get database consistency-test -o yaml | grep -A 5 status:

# Wait for Database to be Ready (consistency checks should pass)
kubectl wait --for=jsonpath='{.status.phase}'=Ready database/consistency-test --timeout=300s

# Verify all replicas are ready
kubectl get statefulset consistency-test -o jsonpath='{.status.readyReplicas}/{.spec.replicas}'
echo

# Verify pods are running
kubectl get pods -l app=database,name=consistency-test

# Check operator logs for consistency check activity
kubectl logs -n postgres-operator-system -l control-plane=controller-manager | grep -i "consistency\|replica"
```

**Test consistency check failure scenario:**

```bash
# Scale down replicas to simulate inconsistency
kubectl patch database consistency-test --type=merge -p '{"spec":{"replicas":1}}'

# Wait for StatefulSet to scale down
kubectl wait --for=jsonpath='{.status.readyReplicas}'=1 statefulset/consistency-test --timeout=60s

# Scale back up
kubectl patch database consistency-test --type=merge -p '{"spec":{"replicas":3}}'

# Watch Database status during consistency checks
kubectl get database consistency-test -w

# Check logs for consistency check retries
kubectl logs -n postgres-operator-system -l control-plane=controller-manager | grep -i "consistency"
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

> **Using Podman instead of Docker?**
> 
> ```bash
> # Build with podman
> make docker-build IMG=postgres-operator:latest CONTAINER_TOOL=podman
> 
> # Load image into kind (save to tarball, then load)
> podman save localhost/postgres-operator:latest -o /tmp/postgres-operator.tar
> kind load image-archive /tmp/postgres-operator.tar --name k8s-operators-course
> rm /tmp/postgres-operator.tar
> 
> # Deploy with localhost/ prefix
> make deploy IMG=localhost/postgres-operator:latest
> ```

> **Getting `ErrImagePull` or `ImagePullBackOff`?**
> 
> Ensure `imagePullPolicy: IfNotPresent` is set in `config/manager/manager.yaml` and the image name matches what's loaded in kind.

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

