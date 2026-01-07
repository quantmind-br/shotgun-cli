package geminiapi

// ValidModels returns known models for Gemini API.
func ValidModels() []string {
	return []string{
		"gemini-2.5-flash",
		"gemini-2.5-pro",
		"gemini-2.0-flash",
		"gemini-1.5-flash",
		"gemini-1.5-pro",
	}
}

// IsKnownModel always returns true - model validation removed to allow custom/preview models.
// Deprecated: This function no longer validates models. Any model string is accepted.
func IsKnownModel(model, baseURL string) bool {
	return true
}
