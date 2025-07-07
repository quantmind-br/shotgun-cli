package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"shotgun-cli/internal/ui"
)

func Execute() error {
	fmt.Println("🎯 Shotgun CLI - AI Prompt Generator")
	fmt.Println("=====================================")
	
	// Informação adicional para debug no Windows
	if runtime.GOOS == "windows" {
		fmt.Printf("Platform: Windows\n")
		fmt.Printf("Starting TUI interface...\n\n")
	}
	
	app := ui.NewApp()
	if err := app.Run(); err != nil {
		// Se o TUI falhar, mostrar informação útil
		if strings.Contains(err.Error(), "TTY") || strings.Contains(err.Error(), "tty") {
			fmt.Printf("\n❌ Terminal interface error: %v\n", err)
			fmt.Println("\n🔧 Windows Terminal Tips:")
			fmt.Println("1. Try running from Windows Terminal (not Command Prompt)")
			fmt.Println("2. Try running from PowerShell")
			fmt.Println("3. Make sure your terminal supports ANSI colors")
			fmt.Println("4. Try: set TERM=xterm-256color && shotgun")
			return fmt.Errorf("terminal compatibility issue - see tips above")
		}
		return fmt.Errorf("failed to run application: %w", err)
	}
	
	return nil
}