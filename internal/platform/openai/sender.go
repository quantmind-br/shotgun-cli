package openai

import (
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/llm"
)

type sender struct {
	client *llm.BaseClient
}

// NewSender creates a new OpenAI sender
func NewSender(client *llm.BaseClient) llm.Sender {
	return &sender{
		client: client,
	}
}

func (s *sender) BuildRequest(content string) (interface{}, error) {
	return map[string]interface{}{
		"model": s.client.Model,
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

	choices, ok := resp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := choices[0].(map[string]interface{})
	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no message in choice")
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("no content in message")
	}

	usage := &llm.Usage{}
	if usageRaw, ok := message["usage"].(map[string]interface{}); ok {
		if pt, ok := usageRaw["prompt_tokens"].(float64); ok {
			usage.PromptTokens = int(pt)
		}
		if ct, ok := usageRaw["completion_tokens"].(float64); ok {
			usage.CompletionTokens = int(ct)
		}
		if tt, ok := usageRaw["total_tokens"].(float64); ok {
			usage.TotalTokens = int(tt)
		}
	}

	return &llm.Result{
		Response: content,
		Model:    s.client.Model,
		Provider: "OpenAI",
		Usage:    usage,
	}
}

func (s *sender) GetEndpoint() string {
	return "/v1/chat/completions"
}

func (s *sender) GetHeaders() map[string]string {
	return map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + s.client.APIKey,
	}
}

func (s *sender) GetResponseType() interface{} {
	return &chatCompletionResponse{}
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}
