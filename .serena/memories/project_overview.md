# Project Overview

## Purpose
shotgun-cli is a terminal-based prompt generation tool built with Go and BubbleTea. It's designed to generate structured LLM prompts from codebase context using an interactive terminal user interface.

## Key Features
- **Interactive Terminal UI**: Built with BubbleTea for smooth TUI experience
- **Inverse File Selection**: Users exclude files rather than include them (more intuitive for large projects)
- **Advanced File Filtering**: Automatic .gitignore support, built-in ignore patterns, custom ignore rules
- **Template System**: 4 built-in templates + extensible custom templates with YAML frontmatter
- **Progress Tracking**: Real-time progress bars for large projects
- **Cross-Platform**: Windows, macOS, Linux support with platform-specific optimizations

## Target Users
- Developers who need to generate structured prompts for LLMs from their codebase
- Teams working with large codebases who need context-aware prompt generation
- Users who prefer terminal-based tools over GUI applications

## Distribution
- Global installation via npm (`npm install -g shotgun-cli`)
- Cross-platform binaries generated for Windows, Linux, macOS (x64 and ARM64)
- Hybrid npm/Go build system for easy distribution