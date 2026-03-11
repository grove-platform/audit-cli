// Package url provides URL resolution functionality.
package url

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/grove-platform/audit-cli/internal/config"
)

// ResolveFileToURL resolves a source file path to a production URL.
//
// The function identifies the project and version from the file path and
// constructs the appropriate production URL.
//
// Parameters:
//   - filePath: Absolute path to the source .txt file
//   - baseURL: Base production URL (e.g., https://www.mongodb.com/docs)
//
// Returns:
//   - string: The production URL for the file
//   - error: Error if resolution fails
func ResolveFileToURL(filePath string, baseURL string) (string, error) {
	// Validate file extension
	if !strings.HasSuffix(filePath, ".txt") {
		return "", fmt.Errorf("file must have .txt extension: %s", filePath)
	}

	// Parse the file path to extract project, version, and page path
	info, err := parseFilePath(filePath)
	if err != nil {
		return "", err
	}

	// Get the URL slug for this project
	urlSlug, err := getURLSlugForProject(info.projectName)
	if err != nil {
		return "", err
	}

	// Build the URL
	return buildURL(baseURL, urlSlug, info.version, info.pagePath), nil
}

// filePathInfo holds parsed information from a source file path.
type filePathInfo struct {
	projectName string // Snooty project name (from snooty.toml)
	version     string // Version slug (e.g., "manual", "v8.0", "current")
	pagePath    string // Page path relative to source directory
	contentDir  string // Content directory name
}

// parseFilePath extracts project, version, and page path from a source file path.
//
// Expected path patterns:
//   - content/{project}/source/{page}.txt              (non-versioned)
//   - content/{project}/{version}/source/{page}.txt   (versioned)
func parseFilePath(filePath string) (*filePathInfo, error) {
	// Normalize path separators
	filePath = filepath.ToSlash(filePath)

	// Find the content directory in the path
	contentIdx := strings.Index(filePath, "/content/")
	if contentIdx == -1 {
		return nil, fmt.Errorf("file path must contain /content/ directory: %s", filePath)
	}

	// Get the path starting from content/
	relativePath := filePath[contentIdx+9:] // Skip "/content/"

	// Split into parts
	parts := strings.Split(relativePath, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid path structure: %s", filePath)
	}

	info := &filePathInfo{
		contentDir: parts[0],
	}

	// Find "source" directory to determine structure
	sourceIdx := -1
	for i, part := range parts {
		if part == "source" {
			sourceIdx = i
			break
		}
	}

	if sourceIdx == -1 {
		return nil, fmt.Errorf("file path must contain /source/ directory: %s", filePath)
	}

	// Determine if versioned or non-versioned
	switch sourceIdx {
	case 1:
		// Non-versioned: content/{project}/source/{page}.txt
		info.version = ""
	case 2:
		// Versioned: content/{project}/{version}/source/{page}.txt
		info.version = parts[1]
	default:
		return nil, fmt.Errorf("unexpected path structure: %s", filePath)
	}

	// Get the page path (everything after source/)
	pagePathParts := parts[sourceIdx+1:]
	pagePath := strings.Join(pagePathParts, "/")

	// Remove .txt extension
	pagePath = strings.TrimSuffix(pagePath, ".txt")

	// Handle index.txt -> empty page path (will render as trailing slash)
	if pagePath == "index" {
		pagePath = ""
	}

	info.pagePath = pagePath

	// Get project name from snooty.toml
	monorepoPath := extractMonorepoPath(filePath, contentIdx)
	projectName, err := getProjectName(monorepoPath, info.contentDir, info.version)
	if err != nil {
		return nil, err
	}
	info.projectName = projectName

	return info, nil
}

// extractMonorepoPath extracts the monorepo root path from a file path.
func extractMonorepoPath(filePath string, contentIdx int) string {
	return filePath[:contentIdx]
}

// getProjectName reads the snooty.toml to get the project name.
func getProjectName(monorepoPath, contentDir, _ string) (string, error) {
	urlMapping, err := config.GetURLMapping(monorepoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get URL mapping: %w", err)
	}

	// Reverse lookup: find project name for this content directory
	for project, dir := range urlMapping.ProjectToContentDir {
		if dir == contentDir {
			return project, nil
		}
	}

	return "", fmt.Errorf("could not find project name for content directory: %s", contentDir)
}

