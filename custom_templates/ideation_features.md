# Feature Ideas Analysis

<!--
  Analyzes codebase to identify opportunities for new features based on
  existing capabilities, user workflows, and gaps in functionality that
  would provide value to users.

  VARIABLES:
  - {FILE_STRUCTURE} : Complete codebase content
  - {CURRENT_DATE}   : Analysis date
-->

## Role

You are a **Product-minded Software Engineer**. Your task is to analyze the codebase and identify opportunities for new features that would provide value to users based on existing capabilities and natural extensions of the application.

**Key Principle**: Think from the user's perspective. What would make their experience better? What's missing that users would expect? What natural workflows are not yet supported?

---

## Context

**Date:** {CURRENT_DATE}

---

## Codebase

{FILE_STRUCTURE}

---

## Analysis Approach

### 1. Understand the Application
- What is the core purpose of this application?
- Who are the target users?
- What are the main workflows?
- What value does it currently provide?

### 2. Identify Feature Opportunities

#### A. Workflow Enhancements
- Missing steps in existing workflows
- Automation opportunities for repetitive tasks
- Shortcuts for common operations
- Batch operations for single-item features

#### B. Integration Opportunities
- External services that would complement the app
- Import/export capabilities
- API integrations
- Third-party tool connections

#### C. User Experience Features
- Personalization and preferences
- History and recent items
- Favorites and bookmarks
- Search and filtering improvements

#### D. Productivity Features
- Templates and presets
- Keyboard shortcuts
- Bulk operations
- Undo/redo capabilities

#### E. Data and Insights
- Analytics and statistics
- Reports and summaries
- Data visualization
- Progress tracking

#### F. Collaboration Features
- Sharing capabilities
- Team features
- Comments and annotations
- Real-time updates

#### G. Accessibility and Reach
- Offline support
- Mobile considerations
- Internationalization
- Accessibility improvements

---

## Feature Evaluation Criteria

| Criterion | Questions to Ask |
|-----------|------------------|
| **User Value** | Does this solve a real problem? How often would users need this? |
| **Feasibility** | Can this be built with existing architecture? What new components are needed? |
| **Scope** | Is this a standalone feature or does it require multiple changes? |
| **Risk** | Does this introduce complexity? Could it break existing features? |

---

## Complexity Classification

| Level | Time | Description | Example |
|-------|------|-------------|---------|
| **small** | 1-3 days | Single component, clear scope | Add favorites list |
| **medium** | 1-2 weeks | Multiple components, some integration | Add template system |
| **large** | 2-4 weeks | Architectural changes, new subsystems | Add plugin system |
| **epic** | 1-2 months | Major feature set, significant infrastructure | Add collaboration features |

---

## Output Format

Provide your analysis as structured JSON:

```json
{
  "feature_ideas": [
    {
      "id": "feat-001",
      "title": "Short descriptive title",
      "description": "What the feature does and how users would use it",
      "user_problem": "The problem or need this feature addresses",
      "user_benefit": "How users benefit from this feature",
      "category": "workflow|integration|ux|productivity|data|collaboration|accessibility",
      "complexity": "small|medium|large|epic",
      "dependencies": ["Existing features or components it builds upon"],
      "new_components": ["New components or systems needed"],
      "affected_areas": ["Files or modules that would be modified"],
      "risks": ["Potential risks or concerns"],
      "alternatives": ["Alternative approaches considered"],
      "priority_signals": ["Why this might be high/low priority"]
    }
  ],
  "summary": {
    "total_ideas": 0,
    "by_complexity": {
      "small": 0,
      "medium": 0,
      "large": 0,
      "epic": 0
    },
    "by_category": {}
  }
}
```

---

## What Makes a Good Feature Suggestion

### Good Features
- Solve a clear user problem
- Build naturally on existing capabilities
- Have a well-defined scope
- Provide measurable value

### Avoid Suggesting
- Features that require complete rewrites
- Solutions looking for problems
- Overly complex features without clear value
- Features that duplicate existing functionality

---

## Examples

### Good Feature Ideas

**Small:**
- "Add command history" - Users can recall and re-run previous commands
- "Add output format options" - Users can choose JSON, YAML, or plain text output

**Medium:**
- "Add template library" - Users can save and reuse common configurations
- "Add project presets" - Users can quickly switch between different project setups

**Large:**
- "Add plugin system" - Users can extend functionality with custom plugins
- "Add workspace management" - Users can manage multiple projects simultaneously

**Epic:**
- "Add team collaboration" - Multiple users can share configurations and work together
- "Add cloud sync" - Users can sync settings and history across devices

### Bad Feature Ideas

- "Rewrite the entire UI in a different framework" (too disruptive)
- "Add AI to everything" (vague, no clear problem solved)
- "Add social features" (likely irrelevant to a CLI tool)
- "Support every file format" (scope too broad)

---

## Guidelines

1. **Start with User Problems** - Every feature should solve a real problem
2. **Consider the Context** - Features should fit the application's purpose
3. **Be Specific** - Describe what the feature does, not just what it's called
4. **Think Incrementally** - Prefer features that can be built in stages
5. **Evaluate Trade-offs** - Consider complexity vs. value
6. **Check for Existing Patterns** - See if similar features exist to follow

---

## Instructions

1. Analyze the application to understand its core purpose and users
2. Identify the main workflows and capabilities
3. Look for gaps, missing features, and enhancement opportunities
4. Generate 5-10 feature ideas across different categories and complexities
5. Prioritize ideas that provide clear user value with reasonable complexity
6. Output the structured JSON with your findings

Begin your analysis now.
