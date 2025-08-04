package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreakerManager manages multiple circuit breakers for different services
type CircuitBreakerManager struct {
	breakers map[string]*EnhancedCircuitBreaker
	config   CircuitBreakerConfig
	mu       sync.RWMutex
}

// EnhancedCircuitBreaker wraps gobreaker with additional functionality
type EnhancedCircuitBreaker struct {
	*gobreaker.CircuitBreaker
	name      string
	config    CircuitBreakerConfig
	metrics   *CircuitBreakerMetrics
	listeners []CircuitBreakerListener
	mu        sync.RWMutex
}

// CircuitBreakerMetrics tracks circuit breaker performance
type CircuitBreakerMetrics struct {
	TotalRequests      int64         `json:"totalRequests"`
	SuccessfulRequests int64         `json:"successfulRequests"`
	FailedRequests     int64         `json:"failedRequests"`
	TimeoutRequests    int64         `json:"timeoutRequests"`
	CircuitOpenCount   int64         `json:"circuitOpenCount"`
	CircuitCloseCount  int64         `json:"circuitCloseCount"`
	LastStateChange    time.Time     `json:"lastStateChange"`
	AverageLatency     time.Duration `json:"averageLatency"`
	mu                 sync.RWMutex
}

// CircuitBreakerListener defines the interface for circuit breaker event listeners
type CircuitBreakerListener interface {
	OnStateChange(name string, from, to gobreaker.State)
	OnRequestComplete(name string, duration time.Duration, success bool)
	OnCircuitOpen(name string, counts gobreaker.Counts)
	OnCircuitClose(name string)
}

// DefaultCircuitBreakerListener provides default logging for circuit breaker events
type DefaultCircuitBreakerListener struct {
	name string
}

// CircuitBreakerOperation defines the signature for operations executed with circuit breaker
type CircuitBreakerOperation func(ctx context.Context) (interface{}, error)

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(config CircuitBreakerConfig) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*EnhancedCircuitBreaker),
		config:   config,
	}
}

// GetOrCreateBreaker gets an existing circuit breaker or creates a new one
func (cbm *CircuitBreakerManager) GetOrCreateBreaker(name string) *EnhancedCircuitBreaker {
	cbm.mu.RLock()
	if breaker, exists := cbm.breakers[name]; exists {
		cbm.mu.RUnlock()
		return breaker
	}
	cbm.mu.RUnlock()

	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	// Double-check after acquiring write lock
	if breaker, exists := cbm.breakers[name]; exists {
		return breaker
	}

	// Create new circuit breaker
	breaker := NewEnhancedCircuitBreaker(name, cbm.config)
	cbm.breakers[name] = breaker

	return breaker
}

// GetBreaker returns an existing circuit breaker
func (cbm *CircuitBreakerManager) GetBreaker(name string) (*EnhancedCircuitBreaker, bool) {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	breaker, exists := cbm.breakers[name]
	return breaker, exists
}

// ListBreakerNames returns the names of all registered circuit breakers
func (cbm *CircuitBreakerManager) ListBreakerNames() []string {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	names := make([]string, 0, len(cbm.breakers))
	for name := range cbm.breakers {
		names = append(names, name)
	}

	return names
}

// GetAllMetrics returns metrics for all circuit breakers
func (cbm *CircuitBreakerManager) GetAllMetrics() map[string]*CircuitBreakerMetrics {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	metrics := make(map[string]*CircuitBreakerMetrics)
	for name, breaker := range cbm.breakers {
		metrics[name] = breaker.GetMetrics()
	}

	return metrics
}

// NewEnhancedCircuitBreaker creates a new enhanced circuit breaker
func NewEnhancedCircuitBreaker(name string, config CircuitBreakerConfig) *EnhancedCircuitBreaker {
	metrics := &CircuitBreakerMetrics{
		LastStateChange: time.Now(),
	}

	breaker := &EnhancedCircuitBreaker{
		name:      name,
		config:    config,
		metrics:   metrics,
		listeners: make([]CircuitBreakerListener, 0),
	}

	// Configure gobreaker settings
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Default ready to trip logic: > 50% failure rate with at least 5 requests
			return counts.Requests >= 5 && counts.TotalFailures > counts.Requests/2
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			breaker.handleStateChange(from, to)
		},
	}

	// Allow custom OnStateChange if provided
	if config.OnStateChange != nil {
		originalHandler := settings.OnStateChange
		settings.OnStateChange = func(name string, from, to gobreaker.State) {
			originalHandler(name, from, to)
			config.OnStateChange(name, from.String(), to.String())
		}
	}

	breaker.CircuitBreaker = gobreaker.NewCircuitBreaker(settings)

	// Add default listener
	breaker.AddListener(&DefaultCircuitBreakerListener{name: name})

	return breaker
}

