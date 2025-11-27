// Solution: Rate Limiter from Module 7
// This implements rate limiting to prevent API overload

package ratelimiter

import (
	"sync"
	"time"
)

// RateLimiter limits the rate of operations
type RateLimiter struct {
	mu          sync.Mutex
	lastCall    time.Time
	minInterval time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(minInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		minInterval: minInterval,
		lastCall:    time.Time{}, // Zero time, first call will proceed immediately
	}
}

// Wait blocks until the minimum interval has passed since the last call
func (r *RateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastCall)

	if elapsed < r.minInterval {
		time.Sleep(r.minInterval - elapsed)
	}

	r.lastCall = time.Now()
}

// Reset resets the rate limiter
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastCall = time.Time{}
}

// Example usage:
//
// type DatabaseReconciler struct {
//     client.Client
//     Scheme      *runtime.Scheme
//     rateLimiter *ratelimiter.RateLimiter
// }
//
// func NewDatabaseReconciler(mgr ctrl.Manager) *DatabaseReconciler {
//     return &DatabaseReconciler{
//         Client:      mgr.GetClient(),
//         Scheme:      mgr.GetScheme(),
//         rateLimiter: ratelimiter.NewRateLimiter(100 * time.Millisecond),
//     }
// }
//
// func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
//     // Rate limit API calls
//     r.rateLimiter.Wait()
//
//     // ... reconciliation logic ...
// }
