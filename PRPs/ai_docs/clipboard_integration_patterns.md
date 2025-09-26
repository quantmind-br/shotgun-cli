# Cross-Platform Clipboard Integration Patterns for Go

## Critical URLs and Documentation

### Go Libraries
- **golang-design/clipboard**: https://github.com/golang-design/clipboard
- **atotto/clipboard**: https://github.com/atotto/clipboard
- **Command execution**: https://pkg.go.dev/os/exec

### Platform-Specific Tools Documentation
- **Linux - wl-copy**: https://github.com/bugaevc/wl-clipboard
- **Linux - xclip**: https://github.com/astrand/xclip
- **Linux - xsel**: http://www.vergenet.net/~conrad/software/xsel/
- **macOS - pbcopy**: https://ss64.com/osx/pbcopy.html
- **Windows - clip.exe**: https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/clip

## Interface Design Pattern

```go
package clipboard

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "runtime"
    "strings"
    "time"
)

// ClipboardManager provides cross-platform clipboard access
type ClipboardManager interface {
    Copy(content string) error
    CopyWithTimeout(content string, timeout time.Duration) error
    IsAvailable() bool
    GetCommand() (string, []string)
    GetPlatform() string
}

// ClipboardError represents clipboard operation errors
type ClipboardError struct {
    Platform string
    Command  string
    Err      error
}

func (e ClipboardError) Error() string {
    return fmt.Sprintf("clipboard error on %s using %s: %v", e.Platform, e.Command, e.Err)
}

// Manager is the main clipboard manager implementation
type Manager struct {
    platform    string
    availableTools []ClipboardTool
    selectedTool   *ClipboardTool
}

type ClipboardTool struct {
    Name        string
    Command     string
    Args        []string
    Available   bool
    Priority    int  // Lower number = higher priority
}
```

## Platform Detection and Initialization

```go
func NewManager() *Manager {
    platform := runtime.GOOS
    manager := &Manager{
        platform: platform,
    }

    switch platform {
    case "linux":
        manager.availableTools = []ClipboardTool{
            {Name: "wl-copy", Command: "wl-copy", Priority: 1},
            {Name: "xclip", Command: "xclip", Args: []string{"-selection", "clipboard"}, Priority: 2},
            {Name: "xsel", Command: "xsel", Args: []string{"--clipboard", "--input"}, Priority: 3},
        }
    case "darwin":
        manager.availableTools = []ClipboardTool{
            {Name: "pbcopy", Command: "pbcopy", Priority: 1},
        }
    case "windows":
        manager.availableTools = []ClipboardTool{
            {Name: "clip", Command: "clip", Priority: 1},
            {Name: "powershell", Command: "powershell", Args: []string{"-Command", "Set-Clipboard"}, Priority: 2},
        }
    default:
        // Try common Linux tools as fallback
        manager.availableTools = []ClipboardTool{
            {Name: "xclip", Command: "xclip", Args: []string{"-selection", "clipboard"}, Priority: 1},
            {Name: "xsel", Command: "xsel", Args: []string{"--clipboard", "--input"}, Priority: 2},
        }
    }

    // Check availability and select best tool
    manager.checkAvailability()
    return manager
}

func (m *Manager) checkAvailability() {
    for i := range m.availableTools {
        tool := &m.availableTools[i]

        // Check if command exists
        _, err := exec.LookPath(tool.Command)
        tool.Available = (err == nil)

        // Select the first available tool (they're sorted by priority)
        if tool.Available && m.selectedTool == nil {
            m.selectedTool = tool
        }
    }
}
```

## Linux Implementation Pattern

