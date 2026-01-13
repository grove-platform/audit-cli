// Package testablecode provides the testable-code subcommand for the report command.
package testablecode

import (
	"fmt"
	"sync"

	"github.com/grove-platform/audit-cli/internal/rst"
	"github.com/grove-platform/audit-cli/internal/snooty"
)

// PageEntry represents a single page from the analytics CSV.
type PageEntry struct {
	Rank int
	URL  string
}

// CodeExample represents a single code example found in a page.
type CodeExample struct {
	// Type is the directive type: literalinclude, code-block, code, io-code-block
	Type string
	// Language is the programming language (from :language: option or argument, or file extension)
	Language string
	// Product is the MongoDB product context (from tabs, composables, or content directory)
	Product string
	// IsInput indicates if this is an input block (for io-code-block)
	IsInput bool
	// IsOutput indicates if this is an output block (for io-code-block)
	IsOutput bool
	// IsTested indicates if the code example references tested code
	IsTested bool
	// IsTestable indicates if the code example could be tested (based on product)
	IsTestable bool
	// IsMaybeTestable indicates the example uses a language that COULD be testable
	// (javascript, shell) but lacks proper context to determine definitively.
	// These are grey-area examples that may need manual review.
	IsMaybeTestable bool
	// FilePath is the path to the included file (for literalinclude or io-code-block)
	FilePath string
	// SourceFile is the RST file containing this code example
	SourceFile string
}

// PageAnalysis represents the analysis results for a single page.
type PageAnalysis struct {
	Rank         int
	URL          string
	SourcePath   string
	ContentDir   string
	Error        string // Non-empty if page could not be analyzed
	CodeExamples []CodeExample
}

// ProductStats holds statistics for a single product/language.
type ProductStats struct {
	Product            string
	TotalCount         int
	InputCount         int
	OutputCount        int
	TestedCount        int
	TestableCount      int
	MaybeTestableCount int
}

// PageReport holds the complete analysis for a page with aggregated stats.
type PageReport struct {
	Rank               int
	URL                string
	SourcePath         string
	ContentDir         string
	Error              string
	TotalExamples      int
	TotalInput         int
	TotalOutput        int
	TotalTested        int
	TotalTestable      int
	TotalMaybeTestable int
	ByProduct          map[string]*ProductStats
}

// TestableProducts lists the products that have test infrastructure.
//
// WHY THIS EXISTS:
// MongoDB has automated testing infrastructure for code examples in certain driver
// documentation sets. This map identifies which products have that infrastructure,
// so we can report on how many code examples on a page COULD be tested.
//
// WHY RAW LANGUAGES ARE EXCLUDED:
// Raw language values like "javascript" and "shell" are intentionally excluded because
// many code examples use these languages without being actual Driver/Shell examples.
// For example:
//   - A "javascript" code block might be a browser snippet, not a Node.js driver example
//   - A "shell" code block might be a bash command, not a MongoDB Shell example
//
// Only properly contextualized examples are considered testable:
//   - Examples in driver content directories (e.g., content/pymongo-driver)
//   - Examples within driver tab sets (.. tabs-drivers:: with :tabid:)
//   - Examples within composable tutorials with language/interface options
//
// The map includes both human-readable names (e.g., "Python") and internal IDs
// (e.g., "python") to handle both display names and raw values from rstspec.toml.
var TestableProducts = map[string]bool{
	"C#":            true,
	"csharp":        true,
	"Go":            true,
	"go":            true,
	"Java":          true,
	"Java (Sync)":   true,
	"java":          true,
	"java-sync":     true,
	"Node.js":       true,
	"nodejs":        true,
	"Python":        true,
	"python":        true,
	"MongoDB Shell": true,
	"mongosh":       true,
}

// MaybeTestableProducts lists products that COULD be testable but lack proper context.
//
// These are "grey area" examples where the language (javascript, shell) could represent
// testable code (Node.js driver, MongoDB Shell) but could also be non-testable content
// (other JavaScript, bash commands, output examples).
//
// Examples are marked as "maybe testable" when:
//   - Language is "javascript" or "js" but not in a Node.js driver or MongoDB Shell context
//   - Language is "shell" but not in a MongoDB Shell context
//
// These examples need manual review to determine if they should be tested.
var MaybeTestableProducts = map[string]bool{
	"JavaScript": true,
	"Shell":      true,
}

// TestableDrivers lists the driver project names that have test infrastructure.
// Used to highlight which drivers have test infrastructure in --list-drivers output.
// The keys are the Snooty project names (used in URLs and internally).
var TestableDrivers = map[string]bool{
	"csharp":  true, // C# Driver
	"golang":  true, // Go Driver
	"java":    true, // Java Sync Driver
	"node":    true, // Node.js Driver
	"pymongo": true, // Python Driver
	// Note: mongodb-shell has test infrastructure but is not a driver (use --filter mongosh)
}

