package cmd

import (
	"fmt"

	"shotgun-cli/internal/ui"
)

func Execute() error {
	fmt.Println("🎯 Shotgun CLI - AI Prompt Generator")
	fmt.Println("=====================================")
	
	app := ui.NewApp()
	if err := app.Run(); err != nil {
		return fmt.Errorf("failed to run application: %w", err)
	}
	
	return nil
}