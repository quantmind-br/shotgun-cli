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
│   ├── llm.go                    # LLM provider management (NEW)
│   ├── providers.go              # Provider registry initialization (NEW)
│   ├── config_llm.go             # LLM configuration helpers (NEW)
│   ├── completion.go             # Shell completion
│   ├── send.go                   # Send context to LLM
│   ├── gemini.go                 # Legacy Gemini web integration
│   └── *_test.go                 # CLI command tests
│
├── internal/
│   ├── app/                      # Application Layer (NEW)
│   │   ├── service.go            # ContextService - main orchestration
│   │   ├── context.go            # Service types (GenerateConfig, GenerateResult)
│   │   ├── providers.go          # Provider initialization
│   │   └── *_test.go
│   │
│   ├── config/                   # Configuration Management (NEW)
│   │   ├── keys.go               # Configuration key constants
│   │   └── validator.go          # Config validation logic
│   │
│   ├── core/                     # Core business logic
│   │   ├── scanner/              # File system scanning
│   │   ├── context/              # Context generation
│   │   ├── template/             # Template management
│   │   ├── ignore/               # Gitignore pattern matching
│   │   ├── tokens/               # Token estimation
│   │   ├── llm/                  # LLM provider abstraction (NEW)
│   │   │   ├── provider.go       # Provider interface
│   │   │   ├── config.go         # LLM config types
│   │   │   └── registry.go       # Provider registry
│   │   └── diff/                 # Diff splitting utilities
│   │
│   ├── ui/                       # TUI components (Bubble Tea)
│   │   ├── wizard.go             # Main wizard orchestration (WizardModel)
│   │   ├── screens/              # Individual wizard steps
│   │   ├── components/           # Reusable UI components
│   │   └── styles/               # Theme and styling (Lip Gloss)
│   │
│   ├── platform/                 # Platform-specific code
│   │   ├── clipboard/            # Cross-platform clipboard
│   │   ├── openai/               # OpenAI provider implementation (NEW)
│   │   ├── anthropic/            # Anthropic provider implementation (NEW)
│   │   ├── geminiapi/            # Gemini API provider (NEW)
│   │   └── gemini/               # Gemini web integration (legacy)
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

## Layered Architecture

The application follows Clean Architecture/Hexagonal Architecture principles:

### 1. Application Layer (`internal/app/`)
- **ContextService**: Main orchestration service that coordinates scanning, generation, and LLM operations
- Provides high-level API: `Generate()`, `GenerateWithProgress()`, `SendToLLM()`
- Uses functional options pattern for dependency injection
- Located in: `internal/app/service.go`

### 2. Core/Domain Layer (`internal/core/`)
Pure business logic with no external dependencies:
- **scanner**: File system scanning with ignore pattern matching
- **context**: Context generation from file trees
- **template**: Template loading and rendering
- **llm**: LLM provider abstraction interface
- **tokens**: Token estimation utilities
- **ignore**: Gitignore-style pattern engine
- **diff**: Diff splitting utilities

### 3. Infrastructure/Platform Layer (`internal/platform/`, `cmd/`)
External system integrations:
- **openai**: OpenAI API client
- **anthropic**: Anthropic/Claude API client
- **geminiapi**: Google Gemini API client
- **gemini**: Browser-based Gemini integration
- **clipboard**: Cross-platform clipboard operations

### 4. Configuration Layer (`internal/config/`)
- Centralized configuration key constants
- Validation logic

### 5. Presentation Layer (`cmd/`, `internal/ui/`)
- CLI command definitions (Cobra)
- TUI wizard (Bubble Tea)
- User input handling

## Key Interfaces

### Scanner Interface (`internal/core/scanner/scanner.go`)
```go
type Scanner interface {
    Scan(ctx context.Context, config ScanConfig) (*FileNode, error)
    ScanWithProgress(rootPath string, config *ScanConfig, progress chan<- Progress) (*FileNode, error)
}
```

### Context Generator (`internal/core/context/generator.go`)
```go
type ContextGenerator interface {
    Generate(files []*scanner.FileNode, config GenerateConfig) (*ContextData, error)
    GenerateWithProgress(files []*scanner.FileNode, config GenerateConfig, progress chan<- GenProgress) (*ContextData, error)
}
```

### LLM Provider (`internal/core/llm/provider.go`)
```go
type Provider interface {
    Send(ctx context.Context, content string) (*Result, error)
    SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*Result, error)
    Name() string
    IsAvailable() bool
    IsConfigured() bool
    ValidateConfig() error
}
```

### Context Service (`internal/app/service.go`)
```go
type ContextService interface {
    Generate(ctx context.Context, cfg GenerateConfig) (*GenerateResult, error)
    GenerateWithProgress(ctx context.Context, cfg GenerateConfig, progress ProgressCallback) (*GenerateResult, error)
    SendToLLM(ctx context.Context, content string, provider llm.Provider) (*llm.Result, error)
}
```

## Data Flow

1. **CLI Entry Point** → `cmd/*.go` (Cobra commands)
2. **Service Orchestration** → `internal/app/service.go` (ContextService)
3. **Scanning** → `FilesystemScanner.Scan()` → `FileNode` tree
4. **Selection** → User selects files in TUI → selected `FileNode` list
5. **Generation** → `ContextGenerator.Generate()` → `ContextData`
6. **Rendering** → Template + variables → final output
7. **LLM Integration** → Provider.Send() → AI response
8. **Output** → Clipboard copy or file write

## Provider Registry Pattern

The LLM provider system uses a registry pattern for extensibility:

```go
// In cmd/providers.go
providerRegistry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
    return openai.NewClient(cfg)
})
```

Supported providers:
- `ProviderOpenAI`: OpenAI API (GPT-4o, GPT-4, o1, o3)
- `ProviderAnthropic`: Anthropic API (Claude 4, Claude 3.5)
- `ProviderGemini`: Google Gemini API
- `ProviderGeminiWeb`: Browser-based Gemini integration


## TUI Architecture

The TUI Wizard uses the "Model of Models" pattern with dedicated coordinators for asynchronous operations.

### Component Structure

```
internal/ui/
├── wizard.go             # Main orchestrator (WizardModel)
├── scan_coordinator.go   # File scanning state machine
├── generate_coordinator.go # Context generation state machine
├── screens/              # Individual wizard steps
│   ├── file_selection.go
│   ├── template_selection.go
│   ├── task_input.go
│   ├── rules_input.go
│   └── review.go
└── components/           # Reusable UI components
    ├── progress.go
    └── tree.go
```

### Coordinator Responsibilities

- **ScanCoordinator**: Manages file system scanning, progress updates, and results
- **GenerateCoordinator**: Manages context generation, progress updates, and content buffering

Both coordinators follow the Bubble Tea command pattern:
1. `Start()` returns a command to begin async work
2. `Poll()` returns commands to check status
3. `Result()` provides access to completed data