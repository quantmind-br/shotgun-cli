# PRP: Configuration Shortcuts Implementation

## Overview

This Product Requirements Phase (PRP) specification defines the implementation of functional configuration shortcuts (F1, F2, F3, F4) in the OpenAI API configuration screen. Currently, these shortcuts exist as stub implementations and need to be connected to the actual business logic for saving configuration, resetting to defaults, and testing API connections.

## 1. Current State Assessment

### 1.1 Current Stub Implementations

**Location:** `internal/ui/config_form.go` (lines 499-519)

```go
// Current stub implementations - NON-FUNCTIONAL
func (m *ConfigFormModel) saveConfiguration() tea.Cmd {
	return func() tea.Msg {
		// Build configuration from form values
		// This would update the actual config and save it
		return configSavedMsg{}
	}
}

func (m *ConfigFormModel) resetConfiguration() tea.Cmd {
	return func() tea.Msg {
		// Reset to default configuration
		return configResetMsg{}
	}
}

func (m *ConfigFormModel) testConnection() tea.Cmd {
	return func() tea.Msg {
		// Test API connection with current settings
		return connectionTestMsg{success: true, message: "Connection successful"}
	}
}
```

### 1.2 Current Shortcut Bindings

**Location:** `internal/ui/config_form.go` (lines 357-370)

- **F1 (Help):** ✅ **WORKING** - Toggles help display (`m.showHelp = !m.showHelp`)
- **F2 (Save Config):** ❌ **STUB** - Returns empty `configSavedMsg{}`
- **F3 (Reset):** ❌ **STUB** - Returns empty `configResetMsg{}`
- **F4 (Test Connection):** ❌ **STUB** - Returns hardcoded success message

### 1.3 Configuration Infrastructure Analysis

**ConfigManager** (`internal/core/config.go`):
- ✅ Thread-safe configuration management with `sync.RWMutex`
- ✅ JSON-based configuration persistence at `~/.config/shotgun-cli/config.json`
- ✅ Validation and default value handling
- ✅ Methods: `Load()`, `Save()`, `Update()`, `Reset()`, `Get()`

**SecureKeyManager** (`internal/core/keyring.go`):
- ✅ Cross-platform secure key storage (Keychain/Credential Manager/Secret Service)
- ✅ Methods: `StoreAPIKey()`, `GetAPIKey()`, `HasAPIKey()`, `TestAPIKey()`
- ✅ Test functionality with `TestAPIKey()` method for validation

**Translator** (`internal/core/translator.go`):
- ✅ OpenAI API client with connection testing
- ✅ Method: `TestConnection()` for verifying API connectivity
- ✅ Configuration validation with `ValidateConfig()`

### 1.4 Configuration Data Flow

```
ConfigFormModel (UI) → ConfigManager (Core) → File System
     ↓                        ↓
SecureKeyManager ← → System Keyring
     ↓
Translator → OpenAI API
```

### 1.5 Integration Points

**Main App Integration** (`internal/ui/app.go`):
- ✅ Configuration form initialized with `ConfigManager` and `SecureKeyManager`
- ✅ Message handlers exist but perform no actions:
  - `configSavedMsg`: Updates translator config only
  - `configResetMsg`: No action
  - `connectionTestMsg`: No action

## 2. Desired State & Implementation Requirements

### 2.1 F2 (Save Configuration) Implementation

**Functionality:**
1. **Validate Form Data:** Check all required fields are filled and valid
2. **Update Core Configuration:** Apply form values to `ConfigManager`
3. **Store API Key:** Save API key securely via `SecureKeyManager`
4. **Persist Changes:** Save configuration file to disk
5. **Update Runtime State:** Refresh translator and other dependent components
6. **User Feedback:** Display save status (success/failure with details)

**Error Handling:**
- Field validation errors (required fields, format validation)
- API key storage failures (keyring access issues)
- File system errors (permission issues, disk full)
- Configuration validation failures

### 2.2 F3 (Reset Configuration) Implementation

