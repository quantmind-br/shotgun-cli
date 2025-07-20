package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"shotgun-cli/internal/ui"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("shotgun-cli v1.0.0")
		return
	}

	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Print(`shotgun-cli - Terminal-based prompt generation tool

USAGE:
    shotgun-cli [OPTIONS]

OPTIONS:
    --version    Show version information
    --help, -h   Show this help message

DESCRIPTION:
    Interactive terminal tool to generate LLM prompts from codebase context.
    Uses inverse file selection (exclude files rather than include).
    
WORKFLOW:
    1. Choose files to exclude from prompt (uses current directory)
    2. Select prompt template and customize
    3. Generate and save prompt to current directory

KEYBOARD SHORTCUTS:
    hjkl         Navigation (vim-style)
    Space        Toggle file exclusion
    Enter        Confirm selection
    q, Ctrl+C    Quit
    ?            Help
`)
		return
	}

	// Validate current directory before starting UI
	if err := ui.ValidateCurrentDirectory(); err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("\nPlease run shotgun-cli from a valid directory with appropriate permissions.\n")
		os.Exit(1)
	}

	// Initialize BubbleTea application
	m, err := ui.NewModel()
	if err != nil {
		fmt.Printf("Error initializing application: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
