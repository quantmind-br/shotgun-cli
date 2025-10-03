package clipboard

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	toolClip       = "clip"
	toolPowerShell = "powershell"
)

type WindowsClipboard struct {
	preferredTool string
}

func NewWindowsClipboard() *WindowsClipboard {
	wc := &WindowsClipboard{}

	if _, err := exec.LookPath(toolClip); err == nil {
		wc.preferredTool = toolClip
	} else if _, err := exec.LookPath(toolPowerShell); err == nil {
		wc.preferredTool = toolPowerShell
	}

	return wc
}

func (wc *WindowsClipboard) Copy(content string) error {
	switch wc.preferredTool {
	case toolClip:
		return wc.copyWithClip(content)
	case toolPowerShell:
		return wc.copyWithPowerShell(content)
	default:
		return &ClipboardError{
			Platform: platformWindows,
			Command:  "none",
			Err:      fmt.Errorf("no clipboard tools available"),
		}
	}
}

func (wc *WindowsClipboard) CopyWithTimeout(content string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	switch wc.preferredTool {
	case toolClip:
		return wc.copyWithClipContext(ctx, content)
	case toolPowerShell:
		return wc.copyWithPowerShellContext(ctx, content)
	default:
		return &ClipboardError{
			Platform: platformWindows,
			Command:  "none",
			Err:      fmt.Errorf("no clipboard tools available"),
		}
	}
}

func (wc *WindowsClipboard) copyWithClip(content string) error {
	cmd := exec.Command(toolClip)
	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		return &ClipboardError{
			Platform: platformWindows,
			Command:  toolClip,
			Err:      err,
		}
	}

	return nil
}

func (wc *WindowsClipboard) copyWithClipContext(ctx context.Context, content string) error {
	cmd := exec.CommandContext(ctx, toolClip)
	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ClipboardError{
				Platform: platformWindows,
				Command:  toolClip,
				Err:      fmt.Errorf("clipboard operation timed out"),
			}
		}
		return &ClipboardError{
			Platform: platformWindows,
			Command:  toolClip,
			Err:      err,
		}
	}

	return nil
}

func (wc *WindowsClipboard) copyWithPowerShell(content string) error {
	cmd := exec.Command(toolPowerShell, "-NoProfile", "-Command", "Set-Clipboard -Value ([Console]::In.ReadToEnd())")
	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		return &ClipboardError{
			Platform: platformWindows,
			Command:  toolPowerShell,
			Err:      err,
		}
	}

	return nil
}

func (wc *WindowsClipboard) copyWithPowerShellContext(ctx context.Context, content string) error {
	cmd := exec.CommandContext(
		ctx, toolPowerShell, "-NoProfile", "-Command",
		"Set-Clipboard -Value ([Console]::In.ReadToEnd())",
	)
	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ClipboardError{
				Platform: platformWindows,
				Command:  toolPowerShell,
				Err:      fmt.Errorf("clipboard operation timed out"),
			}
		}
		return &ClipboardError{
			Platform: platformWindows,
			Command:  toolPowerShell,
			Err:      err,
		}
	}

	return nil
}

func (wc *WindowsClipboard) IsAvailable() bool {
	if wc.preferredTool == toolClip {
		_, err := exec.LookPath(toolClip)
		return err == nil
	}
	if wc.preferredTool == toolPowerShell {
		_, err := exec.LookPath(toolPowerShell)
		return err == nil
	}
	return false
}

func (wc *WindowsClipboard) GetCommand() (string, []string) {
	switch wc.preferredTool {
	case toolClip:
		return toolClip, []string{}
	case toolPowerShell:
		return toolPowerShell, []string{"-Command", "Set-Clipboard"}
	default:
		return "", nil
	}
}

func (wc *WindowsClipboard) GetPlatform() string {
	return platformWindows
}

func (wc *WindowsClipboard) SetSelectedTool(name string) error {
	switch name {
	case toolClip:
		if _, err := exec.LookPath(toolClip); err != nil {
			return &ClipboardError{
				Platform: platformWindows,
				Command:  name,
				Err:      fmt.Errorf("tool %s is not available", name),
			}
		}
		wc.preferredTool = name
		return nil
	case toolPowerShell:
		if _, err := exec.LookPath(toolPowerShell); err != nil {
			return &ClipboardError{
				Platform: platformWindows,
				Command:  name,
				Err:      fmt.Errorf("tool %s is not available", name),
			}
		}
		wc.preferredTool = name
		return nil
	default:
		return &ClipboardError{
			Platform: platformWindows,
			Command:  name,
			Err:      fmt.Errorf("unknown tool %s, supported tools: clip, powershell", name),
		}
	}
}
