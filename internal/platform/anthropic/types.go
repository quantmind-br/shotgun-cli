// Package anthropic provides a client for Anthropic Claude API.
package anthropic

// MessagesRequest represents the request body for /v1/messages.
type MessagesRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	System    string    `json:"system,omitempty"`
	Stream    bool      `json:"stream,omitempty"`
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

// MessagesResponse represents the API response.
type MessagesResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        UsageAPI       `json:"usage"`
}

// ContentBlock represents a content block in the response.
type ContentBlock struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
}

// UsageAPI represents token usage.
type UsageAPI struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ErrorResponse represents an API error.
type ErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}
