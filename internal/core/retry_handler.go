package core

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// RetryHandler provides intelligent retry logic with exponential backoff and jitter
type RetryHandler struct {
	config       RetryConfig
	errorHandler *ErrorHandler
}

// RetryOperation defines the signature for operations that can be retried
type RetryOperation func(ctx context.Context, attempt int) (interface{}, error)

// RetryContext provides context information during retry operations
type RetryContext struct {
	Attempt       int           `json:"attempt"`
	MaxRetries    int           `json:"maxRetries"`
	LastError     error         `json:"-"`
	TotalDuration time.Duration `json:"totalDuration"`
	NextDelay     time.Duration `json:"nextDelay"`
}

// NewRetryHandler creates a new retry handler with the given configuration
func NewRetryHandler(config RetryConfig) *RetryHandler {
	return &RetryHandler{
		config:       config,
		errorHandler: NewErrorHandler(10), // Keep last 10 errors
	}
}

// Execute performs an operation with intelligent retry logic
func (rh *RetryHandler) Execute(ctx context.Context, operation RetryOperation) (interface{}, error) {
	startTime := time.Now()
	var lastErr error

	for attempt := 0; attempt <= rh.config.MaxRetries; attempt++ {
		// Create attempt-specific context with timeout
		attemptCtx, cancel := context.WithTimeout(ctx, rh.config.TimeoutPerTry)

		// Track retry context for monitoring
		retryCtx := &RetryContext{
			Attempt:       attempt + 1,
			MaxRetries:    rh.config.MaxRetries + 1,
			LastError:     lastErr,
			TotalDuration: time.Since(startTime),
		}

		// Execute the operation
		result, err := operation(attemptCtx, attempt)
		cancel() // Always cancel the attempt context

		// Success case
		if err == nil {
			return result, nil
		}

		lastErr = err
		retryCtx.LastError = err

		// Record the error for analysis
		if translationErr, ok := err.(*TranslationError); ok {
			rh.errorHandler.RecordError(translationErr)
		}

		// Check if we should retry
		if !rh.shouldRetry(err, attempt, retryCtx) {
			break
		}

		// Calculate and apply delay before next attempt
		delay := rh.calculateRetryDelay(attempt, err)
		retryCtx.NextDelay = delay

		// Check context before delay
		if ctx.Err() != nil {
			return nil, NewTranslationErrorWithCause(
				ErrorTypeTimeout,
				"CONTEXT_CANCELLED",
				"operation cancelled during retry delay",
				ctx.Err(),
			)
		}

		// Apply the delay with context cancellation support
		if err := rh.waitWithContext(ctx, delay); err != nil {
			return nil, err
		}
	}

	// All retries exhausted
	return nil, NewMaxRetriesError(rh.config.MaxRetries+1, lastErr)
}

// shouldRetry determines whether an operation should be retried
func (rh *RetryHandler) shouldRetry(err error, attempt int, retryCtx *RetryContext) bool {
	// Don't retry if we've exceeded max attempts
	if attempt >= rh.config.MaxRetries {
		return false
	}

	// Check if error type is retryable
	if translationErr, ok := err.(*TranslationError); ok {
		if !translationErr.IsRetryable() {
			return false
		}

		// Apply error-specific retry logic
		return rh.shouldRetryErrorType(translationErr, retryCtx)
	}

	// Default to retryable for unknown error types (conservative approach)
	return true
}

// shouldRetryErrorType applies error-type-specific retry logic
func (rh *RetryHandler) shouldRetryErrorType(err *TranslationError, retryCtx *RetryContext) bool {
	switch err.Type {
	case ErrorTypeNetwork:
		// Retry network errors with exponential backoff
		return true

	case ErrorTypeTimeout:
		// Retry timeouts but with longer delays
		return true

	case ErrorTypeRateLimit:
		// Retry rate limits but with longer delays and jitter
		return true

	case ErrorTypeServerError:
		// Retry server errors (5xx) but not too aggressively
		return retryCtx.Attempt <= rh.config.MaxRetries/2 // Only retry first half of attempts

	case ErrorTypeAPIResponse:
		// Retry API response errors sparingly
		return retryCtx.Attempt <= 2 // Maximum 2 retries for API response errors

	case ErrorTypeAuth:
		// Don't retry auth errors - they need user intervention
		return false

	case ErrorTypeValidation:
		// Don't retry validation errors - input is wrong
		return false

	case ErrorTypeQuotaExceeded:
		// Don't retry quota errors - need to wait for reset or upgrade
		return false

	case ErrorTypeCircuitBreaker:
		// Don't retry circuit breaker errors - need to wait for reset
		return false

	default:
		// Conservative default - retry unknown errors
		return true
	}
}

