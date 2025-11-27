// Solution: Finalizer Handler from Module 4
// This implements graceful cleanup with finalizers

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
    log := log.FromContext(ctx)

    // Check if finalizer exists
    if !controllerutil.ContainsFinalizer(db, finalizerName) {
        return ctrl.Result{}, nil
    }

    log.Info("Handling deletion", "name", db.Name)

    // Perform cleanup operations
    if err := r.cleanupExternalResources(ctx, db); err != nil {
        log.Error(err, "Failed to cleanup external resources")
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

    log.Info("Finalizer removed, resource will be deleted")
    return ctrl.Result{}, nil
}

// cleanupExternalResources performs actual cleanup
func (r *DatabaseReconciler) cleanupExternalResources(ctx context.Context, db *databasev1.Database) error {
    log := log.FromContext(ctx)

    // Wait for StatefulSet to be deleted (owner reference handles it)
    statefulSet := &appsv1.StatefulSet{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      db.Name,
        Namespace: db.Namespace,
    }, statefulSet)

    if !errors.IsNotFound(err) {
        // StatefulSet still exists, wait for owner reference to delete it
        log.Info("Waiting for StatefulSet to be deleted")
        return fmt.Errorf("StatefulSet still exists")
    }

    // Add other cleanup operations here:
    // - Delete backups in external system
    // - Notify external services
    // - Clean up external resources

    log.Info("Cleanup completed")
    return nil
}

