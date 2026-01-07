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

// IsKnownModel always returns true - model validation removed to allow custom/preview models.
// Deprecated: This function no longer validates models. Any model string is accepted.
func IsKnownModel(model, baseURL string) bool {
	return true
}
