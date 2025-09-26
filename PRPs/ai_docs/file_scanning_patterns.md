# Go File Scanning and Context Generation Patterns

## Critical URLs and Documentation

### Official Go Documentation
- **filepath package**: https://pkg.go.dev/path/filepath
- **io/fs package**: https://pkg.go.dev/io/fs
- **text/template**: https://pkg.go.dev/text/template
- **embed package**: https://pkg.go.dev/embed

### Recommended Libraries
- **Gitignore Parser**: https://github.com/sabhiram/go-gitignore
- **Alternative**: https://github.com/go-git/go-git/tree/master/plumbing/format/gitignore

## High-Performance Directory Scanning Pattern

### Core Scanner Interface
```go
package scanner

import (
    "context"
    "io/fs"
    "path/filepath"
    "strings"
    "sync"
)

type Scanner interface {
    Scan(rootPath string) (*FileNode, error)
    ScanWithProgress(rootPath string, progress chan<- Progress) (*FileNode, error)
}

type FileNode struct {
    Name            string      `json:"name"`
    Path            string      `json:"path"`           // Absolute path
    RelPath         string      `json:"rel_path"`       // Relative from root
    IsDir           bool        `json:"is_dir"`
    Children        []FileNode  `json:"children,omitempty"`
    Selected        bool        `json:"selected"`
    IsGitignored    bool        `json:"is_gitignored"`
    IsCustomIgnored bool        `json:"is_custom_ignored"`
    Size            int64       `json:"size"`
}

type Progress struct {
    Current int64  `json:"current"`
    Total   int64  `json:"total"`
    Stage   string `json:"stage"`    // "scanning", "generating"
    Message string `json:"message"`
}

type ScanConfig struct {
    MaxFileSize   int64         // 10MB default
    MaxFiles      int           // Safety limit
    MaxMemory     int64         // 100MB default
    SkipBinary    bool          // Skip binary files
    IncludeHidden bool          // Include hidden files
    Workers       int           // Concurrent workers
}
```

