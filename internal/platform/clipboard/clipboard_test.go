package clipboard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Compile-time assertions for interface compliance.
var (
	_ ClipboardManager = (*LinuxClipboard)(nil)
	_ ClipboardManager = (*DarwinClipboard)(nil)
	_ ClipboardManager = (*WindowsClipboard)(nil)
	_ ClipboardManager = (*WSLClipboard)(nil)
)

type fakeClipboard struct {
	platform      string
	command       string
	available     bool
	copyCalls     []string
	timeoutCalls  []time.Duration
	errorOnCopy   error
	errorOnTimed  error
	selectedTools []string
}

func (f *fakeClipboard) Copy(content string) error {
	f.copyCalls = append(f.copyCalls, content)
	return f.errorOnCopy
}

func (f *fakeClipboard) CopyWithTimeout(content string, timeout time.Duration) error {
	f.timeoutCalls = append(f.timeoutCalls, timeout)
	f.copyCalls = append(f.copyCalls, content)
	if f.errorOnTimed != nil {
		return f.errorOnTimed
	}
	return f.errorOnCopy
}

func (f *fakeClipboard) IsAvailable() bool { return f.available }

func (f *fakeClipboard) GetCommand() (string, []string) { return f.command, nil }

func (f *fakeClipboard) GetPlatform() string { return f.platform }

func (f *fakeClipboard) SetSelectedTool(name string) error {
	f.selectedTools = append(f.selectedTools, name)
	return nil
}

func TestManager_CopyDelegatesToImplementation(t *testing.T) {
	t.Parallel()

	fake := &fakeClipboard{platform: platformLinux, command: toolXclip, available: true}
	mgr := &Manager{platform: platformLinux, clipboard: fake}

	if err := mgr.Copy("hello"); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if fake.copyCalls[0] != "hello" {
		t.Fatalf("expected copy content to be recorded")
	}

	if err := mgr.CopyWithTimeout("world", time.Second); err != nil {
		t.Fatalf("CopyWithTimeout failed: %v", err)
	}
	if len(fake.timeoutCalls) == 0 {
		t.Fatalf("expected timeout call to be recorded")
	}
	if got := fake.timeoutCalls[len(fake.timeoutCalls)-1]; got != time.Second {
		t.Fatalf("expected explicit timeout, got %v", got)
	}
}

func TestManager_CopyLarge_SizeLimit(t *testing.T) {
	t.Parallel()

	fake := &fakeClipboard{platform: platformLinux, command: toolXclip, available: true}
	mgr := &Manager{platform: platformLinux, clipboard: fake}

	oversized := strings.Repeat("x", MaxClipboardSize+1)
	if err := mgr.CopyLarge(oversized); err == nil {
		t.Fatalf("expected error for oversized clipboard content")
	}

	allowed := strings.Repeat("y", 1024)
	if err := mgr.CopyLarge(allowed); err != nil {
		t.Fatalf("CopyLarge should succeed: %v", err)
	}
	if len(fake.copyCalls) == 0 {
		t.Fatalf("expected underlying copy to be invoked")
	}
}

func TestManager_ForceToolSelection(t *testing.T) {
	t.Parallel()

	fake := &fakeClipboard{platform: platformLinux, command: toolXclip, available: true}
	mgr := &Manager{
		platform:  platformLinux,
		clipboard: fake,
		tools:     []ClipboardTool{{Name: toolXclip, Command: toolXclip, Available: true}},
	}

	if err := mgr.ForceToolSelection(toolXclip); err != nil {
		t.Fatalf("expected tool selection to succeed: %v", err)
	}
	if mgr.selectedTool == nil || mgr.selectedTool.Name != toolXclip {
		t.Fatalf("expected selected tool to be xclip")
	}
	if len(fake.selectedTools) != 1 || fake.selectedTools[0] != toolXclip {
		t.Fatalf("expected clipboard implementation to receive SetSelectedTool call")
	}

	if err := mgr.ForceToolSelection("unknown"); err == nil {
		t.Fatalf("expected error for unknown tool")
	}
}

