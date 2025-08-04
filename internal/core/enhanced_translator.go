package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

// OpenAIInterface defines the contract for OpenAI API interactions (testable)
type OpenAIInterface interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// OpenAIClientWrapper wraps the actual OpenAI client to implement our interface
type OpenAIClientWrapper struct {
	client *openai.Client
}

func (w *OpenAIClientWrapper) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return w.client.CreateChatCompletion(ctx, req)
}

// EnhancedTranslationService provides resilient translation with caching and monitoring
type EnhancedTranslationService struct {
	client      OpenAIInterface
	config      *EnhancedConfig
	keyManager  *SecureKeyManager
	retryConfig RetryConfig
	breaker     *gobreaker.CircuitBreaker
	cache       *EnhancedTranslationCache
	rateLimiter *rate.Limiter
	metrics     *TranslationMetrics
	mu          sync.RWMutex
}

// EnhancedTranslationResult contains comprehensive translation result information
type EnhancedTranslationResult struct {
	// Core translation data
	OriginalText   string    `json:"originalText"`
	TranslatedText string    `json:"translatedText"`
	SourceLanguage string    `json:"sourceLanguage"`
	TargetLanguage string    `json:"targetLanguage"`
	Timestamp      time.Time `json:"timestamp"`

	// Enhanced metadata
	TokensUsed       int           `json:"tokensUsed"`
	Model            string        `json:"model"`
	Duration         time.Duration `json:"duration"`
	Cached           bool          `json:"cached"`
	Confidence       float64       `json:"confidence,omitempty"`
	DetectedLanguage string        `json:"detectedLanguage,omitempty"`

	// Performance metrics
	AttemptCount   int    `json:"attemptCount"`
	ApiProvider    string `json:"apiProvider"`
	CircuitBreaker string `json:"circuitBreakerState,omitempty"`
}

// EnhancedTranslationCache provides intelligent translation caching
type EnhancedTranslationCache struct {
	entries       map[string]*CacheEntry
	maxSize       int
	ttl           time.Duration
	accessMap     map[string]time.Time
	mu            sync.RWMutex
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// CacheEntry represents a cached translation result
type CacheEntry struct {
	Result     *EnhancedTranslationResult
	CreatedAt  time.Time
	AccessedAt time.Time
	HitCount   int
}

// TranslationMetrics tracks translation service performance
type TranslationMetrics struct {
	TotalRequests       int64         `json:"totalRequests"`
	SuccessfulRequests  int64         `json:"successfulRequests"`
	FailedRequests      int64         `json:"failedRequests"`
	CacheHits           int64         `json:"cacheHits"`
	CacheMisses         int64         `json:"cacheMisses"`
	AverageLatency      time.Duration `json:"averageLatency"`
	TotalTokensUsed     int64         `json:"totalTokensUsed"`
	CircuitBreakerTrips int64         `json:"circuitBreakerTrips"`

	// Error breakdown
	ErrorsByType map[string]int64 `json:"errorsByType"`

	mu sync.RWMutex
}

// NewEnhancedTranslationService creates a new enhanced translation service
func NewEnhancedTranslationService(config *EnhancedConfig, keyManager *SecureKeyManager) (*EnhancedTranslationService, error) {
	// Get API key from keyring
	apiKey, err := keyManager.GetAPIKey(config.OpenAI.APIKeyAlias)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve API key: %w", err)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("no API key configured")
	}

	// Create OpenAI client with configuration
	clientConfig := openai.DefaultConfig(apiKey)
	clientConfig.BaseURL = config.OpenAI.BaseURL
	client := openai.NewClientWithConfig(clientConfig)

	// Wrap client to implement interface
	wrappedClient := &OpenAIClientWrapper{client: client}

	// Configure circuit breaker
	breakerSettings := gobreaker.Settings{
		Name:        "openai-translation",
		MaxRequests: 5,
		Interval:    time.Minute,
		Timeout:     2 * time.Minute,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip if failure rate > 60% and we have at least 5 requests
			return counts.Requests >= 5 && counts.TotalFailures > counts.Requests*60/100
		},
	}

	breaker := gobreaker.NewCircuitBreaker(breakerSettings)

	// Create translation cache
	cache := NewEnhancedTranslationCache(config.Translation.CacheSize, time.Duration(config.Translation.CacheTTL)*time.Second)

	// Create rate limiter (60 requests per minute for OpenAI)
	rateLimiter := rate.NewLimiter(rate.Every(time.Second), 10)

	// Initialize metrics
	metrics := &TranslationMetrics{
		ErrorsByType: make(map[string]int64),
	}

	service := &EnhancedTranslationService{
		client:      wrappedClient,
		config:      config,
		keyManager:  keyManager,
		retryConfig: DefaultRetryConfig(),
		breaker:     breaker,
		cache:       cache,
		rateLimiter: rateLimiter,
		metrics:     metrics,
	}

	return service, nil
}

