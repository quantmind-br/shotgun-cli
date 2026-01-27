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
│   ├── llm.go                    # LLM provider management
│   ├── providers.go              # Provider registry initialization
│   ├── config_llm.go             # LLM configuration helpers
│   ├── completion.go             # Shell completion
│   ├── send.go                   # Send context to LLM
│   └── *_test.go                 # CLI command tests
│
├── internal/
│   ├── app/                      # Application Layer
│   │   ├── service.go            # ContextService - main orchestration
│   │   ├── context.go            # Service types (GenerateConfig, GenerateResult, LLMSendConfig)
│   │   ├── config.go             # CLI config types (CLIConfig, ProgressMode)
│   │   ├── providers.go          # DefaultProviderRegistry initialization
│   │   └── *_test.go
│   │
│   ├── config/                   # Configuration Management
│   │   ├── keys.go               # Configuration key constants
│   │   ├── validator.go          # Config validation logic
│   │   └── metadata.go           # Config metadata (types, categories, descriptions)
│   │
│   ├── core/                     # Core business logic
│   │   ├── scanner/              # File system scanning
│   │   ├── context/              # Context generation
│   │   ├── template/             # Template management
│   │   ├── ignore/               # Gitignore pattern matching
│   │   ├── tokens/               # Token estimation
│   │   ├── llm/                  # LLM provider abstraction
│   │   │   ├── provider.go       # Provider interface
│   │   │   ├── config.go         # LLM config types
│   │   │   └── registry.go       # Provider registry
│   │   └── diff/                 # Diff splitting utilities
│   │
│   ├── ui/                       # TUI components (Bubble Tea)
│   │   ├── wizard.go             # Main wizard orchestration (WizardModel)
│   │   ├── config_wizard.go      # Configuration TUI
│   │   ├── scan_coordinator.go   # File scanning state machine
│   │   ├── generate_coordinator.go # Context generation state machine
│   │   ├── screens/              # Individual wizard steps
│   │   ├── components/           # Reusable UI components
│   │   └── styles/               # Theme and styling (Lip Gloss)
│   │
│   ├── platform/                 # Infrastructure implementations
│   │   ├── openai/               # OpenAI provider implementation
│   │   ├── anthropic/            # Anthropic provider implementation
│   │   ├── geminiapi/            # Gemini API provider
│   │   ├── llmbase/              # Shared LLM base client (Strategy pattern)
│   │   │   ├── base_client.go    # BaseClient with common Provider logic
│   │   │   └── sender.go         # Sender interface for provider-specific logic
│   │   ├── http/                 # Shared HTTP client (JSONClient)
│   │   └── clipboard/            # Cross-platform clipboard
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
- Provides high-level API: `Generate()`, `GenerateWithProgress()`, `SendToLLM()`, `SendToLLMWithProgress()`
- Uses functional options pattern for dependency injection
- **DefaultProviderRegistry**: Unified registry for creating LLM providers

### 2. Core/Domain Layer (`internal/core/`)
Pure business logic with no external dependencies:
- **scanner**: File system scanning with ignore pattern matching
- **context**: Context generation from file trees
- **template**: Template loading and rendering
- **llm**: LLM provider abstraction interface and registry
- **tokens**: Token estimation utilities
- **ignore**: Gitignore-style pattern engine
- **diff**: Diff splitting utilities

### 3. Infrastructure/Platform Layer (`internal/platform/`)
External system integrations:
- **openai**: OpenAI API client
- **anthropic**: Anthropic/Claude API client
- **geminiapi**: Google Gemini API client
- **llmbase**: Shared base client implementation
- **http**: Shared HTTP/JSON client
- **clipboard**: Cross-platform clipboard operations

### 4. Configuration Layer (`internal/config/`)
- Centralized configuration key constants
- Validation logic for all config values
- Config metadata with types, categories, and descriptions

### 5. Presentation Layer (`cmd/`, `internal/ui/`)
- CLI command definitions (Cobra)
- TUI wizard (Bubble Tea)
- User input handling

## Key Interfaces

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

### LLM Base Client (`internal/platform/llmbase/`)
A shared base implementation for HTTP-based providers that uses the Strategy pattern.

- **BaseClient** (`base_client.go`): Implements common logic for `llm.Provider` interface
- **Sender** (`sender.go`): Interface for provider-specific logic

```go
type Sender interface {
    BuildRequest(content string) (interface{}, error)
    ParseResponse(response interface{}, rawJSON []byte) (*llm.Result, error)
    GetEndpoint() string
    GetHeaders() map[string]string
    NewResponse() interface{}
    GetProviderName() string
}
```

### Context Service (`internal/app/context.go`)
```go
type ContextService interface {
    Generate(ctx context.Context, cfg GenerateConfig) (*GenerateResult, error)
    GenerateWithProgress(ctx context.Context, cfg GenerateConfig, progress ProgressCallback) (*GenerateResult, error)
    SendToLLM(ctx context.Context, content string, provider llm.Provider) (*llm.Result, error)
    SendToLLMWithProgress(ctx context.Context, content string, cfg LLMSendConfig, progress LLMProgressCallback) (*llm.Result, error)
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
// In internal/app/providers.go
DefaultProviderRegistry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
    return openai.NewClient(cfg)
})
```

Supported providers (3):
- `ProviderOpenAI`: OpenAI API (GPT-4o, GPT-4, o1, o3)
- `ProviderAnthropic`: Anthropic API (Claude 4, Claude 3.5)
- `ProviderGemini`: Google Gemini API

## TUI Architecture

The TUI Wizard uses the "Model of Models" pattern with dedicated coordinators.

### Coordinator Responsibilities

- **ScanCoordinator**: Manages file system scanning, progress updates, and results
- **GenerateCoordinator**: Manages context generation, progress updates, and content buffering

Both coordinators follow the Bubble Tea command pattern:
1. `Start()` returns a command to begin async work
2. `Poll()` returns commands to check status
3. `Result()` provides access to completed data
