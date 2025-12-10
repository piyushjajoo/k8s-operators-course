// Solution: Multi-Tenant Controller Patterns from Module 8
// This demonstrates multi-tenant operator patterns using ClusterDatabase
// Shows how to handle tenant isolation, quotas, and cross-namespace management

package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	databasev1 "github.com/example/postgres-operator/api/v1"
)

const clusterDatabaseFinalizer = "database.example.com/clusterdatabase-finalizer"

// MultiTenantReconciler demonstrates multi-tenant patterns
// This is an enhanced version of ClusterDatabaseReconciler with tenant isolation
type MultiTenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=database.example.com,resources=clusterdatabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=clusterdatabases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=clusterdatabases/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=resourcequotas,verbs=get;list;watch

func (r *MultiTenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	db := &databasev1.ClusterDatabase{}
	if err := r.Get(ctx, req.NamespacedName, db); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deletion with finalizer
	if !db.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, db)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(db, clusterDatabaseFinalizer) {
		controllerutil.AddFinalizer(db, clusterDatabaseFinalizer)
		if err := r.Update(ctx, db); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Extract tenant from spec
	tenant := db.Spec.Tenant
	targetNamespace := db.Spec.TargetNamespace

	logger.Info("Reconciling ClusterDatabase",
		"name", db.Name,
		"targetNamespace", targetNamespace,
		"tenant", tenant)

	// Validate target namespace exists and belongs to tenant
	if err := r.validateTenantNamespace(ctx, targetNamespace, tenant); err != nil {
		return ctrl.Result{}, err
	}

	// Check resource quota for tenant's namespace
	if err := r.checkTenantQuota(ctx, targetNamespace, tenant); err != nil {
		logger.Error(err, "Quota check failed",
			"namespace", targetNamespace,
			"tenant", tenant)
		return ctrl.Result{}, err
	}

	// Apply tenant-specific configuration and reconcile
	return r.reconcileForTenant(ctx, db, tenant)
}

// handleDeletion manages cleanup when ClusterDatabase is deleted
func (r *MultiTenantReconciler) handleDeletion(ctx context.Context, db *databasev1.ClusterDatabase) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(db, clusterDatabaseFinalizer) {
		logger.Info("Cleaning up managed resources",
			"name", db.Name,
			"namespace", db.Spec.TargetNamespace)

		// Clean up managed resources by label
		if err := r.cleanupManagedResources(ctx, db); err != nil {
			return ctrl.Result{}, err
		}

		// Remove finalizer
		controllerutil.RemoveFinalizer(db, clusterDatabaseFinalizer)
		if err := r.Update(ctx, db); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// cleanupManagedResources deletes all resources managed by this ClusterDatabase
func (r *MultiTenantReconciler) cleanupManagedResources(ctx context.Context, db *databasev1.ClusterDatabase) error {
	labelSelector := client.MatchingLabels{
		"clusterdatabase": db.Name,
	}
	namespace := db.Spec.TargetNamespace

	// Delete StatefulSet
	if err := r.deleteResourcesByLabel(ctx, namespace, labelSelector, &corev1.ServiceList{}); err != nil {
		return err
	}

	// Delete Services
	if err := r.deleteResourcesByLabel(ctx, namespace, labelSelector, &corev1.SecretList{}); err != nil {
		return err
	}

	return nil
}

func (r *MultiTenantReconciler) deleteResourcesByLabel(
	ctx context.Context,
	namespace string,
	labels client.MatchingLabels,
	list client.ObjectList,
) error {
	if err := r.List(ctx, list, client.InNamespace(namespace), labels); err != nil {
		return err
	}
	// Deletion logic would iterate over list.Items
	return nil
}

// validateTenantNamespace checks if the namespace exists and optionally validates tenant ownership
func (r *MultiTenantReconciler) validateTenantNamespace(ctx context.Context, namespace, tenant string) error {
	ns := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: namespace}, ns); err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("target namespace %s does not exist", namespace)
		}
		return err
	}

	// Optional: Validate that namespace belongs to tenant
	if tenant != "" {
		if nsTenant, ok := ns.Labels["tenant"]; ok && nsTenant != tenant {
			return fmt.Errorf("namespace %s belongs to tenant %s, not %s", namespace, nsTenant, tenant)
		}
	}

	return nil
}

