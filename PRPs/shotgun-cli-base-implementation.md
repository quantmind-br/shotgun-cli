# Shotgun-CLI Base Implementation PRP

## Goal

**Feature Goal**: Implement complete Shotgun-CLI multiplatform terminal tool with both interactive TUI wizard and headless CLI modes for generating LLM-optimized codebase contexts.

**Deliverable**: Single cross-platform binary (~5MB) supporting:
- 5-step TUI wizard (`shotgun-cli` with no args)
- Headless CLI commands (`shotgun-cli context generate`, `shotgun-cli template render`, etc.)
- File scanning with layered ignore rules (built-in → .gitignore → custom)
- ASCII tree generation and context formatting with 10MB size limits
- Cross-platform clipboard integration with graceful fallbacks
- Template system with embedded defaults and variable substitution

**Success Definition**:
- Zero-config launch to prompt generation in <5 steps
- Process 10,000+ files in <2 seconds
- Cross-platform compatibility (Linux, macOS, Windows, WSL)
- Memory usage <100MB for typical projects
- Single binary distribution with embedded assets

## User Persona

**Target User**: Software developers and DevOps engineers using LLMs for code assistance

**Use Case**: Generate structured codebase representations that fit within LLM context windows for code analysis, planning, and implementation tasks

**User Journey**:
1. Navigate to project directory
2. Run `shotgun-cli` (launches TUI wizard)
3. Select files using tree navigation (Step 1)
4. Choose template (Step 2)
5. Enter task description (Step 3)
6. Add rules/constraints (Step 4)
7. Review and generate (Step 5)
8. Context saved as `shotgun-prompt-YYYYMMDD-HHMMSS.md` and copied to clipboard

**Pain Points Addressed**:
- Manual file selection for LLM context is tedious
- Context often exceeds LLM token limits
- Inconsistent prompt formatting across team
- No automation support for CI/CD workflows

## Why

- **Business Value**: Accelerates LLM-assisted development by providing structured, consistent codebase context
- **Integration**: Complements existing development workflows without requiring changes to IDE or tooling
- **Problems Solved**:
  - Context size management (10MB limit enforcement)
  - Cross-platform file scanning with intelligent ignore rules
  - Template-based prompt standardization for teams
  - Automation support for CI/CD pipelines

## What

**User-Visible Behavior**:
- Interactive TUI wizard with keyboard navigation (F8/F10 for steps, arrow keys, space/enter for selection)
- File tree with checkbox selection, ignore status indicators `(g)` gitignored, `(c)` custom ignored
- Real-time progress bars during scanning and generation
- Automatic clipboard copy with platform detection
- Generated files auto-excluded from future scans

**Technical Requirements**:
- Go 1.24+ with Bubble Tea ≥1.3.5 (Windows F-key support)
- Cross-compilation for Linux, macOS, Windows, WSL
- Embedded templates and assets using `go:embed`
- Layered ignore rules: built-in → .gitignore → custom → explicit
- ASCII tree generation with file content blocks
- Template variable substitution: `{TASK}`, `{RULES}`, `{FILE_STRUCTURE}`, `{CURRENT_DATE}`

### Success Criteria

- [ ] TUI wizard launches with `shotgun-cli` (no arguments)
- [ ] Headless commands work: `shotgun-cli context generate --root . --include "*.go"`
- [ ] File scanning processes 10K+ files in <2 seconds with progress reporting
- [ ] Generated context includes ASCII tree + file content blocks within 10MB limit
- [ ] Cross-platform clipboard integration with graceful fallback messages
- [ ] Templates render with variable substitution from embedded assets
- [ ] Auto-exclusion of `shotgun-prompt-*.md` files from scans
- [ ] Memory usage stays under 100MB for typical projects

## All Needed Context

### Context Completeness Check

_"If someone knew nothing about this codebase, would they have everything needed to implement this successfully?"_

✅ **YES** - This PRP provides:
- Complete project structure with file responsibilities
- Specific implementation patterns from real-world examples
- All library integration details with version requirements
- Cross-platform compatibility patterns
- Performance optimization techniques
- Comprehensive validation commands
- Common gotchas and solutions

