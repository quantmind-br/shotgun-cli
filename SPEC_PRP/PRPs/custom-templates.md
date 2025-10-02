# SPEC PRP: Custom Templates Support

> Ingest the information from this file, implement the Low-Level Tasks, and generate the code that will satisfy the High and Mid-Level Objectives.

## Current State

### Files
- `internal/core/template/manager.go` - Only loads embedded templates from `assets.Templates`
- `internal/core/template/template.go` - Template struct with `IsEmbedded` field (unused)
- `internal/core/template/renderer.go` - Variable substitution logic
- `cmd/template.go` - CLI commands for `list` and `render`
- `cmd/config.go` - Config management with `template.custom-path` (defined but not used)
- `templates/*.md` - 4 embedded templates (analyzeBug, makeDiffGitFormat, makePlan, projectManager)

### Behavior
- Templates are **only** loaded from embedded filesystem at startup
- Manager uses `fs.Sub(assets.Templates, "templates")` to access embedded templates
- Config has `template.custom-path` key defined but never read
- `IsEmbedded` field in Template struct exists but is never set or used
- No filesystem scanning for user-provided templates

### Issues
- Users cannot add custom templates without rebuilding the application
- `template.custom-path` configuration is a dead feature
- No XDG Base Directory compliance for user templates
- Template system is inflexible and requires code changes for new templates

## Desired State

### Files
- `internal/core/template/manager.go` - Load templates from both embedded FS and custom paths
- `internal/core/template/loader.go` - NEW: Template loading abstraction layer
- `internal/core/template/template.go` - `IsEmbedded` field properly utilized
- `cmd/template.go` - Enhanced with `template import/export` commands
- Config properly reads and uses `template.custom-path`
- User templates directory: `~/.config/shotgun-cli/templates/` (XDG compliant)

### Behavior
- Templates loaded from multiple sources with priority: custom > default config > embedded
- Users can place `.md` files in `~/.config/shotgun-cli/templates/`
- `template.custom-path` config allows additional template directories
- `template import <file>` copies template to user directory
- `template export <name> <file>` exports template to file
- `template list` shows origin (embedded/custom) for each template
- Name conflicts: custom templates override embedded ones

### Benefits
- User extensibility without rebuilding
- XDG Base Directory compliance
- Multiple template sources support
- Easy template sharing via import/export
- Clear template provenance in listings

## High-Level Objective

Enable users to create, import, and manage custom prompt templates from filesystem locations without modifying application code, while maintaining backward compatibility with embedded templates.

## Mid-Level Objectives

1. **Template Source Abstraction** - Create a template loading system that supports multiple sources (embedded FS, user config dir, custom paths)
2. **Configuration Integration** - Read and properly utilize `template.custom-path` configuration with XDG Base Directory support
3. **Template Import/Export** - Provide CLI commands to copy templates between filesystem and user config directory
4. **Enhanced Template Listing** - Display template origin (embedded/custom/path) in `template list` output
5. **Backward Compatibility** - Ensure existing embedded templates continue working without changes

## Implementation Notes

### Technical Details
- Use XDG Base Directory spec: `~/.config/shotgun-cli/templates/`
- Template loading priority: custom path > user config dir > embedded
- Name conflicts: later sources override earlier ones
- Use `os.DirFS` for filesystem template loading
- Maintain thread-safety with existing `sync.RWMutex`

### Dependencies
- `github.com/adrg/xdg` - XDG Base Directory support (ADD to go.mod)
- Existing `embed` package for embedded templates
- `os` package for filesystem operations
- `io/fs` for filesystem abstraction

### Coding Standards
- Follow existing Manager pattern with mutex protection
- Maintain zero-breaking changes to public API
- Use existing `parseTemplate` function for all template sources
- Add comprehensive error messages for filesystem issues
- Follow gitignore-style patterns for template discovery

### Testing Strategy
- Unit tests for multi-source loading
- Test priority/override behavior
- Test XDG directory creation
- E2E tests for import/export commands
- Backward compatibility tests

