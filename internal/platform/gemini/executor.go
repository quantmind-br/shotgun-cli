package gemini

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Result contains the result of geminiweb execution.
type Result struct {
	// Response is the Gemini response text.
	Response string

	// RawResponse is the unprocessed response from geminiweb.
	RawResponse string

	// Model used in the request.
	Model string

	// Duration of the execution.
	Duration time.Duration
}

// Executor manages geminiweb execution.
type Executor struct {
	config Config
}

// NewExecutor creates a new executor with the given configuration.
func NewExecutor(config Config) *Executor {
	return &Executor{config: config}
}

// Send sends content to Gemini and returns the response.
func (e *Executor) Send(ctx context.Context, content string) (*Result, error) {
	binaryPath, err := e.config.FindBinary()
	if err != nil {
		return nil, fmt.Errorf("geminiweb not available: %w", err)
	}

	if !IsConfigured() {
		return nil, fmt.Errorf("geminiweb not configured. Run: geminiweb auto-login")
	}

	// Build arguments
	args := e.buildArgs()

	// Create context with timeout
	timeout := time.Duration(e.config.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	log.Debug().
		Str("binary", binaryPath).
		Strs("args", args).
		Int("content_length", len(content)).
		Msg("Executing geminiweb")

	startTime := time.Now()

	// Create command
	cmd := exec.CommandContext(ctx, binaryPath, args...)

	// Configure stdin with content
	cmd.Stdin = strings.NewReader(content)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("geminiweb timed out after %v", timeout)
		}

		stderrStr := stderr.String()
		if stderrStr != "" {
			return nil, fmt.Errorf("geminiweb execution failed: %w\nstderr: %s", err, stderrStr)
		}
		return nil, fmt.Errorf("geminiweb execution failed: %w", err)
	}

	duration := time.Since(startTime)
	rawResponse := stdout.String()

	// Strip ANSI codes and parse response
	cleanResponse := StripANSI(rawResponse)
	parsedResponse := ParseResponse(cleanResponse)

	log.Debug().
		Dur("duration", duration).
		Int("response_length", len(parsedResponse)).
		Msg("geminiweb completed")

	return &Result{
		Response:    parsedResponse,
		RawResponse: rawResponse,
		Model:       e.config.Model,
		Duration:    duration,
	}, nil
}

// SendWithProgress sends content with progress callback.
func (e *Executor) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*Result, error) {
	progress("Locating geminiweb...")

	binaryPath, err := e.config.FindBinary()
	if err != nil {
		return nil, fmt.Errorf("geminiweb not available: %w", err)
	}

	if !IsConfigured() {
		return nil, fmt.Errorf("geminiweb not configured. Run: geminiweb auto-login")
	}

	progress("Preparing request...")

	args := e.buildArgs()

	timeout := time.Duration(e.config.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	progress(fmt.Sprintf("Sending to Gemini (%s)...", e.config.Model))

	startTime := time.Now()
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Stdin = strings.NewReader(content)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout after %v", timeout)
		}
		stderrStr := stderr.String()
		if stderrStr != "" {
			return nil, fmt.Errorf("failed: %s", stderrStr)
		}
		return nil, fmt.Errorf("failed: %w", err)
	}

	progress("Processing response...")

	duration := time.Since(startTime)
	rawResponse := stdout.String()
	cleanResponse := StripANSI(rawResponse)
	parsedResponse := ParseResponse(cleanResponse)

	return &Result{
		Response:    parsedResponse,
		RawResponse: rawResponse,
		Model:       e.config.Model,
		Duration:    duration,
	}, nil
}

// buildArgs builds command-line arguments for geminiweb.
func (e *Executor) buildArgs() []string {
	var args []string

	if e.config.Model != "" {
		args = append(args, "-m", e.config.Model)
	}

	if e.config.BrowserRefresh != "" {
		args = append(args, "--browser-refresh", e.config.BrowserRefresh)
	}

	return args
}

// StripANSI removes ANSI escape codes from a string.
func StripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Skip ANSI escape sequence until terminating letter
			j := i + 2
			for j < len(s) && !isTerminatingByte(s[j]) {
				j++
			}
			if j < len(s) {
				j++ // Include the terminating letter
			}
			i = j
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}

// isTerminatingByte checks if byte terminates an ANSI sequence.
func isTerminatingByte(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
