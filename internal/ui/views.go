package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"shotgun-cli/internal/core"
)

// Translation message types
type translationCompleteMsg struct {
	textType string // "task" or "rules"
	result   *core.EnhancedTranslationResult
	err      error
}

type translationStartMsg struct {
	textType string // "task" or "rules"
}

// Translation commands
func (m *Model) translateText(textType, text string) tea.Cmd {
	if m.translator == nil {
		// Translator not initialized (likely due to missing API key or config issues)
		return func() tea.Msg {
			return translationCompleteMsg{
				textType: textType,
				result: &core.EnhancedTranslationResult{
					OriginalText:   text,
					TranslatedText: text, // Use original text when translation unavailable
					TargetLanguage: "en",
					Timestamp:      time.Now(),
					TokensUsed:     0,
					Model:          "unavailable",
					Duration:       0,
					Cached:         false,
					AttemptCount:   0,
					ApiProvider:    "none",
				},
				err: fmt.Errorf("translation unavailable: translator not initialized (check API key configuration)"),
			}
		}
	}

	if !m.translator.IsConfigured() {
		// Translator exists but is not properly configured
		return func() tea.Msg {
			return translationCompleteMsg{
				textType: textType,
				result: &core.EnhancedTranslationResult{
					OriginalText:   text,
					TranslatedText: text, // Use original text when translation unavailable
					TargetLanguage: "en",
					Timestamp:      time.Now(),
					TokensUsed:     0,
					Model:          "not_configured",
					Duration:       0,
					Cached:         false,
					AttemptCount:   0,
					ApiProvider:    "none",
				},
				err: fmt.Errorf("translation unavailable: translator not properly configured (check API key)"),
			}
		}
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var result *core.EnhancedTranslationResult
		var err error

		// All translation now goes through TranslateText with appropriate context
		result, err = m.translator.TranslateText(ctx, text, textType)

		return translationCompleteMsg{
			textType: textType,
			result:   result,
			err:      err,
		}
	}
}

// File Exclusion View
func (m *Model) updateFileExclusion(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		// Continue to next step
		m.includedFiles = core.GetIncludedFiles(m.fileTree_root, m.selection)
		m.currentView = ViewTemplateSelection
		return m, nil
	}

	// Let the file tree handle other keys
	var cmd tea.Cmd
	m.fileTree, cmd = m.fileTree.Update(msg)
	return m, cmd
}

func (m *Model) renderFileExclusion() string {
	title := titleStyle.Render("shotgun-cli - File Exclusion")

	// Show full current directory path
	dirPath := m.selectedDir
	if len(dirPath) > 60 {
		dirPath = "..." + dirPath[len(dirPath)-57:]
	}
	subtitle := subtitleStyle.Render(fmt.Sprintf("Working Directory: %s", dirPath))

	content := []string{
		title,
		"",
		subtitle,
		"",
	}

	if m.fileTree_root != nil {
		content = append(content, m.fileTree.View())
	} else {
		content = append(content, "Scanning directory...")
	}

	if m.lastError != nil {
		content = append(content, "", errorStyle.Render("Error: "+m.lastError.Error()))
	}

	return strings.Join(content, "\n")
}

