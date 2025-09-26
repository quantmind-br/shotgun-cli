package clipboard

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type DarwinClipboard struct{}

func NewDarwinClipboard() *DarwinClipboard {
	return &DarwinClipboard{}
}

func (dc *DarwinClipboard) Copy(content string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		return &ClipboardError{
			Platform: "darwin",
			Command:  "pbcopy",
			Err:      err,
		}
	}

	return nil
}

func (dc *DarwinClipboard) CopyWithTimeout(content string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pbcopy")
	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ClipboardError{
				Platform: "darwin",
				Command:  "pbcopy",
				Err:      fmt.Errorf("clipboard operation timed out"),
			}
		}
		return &ClipboardError{
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

func (dc *DarwinClipboard) GetPlatform() string {
	return "darwin"
}

func (dc *DarwinClipboard) SetSelectedTool(name string) error {
	if name != "pbcopy" {
		return &ClipboardError{
			Platform: "darwin",
			Command:  name,
			Err:      fmt.Errorf("unknown tool %s, only pbcopy is supported", name),
		}
	}
	return nil
}