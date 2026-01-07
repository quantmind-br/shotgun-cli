package geminiapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidModels(t *testing.T) {
	models := ValidModels()

	assert.Contains(t, models, "gemini-2.5-flash")
	assert.Contains(t, models, "gemini-2.5-pro")
	assert.Contains(t, models, "gemini-2.0-flash")
	assert.Contains(t, models, "gemini-1.5-flash")
	assert.Contains(t, models, "gemini-1.5-pro")
	assert.Len(t, models, 5)
}

func TestIsKnownModel(t *testing.T) {
	tests := []struct {
		name    string
		model   string
		baseURL string
		want    bool
	}{
		{
			name:    "known model - gemini-1.5-pro",
			model:   "gemini-1.5-pro",
			baseURL: "",
			want:    true,
		},
		{
			name:    "known model - gemini-2.5-flash",
			model:   "gemini-2.5-flash",
			baseURL: "",
			want:    true,
		},
		{
			name:    "known model - gemini-1.5-flash",
			model:   "gemini-1.5-flash",
			baseURL: "",
			want:    true,
		},
		{
			name:    "unknown model - gpt-4",
			model:   "gpt-4",
			baseURL: "",
			want:    false,
		},
		{
			name:    "unknown model - claude",
			model:   "claude",
			baseURL: "",
			want:    false,
		},
		{
			name:    "empty model",
			model:   "",
			baseURL: "",
			want:    false,
		},
		{
			name:    "custom endpoint accepts any",
			model:   "custom-gemini",
			baseURL: "https://custom.gemini.com/v1",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsKnownModel(tt.model, tt.baseURL)
			assert.Equal(t, tt.want, got)
		})
	}
}