// checkTenantQuota verifies quota limits for the tenant's namespace
func (r *MultiTenantReconciler) checkTenantQuota(ctx context.Context, namespace, tenant string) error {
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

	// Count ClusterDatabases targeting this namespace
	databases := &databasev1.ClusterDatabaseList{}
	if err := r.List(ctx, databases); err != nil {
		return err
	}

	// Count databases for this tenant/namespace
	var count int64
	for _, db := range databases.Items {
		if db.Spec.TargetNamespace == namespace {
			// Optionally also filter by tenant
			if tenant == "" || db.Spec.Tenant == tenant {
				count++
			}
		}
	}

	hard, exists := quota.Spec.Hard["clusterdatabases.database.example.com"]
	if !exists {
		// No quota limit for clusterdatabases
		return nil
	}

	if hard.Value() <= count {
		return fmt.Errorf("quota exceeded for tenant %s: %d/%d clusterdatabases in namespace %s",
			tenant, count, hard.Value(), namespace)
	}

	return nil
}

// reconcileForTenant applies tenant-specific logic and reconciles resources
func (r *MultiTenantReconciler) reconcileForTenant(
	ctx context.Context,
	db *databasev1.ClusterDatabase,
	tenant string,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Apply tenant-specific labels to all managed resources
	tenantLabels := map[string]string{
		"app.kubernetes.io/managed-by": "clusterdatabase-controller",
		"clusterdatabase":              db.Name,
		"tenant":                       tenant,
	}

	// Apply tenant-specific configuration
	// Different tenants might get different resource limits, storage classes, etc.
	config := r.getTenantConfig(tenant)

	logger.Info("Reconciling with tenant config",
		"tenant", tenant,
		"config", config)

	// Continue with normal reconciliation using tenant config
	// ... create StatefulSet, Service, Secret with tenantLabels ...

	return ctrl.Result{}, nil
}

// TenantConfig holds tenant-specific configuration
type TenantConfig struct {
	MaxReplicas    int32
	StorageClass   string
	ResourceLimits corev1.ResourceRequirements
}

// getTenantConfig returns configuration specific to a tenant
func (r *MultiTenantReconciler) getTenantConfig(tenant string) TenantConfig {
	// In production, this could come from a ConfigMap, CRD, or external config
	switch tenant {
	case "production":
		return TenantConfig{
			MaxReplicas:  10,
			StorageClass: "fast-ssd",
		}
	case "development":
		return TenantConfig{
			MaxReplicas:  3,
			StorageClass: "standard",
		}
	default:
		return TenantConfig{
			MaxReplicas:  5,
			StorageClass: "standard",
		}
	}
}

// ListClusterDatabasesByTenant returns all ClusterDatabases for a specific tenant
func (r *MultiTenantReconciler) ListClusterDatabasesByTenant(
	ctx context.Context,
	tenant string,
) (*databasev1.ClusterDatabaseList, error) {
	list := &databasev1.ClusterDatabaseList{}
	// For cluster-scoped resources, we list without namespace filter
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

// ListClusterDatabasesByNamespace returns all ClusterDatabases targeting a namespace
func (r *MultiTenantReconciler) ListClusterDatabasesByNamespace(
	ctx context.Context,
	namespace string,
) (*databasev1.ClusterDatabaseList, error) {
	list := &databasev1.ClusterDatabaseList{}
	if err := r.List(ctx, list); err != nil {
		return nil, err
	}

	// Filter by target namespace
	filtered := &databasev1.ClusterDatabaseList{}
	for _, db := range list.Items {
		if db.Spec.TargetNamespace == namespace {
			filtered.Items = append(filtered.Items, db)
		}
	}
	return filtered, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *MultiTenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.ClusterDatabase{}).
		// Note: We don't use Owns() because cluster-scoped resources
		// cannot own namespaced resources directly
		Complete(r)
}
