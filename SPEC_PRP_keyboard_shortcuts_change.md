# PRP Specification: Keyboard Shortcuts Change - Fix Configuration Access Conflict

**PRP ID:** SPEC_PRP_KEYBOARD_SHORTCUTS_001  
**Date:** 2025-07-20  
**Author:** Claude Code Assistant  
**Status:** Draft  
**Priority:** Medium  

## Executive Summary

This PRP addresses a critical usability issue where the keyboard shortcut "c" is being used for both configuration access and continuing in the file exclusion view, creating user confusion and workflow disruption. The solution involves changing configuration access from "c" to "o" (for "options") and restoring "c" for its original continue functionality in the file exclusion view.

## Problem Statement

### Current Issue
The current implementation has a keyboard shortcut conflict:
- **Global "c" shortcut** (lines 251-256 in `internal/ui/app.go`): Used to access configuration from any main view
- **File exclusion "c" shortcut** (lines 72-76 in `internal/ui/views.go`): Used to continue to the next step from file exclusion view

This creates confusion because:
1. Users expect "c" to mean "continue" in file exclusion (as shown in help text)
2. The global "c" for configuration can override the continue functionality
3. Help documentation shows conflicting usage of "c"

### Impact
- **User Experience**: Confusing and inconsistent keyboard navigation
- **Workflow Disruption**: Users may accidentally access configuration when trying to continue
- **Documentation Inconsistency**: Help text and README show conflicting key mappings

## Current State Assessment

### Keyboard Shortcut Analysis

#### Global Shortcuts (app.go:251-256)
```go
case "c":
    // Access configuration (only from main views, not during generation)
    if m.currentView != ViewGeneration && m.currentView != ViewConfiguration {
        m.currentView = ViewConfiguration
        return m, nil
    }
```

#### File Exclusion Shortcuts (views.go:72-76)
```go
case "c":
    // Continue to next step
    m.includedFiles = core.GetIncludedFiles(m.fileTree_root, m.selection)
    m.currentView = ViewTemplateSelection
    return m, nil
```

#### Help Text References
1. **app.go:451**: `c            Continue to next step`
2. **filetree.go:136**: `Space: toggle | hjkl: navigate | Enter: expand | r: reset | a/A: exclude/include all | c: continue`

#### Documentation References
1. **README.md:137**: `c: Continue to next step`
2. **CLAUDE.md**: Contains keyboard shortcut documentation

### Configuration System Architecture
- Configuration access is implemented in `ViewConfiguration` state
- Configuration form is handled by `ConfigFormModel` in `internal/ui/config_form.go`
- Configuration views are in `internal/ui/config_views.go`
- Configuration backend is in `internal/core/config.go`

## Desired State

### New Keyboard Mapping
- **"o" shortcut**: Access configuration (Options)
- **"c" shortcut**: Continue functionality in file exclusion view only
- **Consistent behavior**: "c" always means continue where applicable

### User Experience Goals
1. Intuitive keyboard shortcuts that match user expectations
2. Consistent "c" for continue across all applicable views
3. Memorable "o" for configuration (options/settings)
4. Clear separation between navigation and configuration access

## Implementation Plan

### Phase 1: Core Shortcut Changes
**Duration:** 1 day  
**Risk Level:** Low  

#### Task 1.1: Update Global Configuration Shortcut
**File:** `internal/ui/app.go`  
**Lines:** 251-256  
**Changes:**
```go
// OLD:
case "c":
    // Access configuration (only from main views, not during generation)

// NEW:
case "o":
    // Access configuration (options) (only from main views, not during generation)
```

**Validation Criteria:**
- [ ] "o" key opens configuration from any main view
- [ ] "c" key no longer opens configuration globally
- [ ] Configuration access still blocked during generation

#### Task 1.2: Verify File Exclusion Continue Shortcut
**File:** `internal/ui/views.go`  
**Lines:** 72-76  
**Changes:** No changes needed - "c" for continue is already implemented correctly
**Validation Criteria:**
- [ ] "c" key continues from file exclusion to template selection
- [ ] Continue functionality works as expected

### Phase 2: Help Text Updates
**Duration:** 0.5 days  
**Risk Level:** Low  

#### Task 2.1: Update In-App Help Text
**File:** `internal/ui/app.go`  
**Lines:** 435-477 (renderHelp function)  
**Changes:**
```markdown
# OLD help text:
  General:
    Ctrl+Q, Ctrl+C    Quit application
    ?            Toggle this help
    Esc          Go back to previous step

# NEW help text:
  General:
    Ctrl+Q, Ctrl+C    Quit application
    ?            Toggle this help
    o            Access configuration/settings
    Esc          Go back to previous step
```

**Validation Criteria:**
- [ ] Help text shows "o" for configuration access
- [ ] Help text maintains "c" for continue in file exclusion section
- [ ] No references to conflicting "c" usage

#### Task 2.2: Update File Tree Help Text
**File:** `internal/ui/filetree.go`  
**Line:** 136  
**Changes:** No changes needed - "c: continue" is correct
**Validation Criteria:**
- [ ] File tree help text still shows "c: continue"
- [ ] Help text is consistent with actual behavior

### Phase 3: Documentation Updates
**Duration:** 0.5 days  
**Risk Level:** Low  

