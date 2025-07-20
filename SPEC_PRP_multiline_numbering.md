# PRP Specification: Enhanced Multiline Input with Auto-Numbering

## Current State Assessment

### Current Implementation
```yaml
current_state:
  files: 
    - internal/ui/app.go          # Main UI model with textinput/textarea
    - internal/ui/views.go        # View rendering and input handling
    - internal/core/types.go      # TemplateData structure
    - internal/core/template*.go  # Template processing logic
  
  behavior:
    task_input: "Single-line textinput.Model with 500 char limit"
    rules_input: "Multiline textarea.Model (60x5) but no auto-numbering"
    template_processing: "Simple {TASK} and {RULES} placeholder replacement"
    user_workflow: "Tab to switch focus, Enter to generate prompt"
  
  issues:
    - Task Description limited to single line (500 chars)
    - No support for multiple tasks with automatic numbering
    - Custom Rules lack structured numbering system
    - Templates expect simple text substitution without formatting
    - User cannot create organized task lists or rule sets
```

### Pain Points Identified
1. **Single-line Task Limitation**: Users cannot describe complex multi-step tasks
2. **No Task Numbering**: Multiple tasks appear as unstructured text
3. **Unstructured Rules**: Custom rules lack organization and clarity
4. **Poor UX for Complex Prompts**: Large projects need structured input
5. **Template Compatibility**: Current templates expect unformatted text

## Desired State Specification

### Target Architecture
```yaml
desired_state:
  files:
    - internal/ui/app.go          # Enhanced with new multiline components
    - internal/ui/views.go        # Updated input handling and rendering
    - internal/ui/components.go   # NEW: Custom numbered textarea components
    - internal/core/types.go      # Enhanced TemplateData with formatting
    - internal/core/formatter.go # NEW: Auto-numbering and formatting logic
  
  behavior:
    task_input: "Multiline textarea with auto-numbering on Enter key"
    rules_input: "Enhanced multiline textarea with auto-numbering"
    auto_numbering: "Automatic sequential numbering (1., 2., 3., etc.)"
    template_processing: "Smart formatting with numbered lists"
    user_workflow: "Enhanced UX with visual numbering feedback"
  
  benefits:
    - Support for complex multi-task descriptions
    - Organized, numbered task and rule lists
    - Better prompt structure and readability
    - Maintained template compatibility
    - Enhanced user productivity
```

## Hierarchical Objectives

### 1. High-Level Goal
Transform shotgun-cli input system to support structured, numbered multiline input for both Task Description and Custom Rules fields, enabling users to create organized, multi-item prompts.

### 2. Mid-Level Milestones

#### Milestone A: Enhanced Input Components
- Create auto-numbering textarea components
- Implement smart Enter key handling for numbering
- Add visual feedback for numbered items

#### Milestone B: Core Logic Enhancement  
- Develop formatting logic for numbered lists
- Enhance TemplateData processing
- Maintain backward compatibility with existing templates

#### Milestone C: UI Integration
- Replace existing input components
- Update view rendering logic
- Enhance keyboard navigation and focus management

#### Milestone D: Testing & Validation
- Comprehensive testing of new input behavior
- Template compatibility verification
- User experience validation

### 3. Low-Level Implementation Tasks

## Implementation Tasks with Dense Keywords

### Task 1: Create Auto-Numbering Components
```yaml
task_1_create_numbered_textarea:
  action: CREATE
  file: internal/ui/components.go
  changes: |
    - CREATE new file with NumberedTextArea struct
    - IMPLEMENT auto-numbering logic on Enter key press
    - ADD visual indicators for numbered items
    - INCLUDE methods: NewNumberedTextArea(), Update(), View(), GetNumberedValue()
    - HANDLE edge cases: empty lines, backspace on numbers, cursor positioning
  validation:
    - command: "go build ."
    - expect: "compilation success"
    - command: "go test ./internal/ui"
    - expect: "all tests pass"
```

