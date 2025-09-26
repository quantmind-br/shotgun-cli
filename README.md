# shotgun-cli

A cross-platform CLI tool that generates LLM-optimized codebase contexts with both TUI wizard and headless CLI modes. The tool scans your codebase, applies intelligent filtering patterns, and generates structured text representations optimized for Large Language Model consumption.

## Features

- **Interactive TUI Wizard**: 5-step guided interface for easy context generation
- **Headless CLI Mode**: Command-line interface for automation and scripting
- **Smart File Filtering**: Gitignore-style pattern matching with include/exclude support
- **Size Management**: Configurable size limits with enforcement options
- **Multi-format Output**: Markdown output with optional clipboard integration
- **Cross-platform**: Works on Linux, macOS, and Windows
- **Template System**: Built-in prompt templates for different use cases
- **Diff Management**: Tools for splitting large diff files into manageable chunks

## Installation

### Option 1: Install from Source (Go)
```bash
go install github.com/quantmind-br/shotgun-cli@latest
```

### Option 2: Build and Install Locally
```bash
git clone https://github.com/quantmind-br/shotgun-cli
cd shotgun-cli

# Install to GOPATH/bin (user-local)
make install

# Or install system-wide (requires sudo)
make install-system
```

### Option 3: Manual Build
```bash
git clone https://github.com/quantmind-br/shotgun-cli
cd shotgun-cli
make build
# Binary will be in build/shotgun-cli
```

### Installation Options

- **`make install`**: Installs to `$GOPATH/bin` (default, user-local)
- **`make install-local`**: Same as above, explicitly local installation
- **`make install-system`**: Installs to `/usr/local/bin` (system-wide, requires sudo)
- **`make uninstall`**: Removes system-wide installation

### Verify Installation
```bash
shotgun-cli --version
```

## Usage

### Interactive Mode (TUI Wizard)

When called without arguments, shotgun-cli launches an interactive 5-step wizard:

```bash
shotgun-cli
```

The wizard guides you through:
1. **File Selection**: Choose which files to include in the context
2. **Template Selection**: Pick from built-in prompt templates
3. **Task Input**: Define your specific task or question
4. **Rules Input**: Add custom rules or constraints
5. **Review & Generate**: Preview and generate the final context

### Headless CLI Mode

#### Context Generation

Generate context from your codebase with full control:

```bash
# Basic usage
shotgun-cli context generate

# Advanced usage with options
shotgun-cli context generate \
  --root ./src \
  --include "*.go,*.js,*.ts" \
  --exclude "vendor/*,node_modules/*,*.test.go" \
  --output my-context.md \
  --max-size 5MB \
  --no-enforce-limit
```

**Context Generate Options:**
- `--root, -r`: Root directory to scan (default: current directory)
- `--include, -i`: File patterns to include (glob patterns, default: `["*"]`)
- `--exclude, -e`: File patterns to exclude (glob patterns)
- `--output, -o`: Output file (default: `shotgun-prompt-YYYYMMDD-HHMMSS.md`)
- `--max-size`: Maximum context size (default: `10MB`, formats: `1MB`, `5GB`, `500KB`)
- `--enforce-limit`: Fail if output exceeds max-size (default: true)
- `--no-enforce-limit`: Allow generation that exceeds size limit with warning

**Examples:**
```bash
# Generate context for a Go project
shotgun-cli context generate --include "*.go" --exclude "vendor/*,*.test.go"

# Generate with custom output and size limit
shotgun-cli context generate --output project-context.md --max-size 2MB

# Generate allowing size limit exceeded
shotgun-cli context generate --no-enforce-limit --max-size 1MB

# Generate for specific directory
shotgun-cli context generate --root ./backend --include "*.py,*.yaml"
```

#### Template Management

List and use built-in prompt templates:

```bash
# List available templates
shotgun-cli template list

# Render a template with variables
shotgun-cli template render prompt_makePlan --var PROJECT_NAME=myapp --var TASK="implement auth"

# Render to file
shotgun-cli template render prompt_analyzeBug --var BUG_DESCRIPTION="login fails" --output bug-analysis.md
```

**Template Render Options:**
- `--var`: Template variables as key=value pairs (can be used multiple times)
- `--output, -o`: Output file (default: stdout)

**Built-in Templates:**
- `prompt_analyzeBug`: For analyzing and debugging issues
- `prompt_makeDiffGitFormat`: For creating git-format diff files
- `prompt_makePlan`: For project planning and task breakdown
- `prompt_projectManager`: For project management tasks

#### Diff Management

Split large diff files into manageable chunks:

```bash
# Split a diff file
shotgun-cli diff split --input large.diff --output-dir chunks --approx-lines 500

# Split without headers
shotgun-cli diff split --input large.diff --output-dir chunks --approx-lines 300 --no-header
```

**Diff Split Options:**
- `--input, -i`: Input diff file (required)
- `--output-dir, -o`: Output directory for chunks (default: `chunks`)
- `--approx-lines`: Approximate lines per chunk (default: 500)
- `--no-header`: Don't add chunk headers to output files

#### Configuration Management

View and modify configuration settings:

```bash
# Show current configuration
shotgun-cli config show

# Set configuration values
shotgun-cli config set scanner.max-files 1000
shotgun-cli config set output.clipboard true
```

**Available Config Keys:**
- `scanner.max-files`: Maximum number of files to process
- `scanner.max-file-size`: Maximum size per file
- `output.clipboard`: Auto-copy to clipboard
- Logging levels and other settings

#### Shell Completion

Generate shell completion scripts:

```bash
# Bash
shotgun-cli completion bash > /etc/bash_completion.d/shotgun-cli

# Zsh
shotgun-cli completion zsh > "${fpath[1]}/_shotgun-cli"

# Fish
shotgun-cli completion fish > ~/.config/fish/completions/shotgun-cli.fish

# PowerShell
shotgun-cli completion powershell > shotgun-cli.ps1
```