// Template Selection View
func (m *Model) updateTemplateSelection(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Confirm template selection and move to task description
		m.currentView = ViewTaskDescription
		return m, m.taskInput.Focus()

	case "up", "k":
		// Navigate to previous template
		if m.templateIndex > 0 {
			m.templateIndex--
		} else {
			m.templateIndex = len(core.AvailableTemplates) - 1 // Wrap to last
		}
		m.currentTemplate = core.AvailableTemplates[m.templateIndex].Key
		return m, nil

	case "down", "j":
		// Navigate to next template
		if m.templateIndex < len(core.AvailableTemplates)-1 {
			m.templateIndex++
		} else {
			m.templateIndex = 0 // Wrap to first
		}
		m.currentTemplate = core.AvailableTemplates[m.templateIndex].Key
		return m, nil

	case "1":
		m.templateIndex = 0
		m.currentTemplate = core.AvailableTemplates[0].Key
		return m, nil
	case "2":
		if len(core.AvailableTemplates) > 1 {
			m.templateIndex = 1
			m.currentTemplate = core.AvailableTemplates[1].Key
		}
		return m, nil
	case "3":
		if len(core.AvailableTemplates) > 2 {
			m.templateIndex = 2
			m.currentTemplate = core.AvailableTemplates[2].Key
		}
		return m, nil
	case "4":
		if len(core.AvailableTemplates) > 3 {
			m.templateIndex = 3
			m.currentTemplate = core.AvailableTemplates[3].Key
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) renderTemplateSelection() string {
	title := titleStyle.Render("shotgun-cli - Template Selection")
	subtitle := subtitleStyle.Render(fmt.Sprintf("Files selected: %d | Step 2 of 4", len(m.includedFiles)))

	// Enhanced template selection with full descriptions
	templateSection := "Select a template for your prompt:\n\n"

	for i, tmpl := range core.AvailableTemplates {
		numberPrefix := fmt.Sprintf("%d. ", i+1)

		if i == m.templateIndex {
			indicator := "▶ "
			// Highlight selected template with description on same line
			templateSection += statusStyle.Render(fmt.Sprintf("%s%s%s - %s\n", indicator, numberPrefix, tmpl.Name, tmpl.Description))
		} else {
			indicator := "  "
			// Regular template with description on same line
			templateSection += fmt.Sprintf("%s%s%s - ", indicator, numberPrefix, tmpl.Name)
			templateSection += subtitleStyle.Render(tmpl.Description) + "\n"
		}
		templateSection += "\n"
	}

	content := []string{
		title,
		"",
		subtitle,
		"",
		templateSection,
		helpStyle.Render("↑/↓ (k/j): navigate | 1-4: quick select | Enter: confirm | Esc: back"),
	}

	return strings.Join(content, "\n")
}

// Task Description View
func (m *Model) updateTaskDescription(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.String() {
	case "f5":
		// Capture task text and check if translation is enabled
		taskText := m.taskInput.Value()
		m.taskText = taskText

		config := m.configMgr.Get()
		if config.Translation.Enabled && strings.TrimSpace(taskText) != "" {
			// Start translation
			m.translating = true
			m.translationStatus = "Translating task..."
			cmd := m.translateText("task", taskText)
			return m, cmd
		} else {
			// No translation needed, continue to next step
			m.currentView = ViewCustomRules
			return m, m.rulesInput.Focus()
		}

	case "tab":
		// Toggle focus on the task input (for consistency)
		if m.taskInput.Focused() {
			m.taskInput.Blur()
		} else {
			return m, m.taskInput.Focus()
		}
		return m, nil
	}

	// Handle task input updates
	var cmd tea.Cmd
	m.taskInput, cmd = m.taskInput.Update(msg)
	return m, cmd
}

func (m *Model) renderTaskDescription() string {
	// Adjust textarea for available screen space
	if m.width > 0 && m.height > 0 {
		m.taskInput.SetFullScreenMode(m.width, m.height)
	}

	title := titleStyle.Render("shotgun-cli - Task Description")

	// Show selected template for context
	selectedTemplate := core.AvailableTemplates[m.templateIndex]
	templateInfo := subtitleStyle.Render(fmt.Sprintf("Template: %s | Step 3 of 4", selectedTemplate.Name))

	// Task input section with instructions
	instructions := "Describe what you want to accomplish:"

	taskSection := fmt.Sprintf("%s\n\n%s", instructions, m.taskInput.View())

	// Examples section (more compact for full-screen)
	examplesSection := helpStyle.Render(`Examples: • Implement user authentication • Add validation to forms • Fix memory leak in data processing`)

	content := []string{
		title,
		"",
		templateInfo,
		"",
		taskSection,
		"",
		examplesSection,
		"",
	}

	// Add translation status if active
	if m.translating && m.translationStatus != "" {
		translationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33")) // Yellow
		content = append(content, translationStyle.Render("🔄 "+m.translationStatus), "")
	}

	content = append(content, helpStyle.Render("Enter: auto-number | Tab: focus field | F5: continue | Esc: back"))

	return strings.Join(content, "\n")
}

// Custom Rules View
func (m *Model) updateCustomRules(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.String() {
	case "f5":
		// Capture both task and rules text
		m.taskText = m.taskInput.Value()
		rulesText := m.rulesInput.Value()
		if rulesText == "" {
			rulesText = "no additional rules"
		}
		m.rulesText = rulesText

		config := m.configMgr.Get()
		if config.Translation.Enabled && strings.TrimSpace(rulesText) != "" && rulesText != "no additional rules" {
			// Start translation for rules
			m.translating = true
			m.translationStatus = "Translating rules..."
			cmd := m.translateText("rules", rulesText)
			return m, cmd
		} else {
			// No translation needed, proceed to generation
			m.currentView = ViewGeneration
			return m, m.generatePrompt()
		}

	case "tab":
		// Toggle focus on the rules input (for consistency)
		if m.rulesInput.Focused() {
			m.rulesInput.Blur()
		} else {
			return m, m.rulesInput.Focus()
		}
		return m, nil
	}

	// Handle rules input updates
	var cmd tea.Cmd
	m.rulesInput, cmd = m.rulesInput.Update(msg)
	return m, cmd
}

func (m *Model) renderCustomRules() string {
	// Adjust textarea for available screen space
	if m.width > 0 && m.height > 0 {
		m.rulesInput.SetFullScreenMode(m.width, m.height)
	}

	title := titleStyle.Render("shotgun-cli - Custom Rules")

	// Show selected template and progress for context
	selectedTemplate := core.AvailableTemplates[m.templateIndex]
	templateInfo := subtitleStyle.Render(fmt.Sprintf("Template: %s | Step 4 of 4", selectedTemplate.Name))

	// Rules input section with instructions
	instructions := "Add custom rules and constraints (optional):"

	rulesSection := fmt.Sprintf("%s\n\n%s", instructions, m.rulesInput.View())

	// Examples section for rules (more compact for full-screen)
	examplesSection := helpStyle.Render(`Examples: • Use TypeScript • Focus on performance • Follow existing patterns • Consider mobile
Optional: Leave empty if no specific rules needed`)

	content := []string{
		title,
		"",
		templateInfo,
		"",
		rulesSection,
		"",
		examplesSection,
		"",
	}

	// Add translation status if active
	if m.translating && m.translationStatus != "" {
		translationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33")) // Yellow
		content = append(content, translationStyle.Render("🔄 "+m.translationStatus), "")
	}

	content = append(content, helpStyle.Render("Tab: focus field | F5: generate | Esc: back"))

	return strings.Join(content, "\n")
}

// Configuration View
func (m *Model) updateConfiguration(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Exit configuration, return to file exclusion
		m.currentView = ViewFileExclusion
		return m, nil
	}

	// Let the configuration form handle the message
	var cmd tea.Cmd
	m.configForm, cmd = m.configForm.Update(tea.KeyMsg(msg))
	return m, cmd
}

