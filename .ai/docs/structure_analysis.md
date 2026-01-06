# Code Structure Analysis

## Architectural Overview
The `shotgun-cli` repository is a Go-based command-line tool designed to generate LLM-optimized context from codebases. It features a dual-mode architecture:
- **Interactive TUI Mode**: A 5-step wizard built with the `Bubble Tea` (Model-View-Update) framework, providing a user-friendly experience for file selection and context configuration.
- **Headless CLI Mode**: A command-driven interface built using `Cobra` and `Viper` for automated workflows and terminal-based interactions.

The project follows a modular structure where core business logic is decoupled from both the CLI interface and the TUI components through well-defined interfaces. It adheres to standard Go project layouts, with entry points in `cmd/` and internal library logic in `internal/`.

## Core Components
- **Scanner (`internal/core/scanner`)**: Responsible for traversing the filesystem and building a hierarchical tree representation (`FileNode`) of the codebase. It supports concurrent scanning and progress reporting.
- **Ignore Engine (`internal/core/ignore`)**: A layered filtering system that applies exclusion rules from multiple sources (built-in, `.gitignore`, `.shotgunignore`, and custom patterns) using Git-style matching semantics.
- **Context Generator (`internal/core/context`)**: The orchestrator that aggregates selected file contents, directory structures, and metadata into a final prompt string for LLMs.
- **Template Manager (`internal/core/template`)**: Handles the loading, validation, and rendering of Markdown-based prompt templates from embedded assets, XDG config directories, and custom paths.
- **LLM Provider Registry (`internal/core/llm`)**: A unified management layer for different LLM backends, providing a consistent interface for sending prompts and receiving structured results.
- **TUI Wizard (`internal/ui`)**: Orchestrates the interactive user flow, managing the state transitions between file selection, template choice, and input gathering.

## Service Definitions
- **File System Scanner**: Provides the ability to discover and index files while respecting complex ignore rules and memory constraints.
- **Prompt Renderer**: A specialized service that transforms structured codebase data into formatted Markdown prompts using a custom template engine.
- **Token Estimator**: A heuristic-based service in `internal/core/tokens` that approximates token counts to help users stay within LLM context windows without requiring heavy external tokenizer dependencies.
- **Platform Clients**: Located in `internal/platform/`, these services provide low-level API integrations for Anthropic, OpenAI, and Gemini, as well as system-level utilities like clipboard management.

## Interface Contracts
- **`Scanner`**: Defines methods for scanning filesystems (`Scan`, `ScanWithProgress`) and returning a `FileNode` tree.
- **`IgnoreEngine`**: Contracts for checking if paths should be ignored and managing rule layers.
- **`ContextGenerator`**: Defines the interface for transforming a file tree and selections into a prompt string.
- **`Provider`**: The abstraction for LLM services, requiring methods like `Send`, `IsAvailable`, and `Name`.
- **`TemplateManager`**: Interface for discovering and managing prompt templates.

## Design Patterns Identified
- **Strategy Pattern**: Employed in the `llm.Provider` interface, allowing the application to switch between different LLM backends (OpenAI, Anthropic, Gemini) interchangeably.
- **Composite Pattern**: The `FileNode` structure represents the filesystem as a recursive tree of nodes, simplifying operations like size calculation and tree rendering.
- **Model-View-Update (MVU)**: The core pattern for the TUI wizard, ensuring predictable state management across complex user interactions.
- **Registry Pattern**: Used in the LLM and Template packages to manage and retrieve available providers or templates by name.
- **Layered Rules Engine**: Specifically used in the `IgnoreEngine` to handle the priority of different ignore sources (e.g., explicit excludes overriding `.gitignore`).

## Component Relationships
1. **CLI/TUI -> Scanner**: The UI layers initiate scans to display the project structure to the user.
2. **Scanner -> Ignore Engine**: The scanner consults the ignore engine for every file/directory encountered to determine if it should be indexed.
3. **Context Generator -> Template Manager**: The generator uses the manager to fetch and render the user's chosen prompt template.
4. **Context Generator -> Tree Renderer**: Internal utilities transform the `FileNode` tree into ASCII representations for inclusion in the prompt.
5. **TUI Wizard -> LLM Provider**: The final step of the wizard uses the provider registry to send the generated context to the selected AI service.

## Key Methods & Functions
- **`cmd.Execute()`**: The entry point for the CLI, handling configuration loading and subcommand routing.
- **`scanner.ScanWithProgress()`**: Performs the heavy lifting of project indexing with asynchronous feedback.
- **`context.GenerateWithProgressEx()`**: The primary function for assembling the final LLM-optimized payload.
- **`ignore.ShouldIgnore()`**: The core logic for the layered ignore system.
- **`tokens.Estimate()`**: Provides real-time feedback on prompt size relative to model limits.
- **`ui.NewWizard()`**: Initializes the TUI state machine and launches the interactive interface.