## Context

### Beginning Context

**Files that exist:**
- `internal/core/template/manager.go` (embedded-only loading)
- `internal/core/template/template.go` (IsEmbedded unused)
- `internal/core/template/renderer.go` (variable substitution)
- `cmd/template.go` (list/render commands)
- `cmd/config.go` (template.custom-path defined)
- `templates/prompt_*.md` (4 embedded templates)

**Current behavior:**
- Only embedded templates available
- `template.custom-path` config exists but ignored
- No user extensibility

### Ending Context

**Files that will exist:**
- `internal/core/template/manager.go` (multi-source loading)
- `internal/core/template/loader.go` (NEW - loading abstraction)
- `internal/core/template/template.go` (IsEmbedded properly set)
- `internal/core/template/loader_test.go` (NEW - loader tests)
- `cmd/template.go` (added import/export commands)
- User template directory: `~/.config/shotgun-cli/templates/` (created on first use)
- All existing files remain functional

**New behavior:**
- Templates load from multiple sources
- Users can add `.md` files to config directory
- `template import/export` commands work
- `template list` shows template origin
- Custom templates override embedded ones

## Low-Level Tasks

> Ordered by dependency and implementation logic

### 1. Add XDG Base Directory dependency

**Action:** ADD
**File:** `go.mod`
**Changes:**
- Add `github.com/adrg/xdg v0.6.0` dependency
**Validation:**
```bash
go mod tidy
go list -m github.com/adrg/xdg
# Expected: github.com/adrg/xdg v0.6.0
```

### 2. Create template loader abstraction

**Action:** CREATE
**File:** `internal/core/template/loader.go`
**Changes:**
- Create `TemplateSource` interface with `LoadTemplates() (map[string]*Template, error)` method
- Implement `EmbeddedSource` struct that wraps embedded FS loading
- Implement `FilesystemSource` struct that loads from a directory path
- Add `loadTemplatesFromFS(fsys fs.FS, basePath string) (map[string]*Template, error)` helper
- Set `IsEmbedded` field correctly based on source type
**Validation:**
```bash
go test ./internal/core/template -run TestTemplateLoader
# Expected: PASS
```

### 3. Create loader unit tests

**Action:** CREATE
**File:** `internal/core/template/loader_test.go`
**Changes:**
- Test `EmbeddedSource` loads embedded templates correctly
- Test `FilesystemSource` loads from temp directory
- Test template name extraction (strip "prompt_" prefix, .md suffix)
- Test invalid template handling (malformed content)
- Test empty directory handling
- Test `IsEmbedded` field is set correctly
**Validation:**
```bash
go test ./internal/core/template -v -run TestLoader
# Expected: All loader tests pass
```

### 4. Update Template struct usage

**Action:** MODIFY
**File:** `internal/core/template/template.go`
**Changes:**
- Ensure `IsEmbedded` field is exported and documented
- Add `Source` field to Template struct: `Source string // "embedded", "user", or custom path`
**Validation:**
```bash
go build ./internal/core/template
# Expected: Builds without errors
```

### 5. Refactor Manager to use multi-source loading

**Action:** MODIFY
**File:** `internal/core/template/manager.go`
**Changes:**
- Replace single `loadTemplates()` with `loadFromSources(sources []TemplateSource) error`
- In `NewManager()`: create sources slice with embedded source, user dir source, custom path source (if configured)
- Read `template.custom-path` from viper config
- Use `xdg.ConfigHome` to get `~/.config/shotgun-cli/templates/`
- Create user templates directory if it doesn't exist
- Load templates with priority: custom path > user config > embedded
- Later sources override earlier ones by name
- Set `IsEmbedded` and `Source` fields appropriately
**Validation:**
```bash
go test ./internal/core/template -run TestManager
# Expected: All manager tests pass
```

### 6. Add Manager tests for multi-source loading

