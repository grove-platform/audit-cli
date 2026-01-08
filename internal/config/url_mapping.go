// Package config provides configuration management for audit-cli.
// This file handles URL-to-source-file mapping for MongoDB documentation.

package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// SnootyDataAPIURL is the endpoint for fetching project metadata.
const SnootyDataAPIURL = "https://snooty-data-api.mongodb.com/prod/projects"

// CacheTTL is the time-to-live for the cached URL mapping (24 hours).
const CacheTTL = 24 * time.Hour

// CacheDir is the directory for storing cache files.
const CacheDir = ".audit-cli"

// CacheFileName is the name of the URL mapping cache file.
const CacheFileName = "url-mapping-cache.json"

// URLMappingCache represents the cached URL mapping data.
type URLMappingCache struct {
	Timestamp   time.Time           `json:"timestamp"`
	Mapping     map[string]string   `json:"mapping"`      // URL slug -> snooty project name
	Branches    map[string][]string `json:"branches"`     // project name -> list of version slugs
	DriverSlugs []string            `json:"driver_slugs"` // URL slugs for driver documentation
}

// SnootyAPIResponse represents the response from the Snooty Data API.
type SnootyAPIResponse struct {
	Data []SnootyProject `json:"data"`
}

// SnootyProject represents a project in the Snooty Data API response.
type SnootyProject struct {
	Project     string         `json:"project"`
	DisplayName string         `json:"displayName"`
	RepoName    string         `json:"repoName"`
	Branches    []SnootyBranch `json:"branches"`
}

// SnootyBranch represents a branch in a Snooty project.
type SnootyBranch struct {
	GitBranchName  string `json:"gitBranchName"`
	Label          string `json:"label"`
	Active         any    `json:"active"` // Can be bool or string "true"
	FullURL        string `json:"fullUrl"`
	IsStableBranch bool   `json:"isStableBranch"`
}

// SnootyToml represents the relevant fields from a snooty.toml file.
type SnootyToml struct {
	Name string `toml:"name"`
}

// URLMapping provides URL-to-source-file resolution.
type URLMapping struct {
	// URLSlugToProject maps URL slugs to snooty project names
	URLSlugToProject map[string]string
	// ProjectToContentDir maps snooty project names to content directories
	ProjectToContentDir map[string]string
	// ProjectBranches maps project names to available version slugs
	ProjectBranches map[string][]string
	// DriverSlugs contains URL slugs for driver documentation (excludes mongodb-shell)
	DriverSlugs []string
	// MonorepoPath is the path to the docs monorepo
	MonorepoPath string
}

// getCachePath returns the path to the cache file.
func getCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, CacheDir, CacheFileName), nil
}

// loadCache loads the URL mapping from the cache file.
func loadCache() (*URLMappingCache, error) {
	cachePath, err := getCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var cache URLMappingCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %w", err)
	}

	// Check if cache is expired
	if time.Since(cache.Timestamp) > CacheTTL {
		return nil, fmt.Errorf("cache expired")
	}

	return &cache, nil
}

