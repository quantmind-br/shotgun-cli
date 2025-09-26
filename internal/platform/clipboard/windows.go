package clipboard

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type WindowsClipboard struct {
	preferredTool string
}

func NewWindowsClipboard() *WindowsClipboard {
	wc := &WindowsClipboard{}

	if _, err := exec.LookPath("clip"); err == nil {
		wc.preferredTool = "clip"
	} else if _, err := exec.LookPath("powershell"); err == nil {
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
		return &ClipboardError{
			Platform: "windows",
			Command:  "none",
			Err:      fmt.Errorf("no clipboard tools available"),
		}
	}
}

func (wc *WindowsClipboard) CopyWithTimeout(content string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	switch wc.preferredTool {
	case "clip":
		return wc.copyWithClipContext(ctx, content)
	case "powershell":
		return wc.copyWithPowerShellContext(ctx, content)
	default:
		return &ClipboardError{
			Platform: "windows",
			Command:  "none",
			Err:      fmt.Errorf("no clipboard tools available"),
		}
	}
}

func (wc *WindowsClipboard) copyWithClip(content string) error {
	cmd := exec.Command("clip")
	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		return &ClipboardError{
			Platform: "windows",
			Command:  "clip",
			Err:      err,
		}
	}

	return nil
}

func (wc *WindowsClipboard) copyWithClipContext(ctx context.Context, content string) error {
	cmd := exec.CommandContext(ctx, "clip")
	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ClipboardError{
				Platform: "windows",
				Command:  "clip",
				Err:      fmt.Errorf("clipboard operation timed out"),
			}
		}
		return &ClipboardError{
			Platform: "windows",
			Command:  "clip",
			Err:      err,
		}
	}

	return nil
}

func (wc *WindowsClipboard) copyWithPowerShell(content string) error {
	escapedContent := strings.ReplaceAll(content, "'", "''")
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Set-Clipboard -Value '%s'", escapedContent))

	if err := cmd.Run(); err != nil {
		return &ClipboardError{
			Platform: "windows",
			Command:  "powershell",
			Err:      err,
		}
	}

	return nil
}

func (wc *WindowsClipboard) copyWithPowerShellContext(ctx context.Context, content string) error {
	escapedContent := strings.ReplaceAll(content, "'", "''")
	cmd := exec.CommandContext(ctx, "powershell", "-Command", fmt.Sprintf("Set-Clipboard -Value '%s'", escapedContent))

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ClipboardError{
				Platform: "windows",
				Command:  "powershell",
				Err:      fmt.Errorf("clipboard operation timed out"),
			}
		}
		return &ClipboardError{
			Platform: "windows",
			Command:  "powershell",
			Err:      err,
		}
	}

	return nil
}

func (wc *WindowsClipboard) IsAvailable() bool {
	if wc.preferredTool == "clip" {
		_, err := exec.LookPath("clip")
		return err == nil
	}
	if wc.preferredTool == "powershell" {
		_, err := exec.LookPath("powershell")
		return err == nil
	}
	return false
}

func (wc *WindowsClipboard) GetCommand() (string, []string) {
	switch wc.preferredTool {
	case "clip":
		return "clip", []string{}
	case "powershell":
		return "powershell", []string{"-Command", "Set-Clipboard"}
	default:
		return "", nil
	}
}

func (wc *WindowsClipboard) GetPlatform() string {
	return "windows"
}