### Documentation & References

```yaml
# MUST READ - Critical library documentation
- docfile: PRPs/ai_docs/bubble_tea_patterns.md
  why: Complete TUI wizard implementation with 5-step navigation, file tree, progress bars
  critical: Multi-step state management, keyboard navigation, cross-platform F-key support

- docfile: PRPs/ai_docs/cobra_cli_patterns.md
  why: CLI framework patterns for headless commands and root command TUI launch
  critical: Flag validation, subcommand structure, configuration integration

- docfile: PRPs/ai_docs/file_scanning_patterns.md
  why: High-performance directory scanning, ignore rules, tree generation, context formatting
  critical: Memory management for 10K+ files, layered ignore rules, size limit enforcement

- docfile: PRPs/ai_docs/clipboard_integration_patterns.md
  why: Cross-platform clipboard with fallback strategies for Linux/macOS/Windows/WSL
  critical: Platform detection, tool availability checking, graceful degradation

# MUST READ - Official documentation with specific sections
- url: https://github.com/charmbracelet/bubbletea
  why: TUI framework for 5-step wizard implementation
  critical: Model-Update-View pattern, keyboard event handling, command patterns

- url: https://github.com/spf13/cobra
  why: CLI framework for headless commands and root command structure
  critical: Command composition, flag validation, configuration integration

- url: https://pkg.go.dev/embed
  why: Template embedding for self-contained binary distribution
  critical: Asset embedding patterns, template loading from embedded filesystem

- url: https://github.com/sabhiram/go-gitignore
  why: Gitignore rule parsing and application for file filtering
  critical: Rule compilation, path matching, layered ignore rule precedence

- url: https://pkg.go.dev/path/filepath#WalkDir
  why: Efficient directory traversal for file scanning
  critical: Performance optimization, error handling, early termination

- url: https://goreleaser.com/quick-start/
  why: Cross-platform binary building and distribution
  critical: Multi-platform compilation, binary optimization, asset embedding
```

### Current Codebase Tree

Currently empty - implementing from scratch:

```bash
shotgun-cli/
├── (empty directory - new project)
```

### Desired Codebase Tree

