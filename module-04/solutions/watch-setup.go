// Solution: Watch Setup from Module 4
// This shows how to set up watches for dependent resources

package controller

import (
    "context"

    "k8s.io/apimachinery/pkg/types"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/handler"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
    "sigs.k8s.io/controller-runtime/pkg/source"
    ctrl "sigs.k8s.io/controller-runtime"

    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    databasev1 "github.com/example/postgres-operator/api/v1"
)

// SetupWithManager sets up the controller with watches
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1.Database{}).
        // Watch owned resources (automatically reconciles owner when child changes)
        Owns(&appsv1.StatefulSet{}).
        Owns(&corev1.Service{}).
        // Watch non-owned resources
        Watches(
            &source.Kind{Type: &corev1.Secret{}},
            handler.EnqueueRequestsFromMapFunc(r.findDatabasesForSecret),
        ).
        Complete(r)
}

// findDatabasesForSecret finds all Databases that reference a Secret
func (r *DatabaseReconciler) findDatabasesForSecret(secret client.Object) []reconcile.Request {
    databases := &databasev1.DatabaseList{}
    r.List(context.Background(), databases)

    var requests []reconcile.Request
    for _, db := range databases.Items {
        // If Database references this Secret
        if db.Spec.SecretName == secret.GetName() &&
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