#### Task 3.1: Update README.md
**File:** `README.md`  
**Lines:** 125-154 (Keyboard Shortcuts section)  
**Changes:**
```markdown
# OLD:
### General
- `Ctrl+Q`, `Ctrl+C`: Quit application
- `?`: Toggle help
- `Esc`: Go back to previous step

# NEW:
### General
- `Ctrl+Q`, `Ctrl+C`: Quit application
- `?`: Toggle help
- `o`: Access configuration/settings
- `Esc`: Go back to previous step
```

**Validation Criteria:**
- [ ] README shows "o" for configuration
- [ ] README maintains "c" for continue in file exclusion
- [ ] All keyboard shortcuts are documented correctly

#### Task 3.2: Update CLAUDE.md
**File:** `CLAUDE.md`  
**Section:** Input System keyboard shortcuts  
**Changes:** Update keyboard shortcut documentation to reflect new mapping
**Validation Criteria:**
- [ ] CLAUDE.md reflects new "o" for configuration
- [ ] Development documentation is consistent

### Phase 4: Testing and Validation
**Duration:** 0.5 days  
**Risk Level:** Low  

#### Task 4.1: Manual Testing
**Test Cases:**
1. **Configuration Access Test**:
   - Press "o" from file exclusion view → should open configuration
   - Press "o" from template selection → should open configuration
   - Press "o" from task description → should open configuration
   - Press "o" from custom rules → should open configuration
   - Press "o" during generation → should be ignored

2. **Continue Functionality Test**:
   - Press "c" from file exclusion view → should continue to template selection
   - Press "c" from other views → should not open configuration

3. **Help Text Verification**:
   - Press "?" → verify help shows "o" for configuration
   - Check file tree help → verify shows "c" for continue

#### Task 4.2: User Experience Testing
**Scenarios:**
1. New user workflow through entire application
2. Experienced user muscle memory adaptation
3. Edge case testing (rapid key presses, etc.)

**Validation Criteria:**
- [ ] Configuration access is intuitive with "o"
- [ ] Continue functionality works consistently with "c"
- [ ] No keyboard conflicts or confusion
- [ ] Help text matches actual behavior

## Risk Assessment

### Technical Risks
| Risk | Probability | Impact | Mitigation |
|------|------------|---------|------------|
| Keyboard conflict remains | Low | Medium | Thorough testing of all views |
| User confusion during transition | Medium | Low | Clear documentation updates |
| Muscle memory disruption | Medium | Low | Intuitive "o" for options |

### Rollback Strategy
If issues arise after implementation:
1. **Immediate Rollback**: Revert `app.go` changes to restore "c" for configuration
2. **Partial Rollback**: Keep "o" for configuration, restore global "c" override
3. **Documentation Rollback**: Revert help text and documentation changes

### Dependencies
- No external dependencies
- No breaking API changes
- No database schema changes
- Compatible with existing configuration system

## Testing Strategy

### Unit Tests
- No unit tests required (UI keyboard handling)

### Integration Tests
- Manual testing of keyboard navigation flows
- End-to-end workflow testing

### User Acceptance Criteria
- [ ] "o" reliably opens configuration from all main views
- [ ] "c" reliably continues from file exclusion view
- [ ] No keyboard shortcut conflicts exist
- [ ] Help documentation matches behavior
- [ ] User workflows are intuitive and consistent

## Success Metrics

### Technical Metrics
- Zero keyboard shortcut conflicts
- 100% consistency between help text and behavior
- All views respond correctly to new shortcuts

### User Experience Metrics
- Intuitive navigation (subjective assessment)
- Reduced confusion about keyboard shortcuts
- Consistent "continue" behavior across applicable views

## Migration Considerations

### User Communication
- Update any user guides or documentation
- Consider adding temporary help text about shortcut change
- No breaking changes to saved configurations

### Backward Compatibility
- No API changes required
- Configuration system remains unchanged
- Existing user configurations preserved

## Implementation Timeline

| Task | Duration | Dependencies |
|------|----------|--------------|
| Update global configuration shortcut | 2 hours | None |
| Verify file exclusion continue shortcut | 1 hour | Task 1 |
| Update in-app help text | 2 hours | Task 1-2 |
| Update file tree help text verification | 30 minutes | Task 3 |
| Update README.md | 1 hour | Task 3 |
| Update CLAUDE.md | 1 hour | Task 4 |
| Manual testing | 3 hours | All tasks |
| User experience testing | 2 hours | Task 7 |

**Total Estimated Time:** 2.5 days

## Conclusion

This PRP resolves a significant usability issue by:
1. Eliminating keyboard shortcut conflicts
2. Providing intuitive and consistent navigation
3. Maintaining expected "continue" behavior
4. Introducing memorable "options" access with "o"

The implementation is low-risk with clear rollback options and comprehensive testing. The change improves user experience without affecting the underlying functionality or configuration system.

## Appendix

### Code Locations Summary
- **Global shortcuts**: `internal/ui/app.go:241-271`
- **File exclusion shortcuts**: `internal/ui/views.go:70-83`
- **Help rendering**: `internal/ui/app.go:435-477`
- **File tree help**: `internal/ui/filetree.go:136`
- **Configuration system**: `internal/ui/config_form.go`, `internal/ui/config_views.go`
- **Documentation**: `README.md`, `CLAUDE.md`

### Alternative Considerations
- **Alternative keys considered**: "s" (settings), "p" (preferences), "," (common settings key)
- **Chosen "o"**: Intuitive for "options", available, not conflicting with existing shortcuts