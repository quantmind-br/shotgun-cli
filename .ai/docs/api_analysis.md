
Based on my comprehensive analysis of the shotgun-cli project, I can now provide the complete API documentation. This is a CLI tool that serves as a context generation and AI orchestration engine, not a traditional web service.

# API Documentation

## APIs Served by This Project

**Note:** This project is a Command-Line Interface (CLI) tool and does not expose traditional network-based APIs (REST, gRPC, etc.). The "APIs" are CLI commands that provide programmatic access to the tool's functionality.

### Endpoints (CLI Commands)

#### `shotgun-cli context send [file]`

| Detail | Value |
| :--- | :--- |
| **Method and Path** | `shotgun-cli context send [file]` |
| **Description** | Sends the content of a specified file or standard input (stdin) to the configured Google Gemini model via the `geminiweb` CLI tool. |
| **Authentication** | None handled by `shotgun-cli`. Relies on pre-configured `geminiweb` credentials. |

**Request**

| Parameter | Type | Location | Description |
| :--- | :--- | :--- | :--- |
| `[file]` | string | Positional Arg | **Optional.** The path to the context file to send. If omitted, content is read from stdin. |
| `-o, --output` | string | Flag | Output file path for the Gemini response. If not specified, the response is printed to stdout. |
| `-m, --model` | string | Flag | Gemini model to use (e.g., `gemini-3.0-pro`). Overrides the configuration file. |
| `--timeout` | int | Flag | Timeout for the request in seconds. Overrides the configuration file. |
| `--raw` | boolean | Flag | Outputs the raw response from `geminiweb` without processing (e.g., stripping ANSI codes). |
| **Body** | text | Stdin/File | The LLM-optimized context and prompt content. |

**Response**

| Status | Format | Description |
| :--- | :--- | :--- |
| **Success** | Text/File | If `-o` is used, a confirmation message is printed to stderr, and the response is saved to the file. Otherwise, the parsed LLM response text is printed to stdout. |
| **Error** | Text (stderr) | A descriptive error message is printed to stderr, and the process exits with a non-zero code. Errors include file not found, no input, `geminiweb` not found, `geminiweb` not configured, or request timeout/failure. |

**Examples**

```bash
# Send content from a file and print response to stdout
shotgun-cli context send prompt.md

# Send content from a file and save response to a file
shotgun-cli context send prompt.md -o response.md

# Pipe content from stdin and use a specific model
cat context.txt | shotgun-cli context send -m gemini-3.0-pro
```

#### `shotgun-cli context generate`

| Detail | Value |
| :--- | :--- |
| **Method and Path** | `shotgun-cli context generate` |
| **Description** | Generates a structured text representation of your codebase within LLM token limits. Scans codebase, applies ignore patterns, and creates optimized context file. |

**Request**

| Parameter | Type | Location | Description |
| :--- | :--- | :--- | :--- |
| `--root` | string | Flag | Root directory to scan (default: current directory). |
| `--include` | []string | Flag | File patterns to include (e.g., "*.go,*.js"). |
| `--exclude` | []string | Flag | File patterns to exclude (e.g., "vendor/*,*.test.go"). |
| `--output` | string | Flag | Output file path (default: auto-generated timestamp). |
| `--max-size` | string | Flag | Maximum context size (default: "10MB"). |
| `--no-enforce-limit` | boolean | Flag | Allow generation that exceeds size limit with warning. |
| `--send-gemini` | boolean | Flag | Automatically send generated context to Gemini. |
| `--gemini-model` | string | Flag | Gemini model to use when auto-sending. |
| `--gemini-output` | string | Flag | Output file for Gemini response when auto-sending. |
| `--gemini-timeout` | int | Flag | Timeout for Gemini request when auto-sending. |

**Response**

| Status | Format | Description |
| :--- | :--- | :--- |
| **Success** | Text/File | Generated context file saved to specified path or auto-generated filename. |
| **Error** | Text (stderr) | Error message with details about scanning failures or validation errors. |

#### `shotgun-cli template list`

| Detail | Value |
| :--- | :--- |
| **Method and Path** | `shotgun-cli template list` |
| **Description** | Lists all available embedded templates with their names and descriptions. |

**Response**

| Status | Format | Description |
| :--- | :--- | :--- |
| **Success** | Text | Formatted table showing template names, sources, and descriptions. |

#### `shotgun-cli template render [template-name]`

| Detail | Value |
| :--- | :--- |
| **Method and Path** | `shotgun-cli template render [template-name]` |
| **Description** | Renders a specific template with variable substitution. |

