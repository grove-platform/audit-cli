# Changelog

All notable changes to audit-cli will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2025-01-07

### Added

#### Report Commands
- `report testable-code` - Analyze testable code examples on pages from analytics data
  - Takes a CSV file with page rankings and URLs from analytics
  - Resolves URLs to source files using the Snooty Data API
  - Collects code examples (literalinclude, code, code-block, io-code-block) from each page
  - Determines product context from tabs, composables, and content directories
  - Identifies tested vs testable vs "maybe testable" code examples
  - Supports multiple output formats: text, JSON, CSV
  - Flags:
    - `--format, -f` - Output format (text, json, csv)
    - `--output, -o` - Output file path (default: stdout)
    - `--details` - Show detailed per-product breakdown

#### Internal Packages
- `internal/language` - Programming language utilities (refactored from code-examples)
  - Language normalization (e.g., "ts" → "typescript", "py" → "python")
  - File extension mapping for all supported languages
  - Language-to-product mapping for MongoDB drivers
  - Non-driver language detection (bash, json, yaml, etc.)
  - MongoDB Shell language detection
  - Language resolution with priority: argument > option > file extension
- `internal/snooty` - Snooty.toml parsing utilities
  - Parse snooty.toml configuration files
  - Find project snooty.toml from source file paths
  - Build composable ID-to-title mappings
  - Extract project and version from snooty.toml paths
- `internal/config/url_mapping.go` - URL-to-source-file mapping
  - Fetches project metadata from Snooty Data API
  - Resolves documentation URLs to source file paths
  - Caches API responses for 24 hours in `~/.audit-cli/`
  - Supports offline usage with expired cache fallback
- `internal/projectinfo/products.go` - Content directory to product mapping
  - Maps driver content directories to display product names
  - Supports all MongoDB driver documentation projects
- `internal/rst/yaml_steps_parser.go` - YAML steps file parsing
  - Parses legacy YAML-native code examples in steps files
  - Extracts code blocks with language and content
- `internal/rst/directive_parser.go` - Enhanced directive parsing
  - Added `ResolveLanguage()` method to Directive type
  - Added `ResolveLanguage()` method to SubDirective type
  - Language resolution follows priority: argument > option > file extension

### Changed

- Refactored language handling from `commands/extract/code-examples/language.go` to `internal/language` package
  - All language-related utilities now centralized and reusable
  - Added product mapping and non-driver language detection
- Enhanced `internal/rst` directive parsing with language resolution methods
- Updated `analyze usage` to use new language package for file extension handling

## [0.2.0] - 2025-12-12

### Added

#### Analyze Commands
- `analyze composables` - Analyze composable definitions in snooty.toml files
  - Inventory all composables across projects and versions
  - Identify identical composables (same ID, title, and options) across different projects/versions
  - Find similar composables with different IDs but overlapping option sets using Jaccard similarity (60% threshold)
  - Track composable usage in RST files via `composable-tutorial` directives
  - Identify unused composables that may be candidates for removal
  - Flags:
    - `--for-project` - Filter to a specific project
    - `--current-only` - Only analyze current versions
    - `--verbose` - Show full option details with titles
    - `--find-similar` - Show identical and similar composables for consolidation
    - `--find-usages` - Show where each composable is used in RST files with file paths
    - `--with-rstspec` - Show canonical composable definitions from rstspec.toml

#### Configuration System
- Monorepo path configuration via three methods (priority order):
  1. Command-line argument (highest priority)
  2. Environment variable `AUDIT_CLI_MONOREPO_PATH`
  3. Config file `.audit-cli.yaml` in current or home directory (lowest priority)
- Config file format:
  ```yaml
  monorepo_path: /path/to/docs-monorepo
  ```
- Applies to commands: `analyze composables`, `count tested-examples`, `count pages`

#### File Path Resolution
- Flexible path resolution for all file-based commands
- Supports three path types (priority order):
  1. Absolute paths - Used as-is
  2. Relative to monorepo root - If monorepo configured and file exists there
  3. Relative to current directory - Fallback
