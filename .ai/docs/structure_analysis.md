
# Code Structure Analysis
## Architectural Overview
Shotgun CLI is a sophisticated Command-Line Interface tool written in Go that implements a **Clean Architecture/Hexagonal Architecture** pattern. The system functions as a **Context Generation and AI Orchestration Engine** designed to bridge complex codebases with AI language models. The architecture follows clear separation of concerns across three main layers:

1. **Presentation/Adapter Layer** (`cmd`, `internal/ui`): Handles user interaction through CLI commands and interactive TUI wizard using Cobra and Bubble Tea frameworks
2. **Core/Domain Layer** (`internal/core`): Contains pure business logic for context generation, file scanning, template management, and token estimation
3. **Infrastructure/Platform Layer** (`internal/platform`): Implements external system integrations (Gemini AI API, clipboard operations)

The tool provides both interactive (TUI) and programmatic (CLI) interfaces for generating LLM-optimized codebase contexts and facilitating AI-assisted development workflows.

## Core Components
### `internal/core/context`
- **Purpose**: Aggregates all necessary information into structured AI context payloads
- **Key Files**: `generator.go` (main context builder), `tree.go` (structure representation), `content.go` (file content handling), `template.go` (template integration)
- **Responsibilities**: File tree generation, content collection with size limits, template variable management, context optimization

### `internal/core/scanner`
- **Purpose**: Provides intelligent file system traversal capabilities with layered filtering
- **Key Files**: `scanner.go` (interface and types), `filesystem.go` (implementation)
- **Responsibilities**: Recursive directory scanning, file metadata collection, ignore rule application, progress reporting

### `internal/core/template`
- **Purpose**: Manages AI prompt template lifecycle and rendering with multi-source support
- **Key Files**: `manager.go` (template management), `loader.go` (template loading), `renderer.go` (template rendering), `template.go` (template structures)
- **Responsibilities**: Template discovery from multiple sources, validation, variable substitution, priority-based loading

### `internal/core/ignore`
- **Purpose**: Implements sophisticated layered ignore rule processing
- **Key Files**: `engine.go` (ignore engine implementation)
- **Responsibilities**: Gitignore parsing, custom pattern matching, explicit include/exclude handling, built-in pattern management

### `internal/core/tokens`
- **Purpose**: Provides token estimation for LLM context optimization
- **Key Files**: `estimator.go` (token calculation utilities)
- **Responsibilities**: Token count estimation using heuristics, context window validation, size formatting

## Service Definitions
### `platform/gemini` (AI Execution Service)
- **Responsibility**: Handles Gemini AI API integration through external tool execution, request serialization, response parsing
- **Key Files**: `executor.go` (request execution), `config.go` (configuration management), `parser.go` (response processing)
- **Interface**: Abstracts AI provider interactions for potential multi-provider support
- **Pattern**: Strategy pattern implementation for AI provider abstraction

### `internal/ui` (TUI Service)
- **Responsibility**: Provides interactive terminal interface for complex user inputs using MVU pattern
- **Key Files**: `wizard.go` (main TUI orchestrator), `screens/` (individual screens), `components/` (reusable UI elements), `styles/` (theming)
- **Framework**: Built on Bubble Tea MVU pattern with message-driven state management

### `cmd` (Command Orchestration Services)
- **Responsibility**: CLI command handling and application flow orchestration
- **Key Files**: `root.go` (main entry point), `send.go` (AI interaction), `diff.go` (diff generation), `template.go` (template commands), `context.go` (context management)
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
- **Implementation**: Each command file (`send.go`, `diff.go`, `template.go`, `context.go`) implements specific command handlers dispatched by `root.go`

### 2. Builder/Generator Pattern
- **Location**: `internal/core/context/generator.go`
- **Implementation**: Complex context objects built through step-by-step assembly of file structure, content, and template variables

### 3. Strategy Pattern
- **Location**: `internal/platform/gemini` and implied multi-provider support
- **Implementation**: AI execution abstracted through interfaces, allowing different AI providers to be swapped

### 4. MVU Pattern (Model-View-Update)
- **Location**: `internal/ui` with Bubble Tea framework
- **Implementation**: TUI state management through message-driven updates with clear separation of state and rendering

### 5. Template Method Pattern
- **Location**: `internal/core/template/renderer.go`
- **Implementation**: Template rendering follows standardized process with variable substitution hooks

### 6. Layered Architecture Pattern
- **Location**: Overall package structure
- **Implementation**: Clear separation between presentation, business logic, and infrastructure layers with dependency inversion

### 7. Factory Pattern
- **Location**: `internal/core/template/manager.go`
- **Implementation**: Template manager creates and manages multiple template sources with priority-based loading

