
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

| Configuration Key | Default | Description |
| :--- | :--- | :--- |
| `gemini.model` | `gemini-2.5-flash` | Default Gemini model to use. |
| `gemini.timeout` | `300` | Default timeout in seconds. |
| `gemini.binary-path` | `""` | Explicit path to `geminiweb` binary. |
| `gemini.browser-refresh` | `auto` | Browser refresh strategy for authentication. |

**Endpoints Used**

The `geminiweb` tool abstracts the actual Gemini API endpoints. The CLI tool sends content to `geminiweb` via stdin, which then communicates with the Gemini API.

**Authentication Method**

*   **Type:** Delegated to `geminiweb` CLI tool.
*   **Mechanism:** The `geminiweb` tool handles authentication, likely using browser-based OAuth flows or API keys configured separately.
*   **Validation:** `shotgun-cli` checks for authentication readiness by verifying the existence of `~/.geminiweb/cookies.json`.

**Error Handling**

*   **Binary Not Found:** Clear error message with installation instructions.
*   **Not Configured:** Error message directing user to run `geminiweb auto-login`.
*   **Execution Timeout:** Context cancellation with timeout error message.
*   **API Errors:** Stderr from `geminiweb` is captured and displayed to the user.

**Retry/Circuit Breaker Configuration**

*   **Retry Logic:** Not implemented within `shotgun-cli`. Retries must be handled by the `geminiweb` tool or managed manually by the user.
*   **Circuit Breaker:** Not implemented.

### Integration Patterns

**Command Execution Pattern**

The integration uses Go's `os/exec` package to run the external `geminiweb` binary:

```go
cmd := exec.CommandContext(ctx, binaryPath, args...)
cmd.Stdin = strings.NewReader(content)
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr
err := cmd.Run()
```

**Response Processing Pipeline**

1. **Raw Capture:** Stdout and stderr are captured into byte buffers.
2. **ANSI Stripping:** ANSI escape codes are removed from the response.
3. **Response Parsing:** The clean text is parsed to extract the relevant LLM response.
4. **Duration Tracking:** Execution time is measured and reported.

**Configuration Management**

*   **Discovery:** The tool searches for the `geminiweb` binary in standard locations (PATH, GOPATH/bin, /usr/local/bin).
*   **Validation:** Pre-flight checks ensure the binary exists and is configured before attempting to send requests.
*   **Flexibility:** Users can override binary paths, models, and timeouts via command-line flags or configuration files.

## Available Documentation

### API Specifications and Integration Guides

| Document Path | Content Type | Quality Evaluation |
|---------------|--------------|-------------------|
| `README.md` | Project Overview | **Excellent** - Comprehensive setup and usage instructions |
| `.ai/docs/api_analysis.md` | API Documentation | **Excellent** - Detailed CLI command reference and external API integration |
| `.ai/docs/data_flow_analysis.md` | Data Flow Analysis | **Excellent** - Complete data transformation pipeline documentation |
| `.ai/docs/dependency_analysis.md` | Dependency Analysis | **Excellent** - Thorough dependency mapping and coupling assessment |
| `.ai/docs/request_flow_analysis.md` | Request Flow Analysis | **Excellent** - Complete command execution flow documentation |
| `.ai/docs/structure_analysis.md` | Architecture Analysis | **Excellent** - Comprehensive structural analysis with component mapping |

### Configuration and Setup Documentation

| Document Path | Content Type | Quality Evaluation |
|---------------|--------------|-------------------|
| `CLAUDE.md` | Claude AI Integration Guide | **Excellent** - Specific setup instructions for Claude integration |
| `GEMINI.md` | Gemini AI Integration Guide | **Excellent** - Detailed Gemini configuration and usage |
| `AGENTS.md` | AI Agent Configuration | **Good** - Agent setup and configuration guidelines |

### Workflow and Development Documentation

| Document Path | Content Type | Quality Evaluation |
|---------------|--------------|-------------------|
| `.claude/commands/` | Workflow Definitions | **Excellent** - Extensive collection of markdown files defining application capabilities |
| `.claude/commands/prp-*` | Project Management Workflows | **Excellent** - Well-structured project planning and execution workflows |
| `.claude/commands/development/` | Development Workflows | **Excellent** - Comprehensive development process documentation |
| `templates/` | Prompt Templates | **Excellent** - Concrete examples of tool usage patterns and expected outputs |

### Documentation Quality Assessment

**Strengths:**
*   **Comprehensive Coverage:** Documentation spans from high-level architecture to detailed command reference
*   **Practical Examples:** Rich with concrete usage examples and configuration samples
*   **Multiple Perspectives:** Covers both user-facing documentation and developer-focused architectural analysis
*   **AI Integration Focus:** Excellent documentation for AI provider integrations (Gemini, Claude)
*   **Workflow Documentation:** Extensive collection of predefined workflows and operational patterns

**Areas for Improvement:**
*   **OpenAPI/Swagger Specs:** Not applicable as this is a CLI tool, not a web service
*   **Interactive Documentation:** Could benefit from integrated help system improvements
*   **Error Code Reference:** Could include a comprehensive error code reference guide

**Overall Assessment:** The documentation quality is **excellent** for a CLI tool, providing comprehensive coverage of all aspects from setup to advanced usage patterns. The `.ai/docs/` directory contains particularly high-quality technical documentation that would be valuable for developers extending or integrating with the tool.