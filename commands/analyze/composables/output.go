// Package composables provides functionality for analyzing composables in snooty.toml files.
package composables

import (
	"fmt"
	"sort"
	"strings"

	"github.com/grove-platform/audit-cli/internal/snooty"
)

// PrintResults prints the analysis results in a formatted table.
func PrintResults(result *AnalysisResult, verbose bool, findSimilar bool, findUsages bool, usages map[string]*ComposableUsage) {
	fmt.Printf("Composables Analysis\n")
	fmt.Printf("====================\n\n")

	fmt.Printf("Total composable definitions found: %d\n", len(result.AllComposables))
	fmt.Printf("(Each [[composables]] stanza in snooty.toml/rstspec.toml files)\n\n")

	// Print summary by ID
	printSummaryByID(result)

	// Print identical and similar groups only if requested
	if findSimilar {
		// Print identical groups
		if len(result.IdenticalGroups) > 0 {
			fmt.Printf("\nIdentical Composables (Consolidation Candidates)\n")
			fmt.Printf("================================================\n\n")
			for i, group := range result.IdenticalGroups {
				printComposableGroup(group, true, verbose)
				// Add separator between groups (but not after the last one)
				if i < len(result.IdenticalGroups)-1 {
					fmt.Printf("\n%s\n\n", strings.Repeat("-", 80))
				}
			}
		}

		// Print similar groups
		if len(result.SimilarGroups) > 0 {
			fmt.Printf("\nSimilar Composables (Review Recommended)\n")
			fmt.Printf("========================================\n\n")
			for i, group := range result.SimilarGroups {
				printComposableGroup(group, false, verbose)
				// Add separator between groups (but not after the last one)
				if i < len(result.SimilarGroups)-1 {
					fmt.Printf("\n%s\n\n", strings.Repeat("-", 80))
				}
			}
		}
	}

	// Print usage information if requested
	if findUsages && usages != nil {
		fmt.Printf("\nComposable Usages\n")
		fmt.Printf("=================\n\n")
		printUsageInformation(result.AllComposables, usages, verbose)
	}

	// Print all composables table
	fmt.Printf("\nAll Composables\n")
	fmt.Printf("===============\n\n")
	printAllComposablesTable(result.AllComposables, verbose)
}

