// Package composables provides functionality for analyzing composables in snooty.toml files.
package composables

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

// SnootyConfig represents the structure of a snooty.toml file.
type SnootyConfig struct {
	Composables []Composable `toml:"composables"`
}

// ComposableLocation tracks where a composable was found.
type ComposableLocation struct {
	Project    string
	Version    string // Empty for non-versioned projects
	Composable Composable
	FilePath   string
	Source     string // "snooty.toml" or "rstspec.toml"
}

// ComposableGroup represents a group of similar composables.
type ComposableGroup struct {
	ID        string
	Locations []ComposableLocation
	// Similarity score (1.0 = identical, < 1.0 = similar)
	Similarity float64
}

// AnalysisResult contains the results of analyzing composables.
type AnalysisResult struct {
	// All composables found
	AllComposables []ComposableLocation
	// Groups of identical composables (same ID, same options)
	IdenticalGroups []ComposableGroup
	// Groups of similar composables (same ID, different options)
	SimilarGroups []ComposableGroup
}

// ComposableUsage tracks where a composable is used in RST files.
type ComposableUsage struct {
	ComposableID string
	Project      string
	Version      string
	UsageCount   int
	FilePaths    []string // Relative paths from monorepo root
}