```go
// LinuxClipboard handles Linux clipboard operations
type LinuxClipboard struct {
    tools []ClipboardTool
    selectedTool *ClipboardTool
}

func NewLinuxClipboard() *LinuxClipboard {
    tools := []ClipboardTool{
        {
            Name:     "wl-copy",
            Command:  "wl-copy",
            Args:     []string{},
            Priority: 1,
        },
        {
            Name:     "xclip",
            Command:  "xclip",
            Args:     []string{"-selection", "clipboard"},
            Priority: 2,
        },
        {
            Name:     "xsel",
            Command:  "xsel",
            Args:     []string{"--clipboard", "--input"},
            Priority: 3,
        },
    }

    lc := &LinuxClipboard{tools: tools}
    lc.detectAvailableTools()
    return lc
}

func (lc *LinuxClipboard) detectAvailableTools() {
    // Special handling for Wayland vs X11
    if os.Getenv("WAYLAND_DISPLAY") != "" {
        // Prioritize wl-copy on Wayland
        lc.checkTool("wl-copy")
    } else if os.Getenv("DISPLAY") != "" {
        // Prioritize xclip/xsel on X11
        lc.checkTool("xclip")
        if lc.selectedTool == nil {
            lc.checkTool("xsel")
        }
    }

    // Fallback: check all tools
    if lc.selectedTool == nil {
        for _, tool := range lc.tools {
            if lc.checkTool(tool.Name) {
                break
            }
        }
    }
}

func (lc *LinuxClipboard) checkTool(toolName string) bool {
    for i, tool := range lc.tools {
        if tool.Name == toolName {
            if _, err := exec.LookPath(tool.Command); err == nil {
                lc.tools[i].Available = true
                lc.selectedTool = &lc.tools[i]
                return true
            }
        }
    }
    return false
}

func (lc *LinuxClipboard) Copy(content string) error {
    if lc.selectedTool == nil {
        return fmt.Errorf("no clipboard tool available on Linux")
    }

    return lc.copyWithTool(*lc.selectedTool, content)
}

func (lc *LinuxClipboard) copyWithTool(tool ClipboardTool, content string) error {
    cmd := exec.Command(tool.Command, tool.Args...)
    cmd.Stdin = strings.NewReader(content)

    if err := cmd.Run(); err != nil {
        return ClipboardError{
            Platform: "linux",
            Command:  tool.Command,
            Err:      err,
        }
    }

    return nil
}
```

## macOS Implementation Pattern

```go
// DarwinClipboard handles macOS clipboard operations
type DarwinClipboard struct{}

func NewDarwinClipboard() *DarwinClipboard {
    return &DarwinClipboard{}
}

func (dc *DarwinClipboard) Copy(content string) error {
    cmd := exec.Command("pbcopy")
    cmd.Stdin = strings.NewReader(content)

    if err := cmd.Run(); err != nil {
        return ClipboardError{
            Platform: "darwin",
            Command:  "pbcopy",
            Err:      err,
        }
    }

    return nil
}

func (dc *DarwinClipboard) IsAvailable() bool {
    _, err := exec.LookPath("pbcopy")
    return err == nil
}

func (dc *DarwinClipboard) GetCommand() (string, []string) {
    return "pbcopy", []string{}
}
```

## Windows Implementation Pattern

```go
// WindowsClipboard handles Windows clipboard operations
type WindowsClipboard struct {
    preferredTool string
}

func NewWindowsClipboard() *WindowsClipboard {
    wc := &WindowsClipboard{}

    // Check availability and prefer clip.exe over PowerShell
    if _, err := exec.LookPath("clip"); err == nil {
        wc.preferredTool = "clip"
    } else {
        wc.preferredTool = "powershell"
    }

    return wc
}

func (wc *WindowsClipboard) Copy(content string) error {
    switch wc.preferredTool {
    case "clip":
        return wc.copyWithClip(content)
    case "powershell":
        return wc.copyWithPowerShell(content)
    default:
        return fmt.Errorf("no clipboard tool available on Windows")
    }
}

func (wc *WindowsClipboard) copyWithClip(content string) error {
    cmd := exec.Command("clip")
    cmd.Stdin = strings.NewReader(content)

    if err := cmd.Run(); err != nil {
        return ClipboardError{
            Platform: "windows",
            Command:  "clip",
            Err:      err,
        }
    }

    return nil
}

func (wc *WindowsClipboard) copyWithPowerShell(content string) error {
    // Escape content for PowerShell
    escaped := strings.ReplaceAll(content, "'", "''")

    cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Set-Clipboard -Value '%s'", escaped))

    if err := cmd.Run(); err != nil {
        return ClipboardError{
            Platform: "windows",
            Command:  "powershell",
            Err:      err,
        }
    }

    return nil
}

func (wc *WindowsClipboard) IsAvailable() bool {
    // Check both clip.exe and PowerShell
    _, clipErr := exec.LookPath("clip")
    _, psErr := exec.LookPath("powershell")

    return clipErr == nil || psErr == nil
}
```

## WSL Detection and Handling

