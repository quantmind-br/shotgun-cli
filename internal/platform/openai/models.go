package openai

// ValidModels returns known models for OpenAI.
// Note: We don't restrict to these since custom endpoints may have others.
func ValidModels() []string {
	return []string{
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
		"o1-preview",
		"o1-mini",
		"o1",
		"o3-mini",
	}
}

// IsKnownModel checks if it's a known model.
// Returns true for any model if base-url is custom.
func IsKnownModel(model, baseURL string) bool {
	// If using custom endpoint, accept any model.
	if baseURL != "" && baseURL != "https://api.openai.com/v1" {
		return true
	}

	for _, known := range ValidModels() {
		if model == known {
			return true
		}
	}
	return false
}