### Optimized Directory Walker
```go
type FileSystemScanner struct {
    config       ScanConfig
    ignoreEngine *IgnoreEngine
    mutex        sync.RWMutex
    stats        ScanStats
}

type ScanStats struct {
    FilesProcessed int64
    DirsProcessed  int64
    BytesProcessed int64
    SkippedFiles   int64
}

func (s *FileSystemScanner) ScanWithProgress(rootPath string, progress chan<- Progress) (*FileNode, error) {
    rootPath, err := filepath.Abs(rootPath)
    if err != nil {
        return nil, fmt.Errorf("invalid root path: %w", err)
    }

    // Create root node
    root := &FileNode{
        Name:    filepath.Base(rootPath),
        Path:    rootPath,
        RelPath: ".",
        IsDir:   true,
    }

    // First pass: count total items for progress
    totalItems, err := s.countItems(rootPath)
    if err != nil {
        return nil, fmt.Errorf("failed to count items: %w", err)
    }

    progress <- Progress{
        Total:   totalItems,
        Stage:   "scanning",
        Message: "Starting directory scan...",
    }

    // Second pass: build tree with progress reporting
    err = s.walkAndBuild(rootPath, root, progress)
    if err != nil {
        return nil, fmt.Errorf("scan failed: %w", err)
    }

    // Sort children: directories first, then files
    s.sortChildren(root)

    progress <- Progress{
        Current: totalItems,
        Total:   totalItems,
        Stage:   "complete",
        Message: fmt.Sprintf("Scanned %d files and %d directories", s.stats.FilesProcessed, s.stats.DirsProcessed),
    }

    return root, nil
}

func (s *FileSystemScanner) walkAndBuild(rootPath string, node *FileNode, progress chan<- Progress) error {
    return filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            // Log error but continue scanning
            log.Warn().Err(err).Str("path", path).Msg("Error accessing path")
            return nil
        }

        // Calculate relative path
        relPath, err := filepath.Rel(rootPath, path)
        if err != nil {
            return err
        }

        // Skip root directory itself
        if path == rootPath {
            return nil
        }

        // Check ignore rules
        ignored, reason := s.ignoreEngine.ShouldIgnore(relPath)

        if ignored {
            s.stats.SkippedFiles++

            // Skip entire directory if it's ignored
            if d.IsDir() {
                return filepath.SkipDir
            }
            return nil
        }

        // Create node
        fileNode := FileNode{
            Name:            d.Name(),
            Path:            path,
            RelPath:         relPath,
            IsDir:           d.IsDir(),
            IsGitignored:    reason == IgnoreReasonGitignore,
            IsCustomIgnored: reason == IgnoreReasonCustom,
        }

        // Get file size for regular files
        if !d.IsDir() {
            if info, err := d.Info(); err == nil {
                fileNode.Size = info.Size()
                s.stats.BytesProcessed += info.Size()
            }
        }

        // Add to parent node
        parentNode := s.findParentNode(node, relPath)
        if parentNode != nil {
            parentNode.Children = append(parentNode.Children, fileNode)
        }

        // Update progress
        if d.IsDir() {
            s.stats.DirsProcessed++
        } else {
            s.stats.FilesProcessed++
        }

        // Send progress update every 100 items
        if (s.stats.FilesProcessed+s.stats.DirsProcessed)%100 == 0 {
            progress <- Progress{
                Current: s.stats.FilesProcessed + s.stats.DirsProcessed,
                Stage:   "scanning",
                Message: fmt.Sprintf("Processed %d items...", s.stats.FilesProcessed+s.stats.DirsProcessed),
            }
        }

        return nil
    })
}

func (s *FileSystemScanner) sortChildren(node *FileNode) {
    if !node.IsDir {
        return
    }

    // Sort children: directories first, then files, alphabetically within each group
    sort.Slice(node.Children, func(i, j int) bool {
        if node.Children[i].IsDir && !node.Children[j].IsDir {
            return true
        }
        if !node.Children[i].IsDir && node.Children[j].IsDir {
            return false
        }
        return strings.ToLower(node.Children[i].Name) < strings.ToLower(node.Children[j].Name)
    })

    // Recursively sort children
    for i := range node.Children {
        s.sortChildren(&node.Children[i])
    }
}
```

## Layered Ignore Rule Engine

