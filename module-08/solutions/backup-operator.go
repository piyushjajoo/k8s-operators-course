// Solution: Backup Operator from Module 8
// This demonstrates operator composition with backup functionality

package controller

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/api/meta"
	backupv1 "github.com/example/backup-operator/api/v1"
	databasev1 "github.com/example/postgres-operator/api/v1"
)

type BackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *BackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	backup := &backupv1.Backup{}
	if err := r.Get(ctx, req.NamespacedName, backup); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get Database
	db := &databasev1.Database{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      backup.Spec.DatabaseRef.Name,
		Namespace: backup.Namespace,
	}, db)

	if errors.IsNotFound(err) {
		// Database not found, wait
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	if err != nil {
		return ctrl.Result{}, err
	}

	// Check if database is ready
	if db.Status.Phase != "Ready" {
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Perform backup
	return r.performBackup(ctx, db, backup)
}

func (r *BackupReconciler) performBackup(ctx context.Context, db *databasev1.Database, backup *backupv1.Backup) (ctrl.Result, error) {
	// Update status to in progress
	backup.Status.Phase = "InProgress"
	meta.SetStatusCondition(&backup.Status.Conditions, metav1.Condition{
		Type:    "BackupReady",
		Status:  metav1.ConditionFalse,
		Reason:  "BackupInProgress",
		Message: "Backup in progress",
	})
	r.Status().Update(ctx, backup)

	// Perform actual backup (simplified)
	backupLocation, err := r.createBackup(ctx, db, backup)
	if err != nil {
		backup.Status.Phase = "Failed"
		meta.SetStatusCondition(&backup.Status.Conditions, metav1.Condition{
			Type:    "BackupReady",
			Status:  metav1.ConditionFalse,
			Reason:  "BackupFailed",
			Message: err.Error(),
		})
		r.Status().Update(ctx, backup)
		return ctrl.Result{}, err
	}

	// Update status to completed
	backup.Status.Phase = "Completed"
	backup.Status.BackupTime = metav1.Now()
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
		return ctrl.Result{RequeueAfter: 24 * time.Hour}, r.Status().Update(ctx, backup)
	}

	return ctrl.Result{}, r.Status().Update(ctx, backup)
}

func (r *BackupReconciler) createBackup(ctx context.Context, db *databasev1.Database, backup *backupv1.Backup) (string, error) {
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
		For(&backupv1.Backup{}).
		Complete(r)
}