**Functionality:**
1. **Load Default Values:** Reset all form fields to default configuration
2. **Preserve Existing API Key:** Option to keep or clear stored API key
3. **Reset Core Configuration:** Update `ConfigManager` with defaults
4. **UI State Reset:** Clear validation errors and editing state
5. **User Confirmation:** Optional confirmation dialog for destructive action
6. **User Feedback:** Display reset confirmation

### 2.3 F4 (Test Connection) Implementation

**Functionality:**
1. **Validate Current Form:** Check required fields for API connection
2. **Create Test Configuration:** Build temporary configuration from form
3. **Test API Key Storage:** Verify keyring access and API key
4. **Test API Connection:** Make test API call to verify connectivity
5. **Connection Diagnostics:** Detailed error reporting for failures
6. **User Feedback:** Display test results with specific error messages

### 2.4 Enhanced User Feedback System

**Status Message Display:**
- Success notifications (green, with checkmark)
- Error messages (red, with error details)
- Warning messages (yellow, for non-critical issues)
- Progress indicators for async operations

## 3. Technical Implementation Plan

### 3.1 Task Breakdown with Dependencies

#### Task 1: Form Data Extraction and Validation
**Scope:** Create methods to convert UI form data to configuration structures
**Files:** `internal/ui/config_form.go`
**Dependencies:** None
**Effort:** 2-3 hours

**Implementation Details:**
```go
// Extract form data into configuration structs
func (m *ConfigFormModel) extractConfigurationData() (*core.Config, error)
func (m *ConfigFormModel) extractAPIKeyData() (string, error)
func (m *ConfigFormModel) validateFormData() map[string]string
```

#### Task 2: F2 Save Configuration Implementation
**Scope:** Implement actual save functionality
**Files:** `internal/ui/config_form.go`
**Dependencies:** Task 1
**Effort:** 3-4 hours

**Implementation Details:**
```go
func (m *ConfigFormModel) saveConfiguration() tea.Cmd {
    return func() tea.Msg {
        // 1. Validate form data
        // 2. Extract configuration and API key
        // 3. Update ConfigManager
        // 4. Store API key in keyring
        // 5. Save configuration file
        // 6. Return success/failure message
    }
}
```

#### Task 3: F3 Reset Configuration Implementation
**Scope:** Implement configuration reset functionality
**Files:** `internal/ui/config_form.go`
**Dependencies:** None
**Effort:** 2-3 hours

**Implementation Details:**
```go
func (m *ConfigFormModel) resetConfiguration() tea.Cmd {
    return func() tea.Msg {
        // 1. Load default configuration
        // 2. Update form fields with defaults
        // 3. Clear validation errors
        // 4. Optionally preserve API key
        // 5. Return reset confirmation
    }
}
```

#### Task 4: F4 Test Connection Implementation
**Scope:** Implement API connection testing
**Files:** `internal/ui/config_form.go`
**Dependencies:** Task 1
**Effort:** 3-4 hours

**Implementation Details:**
```go
func (m *ConfigFormModel) testConnection() tea.Cmd {
    return func() tea.Msg {
        // 1. Validate required fields for connection
        // 2. Create temporary translator configuration
        // 3. Test keyring access and API key
        // 4. Test API connectivity
        // 5. Return detailed test results
    }
}
```

#### Task 5: Enhanced Message Handling
**Scope:** Improve app-level message handling for configuration operations
**Files:** `internal/ui/app.go`
**Dependencies:** Tasks 2-4
**Effort:** 2 hours

**Implementation Details:**
```go
// Enhanced message types with detailed information
type configSavedMsg struct {
    success bool
    message string
    errors  map[string]string
}

type configResetMsg struct {
    success bool
}

type connectionTestMsg struct {
    success bool
    message string
    details string
}
```

#### Task 6: UI Feedback Enhancement
**Scope:** Add status message display and improved user feedback
**Files:** `internal/ui/config_views.go`
**Dependencies:** Task 5
**Effort:** 2-3 hours

**Implementation Details:**
- Status message rendering
- Error message display improvements
- Progress indicators for async operations

### 3.2 Data Structures and Interfaces