## Component Relationships
| Component A | Relationship | Component B | Description |
|-------------|--------------|-------------|-------------|
| `cmd` | Orchestrates | `internal/core/context` | Initiates context generation process through command execution |
| `cmd` | Depends on | `internal/platform/gemini` | Sends prepared requests to AI service via executor |
| `cmd` | Uses | `internal/ui` | Invokes TUI screens for interactive input in wizard mode |
| `core/context` | Aggregates from | `core/scanner` & `core/template` | Combines file data and template formatting for context generation |
| `core/scanner` | Uses | `core/ignore` | Filters files based on layered ignore rules |
| `core/context` | Consults | `core/tokens` | Validates context size before execution |
| `platform/gemini` | Communicates with | External API | Handles network communication with Gemini through external tool |
| `internal/ui` | Renders | `internal/ui/components` | Uses reusable UI building blocks for consistent interface |
| `core/template` | Loads from | Multiple Sources | Embedded, user config, and custom template directories |

## Key Methods & Functions
### `cmd/root.go`
- `Execute()`: Main CLI entry point, parses flags and dispatches commands
- `launchTUIWizard()`: Initializes and runs interactive TUI mode with scanner configuration
- `initConfig()`: Configuration loading and validation with platform-specific paths

### `internal/core/context/generator.go`
- `Generate()`: Primary context generation method
- `GenerateWithProgressEx()`: Progress-enabled context generation with structured progress reporting
- `collectFileContents()`: File content collection with size limits and binary file filtering
- `buildCompleteFileStructure()`: Combines tree structure with file content blocks

### `internal/core/scanner/scanner.go`
- `Scan()`: Basic file system scanning without progress
- `ScanWithProgress()`: Progress-enabled scanning with real-time updates
- `DefaultScanConfig()`: Creates default scanning configuration with sensible defaults

### `internal/core/template/manager.go`
- `NewManager()`: Template manager initialization with multi-source loading
- `loadFromSources()`: Multi-source template loading with priority override system
- `RenderTemplate()`: Template rendering with variable substitution and validation

### `internal/platform/gemini/executor.go`
- `Send()`: Sends content to Gemini API with timeout and error handling
- `SendWithProgress()`: Progress-enabled API communication with stage reporting
- `buildArgs()`: Command-line argument construction for geminiweb tool

### `internal/ui/wizard.go`
- `Init()`: TUI initialization with scan command startup
- `Update()`: Message-driven state updates handling various UI events
- `handleKeyPress()`: User input processing with keyboard navigation

### `internal/core/ignore/engine.go`
- `ShouldIgnore()`: Layered ignore rule evaluation with priority system
- `LoadGitignore()`: Gitignore file parsing and compilation
- `AddCustomRule()`: Dynamic pattern addition for runtime configuration

### `internal/core/tokens/estimator.go`
- `Estimate()`: Token count approximation using byte-based heuristics
- `CheckContextFit()`: Context window validation with percentage calculation
- `FormatTokens()`: Human-readable token formatting with K/M suffixes

## Available Documentation
### High-Quality Documentation
| Document Path | Content Type | Quality Evaluation |
|---------------|--------------|-------------------|
| `README.md` | Project Overview | **Excellent** - Comprehensive project overview with architecture, features, and usage examples |
| `CLAUDE.md`, `GEMINI.md` | AI Model Guides | **High Quality** - Dedicated guides for specific AI provider configuration and setup |
| `.ai/docs/structure_analysis.md` | Architecture Analysis | **Comprehensive** - Detailed structural analysis with component mapping and relationships |
| `.ai/docs/api_analysis.md` | API Documentation | **Thorough** - Complete CLI command documentation and external API integration details |
| `.ai/docs/data_flow_analysis.md` | Data Flow Analysis | **Detailed** - Complete data transformation pipeline documentation |
| `.ai/docs/dependency_analysis.md` | Dependency Analysis | **Comprehensive** - Thorough dependency mapping and coupling assessment |
| `.ai/docs/request_flow_analysis.md` | Request Flow Analysis | **Complete** - Full command execution flow documentation |

### Workflow Documentation
| Document Path | Content Type | Quality Evaluation |
|---------------|--------------|-------------------|
| `.claude/commands/` | Workflow Definitions | **Excellent** - Extensive collection of markdown files defining application capabilities and operational logic |
| `.claude/commands/prp-*` | Project Management | **Well-Structured** - Comprehensive project planning and execution workflows |
| `.claude/commands/development/` | Development Workflows | **Comprehensive** - Detailed development process documentation |
| `templates/` | Prompt Templates | **Practical** - Concrete examples of tool usage patterns and expected outputs |

### Configuration and CI/CD
| Document Path | Content Type | Quality Evaluation |
|---------------|--------------|-------------------|
| `.github/workflows/` | CI/CD Configuration | **Professional** - Complete GitHub Actions workflows for testing, releases, and code review |
| `go.mod` | Dependency Management | **Standard** - Well-organized Go module with appropriate dependencies |
| `Makefile` | Build Configuration | **Comprehensive** - Build targets for development, testing, and deployment |

The documentation quality is exceptionally high, with comprehensive coverage of architecture, workflows, and operational procedures. The `.ai/docs/` directory contains particularly valuable technical documentation that provides deep insights into the system's design and implementation.