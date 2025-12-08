// Solution: Watch Setup from Module 4
// This shows how to set up watches for dependent resources

package controller

import (
    "context"
    "fmt"

    "k8s.io/apimachinery/pkg/types"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/handler"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
    ctrl "sigs.k8s.io/controller-runtime"

    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    databasev1 "github.com/example/postgres-operator/api/v1"
)

// SetupWithManager sets up the controller with watches and indexes
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    // Create index for image field - allows efficient lookup of Databases by image
    if err := mgr.GetFieldIndexer().IndexField(
        context.Background(),
        &databasev1.Database{},
        "spec.image",
        func(obj client.Object) []string {
            db, ok := obj.(*databasev1.Database)
            if !ok {
                return nil
            }
            if db.Spec.Image != "" {
                return []string{db.Spec.Image}
            }
            return nil
        },
    ); err != nil {
        return err
    }

    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        // Watch owned resources (automatically reconciles owner when child changes)
        Owns(&appsv1.StatefulSet{}).
        Owns(&corev1.Service{}).
        // Watch non-owned resources (Secrets)
        Watches(
            &corev1.Secret{},
            handler.EnqueueRequestsFromMapFunc(r.findDatabasesForSecret),
        ).
        Complete(r)
}

// secretName returns the name of the Secret for a Database
func (r *DatabaseReconciler) secretName(db *databasev1.Database) string {
    return fmt.Sprintf("%s-credentials", db.Name)
}

// findDatabasesForSecret finds all Databases that use a Secret
// The Secret name is derived from the Database name (e.g., "test-db" -> "test-db-credentials")
func (r *DatabaseReconciler) findDatabasesForSecret(ctx context.Context, secret client.Object) []reconcile.Request {
    databases := &databasev1.DatabaseList{}
    r.List(ctx, databases)

    var requests []reconcile.Request
    for _, db := range databases.Items {
        // Check if this Secret belongs to this Database
        // Secret name is derived: {db-name}-credentials
        if r.secretName(&db) == secret.GetName() &&
            db.Namespace == secret.GetNamespace() {
            requests = append(requests, reconcile.Request{
                NamespacedName: types.NamespacedName{
                    Name:      db.Name,
                    Namespace: db.Namespace,
                },
            })
        }
    }
    return requests
}

// findDatabasesByImage finds all Databases using a specific PostgreSQL image
// Uses the index for efficient lookup
func (r *DatabaseReconciler) findDatabasesByImage(ctx context.Context, image string) ([]databasev1.Database, error) {
    databases := &databasev1.DatabaseList{}
    err := r.List(ctx, databases, client.MatchingFields{
        "spec.image": image,
    })

    if err != nil {
        return nil, err
    }

    return databases.Items, nil
}

// Example with event predicates (only reconcile on spec changes):
//
// Owns(&appsv1.StatefulSet{}, builder.WithPredicates(predicate.Funcs{
//     UpdateFunc: func(e event.UpdateEvent) bool {
//         oldSS := e.ObjectOld.(*appsv1.StatefulSet)
//         newSS := e.ObjectNew.(*appsv1.StatefulSet)
//         return oldSS.Generation != newSS.Generation
//     },
//     CreateFunc: func(e event.CreateEvent) bool {
//         return true
//     },
//     DeleteFunc: func(e event.DeleteEvent) bool {
//         return true
//     },
// }))