```go
package ignore

import (
    "bufio"
    "os"
    "path/filepath"
    "strings"
    "github.com/sabhiram/go-gitignore/ignore"
)

type IgnoreReason string

const (
    IgnoreReasonNone      IgnoreReason = ""
    IgnoreReasonBuiltIn   IgnoreReason = "builtin"
    IgnoreReasonGitignore IgnoreReason = "gitignore"
    IgnoreReasonCustom    IgnoreReason = "custom"
    IgnoreReasonExplicit  IgnoreReason = "explicit"
)

type IgnoreEngine struct {
    builtInIgnores   *ignore.GitIgnore
    gitignoreRules   *ignore.GitIgnore
    customRules      *ignore.GitIgnore
    explicitIncludes []string
    explicitExcludes []string
}

// Default patterns that shotgun-cli always ignores
var DefaultIgnorePatterns = []string{
    "shotgun-prompt-*.md",  // Auto-generated files
    ".git/",
    ".svn/",
    ".hg/",
    "node_modules/",
    "__pycache__/",
    "*.pyc",
    "*.pyo",
    ".DS_Store",
    "Thumbs.db",
    "*.tmp",
    "*.temp",
    "*.log",
    ".vscode/",
    ".idea/",
    "*.swp",
    "*.swo",
    "*~",
}

func NewIgnoreEngine() *IgnoreEngine {
    // Initialize built-in ignores
    builtIn, err := ignore.CompileIgnoreLines(DefaultIgnorePatterns...)
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to compile built-in ignore patterns")
    }

    return &IgnoreEngine{
        builtInIgnores: builtIn,
    }
}

func (ie *IgnoreEngine) LoadGitignore(rootPath string) error {
    gitignorePath := filepath.Join(rootPath, ".gitignore")

    // Check if .gitignore exists
    if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
        ie.gitignoreRules = ignore.CompileIgnoreLines() // Empty rules
        return nil
    }

    // Load .gitignore rules
    gitIgnore, err := ignore.CompileIgnoreFile(gitignorePath)
    if err != nil {
        return fmt.Errorf("failed to load .gitignore: %w", err)
    }

    ie.gitignoreRules = gitIgnore
    return nil
}

func (ie *IgnoreEngine) AddCustomRules(patterns []string) error {
    if len(patterns) == 0 {
        ie.customRules = ignore.CompileIgnoreLines()
        return nil
    }

    customIgnore, err := ignore.CompileIgnoreLines(patterns...)
    if err != nil {
        return fmt.Errorf("failed to compile custom ignore patterns: %w", err)
    }

    ie.customRules = customIgnore
    return nil
}

func (ie *IgnoreEngine) SetExplicitPatterns(includes, excludes []string) {
    ie.explicitIncludes = includes
    ie.explicitExcludes = excludes
}

func (ie *IgnoreEngine) ShouldIgnore(relPath string) (bool, IgnoreReason) {
    // 1. Check explicit excludes first (highest priority)
    for _, pattern := range ie.explicitExcludes {
        if matched, err := filepath.Match(pattern, relPath); err == nil && matched {
            return true, IgnoreReasonExplicit
        }
        // Also check directory patterns
        if matched, err := filepath.Match(pattern, relPath+"/"); err == nil && matched {
            return true, IgnoreReasonExplicit
        }
    }

    // 2. Check explicit includes (override other ignores)
    for _, pattern := range ie.explicitIncludes {
        if matched, err := filepath.Match(pattern, relPath); err == nil && matched {
            return false, IgnoreReasonNone
        }
    }

    // 3. Check ignore rules in priority order
    if ie.builtInIgnores.MatchesPath(relPath) {
        return true, IgnoreReasonBuiltIn
    }

    if ie.gitignoreRules != nil && ie.gitignoreRules.MatchesPath(relPath) {
        return true, IgnoreReasonGitignore
    }

    if ie.customRules != nil && ie.customRules.MatchesPath(relPath) {
        return true, IgnoreReasonCustom
    }

    return false, IgnoreReasonNone
}
```

## ASCII Tree Generation

```go
package tree

type TreeRenderer struct {
    showIgnored bool
    maxDepth    int
}

func (tr *TreeRenderer) RenderTree(root *FileNode) string {
    var builder strings.Builder

    // Project root header
    builder.WriteString(fmt.Sprintf("Project: %s\n", root.Name))
    builder.WriteString(strings.Repeat("=", len(root.Name)+9) + "\n\n")

    tr.renderNode(&builder, root, "", true, 0)

    return builder.String()
}

func (tr *TreeRenderer) renderNode(builder *strings.Builder, node *FileNode, prefix string, isLast bool, depth int) {
    // Skip ignored files unless explicitly shown
    if !tr.showIgnored && (node.IsGitignored || node.IsCustomIgnored) {
        return
    }

    // Respect max depth
    if tr.maxDepth > 0 && depth > tr.maxDepth {
        return
    }

    // Skip root node rendering
    if node.RelPath == "." {
        for i, child := range node.Children {
            tr.renderNode(builder, &child, "", i == len(node.Children)-1, depth)
        }
        return
    }

    // Tree drawing characters
    connector := "├── "
    if isLast {
        connector = "└── "
    }

    // Build line
    line := prefix + connector + node.Name

    // Add directory indicator
    if node.IsDir {
        line += "/"
    }

    // Add ignore status indicators
    if node.IsGitignored {
        line += " (g)"
    } else if node.IsCustomIgnored {
        line += " (c)"
    }

    // Add file size for regular files
    if !node.IsDir && node.Size > 0 {
        line += fmt.Sprintf(" [%s]", formatFileSize(node.Size))
    }

    builder.WriteString(line + "\n")

    // Render children if directory
    if node.IsDir && len(node.Children) > 0 {
        childPrefix := prefix
        if isLast {
            childPrefix += "    "
        } else {
            childPrefix += "│   "
        }

        for i, child := range node.Children {
            tr.renderNode(builder, &child, childPrefix, i == len(node.Children)-1, depth+1)
        }
    }
}

func formatFileSize(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%dB", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
```

