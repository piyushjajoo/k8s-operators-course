// Solution: Performance Optimizations from Module 7
// This demonstrates various performance optimization techniques for kubebuilder operators
// File location: internal/controller/database_controller.go (additions)
//
// Directory structure reminder:
// internal/
// ├── controller/
// │   └── database_controller.go  <- Add these optimizations here
// └── webhook/
//     └── database_webhook.go     <- Webhooks from Module 5

package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	databasev1 "github.com/example/postgres-operator/api/v1"
)

// DatabaseReconciler reconciles a Database object
// This version includes performance optimizations
type DatabaseReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	APILimiter *rate.Limiter // Optional: for external API rate limiting
}

// Example 1: Controller Options with Rate Limiting
// Configure MaxConcurrentReconciles and custom rate limiter in SetupWithManager
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.Database{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		WithOptions(controller.Options{
			// Limit concurrent reconciliations to prevent overload
			MaxConcurrentReconciles: 2,
			// Custom rate limiter with exponential backoff
			RateLimiter: workqueue.NewItemExponentialFailureRateLimiter(
				time.Millisecond*5,  // Base delay
				time.Second*1000,    // Max delay
			),
		}).
		Complete(r)
}

// Example 2: Setting up Field Indexes (call from cmd/main.go)
// This enables fast filtered queries using client.MatchingFields
func SetupIndexes(mgr ctrl.Manager) error {
	// Index databases by environment for fast lookup
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&databasev1.Database{},
		"spec.environment",
		func(obj client.Object) []string {
			db := obj.(*databasev1.Database)
			if db.Spec.Environment == "" {
				return nil
			}
			return []string{db.Spec.Environment}
		},
	); err != nil {
		return fmt.Errorf("failed to create environment index: %w", err)
	}
	return nil
}

// Example 3: Optimized Queries with Field Selectors
// These queries use indexes set up in SetupIndexes
func (r *DatabaseReconciler) getDatabasesByEnvironment(ctx context.Context, env string) (*databasev1.DatabaseList, error) {
	databases := &databasev1.DatabaseList{}
	// This uses the index created above for fast filtering
	err := r.List(ctx, databases, client.MatchingFields{
		"spec.environment": env,
	})
	return databases, err
}

// Example 4: Batch Operations for Efficiency
func (r *DatabaseReconciler) reconcileBatch(ctx context.Context, databases []databasev1.Database) error {
	// Group by operation type
	var toCreate, toUpdate []databasev1.Database

	for _, db := range databases {
		if db.Status.Phase == "" {
			toCreate = append(toCreate, db)
		} else {
			toUpdate = append(toUpdate, db)
		}
	}

	// Batch create
	for _, db := range toCreate {
		if err := r.reconcileDatabase(ctx, &db); err != nil {
			return err
		}
	}

	// Batch update
	for _, db := range toUpdate {
		if err := r.reconcileDatabase(ctx, &db); err != nil {
			return err
		}
	}

	return nil
}

// Example 5: Parallel Processing (use with caution)
func (r *DatabaseReconciler) reconcileParallel(ctx context.Context, requests []ctrl.Request) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(requests))

	for _, req := range requests {
		wg.Add(1)
		go func(request ctrl.Request) {
			defer wg.Done()
			_, err := r.Reconcile(ctx, request)
			if err != nil {
				errChan <- err
			}
		}(req)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("reconciliation errors: %v", errors)
	}

	return nil
}

// Example 6: Custom Performance Metrics
// Register these with controller-runtime's metrics registry
var (
	reconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_operator_reconcile_duration_seconds",
			Help:    "Duration of reconciliations",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"result"},
	)

	databasesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_operator_databases_total",
			Help: "Total number of Database resources by phase",
		},
		[]string{"phase"},
	)
)

func init() {
	// Register custom metrics with controller-runtime's registry
	metrics.Registry.MustRegister(reconcileDuration, databasesTotal)
}

// Example 7: Reconcile with metrics instrumentation
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	start := time.Now()
	var err error
	result := "success"

	defer func() {
		duration := time.Since(start).Seconds()
		if err != nil {
			result = "error"
		}
		reconcileDuration.WithLabelValues(result).Observe(duration)
	}()

	// Optional: Wait for external API rate limiter if configured
	if r.APILimiter != nil {
		if err = r.APILimiter.Wait(ctx); err != nil {
			return ctrl.Result{}, err
		}
	}

	// ... reconciliation logic ...

	return ctrl.Result{}, err
}

// Helper function - placeholder for actual reconciliation
func (r *DatabaseReconciler) reconcileDatabase(ctx context.Context, db *databasev1.Database) error {
	// Actual reconciliation logic would go here
	return nil
}
