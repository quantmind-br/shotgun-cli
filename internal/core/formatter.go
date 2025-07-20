package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// NumberedLine represents a parsed numbered line
type NumberedLine struct {
	Number  int
	Content string
	IsEmpty bool
}

// FormattingOptions contains options for text formatting
type FormattingOptions struct {
	EnableNumbering bool
	StartNumber     int
	NumberFormat    string // e.g., "%d. " for "1. "
	SkipEmptyLines  bool
}

// DefaultFormattingOptions returns default formatting options
func DefaultFormattingOptions() FormattingOptions {
	return FormattingOptions{
		EnableNumbering: true,
		StartNumber:     1,
		NumberFormat:    "%d. ",
		SkipEmptyLines:  true,
	}
}

// ParseNumberedLines parses text and extracts numbered lines
func ParseNumberedLines(text string) ([]NumberedLine, error) {
	if text == "" {
		return []NumberedLine{}, nil
	}
	
	lines := strings.Split(text, "\n")
	result := make([]NumberedLine, 0, len(lines))
	
	// Regex to match numbered lines: "1. content", "2. content", etc.
	numberingRegex := regexp.MustCompile(`^\s*(\d+)\.\s*(.*)$`)
	
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Check if line is empty
		if trimmedLine == "" {
			result = append(result, NumberedLine{
				Number:  0,
				Content: "",
				IsEmpty: true,
			})
			continue
		}
		
		// Check if line is numbered
		if matches := numberingRegex.FindStringSubmatch(line); matches != nil {
			number, err := strconv.Atoi(matches[1])
			if err != nil {
				// Treat as unnumbered content if number parsing fails
				result = append(result, NumberedLine{
					Number:  0,
					Content: trimmedLine,
					IsEmpty: false,
				})
			} else {
				result = append(result, NumberedLine{
					Number:  number,
					Content: strings.TrimSpace(matches[2]),
					IsEmpty: false,
				})
			}
		} else {
			// Unnumbered content
			result = append(result, NumberedLine{
				Number:  0,
				Content: trimmedLine,
				IsEmpty: false,
			})
		}
	}
	
	return result, nil
}

// FormatNumberedText formats text with consistent numbering
func FormatNumberedText(text string, options FormattingOptions) (string, error) {
	if text == "" {
		return "", nil
	}
	
	// Parse existing content
	lines, err := ParseNumberedLines(text)
	if err != nil {
		return "", fmt.Errorf("failed to parse numbered lines: %w", err)
	}
	
	if !options.EnableNumbering {
		// Just return the content without numbering
		var result []string
		for _, line := range lines {
			if line.IsEmpty {
				result = append(result, "")
			} else {
				result = append(result, line.Content)
			}
		}
		return strings.Join(result, "\n"), nil
	}
	
	// Apply consistent numbering
	return applyNumbering(lines, options), nil
}

// ReformatWithNumbers applies automatic numbering to text
func ReformatWithNumbers(text string) (string, error) {
	options := DefaultFormattingOptions()
	return FormatNumberedText(text, options)
}

// applyNumbering applies numbering to parsed lines
func applyNumbering(lines []NumberedLine, options FormattingOptions) string {
	var result []string
	currentNumber := options.StartNumber
	
	for _, line := range lines {
		if line.IsEmpty {
			if !options.SkipEmptyLines {
				result = append(result, "")
			}
			continue
		}
		
		// Apply numbering format
		numberedLine := fmt.Sprintf(options.NumberFormat, currentNumber) + line.Content
		result = append(result, numberedLine)
		currentNumber++
	}
	
	return strings.Join(result, "\n")
}

// ValidateNumberedFormat checks if text has consistent numbering
func ValidateNumberedFormat(text string) (bool, []string, error) {
	lines, err := ParseNumberedLines(text)
	if err != nil {
		return false, nil, err
	}
	
	var issues []string
	expectedNumber := 1
	hasNumberedContent := false
	
	for i, line := range lines {
		if line.IsEmpty {
			continue
		}
		
		if line.Number > 0 {
			hasNumberedContent = true
			if line.Number != expectedNumber {
				issues = append(issues, fmt.Sprintf("Line %d: expected number %d, got %d", i+1, expectedNumber, line.Number))
			}
			expectedNumber++
		} else if hasNumberedContent {
			// Found unnumbered content after numbered content
			issues = append(issues, fmt.Sprintf("Line %d: unnumbered content after numbered content", i+1))
		}
	}
	
	return len(issues) == 0, issues, nil
}

// HasNumberedContent checks if text contains any numbered lines
func HasNumberedContent(text string) bool {
	lines, err := ParseNumberedLines(text)
	if err != nil {
		return false
	}
	
	for _, line := range lines {
		if line.Number > 0 {
			return true
		}
	}
	return false
}

// ExtractContentOnly removes numbering and returns just the content
func ExtractContentOnly(text string) (string, error) {
	lines, err := ParseNumberedLines(text)
	if err != nil {
		return "", err
	}
	
	var result []string
	for _, line := range lines {
		if line.IsEmpty {
			result = append(result, "")
		} else {
			result = append(result, line.Content)
		}
	}
	
	return strings.Join(result, "\n"), nil
}

// MergeNumberedTexts combines multiple numbered texts into one
func MergeNumberedTexts(texts []string, options FormattingOptions) (string, error) {
	var allLines []NumberedLine
	
	for _, text := range texts {
		if text == "" {
			continue
		}
		
		lines, err := ParseNumberedLines(text)
		if err != nil {
			return "", fmt.Errorf("failed to parse text: %w", err)
		}
		
		allLines = append(allLines, lines...)
	}
	
	if !options.EnableNumbering {
		// Return content without numbering
		var result []string
		for _, line := range allLines {
			if line.IsEmpty {
				result = append(result, "")
			} else {
				result = append(result, line.Content)
			}
		}
		return strings.Join(result, "\n"), nil
	}
	
	return applyNumbering(allLines, options), nil
}

// SplitIntoTasks splits numbered text into individual tasks
func SplitIntoTasks(text string) ([]string, error) {
	lines, err := ParseNumberedLines(text)
	if err != nil {
		return nil, err
	}
	
	var tasks []string
	for _, line := range lines {
		if !line.IsEmpty && line.Content != "" {
			tasks = append(tasks, line.Content)
		}
	}
	
	return tasks, nil
}

// CountNumberedItems returns the count of numbered items in text
func CountNumberedItems(text string) (int, error) {
	lines, err := ParseNumberedLines(text)
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, line := range lines {
		if line.Number > 0 {
			count++
		}
	}
	
	return count, nil
}