// Package llmbase provides shared infrastructure for HTTP-based LLM providers.
// This package exists separately from internal/core/llm to avoid import cycles.
package llmbase

import (
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

// Sender defines the provider-specific operations that each LLM provider must implement.
// This interface follows the Strategy pattern - BaseClient handles common logic while
// concrete providers implement these methods for their specific API requirements.
type Sender interface {
	// BuildRequest creates the provider-specific request payload.
	// The returned value should be JSON-serializable.
	BuildRequest(content string) (interface{}, error)

	// ParseResponse extracts llm.Result from the provider's response.
	// The response parameter contains the unmarshaled API response.
	ParseResponse(response interface{}, rawJSON []byte) (*llm.Result, error)

	// GetEndpoint returns the API endpoint path (e.g., "/v1/chat/completions").
	// For providers that include API key in URL (like Gemini), include it here.
	GetEndpoint() string

	// GetHeaders returns provider-specific HTTP headers.
	// Common headers like Content-Type are handled by BaseClient.
	GetHeaders() map[string]string

	// NewResponse returns a new instance of the response struct for JSON unmarshaling.
	NewResponse() interface{}

	// GetProviderName returns the display name for this provider (e.g., "OpenAI", "Anthropic").
	GetProviderName() string
}
