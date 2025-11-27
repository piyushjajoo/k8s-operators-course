// Solution: Condition Helper Functions from Module 4
// These helpers make it easy to manage conditions in your operator

package controller

import (
    "k8s.io/apimachinery/pkg/api/meta"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// setCondition sets a condition on the Database
func (r *DatabaseReconciler) setCondition(db *databasev1.Database, conditionType string, status metav1.ConditionStatus, reason, message string) {
    condition := metav1.Condition{
        Type:               conditionType,
        Status:             status,
        Reason:             reason,
        Message:            message,
        LastTransitionTime: metav1.Now(),
        ObservedGeneration: db.Generation,
    }

    meta.SetStatusCondition(&db.Status.Conditions, condition)
}

// getCondition gets a condition by type
func (r *DatabaseReconciler) getCondition(db *databasev1.Database, conditionType string) *metav1.Condition {
    return meta.FindStatusCondition(db.Status.Conditions, conditionType)
}

// Example usage in Reconcile:
//
// // Check StatefulSet status
// statefulSet := &appsv1.StatefulSet{}
// err := r.Get(ctx, client.ObjectKey{Name: db.Name, Namespace: db.Namespace}, statefulSet)
//
// if errors.IsNotFound(err) {
//     r.setCondition(db, "Ready", metav1.ConditionFalse, "StatefulSetNotFound", "StatefulSet not found")
//     r.setCondition(db, "Progressing", metav1.ConditionTrue, "Creating", "Creating StatefulSet")
// } else if statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas {
//     r.setCondition(db, "Ready", metav1.ConditionTrue, "AllReplicasReady", "All replicas are ready")
//     r.setCondition(db, "Progressing", metav1.ConditionFalse, "ReconciliationComplete", "Reconciliation complete")
// } else {
//     r.setCondition(db, "Ready", metav1.ConditionFalse, "ReplicasNotReady",
//         fmt.Sprintf("%d/%d replicas ready", statefulSet.Status.ReadyReplicas, *statefulSet.Spec.Replicas))
//     r.setCondition(db, "Progressing", metav1.ConditionTrue, "Scaling", "Waiting for replicas to be ready")
// }
//
// db.Status.ObservedGeneration = db.Generation
// return ctrl.Result{}, r.Status().Update(ctx, db)