// saveCache saves the URL mapping to the cache file.
func saveCache(cache *URLMappingCache) error {
	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	// Ensure cache directory exists
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// isActive checks if a branch is active (handles both bool and string "true").
func isActive(active any) bool {
	switch v := active.(type) {
	case bool:
		return v
	case string:
		return v == "true"
	default:
		return false
	}
}

// fetchFromAPI fetches URL mapping from the Snooty Data API.
func fetchFromAPI() (*URLMappingCache, error) {
	resp, err := http.Get(SnootyDataAPIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	var apiResp SnootyAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	cache := &URLMappingCache{
		Timestamp:   time.Now(),
		Mapping:     make(map[string]string),
		Branches:    make(map[string][]string),
		DriverSlugs: []string{},
	}

	// Regex to extract URL slug from fullUrl
	slugRegex := regexp.MustCompile(`/docs/(.+?)/?$`)

	// Track driver slugs using a map to avoid duplicates
	driverSlugSet := make(map[string]bool)

	for _, project := range apiResp.Data {
		var versionSlugs []string
		var baseSlugForProject string

		for _, branch := range project.Branches {
			if !isActive(branch.Active) {
				continue
			}

			match := slugRegex.FindStringSubmatch(branch.FullURL)
			if match == nil {
				continue
			}

			fullPath := match[1]
			// Extract base slug (without version) and version
			// e.g., "drivers/go/current" -> base="drivers/go", version="current"
			parts := strings.Split(fullPath, "/")

			// Check if last part looks like a version
			lastPart := parts[len(parts)-1]
			if isVersionSlug(lastPart) {
				versionSlugs = append(versionSlugs, lastPart)
				// Use base path (without version) as the slug
				baseSlug := strings.Join(parts[:len(parts)-1], "/")
				if baseSlug != "" {
					cache.Mapping[baseSlug] = project.Project
					baseSlugForProject = baseSlug
				}
			}
			// Also map the full path
			cache.Mapping[fullPath] = project.Project
		}
		if len(versionSlugs) > 0 {
			cache.Branches[project.Project] = versionSlugs
		}

		// Identify driver projects by URL pattern or displayName
		// Exclude mongodb-shell as it's not a driver
		if baseSlugForProject != "" && project.Project != "mongodb-shell" {
			if isDriverSlug(baseSlugForProject, project.DisplayName) {
				driverSlugSet[baseSlugForProject] = true
			}
		}
	}

	// Convert driver slug set to sorted slice
	for slug := range driverSlugSet {
		cache.DriverSlugs = append(cache.DriverSlugs, slug)
	}
	// Sort for deterministic output
	sortStrings(cache.DriverSlugs)

	return cache, nil
}

// isDriverSlug determines if a URL slug represents driver documentation.
// A slug is considered a driver if:
//   - It starts with "drivers/" or "languages/"
//   - OR the displayName contains "Driver" (case-insensitive)
//   - OR it's in the standaloneDriverSlugs list (for edge cases)
//
// Excludes mongodb-shell which is handled separately (use --filter mongosh).
// Excludes ODMs (Mongoid, Entity Framework), connectors (Spark, Kafka), and other
// non-driver projects - we only want actual MongoDB drivers.
func isDriverSlug(slug, displayName string) bool {
	// Check URL patterns - most drivers use "drivers/" or "languages/" prefixes
	if strings.HasPrefix(slug, "drivers/") || strings.HasPrefix(slug, "languages/") {
		return true
	}

	// Check displayName for "Driver" (handles standalone drivers like ruby-driver
	// which has URL slug "ruby-driver" and displayName "Ruby Driver")
	if strings.Contains(strings.ToLower(displayName), "driver") {
		return true
	}

	// Standalone driver slugs that don't match the above patterns.
	// These are edge cases where the URL slug doesn't start with "drivers/" or
	// "languages/" AND the displayName doesn't contain "Driver".
	//
	// As of 2026-01-08, the only such case is:
	//   - php-library: displayName is "PHP Library", URL is "php-library"
	//
	// NOT included (these are ODMs/connectors, not drivers):
	//   - mongoid: ODM for Ruby (displayName: "Mongoid")
	//   - entity-framework: ORM for C# (displayName: "Entity Framework")
	//   - spark-connector, kafka-connector: data connectors
	standaloneDriverSlugs := map[string]bool{
		"php-library": true,
	}
	return standaloneDriverSlugs[slug]
}

// sortStrings sorts a slice of strings in place - used to display the list of filters in alphabetical order.
func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// isVersionSlug checks if a string looks like a version slug.
func isVersionSlug(s string) bool {
	versionPatterns := []string{
		"current", "upcoming", "stable", "master", "latest",
		"manual", // MongoDB Manual uses "manual" as the current version directory
	}
	for _, p := range versionPatterns {
		if s == p {
			return true
		}
	}
	// Check for version patterns like v8.0, v1.13, etc.
	matched, _ := regexp.MatchString(`^v?\d+(\.\d+)*$`, s)
	return matched
}

// scanSnootyTomlFiles scans the monorepo for snooty.toml files and builds
// a mapping from snooty project name to content directory.
func scanSnootyTomlFiles(monorepoPath string) (map[string]string, error) {
	projectToDir := make(map[string]string)
	contentDir := filepath.Join(monorepoPath, "content")

	entries, err := os.ReadDir(contentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read content directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		dirPath := filepath.Join(contentDir, dirName)

		// Check for snooty.toml directly in the project directory
		snootyPath := filepath.Join(dirPath, "snooty.toml")
		if name, err := parseSnootyName(snootyPath); err == nil {
			projectToDir[name] = dirName
		}

		// Check for versioned subdirectories
		subEntries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}

		for _, subEntry := range subEntries {
			if !subEntry.IsDir() {
				continue
			}
			subDirName := subEntry.Name()
			subSnootyPath := filepath.Join(dirPath, subDirName, "snooty.toml")
			if name, err := parseSnootyName(subSnootyPath); err == nil {
				// For versioned projects, store just the base directory name
				// The version will be added from the URL during resolution
				// Only set if not already set (prefer non-versioned snooty.toml)
				if _, exists := projectToDir[name]; !exists {
					projectToDir[name] = dirName
				}
			}
		}
	}

	return projectToDir, nil
}

// parseSnootyName extracts the name field from a snooty.toml file.
func parseSnootyName(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var snootyToml SnootyToml
	if _, err := toml.Decode(string(data), &snootyToml); err != nil {
		return "", err
	}

	if snootyToml.Name == "" {
		return "", fmt.Errorf("no name field in snooty.toml")
	}

	return snootyToml.Name, nil
}

// GetURLMapping returns a URLMapping instance for resolving URLs to source files.
// It uses cached data if available and not expired, otherwise fetches from the API.
// Falls back to static mapping if API is unavailable.
func GetURLMapping(monorepoPath string) (*URLMapping, error) {
	var cache *URLMappingCache
	var err error

	// Try to load from cache first
	cache, err = loadCache()
	if err != nil {
		// Cache miss or expired, try to fetch from API
		cache, err = fetchFromAPI()
		if err != nil {
			// API failed, use static fallback
			fmt.Fprintf(os.Stderr, "Warning: Could not fetch URL mapping from API (%v), using static fallback\n", err)
			cache = getStaticFallback()
		} else {
			// Save to cache for next time
			if saveErr := saveCache(cache); saveErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not save URL mapping cache: %v\n", saveErr)
			}
		}
	}

	// Merge special cases that aren't in the API data
	mergeSpecialCases(cache)

	// Scan snooty.toml files to build project -> content dir mapping
	projectToDir, err := scanSnootyTomlFiles(monorepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan snooty.toml files: %w", err)
	}

	return &URLMapping{
		URLSlugToProject:    cache.Mapping,
		ProjectToContentDir: projectToDir,
		ProjectBranches:     cache.Branches,
		DriverSlugs:         cache.DriverSlugs,
		MonorepoPath:        monorepoPath,
	}, nil
}

