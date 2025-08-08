# Technology Stack

## Core Technologies
- **Go 1.24.5**: Main programming language
- **BubbleTea**: Terminal UI framework from Charm
- **Lipgloss**: Styling and layout for TUI components
- **Node.js/npm**: Build system orchestration and distribution

## Key Dependencies

### UI Framework
- `github.com/charmbracelet/bubbletea`: Terminal UI framework
- `github.com/charmbracelet/lipgloss`: Styling library
- `github.com/charmbracelet/bubbles`: Pre-built UI components

### Core Functionality
- `github.com/sabhiram/go-gitignore`: .gitignore parsing
- `github.com/adrg/xdg`: Cross-platform directory paths
- `github.com/fsnotify/fsnotify`: File system watching
- `gopkg.in/yaml.v3`: YAML parsing for template frontmatter

### Configuration & Storage
- `github.com/knadh/koanf/v2`: Configuration management
- `github.com/99designs/keyring`: Secure credential storage
- `github.com/go-playground/validator/v10`: Configuration validation

### Resilience & Reliability  
- `github.com/sony/gobreaker`: Circuit breaker pattern
- `golang.org/x/time`: Rate limiting and timeouts

### AI Integration
- `github.com/sashabaranov/go-openai`: OpenAI API client for translation features

### Testing
- `github.com/stretchr/testify`: Testing framework

## Architecture Pattern
- **MVC-like**: Separation between UI (internal/ui), business logic (internal/core), and main entry point
- **BubbleTea Model-View-Update**: Standard pattern for terminal UIs
- **Modular Design**: Clear separation of concerns between scanning, generation, templating, and configuration