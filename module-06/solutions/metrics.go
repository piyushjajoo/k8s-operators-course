// Solution: Prometheus Metrics from Module 6
// This demonstrates how to expose metrics for observability

package controller

import (
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
    // reconcileTotal counts the total number of reconciliations
    reconcileTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "database_reconcile_total",
            Help: "Total number of reconciliations",
        },
        []string{"result"}, // success, error
    )
    
    // reconcileDuration measures the duration of reconciliations
    reconcileDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "database_reconcile_duration_seconds",
            Help: "Duration of reconciliations in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"result"},
    )
    
    // databaseCount tracks the number of databases
    databaseCount = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "database_count",
            Help: "Number of databases",
        },
        []string{"phase"}, // Pending, Creating, Ready, Failed
    )
)

func init() {
    // Register metrics with the global registry
    metrics.Registry.MustRegister(
        reconcileTotal,
        reconcileDuration,
        databaseCount,
    )
}

// Example usage in Reconcile function:
//
// func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
//     start := time.Now()
//     var err error
//     result := "success"
//     
//     defer func() {
//         duration := time.Since(start).Seconds()
//         if err != nil {
//             result = "error"
//         }
//         reconcileDuration.WithLabelValues(result).Observe(duration)
//         reconcileTotal.WithLabelValues(result).Inc()
//     }()
//     
//     // ... reconciliation logic ...
//     
//     // Update database count
//     databases := &databasev1.DatabaseList{}
//     r.List(ctx, databases)
//     for _, db := range databases.Items {
//         databaseCount.WithLabelValues(db.Status.Phase).Inc()
//     }
//     
//     return ctrl.Result{}, err
// }

