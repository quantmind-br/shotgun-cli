# Shotgun CLI

[![Tests](https://github.com/quantmind-br/shotgun-cli/actions/workflows/test.yml/badge.svg)](https://github.com/quantmind-br/shotgun-cli/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/quantmind-br/shotgun-cli/graph/badge.svg)](https://codecov.io/gh/quantmind-br/shotgun-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/quantmind-br/shotgun-cli)](https://goreportcard.com/report/github.com/quantmind-br/shotgun-cli)

## Project Overview

**Shotgun CLI** is a sophisticated Command-Line Interface tool written in Go that functions as a **Context Generation and AI Orchestration Engine**. The tool bridges complex codebases with AI language models, providing both interactive (TUI) and programmatic (CLI) interfaces for generating LLM-optimized codebase contexts and facilitating AI-assisted development workflows.

### Purpose and Main Functionality

The primary purpose of Shotgun CLI is to transform complex codebases into structured, LLM-optimized contexts that can be sent to AI models like Google Gemini. It intelligently scans file systems, applies layered ignore patterns, and generates comprehensive context files that include file structure representations and relevant content.

### Key Features and Capabilities

- **Interactive TUI Wizard**: 5-step guided workflow using Bubble Tea framework for intuitive user interaction
- **CLI Commands**: Programmatic interface for automation and scripting
- **Intelligent File Scanning**: Recursive directory traversal with layered ignore rule processing
- **Template Management**: Multi-source template loading with variable substitution
- **Context Optimization**: Token estimation and size validation for LLM compatibility
- **AI Integration**: Seamless integration with Google Gemini API via external tool execution
- **Cross-platform Support**: Works across different operating systems with platform-specific optimizations

### Likely Intended Use Cases

- **Code Review Automation**: Generate comprehensive context for AI-powered code analysis
- **Documentation Generation**: Create structured representations of codebases for AI-assisted documentation
- **Refactoring Assistance**: Provide AI models with complete context for intelligent refactoring suggestions
- **Onboarding Tools**: Help new developers understand complex codebase structures
- **Code Migration**: Facilitate AI-assisted code migration between languages or frameworks

## Table of Contents

- [Architecture](#architecture)
- [LLM Provider Configuration](#llm-provider-configuration)
- [Config Commands](#config-commands)
- [CMD Helper Functions](#cmd-helper-functions)
- [TUI Wizard Usage](#tui-wizard-usage)
- [TUI Wizard Helper Functions](#tui-wizard-helper-functions)
- [TUI Wizard State Transitions](#tui-wizard-state-transitions)
- [C4 Model Architecture](#c4-model-architecture)
- [Repository Structure](#repository-structure)
- [Dependencies and Integration](#dependencies-and-integration)
- [API Documentation](#api-documentation)
- [Development Notes](#development-notes)
- [Known Issues and Limitations](#known-issues-and-limitations)
- [Additional Documentation](#additional-documentation)

## Architecture

### High-level Architecture Overview

Shotgun CLI implements a **Clean Architecture/Hexagonal Architecture** pattern with clear separation of concerns across three main layers:

1. **Presentation/Adapter Layer** (`cmd`, `internal/ui`): Handles user interaction through CLI commands and interactive TUI wizard
2. **Application Layer** (`internal/app`): Orchestrates business logic and provides a unified API (`ContextService`) for presentation layers
3. **Core/Domain Layer** (`internal/core`): Contains pure business logic for context generation, file scanning, template management, and token estimation
4. **Infrastructure/Platform Layer** (`internal/platform`): Implements external system integrations (LLM providers, `http` client, `geminiweb`, clipboard operations)

### Technology Stack and Frameworks

- **Language**: Go
- **CLI Framework**: Cobra (`github.com/spf13/cobra`)
- **Configuration**: Viper (`github.com/spf13/viper`)
- **TUI Framework**: Bubble Tea (`github.com/charmbracelet/bubbletea`)
- **Logging**: Zerolog (`github.com/rs/zerolog`)
- **Template Engine**: Go standard library templates
- **Ignore Processing**: go-gitignore (`github.com/sabhiram/go-gitignore`)

### Component Relationships

```mermaid
graph TD
    A[main.go] → B[cmd]
    B → O[internal/app]
    O → C[internal/core/context]
    O → D[internal/core/scanner]
    O → E[internal/core/template]
    O → H[internal/platform/llm]
    
    B → C
    B → D
    B → E
    B → F[internal/core/ignore]
    B → G[internal/core/tokens]
    B → H
    B → I[internal/platform/clipboard]
    B → J[internal/ui/wizard]
    B → K[internal/utils]
    
    C → D
    C → E
    C → F
    C → G
    
    D → F
    
    J → SC[internal/ui/scan_coordinator]
    J → GC[internal/ui/generate_coordinator]
    J → O
    J → L[internal/ui/screens]
    J → M[internal/ui/components]
    
    SC → D
    GC → C
    
    L → M
    L → N[internal/ui/styles]
    
    M → N
    
    style A fill:#e1f5fe
    style B fill:#f3e5f5
    style O fill:#fff9c4
    style C fill:#e8f5e8
    style D fill:#e8f5e8
    style E fill:#e8f5e8
    style F fill:#e8f5e8
    style G fill:#e8f5e8
    style H fill:#fff3e0
    style I fill:#fff3e0
    style J fill:#fce4ec
    style K fill:#f5f5f5
```

### Key Design Patterns

- **Command Pattern**: CLI command structure in `cmd` package
- **Builder/Generator Pattern**: Context generation in `internal/core/context`
- **Strategy Pattern**: AI provider abstraction for multi-provider support
- **MVU Pattern**: TUI state management with Bubble Tea
- **Template Method Pattern**: Standardized template rendering process
- **Layered Architecture**: Clear separation between presentation, business logic, and infrastructure
- **Factory Pattern**: Template manager and scanner creation

### Platform Layer Architecture

The Platform layer provides a standardized way for LLM providers to communicate with external APIs.

#### Shared HTTP Client

Most LLM providers use the shared `JSONClient` located in `internal/platform/http/client.go`. This client provides:

- **Standardized Requests**: Using `PostJSON()` for consistent API interaction
- **Error Handling**: Standardized `HTTPError` type that captures status codes and response bodies
- **Configuration**: Unified timeout and base URL handling

#### Provider Implementation

By using the shared client, LLM providers (OpenAI, Anthropic, GeminiAPI) only need to implement:
1. **Request Building**: Constructing the provider-specific JSON request structure
2. **Response Mapping**: Defining the target structure for JSON unmarshaling
3. **Error Handling**: Mapping `HTTPError` to provider-specific error messages
4. **Authentication**: Setting the required headers (e.g., `Authorization`, `x-api-key`)

## LLM Provider Configuration

shotgun-cli supports multiple LLM providers including OpenAI, Anthropic, Google Gemini API, and GeminiWeb (browser-based).

### GeminiWeb Provider

GeminiWeb is a browser-based integration that uses the `geminiweb` CLI tool for Google Gemini access without requiring an API key.

#### Setup

1. **Install geminiweb**:
   ```bash
   go install github.com/diogo/geminiweb/cmd/geminiweb@latest
   ```

2. **Configure authentication** (browser-based):
   ```bash
   geminiweb auto-login
   ```
   This will open a browser window for Google authentication and store cookies in `~/.geminiweb/cookies.json`.

3. **Configure shotgun-cli**:
   ```bash
   shotgun-cli config set llm.provider geminiweb
   shotgun-cli config set gemini.enabled true
   ```

#### Cookies File Setup

The `geminiweb` tool stores authentication cookies in:
- **Linux/macOS**: `~/.geminiweb/cookies.json`
- **Windows**: `%USERPROFILE%\.geminiweb\cookies.json`

The cookies file is created automatically when you run `geminiweb auto-login`. If authentication expires, simply run the command again to refresh the cookies.

#### Verification

Check if the provider is configured correctly:
```bash
shotgun-cli llm status
shotgun-cli llm doctor
```

#### Configuration Options

- `llm.provider`: Set to `geminiweb`
- `gemini.binary-path`: Optional custom path to geminiweb binary
- `gemini.browser-refresh`: Auto-refresh cookies using browser (`auto`, `chrome`, `firefox`, `edge`, etc.)
- `gemini.model`: Model to use (default: `gemini-2.5-flash`)
- `gemini.timeout`: Request timeout in seconds (default: `300`)

For more information on LLM providers, see `.serena/memories/llm_providers.md`.

### LLM Diagnostic Commands

Shotgun CLI provides diagnostic commands to help troubleshoot and verify LLM provider configuration.

#### `shotgun-cli llm status`

Display the current LLM provider configuration and status.

```bash
shotgun-cli llm status
```

**Output format**:
```
=== LLM Configuration ===

Provider:  anthropic
Model:     claude-sonnet-4-20250514
Base URL:  (default: https://api.anthropic.com)
API Key:   sk-a...-key
Timeout:   300s

Status: Ready
```

The status command shows:
- **Provider**: Currently selected LLM provider
- **Model**: Model name configured for use
- **Base URL**: API endpoint (shows default or custom URL)
- **API Key**: Masked API key for security
- **Timeout**: Request timeout in seconds
- **Status**: One of `Ready`, `Not configured`, `Not available`, or `Not ready`

#### `shotgun-cli llm doctor`

Run diagnostics on the LLM provider configuration and provide specific guidance for fixing issues.

```bash
shotgun-cli llm doctor
```

**Output format**:
```
Running diagnostics for anthropic...

Checking provider... anthropic
Checking API key... configured
Checking model... claude-sonnet-4-20250514
Checking provider availability... OK
Checking provider configuration... OK

No issues found! anthropic is ready.
```

The doctor command checks:
- Provider type is valid
- API key is configured (if required)
- Model is set
- Provider availability
- Provider configuration completeness

When issues are found, the doctor provides specific next steps for each provider:
- **OpenAI**: API key setup link and configuration commands
- **Anthropic**: API key setup link and configuration commands
- **Gemini**: API key setup link and configuration commands
- **GeminiWeb**: Installation and authentication steps

#### `shotgun-cli llm list`

List all supported LLM providers with descriptions and configuration information.

```bash
shotgun-cli llm list
```

**Output format**:
```
Supported LLM Providers:

  openai      - OpenAI (GPT-4o, GPT-4, o1, o3)
* anthropic   - Anthropic (Claude 4, Claude 3.5)
  gemini      - Google Gemini (Gemini 2.5, Gemini 2.0)
  geminiweb   - GeminiWeb (Browser-based (no API key))

Configure with:
  shotgun-cli config set llm.provider <provider>
  shotgun-cli config set llm.api-key <your-api-key>

For custom endpoints (OpenRouter, Azure, etc.):
  shotgun-cli config set llm.base-url https://openrouter.ai/api/v1
```

The current provider is marked with `*` in the list.

## Config Commands

Shotgun CLI provides a configuration system built on Viper that allows users to customize scanner behavior, LLM settings, and output preferences.

### Configuration File Location

The configuration file is stored at:
- **Linux/macOS**: `$XDG_CONFIG_HOME/shotgun-cli/config.yaml` (defaults to `~/.config/shotgun-cli/config.yaml`)
- **Windows**: `%APPDATA%\shotgun-cli\config.yaml`

### Configuration Sources

Configuration values are loaded from multiple sources in order of priority (highest to lowest):

1. **Command-line flags**: Highest priority, override all other sources
2. **Environment variables**: Override config file values
3. **Config file**: Persistent settings stored in `config.yaml`
4. **Defaults**: Built-in default values used if no other source specifies a value

### Available Commands

#### `shotgun-cli config show`

Display current configuration values with their sources.

```bash
shotgun-cli config show
```

**Output format** (human-readable):
```
scanner.max-files: 1000 (default)
scanner.max-file-size: 10MB (default)
llm.provider: anthropic (config file)
gemini.enabled: true (config file)
output.format: markdown (config file)
```

**Output format** (JSON):
```bash
shotgun-cli config show --format json
```

#### `shotgun-cli config set <key> <value>`

Set a configuration value. The value is validated and written to the config file.

```bash
# Set scanner max files
shotgun-cli config set scanner.max-files 5000

# Set LLM provider
shotgun-cli config set llm.provider openai

# Set API key
shotgun-cli config set llm.api-key sk-...

# Set Gemini model
shotgun-cli config set gemini.model gemini-2.5-pro
```

**Validation**: Values are validated before being saved. Invalid values will return an error:
```bash
$ shotgun-cli config set scanner.max-files invalid
Error: failed to parse integer value
```

### Configuration Keys

#### Scanner Settings

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `scanner.max-files` | int | 1000 | Maximum number of files to scan |
| `scanner.max-file-size` | size | 10MB | Maximum size per file (e.g., 10MB, 500KB) |
| `scanner.respect-gitignore` | bool | true | Respect .gitignore files during scanning |
| `scanner.skip-binary` | bool | true | Skip binary files during scanning |
| `scanner.workers` | int | 4 | Number of parallel scanner workers (1-32) |
| `scanner.include-hidden` | bool | false | Include hidden files (starting with .) |
| `scanner.include-ignored` | bool | false | Include git-ignored files |
| `scanner.respect-shotgunignore` | bool | true | Respect .shotgunignore files |
| `scanner.max-memory` | size | 100MB | Maximum memory usage for scanning |

#### Context Settings

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `context.max-size` | size | 10MB | Maximum size of generated context (e.g., 1MB, 500KB) |
| `context.include-tree` | bool | true | Include file tree in context |
| `context.include-summary` | bool | true | Include file summary in context |

#### Template Settings

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `template.custom-path` | path | - | Custom path to template directory |

#### Output Settings

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `output.format` | string | markdown | Output format: `markdown` or `text` |
| `output.clipboard` | bool | false | Copy generated context to clipboard |

#### LLM Provider Settings

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `llm.provider` | string | - | LLM provider: `openai`, `anthropic`, `gemini`, `geminiweb` |
| `llm.api-key` | string | - | API key for the provider |
| `llm.base-url` | URL | - | Custom base URL for API requests |
| `llm.model` | string | - | Model name to use |
| `llm.timeout` | int | 300 | Request timeout in seconds (1-3600) |

#### Gemini Integration Settings

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `gemini.enabled` | bool | false | Enable Gemini integration |
| `gemini.binary-path` | path | geminiweb | Path to geminiweb binary |
| `gemini.model` | string | gemini-2.5-flash | Gemini model to use |
| `gemini.timeout` | int | 120 | Gemini request timeout in seconds |
| `gemini.browser-refresh` | string | auto | Browser for cookie refresh: `auto`, `chrome`, `firefox`, `edge`, `chromium`, `opera` |
| `gemini.auto-send` | bool | false | Automatically send to Gemini after generation |
| `gemini.save-response` | bool | false | Save Gemini response to file |

### Configuration Validation

The configuration system provides centralized validation through `internal/config/validator.go`. All values are validated before being saved to the configuration file.

#### Validation Functions

| Key | Validator | Rules | Error Messages |
|-----|-----------|-------|----------------|
| `scanner.max-files` | `validateMaxFiles` | Positive integer, rejects size formats | "expected a positive integer", "expected a number, got size format", "must be positive" |
| `scanner.max-file-size` | `validateSizeFormat` | Size format (KB/MB/GB/B) or plain number | "expected size format (e.g., 1MB, 500KB)" |
| `scanner.workers` | `validateWorkers` | Integer between 1 and 32 | "must be between 1 and 32", "expected a positive integer" |
| `scanner.respect-gitignore` | `validateBooleanValue` | "true" or "false" (case-insensitive) | "expected 'true' or 'false'" |
| `scanner.skip-binary` | `validateBooleanValue` | "true" or "false" (case-insensitive) | "expected 'true' or 'false'" |
| `scanner.include-hidden` | `validateBooleanValue` | "true" or "false" (case-insensitive) | "expected 'true' or 'false'" |
| `scanner.include-ignored` | `validateBooleanValue` | "true" or "false" (case-insensitive) | "expected 'true' or 'false'" |
| `scanner.respect-shotgunignore` | `validateBooleanValue` | "true" or "false" (case-insensitive) | "expected 'true' or 'false'" |
| `scanner.max-memory` | `validateSizeFormat` | Size format (KB/MB/GB/B) or plain number | "expected size format (e.g., 1MB, 500KB)" |
| `context.max-size` | `validateSizeFormat` | Size format (KB/MB/GB/B) or plain number | "expected size format (e.g., 1MB, 500KB)" |
| `context.include-tree` | `validateBooleanValue` | "true" or "false" (case-insensitive) | "expected 'true' or 'false'" |
| `context.include-summary` | `validateBooleanValue` | "true" or "false" (case-insensitive) | "expected 'true' or 'false'" |
| `template.custom-path` | `validatePath` | Valid path (empty allowed) | "failed to expand home directory", "parent path exists but is not a directory" |
| `output.format` | `validateOutputFormat` | "markdown" or "text" | "expected 'markdown' or 'text'" |
| `output.clipboard` | `validateBooleanValue` | "true" or "false" (case-insensitive) | "expected 'true' or 'false'" |
| `llm.provider` | `validateLLMProvider` | openai, anthropic, gemini, geminiweb | "expected one of: openai, anthropic, gemini, geminiweb" |
| `llm.api-key` | None | Any string | N/A |
| `llm.base-url` | `validateURL` | Empty or starts with http:// or https:// | "URL must start with http:// or https://" |
| `llm.model` | None | Any string (provider-specific validation) | N/A |
| `llm.timeout` | `validateTimeout` | Integer between 1 and 3600 seconds | "timeout must be positive", "timeout too large (max 3600 seconds)" |
| `gemini.enabled` | `validateBooleanValue` | "true" or "false" (case-insensitive) | "expected 'true' or 'false'" |
| `gemini.binary-path` | `validatePath` | Valid path (empty allowed) | "failed to expand home directory", "parent path exists but is not a directory" |
| `gemini.model` | `validateGeminiModel` | gemini-2.5-flash, gemini-2.5-pro, gemini-3.0-pro | "expected one of: gemini-2.5-flash, gemini-2.5-pro, gemini-3.0-pro" |
| `gemini.timeout` | `validateTimeout` | Integer between 1 and 3600 seconds | "timeout must be positive", "timeout too large (max 3600 seconds)" |
| `gemini.browser-refresh` | `validateBrowserRefresh` | Empty, auto, chrome, firefox, edge, chromium, opera | "expected one of: auto, chrome, firefox, edge, chromium, opera (or empty to disable)" |
| `gemini.auto-send` | `validateBooleanValue` | "true" or "false" (case-insensitive) | "expected 'true' or 'false'" |
| `gemini.save-response` | `validateBooleanValue` | "true" or "false" (case-insensitive) | "expected 'true' or 'false'" |

#### Validation Rules Detail

**Integer Validation** (`scanner.max-files`, `scanner.workers`):
- Must be a valid integer format (e.g., `100`, `5000`)
- Workers: Range 1-32
- Max-files: Must be positive
- Max-files specifically rejects size formats like "10MB" or "1KB"

**Size Format Validation** (`scanner.max-file-size`, `context.max-size`, `scanner.max-memory`):
- Supports suffixes: `KB`, `MB`, `GB`, `B`
- Also accepts plain numbers (bytes)
- Examples: `100`, `1KB`, `10MB`, `1GB`, `500KB`
- Case-insensitive suffixes
- Uses `utils.ParseSize()` for parsing

**Boolean Validation** (all `*enabled`, `*include*`, `*skip*`, `*respect*`, `auto-send`, `save-response`, `clipboard`):
- Only accepts: `true` or `false`
- Case-insensitive: `true`, `True`, `TRUE`, `false`, `False`, `FALSE`
- Does NOT accept: `yes`, `no`, `1`, `0`, `on`, `off`

**String Enum Validation**:
- `llm.provider`: `openai`, `anthropic`, `gemini`, `geminiweb`
- `output.format`: `markdown`, `text`
- `gemini.model`: `gemini-2.5-flash`, `gemini-2.5-pro`, `gemini-3.0-pro`
- `gemini.browser-refresh`: `""`, `auto`, `chrome`, `firefox`, `edge`, `chromium`, `opera` (case-sensitive)

**Path Validation** (`template.custom-path`, `gemini.binary-path`):
- Empty string is allowed
- Expands `~/` to home directory
- Parent directory must exist or be creatable
- Validates that existing parent is a directory

**URL Validation** (`llm.base-url`):
- Empty string is allowed
- Must start with `http://` or `https://`
- Basic URL format validation only

**Timeout Validation** (`llm.timeout`, `gemini.timeout`):
- Must be a positive integer
- Range: 1-3600 seconds (1 hour max)

#### Type Conversion

The `ConvertValue()` function converts validated string values to their appropriate Go types:

| Input Type | Conversion | Result Type |
|------------|------------|-------------|
| Integer keys | `fmt.Sscanf` | `int` |
| Boolean keys | `strings.ToLower == "true"` | `bool` |
| All other keys | Identity | `string` |

**Integer Keys**: `scanner.max-files`, `scanner.workers`, `llm.timeout`, `gemini.timeout`

**Boolean Keys**: All `*enabled`, `*include*`, `*skip*`, `*respect*`, `auto-send`, `save-response`, `clipboard` keys

#### Testing

Configuration validation is tested in `internal/config/validator_test.go`:

- `TestIsValidKey`: Valid key detection
- `TestValidKeys`: No duplicates in valid keys list
- `TestValidateValue_*`: Per-key validation tests (workers, max-files, size format, boolean, provider, format, timeout, URL, browser-refresh, path, model)
- `TestConvertValue_*`: Type conversion tests
- `TestValidatePath_*`: Path validation with existing files
- `TestValidateBrowserRefresh_Direct`: Direct validator function tests
- `TestValidateGeminiModel`: Gemini model validation

All validation tests use `t.Parallel()` for efficient execution.

### Gemini Status

The `getGeminiStatusSummary` function provides a status summary for Gemini integration:

- **disabled**: Gemini integration is disabled
- **✗ geminiweb not found**: geminiweb binary not in PATH
- **✓ configured**: Gemini is properly configured
- **⚠ needs configuration**: Gemini enabled but not fully configured

Check status with:
```bash
shotgun-cli llm status
```

### Testing

Configuration functions are tested in `cmd/config_test.go`:
- `TestShowCurrentConfig_*`: Display configuration in various formats
- `TestSetConfigValue_*`: Set and validate configuration values
- `TestGetDefaultConfigPath_*`: Config file path resolution
- `TestGetGeminiStatusSummary_*`: Gemini status detection
- `TestGetConfigSource_*`: Configuration source detection

## CMD Helper Functions

The CMD package (`cmd/`) contains helper functions used throughout the CLI commands. This section documents key helper functions for formatting, display, and progress reporting.

### Overview

CMD helper functions provide utility functionality for:
- **Duration Formatting**: Converting time durations to human-readable strings
- **URL Display**: Showing configuration URLs with appropriate defaults
- **Progress Reporting**: Outputting progress in human or JSON format

### Key Helper Functions

#### `formatDuration(d time.Duration) string`

**Location**: `cmd/send.go`

**Purpose**: Formats a duration for display in the CLI output.

**Behavior**:
- Durations < 1 second: Returns milliseconds (e.g., "500ms")
- Durations >= 1 second: Returns seconds with 1 decimal place (e.g., "1.5s")

**Examples**:
```go
formatDuration(500 * time.Millisecond)  // "500ms"
formatDuration(1500 * time.Millisecond) // "1.5s"
formatDuration(60 * time.Second)        // "60.0s"
formatDuration(5 * time.Minute)         // "300.0s"
```

**Use Case**: Displaying API request duration after sending content to an LLM.

**Tests**: `cmd/send_test.go` - 13 test cases covering milliseconds, seconds, and minutes

---

#### `displayURL(url string, provider llm.ProviderType) string`

**Location**: `cmd/llm.go`

**Purpose**: Displays a base URL with appropriate default fallback for each provider.

**Behavior**:
- If URL is empty and provider has a default BaseURL: Returns "(default: <url>)"
- If URL is empty and provider has no default BaseURL: Returns "(default)"
- If URL is provided: Returns the URL as-is

**Examples**:
```go
displayURL("", llm.ProviderOpenAI)     // "(default: https://api.openai.com/v1)"
displayURL("", llm.ProviderAnthropic)  // "(default: https://api.anthropic.com)"
displayURL("", llm.ProviderGeminiWeb)  // "(default)"
displayURL("https://custom.proxy.com", llm.ProviderOpenAI) // "https://custom.proxy.com"
```

**Use Case**: Showing LLM endpoint configuration in status output.

**Tests**: `cmd/llm_test.go` - 10 test cases covering all providers and custom URLs

---

#### Progress Reporting Functions

##### `renderProgressHuman(p ProgressOutput)`

**Location**: `cmd/context.go`

**Purpose**: Renders progress output in human-readable format to stderr.

**Behavior**:
- With Total > 0: Outputs `[Stage] Message: Current/Total (Percent%)`
- Without Total: Outputs `[Stage] Message`

**Output Format**:
```
[scanning] Processing files: 50/100 (50.0%)
[generating] Creating context
```

**Tests**: `cmd/context_test.go` - Existing tests for various progress states

---

##### `renderProgressJSON(p ProgressOutput)`

**Location**: `cmd/context.go`

**Purpose**: Renders progress output as JSON to stderr (one line per event).

**Behavior**: Marshals the `ProgressOutput` struct to JSON and outputs to stderr.

**Output Format**:
```json
{"timestamp":"2024-01-01T12:00:00Z","stage":"scanning","message":"Processing files","current":50,"total":100,"percent":50}
```

**Use Case**: Programmatic progress monitoring in CI/CD pipelines.

**Tests**: `cmd/context_test.go` - 4 test cases covering full progress, partial progress, and edge cases

---

##### `renderProgress(mode ProgressMode, p ProgressOutput)`

**Location**: `cmd/context.go`

**Purpose**: Routes progress output to the appropriate renderer based on mode.

**Behavior**:
- `ProgressHuman`: Calls `renderProgressHuman()`
- `ProgressJSON`: Calls `renderProgressJSON()`
- `ProgressNone`: No output

**Example**:
```go
renderProgress(ProgressHuman, progress)  // Human-readable output
renderProgress(ProgressJSON, progress)   // JSON output
renderProgress(ProgressNone, progress)   // No output
```

**Tests**: `cmd/context_test.go` - 4 test cases covering all modes

---

### ProgressOutput Struct

**Location**: `cmd/context.go`

```go
type ProgressOutput struct {
    Timestamp string  `json:"timestamp"`
    Stage     string  `json:"stage"`
    Message   string  `json:"message"`
    Current   int64   `json:"current,omitempty"`
    Total     int64   `json:"total,omitempty"`
    Percent   float64 `json:"percent,omitempty"`
}
```

### Testing

CMD helper functions are tested in:
- `cmd/send_test.go`: `TestFormatDuration`
- `cmd/llm_test.go`: `TestDisplayURL`
- `cmd/context_test.go`: `TestRenderProgressJSON`, `TestRenderProgress`, and existing `TestRenderProgressHuman_*` tests

All tests use stdout/stderr capture to verify output format.

### Cross-References

- **CMD Package**: `cmd/`
- **LLM Types**: `internal/core/llm/`
- **Context Commands**: "Context Commands" section

## TUI Wizard Usage

The TUI Wizard provides an interactive 5-step workflow for generating LLM-optimized codebase contexts. This section covers keyboard shortcuts, terminal requirements, and usage tips.

### Terminal Requirements

**Minimum terminal size: 40 columns x 10 rows**

If your terminal window is too small, the wizard will display a warning overlay asking you to resize. The warning shows your current dimensions and the minimum required size.

```
Terminal too small

Current:  30x8
Required: 40x10

Please resize your terminal
```

### Keyboard Shortcuts

#### Global Navigation

| Key | Action |
|-----|--------|
| F1 | Toggle help screen |
| F7 / Ctrl+P | Previous step |
| F8 / Ctrl+N | Next step |
| Ctrl+Q | Quit application |

#### File Selection (Step 1)

| Key | Action |
|-----|--------|
| ↑/↓ or k/j | Navigate up/down |
| ←/→ or h/l | Collapse/Expand directory |
| Space | Toggle selection (file or directory) |
| a | Select all visible files |
| A | Deselect all visible files |
| i | Toggle showing ignored files |
| / | Enter filter mode (fuzzy search) |
| Ctrl+C | Clear filter |
| F5 | Rescan directory |

**Filter Mode**: When a filter is active, the status bar displays the match count in the format `X/Y files` (e.g., "12/45 files"), showing how many files match the filter out of the total available files.

#### Template Selection (Step 2)

| Key | Action |
|-----|--------|
| ↑/↓ or k/j | Navigate templates |
| Enter | Select template |
| v | View full template (opens modal) |

**Template Preview Modal**:

| Key | Action |
|-----|--------|
| j/k | Scroll up/down |
| PgUp/PgDown | Page scroll |
| g/G | Jump to top/bottom |
| Esc/q | Close modal |

#### Text Input (Steps 3-4)

| Key | Action |
|-----|--------|
| Type | Enter text |
| Enter | New line |
| Backspace | Delete character |

#### Review (Step 5)

| Key | Action |
|-----|--------|
| F8 | Generate context |
| c | Copy to clipboard |
| F9 | Send to LLM (if configured) |

### Visual Feedback

- **Loading Spinner**: During directory scanning, an animated spinner is displayed with "Scanning directory..." message
- **Progress Indicators**: Progress bars show scan and generation progress with current/total counts
- **Filter Match Count**: When filtering files, the stats bar shows "X/Y files" indicating matches vs total

## TUI Wizard Helper Functions

The TUI Wizard (`internal/ui/wizard.go`) implements a 5-step interactive workflow using Bubble Tea. This section documents the key helper functions that power the wizard's internal operations.

### Overview

The wizard follows the MVU (Model-View-Update) pattern from Bubble Tea, with helper functions handling specific aspects of the workflow:

- **State Management**: Tracking the current step and wizard state
- **Scan Operations**: File system scanning and tree building
- **Generation Operations**: Context generation and file output
- **Message Handling**: Processing Bubble Tea messages
- **Validation**: Ensuring data integrity before state transitions

### TUI Coordinator Pattern

The TUI Wizard uses the "Model of Models" pattern with dedicated coordinators for asynchronous operations. This separates the complex state management of scanning and generation from the main UI logic.

#### ScanCoordinator

Manages the file system scanning state machine in `internal/ui/scan_coordinator.go`.

- **Start**: Initiates async scan with `Start(rootPath, config)`
- **Poll**: Checks for progress updates via `Poll()`
- **Result**: Returns `(*scanner.FileNode, error)` via `Result()`
- **State**: Tracks `started`, `done`, and `progress` channels

#### GenerateCoordinator

Manages the context generation state machine in `internal/ui/generate_coordinator.go`.

- **Start**: Initiates async generation with `Start(config)`
- **Poll**: Checks for progress updates via `Poll()`
- **Result**: Returns `(string, error)` via `Result()`
- **State**: Tracks generation progress and content buffering

#### Message Flow

Both coordinators follow the Bubble Tea command pattern:

1. **Start**: `wizard.Update` calls `coordinator.Start()` → returns `tea.Cmd`
2. **Poll**: `coordinator.Poll()` checks channels and returns `Batch(Msg, NextPoll)`
3. **Progress**: UI receives progress messages (`ScanProgressMsg`, `GenerationProgressMsg`)
4. **Completion**: Coordinator signals completion, UI retrieves result via `Result()`

### Message Handler Functions

##### `handleTemplateMessage(msg tea.Msg) tea.Cmd`

**Purpose**: Routes template-related messages to the template selection component.

**Signature**: `func (m *WizardModel) handleTemplateMessage(msg tea.Msg) tea.Cmd`

**Behavior**:
- Only processes messages when in `StepTemplateSelection`
- Delegates to `templateSelection.HandleMessage()` if component exists
- Returns nil otherwise (ignores message)

**Error Handling**: No explicit errors; relies on template selection component.

**Example**:
```go
cmd := wizard.handleTemplateMessage(TemplateSelectedMsg{
    Template: &template.Template{Name: "code-review"},
})
```

**Related Tests**: `TestWizardHandleTemplateMessage`, `TestWizardHandleTemplateMessage_WrongStep`, `TestWizardHandleTemplateMessage_NilTemplateSelection`, `TestWizardHandleTemplateMessage_CorrectStepWithSelection`

### Testing Approach

The wizard helper functions are tested comprehensively in `internal/ui/wizard_test.go`:

1. **Unit Tests**: Each helper function has dedicated tests covering success and error paths
2. **Table-Driven Tests**: Complex functions use table-driven tests for multiple scenarios
3. **Parallel Execution**: Tests use `t.Parallel()` for efficient execution
4. **State Validation**: Tests verify internal state changes and message types
5. **Error Simulation**: Tests simulate error conditions using mocked state

**Coverage**: The wizard package has comprehensive test coverage for all helper functions, including edge cases like nil state, empty content, and size validation.

### Common Usage Patterns

#### Starting a Scan

```go
// Initialize scan state
model, cmd := wizard.Update(startScanMsg{
    rootPath: rootPath,
    config:   scanConfig,
})

// Process progress messages
model, _ = wizard.Update(ScanProgressMsg{
    Current: 50,
    Total:   100,
    Stage:   "scanning",
})

// Complete scan
model, _ = wizard.Update(ScanCompleteMsg{
    Tree: fileTree,
})
```

#### Starting a Generation

```go
// Initialize generation state
model, cmd := wizard.Update(startGenerationMsg{
    fileTree:      wizard.fileTree,
    selectedFiles: wizard.selectedFiles,
    template:      wizard.template,
    taskDesc:      wizard.taskDesc,
    rules:         wizard.rules,
    rootPath:      rootPath,
})

// Process progress
model, _ = wizard.Update(GenerationProgressMsg{
    Stage:   "render",
    Message: "Rendering template",
})

// Complete generation
model, _ = wizard.Update(GenerationCompleteMsg{
    Content:  generatedContent,
    FilePath: outputPath,
})
```

### Cross-References

- **Test File**: `internal/ui/wizard_test.go`
- **Wizard Implementation**: `internal/ui/wizard.go`
- **Screen Components**: `internal/ui/screens/`
- **Architecture Memory**: `.serena/memories/architecture.md`

## TUI Wizard State Transitions

The TUI Wizard implements a state machine that guides users through a 5-step interactive workflow. This section documents the state transitions, message flows, and testing patterns used to ensure reliable wizard behavior.

### Wizard State Machine

The wizard maintains a linear progression through five states:

| Step | Constant | Screen | Purpose |
|------|----------|--------|---------|
| 1 | `StepFileSelection` | File Selection | Select root directory and configure scan options |
| 2 | `StepTemplateSelection` | Template Selection | Choose a prompt template for generation |
| 3 | `StepTaskInput` | Task Input | Describe the task/context for generation |
| 4 | `StepRulesInput` | Rules Input | (Optional) Add specific rules or constraints |
| 5 | `StepReview` | Review | Review selections and trigger generation |

### State Transition Logic

**Forward Movement**: Users progress through steps using:
- **Enter**: Confirm and move to next step
- **Tab**: Navigate between interactive elements
- **Ctrl+N**: Skip to next step (when applicable)

**Backward Movement**: Users can return to previous steps using:
- **Esc**: Go back to previous step
- **Ctrl+B**: Explicit back navigation

**State Guards**:
- Cannot proceed from Step 1 without valid file selection
- Cannot proceed from Step 2 without template selection
- Cannot proceed from Step 3 with empty task description
- Step 4 (Rules) is optional - can be skipped with empty rules
- Step 5 requires successful scan completion before generation

### Iterative Command Patterns

The wizard uses Bubble Tea's command pattern for asynchronous operations. Key iterative patterns include:

#### Scan Iterative Pattern

```
User Action → startScanMsg → scanner.Scan() (iterative)
                              ↓
                         ScanProgressMsg → UI Update
                              ↓
                         ScanCompleteMsg → finishScan() → Store Result
```

**Components**:
- `startScanMsg`: Initiates scan with root path and config
- `scanner.Scan()`: Returns `tea.Cmd` that yields progress messages
- `ScanProgressMsg`: Updates UI with current/total file count
- `ScanCompleteMsg`: Finalizes scan with file tree result
- `ScanErrorMsg`: Handles scan failures

#### Generation Iterative Pattern

```
User Action → startGenerationMsg → context.Generate() (iterative)
                                        ↓
                                   GenerationProgressMsg → UI Update
                                        ↓
                                   GenerationCompleteMsg → finishGeneration() → Write File
```

**Components**:
- `startGenerationMsg`: Initiates generation with template and context
- `context.Generate()`: Returns `tea.Cmd` that yields progress messages
- `GenerationProgressMsg`: Updates UI with generation stage
- `GenerationCompleteMsg`: Finalizes generation with content and file path
- `GenerationErrorMsg`: Handles generation failures

### Message Types for State Transitions

| Message | Type | Source | Handler | Purpose |
|---------|------|--------|---------|---------|
| `startScanMsg` | Internal | Step 1 | `handleStartScan` | Trigger file scan |
| `ScanProgressMsg` | Internal | Scanner | `handleScanProgress` | Update scan UI |
| `ScanCompleteMsg` | Internal | Scanner | `handleScanComplete` | Store scan results |
| `ScanErrorMsg` | Internal | Scanner | `handleScanError` | Display scan failure |
| `startGenerationMsg` | Internal | Step 5 | `handleStartGeneration` | Trigger generation |
| `GenerationProgressMsg` | Internal | Generator | `handleGenerationProgress` | Update generation UI |
| `GenerationCompleteMsg` | Internal | Generator | `handleGenerationComplete` | Store generation results |
| `GenerationErrorMsg` | Internal | Generator | `handleGenerationError` | Display generation failure |
| `StepBackMsg` | Key User | Keyboard | `Update` | Navigate to previous step |
| `StepNextMsg` | Key User | Keyboard | `Update` | Navigate to next step |
| `QuitMsg` | Key User | Keyboard | `Update` | Exit wizard |

### Example Test Coverage Table

| Function | Test Coverage | Test Cases |
|----------|---------------|------------|
| `handleStartScan` | 100% | 3 cases |
| `finishScan` | 100% | 3 cases (success, error, nil state) |
| `handleStartGeneration` | 100% | 4 cases |
| `finishGeneration` | 100% | 4 cases (success, empty, size error, nil) |
| `validateContentSize` | 100% | 10 cases (boundaries, errors, empty) |
| `parseSize` | 100% | 28 cases (valid, invalid, edge) |
| `handleTemplateMessage` | 100% | 3 cases (wrong step, nil, correct) |

### Testing Patterns

#### State-Based Testing

Tests verify wizard state at each transition point:

```go
func TestWizardStateTransitions(t *testing.T) {
    wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil)

    // Initial state
    require.Equal(t, StepFileSelection, wizard.step)

    // Transition to Step 2
    wizard.step = StepTemplateSelection
    require.Equal(t, StepTemplateSelection, wizard.step)
}
```

#### Message-Driven Testing

Tests verify correct message handling for each state:

```go
func TestWizardHandleScanComplete(t *testing.T) {
    wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil)
    wizard.scanState = &scanState{}

    msg := ScanCompleteMsg{
        Tree: &scanner.FileNode{Name: "root", Path: "/tmp", IsDir: true},
    }

    cmd := wizard.handleScanComplete(msg)
    require.NotNil(t, cmd)

    // Verify state updated
    require.NotNil(t, wizard.fileTree)
}
```

#### Command Result Testing

Tests verify that commands return expected message types:

```go
func TestWizardFinishScan(t *testing.T) {
    wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil)
    wizard.scanState = &scanState{
        result: &scanner.FileNode{Name: "root"},
    }

    cmd := wizard.finishScan()
    msg := cmd()

    scanComplete, ok := msg.(ScanCompleteMsg)
    require.True(t, ok, "command should return ScanCompleteMsg")
    require.NotNil(t, scanComplete.Tree)
}
```

#### Error Simulation

Tests verify error handling paths:

```go
func TestWizardFinishScan_WithError(t *testing.T) {
    wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil)
    wizard.scanState = &scanState{
        scanErr: errors.New("scan failed"),
    }

    cmd := wizard.finishScan()
    msg := cmd()

    scanErr, ok := msg.(ScanErrorMsg)
    require.True(t, ok, "command should return ScanErrorMsg on error")
    require.Contains(t, scanErr.Error(), "scan failed")
}
```

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        Wizard Model                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │   Step 1     │───▶│   Step 2     │───▶│   Step 3     │      │
│  │  File Select │    │  Template    │    │   Task Input │      │
│  └──────────────┘    └──────────────┘    └──────────────┘      │
│         │                                       │               │
│         │ startScanMsg                          │               │
│         ▼                                       │               │
│  ┌──────────────┐                              │               │
│  │ Scan Process │◀─────────────┐                │               │
│  │  (Iterative) │               │                │               │
│  └──────────────┘               │                │               │
│         │                        │                │               │
│         │ ScanProgressMsg        │                │               │
│         ▼                        │                │               │
│  ┌──────────────┐                │                │               │
│  │ ScanComplete │───────────────▶│                │               │
│  └──────────────┘                │                │               │
│                                   │                │               │
│                                   │                ▼               │
│                                   │    ┌──────────────┐           │
│                                   │    │   Step 4     │           │
│                                   │    │  Rules Input │           │
│                                   │    └──────────────┘           │
│                                   │         │                    │
│                                   │         │ (optional)         │
│                                   │         ▼                    │
│                                   │    ┌──────────────┐           │
│                                   │    │   Step 5     │           │
│                                   └───▶│    Review    │           │
│                                        └──────────────┘           │
│                                               │                    │
│                                               │ startGenerationMsg │
│                                               ▼                    │
│                                        ┌──────────────┐           │
│                                        │ Generation   │           │
│                                        │  (Iterative) │           │
│                                        └──────────────┘           │
│                                               │                    │
│                                               │ GenerationComplete │
│                                               ▼                    │
│                                        ┌──────────────┐           │
│                                        │  File Output │           │
│                                        └──────────────┘           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Cross-References

- **Wizard Implementation**: `internal/ui/wizard.go`
- **Test Suite**: `internal/ui/wizard_test.go`
- **Helper Functions**: "TUI Wizard Helper Functions" section
- **Scanner Package**: `internal/scanner/`
- **Context Package**: `internal/context/`

## TUI LLM Integration

The TUI provides integrated LLM functionality that allows users to send generated context to AI models directly from the wizard interface. This section documents the LLM integration architecture, configuration, and testing.

### LLM Integration Architecture

The LLM integration follows a layered architecture where the UI delegates orchestration to the application layer:

```
┌─────────────────────────────────────────────────────────────────┐
│                    TUI Wizard (Review Screen)                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Generated Content ────▶  [Send to LLM] Button                  │
│                           │                                     │
│                           ▼                                     │
│  ┌──────────────────────────────────────────────────────┐      │
│  │           handleSendToLLM()                          │      │
│  │  • Validates wizard state                            │      │
│  │  • Prepares LLMSendConfig                            │      │
│  │  • Calls svc.SendToLLMWithProgress()                 │      │
│  └──────────────────────────────────────────────────────┘      │
│                           │                                     │
│                           ▼                                     │
│  ┌──────────────────────────────────────────────────────┐      │
│  │           app.ContextService                         │      │
│  │  • Orchestrates provider creation                    │      │
│  │  • Handles progress reporting via callback           │      │
│  │  • Manages response saving                           │      │
│  │  • Returns LLMCompleteMsg or LLMErrorMsg            │      │
│  └──────────────────────────────────────────────────────┘      │
│                           │                                     │
│                           ▼                                     │
│                     Response File                              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### LLM Provider Configuration

The wizard supports multiple LLM providers through a unified configuration:

| Provider | Config Key | Required Fields | Optional Fields |
|----------|-----------|-----------------|-----------------|
| OpenAI | `openai` | `api_key` | `base_url`, `model`, `timeout` |
| Anthropic | `anthropic` | `api_key` | `base_url`, `model`, `timeout` |
| Gemini | `gemini` | `api_key` | `base_url`, `model`, `timeout` |
| GeminiWeb | `geminiweb` | (none - uses browser auth) | `binary_path`, `browser_refresh`, `model` |

**Configuration Example**:

```yaml
llm:
  provider: "anthropic"
  api_key: "sk-ant-..."
  model: "claude-sonnet-4-20250514"
  timeout: 300
  save_response: true
```

**GeminiWeb Legacy Config** (automatically merged):

```yaml
gemini:
  enabled: true
  binary_path: "/path/to/geminiweb"
  model: "gemini-2.0-pro"
  timeout: 300
  browser_refresh: "auto"
  save_response: true
```

### LLM Integration Message Flow

> **Note**: Message types were renamed in a recent refactoring to be provider-agnostic.

| Message | Type | Source | Handler | Purpose |
|---------|------|--------|---------|---------|
| `LLMProgressMsg` | Internal | Service | `handleLLMProgress` | Update progress UI |
| `LLMCompleteMsg` | Internal | Service | `handleLLMComplete` | Store response, update UI |
| `LLMErrorMsg` | Internal | Service | `handleLLMError` | Display error, update UI |
| `RescanRequestMsg` | Key User | Review | `handleRescanRequest` | Trigger new scan |

### Service Delegation

The wizard now delegates all LLM operations to the `app.ContextService`. The `NewWizard` constructor accepts this service as a dependency.

```go
// Using ContextService for LLM operations in the wizard
svc := app.NewContextService()
wizard := wizard.NewWizard(rootPath, scanConfig, templateMgr, svc)
```

### Send Flow

1. **User Action**: User presses "Send to LLM" on review screen
2. **Validation**: `handleSendToLLM()` validates:
   - Wizard is on review step
   - Generated content exists
   - Not already sending
3. **Execution**: `svc.SendToLLMWithProgress()` is called:
   - Orchestrates the entire send process
   - Progress callback updates UI during send
   - On success: Saves response (if configured)
   - Returns completion message with response file path

### Test Coverage

| Function | Test Coverage | Test Cases |
|----------|---------------|------------|
| `handleSendToLLM` | 100% | 5 cases (step validation, states, errors) |
| `handleRescanRequest` | 100% | 3 cases (file selection, other steps, all steps) |
| `handleLLMProgress` | 100% | 1 case |
| `handleLLMComplete` | 100% | 1 case |
| `handleLLMError` | 100% | 2 cases |

### Error Handling

The LLM integration handles various error scenarios through the service layer:

| Error | Source | User Experience |
|-------|---------|-----------------|
| Invalid provider | Service | Returns error, displayed in review |
| Provider unavailable | Service | Error shown, send prevented |
| Not configured | Service | Error shown, send prevented |
| Send timeout | Service | Returns `LLMErrorMsg` |
| Save failure | Service | Returns error with context |

### Testing Examples

#### Service Usage Example

```go
// Using ContextService for LLM operations
svc := app.NewContextService()
result, err := svc.SendToLLMWithProgress(ctx, content, app.LLMSendConfig{
    Provider:     llm.ProviderOpenAI,
    APIKey:       "your-api-key",
    Model:        "gpt-4o",
    SaveResponse: true,
    OutputPath:   "./response.md",
}, func(stage string) {
    fmt.Printf("Progress: %s\n", stage)
})
```

#### Send Handler Test

```go
func TestWizardHandleSendToLLM_NotReviewStep(t *testing.T) {
    svc := &mockContextService{}
    wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, svc)
    wizard.step = StepFileSelection  // Wrong step
    wizard.generatedContent = "content"
    
    cmd := wizard.handleSendToLLM()
    assert.Nil(t, cmd)  // Should return nil for wrong step
}
```

### Cross-References

- **LLM Package**: `internal/core/llm/`
- **Application Service**: `internal/app/context.go`
- **Wizard LLM Tests**: `internal/ui/wizard_test.go`
- **LLM Commands Documentation**: "LLM Diagnostic Commands" section

## C4 Model Architecture

### Context Diagram

</arg_value>
</tool_call>