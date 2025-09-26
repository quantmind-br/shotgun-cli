package clipboard

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestClipboardManager_Interface(t *testing.T) {
	manager := NewManager()

	var _ ClipboardManager = manager.clipboard

	if manager.clipboard == nil {
		t.Fatal("clipboard implementation should not be nil")
	}

	if manager.GetPlatform() != runtime.GOOS {
		t.Errorf("expected platform %s, got %s", runtime.GOOS, manager.GetPlatform())
	}
}

func TestManager_PlatformDetection(t *testing.T) {
	tests := []struct {
		name         string
		platform     string
		expectedType string
	}{
		{"Linux", "linux", "*clipboard.LinuxClipboard"},
		{"Darwin", "darwin", "*clipboard.DarwinClipboard"},
		{"Windows", "windows", "*clipboard.WindowsClipboard"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{platform: tt.platform}

			switch tt.platform {
			case "linux":
				if !mockIsWSL(false) {
					manager.clipboard = NewLinuxClipboard()
				} else {
					manager.clipboard = NewWSLClipboard()
				}
			case "darwin":
				manager.clipboard = NewDarwinClipboard()
			case "windows":
				manager.clipboard = NewWindowsClipboard()
			}

			if manager.clipboard == nil {
				t.Fatalf("clipboard implementation should not be nil for platform %s", tt.platform)
			}
		})
	}
}

func TestClipboardError(t *testing.T) {
	err := &ClipboardError{
		Platform: "test",
		Command:  "testcmd",
		Err:      fmt.Errorf("test error"),
	}

	expected := "clipboard error on test using testcmd: test error"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func TestLinuxClipboard_ToolDetection(t *testing.T) {
	lc := &LinuxClipboard{
		tools: []ClipboardTool{
			{Name: "wl-copy", Command: "wl-copy", Args: []string{}, Priority: 1},
			{Name: "xclip", Command: "xclip", Args: []string{"-selection", "clipboard"}, Priority: 2},
			{Name: "xsel", Command: "xsel", Args: []string{"--clipboard", "--input"}, Priority: 3},
		},
	}

	lc.detectAvailableTools()

	foundAvailableTool := false
	for _, tool := range lc.tools {
		if tool.Available {
			foundAvailableTool = true
			break
		}
	}

	if foundAvailableTool && lc.selectedTool == nil {
		t.Error("expected a tool to be selected when available tools exist")
	}
}

func TestDarwinClipboard_Basic(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping macOS-specific test")
	}

	dc := NewDarwinClipboard()

	if dc.GetPlatform() != "darwin" {
		t.Errorf("expected platform darwin, got %s", dc.GetPlatform())
	}

	cmd, args := dc.GetCommand()
	if cmd != "pbcopy" {
		t.Errorf("expected command pbcopy, got %s", cmd)
	}
	if len(args) != 0 {
		t.Errorf("expected no args, got %v", args)
	}
}

func TestWindowsClipboard_ToolSelection(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("skipping Windows-specific test")
	}

	wc := NewWindowsClipboard()

	if wc.preferredTool == "" {
		t.Error("expected a preferred tool to be selected")
	}

	cmd, args := wc.GetCommand()
	if cmd == "" {
		t.Error("expected a command to be returned")
	}

	if wc.preferredTool == "clip" && cmd != "clip" {
		t.Errorf("expected clip command when preferred tool is clip, got %s", cmd)
	}

	if wc.preferredTool == "powershell" && cmd != "powershell" {
		t.Errorf("expected powershell command when preferred tool is powershell, got %s", cmd)
	}

	if wc.preferredTool == "powershell" {
		expectedArgs := []string{"-Command", "Set-Clipboard"}
		if len(args) != len(expectedArgs) {
			t.Errorf("expected args %v, got %v", expectedArgs, args)
		}
		for i, arg := range expectedArgs {
			if i < len(args) && args[i] != arg {
				t.Errorf("expected arg[%d] %s, got %s", i, arg, args[i])
			}
		}
	}
}

func TestWSLClipboard_Detection(t *testing.T) {
	wsl := NewWSLClipboard()

	if wsl.detectWSL() && !wsl.isWSL {
		t.Error("WSL detection mismatch")
	}

	if wsl.GetPlatform() != "wsl" {
		t.Errorf("expected platform wsl, got %s", wsl.GetPlatform())
	}

	if wsl.isWSL {
		cmd, args := wsl.GetCommand()
		if cmd != "clip.exe" {
			t.Errorf("expected clip.exe command in WSL, got %s", cmd)
		}
		if len(args) != 0 {
			t.Errorf("expected no args for clip.exe, got %v", args)
		}
	}
}

