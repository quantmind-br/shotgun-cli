package clipboard

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	toolXclip   = "xclip"
	toolXsel    = "xsel"
	toolWlCopy  = "wl-copy"
)

type LinuxClipboard struct {
	tools        []ClipboardTool
	selectedTool *ClipboardTool
}

func NewLinuxClipboard() *LinuxClipboard {
	lc := &LinuxClipboard{
		tools: []ClipboardTool{
			{Name: toolWlCopy, Command: toolWlCopy, Args: []string{}, Priority: 1},
			{Name: toolXclip, Command: toolXclip, Args: []string{"-selection", "clipboard"}, Priority: 2},
			{Name: toolXsel, Command: toolXsel, Args: []string{"--clipboard", "--input"}, Priority: 3},
		},
	}

	lc.detectAvailableTools()
	return lc
}

func (lc *LinuxClipboard) detectAvailableTools() {
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	xDisplay := os.Getenv("DISPLAY")

	// Check tool availability
	for i := range lc.tools {
		lc.tools[i].Available = lc.checkTool(lc.tools[i].Command)
	}

	// Try Wayland first if available
	if waylandDisplay != "" && lc.selectWaylandTool() {
		return
	}

	// Try X11 tools if display is set
	if xDisplay != "" && lc.selectX11Tool() {
		return
	}

	// Fallback to any available tool
	lc.selectAnyAvailableTool()
}

func (lc *LinuxClipboard) selectWaylandTool() bool {
	for i := range lc.tools {
		tool := &lc.tools[i]
		if tool.Name == toolWlCopy && tool.Available {
			lc.selectedTool = tool
			return true
		}
	}
	return false
}

func (lc *LinuxClipboard) selectX11Tool() bool {
	for i := range lc.tools {
		tool := &lc.tools[i]
		if (tool.Name == toolXclip || tool.Name == toolXsel) && tool.Available {
			if lc.selectedTool == nil || tool.Priority < lc.selectedTool.Priority {
				lc.selectedTool = tool
			}
		}
	}
	return lc.selectedTool != nil
}

func (lc *LinuxClipboard) selectAnyAvailableTool() {
	for i := range lc.tools {
		tool := &lc.tools[i]
		if tool.Available {
			if lc.selectedTool == nil || tool.Priority < lc.selectedTool.Priority {
				lc.selectedTool = tool
			}
		}
	}
}

func (lc *LinuxClipboard) checkTool(toolName string) bool {
	_, err := exec.LookPath(toolName)
	return err == nil
}

func (lc *LinuxClipboard) Copy(content string) error {
	if lc.selectedTool == nil {
		return &ClipboardError{
			Platform: platformLinux,
			Command:  "none",
			Err:      fmt.Errorf("no clipboard tools available"),
		}
	}
	return lc.copyWithTool(content)
}

func (lc *LinuxClipboard) CopyWithTimeout(content string, timeout time.Duration) error {
	if lc.selectedTool == nil {
		return &ClipboardError{
			Platform: platformLinux,
			Command:  "none",
			Err:      fmt.Errorf("no clipboard tools available"),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return lc.copyWithContext(ctx, content)
}

func (lc *LinuxClipboard) copyWithTool(content string) error {
	cmd := exec.Command(lc.selectedTool.Command, lc.selectedTool.Args...)
	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		return &ClipboardError{
			Platform: platformLinux,
			Command:  lc.selectedTool.Command,
			Err:      err,
		}
	}

	return nil
}

func (lc *LinuxClipboard) copyWithContext(ctx context.Context, content string) error {
	cmd := exec.CommandContext(ctx, lc.selectedTool.Command, lc.selectedTool.Args...)
	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ClipboardError{
				Platform: platformLinux,
				Command:  lc.selectedTool.Command,
				Err:      fmt.Errorf("clipboard operation timed out"),
			}
		}
		return &ClipboardError{
			Platform: platformLinux,
			Command:  lc.selectedTool.Command,
			Err:      err,
		}
	}

	return nil
}

func (lc *LinuxClipboard) IsAvailable() bool {
	return lc.selectedTool != nil && lc.selectedTool.Available
}

func (lc *LinuxClipboard) GetCommand() (string, []string) {
	if lc.selectedTool != nil {
		return lc.selectedTool.Command, lc.selectedTool.Args
	}
	return "", nil
}

func (lc *LinuxClipboard) GetPlatform() string {
	return platformLinux
}

func (lc *LinuxClipboard) SetSelectedTool(name string) error {
	for i := range lc.tools {
		tool := &lc.tools[i]
		if tool.Name == name {
			if !tool.Available {
				return &ClipboardError{
					Platform: "linux",
					Command:  name,
					Err:      fmt.Errorf("tool %s is not available", name),
				}
			}
			lc.selectedTool = tool
			return nil
		}
	}

	return &ClipboardError{
		Platform: "linux",
		Command:  name,
		Err:      fmt.Errorf("unknown tool %s", name),
	}
}