// printSummaryByID prints a summary of composables grouped by ID.
func printSummaryByID(result *AnalysisResult) {
	// Group by ID
	countByID := make(map[string]int)
	for _, loc := range result.AllComposables {
		countByID[loc.Composable.ID]++
	}

	// Sort IDs
	var ids []string
	for id := range countByID {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	fmt.Printf("Composables by ID:\n")
	for _, id := range ids {
		count := countByID[id]
		status := ""
		if count > 1 {
			status = " (multiple instances)"
		}
		fmt.Printf("  - %s: %d%s\n", id, count, status)
	}
}

// printComposableGroup prints a group of composables.
func printComposableGroup(group ComposableGroup, isIdentical bool, verbose bool) {
	if isIdentical {
		// For identical composables, they all have the same ID
		fmt.Printf("ID: %s\n", group.ID)
		fmt.Printf("Occurrences: %d\n", len(group.Locations))

		// Get the first composable as reference
		ref := group.Locations[0].Composable
		fmt.Printf("Title: %s\n", ref.Title)
		fmt.Printf("Default: %s\n", ref.Default)

		if verbose {
			fmt.Printf("Options:\n")
			printOptionsVerbose(ref.Options, "  ")
		} else {
			fmt.Printf("Options: %s\n", formatOptions(ref.Options))
		}

		fmt.Printf("\nFound in:\n")
		for _, loc := range group.Locations {
			location := loc.Project
			if loc.Version != "" {
				location += "/" + loc.Version
			}
			fmt.Printf("  - %s (%s)\n", location, loc.Source)
		}
	} else {
		// For similar composables with different IDs, show each one
		fmt.Printf("Group: %.1f%% Similarity\n", group.Similarity*100)
		fmt.Printf("%s\n", strings.Repeat("=", 40))
		fmt.Printf("Composables: %d\n", len(group.Locations))

		fmt.Printf("\nComposables in this group:\n")
		for i, loc := range group.Locations {
			location := loc.Project
			if loc.Version != "" {
				location += "/" + loc.Version
			}

			fmt.Printf("\n  %d. ID: %s\n", i+1, loc.Composable.ID)
			fmt.Printf("     Location: %s (%s)\n", location, loc.Source)
			fmt.Printf("     Title: %s\n", loc.Composable.Title)
			if loc.Composable.Default != "" {
				fmt.Printf("     Default: %s\n", loc.Composable.Default)
			}

			if verbose {
				fmt.Printf("     Options:\n")
				printOptionsVerbose(loc.Composable.Options, "       ")
			} else {
				fmt.Printf("     Options: %s\n", formatOptions(loc.Composable.Options))
			}
		}

		// Show common options
		if verbose {
			commonOptions := findCommonOptions(group.Locations)
			if len(commonOptions) > 0 {
				fmt.Printf("\n  Common options across all:\n")
				printOptionsVerbose(commonOptions, "    ")
			}
		}
	}

}

// printAllComposablesTable prints all composables in a table format.
func printAllComposablesTable(locations []ComposableLocation, verbose bool) {
	// Sort by project, version, then ID
	sorted := make([]ComposableLocation, len(locations))
	copy(sorted, locations)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Project != sorted[j].Project {
			return sorted[i].Project < sorted[j].Project
		}
		if sorted[i].Version != sorted[j].Version {
			return sorted[i].Version < sorted[j].Version
		}
		return sorted[i].Composable.ID < sorted[j].Composable.ID
	})

	if verbose {
		// Verbose table format with multi-line options
		fmt.Printf("%-20s %-15s %-15s %-30s %-30s %-15s %s\n", "Project", "Version", "Source", "ID", "Title", "Default", "Options")
		fmt.Printf("%s\n", strings.Repeat("-", 155))

		for i, loc := range sorted {
			version := loc.Version
			if version == "" {
				version = "(none)"
			}

			// Format options as bullet list
			optionLines := formatOptionsAsBullets(loc.Composable.Options)

			// Print first line with all columns
			if len(optionLines) > 0 {
				fmt.Printf("%-20s %-15s %-15s %-30s %-30s %-15s %s\n",
					truncate(loc.Project, 20),
					truncate(version, 15),
					truncate(loc.Source, 15),
					truncate(loc.Composable.ID, 30),
					truncate(loc.Composable.Title, 30),
					truncate(loc.Composable.Default, 15),
					optionLines[0])

				// Print continuation lines with options only
				for j := 1; j < len(optionLines); j++ {
					fmt.Printf("%-20s %-15s %-15s %-30s %-30s %-15s %s\n", "", "", "", "", "", "", optionLines[j])
				}
			} else {
				fmt.Printf("%-20s %-15s %-15s %-30s %-30s %-15s\n",
					truncate(loc.Project, 20),
					truncate(version, 15),
					truncate(loc.Source, 15),
					truncate(loc.Composable.ID, 30),
					truncate(loc.Composable.Title, 30),
					truncate(loc.Composable.Default, 15))
			}

			// Add separator line between rows (but not after the last row)
			if i < len(sorted)-1 {
				fmt.Printf("%s\n", strings.Repeat("-", 155))
			}
		}
	} else {
		// Compact table format
		fmt.Printf("%-20s %-15s %-15s %-30s %-30s %s\n", "Project", "Version", "Source", "ID", "Title", "Options")
		fmt.Printf("%s\n", strings.Repeat("-", 135))

		for _, loc := range sorted {
			version := loc.Version
			if version == "" {
				version = "(none)"
			}
			options := formatOptions(loc.Composable.Options)
			fmt.Printf("%-20s %-15s %-15s %-30s %-30s %s\n",
				truncate(loc.Project, 20),
				truncate(version, 15),
				truncate(loc.Source, 15),
				truncate(loc.Composable.ID, 30),
				truncate(loc.Composable.Title, 30),
				truncate(options, 40))
		}
	}
}

// formatOptions formats options as a comma-separated list of IDs.
func formatOptions(options []snooty.ComposableOption) string {
	var ids []string
	for _, opt := range options {
		ids = append(ids, opt.ID)
	}
	return strings.Join(ids, ", ")
}

// formatOptionsAsBullets formats options as bullet points for table display.
func formatOptionsAsBullets(options []snooty.ComposableOption) []string {
	var lines []string
	for _, opt := range options {
		lines = append(lines, fmt.Sprintf("• %s: %s", opt.ID, opt.Title))
	}
	return lines
}

// printOptionsVerbose prints options in verbose format with wrapping.
func printOptionsVerbose(options []snooty.ComposableOption, indent string) {
	const maxWidth = 100 // Maximum width for wrapped text

	for _, opt := range options {
		optText := fmt.Sprintf("%s- %s: %s", indent, opt.ID, opt.Title)

		// Wrap text if it exceeds maxWidth
		if len(optText) > maxWidth {
			// Print the first line up to maxWidth
			fmt.Println(optText[:maxWidth])

			// Print continuation lines
			remaining := optText[maxWidth:]
			continuationIndent := indent + "  "
			for len(remaining) > 0 {
				if len(remaining) <= maxWidth-len(continuationIndent) {
					fmt.Printf("%s%s\n", continuationIndent, remaining)
					break
				}
				// Find a good break point (space or comma)
				breakPoint := findBreakPoint(remaining, maxWidth-len(continuationIndent))
				fmt.Printf("%s%s\n", continuationIndent, remaining[:breakPoint])
				remaining = strings.TrimSpace(remaining[breakPoint:])
			}
		} else {
			fmt.Println(optText)
		}
	}
}

