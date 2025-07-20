package main

import (
	"fmt"
	"os"
	"runtime"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"shotgun-cli/internal/ui"
)

func main() {
	// Ensure UTF-8 support on Windows
	if runtime.GOOS == "windows" {
		// Set environment variables for UTF-8 support
		os.Setenv("LANG", "en_US.UTF-8")
		os.Setenv("LC_ALL", "en_US.UTF-8")

		// Try to set UTF-8 code page for better Unicode support
		if err := enableUTF8Windows(); err != nil {
			// If system call fails, try alternative approach
			fmt.Printf("Warning: Could not set UTF-8 mode: %v\n", err)
		}
	}

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
    Ctrl+Q, Ctrl+C    Quit
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

	// Configure BubbleTea with explicit input handling for UTF-8
	opts := []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithInput(os.Stdin),
	}

	// On Windows, add additional options for UTF-8 support
	if runtime.GOOS == "windows" {
		// Force UTF-8 mode
		opts = append(opts, tea.WithoutSignalHandler())
	}

	p := tea.NewProgram(m, opts...)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}

// enableUTF8Windows attempts to enable UTF-8 support on Windows
func enableUTF8Windows() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleCP := kernel32.NewProc("SetConsoleCP")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	getStdHandle := kernel32.NewProc("GetStdHandle")

	// UTF-8 code page is 65001
	const CP_UTF8 = 65001
	const STD_INPUT_HANDLE = ^uintptr(10 - 1) // -10 in two's complement
	const STD_OUTPUT_HANDLE = ^uintptr(7 - 1) // -7 in two's complement
	const ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
	const ENABLE_VIRTUAL_TERMINAL_INPUT = 0x0200

	// Set input code page to UTF-8
	ret, _, _ := setConsoleCP.Call(CP_UTF8)
	if ret == 0 {
		return fmt.Errorf("failed to set console input code page")
	}

	// Set output code page to UTF-8
	ret, _, _ = setConsoleOutputCP.Call(CP_UTF8)
	if ret == 0 {
		return fmt.Errorf("failed to set console output code page")
	}

	// Enable virtual terminal processing for better Unicode support
	hStdOut, _, _ := getStdHandle.Call(STD_OUTPUT_HANDLE)
	if hStdOut != 0 {
		setConsoleMode.Call(hStdOut, ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	}

	hStdIn, _, _ := getStdHandle.Call(STD_INPUT_HANDLE)
	if hStdIn != 0 {
		setConsoleMode.Call(hStdIn, ENABLE_VIRTUAL_TERMINAL_INPUT)
	}

	return nil
}
