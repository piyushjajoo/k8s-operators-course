// Solution: Complete Database Controller from Module 3
// This implements reconciliation logic for the PostgreSQL operator

package controller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

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

	if statefulSet.Spec.Replicas != desiredStatefulSet.Spec.Replicas {
		return r.patchStatefulSetReplicas(ctx, statefulSet, *desiredStatefulSet.Spec.Replicas)
	}

	// Update if needed
	if statefulSet.Spec.Replicas != desiredStatefulSet.Spec.Replicas ||
		statefulSet.Spec.Template.Spec.Containers[0].Image != desiredStatefulSet.Spec.Template.Spec.Containers[0].Image {
		statefulSet.Spec = desiredStatefulSet.Spec
		logger.Info("Updating StatefulSet", "name", statefulSet.Name)
		// Use retry logic for updates
		return r.updateWithRetry(ctx, statefulSet, 3)
	}

	return nil
}

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

func (r *DatabaseReconciler) listDatabasesInNamespace(ctx context.Context, namespace string) (*databasev1.DatabaseList, error) {
	list := &databasev1.DatabaseList{}
	err := r.List(ctx, list, client.InNamespace(namespace))
	return list, err
}

func (r *DatabaseReconciler) listDatabasesByLabel(ctx context.Context, labels map[string]string) (*databasev1.DatabaseList, error) {
	list := &databasev1.DatabaseList{}
	err := r.List(ctx, list, client.MatchingLabels(labels))
	return list, err
}

func (r *DatabaseReconciler) findDatabasesByOwner(ctx context.Context, ownerName string) (*databasev1.DatabaseList, error) {
	list := &databasev1.DatabaseList{}
	err := r.List(ctx, list, client.MatchingFields{
		".metadata.ownerReferences[0].name": ownerName,
	})
	return list, err
}

func (r *DatabaseReconciler) patchStatefulSetReplicas(ctx context.Context, statefulSet *appsv1.StatefulSet, replicas int32) error {
	patch := client.MergeFrom(statefulSet.DeepCopy())
	statefulSet.Spec.Replicas = &replicas
	return r.Patch(ctx, statefulSet, patch)
}

func (r *DatabaseReconciler) updateWithRetry(ctx context.Context, obj client.Object, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		err := r.Update(ctx, obj)
		if err == nil {
			return nil
		}

		if !errors.IsConflict(err) {
			return err
		}

		// Conflict - get fresh version and retry
		key := client.ObjectKeyFromObject(obj)
		if err := r.Get(ctx, key, obj); err != nil {
			return err
		}

		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("max retries exceeded")
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.Database{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
