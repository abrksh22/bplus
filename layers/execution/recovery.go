package execution

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/abrksh22/bplus/internal/errors"
	"github.com/abrksh22/bplus/internal/logging"
)

// CircuitBreaker implements the circuit breaker pattern to prevent cascading failures.
type CircuitBreaker struct {
	mu              sync.RWMutex
	maxFailures     int
	resetTimeout    time.Duration
	failures        int
	lastFailureTime time.Time
	state           CircuitState
	logger          *logging.Logger
}

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // Normal operation
	CircuitOpen                         // Too many failures, rejecting requests
	CircuitHalfOpen                     // Testing if service has recovered
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
		logger:       logging.NewDefaultLogger().WithComponent("circuit_breaker"),
	}
}

// Execute runs the given function with circuit breaker protection.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	cb.mu.Lock()

	// Check if we should transition from open to half-open
	if cb.state == CircuitOpen && time.Since(cb.lastFailureTime) >= cb.resetTimeout {
		cb.logger.Info("Circuit breaker transitioning to half-open", "failures", cb.failures)
		cb.state = CircuitHalfOpen
	}

	// Reject if circuit is open
	if cb.state == CircuitOpen {
		cb.mu.Unlock()
		return errors.New(errors.ErrCodeInternal, "circuit breaker is open")
	}

	cb.mu.Unlock()

	// Execute function
	err := fn(ctx)

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// recordFailure records a failure and potentially opens the circuit.
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.logger.Warn("Circuit breaker opening", "failures", cb.failures, "max", cb.maxFailures)
		cb.state = CircuitOpen
	}
}

// recordSuccess records a success and potentially closes the circuit.
func (cb *CircuitBreaker) recordSuccess() {
	if cb.state == CircuitHalfOpen {
		cb.logger.Info("Circuit breaker closing after successful test")
	}
	cb.failures = 0
	cb.state = CircuitClosed
}

// GetState returns the current state of the circuit breaker.
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset manually resets the circuit breaker.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = CircuitClosed
	cb.logger.Info("Circuit breaker manually reset")
}

// RetryPolicy defines how retries should be performed.
type RetryPolicy struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       bool
}

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// RetryWithPolicy executes a function with retry logic.
func RetryWithPolicy(ctx context.Context, policy *RetryPolicy, fn func(context.Context) error) error {
	if policy == nil {
		policy = DefaultRetryPolicy()
	}

	logger := logging.NewDefaultLogger().WithComponent("retry")

	var lastErr error
	delay := policy.InitialDelay

	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		// Try to execute
		err := fn(ctx)
		if err == nil {
			if attempt > 1 {
				logger.Info("Operation succeeded after retry", "attempt", attempt)
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryable(err) {
			logger.Debug("Error is not retryable", "attempt", attempt, "error", err.Error())
			return err
		}

		// Check if we've exhausted retries
		if attempt >= policy.MaxAttempts {
			logger.Warn("Max retry attempts reached", "attempts", attempt, "error", err.Error())
			break
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Calculate delay with exponential backoff
		currentDelay := delay
		if policy.Jitter {
			currentDelay = addJitter(delay, 0.2)
		}

		logger.Info("Retrying after error", "attempt", attempt, "delay", currentDelay.String(), "error", err.Error())

		// Wait before retrying
		select {
		case <-time.After(currentDelay):
		case <-ctx.Done():
			return ctx.Err()
		}

		// Exponential backoff
		delay = time.Duration(float64(delay) * policy.Multiplier)
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}
	}

	return errors.Wrapf(lastErr, errors.ErrCodeInternal, "operation failed after %d attempts", policy.MaxAttempts)
}

// isRetryable determines if an error is retryable.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for known retryable error types
	if errors.IsRetryable(err) {
		return true
	}

	// Check error message for retryable patterns
	errMsg := err.Error()
	retryablePatterns := []string{
		"timeout",
		"temporary",
		"connection refused",
		"connection reset",
		"rate limit",
		"too many requests",
		"service unavailable",
		"gateway timeout",
		"deadline exceeded",
	}

	for _, pattern := range retryablePatterns {
		if contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && anyContains(s, substr)))
}

func anyContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ErrorRecoveryContext manages error recovery state for an agent.
type ErrorRecoveryContext struct {
	mu                   sync.RWMutex
	circuitBreakers      map[string]*CircuitBreaker
	consecutiveErrors    int
	maxConsecutiveErrors int
	logger               *logging.Logger
}

// NewErrorRecoveryContext creates a new error recovery context.
func NewErrorRecoveryContext(maxConsecutiveErrors int) *ErrorRecoveryContext {
	return &ErrorRecoveryContext{
		circuitBreakers:      make(map[string]*CircuitBreaker),
		maxConsecutiveErrors: maxConsecutiveErrors,
		logger:               logging.NewDefaultLogger().WithComponent("error_recovery"),
	}
}

// GetCircuitBreaker returns a circuit breaker for the given key.
func (erc *ErrorRecoveryContext) GetCircuitBreaker(key string) *CircuitBreaker {
	erc.mu.Lock()
	defer erc.mu.Unlock()

	if cb, ok := erc.circuitBreakers[key]; ok {
		return cb
	}

	// Create new circuit breaker
	cb := NewCircuitBreaker(3, 30*time.Second)
	erc.circuitBreakers[key] = cb
	return cb
}

// RecordError records an error and checks if recovery is needed.
func (erc *ErrorRecoveryContext) RecordError(err error) bool {
	erc.mu.Lock()
	defer erc.mu.Unlock()

	erc.consecutiveErrors++
	erc.logger.Debug("Error recorded", "consecutive", erc.consecutiveErrors, "max", erc.maxConsecutiveErrors, "error", err)

	return erc.consecutiveErrors >= erc.maxConsecutiveErrors
}

// RecordSuccess records a successful operation.
func (erc *ErrorRecoveryContext) RecordSuccess() {
	erc.mu.Lock()
	defer erc.mu.Unlock()

	if erc.consecutiveErrors > 0 {
		erc.logger.Debug("Success after errors", "consecutive_errors", erc.consecutiveErrors)
		erc.consecutiveErrors = 0
	}
}

// ShouldStopDueToErrors checks if execution should stop due to errors.
func (erc *ErrorRecoveryContext) ShouldStopDueToErrors() bool {
	erc.mu.RLock()
	defer erc.mu.RUnlock()
	return erc.consecutiveErrors >= erc.maxConsecutiveErrors
}

// Reset resets the error recovery context.
func (erc *ErrorRecoveryContext) Reset() {
	erc.mu.Lock()
	defer erc.mu.Unlock()

	erc.consecutiveErrors = 0
	for _, cb := range erc.circuitBreakers {
		cb.Reset()
	}
	erc.logger.Info("Error recovery context reset")
}

// GetStatus returns the current status of the error recovery context.
func (erc *ErrorRecoveryContext) GetStatus() map[string]interface{} {
	erc.mu.RLock()
	defer erc.mu.RUnlock()

	circuitStatuses := make(map[string]string)
	for key, cb := range erc.circuitBreakers {
		circuitStatuses[key] = cb.GetState().String()
	}

	return map[string]interface{}{
		"consecutive_errors":     erc.consecutiveErrors,
		"max_consecutive_errors": erc.maxConsecutiveErrors,
		"circuit_breakers":       circuitStatuses,
		"should_stop":            erc.consecutiveErrors >= erc.maxConsecutiveErrors,
	}
}

// addJitter adds random jitter to a duration.
func addJitter(duration time.Duration, factor float64) time.Duration {
	jitter := float64(duration) * factor * (rand.Float64()*2 - 1)
	return time.Duration(float64(duration) + jitter)
}
