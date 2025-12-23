package gemini

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ParseResponse extracts the relevant content from geminiweb output.
// It removes the Gemini header and cleans up the response.
func ParseResponse(raw string) string {
	if raw == "" {
		return ""
	}

	lines := strings.Split(raw, "\n")
	result := make([]string, 0, len(lines))
	skipNext := false

	for i, line := range lines {
		// Skip the Gemini header line (contains "✦ Gemini" or similar)
		if strings.Contains(line, "Gemini") && i < 3 {
			skipNext = true

			continue
		}

		// Skip empty line after header
		if skipNext && strings.TrimSpace(line) == "" {
			skipNext = false

			continue
		}

		skipNext = false
		result = append(result, line)
	}

	response := strings.Join(result, "\n")

	// Trim leading/trailing whitespace
	response = strings.TrimSpace(response)

	return response
}

// CodeBlock represents a code block extracted from the response.
type CodeBlock struct {
	Language string
	Code     string
}

// ExtractCodeBlocks extracts fenced code blocks from the response.
func ExtractCodeBlocks(response string) []CodeBlock {
	re := regexp.MustCompile("```(\\w*)\\n([\\s\\S]*?)```")
	matches := re.FindAllStringSubmatch(response, -1)

	blocks := make([]CodeBlock, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 3 {
			blocks = append(blocks, CodeBlock{
				Language: match[1],
				Code:     strings.TrimSpace(match[2]),
			})
		}
	}

	return blocks
}

// ExtractThoughts extracts thinking/reasoning sections from the response.
// Some Gemini models include a "thoughts" section prefixed with certain markers.
func ExtractThoughts(response string) (thoughts string, content string) {
	// Look for thought markers (commonly used in reasoning models)
	markers := []string{"<thinking>", "💭", "**Thinking:**", "## Thinking"}

	for _, marker := range markers {
		if idx := strings.Index(response, marker); idx != -1 {
			// Find end of thoughts section
			var endIdx int
			switch marker {
			case "<thinking>":
				endIdx = strings.Index(response[idx:], "</thinking>")
				if endIdx != -1 {
					endIdx += idx + len("</thinking>")
					thoughts = response[idx:endIdx]
					content = strings.TrimSpace(response[:idx] + response[endIdx:])

					return
				}
			default:
				// For other markers, find next double newline
				endIdx = strings.Index(response[idx:], "\n\n")
				if endIdx != -1 {
					endIdx += idx
					thoughts = strings.TrimSpace(response[idx:endIdx])
					content = strings.TrimSpace(response[:idx] + response[endIdx:])

					return
				}
			}
		}
	}

	return "", response
}

// SummarizeResponse creates a brief summary of the response.
func SummarizeResponse(response string, maxLength int) string {
	if len(response) <= maxLength {
		return response
	}

	// Try to cut at a sentence boundary
	truncated := response[:maxLength]
	lastPeriod := strings.LastIndex(truncated, ". ")
	if lastPeriod > maxLength/2 {
		return truncated[:lastPeriod+1] + "..."
	}

	// Fall back to word boundary
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > maxLength/2 {
		return truncated[:lastSpace] + "..."
	}

	return truncated + "..."
}

// ContainsError checks if the response indicates an error from Gemini.
func ContainsError(response string) (bool, string) {
	errorIndicators := []string{
		"I apologize",
		"I cannot",
		"I'm unable",
		"Error:",
		"error occurred",
		"rate limit",
		"quota exceeded",
		"authentication failed",
		"unauthorized",
	}

	lowerResponse := strings.ToLower(response)
	for _, indicator := range errorIndicators {
		if strings.Contains(lowerResponse, strings.ToLower(indicator)) {
			// Try to extract the error message
			lines := strings.Split(response, "\n")
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), strings.ToLower(indicator)) {
					return true, strings.TrimSpace(line)
				}
			}

			return true, indicator
		}
	}

	return false, ""
}

// FormatDuration formats a duration for display.
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}

	return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
}
