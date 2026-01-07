// Package composables provides functionality for analyzing composables in snooty.toml files.
package composables

import (
	"github.com/grove-platform/audit-cli/internal/snooty"
)

// ComposableLocation tracks where a composable was found.
type ComposableLocation struct {
	Project    string
	Version    string // Empty for non-versioned projects
	Composable snooty.Composable
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

