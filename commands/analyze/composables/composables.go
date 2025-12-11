// Package composables provides functionality for analyzing composables in snooty.toml files.
//
// This package implements the "analyze composables" subcommand, which scans the MongoDB
// documentation monorepo for snooty.toml files and analyzes the composables defined in them.
// It helps identify opportunities for consolidation by finding identical or similar composables
// across projects and versions.
package composables

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewComposablesCommand creates the composables subcommand for analysis.
//
// This command analyzes composables defined in snooty.toml files across the MongoDB
// documentation monorepo. It identifies:
//   - All composables and their locations
//   - Identical composables that could be consolidated
//   - Similar composables that should be reviewed
//
// Usage:
//
//	analyze composables /path/to/docs-monorepo
//	analyze composables /path/to/docs-monorepo --for-project manual
//	analyze composables /path/to/docs-monorepo --current-only
//
// Flags:
//   - --for-project: Only analyze composables for a specific project
//   - --current-only: Only analyze composables in current versions
//   - --verbose: Show full option details with titles
//   - --find-consolidation-candidates: Show identical and similar composables for consolidation
//   - --find-usages: Show where each composable is used in RST files
func NewComposablesCommand() *cobra.Command {
	var (
		forProject                  string
		currentOnly                 bool
		verbose                     bool
		findConsolidationCandidates bool
		findUsages                  bool
	)

	cmd := &cobra.Command{
		Use:   "composables [monorepo-path]",
		Short: "Analyze composables in snooty.toml files",
		Long: `Analyze composables defined in snooty.toml files across the MongoDB documentation monorepo.

This command scans all snooty.toml files in the monorepo and analyzes the composables
defined in them.

Composables are configuration elements in snooty.toml that define content variations
for different contexts (e.g., different programming languages, deployment types, or interfaces).

By default, the output includes:
  - A summary of all composables grouped by ID
  - A detailed table of all composables found

With --find-consolidation-candidates, the output also includes:
  - Identical composables (same ID, title, and options) across different projects/versions
  - Similar composables (different IDs but similar option sets) that may be consolidation candidates

With --find-usages, the output also includes:
  - Usage count for each composable
  - File paths where each composable is used in composable-tutorial directives

Examples:
  # Analyze all composables in the monorepo
  analyze composables /path/to/docs-monorepo

  # Analyze composables for a specific project
  analyze composables /path/to/docs-monorepo --for-project manual

  # Analyze only current versions
  analyze composables /path/to/docs-monorepo --current-only

  # Show full option details
  analyze composables /path/to/docs-monorepo --verbose

  # Find consolidation candidates
  analyze composables /path/to/docs-monorepo --find-consolidation-candidates

  # Find where composables are used
  analyze composables /path/to/docs-monorepo --find-usages

  # Combine flags
  analyze composables /path/to/docs-monorepo --for-project atlas --find-consolidation-candidates --find-usages --verbose`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runComposables(args[0], forProject, currentOnly, verbose, findConsolidationCandidates, findUsages)
		},
	}

	cmd.Flags().StringVar(&forProject, "for-project", "", "Only analyze composables for a specific project")
	cmd.Flags().BoolVar(&currentOnly, "current-only", false, "Only analyze composables in current versions")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full option details with titles")
	cmd.Flags().BoolVar(&findConsolidationCandidates, "find-consolidation-candidates", false, "Show identical and similar composables for consolidation")
	cmd.Flags().BoolVar(&findUsages, "find-usages", false, "Show where each composable is used in RST files")

	return cmd
}

// runComposables executes the composables analysis operation.
func runComposables(monorepoPath string, forProject string, currentOnly bool, verbose bool, findConsolidationCandidates bool, findUsages bool) error {
	// Find all snooty.toml files and extract composables
	locations, err := FindSnootyTOMLFiles(monorepoPath, forProject, currentOnly)
	if err != nil {
		return fmt.Errorf("failed to find snooty.toml files: %w", err)
	}

	if len(locations) == 0 {
		fmt.Println("No composables found in the monorepo.")
		return nil
	}

	// Analyze the composables
	result := AnalyzeComposables(locations)

	// Find usages if requested
	var usages map[string]*ComposableUsage
	if findUsages {
		usages, err = FindComposableUsages(monorepoPath, result.AllComposables, forProject, currentOnly)
		if err != nil {
			return fmt.Errorf("failed to find composable usages: %w", err)
		}
	}

	// Print the results
	PrintResults(result, verbose, findConsolidationCandidates, findUsages, usages)

	return nil
}

