# Project Overview: shotgun-cli

## Purpose
shotgun-cli is a terminal-based prompt generation tool designed to help developers generate structured LLM prompts from codebase context. It's a command-line interface version of the Shotgun application that provides an interactive TUI experience for creating context-rich prompts for AI coding assistants.

## Key Features
- **Interactive Terminal UI**: Built with BubbleTea for smooth TUI experience
- **Inverse File Selection**: Select files to exclude rather than include (more intuitive for large projects)
- **Advanced File Filtering**: Automatic .gitignore support, built-in ignore patterns, custom ignore rules
- **Multiple Prompt Templates**: 
  - Dev: Generate git diffs for code changes
  - Architect: Create design plans and architecture
  - Debug: Bug analysis and debugging
  - Project Manager: Documentation sync and task management
- **Progress Tracking**: Real-time progress bars for large projects
- **Cross-Platform**: Works on Windows, macOS, and Linux

## Business Value
The tool helps developers quickly generate comprehensive, context-aware prompts for LLM interactions by automatically processing codebase files and applying specialized prompt templates based on the task type.