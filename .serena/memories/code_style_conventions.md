# Code Style and Conventions

## Go Conventions
- **Package Structure**: Use `internal/` for private packages not meant for external import
- **Error Handling**: 
  - Use structured errors with `ShotgunError` struct containing operation and path context
  - `ErrorCollector` for aggregating multiple errors during batch operations
  - Always validate directory access before operations
- **Naming**: Standard Go naming conventions (PascalCase for exported, camelCase for unexported)
- **Concurrency**: 
  - Use channels for progress updates between goroutines and UI
  - Worker pools (`workerPool chan struct{}`) to limit concurrent file operations
  - Mutex protection for shared state (selection maps, progress counters)

## Project Structure Conventions
- **Core Business Logic**: `internal/core/` contains all business logic components
- **UI/Presentation Logic**: `internal/ui/` contains BubbleTea UI components
- **External Assets**: Templates stored in `templates/` directory
- **Build Artifacts**: Generated binaries in `bin/` directory

## Architecture Patterns
- **State Management**: 
  - `ViewState` enum for UI navigation
  - `SelectionState` for thread-safe file inclusion/exclusion
  - Progress updates via channels for real-time feedback
- **Template System**: Simple string replacement using placeholders like `{TASK}`, `{RULES}`, `{FILE_STRUCTURE}`, `{CURRENT_DATE}`
- **Component Design**: Inverse file selection (exclude rather than include files)

## File Organization
- **Types**: Core data structures in `types.go`
- **Scanning**: Directory traversal logic in `scanner.go`
- **Generation**: Context generation in `generator.go`
- **Templates**: Template processing in `template.go` and `template_simple.go`
- **UI Components**: Separate files for different UI concerns (`app.go`, `filetree.go`, `views.go`, `components.go`)