# Sample Project Fixture

This fixture simulates a moderately sized Go application used for integration tests.

## Structure Overview

- `cmd/app`: entry point and CLI wiring
- `internal/handlers`: HTTP handlers and business logic
- `internal/models`: domain models and validation helpers
- `internal/utils`: shared utilities with nested packages
- `pkg/storage`: storage abstraction layer
- `configs`, `config`: configuration files in YAML and TOML formats
- `docs`: documentation tree with architecture notes, API descriptions, and ADRs
- `scripts`: helper shell scripts used by automation
- `templates`: HTML/email templates for rendering tasks
- `web/static`: frontend assets spanning JS/CSS/HTML
- `data`: sample structured data files
- `tests`: unit/e2e/integration test artifacts
- `build`, `dist`: directories that should typically be ignored
- `vendor`, `third_party`: dependency placeholders
- `migrations`: database migration SQL files
- `public`: static assets served to clients

The directory contains a cross-section of file types—including Go source, markdown, YAML/TOML, JSON/CSV, shell scripts, and binary placeholders—to exercise scanner rules, language detection, and ignore logic.

Overall the fixture has roughly 70 files spread across multiple nested folders to emulate a realistic repository for end-to-end tests.