- Applies to commands: `extract code-examples`, `extract procedures`, `analyze includes`, `analyze usage`, `search find-string`, `compare file-contents`
- Eliminates need to type full paths when working with monorepo files

#### Internal Packages
- `internal/config` - Configuration management
  - Config file loading from `.audit-cli.yaml`
  - Environment variable support
  - Monorepo path resolution with priority order
  - File path resolution with flexible resolution
- `internal/rst` - Enhanced RST parsing
  - `FetchRstspec()` - Fetches canonical composable definitions from snooty-parser rstspec.toml
  - Provides standard composable IDs, titles, defaults, and options

## [0.1.0] - 2025-12-10

### Added

Initial release after splitting from the MongoDB code-example-tooling monorepo.

#### Extract Commands
- `extract code-examples` - Extract code examples from RST files
  - Supports `literalinclude`, `code-block`, and `io-code-block` directives
  - Handles partial file extraction with `:start-after:`, `:end-before:`, `:lines:` options
  - Automatic language detection and normalization
  - Recursive directory scanning
  - Follow include directives to process entire documentation trees
- `extract procedures` - Extract procedures from RST files
  - Supports `procedure` directive with `step` sub-directives
  - Supports ordered lists (numbered and lettered)
  - Detects and extracts procedure variations (tabs, composable tutorials)
  - Content-based deduplication using hashing
  - Optional include directive expansion

#### Search Commands
- `search find-string` - Search for substrings in documentation files
  - Case-sensitive and case-insensitive search modes
  - Exact word matching or partial matching
  - Recursive directory scanning
  - Follow include directives
  - Language breakdown in verbose mode

#### Analyze Commands
- `analyze includes` - Analyze include directive relationships
  - Tree view of include dependencies
  - Flat list of all included files
  - Circular include detection
  - Maximum depth tracking
- `analyze usage` - Find all files that use a target file
  - Tracks `include`, `literalinclude`, and `io-code-block` references
  - Optional toctree entry tracking
  - Recursive mode to find all documentation pages that ultimately use a file
  - Path exclusion support
- `analyze procedures` - Analyze procedure variations and statistics
  - Count procedures and variations
  - Detect implementation types (directive vs ordered list)
  - Step count analysis
  - Sub-procedure detection
  - Variation listing (composable tutorial selections and tabids)

#### Compare Commands
- `compare file-contents` - Compare file contents across versions
  - Direct comparison between two files
  - Version comparison mode with auto-discovery
  - Unified diff output
  - Progressive detail levels (summary, paths, diffs)

#### Count Commands
- `count tested-examples` - Count tested code examples in the monorepo
  - Total count across all products
  - Per-product breakdown
  - Per-language breakdown
  - Supports: pymongo, mongosh, go/driver, go/atlas-sdk, javascript/driver, java/driver-sync, csharp/driver
- `count pages` - Count documentation pages (.txt files)
  - Total count across all projects
  - Per-project breakdown
  - Per-version breakdown
  - Automatic exclusions (code-examples, meta directories)
  - Custom directory exclusions
  - Current version only mode

#### Internal Packages
- `internal/rst` - RST parsing utilities
  - Directive parsing (literalinclude, code-block, io-code-block, procedure, step, tabs, composable-tutorial)
  - Include directive following with circular detection
  - Procedure parsing with variation detection
  - Content extraction with partial file support
- `internal/projectinfo` - MongoDB documentation structure utilities
  - Source directory detection
  - Product directory detection
  - Version path resolution

#### Documentation
- Comprehensive README.md with usage examples
- PROCEDURE_PARSING.md with detailed procedure parsing business logic
- AGENTS.md for LLM development assistance

### Technical Details
- Built with Go 1.24
- Uses spf13/cobra v1.10.1 for CLI framework
- Uses aymanbagabas/go-udiff v0.3.1 for diff generation
- Comprehensive test coverage with deterministic testing for procedure parsing