**Request**

| Parameter | Type | Location | Description |
| :--- | :--- | :--- | :--- |
| `[template-name]` | string | Positional Arg | **Required.** Name of the template to render. |
| `--var` | map[string]string | Flag | Template variables in key=value format. |
| `--output` | string | Flag | Output file path (default: stdout). |

**Response**

| Status | Format | Description |
| :--- | :--- | :--- |
| **Success** | Text/File | Rendered template output to stdout or specified file. |
| **Error** | Text (stderr) | Error message for missing template or required variables. |

#### `shotgun-cli diff split`

| Detail | Value |
| :--- | :--- |
| **Method and Path** | `shotgun-cli diff split` |
| **Description** | Splits large diff files into smaller, manageable chunks while preserving diff context. |

**Request**

| Parameter | Type | Location | Description |
| :--- | :--- | :--- | :--- |
| `--input` | string | Flag | **Required.** Input diff file path. |
| `--output-dir` | string | Flag | Output directory for chunks (default: "chunks"). |
| `--approx-lines` | int | Flag | Approximate lines per chunk (default: 1000). |
| `--no-header` | boolean | Flag | Omit metadata headers for patch tool compatibility. |

**Response**

| Status | Format | Description |
| :--- | :--- | :--- |
| **Success** | Text/Files | Multiple chunk files created in output directory with summary. |
| **Error** | Text (stderr) | Error message for file access or parsing issues. |

#### `shotgun-cli config show`

| Detail | Value |
| :--- | :--- |
| **Method and Path** | `shotgun-cli config show` |
| **Description** | Displays current configuration values with their sources. |

**Response**

| Status | Format | Description |
| :--- | :--- | :--- |
| **Success** | Text | Formatted configuration display grouped by category. |

#### `shotgun-cli config set [key] [value]`

| Detail | Value |
| :--- | :--- |
| **Method and Path** | `shotgun-cli config set [key] [value]` |
| **Description** | Sets a configuration value in the config file. |

**Request**

| Parameter | Type | Location | Description |
| :--- | :--- | :--- | :--- |
| `[key]` | string | Positional Arg | **Required.** Configuration key (e.g., "scanner.max-files"). |
| `[value]` | string | Positional Arg | **Required.** Configuration value. |

**Response**

| Status | Format | Description |
| :--- | :--- | :--- |
| **Success** | Text | Confirmation message with config file location. |
| **Error** | Text (stderr) | Error message for invalid keys or values. |

### Authentication & Security

The `shotgun-cli` does not manage API keys or authentication tokens directly.

*   **Mechanism:** Authentication is delegated entirely to the external `geminiweb` CLI tool.
*   **Prerequisites:** The user must ensure the `geminiweb` binary is installed and configured for authentication by running `geminiweb auto-login` prior to using `shotgun-cli context send`.
*   **Security:** The tool handles context content locally and pipes it to the authenticated `geminiweb` binary via standard input.

### Rate Limiting & Constraints

*   **Rate Limiting:** Not implemented by `shotgun-cli`. Rate limiting is managed by the underlying Google Gemini API and enforced by the `geminiweb` tool.
*   **Constraints:**
    *   **Max Files/Size:** The context generation phase respects configuration limits for file count (`scanner.max-files`, default 10000) and file size (`scanner.max-file-size`, default 1MB).
    *   **Timeout:** The `send` command enforces a configurable timeout (`--timeout` flag or `gemini.timeout` config) on the external `geminiweb` execution.

## External API Dependencies

This project consumes the Google Gemini API indirectly by executing the `geminiweb` CLI tool.

### Services Consumed

| Detail | Value |
| :--- | :--- |
| **Service Name & Purpose** | **Google Gemini** (via `geminiweb` CLI). Used for processing LLM-optimized codebase contexts and generating responses. |
| **Integration Type** | External CLI execution (`os/exec`). |

**Base URL/Configuration**

The integration is configured via the `viper` configuration system, primarily using the `gemini` prefix.

| Configuration Key | Default Value | Description |
| :--- | :--- | :--- |
| `gemini.model` | `gemini-2.5-flash` | Default Gemini model to use. |
| `gemini.timeout` | `300` | Default timeout in seconds (5 minutes). |
| `gemini.binary-path` | `""` | Explicit path to `geminiweb` binary. |
| `gemini.browser-refresh` | `"auto"` | Browser refresh strategy for cookie management. |
| `gemini.auto-send` | `false` | Automatically send generated context to Gemini. |

