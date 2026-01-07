# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds



## Architecture & Application Logic

### Application Service Layer

The project uses a dedicated application layer in `internal/app` to orchestrate business logic shared between CLI and TUI.

*   **`ContextService` (internal/app/context.go)**: The primary orchestrator for context generation.
    *   `Generate(ctx, cfg)`: Synchronous generation (used by CLI).
    *   `GenerateWithProgress(ctx, cfg, callback)`: Async generation with progress updates (used by TUI).
    *   `SendToLLM(ctx, content, provider)`: Standardized way to send content to an LLM.

*   **`DefaultProviderRegistry` (internal/app/providers.go)**: Centralized registry for LLM providers.
    *   To add a new provider: Implement `llm.Provider`, then register it in `internal/app/providers.go`.
    *   Both CLI and TUI consume this registry to create providers.

### Core Logic Additions

*   **`internal/core/diff`**: Pure business logic for git diff operations. Use `diff.IntelligentSplit` to split large diffs while preserving file boundaries.
*   **`internal/core/scanner`**: Use `scanner.CollectSelections` or `scanner.NewSelectAll` for handling file selections in a tree.

---

## CI/CD and Testing

### Automated Pipeline

Every push and PR runs the following checks via GitHub Actions:

1. **Tests**: All tests with race detection and coverage
2. **Coverage**: Must be >= 85% or CI fails
3. **Lint**: golangci-lint checks
4. **Build**: Compilation verification

### Running Tests Locally

```bash
# Run all tests
go test ./...

# Run with race detection (recommended)
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total
```

### Coverage Requirements

- **Minimum threshold**: 85% (enforced by CI)
- **Target for new code**: 90%
- **Core packages target**: 95%

Before submitting a PR:
1. Run tests locally
2. Check coverage meets threshold
3. Run linter: `golangci-lint run`

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed testing guidelines.

---

<!-- BEGIN BEADS INTEGRATION -->
## Issue Tracking with bd (beads)

**IMPORTANT**: This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

### Why bd?

- Dependency-aware: Track blockers and relationships between issues
- Git-friendly: Auto-syncs to JSONL for version control
- Agent-optimized: JSON output, ready work detection, discovered-from links
- Prevents duplicate tracking systems and confusion

### Quick Start

**Check for ready work:**

```bash
bd ready --json
```

**Create new issues:**

```bash
bd create "Issue title" --description="Detailed context" -t bug|feature|task -p 0-4 --json
bd create "Issue title" --description="What this issue is about" -p 1 --deps discovered-from:bd-123 --json
```

**Claim and update:**

```bash
bd update bd-42 --status in_progress --json
bd update bd-42 --priority 1 --json
```

**Complete work:**

```bash
bd close bd-42 --reason "Completed" --json
```

### Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks
- `chore` - Maintenance (dependencies, tooling)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

### Workflow for AI Agents

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task**: `bd update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" --description="Details about what was found" -p 1 --deps discovered-from:<parent-id>`
5. **Complete**: `bd close <id> --reason "Done"`

### Auto-Sync

bd automatically syncs with git:

- Exports to `.beads/issues.jsonl` after changes (5s debounce)
- Imports from JSONL when newer (e.g., after `git pull`)
- No manual export/import needed!

### Important Rules

- ✅ Use bd for ALL task tracking
- ✅ Always use `--json` flag for programmatic use
- ✅ Link discovered work with `discovered-from` dependencies
- ✅ Check `bd ready` before asking "what should I work on?"
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT use external issue trackers
- ❌ Do NOT duplicate tracking systems

For more details, see README.md and docs/QUICKSTART.md.

<!-- END BEADS INTEGRATION -->
