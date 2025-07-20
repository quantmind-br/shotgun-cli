# PRP: OpenAI Translation Integration for shotgun-cli

**Date:** 2025-07-20  
**Version:** 1.0  
**Priority:** High  
**Complexity:** Medium-High  

## Executive Summary

Integrate OpenAI-compatible APIs to provide automatic translation of user input fields (tasks and rules) from native language to English, ensuring final prompts are consistently in English while preserving user's ability to input in their preferred language.

## Current State Analysis

### Current Implementation
```yaml
current_state:
  files:
    - internal/ui/components.go    # NumberedTextArea input component
    - internal/ui/views.go         # Task/Rules view handlers and data capture
    - internal/ui/app.go           # State management (taskText, rulesText)
    - internal/core/template_simple.go # Template processing with {TASK}/{RULES}
    - templates/*.md               # Four template files using placeholders

  behavior:
    - User inputs task description in ViewTaskDescription (Step 3/4)
    - User inputs custom rules in ViewCustomRules (Step 4/4)
    - F5 triggers data capture: m.taskText = m.taskInput.Value()
    - Template processing replaces {TASK} and {RULES} with raw user input
    - No validation, translation, or language processing

  data_flow: |
    Input Capture (views.go:214-215) →
    Template Data Creation (views.go:410-415) →
    String Replacement (template_simple.go:69-75) →
    Final Prompt Generation

  issues:
    - No internationalization support
    - Users must write in English for optimal LLM performance
    - No configuration system exists
    - No API integration capability
    - Security concerns for API key storage not addressed
```

### Architecture Strengths
- Clean separation of concerns (UI → State → Processing)
- Simple data structures (plain strings)
- Unicode/UTF-8 support already implemented
- Thread-safe with mutex protection
- Multiple clear integration points available

## Desired State Specification

### Target Implementation
```yaml
desired_state:
  files:
    - internal/core/config.go           # Configuration management system
    - internal/core/keyring.go          # Secure API key storage
    - internal/core/translator.go       # OpenAI translation client
    - internal/core/types.go            # Enhanced with Config structs
    - internal/ui/config_form.go        # Configuration UI component
    - internal/ui/config_views.go       # Configuration view rendering
    - internal/ui/views.go              # Enhanced with translation calls
    - internal/ui/app.go                # Enhanced with config integration

  behavior:
    - User configures API settings via dedicated configuration screen
    - Automatic translation occurs after task input (F5 press)
    - Automatic translation occurs after rules input (F5 press)
    - Translation can be toggled on/off globally
    - Secure API key storage using OS keyring
    - Support for OpenAI-compatible APIs (custom base URLs)
    - Graceful fallback when translation fails

  enhanced_data_flow: |
    Input Capture →
    Translation Check (if enabled) →
    API Translation Call (task/rules separately) →
    Template Data Creation (with translated text) →
    String Replacement →
    Final Prompt Generation

  benefits:
    - Native language input support
    - Consistent English output for optimal LLM performance
    - Configurable translation settings
    - Secure credential management
    - Support for multiple OpenAI-compatible providers
    - Non-breaking changes to existing workflow
```

## Hierarchical Objectives

### 1. High-Level Goal
**Objective:** Enable seamless native-language input with automatic English translation  
**Success Criteria:** Users can write tasks/rules in any language and get English prompts
**Impact:** Significantly improved user experience for non-English speakers

### 2. Mid-Level Milestones

#### 2.1 Configuration Infrastructure
**Objective:** Implement robust configuration system with secure key management  
**Components:** Config files, keyring integration, UI forms  
**Validation:** Configuration persists across sessions and keys are secure

#### 2.2 Translation Service Integration  
**Objective:** Integrate OpenAI-compatible APIs for translation  
**Components:** HTTP client, error handling, retry logic  
**Validation:** Successful translation with fallback mechanisms

#### 2.3 UI Integration & User Experience
**Objective:** Seamless integration with existing BubbleTea interface  
**Components:** Configuration views, status indicators, error messaging  
**Validation:** Intuitive workflow without disrupting existing UX

### 3. Low-Level Implementation Tasks

## Implementation Tasks

