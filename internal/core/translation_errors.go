package core

import (
	"fmt"
	"time"
)

// TranslationErrorType represents different categories of translation errors
type TranslationErrorType int

const (
	ErrorTypeUnknown TranslationErrorType = iota
	ErrorTypeValidation
	ErrorTypeAuth
	ErrorTypeNetwork
	ErrorTypeTimeout
	ErrorTypeRateLimit
	ErrorTypeQuotaExceeded
	ErrorTypeServerError
	ErrorTypeAPIResponse
	ErrorTypeCircuitBreaker
	ErrorTypeMaxRetriesExceeded
)

// String returns the string representation of TranslationErrorType
func (t TranslationErrorType) String() string {
	switch t {
	case ErrorTypeValidation:
		return "validation"
	case ErrorTypeAuth:
		return "authentication"
	case ErrorTypeNetwork:
		return "network"
	case ErrorTypeTimeout:
		return "timeout"
	case ErrorTypeRateLimit:
		return "rate_limit"
	case ErrorTypeQuotaExceeded:
		return "quota_exceeded"
	case ErrorTypeServerError:
		return "server_error"
	case ErrorTypeAPIResponse:
		return "api_response"
	case ErrorTypeCircuitBreaker:
		return "circuit_breaker"
	case ErrorTypeMaxRetriesExceeded:
		return "max_retries_exceeded"
	default:
		return "unknown"
	}
}

// IsRetryable returns whether this error type should trigger retries
func (t TranslationErrorType) IsRetryable() bool {
	switch t {
	case ErrorTypeNetwork, ErrorTypeTimeout, ErrorTypeRateLimit, ErrorTypeServerError:
		return true
	case ErrorTypeAuth, ErrorTypeValidation, ErrorTypeQuotaExceeded, ErrorTypeCircuitBreaker, ErrorTypeMaxRetriesExceeded:
		return false
	default:
		return false // Conservative approach for unknown errors
	}
}

// Severity returns the severity level of the error
func (t TranslationErrorType) Severity() string {
	switch t {
	case ErrorTypeValidation:
		return "warning"
	case ErrorTypeAuth, ErrorTypeQuotaExceeded:
		return "critical"
	case ErrorTypeNetwork, ErrorTypeTimeout, ErrorTypeRateLimit:
		return "error"
	case ErrorTypeServerError, ErrorTypeAPIResponse:
		return "error"
	case ErrorTypeCircuitBreaker, ErrorTypeMaxRetriesExceeded:
		return "critical"
	default:
		return "error"
	}
}

// TranslationError represents a comprehensive translation error with classification
type TranslationError struct {
	Type        TranslationErrorType `json:"type"`
	Message     string               `json:"message"`
	Code        string               `json:"code"`
	Cause       error                `json:"-"` // Original error, not serializable
	Timestamp   time.Time            `json:"timestamp"`
	Context     map[string]string    `json:"context,omitempty"`
	Suggestions []string             `json:"suggestions,omitempty"`
}

// Error implements the error interface
func (e *TranslationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s (caused by: %v)", e.Type.String(), e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Type.String(), e.Code, e.Message)
}

// Unwrap returns the underlying error for error chain compatibility
func (e *TranslationError) Unwrap() error {
	return e.Cause
}

// IsRetryable returns whether this specific error should trigger retries
func (e *TranslationError) IsRetryable() bool {
	return e.Type.IsRetryable()
}

// GetSeverity returns the severity level of this error
func (e *TranslationError) GetSeverity() string {
	return e.Type.Severity()
}

// WithContext adds contextual information to the error
func (e *TranslationError) WithContext(key, value string) *TranslationError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

