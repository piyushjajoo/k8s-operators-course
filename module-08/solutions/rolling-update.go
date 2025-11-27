// Solution: Rolling Update Handling from Module 8
// This demonstrates how to handle rolling updates for stateful applications

package controller

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	databasev1 "github.com/example/postgres-operator/api/v1"
)

func (r *DatabaseReconciler) updateStatefulSet(ctx context.Context, db *databasev1.Database) error {
	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      db.Name,
		Namespace: db.Namespace,
	}, statefulSet)

	if errors.IsNotFound(err) {
		// StatefulSet doesn't exist, create it
		return r.createStatefulSet(ctx, db)
	}

	if err != nil {
		return err
	}

	// Check if update needed
	desiredImage := db.Spec.Image
	currentImage := statefulSet.Spec.Template.Spec.Containers[0].Image

	if desiredImage != currentImage {
		// Update image
		statefulSet.Spec.Template.Spec.Containers[0].Image = desiredImage

		// Update StatefulSet (will trigger rolling update)
		if err := r.Update(ctx, statefulSet); err != nil {
			return err
		}

		// Wait for update to complete
		return r.waitForRollingUpdate(ctx, statefulSet)
	}

	// Check if replicas need updating
	if *statefulSet.Spec.Replicas != db.Spec.Replicas {
		statefulSet.Spec.Replicas = &db.Spec.Replicas
		if err := r.Update(ctx, statefulSet); err != nil {
			return err
		}
	}

	return nil
}

func (r *DatabaseReconciler) waitForRollingUpdate(ctx context.Context, ss *appsv1.StatefulSet) error {
	return wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		err := r.Get(ctx, client.ObjectKeyFromObject(ss), ss)
		if err != nil {
			return false, err
		}

		// Check if update complete
		// All replicas should be updated and ready
		if ss.Status.UpdatedReplicas == *ss.Spec.Replicas &&
			ss.Status.ReadyReplicas == *ss.Spec.Replicas {
			return true, nil
		}

		return false, nil
	})
}

func (r *DatabaseReconciler) createStatefulSet(ctx context.Context, db *databasev1.Database) error {
	// Create StatefulSet for database
	// This is a simplified example
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db.Name,
			Namespace: db.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &db.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": db.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "database",
							Image: db.Spec.Image,
						},
					},
				},
			},
		},
	}

	return r.Create(ctx, statefulSet)
}