#### Enhanced Message Types
```go
type ConfigOperationResult struct {
    Success bool              `json:"success"`
    Message string            `json:"message"`
    Errors  map[string]string `json:"errors,omitempty"`
    Details string            `json:"details,omitempty"`
}

type configSavedMsg struct {
    ConfigOperationResult
}

type configResetMsg struct {
    ConfigOperationResult
}

type connectionTestMsg struct {
    ConfigOperationResult
    LatencyMs int `json:"latency_ms,omitempty"`
}
```

#### Form Validation Interface
```go
type FieldValidator interface {
    ValidateField(field ConfigField) error
}

type FormValidator struct {
    validators map[string]FieldValidator
}
```

### 3.3 Configuration Field Mapping

**OpenAI Configuration Section:**
```go
OpenAI: core.OpenAIConfig{
    APIKeyAlias: extractFieldValue("API Key"),
    BaseURL:     extractFieldValue("Base URL"),
    Model:       extractFieldValue("Model"),
    Timeout:     parseInt(extractFieldValue("Timeout (seconds)")),
    MaxTokens:   parseInt(extractFieldValue("Max Tokens")),
    Temperature: parseFloat(extractFieldValue("Temperature")),
}
```

**Translation Configuration Section:**
```go
Translation: core.TranslationConfig{
    Enabled:        parseBool(extractFieldValue("Enable Translation")),
    TargetLanguage: extractFieldValue("Target Language"),
    ContextPrompt:  extractFieldValue("Custom Translation Prompt"),
}
```

**Application Configuration Section:**
```go
App: core.AppConfig{
    Theme:           extractFieldValue("Theme"),
    AutoSave:        parseBool(extractFieldValue("Auto Save Config")),
    ShowLineNumbers: parseBool(extractFieldValue("Show Line Numbers")),
    DefaultTemplate: extractFieldValue("Default Template"),
}
```

## 4. Validation Criteria

### 4.1 F2 (Save Configuration) Success Criteria

**Functional Requirements:**
- [ ] All form fields are validated before saving
- [ ] Configuration is persisted to `~/.config/shotgun-cli/config.json`
- [ ] API key is stored securely in system keyring
- [ ] Invalid configurations are rejected with specific error messages
- [ ] Translator is reinitialized with new configuration
- [ ] Success/failure feedback is displayed to user

**Test Cases:**
1. **Valid Configuration Save:**
   - Fill all required fields with valid values
   - Press F2
   - Verify configuration file is updated
   - Verify API key is stored in keyring
   - Verify success message is displayed

2. **Invalid Configuration Rejection:**
   - Leave required fields empty
   - Press F2
   - Verify validation errors are shown
   - Verify configuration is not saved

3. **API Key Storage Failure:**
   - Simulate keyring access failure
   - Press F2
   - Verify appropriate error message is displayed

### 4.2 F3 (Reset Configuration) Success Criteria

**Functional Requirements:**
- [ ] All form fields are reset to default values
- [ ] Default configuration matches `core.DefaultConfig()`
- [ ] Validation errors are cleared
- [ ] User receives confirmation of reset
- [ ] API key handling preference is respected (keep/clear)

**Test Cases:**
1. **Complete Reset:**
   - Modify configuration values
   - Press F3
   - Verify all fields show default values
   - Verify confirmation message is displayed

2. **API Key Preservation:**
   - Set API key
   - Press F3
   - Verify API key field still shows "set" status

### 4.3 F4 (Test Connection) Success Criteria

**Functional Requirements:**
- [ ] Current form values are used for testing (not saved values)
- [ ] API key from form is validated
- [ ] Network connectivity to OpenAI API is tested
- [ ] Detailed error messages for different failure types
- [ ] Success message with connection details

**Test Cases:**
1. **Successful Connection:**
   - Enter valid API key and settings
   - Press F4
   - Verify success message with connection details

2. **Invalid API Key:**
   - Enter invalid API key
   - Press F4
   - Verify API key error message

3. **Network Connectivity Issues:**
   - Configure invalid base URL
   - Press F4
   - Verify network error message

4. **Missing Required Fields:**
   - Leave API key empty
   - Press F4
   - Verify validation error message

