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

// IsKnownModel always returns true - model validation removed to allow custom/preview models.
// Deprecated: This function no longer validates models. Any model string is accepted.
func IsKnownModel(model, baseURL string) bool {
	return true
}
