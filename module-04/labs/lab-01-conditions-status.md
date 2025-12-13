---
layout: default
title: "Lab 04.1: Conditions Status"
nav_order: 11
parent: "Module 4: Advanced Reconciliation"
grand_parent: Modules
permalink: /module-04/labs/conditions-status/
mermaid: true
---

# Lab 4.1: Implementing Status Conditions

**Related Lesson:** [Lesson 4.1: Conditions and Status Management](../lessons/01-conditions-status.md)  
**Navigation:** [Module Overview](../README.md) | [Next Lab: Finalizers →](lab-02-finalizers-cleanup.md)

## Objectives

- Add conditions to your Database operator
- Implement condition helper functions
- Update conditions based on resource state
- Observe condition transitions

## Prerequisites

- Completion of [Module 3](../../module-03/README.md)
- PostgreSQL operator from Module 3
- Understanding of status management

## Exercise 1: Add Conditions to Status

### Task 1.1: Update Status Type

Edit `api/v1/database_types.go`:

```go
// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
	// Phase is the current phase
	// +kubebuilder:validation:Enum=Pending;Creating;Ready;Failed
	Phase string `json:"phase,omitempty"`

	// Ready indicates if the database is ready
	Ready bool `json:"ready,omitempty"`

	// Endpoint is the database endpoint
	Endpoint string `json:"endpoint,omitempty"`

	// SecretName is the name of the Secret containing database credentials
	SecretName string `json:"secretName,omitempty"`

	// Conditions represent the latest observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration tracks the generation this status applies to
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}
```

### Task 1.2: Regenerate Code

```bash
# Regenerate code
make generate
make manifests
```

## Exercise 2: Implement Condition Helpers

### Task 2.1: Add Helper Functions

Add to `internal/controller/database_controller.go`:

```go
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
```

## Exercise 3: Update Reconciliation Logic

### Task 3.1: Add Conditions to Reconcile

Modify your `reconcileStatefulSet` and `updateStatus` function as below:

```go
func (r *DatabaseReconciler) reconcileStatefulSet(ctx context.Context, db *databasev1.Database) error {
	// ... existing code

	if errors.IsNotFound(err) {
		// Set owner reference
		if err := ctrl.SetControllerReference(db, desiredStatefulSet, r.Scheme); err != nil {
			return err
		}
		logger.Info("Creating StatefulSet", "name", desiredStatefulSet.Name)
		r.setCondition(db, "Ready", metav1.ConditionFalse, "StatefulSetNotFound", "StatefulSet not found")
		r.setCondition(db, "Progressing", metav1.ConditionTrue, "Creating", "Creating StatefulSet")
		return r.Create(ctx, desiredStatefulSet)
	} else if err != nil {
		r.setCondition(db, "Ready", metav1.ConditionFalse, "Error", err.Error())
		return err
	}

	// ... existing code

	return nil
}

func (r *DatabaseReconciler) updateStatus(ctx context.Context, db *databasev1.Database) error {
	// ... existing code
    
	if err != nil {
		db.Status.Phase = "Pending"
		db.Status.Ready = false
	} else {
		if statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas {
			db.Status.Phase = "Ready"
			db.Status.Ready = true
			db.Status.Endpoint = fmt.Sprintf("%s.%s.svc.cluster.local:5432", db.Name, db.Namespace)
			r.setCondition(db, "Ready", metav1.ConditionTrue, "AllReplicasReady", "All replicas are ready")
			r.setCondition(db, "Progressing", metav1.ConditionFalse, "ReconciliationComplete", "Reconciliation complete")
		} else {
			db.Status.Phase = "Creating"
			db.Status.Ready = false
			r.setCondition(db, "Ready", metav1.ConditionFalse, "ReplicasNotReady",
				fmt.Sprintf("%d/%d replicas ready", statefulSet.Status.ReadyReplicas, *statefulSet.Spec.Replicas))
			r.setCondition(db, "Progressing", metav1.ConditionTrue, "Scaling", "Waiting for replicas to be ready")
		}
	}

	return r.Status().Update(ctx, db)
}
```

## Exercise 4: Test Conditions

### Task 4.1: Install and Run

```bash
# Install CRD
make install

# Run operator
make run
```

### Task 4.2: Create Database

```bash
# Create Database
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-db
spec:
  image: postgres:14
  replicas: 1
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF
```

### Task 4.3: Observe Conditions

```bash
# Watch conditions
kubectl get database test-db -o jsonpath='{.status.conditions}' | jq '.'

# Watch condition transitions
watch -n 1 'kubectl get database test-db -o jsonpath="{.status.conditions[?(@.type==\"Ready\")]}"'
```

## Exercise 5: Test Condition Transitions

### Task 5.1: Scale Database

```bash
# Scale up
kubectl patch database test-db --type merge -p '{"spec":{"replicas":3}}'

# Watch Progressing condition
kubectl get database test-db -o jsonpath='{.status.conditions[?(@.type=="Progressing")]}'
```

### Task 5.2: Check Observed Generation

```bash
# Get generation
kubectl get database test-db -o jsonpath='{.metadata.generation}'

# Get observed generation
kubectl get database test-db -o jsonpath='{.status.conditions[0].observedGeneration}'

# They should match when reconciliation is complete
```

## Cleanup

```bash
# Delete Database
kubectl delete database test-db
```

## Lab Summary

In this lab, you:
- Added conditions to Database status
- Implemented condition helper functions
- Updated conditions in reconciliation
- Observed condition transitions
- Tested condition updates

## Key Learnings

1. Conditions provide structured status reporting
2. Use meta.SetStatusCondition for updates
3. Track observed generation
4. Update conditions based on actual state
5. Conditions transition through states
6. Standard condition types improve UX

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [Condition Helpers](../solutions/conditions-helpers.go) - Helper functions for managing conditions

## Next Steps

Now let's implement finalizers for graceful cleanup!

**Navigation:** [← Module Overview](../README.md) | [Related Lesson](../lessons/01-conditions-status.md) | [Next Lab: Finalizers →](lab-02-finalizers-cleanup.md)