// NewEnhancedTranslationCache creates a new enhanced translation cache
func NewEnhancedTranslationCache(maxSize int, ttl time.Duration) *EnhancedTranslationCache {
	cache := &EnhancedTranslationCache{
		entries:     make(map[string]*CacheEntry),
		maxSize:     maxSize,
		ttl:         ttl,
		accessMap:   make(map[string]time.Time),
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup routine
	cache.cleanupTicker = time.NewTicker(15 * time.Minute)
	go cache.cleanupRoutine()

	return cache
}

// TranslateText performs enhanced translation with resilience features
func (s *EnhancedTranslationService) TranslateText(ctx context.Context, text, textType string) (*EnhancedTranslationResult, error) {
	startTime := time.Now()

	s.metrics.mu.Lock()
	s.metrics.TotalRequests++
	s.metrics.mu.Unlock()

	if text == "" {
		return nil, &TranslationError{
			Type:    ErrorTypeValidation,
			Message: "text cannot be empty",
			Code:    "EMPTY_TEXT",
		}
	}

	// Check cache first
	cacheKey := s.generateCacheKey(text, textType, s.config.Translation.TargetLanguage)
	if cached := s.cache.Get(cacheKey); cached != nil {
		s.metrics.mu.Lock()
		s.metrics.CacheHits++
		s.metrics.SuccessfulRequests++
		s.metrics.mu.Unlock()

		// Return cached result with updated metadata
		result := *cached.Result
		result.Cached = true
		result.Duration = time.Since(startTime)
		return &result, nil
	}

	s.metrics.mu.Lock()
	s.metrics.CacheMisses++
	s.metrics.mu.Unlock()

	// Apply rate limiting
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, &TranslationError{
			Type:    ErrorTypeRateLimit,
			Message: "rate limit exceeded",
			Code:    "RATE_LIMITED",
			Cause:   err,
		}
	}

	// Execute translation with circuit breaker
	result, err := s.executeWithCircuitBreaker(ctx, text, textType, startTime)
	if err != nil {
		s.metrics.mu.Lock()
		s.metrics.FailedRequests++
		s.metrics.recordError(err)
		s.metrics.mu.Unlock()
		return nil, err
	}

	// Cache successful result
	s.cache.Set(cacheKey, result)

	// Update metrics
	s.metrics.mu.Lock()
	s.metrics.SuccessfulRequests++
	s.metrics.TotalTokensUsed += int64(result.TokensUsed)
	s.metrics.updateAverageLatency(result.Duration)
	s.metrics.mu.Unlock()

	return result, nil
}

// executeWithCircuitBreaker executes translation with circuit breaker protection
func (s *EnhancedTranslationService) executeWithCircuitBreaker(ctx context.Context, text, textType string, startTime time.Time) (*EnhancedTranslationResult, error) {
	result, err := s.breaker.Execute(func() (interface{}, error) {
		return s.executeTranslationWithRetry(ctx, text, textType, startTime)
	})

	if err != nil {
		// Check if circuit breaker tripped
		if s.breaker.State() == gobreaker.StateOpen {
			s.metrics.mu.Lock()
			s.metrics.CircuitBreakerTrips++
			s.metrics.mu.Unlock()

			return nil, &TranslationError{
				Type:    ErrorTypeCircuitBreaker,
				Message: "circuit breaker is open - service temporarily unavailable",
				Code:    "CIRCUIT_BREAKER_OPEN",
				Cause:   err,
			}
		}

		return nil, err
	}

	return result.(*EnhancedTranslationResult), nil
}

// executeTranslationWithRetry performs translation with intelligent retry logic
func (s *EnhancedTranslationService) executeTranslationWithRetry(ctx context.Context, text, textType string, startTime time.Time) (*EnhancedTranslationResult, error) {
	var lastErr error
	attemptCount := 0

	for attempt := 0; attempt <= s.retryConfig.MaxRetries; attempt++ {
		attemptCount++

		// Apply retry delay with jitter (except for first attempt)
		if attempt > 0 {
			delay := s.calculateRetryDelay(attempt)
			select {
			case <-ctx.Done():
				return nil, &TranslationError{
					Type:    ErrorTypeTimeout,
					Message: "context cancelled during retry delay",
					Code:    "CONTEXT_CANCELLED",
					Cause:   ctx.Err(),
				}
			case <-time.After(delay):
				// Continue with retry
			}
		}

		// Execute single translation attempt
		result, err := s.performSingleTranslation(ctx, text, textType, attemptCount, startTime)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !s.isRetryableError(err) {
			break
		}

		// Check context before next retry
		if ctx.Err() != nil {
			return nil, &TranslationError{
				Type:    ErrorTypeTimeout,
				Message: "context cancelled",
				Code:    "CONTEXT_CANCELLED",
				Cause:   ctx.Err(),
			}
		}
	}

	return nil, &TranslationError{
		Type:    ErrorTypeMaxRetriesExceeded,
		Message: fmt.Sprintf("translation failed after %d attempts", attemptCount),
		Code:    "MAX_RETRIES_EXCEEDED",
		Cause:   lastErr,
	}
}