```bash
shotgun-cli/
├── main.go                                # Entry point
├── cmd/
│   ├── root.go                           # Root command with TUI launch
│   ├── context.go                        # context generate command
│   ├── template.go                       # template list/render commands
│   ├── diff.go                           # diff split command
│   └── config.go                         # config show/set commands
├── internal/
│   ├── core/
│   │   ├── scanner/
│   │   │   ├── scanner.go                # Directory scanning interface
│   │   │   ├── filesystem.go             # File system walker
│   │   │   └── scanner_test.go           # Unit tests
│   │   ├── ignore/
│   │   │   ├── engine.go                 # Layered ignore rule engine
│   │   │   └── engine_test.go            # Unit tests
│   │   ├── context/
│   │   │   ├── generator.go              # Context generation with size limits
│   │   │   └── generator_test.go         # Unit tests
│   │   ├── template/
│   │   │   ├── manager.go                # Template loading and rendering
│   │   │   └── manager_test.go           # Unit tests
│   │   └── diff/
│   │       ├── splitter.go               # Diff file splitting
│   │       └── splitter_test.go          # Unit tests
│   ├── ui/
│   │   ├── wizard.go                     # Main TUI orchestrator
│   │   ├── components/
│   │   │   ├── tree/                     # File tree component
│   │   │   │   ├── tree.go
│   │   │   │   └── tree_test.go
│   │   │   ├── progress/                 # Progress bar component
│   │   │   │   ├── progress.go
│   │   │   │   └── progress_test.go
│   │   │   └── input/                    # Text input component
│   │   │       ├── input.go
│   │   │       └── input_test.go
│   │   ├── screens/                      # Wizard steps
│   │   │   ├── file_selection.go         # Step 1: File tree selection
│   │   │   ├── template_selection.go     # Step 2: Template choice
│   │   │   ├── task_input.go             # Step 3: Task description
│   │   │   ├── rules_input.go            # Step 4: Rules input
│   │   │   └── review.go                 # Step 5: Review and generate
│   │   └── styles/
│   │       └── theme.go                  # Lip Gloss styling
│   ├── platform/
│   │   ├── clipboard/
│   │   │   ├── clipboard.go              # Cross-platform interface
│   │   │   ├── linux.go                  # Linux clipboard (wl-copy/xclip/xsel)
│   │   │   ├── darwin.go                 # macOS clipboard (pbcopy)
│   │   │   ├── windows.go                # Windows clipboard (clip/powershell)
│   │   │   └── clipboard_test.go         # Unit tests
│   │   └── config/
│   │       ├── paths.go                  # XDG/AppData configuration paths
│   │       └── config.go                 # Configuration management
│   └── utils/
│       ├── fs/
│       │   ├── helpers.go                # Filesystem utilities
│       │   └── helpers_test.go           # Unit tests
│       └── progress/
│           ├── reporter.go               # Progress reporting utilities
│           └── reporter_test.go          # Unit tests
├── assets/                               # Embedded resources
│   ├── templates/
│   │   ├── makeDiffGitFormat.tmpl       # Git diff formatting template
│   │   ├── makePlan.tmpl                # Planning template
│   │   ├── analyzeBug.tmpl              # Bug analysis template
│   │   └── projectManager.tmpl          # Project management template
│   └── ignore.glob                       # Default ignore patterns
├── test/
│   ├── fixtures/                        # Test file structures
│   │   ├── sample-project/              # Mock project for testing
│   │   └── large-project/               # Performance testing
│   └── e2e/
│       ├── wizard_test.go               # End-to-end TUI tests
│       └── cli_test.go                  # End-to-end CLI tests
├── .github/
│   └── workflows/
│       ├── ci.yaml                      # Test and lint workflow
│       └── release.yaml                 # Goreleaser release workflow
├── .goreleaser.yaml                     # Release configuration
├── Makefile                             # Build automation
├── go.mod                               # Dependencies
├── go.sum                               # Dependency checksums
└── README.md                            # Project documentation
```

### Known Gotchas & Library Quirks

```go
// CRITICAL: Bubble Tea ≥1.3.5 required for Windows F-key support
// Use specific version in go.mod:
// github.com/charmbracelet/bubbletea v1.3.5

// CRITICAL: filepath.WalkDir (Go 1.16+) is more efficient than filepath.Walk
// Always use WalkDir to avoid unnecessary os.Lstat calls

// CRITICAL: Cross-platform clipboard requires different tools:
// Linux: wl-copy (Wayland) > xclip > xsel (X11)
// macOS: pbcopy (always available)
// Windows: clip.exe > powershell Set-Clipboard
// WSL: clip.exe (Windows binary from WSL)

// CRITICAL: File tree memory management for large projects
// Use lazy loading and virtual scrolling for trees with 10K+ files

// CRITICAL: gitignore rules must be applied at directory level
// Use fs.SkipDir to avoid traversing ignored directories entirely

// CRITICAL: Template variable substitution security
// Always validate template content before substitution to prevent injection
```

## Implementation Blueprint

### Data Models and Structure

