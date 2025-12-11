# Changelog

All notable changes to audit-cli will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-12-11

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
    - `--find-consolidation-candidates` - Show identical and similar composables for consolidation
    - `--find-usages` - Show where each composable is used in RST files with file paths

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

