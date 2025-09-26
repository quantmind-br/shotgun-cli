# Shotgun-CLI Development Guide

## Executive Summary

Shotgun-CLI is a multiplatform terminal tool that generates and manipulates project code contexts for use with LLMs. It features both an interactive TUI wizard (Bubble Tea) and headless CLI modes, enabling consistent experience across Linux, macOS, Windows, and WSL for both local workflows and CI/CD automation.

### Key Goals
- Generate structured text representations of codebases within LLM token limits (10MB max)
- Provide 5-step interactive wizard with automatic project scanning
- Support template-based prompt generation with dynamic variable substitution
- Enable efficient context management for large codebases with intelligent chunking

## Core Requirements

### Functional Requirements

| ID | Requirement | Priority |
|----|-------------|----------|
| FR1 | Launch complete 5-step TUI wizard when running `shotgun-cli` without arguments | CRITICAL |
| FR2 | Save prompts to root as `shotgun-prompt-YYYYMMDD-HHMMSS.md` | CRITICAL |
| FR3 | Auto-exclude previously generated `shotgun-prompt-*.md` files from scans | HIGH |
| FR4 | Load templates via Go's `embed` package with placeholders `{TASK}`, `{RULES}`, `{FILE_STRUCTURE}`, `{CURRENT_DATE}` | HIGH |
| FR5 | Apply layered ignore rules: built-in → .gitignore → custom → explicit | CRITICAL |
| FR6 | Display navigable file tree with granular selection and progress feedback | CRITICAL |
| FR7 | Generate ASCII tree representation followed by `<file>` content blocks | CRITICAL |
| FR8 | Enforce 10MB size limit with cumulative counting and error reporting | CRITICAL |
| FR9 | Support headless CLI commands for automation | HIGH |
| FR10 | Cross-platform clipboard integration with fallback strategies | MEDIUM |

### Non-Functional Requirements

| ID | Requirement | Target |
|----|-------------|--------|
| NFR1 | Use Bubble Tea ≥ 1.3.5 for Windows F-key support | Required |
| NFR2 | Cross-platform compatibility (Linux, macOS, Windows, WSL) | Required |
| NFR3 | Process 10,000 files in <2 seconds | Performance |
| NFR4 | Memory usage <100MB for typical projects | Performance |
| NFR5 | Single binary <5MB | Deployment |

## Technical Architecture

### Technology Stack

```yaml
Language: Go 1.24
TUI Framework: Bubble Tea ≥1.3.5 with Lip Gloss
CLI Framework: Cobra
Logging: zerolog
Build: Make/Mage
Distribution: Goreleaser
Embedded Assets: Go embed package
```

### Project Structure

```text
shotgun-cli/
├── cmd/
│   └── shotgun-cli/
│       ├── main.go                 # Entry point
│       └── commands/                # Cobra commands
│           ├── root.go
│           ├── context.go          # context generate command
│           ├── diff.go              # diff split command
│           ├── template.go          # template commands
│           └── config.go            # config commands
├── internal/
│   ├── core/                       # Business logic
│   │   ├── scanner/                # Directory scanning
│   │   │   ├── scanner.go
│   │   │   └── scanner_test.go
│   │   ├── ignore/                 # Ignore engine
│   │   │   ├── engine.go
│   │   │   └── engine_test.go
│   │   ├── context/                # Context generation
│   │   │   ├── generator.go
│   │   │   └── generator_test.go
│   │   ├── template/               # Template management
│   │   │   ├── manager.go
│   │   │   └── manager_test.go
│   │   └── diff/                   # Diff splitting
│   │       ├── splitter.go
│   │       └── splitter_test.go
│   ├── ui/                         # TUI components
│   │   ├── components/
│   │   │   ├── tree/               # File tree component
│   │   │   ├── progress/           # Progress bar
│   │   │   ├── input/              # Text input
│   │   │   └── shortcuts/          # Keyboard shortcuts
│   │   ├── screens/                # Wizard steps
│   │   │   ├── file_selection.go   # Step 1
│   │   │   ├── template_selection.go # Step 2
│   │   │   ├── task_input.go       # Step 3
│   │   │   ├── rules_input.go      # Step 4
│   │   │   └── review.go           # Step 5
│   │   ├── styles/
│   │   │   └── theme.go            # Lip Gloss styles
│   │   └── wizard.go               # Main TUI orchestrator
│   ├── platform/                   # Platform-specific code
│   │   ├── clipboard/
│   │   │   ├── clipboard.go        # Interface
│   │   │   ├── linux.go            # wl-copy/xclip/xsel
│   │   │   ├── darwin.go           # pbcopy
│   │   │   └── windows.go          # PowerShell/clip.exe
│   │   └── config/
│   │       ├── paths.go            # XDG/AppData paths
│   │       └── config.go           # Config management
│   └── utils/
│       ├── fs/                     # Filesystem helpers
│       └── progress/               # Progress reporting
├── assets/                         # Embedded resources
│   ├── templates/
│   │   ├── makeDiffGitFormat.tmpl
│   │   ├── makePlan.tmpl
│   │   ├── analyzeBug.tmpl
│   │   └── projectManager.tmpl
│   └── ignore.glob                 # Default ignore patterns
├── test/
│   ├── fixtures/                   # Test file structures
│   └── e2e/                        # End-to-end tests
├── .github/
│   └── workflows/
│       ├── ci.yaml                 # Test and lint
│       └── release.yaml            # Goreleaser workflow
├── Makefile
├── go.mod
└── README.md
```

