# Request Flow Analysis

The `shotgun-cli` application is a CLI tool designed to generate LLM-optimized codebase contexts. It operates in two primary modes: an interactive TUI (Terminal User Interface) wizard and a headless CLI mode. The request flow follows a command-driven architecture rather than a traditional HTTP server-request model.

## API Endpoints

As a CLI application, "endpoints" correspond to subcommands provided via the `cobra` library:

| Command | Subcommand | Method | Description |
|---------|------------|--------|-------------|
| `shotgun-cli` | (none) | CLI | Launches the interactive TUI Wizard (5 steps) |
| `shotgun-cli` | `context` | CLI | Base command for context operations |
| `shotgun-cli` | `context send [file]`| CLI | Sends a context file/stdin to Google Gemini |
| `shotgun-cli` | `diff` | CLI | Generates a diff-based context |
| `shotgun-cli` | `template` | CLI | Manages context templates (list, show, create) |
| `shotgun-cli` | `gemini` | CLI | Gemini integration utilities (doctor, models) |
| `shotgun-cli` | `config` | CLI | Manages application configuration |

## Request Processing Pipeline

The "pipeline" for a request (command execution) involves several stages:

1.  **Initialization (`cobra.OnInitialize`):** Loads configuration via `viper` from various locations (`/etc`, `$HOME/.config`, local `.`).
2.  **Validation (`PreRunE`):** Validates arguments and flags (e.g., checking if a file exists before sending to Gemini).
3.  **Command Execution (`Run` / `RunE`):** The primary logic for the command is executed.
4.  **Interactive Wizard (Bubble Tea):** If the TUI is launched, the "request" becomes an event loop managed by `bubbletea`:
    *   **Messages (`tea.Msg`):** Keyboard input, window resizing, or internal background process completion (e.g., `ScanCompleteMsg`).
    *   **Commands (`tea.Cmd`):** Background I/O operations like scanning the filesystem or calling the Gemini API.

## Routing Logic

Routing is handled by the `cobra` command hierarchy defined in the `cmd/` directory:

*   **Entry Point:** `main.go` calls `cmd.Execute()`.
*   **Command Matching:** `cobra` parses `os.Args` and routes the request to the specific `Run` or `RunE` function of the matched command.
*   **TUI Routing:** Inside the wizard, the "route" is maintained by a state variable (`m.step`) within the `WizardModel`. Navigation between screens (File Selection -> Template Selection -> Task Input, etc.) is handled by updating this state based on user input.

## Response Generation

Responses are generated and returned to the user in several ways:

1.  **Console Output:** Most commands use `fmt.Printf` or `fmt.Println` to output results directly to `stdout`.
2.  **File Generation:** The `context` and `diff` commands generate markdown files containing the codebase context.
3.  **Clipboard:** The TUI wizard can copy the generated context directly to the system clipboard using the `internal/platform/clipboard` package.
4.  **External Process:** The `context send` command pipes content to the `geminiweb` binary and captures its `stdout` as the final response.

## Error Handling

Error handling is implemented at multiple levels:

*   **CLI Level:** `cobra` catches errors returned from `RunE` and prints them to `stderr` before exiting with a non-zero code.
*   **Internal Logic:** Functions return standard Go `error` objects which are wrapped with context (e.g., `fmt.Errorf("failed to read file: %w", err)`).
*   **TUI Wizard:** Errors are captured in the `WizardModel.error` field. When an error occurs during background tasks (like `ScanErrorMsg` or `GenerationErrorMsg`), the TUI displays a specific error view to the user.
*   **Logging:** The `zerolog` package is used throughout the application to log debug and error information to `stderr`, which can be controlled by `--verbose` or `--quiet` flags.