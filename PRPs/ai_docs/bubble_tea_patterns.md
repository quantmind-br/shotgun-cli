# Bubble Tea TUI Implementation Patterns for Shotgun-CLI

## Critical URLs and Documentation

### Official Documentation
- **Main Site**: https://github.com/charmbracelet/bubbletea
- **Examples**: https://github.com/charmbracelet/bubbletea/tree/master/examples
- **Tutorial**: https://github.com/charmbracelet/bubbletea/tree/master/tutorials
- **API Reference**: https://pkg.go.dev/github.com/charmbracelet/bubbletea

### Key Version Requirements
- **Bubble Tea**: ≥1.3.5 (for Windows F-key support)
- **Lip Gloss**: Latest for styling integration

## Multi-Step Wizard Implementation Pattern

### Core Model Structure
```go
type WizardModel struct {
    step          int                    // 1-5 current step
    fileTree      *FileNode             // File tree state
    selectedFiles map[string]bool       // File selections
    template      *TemplateDescriptor   // Selected template
    taskDesc      string               // User task input
    rules         string               // User rules input
    progress      Progress             // Progress state
    error         error                // Current error
    width         int                  // Terminal width
    height        int                  // Terminal height
}

func (m WizardModel) Init() tea.Cmd {
    return tea.Batch(
        tea.EnterAltScreen,
        scanDirectoryCmd(),
    )
}

func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyInput(msg)
    case tea.WindowSizeMsg:
        m.width, m.height = msg.Width, msg.Height
    case ScanCompleteMsg:
        m.fileTree = msg.FileTree
    }

    return m, nil
}

func (m WizardModel) View() string {
    switch m.step {
    case 1:
        return m.renderFileSelection()
    case 2:
        return m.renderTemplateSelection()
    case 3:
        return m.renderTaskInput()
    case 4:
        return m.renderRulesInput()
    case 5:
        return m.renderReview()
    default:
        return "Unknown step"
    }
}
```

### Navigation Pattern
```go
func (m WizardModel) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // Global shortcuts
    switch msg.String() {
    case "ctrl+q":
        return m, tea.Quit
    case "?":
        m.showHelp = !m.showHelp
        return m, nil
    case "f8", "ctrl+pgdown":
        if m.canAdvance() {
            m.step++
        }
        return m, nil
    case "f10", "ctrl+pgup":
        if m.step > 1 {
            m.step--
        }
        return m, nil
    }

    // Step-specific handling
    switch m.step {
    case 1:
        return m.handleFileSelectionInput(msg)
    case 2:
        return m.handleTemplateSelectionInput(msg)
    case 3:
        return m.handleTaskInputInput(msg)
    case 4:
        return m.handleRulesInputInput(msg)
    case 5:
        return m.handleReviewInput(msg)
    }

    return m, nil
}
```

## File Tree Component Pattern

### Tree Node Structure
```go
type FileNode struct {
    Name            string
    Path            string      // Absolute path
    RelPath         string      // Relative from root
    IsDir           bool
    Children        []FileNode
    Selected        bool
    IsGitignored    bool
    IsCustomIgnored bool
    Size            int64
    Expanded        bool        // For tree navigation
}

type FileTreeModel struct {
    tree          *FileNode
    cursor        int           // Current position
    selections    map[string]bool
    showIgnored   bool
    filter        string
}
```

### Tree Navigation Implementation
```go
func (m FileTreeModel) handleInput(msg tea.KeyMsg) (FileTreeModel, tea.Cmd) {
    switch msg.String() {
    case "up", "k":
        if m.cursor > 0 {
            m.cursor--
        }
    case "down", "j":
        m.cursor++
    case "left", "h":
        m.collapseCurrentNode()
    case "right", "l":
        m.expandCurrentNode()
    case " ":
        m.toggleSelection()
    case "/":
        return m, m.startFilterMode()
    case "d":
        m.toggleDirectorySelection()
    case "i":
        m.showIgnored = !m.showIgnored
    case "f5":
        return m, m.rescanDirectory()
    }
    return m, nil
}

func (m FileTreeModel) View() string {
    var builder strings.Builder

    // Render header
    builder.WriteString(lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("14")).
        Render("File Selection") + "\n\n")

    // Render tree
    m.renderTree(&builder, m.tree, "", 0)

    // Render footer with shortcuts
    builder.WriteString(m.renderFooter())

    return builder.String()
}

func (m FileTreeModel) renderTree(builder *strings.Builder, node *FileNode, prefix string, index int) {
    if !m.showIgnored && (node.IsGitignored || node.IsCustomIgnored) {
        return
    }

    // Apply filter
    if m.filter != "" && !strings.Contains(strings.ToLower(node.Name), strings.ToLower(m.filter)) {
        return
    }

    // Cursor highlighting
    style := lipgloss.NewStyle()
    if index == m.cursor {
        style = style.Background(lipgloss.Color("240"))
    }

    // Selection checkbox
    checkbox := "[ ]"
    if node.Selected {
        checkbox = "[x]"
    }

    // Directory indicator
    name := node.Name
    if node.IsDir {
        name += "/"
    }

    // Ignore status indicators
    status := ""
    if node.IsGitignored {
        status = " (g)"
    } else if node.IsCustomIgnored {
        status = " (c)"
    }

    line := fmt.Sprintf("%s%s %s%s", prefix, checkbox, name, status)
    builder.WriteString(style.Render(line) + "\n")

    // Render children if expanded
    if node.IsDir && node.Expanded {
        childPrefix := prefix + "  "
        for _, child := range node.Children {
            m.renderTree(builder, &child, childPrefix, index+1)
        }
    }
}
```

