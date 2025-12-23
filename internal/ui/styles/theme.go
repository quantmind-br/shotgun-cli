// Package styles provides visual styling and theming for the TUI components.
package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Nord Color Palette - Professional dark theme for TUI
// https://www.nordtheme.com/docs/colors-and-palettes
var (
	// Polar Night - Dark shades for backgrounds
	Nord0 = lipgloss.Color("#2E3440") // Darkest background
	Nord1 = lipgloss.Color("#3B4252") // Darker elements
	Nord2 = lipgloss.Color("#434C5E") // UI elements
	Nord3 = lipgloss.Color("#4C566A") // Inactive elements

	// Snow Storm - Light shades for text
	Nord4 = lipgloss.Color("#D8DEE9") // Main text
	Nord5 = lipgloss.Color("#E5E9F0") // Brighter text
	Nord6 = lipgloss.Color("#ECEFF4") // Brightest text

	// Frost - Accent colors
	Nord7  = lipgloss.Color("#8FBCBB") // Teal (secondary)
	Nord8  = lipgloss.Color("#88C0D0") // Light blue (primary)
	Nord9  = lipgloss.Color("#81A1C1") // Blue (links, navigation)
	Nord10 = lipgloss.Color("#5E81AC") // Dark blue (selections)

	// Aurora - Semantic colors
	Nord11 = lipgloss.Color("#BF616A") // Red (errors, warnings)
	Nord12 = lipgloss.Color("#D08770") // Orange (attention)
	Nord13 = lipgloss.Color("#EBCB8B") // Yellow (warnings, partial)
	Nord14 = lipgloss.Color("#A3BE8C") // Green (success, selected)
	Nord15 = lipgloss.Color("#B48EAD") // Purple (special)

	// Semantic color aliases
	PrimaryColor   = Nord8  // Light blue - main interactive elements
	SecondaryColor = Nord10 // Dark blue - secondary elements
	AccentColor    = Nord14 // Green - success/selected
	ErrorColor     = Nord11 // Red - errors
	WarningColor   = Nord13 // Yellow - warnings
	SuccessColor   = Nord14 // Green - success
	MutedColor     = Nord3  // Gray - muted/inactive
	BorderColor    = Nord2  // Dark gray - borders
	TextColor      = Nord4  // Light gray - main text
	BrightText     = Nord6  // White - bright text
	DimText        = Nord3  // Dim text for secondary info

	// Base styles with improved Nord colors
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor)

	SubtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Nord9)

	SelectedStyle = lipgloss.NewStyle().
			Background(Nord10).
			Foreground(Nord6).
			Bold(true)

	CursorStyle = lipgloss.NewStyle().
			Background(Nord10).
			Foreground(Nord6)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true)

	// Border styles with rounded corners
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1)

	FocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PrimaryColor).
				Padding(1)

	BlurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(MutedColor).
				Padding(1)

	ProgressStyle = lipgloss.NewStyle().
			Foreground(AccentColor)

	TreeStyle = lipgloss.NewStyle().
			Foreground(TextColor)

	// Status indicator styles
	StatusActiveStyle = lipgloss.NewStyle().
				Foreground(SuccessColor).
				Bold(true)

	StatusInactiveStyle = lipgloss.NewStyle().
				Foreground(MutedColor)

	StatusWarningStyle = lipgloss.NewStyle().
				Foreground(WarningColor)

	// Code/content styles
	CodeStyle = lipgloss.NewStyle().
			Foreground(Nord7)

	PathStyle = lipgloss.NewStyle().
			Foreground(Nord9)

	// Input field styles
	InputLabelStyle = lipgloss.NewStyle().
			Foreground(Nord9).
			Bold(true)

	InputPlaceholderStyle = lipgloss.NewStyle().
				Foreground(DimText).
				Italic(true)

	// Stats/metrics styles
	StatsLabelStyle = lipgloss.NewStyle().
			Foreground(Nord9)

	StatsValueStyle = lipgloss.NewStyle().
			Foreground(Nord6).
			Bold(true)

	TokenCountStyle = lipgloss.NewStyle().
			Foreground(Nord15).
			Bold(true)
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
	FileUnselectedColor = MutedColor   // Gray
	FileSelectedColor   = SuccessColor // Green
	FilePartialColor    = WarningColor // Yellow

	// Styles for file/directory names based on selection state
	UnselectedNameStyle = lipgloss.NewStyle().
				Foreground(FileUnselectedColor)

	SelectedNameStyle = lipgloss.NewStyle().
				Foreground(FileSelectedColor).
				Bold(true)

	PartialNameStyle = lipgloss.NewStyle().
				Foreground(FilePartialColor).
				Bold(true)

	// Styles for ignored files
	GitIgnoredStyle = lipgloss.NewStyle().
			Foreground(Nord11). // Red for gitignored
			Italic(true)

	CustomIgnoredStyle = lipgloss.NewStyle().
				Foreground(Nord12). // Orange for custom ignored
				Italic(true)
)

