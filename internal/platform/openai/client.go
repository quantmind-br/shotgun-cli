package openai

import (
	"context"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/http"
	"github.com/quantmind-br/shotgun-cli/internal/platform/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/openai"
)

type sender struct {
	client *llm.BaseClient
	model  string
}

// NewSender creates a new OpenAI sender
func NewSender(client *llm.BaseClient) llm.Sender {
	return &sender{
		client: client,
		model: "gpt-4o",
	}
}

func (s *sender) BuildRequest(content string) (interface{}, error) {
	return map[string]interface{}{
		"model":    s.model,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": content,
			},
		},
	}
	}
}

func (s *sender) ParseResponse(response interface{}) (*llm.Result, error) {
	resp, ok := response.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	choices, ok := resp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := choices[0].(map[string]interface{})
	message, ok := choice["message"].(string)
	if !ok {
		return nil, fmt.Errorf("no message in choice")
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("no content in message")
	}

	usage := &llm.Usage{
		CompletionTokens: len(content),
	}

	return &llm.Result{
		Response:    content,
		Model:    s.model,
			Provider:    "OpenAI",
		Usage:    usage,
	}
}

func (s *sender) GetEndpoint() string {
	return "/v1/chat/completions"
}

func (s *sender) GetHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
		"Authorization": "Bearer " + s.client.APIKey,
	}
}

func (s *sender) GetResponseType() interface{} {
	return &openaiResponse{}
}

type openaiResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
}
