package anthropic

import (
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/llm"
)

type sender struct {
	client *llm.BaseClient
	model  string
}

// NewSender creates a new Anthropic sender
func NewSender(client *llm.BaseClient, model string) llm.Sender {
	return &sender{
		client: client,
		model:  model,
	}
}

func (s *sender) BuildRequest(content string) (interface{}, error) {
	return map[string]interface{}{
		"model":      s.model,
		"max_tokens": s.client.MaxTokens,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": content,
			},
		},
	}
}

func (s *sender) ParseResponse(response interface{}) (*llm.Result, error) {
	resp, ok := response.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	content, ok := resp["content"].(string)
	if !ok {
		return nil, fmt.Errorf("no content in response")
	}

	usage := &llm.Usage{
		PromptTokens:     0,
		CompletionTokens: len(content),
		TotalTokens:      len(content),
	}

	return &llm.Result{
		Response: content,
		Model:    s.model,
		Provider: "Anthropic",
		Usage:    usage,
	}
}

func (s *sender) GetEndpoint() string {
	return "/v1/messages"
}

func (s *sender) GetHeaders() map[string]string {
	return map[string]string{
		"Content-Type":      "application/json",
		"x-api-key":         s.client.APIKey,
		"anthropic-version": "2023-06-01",
	}
}

func (s *sender) GetResponseType() interface{} {
	return &anthropicResponse{}
}

type anthropicResponse struct {
	Content string `json:"content"`
}
