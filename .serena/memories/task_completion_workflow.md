# Task Completion Workflow

## Standard Development Task Completion

### 1. Code Quality Checks
```bash
# Format code according to Go standards
npm run format           # Runs: go fmt ./...

# Check for potential issues
npm run lint             # Runs: go vet ./...
```

### 2. Testing
```bash
# Run all tests
npm test                 # Runs: go test ./...

# Run tests with coverage (when needed)
go test -v -cover ./...

# Run specific tests (for targeted testing)
go test -v ./internal/core -run TestSpecificFunction
```

### 3. Functionality Verification
```bash
# Test the application in development mode
npm run dev              # Runs: go run .

# Test specific functionality through the TUI interface
# - Navigate through file exclusion
# - Test template selection
# - Verify prompt generation
```

### 4. Build Verification (when applicable)
```bash
# For local changes
npm run build:local      # Quick local build test

# For cross-platform changes
npm run build:all        # Ensure all platforms build successfully
```

### 5. Template Changes (if applicable)
When modifying templates in `templates/`:
- Verify template syntax with placeholders: `{TASK}`, `{RULES}`, `{FILE_STRUCTURE}`, `{CURRENT_DATE}`
- Test template rendering through the application
- Ensure output format matches expected structure (git diff for Dev, markdown for others)

## Pre-Commit Checklist
- [ ] Code formatted with `npm run format`
- [ ] No linting errors from `npm run lint`
- [ ] All tests pass with `npm test`
- [ ] Application runs correctly with `npm run dev`
- [ ] New functionality tested through UI
- [ ] Documentation updated if needed (README.md, CLAUDE.md)

## Error Resolution
If any step fails:
1. **Format/Lint errors**: Fix code style issues reported
2. **Test failures**: Address failing tests, add missing test cases
3. **Build failures**: Check Go syntax, imports, and cross-platform compatibility
4. **Runtime errors**: Test with `npm run dev` and debug through UI interaction

## Special Considerations
- **Windows Development**: Be aware of path separators and UTF-8 encoding issues
- **Cross-Platform**: Test builds on Windows target when making file handling changes
- **UI Changes**: Test keyboard shortcuts and TUI responsiveness
- **Template Changes**: Verify placeholder substitution and output formatting