// GetURLMappingWithoutMonorepo returns a URLMapping instance without requiring a monorepo path.
// This is useful for operations that only need API data (like listing drivers) and don't need
// to resolve local file paths.
func GetURLMappingWithoutMonorepo() (*URLMapping, error) {
	var cache *URLMappingCache
	var err error

	// Try to load from cache first
	cache, err = loadCache()
	if err != nil {
		// Cache miss or expired, try to fetch from API
		cache, err = fetchFromAPI()
		if err != nil {
			// API failed, use static fallback
			fmt.Fprintf(os.Stderr, "Warning: Could not fetch URL mapping from API (%v), using static fallback\n", err)
			cache = getStaticFallback()
		} else {
			// Save to cache for next time
			if saveErr := saveCache(cache); saveErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not save URL mapping cache: %v\n", saveErr)
			}
		}
	}

	// Merge special cases that aren't in the API data
	mergeSpecialCases(cache)

	return &URLMapping{
		URLSlugToProject:    cache.Mapping,
		ProjectToContentDir: make(map[string]string), // Empty - no monorepo scanning
		ProjectBranches:     cache.Branches,
		DriverSlugs:         cache.DriverSlugs,
		MonorepoPath:        "",
	}, nil
}

// mergeSpecialCases adds special URL mappings that aren't in the API data.
// These are edge cases where the URL slug doesn't follow the standard pattern.
func mergeSpecialCases(cache *URLMappingCache) {
	specialCases := map[string]string{
		// Note: get-started is handled specially in ResolveURL because
		// the slug itself is the page path (get-started.txt, not index.txt)
	}

	for slug, project := range specialCases {
		if _, exists := cache.Mapping[slug]; !exists {
			cache.Mapping[slug] = project
		}
	}
}

