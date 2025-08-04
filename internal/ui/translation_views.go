package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"shotgun-cli/internal/core"
)

// TranslationView handles translation UI with async support and progress indicators
type TranslationView struct {
	translationService *core.EnhancedTranslationService
	configManager      *core.EnhancedConfigManager

	// UI state
	isTranslating bool
	progress      progress.Model
	spinner       spinner.Model
	lastResult    *core.EnhancedTranslationResult
	lastError     error

	// Content tracking
	originalText    string
	translatedText  string
	translationType string

	// Display settings
	width         int
	height        int
	showDetails   bool
	autoTranslate bool

	// Cache and retry state
	cacheStats *TranslationCacheStats
	retryCount int
	maxRetries int
}

// TranslationCacheStats contains cache performance information
type TranslationCacheStats struct {
	CacheHits     int64
	CacheMisses   int64
	TotalRequests int64
	HitRate       float64
}

// TranslationProgressMsg represents translation progress updates
type TranslationProgressMsg struct {
	Progress    float64
	Status      string
	AttemptNum  int
	MaxAttempts int
}

// TranslationCompleteMsg represents completed translation
type TranslationCompleteMsg struct {
	Result *core.EnhancedTranslationResult
	Error  error
}

// TranslationRetryMsg represents a retry attempt
type TranslationRetryMsg struct {
	AttemptNum int
	Reason     string
}

// NewTranslationView creates a new translation view
func NewTranslationView(translationService *core.EnhancedTranslationService, configManager *core.EnhancedConfigManager) *TranslationView {
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 40

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &TranslationView{
		translationService: translationService,
		configManager:      configManager,
		progress:           p,
		spinner:            s,
		maxRetries:         3,
		cacheStats:         &TranslationCacheStats{},
		width:              80,
		height:             24,
	}
}

// SetSize sets the view dimensions
func (tv *TranslationView) SetSize(width, height int) {
	tv.width = width
	tv.height = height
	tv.progress.Width = width - 20
}

// StartTranslation begins async translation with progress tracking
func (tv *TranslationView) StartTranslation(text, textType string) tea.Cmd {
	tv.originalText = text
	tv.translationType = textType
	tv.isTranslating = true
	tv.retryCount = 0
	tv.lastError = nil

	return tea.Batch(
		tv.spinner.Tick,
		tv.performTranslation(text, textType),
	)
}

// RetryTranslation retries the last failed translation
func (tv *TranslationView) RetryTranslation() tea.Cmd {
	if tv.retryCount >= tv.maxRetries {
		return nil
	}

	tv.retryCount++
	tv.isTranslating = true
	tv.lastError = nil

	return tea.Batch(
		tv.spinner.Tick,
		tv.performTranslation(tv.originalText, tv.translationType),
		func() tea.Msg {
			return TranslationRetryMsg{
				AttemptNum: tv.retryCount,
				Reason:     "User requested retry",
			}
		},
	)
}

// performTranslation executes the translation in background
func (tv *TranslationView) performTranslation(text, textType string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Check if translation service is configured
		if !tv.translationService.IsConfigured() {
			return TranslationCompleteMsg{
				Error: fmt.Errorf("translation service not configured - please check OpenAI API settings"),
			}
		}

		// Update progress
		go func() {
			for i := 0; i < 100; i += 10 {
				time.Sleep(200 * time.Millisecond)
				// Note: In a real implementation, this would be sent through a channel
				// For now, we'll simulate progress
			}
		}()

		// Perform translation
		result, err := tv.translationService.TranslateText(ctx, text, textType)

		return TranslationCompleteMsg{
			Result: result,
			Error:  err,
		}
	}
}

// Update handles translation view updates
func (tv *TranslationView) Update(msg tea.Msg) (*TranslationView, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case TranslationProgressMsg:
		cmd = tv.progress.SetPercent(msg.Progress)
		return tv, cmd

	case TranslationCompleteMsg:
		tv.isTranslating = false
		tv.lastResult = msg.Result
		tv.lastError = msg.Error

		if msg.Result != nil {
			tv.translatedText = msg.Result.TranslatedText
			tv.updateCacheStats()
		}
		return tv, nil

	case TranslationRetryMsg:
		// Handle retry message display
		return tv, nil

	case tea.KeyMsg:
		if tv.isTranslating {
			switch msg.String() {
			case "ctrl+c", "esc":
				// Cancel translation
				tv.isTranslating = false
				tv.lastError = fmt.Errorf("translation cancelled by user")
				return tv, nil
			}
		} else {
			switch msg.String() {
			case "r":
				// Retry translation
				if tv.lastError != nil {
					return tv, tv.RetryTranslation()
				}
			case "d":
				// Toggle details view
				tv.showDetails = !tv.showDetails
				return tv, nil
			case "c":
				// Clear translation
				tv.clearTranslation()
				return tv, nil
			}
		}
	}

	// Update spinner during translation
	if tv.isTranslating {
		tv.spinner, cmd = tv.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return tv, tea.Batch(cmds...)
}