### Global Options

These options work with all commands:
- `--config`: Config file path (default: `~/.config/shotgun-cli/config.yaml`)
- `--help, -h`: Show help for any command
- `--version`: Show version information

## Configuration

### Ignore Patterns

The tool uses **gitignore syntax** via the `github.com/sabhiram/go-gitignore` library for powerful file filtering.

**Examples of gitignore-style patterns:**

- `*.log` - Ignore all .log files
- `dir/` - Ignore directories named "dir"
- `!/keep.go` - Explicitly include files that would otherwise be ignored
- `**/vendor/` - Ignore vendor directories at any depth
- `build/` - Ignore build directory
- `*.tmp` - Ignore all temporary files
- `!important.tmp` - But keep this specific temporary file

**Advanced gitignore features:**

- **Directory matching**: Patterns ending with `/` only match directories
- **Negation**: Patterns starting with `!` negate (include) previously ignored files
- **Nested patterns**: Use `**/` for matching at any directory depth
- **Relative paths**: Patterns starting with `/` are anchored to the repository root

### Configuration File

The tool supports a YAML configuration file located at `~/.config/shotgun-cli/config.yaml`:

```yaml
scanner:
  max-files: 10000
  max-file-size: "1MB"

output:
  clipboard: true

logging:
  level: "info"
```

### Environment Variables

- `SHOTGUN_CONFIG`: Override config file path
- `SHOTGUN_LOG_LEVEL`: Set logging level (debug, info, warn, error)

## Architecture

### Core Components

- **Scanner**: File system scanning with intelligent filtering
- **Context Generator**: Creates optimized text representations
- **Template Engine**: Renders prompt templates with variables
- **TUI Wizard**: Interactive Bubble Tea-based interface
- **CLI Commands**: Cobra-based command structure

### Project Structure

```
shotgun-cli/
├── cmd/                    # CLI commands and configuration
│   ├── root.go            # Main command and TUI launcher
│   ├── context.go         # Context generation commands
│   ├── template.go        # Template management
│   ├── diff.go            # Diff splitting tools
│   └── config.go          # Configuration management
├── internal/
│   ├── core/              # Core business logic
│   │   ├── scanner/       # File system scanning
│   │   ├── context/       # Context generation
│   │   ├── template/      # Template management
│   │   └── ignore/        # Gitignore pattern matching
│   ├── ui/                # TUI components
│   │   ├── wizard.go      # Main wizard orchestration
│   │   ├── screens/       # Individual wizard steps
│   │   └── components/    # Reusable UI components
│   ├── platform/          # Platform-specific code
│   │   └── clipboard/     # Clipboard integration
│   └── utils/             # Utility functions
├── templates/             # Built-in prompt templates
└── assets/               # Embedded assets
```

## Development

### Prerequisites

- Go 1.24 or later
- Make (for build automation)

### Building

```bash
# Build for current platform
make build

# Cross-compile for all platforms
make build-all

# Install dependencies
make deps
```

### Testing

```bash
# Run unit tests
make test

# Run tests with race detection
make test-race

# Run benchmarks
make test-bench

# Run end-to-end tests
make test-e2e

# Generate coverage report
make coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run static analysis
make vet

# Run all quality checks
make fmt lint vet test
```

### Release

```bash
# Build release artifacts
make release
```

## Dependencies

### Core Libraries

- **CLI Framework**: [Cobra](https://github.com/spf13/cobra) for command structure
- **Configuration**: [Viper](https://github.com/spf13/viper) for config management
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) for interactive interface
- **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss) for TUI styling
- **Logging**: [Zerolog](https://github.com/rs/zerolog) for structured logging
- **Gitignore**: [go-gitignore](https://github.com/sabhiram/go-gitignore) for pattern matching

### Platform Support

- **Clipboard**: Cross-platform clipboard integration
- **File System**: Native file system operations
- **Terminal**: Advanced terminal features and detection

## Examples

### Basic Workflow

1. **Interactive Mode**:
   ```bash
   cd your-project
   shotgun-cli
   # Follow the 5-step wizard
   ```

2. **CLI Mode for Automation**:
   ```bash
   shotgun-cli context generate \
     --include "*.go,*.md" \
     --exclude "vendor/*,*.test.go" \
     --max-size 3MB \
     --output project-context.md
   ```

3. **Template-based Prompts**:
   ```bash
   shotgun-cli template render prompt_makePlan \
     --var PROJECT_NAME="my-api" \
     --var TASK="implement JWT authentication" \
     --output plan.md
   ```

### Integration Examples

#### CI/CD Pipeline
```yaml
- name: Generate codebase context
  run: |
    shotgun-cli context generate \
      --include "*.go,*.yaml,*.md" \
      --exclude "vendor/*,*.test.go" \
      --output codebase-context.md \
      --max-size 5MB
```

#### Pre-commit Hook
```bash
#!/bin/bash
shotgun-cli context generate \
  --include "*.go" \
  --max-size 2MB \
  --output .context/latest.md
```

## Migration from filepath.Match patterns

If you were previously using `filepath.Match`-style patterns, you may need to update them:

| Old Pattern (filepath.Match) | New Pattern (gitignore) | Notes |
|------------------------------|------------------------|--------|
| `*.log` | `*.log` | ✅ Same |
| `test*` | `test*` | ✅ Same |
| `[abc].txt` | `[abc].txt` | ✅ Same |
| Custom complex patterns | Check gitignore docs | May need adjustment |

For complete gitignore pattern documentation, see the [go-gitignore library documentation](https://github.com/sabhiram/go-gitignore).

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and quality checks: `make fmt lint vet test`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.