/*
State Machine Controller - Complete Implementation

This file demonstrates multi-phase reconciliation using a state machine pattern.
It shows the complete implementation including:
- State definitions
- State machine dispatcher
- State handler functions
- Resource creation during appropriate phases
- Status updates with conditions

State Flow:
Pending → Provisioning → Configuring → Deploying → Verifying → Ready
                                                              ↓
                                                           Failed (on error)

IMPORTANT: Before using this code, you must update your API types (api/v1/database_types.go)
to allow the new phase values. Update the Phase field enum:

    // +kubebuilder:validation:Enum=Pending;Provisioning;Configuring;Deploying;Verifying;Ready;Failed
    Phase string `json:"phase,omitempty"`

Then run: make manifests && make install
*/
package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	databasev1 "github.com/example/postgres-operator/api/v1"
)

// ============================================================================
// State Definitions
// ============================================================================

// DatabaseState represents the possible states of a Database resource
type DatabaseState string

const (
	StatePending      DatabaseState = "Pending"
	StateProvisioning DatabaseState = "Provisioning"
	StateConfiguring  DatabaseState = "Configuring"
	StateDeploying    DatabaseState = "Deploying"
	StateVerifying    DatabaseState = "Verifying"
	StateReady        DatabaseState = "Ready"
	StateFailed       DatabaseState = "Failed"
)

const finalizerName = "database.example.com/finalizer"

// ============================================================================
// Main Reconcile Function - Entry Point
// ============================================================================

// Reconcile is the main entry point that delegates to the state machine
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Read Database resource
	db := &databasev1.Database{}
	if err := r.Get(ctx, req.NamespacedName, db); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(db, finalizerName) {
		controllerutil.AddFinalizer(db, finalizerName)
		if err := r.Update(ctx, db); err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("Added finalizer", "name", db.Name)
	}

	// Check if resource is being deleted
	if !db.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, db)
	}

	logger.Info("Reconciling Database", "name", db.Name)

	// IMPORTANT: Use state machine for multi-phase reconciliation
	// This is the key change from direct reconciliation
	return r.reconcileWithStateMachine(ctx, db)
}

// ============================================================================
// State Machine Dispatcher
// ============================================================================

// reconcileWithStateMachine routes reconciliation to the appropriate state handler
func (r *DatabaseReconciler) reconcileWithStateMachine(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
	currentState := DatabaseState(db.Status.Phase)
	if currentState == "" {
		currentState = StatePending
	}

	logger := log.FromContext(ctx)
	logger.Info("Reconciling", "state", currentState)

	switch currentState {
	case StatePending:
		return r.transitionToProvisioning(ctx, db)
	case StateProvisioning:
		return r.handleProvisioning(ctx, db)
	case StateConfiguring:
		return r.handleConfiguring(ctx, db)
	case StateDeploying:
		return r.handleDeploying(ctx, db)
	case StateVerifying:
		return r.handleVerifying(ctx, db)
	case StateReady:
		return r.handleReady(ctx, db)
	case StateFailed:
		return r.handleFailed(ctx, db)
	default:
		logger.Info("Unknown state, resetting to Pending", "state", currentState)
		db.Status.Phase = string(StatePending)
		return ctrl.Result{}, r.Status().Update(ctx, db)
	}
}

// ============================================================================
// State Handlers
// ============================================================================

// transitionToProvisioning moves from Pending to Provisioning
func (r *DatabaseReconciler) transitionToProvisioning(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("STATE TRANSITION: Pending -> Provisioning", "database", db.Name)

	db.Status.Phase = string(StateProvisioning)
	db.Status.Ready = false
	r.setCondition(db, "Progressing", metav1.ConditionTrue, "Provisioning", "Starting provisioning")
	r.setCondition(db, "Ready", metav1.ConditionFalse, "Provisioning", "Database is being provisioned")
	if err := r.Status().Update(ctx, db); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Waiting 15 seconds before next reconciliation (for visualization)", "currentPhase", db.Status.Phase)
	// Delay to visualize state transition (remove in production)
	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}

