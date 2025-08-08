# Windows Development Notes

## Windows-Specific Considerations

### System Commands
Since the development system is Windows, use these commands instead of Unix equivalents:

```cmd
# File operations
dir                    # Instead of ls
type filename.txt      # Instead of cat
findstr "pattern" *.go # Instead of grep
where go               # Instead of which go
```

### Path Handling in Code
- **Always use `filepath.Join()`** for cross-platform compatibility
- **Be aware of backslash vs forward slash** in path strings
- **Use `filepath.Separator`** when needed for platform-specific separators

### UTF-8 Support
The application includes special Windows UTF-8 handling in `main.go`:
- Sets console code pages to UTF-8 (65001)
- Configures environment variables for UTF-8
- Enables virtual terminal processing
- This is critical for proper display of Unicode characters in the TUI

### Windows-Specific Features in Codebase

#### Console Configuration
- `enableUTF8Windows()` function in main.go
- Windows-specific BubbleTea options (WithoutSignalHandler)
- Console code page manipulation via syscalls

#### File System
- XDG directory support via `github.com/adrg/xdg` library
- Configuration stored in `%LOCALAPPDATA%\shotgun-cli\`
- Custom templates in `%LOCALAPPDATA%\shotgun-cli\templates\`

#### Keyring Integration
- Uses Windows Credential Manager via `github.com/99designs/keyring`
- Secure storage for API keys and sensitive configuration

### Development Environment Setup

#### Prerequisites
- **Go 1.24.5+**: Required for building
- **Node.js 14.0.0+**: Required for npm build scripts
- **Git**: For version control

#### Environment Variables
```cmd
# Enable debug logging
set DEBUG=1
go run .

# UTF-8 support (automatically handled by application)
set LANG=en_US.UTF-8
set LC_ALL=en_US.UTF-8
```

### Windows Build Process
The build system is designed to work well on Windows:
- npm scripts handle cross-platform builds
- GoReleaser configuration supports Windows builds
- Binary outputs include `.exe` extension automatically
- Windows-specific optimizations in build flags

### Testing Considerations
- File path tests must work with Windows path separators
- UTF-8 and Unicode handling tests are important
- Console input/output tests for Windows terminal behavior
- File permission tests (Windows has different permission model)

### Common Windows Issues
1. **Path length limits**: Windows has historically had 260-character path limits
2. **File locking**: Windows locks files more aggressively than Unix systems
3. **Case sensitivity**: Windows file system is case-insensitive by default
4. **Line endings**: CRLF vs LF (handled automatically by Git and Go)

### Performance Notes
- Windows I/O can be slower than Unix systems for many small files
- Directory traversal performance differs
- The application includes progress tracking for large directory scans