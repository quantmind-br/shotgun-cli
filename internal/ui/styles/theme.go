package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	PrimaryColor   = lipgloss.Color("#00ADD8")
	SecondaryColor = lipgloss.Color("#5E81AC")
	AccentColor    = lipgloss.Color("#A3BE8C")
	ErrorColor     = lipgloss.Color("#BF616A")
	WarningColor   = lipgloss.Color("#EBCB8B")
	SuccessColor   = lipgloss.Color("#A3BE8C")
	MutedColor     = lipgloss.Color("#5C7E8C")
	BorderColor    = lipgloss.Color("#434C5E")

	// Base styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor)

	SelectedStyle = lipgloss.NewStyle().
			Background(PrimaryColor).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1)

	ProgressStyle = lipgloss.NewStyle().
			Foreground(AccentColor)

	TreeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ECEFF4"))
)
// SelectionState represents the selection state of a file or directory node
type SelectionState int

const (
	SelectionUnselected SelectionState = iota
	SelectionPartial
	SelectionSelected
)

// Selection state colors for file tree
var (
	FileUnselectedColor = lipgloss.Color("#5C7E8C") // Muted blue-gray
	FileSelectedColor   = lipgloss.Color("#A3BE8C") // Success green
	FilePartialColor    = lipgloss.Color("#EBCB8B") // Warning yellow

	// Styles for file/directory names based on selection state
	UnselectedNameStyle = lipgloss.NewStyle().
		Foreground(FileUnselectedColor)

	SelectedNameStyle = lipgloss.NewStyle().
		Foreground(FileSelectedColor).
		Bold(true)

	PartialNameStyle = lipgloss.NewStyle().
		Foreground(FilePartialColor).
		Bold(true)
)

// Helper functions for common styling operations

func RenderHeader(step int, title string) string {
	header := fmt.Sprintf("Step %d/5: %s", step, title)
	return TitleStyle.Render(header)
}

func RenderFooter(shortcuts []string) string {
	parts := make([]string, 0, len(shortcuts))
	for _, shortcut := range shortcuts {
		parts = append(parts, HelpStyle.Render(shortcut))
	}
	return strings.Join(parts, " • ")
}

func RenderModal(content string) string {
	return BorderStyle.Render(content)
}

func RenderError(message string) string {
	return ErrorStyle.Render("❌ " + message)
}

func RenderSuccess(message string) string {
	return SuccessStyle.Render("✅ " + message)
}

func RenderWarning(message string) string {
	warningStyle := lipgloss.NewStyle().
		Foreground(WarningColor).
		Bold(true)
	return warningStyle.Render("⚠ " + message)
}

func RenderProgressBar(current, total int64, width int) string {
	if total <= 0 || width <= 0 {
		return strings.Repeat("░", width)
	}

	filled := int(float64(width) * float64(current) / float64(total))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return ProgressStyle.Render(bar)
}

func RenderBox(content string, title string) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(1, 2).
		Margin(1, 0)

	if title != "" {
		titleStyle := TitleStyle.
			Padding(0, 1).
			Background(lipgloss.Color("#FFFFFF")).
			Foreground(PrimaryColor)

		styledTitle := titleStyle.Render(title)
		content = styledTitle + "\n\n" + content
	}

	return boxStyle.Render(content)
}

func RenderList(items []string, selected int) string {
	var content strings.Builder

	for i, item := range items {
		if i == selected {
			line := SelectedStyle.Render("▶ " + item)
			content.WriteString(line)
		} else {
			content.WriteString("  " + item)
		}

		if i < len(items)-1 {
			content.WriteString("\n")
		}
	}

	return content.String()
}

func RenderTable(headers []string, rows [][]string) string {
	if len(headers) == 0 || len(rows) == 0 {
		return ""
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	var content strings.Builder

	// Render header
	for i, header := range headers {
		padding := colWidths[i] - len(header)
		content.WriteString(TitleStyle.Render(header))
		content.WriteString(strings.Repeat(" ", padding))
		if i < len(headers)-1 {
			content.WriteString(" | ")
		}
	}
	content.WriteString("\n")

	// Render separator
	for i, width := range colWidths {
		content.WriteString(strings.Repeat("─", width))
		if i < len(colWidths)-1 {
			content.WriteString("─┼─")
		}
	}
	content.WriteString("\n")

	// Render rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				padding := colWidths[i] - len(cell)
				content.WriteString(cell)
				content.WriteString(strings.Repeat(" ", padding))
				if i < len(row)-1 && i < len(colWidths)-1 {
					content.WriteString(" | ")
				}
			}
		}
		content.WriteString("\n")
	}

	return content.String()
}

func RenderKeyValue(key, value string) string {
	keyStyle := TitleStyle.Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ECEFF4"))

	return keyStyle.Render(key+":") + " " + valueStyle.Render(value)
}

func RenderSeparator(width int) string {
	if width <= 0 {
		width = 50
	}
	return HelpStyle.Render(strings.Repeat("─", width))
}

func RenderCenter(text string, width int) string {
	if width <= 0 || len(text) >= width {
		return text
	}

	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text
}

func RenderFileTree(isDirectory bool, isSelected bool, isExpanded bool, name string) string {
	var icon string
	if isDirectory {
		if isExpanded {
			icon = "📂"
		} else {
			icon = "📁"
		}
	} else {
		icon = "📄"
	}

	checkbox := ""
	if !isDirectory {
		if isSelected {
			checkbox = "[✓] "
		} else {
			checkbox = "[ ] "
		}
	}

	return TreeStyle.Render(checkbox + icon + " " + name)
}

// RenderFileName applies color styling to file/directory names based on selection state
func RenderFileName(name string, selectionState SelectionState) string {
	switch selectionState {
	case SelectionSelected:
		return SelectedNameStyle.Render(name)
	case SelectionPartial:
		return PartialNameStyle.Render(name)
	case SelectionUnselected:
		return UnselectedNameStyle.Render(name)
	default:
		return TreeStyle.Render(name)
	}
}
