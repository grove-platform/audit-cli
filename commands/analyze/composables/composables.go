// Package composables provides functionality for analyzing composables in snooty.toml files.
//
// This package implements the "analyze composables" subcommand, which scans the MongoDB
// documentation monorepo for snooty.toml files and analyzes the composables defined in them.
// It helps identify opportunities for consolidation by finding identical or similar composables
// across projects and versions.
package composables

import (
	"fmt"

	"github.com/grove-platform/audit-cli/internal/config"
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
//   - --with-rstspec: Include composables from the canonical rstspec.toml file
func NewComposablesCommand() *cobra.Command {
	var (
		forProject                  string
		currentOnly                 bool
		verbose                     bool
		findConsolidationCandidates bool
		findUsages                  bool
		withRstspec                 bool
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

With --with-rstspec, the analysis also includes:
  - Composables from the canonical rstspec.toml file in the snooty-parser repository
  - Helps identify duplication between local snooty.toml files and the canonical definitions

Monorepo Path Configuration:
  The monorepo path can be specified in three ways (in order of priority):
    1. Command-line argument: analyze composables /path/to/monorepo
    2. Environment variable: export AUDIT_CLI_MONOREPO_PATH=/path/to/monorepo
    3. Config file (.audit-cli.yaml):
       monorepo_path: /path/to/monorepo

Examples:
  # Analyze all composables in the monorepo
  analyze composables /path/to/docs-monorepo

  # Use configured monorepo path
  analyze composables

  # Analyze composables for a specific project
  analyze composables --for-project manual

  # Analyze only current versions
  analyze composables --current-only

  # Show full option details
  analyze composables --verbose

  # Find consolidation candidates
  analyze composables --find-consolidation-candidates

  # Find where composables are used
  analyze composables --find-usages

  # Include canonical rstspec.toml composables
  analyze composables --with-rstspec --find-consolidation-candidates

  # Combine flags
  analyze composables --for-project atlas --find-consolidation-candidates --find-usages --verbose`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve monorepo path from args, env, or config
			var cmdLineArg string
			if len(args) > 0 {
				cmdLineArg = args[0]
			}
			monorepoPath, err := config.GetMonorepoPath(cmdLineArg)
			if err != nil {
				return err
			}
			return runComposables(monorepoPath, forProject, currentOnly, verbose, findConsolidationCandidates, findUsages, withRstspec)
		},
	}

	cmd.Flags().StringVar(&forProject, "for-project", "", "Only analyze composables for a specific project")
	cmd.Flags().BoolVar(&currentOnly, "current-only", false, "Only analyze composables in current versions")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full option details with titles")
	cmd.Flags().BoolVar(&findConsolidationCandidates, "find-consolidation-candidates", false, "Show identical and similar composables for consolidation")
	cmd.Flags().BoolVar(&findUsages, "find-usages", false, "Show where each composable is used in RST files")
	cmd.Flags().BoolVar(&withRstspec, "with-rstspec", false, "Include composables from the canonical rstspec.toml file")

	return cmd
}

// runComposables executes the composables analysis operation.
func runComposables(monorepoPath string, forProject string, currentOnly bool, verbose bool, findConsolidationCandidates bool, findUsages bool, withRstspec bool) error {
	// Find all snooty.toml files and extract composables
	locations, err := FindSnootyTOMLFiles(monorepoPath, forProject, currentOnly)
	if err != nil {
		return fmt.Errorf("failed to find snooty.toml files: %w", err)
	}

	// Fetch rstspec.toml composables if requested
	if withRstspec {
		fmt.Println("Fetching composables from rstspec.toml...")
		rstspecLocations, err := FetchRstspecComposables()
		if err != nil {
			return fmt.Errorf("failed to fetch rstspec.toml composables: %w", err)
		}
		fmt.Printf("Found %d composables in rstspec.toml\n", len(rstspecLocations))
		locations = append(locations, rstspecLocations...)
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