// performSingleTranslation executes a single translation request
func (s *EnhancedTranslationService) performSingleTranslation(ctx context.Context, text, textType string, attemptCount int, startTime time.Time) (*EnhancedTranslationResult, error) {
	// Create context with timeout
	requestCtx, cancel := context.WithTimeout(ctx, s.retryConfig.TimeoutPerTry)
	defer cancel()

	// Build translation prompt
	prompt := s.buildEnhancedTranslationPrompt(text, textType, s.config.Translation.TargetLanguage, s.config.Translation.ContextPrompt)

	// Prepare OpenAI request
	req := openai.ChatCompletionRequest{
		Model: s.config.OpenAI.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   s.config.OpenAI.MaxTokens,
		Temperature: float32(s.config.OpenAI.Temperature),
	}

	// Execute API call
	resp, err := s.client.CreateChatCompletion(requestCtx, req)
	if err != nil {
		return nil, s.classifyError(err)
	}

	if len(resp.Choices) == 0 {
		return nil, &TranslationError{
			Type:    ErrorTypeAPIResponse,
			Message: "no translation choices returned from API",
			Code:    "NO_CHOICES",
		}
	}

	translatedText := resp.Choices[0].Message.Content
	duration := time.Since(startTime)

	return &EnhancedTranslationResult{
		OriginalText:   text,
		TranslatedText: translatedText,
		SourceLanguage: "auto-detected",
		TargetLanguage: s.config.Translation.TargetLanguage,
		Timestamp:      time.Now(),
		TokensUsed:     resp.Usage.TotalTokens,
		Model:          s.config.OpenAI.Model,
		Duration:       duration,
		Cached:         false,
		AttemptCount:   attemptCount,
		ApiProvider:    s.config.OpenAI.BaseURL,
		CircuitBreaker: s.breaker.State().String(),
	}, nil
}

// Helper methods for enhanced functionality

// generateCacheKey creates a unique cache key for translation requests
func (s *EnhancedTranslationService) generateCacheKey(text, textType, targetLanguage string) string {
	return fmt.Sprintf("%s:%s:%s:%x", textType, targetLanguage, s.config.OpenAI.Model,
		[]byte(text)[:min(len(text), 50)]) // Use first 50 chars as key component
}

// calculateRetryDelay calculates retry delay with exponential backoff and jitter
func (s *EnhancedTranslationService) calculateRetryDelay(attempt int) time.Duration {
	baseDelay := s.retryConfig.BaseDelay
	delay := time.Duration(float64(baseDelay) * s.retryConfig.BackoffFactor * float64(attempt))

	// Cap at max delay
	if delay > s.retryConfig.MaxDelay {
		delay = s.retryConfig.MaxDelay
	}

	// Add jitter if enabled
	if s.retryConfig.JitterEnabled {
		jitter := time.Duration(float64(delay) * 0.1) // 10% jitter
		delay += time.Duration(float64(jitter) * (2.0*float64(time.Now().UnixNano()%1000)/1000.0 - 1.0))
	}

	return delay
}

// buildEnhancedTranslationPrompt creates an enhanced translation prompt
func (s *EnhancedTranslationService) buildEnhancedTranslationPrompt(text, textType, targetLanguage, contextPrompt string) string {
	if contextPrompt != "" {
		return fmt.Sprintf("%s\n\nText to translate (%s):\n%s", contextPrompt, textType, text)
	}

	return fmt.Sprintf("Translate the following %s to %s, preserving technical terms and maintaining the original meaning:\n\n%s",
		textType, targetLanguage, text)
}

// isRetryableError determines if an error should trigger a retry
func (s *EnhancedTranslationService) isRetryableError(err error) bool {
	if translationErr, ok := err.(*TranslationError); ok {
		switch translationErr.Type {
		case ErrorTypeNetwork, ErrorTypeTimeout, ErrorTypeRateLimit, ErrorTypeServerError:
			return true
		case ErrorTypeAuth, ErrorTypeValidation, ErrorTypeQuotaExceeded:
			return false
		default:
			return true // Default to retryable for unknown errors
		}
	}

	return true // Retry unknown error types
}

