// Solution: Watch Setup from Module 4
// This shows how to set up watches for dependent resources

package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	databasev1 "github.com/example/postgres-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// ... existing code

func (r *DatabaseReconciler) findDatabasesForSecret(ctx context.Context, secret client.Object) []reconcile.Request {
	databases := &databasev1.DatabaseList{}
	r.List(context.Background(), databases)

	var requests []reconcile.Request
	for _, db := range databases.Items {
		// If Database references this Secret
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

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create index for image field
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
		Owns(&appsv1.StatefulSet{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldSS := e.ObjectOld.(*appsv1.StatefulSet)
				newSS := e.ObjectNew.(*appsv1.StatefulSet)
				// Reconcile on spec changes (Generation) OR status changes (ReadyReplicas)
				// Without checking ReadyReplicas, Database status would never update to Ready!
				return oldSS.Generation != newSS.Generation ||
					oldSS.Status.ReadyReplicas != newSS.Status.ReadyReplicas
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return true
			},
		})).
		Owns(&corev1.Service{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findDatabasesForSecret),
		).
		Complete(r)
}