## 5. Risk Assessment and Mitigation

### 5.1 High Risk Areas

#### Risk 1: API Key Security
**Description:** Potential exposure of API keys during storage or testing
**Impact:** High - API key compromise
**Mitigation:**
- Use secure keyring APIs exclusively
- Never log API keys
- Validate keyring availability before operations
- Clear temporary API key variables immediately after use

#### Risk 2: Configuration File Corruption
**Description:** Invalid JSON or file permission issues during save
**Impact:** Medium - Configuration loss
**Mitigation:**
- Atomic file writes (write to temp file, then rename)
- Configuration validation before saving
- Backup current configuration before overwriting
- Graceful fallback to default configuration

#### Risk 3: UI State Inconsistency
**Description:** Form state and actual configuration becoming out of sync
**Impact:** Medium - User confusion, data loss
**Mitigation:**
- Reload form state after successful save operations
- Clear error states appropriately
- Validate state transitions
- Comprehensive error handling

### 5.2 Medium Risk Areas

#### Risk 4: Network Timeouts During Testing
**Description:** Connection tests hanging or failing due to network issues
**Impact:** Low - Poor user experience
**Mitigation:**
- Implement reasonable timeouts (30 seconds)
- Show progress indicators
- Allow cancellation of test operations
- Provide detailed error messages

#### Risk 5: Cross-Platform Keyring Compatibility
**Description:** Keyring operations failing on specific platforms
**Impact:** Medium - Feature unavailable on some systems
**Mitigation:**
- Test on all supported platforms
- Provide clear error messages for keyring failures
- Fallback to file-based storage when keyring unavailable
- Document keyring requirements

## 6. Rollback Strategy

### 6.1 Implementation Rollback Plan

**Phase 1 - Individual Feature Rollback:**
- Each shortcut (F2, F3, F4) can be individually disabled
- Revert to stub implementations if issues arise
- Configuration form remains functional for manual editing

**Phase 2 - Complete Feature Rollback:**
- Disable configuration shortcuts entirely
- Preserve existing configuration display functionality
- Users can still view configuration but not modify via shortcuts

**Phase 3 - Emergency Rollback:**
- Revert to previous stable configuration system
- Disable configuration form entirely if necessary
- Provide command-line configuration fallback

### 6.2 Data Recovery Procedures

**Configuration Backup:**
- Automatic backup of configuration before any save operation
- Backup location: `~/.config/shotgun-cli/config.backup.json`
- Manual recovery instructions in documentation

**API Key Recovery:**
- Keyring operations are non-destructive by default
- API keys remain in keyring even if configuration operations fail
- Manual keyring inspection tools for recovery

## 7. Integration Testing Approach

### 7.1 Unit Testing

**Configuration Operations:**
```go
func TestConfigFormModel_SaveConfiguration(t *testing.T)
func TestConfigFormModel_ResetConfiguration(t *testing.T)
func TestConfigFormModel_TestConnection(t *testing.T)
func TestConfigFormModel_ExtractConfigurationData(t *testing.T)
func TestConfigFormModel_ValidateFormData(t *testing.T)
```

**Mock Dependencies:**
- Mock `ConfigManager` for testing save/load operations
- Mock `SecureKeyManager` for testing keyring operations
- Mock `Translator` for testing connection operations

### 7.2 Integration Testing

**End-to-End Scenarios:**
1. **Complete Configuration Workflow:**
   - Start with default configuration
   - Modify settings via form
   - Save with F2
   - Verify persistence
   - Reset with F3
   - Verify default restoration

2. **Error Handling Workflows:**
   - Invalid API key handling
   - Network connectivity issues
   - File permission problems
   - Keyring access failures

3. **Cross-Platform Testing:**
   - Windows (Credential Manager)
   - macOS (Keychain)
   - Linux (Secret Service)

### 7.3 Manual Testing Checklist

**F2 (Save Configuration) Testing:**
- [ ] Valid configuration saves successfully
- [ ] Invalid configurations are rejected
- [ ] API key is stored in keyring
- [ ] Configuration file is updated
- [ ] UI shows appropriate feedback
- [ ] Translator is reinitialized

