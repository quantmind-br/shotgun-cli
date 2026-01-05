# Data Flow Analysis

## Data Models

The system architecture revolves around several core data structures that facilitate the scanning, transformation, and delivery of project context:

- **FileNode**: A recursive tree structure representing the file system.
    - `Name`, `Path`, `RelPath`, `IsDir`, `Size`, `Children` (slice of `FileNode`).
    - `IsGitignored`, `IsCustomIgnored`: Boolean flags for filtering state.
- **ScanConfig**: Configuration for the filesystem crawler.
    - Limits: `MaxFileSize`, `MaxFiles`, `MaxMemory`.
    - Filters: `SkipBinary`, `IncludeHidden`, `IgnorePatterns`, `IncludePatterns`.
- **ContextData**: The aggregate model passed to the template engine.
    - `Task`: User-provided task description.
    - `Rules`: User-provided constraints/rules.
    - `FileStructure`: ASCII tree representation of the selected files.
    - `Files`: Slice of `FileContent` objects containing path and full text.
    - `CurrentDate`: Generated timestamp.
- **Template**: Structure representing a prompt template.
    - `Name`, `Description`, `Content`, `Source`.

## Input Sources

Data enters the system through multiple channels:

- **File System**: The primary source of data. The `scanner` module traverses the local directory to build the `FileNode` tree and read file contents.
- **CLI Arguments/Flags**: User preferences, model selection, and paths provided via `cobra` commands.
- **Interactive UI (TUI)**: User input for task descriptions, custom rules, and manual file selection/deselection via `bubbletea` screens.
- **Configuration Files**: `shotgun.yaml` and `.shotgunignore` files, as well as global config via `viper`.
- **Templates**: Pre-defined or user-provided Markdown templates (embedded, XDG config dir, or custom path).

## Data Transformations

The system performs several stages of data transformation:

1.  **Scanning to Tree**: Flat file system paths are transformed into a hierarchical `FileNode` tree.
2.  **Filtering**: The tree is pruned based on `.gitignore`, `.shotgunignore`, and user-defined patterns.
3.  **Context Assembly**: Selected files are read and bundled into a `ContextData` struct.
4.  **Template Rendering**: The `ContextData` is merged into a Markdown template.
    - Converts `{VARIABLE}` placeholders to `{{.Variable}}` Go template syntax.
    - Uses `text/template` for the final string interpolation.
5.  **Output Cleaning**: If sending to Gemini, the raw response from the `geminiweb` binary is processed to strip ANSI escape codes and parse structured blocks.

## Storage Mechanisms

The system is primarily transient and focused on "piping" data, but uses:

- **System Clipboard**: Temporary storage for the generated context string.
- **XDG Config Directory**: Local storage for user templates and configuration (`~/.config/shotgun-cli/`).
- **Memory Cache**: The `TemplateManager` maintains an in-memory map of loaded templates to avoid redundant I/O.
- **Gemini Cookies**: `geminiweb` stores authentication state in `~/.geminiweb/cookies.json`.

## Data Validation

Validation occurs at multiple points in the lifecycle:

- **Config Validation**: `validateConfig` ensures limits (MaxFileSize, etc.) are within sane bounds.
- **Template Validation**: Ensures templates contain required variables and valid syntax before rendering.
- **Token Estimation**: Counts tokens in the generated context to warn if it exceeds model limits.
- **Size Limits**: Final generated context is checked against `MaxTotalSize` before being output or copied.
- **Binary Detection**: Files are checked for binary content before reading to prevent corrupting the prompt context.

## Output Formats

The final processed data leaves the system in three main ways:

- **Standard Output (Stdout)**: The generated context or the LLM response is printed to the terminal.
- **System Clipboard**: The generated context is copied for manual pasting into LLM web interfaces.
- **External Process (Stdin)**: The generated context is piped into the `geminiweb` binary to interact with Google Gemini models.
- **Markdown Files**: Templates and generated results are typically formatted as Markdown for readability.