```go
// Core data models ensuring type safety and consistency

// FileNode - Core tree structure with selection state
type FileNode struct {
    Name            string      `json:"name"`
    Path            string      `json:"path"`        // Absolute path
    RelPath         string      `json:"rel_path"`    // Relative from root
    IsDir           bool        `json:"is_dir"`
    Children        []FileNode  `json:"children,omitempty"`
    Selected        bool        `json:"selected"`
    IsGitignored    bool        `json:"is_gitignored"`
    IsCustomIgnored bool        `json:"is_custom_ignored"`
    Size            int64       `json:"size"`
    Expanded        bool        `json:"expanded"`    // For TUI tree navigation
}

// TemplateDescriptor - Template metadata with validation
type TemplateDescriptor struct {
    Name         string   `json:"name"`
    Description  string   `json:"description"`
    Content      string   `json:"content"`
    RequiredVars []string `json:"required_vars"`  // {TASK}, {RULES}, etc
    IsCustom     bool     `json:"is_custom"`
}

// GenerateConfig - Context generation configuration
type GenerateConfig struct {
    RootPath     string            `json:"root_path"`
    MaxSize      int64             `json:"max_size"`      // 10MB default
    MaxFiles     int               `json:"max_files"`
    Include      []string          `json:"include"`
    Exclude      []string          `json:"exclude"`
    SkipBinary   bool              `json:"skip_binary"`
    TemplateVars map[string]string `json:"template_vars"`
}

// Progress - Progress reporting with stages
type Progress struct {
    Current int64  `json:"current"`
    Total   int64  `json:"total"`
    Stage   string `json:"stage"`    // "scanning", "generating", "complete"
    Message string `json:"message"`
}

// WizardModel - TUI state management
type WizardModel struct {
    step          int                      // 1-5 current step
    fileTree      *FileNode               // File tree state
    selectedFiles map[string]bool         // File selections
    template      *TemplateDescriptor     // Selected template
    taskDesc      string                  // User task input
    rules         string                  // User rules input
    progress      Progress                // Progress state
    error         error                   // Current error
    width         int                     // Terminal dimensions
    height        int
    showHelp      bool                    // Help overlay state
}
```

### Implementation Tasks (Ordered by Dependencies)