**Action:** MODIFY
**File:** `internal/core/template/manager_test.go`
**Changes:**
- Test embedded templates still load
- Test user directory templates load
- Test custom path templates load
- Test priority/override behavior (custom > user > embedded)
- Test missing directories are handled gracefully
- Test `IsEmbedded` and `Source` fields are correct
**Validation:**
```bash
go test ./internal/core/template -v
# Expected: All tests pass including new multi-source tests
```

### 7. Update template list command to show source

**Action:** MODIFY
**File:** `cmd/template.go`
**Changes:**
- In `templateListCmd.RunE`: modify output to show template source
- Format: `name (source)  description`
- Sources: "embedded", "user", or custom path basename
- Update column width calculation to account for source display
**Validation:**
```bash
go build -o build/shotgun-cli ./main.go
./build/shotgun-cli template list
# Expected: Shows templates with source indicators
```

### 8. Create template import command

**Action:** ADD
**File:** `cmd/template.go`
**Changes:**
- Create `templateImportCmd` with usage: `import <file> [name]`
- Read template file from filesystem
- Validate template content with `parseTemplate`
- Extract name from file or use provided name
- Copy to `~/.config/shotgun-cli/templates/`
- Show success message with location
**Validation:**
```bash
echo "## TEMPLATE\nTest\n{TASK}" > /tmp/test.md
./build/shotgun-cli template import /tmp/test.md mytest
./build/shotgun-cli template list | grep mytest
# Expected: mytest appears in list with "user" source
```

### 9. Create template export command

**Action:** ADD
**File:** `cmd/template.go`
**Changes:**
- Create `templateExportCmd` with usage: `export <name> <file>`
- Get template content from manager
- Write to specified file path
- Overwrite confirmation if file exists
- Show success message
**Validation:**
```bash
./build/shotgun-cli template export analyzeBug /tmp/exported.md
test -f /tmp/exported.md
# Expected: File exists with template content
```

### 10. Register new template subcommands

**Action:** MODIFY
**File:** `cmd/template.go`
**Changes:**
- Add `templateCmd.AddCommand(templateImportCmd)` in init()
- Add `templateCmd.AddCommand(templateExportCmd)` in init()
- Update template command long description to mention import/export
**Validation:**
```bash
./build/shotgun-cli template --help
# Expected: Shows import and export subcommands
```

### 11. Update config validation for template.custom-path

**Action:** MODIFY
**File:** `cmd/config.go`
**Changes:**
- In `validateConfigValue`: allow non-existent paths for `template.custom-path` (will be created)
- Add helpful message about path creation on first use
**Validation:**
```bash
./build/shotgun-cli config set template.custom-path /tmp/my-templates
# Expected: Success message, no error about non-existent path
```

### 12. Create E2E test for custom templates

**Action:** CREATE
**File:** `test/e2e/custom_templates_test.sh`
**Changes:**
- Test import command with valid template
- Test import command with invalid template
- Test export command
- Test list command shows sources
- Test custom path in config
- Test template override behavior
- Test render works with custom templates
**Validation:**
```bash
make test-e2e
# Expected: New e2e test passes
```

### 13. Update documentation

**Action:** MODIFY
**File:** `CLAUDE.md`
**Changes:**
- Document custom templates feature in "Architecture" section
- Note template loading priority
- Document XDG compliance for user templates
- Add examples of import/export commands
**Validation:**
```bash
grep -i "custom template" CLAUDE.md
# Expected: Documentation present
```

### 14. Integration test - full workflow

**Action:** MODIFY
**File:** `internal/core/template/manager_test.go`
**Changes:**
- Add integration test `TestCustomTemplateWorkflow`:
  - Create temp config directory
  - Place custom template file
  - Initialize manager
  - Verify custom template loads
  - Verify custom template overrides embedded if same name
  - Verify template can be rendered
**Validation:**
```bash
go test ./internal/core/template -v -run TestCustomTemplateWorkflow
# Expected: PASS
```

### 15. Build and final validation