```go
// WSLClipboard handles clipboard operations in WSL environment
type WSLClipboard struct {
    isWSL bool
}

func NewWSLClipboard() *WSLClipboard {
    return &WSLClipboard{
        isWSL: detectWSL(),
    }
}

func detectWSL() bool {
    // Check for WSL-specific indicators
    if _, err := os.Stat("/proc/version"); err == nil {
        if content, err := os.ReadFile("/proc/version"); err == nil {
            version := strings.ToLower(string(content))
            return strings.Contains(version, "microsoft") || strings.Contains(version, "wsl")
        }
    }

    // Check environment variable
    if os.Getenv("WSL_DISTRO_NAME") != "" {
        return true
    }

    return false
}

func (wc *WSLClipboard) Copy(content string) error {
    if !wc.isWSL {
        return fmt.Errorf("not running in WSL environment")
    }

    // Use Windows clip.exe from within WSL
    cmd := exec.Command("clip.exe")
    cmd.Stdin = strings.NewReader(content)

    if err := cmd.Run(); err != nil {
        return ClipboardError{
            Platform: "wsl",
            Command:  "clip.exe",
            Err:      err,
        }
    }

    return nil
}

func (wc *WSLClipboard) IsAvailable() bool {
    if !wc.isWSL {
        return false
    }

    _, err := exec.LookPath("clip.exe")
    return err == nil
}
```

## Unified Manager with Timeout Support

```go
func (m *Manager) Copy(content string) error {
    return m.CopyWithTimeout(content, 10*time.Second)
}

func (m *Manager) CopyWithTimeout(content string, timeout time.Duration) error {
    if m.selectedTool == nil {
        return fmt.Errorf("no clipboard tool available")
    }

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    return m.copyWithContext(ctx, content)
}

func (m *Manager) copyWithContext(ctx context.Context, content string) error {
    args := append(m.selectedTool.Args)
    cmd := exec.CommandContext(ctx, m.selectedTool.Command, args...)
    cmd.Stdin = strings.NewReader(content)

    if err := cmd.Run(); err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("clipboard operation timed out")
        }
        return ClipboardError{
            Platform: m.platform,
            Command:  m.selectedTool.Command,
            Err:      err,
        }
    }

    return nil
}

func (m *Manager) IsAvailable() bool {
    return m.selectedTool != nil && m.selectedTool.Available
}

func (m *Manager) GetCommand() (string, []string) {
    if m.selectedTool == nil {
        return "", nil
    }
    return m.selectedTool.Command, m.selectedTool.Args
}

func (m *Manager) GetPlatform() string {
    return m.platform
}

// GetStatus returns detailed information about clipboard availability
func (m *Manager) GetStatus() ClipboardStatus {
    status := ClipboardStatus{
        Platform:  m.platform,
        Available: m.selectedTool != nil,
        Tools:     make([]ToolStatus, len(m.availableTools)),
    }

    for i, tool := range m.availableTools {
        status.Tools[i] = ToolStatus{
            Name:      tool.Name,
            Command:   tool.Command,
            Available: tool.Available,
            Selected:  m.selectedTool != nil && tool.Name == m.selectedTool.Name,
        }
    }

    return status
}

type ClipboardStatus struct {
    Platform  string       `json:"platform"`
    Available bool         `json:"available"`
    Tools     []ToolStatus `json:"tools"`
}

type ToolStatus struct {
    Name      string `json:"name"`
    Command   string `json:"command"`
    Available bool   `json:"available"`
    Selected  bool   `json:"selected"`
}
```

## Large Content Handling