```yaml
Task 1: CREATE go.mod and project structure
  - IMPLEMENT: Go module initialization with required dependencies
  - DEPENDENCIES: Go 1.24+, Bubble Tea ≥1.3.5, Cobra, Lip Gloss, zerolog
  - COMMANDS: |
    go mod init shotgun-cli
    go get github.com/charmbracelet/bubbletea@v1.3.5
    go get github.com/charmbracelet/lipgloss@latest
    go get github.com/spf13/cobra@latest
    go get github.com/rs/zerolog@latest
    go get github.com/sabhiram/go-gitignore@latest
  - PLACEMENT: Root directory

Task 2: CREATE internal/core/scanner package
  - IMPLEMENT: FileNode struct, Scanner interface, filesystem walker
  - FOLLOW pattern: PRPs/ai_docs/file_scanning_patterns.md (high-performance scanning)
  - NAMING: Scanner interface, FileSystemScanner struct, ScanWithProgress methods
  - PLACEMENT: internal/core/scanner/
  - FILES: scanner.go, filesystem.go, scanner_test.go

Task 3: CREATE internal/core/ignore package
  - IMPLEMENT: IgnoreEngine with layered rule support (built-in → .gitignore → custom)
  - FOLLOW pattern: PRPs/ai_docs/file_scanning_patterns.md (layered ignore rules)
  - DEPENDENCIES: github.com/sabhiram/go-gitignore
  - NAMING: IgnoreEngine struct, ShouldIgnore methods, IgnoreReason enum
  - PLACEMENT: internal/core/ignore/
  - FILES: engine.go, engine_test.go

Task 4: CREATE internal/core/context package
  - IMPLEMENT: ContextGenerator with template rendering and size limits
  - FOLLOW pattern: PRPs/ai_docs/file_scanning_patterns.md (context generation)
  - DEPENDENCIES: text/template, Task 2 (scanner), Task 3 (ignore)
  - NAMING: ContextGenerator struct, Generate methods, GenerateConfig
  - PLACEMENT: internal/core/context/
  - FILES: generator.go, generator_test.go

Task 5: CREATE internal/platform/clipboard package
  - IMPLEMENT: Cross-platform clipboard with fallback chains
  - FOLLOW pattern: PRPs/ai_docs/clipboard_integration_patterns.md (platform detection)
  - NAMING: ClipboardManager interface, Manager struct, platform-specific implementations
  - PLACEMENT: internal/platform/clipboard/
  - FILES: clipboard.go, linux.go, darwin.go, windows.go, clipboard_test.go

Task 6: CREATE internal/core/template package
  - IMPLEMENT: Template manager with embedded assets and variable substitution
  - DEPENDENCIES: embed package, text/template
  - FOLLOW pattern: go:embed for asset loading
  - NAMING: TemplateManager interface, Manager struct, TemplateDescriptor
  - PLACEMENT: internal/core/template/
  - FILES: manager.go, manager_test.go

Task 7: CREATE assets/ directory with embedded templates
  - IMPLEMENT: Template files with variable placeholders {TASK}, {RULES}, etc.
  - FOLLOW pattern: Text template syntax with safe variable substitution
  - PLACEMENT: assets/templates/
  - FILES: makePlan.tmpl, analyzeBug.tmpl, makeDiffGitFormat.tmpl, projectManager.tmpl

Task 8: CREATE cmd/ package with Cobra commands
  - IMPLEMENT: Root command launching TUI, subcommands for headless operation
  - FOLLOW pattern: PRPs/ai_docs/cobra_cli_patterns.md (command structure)
  - DEPENDENCIES: github.com/spf13/cobra, all core packages
  - NAMING: rootCmd, contextCmd, templateCmd with proper flag validation
  - PLACEMENT: cmd/
  - FILES: root.go, context.go, template.go, diff.go, config.go

Task 9: CREATE internal/ui/ package with Bubble Tea components
  - IMPLEMENT: 5-step wizard with file tree, progress bars, text input
  - FOLLOW pattern: PRPs/ai_docs/bubble_tea_patterns.md (multi-step wizard)
  - DEPENDENCIES: Bubble Tea ≥1.3.5, Lip Gloss
  - NAMING: WizardModel, screen-specific models, component models
  - PLACEMENT: internal/ui/
  - FILES: wizard.go, screens/*.go, components/*.go, styles/theme.go

Task 10: CREATE main.go entry point
  - IMPLEMENT: Application entry point with proper error handling
  - FOLLOW pattern: cobra.Execute() with logging setup
  - DEPENDENCIES: cmd package, zerolog
  - INTEGRATION: Root command execution with graceful error handling
  - PLACEMENT: Root directory

Task 11: CREATE test infrastructure
  - IMPLEMENT: Unit tests for all packages, fixtures for e2e testing
  - FOLLOW pattern: Standard Go testing with table-driven tests
  - NAMING: *_test.go files, TestMain functions, benchmark tests
  - PLACEMENT: test/, alongside source files
  - COVERAGE: All public interfaces and critical paths

Task 12: CREATE build automation
  - IMPLEMENT: Makefile, Goreleaser config, GitHub Actions
  - FOLLOW pattern: Cross-platform builds with asset embedding
  - NAMING: Standard build targets (build, test, lint, release)
  - PLACEMENT: Root directory, .github/workflows/
  - FILES: Makefile, .goreleaser.yaml, .github/workflows/ci.yaml
```

### Implementation Patterns & Key Details

```go
// Scanner pattern with progress reporting
func (s *FileSystemScanner) ScanWithProgress(rootPath string, progress chan<- Progress) (*FileNode, error) {
    // PATTERN: Two-pass scanning - count then build (PRPs/ai_docs/file_scanning_patterns.md)
    totalItems, err := s.countItems(rootPath)
    if err != nil {
        return nil, fmt.Errorf("failed to count items: %w", err)
    }

    // CRITICAL: Use filepath.WalkDir for performance
    // GOTCHA: Apply ignore rules at directory level with fs.SkipDir
    // PATTERN: Report progress every 100 items to avoid UI blocking
}

// TUI wizard navigation pattern
func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // CRITICAL: F-key support for Windows (Bubble Tea ≥1.3.5)
        switch msg.String() {
        case "f8", "ctrl+pgdown":  // Next step with Windows fallback
            if m.canAdvance() { m.step++ }
        case "f10", "ctrl+pgup":   // Previous step with Windows fallback
            if m.step > 1 { m.step-- }
        }
    }
    // PATTERN: Delegate to step-specific handlers
    return m.handleStepInput(msg)
}

// Clipboard integration with fallbacks
func (m *Manager) Copy(content string) error {
    // PATTERN: Platform detection and tool availability checking
    // CRITICAL: Graceful degradation with user feedback
    if !m.IsAvailable() {
        return fmt.Errorf("no clipboard tool available on %s", m.platform)
    }

    // GOTCHA: WSL uses clip.exe (Windows binary)
    // GOTCHA: Linux needs different tools for Wayland vs X11
    return m.copyWithTimeout(content, 10*time.Second)
}

// Template rendering with security
func (tm *TemplateManager) Render(template *TemplateDescriptor, vars map[string]string) (string, error) {
    // CRITICAL: Validate all required variables present
    for _, required := range template.RequiredVars {
        if _, exists := vars[required]; !exists {
            return "", fmt.Errorf("missing required variable: %s", required)
        }
    }

    // PATTERN: Safe template execution with custom functions
    tmpl, err := text.Template.New(template.Name).Funcs(tm.getFuncMap()).Parse(template.Content)
    // RETURN: Rendered template with substituted variables
}
```

