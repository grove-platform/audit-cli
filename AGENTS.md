# AGENTS.md - LLM Development Guide for audit-cli

This document provides essential context for LLMs performing development tasks in the `audit-cli` repository.

## Repository Overview

**Purpose**: A Go CLI tool for auditing and analyzing MongoDB's reStructuredText (RST) documentation.

**Key Capabilities**:
- Extract code examples and procedures from RST files
- Search documentation for patterns
- Analyze file dependencies and relationships
- Analyze composable definitions and usage across projects
- Compare files across documentation versions
- Count documentation pages and tested code examples

**Target Users**: MongoDB technical writers performing maintenance, scoping work, and reporting.

## Project Structure

```
audit-cli/
├── main.go                    # CLI entry point using cobra
├── go.mod                     # Module: github.com/grove-platform/audit-cli
├── commands/                  # Command implementations (parent + subcommands)
│   ├── extract/              # Extract content from RST files
│   │   ├── extract.go        # Parent command
│   │   ├── code-examples/    # Extract code examples subcommand
│   │   └── procedures/       # Extract procedures subcommand
│   ├── search/               # Search through files
│   │   └── find-string/      # Find string subcommand
│   ├── analyze/              # Analyze RST structures
│   │   ├── includes/         # Analyze include relationships
│   │   ├── usage/            # Find file usages
│   │   ├── procedures/       # Analyze procedure variations
│   │   └── composables/      # Analyze composable definitions and usage
│   ├── compare/              # Compare files across versions
│   │   └── file-contents/    # Compare file contents
│   └── count/                # Count documentation content
│       ├── tested-examples/  # Count tested code examples
│       └── pages/            # Count documentation pages
├── internal/                 # Internal packages (not importable externally)
│   ├── projectinfo/          # MongoDB docs project structure utilities
│   │   ├── pathresolver.go  # Path resolution
│   │   ├── source_finder.go # Source directory detection
│   │   └── version_resolver.go # Version path resolution
│   └── rst/                  # RST parsing utilities
│       ├── parser.go         # Generic parsing with includes
│       ├── directive_parser.go # Directive parsing
│       ├── directive_regex.go  # Regex patterns for directives
│       ├── parse_procedures.go # Procedure parsing (core logic)
│       ├── get_procedure_variations.go # Variation extraction
│       └── rstspec.go        # Fetch and parse canonical rstspec.toml
├── testdata/                 # Test fixtures (auto-ignored by Go build)
│   ├── input-files/source/   # Test RST files
│   ├── expected-output/      # Expected extraction results
│   ├── compare/              # Compare command test data
│   └── count-test-monorepo/  # Count command test data
├── bin/                      # Build output directory
├── docs/                     # Additional documentation
│   └── PROCEDURE_PARSING.md  # Detailed procedure parsing logic
└── README.md                 # Comprehensive user documentation

```

## Technology Stack