// Helper functions for common styling operations

func RenderHeader(step int, title string) string {
	stepIndicator := lipgloss.NewStyle().
		Foreground(Nord15).
		Bold(true).
		Render(fmt.Sprintf("Step %d/5", step))

	titleStyled := TitleStyle.Render(title)
	separator := lipgloss.NewStyle().
		Foreground(MutedColor).
		Render(" • ")

	return stepIndicator + separator + titleStyled
}

func RenderFooter(shortcuts []string) string {
	parts := make([]string, 0, len(shortcuts))
	for i, shortcut := range shortcuts {
		// Alternate between two colors for visual separation
		var style lipgloss.Style
		if i%2 == 0 {
			style = lipgloss.NewStyle().Foreground(Nord9)
		} else {
			style = lipgloss.NewStyle().Foreground(Nord7)
		}
		parts = append(parts, style.Render(shortcut))
	}
	separator := lipgloss.NewStyle().Foreground(MutedColor).Render(" │ ")
	return strings.Join(parts, separator)
}

func RenderModal(content string) string {
	return BorderStyle.Render(content)
}

func RenderError(message string) string {
	icon := lipgloss.NewStyle().Foreground(ErrorColor).Render("✖")
	text := ErrorStyle.Render(message)
	return icon + " " + text
}

func RenderSuccess(message string) string {
	icon := lipgloss.NewStyle().Foreground(SuccessColor).Render("✔")
	text := SuccessStyle.Render(message)
	return icon + " " + text
}

func RenderWarning(message string) string {
	icon := lipgloss.NewStyle().Foreground(WarningColor).Render("⚠")
	text := WarningStyle.Render(message)
	return icon + " " + text
}

func RenderInfo(message string) string {
	icon := lipgloss.NewStyle().Foreground(PrimaryColor).Render("ℹ")
	text := lipgloss.NewStyle().Foreground(TextColor).Render(message)
	return icon + " " + text
}

func RenderProgressBar(current, total int64, width int) string {
	if total <= 0 || width <= 0 {
		return lipgloss.NewStyle().Foreground(MutedColor).Render(strings.Repeat("░", width))
	}

	filled := int(float64(width) * float64(current) / float64(total))
	if filled > width {
		filled = width
	}

	filledStyle := lipgloss.NewStyle().Foreground(AccentColor)
	emptyStyle := lipgloss.NewStyle().Foreground(MutedColor)

	bar := filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", width-filled))
	return bar
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
			Background(Nord1).
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
			cursor := lipgloss.NewStyle().Foreground(PrimaryColor).Render("▶")
			line := SelectedStyle.Render(" " + item)
			content.WriteString(cursor + line)
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

	colWidths := calculateColumnWidths(headers, rows)

	var content strings.Builder
	renderTableHeader(&content, headers, colWidths)
	renderTableSeparator(&content, colWidths)
	renderTableRows(&content, rows, colWidths)

	return content.String()
}

