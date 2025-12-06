# Lab 3.3: Building PostgreSQL Operator

**Related Lesson:** [Lesson 3.3: Implementing Reconciliation Logic](../lessons/03-reconciliation-logic.md)  
**Navigation:** [← Previous Lab: Designing API](lab-02-designing-api.md) | [Module Overview](../README.md) | [Next Lab: Client-Go →](lab-04-client-go.md)

## Objectives

- Implement reconciliation logic for PostgreSQL operator
- Handle resource creation and updates
- Use owner references
- Manage Secrets for database credentials
- Test idempotency

## Prerequisites

- Completion of [Lab 3.2](lab-02-designing-api.md)
- Database API defined
- Understanding of reconciliation patterns

## Exercise 1: Implement Basic Reconciliation

### Task 1.1: Set Up Controller Structure

Edit `internal/controller/database_controller.go`:

```go
package controller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	databasev1 "github.com/example/postgres-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=database.example.com,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=databases/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    
    // Read Database resource
    db := &databasev1.Database{}
    if err := r.Get(ctx, req.NamespacedName, db); err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }
    
    logger.Info("Reconciling Database", "name", db.Name)
    
    // Reconcile Secret (must be done before StatefulSet)
    if err := r.reconcileSecret(ctx, db); err != nil {
        return ctrl.Result{}, err
    }
    
    // Reconcile StatefulSet
    if err := r.reconcileStatefulSet(ctx, db); err != nil {
        return ctrl.Result{}, err
    }
    
    // Reconcile Service
    if err := r.reconcileService(ctx, db); err != nil {
        return ctrl.Result{}, err
    }
    
    // Update status
    if err := r.updateStatus(ctx, db); err != nil {
        return ctrl.Result{}, err
    }
    
    return ctrl.Result{}, nil
}
```

## Exercise 2: Implement Secret Management

The controller automatically generates a random password and stores it in a Kubernetes Secret.
This is more secure than requiring users to specify passwords in plain text.

### Task 2.1: Helper Functions

Add helper functions for Secret management:

```go
// secretName returns the name of the Secret for this Database
func (r *DatabaseReconciler) secretName(db *databasev1.Database) string {
	return fmt.Sprintf("%s-credentials", db.Name)
}

// generatePassword generates a random password
func generatePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
```

### Task 2.2: Reconcile Secret

```go
// reconcileSecret ensures the credentials Secret exists
func (r *DatabaseReconciler) reconcileSecret(ctx context.Context, db *databasev1.Database) error {
	logger := log.FromContext(ctx)
	secretName := r.secretName(db)

	secret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: db.Namespace,
	}, secret)

	if errors.IsNotFound(err) {
		// Generate random password
		password, err := generatePassword(16)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}

		// Create new secret
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: db.Namespace,
			},
			Type: corev1.SecretTypeOpaque,
			StringData: map[string]string{
				"username": db.Spec.Username,
				"password": password,
				"database": db.Spec.DatabaseName,
			},
		}

		// Set owner reference
		if err := ctrl.SetControllerReference(db, secret, r.Scheme); err != nil {
			return err
		}

		logger.Info("Creating Secret", "name", secretName)
		return r.Create(ctx, secret)
	} else if err != nil {
		return err
	}

	// Secret already exists, don't update password
	return nil
}
```

## Exercise 3: Implement StatefulSet Reconciliation

### Task 3.1: Build StatefulSet

Add helper function to build StatefulSet. Note how we reference the password from the Secret:

```go
func (r *DatabaseReconciler) buildStatefulSet(db *databasev1.Database) *appsv1.StatefulSet {
	replicas := int32(1)
	if db.Spec.Replicas != nil {
		replicas = *db.Spec.Replicas
	}

	image := db.Spec.Image
	if image == "" {
		image = "postgres:14"
	}

	secretName := r.secretName(db)

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db.Name,
			Namespace: db.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":      "database",
					"database": db.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":      "database",
						"database": db.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "postgres",
							Image: image,
							Env: []corev1.EnvVar{
								{
									Name:  "POSTGRES_DB",
									Value: db.Spec.DatabaseName,
								},
								{
									Name: "POSTGRES_USER",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: secretName,
											},
											Key: "username",
										},
									},
								},
								{
									Name: "POSTGRES_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: secretName,
											},
											Key: "password",
										},
									},
								},
								{
									Name:  "PGDATA",
									Value: "/var/lib/postgresql/data/pgdata",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/var/lib/postgresql/data",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(db.Spec.Storage.Size),
							},
						},
					},
				},
			},
		},
	}
}
```

### Task 3.2: Reconcile StatefulSet

```go
func (r *DatabaseReconciler) reconcileStatefulSet(ctx context.Context, db *databasev1.Database) error {
	logger := log.FromContext(ctx)

	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      db.Name,
		Namespace: db.Namespace,
	}, statefulSet)

	desiredStatefulSet := r.buildStatefulSet(db)

	if errors.IsNotFound(err) {
		// Set owner reference
		if err := ctrl.SetControllerReference(db, desiredStatefulSet, r.Scheme); err != nil {
			return err
		}
		logger.Info("Creating StatefulSet", "name", desiredStatefulSet.Name)
		return r.Create(ctx, desiredStatefulSet)
	} else if err != nil {
		return err
	}

	// Update if needed
	if statefulSet.Spec.Replicas != desiredStatefulSet.Spec.Replicas ||
		statefulSet.Spec.Template.Spec.Containers[0].Image != desiredStatefulSet.Spec.Template.Spec.Containers[0].Image {
		statefulSet.Spec = desiredStatefulSet.Spec
		logger.Info("Updating StatefulSet", "name", statefulSet.Name)
		return r.Update(ctx, statefulSet)
	}

	return nil
}
```

