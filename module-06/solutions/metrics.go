/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Solution: Prometheus Metrics from Module 6 Lab 4
// Location: internal/controller/metrics.go
//
// This file defines custom Prometheus metrics for the database operator.
// Metrics are automatically exposed at the /metrics endpoint.

package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// ReconcileTotal counts the total number of reconciliations
	ReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_reconcile_total",
			Help: "Total number of reconciliations per controller",
		},
		[]string{"result"}, // success, error, requeue
	)

	// ReconcileDuration measures the duration of reconciliations
	ReconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_reconcile_duration_seconds",
			Help:    "Duration of reconciliations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"},
	)

	// DatabasesTotal tracks the current number of Database resources
	DatabasesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_resources_total",
			Help: "Current number of Database resources by phase",
		},
		[]string{"phase"},
	)

	// DatabaseInfo provides information about each database
	DatabaseInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_info",
			Help: "Information about Database resources",
		},
		[]string{"name", "namespace", "image", "phase"},
	)
)

func init() {
	// Register custom metrics with the global registry
	metrics.Registry.MustRegister(
		ReconcileTotal,
		ReconcileDuration,
		DatabasesTotal,
		DatabaseInfo,
	)
}

// Usage in Reconcile function:
//
// func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
//     start := time.Now()
//     reconcileResult := "success"
//
//     // Defer metrics recording
//     defer func() {
//         duration := time.Since(start).Seconds()
//         ReconcileDuration.WithLabelValues(reconcileResult).Observe(duration)
//         ReconcileTotal.WithLabelValues(reconcileResult).Inc()
//     }()
//
//     // ... reconciliation logic ...
//
//     // On error, update the result label
//     if err != nil {
//         reconcileResult = "error"
//         return ctrl.Result{}, err
//     }
//
//     // Record database info
//     DatabaseInfo.WithLabelValues(
//         db.Name,
//         db.Namespace,
//         db.Spec.Image,
//         db.Status.Phase,
//     ).Set(1)
//
//     return ctrl.Result{}, nil
// }
