## Shotgun CLI Agent Guidelines

### Build & Test

- **Build:** `npm run build`
- **Test:** `npm run test`
- **Run a single test:** `go test -run ^TestMyFunction$`
- **Lint:** `golangci-lint run` (assumed, not specified)

### Code Style

- **Formatting:** Use `gofmt` or `goimports`.
- **Imports:** Group standard library, third-party, and internal packages.
- **Types:** Use descriptive type names.
- **Naming:** Use camelCase for variables and functions. Use PascalCase for exported functions and types.
- **Error Handling:** Use `fmt.Errorf` with `%w` to wrap errors.
- **Logging:** Use the standard `log` package for logging.
- **Comments:** Add comments to explain complex logic.
- **Dependencies:** Manage Go dependencies with Go Modules.
- **UI:** The project uses `tview` for the terminal UI.
