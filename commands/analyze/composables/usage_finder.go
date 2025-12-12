// Package composables provides functionality for analyzing composables in snooty.toml files.
package composables

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/grove-platform/audit-cli/internal/projectinfo"
)

// composableTutorialRegex matches .. composable-tutorial:: directives
var composableTutorialRegex = regexp.MustCompile(`^\.\.\s+composable-tutorial::`)

// optionsRegex matches :options: lines in composable-tutorial directives
var optionsRegex = regexp.MustCompile(`^\s*:options:\s+(.+)$`)

// FindComposableUsages finds all usages of composables in RST files.
// It scans all .txt and .rst files in the monorepo and looks for composable-tutorial directives.
func FindComposableUsages(monorepoPath string, composables []ComposableLocation, forProject string, currentOnly bool) (map[string]*ComposableUsage, error) {
	// Create a map to track usages by composable ID + project + version
	usageMap := make(map[string]*ComposableUsage)

	// Walk through the content directory
	contentDir := filepath.Join(monorepoPath, "content")
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .txt and .rst files
		ext := filepath.Ext(path)
		if ext != ".txt" && ext != ".rst" {
			return nil
		}

		// Extract project and version from path
		project, version := extractProjectAndVersionFromPath(path, monorepoPath)
		if project == "" {
			return nil
		}

		// Apply filters
		if forProject != "" && project != forProject {
			return nil
		}

		if currentOnly && !projectinfo.IsCurrentVersion(version) {
			return nil
		}

		// Parse the file for composable-tutorial directives
		composableIDs, err := extractComposableIDsFromFile(path)
		if err != nil {
			// Skip files that can't be read
			return nil
		}

		// Track usages
		for _, composableID := range composableIDs {
			key := fmt.Sprintf("%s::%s::%s", project, version, composableID)
			if usage, exists := usageMap[key]; exists {
				usage.UsageCount++
				usage.FilePaths = append(usage.FilePaths, getRelativePath(path, monorepoPath))
			} else {
				usageMap[key] = &ComposableUsage{
					ComposableID: composableID,
					Project:      project,
					Version:      version,
					UsageCount:   1,
					FilePaths:    []string{getRelativePath(path, monorepoPath)},
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return usageMap, nil
}

// extractProjectAndVersionFromPath extracts project and version from a file path.
// Example: /path/to/content/atlas/source/file.txt -> project: atlas, version: ""
// Example: /path/to/content/manual/v7.0/source/file.txt -> project: manual, version: v7.0
func extractProjectAndVersionFromPath(filePath string, monorepoPath string) (string, string) {
	// Get relative path from monorepo root
	relPath, err := filepath.Rel(monorepoPath, filePath)
	if err != nil {
		return "", ""
	}

	// Split path into parts
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 3 || parts[0] != "content" {
		return "", ""
	}

	project := parts[1]

	// Check if there's a version directory
	if len(parts) >= 4 && parts[2] != "source" {
		// Versioned project: content/{project}/{version}/source/...
		return project, parts[2]
	}

	// Non-versioned project: content/{project}/source/...
	return project, ""
}

// extractComposableIDsFromFile parses an RST file and extracts composable IDs from composable-tutorial directives.
func extractComposableIDsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var composableIDs []string
	scanner := bufio.NewScanner(file)
	inComposableTutorial := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check for composable-tutorial directive
		if composableTutorialRegex.MatchString(trimmedLine) {
			inComposableTutorial = true
			continue
		}

		// If we're in a composable-tutorial, look for :options: line
		if inComposableTutorial {
			if matches := optionsRegex.FindStringSubmatch(line); len(matches) > 1 {
				// Parse the comma-separated list of composable IDs
				optionsStr := strings.TrimSpace(matches[1])
				ids := strings.Split(optionsStr, ",")
				for _, id := range ids {
					composableIDs = append(composableIDs, strings.TrimSpace(id))
				}
				inComposableTutorial = false
			}
		}
	}

	return composableIDs, scanner.Err()
}

// getRelativePath returns the path relative to the monorepo root.
func getRelativePath(path string, monorepoPath string) string {
	relPath, err := filepath.Rel(monorepoPath, path)
	if err != nil {
		return path
	}
	return relPath
}

