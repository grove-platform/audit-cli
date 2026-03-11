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

// DocsSetError represents an error that occurred while processing a specific docs set.
type DocsSetError struct {
	DocsSet string
	Err     error
}

// ScanResult contains the results of scanning docs sets, including any errors encountered.
type ScanResult struct {
	Entries []PageEntry
	Errors  []DocsSetError
}

// HasErrors returns true if any errors occurred during scanning.
func (r *ScanResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// PrintErrorReport prints a summary of all errors encountered during scanning.
func (r *ScanResult) PrintErrorReport() {
	if !r.HasErrors() {
		return
	}

	fmt.Fprintf(os.Stderr, "\n=== Docs Set Scanning Errors ===\n")
	fmt.Fprintf(os.Stderr, "%d docs set(s) had errors:\n\n", len(r.Errors))
	for _, e := range r.Errors {
		fmt.Fprintf(os.Stderr, "  • %s: %v\n", e.DocsSet, e.Err)
	}
	fmt.Fprintf(os.Stderr, "================================\n\n")
}

// ScanDocsSets scans one or more documentation sets (content directories) and returns
// PageEntry structs for all pages found, with URLs resolved using the resolve url logic.
//
// The function continues scanning even if individual docs sets fail, accumulating errors
// in the ScanResult. Check ScanResult.HasErrors() and use PrintErrorReport() to display
// any errors that occurred.
//
// Parameters:
//   - monorepoPath: Path to the docs monorepo root
//   - docsSets: List of content directory names to scan (e.g., "cloud-docs", "golang")
//   - versionFilter: Version filter - "" for all versions, "current" for current only, or specific version like "v8.0"
//   - baseURL: Base URL for production documentation (e.g., https://www.mongodb.com/docs)
//
// Returns:
//   - *ScanResult: Contains entries from successful scans and any errors encountered
//   - error: Only returned for fatal errors (e.g., no docs sets specified, content dir missing)
func ScanDocsSets(monorepoPath string, docsSets []string, versionFilter string, baseURL string) (*ScanResult, error) {
	if len(docsSets) == 0 {
		return nil, fmt.Errorf("at least one docs set must be specified")
	}

	contentDir := filepath.Join(monorepoPath, "content")
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("content directory not found: %s", contentDir)
	}

	result := &ScanResult{
		Entries: []PageEntry{},
		Errors:  []DocsSetError{},
	}
	rank := 1

	for _, docsSet := range docsSets {
		docsSetPath := filepath.Join(contentDir, docsSet)
		if _, err := os.Stat(docsSetPath); os.IsNotExist(err) {
			result.Errors = append(result.Errors, DocsSetError{
				DocsSet: docsSet,
				Err:     fmt.Errorf("couldn't find docs set at path %q", docsSetPath),
			})
			continue
		}

		entries, err := scanSingleDocsSet(monorepoPath, docsSet, versionFilter, baseURL, &rank)
		if err != nil {
			result.Errors = append(result.Errors, DocsSetError{
				DocsSet: docsSet,
				Err:     err,
			})
			continue
		}
		result.Entries = append(result.Entries, entries...)
	}

	return result, nil
}

// scanSingleDocsSet scans a single docs set and returns PageEntry structs.
// The rank parameter is a pointer so it can be incremented across multiple doc sets.
// versionFilter can be: "" (all versions), "current" (only current), or a specific version like "v8.0"
func scanSingleDocsSet(monorepoPath, docsSet string, versionFilter string, baseURL string, rank *int) ([]PageEntry, error) {
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

		// Apply version filter if specified
		if versionFilter != "" {
			version := extractVersionFromFilePath(relPath)
			if versionFilter == "current" {
				// Filter to only current version
				if version != "" && !projectinfo.IsCurrentVersion(version) {
					return nil
				}
			} else {
				// Filter to specific version
				if version != versionFilter {
					return nil
				}
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
