package core

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

// TranslationResult represents the result of a translation operation
type TranslationResult struct {
	OriginalText   string
	TranslatedText string
	SourceLanguage string
	TargetLanguage string
	Timestamp      time.Time
	ApiProvider    string
}

// Translator handles text translation using OpenAI-compatible APIs
type Translator struct {
	client    *openai.Client
	openaiConfig    OpenAIConfig
	translationConfig TranslationConfig
	keyManager *SecureKeyManager
	mu        sync.RWMutex
}

// NewTranslator creates a new translator instance
func NewTranslator(openaiConfig OpenAIConfig, translationConfig TranslationConfig, keyManager *SecureKeyManager) (*Translator, error) {
	if keyManager == nil {
		return nil, fmt.Errorf("key manager cannot be nil")
	}
	
	translator := &Translator{
		openaiConfig:      openaiConfig,
		translationConfig: translationConfig,
		keyManager:        keyManager,
	}
	
	// Initialize client if API key is available
	if err := translator.initializeClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize translator: %w", err)
	}
	
	return translator, nil
}

// initializeClient initializes the OpenAI client with the stored API key
func (t *Translator) initializeClient() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.openaiConfig.APIKeyAlias == "" {
		return fmt.Errorf("no API key alias configured")
	}
	
	apiKey, err := t.keyManager.GetAPIKey(t.openaiConfig.APIKeyAlias)
	if err != nil {
		return fmt.Errorf("failed to retrieve API key: %w", err)
	}
	
	// Create OpenAI client configuration
	clientConfig := openai.DefaultConfig(apiKey)
	
	// Set custom base URL if provided
	if t.openaiConfig.BaseURL != "" && t.openaiConfig.BaseURL != "https://api.openai.com/v1" {
		clientConfig.BaseURL = t.openaiConfig.BaseURL
	}
	
	t.client = openai.NewClientWithConfig(clientConfig)
	
	return nil
}

// UpdateConfig updates the translator configuration and reinitializes the client
func (t *Translator) UpdateConfig(openaiConfig OpenAIConfig, translationConfig TranslationConfig) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.openaiConfig = openaiConfig
	t.translationConfig = translationConfig
	
	// Reinitialize client with new configuration
	if t.openaiConfig.APIKeyAlias != "" {
		return t.initializeClient()
	}
	
	return nil
}

// TranslateText translates text from any language to the target language
func (t *Translator) TranslateText(ctx context.Context, text, textType string) (*TranslationResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}
	
	t.mu.RLock()
	client := t.client
	openaiConfig := t.openaiConfig
	translationConfig := t.translationConfig
	t.mu.RUnlock()
	
	if client == nil {
		return nil, fmt.Errorf("translator not properly initialized")
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(openaiConfig.Timeout)*time.Second)
	defer cancel()
	
	// Build translation prompt based on text type
	prompt := t.buildTranslationPrompt(text, textType, translationConfig.TargetLanguage)
	
	// Perform translation with retry logic
	translatedText, err := t.translateWithRetry(ctx, prompt, openaiConfig)
	if err != nil {
		return nil, fmt.Errorf("translation failed: %w", err)
	}
	
	return &TranslationResult{
		OriginalText:   text,
		TranslatedText: translatedText,
		SourceLanguage: "auto-detected",
		TargetLanguage: translationConfig.TargetLanguage,
		Timestamp:      time.Now(),
		ApiProvider:    openaiConfig.BaseURL,
	}, nil
}

// TranslateTask translates a task description with task-specific context
func (t *Translator) TranslateTask(ctx context.Context, taskText string) (*TranslationResult, error) {
	return t.TranslateText(ctx, taskText, "task")
}

// TranslateRules translates custom rules with rules-specific context
func (t *Translator) TranslateRules(ctx context.Context, rulesText string) (*TranslationResult, error) {
	return t.TranslateText(ctx, rulesText, "rules")
}