// View renders the translation view
func (tv *TranslationView) View() string {
	var b strings.Builder

	// Header
	b.WriteString(translationHeaderStyle.Render("Translation Service"))
	b.WriteString("\n\n")

	// Status and progress
	if tv.isTranslating {
		b.WriteString(tv.renderTranslationProgress())
	} else if tv.lastError != nil {
		b.WriteString(tv.renderError())
	} else if tv.lastResult != nil {
		b.WriteString(tv.renderTranslationResult())
	} else {
		b.WriteString(tv.renderWelcome())
	}

	// Details section
	if tv.showDetails && tv.lastResult != nil {
		b.WriteString("\n\n")
		b.WriteString(tv.renderTranslationDetails())
	}

	// Cache stats
	if tv.cacheStats.TotalRequests > 0 {
		b.WriteString("\n\n")
		b.WriteString(tv.renderCacheStats())
	}

	// Help section
	b.WriteString("\n\n")
	b.WriteString(tv.renderHelp())

	return b.String()
}

// renderTranslationProgress shows translation in progress
func (tv *TranslationView) renderTranslationProgress() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%s Translating...", tv.spinner.View()))
	b.WriteString("\n\n")

	if tv.retryCount > 0 {
		b.WriteString(translationRetryStyle.Render(fmt.Sprintf("Retry attempt %d/%d", tv.retryCount, tv.maxRetries)))
		b.WriteString("\n")
	}

	b.WriteString(tv.progress.View())
	b.WriteString("\n\n")

	b.WriteString(translationStatusStyle.Render("Status: Contacting OpenAI API..."))

	return b.String()
}

// renderError shows translation errors with user-friendly messages
func (tv *TranslationView) renderError() string {
	var b strings.Builder

	b.WriteString(translationErrorStyle.Render("Translation Failed"))
	b.WriteString("\n\n")

	// User-friendly error message
	errorMsg := tv.getUserFriendlyError(tv.lastError)
	b.WriteString(translationErrorDetailStyle.Render(errorMsg))
	b.WriteString("\n\n")

	// Retry button
	if tv.retryCount < tv.maxRetries {
		b.WriteString(translationRetryButtonStyle.Render("Press 'r' to retry"))
	} else {
		b.WriteString(translationMaxRetriesStyle.Render("Maximum retries reached. Check configuration."))
	}

	return b.String()
}

// renderTranslationResult shows successful translation
func (tv *TranslationView) renderTranslationResult() string {
	var b strings.Builder

	b.WriteString(translationSuccessStyle.Render("Translation Complete"))
	b.WriteString("\n\n")

	// Original text preview
	originalPreview := tv.originalText
	if len(originalPreview) > 100 {
		originalPreview = originalPreview[:97] + "..."
	}
	b.WriteString(translationOriginalStyle.Render("Original: " + originalPreview))
	b.WriteString("\n\n")

	// Translated text
	b.WriteString(translationResultStyle.Render("Translation:"))
	b.WriteString("\n")
	b.WriteString(translationTextStyle.Render(tv.translatedText))
	b.WriteString("\n\n")

	// Basic metadata
	if tv.lastResult != nil {
		metadata := fmt.Sprintf("Model: %s | Tokens: %d | Duration: %v",
			tv.lastResult.Model,
			tv.lastResult.TokensUsed,
			tv.lastResult.Duration.Round(time.Millisecond))

		if tv.lastResult.Cached {
			metadata += " | Cached ✓"
		}

		b.WriteString(translationMetadataStyle.Render(metadata))
	}

	return b.String()
}

// renderWelcome shows welcome message
func (tv *TranslationView) renderWelcome() string {
	var b strings.Builder

	b.WriteString(translationWelcomeStyle.Render("Ready for Translation"))
	b.WriteString("\n\n")

	config := tv.configManager.GetEnhanced()
	if config.Translation.Enabled {
		b.WriteString(fmt.Sprintf("Target Language: %s", config.Translation.TargetLanguage))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Model: %s", config.OpenAI.Model))

		if config.Translation.CacheEnabled {
			b.WriteString("\nTranslation cache enabled")
		}
	} else {
		b.WriteString(translationDisabledStyle.Render("Translation is currently disabled"))
		b.WriteString("\nEnable it in the configuration settings")
	}

	return b.String()
}

// renderTranslationDetails shows detailed translation information
func (tv *TranslationView) renderTranslationDetails() string {
	if tv.lastResult == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(translationDetailsHeaderStyle.Render("Translation Details"))
	b.WriteString("\n\n")

	details := []string{
		fmt.Sprintf("Source Language: %s", tv.lastResult.SourceLanguage),
		fmt.Sprintf("Target Language: %s", tv.lastResult.TargetLanguage),
		fmt.Sprintf("Model Used: %s", tv.lastResult.Model),
		fmt.Sprintf("Tokens Used: %d", tv.lastResult.TokensUsed),
		fmt.Sprintf("Duration: %v", tv.lastResult.Duration.Round(time.Millisecond)),
		fmt.Sprintf("Attempt Count: %d", tv.lastResult.AttemptCount),
		fmt.Sprintf("Timestamp: %s", tv.lastResult.Timestamp.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("Cached: %v", tv.lastResult.Cached),
	}

	if tv.lastResult.CircuitBreaker != "" {
		details = append(details, fmt.Sprintf("Circuit Breaker: %s", tv.lastResult.CircuitBreaker))
	}

	for _, detail := range details {
		b.WriteString("• " + detail + "\n")
	}

	return b.String()
}

// renderCacheStats shows cache performance information
func (tv *TranslationView) renderCacheStats() string {
	var b strings.Builder

	b.WriteString(translationCacheHeaderStyle.Render("Cache Statistics"))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("Total Requests: %d | Cache Hits: %d | Hit Rate: %.1f%%",
		tv.cacheStats.TotalRequests,
		tv.cacheStats.CacheHits,
		tv.cacheStats.HitRate*100))

	return b.String()
}

