package geminiapi

import (
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/http"
	"github.com/quantmind-br/shotgun-cli/internal/platform/llm"
)

const (
	defaultBaseURL   = "https://generativelanguage.googleapis.com/v1beta"
	defaultModel     = "gemini-2.5-flash"
	defaultMaxTokens = 8192
)

type sender struct {
	client *llm.BaseClient
}

// NewSender creates a new Gemini API sender
func NewSender(client *llm.BaseClient) llm.Sender {
	return &sender{
		client: client,
	}
}

func (s *sender) BuildRequest(content string) (interface{}, error) {
	return map[string]interface{}{
		"contents": []Content{
			{
				"Parts": []string{content},
			},
		},
	}
}

func (s *sender) ParseResponse(response interface{}) (*llm.Result, error) {
	resp, ok := response.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	contents, ok := resp["candidates"].([]interface{})
	if !ok || len(contents) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	content, ok := contents[0].(map[string]interface{})
	parts, ok := content["Parts"].([]interface{})
	if !ok || len(parts) == 0 {
		return nil, fmt.Errorf("no parts in response")
	}

	var responseText string
	for _, part := range parts {
		responseText += part.Text
	}

	usage := &llm.Usage{
		CompletionTokens: len(responseText),
		TotalTokens:      len(responseText),
	}

	return &llm.Result{
		Response: responseText,
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
	Candidates []Candidate `json:"candidates"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}
