// Package gemini provides integration with geminiweb-go for sending context to Google Gemini.
package gemini

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Config contains configuration for geminiweb integration.
type Config struct {
	// BinaryPath is the path to geminiweb executable.
	// If empty, searches in PATH and common locations.
	BinaryPath string

	// Model is the Gemini model to use (gemini-2.5-flash, gemini-2.5-pro, gemini-3.0-pro).
	Model string

	// Timeout in seconds for execution.
	Timeout int

	// BrowserRefresh enables automatic cookie refresh via browser.
	// Valid values: "auto", "chrome", "firefox", "edge", "chromium", "opera", or empty to disable.
	BrowserRefresh string

	// Verbose enables detailed logging.
	Verbose bool
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		BinaryPath:     "",
		Model:          "gemini-2.5-flash",
		Timeout:        300, // 5 minutes
		BrowserRefresh: "auto",
		Verbose:        false,
	}
}

// ValidModels returns the list of valid Gemini model names.
func ValidModels() []string {
	return []string{
		"gemini-2.5-flash",
		"gemini-2.5-pro",
		"gemini-3.0-pro",
	}
}

// IsValidModel checks if the model name is valid.
func IsValidModel(model string) bool {
	for _, valid := range ValidModels() {
		if model == valid {
			return true
		}
	}
	return false
}

// FindBinary locates the geminiweb executable.
func (c *Config) FindBinary() (string, error) {
	// Check explicit path first
	if c.BinaryPath != "" {
		if _, err := os.Stat(c.BinaryPath); err == nil {
			return c.BinaryPath, nil
		}
		return "", fmt.Errorf("geminiweb binary not found at specified path: %s", c.BinaryPath)
	}

	// Search in PATH
	path, err := exec.LookPath("geminiweb")
	if err == nil {
		return path, nil
	}

	// Search in common locations
	home, _ := os.UserHomeDir()
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(home, "go")
	}

	commonPaths := []string{
		filepath.Join(gopath, "bin", "geminiweb"),
		filepath.Join(home, "go", "bin", "geminiweb"),
		"/usr/local/bin/geminiweb",
		"/usr/bin/geminiweb",
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("geminiweb binary not found in PATH or common locations. Install with: go install github.com/diogo/geminiweb/cmd/geminiweb@latest")
}

// IsAvailable checks if geminiweb binary is available.
func IsAvailable() bool {
	cfg := DefaultConfig()
	_, err := cfg.FindBinary()
	return err == nil
}

// IsConfigured checks if geminiweb is configured (cookies exist).
func IsConfigured() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	cookiesPath := filepath.Join(home, ".geminiweb", "cookies.json")
	info, err := os.Stat(cookiesPath)
	if err != nil {
		return false
	}
	// Check if file has some content (not empty)
	return info.Size() > 2 // At least "{}"
}

// GetCookiesPath returns the path to the geminiweb cookies file.
func GetCookiesPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".geminiweb", "cookies.json")
}

// Status represents the current status of geminiweb integration.
type Status struct {
	Available   bool
	Configured  bool
	BinaryPath  string
	CookiesPath string
	Error       string
}

// GetStatus returns the current status of geminiweb integration.
func GetStatus() Status {
	cfg := DefaultConfig()
	status := Status{
		CookiesPath: GetCookiesPath(),
	}

	binaryPath, err := cfg.FindBinary()
	if err != nil {
		status.Error = err.Error()
		return status
	}

	status.Available = true
	status.BinaryPath = binaryPath
	status.Configured = IsConfigured()

	if !status.Configured {
		status.Error = "geminiweb not configured. Run: geminiweb auto-login"
	}

	return status
}
