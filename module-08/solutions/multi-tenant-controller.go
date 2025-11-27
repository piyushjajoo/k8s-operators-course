// Solution: Multi-Tenant Controller from Module 8
// This demonstrates multi-tenant operator patterns

package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	databasev1 "github.com/example/postgres-operator/api/v1"
)

type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	db := &databasev1.Database{}
	if err := r.Get(ctx, req.NamespacedName, db); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Extract tenant from spec or namespace
	tenant := db.Spec.Tenant
	if tenant == "" {
		// For namespaced resources, use namespace as tenant
		tenant = req.Namespace
	}

	// Check resource quota for tenant
	if err := r.checkQuota(ctx, tenant); err != nil {
		return ctrl.Result{}, err
	}

	// Apply tenant-specific logic
	return r.reconcileForTenant(ctx, db, tenant)
}

func (r *DatabaseReconciler) checkQuota(ctx context.Context, namespace string) error {
	quota := &corev1.ResourceQuota{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      "database-quota",
		Namespace: namespace,
	}, quota)

	if errors.IsNotFound(err) {
		// No quota, proceed
		return nil
	}

	if err != nil {
		return err
	}

	// Count existing databases for this tenant
	databases := &databasev1.DatabaseList{}
	err = r.List(ctx, databases, client.InNamespace(namespace))
	if err != nil {
		return err
	}

	used := int64(len(databases.Items))
	hard, exists := quota.Spec.Hard["databases.database.example.com"]
	if !exists {
		// No quota limit for databases
		return nil
	}

	if hard.Value() <= used {
		return fmt.Errorf("quota exceeded: %d/%d databases", used, hard.Value())
	}

	return nil
}

func (r *DatabaseReconciler) reconcileForTenant(ctx context.Context, db *databasev1.Database, tenant string) (ctrl.Result, error) {
	// Apply tenant-specific labels
	if db.Labels == nil {
		db.Labels = make(map[string]string)
	}
	db.Labels["tenant"] = tenant

	// Apply tenant-specific configuration
	// For example, different resource limits per tenant
	if tenant == "production" {
		// Production tenant gets more resources
	} else if tenant == "development" {
		// Development tenant gets fewer resources
	}

	// Continue with normal reconciliation
	return r.reconcileDatabase(ctx, db)
}

func (r *DatabaseReconciler) reconcileDatabase(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
	// Normal database reconciliation logic
	// ... create StatefulSet, Service, etc. ...
	return ctrl.Result{}, nil
}

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.Database{}).
		Complete(r)
}

