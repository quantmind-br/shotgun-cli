# Shotgun CLI

## Project Overview

**Shotgun CLI** is a sophisticated Command-Line Interface tool written in Go that functions as a **Context Generation and AI Orchestration Engine**. The tool bridges complex codebases with AI language models, providing both interactive (TUI) and programmatic (CLI) interfaces for generating LLM-optimized codebase contexts and facilitating AI-assisted development workflows.

### Purpose and Main Functionality

The primary purpose of Shotgun CLI is to transform complex codebases into structured, LLM-optimized contexts that can be sent to AI models like Google Gemini. It intelligently scans file systems, applies layered ignore patterns, and generates comprehensive context files that include file structure representations and relevant content.

### Key Features and Capabilities

- **Interactive TUI Wizard**: 5-step guided workflow using Bubble Tea framework for intuitive user interaction
- **CLI Commands**: Programmatic interface for automation and scripting
- **Intelligent File Scanning**: Recursive directory traversal with layered ignore rule processing
- **Template Management**: Multi-source template loading with variable substitution
- **Context Optimization**: Token estimation and size validation for LLM compatibility
- **AI Integration**: Seamless integration with Google Gemini API via external tool execution
- **Cross-platform Support**: Works across different operating systems with platform-specific optimizations

### Likely Intended Use Cases

- **Code Review Automation**: Generate comprehensive context for AI-powered code analysis
- **Documentation Generation**: Create structured representations of codebases for AI-assisted documentation
- **Refactoring Assistance**: Provide AI models with complete context for intelligent refactoring suggestions
- **Onboarding Tools**: Help new developers understand complex codebase structures
- **Code Migration**: Facilitate AI-assisted code migration between languages or frameworks

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

Shotgun CLI implements a **Clean Architecture/Hexagonal Architecture** pattern with clear separation of concerns across three main layers:

1. **Presentation/Adapter Layer** (`cmd`, `internal/ui`): Handles user interaction through CLI commands and interactive TUI wizard
2. **Core/Domain Layer** (`internal/core`): Contains pure business logic for context generation, file scanning, template management, and token estimation
3. **Infrastructure/Platform Layer** (`internal/platform`): Implements external system integrations (Gemini AI API, clipboard operations)

### Technology Stack and Frameworks

- **Language**: Go
- **CLI Framework**: Cobra (`github.com/spf13/cobra`)
- **Configuration**: Viper (`github.com/spf13/viper`)
- **TUI Framework**: Bubble Tea (`github.com/charmbracelet/bubbletea`)
- **Logging**: Zerolog (`github.com/rs/zerolog`)
- **Template Engine**: Go standard library templates
- **Ignore Processing**: go-gitignore (`github.com/sabhiram/go-gitignore`)

### Component Relationships

```mermaid
graph TD
    A[main.go] → B[cmd]
    B → C[internal/core/context]
    B → D[internal/core/scanner]
    B → E[internal/core/template]
    B → F[internal/core/ignore]
    B → G[internal/core/tokens]
    B → H[internal/platform/gemini]
    B → I[internal/platform/clipboard]
    B → J[internal/ui/wizard]
    B → K[internal/utils]
    
    C → D
    C → E
    C → F
    C → G
    
    D → F
    
    J → C
    J → D
    J → E
    J → H
    J → I
    J → L[internal/ui/screens]
    J → M[internal/ui/components]
    
    L → M
    L → N[internal/ui/styles]
    
    M → N
    
    style A fill:#e1f5fe
    style B fill:#f3e5f5
    style C fill:#e8f5e8
    style D fill:#e8f5e8
    style E fill:#e8f5e8
    style F fill:#e8f5e8
    style G fill:#e8f5e8
    style H fill:#fff3e0
    style I fill:#fff3e0
    style J fill:#fce4ec
    style K fill:#f5f5f5
```

### Key Design Patterns

- **Command Pattern**: CLI command structure in `cmd` package
- **Builder/Generator Pattern**: Context generation in `internal/core/context`
- **Strategy Pattern**: AI provider abstraction for multi-provider support
- **MVU Pattern**: TUI state management with Bubble Tea
- **Template Method Pattern**: Standardized template rendering process
- **Layered Architecture**: Clear separation between presentation, business logic, and infrastructure
- **Factory Pattern**: Template manager and scanner creation

## C4 Model Architecture

### Context Diagram

</arg_value>
</tool_call>