## Exercise 4: Implement Service Reconciliation

### Task 4.1: Build Service

```go
func (r *DatabaseReconciler) buildService(db *databasev1.Database) *corev1.Service {
    return &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:      db.Name,
            Namespace: db.Namespace,
        },
        Spec: corev1.ServiceSpec{
            Selector: map[string]string{
                "app":      "database",
                "database": db.Name,
            },
            Ports: []corev1.ServicePort{
                {
                    Port: 5432,
                    Name: "postgres",
                },
            },
        },
    }
}
```

### Task 4.2: Reconcile Service

```go
func (r *DatabaseReconciler) reconcileService(ctx context.Context, db *databasev1.Database) error {
    service := &corev1.Service{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, service)
    
    desiredService := r.buildService(db)
    
    if errors.IsNotFound(err) {
        if err := ctrl.SetControllerReference(db, desiredService, r.Scheme); err != nil {
            return err
        }
        return r.Create(ctx, desiredService)
    } else if err != nil {
        return err
    }
    
    // Service updates are less common, but handle if needed
    return nil
}
```

## Exercise 5: Update Status

### Task 5.1: Implement Status Update

The status includes the Secret name so users know where to find credentials:

```go
func (r *DatabaseReconciler) updateStatus(ctx context.Context, db *databasev1.Database) error {
    // Set the secret name in status
    db.Status.SecretName = r.secretName(db)

    // Check StatefulSet status
    statefulSet := &appsv1.StatefulSet{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet)
    
    if err != nil {
        db.Status.Phase = "Pending"
        db.Status.Ready = false
    } else {
        if statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas {
            db.Status.Phase = "Ready"
            db.Status.Ready = true
            db.Status.Endpoint = fmt.Sprintf("%s.%s.svc.cluster.local:5432", db.Name, db.Namespace)
        } else {
            db.Status.Phase = "Creating"
            db.Status.Ready = false
        }
    }
    
    return r.Status().Update(ctx, db)
}
```

## Exercise 6: Set Up the Controller Manager

For the controller to receive events when owned resources change (e.g., when StatefulSet becomes ready), we must tell the manager to watch those resources.

### Task 6.1: Configure Watches

Add the `SetupWithManager` function at the end of your controller:

```go
// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.Database{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
```

**Key Points:**
- `For(&databasev1.Database{})` - Watch Database resources (primary resource)
- `Owns(&appsv1.StatefulSet{})` - Watch StatefulSets owned by Database (via owner reference)
- `Owns(&corev1.Service{})` - Watch Services owned by Database
- `Owns(&corev1.Secret{})` - Watch Secrets owned by Database

This ensures that when a StatefulSet's status changes (pods become ready), the controller is notified and reconciles the parent Database to update its status.

## Exercise 7: Test the Operator

### Task 7.1: Install and Run

```bash
# Install CRD
make install

# Run operator
make run
```

### Task 7.2: Create Database

```bash
# Create Database resource (no password needed - it's auto-generated!)
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: my-database
spec:
  image: postgres:14
  replicas: 1
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF
```

### Task 7.3: Observe Reconciliation

```bash
# Watch Database status
kubectl get database my-database -w

# Check StatefulSet
kubectl get statefulset my-database

# Check Service
kubectl get service my-database

# Check the auto-generated Secret
kubectl get secret my-database-credentials

# View the generated password (base64 decoded)
kubectl get secret my-database-credentials -o jsonpath='{.data.password}' | base64 -d

# Check operator logs
```

## Exercise 8: Test Idempotency

### Task 8.1: Apply Multiple Times

```bash
# Apply the same resource multiple times
for i in {1..3}; do
  kubectl apply -f database.yaml
  sleep 2
done

# Verify only one StatefulSet exists
kubectl get statefulsets | grep my-database
```

### Task 8.2: Test Updates

```bash
# Update replicas
kubectl patch database my-database --type merge -p '{"spec":{"replicas":2}}'

# Verify StatefulSet was updated
kubectl get statefulset my-database -o jsonpath='{.spec.replicas}'
```

## Cleanup

```bash
# Delete Database (should cascade delete StatefulSet, Service, and Secret)
kubectl delete database my-database

# Verify resources were deleted
kubectl get statefulset my-database
kubectl get service my-database
kubectl get secret my-database-credentials
```

## Lab Summary

In this lab, you:
- Implemented complete reconciliation logic
- Created Secret with auto-generated password
- Created StatefulSet and Service
- Used owner references for all resources
- Configured watches with `Owns()` to react to owned resource changes
- Updated status with Secret name
- Tested idempotency
- Verified cascade deletion

## Key Learnings

1. Reconciliation follows: read, compare, create/update, status
2. Owner references ensure cascade deletion
3. **Use `Owns()` to watch owned resources** - without this, the controller won't be notified when StatefulSet/Service/Secret status changes
4. Idempotency is crucial
5. Secrets should be auto-generated, not user-provided in plain text
6. Status updates reflect actual state and provide useful info (like Secret name)
7. Error handling is important
8. Logging helps debugging

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Database Types](../solutions/database-types.go) - Complete Database API type definitions
- [Database Controller](../solutions/database-controller.go) - Complete controller with Secret/StatefulSet/Service reconciliation

## Next Steps

Now let's learn advanced client operations for more sophisticated controllers!

**Navigation:** [← Previous Lab: Designing API](lab-02-designing-api.md) | [Related Lesson](../lessons/03-reconciliation-logic.md) | [Next Lab: Client-Go →](lab-04-client-go.md)
