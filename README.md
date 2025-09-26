# shotgun-cli

A cross-platform CLI tool that generates LLM-optimized codebase contexts with both TUI wizard and headless CLI modes.

## Installation

```bash
go install github.com/your-org/shotgun-cli@latest
```

## Usage

### Interactive Mode
When called without arguments, shotgun-cli launches an interactive 5-step wizard:

```bash
shotgun-cli
```

### Configuration

The tool can be configured using various options, including custom ignore patterns.

#### Ignore Patterns

The `--ignore` flag and `ScanConfig.IgnorePatterns` now use **gitignore syntax** via the `github.com/sabhiram/go-gitignore` library.

**Important**: As of the latest version, ignore patterns follow gitignore rules instead of simple `filepath.Match` patterns.

##### Examples of gitignore-style patterns:

- `*.log` - Ignore all .log files
- `dir/` - Ignore directories named "dir"
- `!/keep.go` - Explicitly include files that would otherwise be ignored
- `**/vendor/` - Ignore vendor directories at any depth
- `build/` - Ignore build directory
- `*.tmp` - Ignore all temporary files
- `!important.tmp` - But keep this specific temporary file

##### Migration from filepath.Match patterns:

If you were previously using `filepath.Match`-style patterns, you may need to update them:

| Old Pattern (filepath.Match) | New Pattern (gitignore) | Notes |
|------------------------------|------------------------|--------|
| `*.log` | `*.log` | ✅ Same |
| `test*` | `test*` | ✅ Same |
| `[abc].txt` | `[abc].txt` | ✅ Same |
| Custom complex patterns | Check gitignore docs | May need adjustment |

##### Advanced gitignore features:

- **Directory matching**: Patterns ending with `/` only match directories
- **Negation**: Patterns starting with `!` negate (include) previously ignored files
- **Nested patterns**: Use `**/` for matching at any directory depth
- **Relative paths**: Patterns starting with `/` are anchored to the repository root

For complete gitignore pattern documentation, see the [go-gitignore library documentation](https://github.com/sabhiram/go-gitignore).

## Development

### Building

```bash
go build -o shotgun-cli
```

### Testing

```bash
go test ./...
```