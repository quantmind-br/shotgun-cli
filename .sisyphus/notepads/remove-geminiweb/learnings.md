# Learnings from Removing GeminiWeb

## Overview
We removed `ProviderGeminiWeb` (the browser-based integration) while keeping `ProviderGemini` (the API integration).

## Key Changes
- Removed `ProviderGeminiWeb` constant from `internal/core/llm`.
- Cleaned up `config.go` and tests in `internal/core/llm`.
- Removed registration in `internal/app/providers.go` and `cmd/providers.go`.
- Removed CLI commands and logic in `cmd/llm.go` and `cmd/context.go`.
- Removed "legacy" `sendToGemini` logic in `cmd/context.go` which relied on a missing `internal/platform/geminiweb` package.

## Issues Encountered
- `go test` reported a missing file `cmd/gemini.go` which was actually a phantom artifact in the build cache. `go clean -cache` fixed it.
- `internal/platform/geminiweb` directory was missing, but `cmd/context.go` still imported it, causing build failures. Removed the import and usage.

## Architecture Notes
- The project distinguishes between `gemini` (API) and `geminiweb` (browser automation).
- `geminiweb` was integrated via a separate binary and "legacy" direct calls in `cmd/context.go`, bypassing the `llm.Provider` interface in some places, but also implemented it in `internal/platform/geminiweb`.
- Cleaning up the provider required touching multiple layers (`core`, `app`, `cmd`, `ui`).