**F3 (Reset Configuration) Testing:**
- [ ] All fields reset to defaults
- [ ] Validation errors are cleared
- [ ] API key handling works correctly
- [ ] UI shows reset confirmation

**F4 (Test Connection) Testing:**
- [ ] Valid configuration tests successfully
- [ ] Invalid API key shows error
- [ ] Network issues are detected
- [ ] Progress is shown during testing
- [ ] Detailed error messages are provided

## 8. Implementation Timeline

### 8.1 Development Phases

**Phase 1 (Days 1-2): Foundation**
- Task 1: Form data extraction and validation
- Unit tests for extraction methods
- Basic error handling infrastructure

**Phase 2 (Days 3-4): Core Implementation**
- Task 2: F2 Save Configuration
- Task 3: F3 Reset Configuration  
- Integration with ConfigManager and SecureKeyManager

**Phase 3 (Days 5-6): Connection Testing**
- Task 4: F4 Test Connection
- Integration with Translator
- Network error handling

**Phase 4 (Days 7-8): UI Enhancement**
- Task 5: Enhanced message handling
- Task 6: UI feedback improvements
- Status message display

**Phase 5 (Days 9-10): Testing & Polish**
- Comprehensive testing
- Bug fixes
- Documentation updates
- Performance optimization

### 8.2 Delivery Milestones

**Milestone 1:** F2 Save functionality working (Day 4)
**Milestone 2:** F3 Reset functionality working (Day 4)
**Milestone 3:** F4 Test functionality working (Day 6)
**Milestone 4:** Enhanced UI feedback complete (Day 8)
**Milestone 5:** Full testing and validation complete (Day 10)

## 9. Acceptance Criteria

### 9.1 Functional Acceptance

**Primary Criteria:**
- [ ] F2 saves configuration to file system and keyring
- [ ] F3 resets all fields to default values
- [ ] F4 tests API connection with current form values
- [ ] All operations provide clear user feedback
- [ ] Error conditions are handled gracefully
- [ ] No data loss under normal operation

**Secondary Criteria:**
- [ ] Performance is acceptable (operations complete within 3 seconds)
- [ ] UI remains responsive during operations
- [ ] Cross-platform compatibility maintained
- [ ] Existing functionality is not disrupted
- [ ] Help text (F1) accurately describes new functionality

### 9.2 Quality Acceptance

**Code Quality:**
- [ ] Code follows existing project patterns and style
- [ ] Comprehensive error handling
- [ ] Thread-safe operations where required
- [ ] Memory leaks are avoided
- [ ] No security vulnerabilities introduced

**Testing Coverage:**
- [ ] Unit tests for all new functions
- [ ] Integration tests for complete workflows
- [ ] Manual testing on all supported platforms
- [ ] Error condition testing
- [ ] Performance testing under load

## 10. Dependencies and Prerequisites

### 10.1 External Dependencies

**Go Modules:**
- No new dependencies required
- Existing dependencies sufficient:
  - `github.com/99designs/keyring` (keyring operations)
  - `github.com/adrg/xdg` (configuration directory)
  - `github.com/sashabaranov/go-openai` (API testing)

### 10.2 System Requirements

**Keyring Support:**
- Windows: Credential Manager available
- macOS: Keychain access enabled
- Linux: Secret Service running (GNOME Keyring/KDE Wallet)

**File System Access:**
- Read/write access to user configuration directory
- Temporary file creation for atomic operations

### 10.3 Development Prerequisites

**Development Environment:**
- Go 1.21+ development environment
- Test keyring access on target platforms
- OpenAI API key for testing (development/testing key)
- Network connectivity for API testing

## Conclusion

This PRP specification provides a comprehensive roadmap for implementing the configuration shortcuts functionality in shotgun-cli. The implementation focuses on robust error handling, secure API key management, and excellent user experience while maintaining the existing architecture and code quality standards.

The phased approach allows for incremental delivery and testing, reducing implementation risk while ensuring that each feature is thoroughly validated before moving to the next phase.