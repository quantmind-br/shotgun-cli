# Dependency Analysis

## Internal Dependencies

The application follows a layered architecture with clear boundaries between command-line interfaces, user interface components, and core domain logic.

### Module Map
- **cmd**: Entry point (Cobra). Depends on `internal/ui` (for wizard mode), `internal/core/scanner` (for config), and `internal/platform/gemini`.
- **internal/ui**: TUI implementation (Bubble Tea). High-level orchestrator.
  - Depends on `internal/ui/screens`, `internal/ui/components` for UI layout.
  - Depends on `internal/core/scanner`, `internal/core/context`, `internal/core/template` for domain logic.
  - Depends on `internal/platform/clipboard`, `internal/platform/gemini` for side-effect operations.
- **internal/core/context**: Handles the logic of combining files and templates.
  - Depends on `internal/core/scanner` for the data structures (`FileNode`).
  - Contains `TreeRenderer` and `TemplateRenderer`.
- **internal/core/scanner**: Filesystem traversal logic.
  - Depends on `internal/core/ignore` for filtering logic.
- **internal/core/template**: Template loading and management.
  - Depends on `internal/assets` for embedded templates.
- **internal/core/ignore**: Abstraction over ignore rules (gitignore, custom, built-in).
  - Standalone domain logic focused on path matching.
- **internal/platform**: Infrastructure and external integrations.
  - **gemini**: Wrapper around the `geminiweb` CLI.
  - **clipboard**: OS-level clipboard interaction.
- **internal/assets**: Static assets and embedded Go templates.

## External Dependencies

### Core Frameworks
- **github.com/spf13/cobra**: CLI command structure and argument parsing.
- **github.com/spf13/viper**: Configuration management (YAML, Env, Flags).
- **github.com/charmbracelet/bubbletea**: TUI framework (The Elm Architecture in Go).
- **github.com/charmbracelet/bubbles & lipgloss**: TUI components and styling.

### Utilities
- **github.com/adrg/xdg**: Cross-platform XDG base directory support (config/cache paths).
- **github.com/atotto/clipboard**: Cross-platform clipboard access.
- **github.com/rs/zerolog**: Structured logging.
- **github.com/sabhiram/go-gitignore**: Parsing and matching `.gitignore` files.
- **golang.org/x/text**: Language detection and string casing.
- **github.com/stretchr/testify**: Testing assertions and mocks.

## Dependency Graph

The dependency flow is primarily **Top-Down**:

1.  **UI/CLI Layer** (`cmd`, `internal/ui`) -> **Domain/Core Layer** (`scanner`, `context`, `template`)
2.  **Domain/Core Layer** -> **Infrastructure Layer** (`platform`, `ignore`, `assets`)
3.  **Cross-cutting** (`utils`, `styles`) are used across layers.

The structure avoids circular dependencies by using a "Star" pattern where `internal/ui` acts as the mediator between different core modules, and the core modules communicate via shared data structures (like `scanner.FileNode`) defined in lower-level packages.

## Dependency Injection

The project uses **Constructor Injection** and **Interface-based abstractions**:

- **Interface Abstraction**: `ContextGenerator` and `Scanner` are defined as interfaces, allowing for different implementations (though `DefaultContextGenerator` and `FileSystemScanner` are the primary ones).
- **Manual Injection**: `WizardModel` is initialized with a `ScanConfig` and creates its own instances of screens/components.
- **Functional Injection**: Progress reporting is handled via callbacks (`func(GenProgress)`) or channels (`chan Progress`), decoupling the core logic from UI update mechanisms.
- **Service Management**: `TemplateManager` encapsulates the complexity of loading templates from multiple sources (embedded, XDG, custom) and is injected into components that need template access.

## Potential Issues

- **Tight Coupling to Gemini CLI**: The `internal/platform/gemini` package depends on the availability of an external binary (`geminiweb`). While abstracted, the application logic in `internal/ui/wizard.go` has specific message types and states tied to this integration.
- **Fat Wizard Model**: `internal/ui/wizard.go` acts as a "God Object" for the TUI, holding state for scanning, generation, and Gemini communication. While typical for Bubble Tea `Update` functions, it creates high coupling between all UI screens.
- **Reflective Template Rendering**: The `ContextGenerator` performs string replacements (`convertTemplateVariables`) to bridge custom template syntax to Go's `text/template`, which adds a runtime dependency on specific string patterns.
- **Global Logger/Viper**: Extensive use of global `log` (zerolog) and `viper` across packages makes unit testing components in isolation slightly more difficult as they depend on global state.