func TestManager_Timeout(t *testing.T) {
	manager := NewManager()

	if !manager.IsAvailable() {
		t.Skip("no clipboard tools available")
	}

	testContent := "timeout test content"
	timeout := time.Millisecond * 100

	err := manager.CopyWithTimeout(testContent, timeout)
	if err != nil {
		clipErr, ok := err.(*ClipboardError)
		if ok && strings.Contains(clipErr.Err.Error(), "timed out") {
			t.Log("timeout test passed - operation timed out as expected")
		} else {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestManager_LargeContent(t *testing.T) {
	manager := NewManager()

	if !manager.IsAvailable() {
		t.Skip("no clipboard tools available")
	}

	largeContent := strings.Repeat("a", MaxClipboardSize+1)

	err := manager.CopyLarge(largeContent)
	if err == nil {
		t.Error("expected error for content exceeding maximum size")
	}

	clipErr, ok := err.(*ClipboardError)
	if !ok {
		t.Errorf("expected ClipboardError, got %T", err)
	}

	if !strings.Contains(clipErr.Err.Error(), "too large") {
		t.Errorf("expected 'too large' in error message, got: %v", clipErr.Err)
	}
}

func TestManager_Status(t *testing.T) {
	manager := NewManager()
	status := manager.GetStatus()

	if status.Platform != runtime.GOOS {
		t.Errorf("expected platform %s, got %s", runtime.GOOS, status.Platform)
	}

	if len(status.Tools) == 0 {
		t.Error("expected at least one tool in status")
	}

	selectedCount := 0
	for _, tool := range status.Tools {
		if tool.Selected {
			selectedCount++
		}
		if tool.Name == "" {
			t.Error("tool name should not be empty")
		}
		if tool.Command == "" {
			t.Error("tool command should not be empty")
		}
	}

	if manager.IsAvailable() && selectedCount != 1 {
		t.Errorf("expected exactly 1 selected tool when available, got %d", selectedCount)
	}
}

func TestManager_ToolSelection(t *testing.T) {
	manager := NewManager()

	availableTools := manager.GetAvailableTools()
	selectedTool := manager.GetSelectedTool()

	if manager.IsAvailable() {
		if len(availableTools) == 0 {
			t.Error("expected available tools when clipboard is available")
		}
		if selectedTool == "" {
			t.Error("expected selected tool when clipboard is available")
		}

		found := false
		for _, tool := range availableTools {
			if tool == selectedTool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("selected tool %s not found in available tools %v", selectedTool, availableTools)
		}
	}
}

func TestManager_ForceToolSelection(t *testing.T) {
	manager := NewManager()

	if !manager.IsAvailable() {
		t.Skip("no clipboard tools available")
	}

	availableTools := manager.GetAvailableTools()
	if len(availableTools) == 0 {
		t.Skip("no available tools to test")
	}

	firstTool := availableTools[0]
	err := manager.ForceToolSelection(firstTool)
	if err != nil {
		t.Errorf("unexpected error forcing tool selection: %v", err)
	}

	if manager.GetSelectedTool() != firstTool {
		t.Errorf("expected selected tool %s, got %s", firstTool, manager.GetSelectedTool())
	}

	err = manager.ForceToolSelection("nonexistent-tool")
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}
}

func TestCreateTempFile(t *testing.T) {
	content := "test content for temp file"
	suffix := ".txt"

	tmpPath, cleanup, err := CreateTempFile(content, suffix)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer cleanup()

	if !strings.HasSuffix(tmpPath, suffix) {
		t.Errorf("expected temp file to have suffix %s, got %s", suffix, tmpPath)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	if string(data) != content {
		t.Errorf("expected temp file content %q, got %q", content, string(data))
	}

	cleanup()

	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("expected temp file to be cleaned up")
	}
}

func BenchmarkClipboard_SmallContent(b *testing.B) {
	manager := NewManager()
	if !manager.IsAvailable() {
		b.Skip("no clipboard tools available")
	}

	content := "small test content"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := manager.Copy(content)
		if err != nil {
			b.Fatalf("clipboard copy failed: %v", err)
		}
	}
}

func BenchmarkClipboard_LargeContent(b *testing.B) {
	manager := NewManager()
	if !manager.IsAvailable() {
		b.Skip("no clipboard tools available")
	}

	content := strings.Repeat("large content test ", 10000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := manager.Copy(content)
		if err != nil {
			b.Fatalf("clipboard copy failed: %v", err)
		}
	}
}

func mockIsWSL(isWSL bool) bool {
	return isWSL
}

func mockExecLookPath(cmd string, available bool) error {
	if available {
		return nil
	}
	return exec.ErrNotFound
}

func TestClipboard_CrossPlatformBuild(t *testing.T) {
	platforms := []string{"linux", "darwin", "windows"}

	for _, platform := range platforms {
		t.Run(fmt.Sprintf("Platform_%s", platform), func(t *testing.T) {
			manager := &Manager{platform: platform}

			switch platform {
			case "linux":
				manager.clipboard = NewLinuxClipboard()
			case "darwin":
				manager.clipboard = NewDarwinClipboard()
			case "windows":
				manager.clipboard = NewWindowsClipboard()
			}

			if manager.clipboard == nil {
				t.Fatalf("failed to create clipboard for platform %s", platform)
			}

			if manager.clipboard.GetPlatform() != platform {
				t.Errorf("platform mismatch: expected %s, got %s", platform, manager.clipboard.GetPlatform())
			}
		})
	}
}