### Integration Points

```yaml
EMBEDDED_ASSETS:
  - location: assets/templates/*.tmpl
  - pattern: "//go:embed assets"
  - access: "templateFS.ReadFile()"

CONFIGURATION:
  - paths: ~/.config/shotgun-cli/ (Linux), ~/Library/Application Support/shotgun-cli/ (macOS), %APPDATA%/shotgun-cli/ (Windows)
  - format: YAML configuration files
  - pattern: "viper.AddConfigPath() with XDG compliance"

LOGGING:
  - library: zerolog with console writer
  - levels: info (default), debug (verbose flag)
  - pattern: "Structured logging with context fields"

CLIPBOARD:
  - linux: wl-copy → xclip → xsel fallback chain
  - macos: pbcopy (always available)
  - windows: clip.exe → powershell fallback
  - wsl: clip.exe with environment detection
```

## Validation Loop

### Level 1: Syntax & Style (Immediate Feedback)

```bash
# Run after each package creation
go mod tidy                                    # Clean dependencies
go vet ./...                                   # Static analysis
golangci-lint run                              # Comprehensive linting
go fmt ./...                                   # Code formatting

# Build validation
go build -o shotgun-cli ./main.go             # Build binary
./shotgun-cli --help                          # Test basic functionality

# Expected: Clean build, no lint errors, help text displays correctly
```

### Level 2: Unit Tests (Component Validation)

```bash
# Test each package as created
go test ./internal/core/scanner -v            # Scanner package tests
go test ./internal/core/ignore -v             # Ignore engine tests
go test ./internal/core/context -v            # Context generator tests
go test ./internal/platform/clipboard -v      # Clipboard tests
go test ./internal/ui -v                      # TUI component tests

# Full test suite with coverage
go test ./... -v -race -cover                 # All packages with race detection
go test ./... -bench=. -benchmem             # Performance benchmarks

# Expected: All tests pass, >90% coverage for core packages
```

### Level 3: Integration Testing (System Validation)

```bash
# TUI wizard integration
echo "Test TUI wizard launch"
./shotgun-cli                                 # Should launch TUI wizard
# Manual test: Navigate through 5 steps, verify file tree, template selection

# Headless CLI integration
./shotgun-cli context generate --root . --include "*.go" --output test-context.md
ls -la shotgun-prompt-*.md                   # Verify file created
grep -q "# Project Context" shotgun-prompt-*.md  # Verify content format

# Template system integration
./shotgun-cli template list                   # Should show embedded templates
./shotgun-cli template render makePlan --var TASK="Test task" --var RULES="Test rules"

# Clipboard integration (platform-specific)
./shotgun-cli context generate --root . | head -c 1000  # Generate small context
# Manual test: Verify clipboard contains generated content

# Performance validation
find . -name "*.go" | wc -l                  # Count Go files
time ./shotgun-cli context generate --root . # Should complete in <2s for typical projects

# Expected: All commands work, TUI navigable, files generated, clipboard functional
```

### Level 4: Cross-Platform & Performance Validation