```go
const MaxClipboardSize = 10 * 1024 * 1024 // 10MB limit

func (m *Manager) CopyLarge(content string) error {
    if len(content) > MaxClipboardSize {
        return fmt.Errorf("content too large for clipboard: %d bytes (max: %d)", len(content), MaxClipboardSize)
    }

    // For very large content, consider chunking or streaming
    if len(content) > 1024*1024 { // 1MB
        return m.copyLargeContent(content)
    }

    return m.Copy(content)
}

func (m *Manager) copyLargeContent(content string) error {
    // Use a temporary file for very large content on some platforms
    if m.platform == "linux" && len(content) > 5*1024*1024 {
        return m.copyViaTempFile(content)
    }

    return m.Copy(content)
}

func (m *Manager) copyViaTempFile(content string) error {
    // Create temporary file
    tmpFile, err := os.CreateTemp("", "shotgun-clipboard-*.txt")
    if err != nil {
        return fmt.Errorf("failed to create temp file: %w", err)
    }
    defer os.Remove(tmpFile.Name())
    defer tmpFile.Close()

    // Write content to temp file
    if _, err := tmpFile.WriteString(content); err != nil {
        return fmt.Errorf("failed to write to temp file: %w", err)
    }

    tmpFile.Close()

    // Use appropriate command to copy from file
    var cmd *exec.Cmd
    switch m.selectedTool.Name {
    case "xclip":
        cmd = exec.Command("xclip", "-selection", "clipboard", "-i", tmpFile.Name())
    case "xsel":
        cmd = exec.Command("xsel", "--clipboard", "--input", "--file", tmpFile.Name())
    default:
        // Fallback to reading file and copying normally
        content, err := os.ReadFile(tmpFile.Name())
        if err != nil {
            return err
        }
        return m.Copy(string(content))
    }

    if err := cmd.Run(); err != nil {
        return ClipboardError{
            Platform: m.platform,
            Command:  cmd.String(),
            Err:      err,
        }
    }

    return nil
}
```

## Usage Pattern for Shotgun-CLI

```go
package main

import (
    "log"
    "github.com/rs/zerolog"
    "your-project/internal/clipboard"
)

func saveAndCopyContext(context string, outputPath string) error {
    // Save to file
    if err := saveContextToFile(context, outputPath); err != nil {
        return fmt.Errorf("failed to save context: %w", err)
    }

    log.Info().Str("file", outputPath).Msg("Context saved to file")

    // Copy to clipboard
    clipboardMgr := clipboard.NewManager()

    if !clipboardMgr.IsAvailable() {
        log.Warn().Msg("Clipboard not available, skipping copy operation")
        return nil
    }

    log.Info().Str("tool", clipboardMgr.GetStatus().Tools[0].Name).Msg("Copying to clipboard...")

    if err := clipboardMgr.CopyLarge(context); err != nil {
        log.Warn().Err(err).Msg("Failed to copy to clipboard, but file was saved successfully")
        return nil // Don't fail the entire operation if clipboard fails
    }

    log.Info().Msg("Context copied to clipboard successfully")
    return nil
}

func displayClipboardStatus(clipboardMgr *clipboard.Manager) {
    status := clipboardMgr.GetStatus()

    fmt.Printf("Clipboard Status:\n")
    fmt.Printf("  Platform: %s\n", status.Platform)
    fmt.Printf("  Available: %v\n", status.Available)
    fmt.Printf("  Tools:\n")

    for _, tool := range status.Tools {
        marker := " "
        if tool.Selected {
            marker = "*"
        }
        available := "✗"
        if tool.Available {
            available = "✓"
        }

        fmt.Printf("  %s %s %s (%s)\n", marker, available, tool.Name, tool.Command)
    }
}
```

## Error Handling Best Practices

```go
func handleClipboardErrors(err error) {
    if err == nil {
        return
    }

    var clipErr ClipboardError
    if errors.As(err, &clipErr) {
        switch clipErr.Platform {
        case "linux":
            log.Warn().Err(err).Msg("Clipboard failed. Install xclip, xsel, or wl-clipboard")
        case "darwin":
            log.Warn().Err(err).Msg("Clipboard failed. pbcopy not available")
        case "windows":
            log.Warn().Err(err).Msg("Clipboard failed. Neither clip.exe nor PowerShell available")
        default:
            log.Warn().Err(err).Msg("Clipboard operation failed")
        }
    } else {
        log.Error().Err(err).Msg("Unexpected clipboard error")
    }
}
```

## Key Implementation Gotchas

### WSL Environment Detection
- Check `/proc/version` for "microsoft" or "wsl"
- Use `clip.exe` (not `clip`) in WSL
- Handle path translation if needed

### Large Content Performance
- Set reasonable size limits (10MB)
- Use temporary files for very large content on Linux
- Consider streaming for clipboard operations

### Platform-Specific Quirks
- **Linux**: Wayland vs X11 requires different tools
- **Windows**: PowerShell needs content escaping
- **macOS**: pbcopy is simple but check availability

### Error Handling
- Graceful degradation when clipboard unavailable
- Timeout operations to prevent hanging
- Provide clear error messages for missing tools

This documentation provides all essential patterns for robust cross-platform clipboard integration in Go applications, with specific optimizations for the Shotgun-CLI use case.