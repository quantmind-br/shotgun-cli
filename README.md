# shotgun-cli

Terminal-based prompt generation tool built with Go and BubbleTea. A command-line interface version of the Shotgun application for generating structured LLM prompts from codebase context.

## Features

- **Interactive Terminal UI**: Built with [BubbleTea](https://github.com/charmbracelet/bubbletea) for a smooth TUI experience
- **Inverse File Selection**: Instead of selecting files to include, select files to exclude (more intuitive for large projects)
- **Advanced File Filtering**: 
  - Automatic `.gitignore` support
  - Built-in ignore patterns for common file types
  - Custom ignore rules
- **Multiple Prompt Templates**: 
  - **Dev**: Generate git diffs for code changes
  - **Architect**: Create design plans and architecture
  - **Debug**: Bug analysis and debugging
  - **Project Manager**: Documentation sync and task management
- **Progress Tracking**: Real-time progress bars for large projects
- **Cross-Platform**: Works on Windows, macOS, and Linux

## Installation

### Global Installation via npm

For end users:
```bash
npm install -g shotgun-cli
```

### Local Development Installation

For development or local installation:
```bash
git clone https://github.com/your-username/shotgun-cli.git
cd shotgun-cli
npm install -g .
```

This will:
1. Build the Go binary automatically

## Development

### Build Commands

The project includes comprehensive npm scripts for building across platforms:

```bash
# Build for current platform
npm run build

# Build for all platforms (Windows, Linux, macOS, macOS ARM64)
npm run build:all

# Build for specific platforms
npm run build:windows
npm run build:linux
npm run build:macos

# Build locally (development)
npm run build:local

# Clean build artifacts
npm run clean

# Run tests
npm test

# Run in development mode
npm run dev

# Code quality
npm run lint
npm run format
```

### Build Outputs

The build process creates optimized binaries for each platform:

- `bin/shotgun-cli.exe` - Windows (x64)
- `bin/shotgun-cli-linux` - Linux (x64)
- `bin/shotgun-cli-macos` - macOS (x64)
- `bin/shotgun-cli-macos-arm64` - macOS (ARM64)

All binaries are optimized with `-ldflags="-s -w"` for smaller file sizes.

This will:
1. Build the Go binary automatically
2. Install the `shotgun-cli` command globally
3. Make it available from any directory

### Manual Build

To build manually:
```bash
go build -o bin/shotgun-cli .
```

## Usage

Simply run the command in any directory:

```bash
shotgun-cli
```

### Command Line Options

```bash
shotgun-cli --version    # Show version
shotgun-cli --help       # Show help
```

## Workflow

1. **File Exclusion**: Choose files to exclude from the prompt (inverse selection)
2. **Template Selection**: Select from 4 specialized prompt templates with full descriptions
3. **Task Description**: Describe your task in dedicated full-screen interface
4. **Custom Rules**: Add optional rules and constraints in dedicated full-screen interface
5. **Generate**: Creates and saves the prompt to current directory

## Keyboard Shortcuts

### General
- `Ctrl+Q`, `Ctrl+C`: Quit application
- `?`: Toggle help
- `o`: Access configuration/settings
- `Esc`: Go back to previous step

### Directory Selection
- `↑↓`: Navigate directories
- `Enter`: Select directory

### File Exclusion
- `hjkl` or `↑↓←→`: Navigate file tree
- `Space`: Toggle file/directory exclusion
- `c`: Continue to next step
- `r`: Reset all exclusions
- `a`: Exclude all files
- `A`: Include all files

### Template Selection
- `↑/↓` (or `k/j`): Navigate template options
- `1-4`: Quick select template by number
- `Enter`: Confirm selection and continue

### Task Description
- `Tab`: Focus/unfocus input field
- `F5`: Continue to Custom Rules

### Custom Rules
- `Tab`: Focus/unfocus input field
- `F5`: Generate prompt

## Template Types

### 1. Dev Template (`prompt_makeDiffGitFormat.md`)
- **Purpose**: Generate git diff formatted code changes
- **Output**: Standard unified git diff format
- **Use Case**: Code implementation requests

### 2. Architect Template (`prompt_makePlan.md`)
- **Purpose**: Create refactoring and design plans
- **Output**: Structured Markdown planning document
- **Use Case**: Architecture and planning tasks

### 3. Debug Template (`prompt_analyzeBug.md`)
- **Purpose**: Debug analysis and root cause identification
- **Output**: Comprehensive bug analysis report
- **Use Case**: Debugging and troubleshooting

### 4. Project Manager Template (`prompt_projectManager.md`)
- **Purpose**: Documentation synchronization and task management
- **Output**: Git diff for documentation updates
- **Use Case**: Project management and documentation

## File Filtering

The application uses a three-layer filtering system:

1. **Built-in Ignore Patterns**: Common files like `node_modules/`, `*.jpg`, `*.exe`, etc.
2. **Project .gitignore**: Automatically respects your project's `.gitignore` file
3. **Custom Rules**: User-defined ignore patterns

### Default Ignored File Types

- **Media**: `*.jpg`, `*.png`, `*.mp4`, etc.
- **Dependencies**: `node_modules/`, `vendor/`, `__pycache__/`
- **IDE Files**: `.idea/`, `.vscode/`, `.vs/`
- **Temporary**: `*.tmp`, `*.bak`, `*.swp`
- **Archives**: `*.zip`, `*.rar`, `*.tar`
- **Binaries**: `*.exe`, `*.dll`, `*.so`

## Configuration

Configuration is stored in:
- **Windows**: `%APPDATA%\shotgun-code\settings.json`
- **macOS**: `~/Library/Application Support/shotgun-code/settings.json`
- **Linux**: `~/.config/shotgun-code/settings.json`

## Output

Generated prompts are saved in the current directory with timestamped filenames:
```
shotgun_prompt_2025-01-20_143052.md
```

## Performance

- **Startup time**: <2 seconds
- **Directory scanning**: <5 seconds for 1000 files
- **Context generation**: Varies by project size
- **Memory usage**: Scales with project size
- **Size limit**: No limit (generate prompts of any size)

## Development

### Building

```bash
go build -o shotgun-cli .
```

### Testing

```bash
go test -v ./...
```

### Code Quality

```bash
go fmt ./...
go vet ./...
```

## Architecture

```
shotgun-cli/
├── main.go                   # Application entry point
├── internal/
│   ├── ui/                   # BubbleTea UI components
│   │   ├── app.go           # Main application model
│   │   ├── filetree.go      # File tree with exclusion
│   │   └── views.go         # View handlers
│   ├── core/                # Business logic
│   │   ├── scanner.go       # Directory scanning
│   │   ├── generator.go     # Context generation
│   │   ├── template.go      # Template processing
│   │   └── types.go         # Core data structures
│   └── config/              # Configuration
├── templates/               # Prompt templates
├── package.json            # NPM configuration
└── .goreleaser.yaml        # Release configuration
```

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Troubleshooting

### Common Issues

**Template not found**: Ensure templates are in the correct path relative to the executable.

**Permission denied**: Make sure you have read access to the project directory.

**Out of memory**: Large projects may exceed the 10MB context limit. Use file exclusion to reduce size.

**Binary not found**: After npm install, ensure your PATH includes npm global binaries.

### Getting Help

- Check `shotgun-cli --help` for usage information
- Press `?` in the application for keyboard shortcuts
- Report issues at: https://github.com/your-username/shotgun-cli/issues

## Acknowledgments

- Built with [BubbleTea](https://github.com/charmbracelet/bubbletea) TUI framework
- Inspired by the original Shotgun desktop application
- File filtering powered by [go-gitignore](https://github.com/sabhiram/go-gitignore)