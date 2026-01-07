// Package testablecode provides the testable-code subcommand for the report command.
//
// This command analyzes code examples on documentation pages based on analytics data.
// It takes a CSV file with page rankings and URLs, resolves each URL to its source file,
// collects code examples, and generates a report with testability information.
//
// # Purpose
//
// The primary goal is to help identify which high-traffic documentation pages have
// code examples that COULD be tested but currently ARE NOT. This helps prioritize
// efforts to add test coverage to the most impactful pages.
//
// # Key Concepts
//
// Product vs Language:
// A "product" is a MongoDB driver or tool (e.g., "Python", "Node.js", "MongoDB Shell").
// A "language" is the programming language of a code example (e.g., "python", "javascript").
// The same language can map to different products depending on context.
//
// Testable vs Tested:
// "Testable" means the code example is for a product that has test infrastructure.
// "Tested" means the code example actually references tested code (literalinclude from
// the tested code examples directory).
//
// Context Inheritance:
// Code examples in included files inherit context from their parent. For example,
// if a file is included within a `.. selected-content:: :selections: python` block,
// all code examples in that file are attributed to "Python".
//
// # Special Cases
//
// The command handles several special cases documented in internal/language and code_collector.go:
//   - NonDriverLanguages: Languages that bypass context inheritance (bash, json, etc.) - see internal/language
//   - MongoShellLanguages: Languages that need MongoDB Shell context checking - see internal/language
//   - Content directory mapping: Driver content dirs map to products - see internal/projectinfo
package testablecode

import (
	"fmt"
	"os"

	"github.com/grove-platform/audit-cli/internal/config"
	"github.com/spf13/cobra"
)

// NewTestableCodeCommand creates the testable-code subcommand.
func NewTestableCodeCommand() *cobra.Command {
	var outputFormat string
	var showDetails bool
	var outputFile string

	cmd := &cobra.Command{
		Use:   "testable-code <csv-file> [monorepo-path]",
		Short: "Analyze testable code examples on pages from analytics data",
		Long: `Analyze testable code examples on documentation pages based on analytics CSV data.

Takes a CSV file with page rankings and URLs, resolves each URL to its source file
in the monorepo, collects code examples (literalinclude, code-block, io-code-block),
and generates a report with:
  - Total code examples per page
  - Breakdown by product/language
  - Input vs output counts (for io-code-block)
  - Tested vs untested counts
  - Testable count (examples that could be tested based on product)
  - Maybe testable count (javascript/shell examples without clear context)

The CSV file should have columns for rank and URL. The first row is treated as a header.

Example CSV format:
  rank,url
  1,www.mongodb.com/docs/atlas/some-page/
  2,www.mongodb.com/docs/manual/tutorial/install/

Testable products (have test infrastructure):
  - C#, Go, Java (Sync), Node.js, Python, MongoDB Shell

Output formats:
  - text: Human-readable report with summary and detailed sections
  - json: Machine-readable JSON output
  - csv: Comma-separated values (summary by default, use --details for per-product breakdown)`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			csvPath := args[0]

			// Get monorepo path
			var cmdLineArg string
			if len(args) > 1 {
				cmdLineArg = args[1]
			}
			monorepoPath, err := config.GetMonorepoPath(cmdLineArg)
			if err != nil {
				return err
			}

			return runTestableCode(csvPath, monorepoPath, outputFormat, showDetails, outputFile)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format: text, json, or csv")
	cmd.Flags().BoolVar(&showDetails, "details", false, "Show detailed per-product breakdown (for csv: one row per product per page)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")

	return cmd
}

// runTestableCode is the main entry point for the testable-code command.
func runTestableCode(csvPath, monorepoPath, outputFormat string, showDetails bool, outputFile string) error {
	// Parse CSV file
	entries, err := ParseCSV(csvPath)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Parsed %d pages from CSV\n", len(entries))

	// Load product mappings from rstspec.toml
	fmt.Fprintf(os.Stderr, "Loading product mappings from rstspec.toml...\n")
	mappings, err := LoadProductMappings()
	if err != nil {
		return fmt.Errorf("failed to load product mappings: %w", err)
	}

	// Get URL mapping
	urlMapping, err := config.GetURLMapping(monorepoPath)
	if err != nil {
		return fmt.Errorf("failed to get URL mapping: %w", err)
	}

	// Analyze each page
	var reports []PageReport
	for i, entry := range entries {
		fmt.Fprintf(os.Stderr, "Analyzing page %d/%d: %s\n", i+1, len(entries), entry.URL)

		analysis, err := AnalyzePage(entry, urlMapping, mappings)
		if err != nil {
			// Log error but continue with other pages
			fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
			reports = append(reports, PageReport{
				Rank:  entry.Rank,
				URL:   entry.URL,
				Error: err.Error(),
			})
			continue
		}

		report := BuildPageReport(analysis)
		reports = append(reports, report)
	}

	// Determine output writer
	var writer *os.File
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		writer = f
		fmt.Fprintf(os.Stderr, "Writing output to %s\n", outputFile)
	} else {
		writer = os.Stdout
	}

	// Output report
	switch outputFormat {
	case "json":
		return OutputJSON(writer, reports)
	case "csv":
		return OutputCSV(writer, reports, showDetails)
	default:
		return OutputText(writer, reports)
	}
}