### Task 2: Enhanced Formatter Logic
```yaml
task_2_create_formatter:
  action: CREATE  
  file: internal/core/formatter.go
  changes: |
    - CREATE FormatNumberedText() function
    - IMPLEMENT ParseNumberedLines() for existing numbered text
    - ADD ReformatWithNumbers() for auto-numbering
    - INCLUDE validation for numbered format consistency
    - HANDLE mixed numbered/unnumbered content gracefully
  validation:
    - command: "go test ./internal/core -run TestFormatter"
    - expect: "formatter tests pass"
```

### Task 3: Enhance TemplateData Structure
```yaml
task_3_enhance_template_data:
  action: MODIFY
  file: internal/core/types.go
  changes: |
    - ADD FormattedTask field to TemplateData struct
    - ADD FormattedRules field to TemplateData struct  
    - MODIFY existing Task and Rules fields (maintain compatibility)
    - ADD formatting options struct for future extensibility
  validation:
    - command: "go build ."
    - expect: "compilation success"
    - command: "grep -n 'FormattedTask\|FormattedRules' internal/core/types.go"
    - expect: "new fields present"
```

### Task 4: Replace Task Input Component
```yaml
task_4_replace_task_input:
  action: MODIFY
  file: internal/ui/app.go
  changes: |
    - REPLACE textinput.Model taskInput with NumberedTextArea
    - MODIFY NewModel() to initialize new numbered component
    - UPDATE component dimensions and styling
    - REMOVE textinput-specific imports and dependencies
    - ADD new component import and initialization
  validation:
    - command: "go build ."
    - expect: "compilation success"
    - command: "npm run dev"
    - expect: "application starts successfully"
```

### Task 5: Enhance Rules Input Component  
```yaml
task_5_enhance_rules_input:
  action: MODIFY
  file: internal/ui/app.go
  changes: |
    - REPLACE textarea.Model rulesInput with NumberedTextArea
    - MODIFY existing rulesInput initialization
    - UPDATE placeholder text to indicate numbering capability
    - MAINTAIN existing dimensions as baseline
  validation:
    - command: "go build ."
    - expect: "compilation success"
```

### Task 6: Update View Rendering Logic
```yaml
task_6_update_view_rendering:
  action: MODIFY
  file: internal/ui/views.go
  changes: |
    - MODIFY updatePromptComposition() for new component handling
    - UPDATE renderPromptComposition() to show numbered content
    - MODIFY tab switching logic for new components
    - UPDATE value extraction to use GetNumberedValue()
    - ADD visual feedback for numbered vs unnumbered content
  validation:
    - command: "go build ."
    - expect: "compilation success"
    - command: "npm run dev"
    - expect: "UI renders correctly with new components"
```

### Task 7: Enhanced Template Processing
```yaml
task_7_enhance_template_processing:
  action: MODIFY
  file: internal/core/template_simple.go
  changes: |
    - MODIFY GeneratePrompt() to use formatted fields when available
    - ADD fallback to original fields for backward compatibility
    - IMPLEMENT smart formatting detection
    - ENSURE templates receive properly formatted numbered content
  validation:
    - command: "go test ./internal/core -run TestTemplate"
    - expect: "template tests pass"
    - command: "echo '1. Test task\n2. Another task' | npm run dev"
    - expect: "numbered content appears in generated prompt"
```

### Task 8: Update Generation Logic
```yaml
task_8_update_generation_logic:
  action: MODIFY
  file: internal/ui/views.go
  changes: |
    - MODIFY generatePrompt() to use formatter for TemplateData
    - ADD FormattedTask and FormattedRules population
    - IMPLEMENT proper numbering before template processing
    - MAINTAIN existing taskText and rulesText for compatibility
  validation:
    - command: "npm run dev"
    - expect: "prompt generation works with numbered content"
    - command: "cat shotgun_prompt_*.md | grep -E '^[0-9]+\.'"
    - expect: "numbered items present in output"
```