// classifyError classifies API errors into specific error types
func (s *EnhancedTranslationService) classifyError(err error) *TranslationError {
	// This would contain sophisticated error classification logic
	// For now, provide basic classification

	errMsg := err.Error()

	// Network/timeout errors
	if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "context deadline exceeded") {
		return &TranslationError{
			Type:    ErrorTypeTimeout,
			Message: "request timeout",
			Code:    "TIMEOUT",
			Cause:   err,
		}
	}

	// Authentication errors
	if strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "invalid_api_key") {
		return &TranslationError{
			Type:    ErrorTypeAuth,
			Message: "authentication failed",
			Code:    "AUTH_FAILED",
			Cause:   err,
		}
	}

	// Rate limit errors
	if strings.Contains(errMsg, "rate_limit") || strings.Contains(errMsg, "too_many_requests") {
		return &TranslationError{
			Type:    ErrorTypeRateLimit,
			Message: "rate limit exceeded",
			Code:    "RATE_LIMITED",
			Cause:   err,
		}
	}

	// Default to network error
	return &TranslationError{
		Type:    ErrorTypeNetwork,
		Message: "network error",
		Code:    "NETWORK_ERROR",
		Cause:   err,
	}
}

// Cache methods

// Get retrieves a cached translation result
func (c *EnhancedTranslationCache) Get(key string) *CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil
	}

	// Check TTL
	if time.Since(entry.CreatedAt) > c.ttl {
		// Entry expired, remove it
		go func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			delete(c.entries, key)
			delete(c.accessMap, key)
		}()
		return nil
	}

	// Update access time
	entry.AccessedAt = time.Now()
	entry.HitCount++
	c.accessMap[key] = entry.AccessedAt

	return entry
}

// Set stores a translation result in cache
func (c *EnhancedTranslationCache) Set(key string, result *EnhancedTranslationResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if cache is full
	if len(c.entries) >= c.maxSize {
		c.evictLRU()
	}

	entry := &CacheEntry{
		Result:     result,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		HitCount:   1,
	}

	c.entries[key] = entry
	c.accessMap[key] = entry.AccessedAt
}

// evictLRU evicts the least recently used cache entry
func (c *EnhancedTranslationCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time = time.Now()

	for key, accessTime := range c.accessMap {
		if accessTime.Before(oldestTime) {
			oldestTime = accessTime
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
		delete(c.accessMap, oldestKey)
	}
}

// cleanupRoutine periodically cleans expired entries
func (c *EnhancedTranslationCache) cleanupRoutine() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.cleanupExpired()
		case <-c.stopCleanup:
			c.cleanupTicker.Stop()
			return
		}
	}
}

// cleanupExpired removes expired cache entries
func (c *EnhancedTranslationCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.Sub(entry.CreatedAt) > c.ttl {
			delete(c.entries, key)
			delete(c.accessMap, key)
		}
	}
}

// Metrics methods

// recordError records an error in metrics
func (m *TranslationMetrics) recordError(err error) {
	if translationErr, ok := err.(*TranslationError); ok {
		m.ErrorsByType[translationErr.Type.String()]++
	} else {
		m.ErrorsByType["unknown"]++
	}
}

// updateAverageLatency updates the running average latency
func (m *TranslationMetrics) updateAverageLatency(duration time.Duration) {
	// Simple moving average calculation
	if m.SuccessfulRequests == 1 {
		m.AverageLatency = duration
	} else {
		m.AverageLatency = time.Duration((int64(m.AverageLatency)*(m.SuccessfulRequests-1) + int64(duration)) / m.SuccessfulRequests)
	}
}

// GetMetrics returns current translation metrics
func (s *EnhancedTranslationService) GetMetrics() *TranslationMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	// Return a copy to avoid race conditions
	metricsCopy := *s.metrics
	metricsCopy.ErrorsByType = make(map[string]int64)
	for k, v := range s.metrics.ErrorsByType {
		metricsCopy.ErrorsByType[k] = v
	}

	return &metricsCopy
}

// IsConfigured checks if the translation service is properly configured
func (s *EnhancedTranslationService) IsConfigured() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.client != nil && s.config != nil && s.config.Translation.Enabled
}

// TestConnection tests the API connection
func (s *EnhancedTranslationService) TestConnection(ctx context.Context) error {
	// Simple test translation
	testResult, err := s.TranslateText(ctx, "Hello, world!", "test")
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	if testResult.TranslatedText == "" {
		return fmt.Errorf("connection test returned empty result")
	}

	return nil
}

// Close cleans up resources
func (s *EnhancedTranslationService) Close() error {
	if s.cache != nil {
		close(s.cache.stopCleanup)
	}
	return nil
}

// Utility function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
