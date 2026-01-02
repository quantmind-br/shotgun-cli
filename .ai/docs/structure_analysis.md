# Code Structure Analysis

## Architectural Overview
The codebase follows a layered, modular CLI architecture in Go, utilizing a "Hexagonal-lite" approach where core business logic is decoupled from user interface and external platform integrations. The application supports two primary modes: an interactive TUI (Terminal User Interface) wizard and a headless CLI mode. 

The architecture is divided into four main layers:
1.  **Command Layer (`cmd/`)**: Entry point using the Cobra framework, responsible for CLI argument parsing, configuration loading, and orchestration.
2.  **UI Layer (`internal/ui/`)**: Implements the interactive wizard using the Bubble Tea framework (MVU pattern - Model-View-Update).
3.  **Core Domain Layer (`internal/core/`)**: Contains the primary business logic for file scanning, template management, and context generation.
4.  **Platform/Infrastructure Layer (`internal/platform/`)**: Handles external integrations like the Gemini LLM executor and system clipboard.

## Core Components
*   **Scanner (`internal/core/scanner`)**: Responsibile for walking the filesystem, respecting ignore rules, and building a `FileNode` tree representation of the codebase.
*   **Context Generator (`internal/core/context`)**: The heart of the application; it aggregates selected file contents, renders the project structure, and applies templates to produce the final LLM-optimized prompt.
*   **Ignore Engine (`internal/core/ignore`)**: A layered rule processor that handles built-in, `.gitignore`, and custom `.shotgunignore` patterns to filter files during scanning.
*   **Template Manager (`internal/core/template`)**: Manages the discovery and loading of templates from embedded assets, user config directories, and custom paths.
*   **Wizard (`internal/ui/wizard.go`)**: Orchestrates the multi-step user flow (File Selection -> Template Selection -> Task Input -> Rules Input -> Review/Generate).

## Service Definitions
*   **FileSystemScanner**: A service that implements the `Scanner` interface to provide concurrent file discovery with progress reporting via channels.
*   **DefaultContextGenerator**: Implements the `ContextGenerator` interface to transform a file tree and selected files into a formatted prompt string.
*   **Gemini Executor**: A bridge service that wraps the `geminiweb` external binary to send generated contexts directly to the LLM and retrieve responses.
*   **Token Estimator**: Provides heuristic-based token counting (e.g., 4 bytes per token) to help users stay within LLM context window limits.

## Interface Contracts
*   **`Scanner`**: Defines `Scan` and `ScanWithProgress`. Allows for different scanning implementations (e.g., local vs. remote).
*   **`ContextGenerator`**: Defines methods for prompt generation (`Generate`, `GenerateWithProgress`). It ensures the generation logic is consistent regardless of whether it's called from the TUI or headless CLI.
*   **`IgnoreEngine`**: Defines the contract for rule-based path filtering, supporting dynamic loading of ignore files.
*   **`TemplateManager`**: Orchestrates template lifecycle, including discovery across multiple prioritized sources.

## Design Patterns Identified
*   **Model-View-Update (MVU)**: Used throughout the TUI layer via the Bubble Tea framework to manage UI state and transitions.
*   **Strategy Pattern**: Employed in the `IgnoreEngine` where different matchers (built-in, gitignore, custom) are layered to decide if a path should be filtered.
*   **Composition**: The `WizardModel` composes various screen models (`FileSelectionModel`, `ReviewModel`, etc.) to delegate UI responsibilities.
*   **Singleton/Manager**: The `TemplateManager` acts as a central registry for all prompt templates available to the system.
*   **Dependency Injection**: Services like the `Scanner` and `ContextGenerator` are instantiated and passed into the UI components, facilitating testability.

## Component Relationships
*   **Orchestration**: `cmd/root.go` initializes the configuration and launches either the `Wizard` (TUI) or specific subcommands.
*   **Data Flow**: The `Scanner` produces a `FileNode` tree. The UI allows users to toggle selection flags on these nodes. The `ContextGenerator` then consumes the `FileNode` tree and the map of selections to produce the final text output.
*   **Feedback Loop**: Core services use Go channels (`chan Progress`) to send real-time updates back to the UI layer, allowing the Bubble Tea `Update` loop to refresh progress bars and status messages.

## Key Methods & Functions
*   **`Scanner.ScanWithProgress`**: Asynchronously crawls the filesystem and streams progress updates.
*   **`ContextGenerator.GenerateWithProgressEx`**: Orchestrates the rendering of the ASCII tree structure and file content blocks into a cohesive prompt.
*   **`IgnoreEngine.ShouldIgnore`**: The central logic for path filtering, applying priority-based matching across multiple rule layers.
*   **`Executor.Send`**: Executes the external `geminiweb` process, handling stdin/stdout and timeout management.
*   **`EstimateFromBytes`**: The primary heuristic function for calculating token usage without requiring heavy tokenizer libraries.