### Task 9: Comprehensive Testing
```yaml
task_9_comprehensive_testing:
  action: CREATE
  file: internal/ui/components_test.go
  changes: |
    - CREATE unit tests for NumberedTextArea component
    - ADD tests for auto-numbering behavior
    - INCLUDE edge case testing (empty lines, backspace, etc.)
    - ADD integration tests with existing UI workflow
  validation:
    - command: "go test ./internal/ui -v"
    - expect: "all new component tests pass"
```

### Task 10: Documentation Update
```yaml
task_10_update_documentation:
  action: MODIFY
  file: CLAUDE.md
  changes: |
    - ADD documentation for new numbered input system
    - UPDATE architecture overview with new components
    - MODIFY keyboard shortcuts section for new features
    - ADD examples of numbered task/rule input
  validation:
    - command: "grep -n 'NumberedTextArea\|auto-numbering' CLAUDE.md"
    - expect: "documentation includes new features"
```

## Implementation Strategy

### Dependency Order
```
1. Task 1 (Components) → Foundation for all other changes
2. Task 2 (Formatter) → Required by template processing
3. Task 3 (TemplateData) → Required by generation logic
4. Task 4,5 (Replace Inputs) → Can be done in parallel
5. Task 6 (View Logic) → Depends on new components
6. Task 7,8 (Template/Generation) → Depends on formatter and data structure
7. Task 9 (Testing) → Validation of all components
8. Task 10 (Documentation) → Final step
```

### Progressive Enhancement Strategy
1. **Phase 1**: Create new components without replacing existing (parallel development)
2. **Phase 2**: Replace components one at a time with fallback options
3. **Phase 3**: Enhanced template processing with backward compatibility
4. **Phase 4**: Full integration and testing

### Rollback Plan
- Keep original textinput/textarea components available
- Feature flag system for enabling/disabling new components
- Comprehensive backup of working implementation
- Quick rollback via git branch strategy

## Risk Assessment and Mitigations

### Identified Risks

#### High Risk: Template Compatibility
**Risk**: New numbered format breaks existing templates
**Mitigation**: 
- Dual-field approach (original + formatted)
- Backward compatibility testing
- Fallback to original behavior

#### Medium Risk: User Experience Disruption  
**Risk**: Changed input behavior confuses existing users
**Mitigation**:
- Gradual introduction with clear visual feedback
- Help text updates
- Intuitive Enter key behavior

#### Low Risk: Performance Impact
**Risk**: Auto-numbering adds processing overhead
**Mitigation**:
- Efficient string processing algorithms
- Debounced numbering updates
- Performance testing

### Go/No-Go Criteria
- ✅ All existing functionality preserved
- ✅ Template output maintains compatibility  
- ✅ New numbering works reliably
- ✅ Performance impact < 100ms
- ✅ User testing shows positive feedback

## Quality Checklist

- [x] Current state fully documented
- [x] Desired state clearly defined  
- [x] All objectives measurable
- [x] Tasks ordered by dependency
- [x] Each task has validation that AI can run
- [x] Risks identified with mitigations
- [x] Rollback strategy included
- [x] Integration points noted
- [x] Progressive enhancement strategy
- [x] Backward compatibility maintained

## Success Metrics

1. **Functional**: Users can create numbered task lists via Enter key
2. **Compatibility**: All existing templates work without modification
3. **Performance**: Input responsiveness maintained < 100ms
4. **Usability**: Tab navigation and focus management work seamlessly
5. **Output Quality**: Generated prompts contain properly formatted numbered content

## User Acceptance Criteria

1. **As a user**, I can press Enter in Task Description to create numbered tasks
2. **As a user**, I can press Enter in Custom Rules to create numbered rules  
3. **As a user**, numbered content appears properly formatted in generated prompts
4. **As a user**, existing single-line task input still works for simple cases
5. **As a user**, Tab navigation between fields works as expected
6. **As a user**, all existing templates produce the same quality output

---

This specification provides a comprehensive roadmap for transforming the shotgun-cli input system while maintaining backward compatibility and ensuring robust, user-friendly numbered input capabilities.