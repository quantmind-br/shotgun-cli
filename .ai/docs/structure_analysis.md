
# Code Structure Analysis
## Architectural Overview
The codebase implements a sophisticated Command-Line Interface (CLI) tool written in Go, designed to generate LLM-optimized codebase contexts and facilitate AI-assisted development workflows. The architecture follows **Clean Architecture/Hexagonal Architecture** principles with clear separation of concerns:

1. **Presentation/Adapter Layer** (`cmd`, `internal/ui`): Handles user interaction through CLI commands and interactive TUI wizard
2. **Core/Domain Layer** (`internal/core`): Contains business logic for context generation, file scanning, template management, and token estimation
3. **Infrastructure/Platform Layer** (`internal/platform`): Implements external system integrations (Gemini AI API, clipboard)

The system functions as a **Context Generation and AI Orchestration Engine** that gathers comprehensive codebase context and dispatches it to AI models for processing.

## Core Components
### `internal/core/context`
- **Purpose**: Aggregates all necessary information into structured AI context payloads
- **Key Files**: `generator.go` (main context builder), `tree.go` (structure representation), `content.go` (file content handling)
- **Responsibilities**: File tree generation, content collection, template variable management

### `internal/core/scanner`
- **Purpose**: Provides file system traversal capabilities with intelligent filtering
- **Key Files**: `scanner.go` (interface and types), `filesystem.go` (implementation)
- **Responsibilities**: Recursive directory scanning, file metadata collection, ignore rule application

### `internal/core/template`
- **Purpose**: Manages AI prompt template lifecycle and rendering
- **Key Files**: `manager.go` (template management), `loader.go` (template loading), `renderer.go` (template rendering)
- **Responsibilities**: Template discovery, validation, variable substitution, multi-source loading

### `internal/core/ignore`
- **Purpose**: Implements layered ignore rule processing
- **Key Files**: `engine.go` (ignore engine implementation)
- **Responsibilities**: Gitignore parsing, custom pattern matching, explicit include/exclude handling

### `internal/core/tokens`
- **Purpose**: Provides token estimation for LLM context optimization
- **Key Files**: `estimator.go` (token calculation utilities)
- **Responsibilities**: Token count estimation, context window validation, size formatting

## Service Definitions
### `platform/gemini` (AI Execution Service)
- **Responsibility**: Handles Gemini AI API integration, request serialization, response parsing
- **Key Files**: `executor.go` (request execution), `config.go` (configuration management), `parser.go` (response processing)
- **Interface**: Abstracts AI provider interactions for potential multi-provider support

### `internal/ui` (TUI Service)
- **Responsibility**: Provides interactive terminal interface for complex user inputs
- **Key Files**: `wizard.go` (main TUI orchestrator), `screens/` (individual screens), `components/` (reusable UI elements)
- **Framework**: Built on Bubble Tea MVU pattern

### `cmd` (Command Orchestration Services)
- **Responsibility**: CLI command handling and application flow orchestration
- **Key Files**: `root.go` (main entry point), `send.go` (AI interaction), `diff.go` (diff generation), `template.go` (template commands)
- **Pattern**: Command pattern implementation with Cobra framework

### `platform/clipboard` (Utility Service)
- **Responsibility**: Cross-platform clipboard access for generated output
- **Key Files**: `clipboard.go` (clipboard operations)

## Interface Contracts
### `Scanner` Interface
```go
type Scanner interface {
    Scan(rootPath string, config *ScanConfig) (*FileNode, error)
    ScanWithProgress(rootPath string, config *ScanConfig, progress chan<- Progress) (*FileNode, error)
}
```

### `ContextGenerator` Interface
```go
type ContextGenerator interface {
    Generate(root *scanner.FileNode, selections map[string]bool, config GenerateConfig) (string, error)
    GenerateWithProgress(root *scanner.FileNode, selections map[string]bool, config GenerateConfig, progress func(string)) (string, error)
    GenerateWithProgressEx(root *scanner.FileNode, selections map[string]bool, config GenerateConfig, progress func(GenProgress)) (string, error)
}
```

### `TemplateManager` Interface
```go
type TemplateManager interface {
    ListTemplates() ([]Template, error)
    GetTemplate(name string) (*Template, error)
    RenderTemplate(name string, vars map[string]string) (string, error)
    ValidateTemplate(name string) error
    GetRequiredVariables(name string) ([]string, error)
}
```

### `IgnoreEngine` Interface
```go
type IgnoreEngine interface {
    ShouldIgnore(relPath string) (bool, IgnoreReason)
    LoadGitignore(rootDir string) error
    AddCustomRule(pattern string) error
    IsGitignored(relPath string) bool
    LoadShotgunignore(rootDir string) error
    AddExplicitExclude(pattern string) error
    AddExplicitInclude(pattern string) error
}
```

## Design Patterns Identified
### 1. Command Pattern
- **Location**: `cmd` package structure
- **Implementation**: Each command file (`send.go`, `diff.go`, `template.go`) implements specific command handlers dispatched by `root.go`

### 2. Builder/Generator Pattern
- **Location**: `internal/core/context/generator.go`
- **Implementation**: Complex context objects built through step-by-step assembly of file structure, content, and template variables

