// Solution: ClusterDatabase Controller from Module 8
// This implements a cluster-scoped controller for managing databases across namespaces
// Unlike the namespace-scoped DatabaseReconciler, this controller:
// 1. Watches cluster-scoped ClusterDatabase resources
// 2. Creates resources in the specified targetNamespace
// 3. Supports multi-tenant patterns with tenant isolation

package controller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	databasev1 "github.com/example/postgres-operator/api/v1"
)

// ClusterDatabaseReconciler reconciles a ClusterDatabase object
type ClusterDatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=database.example.com,resources=clusterdatabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=clusterdatabases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=clusterdatabases/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=resourcequotas,verbs=get;list;watch

func (r *ClusterDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Read ClusterDatabase resource (cluster-scoped, so no namespace in the request)
	db := &databasev1.ClusterDatabase{}
	if err := r.Get(ctx, req.NamespacedName, db); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling ClusterDatabase",
		"name", db.Name,
		"targetNamespace", db.Spec.TargetNamespace,
		"tenant", db.Spec.Tenant)

	// Validate target namespace exists
	if err := r.validateNamespace(ctx, db.Spec.TargetNamespace); err != nil {
		return ctrl.Result{}, err
	}

	// Check resource quota for the target namespace
	if err := r.checkQuota(ctx, db.Spec.TargetNamespace); err != nil {
		logger.Error(err, "Quota check failed", "namespace", db.Spec.TargetNamespace)
		return ctrl.Result{}, err
	}

	// Reconcile resources in the target namespace
	if err := r.reconcileSecret(ctx, db); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileStatefulSet(ctx, db); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileService(ctx, db); err != nil {
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateStatus(ctx, db); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// validateNamespace checks if the target namespace exists
func (r *ClusterDatabaseReconciler) validateNamespace(ctx context.Context, namespace string) error {
	ns := &corev1.Namespace{}
	err := r.Get(ctx, client.ObjectKey{Name: namespace}, ns)
	if errors.IsNotFound(err) {
		return fmt.Errorf("target namespace %s does not exist", namespace)
	}
	return err
}

// checkQuota verifies quota limits for the namespace
func (r *ClusterDatabaseReconciler) checkQuota(ctx context.Context, namespace string) error {
	quota := &corev1.ResourceQuota{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      "database-quota",
		Namespace: namespace,
	}, quota)

	if errors.IsNotFound(err) {
		// No quota defined, proceed
		return nil
	}
	if err != nil {
		return err
	}

	// Count existing ClusterDatabases targeting this namespace
	databases := &databasev1.ClusterDatabaseList{}
	if err := r.List(ctx, databases); err != nil {
		return err
	}

	// Count databases targeting this namespace
	var count int64
	for _, db := range databases.Items {
		if db.Spec.TargetNamespace == namespace {
			count++
		}
	}

	hard, exists := quota.Spec.Hard["clusterdatabases.database.example.com"]
	if !exists {
		// No quota limit for clusterdatabases
		return nil
	}

	if hard.Value() <= count {
		return fmt.Errorf("quota exceeded: %d/%d clusterdatabases in namespace %s",
			count, hard.Value(), namespace)
	}

	return nil
}

// secretName returns the name of the Secret for this ClusterDatabase
func (r *ClusterDatabaseReconciler) secretName(db *databasev1.ClusterDatabase) string {
	return fmt.Sprintf("%s-credentials", db.Name)
}

// generatePassword generates a random password
// NOTE: if you have database_controller.go from earlier labs, this function already exists, you will see compiler errors, just deletion this duplicate function
func generatePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// reconcileSecret ensures the credentials Secret exists in the target namespace
func (r *ClusterDatabaseReconciler) reconcileSecret(ctx context.Context, db *databasev1.ClusterDatabase) error {
	logger := log.FromContext(ctx)
	secretName := r.secretName(db)

	secret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: db.Spec.TargetNamespace, // Create in target namespace
	}, secret)

	if errors.IsNotFound(err) {
		password, err := generatePassword(16)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}

		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: db.Spec.TargetNamespace,
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "clusterdatabase-controller",
					"clusterdatabase":              db.Name,
					"tenant":                       db.Spec.Tenant,
				},
			},
			Type: corev1.SecretTypeOpaque,
			StringData: map[string]string{
				"username": db.Spec.Username,
				"password": password,
				"database": db.Spec.DatabaseName,
			},
		}

		// Note: Cannot use SetControllerReference for cluster-scoped owner
		// and namespaced resource. Use labels for tracking instead.

		logger.Info("Creating Secret", "name", secretName, "namespace", db.Spec.TargetNamespace)
		return r.Create(ctx, secret)
	}

	return err
}

