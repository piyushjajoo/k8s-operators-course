// Solution: Operator Coordination from Module 8
// This demonstrates how operators coordinate through resources and conditions

package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	databasev1 "github.com/example/postgres-operator/api/v1"
)

// Example 1: Database operator checks backup status
func (r *DatabaseReconciler) checkBackupStatus(ctx context.Context, db *databasev1.Database) error {
	if db.Spec.BackupRef == nil {
		return nil
	}

	backup := &databasev1.Backup{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      db.Spec.BackupRef.Name,
		Namespace: db.Namespace,
	}, backup)

	if errors.IsNotFound(err) {
		return fmt.Errorf("backup %s not found", db.Spec.BackupRef.Name)
	}

	if err != nil {
		return err
	}

	// Check backup condition
	condition := meta.FindStatusCondition(backup.Status.Conditions, "BackupReady")
	if condition == nil || condition.Status != metav1.ConditionTrue {
		return fmt.Errorf("backup not ready: %s", condition.Message)
	}

	return nil
}

// Example 2: Database operator waits for backup
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	db := &databasev1.Database{}
	if err := r.Get(ctx, req.NamespacedName, db); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if backup is required
	if db.Spec.BackupRef != nil {
		if err := r.checkBackupStatus(ctx, db); err != nil {
			// Backup not ready, requeue
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
	}

	// Continue with database reconciliation
	return r.reconcileDatabase(ctx, db)
}

// Example 3: Emit events for coordination
func (r *DatabaseReconciler) emitCoordinationEvent(db *databasev1.Database, eventType, reason, message string) {
	r.Recorder.Event(db, eventType, reason, message)
}

// Example 4: Update status for other operators to read
func (r *DatabaseReconciler) updateStatusForCoordination(ctx context.Context, db *databasev1.Database) error {
	// Set condition that other operators can check
	meta.SetStatusCondition(&db.Status.Conditions, metav1.Condition{
		Type:    "ReadyForBackup",
		Status:  metav1.ConditionTrue,
		Reason:  "DatabaseReady",
		Message: "Database is ready for backup",
	})

	return r.Status().Update(ctx, db)
}

