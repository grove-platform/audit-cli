package url

import (
	"testing"
)

// TestBuildURL tests the URL construction logic with various combinations
// of URL slug, version, and page path.
func TestBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		urlSlug  string
		version  string
		pagePath string
		expected string
	}{
		{
			name:     "non-versioned project with page",
			baseURL:  "https://www.mongodb.com/docs",
			urlSlug:  "atlas",
			version:  "",
			pagePath: "manage-clusters",
			expected: "https://www.mongodb.com/docs/atlas/manage-clusters/",
		},
		{
			name:     "non-versioned project index",
			baseURL:  "https://www.mongodb.com/docs",
			urlSlug:  "compass",
			version:  "",
			pagePath: "",
			expected: "https://www.mongodb.com/docs/compass/",
		},
		{
			name:     "versioned project with page",
			baseURL:  "https://www.mongodb.com/docs",
			urlSlug:  "drivers/go",
			version:  "current",
			pagePath: "atlas-search",
			expected: "https://www.mongodb.com/docs/drivers/go/current/atlas-search/",
		},
		{
			name:     "versioned project index",
			baseURL:  "https://www.mongodb.com/docs",
			urlSlug:  "drivers/node",
			version:  "current",
			pagePath: "",
			expected: "https://www.mongodb.com/docs/drivers/node/current/",
		},
		{
			name:     "MongoDB Manual (empty slug) with version",
			baseURL:  "https://www.mongodb.com/docs",
			urlSlug:  "",
			version:  "manual",
			pagePath: "indexes",
			expected: "https://www.mongodb.com/docs/manual/indexes/",
		},
		{
			name:     "MongoDB Manual index",
			baseURL:  "https://www.mongodb.com/docs",
			urlSlug:  "",
			version:  "manual",
			pagePath: "",
			expected: "https://www.mongodb.com/docs/manual/",
		},
		{
			name:     "MongoDB Manual specific version",
			baseURL:  "https://www.mongodb.com/docs",
			urlSlug:  "",
			version:  "v8.0",
			pagePath: "tutorial/install-mongodb-on-os-x",
			expected: "https://www.mongodb.com/docs/v8.0/tutorial/install-mongodb-on-os-x/",
		},
		{
			name:     "nested URL slug",
			baseURL:  "https://www.mongodb.com/docs",
			urlSlug:  "atlas/app-services",
			version:  "",
			pagePath: "logs",
			expected: "https://www.mongodb.com/docs/atlas/app-services/logs/",
		},
		{
			name:     "base URL with trailing slash",
			baseURL:  "https://www.mongodb.com/docs/",
			urlSlug:  "atlas",
			version:  "",
			pagePath: "clusters",
			expected: "https://www.mongodb.com/docs/atlas/clusters/",
		},
		{
			name:     "custom base URL (staging)",
			baseURL:  "https://docs-staging.mongodb.com",
			urlSlug:  "atlas",
			version:  "",
			pagePath: "index",
			expected: "https://docs-staging.mongodb.com/atlas/index/",
		},
		{
			name:     "deeply nested page path",
			baseURL:  "https://www.mongodb.com/docs",
			urlSlug:  "drivers/java/sync",
			version:  "current",
			pagePath: "fundamentals/connection/connection-options",
			expected: "https://www.mongodb.com/docs/drivers/java/sync/current/fundamentals/connection/connection-options/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildURL(tt.baseURL, tt.urlSlug, tt.version, tt.pagePath)
			if result != tt.expected {
				t.Errorf("buildURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestGetURLSlugForProject tests the project name to URL slug lookup.
func TestGetURLSlugForProject(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		expected    string
		expectError bool
	}{
		{
			name:        "Atlas project",
			projectName: "cloud-docs",
			expected:    "atlas",
			expectError: false,
		},
		{
			name:        "Go driver",
			projectName: "golang",
			expected:    "drivers/go",
			expectError: false,
		},
		{
			name:        "MongoDB Manual (empty slug)",
			projectName: "docs",
			expected:    "",
			expectError: false,
		},
		{
			name:        "Compass",
			projectName: "compass",
			expected:    "compass",
			expectError: false,
		},
		{
			name:        "App Services",
			projectName: "atlas-app-services",
			expected:    "atlas/app-services",
			expectError: false,
		},
		{
			name:        "unknown project",
			projectName: "nonexistent-project",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getURLSlugForProject(tt.projectName)

			if tt.expectError {
				if err == nil {
					t.Errorf("getURLSlugForProject() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("getURLSlugForProject() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("getURLSlugForProject() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

// TestProjectToURLSlugCoverage verifies important projects are mapped.
func TestProjectToURLSlugCoverage(t *testing.T) {
	// These are critical projects that must always have mappings
	criticalProjects := []string{
		// Atlas products
		"cloud-docs",
		"atlas-cli",
		"atlas-operator",
		"atlas-app-services",
		// Server
		"docs",
		// Key drivers
		"golang",
		"node",
		"java",
		"csharp",
		"pymongo",
		"rust",
		// Tools
		"compass",
		"mongodb-shell",
		"database-tools",
	}

	for _, project := range criticalProjects {
		t.Run(project, func(t *testing.T) {
			_, exists := projectToURLSlug[project]
			if !exists {
				t.Errorf("critical project %q is missing from projectToURLSlug map", project)
			}
		})
	}
}

// TestExtractMonorepoPath tests extraction of the monorepo path.
func TestExtractMonorepoPath(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		contentIdx int
		expected   string
	}{
		{
			name:       "typical path",
			filePath:   "/Users/user/docs-mongodb-internal/content/atlas/source/index.txt",
			contentIdx: 33, // Position where /content/ starts
			expected:   "/Users/user/docs-mongodb-internal",
		},
		{
			name:       "root path",
			filePath:   "/content/atlas/source/index.txt",
			contentIdx: 0,
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMonorepoPath(tt.filePath, tt.contentIdx)
			if result != tt.expected {
				t.Errorf("extractMonorepoPath() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestBuildURLEdgeCases tests edge cases in URL construction.
func TestBuildURLEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		urlSlug  string
		version  string
		pagePath string
		expected string
	}{
		{
			name:     "all empty except base produces double slash",
			baseURL:  "https://www.mongodb.com/docs",
			urlSlug:  "",
			version:  "",
			pagePath: "",
			// Note: This edge case produces a double slash, but it never occurs in practice
			// because there's always either a slug or version for any real project
			expected: "https://www.mongodb.com/docs//",
		},
		{
			name:     "only page path",
			baseURL:  "https://www.mongodb.com/docs",
			urlSlug:  "",
			version:  "",
			pagePath: "some-page",
			expected: "https://www.mongodb.com/docs/some-page/",
		},
		{
			name:     "multiple trailing slashes in base - trims only one",
			baseURL:  "https://www.mongodb.com/docs///",
			urlSlug:  "atlas",
			version:  "",
			pagePath: "",
			// TrimSuffix only removes one trailing slash
			expected: "https://www.mongodb.com/docs///atlas/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildURL(tt.baseURL, tt.urlSlug, tt.version, tt.pagePath)
			if result != tt.expected {
				t.Errorf("buildURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestURLSlugNeverDoubleSlash verifies that no URL slug produces double slashes.
func TestURLSlugNeverDoubleSlash(t *testing.T) {
	baseURL := "https://www.mongodb.com/docs"

	for project, slug := range projectToURLSlug {
		t.Run(project, func(t *testing.T) {
			// Test with version
			url := buildURL(baseURL, slug, "current", "page")
			if containsDoubleSlash(url) {
				t.Errorf("URL for project %q contains double slash: %s", project, url)
			}

			// Test without version
			url = buildURL(baseURL, slug, "", "page")
			if containsDoubleSlash(url) {
				t.Errorf("URL for project %q (no version) contains double slash: %s", project, url)
			}
		})
	}
}

// containsDoubleSlash checks if the URL path contains //
func containsDoubleSlash(url string) bool {
	// Skip the protocol part (https://)
	if len(url) < 8 {
		return false
	}
	pathPart := url[8:] // Skip "https://"
	for i := 0; i < len(pathPart)-1; i++ {
		if pathPart[i] == '/' && pathPart[i+1] == '/' {
			return true
		}
	}
	return false
}