// calculateRetryDelay calculates the delay before the next retry attempt
func (rh *RetryHandler) calculateRetryDelay(attempt int, err error) time.Duration {
	// Base calculation with exponential backoff
	baseDelay := rh.config.BaseDelay
	backoffFactor := rh.config.BackoffFactor

	// Calculate exponential backoff
	delay := time.Duration(float64(baseDelay) * math.Pow(backoffFactor, float64(attempt)))

	// Apply error-specific delay modifications
	if translationErr, ok := err.(*TranslationError); ok {
		delay = rh.adjustDelayForErrorType(delay, translationErr)
	}

	// Cap at maximum delay
	if delay > rh.config.MaxDelay {
		delay = rh.config.MaxDelay
	}

	// Add jitter if enabled
	if rh.config.JitterEnabled {
		delay = rh.addJitter(delay)
	}

	return delay
}

// adjustDelayForErrorType adjusts retry delay based on specific error types
func (rh *RetryHandler) adjustDelayForErrorType(baseDelay time.Duration, err *TranslationError) time.Duration {
	switch err.Type {
	case ErrorTypeRateLimit:
		// For rate limits, use longer delays
		multiplier := 2.0
		if retryAfter, exists := err.Context["retry_after"]; exists {
			if duration, parseErr := time.ParseDuration(retryAfter); parseErr == nil {
				return duration
			}
		}
		return time.Duration(float64(baseDelay) * multiplier)

	case ErrorTypeTimeout:
		// For timeouts, use moderate increase
		return time.Duration(float64(baseDelay) * 1.5)

	case ErrorTypeNetwork:
		// For network errors, use standard delay
		return baseDelay

	case ErrorTypeServerError:
		// For server errors, use longer delays to avoid overwhelming the server
		return time.Duration(float64(baseDelay) * 3.0)

	default:
		return baseDelay
	}
}

// addJitter adds randomized jitter to the delay to avoid thundering herd problem
func (rh *RetryHandler) addJitter(delay time.Duration) time.Duration {
	// Calculate jitter range (±10% of delay)
	jitterRange := float64(delay) * 0.1

	// Generate random jitter
	jitter := time.Duration((rand.Float64()*2.0 - 1.0) * jitterRange)

	// Apply jitter
	finalDelay := delay + jitter

	// Ensure delay is not negative
	if finalDelay < 0 {
		finalDelay = delay / 2
	}

	return finalDelay
}

// waitWithContext waits for the specified duration with context cancellation support
func (rh *RetryHandler) waitWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return NewTranslationErrorWithCause(
			ErrorTypeTimeout,
			"CONTEXT_CANCELLED",
			"operation cancelled during retry delay",
			ctx.Err(),
		)
	case <-timer.C:
		return nil
	}
}

