// Package composables provides functionality for analyzing composables in snooty.toml files.
package composables

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/grove-platform/audit-cli/internal/snooty"
)

// ParseSnootyTOML parses a snooty.toml file and extracts composables.
//
// Parameters:
//   - filePath: Path to the snooty.toml file
//
// Returns:
//   - []snooty.Composable: Slice of composables found in the file
//   - error: Any error encountered during parsing
func ParseSnootyTOML(filePath string) ([]snooty.Composable, error) {
	config, err := snooty.ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	return config.Composables, nil
}

// FindSnootyTOMLFiles finds all snooty.toml files in the monorepo.
//
// Parameters:
//   - monorepoPath: Path to the MongoDB documentation monorepo
//   - forProject: If non-empty, only find files for this project
//   - currentOnly: If true, only find files in current versions
//
// Returns:
//   - []ComposableLocation: Slice of all composables found with their locations
//   - error: Any error encountered during discovery
func FindSnootyTOMLFiles(monorepoPath string, forProject string, currentOnly bool) ([]ComposableLocation, error) {
	// Get absolute path
	absPath, err := filepath.Abs(monorepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Find the content directory
	contentDir, err := findContentDirectory(absPath)
	if err != nil {
		return nil, err
	}

	var locations []ComposableLocation

	// Walk through the content directory
	err = filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process snooty.toml files
		if info.Name() != "snooty.toml" {
			return nil
		}

		// Extract project and version information
		relPath, err := filepath.Rel(contentDir, path)
		if err != nil {
			return err
		}

		projectName, versionName := snooty.ExtractProjectAndVersion(relPath)
		if projectName == "" {
			return nil
		}

		// Filter by project if specified
		if forProject != "" && projectName != forProject {
			return nil
		}

		// Filter by current version if specified
		if currentOnly && versionName != "" {
			if !snooty.IsCurrentVersion(versionName) {
				return nil
			}
		}

		// Parse the snooty.toml file
		composables, err := ParseSnootyTOML(path)
		if err != nil {
			// Skip files that can't be parsed
			return nil
		}

		// Add each composable to the locations
		for _, comp := range composables {
			locations = append(locations, ComposableLocation{
				Project:    projectName,
				Version:    versionName,
				Composable: comp,
				FilePath:   path,
				Source:     "snooty.toml",
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk content directory: %w", err)
	}

	return locations, nil
}

// findContentDirectory finds the content directory from the given path.
func findContentDirectory(dirPath string) (string, error) {
	// Check if this is already a content directory
	if filepath.Base(dirPath) == "content" {
		return dirPath, nil
	}

	// Check if there's a content subdirectory
	contentDir := filepath.Join(dirPath, "content")
	if _, err := os.Stat(contentDir); err == nil {
		return contentDir, nil
	}

	return "", fmt.Errorf("content directory not found in: %s", dirPath)
}
