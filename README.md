# Shotgun CLI

🎯 A powerful CLI tool for generating optimized prompts for AI interactions by combining task templates, user input, and project file context.

## Overview

Shotgun CLI streamlines the process of creating comprehensive prompts for AI assistants by:

- **Quick Task Selection**: Choose from predefined task types (architect, dev, find bug, docs-sync)
- **Structured Input**: Multi-line text input with automatic line numbering
- **Smart File Selection**: Interactive file tree with gitignore integration
- **Template Processing**: Automatic prompt generation using predefined templates
- **Offline Operation**: Works entirely locally without API keys or external services
- **Cross-Platform**: Optimized for Windows, Linux, and WSL.

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

- Go 1.23+ (for building from source if not installing via npm pre-compiled binary)
- Node.js and npm (required for build scripts and the installation process via `npm install`)

### Recommended Installation (npm)

The easiest way to install `shotgun-cli` is via npm. This method will attempt to download a pre-compiled binary for your platform (Windows, Linux, macOS). If a binary is not available for your specific platform/architecture combination, it will fall back to building from source automatically.

```bash
# If published to npm registry (replace 'shotgun-cli' with actual published name if different)
npm install -g shotgun-cli

# For installing directly from a cloned repository:
git clone https://github.com/your-username/shotgun-cli.git # Replace with actual URL
cd shotgun-cli
npm install -g .
# This command runs 'node install.js || npm run build:platform'
# 'install.js' tries to download a binary. If it fails, 'build:platform' compiles it.
```

After installation, the `shotgun` or `shotgun-cli` command should be available in your PATH.

### Building from Source (Alternative)