// buildTranslationPrompt creates an optimized prompt for translation
func (t *Translator) buildTranslationPrompt(text, textType, targetLanguage string) string {
	var contextPrompt string
	
	switch textType {
	case "task":
		contextPrompt = "You are translating a software development task description. Preserve technical terms, programming concepts, and maintain clarity for developers."
	case "rules":
		contextPrompt = "You are translating custom rules or constraints for a software development task. Preserve technical requirements and maintain precise meaning."
	default:
		contextPrompt = "You are translating text for a software development context. Preserve technical terms and maintain the original meaning."
	}
	
	// Use custom context prompt if provided
	if t.translationConfig.ContextPrompt != "" {
		contextPrompt = t.translationConfig.ContextPrompt
	}
	
	return fmt.Sprintf(`%s

Please translate the following text to %s. Requirements:
1. Maintain technical accuracy
2. Preserve the original meaning and intent
3. Keep technical terms in English when appropriate
4. Respond with ONLY the translated text, no explanations

Text to translate:
%s`, contextPrompt, targetLanguage, text)
}

// translateWithRetry performs translation with exponential backoff retry logic
func (t *Translator) translateWithRetry(ctx context.Context, prompt string, config OpenAIConfig) (string, error) {
	var lastErr error
	
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff delay
			delay := time.Duration(config.RetryDelay) * time.Second * time.Duration(math.Pow(2, float64(attempt-1)))
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("context cancelled during retry delay: %w", ctx.Err())
			case <-time.After(delay):
				// Continue with retry
			}
		}
		
		translatedText, err := t.performTranslation(ctx, prompt, config)
		if err == nil {
			return translatedText, nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !t.isRetryableError(err) {
			break
		}
	}
	
	return "", fmt.Errorf("translation failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

// performTranslation makes a single translation API call
func (t *Translator) performTranslation(ctx context.Context, prompt string, config OpenAIConfig) (string, error) {
	request := openai.ChatCompletionRequest{
		Model: config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   config.MaxTokens,
		Temperature: float32(config.Temperature),
	}
	
	response, err := t.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from API")
	}
	
	translatedText := strings.TrimSpace(response.Choices[0].Message.Content)
	if translatedText == "" {
		return "", fmt.Errorf("empty translation received from API")
	}
	
	return translatedText, nil
}

// isRetryableError determines if an error should trigger a retry
func (t *Translator) isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	
	// Retryable conditions
	retryableConditions := []string{
		"timeout",
		"connection",
		"rate limit",
		"server error",
		"503",
		"502",
		"504",
		"429", // Too Many Requests
	}
	
	for _, condition := range retryableConditions {
		if strings.Contains(errStr, condition) {
			return true
		}
	}
	
	return false
}

// IsConfigured checks if the translator is properly configured and ready to use
func (t *Translator) IsConfigured() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	return t.client != nil && t.openaiConfig.APIKeyAlias != ""
}

// TestConnection verifies that the translator can connect to the API
func (t *Translator) TestConnection(ctx context.Context) error {
	if !t.IsConfigured() {
		return fmt.Errorf("translator not configured")
	}
	
	// Test with a simple translation
	testText := "Hello, world!"
	result, err := t.TranslateText(ctx, testText, "test")
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	
	if result.TranslatedText == "" {
		return fmt.Errorf("connection test returned empty result")
	}
	
	return nil
}

// GetSupportedModels returns a list of supported OpenAI models for translation
func (t *Translator) GetSupportedModels() []string {
	return []string{
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
	}
}

// ValidateConfig validates the translator configuration
func (t *Translator) ValidateConfig(config OpenAIConfig) error {
	if config.APIKeyAlias == "" {
		return fmt.Errorf("API key alias is required")
	}
	
	if config.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	
	if config.Model == "" {
		return fmt.Errorf("model is required")
	}
	
	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	if config.MaxTokens <= 0 {
		return fmt.Errorf("max tokens must be positive")
	}
	
	if config.Temperature < 0 || config.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	
	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	
	if config.RetryDelay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}
	
	return nil
}

// EstimateTokens provides a rough estimate of token usage for a text
func (t *Translator) EstimateTokens(text string) int {
	// Rough approximation: 1 token ≈ 4 characters for English
	// This is a simplified estimation
	return len(text) / 4
}