## Core Components Implementation

### Data Models

```go
// FileNode - Core tree structure
type FileNode struct {
    Name            string
    Path            string      // Absolute path
    RelPath         string      // Relative from root
    IsDir           bool
    Children        []FileNode
    Selected        bool
    IsGitignored    bool
    IsCustomIgnored bool
    Size            int64
}

// TemplateDescriptor - Template metadata
type TemplateDescriptor struct {
    Name         string
    Description  string
    Content      string
    RequiredVars []string    // {TASK}, {RULES}, etc
    IsCustom     bool
}

// Progress - Progress reporting
type Progress struct {
    Current int
    Total   int
    Stage   string          // "scanning", "generating"
    Message string
}

// Config - User configuration
type Config struct {
    Version     string
    IgnoreRules []string
    Theme       string
    KeyBindings map[string]string
    MaxSize     int64       // Default 10MB
}

// WizardModel - TUI state
type WizardModel struct {
    step          int                    // 1-5
    fileTree      *FileNode
    selectedFiles map[string]bool
    template      *TemplateDescriptor
    taskDesc      string
    rules         string
    progress      Progress
    error         error
}
```

### Component Interfaces

```go
// Scanner - Directory scanning
type Scanner interface {
    Scan(rootPath string) (*FileNode, error)
    ScanWithProgress(rootPath string, progress chan<- Progress) (*FileNode, error)
}

// IgnoreEngine - Ignore rule processing
type IgnoreEngine interface {
    ShouldIgnore(path string) (bool, IgnoreReason)
    AddCustomRule(pattern string) error
    LoadGitignore(path string) error
}

// ContextGenerator - Context generation
type ContextGenerator interface {
    Generate(root *FileNode, config GenerateConfig) (string, error)
    GenerateWithLimit(root *FileNode, limit int64) (string, error)
}

// TemplateManager - Template management
type TemplateManager interface {
    ListTemplates() []TemplateDescriptor
    LoadTemplate(name string) (*TemplateDescriptor, error)
    Render(template *TemplateDescriptor, vars map[string]string) (string, error)
}

// ClipboardManager - Cross-platform clipboard
type ClipboardManager interface {
    Copy(content string) error
    GetCommand() (string, []string)
    IsAvailable() bool
}
```

## User Flows

### Main Flow: TUI Wizard

```text
1. Launch: shotgun-cli (no args)
   → Auto-scan project directory
   → Display Step 1

2. Step 1: File Selection
   - Tree view with checkboxes
   - [x] Selected, [ ] Unselected
   - (g) gitignored, (p) previous prompt, (c) custom rules
   - Navigation: ↑↓ move, Space toggle, →← expand/collapse
   - /: Filter files incrementally, d: Toggle directory selection
   - i: Toggle ignored files visibility, F5: Re-scan, F8: Next step

3. Step 2: Template Selection
   - List templates with descriptions
   - Show required variables
   - Navigation: ↑↓ select, Enter confirm
   - F10: Back, F8: Next

4. Step 3: Task Description
   - Multiline text input
   - Character/token counter
   - All keys for typing
   - F10: Back, F8: Next

5. Step 4: Rules Input (Optional)
   - Multiline text input
   - F10: Back, F8: Next

6. Step 5: Review & Generate
   - Summary display
   - Estimated size/tokens
   - Ctrl+Enter: Generate
   - Esc: Back

7. Generation
   - Progress bar
   - Save to shotgun-prompt-YYYYMMDD-HHMMSS.md
   - Copy to clipboard
   - Auto-exit on success
```

### CLI Commands

```bash
# Context generation
shotgun-cli context generate \
  --root . \
  --include "*.go" \
  --exclude "vendor/*" \
  --output context.md \
  --max-size 5MB

# Template operations
shotgun-cli template list
shotgun-cli template render makePlan \
  --var TASK="Implement feature X" \
  --var RULES="Follow coding standards"

# Diff splitting
shotgun-cli diff split \
  --input large.diff \
  --approx-lines 500 \
  --output-dir chunks/

# Configuration
shotgun-cli config show
shotgun-cli config set theme dark
```

## Implementation Priorities

### Phase 1: Core Foundation (Week 1)
1. Project setup and structure
2. FileNode and Scanner implementation
3. Basic ignore engine (.gitignore support)
4. Context generator with size limits
5. Basic CLI with context command

### Phase 2: Template System (Week 2)
1. Template manager with embedded defaults
2. Variable substitution
3. Template CLI commands
4. Diff splitter implementation
5. Configuration system