// renderHelp shows available keyboard shortcuts
func (tv *TranslationView) renderHelp() string {
	var b strings.Builder

	if tv.isTranslating {
		b.WriteString(translationHelpStyle.Render("Press Ctrl+C or Esc to cancel translation"))
	} else {
		helps := []string{}

		if tv.lastError != nil && tv.retryCount < tv.maxRetries {
			helps = append(helps, "r: retry translation")
		}

		if tv.lastResult != nil {
			helps = append(helps, "d: toggle details")
			helps = append(helps, "c: clear translation")
		}

		if len(helps) > 0 {
			b.WriteString(translationHelpStyle.Render(strings.Join(helps, " | ")))
		}
	}

	return b.String()
}

// Helper methods

// getUserFriendlyError converts technical errors to user-friendly messages
func (tv *TranslationView) getUserFriendlyError(err error) string {
	if err == nil {
		return "Unknown error occurred"
	}

	errMsg := err.Error()

	// Check for common error patterns and provide helpful messages
	switch {
	case strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "unauthorized"):
		return "API key authentication failed. Please check your OpenAI API key in settings."

	case strings.Contains(errMsg, "rate_limit") || strings.Contains(errMsg, "too_many_requests"):
		return "Rate limit exceeded. Please wait a moment and try again, or upgrade your OpenAI plan."

	case strings.Contains(errMsg, "quota_exceeded") || strings.Contains(errMsg, "insufficient_quota"):
		return "API quota exceeded. Please check your OpenAI billing and usage limits."

	case strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection"):
		return "Network connection failed. Please check your internet connection and try again."

	case strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline exceeded"):
		return "Request timed out. The API might be slow - please try again."

	case strings.Contains(errMsg, "circuit breaker"):
		return "Service temporarily unavailable due to repeated failures. Please wait and try again."

	case strings.Contains(errMsg, "not configured"):
		return "Translation service not configured. Please set up your OpenAI API key in settings."

	default:
		// Return the original error but make it more user-friendly
		return fmt.Sprintf("Translation failed: %s", errMsg)
	}
}

// updateCacheStats updates cache performance statistics
func (tv *TranslationView) updateCacheStats() {
	if tv.translationService == nil {
		return
	}

	metrics := tv.translationService.GetMetrics()
	tv.cacheStats.TotalRequests = metrics.TotalRequests
	tv.cacheStats.CacheHits = metrics.CacheHits
	tv.cacheStats.CacheMisses = metrics.CacheMisses

	if tv.cacheStats.TotalRequests > 0 {
		tv.cacheStats.HitRate = float64(tv.cacheStats.CacheHits) / float64(tv.cacheStats.TotalRequests)
	}
}

// clearTranslation resets the translation state
func (tv *TranslationView) clearTranslation() {
	tv.originalText = ""
	tv.translatedText = ""
	tv.lastResult = nil
	tv.lastError = nil
	tv.retryCount = 0
	tv.progress.SetPercent(0)
}

// IsTranslating returns whether translation is in progress
func (tv *TranslationView) IsTranslating() bool {
	return tv.isTranslating
}

// HasResult returns whether there's a translation result available
func (tv *TranslationView) HasResult() bool {
	return tv.lastResult != nil
}

// HasError returns whether there's a translation error
func (tv *TranslationView) HasError() bool {
	return tv.lastError != nil
}

// GetTranslatedText returns the current translated text
func (tv *TranslationView) GetTranslatedText() string {
	return tv.translatedText
}

// Styles for translation view
var (
	translationHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("12")).
				Bold(true).
				Padding(0, 1)

	translationSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)

	translationErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Bold(true)

	translationErrorDetailStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("9")).
					Italic(true)

	translationRetryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("3")).
				Bold(true)

	translationRetryButtonStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("10")).
					Background(lipgloss.Color("8")).
					Padding(0, 1)

	translationMaxRetriesStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("1")).
					Italic(true)

	translationStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Italic(true)

	translationOriginalStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("8")).
					Italic(true)

	translationResultStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("14")).
				Bold(true)

	translationTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("8")).
				Padding(1, 2)

	translationMetadataStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("8")).
					Italic(true)

	translationWelcomeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("14")).
				Bold(true)

	translationDisabledStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("3")).
					Italic(true)

	translationDetailsHeaderStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("13")).
					Bold(true)

	translationCacheHeaderStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("11")).
					Bold(true)

	translationHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Italic(true)
)
