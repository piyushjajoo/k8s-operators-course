// Solution: Restore Controller from Module 8
// This demonstrates restore functionality for stateful applications.
//
// Use kubebuilder to scaffold the API and controller first (same group as Database):
//   kubebuilder create api --group database --version v1 --kind Restore --resource --controller
//
// Then replace the generated controller with this implementation.

package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	databasev1 "github.com/example/postgres-operator/api/v1"
	restorePkg "github.com/example/postgres-operator/internal/restore"
)

// RestoreReconciler reconciles a Restore object
type RestoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=database.example.com,resources=restores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=restores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=restores/finalizers,verbs=update
// +kubebuilder:rbac:groups=database.example.com,resources=databases,verbs=get;list;watch
// +kubebuilder:rbac:groups=database.example.com,resources=backups,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

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

	// Skip if failed (don't retry failed restores automatically)
	// Users can delete and recreate the Restore resource if they want to retry
	if rst.Status.Phase == "Failed" {
		log.Info("Restore failed, skipping reconciliation", "restore", rst.Name)
		return ctrl.Result{}, nil
	}

	// Get Database
	db := &databasev1.Database{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      rst.Spec.DatabaseRef.Name,
		Namespace: rst.Namespace,
	}, db)
	if errors.IsNotFound(err) {
		log.Info("Database not found, waiting", "database", rst.Spec.DatabaseRef.Name)
		// Re-read restore to ensure we have the latest version
		if getErr := r.Get(ctx, req.NamespacedName, rst); getErr != nil {
			return ctrl.Result{}, getErr
		}
		rst.Status.Phase = "Pending"
		meta.SetStatusCondition(&rst.Status.Conditions, metav1.Condition{
			Type:    "RestoreReady",
			Status:  metav1.ConditionFalse,
			Reason:  "DatabaseNotFound",
			Message: fmt.Sprintf("Waiting for database %s to be created", rst.Spec.DatabaseRef.Name),
		})
		if updateErr := r.Status().Update(ctx, rst); updateErr != nil {
			if errors.IsConflict(updateErr) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	// Check if database is ready
	if db.Status.Phase != "Ready" {
		log.Info("Database not ready, waiting", "database", db.Name, "phase", db.Status.Phase)
		// Re-read restore to ensure we have the latest version
		if getErr := r.Get(ctx, req.NamespacedName, rst); getErr != nil {
			return ctrl.Result{}, getErr
		}
		rst.Status.Phase = "Pending"
		meta.SetStatusCondition(&rst.Status.Conditions, metav1.Condition{
			Type:    "RestoreReady",
			Status:  metav1.ConditionFalse,
			Reason:  "DatabaseNotReady",
			Message: fmt.Sprintf("Waiting for database %s to be ready (current phase: %s)", db.Name, db.Status.Phase),
		})
		if updateErr := r.Status().Update(ctx, rst); updateErr != nil {
			if errors.IsConflict(updateErr) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Get Backup
	backup := &databasev1.Backup{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      rst.Spec.BackupRef.Name,
		Namespace: rst.Namespace,
	}, backup)
	if errors.IsNotFound(err) {
		log.Info("Backup not found, waiting", "backup", rst.Spec.BackupRef.Name)
		// Re-read restore to ensure we have the latest version
		if getErr := r.Get(ctx, req.NamespacedName, rst); getErr != nil {
			return ctrl.Result{}, getErr
		}
		rst.Status.Phase = "Pending"
		meta.SetStatusCondition(&rst.Status.Conditions, metav1.Condition{
			Type:    "RestoreReady",
			Status:  metav1.ConditionFalse,
			Reason:  "BackupNotFound",
			Message: fmt.Sprintf("Waiting for backup %s to be created", rst.Spec.BackupRef.Name),
		})
		if updateErr := r.Status().Update(ctx, rst); updateErr != nil {
			if errors.IsConflict(updateErr) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	// Check backup is completed
	if backup.Status.Phase != "Completed" {
		log.Info("Waiting for backup to complete", "backup", backup.Name, "phase", backup.Status.Phase)
		// Re-read restore to ensure we have the latest version
		if getErr := r.Get(ctx, req.NamespacedName, rst); getErr != nil {
			return ctrl.Result{}, getErr
		}
		rst.Status.Phase = "Pending"
		meta.SetStatusCondition(&rst.Status.Conditions, metav1.Condition{
			Type:    "RestoreReady",
			Status:  metav1.ConditionFalse,
			Reason:  "BackupNotCompleted",
			Message: fmt.Sprintf("Waiting for backup %s to complete (current phase: %s)", backup.Name, backup.Status.Phase),
		})
		if updateErr := r.Status().Update(ctx, rst); updateErr != nil {
			if errors.IsConflict(updateErr) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Perform restore
	return r.performRestore(ctx, req, db, backup, rst)
}

func (r *RestoreReconciler) performRestore(ctx context.Context, req ctrl.Request, db *databasev1.Database, backup *databasev1.Backup, rst *databasev1.Restore) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Re-read restore to ensure we have the latest version before updating
	if err := r.Get(ctx, req.NamespacedName, rst); err != nil {
		return ctrl.Result{}, err
	}

	// Check if already completed
	if rst.Status.Phase == "Completed" {
		log.Info("Restore already completed, skipping", "restore", rst.Name)
		return ctrl.Result{}, nil
	}
	// Note: We don't skip InProgress here because the restore operation might have completed
	// but failed to update status due to a conflict. We'll proceed to check/retry.

	// Update status to in progress (only if not already InProgress)
	if rst.Status.Phase != "InProgress" {
		rst.Status.Phase = "InProgress"
		meta.SetStatusCondition(&rst.Status.Conditions, metav1.Condition{
			Type:    "RestoreReady",
			Status:  metav1.ConditionFalse,
			Reason:  "RestoreInProgress",
			Message: "Restore in progress",
		})
		if err := r.Status().Update(ctx, rst); err != nil {
			if errors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
		// Re-read after updating to get latest version
		if err := r.Get(ctx, req.NamespacedName, rst); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Get backup location from Backup status
	if backup.Status.BackupLocation == "" {
		err := fmt.Errorf("backup location not available in backup %s", backup.Name)
		// Re-read restore before updating status on error
		if getErr := r.Get(ctx, req.NamespacedName, rst); getErr != nil {
			return ctrl.Result{}, getErr
		}
		rst.Status.Phase = "Failed"
		meta.SetStatusCondition(&rst.Status.Conditions, metav1.Condition{
			Type:    "RestoreReady",
			Status:  metav1.ConditionFalse,
			Reason:  "BackupLocationMissing",
			Message: err.Error(),
		})
		if updateErr := r.Status().Update(ctx, rst); updateErr != nil {
			if errors.IsConflict(updateErr) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{}, err
	}

	// Perform actual restore using restore package
	// Note: PerformRestore requires k8sClient to retrieve password from Secret
	err := restorePkg.PerformRestore(ctx, r.Client, db, backup.Status.BackupLocation)
	if err != nil {
		log.Error(err, "Restore failed", "database", db.Name, "backup", backup.Name)
		// Re-read restore before updating status on error
		if getErr := r.Get(ctx, req.NamespacedName, rst); getErr != nil {
			return ctrl.Result{}, getErr
		}
		rst.Status.Phase = "Failed"
		meta.SetStatusCondition(&rst.Status.Conditions, metav1.Condition{
			Type:    "RestoreReady",
			Status:  metav1.ConditionFalse,
			Reason:  "RestoreFailed",
			Message: err.Error(),
		})
		if updateErr := r.Status().Update(ctx, rst); updateErr != nil {
			if errors.IsConflict(updateErr) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{}, err
	}

	// Re-read restore before final status update
	if err := r.Get(ctx, req.NamespacedName, rst); err != nil {
		return ctrl.Result{}, err
	}

	// Update status to completed
	rst.Status.Phase = "Completed"
	now := metav1.Now()
	rst.Status.RestoreTime = &now
	meta.SetStatusCondition(&rst.Status.Conditions, metav1.Condition{
		Type:    "RestoreReady",
		Status:  metav1.ConditionTrue,
		Reason:  "RestoreCompleted",
		Message: "Restore completed successfully",
	})

	log.Info("Restore completed", "database", db.Name, "backup", backup.Name)
	if err := r.Status().Update(ctx, rst); err != nil {
		if errors.IsConflict(err) {
			log.Info("Conflict updating restore status, requeuing", "restore", rst.Name)
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.Restore{}).
		Complete(r)
}

