package geminiapi

const (
	defaultBaseURL   = "https://generativelanguage.googleapis.com/v1beta"
	defaultModel     = "gemini-2.5-flash"
	defaultMaxTokens = 8192
)

// ValidModels returns known Gemini models
func ValidModels() []string {
	return []string{
		"gemini-2.5-flash",
		"gemini-2.0-flash-exp",
	}
}
