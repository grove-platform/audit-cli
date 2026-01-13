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
	"sort"
	"strings"

	"github.com/grove-platform/audit-cli/internal/config"
	"github.com/spf13/cobra"
)

// NewTestableCodeCommand creates the testable-code subcommand.
func NewTestableCodeCommand() *cobra.Command {
	var outputFormat string
	var showDetails bool
	var outputFile string
	var filters []string
	var listDrivers bool

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

Filters (use --filter to focus on specific product areas):
  - search: Pages with "atlas-search" or "search" in URL (excludes vector-search)
  - vector-search: Pages with "vector-search" in URL
  - drivers: All MongoDB driver documentation pages
  - driver:<name>: Specific driver. Testable values include:
      csharp, golang, java, node, pymongo
    For the full list of options, use the --list-drivers flag.
  - mongosh: MongoDB Shell documentation pages

Multiple filters can be specified to include pages matching any filter.

Use --list-drivers to see available Driver filter options

Output formats:
  - text: Human-readable report with summary and detailed sections
  - json: Machine-readable JSON output
  - csv: Comma-separated values (summary by default, use --details for per-product breakdown)`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle --list-drivers flag
			if listDrivers {
				return runListDrivers()
			}

			// Require CSV file if not listing drivers
			if len(args) < 1 {
				return fmt.Errorf("requires at least 1 arg(s), only received 0")
			}

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

			return runTestableCode(csvPath, monorepoPath, outputFormat, showDetails, outputFile, filters)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format: text, json, or csv")
	cmd.Flags().BoolVar(&showDetails, "details", false, "Show detailed per-product breakdown (for csv: one row per product per page)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().StringSliceVar(&filters, "filter", nil, "Filter pages by product area (search, vector-search, drivers, driver:<name>, mongosh)")
	cmd.Flags().BoolVar(&listDrivers, "list-drivers", false, "List all drivers from the Snooty Data API")

	return cmd
}

// runListDrivers lists all drivers from the Snooty Data API.
func runListDrivers() error {
	// Use the version that doesn't require a monorepo path
	urlMapping, err := config.GetURLMappingWithoutMonorepo()
	if err != nil {
		return fmt.Errorf("failed to get URL mapping: %w", err)
	}

	driverSlugs := urlMapping.GetDriverSlugs()
	if len(driverSlugs) == 0 {
		fmt.Println("No drivers found in the Snooty Data API.")
		return nil
	}

	// Build a list of driver info and sort by project name (the filter value)
	type driverInfo struct {
		projectName string
		slug        string
		hasTestInfra bool
	}
	drivers := make([]driverInfo, 0, len(driverSlugs))
	for _, slug := range driverSlugs {
		projectName := urlMapping.URLSlugToProject[slug]
		drivers = append(drivers, driverInfo{
			projectName:  projectName,
			slug:         slug,
			hasTestInfra: TestableDrivers[projectName],
		})
	}
	// Sort alphabetically by project name
	sort.Slice(drivers, func(i, j int) bool {
		return drivers[i].projectName < drivers[j].projectName
	})

	fmt.Println("Available driver filters:")
	fmt.Println("=========================")
	fmt.Println()
	fmt.Println("Use --filter driver:<name> with any of these values:")
	fmt.Println()
	for _, d := range drivers {
		testableMarker := ""
		if d.hasTestInfra {
			testableMarker = " (has test infrastructure)"
		}
		fmt.Printf("  --filter driver:%-20s  (URL slug: %s)%s\n", d.projectName, d.slug, testableMarker)
	}
	fmt.Println()
	fmt.Println("Drivers with test infrastructure:")
	fmt.Printf("  %s\n", strings.Join(getTestableDriverNames(), ", "))
	fmt.Println()
	fmt.Println("Note: mongodb-shell is not a driver. Use --filter mongosh instead.")

	return nil
}

// runTestableCode is the main entry point for the testable-code command.
func runTestableCode(csvPath, monorepoPath, outputFormat string, showDetails bool, outputFile string, filters []string) error {
	// Parse CSV file
	entries, err := ParseCSV(csvPath)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Parsed %d pages from CSV\n", len(entries))

	// Get URL mapping early - needed for driver filters
	urlMapping, err := config.GetURLMapping(monorepoPath)
	if err != nil {
		return fmt.Errorf("failed to get URL mapping: %w", err)
	}

	// Validate filters before applying
	if err := validateFilters(filters); err != nil {
		return err
	}

	// Apply URL filters if specified
	if len(filters) > 0 {
		originalCount := len(entries)
		entries = filterEntries(entries, filters, urlMapping)
		fmt.Fprintf(os.Stderr, "Filtered to %d pages matching filter(s): %v\n", len(entries), filters)
		if len(entries) == 0 {
			fmt.Fprintf(os.Stderr, "Warning: No pages matched the specified filter(s). Original count: %d\n", originalCount)
		}
	}

	// Load product mappings from rstspec.toml
	fmt.Fprintf(os.Stderr, "Loading product mappings from rstspec.toml...\n")
	mappings, err := LoadProductMappings()
	if err != nil {
		return fmt.Errorf("failed to load product mappings: %w", err)
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

// filterEntries filters page entries based on the specified filters.
// Returns entries that match any of the specified filters.
func filterEntries(entries []PageEntry, filters []string, urlMapping *config.URLMapping) []PageEntry {
	var filtered []PageEntry
	for _, entry := range entries {
		if matchesAnyFilter(entry.URL, filters, urlMapping) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// matchesAnyFilter checks if a URL matches any of the specified filters.
func matchesAnyFilter(url string, filters []string, urlMapping *config.URLMapping) bool {
	for _, filter := range filters {
		if matchesFilter(url, filter, urlMapping) {
			return true
		}
	}
	return false
}

// validateFilters validates that all specified filters are valid.
// Returns an error if any filter is invalid.
func validateFilters(filters []string) error {
	for _, filter := range filters {
		filterLower := strings.ToLower(filter)

		// Check for driver:<name> pattern - any driver name is valid
		if strings.HasPrefix(filterLower, "driver:") {
			driverName := strings.TrimPrefix(filterLower, "driver:")
			// mongodb-shell should use mongosh filter since it's not a driver
			if driverName == "mongodb-shell" {
				return fmt.Errorf("invalid filter %q: mongodb-shell is not a driver, use --filter mongosh instead", filter)
			}
			// Any other driver name is valid - will just return no results if not found
			continue
		}

		// Check known filters
		switch filterLower {
		case "search", "vector-search", "drivers", "mongosh":
			// Valid filters
		default:
			return fmt.Errorf("unknown filter %q.\nValid filters: search, vector-search, drivers, driver:<name>, mongosh\nUse --list-drivers to see available driver names", filter)
		}
	}
	return nil
}

// getTestableDriverNames returns a sorted list of driver names with test infrastructure.
func getTestableDriverNames() []string {
	var names []string
	for name := range TestableDrivers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// matchesFilter checks if a URL matches a specific filter.
// Matching is case-insensitive.
//
// Supported filters:
//   - "search": matches URLs containing "atlas-search" or "search" but NOT "vector-search"
//   - "vector-search": matches URLs containing "vector-search"
//   - "drivers": matches all driver documentation URLs (excludes mongodb-shell)
//   - "driver:<name>": matches a specific driver by project name (e.g., driver:pymongo)
//   - "mongosh": matches MongoDB Shell documentation URLs
func matchesFilter(url string, filter string, urlMapping *config.URLMapping) bool {
	urlLower := strings.ToLower(url)
	filterLower := strings.ToLower(filter)

	// Check for driver:<name> pattern
	if strings.HasPrefix(filterLower, "driver:") {
		driverName := strings.TrimPrefix(filterLower, "driver:")
		return urlMapping.IsSpecificDriverURL(url, driverName)
	}

	switch filterLower {
	case "search":
		// Match "atlas-search" or "search" but exclude "vector-search"
		if strings.Contains(urlLower, "vector-search") {
			return false
		}
		return strings.Contains(urlLower, "atlas-search") || strings.Contains(urlLower, "search")
	case "vector-search":
		return strings.Contains(urlLower, "vector-search")
	case "drivers":
		return urlMapping.IsDriverURL(url)
	case "mongosh":
		return urlMapping.IsMongoshURL(url)
	default:
		// This shouldn't happen if validateFilters was called first
		return false
	}
}
