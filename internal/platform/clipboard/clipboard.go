package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

type ClipboardManager interface {
	Copy(content string) error
	CopyWithTimeout(content string, timeout time.Duration) error
	IsAvailable() bool
	GetCommand() (string, []string)
	GetPlatform() string
}

type ClipboardError struct {
	Platform string
	Command  string
	Err      error
}

func (e *ClipboardError) Error() string {
	return fmt.Sprintf("clipboard error on %s using %s: %v", e.Platform, e.Command, e.Err)
}

type ClipboardTool struct {
	Name      string
	Command   string
	Args      []string
	Available bool
	Priority  int
}

type Manager struct {
	platform     string
	tools        []ClipboardTool
	selectedTool *ClipboardTool
	clipboard    ClipboardManager
}

func NewManager() *Manager {
	manager := &Manager{
		platform: runtime.GOOS,
	}

	switch manager.platform {
	case "linux":
		if isWSL() {
			manager.clipboard = NewWSLClipboard()
		} else {
			manager.clipboard = NewLinuxClipboard()
		}
	case "darwin":
		manager.clipboard = NewDarwinClipboard()
	case "windows":
		manager.clipboard = NewWindowsClipboard()
	default:
		manager.clipboard = NewLinuxClipboard()
	}

	manager.initializeTools()
	return manager
}

func (m *Manager) initializeTools() {
	switch m.platform {
	case "linux":
		m.tools = []ClipboardTool{
			{Name: "wl-copy", Command: "wl-copy", Args: []string{}, Priority: 1},
			{Name: "xclip", Command: "xclip", Args: []string{"-selection", "clipboard"}, Priority: 2},
			{Name: "xsel", Command: "xsel", Args: []string{"--clipboard", "--input"}, Priority: 3},
		}
	case "darwin":
		m.tools = []ClipboardTool{
			{Name: "pbcopy", Command: "pbcopy", Args: []string{}, Priority: 1},
		}
	case "windows":
		m.tools = []ClipboardTool{
			{Name: "clip", Command: "clip", Args: []string{}, Priority: 1},
			{Name: "powershell", Command: "powershell", Args: []string{"-Command", "Set-Clipboard"}, Priority: 2},
		}
	}

	m.checkAvailability()
}

func (m *Manager) checkAvailability() {
	var bestTool *ClipboardTool

	for i := range m.tools {
		tool := &m.tools[i]
		_, err := exec.LookPath(tool.Command)
		tool.Available = err == nil

		if tool.Available && (bestTool == nil || tool.Priority < bestTool.Priority) {
			bestTool = tool
		}
	}

	m.selectedTool = bestTool
}

func (m *Manager) Copy(content string) error {
	if m.clipboard != nil {
		return m.clipboard.Copy(content)
	}
	return &ClipboardError{
		Platform: m.platform,
		Command:  "none",
		Err:      fmt.Errorf("no clipboard implementation available"),
	}
}

func (m *Manager) CopyWithTimeout(content string, timeout time.Duration) error {
	if m.clipboard != nil {
		return m.clipboard.CopyWithTimeout(content, timeout)
	}
	return &ClipboardError{
		Platform: m.platform,
		Command:  "none",
		Err:      fmt.Errorf("no clipboard implementation available"),
	}
}

func (m *Manager) IsAvailable() bool {
	if m.clipboard != nil {
		return m.clipboard.IsAvailable()
	}
	return false
}

func (m *Manager) GetCommand() (string, []string) {
	if m.clipboard != nil {
		return m.clipboard.GetCommand()
	}
	return "", nil
}

func (m *Manager) GetPlatform() string {
	return m.platform
}

func isWSL() bool {
	wslClipboard := &WSLClipboard{}
	return wslClipboard.detectWSL()
}