### Phase 1: Configuration Foundation
#### Task 1.1: Core Configuration Structure
```yaml
task_name: implement_config_types
action: ADD
file: internal/core/types.go
changes: |
  - Add Config struct with OpenAI, Translation, and App settings
  - Add OpenAIConfig with API key alias, base URL, model, timeouts
  - Add TranslationConfig with enabled flag and target language
  - Add AppConfig for application preferences
  - Include JSON tags for serialization
validation:
  - command: "go build ."
  - expect: "successful compilation"
```

#### Task 1.2: Configuration Manager Implementation
```yaml
task_name: implement_config_manager
action: CREATE
file: internal/core/config.go
changes: |
  - Implement ConfigManager struct with CRUD operations
  - Add XDG-compliant directory resolution (github.com/adrg/xdg)
  - Implement JSON file read/write with error handling
  - Add configuration validation and defaults
  - Thread-safe operations with mutex protection
validation:
  - command: "go test ./internal/core -run TestConfigManager"
  - expect: "all tests pass"
```

#### Task 1.3: Secure Key Management
```yaml
task_name: implement_keyring_manager
action: CREATE
file: internal/core/keyring.go
changes: |
  - Implement SecureKeyManager using github.com/99designs/keyring
  - Support OS keyring (macOS Keychain, Windows Credential Manager, Linux Secret Service)
  - Encrypted file fallback for unsupported platforms
  - API key storage/retrieval with alias system
validation:
  - command: "go test ./internal/core -run TestSecureKeyManager"
  - expect: "key storage and retrieval successful"
```

#### Task 1.4: Update Dependencies
```yaml
task_name: add_dependencies
action: MODIFY
file: go.mod
changes: |
  - Add github.com/adrg/xdg for cross-platform directories
  - Add github.com/99designs/keyring for secure key storage
  - Add go-openai or official openai-go for API client
validation:
  - command: "go mod tidy && go mod verify"
  - expect: "dependencies resolved successfully"
```

### Phase 2: Translation Service
#### Task 2.1: OpenAI Client Implementation
```yaml
task_name: implement_translator
action: CREATE
file: internal/core/translator.go
changes: |
  - Implement Translator struct with OpenAI client
  - Add TranslateText method with context and timeout support
  - Implement retry logic with exponential backoff
  - Add translation prompt optimization for task/rules context
  - Error handling and fallback mechanisms
validation:
  - command: "go test ./internal/core -run TestTranslator"
  - expect: "translation successful with mock API"
```

#### Task 2.2: Translation Integration Point
```yaml
task_name: integrate_translation_workflow
action: MODIFY
file: internal/ui/views.go
changes: |
  - Add translation calls after task input capture (line ~214)
  - Add translation calls after rules input capture (line ~215)
  - Implement loading indicators during translation
  - Add error handling with user notification
  - Preserve original text and show translated version
validation:
  - command: "npm run dev"
  - expect: "translation workflow functional in UI"
```

### Phase 3: Configuration UI
#### Task 3.1: Configuration Form Component
```yaml
task_name: implement_config_form
action: CREATE
file: internal/ui/config_form.go
changes: |
  - Implement ConfigFormModel with form fields
  - Add field validation and input masking for passwords
  - Support different field types (text, password, select, toggle)
  - Keyboard navigation and editing modes
validation:
  - command: "go build ./internal/ui"
  - expect: "UI components compile successfully"
```

#### Task 3.2: Configuration Views
```yaml
task_name: implement_config_views
action: CREATE
file: internal/ui/config_views.go
changes: |
  - Add configuration view rendering with Lipgloss styling
  - Implement tabbed interface for different config sections
  - Add help text and field descriptions
  - Status indicators for connection testing
validation:
  - command: "npm run dev"
  - expect: "configuration UI accessible and functional"
```

#### Task 3.3: View State Integration
```yaml
task_name: integrate_config_view_state
action: MODIFY
file: internal/ui/app.go
changes: |
  - Add ViewConfiguration to ViewState enum
  - Add configManager and keyManager to Model struct
  - Add keyboard shortcut ('c') for configuration access
  - Implement navigation to/from configuration screen
validation:
  - command: "npm run dev"
  - expect: "configuration accessible from any view"
```

### Phase 4: User Experience Enhancement
#### Task 4.1: Status Indicators
```yaml
task_name: add_translation_status_indicators
action: MODIFY
file: internal/ui/views.go
changes: |
  - Add translation status in task/rules views
  - Show loading spinner during API calls
  - Display success/error states with appropriate colors
  - Add translated text preview capability
validation:
  - command: "npm run dev"
  - expect: "clear visual feedback during translation"
```