**Action:** MODIFY
**File:** `Makefile`
**Changes:**
- Ensure `make build` includes new dependencies
- Ensure `make test` runs all new tests
**Validation:**
```bash
make clean
make deps
make build
make test
./build/shotgun-cli template list
mkdir -p ~/.config/shotgun-cli/templates
echo -e "## Test Custom Template\n{TASK}\n\nCustom template content" > ~/.config/shotgun-cli/templates/custom_test.md
./build/shotgun-cli template list | grep custom_test
./build/shotgun-cli template render custom_test --var TASK="test task"
# Expected: All commands succeed, custom template appears and renders
```

## Risk Assessment

### Identified Risks

1. **Filesystem Access Failures**
   - Risk: User config directory not writable
   - Mitigation: Graceful degradation to embedded-only mode with warning log
   - Go/No-Go: Warn user but continue with embedded templates

2. **Template Name Conflicts**
   - Risk: Custom template same name as embedded
   - Mitigation: Custom templates explicitly override embedded (documented behavior)
   - Go/No-Go: Feature, not bug - document clearly

3. **Invalid User Templates**
   - Risk: Malformed templates break loading
   - Mitigation: Skip invalid templates with error log, continue loading others
   - Go/No-Go: Individual template failures don't stop application

4. **XDG Dependency Issues**
   - Risk: XDG library has platform issues
   - Mitigation: Fallback to hardcoded `~/.config` if XDG fails
   - Go/No-Go: Test on Linux/macOS/Windows before merge

### Rollback Strategy

- All changes are additive to existing API
- Feature flag possible: skip custom loading if `template.custom-path` not set and user dir empty
- Revert commits in reverse task order if critical issue found
- Embedded templates continue working regardless of custom template issues

## Progressive Enhancement

1. **Phase 1** (Tasks 1-6): Core multi-source loading - embedded templates still work
2. **Phase 2** (Tasks 7-10): CLI enhancements - import/export commands
3. **Phase 3** (Tasks 11-15): Testing and documentation

Each phase is independently functional. Can merge after each phase if needed.

## Validation Commands Summary

```bash
# After implementation, run this sequence to validate everything works:

# 1. Build
make clean && make deps && make build

# 2. Run tests
make test

# 3. Test embedded templates still work
./build/shotgun-cli template list

# 4. Test custom template directory
mkdir -p ~/.config/shotgun-cli/templates
echo -e "## My Custom Template\n{TASK}\n\nDo the thing: {TASK}" > ~/.config/shotgun-cli/templates/mycustom.md
./build/shotgun-cli template list | grep mycustom

# 5. Test import command
echo -e "## Imported Template\n{TASK}" > /tmp/import-test.md
./build/shotgun-cli template import /tmp/import-test.md imported
./build/shotgun-cli template list | grep imported

# 6. Test export command
./build/shotgun-cli template export analyzeBug /tmp/exported.md
test -f /tmp/exported.md && echo "Export OK"

# 7. Test render with custom template
./build/shotgun-cli template render mycustom --var TASK="test task"

# 8. Test config integration
./build/shotgun-cli config set template.custom-path /tmp/custom-templates
mkdir -p /tmp/custom-templates
echo -e "## Path Template\n{TASK}" > /tmp/custom-templates/pathtest.md
./build/shotgun-cli template list | grep pathtest

# 9. Run E2E tests
make test-e2e

# All commands should succeed without errors
```

## Quality Checklist

- [x] Current state fully documented (embedded-only system)
- [x] Desired state clearly defined (multi-source with XDG support)
- [x] All objectives measurable (validation commands per task)
- [x] Tasks ordered by dependency (loader -> manager -> CLI -> tests)
- [x] Each task has validation that AI can run
- [x] Risks identified with mitigations (filesystem, conflicts, invalid templates)
- [x] Rollback strategy included (additive changes, phase-based)
- [x] Integration points noted (config system, CLI commands, template system)
- [x] Progressive enhancement phases defined
- [x] Backward compatibility ensured (embedded templates continue working)
