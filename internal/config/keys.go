// Package config provides centralized configuration key constants.
package config

const (
	// Scanner
	KeyScannerMaxFiles             = "scanner.max-files"
	KeyScannerMaxFileSize          = "scanner.max-file-size"
	KeyScannerMaxMemory            = "scanner.max-memory"
	KeyScannerSkipBinary           = "scanner.skip-binary"
	KeyScannerIncludeHidden        = "scanner.include-hidden"
	KeyScannerIncludeIgnored       = "scanner.include-ignored"
	KeyScannerWorkers              = "scanner.workers"
	KeyScannerRespectGitignore     = "scanner.respect-gitignore"
	KeyScannerRespectShotgunignore = "scanner.respect-shotgunignore"

	// LLM
	KeyLLMProvider = "llm.provider"
	KeyLLMAPIKey   = "llm.api-key"
	KeyLLMBaseURL  = "llm.base-url"
	KeyLLMModel    = "llm.model"
	KeyLLMTimeout  = "llm.timeout"

	// Gemini
	KeyGeminiEnabled        = "gemini.enabled"
	KeyGeminiModel          = "gemini.model"
	KeyGeminiTimeout        = "gemini.timeout"
	KeyGeminiBinaryPath     = "gemini.binary-path"
	KeyGeminiBrowserRefresh = "gemini.browser-refresh"
	KeyGeminiAutoSend       = "gemini.auto-send"
	KeyGeminiSaveResponse   = "gemini.save-response"

	// Context
	KeyContextIncludeTree    = "context.include-tree"
	KeyContextIncludeSummary = "context.include-summary"
	KeyContextMaxSize        = "context.max-size"

	// Template
	KeyTemplateCustomPath = "template.custom-path"

	// Output
	KeyOutputFormat    = "output.format"
	KeyOutputClipboard = "output.clipboard"

	// Global
	KeyVerbose = "verbose"
	KeyQuiet   = "quiet"
)
