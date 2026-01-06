# Data Flow Analysis

## Data Models

The system centers on several key data structures that facilitate the transformation of codebase information into LLM prompts:

*   **`scanner.FileNode`**: A hierarchical representation of the project's file system. It stores metadata such as path, name, directory status, and ignore flags (`IsGitignored`, `IsCustomIgnored`).
*   **`context.FileContent`**: Represents a single file's data for context generation.
    *   `Path`, `RelPath`: Location of the file.
    *   `Language`: Detected programming language.
    *   `Content`: Full text content of the file.
    *   `Size`: File size in bytes.
*   **`context.ContextData`**: The primary data object used for rendering templates. It aggregates the task description, rules, file structure (ASCII tree), and the list of `FileContent` objects.
*   **`llm.Result`**: Standardized output from any LLM provider, containing the processed response, raw API response, model used, and usage metrics (tokens, duration).
*   **`llm.Config`**: Encapsulates provider settings like API keys, base URLs, model names, and timeouts.

## Input Sources

Data enters the system through multiple channels:

*   **Local Filesystem**: The primary source of data. The system recursively scans the directory tree starting from the current working directory or a specified path.
*   **TUI Wizard (Interactive)**:
    *   **File Selection**: Users interactively toggle files/directories in the `FileNode` tree.
    *   **Template Selection**: Users choose from predefined or custom Markdown prompt templates.
    *   **Task/Rules Input**: Users provide textual descriptions of the task and specific constraints.
*   **CLI Arguments & Stdin (Headless)**:
    *   Existing context files can be passed as arguments.
    *   Piped content via `stdin` allows for integration with other tools.
*   **Configuration Files**: YAML files (managed via Viper) provide persistent settings for scanners, LLM providers, and display preferences.
*   **Environment Variables**: Prefix `SHOTGUN_` variables can override configuration (e.g., `SHOTGUN_LLM_API_KEY`).

## Data Transformations

Information undergoes several transformation stages:

1.  **Scanning & Filtering**: The raw filesystem is walked to build a `FileNode` tree. Files are filtered based on ignore patterns (.gitignore, .shotgunignore), size limits, and binary checks.
2.  **Language Detection**: Based on file extensions and basenames (e.g., `Dockerfile`, `go.mod`), the system assigns a language tag to each file for better syntax highlighting in the prompt.
3.  **Context Assembly**: Selected files and metadata are gathered into a `ContextData` struct. A tree renderer converts the hierarchical `FileNode` structure into an ASCII representation.
4.  **Template Rendering**:
    *   **Variable Substitution**: Placeholders like `{TASK}` and `{RULES}` are replaced with user-provided text.
    *   **Go Template Execution**: The system uses `text/template` for complex rendering, including logic for looping over files and conditional blocks.
5.  **LLM Request Mapping**: The final context string is wrapped into provider-specific JSON structures (e.g., OpenAI's `ChatCompletionRequest`) before being sent via HTTP.

## Storage Mechanisms

The application is primarily stateless but utilizes the following for persistence and temporary storage:

*   **Configuration Files**: Stored in `~/.config/shotgun-cli/config.yaml` or a user-specified path.
*   **Template Files**: Loaded from embedded assets, `~/.config/shotgun-cli/templates/`, or a custom path.
*   **Local Output Files**: Users can explicitly save generated contexts or LLM responses to `.md` files.
*   **System Clipboard**: The generated context can be programmatically copied to the system clipboard for use in web-based LLM interfaces.
*   **Cache**: A hidden `.ai/analysis_cache.json` exists for caching analysis results (used by internal agents).

## Data Validation

Validation occurs at several boundaries:

*   **Configuration Validation**: `validateConfig` checks for valid file counts, size limits, and required provider settings (API keys).
*   **File Integrity**: `isTextFile` peeks at the first 1024 bytes of a file to detect binary content (null bytes or invalid UTF-8) before full ingestion.
*   **Template Validation**: `validateVariables` ensures that all required variables for a selected template are provided by the user before generation.
*   **Provider Validation**: `ValidateConfig` in LLM clients checks for the presence of necessary credentials and model availability.
*   **Size Constraints**: The system enforces `MaxTotalSize` during both file collection and the final context generation to ensure prompts fit within LLM context windows.

## Output Formats

Data leaves the system in several formats:

*   **Markdown**: The primary output format for codebase contexts, utilizing headers, lists, and code blocks.
*   **JSON**: Used for API communication with LLM providers (OpenAI, Anthropic, Gemini).
*   **Interactive TUI**: Real-time progress updates and review screens rendered using the Bubble Tea framework.
*   **Raw Terminal Output**: In headless mode, the context or LLM response is printed directly to `stdout`.
*   **Clipboard**: UTF-8 string data copied to the system's clipboard buffer.