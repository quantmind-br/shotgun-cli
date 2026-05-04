package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidModels(t *testing.T) {
	models := ValidModels()

	assert.Contains(t, models, "gpt-4o-mini")
	assert.Contains(t, models, "gpt-4o-mini")
	assert.Contains(t, models, "gpt-4-turbo")
	assert.Contains(t, models, "gpt-4")
	assert.Contains(t, models, "gpt-3.5-turbo")
	assert.Contains(t, models, "o1-preview")
	assert.Contains(t, models, "o1-mini")
	assert.Contains(t, models, "o3-mini")
	assert.Contains(t, models, "o1")
	assert.Len(t, models, 9)
}

func TestIsKnownModel(t *testing.T) {
	// Model validation removed - all models should now be valid
	tests := []struct {
		name    string
		model   string
		baseURL string
		want    bool
	}{
		{
			name:    "known model - o3-mini",
			model:   "o3-mini",
			baseURL: "",
			want:    true,
		},
		{
			name:    "known model - gpt-4",
			model:   "gpt-4",
			baseURL: "",
			want:    true,
		},
		{
			name:    "known model - gpt-3.5-turbo",
			model:   "gpt-3.5-turbo",
			baseURL: "",
			want:    true,
		},
		{
			name:    "any model - gpt-5",
			model:   "gpt-5",
			baseURL: "",
			want:    true,
		},
		{
			name:    "any model - claude",
			model:   "claude",
			baseURL: "",
			want:    true,
		},
		{
			name:    "empty model",
			model:   "",
			baseURL: "",
			want:    true,
		},
		{
			name:    "custom endpoint",
			model:   "custom-gpt-model",
			baseURL: "https://custom.openai.com/v1",
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
