// Package projectinfo provides utilities for working with MongoDB documentation project structure.
package projectinfo

// ContentDirToProduct maps content directory names to their display product names.
// This is used for reporting and analysis purposes.
//
// Content directories are the top-level directories under content/ in the docs monorepo
// that contain driver or product documentation (e.g., "pymongo-driver", "node", "golang").
//
// Note: This map should include ALL driver/product content directories, not just
// those with tested code examples. Testability is determined separately.
var ContentDirToProduct = map[string]string{
	"c-driver":        "C",
	"cpp-driver":      "C++",
	"csharp":          "C#",
	"golang":          "Go",
	"java":            "Java (Sync)",
	"java-rs":         "Java (Reactive Streams)",
	"kotlin":          "Kotlin (Coroutine)",
	"kotlin-sync":     "Kotlin (Sync)",
	"laravel-mongodb": "Laravel",
	"mongodb-shell":   "MongoDB Shell",
	"node":            "Node.js",
	"php-library":     "PHP",
	"pymongo-arrow":   "PyMongo Arrow",
	"pymongo-driver":  "Python",
	"ruby-driver":     "Ruby",
	"rust":            "Rust",
	"scala-driver":    "Scala",
	"swift":           "Swift",
}

// GetProductFromContentDir returns the display product name for a content directory.
// Returns the product name if found, or empty string if the content directory is not recognized.
func GetProductFromContentDir(contentDir string) string {
	if product, ok := ContentDirToProduct[contentDir]; ok {
		return product
	}
	return ""
}

// GetAllContentDirs returns a slice of all known content directory names.
// Useful for validation or iteration.
func GetAllContentDirs() []string {
	dirs := make([]string, 0, len(ContentDirToProduct))
	for dir := range ContentDirToProduct {
		dirs = append(dirs, dir)
	}
	return dirs
}

// GetAllProducts returns a slice of all known product display names.
// Useful for validation or reporting.
func GetAllProducts() []string {
	// Use a map to deduplicate (in case multiple dirs map to same product)
	seen := make(map[string]bool)
	products := make([]string, 0, len(ContentDirToProduct))
	for _, product := range ContentDirToProduct {
		if !seen[product] {
			seen[product] = true
			products = append(products, product)
		}
	}
	return products
}