// ExecuteWithCallback performs an operation with retry logic and progress callbacks
func (rh *RetryHandler) ExecuteWithCallback(
	ctx context.Context,
	operation RetryOperation,
	onRetry func(RetryContext),
) (interface{}, error) {
	startTime := time.Now()
	var lastErr error

	for attempt := 0; attempt <= rh.config.MaxRetries; attempt++ {
		// Create attempt-specific context with timeout
		attemptCtx, cancel := context.WithTimeout(ctx, rh.config.TimeoutPerTry)

		// Track retry context
		retryCtx := RetryContext{
			Attempt:       attempt + 1,
			MaxRetries:    rh.config.MaxRetries + 1,
			LastError:     lastErr,
			TotalDuration: time.Since(startTime),
		}

		// Execute the operation
		result, err := operation(attemptCtx, attempt)
		cancel()

		// Success case
		if err == nil {
			return result, nil
		}

		lastErr = err
		retryCtx.LastError = err

		// Record error
		if translationErr, ok := err.(*TranslationError); ok {
			rh.errorHandler.RecordError(translationErr)
		}

		// Check if we should retry
		if !rh.shouldRetry(err, attempt, &retryCtx) {
			break
		}

		// Calculate delay
		delay := rh.calculateRetryDelay(attempt, err)
		retryCtx.NextDelay = delay

		// Call retry callback if provided
		if onRetry != nil {
			onRetry(retryCtx)
		}

		// Check context before delay
		if ctx.Err() != nil {
			return nil, NewTranslationErrorWithCause(
				ErrorTypeTimeout,
				"CONTEXT_CANCELLED",
				"operation cancelled during retry delay",
				ctx.Err(),
			)
		}

		// Apply delay
		if err := rh.waitWithContext(ctx, delay); err != nil {
			return nil, err
		}
	}

	return nil, NewMaxRetriesError(rh.config.MaxRetries+1, lastErr)
}

// GetRetryStats returns statistics about retry operations
func (rh *RetryHandler) GetRetryStats() *RetryStats {
	return &RetryStats{
		ErrorCounts:   rh.errorHandler.ErrorCounts,
		LastErrors:    rh.errorHandler.LastErrors,
		Configuration: rh.config,
	}
}

// RetryStats contains statistics about retry operations
type RetryStats struct {
	ErrorCounts   map[TranslationErrorType]int64 `json:"errorCounts"`
	LastErrors    []*TranslationError            `json:"lastErrors"`
	Configuration RetryConfig                    `json:"configuration"`
}

// AdaptiveRetryHandler provides adaptive retry logic that adjusts based on error patterns
type AdaptiveRetryHandler struct {
	*RetryHandler
	successRate     float64
	recentFailures  int
	adaptiveConfig  AdaptiveRetryConfig
	lastStatsUpdate time.Time
}

// AdaptiveRetryConfig contains configuration for adaptive retry behavior
type AdaptiveRetryConfig struct {
	MinRetries            int           `json:"minRetries"`            // Minimum number of retries
	MaxRetries            int           `json:"maxRetries"`            // Maximum number of retries
	SuccessRateThreshold  float64       `json:"successRateThreshold"`  // Success rate below which to increase retries
	FailureCountThreshold int           `json:"failureCountThreshold"` // Consecutive failures to trigger adaptation
	AdaptationWindow      time.Duration `json:"adaptationWindow"`      // Time window for calculating success rate
	BaseDelayMultiplier   float64       `json:"baseDelayMultiplier"`   // Multiplier for base delay during adaptation
	MaxDelayMultiplier    float64       `json:"maxDelayMultiplier"`    // Maximum delay multiplier
}

// DefaultAdaptiveRetryConfig returns sensible defaults for adaptive retry
func DefaultAdaptiveRetryConfig() AdaptiveRetryConfig {
	return AdaptiveRetryConfig{
		MinRetries:            1,
		MaxRetries:            10,
		SuccessRateThreshold:  0.8, // 80% success rate
		FailureCountThreshold: 3,   // 3 consecutive failures
		AdaptationWindow:      5 * time.Minute,
		BaseDelayMultiplier:   1.5, // 50% longer delays
		MaxDelayMultiplier:    3.0, // Up to 3x longer delays
	}
}

// NewAdaptiveRetryHandler creates a new adaptive retry handler
func NewAdaptiveRetryHandler(baseConfig RetryConfig, adaptiveConfig AdaptiveRetryConfig) *AdaptiveRetryHandler {
	return &AdaptiveRetryHandler{
		RetryHandler:    NewRetryHandler(baseConfig),
		successRate:     1.0, // Start optimistic
		recentFailures:  0,
		adaptiveConfig:  adaptiveConfig,
		lastStatsUpdate: time.Now(),
	}
}

