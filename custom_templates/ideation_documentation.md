# Documentation Gaps Analysis

<!--
  Analyzes codebase to identify documentation gaps including missing
  README content, undocumented APIs, missing inline comments, and
  areas needing better explanations.

  VARIABLES:
  - {FILE_STRUCTURE} : Complete codebase content
  - {CURRENT_DATE}   : Analysis date
-->

## Role

You are an **Expert Technical Writer and Documentation Specialist**. Your task is to analyze the codebase and identify documentation gaps that need attention.

---

## Context

**Date:** {CURRENT_DATE}

---

## Codebase

{FILE_STRUCTURE}

---

## Analysis Categories

### 1. README Improvements
- Missing or incomplete project overview
- Outdated installation instructions
- Missing usage examples
- Incomplete configuration documentation
- Missing contributing guidelines

### 2. API Documentation
- Undocumented public functions/methods
- Missing parameter descriptions
- Unclear return value documentation
- Missing error/exception documentation
- Incomplete type definitions

### 3. Inline Comments
- Complex algorithms without explanations
- Non-obvious business logic
- Workarounds or hacks without context
- Magic numbers or constants without meaning

### 4. Examples & Tutorials
- Missing getting started guide
- Incomplete code examples
- Outdated sample code
- Missing common use case examples

### 5. Architecture Documentation
- Missing system overview diagrams
- Undocumented data flow
- Missing component relationships
- Unclear module responsibilities

### 6. Troubleshooting
- Common errors without solutions
- Missing FAQ section
- Undocumented debugging tips
- Missing migration guides

---

## Target Audiences

| Audience | Focus |
|----------|-------|
| **developers** | Internal team members working on the codebase |
| **users** | End users of the application/library |
| **contributors** | Open source contributors or new team members |
| **maintainers** | Long-term maintenance and operations |

---

## Priority Classification

| Priority | Description |
|----------|-------------|
| **high** | Entry points, frequently used APIs, onboarding blockers |
| **medium** | Complex areas, less-used features |
| **low** | Nice to have, edge cases |

---

## Output Format

Provide your analysis as structured JSON:

```json
{
  "documentation_gaps": [
    {
      "id": "doc-001",
      "title": "Add API documentation for authentication module",
      "description": "The auth/ module exports 12 functions but only 3 have JSDoc comments. Key functions like validateToken() and refreshSession() are undocumented.",
      "rationale": "Authentication is a critical module used throughout the app. Developers must read source code to understand it.",
      "category": "readme|api_docs|inline_comments|examples|architecture|troubleshooting",
      "target_audience": "developers|users|contributors|maintainers",
      "affected_areas": ["src/auth/token.ts", "src/auth/session.ts"],
      "current_documentation": "Only basic type exports are documented",
      "proposed_content": "Add JSDoc for all public functions including parameters, return values, errors thrown, and usage examples",
      "priority": "high|medium|low",
      "estimated_effort": "trivial|small|medium|large"
    }
  ],
  "summary": {
    "files_analyzed": 0,
    "documented_functions": 0,
    "undocumented_functions": 0,
    "issues_by_category": {},
    "issues_by_priority": {}
  }
}
```

---

## What to Look For

### README Quality Checklist
- [ ] Clear project description
- [ ] Installation instructions
- [ ] Basic usage example
- [ ] Configuration options
- [ ] API overview (if applicable)
- [ ] Contributing guidelines
- [ ] License information

### API Documentation Signals
```typescript
// GOOD - Well documented
/**
 * Validates a JWT token and returns the decoded payload.
 * @param token - The JWT string to validate
 * @returns The decoded token payload
 * @throws TokenExpiredError if token has expired
 * @throws InvalidTokenError if token signature is invalid
 * @example
 * const payload = validateToken('eyJ...');
 */
function validateToken(token: string): TokenPayload { }

// BAD - Undocumented
function validateToken(token: string): TokenPayload { }
```

### Complex Code Needing Comments
```typescript
// Needs explanation - magic numbers
const result = value * 0.0254 * 1.3;

// Needs explanation - non-obvious logic
if (date.getDay() === 0 || date.getDay() === 6) {
  date.setDate(date.getDate() + (8 - date.getDay()));
}

// Needs explanation - workaround
// @ts-ignore
response.headers['x-custom'] = value;
```

---

## Guidelines

1. **Be Specific**: Point to exact files and functions, not vague areas
2. **Prioritize Impact**: Focus on what helps new developers most
3. **Consider Audience**: Distinguish between user docs and contributor docs
4. **Realistic Scope**: Each idea should be completable in one session
5. **Avoid Redundancy**: Don't suggest docs that exist in different form

---

## Instructions

1. Scan all documentation files (README, docs/, markdown files)
2. Analyze code for JSDoc/docstring coverage
3. Identify public APIs without documentation
4. Find complex code lacking explanations
5. Cross-reference documented vs undocumented code
6. Prioritize by impact on onboarding and maintenance
7. Output the structured JSON with your findings

Begin your analysis now.
