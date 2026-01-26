# Documentation Standards

## GoDoc Style
- Start with the symbol name: `// ContextService defines ...`
- Be concise but complete.
- For interfaces, explain the contract.
- For structs, explain the purpose and key fields.
- For methods, explain parameters and return values if non-obvious.

## Inline Comments
- Explain WHY, not WHAT.
- Use full sentences.
- Group logic blocks with section headers if useful.

## Wizard State Machine Documentation
- Documented the `Update` function in `internal/ui/wizard.go`.
- The wizard uses a message-driven architecture where state transitions are triggered by specific message types (ScanCompleteMsg, GenerationCompleteMsg, LLMCompleteMsg).
- Asynchronous operations (Scanning, Generation) use a pattern of StartMsg -> ProgressMsg(s) -> CompleteMsg.
- Critical transitions (like finishing a scan or generation) are now clearly marked with comments to aid future maintenance.

## Tree Component Documentation
- Added documentation for `recomputeSelectionStates` (post-order traversal) and `buildVisibleItems` (flattening/virtualization).
- Explicitly documenting the "flattening" strategy for TUI virtualization is critical for understanding how the component handles large file trees without performance degradation.
- Documenting the state propagation logic helps future developers understand why directory checkboxes behave as tri-state controls.

- **Parallel Edits**: When using the `edit` tool in parallel on the same file, subsequent edits will fail with "File has been modified" if the first one succeeds. It is safer to apply edits to the same file sequentially or batch them if possible.
- **GoDoc Verification**: Always verify documentation updates with `go doc -all <package>` to ensuring comments are correctly associated with symbols.

## Platform & Application Layer Documentation
- Reviewed and verified documentation for `internal/platform` (http, anthropic, geminiapi).
- Reviewed and verified documentation for `internal/app` (context, service, config).
- Added missing GoDoc for `internal/platform/http` client methods.
- Most exported symbols in these layers were already well-documented, confirming the project's adherence to documentation standards.
- Verified that `golangci-lint` passes for all documentation changes.