- **Language**: Go
- **CLI Framework**: [spf13/cobra](https://github.com/spf13/cobra)
- **Diff Library**: [aymanbagabas/go-udiff](https://github.com/aymanbagabas/go-udiff)
- **YAML Parsing**: gopkg.in/yaml.vX
- **TOML Parsing**: [github.com/BurntSushi/toml](https://github.com/BurntSushi/toml) v1.5.0
- **Testing**: Go standard library (`testing` package)

Refer to the `go.mod` for version info.

## Domain Knowledge: MongoDB Documentation

### RST Directives Supported

**Code Example Directives**:
- `.. literalinclude::` - Transclude code from external files
- `.. code-block::` - Inline code blocks
- `.. io-code-block::` - Input/output code examples with `.. input::` and `.. output::` sub-directives

**Procedure Directives**:
- `.. procedure::` with `.. step::` - Structured procedures
- Ordered lists (numbered `1.` or lettered `a.`) - Simple procedures
- `#.` - Continuation marker (auto-numbered)
- YAML steps files - Converted to procedures during build

**Variation Mechanisms**:
- `.. composable-tutorial::` with `.. selected-content::` - Content variations by selection
- `.. tabs::` with `.. tab::` and `:tabid:` - Tabbed content variations

**Content Inclusion**:
- `.. include::` - Include RST content from other files
- `.. toctree::` - Table of contents (navigation, not content inclusion)

**Composables**:
- Defined in `snooty.toml` files at project/version root
- Canonical definitions also exist in `rstspec.toml` in the snooty-parser repository
- Used in `.. composable-tutorial::` directives with `:options:` parameter
- Enable context-specific documentation (e.g., different languages, deployment types)
- Each composable has an ID, title, default, and list of options
- The `internal/rst` module provides `FetchRstspec()` to retrieve canonical definitions

### MongoDB Documentation Structure

**Versioned Projects**: `content/{project}/{version}/source/`
- Versions: `manual`, `current`, `upcoming`, `v8.0`, `v7.0`, etc.

**Non-versioned Projects**: `content/{project}/source/`

**Tested Code Examples**: `content/code-examples/tested/{language}/{product}/`
- Products: `pymongo`, `mongosh`, `go/driver`, `go/atlas-sdk`, `javascript/driver`, `java/driver-sync`, `csharp/driver`

## Building and Running

### Build from Source
```bash
cd bin
go build ../
./audit-cli --help
```

### Run Without Building
```bash
go run main.go [command] [flags]
```

### Check Version
```bash
./audit-cli --version
# Output: audit-cli version 0.1.0
```

### Run Tests
```bash
# All tests
go test ./...

# Specific package
go test ./commands/extract/code-examples -v

# Specific test
go test ./commands/extract/code-examples -run TestLiteralIncludeDirective -v

# With coverage
go test ./... -cover
```

## Versioning

This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html) (SemVer):

- **MAJOR** version (X.0.0): Incompatible API changes or breaking changes to command behavior
- **MINOR** version (0.X.0): New functionality added in a backward-compatible manner
- **PATCH** version (0.0.X): Backward-compatible bug fixes

**Note**: While in `0.x.x` versions, breaking changes may occur in minor releases. Version `1.0.0` will signal a stable, production-ready release.

### When to Increment Versions

- **MAJOR** (e.g., 0.5.0 → 1.0.0):
  - Breaking changes to command syntax or flags
  - Removal of commands or features
  - Changes to output format that break existing scripts
  - First stable release (0.x.x → 1.0.0)

- **MINOR** (e.g., 0.1.0 → 0.2.0):
  - New commands or subcommands
  - New flags or options
  - New RST directive support
  - New output formats (when existing formats remain unchanged)
  - Significant new features

- **PATCH** (e.g., 0.1.0 → 0.1.1):
  - Bug fixes
  - Performance improvements
  - Documentation updates
  - Internal refactoring with no user-facing changes

### Releasing a New Version

When releasing a new version, follow these steps:

1. **Update the version constant** in `main.go`:
   ```go
   const version = "0.2.0"  // Update this line
   ```

2. **Update CHANGELOG.md** following the [Keep a Changelog](https://keepachangelog.com/) format:
   ```markdown
   ## [0.2.0] - YYYY-MM-DD

   ### Added
   - New feature descriptions

   ### Changed
   - Modified behavior descriptions

   ### Fixed
   - Bug fix descriptions

   ### Removed
   - Removed feature descriptions (if any)
   ```

3. **Test the version output**:
   ```bash
   go run main.go --version
   # Should display: audit-cli version 0.2.0
   ```

4. **Commit the changes**:
   ```bash
   git add main.go CHANGELOG.md
   git commit -m "Release version 0.2.0"
   ```

5. **Tag the release** (optional but recommended):
   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```

### CHANGELOG Format

The CHANGELOG.md follows the [Keep a Changelog](https://keepachangelog.com/) format with these sections:

- **Added**: New features, commands, or capabilities
- **Changed**: Changes to existing functionality
- **Deprecated**: Features that will be removed in future versions
- **Removed**: Features that have been removed
- **Fixed**: Bug fixes
- **Security**: Security-related changes

Each version entry should include:
- Version number in square brackets: `[0.2.0]`
- Release date in ISO format: `YYYY-MM-DD`
- Organized sections with bullet points describing changes
- User-facing language (avoid technical jargon when possible)

## Development Patterns

### Command Structure

**Parent-Subcommand Pattern**: All commands follow a two-level hierarchy:
- Parent command (e.g., `extract`, `analyze`) - defined in `commands/{parent}/{parent}.go`
- Subcommand (e.g., `code-examples`, `procedures`) - defined in `commands/{parent}/{subcommand}/{subcommand}.go`

**File Organization per Subcommand**:
```
commands/{parent}/{subcommand}/
├── {subcommand}.go       # Command definition and RunE function
├── {subcommand}_test.go  # Tests
├── types.go              # Type definitions
├── parser.go or analyzer.go  # Core logic
├── output.go or report.go    # Output formatting
└── (other domain-specific files)
```

**Command Registration**: Parent commands register subcommands in their `New{Parent}Command()` function:
```go
func NewExtractCommand() *cobra.Command {
    cmd := &cobra.Command{Use: "extract", Short: "..."}
    cmd.AddCommand(codeexamples.NewCodeExamplesCommand())
    cmd.AddCommand(procedures.NewProceduresCommand())
    return cmd
}
```

### Error Handling

- Use `fmt.Errorf()` for error wrapping with context
- Return errors from functions; handle at command level
- Print errors to stderr using `fmt.Fprintf(os.Stderr, ...)`
- Exit with non-zero status on errors

### Testing Conventions

**Test Data Location**: `testdata/` directory (auto-ignored by Go build)
- Input files: `testdata/input-files/source/`
- Expected output: `testdata/expected-output/`
- Relative path from test: `filepath.Join("..", "..", "..", "testdata")`

**Test Patterns**:
- Table-driven tests for multiple scenarios
- Temporary directories for output: `os.MkdirTemp("", "test-*")`
- Clean up with `defer os.RemoveAll(tempDir)`
- Byte-for-byte content comparison for extracted files

**Deterministic Testing**: Critical for procedure parsing
- All map iterations must be sorted
- Content hashing must use sorted keys
- Run tests multiple times to verify determinism

### Code Patterns

**Path Resolution**:
- Use `filepath.Join()` for cross-platform paths
- Use `filepath.Abs()` to get absolute paths
- Use `internal/projectinfo` for MongoDB-specific path resolution

**RST Parsing**:
- Use `internal/rst` package for directive parsing
- Use regex patterns from `internal/rst/directive_regex.go`
- Handle include directives with `ParseFileWithIncludes()`

**Output Formatting**:
- Separate output logic into `output.go` or `report.go`
- Support multiple output formats (text, JSON) where applicable
- Use consistent formatting (headers with `=` separators, indentation)

## Key Design Decisions

### RST Parsing Strategy

**Incremental, Opportunistic Parsing**: RST parsing capabilities are added incrementally as needed by various commands, rather than using a complete AST-based parser.

**Rationale**:
- MongoDB documentation uses many custom directives not supported by standard reStructuredText parsing libraries
- A complete parser converting RST to an AST would require significant operational overhead that is not needed at this time
- Targeted parsing for specific directives is more maintainable and performant for this use case

**Critical Rule**: All new RST parsing functionality MUST be added to the `internal/rst` module, NOT to individual command modules. This ensures:
- Parsing capabilities can be reused across commands
- Functionality can be expanded incrementally
- Parsing logic remains centralized and maintainable

**Implementation Pattern**:
1. Add regex patterns to `internal/rst/directive_regex.go`
2. Add directive types and parsing logic to `internal/rst/directive_parser.go`
3. Add specialized parsing functions (e.g., `parse_procedures.go`) as separate files in `internal/rst`
4. Commands import and use functions from `internal/rst` package

### Procedure Parsing

**Uniqueness**: Procedures are grouped by heading + content hash
- Same heading, different content → separate procedures
- Same content, multiple selections → one procedure with multiple appearances

**Analysis vs. Extraction**:
- **Analysis**: Groups procedures by heading, shows variations
- **Extraction**: Creates one file per unique procedure (by content hash)

**Determinism**: All operations must produce consistent results
- Sort all map iterations
- Use sorted keys for hashing
- Critical for testing and CI/CD

See `docs/PROCEDURE_PARSING.md` for comprehensive details.

### Include Directive Handling

**Context-Aware Expansion**:
- No composable tutorial → Expand all includes globally
- Composable tutorial with selected-content in main file → Expand includes within blocks
- Composable tutorial with includes in steps → Expand to detect selected-content blocks

### Version Comparison

**Auto-Discovery**: Automatically detects product directory and available versions from file path
- Pattern: `/path/to/{product}/{version}/source/...`
- Discovers all sibling version directories unless `--versions` specified

## Adding New Commands

### 1. Create Subcommand Directory
```bash
mkdir -p commands/{parent}/{subcommand}
```

### 2. Create Command File
Create `commands/{parent}/{subcommand}/{subcommand}.go`:
```go
package subcommand

import (
    "github.com/spf13/cobra"
)

func NewSubcommandCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "subcommand",
        Short: "Brief description",
        Long:  `Detailed description`,
        Args:  cobra.ExactArgs(1), // or cobra.NoArgs, etc.
        RunE:  runSubcommand,
    }

    // Add flags
    cmd.Flags().StringP("output", "o", "./output", "Output directory")

    return cmd
}

func runSubcommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

### 3. Create Supporting Files
- `types.go` - Type definitions
- `{subcommand}_test.go` - Tests
- Domain-specific files (parser, analyzer, output, etc.)

### 4. Register with Parent Command
In `commands/{parent}/{parent}.go`:
```go
import "github.com/mongodb/code-example-tooling/audit-cli/commands/{parent}/{subcommand}"

func New{Parent}Command() *cobra.Command {
    cmd := &cobra.Command{...}
    cmd.AddCommand(subcommand.NewSubcommandCommand())
    return cmd
}
```

### 5. Add Tests
Create test fixtures in `testdata/` and write tests following existing patterns.

## Common Tasks

### Adding a New RST Directive

1. **Add regex pattern** to `internal/rst/directive_regex.go`:
```go
var MyDirectiveRegex = regexp.MustCompile(`^\.\.\s+my-directive::\s*(.*)$`)
```

2. **Add directive type** to `internal/rst/directive_parser.go`:
```go
const (
    MyDirective DirectiveType = "my-directive"
)
```

3. **Add parsing logic** in `ParseDirectives()` function

4. **Add tests** in appropriate test file

### Updating Expected Test Output

When changing parsing logic:
```bash
# Regenerate expected output
./bin/audit-cli extract code-examples testdata/input-files/source/literalinclude-test.rst \
  -o testdata/expected-output

# Verify changes are correct before committing
git diff testdata/expected-output/
```

### Adding Support for a New Product

For `count tested-examples` command:

1. Add product to valid products list in `commands/count/tested-examples/counter.go`
2. Update README.md with new product
3. Add test data to `testdata/count-test-monorepo/content/code-examples/tested/`
4. Update tests in `tested_examples_test.go`

## Important Notes for LLMs

### When Making Changes

1. **Always check existing patterns** - Look at similar commands/functions before implementing new ones
2. **Maintain consistency** - Follow the established file organization and naming conventions
3. **Update tests** - All changes should include corresponding test updates
4. **Check determinism** - For procedure parsing, verify output is consistent across runs
5. **Update documentation** - Keep README.md, AGENTS.md, and PROCEDURE_PARSING.md in sync with code changes

### Common Pitfalls

1. **Map iteration order** - Always sort map keys before iterating (especially in procedure parsing)
2. **Path separators** - Use `filepath.Join()` instead of string concatenation
3. **Relative paths in tests** - Remember tests are 3 levels deep: `commands/{parent}/{subcommand}/`
4. **Include directive resolution** - Use `internal/projectinfo` for MongoDB-specific path conventions
5. **Testdata directory** - This is a special Go convention; files here are ignored during builds

### Testing Requirements

- **Unit tests** for all new functions
- **Integration tests** for command execution
- **Determinism tests** for procedure parsing
- **Table-driven tests** for multiple scenarios
- **Test coverage** should not decrease

### Documentation Requirements

- **Package-level comments** for all packages
- **Function comments** for exported functions
- **README.md updates** for user-facing changes
- **PROCEDURE_PARSING.md updates** for procedure parsing logic changes
- **Inline comments** for complex logic

## Resources

- **README.md**: Comprehensive user documentation and development guide
- **docs/PROCEDURE_PARSING.md**: Detailed procedure parsing business logic
- **Go Cobra Documentation**: https://github.com/spf13/cobra
- **MongoDB RST Conventions**: See examples in `testdata/input-files/source/`

## Quick Reference

### File Naming Conventions
- Commands: `{subcommand}.go`
- Tests: `{subcommand}_test.go`
- Types: `types.go`
- Core logic: `parser.go`, `analyzer.go`, `counter.go`, etc.
- Output: `output.go`, `report.go`

### Import Path
```go
import "github.com/grove-platform/audit-cli/{package}"
```

### Running Specific Tests
```bash
# By package
go test ./internal/rst -v

# By function name
go test ./commands/extract/procedures -run TestParseFileDeterministic -v

# With race detection
go test ./... -race
```

### Build Tags
None currently used in this project.

### Environment Variables
None currently used in this project.

## Continuous Integration

### GitHub Actions Workflow

The project uses GitHub Actions to automatically run tests on pull requests and pushes to `main`.

**Workflow file**: `.github/workflows/run-tests.yml`

**What it does**:
- Runs on all PRs to `main` and pushes to `main`
- Sets up Go 1.24
- Caches Go modules for faster builds
- Runs all tests with race detection: `go test ./... -v -race -coverprofile=coverage.out`
- Displays test coverage summary

**Viewing results**:
- Check the "Actions" tab in GitHub to see workflow runs
- Each PR will show a green checkmark or red X indicating test status
- Click on the workflow run to see detailed test output and coverage

**Local testing before pushing**:
```bash
# Run the same tests that CI runs
go test ./... -v -race -coverprofile=coverage.out

# View coverage summary
go tool cover -func=coverage.out | tail -1
```