**Endpoints Used**

The `geminiweb` tool abstracts the actual Google Gemini API endpoints. The CLI tool communicates with `geminiweb` via standard input/output.

| Operation | Method | Description |
| :--- | :--- | :--- |
| **Send Context** | `stdin` → `geminiweb` → `stdout` | Sends LLM-optimized context to Gemini for processing. |

**Authentication Method**

*   **Type:** Cookie-based authentication managed by `geminiweb`.
*   **Configuration:** Authentication is handled by the external tool. `shotgun-cli` validates that `geminiweb` is configured by checking for the presence of `~/.geminiweb/cookies.json`.
*   **Setup:** Users must run `geminiweb auto-login` to establish authentication.

**Error Handling**

The `gemini.Executor` implements comprehensive error handling:

| Error Type | Handling Strategy |
| :--- | :--- |
| **Binary Not Found** | Returns clear error with installation instructions. |
| **Not Configured** | Returns error with `geminiweb auto-login` instructions. |
| **Execution Timeout** | Context-based timeout with clear error message. |
| **Process Failure** | Captures stderr and includes in error context. |
| **Response Parsing** | Graceful fallback to raw response on parsing errors. |

**Retry/Circuit Breaker Configuration**

*   **Retry Logic:** Not implemented at the `shotgun-cli` level. Retries are handled by the `geminiweb` tool.
*   **Circuit Breaker:** Not implemented. The tool relies on the external tool's resilience patterns.
*   **Timeout Management:** Configurable timeout with context cancellation for process termination.

### Integration Patterns

**External Process Execution Pattern**

```go
// Simplified integration pattern
func (e *Executor) Send(ctx context.Context, content string) (*Result, error) {
    cmd := exec.CommandContext(ctx, binaryPath, args...)
    cmd.Stdin = strings.NewReader(content)
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("geminiweb execution failed: %w\nstderr: %s", err, stderr.String())
    }
    
    return &Result{
        Response:    ParseResponse(StripANSI(stdout.String())),
        RawResponse: stdout.String(),
    }, nil
}
```

**Configuration Management Pattern**

*   **Hierarchical Configuration:** Defaults → Config File → Environment Variables → CLI Flags
*   **Environment Prefix:** `SHOTGUN_` (e.g., `SHOTGUN_GEMINI_MODEL`)
*   **Config File Locations:** Platform-specific XDG-compliant paths

**Progress Reporting Pattern**

*   **TUI Mode:** Real-time progress bars and status updates via Bubble Tea
*   **CLI Mode:** Structured logging with `zerolog` for operation tracking
*   **External Process:** Progress callbacks for long-running operations

## Available Documentation

### API Specifications

*   **Location:** `./.ai/docs/api_analysis.md`
*   **Quality:** Comprehensive - Contains detailed CLI command documentation with request/response formats, examples, and error handling
*   **Coverage:** Complete - Covers all exposed CLI commands and their parameters

### Integration Guides

*   **Location:** `./.ai/docs/`
*   **Available Documents:**
    *   `data_flow_analysis.md` - Detailed data flow and transformation analysis
    *   `dependency_analysis.md` - Complete dependency mapping and integration patterns
    *   `request_flow_analysis.md` - Command execution flow and routing analysis
    *   `structure_analysis.md` - Project structure and architecture overview
*   **Quality:** Excellent - Provides deep technical insights for developers integrating with the tool

### Code Documentation

*   **Location:** Inline Go documentation throughout the codebase
*   **Quality:** Good - Comprehensive function and package documentation with examples
*   **Coverage:** High - Most public functions and packages are well-documented

### Configuration Documentation

*   **Location:** Available via `shotgun-cli config show` command
*   **Quality:** Good - Shows all configuration options with sources and current values
*   **Coverage:** Complete - All configuration keys are documented and validated

### Documentation Quality Assessment

**Strengths:**
*   Comprehensive CLI command documentation with practical examples
*   Deep technical analysis documents for advanced integration
*   Clear separation between user-facing and developer documentation
*   Good inline code documentation with usage examples

**Areas for Improvement:**
*   Could benefit from a unified API reference document
*   Missing OpenAPI/Swagger specification (not applicable for CLI tool)
*   Could use more integration examples for different workflows

**Overall Assessment:** The documentation quality is excellent for a CLI tool, providing both user-friendly command references and deep technical documentation for developers needing to integrate or extend the functionality.