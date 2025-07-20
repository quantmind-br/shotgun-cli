package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"shotgun-cli/internal/core"
)

// File Exclusion View
func (m *Model) updateFileExclusion(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		// Continue to next step
		m.includedFiles = core.GetIncludedFiles(m.fileTree_root, m.selection)
		m.currentView = ViewPromptComposition
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

// Prompt Composition View
func (m *Model) updatePromptComposition(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "g":
		// Generate prompt
		m.taskText = m.taskInput.Value()
		m.rulesText = m.rulesInput.Value()
		if m.rulesText == "" {
			m.rulesText = "no additional rules"
		}
		m.currentView = ViewGeneration
		return m, m.generatePrompt()

	case "tab":
		// Switch focus between inputs
		if m.taskInput.Focused() {
			m.taskInput.Blur()
			m.rulesInput.Focus()
		} else {
			m.rulesInput.Blur()
			m.taskInput.Focus()
		}
		return m, nil

	case "1":
		m.currentTemplate = core.TemplateDevKey
		return m, nil
	case "2":
		m.currentTemplate = core.TemplateArchKey
		return m, nil
	case "3":
		m.currentTemplate = core.TemplateBugKey
		return m, nil
	case "4":
		m.currentTemplate = core.TemplateProjectKey
		return m, nil
	}

	var cmd tea.Cmd
	if m.taskInput.Focused() {
		m.taskInput, cmd = m.taskInput.Update(msg)
	} else {
		m.rulesInput, cmd = m.rulesInput.Update(msg)
	}
	return m, cmd
}

func (m *Model) renderPromptComposition() string {
	title := titleStyle.Render("shotgun-cli - Prompt Composition")
	subtitle := subtitleStyle.Render(fmt.Sprintf("Files selected: %d", len(m.includedFiles)))

	// Template selection
	templateSection := "Select Template:\n"
	for _, tmpl := range core.AvailableTemplates {
		indicator := " "
		if tmpl.Key == m.currentTemplate {
			indicator = ">"
		}
		templateSection += fmt.Sprintf("%s [%s] %s - %s\n",
			indicator,
			strings.ToUpper(tmpl.Key[:1]),
			tmpl.Name,
			tmpl.Description)
	}

	// Task input
	taskSection := "Task Description:\n" + m.taskInput.View()

	// Rules input
	rulesSection := "Custom Rules:\n" + m.rulesInput.View()

	content := []string{
		title,
		"",
		subtitle,
		"",
		templateSection,
		"",
		taskSection,
		"",
		rulesSection,
		"",
		helpStyle.Render("Tab: switch fields | 1-4: select template | Enter/g: generate | Esc: back"),
	}

	if m.lastError != nil {
		content = append(content, "", errorStyle.Render("Error: "+m.lastError.Error()))
	}

	return strings.Join(content, "\n")
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
	case "enter", "q":
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
		helpStyle.Render("Enter/q: quit | n: new prompt | v: view output"),
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

		// Generate prompt
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
