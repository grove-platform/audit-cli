// Package snooty provides utilities for parsing snooty.toml configuration files.
//
// This package provides:
//   - Types for representing snooty.toml structure (composables, options)
//   - Functions for parsing snooty.toml files
//   - Functions for finding snooty.toml files in the monorepo
//   - Functions for finding a project's snooty.toml from a source file path
package snooty

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/grove-platform/audit-cli/internal/projectinfo"
)

// Composable represents a composable definition from a snooty.toml file.
type Composable struct {
	ID           string              `toml:"id"`
	Title        string              `toml:"title"`
	Default      string              `toml:"default"`
	Dependencies []map[string]string `toml:"dependencies"`
	Options      []ComposableOption  `toml:"options"`
}

// ComposableOption represents an option within a composable.
type ComposableOption struct {
	ID    string `toml:"id"`
	Title string `toml:"title"`
}

// Config represents the structure of a snooty.toml file.
type Config struct {
	Name        string       `toml:"name"`
	Title       string       `toml:"title"`
	Composables []Composable `toml:"composables"`
}

// ParseFile parses a snooty.toml file and returns its configuration.
//
// Parameters:
//   - filePath: Path to the snooty.toml file
//
// Returns:
//   - *Config: The parsed configuration
//   - error: Any error encountered during parsing
func ParseFile(filePath string) (*Config, error) {
	var config Config
	_, err := toml.DecodeFile(filePath, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snooty.toml: %w", err)
	}
	return &config, nil
}

// FindProjectSnootyTOML finds the snooty.toml file for a project based on a source file path.
//
// Given a source file path like:
//   - /path/to/monorepo/content/atlas/source/foo.rst
//   - /path/to/monorepo/content/manual/v8.0/source/bar.rst
//
// This function walks up the directory tree to find the snooty.toml file at:
//   - /path/to/monorepo/content/atlas/snooty.toml
//   - /path/to/monorepo/content/manual/v8.0/snooty.toml
//
// Parameters:
//   - sourcePath: Absolute path to a source file within a project
//
// Returns:
//   - string: Path to the snooty.toml file, or empty string if not found
//   - error: Any error encountered during search
func FindProjectSnootyTOML(sourcePath string) (string, error) {
	// Get absolute path
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Start from the directory containing the source file
	dir := filepath.Dir(absPath)

	// Walk up the directory tree looking for snooty.toml
	// Stop when we reach the content directory or filesystem root
	for {
		// Check if snooty.toml exists in this directory
		snootyPath := filepath.Join(dir, "snooty.toml")
		if _, err := os.Stat(snootyPath); err == nil {
			return snootyPath, nil
		}

		// Check if we've reached the content directory (stop here)
		if filepath.Base(dir) == "content" {
			break
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return "", nil // Not found, but not an error
}

// BuildComposableIDToTitleMap builds a map from composable option IDs to titles
// for a specific composable type (e.g., "language", "interface").
//
// Parameters:
//   - composables: Slice of composables from a snooty.toml file
//   - composableID: The ID of the composable to extract (e.g., "language", "interface")
//
// Returns:
//   - map[string]string: Map from option ID to option title
func BuildComposableIDToTitleMap(composables []Composable, composableID string) map[string]string {
	result := make(map[string]string)
	for _, comp := range composables {
		if comp.ID == composableID {
			for _, opt := range comp.Options {
				result[opt.ID] = opt.Title
			}
		}
	}
	return result
}

// IsCurrentVersion checks if a version string represents a "current" version.
// This is a convenience wrapper around projectinfo.IsCurrentVersion.
func IsCurrentVersion(version string) bool {
	return projectinfo.IsCurrentVersion(version)
}

// ExtractProjectAndVersion extracts project and version from a relative path.
// Returns (project, version) where version is empty for non-versioned projects.
//
// Examples:
//   - "manual/v8.0/snooty.toml" -> ("manual", "v8.0")
//   - "atlas/snooty.toml" -> ("atlas", "")
func ExtractProjectAndVersion(relPath string) (string, string) {
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 2 {
		return "", ""
	}

	projectName := parts[0]

	// Check if this is a versioned project
	// Pattern: project/version/snooty.toml (3 parts)
	// Pattern: project/snooty.toml (2 parts)
	if len(parts) == 3 && parts[2] == "snooty.toml" {
		return projectName, parts[1]
	} else if len(parts) == 2 && parts[1] == "snooty.toml" {
		return projectName, ""
	}

	return "", ""
}