// specialPagePaths maps URL slugs to their actual page paths when the slug
// itself should be used as the page path instead of defaulting to "index".
var specialPagePaths = map[string]string{
	"get-started": "get-started", // /docs/get-started/ -> get-started.txt (not index.txt)
}

// specialSlugToProject maps special URL slugs to their project names.
// These are cases not covered by the API data.
var specialSlugToProject = map[string]string{
	"get-started": "landing",
}

// getStaticFallback returns a static URL mapping as a fallback when API is unavailable.
func getStaticFallback() *URLMappingCache {
	return &URLMappingCache{
		Timestamp: time.Now(),
		Mapping: map[string]string{
			"atlas":                                  "cloud-docs",
			"atlas/app-services":                     "atlas-app-services",
			"atlas/architecture":                     "atlas-architecture",
			"atlas/cli":                              "atlas-cli",
			"atlas/device-sdks":                      "realm",
			"atlas/government":                       "cloudgov",
			"atlas/operator":                         "atlas-operator",
			"bi-connector":                           "bi-connector",
			"charts":                                 "charts",
			"cloud-manager":                          "cloud-manager",
			"compass":                                "compass",
			"database-tools":                         "database-tools",
			"drivers":                                "drivers",
			"drivers/csharp":                         "csharp",
			"drivers/go":                             "golang",
			"drivers/java/sync":                      "java",
			"drivers/kotlin/coroutine":               "kotlin",
			"drivers/node":                           "node",
			"drivers/php/laravel-mongodb":            "laravel",
			"drivers/rust":                           "rust",
			"entity-framework":                       "entity-framework",
			"get-started":                            "landing",
			"kafka-connector":                        "kafka-connector",
			"kubernetes":                             "mck",
			"kubernetes-operator":                    "docs-k8s-operator",
			"languages/c/c-driver":                   "c",
			"languages/cpp/cpp-driver":               "cpp-driver",
			"languages/java/mongodb-hibernate":       "hibernate",
			"languages/java/reactive-streams-driver": "java-rs",
			"languages/kotlin/kotlin-sync-driver":    "kotlin-sync",
			"languages/python/django-mongodb":        "django",
			"languages/python/pymongo-arrow-driver":  "pymongo-arrow",
			"languages/python/pymongo-driver":        "pymongo",
			"languages/scala/scala-driver":           "scala",
			"manual":                                 "docs",
			"mcp-server":                             "mcp-server",
			"mongocli":                               "mongocli",
			"mongodb-analyzer":                       "visual-studio-extension",
			"mongodb-intellij":                       "intellij",
			"mongodb-shell":                          "mongodb-shell",
			"mongodb-voyage":                         "voyage",
			"mongodb-vscode":                         "mongodb-vscode",
			"mongoid":                                "mongoid",
			"mongosync":                              "mongosync",
			"ops-manager":                            "ops-manager",
			"php-library":                            "php-library",
			"relational-migrator":                    "docs-relational-migrator",
			"ruby-driver":                            "ruby-driver",
			"spark-connector":                        "spark-connector",
		},
		Branches: map[string][]string{
			"docs": {"manual", "upcoming", "v8.0", "v7.0", "v6.0", "v5.0", "v4.4"},
		},
		DriverSlugs: []string{
			"drivers/csharp",
			"drivers/go",
			"drivers/java/sync",
			"drivers/kotlin/coroutine",
			"drivers/node",
			"drivers/php/laravel-mongodb",
			"drivers/rust",
			"languages/c/c-driver",
			"languages/cpp/cpp-driver",
			"languages/java/mongodb-hibernate",
			"languages/java/reactive-streams-driver",
			"languages/kotlin/kotlin-sync-driver",
			"languages/python/django-mongodb",
			"languages/python/pymongo-arrow-driver",
			"languages/python/pymongo-driver",
			"languages/scala/scala-driver",
			"mongoid",
			"php-library",
			"ruby-driver",
		},
	}
}

