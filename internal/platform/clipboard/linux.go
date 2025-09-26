package clipboard

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type LinuxClipboard struct {
	tools        []ClipboardTool
	selectedTool *ClipboardTool
}

func NewLinuxClipboard() *LinuxClipboard {
	lc := &LinuxClipboard{
		tools: []ClipboardTool{
			{Name: "wl-copy", Command: "wl-copy", Args: []string{}, Priority: 1},
			{Name: "xclip", Command: "xclip", Args: []string{"-selection", "clipboard"}, Priority: 2},
			{Name: "xsel", Command: "xsel", Args: []string{"--clipboard", "--input"}, Priority: 3},
		},
	}

	lc.detectAvailableTools()
	return lc
}

func (lc *LinuxClipboard) detectAvailableTools() {
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	xDisplay := os.Getenv("DISPLAY")

	for i := range lc.tools {
		tool := &lc.tools[i]
		tool.Available = lc.checkTool(tool.Command)

		if waylandDisplay != "" && tool.Name == "wl-copy" && tool.Available {
			lc.selectedTool = tool
			return
		}
	}

	if xDisplay != "" {
		for i := range lc.tools {
			tool := &lc.tools[i]
			if (tool.Name == "xclip" || tool.Name == "xsel") && tool.Available {
				if lc.selectedTool == nil || tool.Priority < lc.selectedTool.Priority {
					lc.selectedTool = tool
				}
			}
		}
	}

	if lc.selectedTool == nil {
		for i := range lc.tools {
			tool := &lc.tools[i]
			if tool.Available {
				if lc.selectedTool == nil || tool.Priority < lc.selectedTool.Priority {
					lc.selectedTool = tool
				}
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
			Platform: "linux",
			Command:  "none",
			Err:      fmt.Errorf("no clipboard tools available"),
		}
	}
	return lc.copyWithTool(content)
}

func (lc *LinuxClipboard) CopyWithTimeout(content string, timeout time.Duration) error {
	if lc.selectedTool == nil {
		return &ClipboardError{
			Platform: "linux",
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
			Platform: "linux",
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
				Platform: "linux",
				Command:  lc.selectedTool.Command,
				Err:      fmt.Errorf("clipboard operation timed out"),
			}
		}
		return &ClipboardError{
			Platform: "linux",
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
	return "linux"
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