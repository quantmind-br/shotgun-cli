# shotgun-cli Architecture

## Package Structure

```
shotgun-cli/
├── cmd/                          # CLI commands (Cobra)
│   ├── root.go                   # Main command, launches TUI wizard
│   ├── context.go                # Context generation commands
│   ├── template.go               # Template management
│   ├── diff.go                   # Diff splitting tools
│   ├── config.go                 # Configuration management
│   ├── completion.go             # Shell completion
│   └── send.go                   # Gemini integration
│
├── internal/
│   ├── core/                     # Core business logic
│   │   ├── scanner/              # File system scanning
│   │   ├── context/              # Context generation
│   │   ├── template/             # Template management
│   │   ├── ignore/               # Gitignore pattern matching
│   │   └── tokens/               # Token estimation
│   │
│   ├── ui/                       # TUI components (Bubble Tea)
│   │   ├── wizard.go             # Main wizard orchestration (WizardModel)
│   │   ├── screens/              # Individual wizard steps
│   │   ├── components/           # Reusable UI components
│   │   └── styles/               # Theme and styling (Lip Gloss)
│   │
│   ├── platform/                 # Platform-specific code
│   │   ├── clipboard/            # Cross-platform clipboard
│   │   └── gemini/               # Gemini API integration
│   │
│   ├── assets/                   # Embedded resources
│   │   ├── embed.go              # go:embed directive
│   │   └── templates/            # Built-in prompt templates
│   │
│   └── utils/                    # Utility functions
│
├── test/
│   ├── e2e/                      # End-to-end tests
│   └── fixtures/                 # Test fixtures (sample-project)
│
└── templates/                    # Template examples
```

## Key Interfaces

### Scanner Interface (`internal/core/scanner/scanner.go`)
```go
type Scanner interface {
    Scan(ctx context.Context, config ScanConfig) (*FileNode, error)
    ScanWithProgress(rootPath string, config *ScanConfig, progress chan<- Progress) (*FileNode, error)
}
```
- `FileNode`: Tree structure representing scanned files
- `ScanConfig`: Configuration for scanning (root path, filters)
- `FilesystemScanner`: Main implementation
- `Progress`: Progress updates with stages: "counting" → "scanning" → "complete"

**Scanning Flow:**
1. `countItems()` - First pass to count total files (sends "counting" stage)
2. `walkAndBuild()` - Second pass building tree (sends "scanning" stage every 100 items)
3. Completion - Sends "complete" stage

### Context Generator (`internal/core/context/generator.go`)
```go
type ContextGenerator interface {
    Generate(files []*scanner.FileNode, config GenerateConfig) (*ContextData, error)
    GenerateWithProgress(files []*scanner.FileNode, config GenerateConfig, progress chan<- GenProgress) (*ContextData, error)
}
```
- `DefaultContextGenerator`: Main implementation
- `ContextData`: Generated context with tree and content
- `GenerateConfig`: Size limits, include tree/summary flags

### Template Manager (`internal/core/template/manager.go`)
```go
type TemplateManager interface {
    ListTemplates() []*Template
    GetTemplate(name string) (*Template, error)
    RenderTemplate(name string, vars map[string]string) (string, error)
}
```
- Multi-source loading: custom path > user directory > embedded
- Template validation with required variable checking

## UI Architecture (Bubble Tea)

### Wizard Model (`internal/ui/wizard.go`)
- `WizardModel`: Main orchestration model
- Manages 5 steps: FileSelection, TemplateSelection, TaskInput, RulesInput, Review
- Handles async operations via Bubble Tea commands

### Screen Models (`internal/ui/screens/`)
- `FileSelectionModel`: File tree with selection, filtering
- `TemplateSelectionModel`: Template picker
- `TaskInputModel`: Task description input
- `RulesInputModel`: Custom rules input
- `ReviewModel`: Preview and generation

### Components (`internal/ui/components/`)
- `FileTreeModel`: Interactive file tree with expand/collapse, selection
- `ProgressModel`: Progress indicator

## Data Flow

1. **Scanning**: `FilesystemScanner.Scan()` → `FileNode` tree
2. **Selection**: User selects files in TUI → selected `FileNode` list
3. **Generation**: `ContextGenerator.Generate()` → `ContextData`
4. **Rendering**: Template + variables → final output
5. **Output**: Clipboard copy or file write

## Template Priority
1. Custom path (set via `template.custom-path` config)
2. User directory: `~/.config/shotgun-cli/templates/`
3. Embedded templates (fallback)
