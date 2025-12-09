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

// Solution: Observability Patterns from Module 6 Lab 4
// This file demonstrates structured logging and event emission patterns.
//
// These patterns are meant to be integrated into your existing controller code,
// not used as a standalone file.

package controller

// =============================================================================
// PART 1: Controller Struct with Event Recorder
// =============================================================================
//
// Update your DatabaseReconciler struct to include the event recorder:
//
// import (
//     "k8s.io/client-go/tools/record"
// )
//
// type DatabaseReconciler struct {
//     client.Client
//     Scheme   *runtime.Scheme
//     Recorder record.EventRecorder  // Add this field
// }

// =============================================================================
// PART 2: Update main.go to Provide Event Recorder
// =============================================================================
//
// In cmd/main.go, update the controller setup:
//
// if err := (&controller.DatabaseReconciler{
//     Client:   mgr.GetClient(),
//     Scheme:   mgr.GetScheme(),
//     Recorder: mgr.GetEventRecorderFor("database-controller"),
// }).SetupWithManager(mgr); err != nil {
//     setupLog.Error(err, "unable to create controller", "controller", "Database")
//     os.Exit(1)
// }

// =============================================================================
// PART 3: Structured Logging Patterns
// =============================================================================
//
// Use log.FromContext(ctx) for consistent logging:
//
// func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
//     logger := log.FromContext(ctx)
//
//     // Info logs with key-value pairs
//     logger.Info("Reconciling Database",
//         "name", req.Name,
//         "namespace", req.Namespace,
//     )
//
//     // Error logs include the error object
//     if err != nil {
//         logger.Error(err, "Failed to get Database",
//             "name", req.Name,
//             "namespace", req.Namespace,
//         )
//     }
//
//     // Log with multiple context fields
//     logger.Info("Database status",
//         "name", db.Name,
//         "phase", db.Status.Phase,
//         "ready", db.Status.Ready,
//         "generation", db.Generation,
//         "observedGeneration", db.Status.ObservedGeneration,
//     )
// }

// =============================================================================
// PART 4: Event Emission Patterns
// =============================================================================
//
// Events provide user-visible information in `kubectl describe`:
//
// Event types:
// - "Normal" - for successful operations
// - "Warning" - for errors or issues
//
// func (r *DatabaseReconciler) handleProvisioning(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
//     // Event on starting provisioning
//     r.Recorder.Event(db, "Normal", "Provisioning", "Starting database provisioning")
//
//     // Event on creating resources
//     if err := r.reconcileStatefulSet(ctx, db); err != nil {
//         r.Recorder.Event(db, "Warning", "CreateFailed",
//             fmt.Sprintf("Failed to create StatefulSet: %v", err))
//         return ctrl.Result{}, err
//     }
//     r.Recorder.Event(db, "Normal", "Created", "StatefulSet created successfully")
//
//     // Event on success
//     r.Recorder.Event(db, "Normal", "Provisioned", "Database provisioning completed")
//     return ctrl.Result{}, nil
// }
//
// func (r *DatabaseReconciler) handleVerifying(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
//     // Event when database becomes ready
//     r.Recorder.Event(db, "Normal", "Ready",
//         fmt.Sprintf("Database is ready at %s", db.Status.Endpoint))
//     return ctrl.Result{}, nil
// }
//
// func (r *DatabaseReconciler) handleDeletion(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {
//     r.Recorder.Event(db, "Normal", "Deleting", "Starting cleanup of database resources")
//
//     // ... cleanup logic ...
//
//     r.Recorder.Event(db, "Normal", "Deleted", "Cleanup completed successfully")
//     return ctrl.Result{}, nil
// }

// =============================================================================
// PART 5: Update Tests to Include Recorder
// =============================================================================
//
// When testing, provide a fake recorder:
//
// import (
//     "k8s.io/client-go/tools/record"
// )
//
// controllerReconciler := &DatabaseReconciler{
//     Client:   k8sClient,
//     Scheme:   k8sClient.Scheme(),
//     Recorder: record.NewFakeRecorder(100),  // Buffer size of 100 events
// }

// =============================================================================
// PART 6: View Events
// =============================================================================
//
// Users can see events with kubectl:
//
// # View events for a specific resource
// kubectl get events --field-selector involvedObject.name=my-database
//
// # View events sorted by time
// kubectl get events --sort-by='.lastTimestamp'
//
// # View events in kubectl describe
// kubectl describe database my-database
//
// Example output in describe:
//
// Events:
//   Type    Reason       Age   From                 Message
//   ----    ------       ----  ----                 -------
//   Normal  Provisioning 2m    database-controller  Starting database provisioning
//   Normal  Created      2m    database-controller  StatefulSet created successfully
//   Normal  Ready        1m    database-controller  Database is ready at my-database.default.svc.cluster.local:5432