// findBreakPoint finds a good place to break a line (at a space or comma).
func findBreakPoint(s string, maxLen int) int {
	if len(s) <= maxLen {
		return len(s)
	}

	// Look for a space or comma near the end
	for i := maxLen; i > maxLen-20 && i > 0; i-- {
		if s[i] == ' ' || s[i] == ',' {
			return i
		}
	}

	// If no good break point found, just break at maxLen
	return maxLen
}

// truncate truncates a string to the specified length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// findCommonOptions finds options that appear in all composables in the group.
func findCommonOptions(locations []ComposableLocation) []snooty.ComposableOption {
	if len(locations) == 0 {
		return nil
	}

	// Start with options from the first composable
	commonMap := make(map[string]snooty.ComposableOption)
	for _, opt := range locations[0].Composable.Options {
		commonMap[opt.ID] = opt
	}

	// Remove options that don't appear in all composables
	for i := 1; i < len(locations); i++ {
		optionsInThis := make(map[string]bool)
		for _, opt := range locations[i].Composable.Options {
			optionsInThis[opt.ID] = true
		}

		// Remove options from commonMap that aren't in this composable
		for optID := range commonMap {
			if !optionsInThis[optID] {
				delete(commonMap, optID)
			}
		}
	}

	// Convert map to sorted slice
	var common []snooty.ComposableOption
	for _, opt := range commonMap {
		common = append(common, opt)
	}

	sort.Slice(common, func(i, j int) bool {
		return common[i].ID < common[j].ID
	})

	return common
}

// printUsageInformation prints usage information for composables.
func printUsageInformation(composables []ComposableLocation, usages map[string]*ComposableUsage, verbose bool) {
	// Calculate total unique pages across all composables
	uniquePages := make(map[string]bool)
	for _, usage := range usages {
		for _, filePath := range usage.FilePaths {
			uniquePages[filePath] = true
		}
	}

	// Print total unique pages count
	fmt.Printf("Total unique pages using composables: %d\n\n", len(uniquePages))

	// Group usages by composable ID
	usagesByID := make(map[string][]*ComposableUsage)
	for _, usage := range usages {
		usagesByID[usage.ComposableID] = append(usagesByID[usage.ComposableID], usage)
	}

	// Get sorted list of composable IDs
	var ids []string
	for id := range usagesByID {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// Print usage for each composable
	for _, id := range ids {
		usageList := usagesByID[id]

		// Calculate total usage count
		totalCount := 0
		for _, usage := range usageList {
			totalCount += usage.UsageCount
		}

		fmt.Printf("Composable ID: %s\n", id)
		fmt.Printf("Total usages: %d\n", totalCount)

		// Sort usages by project/version
		sort.Slice(usageList, func(i, j int) bool {
			if usageList[i].Project != usageList[j].Project {
				return usageList[i].Project < usageList[j].Project
			}
			return usageList[i].Version < usageList[j].Version
		})

		// Print usage by project/version
		for _, usage := range usageList {
			location := usage.Project
			if usage.Version != "" {
				location += "/" + usage.Version
			}

			fmt.Printf("\n  %s: %d usages\n", location, usage.UsageCount)

			if verbose {
				// Print file paths
				for _, filePath := range usage.FilePaths {
					fmt.Printf("    - %s\n", filePath)
				}
			}
		}

		fmt.Printf("\n")
	}

	// Print composables with no usages
	unusedComposables := findUnusedComposables(composables, usagesByID)
	if len(unusedComposables) > 0 {
		fmt.Printf("Unused Composables\n")
		fmt.Printf("------------------\n\n")

		// Group by ID
		unusedByID := make(map[string][]ComposableLocation)
		for _, loc := range unusedComposables {
			unusedByID[loc.Composable.ID] = append(unusedByID[loc.Composable.ID], loc)
		}

		// Sort IDs
		var unusedIDs []string
		for id := range unusedByID {
			unusedIDs = append(unusedIDs, id)
		}
		sort.Strings(unusedIDs)

		for _, id := range unusedIDs {
			locations := unusedByID[id]
			fmt.Printf("  %s:\n", id)
			for _, loc := range locations {
				location := loc.Project
				if loc.Version != "" {
					location += "/" + loc.Version
				}
				fmt.Printf("    - %s\n", location)
			}
		}
		fmt.Printf("\n")
	}
}

// findUnusedComposables finds composables that have no usages.
func findUnusedComposables(composables []ComposableLocation, usagesByID map[string][]*ComposableUsage) []ComposableLocation {
	var unused []ComposableLocation

	for _, loc := range composables {
		// Check if this composable has any usages in the same project/version
		hasUsage := false
		if usageList, exists := usagesByID[loc.Composable.ID]; exists {
			for _, usage := range usageList {
				if usage.Project == loc.Project && usage.Version == loc.Version {
					hasUsage = true
					break
				}
			}
		}

		if !hasUsage {
			unused = append(unused, loc)
		}
	}

	return unused
}