// ResolveURL resolves a documentation URL to a source file path.
// Returns the absolute path to the source file and the content directory.
//
// URL format: www.mongodb.com/docs/{slug}/{version?}/{page-path}
// Examples:
//   - www.mongodb.com/docs/atlas/some-page/ -> content/atlas/source/some-page.txt
//   - www.mongodb.com/docs/v8.0/tutorial/install/ -> content/manual/v8.0/source/tutorial/install.txt
//   - www.mongodb.com/docs/drivers/go/current/usage/ -> content/golang/current/source/usage.txt
func (m *URLMapping) ResolveURL(url string) (sourcePath string, contentDir string, err error) {
	// Parse the URL to extract the path after /docs/
	urlPath := extractDocsPath(url)
	if urlPath == "" {
		return "", "", fmt.Errorf("invalid URL format: %s", url)
	}

	parts := strings.Split(urlPath, "/")
	if len(parts) == 0 {
		return "", "", fmt.Errorf("empty URL path")
	}

	// Try to find the longest matching slug
	var projectName string
	var pagePath string
	var version string

	for i := len(parts); i > 0; i-- {
		candidateSlug := strings.Join(parts[:i], "/")
		if proj, ok := m.URLSlugToProject[candidateSlug]; ok {
			projectName = proj
			remaining := parts[i:]

			// Check if the matched slug ends with a version
			// e.g., "drivers/go/current" matched, extract "current" as version
			slugParts := strings.Split(candidateSlug, "/")
			lastSlugPart := slugParts[len(slugParts)-1]
			if isVersionSlug(lastSlugPart) {
				version = lastSlugPart
				pagePath = strings.Join(remaining, "/")
			} else if len(remaining) > 0 && isVersionSlug(remaining[0]) {
				// Check if first remaining part is a version
				version = remaining[0]
				pagePath = strings.Join(remaining[1:], "/")
			} else {
				pagePath = strings.Join(remaining, "/")
			}
			break
		}
	}

	// Special handling for MongoDB Manual (docs project)
	// URLs like /docs/manual/... or /docs/v8.0/... map to the "docs" project
	if projectName == "" {
		if len(parts) > 0 && (parts[0] == "manual" || isVersionSlug(parts[0])) {
			projectName = "docs"
			version = parts[0]
			pagePath = strings.Join(parts[1:], "/")
		}
	}

	// Check for special slug mappings not in the API data
	if projectName == "" {
		if len(parts) > 0 {
			if proj, ok := specialSlugToProject[parts[0]]; ok {
				projectName = proj
				// For special slugs, the slug itself may be the page path
				if specialPath, ok := specialPagePaths[parts[0]]; ok {
					pagePath = specialPath
				} else {
					pagePath = strings.Join(parts[1:], "/")
				}
			}
		}
	}

	if projectName == "" {
		return "", "", fmt.Errorf("could not resolve URL slug: %s", urlPath)
	}

	// Get content directory for this project
	contentDir, ok := m.ProjectToContentDir[projectName]
	if !ok {
		return "", "", fmt.Errorf("no content directory found for project: %s", projectName)
	}

	// Build the source file path
	// For versioned projects, the content dir already includes the version
	// For non-versioned projects with a version in URL, we need to add it
	sourceDir := filepath.Join(m.MonorepoPath, "content", contentDir)

	// Check if this is a versioned project by looking for version subdirectories
	// If the content directory has version subdirectories and URL has a version, use it
	if version != "" {
		versionedPath := filepath.Join(m.MonorepoPath, "content", contentDir, version)
		if _, err := os.Stat(versionedPath); err == nil {
			sourceDir = versionedPath
		}
	}

	// Add source directory and page path
	if pagePath == "" {
		pagePath = "index"
	}
	sourcePath = filepath.Join(sourceDir, "source", pagePath+".txt")

	return sourcePath, contentDir, nil
}