func calculateColumnWidths(headers []string, rows [][]string) []int {
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

	return colWidths
}

func renderTableHeader(content *strings.Builder, headers []string, colWidths []int) {
	for i, header := range headers {
		padding := colWidths[i] - len(header)
		content.WriteString(TitleStyle.Render(header))
		content.WriteString(strings.Repeat(" ", padding))
		if i < len(headers)-1 {
			separator := lipgloss.NewStyle().Foreground(MutedColor).Render(" │ ")
			content.WriteString(separator)
		}
	}
	content.WriteString("\n")
}

func renderTableSeparator(content *strings.Builder, colWidths []int) {
	for i, width := range colWidths {
		line := lipgloss.NewStyle().Foreground(MutedColor).Render(strings.Repeat("─", width))
		content.WriteString(line)
		if i < len(colWidths)-1 {
			cross := lipgloss.NewStyle().Foreground(MutedColor).Render("─┼─")
			content.WriteString(cross)
		}
	}
	content.WriteString("\n")
}

func renderTableRows(content *strings.Builder, rows [][]string, colWidths []int) {
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				padding := colWidths[i] - len(cell)
				content.WriteString(cell)
				content.WriteString(strings.Repeat(" ", padding))
				if i < len(row)-1 && i < len(colWidths)-1 {
					separator := lipgloss.NewStyle().Foreground(MutedColor).Render(" │ ")
					content.WriteString(separator)
				}
			}
		}
		content.WriteString("\n")
	}
}

func RenderKeyValue(key, value string) string {
	keyStyled := StatsLabelStyle.Render(key + ":")
	valueStyled := StatsValueStyle.Render(value)
	return keyStyled + " " + valueStyled
}

func RenderSeparator(width int) string {
	if width <= 0 {
		width = 50
	}
	return lipgloss.NewStyle().Foreground(MutedColor).Render(strings.Repeat("─", width))
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
			checkbox = SelectedNameStyle.Render("[✓]") + " "
		} else {
			checkbox = UnselectedNameStyle.Render("[ ]") + " "
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

// RenderIgnoreIndicator renders the ignore status indicator with appropriate styling
func RenderIgnoreIndicator(isGitignored, isCustomIgnored bool) string {
	if isGitignored {
		return GitIgnoredStyle.Render(" (gitignore)")
	}
	if isCustomIgnored {
		return CustomIgnoredStyle.Render(" (custom)")
	}
	return ""
}

// RenderTokenStats renders token statistics with appropriate styling
func RenderTokenStats(fileCount int, size, tokens string) string {
	countStyle := StatsValueStyle.Render(fmt.Sprintf("%d", fileCount))
	sizeStyle := StatsValueStyle.Render(size)
	tokenStyle := TokenCountStyle.Render("~" + tokens)

	return fmt.Sprintf("Selected: %s files (%s / %s tokens)", countStyle, sizeStyle, tokenStyle)
}

// RenderStepIndicator renders a step progress indicator
func RenderStepIndicator(current, total int) string {
	var parts []string
	for i := 1; i <= total; i++ {
		if i < current {
			// Completed step
			parts = append(parts, SuccessStyle.Render("●"))
		} else if i == current {
			// Current step
			parts = append(parts, lipgloss.NewStyle().Foreground(PrimaryColor).Bold(true).Render("◉"))
		} else {
			// Future step
			parts = append(parts, lipgloss.NewStyle().Foreground(MutedColor).Render("○"))
		}
	}
	return strings.Join(parts, " ")
}

// RenderSpinnerFrame returns a styled spinner frame
func RenderSpinnerFrame(frame string) string {
	return lipgloss.NewStyle().Foreground(PrimaryColor).Render(frame)
}
