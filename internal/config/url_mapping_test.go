package config

import (
	"testing"
)

// TestIsActive tests the isActive helper function.
func TestIsActive(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"string true", "true", true},
		{"string false", "false", false},
		{"string other", "yes", false},
		{"nil", nil, false},
		{"int", 1, false},
		{"empty string", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isActive(tc.input)
			if result != tc.expected {
				t.Errorf("isActive(%v) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

// TestIsDriverSlug tests the isDriverSlug function.
func TestIsDriverSlug(t *testing.T) {
	testCases := []struct {
		name        string
		slug        string
		displayName string
		expected    bool
	}{
		// Drivers with drivers/ prefix
		{"drivers/csharp", "drivers/csharp", "C#/.NET Driver", true},
		{"drivers/go", "drivers/go", "Go Driver", true},
		{"drivers/node", "drivers/node", "Node.js Driver", true},
		{"drivers/java/sync", "drivers/java/sync", "Java Sync Driver", true},
		{"drivers/kotlin/coroutine", "drivers/kotlin/coroutine", "Kotlin Coroutine", true},

		// Drivers with languages/ prefix
		{"languages/python/pymongo-driver", "languages/python/pymongo-driver", "PyMongo", true},
		{"languages/c/c-driver", "languages/c/c-driver", "C Driver", true},
		{"languages/scala/scala-driver", "languages/scala/scala-driver", "Scala", true},

		// Drivers detected by displayName containing "Driver"
		{"ruby-driver by displayName", "ruby-driver", "Ruby Driver", true},

		// Standalone driver slugs (edge cases)
		{"php-library", "php-library", "PHP Library", true},

		// Non-drivers (should return false)
		{"mongoid ODM", "mongoid", "Mongoid", false},
		{"entity-framework ORM", "entity-framework", "Entity Framework", false},
		{"atlas", "atlas", "MongoDB Atlas", false},
		{"compass", "compass", "MongoDB Compass", false},
		{"mongodb-shell", "mongodb-shell", "MongoDB Shell", false},
		{"kafka-connector", "kafka-connector", "Kafka Connector", false},
		{"spark-connector", "spark-connector", "Spark Connector", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isDriverSlug(tc.slug, tc.displayName)
			if result != tc.expected {
				t.Errorf("isDriverSlug(%q, %q) = %v, expected %v",
					tc.slug, tc.displayName, result, tc.expected)
			}
		})
	}
}

// TestIsVersionSlug tests the isVersionSlug function.
func TestIsVersionSlug(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		// Named versions
		{"current", "current", true},
		{"upcoming", "upcoming", true},
		{"stable", "stable", true},
		{"master", "master", true},
		{"latest", "latest", true},
		{"manual", "manual", true},

		// Numeric versions
		{"v8.0", "v8.0", true},
		{"v7.0", "v7.0", true},
		{"v1.13", "v1.13", true},
		{"v2.30", "v2.30", true},
		{"8.0 without v", "8.0", true},
		{"1.0.0 semver", "1.0.0", true},
		{"v1.0.0 semver with v", "v1.0.0", true},

		// Non-versions
		{"project name", "pymongo", false},
		{"drivers prefix", "drivers", false},
		{"random string", "hello", false},
		{"empty string", "", false},
		{"partial version", "v", false},
		{"invalid version", "vX.Y", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isVersionSlug(tc.input)
			if result != tc.expected {
				t.Errorf("isVersionSlug(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

// TestExtractDocsPath tests the extractDocsPath function.
func TestExtractDocsPath(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected string
	}{
		// Full URLs with https
		{"full URL with https", "https://www.mongodb.com/docs/drivers/go/current/", "drivers/go/current"},
		{"full URL without www", "https://mongodb.com/docs/atlas/search/", "atlas/search"},

		// URLs with http
		{"http URL", "http://www.mongodb.com/docs/manual/tutorial/", "manual/tutorial"},

		// URLs without protocol
		{"no protocol with www", "www.mongodb.com/docs/compass/current/", "compass/current"},
		{"no protocol no www", "mongodb.com/docs/pymongo/", "pymongo"},

		// Edge cases
		{"trailing slash removed", "https://mongodb.com/docs/atlas/", "atlas"},
		{"no trailing slash", "https://mongodb.com/docs/atlas", "atlas"},
		{"deep path", "https://mongodb.com/docs/drivers/node/current/fundamentals/crud/", "drivers/node/current/fundamentals/crud"},

		// Invalid URLs
		{"no docs path", "https://mongodb.com/products/atlas", ""},
		{"empty string", "", ""},
		{"just domain", "mongodb.com", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractDocsPath(tc.url)
			if result != tc.expected {
				t.Errorf("extractDocsPath(%q) = %q, expected %q", tc.url, result, tc.expected)
			}
		})
	}
}

// createTestURLMapping creates a URLMapping for testing with sample driver data.
func createTestURLMapping() *URLMapping {
	return &URLMapping{
		URLSlugToProject: map[string]string{
			"drivers/go":                      "golang",
			"drivers/node":                    "node",
			"drivers/csharp":                  "csharp",
			"languages/python/pymongo-driver": "pymongo",
			"ruby-driver":                     "ruby-driver",
			"php-library":                     "php-library",
			"mongodb-shell":                   "mongodb-shell",
			"atlas":                           "cloud-docs",
		},
		DriverSlugs: []string{
			"drivers/go",
			"drivers/node",
			"drivers/csharp",
			"languages/python/pymongo-driver",
			"ruby-driver",
			"php-library",
		},
		ProjectToContentDir: map[string]string{},
		ProjectBranches:     map[string][]string{},
		MonorepoPath:        "",
	}
}

// TestIsDriverURL tests the IsDriverURL method.
func TestIsDriverURL(t *testing.T) {
	m := createTestURLMapping()

	testCases := []struct {
		name     string
		url      string
		expected bool
	}{
		// Known driver URLs
		{"go driver", "https://mongodb.com/docs/drivers/go/current/", true},
		{"node driver", "https://mongodb.com/docs/drivers/node/current/fundamentals/", true},
		{"csharp driver", "https://mongodb.com/docs/drivers/csharp/current/", true},
		{"pymongo driver", "https://mongodb.com/docs/languages/python/pymongo-driver/current/", true},
		{"ruby driver", "https://mongodb.com/docs/ruby-driver/current/", true},
		{"php library", "https://mongodb.com/docs/php-library/current/", true},

		// Generic drivers/ and languages/ prefixes (for new drivers not in cache)
		{"unknown driver in drivers/", "https://mongodb.com/docs/drivers/unknown/current/", true},
		{"unknown driver in languages/", "https://mongodb.com/docs/languages/java/new-driver/", true},

		// Non-driver URLs
		{"mongodb shell", "https://mongodb.com/docs/mongodb-shell/current/", false},
		{"atlas", "https://mongodb.com/docs/atlas/search/", false},
		{"manual", "https://mongodb.com/docs/manual/tutorial/", false},
		{"compass", "https://mongodb.com/docs/compass/current/", false},

		// Edge cases
		{"empty URL", "", false},
		{"invalid URL", "not-a-url", false},
		{"exact slug match", "https://mongodb.com/docs/drivers/go", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := m.IsDriverURL(tc.url)
			if result != tc.expected {
				t.Errorf("IsDriverURL(%q) = %v, expected %v", tc.url, result, tc.expected)
			}
		})
	}
}

// TestIsSpecificDriverURL tests the IsSpecificDriverURL method.
func TestIsSpecificDriverURL(t *testing.T) {
	m := createTestURLMapping()

	testCases := []struct {
		name       string
		url        string
		driverName string
		expected   bool
	}{
		// Matching driver URLs
		{"golang match", "https://mongodb.com/docs/drivers/go/current/", "golang", true},
		{"node match", "https://mongodb.com/docs/drivers/node/current/", "node", true},
		{"pymongo match", "https://mongodb.com/docs/languages/python/pymongo-driver/current/", "pymongo", true},
		{"ruby-driver match", "https://mongodb.com/docs/ruby-driver/current/", "ruby-driver", true},

		// Case insensitive matching
		{"golang uppercase", "https://mongodb.com/docs/drivers/go/current/", "GOLANG", true},
		{"node mixed case", "https://mongodb.com/docs/drivers/node/current/", "Node", true},

		// Non-matching
		{"wrong driver", "https://mongodb.com/docs/drivers/go/current/", "node", false},
		{"non-driver URL", "https://mongodb.com/docs/atlas/search/", "golang", false},

		// Edge cases
		{"empty URL", "", "golang", false},
		{"empty driver name", "https://mongodb.com/docs/drivers/go/current/", "", false},
		{"unknown driver name", "https://mongodb.com/docs/drivers/go/current/", "unknown", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := m.IsSpecificDriverURL(tc.url, tc.driverName)
			if result != tc.expected {
				t.Errorf("IsSpecificDriverURL(%q, %q) = %v, expected %v",
					tc.url, tc.driverName, result, tc.expected)
			}
		})
	}
}

// TestIsMongoshURL tests the IsMongoshURL method.
func TestIsMongoshURL(t *testing.T) {
	m := createTestURLMapping()

	testCases := []struct {
		name     string
		url      string
		expected bool
	}{
		// MongoDB Shell URLs
		{"mongosh with path", "https://mongodb.com/docs/mongodb-shell/current/", true},
		{"mongosh root", "https://mongodb.com/docs/mongodb-shell/", true},
		{"mongosh exact", "https://mongodb.com/docs/mongodb-shell", true},
		{"mongosh deep path", "https://mongodb.com/docs/mongodb-shell/reference/methods/", true},

		// Case insensitive
		{"mongosh uppercase", "https://mongodb.com/docs/MONGODB-SHELL/current/", true},
		{"mongosh mixed case", "https://mongodb.com/docs/MongoDB-Shell/current/", true},

		// Non-mongosh URLs
		{"driver URL", "https://mongodb.com/docs/drivers/go/current/", false},
		{"atlas URL", "https://mongodb.com/docs/atlas/", false},
		{"manual URL", "https://mongodb.com/docs/manual/", false},

		// Edge cases
		{"empty URL", "", false},
		{"partial match", "https://mongodb.com/docs/mongodb-shell-like/", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := m.IsMongoshURL(tc.url)
			if result != tc.expected {
				t.Errorf("IsMongoshURL(%q) = %v, expected %v", tc.url, result, tc.expected)
			}
		})
	}
}

// TestGetDriverSlugs tests the GetDriverSlugs method.
func TestGetDriverSlugs(t *testing.T) {
	m := createTestURLMapping()

	slugs := m.GetDriverSlugs()

	if len(slugs) != 6 {
		t.Errorf("Expected 6 driver slugs, got %d", len(slugs))
	}

	// Check that expected slugs are present
	expectedSlugs := map[string]bool{
		"drivers/go":                      true,
		"drivers/node":                    true,
		"drivers/csharp":                  true,
		"languages/python/pymongo-driver": true,
		"ruby-driver":                     true,
		"php-library":                     true,
	}

	for _, slug := range slugs {
		if !expectedSlugs[slug] {
			t.Errorf("Unexpected slug in result: %q", slug)
		}
	}
}

// TestSortStrings tests the sortStrings helper function.
func TestSortStrings(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"already sorted", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"reverse order", []string{"c", "b", "a"}, []string{"a", "b", "c"}},
		{"mixed order", []string{"banana", "apple", "cherry"}, []string{"apple", "banana", "cherry"}},
		{"empty slice", []string{}, []string{}},
		{"single element", []string{"only"}, []string{"only"}},
		{"duplicates", []string{"b", "a", "b", "a"}, []string{"a", "a", "b", "b"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Make a copy to avoid modifying the test case
			input := make([]string, len(tc.input))
			copy(input, tc.input)

			sortStrings(input)

			if len(input) != len(tc.expected) {
				t.Errorf("Length mismatch: got %d, expected %d", len(input), len(tc.expected))
				return
			}

			for i, v := range input {
				if v != tc.expected[i] {
					t.Errorf("At index %d: got %q, expected %q", i, v, tc.expected[i])
				}
			}
		})
	}
}