## Context Generation with Templates

```go
package context

import (
    "bytes"
    "fmt"
    "path/filepath"
    "strings"
    "text/template"
    "time"
    "unicode/utf8"
)

type ContextGenerator struct {
    config GenerateConfig
}

type GenerateConfig struct {
    MaxSize       int64    // 10MB default
    MaxFiles      int      // Safety limit
    SkipBinary    bool     // Skip binary files
    TemplateVars  map[string]string
}

type ContextData struct {
    Task          string
    Rules         string
    FileStructure string
    Files         []FileContent
    CurrentDate   string
    Config        GenerateConfig
}

type FileContent struct {
    Path     string
    RelPath  string
    Language string
    Content  string
    Size     int64
}

func (cg *ContextGenerator) Generate(root *FileNode, config GenerateConfig) (string, error) {
    cg.config = config

    // Build context data
    data := ContextData{
        Task:        config.TemplateVars["TASK"],
        Rules:       config.TemplateVars["RULES"],
        CurrentDate: time.Now().Format("2006-01-02 15:04:05"),
        Config:      config,
    }

    // Generate file structure
    renderer := &TreeRenderer{showIgnored: false, maxDepth: 0}
    data.FileStructure = renderer.RenderTree(root)

    // Collect file contents with size limits
    files, err := cg.collectFileContents(root)
    if err != nil {
        return "", fmt.Errorf("failed to collect file contents: %w", err)
    }
    data.Files = files

    // Use default template if none specified
    templateContent := cg.getDefaultTemplate()
    if customTemplate, exists := config.TemplateVars["TEMPLATE"]; exists {
        templateContent = customTemplate
    }

    // Render template
    tmpl, err := template.New("context").Funcs(cg.getTemplateFunctions()).Parse(templateContent)
    if err != nil {
        return "", fmt.Errorf("template parse error: %w", err)
    }

    var buffer bytes.Buffer
    if err := tmpl.Execute(&buffer, data); err != nil {
        return "", fmt.Errorf("template execution error: %w", err)
    }

    result := buffer.String()

    // Check final size limit
    if int64(len(result)) > config.MaxSize {
        return "", fmt.Errorf("generated context exceeds size limit: %d > %d bytes", len(result), config.MaxSize)
    }

    return result, nil
}

func (cg *ContextGenerator) collectFileContents(root *FileNode) ([]FileContent, error) {
    var files []FileContent
    var totalSize int64

    err := cg.walkSelectedFiles(root, func(node *FileNode) error {
        if !node.Selected || node.IsDir {
            return nil
        }

        // Check if we should skip this file
        if cg.shouldSkipFile(node) {
            return nil
        }

        // Read file content
        content, err := os.ReadFile(node.Path)
        if err != nil {
            log.Warn().Err(err).Str("path", node.Path).Msg("Failed to read file")
            return nil // Continue with other files
        }

        // Check if file is binary
        if cg.config.SkipBinary && !isTextFile(content) {
            return nil
        }

        // Check size limits
        if totalSize+int64(len(content)) > cg.config.MaxSize {
            return fmt.Errorf("size limit exceeded: would be %d bytes", totalSize+int64(len(content)))
        }

        totalSize += int64(len(content))

        files = append(files, FileContent{
            Path:     node.Path,
            RelPath:  node.RelPath,
            Language: detectLanguage(node.Name),
            Content:  string(content),
            Size:     int64(len(content)),
        })

        // Safety check for max files
        if len(files) >= cg.config.MaxFiles {
            return fmt.Errorf("maximum file count exceeded: %d", cg.config.MaxFiles)
        }

        return nil
    })

    return files, err
}

func (cg *ContextGenerator) getTemplateFunctions() template.FuncMap {
    return template.FuncMap{
        "truncate": func(s string, length int) string {
            if len(s) <= length {
                return s
            }
            return s[:length] + "..."
        },
        "formatSize": formatFileSize,
        "detectLang": detectLanguage,
        "now": func() string {
            return time.Now().Format("2006-01-02 15:04:05")
        },
        "join": strings.Join,
        "title": strings.Title,
        "upper": strings.ToUpper,
        "lower": strings.ToLower,
    }
}

func (cg *ContextGenerator) getDefaultTemplate() string {
    return `# Project Context

