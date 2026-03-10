// Package testablecode provides the testable-code subcommand for the report command.
// This file handles scanning documentation sets to find all pages.
package testablecode

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	resolveurl "github.com/grove-platform/audit-cli/commands/resolve/url"
	"github.com/grove-platform/audit-cli/internal/projectinfo"
)

// ScanDocsSets scans one or more documentation sets (content directories) and returns
// PageEntry structs for all pages found, with URLs resolved using the resolve url logic.
//
// Parameters:
//   - monorepoPath: Path to the docs monorepo root
//   - docsSets: List of content directory names to scan (e.g., "cloud-docs", "golang")
//   - currentOnly: If true, only scan the current version for versioned projects
//   - baseURL: Base URL for production documentation (e.g., https://www.mongodb.com/docs)
//
// Returns:
//   - []PageEntry: List of page entries with rank (page number within doc set) and URL
//   - error: Any error encountered during scanning
func ScanDocsSets(monorepoPath string, docsSets []string, currentOnly bool, baseURL string) ([]PageEntry, error) {
	if len(docsSets) == 0 {
		return nil, fmt.Errorf("at least one docs set must be specified")
	}

	contentDir := filepath.Join(monorepoPath, "content")
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("content directory not found: %s", contentDir)
	}

	var allEntries []PageEntry
	rank := 1

	for _, docsSet := range docsSets {
		docsSetPath := filepath.Join(contentDir, docsSet)
		if _, err := os.Stat(docsSetPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("docs set not found: %s", docsSet)
		}

		entries, err := scanSingleDocsSet(monorepoPath, docsSet, currentOnly, baseURL, &rank)
		if err != nil {
			return nil, fmt.Errorf("failed to scan docs set %s: %w", docsSet, err)
		}
		allEntries = append(allEntries, entries...)
	}

	return allEntries, nil
}

// scanSingleDocsSet scans a single docs set and returns PageEntry structs.
// The rank parameter is a pointer so it can be incremented across multiple doc sets.
func scanSingleDocsSet(monorepoPath, docsSet string, currentOnly bool, baseURL string, rank *int) ([]PageEntry, error) {
	contentDir := filepath.Join(monorepoPath, "content")
	docsSetPath := filepath.Join(contentDir, docsSet)

	// Collect all .txt files
	var txtFiles []string

	err := filepath.Walk(docsSetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories we shouldn't scan
		if info.IsDir() {
			dirName := info.Name()
			// Skip code-examples directories
			if dirName == "code-examples" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .txt files
		if filepath.Ext(path) != ".txt" {
			return nil
		}

		// Check if this is in a source directory
		relPath, _ := filepath.Rel(docsSetPath, path)
		if !strings.Contains(relPath, "source"+string(filepath.Separator)) {
			return nil
		}

		// If currentOnly, filter to only current version
		if currentOnly {
			version := extractVersionFromFilePath(relPath)
			if version != "" && !projectinfo.IsCurrentVersion(version) {
				return nil
			}
		}

		txtFiles = append(txtFiles, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort files for deterministic output
	sort.Strings(txtFiles)

	// Convert each file to a PageEntry with resolved URL
	var entries []PageEntry
	for _, filePath := range txtFiles {
		url, err := resolveurl.ResolveFileToURL(filePath, baseURL)
		if err != nil {
			// Log warning but continue with other files
			fmt.Fprintf(os.Stderr, "Warning: Could not resolve URL for %s: %v\n", filePath, err)
			continue
		}

		entries = append(entries, PageEntry{
			Rank:    *rank,
			URL:     url,
			DocsSet: docsSet,
		})
		*rank++
	}

	return entries, nil
}

// extractVersionFromFilePath extracts the version from a relative file path.
// For versioned projects: version/source/file.txt -> "version"
// For non-versioned projects: source/file.txt -> ""
func extractVersionFromFilePath(relPath string) string {
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 2 {
		return ""
	}

	// If first part is "source", this is non-versioned
	if parts[0] == "source" {
		return ""
	}

	// Check if first part looks like a version
	if projectinfo.IsVersionDirectory(parts[0]) {
		return parts[0]
	}

	return ""
}