### Phase 3: TUI Wizard (Week 3)
1. Bubble Tea setup with navigation
2. File selection screen (Step 1)
3. Template selection (Step 2)
4. Text input screens (Steps 3-4)
5. Review and generation (Step 5)

### Phase 4: Polish & Platform (Week 4)
1. Cross-platform clipboard
2. Filesystem monitoring
3. Performance optimization
4. Documentation
5. Release pipeline

## Key Implementation Notes

### File Scanning
```go
func (s *Scanner) Scan(rootPath string) (*FileNode, error) {
    // 1. Initialize root node
    // 2. Walk directory recursively
    // 3. Apply ignore rules at each level
    // 4. Sort: directories first, then files
    // 5. Auto-exclude shotgun-prompt-*.md
    // 6. Return complete tree
}
```

### Context Generation
```go
func (g *Generator) Generate(root *FileNode, config GenerateConfig) (string, error) {
    var builder strings.Builder
    var totalSize int64

    // 1. Generate ASCII tree
    builder.WriteString(generateTree(root))

    // 2. Add file contents
    for _, node := range flattenSelected(root) {
        content := readFile(node.Path)

        // Check size limit
        if totalSize + int64(len(content)) > config.MaxSize {
            return "", ErrContextTooLarge
        }

        // Add file block
        fmt.Fprintf(&builder, "\n<file path=\"%s\">\n%s\n</file>\n",
                    node.RelPath, content)
        totalSize += int64(len(content))
    }

    return builder.String(), nil
}
```

### TUI Keyboard Navigation
```go
func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Global shortcuts
        switch msg.String() {
        case "ctrl+q":
            return m, tea.Quit
        case "?":
            m.showHelp = !m.showHelp
        case "f8", "ctrl+pgdown":
            if m.canAdvance() {
                m.step++
            }
        case "f10", "ctrl+pgup":
            if m.step > 1 {
                m.step--
            }
        }

        // Step-specific handling
        switch m.step {
        case StepFileSelection:
            return m.updateFileSelection(msg)
        // ... other steps
        }
    }
    return m, nil
}
```

### Clipboard Integration
```go
// Platform-specific implementations
// linux.go
func (c *LinuxClipboard) Copy(content string) error {
    // Try in order: wl-copy, xclip, xsel
    cmds := [][]string{
        {"wl-copy"},
        {"xclip", "-selection", "clipboard"},
        {"xsel", "--clipboard", "--input"},
    }

    for _, args := range cmds {
        if err := tryCommand(args, content); err == nil {
            return nil
        }
    }
    return ErrNoClipboardTool
}
```

## Testing Strategy

### Unit Tests
- Core logic: scanner, ignore engine, generator
- Components: tree, input, progress
- Platform code: clipboard, config paths

### Integration Tests
- File scanning with fixtures
- Template rendering
- Context generation with size limits

### E2E Tests
- Complete wizard flow
- CLI command execution
- Cross-platform validation

## Error Handling

```go
// Custom errors
var (
    ErrContextTooLarge = errors.New("context exceeds size limit")
    ErrNoFilesSelected = errors.New("no files selected")
    ErrTemplateNotFound = errors.New("template not found")
    ErrMissingVariable = errors.New("required template variable missing")
    ErrNoClipboardTool = errors.New("no clipboard tool available")
)

// Error wrapping for context
return fmt.Errorf("scanning directory %s: %w", path, err)
```

## Performance Targets

- Scan 10,000 files: <2 seconds
- Generate 10MB context: <1 second
- TUI response: <50ms
- Memory usage: <100MB
- Binary size: <5MB

## Platform Considerations

### Windows
- Use Bubble Tea ≥1.3.5 for F-key support
- PowerShell for clipboard, fallback to clip.exe
- %APPDATA% for config location

### macOS
- pbcopy for clipboard
- ~/Library/Application Support for config

### Linux
- Try wl-copy (Wayland), xclip, xsel
- XDG_CONFIG_HOME for config

### WSL
- Detect WSL environment
- Use clip.exe for clipboard
- Linux paths internally

## Development Workflow

```bash
# Setup
go mod init github.com/user/shotgun-cli
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/spf13/cobra@latest
go get github.com/rs/zerolog@latest

# Development
make build        # Build binary
make test         # Run all tests
make run          # Run locally
make lint         # golangci-lint

# Release
git tag v1.0.0
git push --tags   # Triggers Goreleaser
```

## Success Metrics

- Zero-config launch to prompt generation in <5 steps
- Support for projects with 10K+ files
- Cross-platform compatibility verified
- <2s performance for typical projects
- Single binary distribution

## Related Documentation

- [Keyboard Shortcuts Reference](keyboard-shortcuts.md) - Complete keyboard shortcut guide
- [Epic 3: Interactive TUI Wizard](prd/epic-3-interactive-tui-wizard.md) - Detailed TUI specifications
- [User Interface Design Goals](prd/user-interface-design-goals.md) - Design principles and rationale