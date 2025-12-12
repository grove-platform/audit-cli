// Package composables provides functionality for analyzing composables in snooty.toml files.
package composables

import (
	"fmt"

	"github.com/grove-platform/audit-cli/internal/rst"
)

// FetchRstspecComposables fetches and parses composables from the canonical rstspec.toml file.
//
// This function downloads the rstspec.toml file from the snooty-parser repository
// and extracts the composables defined in it. These are the "canonical" or universal
// composables that may be duplicated in local snooty.toml files.
//
// Returns:
//   - A slice of ComposableLocation objects with Source set to "rstspec.toml"
//   - An error if the fetch or parse fails
func FetchRstspecComposables() ([]ComposableLocation, error) {
	// Fetch the rstspec.toml file using the internal/rst package
	config, err := rst.FetchRstspec()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rstspec.toml: %w", err)
	}

	// Convert rstspec composables to ComposableLocation objects
	locations := make([]ComposableLocation, 0, len(config.Composables))
	for _, rstspecComp := range config.Composables {
		// Convert RstspecComposable to Composable
		composable := Composable{
			ID:      rstspecComp.ID,
			Title:   rstspecComp.Title,
			Default: rstspecComp.Default,
			Options: make([]ComposableOption, 0, len(rstspecComp.Options)),
		}

		// Convert options
		for _, rstspecOpt := range rstspecComp.Options {
			composable.Options = append(composable.Options, ComposableOption{
				ID:    rstspecOpt.ID,
				Title: rstspecOpt.Title,
			})
		}

		locations = append(locations, ComposableLocation{
			Project:    "rstspec",
			Version:    "",
			Composable: composable,
			FilePath:   rst.RstspecURL,
			Source:     "rstspec.toml",
		})
	}

	return locations, nil
}

