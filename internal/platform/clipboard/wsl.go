package clipboard

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type WSLClipboard struct {
	isWSL bool
}

func NewWSLClipboard() *WSLClipboard {
	wsl := &WSLClipboard{}
	wsl.isWSL = wsl.detectWSL()
	return wsl
}

func (wsl *WSLClipboard) detectWSL() bool {
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}

	content, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}

	versionStr := strings.ToLower(string(content))
	return strings.Contains(versionStr, "microsoft") || strings.Contains(versionStr, "wsl")
}

func (wsl *WSLClipboard) Copy(content string) error {
	if !wsl.isWSL {
		return &ClipboardError{
			Platform: "wsl",
			Command:  "none",
			Err:      fmt.Errorf("not running in WSL environment"),
		}
	}

	cmd := exec.Command("clip.exe")
	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		return &ClipboardError{
			Platform: "wsl",
			Command:  "clip.exe",
			Err:      err,
		}
	}

	return nil
}

func (wsl *WSLClipboard) CopyWithTimeout(content string, timeout time.Duration) error {
	if !wsl.isWSL {
		return &ClipboardError{
			Platform: "wsl",
			Command:  "none",
			Err:      fmt.Errorf("not running in WSL environment"),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "clip.exe")
	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ClipboardError{
				Platform: "wsl",
				Command:  "clip.exe",
				Err:      fmt.Errorf("clipboard operation timed out"),
			}
		}
		return &ClipboardError{
			Platform: "wsl",
			Command:  "clip.exe",
			Err:      err,
		}
	}

	return nil
}

func (wsl *WSLClipboard) IsAvailable() bool {
	if !wsl.isWSL {
		return false
	}

	_, err := exec.LookPath("clip.exe")
	return err == nil
}

func (wsl *WSLClipboard) GetCommand() (string, []string) {
	if wsl.isWSL {
		return "clip.exe", []string{}
	}
	return "", nil
}

func (wsl *WSLClipboard) GetPlatform() string {
	return "wsl"
}

func (wsl *WSLClipboard) SetSelectedTool(name string) error {
	if name != "clip.exe" {
		return &ClipboardError{
			Platform: "wsl",
			Command:  name,
			Err:      fmt.Errorf("unknown tool %s, only clip.exe is supported in WSL", name),
		}
	}
	return nil
}