If you prefer to build directly from source without using the npm installation flow:

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-username/shotgun-cli.git # Replace with actual URL
    cd shotgun-cli
    ```

2.  **Build for your platform:**

    *   **For Windows (native PowerShell or Windows Terminal recommended):**
        ```bash
        npm run build:windows
        # This creates bin/shotgun.exe (the Go binary) and bin/shotgun.js (the Node.js wrapper).
        # To make it globally available like 'npm install -g .', run:
        build.bat
        # (This script runs 'npm run build:windows' and then 'npm install -g .' using the local package.)
        ```

    *   **For Linux or macOS (or WSL):**
        ```bash
        npm run build:unix
        # This creates bin/shotgun (the Go binary) and bin/shotgun.js (the Node.js wrapper).
        # To make it globally available like 'npm install -g .', run:
        ./build.sh
        # (This script runs 'npm run build:unix' and then 'npm install -g .' using the local package.)
        ```
    *   **Platform-agnostic build (let npm scripts decide):**
        ```bash
        npm run build
        # This is an alias for 'npm run build:platform', which detects your OS
        # and runs either 'build:windows' or 'build:unix'.
        ```
    *   **Direct Go build (manual, for advanced users):**
        This method bypasses the npm scripts that also create the Node.js wrapper.
        ```bash
        # For Windows:
        go build -ldflags="-s -w" -o bin/shotgun.exe main.go
        # For Linux/macOS/WSL:
        go build -ldflags="-s -w" -o bin/shotgun main.go

        # After a direct Go build, the Node.js wrapper (bin/shotgun.js) will NOT be created automatically.
        # If you need the wrapper (e.g., for the 'shotgun' command in package.json's "bin"), run:
        node create-wrapper.js
        ```

**Notes for Specific Environments:**

*   **Windows Users:**
    *   Using `build.bat` after cloning is a convenient way to build and perform a global-like installation.
    *   For the best Terminal User Interface (TUI) experience, use **Windows Terminal** or **PowerShell**. Avoid the legacy `cmd.exe` if possible.
    *   If you experience TTY (teletypewriter) errors or display issues, try setting the `TERM` environment variable before running the application: `set TERM=xterm-256color`.
*   **WSL (Windows Subsystem for Linux) Users:**
    *   You can build and run the Linux version of `shotgun-cli` directly within your WSL distribution. Use the Linux build instructions (e.g., `npm run build:unix` or `./build.sh`).
    *   The `npm install .` process (which runs `install.js`) will detect WSL and attempt to fetch or build the Linux binary.
    *   The `bin/shotgun.js` wrapper is designed to correctly execute either `bin/shotgun` (Linux binary) or `bin/shotgun.exe` (Windows binary) if they are present in the `bin` directory. This can be useful in mixed environments if the `bin` directory is accessed from both Windows and WSL, though running the native binary for the environment is generally recommended.
*   **Global Installation & PATH:**
    *   Commands like `npm install -g .` or the `build.bat`/`build.sh` scripts aim to make `shotgun` (via the `shotgun.js` wrapper) available in your system's PATH.
    *   If the command is not found after installation, ensure your npm global binary directory is included in your PATH. You can find the npm global path by running `npm prefix -g` and then adding its `bin` subdirectory to your PATH.

### Usage

1.  **Run the application** from any directory (if installed globally or if `bin` is in PATH):
    ```bash
    shotgun
    # or explicitly if not in PATH but installed locally:
    # path/to/project/bin/shotgun.js
    ```

   **Note:** Templates are embedded in the Go binary, so the application does not need to be run from a specific project directory to find its own templates.

2.  **Select task type** using arrow keys and Enter.

3.  **Enter task description**:
    *   Type your multi-line task description.
    *   Lines are automatically numbered.
    *   Press `Alt+D` (or `Ctrl+D` on some terminals/platforms) when finished.

4.  **Configure rules** (optional):
    *   Choose Y/N for adding project-specific rules.
    *   If yes, enter multi-line rules and press `Alt+D` (or `Ctrl+D`).

5.  **Select files**:
    *   All files are **selected by default** (marked with ✓).
    *   Navigate the file tree with ↑/↓ arrows.
    *   Press `Space` to **deselect/select** files or directories (toggle ✓/✗).
    *   Press `Enter` to expand/collapse directories.
    *   Press `Alt+D` (or `Ctrl+D`) to continue with selected files.

6.  **Generate prompt**:
    *   The application processes your inputs.
    *   Generates a timestamped markdown file (e.g., `shotgun_prompt_YYYYMMDD_HHMMSS.md`).
    *   Shows the output file path.

### Keyboard Shortcuts

- `↑/↓` - Navigate menus and file tree
- `Enter` - Select option or expand/collapse directory
- `Space` - Toggle file/directory selection (✓ selected / ✗ deselected)
- `Alt+D` or `Ctrl+D` - Confirm input and proceed to next step (Ctrl+D might be more common on Linux/macOS)
- `Esc` - Go back to previous screen or cancel input
- `Ctrl+C` or `q` - Quit application

## Template Files

The application includes the following template files embedded in the binary (originally from the `templates/` directory):

- `templates/prompt_makePlan.md` - For architecture tasks
- `templates/prompt_makeDiffGitFormat.md` - For development tasks
- `templates/prompt_analyzeBug.md` - For debugging tasks
- `templates/prompt_projectManager.md` - For documentation sync tasks

Each template should contain placeholders:
- `{TASK}` - Replaced with your task description
- `{RULES}` - Replaced with project rules (if provided)
- `{FILE_STRUCTURE}` - Replaced with selected file contents and structure

## File Exclusions

The application automatically excludes common files and directories by default (these won't appear in the file tree for selection):

- Git files and directories (`.git/`, `.gitignore` and its patterns)
- Common dependency directories (`node_modules/`, `vendor/`)
- Build artifacts and caches (`build/`, `dist/`, `target/`, `*.pyc`, `__pycache__/`)
- IDE configuration files (`.vscode/`, `.idea/`)
- Log files and temporary files (`*.log`, `*.tmp`, `*.swp`, `*.cache`)
- Common binary file extensions (`*.exe`, `*.dll`, `*.so`, `*.o`, `*.a`, `*.class`, `*.jar`)

**Note**: All remaining files are **selected by default**. You typically only need to deselect files you explicitly *don't* want to include in your prompt context.

## Output

Generated prompts are saved in the current working directory as:
```
shotgun_prompt_YYYYMMDD_HHMMSS.md
```

The output includes:
- The complete processed prompt based on the chosen template.
- Your task description with line numbers.
- Project rules (if specified).
- A file structure tree of the selected files.
- The full content of all selected files.

## Architecture

```
shotgun-cli/
├── main.go                 # Application entry point (Go binary)
├── internal/               # Internal Go packages
│   ├── cmd/                # Cobra command definitions
│   ├── ui/                 # Bubble Tea TUI components (app.go, filetree.go)
│   ├── file/               # File system operations (scanner.go for gitignore, etc.)
│   └── template/           # Template processing (processor.go, embedded templates)
├── bin/                    # (Created during build) Contains compiled binary and wrapper
│   ├── shotgun             # (Linux/macOS) Compiled Go binary
│   ├── shotgun.exe         # (Windows) Compiled Go binary
│   └── shotgun.js          # Node.js wrapper script, entry point for 'shotgun' command
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── package.json            # npm package configuration, scripts
├── install.js              # Node.js script for binary download/fallback build (used by npm postinstall)
├── create-wrapper.js       # Node.js script to generate bin/shotgun.js
├── build.bat               # Windows batch script for building and global install
├── build.sh                # Unix shell script for building and global install
└── README.md               # This file
```

## Contributing

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/AmazingFeature`).
3. Make your changes.
4. Test thoroughly on Windows, Linux, and WSL if possible.
5. Commit your changes (`git commit -m 'Add some AmazingFeature'`).
6. Push to the branch (`git push origin feature/AmazingFeature`).
7. Open a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

Built with ❤️ using [Bubble Tea](https://github.com/charmbracelet/bubbletea) and Go.
