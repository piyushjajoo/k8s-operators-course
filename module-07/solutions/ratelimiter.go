// Solution: Rate Limiting for Kubebuilder Operators from Module 7
// This shows multiple approaches to rate limiting in kubebuilder operators

package ratelimiter

import (
	"context"
	"time"

	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"
)

// =============================================================================
// APPROACH 1: Use controller-runtime's built-in rate limiting (RECOMMENDED)
// =============================================================================
// Configure in SetupWithManager using controller.Options:
//
// func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
//     return ctrl.NewControllerManagedBy(mgr).
//         For(&databasev1.Database{}).
//         WithOptions(controller.Options{
//             MaxConcurrentReconciles: 2,
//             RateLimiter: workqueue.NewItemExponentialFailureRateLimiter(
//                 time.Millisecond * 5,  // Base delay
//                 time.Second * 1000,    // Max delay
//             ),
//         }).
//         Complete(r)
// }

// =============================================================================
// APPROACH 2: Use golang.org/x/time/rate for external API calls
// =============================================================================

// APIRateLimiter wraps the standard rate.Limiter for external API calls
type APIRateLimiter struct {
	limiter *rate.Limiter
}

// NewAPIRateLimiter creates a rate limiter for external API calls
// rps: requests per second
// burst: maximum burst size
func NewAPIRateLimiter(rps float64, burst int) *APIRateLimiter {
	return &APIRateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

// Wait blocks until a request is allowed or context is cancelled
func (r *APIRateLimiter) Wait(ctx context.Context) error {
	return r.limiter.Wait(ctx)
}

// Allow reports whether an event may happen now
func (r *APIRateLimiter) Allow() bool {
	return r.limiter.Allow()
}

// =============================================================================
// APPROACH 3: Custom rate limiter implementing controller-runtime interface
// =============================================================================

// CustomRateLimiter implements ratelimiter.RateLimiter interface
type CustomRateLimiter struct {
	baseDelay time.Duration
	maxDelay  time.Duration
	failures  map[interface{}]int
}

// NewCustomRateLimiter creates a custom rate limiter
func NewCustomRateLimiter(baseDelay, maxDelay time.Duration) ratelimiter.RateLimiter {
	return &CustomRateLimiter{
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
		failures:  make(map[interface{}]int),
	}
}

func (r *CustomRateLimiter) When(item interface{}) time.Duration {
	r.failures[item]++
	delay := r.baseDelay * time.Duration(1<<uint(r.failures[item]-1))
	if delay > r.maxDelay {
		delay = r.maxDelay
	}
	return delay
}

func (r *CustomRateLimiter) NumRequeues(item interface{}) int {
	return r.failures[item]
}

func (r *CustomRateLimiter) Forget(item interface{}) {
	delete(r.failures, item)
}

// =============================================================================
// Example usage in kubebuilder controller:
// =============================================================================
//
// // In internal/controller/database_controller.go
// type DatabaseReconciler struct {
//     client.Client
//     Scheme     *runtime.Scheme
//     APILimiter *APIRateLimiter  // For external API calls
// }
//
// // In cmd/main.go when creating the reconciler:
// if err = (&controller.DatabaseReconciler{
//     Client:     mgr.GetClient(),
//     Scheme:     mgr.GetScheme(),
//     APILimiter: ratelimiter.NewAPIRateLimiter(10, 1), // 10 req/sec
// }).SetupWithManager(mgr); err != nil {
//     setupLog.Error(err, "unable to create controller")
//     os.Exit(1)
// }
//
// // In Reconcile:
// func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
//     // Wait for rate limiter before external API calls
//     if r.APILimiter != nil {
//         if err := r.APILimiter.Wait(ctx); err != nil {
//             return ctrl.Result{}, err
//         }
//     }
//     // ... make external API call ...
// }

// GetWorkqueueRateLimiter returns a pre-configured rate limiter for work queues
// This is a convenience function for common use cases
func GetWorkqueueRateLimiter() workqueue.RateLimiter {
	return workqueue.NewItemExponentialFailureRateLimiter(
		5*time.Millisecond,   // Base delay
		1000*time.Second,     // Max delay
	)
}