### 3. Strategy Pattern
- **Location**: `internal/platform/gemini` and implied multi-provider support
- **Implementation**: AI execution abstracted through interfaces, allowing different AI providers to be swapped

### 4. MVU Pattern (Model-View-Update)
- **Location**: `internal/ui` with Bubble Tea framework
- **Implementation**: TUI state management through message-driven updates

### 5. Template Method Pattern
- **Location**: `internal/core/template/renderer.go`
- **Implementation**: Template rendering follows standardized process with variable substitution hooks

### 6. Layered Architecture Pattern
- **Location**: Overall package structure
- **Implementation**: Clear separation between presentation, business logic, and infrastructure layers

## Component Relationships
| Component A | Relationship | Component B | Description |
|-------------|--------------|-------------|-------------|
| `cmd` | Orchestrates | `internal/core/context` | Initiates context generation process |
| `cmd` | Depends on | `internal/platform/gemini` | Sends prepared requests to AI service |
| `cmd` | Uses | `internal/ui` | Invokes TUI screens for interactive input |
| `core/context` | Aggregates from | `core/scanner` & `core/template` | Combines file data and template formatting |
| `core/scanner` | Uses | `core/ignore` | Filters files based on ignore rules |
| `core/context` | Consults | `core/tokens` | Validates context size before execution |
| `platform/gemini` | Communicates with | External API | Handles network communication with Gemini |
| `internal/ui` | Renders | `internal/ui/components` | Uses reusable UI building blocks |

## Key Methods & Functions
### `cmd/root.go`
- `Execute()`: Main CLI entry point, parses flags and dispatches commands
- `launchTUIWizard()`: Initializes and runs interactive TUI mode
- `initConfig()`: Configuration loading and validation

### `internal/core/context/generator.go`
- `Generate()`: Primary context generation method
- `GenerateWithProgressEx()`: Progress-enabled context generation
- `collectFileContents()`: File content collection with size limits
- `buildCompleteFileStructure()`: Combines tree structure with file content blocks

### `internal/core/scanner/scanner.go`
- `Scan()`: Basic file system scanning
- `ScanWithProgress()`: Progress-enabled scanning
- `DefaultScanConfig()`: Creates default scanning configuration

### `internal/core/template/manager.go`
- `NewManager()`: Template manager initialization
- `loadFromSources()`: Multi-source template loading with priority
- `RenderTemplate()`: Template rendering with variable substitution

### `internal/platform/gemini/executor.go`
- `Send()`: Sends content to Gemini API
- `SendWithProgress()`: Progress-enabled API communication
- `buildArgs()`: Command-line argument construction

### `internal/ui/wizard.go`
- `Init()`: TUI initialization
- `Update()`: Message-driven state updates
- `handleKeyPress()`: User input processing

### `internal/core/ignore/engine.go`
- `ShouldIgnore()`: Layered ignore rule evaluation
- `LoadGitignore()`: Gitignore file parsing
- `AddCustomRule()`: Dynamic pattern addition

### `internal/core/tokens/estimator.go`
- `Estimate()`: Token count approximation
- `CheckContextFit()`: Context window validation
- `FormatTokens()`: Human-readable token formatting

## Available Documentation
### High-Quality Documentation
| Document Path | Content Type | Quality Evaluation |
|---------------|--------------|-------------------|
| `README.md` | Project Overview | Essential starting point for understanding tool purpose and setup |
| `CLAUDE.md`, `GEMINI.md` | AI Model Guides | High-quality dedicated guides for specific AI provider configuration |
| `.ai/docs/structure_analysis.md` | Architecture Analysis | Comprehensive structural analysis with component mapping |
| `.ai/docs/api_analysis.md` | API Documentation | Detailed CLI command documentation and external API integration |
| `.ai/docs/data_flow_analysis.md` | Data Flow Analysis | Complete data transformation pipeline documentation |
| `.ai/docs/dependency_analysis.md` | Dependency Analysis | Thorough dependency mapping and coupling assessment |
| `.ai/docs/request_flow_analysis.md` | Request Flow Analysis | Complete command execution flow documentation |

### Workflow Documentation
| Document Path | Content Type | Quality Evaluation |
|---------------|--------------|-------------------|
| `.claude/commands/` | Workflow Definitions | **Excellent** - Extensive collection of markdown files defining application capabilities and operational logic |
| `.claude/commands/prp-*` | Project Management | Well-structured project planning and execution workflows |
| `.claude/commands/development/` | Development Workflows | Comprehensive development process documentation |
| `templates/` | Prompt Templates | Concrete examples of tool usage patterns and expected outputs |

### Configuration and CI/CD
| Document Path | Content Type | Quality Evaluation |
|---------------|--------------|-------------------|
| `.github/workflows/*.yml` | CI/CD Automation | Documents automated processes including code review integration |
| `go.mod` | Dependencies | Clear dependency specification with version pinning |

**Overall Documentation Quality**: The repository exhibits exceptional documentation quality, particularly in architectural analysis and workflow definitions. The `.ai/docs/` directory provides comprehensive technical documentation, while `.claude/commands/` externalizes operational logic into well-documented prompt templates. This dual approach of technical documentation and workflow externalization represents a sophisticated architectural pattern for AI-assisted development tools.