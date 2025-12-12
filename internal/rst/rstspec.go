// Package rst provides utilities for parsing reStructuredText files and related configuration.
package rst

import (
	"fmt"
	"io"
	"net/http"

	"github.com/BurntSushi/toml"
)

// RstspecURL is the URL to the canonical rstspec.toml file in the snooty-parser repository.
const RstspecURL = "https://raw.githubusercontent.com/mongodb/snooty-parser/refs/heads/main/snooty/rstspec.toml"

// RstspecComposable represents a composable definition from rstspec.toml.
type RstspecComposable struct {
	ID           string                       `toml:"id"`
	Title        string                       `toml:"title"`
	Default      string                       `toml:"default"`
	Dependencies []map[string]string          `toml:"dependencies"`
	Options      []RstspecComposableOption    `toml:"options"`
}

// RstspecComposableOption represents an option within a composable.
type RstspecComposableOption struct {
	ID    string `toml:"id"`
	Title string `toml:"title"`
}

// RstspecConfig represents the structure of the rstspec.toml file.
// This includes all sections, though most commands will only use specific parts.
type RstspecConfig struct {
	Composables []RstspecComposable `toml:"composables"`
	// Additional sections can be added here as needed:
	// Directives  map[string]interface{} `toml:"directive"`
	// Roles       map[string]interface{} `toml:"role"`
	// etc.
}

// FetchRstspec fetches and parses the canonical rstspec.toml file.
//
// This function downloads the rstspec.toml file from the snooty-parser repository
// and parses it into an RstspecConfig structure. This file contains canonical
// definitions for RST directives, roles, composables, and other configuration
// that may be duplicated or extended in local project files.
//
// Returns:
//   - *RstspecConfig: The parsed rstspec configuration
//   - error: Any error encountered during fetch or parse
//
// Example:
//
//	config, err := rst.FetchRstspec()
//	if err != nil {
//	    return fmt.Errorf("failed to fetch rstspec: %w", err)
//	}
//	fmt.Printf("Found %d composables\n", len(config.Composables))
func FetchRstspec() (*RstspecConfig, error) {
	// Fetch the rstspec.toml file
	resp, err := http.Get(RstspecURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rstspec.toml: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch rstspec.toml: HTTP %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read rstspec.toml: %w", err)
	}

	// Parse the TOML
	var config RstspecConfig
	if err := toml.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("failed to parse rstspec.toml: %w", err)
	}

	return &config, nil
}

