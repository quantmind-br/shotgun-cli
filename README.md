# Shotgun CLI

🎯 A powerful CLI tool for generating optimized prompts for AI interactions by combining task templates, user input, and project file context.

## Overview

Shotgun CLI streamlines the process of creating comprehensive prompts for AI assistants by:

- **Quick Task Selection**: Choose from predefined task types (architect, dev, find bug, docs-sync)
- **Structured Input**: Multi-line text input with automatic line numbering
- **Smart File Selection**: Interactive file tree with gitignore integration
- **Template Processing**: Automatic prompt generation using predefined templates
- **Offline Operation**: Works entirely locally without API keys or external services

## Features

### Task Types

- **architect**: Design system architecture and create plans (uses `prompt_makePlan.md`)
- **dev**: Generate code changes and implementations (uses `prompt_makeDiffGitFormat.md`)
- **find bug**: Analyze and debug code issues (uses `prompt_analyzeBug.md`)
- **docs-sync**: Synchronize documentation with code (uses `prompt_projectManager.md`)

### Core Functionality

- ✅ Interactive terminal UI (TUI) with Bubble Tea
- ✅ Task description input with line numbering
- ✅ Optional project rules configuration
- ✅ File tree navigation with selection
- ✅ Gitignore integration for smart exclusions
- ✅ Template-based prompt generation
- ✅ Timestamped output files

## Installation

### Prerequisites

- Go 1.23+ (for building from source)
- The prompt template files in your project directory

### Building from Source

```bash
git clone <repository-url>
cd shotgun-cli/shotgun_cli
go build -o shotgun-cli
```

### Usage

1. **Run the application** in your project directory:
   ```bash
   ./shotgun-cli
   ```

2. **Select task type** using arrow keys and Enter

3. **Enter task description**:
   - Type your multi-line task description
   - Lines are automatically numbered
   - Press `Alt+D` when finished

4. **Configure rules** (optional):
   - Choose Y/N for adding project-specific rules
   - If yes, enter multi-line rules and press `Alt+D`

5. **Select files**:
   - Navigate the file tree with ↑/↓ arrows
   - Press Space to select/deselect files
   - Press Enter to expand/collapse directories
   - Press `Alt+D` to continue

6. **Generate prompt**:
   - The application processes your inputs
   - Generates a timestamped markdown file
   - Shows the output file path

### Keyboard Shortcuts

- `↑/↓` - Navigate menus and file tree
- `Enter` - Select option or expand/collapse directory
- `Space` - Select/deselect files in file tree
- `Alt+D` - Confirm input and proceed to next step
- `Esc` - Go back to previous screen
- `Ctrl+C` or `q` - Quit application

## Template Files

The application requires the following template files in your project directory:

- `prompt_makePlan.md` - For architecture tasks
- `prompt_makeDiffGitFormat.md` - For development tasks
- `prompt_analyzeBug.md` - For debugging tasks
- `prompt_projectManager.md` - For documentation sync tasks

Each template should contain placeholders:
- `{TASK}` - Replaced with your task description
- `{RULES}` - Replaced with project rules
- `{FILE_STRUCTURE}` - Replaced with selected file contents

## File Exclusions

The application automatically excludes common files and directories:

- Git files (`.git`, `.gitignore` contents)
- Dependencies (`node_modules`, build artifacts)
- IDE files (`.vscode`, `.idea`)
- Temporary files (`*.log`, `*.tmp`, `*.cache`)
- Binary files (`*.exe`, `*.dll`, `*.so`)

## Output

Generated prompts are saved as:
```
shotgun_prompt_YYYYMMDD_HHMMSS.md
```

The output includes:
- Complete processed template
- Your task description with line numbers
- Project rules (if specified)
- File structure tree
- Content of all selected files

## Architecture

```
shotgun-cli/
├── main.go                 # Application entry point
├── internal/
│   ├── cmd/               # Command handling
│   ├── ui/                # Terminal user interface
│   │   ├── app.go         # Main application logic
│   │   └── filetree.go    # File tree component
│   ├── file/              # File system operations
│   │   └── scanner.go     # Directory scanning and gitignore
│   └── template/          # Template processing
│       └── processor.go   # Template loading and substitution
├── go.mod                 # Go module definition
└── README.md             # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

---

Built with ❤️ using [Bubble Tea](https://github.com/charmbracelet/bubbletea)