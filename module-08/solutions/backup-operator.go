// Solution: Backup Operator from Module 8
// This demonstrates operator composition with backup functionality.
//
// Use kubebuilder to scaffold the API and controller first (same group as Database):
//   kubebuilder create api --group database --version v1 --kind Backup --resource --controller
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
)

type BackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=database.example.com,resources=backups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=backups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=backups/finalizers,verbs=update
// +kubebuilder:rbac:groups=database.example.com,resources=databases,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile handles Backup resources
func (r *BackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	backup := &databasev1.Backup{}
	if err := r.Get(ctx, req.NamespacedName, backup); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Skip if already completed
	if backup.Status.Phase == "Completed" {
		return ctrl.Result{}, nil
	}

	// Skip if already in progress (another reconciliation is handling it)
	if backup.Status.Phase == "InProgress" {
		log.Info("Backup already in progress, skipping", "backup", backup.Name)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Get Database
	db := &databasev1.Database{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      backup.Spec.DatabaseRef.Name,
		Namespace: backup.Namespace,
	}, db)

	if errors.IsNotFound(err) {
		// Database not found, set pending status and wait
		log.Info("Database not found, waiting", "database", backup.Spec.DatabaseRef.Name)
		// Re-read backup to ensure we have the latest version
		if getErr := r.Get(ctx, req.NamespacedName, backup); getErr != nil {
			return ctrl.Result{}, getErr
		}
		backup.Status.Phase = "Pending"
		meta.SetStatusCondition(&backup.Status.Conditions, metav1.Condition{
			Type:    "BackupReady",
			Status:  metav1.ConditionFalse,
			Reason:  "DatabaseNotFound",
			Message: fmt.Sprintf("Waiting for database %s to be created", backup.Spec.DatabaseRef.Name),
		})
		if updateErr := r.Status().Update(ctx, backup); updateErr != nil {
			if errors.IsConflict(updateErr) {
				// Resource was modified, requeue to retry
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
		// Re-read backup to ensure we have the latest version
		if getErr := r.Get(ctx, req.NamespacedName, backup); getErr != nil {
			return ctrl.Result{}, getErr
		}
		backup.Status.Phase = "Pending"
		meta.SetStatusCondition(&backup.Status.Conditions, metav1.Condition{
			Type:    "BackupReady",
			Status:  metav1.ConditionFalse,
			Reason:  "DatabaseNotReady",
			Message: fmt.Sprintf("Waiting for database %s to be ready (current phase: %s)", db.Name, db.Status.Phase),
		})
		if updateErr := r.Status().Update(ctx, backup); updateErr != nil {
			if errors.IsConflict(updateErr) {
				// Resource was modified, requeue to retry
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Perform backup
	return r.performBackup(ctx, req, db, backup)
}

func (r *BackupReconciler) performBackup(ctx context.Context, req ctrl.Request, db *databasev1.Database, backup *databasev1.Backup) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Re-read backup to ensure we have the latest version before updating
	if err := r.Get(ctx, req.NamespacedName, backup); err != nil {
		return ctrl.Result{}, err
	}

	// Check if already completed or in progress (another reconciliation might have updated it)
	if backup.Status.Phase == "Completed" {
		log.Info("Backup already completed, skipping", "backup", backup.Name)
		return ctrl.Result{}, nil
	}
	if backup.Status.Phase == "InProgress" {
		log.Info("Backup already in progress, skipping", "backup", backup.Name)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Update status to in progress
	backup.Status.Phase = "InProgress"
	meta.SetStatusCondition(&backup.Status.Conditions, metav1.Condition{
		Type:    "BackupReady",
		Status:  metav1.ConditionFalse,
		Reason:  "BackupInProgress",
		Message: "Backup in progress",
	})
	if err := r.Status().Update(ctx, backup); err != nil {
		if errors.IsConflict(err) {
			// Resource was modified, requeue to retry
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	// Perform actual backup (simplified)
	backupLocation, err := r.createBackup(ctx, db, backup)
	if err != nil {
		// Re-read backup before updating status on error
		if getErr := r.Get(ctx, req.NamespacedName, backup); getErr != nil {
			return ctrl.Result{}, getErr
		}
		backup.Status.Phase = "Failed"
		meta.SetStatusCondition(&backup.Status.Conditions, metav1.Condition{
			Type:    "BackupReady",
			Status:  metav1.ConditionFalse,
			Reason:  "BackupFailed",
			Message: err.Error(),
		})
		if updateErr := r.Status().Update(ctx, backup); updateErr != nil {
			if errors.IsConflict(updateErr) {
				// Resource was modified, requeue to retry
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{}, err
	}

	// Re-read backup before final status update
	if err := r.Get(ctx, req.NamespacedName, backup); err != nil {
		return ctrl.Result{}, err
	}

	// Update status to completed
	backup.Status.Phase = "Completed"
	now := metav1.Now()
	backup.Status.BackupTime = &now
	backup.Status.BackupLocation = backupLocation
	meta.SetStatusCondition(&backup.Status.Conditions, metav1.Condition{
		Type:    "BackupReady",
		Status:  metav1.ConditionTrue,
		Reason:  "BackupCompleted",
		Message: "Backup completed successfully",
	})

	// Handle scheduled backups
	if backup.Spec.Schedule != "" {
		// For scheduled backups, requeue after interval
		// In production, you'd parse cron schedule and calculate next time
		if updateErr := r.Status().Update(ctx, backup); updateErr != nil {
			if errors.IsConflict(updateErr) {
				log.Info("Conflict updating backup status, requeuing", "backup", backup.Name)
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{RequeueAfter: 24 * time.Hour}, nil
	}

	if err := r.Status().Update(ctx, backup); err != nil {
		if errors.IsConflict(err) {
			log.Info("Conflict updating backup status, requeuing", "backup", backup.Name)
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

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

func (r *BackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.Backup{}).
		Complete(r)
}