// extractDocsPath extracts the path after /docs/ from a URL.
func extractDocsPath(url string) string {
	// Remove protocol and domain
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.")

	// Find /docs/ in the path
	idx := strings.Index(url, "/docs/")
	if idx == -1 {
		// Try without leading slash
		idx = strings.Index(url, "docs/")
		if idx == -1 {
			return ""
		}
		url = url[idx+5:]
	} else {
		url = url[idx+6:]
	}

	// Remove trailing slash
	url = strings.TrimSuffix(url, "/")

	return url
}

// IsDriverURL checks if a URL is for driver documentation.
// Returns true if the URL matches any known driver slug pattern.
// Excludes mongodb-shell which is handled separately.
func (m *URLMapping) IsDriverURL(url string) bool {
	urlPath := extractDocsPath(url)
	if urlPath == "" {
		return false
	}
	urlPathLower := strings.ToLower(urlPath)

	// Check against known driver slugs
	for _, slug := range m.DriverSlugs {
		slugLower := strings.ToLower(slug)
		if strings.HasPrefix(urlPathLower, slugLower+"/") || urlPathLower == slugLower {
			return true
		}
	}

	// Also check for the generic drivers/ and languages/ prefixes
	// in case a new driver was added that's not in our cached list
	if strings.HasPrefix(urlPathLower, "drivers/") || strings.HasPrefix(urlPathLower, "languages/") {
		return true
	}

	return false
}

// IsSpecificDriverURL checks if a URL is for a specific driver by project name.
// The driverName should be the Snooty project name (e.g., "golang", "pymongo", "node").
func (m *URLMapping) IsSpecificDriverURL(url, driverName string) bool {
	urlPath := extractDocsPath(url)
	if urlPath == "" {
		return false
	}

	// Find the slug for this driver
	for slug, project := range m.URLSlugToProject {
		if strings.EqualFold(project, driverName) {
			slugLower := strings.ToLower(slug)
			urlPathLower := strings.ToLower(urlPath)
			if strings.HasPrefix(urlPathLower, slugLower+"/") || urlPathLower == slugLower {
				return true
			}
		}
	}

	return false
}

// IsMongoshURL checks if a URL is for MongoDB Shell documentation.
func (m *URLMapping) IsMongoshURL(url string) bool {
	urlPath := extractDocsPath(url)
	if urlPath == "" {
		return false
	}
	urlPathLower := strings.ToLower(urlPath)

	return strings.HasPrefix(urlPathLower, "mongodb-shell/") || urlPathLower == "mongodb-shell"
}

// GetDriverSlugs returns the list of known driver URL slugs.
func (m *URLMapping) GetDriverSlugs() []string {
	return m.DriverSlugs
}
