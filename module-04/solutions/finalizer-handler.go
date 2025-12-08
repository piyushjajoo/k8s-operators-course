// Solution: Finalizer Handler from Module 4
// This implements graceful cleanup with finalizers
//
// Key concept: When using finalizers, you must EXPLICITLY delete child resources.
// Owner references only cascade deletes AFTER the parent is deleted, but finalizers
// prevent the parent from being deleted until cleanup completes - causing a deadlock
// if you only wait for resources to disappear.

package controller

import (
    "context"
    "fmt"
    "time"

    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/api/meta"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    "sigs.k8s.io/controller-runtime/pkg/log"
    ctrl "sigs.k8s.io/controller-runtime"

    appsv1 "k8s.io/api/apps/v1"
    databasev1 "github.com/example/postgres-operator/api/v1"
)

const finalizerName = "database.example.com/finalizer"

// Add finalizer in Reconcile:
//
// // Add finalizer if not present
// if !controllerutil.ContainsFinalizer(db, finalizerName) {
//     controllerutil.AddFinalizer(db, finalizerName)
//     if err := r.Update(ctx, db); err != nil {
//         return ctrl.Result{}, err
//     }
// }
//
// // Check if resource is being deleted
// if !db.DeletionTimestamp.IsZero() {
//     return r.handleDeletion(ctx, db)
// }

// handleDeletion performs cleanup before removing finalizer
func (r *DatabaseReconciler) handleDeletion(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // Check if finalizer exists
    if !controllerutil.ContainsFinalizer(db, finalizerName) {
        return ctrl.Result{}, nil
    }

    logger.Info("Handling deletion", "name", db.Name)

    // Perform cleanup operations
    if err := r.cleanupExternalResources(ctx, db); err != nil {
        logger.Error(err, "Failed to cleanup external resources")
        // Update condition if setCondition method exists
        condition := metav1.Condition{
            Type:               "Ready",
            Status:             metav1.ConditionFalse,
            Reason:             "CleanupFailed",
            Message:            err.Error(),
            LastTransitionTime: metav1.Now(),
            ObservedGeneration: db.Generation,
        }
        meta.SetStatusCondition(&db.Status.Conditions, condition)
        r.Status().Update(ctx, db)
        // Retry after delay
        return ctrl.Result{RequeueAfter: 10 * time.Second}, err
    }

    // Cleanup successful, remove finalizer
    controllerutil.RemoveFinalizer(db, finalizerName)
    if err := r.Update(ctx, db); err != nil {
        return ctrl.Result{}, err
    }

    logger.Info("Finalizer removed, resource will be deleted")
    return ctrl.Result{}, nil
}

// cleanupExternalResources performs actual cleanup
// Important: We must explicitly delete child resources. While owner references
// enable automatic garbage collection when a parent is deleted, finalizers
// prevent the parent from being deleted until cleanup completes. This creates
// a deadlock if you only wait for resources to disappear - you must actively
// delete them.
func (r *DatabaseReconciler) cleanupExternalResources(ctx context.Context, db *databasev1.Database) error {
    logger := log.FromContext(ctx)

    // Delete StatefulSet if it exists
    statefulSet := &appsv1.StatefulSet{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet)

<<<<<<< Updated upstream
    if !errors.IsNotFound(err) {
        // StatefulSet still exists, wait for owner reference to delete it
        logger.Info("Waiting for StatefulSet to be deleted")
        return fmt.Errorf("StatefulSet still exists")
=======
    if err == nil {
        // StatefulSet exists, delete it explicitly
        log.Info("Deleting StatefulSet", "name", statefulSet.Name)
        if err := r.Delete(ctx, statefulSet); err != nil && !errors.IsNotFound(err) {
            return fmt.Errorf("failed to delete StatefulSet: %w", err)
        }
        // Requeue to wait for deletion to complete
        return fmt.Errorf("waiting for StatefulSet to be deleted")
    } else if !errors.IsNotFound(err) {
        return fmt.Errorf("failed to get StatefulSet: %w", err)
>>>>>>> Stashed changes
    }

    // Add other cleanup operations here:
    // - Delete Services, Secrets, ConfigMaps
    // - Delete backups in external system
    // - Notify external services
    // - Clean up external resources

    logger.Info("Cleanup completed")
    return nil
}

