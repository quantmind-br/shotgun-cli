# UI Package - TUI Wizard

Parent: [../AGENTS.md](../AGENTS.md)

## OVERVIEW

Bubble Tea-based interactive wizard. **MVU pattern** with composed screen models.

## STRUCTURE

```
ui/
├── wizard.go              # Main orchestrator, 5-step state machine
├── scan_coordinator.go    # Async scan state management
├── generate_coordinator.go # Async generation state management
├── config_wizard.go       # Config TUI
├── screens/               # Individual wizard screens
│   ├── file_selection.go  # File tree selection
│   ├── template_selection.go
│   ├── task_input.go
│   ├── rules_input.go
│   └── review.go
├── components/            # Reusable TUI components
│   ├── tree.go           # File tree renderer
│   ├── progress.go       # Progress indicators
│   └── modal.go          # Modal dialogs
└── styles/               # Theme and styling
    └── theme.go
```

## SCREEN MODEL ARCHITECTURE

Each screen owns its state. `WizardModel` delegates via accessors:

```go
// WizardModel fields
fileSelection      *FileSelectionModel
templateSelection  *TemplateSelectionModel
taskInput          *TaskInputModel
rulesInput         *RulesInputModel
review             *ReviewModel

// Accessors delegate to screen models
func (m *WizardModel) getSelectedFiles() map[string]bool
func (m *WizardModel) getSelectedTemplate() *template.Template
```

## KEY PATTERNS

### Coordinator Pattern
Background operations use coordinators:
```go
scanCoordinator.Start(rootPath, config)     // Initiates scan
generateCoordinator.Start(generateConfig)   // Initiates generation
```

### Message Routing
```go
func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m.step {
    case StepFileSelection:
        return m.fileSelection.Update(msg)
    // ...
    }
}
```

## WIZARD STEPS

1. **File Selection** - Tree navigation, filtering (`/`), selection
2. **Template Selection** - Choose prompt template
3. **Task Input** - Describe the task
4. **Rules Input** - Optional constraints
5. **Review** - Summary, generate, copy, send to LLM

## KEYBOARD SHORTCUTS

| Global | Action |
|--------|--------|
| F1 | Help |
| F7/Ctrl+P | Previous |
| F8/Ctrl+N | Next |
| Ctrl+Q | Quit |

| File Selection | Action |
|----------------|--------|
| ↑/↓/k/j | Navigate |
| ←/→/h/l | Collapse/Expand |
| Space | Toggle selection |
| / | Filter mode |
| a/A | Select/Deselect all |

## TESTING

```bash
# All UI tests
go test -v -race ./internal/ui/...

# Specific screen tests
go test -v -run TestFileSelection ./internal/ui/screens/
go test -v -run TestWizard ./internal/ui/

# Coordinator tests
go test -v -run TestScanCoordinator ./internal/ui/
go test -v -run TestGenerateCoordinator ./internal/ui/
```

## ANTI-PATTERNS

- Storing screen state in WizardModel (use composed models)
- Blocking in Update() (use coordinators)
- Direct service calls from screens (go through wizard)