// Execute runs an operation with circuit breaker protection
func (ecb *EnhancedCircuitBreaker) Execute(ctx context.Context, operation CircuitBreakerOperation) (interface{}, error) {
	startTime := time.Now()

	// Update request metrics
	ecb.metrics.mu.Lock()
	ecb.metrics.TotalRequests++
	ecb.metrics.mu.Unlock()

	// Execute with circuit breaker
	result, err := ecb.CircuitBreaker.Execute(func() (interface{}, error) {
		return operation(ctx)
	})

	duration := time.Since(startTime)
	success := err == nil

	// Update metrics
	ecb.updateMetrics(duration, success, err)

	// Notify listeners
	ecb.notifyRequestComplete(duration, success)

	if err != nil {
		// Handle specific circuit breaker errors
		if ecb.State() == gobreaker.StateOpen {
			return nil, NewCircuitBreakerError(fmt.Sprintf("circuit breaker '%s' is open", ecb.name))
		}

		return nil, err
	}

	return result, nil
}

// ExecuteWithTimeout runs an operation with both circuit breaker and timeout protection
func (ecb *EnhancedCircuitBreaker) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, operation CircuitBreakerOperation) (interface{}, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return ecb.Execute(timeoutCtx, operation)
}

// AddListener adds a circuit breaker event listener
func (ecb *EnhancedCircuitBreaker) AddListener(listener CircuitBreakerListener) {
	ecb.mu.Lock()
	defer ecb.mu.Unlock()

	ecb.listeners = append(ecb.listeners, listener)
}

// RemoveListener removes a circuit breaker event listener
func (ecb *EnhancedCircuitBreaker) RemoveListener(listener CircuitBreakerListener) {
	ecb.mu.Lock()
	defer ecb.mu.Unlock()

	for i, l := range ecb.listeners {
		if l == listener {
			ecb.listeners = append(ecb.listeners[:i], ecb.listeners[i+1:]...)
			break
		}
	}
}

// GetMetrics returns current circuit breaker metrics
func (ecb *EnhancedCircuitBreaker) GetMetrics() *CircuitBreakerMetrics {
	ecb.metrics.mu.RLock()
	defer ecb.metrics.mu.RUnlock()

	// Return a copy to avoid race conditions - manually copy fields to avoid copying mutex
	metricsCopy := &CircuitBreakerMetrics{
		TotalRequests:      ecb.metrics.TotalRequests,
		SuccessfulRequests: ecb.metrics.SuccessfulRequests,
		FailedRequests:     ecb.metrics.FailedRequests,
		TimeoutRequests:    ecb.metrics.TimeoutRequests,
		CircuitOpenCount:   ecb.metrics.CircuitOpenCount,
		CircuitCloseCount:  ecb.metrics.CircuitCloseCount,
		LastStateChange:    ecb.metrics.LastStateChange,
		AverageLatency:     ecb.metrics.AverageLatency,
	}

	return metricsCopy
}

// GetSuccessRate returns the current success rate
func (ecb *EnhancedCircuitBreaker) GetSuccessRate() float64 {
	ecb.metrics.mu.RLock()
	defer ecb.metrics.mu.RUnlock()

	if ecb.metrics.TotalRequests == 0 {
		return 0.0
	}

	return float64(ecb.metrics.SuccessfulRequests) / float64(ecb.metrics.TotalRequests)
}

// GetFailureRate returns the current failure rate
func (ecb *EnhancedCircuitBreaker) GetFailureRate() float64 {
	return 1.0 - ecb.GetSuccessRate()
}

// IsHealthy returns whether the circuit breaker is in a healthy state
func (ecb *EnhancedCircuitBreaker) IsHealthy() bool {
	state := ecb.State()
	return state == gobreaker.StateClosed || state == gobreaker.StateHalfOpen
}

// GetState returns the current circuit breaker state as a string
func (ecb *EnhancedCircuitBreaker) GetState() string {
	return ecb.State().String()
}

// GetUptime returns how long the circuit breaker has been in its current state
func (ecb *EnhancedCircuitBreaker) GetUptime() time.Duration {
	ecb.metrics.mu.RLock()
	defer ecb.metrics.mu.RUnlock()

	return time.Since(ecb.metrics.LastStateChange)
}