// handleProvisioning creates the Secret and StatefulSet
func (r *DatabaseReconciler) handleProvisioning(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling Provisioning phase", "database", db.Name)

	// Ensure Secret exists first (StatefulSet needs it for credentials)
	if err := r.reconcileSecret(ctx, db); err != nil {
		logger.Error(err, "Failed to reconcile Secret")
		return r.transitionToFailed(ctx, db, "SecretCreationFailed", err.Error())
	}

	// Check if StatefulSet exists
	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      db.Name,
		Namespace: db.Namespace,
	}, statefulSet)

	if errors.IsNotFound(err) {
		// Create StatefulSet
		logger.Info("Creating StatefulSet", "database", db.Name)
		if err := r.reconcileStatefulSet(ctx, db); err != nil {
			logger.Error(err, "Failed to create StatefulSet")
			return r.transitionToFailed(ctx, db, "StatefulSetCreationFailed", err.Error())
		}
		logger.Info("StatefulSet created, requeuing")
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// StatefulSet exists, move to next phase
	logger.Info("STATE TRANSITION: Provisioning -> Configuring", "database", db.Name)
	db.Status.Phase = string(StateConfiguring)
	r.setCondition(db, "Progressing", metav1.ConditionTrue, "Configuring", "StatefulSet created, configuring")
	if err := r.Status().Update(ctx, db); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Waiting 15 seconds before next reconciliation (for visualization)", "currentPhase", db.Status.Phase)
	// Delay to visualize state transition (remove in production)
	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}

// handleConfiguring creates the Service and performs configuration
func (r *DatabaseReconciler) handleConfiguring(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling Configuring phase", "database", db.Name)

	// Ensure Service exists
	logger.Info("Creating Service", "database", db.Name)
	if err := r.reconcileService(ctx, db); err != nil {
		logger.Error(err, "Failed to reconcile Service")
		return r.transitionToFailed(ctx, db, "ServiceCreationFailed", err.Error())
	}

	// Configure database (create users, databases, etc.)
	// In a real operator, you might:
	// - Wait for the database to be connectable
	// - Run initialization scripts
	// - Create database users
	// - Set up replication

	logger.Info("STATE TRANSITION: Configuring -> Deploying", "database", db.Name)
	db.Status.Phase = string(StateDeploying)
	r.setCondition(db, "Progressing", metav1.ConditionTrue, "Deploying", "Configuration complete, deploying")
	if err := r.Status().Update(ctx, db); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Waiting 15 seconds before next reconciliation (for visualization)", "currentPhase", db.Status.Phase)
	// Delay to visualize state transition (remove in production)
	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}

// handleDeploying waits for the StatefulSet to be ready
func (r *DatabaseReconciler) handleDeploying(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling Deploying phase", "database", db.Name)

	// Check if StatefulSet is ready
	statefulSet := &appsv1.StatefulSet{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      db.Name,
		Namespace: db.Namespace,
	}, statefulSet); err != nil {
		if errors.IsNotFound(err) {
			// StatefulSet was deleted, go back to provisioning
			logger.Info("StatefulSet not found, transitioning back to Provisioning")
			db.Status.Phase = string(StateProvisioning)
			return ctrl.Result{}, r.Status().Update(ctx, db)
		}
		return ctrl.Result{}, err
	}

	// Check replica readiness
	desiredReplicas := int32(1)
	if statefulSet.Spec.Replicas != nil {
		desiredReplicas = *statefulSet.Spec.Replicas
	}

	if statefulSet.Status.ReadyReplicas >= desiredReplicas {
		logger.Info("STATE TRANSITION: Deploying -> Verifying", "database", db.Name)
		db.Status.Phase = string(StateVerifying)
		r.setCondition(db, "Progressing", metav1.ConditionTrue, "Verifying", "Deployment complete, verifying")
		if err := r.Status().Update(ctx, db); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info("Waiting 15 seconds before next reconciliation (for visualization)", "currentPhase", db.Status.Phase)
		// Delay to visualize state transition (remove in production)
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}

	// Not ready yet - update condition with progress
	logger.Info("Waiting for StatefulSet replicas to be ready",
		"database", db.Name,
		"readyReplicas", statefulSet.Status.ReadyReplicas,
		"desiredReplicas", desiredReplicas)
	r.setCondition(db, "Progressing", metav1.ConditionTrue, "WaitingForReplicas",
		fmt.Sprintf("Waiting for replicas: %d/%d ready", statefulSet.Status.ReadyReplicas, desiredReplicas))
	if err := r.Status().Update(ctx, db); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// handleVerifying performs health checks before marking as Ready