func TestManager_GetStatusReflectsTools(t *testing.T) {
	t.Parallel()

	fake := &fakeClipboard{platform: platformLinux, command: toolXclip, available: true}
	mgr := &Manager{
		platform:  platformLinux,
		clipboard: fake,
		tools:     []ClipboardTool{{Name: toolXclip, Command: toolXclip, Available: true}},
	}
	mgr.selectedTool = &mgr.tools[0]

	status := mgr.GetStatus()
	if !status.Available {
		t.Fatalf("expected status to be available")
	}
	if status.Platform != platformLinux {
		t.Fatalf("unexpected platform: %s", status.Platform)
	}
	if len(status.Tools) != 1 || !status.Tools[0].Selected {
		t.Fatalf("expected selected tool in status")
	}
}

func TestManager_InitializeTools_LinuxPriority(t *testing.T) {

	dir := t.TempDir()
	createFakeCommand(t, dir, "wl-copy")
	createFakeCommand(t, dir, "xclip")
	createFakeCommand(t, dir, "xsel")
	t.Setenv("PATH", dir)
	t.Setenv("WAYLAND_DISPLAY", "wayland-1")
	t.Setenv("DISPLAY", "")

	mgr := &Manager{platform: platformLinux}
	mgr.clipboard = NewLinuxClipboard()
	mgr.initializeTools()

	if mgr.selectedTool == nil || mgr.selectedTool.Name != "wl-copy" {
		t.Fatalf("expected wl-copy to be selected when available: %+v", mgr.selectedTool)
	}
	for _, tool := range mgr.tools {
		if !tool.Available {
			t.Fatalf("expected %s to be available", tool.Name)
		}
	}
}

func TestManager_InitializeTools_LinuxFallback(t *testing.T) {

	dir := t.TempDir()
	createFakeCommand(t, dir, "xclip")
	t.Setenv("PATH", dir)
	t.Setenv("WAYLAND_DISPLAY", "")
	t.Setenv("DISPLAY", ":0")

	mgr := &Manager{platform: platformLinux}
	mgr.clipboard = NewLinuxClipboard()
	mgr.initializeTools()

	if mgr.selectedTool == nil || mgr.selectedTool.Name != toolXclip {
		t.Fatalf("expected xclip fallback selection, got %+v", mgr.selectedTool)
	}
}

func TestWindowsClipboard_ToolPreference(t *testing.T) {

	dir := t.TempDir()
	createFakeCommand(t, dir, "clip")
	createFakeCommand(t, dir, "powershell")
	t.Setenv("PATH", dir)

	wc := NewWindowsClipboard()
	if wc.preferredTool != "clip" {
		t.Fatalf("expected clip to be preferred when available, got %s", wc.preferredTool)
	}
}

func TestDarwinClipboard_IsAvailable(t *testing.T) {

	dir := t.TempDir()
	createFakeCommand(t, dir, "pbcopy")
	t.Setenv("PATH", dir)

	dc := NewDarwinClipboard()
	if !dc.IsAvailable() {
		t.Fatalf("expected pbcopy to be detected as available")
	}
}

func TestWSLClipboard_DetectsViaEnv(t *testing.T) {
	t.Setenv("WSL_DISTRO_NAME", "Ubuntu-20.04")
	wsl := NewWSLClipboard()
	if !wsl.isWSL {
		t.Fatalf("expected WSL detection to succeed when WSL_DISTRO_NAME set")
	}
}

func TestCreateTempFileLifecycle(t *testing.T) {
	t.Parallel()

	content := "hello clipboard"
	path, cleanup, err := CreateTempFile(content, ".txt")
	if err != nil {
		t.Fatalf("CreateTempFile failed: %v", err)
	}
	if !strings.HasSuffix(path, ".txt") {
		t.Fatalf("expected suffix to be applied: %s", path)
	}
	cleanup()
}

func BenchmarkManager_Copy(b *testing.B) {
	fake := &fakeClipboard{platform: platformLinux, command: toolXclip, available: true}
	mgr := &Manager{platform: platformLinux, clipboard: fake}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := mgr.Copy("benchmark"); err != nil {
			b.Fatalf("Copy failed: %v", err)
		}
	}
}

func createFakeCommand(tb testing.TB, dir, name string) {
	tb.Helper()
	script := "#!/bin/sh\nexit 0\n"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		tb.Fatalf("failed to create fake command %s: %v", name, err)
	}
}