## Task Description
{{.Task}}

{{if .Rules}}## Rules and Guidelines
{{.Rules}}
{{end}}

## Project Structure
` + "```" + `
{{.FileStructure}}
` + "```" + `

## File Contents

{{range .Files}}
### {{.RelPath}}
` + "```{{.Language}}" + `
{{.Content}}
` + "```" + `

{{end}}

---
Generated: {{.CurrentDate}}
Total Files: {{len .Files}}
`
}

// Binary file detection
func isTextFile(data []byte) bool {
    // Sample first 1KB to check
    sample := data
    if len(data) > 1024 {
        sample = data[:1024]
    }

    // Check for null bytes (strong binary indicator)
    if bytes.Contains(sample, []byte{0}) {
        return false
    }

    // Check if valid UTF-8
    return utf8.Valid(sample)
}

// Language detection based on file extension
func detectLanguage(filename string) string {
    ext := strings.ToLower(filepath.Ext(filename))

    langMap := map[string]string{
        ".go":     "go",
        ".js":     "javascript",
        ".ts":     "typescript",
        ".py":     "python",
        ".rs":     "rust",
        ".java":   "java",
        ".cpp":    "cpp",
        ".c":      "c",
        ".h":      "c",
        ".hpp":    "cpp",
        ".cs":     "csharp",
        ".php":    "php",
        ".rb":     "ruby",
        ".sh":     "bash",
        ".bash":   "bash",
        ".zsh":    "zsh",
        ".fish":   "fish",
        ".ps1":    "powershell",
        ".html":   "html",
        ".css":    "css",
        ".scss":   "scss",
        ".sass":   "sass",
        ".json":   "json",
        ".yaml":   "yaml",
        ".yml":    "yaml",
        ".xml":    "xml",
        ".md":     "markdown",
        ".txt":    "text",
        ".sql":    "sql",
        ".dockerfile": "dockerfile",
        ".makefile":   "makefile",
    }

    if lang, exists := langMap[ext]; exists {
        return lang
    }

    // Check for special filenames
    baseName := strings.ToLower(filepath.Base(filename))
    switch baseName {
    case "dockerfile":
        return "dockerfile"
    case "makefile":
        return "makefile"
    case "cmakelists.txt":
        return "cmake"
    case "package.json":
        return "json"
    case "tsconfig.json":
        return "json"
    }

    return "text"
}
```

## Performance Optimization Patterns

### Memory Management
```go
type MemoryLimitedProcessor struct {
    maxMemory     int64
    currentMemory int64
    mutex         sync.Mutex
}

func (mlp *MemoryLimitedProcessor) ProcessFile(path string, size int64, processor func([]byte) error) error {
    mlp.mutex.Lock()
    if mlp.currentMemory+size > mlp.maxMemory {
        mlp.mutex.Unlock()
        return fmt.Errorf("memory limit exceeded: %d + %d > %d", mlp.currentMemory, size, mlp.maxMemory)
    }
    mlp.currentMemory += size
    mlp.mutex.Unlock()

    defer func() {
        mlp.mutex.Lock()
        mlp.currentMemory -= size
        mlp.mutex.Unlock()
    }()

    // Read and process file
    content, err := os.ReadFile(path)
    if err != nil {
        return err
    }

    return processor(content)
}
```

### Concurrent Processing
```go
import "golang.org/x/sync/errgroup"

func (cg *ContextGenerator) processFilesConcurrently(files []*FileNode, maxWorkers int) error {
    g, ctx := errgroup.WithContext(context.Background())
    g.SetLimit(maxWorkers)

    for _, file := range files {
        file := file // Capture loop variable
        g.Go(func() error {
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                return cg.processFile(file)
            }
        })
    }

    return g.Wait()
}
```

This documentation provides all the essential patterns for implementing high-performance file scanning, ignore rule processing, tree generation, and context generation in Go, specifically optimized for the Shotgun-CLI requirements.