## Progress Bar Integration Pattern

```go
type Progress struct {
    Current int
    Total   int
    Stage   string          // "scanning", "generating"
    Message string
}

// Background operation with progress
func scanDirectoryCmd() tea.Cmd {
    return tea.Exec(func() tea.Msg {
        progress := make(chan Progress, 100)

        go func() {
            defer close(progress)
            // File scanning with progress updates
            scanner := NewScanner()
            tree, err := scanner.ScanWithProgress(".", progress)
            if err != nil {
                progress <- Progress{Stage: "error", Message: err.Error()}
                return
            }
            progress <- Progress{Stage: "complete", Message: "Scan complete"}
        }()

        return ScanProgressMsg{Progress: progress}
    }, nil)
}
```

## Text Input Component Pattern

### Multiline Text Input with Counter
```go
import "github.com/charmbracelet/bubbles/textarea"

type TextInputModel struct {
    textarea    textarea.Model
    title       string
    placeholder string
    maxChars    int
    required    bool
}

func NewTextInputModel(title, placeholder string) TextInputModel {
    ta := textarea.New()
    ta.Placeholder = placeholder
    ta.CharLimit = 2000
    ta.SetWidth(80)
    ta.SetHeight(10)
    ta.Focus()

    return TextInputModel{
        textarea:    ta,
        title:       title,
        placeholder: placeholder,
        maxChars:    2000,
    }
}

func (m TextInputModel) Update(msg tea.Msg) (TextInputModel, tea.Cmd) {
    var cmd tea.Cmd
    m.textarea, cmd = m.textarea.Update(msg)
    return m, cmd
}

func (m TextInputModel) View() string {
    var builder strings.Builder

    // Title
    builder.WriteString(lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("14")).
        Render(m.title) + "\n\n")

    // Text area
    builder.WriteString(m.textarea.View() + "\n")

    // Character counter
    current := len(m.textarea.Value())
    counter := fmt.Sprintf("%d/%d characters", current, m.maxChars)

    counterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
    if current > m.maxChars {
        counterStyle = counterStyle.Foreground(lipgloss.Color("9"))
    }

    builder.WriteString(counterStyle.Render(counter) + "\n")

    return builder.String()
}
```

## Styling with Lip Gloss Pattern

```go
var (
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("14")).
        Background(lipgloss.Color("235")).
        Padding(0, 1)

    selectedStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("240")).
        Foreground(lipgloss.Color("15"))

    errorStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("9")).
        Bold(true)

    successStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("10")).
        Bold(true)

    helpStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("8")).
        Italic(true)
)

func (m WizardModel) renderFooter() string {
    shortcuts := []string{
        "F8: Next Step",
        "F10: Previous Step",
        "Ctrl+Q: Quit",
        "?: Help",
    }

    return helpStyle.Render(strings.Join(shortcuts, " • "))
}
```

## Key Implementation Gotchas

### Memory Management
- Use `tea.Batch()` for multiple commands
- Implement proper cleanup in `tea.Quit`
- Limit file tree depth for large projects

### Cross-Platform F-Key Support
```go
// Ensure compatibility with Windows
func (m WizardModel) handleFunctionKeys(msg tea.KeyMsg) {
    switch msg.String() {
    case "f8", "ctrl+pgdown":  // Windows fallback
        m.nextStep()
    case "f10", "ctrl+pgup":   // Windows fallback
        m.previousStep()
    }
}
```

### Performance Optimization
- Lazy load file tree children
- Implement virtual scrolling for large lists
- Use debounced filtering for search

### State Management
- Keep step state separate from component state
- Use message passing for component communication
- Validate state transitions between steps

## Integration with Background Operations

```go
func (m WizardModel) generateContextCmd() tea.Cmd {
    return tea.Exec(func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        generator := NewContextGenerator()
        result, err := generator.Generate(ctx, GenerateConfig{
            FileTree: m.fileTree,
            Template: m.template,
            Task:     m.taskDesc,
            Rules:    m.rules,
        })

        if err != nil {
            return GenerationErrorMsg{Error: err}
        }

        return GenerationCompleteMsg{Result: result}
    }, cancel)
}
```

This documentation provides the essential patterns needed to implement a sophisticated multi-step TUI wizard using Bubble Tea, including file tree navigation, progress reporting, and cross-platform compatibility.