// Reset manually resets the circuit breaker to closed state
func (ecb *EnhancedCircuitBreaker) Reset() {
	ecb.CircuitBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        ecb.name,
		MaxRequests: ecb.config.MaxRequests,
		Interval:    ecb.config.Interval,
		Timeout:     ecb.config.Timeout,
		OnStateChange: func(name string, from, to gobreaker.State) {
			ecb.handleStateChange(from, to)
		},
	})

	// Reset metrics
	ecb.metrics.mu.Lock()
	ecb.metrics.LastStateChange = time.Now()
	ecb.metrics.mu.Unlock()

	// Notify listeners
	ecb.notifyCircuitClose()
}

// Private methods

// updateMetrics updates internal metrics based on request results
func (ecb *EnhancedCircuitBreaker) updateMetrics(duration time.Duration, success bool, err error) {
	ecb.metrics.mu.Lock()
	defer ecb.metrics.mu.Unlock()

	if success {
		ecb.metrics.SuccessfulRequests++
	} else {
		ecb.metrics.FailedRequests++

		// Check for timeout errors
		if err == context.DeadlineExceeded {
			ecb.metrics.TimeoutRequests++
		}
	}

	// Update average latency with exponential moving average
	if ecb.metrics.TotalRequests == 1 {
		ecb.metrics.AverageLatency = duration
	} else {
		// 90% weight to previous average, 10% to current request
		ecb.metrics.AverageLatency = time.Duration(
			0.9*float64(ecb.metrics.AverageLatency) + 0.1*float64(duration),
		)
	}
}

// handleStateChange handles circuit breaker state changes
func (ecb *EnhancedCircuitBreaker) handleStateChange(from, to gobreaker.State) {
	ecb.metrics.mu.Lock()
	ecb.metrics.LastStateChange = time.Now()
	ecb.metrics.mu.Unlock()

	// Update state change counters
	if to == gobreaker.StateOpen {
		ecb.metrics.mu.Lock()
		ecb.metrics.CircuitOpenCount++
		ecb.metrics.mu.Unlock()
		ecb.notifyCircuitOpen()
	} else if from == gobreaker.StateOpen && to == gobreaker.StateClosed {
		ecb.metrics.mu.Lock()
		ecb.metrics.CircuitCloseCount++
		ecb.metrics.mu.Unlock()
		ecb.notifyCircuitClose()
	}

	// Notify all listeners
	ecb.notifyStateChange(from, to)
}

// notifyStateChange notifies all listeners of state changes
func (ecb *EnhancedCircuitBreaker) notifyStateChange(from, to gobreaker.State) {
	ecb.mu.RLock()
	listeners := make([]CircuitBreakerListener, len(ecb.listeners))
	copy(listeners, ecb.listeners)
	ecb.mu.RUnlock()

	for _, listener := range listeners {
		go listener.OnStateChange(ecb.name, from, to)
	}
}

// notifyRequestComplete notifies all listeners of request completion
func (ecb *EnhancedCircuitBreaker) notifyRequestComplete(duration time.Duration, success bool) {
	ecb.mu.RLock()
	listeners := make([]CircuitBreakerListener, len(ecb.listeners))
	copy(listeners, ecb.listeners)
	ecb.mu.RUnlock()

	for _, listener := range listeners {
		go listener.OnRequestComplete(ecb.name, duration, success)
	}
}

// notifyCircuitOpen notifies all listeners that the circuit opened
func (ecb *EnhancedCircuitBreaker) notifyCircuitOpen() {
	ecb.mu.RLock()
	listeners := make([]CircuitBreakerListener, len(ecb.listeners))
	copy(listeners, ecb.listeners)
	ecb.mu.RUnlock()

	counts := ecb.Counts()
	for _, listener := range listeners {
		go listener.OnCircuitOpen(ecb.name, counts)
	}
}

// notifyCircuitClose notifies all listeners that the circuit closed
func (ecb *EnhancedCircuitBreaker) notifyCircuitClose() {
	ecb.mu.RLock()
	listeners := make([]CircuitBreakerListener, len(ecb.listeners))
	copy(listeners, ecb.listeners)
	ecb.mu.RUnlock()

	for _, listener := range listeners {
		go listener.OnCircuitClose(ecb.name)
	}
}

// Default listener implementation

// OnStateChange handles circuit breaker state changes
func (dcbl *DefaultCircuitBreakerListener) OnStateChange(name string, from, to gobreaker.State) {
	// In a real implementation, this would log to the application logger
	fmt.Printf("[CircuitBreaker:%s] State changed from %s to %s\n", name, from.String(), to.String())
}

// OnRequestComplete handles request completion events
func (dcbl *DefaultCircuitBreakerListener) OnRequestComplete(name string, duration time.Duration, success bool) {
	status := "SUCCESS"
	if !success {
		status = "FAILURE"
	}
	fmt.Printf("[CircuitBreaker:%s] Request completed in %v with status: %s\n", name, duration, status)
}