// projectToURLSlug maps Snooty project names (contentSite) to their URL base slugs.
// This mapping is derived from the table-of-contents data in the docs-mongodb-internal
// monorepo, which is the source of truth for production URLs.
//
// The URL slug is the path component after /docs/ in the production URL.
// For example, "golang" maps to "drivers/go", so the URL would be:
// https://www.mongodb.com/docs/drivers/go/current/
var projectToURLSlug = map[string]string{
	// Atlas products
	"atlas-architecture": "atlas/architecture",
	"atlas-cli":          "atlas/cli",
	"atlas-operator":     "atlas/operator",
	"cloud-docs":         "atlas",
	"cloudgov":           "atlas/government",

	// Atlas App Services (deprecated but still in monorepo)
	"atlas-app-services": "atlas/app-services",
	"realm":              "atlas/device-sdks",

	// MongoDB Server
	"docs": "", // MongoDB Manual uses empty slug (e.g., /docs/manual/)

	// Drivers and Languages
	"c":             "languages/c/c-driver",
	"cpp-driver":    "languages/cpp/cpp-driver/read",
	"csharp":        "drivers/csharp",
	"django":        "languages/python/django-mongodb",
	"golang":        "drivers/go",
	"hibernate":     "languages/java/mongodb-hibernate",
	"java":          "drivers/java/sync",
	"java-rs":       "languages/java/reactive-streams-driver",
	"kotlin":        "drivers/kotlin/coroutine",
	"kotlin-sync":   "languages/kotlin/kotlin-sync-driver",
	"laravel":       "drivers/php/laravel-mongodb",
	"node":          "drivers/node",
	"php-library":   "php-library",
	"pymongo":       "languages/python/pymongo-driver",
	"pymongo-arrow": "languages/python/pymongo-arrow-driver",
	"ruby-driver":   "ruby-driver",
	"rust":          "drivers/rust",
	"scala":         "languages/scala/scala-driver",

	// Tools and utilities
	"bi-connector":             "bi-connector",
	"charts":                   "charts",
	"cloud-manager":            "cloud-manager",
	"compass":                  "compass",
	"database-tools":           "database-tools",
	"docs-k8s-operator":        "kubernetes-operator",
	"docs-relational-migrator": "relational-migrator",
	"drivers":                  "drivers",
	"entity-framework":         "entity-framework",
	"intellij":                 "mongodb-intellij",
	"kafka-connector":          "kafka-connector",
	"landing":                  "management",
	"mck":                      "kubernetes",
	"mcp-server":               "mcp-server",
	"meta":                     "meta",
	"mongocli":                 "mongocli",
	"mongodb-shell":            "mongodb-shell",
	"mongodb-vscode":           "mongodb-vscode",
	"mongoid":                  "mongoid",
	"mongosync":                "mongosync",
	"ops-manager":              "ops-manager",
	"spark-connector":          "spark-connector",
	"sql-interface":            "sql-interface",
	"visual-studio-extension":  "mongodb-analyzer",
	"voyageai":                 "voyageai",
}

// getURLSlugForProject returns the base URL slug for a given project name.
// The mapping is derived from the table-of-contents data which is the source
// of truth for production URLs.
func getURLSlugForProject(projectName string) (string, error) {
	if slug, ok := projectToURLSlug[projectName]; ok {
		return slug, nil
	}

	return "", fmt.Errorf("could not find URL slug for project: %s", projectName)
}

// buildURL constructs the production URL from components.
//
// Parameters:
//   - baseURL: Base URL (e.g., https://www.mongodb.com/docs)
//   - urlSlug: Project URL slug (e.g., "atlas", "drivers/go") or empty for MongoDB Manual
//   - version: Version slug (e.g., "current", "v8.0", "manual") or empty for non-versioned
//   - pagePath: Page path (e.g., "tutorial/install") or empty for index
//
// Returns the full production URL with trailing slash.
func buildURL(baseURL, urlSlug, version, pagePath string) string {
	// Ensure base URL doesn't have trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Build path components (only include non-empty parts)
	var pathParts []string

	if urlSlug != "" {
		pathParts = append(pathParts, urlSlug)
	}

	// Add version if present (for versioned projects)
	if version != "" {
		pathParts = append(pathParts, version)
	}

	// Add page path if present
	if pagePath != "" {
		pathParts = append(pathParts, pagePath)
	}

	// Join with slashes and add trailing slash
	return baseURL + "/" + strings.Join(pathParts, "/") + "/"
}
