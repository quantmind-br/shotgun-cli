# Windows Development Environment Notes

## Windows-Specific Commands and Utilities

### File Operations
```cmd
# Directory listing
dir                      # List directory contents (equivalent to ls)
dir /s                   # Recursive listing (equivalent to ls -R)

# File viewing
type filename.txt        # View file contents (equivalent to cat)
more filename.txt        # Paginated file viewing (equivalent to less)

# File searching
find "search_text" filename.txt     # Search in specific file
findstr "pattern" *.go              # Search pattern in multiple files (equivalent to grep)

# File system
where shotgun-cli.exe    # Find executable location (equivalent to which)
copy source dest         # Copy files (equivalent to cp)
move source dest         # Move files (equivalent to mv)
del filename             # Delete files (equivalent to rm)
```

### Path Handling
- Use backslashes (`\`) for Windows paths in commands
- Go code handles path separators automatically with `filepath` package
- Be aware of case-insensitive file system behavior

### Development Environment
```cmd
# Go commands work the same on Windows
go version
go build
go test
go mod tidy

# npm commands work the same
npm run dev
npm test
npm run build
```

## Windows-Specific Code Considerations

### UTF-8 Support
The application includes Windows-specific UTF-8 support in `main.go`:
```go
// Ensure UTF-8 support on Windows
if runtime.GOOS == "windows" {
    os.Setenv("LANG", "en_US.UTF-8")
    os.Setenv("LC_ALL", "en_US.UTF-8")
    // Additional UTF-8 setup via Windows API
}
```

### File Path Handling
- Use `filepath.Join()` for cross-platform path construction
- Use `filepath.Separator` for platform-specific separators
- The codebase properly handles Windows paths in scanner and generator components

### Terminal Compatibility
- BubbleTea framework handles Windows terminal differences
- Windows Command Prompt and PowerShell both supported
- UTF-8 and Unicode character support enabled

## Build Artifacts
Windows builds generate:
- `bin/shotgun-cli.exe` - Windows executable
- Cross-platform builds create multiple binaries for different OS/arch combinations

## Common Windows Development Issues
1. **Path separators**: Always use `filepath` package functions
2. **Line endings**: Git handles CRLF/LF conversion automatically
3. **Case sensitivity**: Windows is case-insensitive, but other platforms are not
4. **Executable extensions**: Windows requires `.exe` extension for executables