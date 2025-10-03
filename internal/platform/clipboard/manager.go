package clipboard

import (
	"fmt"
	"os"
	"os/exec"
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

	tmpPath, cleanup, err := CreateTempFile(content, ".txt")
	if err != nil {
		return &ClipboardError{
			Platform: m.platform,
			Command:  "tempfile",
			Err:      fmt.Errorf("failed to create temp file: %v", err),
		}
	}
	defer cleanup()

	f, err := os.Open(tmpPath)
	if err != nil {
		return &ClipboardError{
			Platform: m.platform,
			Command:  "tempfile",
			Err:      fmt.Errorf("failed to open temp file: %v", err),
		}
	}
	defer f.Close()

	var cmd *exec.Cmd
	switch toolName {
	case "xclip":
		cmd = exec.Command("xclip", "-selection", "clipboard", "-i")
	case "xsel":
		cmd = exec.Command("xsel", "--clipboard", "--input")
	default:
		return m.Copy(content)
	}

	cmd.Stdin = f

	if err := cmd.Run(); err != nil {
		return &ClipboardError{
			Platform: m.platform,
			Command:  toolName,
			Err:      err,
		}
	}

	return nil
}


func (m *Manager) GetStatus() ClipboardStatus {
	platform := m.platform
	if m.clipboard != nil {
		platform = m.clipboard.GetPlatform()
	}

	status := ClipboardStatus{
		Platform:  platform,
		Available: m.IsAvailable(),
		Tools:     make([]ToolStatus, 0, len(m.tools)),
	}

	cmd, _ := m.clipboard.GetCommand()

	for _, tool := range m.tools {
		toolStatus := ToolStatus{
			Name:      tool.Name,
			Command:   tool.Command,
			Available: tool.Available,
			Selected:  tool.Command == cmd,
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

			// Also update the platform-specific implementation
			if m.clipboard != nil {
				if err := m.clipboard.SetSelectedTool(toolName); err != nil {
					return err
				}
			}

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
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
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