// OnCircuitOpen handles circuit breaker opening events
func (dcbl *DefaultCircuitBreakerListener) OnCircuitOpen(name string, counts gobreaker.Counts) {
	fmt.Printf("[CircuitBreaker:%s] Circuit opened - Total: %d, Success: %d, Failures: %d\n",
		name, counts.Requests, counts.TotalSuccesses, counts.TotalFailures)
}

// OnCircuitClose handles circuit breaker closing events
func (dcbl *DefaultCircuitBreakerListener) OnCircuitClose(name string) {
	fmt.Printf("[CircuitBreaker:%s] Circuit closed - Service recovered\n", name)
}

// Utility functions for circuit breaker management

// CircuitBreakerHealthCheck provides health check functionality for circuit breakers
type CircuitBreakerHealthCheck struct {
	manager *CircuitBreakerManager
}

// NewCircuitBreakerHealthCheck creates a new health check instance
func NewCircuitBreakerHealthCheck(manager *CircuitBreakerManager) *CircuitBreakerHealthCheck {
	return &CircuitBreakerHealthCheck{
		manager: manager,
	}
}

// CheckHealth returns the health status of all circuit breakers
func (hc *CircuitBreakerHealthCheck) CheckHealth() map[string]bool {
	health := make(map[string]bool)

	for name, breaker := range hc.manager.breakers {
		health[name] = breaker.IsHealthy()
	}

	return health
}

// GetUnhealthyBreakers returns a list of circuit breakers that are not healthy
func (hc *CircuitBreakerHealthCheck) GetUnhealthyBreakers() []string {
	health := hc.CheckHealth()
	unhealthy := make([]string, 0)

	for name, isHealthy := range health {
		if !isHealthy {
			unhealthy = append(unhealthy, name)
		}
	}

	return unhealthy
}

// GetOverallHealth returns whether all circuit breakers are healthy
func (hc *CircuitBreakerHealthCheck) GetOverallHealth() bool {
	unhealthy := hc.GetUnhealthyBreakers()
	return len(unhealthy) == 0
}

// CircuitBreakerDashboard provides monitoring and dashboard functionality
type CircuitBreakerDashboard struct {
	manager *CircuitBreakerManager
}

// NewCircuitBreakerDashboard creates a new dashboard instance
func NewCircuitBreakerDashboard(manager *CircuitBreakerManager) *CircuitBreakerDashboard {
	return &CircuitBreakerDashboard{
		manager: manager,
	}
}

// GetDashboardData returns comprehensive dashboard data
func (cbd *CircuitBreakerDashboard) GetDashboardData() *CircuitBreakerDashboardData {
	metrics := cbd.manager.GetAllMetrics()

	data := &CircuitBreakerDashboardData{
		Timestamp:       time.Now(),
		TotalBreakers:   len(metrics),
		HealthyBreakers: 0,
		OpenBreakers:    0,
		Breakers:        make(map[string]*CircuitBreakerStatus),
	}

	for name, breaker := range cbd.manager.breakers {
		status := &CircuitBreakerStatus{
			Name:        name,
			State:       breaker.GetState(),
			IsHealthy:   breaker.IsHealthy(),
			SuccessRate: breaker.GetSuccessRate(),
			FailureRate: breaker.GetFailureRate(),
			Uptime:      breaker.GetUptime(),
			Metrics:     breaker.GetMetrics(),
		}

		data.Breakers[name] = status

		if status.IsHealthy {
			data.HealthyBreakers++
		}

		if status.State == "OPEN" {
			data.OpenBreakers++
		}
	}

	return data
}

// CircuitBreakerDashboardData contains comprehensive dashboard information
type CircuitBreakerDashboardData struct {
	Timestamp       time.Time                        `json:"timestamp"`
	TotalBreakers   int                              `json:"totalBreakers"`
	HealthyBreakers int                              `json:"healthyBreakers"`
	OpenBreakers    int                              `json:"openBreakers"`
	Breakers        map[string]*CircuitBreakerStatus `json:"breakers"`
}

// CircuitBreakerStatus contains status information for a single circuit breaker
type CircuitBreakerStatus struct {
	Name        string                 `json:"name"`
	State       string                 `json:"state"`
	IsHealthy   bool                   `json:"isHealthy"`
	SuccessRate float64                `json:"successRate"`
	FailureRate float64                `json:"failureRate"`
	Uptime      time.Duration          `json:"uptime"`
	Metrics     *CircuitBreakerMetrics `json:"metrics"`
}