func (r *DatabaseReconciler) handleVerifying(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling Verifying phase", "database", db.Name)

	// Verify database is working
	// In a real operator, you might:
	// - Connect to the database
	// - Run a test query
	// - Check replication status
	// - Verify backups are configured

	// For this example, we assume verification passes
	logger.Info("STATE TRANSITION: Verifying -> Ready", "database", db.Name)

	db.Status.Phase = string(StateReady)
	db.Status.Ready = true
	db.Status.SecretName = r.secretName(db)
	db.Status.Endpoint = fmt.Sprintf("%s.%s.svc.cluster.local:5432", db.Name, db.Namespace)

	r.setCondition(db, "Ready", metav1.ConditionTrue, "AllChecksPassed", "Database is ready")
	r.setCondition(db, "Progressing", metav1.ConditionFalse, "ReconciliationComplete", "Reconciliation complete")

	logger.Info("Database is now READY!", "database", db.Name, "endpoint", db.Status.Endpoint)

	return ctrl.Result{}, r.Status().Update(ctx, db)
}

// handleReady monitors the ready state and handles updates
func (r *DatabaseReconciler) handleReady(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Check if StatefulSet still exists and is healthy
	statefulSet := &appsv1.StatefulSet{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      db.Name,
		Namespace: db.Namespace,
	}, statefulSet); err != nil {
		if errors.IsNotFound(err) {
			// StatefulSet was deleted, need to re-provision
			logger.Info("StatefulSet deleted, transitioning to Provisioning")
			db.Status.Phase = string(StateProvisioning)
			db.Status.Ready = false
			r.setCondition(db, "Ready", metav1.ConditionFalse, "StatefulSetMissing", "StatefulSet was deleted")
			return ctrl.Result{}, r.Status().Update(ctx, db)
		}
		return ctrl.Result{}, err
	}

	// Check if spec changed (e.g., replicas, image)
	if err := r.reconcileStatefulSet(ctx, db); err != nil {
		logger.Error(err, "Failed to reconcile StatefulSet")
		return ctrl.Result{}, err
	}

	// If replicas changed and not all ready, go back to Deploying
	desiredReplicas := int32(1)
	if statefulSet.Spec.Replicas != nil {
		desiredReplicas = *statefulSet.Spec.Replicas
	}

	if statefulSet.Status.ReadyReplicas < desiredReplicas {
		logger.Info("Replicas not ready, transitioning to Deploying")
		db.Status.Phase = string(StateDeploying)
		db.Status.Ready = false
		r.setCondition(db, "Ready", metav1.ConditionFalse, "ScalingInProgress", "Scaling operation in progress")
		return ctrl.Result{}, r.Status().Update(ctx, db)
	}

	// Everything is good
	return ctrl.Result{}, nil
}

// handleFailed handles the failed state with retry logic
func (r *DatabaseReconciler) handleFailed(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Check if we should retry
	// In a real operator, you might:
	// - Check retry count
	// - Implement exponential backoff
	// - Check if the error condition has been resolved

	logger.Info("In Failed state, will retry", "retryAfter", "1m")

	// For now, transition back to Pending to retry
	// A more sophisticated implementation would track retry counts
	db.Status.Phase = string(StatePending)
	r.setCondition(db, "Progressing", metav1.ConditionTrue, "Retrying", "Retrying after failure")

	return ctrl.Result{RequeueAfter: 1 * time.Minute}, r.Status().Update(ctx, db)
}

// transitionToFailed is a helper to transition to the Failed state
func (r *DatabaseReconciler) transitionToFailed(ctx context.Context, db *databasev1.Database, reason, message string) (ctrl.Result, error) {
	db.Status.Phase = string(StateFailed)
	db.Status.Ready = false
	r.setCondition(db, "Ready", metav1.ConditionFalse, reason, message)
	r.setCondition(db, "Progressing", metav1.ConditionFalse, "Failed", "Reconciliation failed")
	return ctrl.Result{}, r.Status().Update(ctx, db)
}

// ============================================================================
// Helper Functions (referenced but defined elsewhere)
// ============================================================================

// These functions should be defined in your controller:
// - reconcileSecret(ctx, db) error
// - reconcileStatefulSet(ctx, db) error
// - reconcileService(ctx, db) error
// - handleDeletion(ctx, db) (ctrl.Result, error)
// - setCondition(db, type, status, reason, message)
// - secretName(db) string

