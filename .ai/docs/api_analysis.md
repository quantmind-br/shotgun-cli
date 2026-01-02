# API Analysis

## Project Type
This is a Go-based CLI application and library designed to generate LLM-optimized codebase contexts. It features an interactive TUI (Terminal User Interface) wizard and a headless CLI mode. It also includes integration with Google Gemini via `geminiweb`.

## Endpoints Overview
No HTTP endpoints - this is a CLI application and Go library.

## Authentication
- **Gemini Integration**: Authentication is handled externally by the `geminiweb` tool. Users must run `geminiweb auto-login` to authenticate with Google.
- **CLI Configuration**: Configuration is stored locally in a YAML file (defaulting to `~/.config/shotgun-cli/config.yaml` or platform-specific equivalent).

## CLI API
The application provides a comprehensive set of commands for context generation, template management, and Gemini interaction.

### context generate
Generates a structured text representation of a codebase.
- **Flags**:
    - `-r, --root <path>`: Root directory to scan (default: `.`)
    - `-i, --include <patterns>`: File patterns to include (glob-style, default: `*`)
    - `-e, --exclude <patterns>`: File patterns to exclude
    - `-o, --output <file>`: Output context file
    - `--max-size <size>`: Maximum context size (e.g., `10MB`, `500KB`)
    - `--enforce-limit`: Fail if output exceeds max-size (default: `true`)
    - `--send-gemini`: Automatically send generated context to Gemini
    - `-t, --template <name>`: Template name to use for formatting
    - `--task <text>`: Task description for the LLM
    - `--rules <text>`: Rules/constraints for the LLM
    - `-V, --var <key=value>`: Custom template variables
    - `--progress <mode>`: Progress reporting mode (`none`, `human`, `json`)

### context send [file]
Sends an existing context file or stdin to Google Gemini.
- **Flags**:
    - `-o, --output <file>`: File to save Gemini response
    - `-m, --model <name>`: Gemini model to use
    - `--timeout <seconds>`: Request timeout
    - `--raw`: Output raw response without processing

### template list
Lists all available embedded and custom templates with their descriptions.

### template render [name]
Renders a specific template with variable substitution.
- **Flags**:
    - `--var <key=value>`: Variables for template substitution
    - `-o, --output <file>`: Output file for rendered content

### template import <file> [name]
Imports a template file into the user's template directory.

### template export <name> <file>
Exports an existing template to a file.

### config show
Displays current configuration values, their sources (default, config file, environment), and Gemini status.

### config set <key> <value>
Updates a configuration value in the local config file.
- **Common Keys**: `scanner.max-files`, `scanner.max-file-size`, `context.max-size`, `gemini.enabled`, `gemini.model`.

### diff split
Splits a large diff file into smaller, manageable chunks while preserving context.
- **Flags**:
    - `-i, --input <file>`: Input diff file (required)
    - `-o, --output-dir <dir>`: Directory for chunks
    - `--approx-lines <n>`: Approximate lines per chunk (default: 500)
    - `--no-header`: Omit metadata headers in chunks

### gemini status | doctor
- `status`: Shows current Gemini integration readiness.
- `doctor`: Diagnoses configuration issues and suggests fixes.

## Programmatic API (Go Library)
The core logic is exposed through internal packages that can be used programmatically within Go.

### Package `scanner`
Provides file system scanning with ignore rule support.
- `NewFileSystemScanner() Scanner`: Creates a new scanner.
- `Scanner.Scan(root string, config *ScanConfig) (*FileNode, error)`: Scans a directory.
- `FileNode`: Represents the generated directory tree structure.

### Package `context`
Handles the generation of LLM context from a file tree.
- `NewDefaultContextGenerator() *DefaultContextGenerator`: Creates a generator.
- `ContextGenerator.Generate(root *scanner.FileNode, selections map[string]bool, config GenerateConfig) (string, error)`: Produces the final context string.

### Package `template`
Manages prompt templates and rendering.
- `NewManager() (*Manager, error)`: Initializes the template manager (handling embedded and local files).
- `Manager.RenderTemplate(name string, variables map[string]string) (string, error)`: Renders a named template.

### Package `gemini` (platform)
Handles interaction with the `geminiweb` binary.
- `NewExecutor(cfg Config) *Executor`: Creates a Gemini request executor.
- `Executor.Send(ctx context.Context, prompt string) (*Result, error)`: Sends a prompt and returns the response and duration.

## Common Patterns
- **Standardized Size Formatting**: All size parameters accept human-readable strings like `1MB`, `500KB`, `2GB`.
- **Ignore Semantics**: The tool respects `.gitignore` and `.shotgunignore` using standard Git ignore semantics.
- **Configuration Precedence**: Order of precedence is Flags > Environment Variables > Config File > Defaults. Environment variables are prefixed with `SHOTGUN_` (e.g., `SHOTGUN_GEMINI_MODEL`).