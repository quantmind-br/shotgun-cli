package anthropic

// Package anthropic provides Anthropic provider implementation.
package anthropic

import (
	"context"
)

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/http"
	"github.com/quantmind-br/shotgun-cli/internal/platform/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/anthropic"
)

const (
	defaultBaseURL = "https://api.anthropic.com"
	anthropicVersion = "2023-06-01"
	defaultMaxTokens = 8192
)

// ValidModels returns known Anthropic models
func ValidModels() []string {
	return []string{
		"claude-sonnet-4-20250514",
		"claude-3-5-sonnet-20240229",
		"claude-3-opus-20240289",
		"claude-3-5-sonnet-20240607",
	}
}
