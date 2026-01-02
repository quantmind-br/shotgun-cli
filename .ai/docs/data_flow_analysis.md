# Data Flow Analysis

## Data Models

The system uses several core data structures to manage the flow of context information:

*   **`FileNode` (Scanner)**: A recursive tree structure representing the project's file system. It includes metadata like path, size, and ignore status (`IsGitignored`, `IsCustomIgnored`).
*   **`FileContent` (Context)**: Represents the actual content of a file, including its relative path, detected programming language, and the raw string content.
*   **`Template` (Template)**: Contains the structure for the prompt sent to the LLM. It includes metadata (name, description) and `RequiredVars` (placeholders like `{TASK}`, `{RULES}`, `{FILE_STRUCTURE}`).
*   **`ContextData` (Generator)**: The final data assembly used for template rendering. It aggregates user inputs (`Task`, `Rules`), the rendered `FileStructure` (ASCII tree + file blocks), and system metadata like `CurrentDate`.
*   **`Result` (Gemini)**: Holds the output from the LLM execution, containing the response text and performance metrics (duration).

## Input Sources

Data enters the system through three primary channels:

1.  **Local File System**: The `Scanner` component traverses the project directory, reading file metadata and contents.
2.  **User Input (CLI/TUI)**: Users provide operational instructions via command-line flags and interactive UI screens (tasks, custom rules, file selections, and template choices).
3.  **Embedded/Local Templates**: Prompt templates are loaded from either embedded assets or a local `templates/` directory.
4.  **External LLM (Gemini)**: The system receives processed responses back from the Gemini LLM via the `geminiweb` CLI utility.

## Data Transformations

Data undergoes several stages of transformation before reaching the output:

1.  **Scanning & Filtering**: The raw file system is transformed into a `FileNode` tree. Files are filtered based on `.gitignore`, `.shotgunignore`, and binary detection.
2.  **Tree Rendering**: The `FileNode` tree is transformed into an ASCII representation for inclusion in the prompt.
3.  **Language Detection**: Filenames are mapped to programming languages (e.g., `.go` -> `go`, `Dockerfile` -> `dockerfile`) for better LLM context.
4.  **Template Variable Substitution**: The `Renderer` performs string replacement, injecting `ContextData` into `{VARIABLE}` placeholders. It also converts `{VAR}` syntax to Go-style `{{.Var}}` if necessary.
5.  **ANSI Stripping**: Responses from the `geminiweb` utility are cleaned of ANSI escape codes and parsed to extract the meaningful LLM response.

## Storage Mechanisms

The application is primarily a transient CLI tool, but it interacts with the following storage:

*   **Memory**: Most data (scanned trees, file contents, rendered templates) resides in memory during the application lifecycle.
*   **System Clipboard**: The generated context or LLM response can be persisted to the system clipboard for user use in other applications.
*   **Local Filesystem**: Templates are stored as `.md` files. The tool also reads configuration from files like `.shotgunignore`.
*   **Shell Pipe/Stdio**: Data is streamed between the CLI and external processes like `geminiweb`.

## Data Validation

Validation occurs at multiple points to ensure integrity:

*   **File Limits**: Checks against `MaxFileSize`, `MaxFiles`, and `MaxTotalSize` during content collection to prevent memory issues or LLM context window overflows.
*   **Binary Detection**: A "peek" mechanism reads the first 1024 bytes to verify if a file is text-based before full ingestion.
*   **Template Validation**: Templates are checked for unmatched braces and malformed variable patterns.
*   **Required Variables**: The renderer ensures that all variables defined in a template are supplied by the system or user before execution.
*   **UTF-8 Verification**: Validates that file content is valid UTF-8 to avoid processing corrupt or incompatible data.

## Output Formats

The system produces data in several formats:

*   **Markdown**: The primary format for generated prompts and LLM responses.
*   **ASCII Tree**: A visual representation of the project structure.
*   **Structured CLI Output**: Progress updates and results are printed to `stdout`.
*   **Clipboard Content**: The final rendered prompt or LLM response is copied to the clipboard.
*   **JSON (Internal/Debug)**: Various structures include JSON tags for potential logging or debugging exports.