// ProductMappings holds the mappings from rstspec.toml for resolving
// tab IDs and composable options to human-readable product names.
//
// WHY RUNTIME LOADING FROM rstspec.toml:
// The rstspec.toml file in the snooty-parser repository is the canonical source
// of truth for all RST directive definitions, including tabs and composables.
// By loading these mappings at runtime (with caching), we ensure:
//  1. Mappings stay in sync with the actual documentation build system
//  2. New drivers/languages are automatically supported without code changes
//  3. We don't have to maintain duplicate hardcoded mappings
//
// The mappings are cached for 24 hours to avoid repeated network requests.
// See internal/rst/rstspec.go for the caching implementation.
type ProductMappings struct {
	// DriversTabIDToProduct maps driver tab IDs to product names.
	// Example: "python" → "Python", "java-sync" → "Java (Sync)"
	// Loaded from [tabs.drivers] in rstspec.toml.
	DriversTabIDToProduct map[string]string

	// ComposableLanguageToProduct maps composable language IDs to product names.
	// Example: "nodejs" → "Node.js", "csharp" → "C#"
	// Loaded from [[composables]] where id="language" in rstspec.toml.
	ComposableLanguageToProduct map[string]string

	// ComposableInterfaceToProduct maps composable interface IDs to product names.
	// Example: "mongosh" → "MongoDB Shell", "compass" → "Compass"
	// Loaded from [[composables]] where id="interface" in rstspec.toml.
	ComposableInterfaceToProduct map[string]string
}

// LoadProductMappings fetches rstspec.toml and builds the product mappings.
//
// This function fetches the canonical rstspec.toml from the snooty-parser repository
// (with 24-hour caching) and extracts the mappings for:
//   - Driver tabs: [tabs.drivers] section
//   - Language composables: [[composables]] where id="language"
//   - Interface composables: [[composables]] where id="interface"
//
// If the network is unavailable, it falls back to an expired cache if available.
func LoadProductMappings() (*ProductMappings, error) {
	rstspec, err := rst.FetchRstspec()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rstspec.toml: %w", err)
	}

	mappings := &ProductMappings{
		DriversTabIDToProduct:        rstspec.BuildTabIDToTitleMap("drivers"),
		ComposableLanguageToProduct:  rstspec.BuildComposableIDToTitleMap("language"),
		ComposableInterfaceToProduct: rstspec.BuildComposableIDToTitleMap("interface"),
	}

	return mappings, nil
}

// snootyCache caches parsed snooty.toml files by their path to avoid re-parsing.
var snootyCache = struct {
	sync.RWMutex
	configs map[string]*snooty.Config
}{configs: make(map[string]*snooty.Config)}

// MergeProjectComposables creates a copy of the base mappings and merges in
// composables from the project's snooty.toml file.
//
// This function:
//  1. Finds the project's snooty.toml by walking up from the source file path
//  2. Parses the snooty.toml (with caching to avoid re-parsing for each page)
//  3. Merges any "language" or "interface" composables into the mappings
//
// Project-specific composables take precedence over rstspec.toml definitions,
// allowing projects like Atlas to define custom composables that override defaults.
//
// Parameters:
//   - baseMappings: The base mappings loaded from rstspec.toml
//   - sourcePath: Absolute path to the source file being analyzed
//
// Returns:
//   - *ProductMappings: A new ProductMappings with project composables merged in
func MergeProjectComposables(baseMappings *ProductMappings, sourcePath string) *ProductMappings {
	// Find the project's snooty.toml
	snootyPath, err := snooty.FindProjectSnootyTOML(sourcePath)
	if err != nil || snootyPath == "" {
		// No snooty.toml found, return base mappings as-is
		return baseMappings
	}

	// Check cache first
	snootyCache.RLock()
	cachedConfig, found := snootyCache.configs[snootyPath]
	snootyCache.RUnlock()

	var config *snooty.Config
	if found {
		config = cachedConfig
	} else {
		// Parse the snooty.toml
		config, err = snooty.ParseFile(snootyPath)
		if err != nil {
			// Failed to parse, return base mappings
			return baseMappings
		}

		// Cache the parsed config
		snootyCache.Lock()
		snootyCache.configs[snootyPath] = config
		snootyCache.Unlock()
	}

	// If no composables defined, return base mappings
	if len(config.Composables) == 0 {
		return baseMappings
	}

	// Create a copy of the base mappings
	merged := &ProductMappings{
		DriversTabIDToProduct:        make(map[string]string),
		ComposableLanguageToProduct:  make(map[string]string),
		ComposableInterfaceToProduct: make(map[string]string),
	}

	// Copy base mappings
	for k, v := range baseMappings.DriversTabIDToProduct {
		merged.DriversTabIDToProduct[k] = v
	}
	for k, v := range baseMappings.ComposableLanguageToProduct {
		merged.ComposableLanguageToProduct[k] = v
	}
	for k, v := range baseMappings.ComposableInterfaceToProduct {
		merged.ComposableInterfaceToProduct[k] = v
	}

	// Merge project-specific composables (project takes precedence)
	projectLanguage := snooty.BuildComposableIDToTitleMap(config.Composables, "language")
	for k, v := range projectLanguage {
		merged.ComposableLanguageToProduct[k] = v
	}

	projectInterface := snooty.BuildComposableIDToTitleMap(config.Composables, "interface")
	for k, v := range projectInterface {
		merged.ComposableInterfaceToProduct[k] = v
	}

	return merged
}
