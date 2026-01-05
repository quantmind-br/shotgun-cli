package anthropic

// ValidModels returns known models for Anthropic.
func ValidModels() []string {
	return []string{
		"claude-sonnet-4-20250514",
		"claude-3-5-sonnet-latest",
		"claude-3-5-sonnet-20241022",
		"claude-3-5-haiku-latest",
		"claude-3-opus-latest",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}
}

// IsKnownModel checks if it's a known model.
func IsKnownModel(model, baseURL string) bool {
	// If using custom endpoint, accept any model.
	if baseURL != "" && baseURL != "https://api.anthropic.com" {
		return true
	}

	for _, known := range ValidModels() {
		if model == known {
			return true
		}
	}
	return false
}
