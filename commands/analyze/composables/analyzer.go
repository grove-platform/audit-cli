// Package composables provides functionality for analyzing composables in snooty.toml files.
package composables

import (
	"sort"

	"github.com/grove-platform/audit-cli/internal/snooty"
)

// AnalyzeComposables analyzes composables and groups them by similarity.
//
// This function identifies:
// 1. Identical composables across projects (same ID, same options) - consolidation candidates
// 2. Similar composables with different IDs but overlapping options - potential consolidation
//
// Parameters:
//   - locations: All composable locations found in the monorepo
//
// Returns:
//   - *AnalysisResult: Analysis results with grouped composables
func AnalyzeComposables(locations []ComposableLocation) *AnalysisResult {
	result := &AnalysisResult{
		AllComposables:  locations,
		IdenticalGroups: []ComposableGroup{},
		SimilarGroups:   []ComposableGroup{},
	}

	// Group composables by ID
	groupsByID := make(map[string][]ComposableLocation)
	for _, loc := range locations {
		id := loc.Composable.ID
		groupsByID[id] = append(groupsByID[id], loc)
	}

	// Find identical composables (same ID appearing in multiple projects)
	for id, locs := range groupsByID {
		if len(locs) <= 1 {
			continue
		}

		// Check if all composables with this ID are identical
		if areComposablesIdentical(locs) {
			result.IdenticalGroups = append(result.IdenticalGroups, ComposableGroup{
				ID:         id,
				Locations:  locs,
				Similarity: 1.0,
			})
		}
	}

	// Find similar composables (different IDs but similar option sets)
	result.SimilarGroups = findSimilarComposables(locations, groupsByID)

	// Sort groups by ID for consistent output
	sort.Slice(result.IdenticalGroups, func(i, j int) bool {
		return result.IdenticalGroups[i].ID < result.IdenticalGroups[j].ID
	})
	sort.Slice(result.SimilarGroups, func(i, j int) bool {
		// Sort by similarity (descending), then by first ID
		if result.SimilarGroups[i].Similarity != result.SimilarGroups[j].Similarity {
			return result.SimilarGroups[i].Similarity > result.SimilarGroups[j].Similarity
		}
		return result.SimilarGroups[i].ID < result.SimilarGroups[j].ID
	})

	return result
}

// areComposablesIdentical checks if all composables in a group are identical.
func areComposablesIdentical(locs []ComposableLocation) bool {
	if len(locs) <= 1 {
		return true
	}

	first := locs[0].Composable
	for i := 1; i < len(locs); i++ {
		if !composablesEqual(first, locs[i].Composable) {
			return false
		}
	}
	return true
}

// composablesEqual checks if two composables are identical.
func composablesEqual(a, b snooty.Composable) bool {
	// Compare basic fields
	if a.ID != b.ID || a.Title != b.Title || a.Default != b.Default {
		return false
	}

	// Compare options
	if len(a.Options) != len(b.Options) {
		return false
	}

	// Create sorted option strings for comparison
	aOpts := optionsToSortedStrings(a.Options)
	bOpts := optionsToSortedStrings(b.Options)

	for i := range aOpts {
		if aOpts[i] != bOpts[i] {
			return false
		}
	}

	return true
}

// optionsToSortedStrings converts options to sorted strings for comparison.
func optionsToSortedStrings(options []snooty.ComposableOption) []string {
	var strs []string
	for _, opt := range options {
		strs = append(strs, opt.ID+":"+opt.Title)
	}
	sort.Strings(strs)
	return strs
}

// findSimilarComposables finds composables with different IDs but similar option sets.
// This helps identify potential consolidation opportunities across different composable IDs.
func findSimilarComposables(locations []ComposableLocation, groupsByID map[string][]ComposableLocation) []ComposableGroup {
	const similarityThreshold = 0.6 // At least 60% option overlap to be considered similar

	var similarGroups []ComposableGroup

	// Get unique composables (one per ID, preferring the one with most options)
	uniqueComposables := make(map[string]ComposableLocation)
	for id, locs := range groupsByID {
		// Pick the composable with the most options as the representative
		representative := locs[0]
		for _, loc := range locs {
			if len(loc.Composable.Options) > len(representative.Composable.Options) {
				representative = loc
			}
		}
		uniqueComposables[id] = representative
	}

	// Get sorted list of IDs for deterministic iteration
	var ids []string
	for id := range uniqueComposables {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// Compare each pair of composables with different IDs
	processed := make(map[string]bool)
	for i := 0; i < len(ids); i++ {
		id1 := ids[i]
		if processed[id1] {
			continue
		}

		loc1 := uniqueComposables[id1]
		var similarLocs []ComposableLocation
		similarLocs = append(similarLocs, loc1)

		for j := i + 1; j < len(ids); j++ {
			id2 := ids[j]
			if processed[id2] {
				continue
			}

			loc2 := uniqueComposables[id2]
			similarity := calculateOptionSimilarity(loc1.Composable, loc2.Composable)

			if similarity >= similarityThreshold {
				similarLocs = append(similarLocs, loc2)
				processed[id2] = true
			}
		}

		// If we found similar composables, create a group
		if len(similarLocs) > 1 {
			processed[id1] = true

			// Calculate average similarity across all in the group
			avgSimilarity := calculateGroupSimilarity(similarLocs)

			// Create a combined ID showing all the IDs in the group
			var combinedIDs []string
			for _, loc := range similarLocs {
				combinedIDs = append(combinedIDs, loc.Composable.ID)
			}
			sort.Strings(combinedIDs)

			similarGroups = append(similarGroups, ComposableGroup{
				ID:         combinedIDs[0], // Use first ID for sorting
				Locations:  similarLocs,
				Similarity: avgSimilarity,
			})
		}
	}

	return similarGroups
}

// calculateOptionSimilarity calculates the Jaccard similarity between two composables' option sets.
// Returns a value between 0 and 1, where 1 means identical option sets.
func calculateOptionSimilarity(a, b snooty.Composable) float64 {
	// Get option IDs for both composables
	aOptions := make(map[string]bool)
	for _, opt := range a.Options {
		aOptions[opt.ID] = true
	}

	bOptions := make(map[string]bool)
	for _, opt := range b.Options {
		bOptions[opt.ID] = true
	}

	// Calculate intersection and union
	intersection := 0
	union := make(map[string]bool)

	for opt := range aOptions {
		union[opt] = true
		if bOptions[opt] {
			intersection++
		}
	}

	for opt := range bOptions {
		union[opt] = true
	}

	if len(union) == 0 {
		return 0.0
	}

	// Jaccard similarity = intersection / union
	return float64(intersection) / float64(len(union))
}

// calculateGroupSimilarity calculates the average pairwise similarity within a group.
func calculateGroupSimilarity(locs []ComposableLocation) float64 {
	if len(locs) <= 1 {
		return 1.0
	}

	totalSimilarity := 0.0
	comparisons := 0

	for i := 0; i < len(locs); i++ {
		for j := i + 1; j < len(locs); j++ {
			similarity := calculateOptionSimilarity(locs[i].Composable, locs[j].Composable)
			totalSimilarity += similarity
			comparisons++
		}
	}

	if comparisons == 0 {
		return 1.0
	}

	return totalSimilarity / float64(comparisons)
}

