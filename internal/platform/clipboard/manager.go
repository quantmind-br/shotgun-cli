package clipboard

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	MaxClipboardSize = 10 * 1024 * 1024 // 10MB
	LargeContentSize = 5 * 1024 * 1024   // 5MB
)

type ClipboardStatus struct {
	Platform  string
	Available bool
	Tools     []ToolStatus
}

type ToolStatus struct {
	Name      string
	Command   string
	Available bool
	Selected  bool
}

func (m *Manager) CopyLarge(content string) error {
	if len(content) > MaxClipboardSize {
		return &ClipboardError{
			Platform: m.platform,
			Command:  "size-check",
			Err:      fmt.Errorf("content too large (%d bytes, max %d bytes)", len(content), MaxClipboardSize),
		}
	}

	if len(content) > LargeContentSize && m.platform == "linux" {
		return m.copyViaTempFile(content)
	}

	return m.Copy(content)
}

func (m *Manager) copyLargeContent(content string) error {
	if len(content) > MaxClipboardSize {
		return &ClipboardError{
			Platform: m.platform,
			Command:  "size-check",
			Err:      fmt.Errorf("content exceeds maximum clipboard size"),
		}
	}

	return m.copyViaTempFile(content)
}

func (m *Manager) copyViaTempFile(content string) error {
	if m.platform != "linux" {
		return m.Copy(content)
	}

	linuxClipboard, ok := m.clipboard.(*LinuxClipboard)
	if !ok || linuxClipboard.selectedTool == nil {
		return m.Copy(content)
	}

	toolName := linuxClipboard.selectedTool.Name
	if toolName != "xclip" && toolName != "xsel" {
		return m.Copy(content)
	}

	tmpFile, err := os.CreateTemp("", "clipboard-*.txt")
	if err != nil {
		return &ClipboardError{
			Platform: m.platform,
			Command:  "tempfile",
			Err:      fmt.Errorf("failed to create temp file: %v", err),
		}
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if _, err := tmpFile.WriteString(content); err != nil {
		return &ClipboardError{
			Platform: m.platform,
			Command:  "tempfile",
			Err:      fmt.Errorf("failed to write temp file: %v", err),
		}
	}
	tmpFile.Close()

	var cmd *exec.Cmd
	switch toolName {
	case "xclip":
		cmd = exec.Command("xclip", "-selection", "clipboard", tmpFile.Name())
	case "xsel":
		cmd = exec.Command("xsel", "--clipboard", "--input", "<", tmpFile.Name())
	default:
		return m.Copy(content)
	}

	if err := cmd.Run(); err != nil {
		return &ClipboardError{
			Platform: m.platform,
			Command:  toolName,
			Err:      err,
		}
	}

	return nil
}

func (m *Manager) copyWithContext(ctx context.Context, content string) error {
	if m.clipboard == nil {
		return &ClipboardError{
			Platform: m.platform,
			Command:  "none",
			Err:      fmt.Errorf("no clipboard implementation available"),
		}
	}

	if m.selectedTool == nil {
		return &ClipboardError{
			Platform: m.platform,
			Command:  "none",
			Err:      fmt.Errorf("no clipboard tools available"),
		}
	}

	cmd := exec.CommandContext(ctx, m.selectedTool.Command, m.selectedTool.Args...)
	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ClipboardError{
				Platform: m.platform,
				Command:  m.selectedTool.Command,
				Err:      fmt.Errorf("clipboard operation timed out"),
			}
		}
		return &ClipboardError{
			Platform: m.platform,
			Command:  m.selectedTool.Command,
			Err:      err,
		}
	}

	return nil
}

func (m *Manager) GetStatus() ClipboardStatus {
	status := ClipboardStatus{
		Platform:  m.platform,
		Available: m.IsAvailable(),
		Tools:     make([]ToolStatus, 0, len(m.tools)),
	}

	for _, tool := range m.tools {
		toolStatus := ToolStatus{
			Name:      tool.Name,
			Command:   tool.Command,
			Available: tool.Available,
			Selected:  m.selectedTool != nil && m.selectedTool.Name == tool.Name,
		}
		status.Tools = append(status.Tools, toolStatus)
	}

	return status
}

func (m *Manager) GetAvailableTools() []string {
	var tools []string
	for _, tool := range m.tools {
		if tool.Available {
			tools = append(tools, tool.Name)
		}
	}
	return tools
}

func (m *Manager) GetSelectedTool() string {
	if m.selectedTool != nil {
		return m.selectedTool.Name
	}
	return ""
}

func (m *Manager) ForceToolSelection(toolName string) error {
	for i := range m.tools {
		tool := &m.tools[i]
		if tool.Name == toolName {
			if !tool.Available {
				return &ClipboardError{
					Platform: m.platform,
					Command:  toolName,
					Err:      fmt.Errorf("tool %s is not available", toolName),
				}
			}
			m.selectedTool = tool
			return nil
		}
	}

	return &ClipboardError{
		Platform: m.platform,
		Command:  toolName,
		Err:      fmt.Errorf("unknown tool %s", toolName),
	}
}

func (m *Manager) RefreshAvailability() {
	m.checkAvailability()
}

func CreateTempFile(content string, suffix string) (string, func(), error) {
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("clipboard-*%s", suffix))
	if err != nil {
		return "", nil, err
	}

	cleanup := func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		cleanup()
		return "", nil, err
	}

	if err := tmpFile.Close(); err != nil {
		cleanup()
		return "", nil, err
	}

	return tmpFile.Name(), cleanup, nil
}