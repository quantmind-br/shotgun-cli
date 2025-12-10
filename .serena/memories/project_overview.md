# shotgun-cli Project Overview

## Purpose
shotgun-cli is a cross-platform CLI tool that generates LLM-optimized codebase contexts. It scans codebases, applies intelligent filtering patterns, and generates structured text representations optimized for Large Language Model consumption.

## Key Features
- **Interactive TUI Wizard**: 5-step Bubble Tea interface for guided context generation
- **Headless CLI Mode**: Commands for automation (`context generate`, `template`, `diff`, `config`)
- **Smart File Filtering**: Gitignore-style pattern matching with include/exclude support
- **Template System**: Multi-source template loading with custom overrides
- **Cross-platform**: Linux, macOS, and Windows support

## Two Modes of Operation
1. **TUI Wizard Mode** (default): 5-step guided interface
   - Step 1: File Selection (tree with Vim-style navigation)
   - Step 2: Template Selection
   - Step 3: Task Input
   - Step 4: Rules Input (optional)
   - Step 5: Review & Generate

2. **Headless CLI Mode**: Direct commands for automation
   - `shotgun-cli context generate` - Generate context
   - `shotgun-cli template [list|render|import|export]` - Template management
   - `shotgun-cli diff split` - Split large diffs
   - `shotgun-cli config [show|set]` - Configuration management

## Output Format
Generated context includes:
- File tree structure (hierarchical view)
- File content blocks in XML-like format: `<file path="...">content</file>`
- Template variables (task, rules, metadata)

## Configuration
- Config file: `~/.config/shotgun-cli/config.yaml`
- Environment prefix: `SHOTGUN_`
- XDG-compliant paths via `adrg/xdg`

## Technology Stack
- **Language**: Go 1.24
- **CLI**: Cobra + Viper
- **TUI**: Bubble Tea + Lip Gloss + Bubbles
- **Logging**: Zerolog
- **Patterns**: go-gitignore for gitignore-style filtering
