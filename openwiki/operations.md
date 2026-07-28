# Operations Guide

## Build

```bash
make build                    # → build/shotgun-cli (current platform)
make build-all                # cross-compile: linux/darwin/windows × amd64/arm64
make install                  # build + copy to ~/.local/bin
```

The binary is placed in `build/` by default. `make install` copies it to `$(HOME)/.local/bin`.

### Docker

```bash
docker build -t shotgun-cli .
docker run --rm -v $(pwd):/workspace shotgun-cli context generate
```

Source: `Makefile`, `Dockerfile`

---

## Test

```bash
# Unit tests
go test ./...
make test

# Tests with race detector (preferred for CI)
go test -race ./...
make test-race

# Single test
go test -v -run TestName_Scenario ./internal/core/scanner/...

# E2E tests (requires `make build` first)
make test-e2e

# Benchmarks
make test-bench

# Coverage (target: 85%+ overall, 90%+ new code)
make coverage
```

**Test conventions**:
- Table-driven tests with `t.Parallel()` and `stretchr/testify` (`require` for fatal, `assert` for soft checks)
- Naming: `TestFunc_Scenario`
- Three import groups (stdlib → external → internal), blank-line separated

Source: `Makefile`, `CLAUDE.md`, `AGENTS.md`

---

## Lint

```bash
make lint                     # uses .golangci.yml
golangci-lint run ./...       # direct invocation
```

**Strict rules**:
- Max line length: 120 characters
- Max cyclomatic complexity: 25
- Checks: errcheck, staticcheck, gosec, goconst, lll, prealloc, unparam, unused

Run lint before considering work done.

Source: `.golangci.yml`

---

## Configuration

### Configuration file

- **Linux/macOS**: `$XDG_CONFIG_HOME/shotgun-cli/config.yaml` (defaults to `~/.config/shotgun-cli/config.yaml`)
- **Windows**: `%APPDATA%\shotgun-cli\config.yaml`

### Configuration sources (priority order, highest first)

1. Command-line flags
2. Environment variables
3. Config file
4. Built-in defaults

### Key configuration keys

| Key | Default | Description |
|-----|---------|-------------|
| `llm.provider` | `gemini` | LLM provider (openai, anthropic, gemini) |
| `llm.api-key` | — | API key for the provider |
| `llm.base-url` | — | Custom API endpoint (OpenRouter, Azure, etc.) |
| `llm.model` | — | Model name |
| `llm.timeout` | `300` | Request timeout in seconds |
| `scanner.max-files` | `10000` | Maximum files to scan |
| `scanner.max-file-size` | `1MB` | Skip files larger than this |
| `scanner.max-memory` | `500MB` | Memory limit |
| `scanner.workers` | `0` (auto) | Parallel workers |
| `scanner.include-hidden` | `false` | Include dot-files |
| `scanner.include-ignored` | `true` | Include ignored files in scan (but they are still marked ignored) |
| `scanner.skip-binary` | `true` | Skip binary file content |
| `scanner.respect-gitignore` | `true` | Honor .gitignore rules |
| `scanner.respect-shotgunignore` | `true` | Honor .shotgunignore rules |
| `context.include-tree` | `true` | Include directory tree in output |
| `context.include-summary` | `true` | Include file summary |
| `context.max-size` | `128KB` | Max output size |

### CLI config commands

```bash
shotgun-cli config set llm.provider anthropic
shotgun-cli config set llm.api-key sk-...
shotgun-cli config get llm.provider

# Interactive config editor
shotgun-cli config --interactive
```

Source: `internal/config/keys.go`, `internal/config/metadata.go`, `internal/config/validator.go`, `cmd/config.go`, `cmd/root.go`

---

## CI

GitHub Actions workflows in `.github/workflows/`:

| Workflow | Trigger | Description |
|----------|---------|-------------|
| `test.yml` | push, PR | Tests, lint, coverage upload |
| `claude-code-review.yml` | PR | AI-powered code review via Claude |

Coverage is uploaded to Codecov. The test matrix covers multiple Go versions and OS platforms.

Source: `.github/workflows/test.yml`, `.github/workflows/claude-code-review.yml`

---

## Release

Releases use GoReleaser (`.goreleaser.yml`):

```bash
make release-snapshot        # Test release build locally
make release-tag             # Tag and push for release
make release-push            # Push to trigger goreleaser
```

Key build flags set via ldflags:
- `main.version` — set from git tag
- `main.commit` — set from git commit
- `main.date` — set from build date

Source: `.goreleaser.yml`, `Makefile`

---

## Logging

- Uses `zerolog` throughout — never `fmt.Println` or stdlib `log`
- Console writer to stderr for human-friendly output
- Error wrapping convention: `fmt.Errorf("doing X: %w", err)`

---

## Ignore patterns

Two ignore files control scanning:
- **`.shotgunignore`** — project-level custom ignore patterns (similar syntax to `.gitignore`)
- **`.gitignore`** — standard Git ignore rules (respected by default)

The ignore engine applies: explicit rules → built-in patterns → `.shotgunignore` → `.gitignore`.

Source: `internal/core/ignore/engine.go`

---

## Working with the codebase

### Pre-requisites
- Go 1.24+
- `golangci-lint` installed
- `make` available

### Common development flows

```bash
# Edit → test → lint cycle
go test -race ./internal/core/scanner/...
make lint
go test -race ./...

# Test specific package
go test -v -race -run TestScanner ./internal/core/scanner/

# Coverage check
make coverage
```

### Things to watch

- **Layer violations**: `core/` must not import project-internal packages. The build will fail if a `core/` package imports `internal/`.
- **Race conditions**: Always test with `-race`. The TUI coordinators (`generate_coordinator.go`, `scan_coordinator.go`) use mutexes to protect shared state.
- **Progress callbacks**: Not optional in the TUI path — dropping them stalls the wizard UI.
- **Error wrapping**: Use `fmt.Errorf("...: %w", err)` consistently.
- **Config keys**: Defined in `internal/config/keys.go` — never use raw string keys.
