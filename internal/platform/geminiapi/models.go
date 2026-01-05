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

// IsKnownModel checks if it's a known model.
func IsKnownModel(model, baseURL string) bool {
	// If using custom endpoint, accept any model.
	if baseURL != "" && baseURL != "https://generativelanguage.googleapis.com/v1beta" {
		return true
	}

	for _, known := range ValidModels() {
		if model == known {
			return true
		}
	}
	return false
}