func (r *ClusterDatabaseReconciler) buildStatefulSet(db *databasev1.ClusterDatabase) *appsv1.StatefulSet {
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
			Namespace: db.Spec.TargetNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "clusterdatabase-controller",
				"clusterdatabase":              db.Name,
				"tenant":                       db.Spec.Tenant,
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":             "clusterdatabase",
					"clusterdatabase": db.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":             "clusterdatabase",
						"clusterdatabase": db.Name,
						"tenant":          db.Spec.Tenant,
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

func (r *ClusterDatabaseReconciler) reconcileStatefulSet(ctx context.Context, db *databasev1.ClusterDatabase) error {
	logger := log.FromContext(ctx)

	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      db.Name,
		Namespace: db.Spec.TargetNamespace,
	}, statefulSet)

	desiredStatefulSet := r.buildStatefulSet(db)

	if errors.IsNotFound(err) {
		logger.Info("Creating StatefulSet",
			"name", desiredStatefulSet.Name,
			"namespace", db.Spec.TargetNamespace)
		return r.Create(ctx, desiredStatefulSet)
	} else if err != nil {
		return err
	}

	// Update if needed
	if *statefulSet.Spec.Replicas != *desiredStatefulSet.Spec.Replicas ||
		statefulSet.Spec.Template.Spec.Containers[0].Image != desiredStatefulSet.Spec.Template.Spec.Containers[0].Image {
		statefulSet.Spec.Replicas = desiredStatefulSet.Spec.Replicas
		statefulSet.Spec.Template.Spec.Containers[0].Image = desiredStatefulSet.Spec.Template.Spec.Containers[0].Image
		logger.Info("Updating StatefulSet", "name", statefulSet.Name)
		return r.Update(ctx, statefulSet)
	}

	return nil
}

func (r *ClusterDatabaseReconciler) buildService(db *databasev1.ClusterDatabase) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db.Name,
			Namespace: db.Spec.TargetNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "clusterdatabase-controller",
				"clusterdatabase":              db.Name,
				"tenant":                       db.Spec.Tenant,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":             "clusterdatabase",
				"clusterdatabase": db.Name,
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

func (r *ClusterDatabaseReconciler) reconcileService(ctx context.Context, db *databasev1.ClusterDatabase) error {
	logger := log.FromContext(ctx)

	service := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      db.Name,
		Namespace: db.Spec.TargetNamespace,
	}, service)

	desiredService := r.buildService(db)

	if errors.IsNotFound(err) {
		logger.Info("Creating Service",
			"name", desiredService.Name,
			"namespace", db.Spec.TargetNamespace)
		return r.Create(ctx, desiredService)
	}

	return err
}

func (r *ClusterDatabaseReconciler) updateStatus(ctx context.Context, db *databasev1.ClusterDatabase) error {
	// Set namespace and secret in status
	db.Status.TargetNamespace = db.Spec.TargetNamespace
	db.Status.SecretName = r.secretName(db)

	// Check StatefulSet status
	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      db.Name,
		Namespace: db.Spec.TargetNamespace,
	}, statefulSet)

	if err != nil {
		db.Status.Phase = "Pending"
		db.Status.Ready = false
	} else {
		if statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas {
			db.Status.Phase = "Ready"
			db.Status.Ready = true
			db.Status.Endpoint = fmt.Sprintf("%s.%s.svc.cluster.local:5432",
				db.Name, db.Spec.TargetNamespace)
		} else {
			db.Status.Phase = "Creating"
			db.Status.Ready = false
		}
	}

	return r.Status().Update(ctx, db)
}

// ListClusterDatabasesByTenant returns all ClusterDatabases for a specific tenant
func (r *ClusterDatabaseReconciler) ListClusterDatabasesByTenant(ctx context.Context, tenant string) (*databasev1.ClusterDatabaseList, error) {
	list := &databasev1.ClusterDatabaseList{}
	// For cluster-scoped resources, we don't filter by namespace
	if err := r.List(ctx, list); err != nil {
		return nil, err
	}

	// Filter by tenant
	filtered := &databasev1.ClusterDatabaseList{}
	for _, db := range list.Items {
		if db.Spec.Tenant == tenant {
			filtered.Items = append(filtered.Items, db)
		}
	}
	return filtered, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *ClusterDatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.ClusterDatabase{}).
		// Note: We don't use Owns() here because cluster-scoped resources
		// cannot own namespaced resources directly. Instead, we use labels
		// to track managed resources.
		Complete(r)
}