```bash
# Cross-platform build testing
GOOS=linux GOARCH=amd64 go build -o shotgun-cli-linux ./main.go
GOOS=darwin GOARCH=amd64 go build -o shotgun-cli-darwin ./main.go
GOOS=windows GOARCH=amd64 go build -o shotgun-cli-windows.exe ./main.go
ls -la shotgun-cli-*                         # Verify all binaries created

# Binary size validation (<5MB requirement)
ls -lh shotgun-cli-* | awk '{print $5, $9}'  # Check file sizes

# Memory usage validation (<100MB requirement)
# Linux/macOS:
/usr/bin/time -v ./shotgun-cli context generate --root . 2>&1 | grep "Maximum resident set size"
# Windows:
# Use Process Explorer or similar tool

# Large project performance test
# Create test fixture with 10K+ files
find test/fixtures/large-project -type f | wc -l  # Should be >10K files
time ./shotgun-cli context generate --root test/fixtures/large-project

# WSL-specific testing (if available)
# Test clip.exe availability and clipboard integration in WSL environment

# Expected: All platforms build successfully, binaries <5MB, memory <100MB, performance targets met
```

## Final Validation Checklist

### Technical Validation

- [ ] All 4 validation levels completed successfully
- [ ] All tests pass: `go test ./... -v -race`
- [ ] No lint errors: `golangci-lint run`
- [ ] No vet issues: `go vet ./...`
- [ ] Clean formatting: `go fmt ./...`
- [ ] Binary builds for all platforms (Linux, macOS, Windows)
- [ ] Binary size <5MB for all platforms

### Feature Validation

- [ ] TUI wizard launches with `shotgun-cli` (no args)
- [ ] 5-step navigation works with F8/F10 and fallback keys
- [ ] File tree shows with selections, ignore indicators (g), (c)
- [ ] Template selection displays embedded templates
- [ ] Text input accepts task description and rules
- [ ] Review step shows summary with estimated size
- [ ] Context generation creates `shotgun-prompt-YYYYMMDD-HHMMSS.md`
- [ ] Headless commands work: `shotgun-cli context generate --help`
- [ ] Template commands work: `shotgun-cli template list`
- [ ] Configuration commands work: `shotgun-cli config show`

### Performance Validation

- [ ] Scans 10K+ files in <2 seconds
- [ ] Memory usage <100MB for typical projects
- [ ] Context generation respects 10MB size limit
- [ ] Progress reporting updates smoothly during operations
- [ ] TUI responds to keyboard input without lag (<50ms)

### Cross-Platform Validation

- [ ] Clipboard works on Linux (wl-copy/xclip/xsel detection)
- [ ] Clipboard works on macOS (pbcopy)
- [ ] Clipboard works on Windows (clip.exe/powershell)
- [ ] WSL detection works with clip.exe fallback
- [ ] F-key support works on Windows (Bubble Tea ≥1.3.5)
- [ ] File paths handle correctly across platforms
- [ ] Configuration directories follow platform conventions

### Code Quality Validation

- [ ] Follows Go conventions and idioms
- [ ] Error handling provides actionable messages
- [ ] Logging levels appropriate (info/debug)
- [ ] No hardcoded paths or platform assumptions
- [ ] Resource cleanup (file handles, goroutines)
- [ ] Graceful handling of interrupted operations (Ctrl+C)

---

## Anti-Patterns to Avoid

- ❌ Don't use `filepath.Walk` - use `filepath.WalkDir` for better performance
- ❌ Don't load entire large files into memory - stream or use size limits
- ❌ Don't ignore cross-platform path handling - always use filepath package
- ❌ Don't hardcode clipboard commands - detect availability and fallback gracefully
- ❌ Don't skip error handling in TUI - provide user feedback for all errors
- ❌ Don't use sync functions in Bubble Tea Update loops - use tea.Cmd for async ops
- ❌ Don't assume .gitignore exists - handle missing files gracefully
- ❌ Don't forget WSL detection - it needs special clipboard handling
- ❌ Don't skip memory limits - implement safeguards for large codebases
- ❌ Don't use global state - pass configuration through context