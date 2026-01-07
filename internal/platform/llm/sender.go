package llm

// Sender interface defines the strategy pattern for provider-specific logic
type Sender interface {
	// BuildRequest creates the provider-specific request payload
	BuildRequest(content string) (interface{}, error)

	// ParseResponse extracts the llm.Result from the provider-specific response
	ParseResponse(response interface{}) (*Result, error)

	// GetEndpoint returns the API endpoint path for the provider
	GetEndpoint() string

	// GetHeaders returns provider-specific HTTP headers
	GetHeaders() map[string]string

	// GetResponseType returns the response type for JSON unmarshaling
	GetResponseType() interface{}
}
