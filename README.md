# Shotgun CLI

## Project Overview

**Shotgun CLI** is a sophisticated Command-Line Interface tool written in Go that serves as a Context Generation and AI Orchestration Engine. The tool is designed to generate LLM-optimized codebase contexts and facilitate AI-assisted development workflows by gathering comprehensive codebase context and dispatching it to AI models for processing.

### Purpose and Main Functionality

The primary purpose of Shotgun CLI is to bridge the gap between complex codebases and AI language models by:

- Scanning and analyzing project structures with intelligent filtering
- Generating optimized context files within LLM token limits
- Providing both interactive (TUI) and programmatic (CLI) interfaces
- Integrating with Google Gemini API for AI-powered code analysis
- Managing templates for different AI interaction patterns

### Key Features and Capabilities

- **Intelligent File Scanning**: Recursive directory traversal with layered ignore rules (.gitignore, .shotgunignore, custom patterns)
- **Interactive TUI Wizard**: 5-step guided workflow for complex configurations using Bubble Tea framework
- **Template Management**: Multi-source template loading with variable substitution and validation
- **Token Optimization**: Built-in token estimation and context window validation
- **AI Integration**: Seamless integration with Google Gemini API via external tool execution
- **Cross-platform Support**: Clipboard operations and filesystem access across different platforms
- **Progress Reporting**: Real-time progress updates for long-running operations

### Likely Intended Use Cases

- **Code Review Automation**: Generate comprehensive context for AI-powered code reviews
- **Documentation Generation**: Create AI-assisted documentation from codebase analysis
- **Refactoring Assistance**: Provide AI models with complete project context for refactoring suggestions
- **Onboarding Tools**: Help new developers understand complex codebases through AI analysis
- **Code Migration**: Assist in migrating code between different frameworks or languages

## Table of Contents

- [Architecture](#architecture)
- [C4 Model Architecture](#c4-model-architecture)
- [Repository Structure](#repository-structure)
- [Dependencies and Integration](#dependencies-and-integration)
- [API Documentation](#api-documentation)
- [Development Notes](#development-notes)
- [Known Issues and Limitations](#known-issues-and-limitations)
- [Additional Documentation](#additional-documentation)

## Architecture

### High-level Architecture Overview

Shotgun CLI follows **Clean Architecture/Hexagonal Architecture** principles with clear separation of concerns across three main layers:

1. **Presentation/Adapter Layer** (`cmd`, `internal/ui`): Handles user interaction through CLI commands and interactive TUI wizard
2. **Core/Domain Layer** (`internal/core`): Contains business logic for context generation, file scanning, template management, and token estimation
3. **Infrastructure/Platform Layer** (`internal/platform`): Implements external system integrations (Gemini AI API, clipboard)

The architecture emphasizes dependency inversion, with core business logic remaining independent of external frameworks and services.

### Technology Stack and Frameworks

- **Language**: Go 1.21+
- **CLI Framework**: Cobra (github.com/spf13/cobra) for command structure and argument parsing
- **Configuration**: Viper (github.com/spf13/viper) for configuration management
- **TUI Framework**: Bubble Tea (github.com/charmbracelet/bubbletea) with Bubbles components
- **Logging**: Zerolog (github.com/rs/zerolog) for structured logging
- **Template Engine**: Go's built-in text/template with custom variable syntax
- **Ignore Processing**: go-gitignore (github.com/sabhiram/go-gitignore) for pattern matching

### Component Relationships

```mermaid
graph TB
    subgraph "Presentation Layer"
        CMD[cmd Package]
        TUI[internal/ui]
    end
    
    subgraph "Core Business Logic"
        Context[internal/core/context]
        Scanner[internal/core/scanner]
        Template[internal/core/template]
        Ignore[internal/core/ignore]
        Tokens[internal/core/tokens]
    end
    
    subgraph "Infrastructure Layer"
        Gemini[internal/platform/gemini]
        Clipboard[internal/platform/clipboard]
    end
    
    subgraph "External Services"
        GeminiAPI[Google Gemini API]
        FileSystem[File System]
        ConfigFiles[Configuration Files]
    end
    
    CMD -. orchestrates .|color:blue| Context
    CMD -. orchestrates .|color:blue| Scanner
    CMD -. orchestrates .|color:blue| Gemini
    CMD -. orchestrates .|color:blue| TUI
    
    TUI -. uses .|color:green| Context
    TUI -. uses .|color:green| Scanner
    TUI -. uses .|color:green| Template
    
    Context -. aggregates from .|color:orange| Scanner
    Context -. uses .|color:orange| Template
    Context -. consults .|color:orange| Tokens
    
    Scanner -. uses .|color:red| Ignore
    Scanner -. reads from .|color:red| FileSystem
    
    Gemini -. communicates with .|color:purple| GeminiAPI
    Clipboard -. accesses .|color:purple| System Clipboard
    
    CMD -. loads .|color:brown| ConfigFiles
```

### Key Design Patterns

1. **Command Pattern**: Implemented in `cmd` package structure with each command file handling specific operations
2. **Builder/Generator Pattern**: Used in `internal/core/context` for complex context object assembly
3. **Strategy Pattern**: Abstracted AI provider interfaces for potential multi-provider support
4. **MVU Pattern** (Model-View-Update): TUI state management through Bubble Tea framework
5. **Template Method Pattern**: Standardized template rendering process with variable substitution
6. **Layered Architecture**: Clear separation between presentation, business logic, and infrastructure

## C4 Model Architecture

### Context Diagram

</arg_value>
</tool_call>