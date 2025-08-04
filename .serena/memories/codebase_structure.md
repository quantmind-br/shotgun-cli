# Codebase Structure

## High-Level Architecture
```
shotgun-cli/
├── main.go                   # Application entry point with Windows UTF-8 support
├── internal/                 # Private packages
│   ├── core/                 # Business logic layer
│   │   ├── types.go          # Core data structures and interfaces
│   │   ├── scanner.go        # Directory scanning with gitignore support
│   │   ├── generator.go      # Context generation and file processing
│   │   ├── template.go       # Complex template processing
│   │   ├── template_simple.go # Simple template processing
│   │   ├── config.go         # Configuration management
│   │   ├── keyring.go        # Secure credential storage
│   │   └── translator.go     # OpenAI translation integration
│   └── ui/                   # BubbleTea UI layer
│       ├── app.go            # Main application model and state management
│       ├── filetree.go       # File tree UI with exclusion controls
│       ├── components.go     # Custom UI components (NumberedTextArea)
│       ├── views.go          # View rendering logic
│       ├── config_form.go    # Configuration form component
│       └── config_views.go   # Configuration view rendering
├── templates/                # Prompt templates
│   ├── prompt_makeDiffGitFormat.md    # Dev template (git diff generation)
│   ├── prompt_makePlan.md             # Architect template (design planning)
│   ├── prompt_analyzeBug.md           # Debug template (bug analysis)
│   └── prompt_projectManager.md       # Project manager template
├── scripts/                  # Build and installation scripts
│   ├── build.js              # Cross-platform build orchestration
│   ├── install.js            # Post-install binary setup
│   └── clean.js              # Cleanup build artifacts
├── bin/                      # Generated binaries (gitignored)
└── package.json              # npm configuration and scripts
```

## Key Component Responsibilities

### Core Layer (`internal/core/`)
- **DirectoryScanner**: Recursive directory traversal with three-layer filtering (built-in patterns, .gitignore, custom rules)
- **ContextGenerator**: File processing with worker pools and progress tracking
- **TemplateProcessor**: Template loading and prompt generation with variable substitution
- **SelectionState**: Thread-safe file inclusion/exclusion management
- **ConfigManager**: Application configuration persistence
- **SecureKeyManager**: Secure storage for API keys using system keyring

### UI Layer (`internal/ui/`)
- **Model**: Main application state machine with ViewState enum
- **FileTreeModel**: Interactive file tree with inverse selection (exclude files)
- **NumberedTextArea**: Enhanced text input component with line numbers
- **ConfigFormModel**: Configuration interface for OpenAI integration

### Template System
Four specialized templates for different use cases:
- Dev: Generates git diff format for code changes
- Architect: Creates structured design plans
- Debug: Provides bug analysis framework
- Project Manager: Handles documentation synchronization