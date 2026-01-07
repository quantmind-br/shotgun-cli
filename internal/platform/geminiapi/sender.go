package geminiapi

import (
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/llm"
)

type sender struct {
	client *llm.BaseClient
}

// NewSender creates a new GeminiAPI sender
func NewSender(client *llm.BaseClient) llm.Sender {
	return &sender{
		client: client,
	}
}

func (s *sender) BuildRequest(content string) (interface{}, error) {
	return map[string]interface{}{
		"contents": []string{content},
	}
}

func (s *sender) ParseResponse(response interface{}) (*llm.Result, error) {
	resp, ok := response.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	contents, ok := resp["contents"].([]interface{})
	if !ok || len(contents) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	content := contents[0].(map[string]interface{})
	text, ok := content["parts"].([]interface{})
	if !ok || len(text) == 0 {
		return nil, fmt.Errorf("no text in response")
	}

	usage := &llm.Usage{
		PromptTokens:     0,
		CompletionTokens: len(text),
		TotalTokens:      len(text),
	}

	return &llm.Result{
		Response: text,
		Model:    s.client.Model,
		Provider: "Gemini",
		Usage:    usage,
	}
}

func (s *sender) GetEndpoint() string {
	return "/v1/models/" + s.client.Model + ":generateContent"
}

func (s *sender) GetHeaders() map[string]string {
	return map[string]string{
		"Content-Type":   "application/json",
		"x-goog-api-key": s.client.APIKey,
	}
}

func (s *sender) GetResponseType() interface{} {
	return &geminiResponse{}
}

type geminiResponse struct {
	Candidates []struct {
		Content string `json:"parts"`
	} `json:"candidates"`
}