// WithSuggestion adds a suggestion for resolving the error
func (e *TranslationError) WithSuggestion(suggestion string) *TranslationError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// NewTranslationError creates a new translation error with default timestamp
func NewTranslationError(errorType TranslationErrorType, code, message string) *TranslationError {
	return &TranslationError{
		Type:      errorType,
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// NewTranslationErrorWithCause creates a new translation error with an underlying cause
func NewTranslationErrorWithCause(errorType TranslationErrorType, code, message string, cause error) *TranslationError {
	return &TranslationError{
		Type:      errorType,
		Code:      code,
		Message:   message,
		Cause:     cause,
		Timestamp: time.Now(),
	}
}

// Common translation errors

// NewValidationError creates a validation error
func NewValidationError(message string) *TranslationError {
	return NewTranslationError(ErrorTypeValidation, "VALIDATION_ERROR", message).
		WithSuggestion("Check input parameters and ensure they meet requirements")
}

// NewAuthError creates an authentication error
func NewAuthError(message string, cause error) *TranslationError {
	return NewTranslationErrorWithCause(ErrorTypeAuth, "AUTH_ERROR", message, cause).
		WithSuggestion("Verify API key configuration and permissions").
		WithSuggestion("Check if API key has expired or been revoked")
}

// NewNetworkError creates a network error
func NewNetworkError(message string, cause error) *TranslationError {
	return NewTranslationErrorWithCause(ErrorTypeNetwork, "NETWORK_ERROR", message, cause).
		WithSuggestion("Check internet connectivity").
		WithSuggestion("Verify API endpoint URL is correct").
		WithSuggestion("Check if firewall is blocking the connection")
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(message string, cause error) *TranslationError {
	return NewTranslationErrorWithCause(ErrorTypeTimeout, "TIMEOUT_ERROR", message, cause).
		WithSuggestion("Increase timeout configuration").
		WithSuggestion("Check network latency to API endpoint").
		WithSuggestion("Retry the request")
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string, retryAfter time.Duration) *TranslationError {
	err := NewTranslationError(ErrorTypeRateLimit, "RATE_LIMIT_ERROR", message).
		WithSuggestion("Wait before retrying the request").
		WithSuggestion("Consider implementing request queuing")

	if retryAfter > 0 {
		err.WithContext("retry_after", retryAfter.String())
		err.WithSuggestion(fmt.Sprintf("Retry after %v", retryAfter))
	}

	return err
}

// NewQuotaExceededError creates a quota exceeded error
func NewQuotaExceededError(message string) *TranslationError {
	return NewTranslationError(ErrorTypeQuotaExceeded, "QUOTA_EXCEEDED", message).
		WithSuggestion("Check your API usage and billing status").
		WithSuggestion("Upgrade your API plan if needed").
		WithSuggestion("Wait for quota reset period")
}

// NewServerError creates a server error
func NewServerError(message string, statusCode int, cause error) *TranslationError {
	err := NewTranslationErrorWithCause(ErrorTypeServerError, "SERVER_ERROR", message, cause).
		WithSuggestion("Retry the request").
		WithSuggestion("Check API service status")

	if statusCode > 0 {
		err.WithContext("status_code", fmt.Sprintf("%d", statusCode))
	}

	return err
}

// NewAPIResponseError creates an API response error
func NewAPIResponseError(message string) *TranslationError {
	return NewTranslationError(ErrorTypeAPIResponse, "API_RESPONSE_ERROR", message).
		WithSuggestion("Check API response format").
		WithSuggestion("Verify request parameters").
		WithSuggestion("Contact API support if issue persists")
}

// NewCircuitBreakerError creates a circuit breaker error
func NewCircuitBreakerError(message string) *TranslationError {
	return NewTranslationError(ErrorTypeCircuitBreaker, "CIRCUIT_BREAKER_ERROR", message).
		WithSuggestion("Wait for circuit breaker to reset").
		WithSuggestion("Check API service health").
		WithSuggestion("Review error patterns that triggered the circuit breaker")
}

// NewMaxRetriesError creates a max retries exceeded error
func NewMaxRetriesError(attempts int, lastError error) *TranslationError {
	message := fmt.Sprintf("maximum retry attempts (%d) exceeded", attempts)
	err := NewTranslationErrorWithCause(ErrorTypeMaxRetriesExceeded, "MAX_RETRIES_EXCEEDED", message, lastError).
		WithSuggestion("Check underlying error cause").
		WithSuggestion("Consider increasing retry limits if appropriate").
		WithSuggestion("Review retry strategy configuration")

	err.WithContext("attempts", fmt.Sprintf("%d", attempts))

	return err
}

// ErrorHandler provides utilities for handling translation errors
type ErrorHandler struct {
	ErrorCounts   map[TranslationErrorType]int64 `json:"errorCounts"`
	LastErrors    []*TranslationError            `json:"lastErrors"`
	MaxLastErrors int                            `json:"maxLastErrors"`
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(maxLastErrors int) *ErrorHandler {
	return &ErrorHandler{
		ErrorCounts:   make(map[TranslationErrorType]int64),
		LastErrors:    make([]*TranslationError, 0, maxLastErrors),
		MaxLastErrors: maxLastErrors,
	}
}

// RecordError records an error for tracking and analysis
func (h *ErrorHandler) RecordError(err *TranslationError) {
	// Update error counts
	h.ErrorCounts[err.Type]++

	// Add to recent errors (with rotation)
	h.LastErrors = append(h.LastErrors, err)
	if len(h.LastErrors) > h.MaxLastErrors {
		h.LastErrors = h.LastErrors[1:]
	}
}

// GetErrorRate returns the error rate for a specific error type
func (h *ErrorHandler) GetErrorRate(errorType TranslationErrorType, totalRequests int64) float64 {
	if totalRequests == 0 {
		return 0.0
	}

	errorCount := h.ErrorCounts[errorType]
	return float64(errorCount) / float64(totalRequests)
}

// GetTopErrors returns the most frequent error types
func (h *ErrorHandler) GetTopErrors(limit int) []struct {
	Type  TranslationErrorType
	Count int64
} {
	type errorStat struct {
		Type  TranslationErrorType
		Count int64
	}

	var stats []errorStat
	for errorType, count := range h.ErrorCounts {
		stats = append(stats, errorStat{Type: errorType, Count: count})
	}

	// Simple sort by count (descending)
	for i := 0; i < len(stats)-1; i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[i].Count < stats[j].Count {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	if limit > 0 && limit < len(stats) {
		stats = stats[:limit]
	}

	// Convert to return type
	result := make([]struct {
		Type  TranslationErrorType
		Count int64
	}, len(stats))

	for i, stat := range stats {
		result[i] = struct {
			Type  TranslationErrorType
			Count int64
		}{Type: stat.Type, Count: stat.Count}
	}

	return result
}

// HasRecentErrors checks if there have been recent errors of a specific type
func (h *ErrorHandler) HasRecentErrors(errorType TranslationErrorType, window time.Duration) bool {
	cutoff := time.Now().Add(-window)

	for _, err := range h.LastErrors {
		if err.Type == errorType && err.Timestamp.After(cutoff) {
			return true
		}
	}

	return false
}

// GetRecentErrorPatterns analyzes recent errors for patterns
func (h *ErrorHandler) GetRecentErrorPatterns(window time.Duration) map[TranslationErrorType]int {
	cutoff := time.Now().Add(-window)
	patterns := make(map[TranslationErrorType]int)

	for _, err := range h.LastErrors {
		if err.Timestamp.After(cutoff) {
			patterns[err.Type]++
		}
	}

	return patterns
}

// SuggestRecoveryAction suggests recovery actions based on error patterns
func (h *ErrorHandler) SuggestRecoveryAction(errorType TranslationErrorType) []string {
	switch errorType {
	case ErrorTypeNetwork:
		return []string{
			"Check network connectivity",
			"Verify DNS resolution",
			"Test with different network",
		}
	case ErrorTypeAuth:
		return []string{
			"Verify API key is valid",
			"Check API key permissions",
			"Regenerate API key if necessary",
		}
	case ErrorTypeRateLimit:
		return []string{
			"Implement request queuing",
			"Add delay between requests",
			"Consider upgrading API plan",
		}
	case ErrorTypeQuotaExceeded:
		return []string{
			"Check billing status",
			"Upgrade API plan",
			"Wait for quota reset",
		}
	case ErrorTypeTimeout:
		return []string{
			"Increase timeout values",
			"Check API endpoint latency",
			"Consider request optimization",
		}
	default:
		return []string{
			"Review error details",
			"Check API documentation",
			"Contact support if issue persists",
		}
	}
}
