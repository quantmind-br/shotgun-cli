# Code Structure Analysis

## Architectural Overview
The `shotgun-cli` is a Go-based command-line tool designed to generate LLM-optimized context from a codebase. It employs a **modular CLI architecture** with a clear separation between the presentation layer (TUI vs. Headless), core business logic (scanning, template management, context generation), and external platform integrations (Gemini, Clipboard).

The application follows a **layered design**:
1.  **Entry Layer**: `main.go` and `cmd/` (Cobra-based CLI handling).
2.  **Orchestration Layer**: `internal/ui/wizard.go` (Bubble Tea-based TUI) manages the workflow state.
3.  **Core Domain Layer**: `internal/core/` handles file scanning, ignore rules, and context assembly.
4.  **Platform/Infrastructure Layer**: `internal/platform/` handles OS-level and external API interactions.

## Core Components
*   **CLI Engine (`cmd/`)**: Built with Cobra and Viper, it handles command routing, flag parsing, and configuration loading. It provides two modes: an interactive TUI wizard and a headless CLI.
*   **Scanner (`internal/core/scanner`)**: Responsible for traversing the filesystem. It builds a hierarchical `FileNode` tree while respecting complex ignore rules.
*   **Template System (`internal/core/template`)**: A sophisticated subsystem for loading, managing, and rendering templates. It supports multiple sources (embedded, XDG config, and custom paths) with a priority-based override mechanism.
*   **Context Generator (`internal/core/context`)**: The primary business logic component that assembles the final LLM prompt. It merges file tree structures, selected file contents, and user-provided instructions into a templated output.
*   **TUI Wizard (`internal/ui`)**: An interactive 5-step workflow (File Selection → Template Selection → Task Input → Rules Input → Review) built using the Bubble Tea (The Elm Architecture) pattern.

## Service Definitions
*   **`Scanner`**: A service that performs filesystem traversal. It provides both standard and progress-aware scanning methods.
*   **`TemplateManager`**: Manages the lifecycle of prompt templates, including discovery across different filesystem locations and variable validation.
*   **`ContextGenerator`**: Orchestrates the transformation of raw file data and user input into a formatted Markdown document optimized for LLM consumption.
*   **`GeminiExecutor`**: An infrastructure service that interacts with the `geminiweb` binary to send generated contexts to the Gemini LLM and parse the responses.

## Interface Contracts
*   **`scanner.Scanner`**: Defines `Scan` and `ScanWithProgress`. This abstraction allows for different scanning implementations (e.g., concurrent vs. sequential).
*   **`template.TemplateManager`**: Contracts for `ListTemplates`, `GetTemplate`, and `RenderTemplate`, decoupling the UI from how templates are stored or rendered.
*   **`context.ContextGenerator`**: Defines the `Generate` methods. It takes a file tree and user selections and returns the final string context.
*   **`tea.Model`**: The `WizardModel` and its sub-screens implement this interface, ensuring a consistent message-passing architecture for the TUI.

## Design Patterns Identified
*   **The Elm Architecture (TEA)**: Heavily used in the `internal/ui` package via Bubble Tea for managing state, updates, and view rendering.
*   **Composite Pattern**: The `FileNode` structure represents the filesystem as a tree of nodes, allowing uniform treatment of files and directories.
*   **Strategy Pattern**: The `TemplateSource` interface (implemented by `EmbeddedSource` and `FilesystemSource`) allows the `TemplateManager` to load templates from different locations using the same logic.
*   **Observer/Reactive Pattern**: Use of Go channels for reporting progress from long-running core operations (scanning, generation) back to the UI.
*   **Factory Pattern**: `NewWizard`, `NewManager`, and `NewDefaultContextGenerator` are used to instantiate complex components with their dependencies.

## Component Relationships
*   **`WizardModel` → `Scanner`**: The wizard initiates the scan at startup to populate the file selection screen.
*   **`WizardModel` → `ContextGenerator`**: The wizard feeds gathered user inputs (task, rules) and file selections into the generator at the final step.
*   **`ContextGenerator` → `TemplateRenderer`**: The generator uses templates to format the final output.
*   **`TemplateManager` → `TemplateSource`**: The manager aggregates templates from multiple sources, applying priority rules.
*   **`cmd` → `WizardModel`**: The root command launches the TUI by initializing the wizard and passing it to the Bubble Tea runtime.

## Key Methods & Functions
*   **`scanner.ScanWithProgress`**: Performs thread-safe filesystem traversal with real-time feedback.
*   **`template.Manager.loadFromSources`**: Implements the priority-based template loading logic.
*   **`context.DefaultContextGenerator.GenerateWithProgressEx`**: The "brain" of the application that sequences tree rendering, content collection, and template interpolation.
*   **`gemini.Executor.Send`**: Manages external process execution (geminiweb), stdin/stdout piping, and ANSI code stripping for LLM interaction.
*   **`ui.WizardModel.Update`**: The central state machine transition function that handles all TUI navigation and asynchronous task completions.