#### Task 4.2: Error Handling Enhancement
```yaml
task_name: enhance_error_handling
action: MODIFY
file: internal/ui/views.go
changes: |
  - Add comprehensive error messages for API failures
  - Implement fallback behavior when translation unavailable
  - Add retry mechanisms with user confirmation
  - Graceful degradation to original text
validation:
  - command: "npm run dev"
  - expect: "robust error handling with clear user guidance"
```

### Phase 5: Testing & Documentation
#### Task 5.1: Unit Tests
```yaml
task_name: implement_comprehensive_tests
action: CREATE
file: internal/core/translator_test.go
changes: |
  - Add unit tests for all translation functionality
  - Mock OpenAI API responses for testing
  - Test error conditions and retry logic
  - Validate configuration management
validation:
  - command: "go test -v ./... -cover"
  - expect: "80%+ test coverage on new code"
```

#### Task 5.2: Integration Tests
```yaml
task_name: add_integration_tests
action: CREATE
file: internal/ui/translation_integration_test.go
changes: |
  - Test complete translation workflow
  - Validate UI state transitions
  - Test configuration persistence
  - End-to-end translation scenarios
validation:
  - command: "go test -v ./internal/ui -tags=integration"
  - expect: "full workflow tests pass"
```

#### Task 5.3: Documentation Updates
```yaml
task_name: update_documentation
action: MODIFY
file: README.md
changes: |
  - Add translation feature documentation
  - Document configuration setup process
  - Add API provider compatibility notes
  - Include troubleshooting section
validation:
  - command: "manual review"
  - expect: "clear documentation for users"
```

## Implementation Strategy

### Dependencies & Order
1. **Foundation First:** Configuration system must be implemented before translation
2. **Security Priority:** Keyring integration before UI to ensure secure design
3. **Incremental UI:** Add configuration views before integrating translation workflow
4. **Progressive Enhancement:** Translation as optional feature, existing workflow unchanged

### Risk Mitigation
| Risk | Impact | Mitigation |
|------|--------|------------|
| API Key Security | High | Use OS keyring with encrypted fallback |
| API Failures | Medium | Retry logic, fallback to original text |
| Performance Impact | Medium | Async translation, loading indicators |
| UI Complexity | Low | Separate configuration view, optional feature |

### Rollback Strategy
- Feature flag for translation (enabled/disabled)
- Configuration validation prevents invalid states
- Original workflow preserved if translation disabled
- Clean separation allows component removal if needed

## Success Metrics

### Functional Requirements
- [ ] User can configure OpenAI API settings securely
- [ ] Tasks and rules are automatically translated to English
- [ ] Translation can be enabled/disabled globally
- [ ] Original workflow preserved when translation disabled
- [ ] Support for OpenAI-compatible APIs (custom base URLs)

### Non-Functional Requirements
- [ ] Translation completes within 30 seconds for typical inputs
- [ ] API keys stored securely using OS keyring
- [ ] UI remains responsive during translation
- [ ] Configuration persists across application restarts
- [ ] Error messages provide clear guidance to users

### Quality Gates
- [ ] Unit test coverage ≥80% for new components
- [ ] Integration tests cover complete workflow
- [ ] Manual testing on Windows, macOS, and Linux
- [ ] Performance testing with various input sizes
- [ ] Security audit of key storage implementation

## Timeline Estimation

**Total Estimated Duration:** 12-15 working days

- **Phase 1 (Configuration):** 3-4 days
- **Phase 2 (Translation):** 3-4 days  
- **Phase 3 (UI Integration):** 3-4 days
- **Phase 4 (UX Enhancement):** 2-3 days
- **Phase 5 (Testing & Docs):** 1-2 days

## Conclusion

This PRP provides a comprehensive roadmap for integrating OpenAI translation capabilities into shotgun-cli while maintaining the existing user experience and ensuring security best practices. The implementation is designed to be non-breaking, optional, and aligned with the application's current architecture patterns.

The hierarchical task breakdown ensures systematic implementation with clear validation criteria at each step, enabling confident progress tracking and quality assurance throughout the development process.