// Execute performs operation with adaptive retry logic
func (arh *AdaptiveRetryHandler) Execute(ctx context.Context, operation RetryOperation) (interface{}, error) {
	// Update adaptive configuration based on recent performance
	arh.updateAdaptiveConfig()

	// Execute with current configuration
	result, err := arh.RetryHandler.Execute(ctx, operation)

	// Update statistics
	arh.updateStatistics(err == nil)

	return result, err
}

// updateAdaptiveConfig adjusts retry configuration based on recent performance
func (arh *AdaptiveRetryHandler) updateAdaptiveConfig() {
	now := time.Now()

	// Only update if enough time has passed
	if now.Sub(arh.lastStatsUpdate) < arh.adaptiveConfig.AdaptationWindow/4 {
		return
	}

	// Calculate recent error patterns
	recentPatterns := arh.errorHandler.GetRecentErrorPatterns(arh.adaptiveConfig.AdaptationWindow)
	totalRecentRequests := 0
	for _, count := range recentPatterns {
		totalRecentRequests += count
	}

	// Adjust retry limits based on success rate
	if arh.successRate < arh.adaptiveConfig.SuccessRateThreshold {
		// Poor success rate - increase retries and delays
		newMaxRetries := int(float64(arh.config.MaxRetries) * 1.5)
		if newMaxRetries > arh.adaptiveConfig.MaxRetries {
			newMaxRetries = arh.adaptiveConfig.MaxRetries
		}
		arh.config.MaxRetries = newMaxRetries

		// Increase base delay
		arh.config.BaseDelay = time.Duration(float64(arh.config.BaseDelay) * arh.adaptiveConfig.BaseDelayMultiplier)
		if arh.config.BaseDelay > time.Duration(float64(DefaultRetryConfig().BaseDelay)*arh.adaptiveConfig.MaxDelayMultiplier) {
			arh.config.BaseDelay = time.Duration(float64(DefaultRetryConfig().BaseDelay) * arh.adaptiveConfig.MaxDelayMultiplier)
		}
	} else if arh.successRate > 0.95 && arh.recentFailures == 0 {
		// Excellent success rate - reduce retries for efficiency
		newMaxRetries := arh.config.MaxRetries - 1
		if newMaxRetries < arh.adaptiveConfig.MinRetries {
			newMaxRetries = arh.adaptiveConfig.MinRetries
		}
		arh.config.MaxRetries = newMaxRetries

		// Reset base delay to default
		arh.config.BaseDelay = DefaultRetryConfig().BaseDelay
	}

	arh.lastStatsUpdate = now
}

// updateStatistics updates internal statistics based on operation results
func (arh *AdaptiveRetryHandler) updateStatistics(success bool) {
	if success {
		arh.recentFailures = 0
		// Update success rate with exponential moving average
		arh.successRate = arh.successRate*0.9 + 0.1 // 90% weight to previous, 10% to current success
	} else {
		arh.recentFailures++
		// Update success rate with exponential moving average
		arh.successRate = arh.successRate * 0.9 // 90% weight to previous, current failure = 0
	}

	// Ensure success rate stays within bounds
	if arh.successRate > 1.0 {
		arh.successRate = 1.0
	} else if arh.successRate < 0.0 {
		arh.successRate = 0.0
	}
}

// GetAdaptiveStats returns current adaptive retry statistics
func (arh *AdaptiveRetryHandler) GetAdaptiveStats() *AdaptiveRetryStats {
	return &AdaptiveRetryStats{
		RetryStats:      *arh.GetRetryStats(),
		SuccessRate:     arh.successRate,
		RecentFailures:  arh.recentFailures,
		AdaptiveConfig:  arh.adaptiveConfig,
		LastStatsUpdate: arh.lastStatsUpdate,
	}
}

// AdaptiveRetryStats contains statistics about adaptive retry operations
type AdaptiveRetryStats struct {
	RetryStats      `json:",inline"`
	SuccessRate     float64             `json:"successRate"`
	RecentFailures  int                 `json:"recentFailures"`
	AdaptiveConfig  AdaptiveRetryConfig `json:"adaptiveConfig"`
	LastStatsUpdate time.Time           `json:"lastStatsUpdate"`
}