// Generation View
func (m *Model) updateGeneration(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Cancel generation
		if m.generationCancel != nil {
			m.generationCancel()
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) renderGeneration() string {
	title := titleStyle.Render("shotgun-cli - Generating Context")
	subtitle := subtitleStyle.Render("Please wait while the prompt is being generated...")

	progressText := ""
	if m.progress.Phase != "" {
		progressText = fmt.Sprintf("%s: %s (%.1f%%)",
			m.progress.Phase,
			m.progress.CurrentFile,
			m.progress.Percentage)
	}

	content := []string{
		title,
		"",
		subtitle,
		"",
		m.progressBar.ViewAs(m.progress.Percentage / 100.0),
		"",
		progressText,
		"",
		helpStyle.Render("Ctrl+C: cancel generation"),
	}

	if m.lastError != nil {
		content = append(content, "", errorStyle.Render("Error: "+m.lastError.Error()))
	}

	return strings.Join(content, "\n")
}

// Complete View
func (m *Model) updateComplete(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m, tea.Quit
	case "v":
		// View the output file (simplified - just show a message)
		return m, nil
	case "n":
		// Start new prompt generation - reset to file exclusion
		m.currentView = ViewFileExclusion
		m.fileTree_root = nil
		m.includedFiles = nil
		m.taskInput.SetValue("")
		m.rulesInput.SetValue("")
		m.selection.Reset()
		m.lastError = nil
		// Reset template selection to default
		m.templateIndex = 0
		m.currentTemplate = core.TemplateDevKey
		return m, m.scanDirectory()
	}
	return m, nil
}

func (m *Model) renderComplete() string {
	title := titleStyle.Render("shotgun-cli - Generation Complete")

	var subtitle string
	if m.outputPath != "" {
		subtitle = subtitleStyle.Render("Prompt saved to: " + m.outputPath)
	} else {
		subtitle = subtitleStyle.Render("Prompt generation completed")
	}

	// Show some statistics
	stats := fmt.Sprintf("Files processed: %d", len(m.includedFiles))
	if m.outputPath != "" {
		if info, err := os.Stat(m.outputPath); err == nil {
			sizeMB := float64(info.Size()) / (1024 * 1024)
			stats += fmt.Sprintf(" | Output size: %.1f MB", sizeMB)
		}
	}

	content := []string{
		title,
		"",
		subtitle,
		"",
		statusStyle.Render("✓ Success!"),
		"",
		stats,
		"",
		helpStyle.Render("Enter: quit | n: new prompt | v: view output"),
	}

	if m.lastError != nil {
		content = append(content, "", errorStyle.Render("Error: "+m.lastError.Error()))
	}

	return strings.Join(content, "\n")
}

// Async operations
func (m *Model) scanDirectory() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		root, err := m.scanner.ScanDirectory(ctx, m.selectedDir)
		if err != nil {
			return errorMsg{err: err}
		}

		return scanCompleteMsg{root: root}
	}
}

func (m *Model) generatePrompt() tea.Cmd {
	return func() tea.Msg {
		// Create cancellable context
		ctx, cancel := context.WithCancel(context.Background())
		m.generationCtx = ctx
		m.generationCancel = cancel
		m.generating = true

		// Generate context
		fileContext, err := m.generator.GenerateContextWithTree(ctx, m.fileTree_root, m.includedFiles)
		if err != nil {
			return errorMsg{err: err}
		}

		templateData := core.TemplateData{
			Task:          m.taskText,
			Rules:         m.rulesText,
			CurrentDate:   time.Now().Format("2006-01-02"),
			FileStructure: fileContext,
		}

		prompt, err := m.templates.GeneratePrompt(m.currentTemplate, templateData)
		if err != nil {
			return errorMsg{err: err}
		}

		// Save to file with timestamp
		timestamp := time.Now().Format("2006-01-02_150405")
		outputFilename := fmt.Sprintf("shotgun_prompt_%s.md", timestamp)

		err = os.WriteFile(outputFilename, []byte(prompt), 0644)
		if err != nil {
			return errorMsg{err: err}
		}

		return generationCompleteMsg(outputFilename)
	}
}
