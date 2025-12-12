package composables

import (
	"testing"
)

// TestFetchRstspecComposables tests fetching and converting rstspec composables.
// This is an integration test that makes a real HTTP request to GitHub.
func TestFetchRstspecComposables(t *testing.T) {
	locations, err := FetchRstspecComposables()
	if err != nil {
		t.Fatalf("Failed to fetch rstspec composables: %v", err)
	}

	if len(locations) == 0 {
		t.Fatal("Expected at least one composable location")
	}

	t.Logf("Successfully fetched %d composable locations from rstspec.toml", len(locations))
}

// TestRstspecComposableLocationsStructure tests the structure of converted composables.
func TestRstspecComposableLocationsStructure(t *testing.T) {
	locations, err := FetchRstspecComposables()
	if err != nil {
		t.Fatalf("Failed to fetch rstspec composables: %v", err)
	}

	for _, loc := range locations {
		// Verify project is set to "rstspec"
		if loc.Project != "rstspec" {
			t.Errorf("Expected project 'rstspec', got '%s'", loc.Project)
		}

		// Verify version is empty
		if loc.Version != "" {
			t.Errorf("Expected empty version, got '%s'", loc.Version)
		}

		// Verify source is set to "rstspec.toml"
		if loc.Source != "rstspec.toml" {
			t.Errorf("Expected source 'rstspec.toml', got '%s'", loc.Source)
		}

		// Verify FilePath is set to the URL
		if loc.FilePath == "" {
			t.Error("Expected non-empty FilePath")
		}

		// Verify composable has required fields
		if loc.Composable.ID == "" {
			t.Error("Composable has empty ID")
		}

		if loc.Composable.Title == "" {
			t.Errorf("Composable %s has empty title", loc.Composable.ID)
		}

		if len(loc.Composable.Options) == 0 {
			t.Errorf("Composable %s has no options", loc.Composable.ID)
		}

		// Verify options are properly converted
		for _, opt := range loc.Composable.Options {
			if opt.ID == "" {
				t.Errorf("Composable %s has option with empty ID", loc.Composable.ID)
			}
			if opt.Title == "" {
				t.Errorf("Composable %s has option %s with empty title", loc.Composable.ID, opt.ID)
			}
		}
	}

	t.Logf("All %d composable locations have valid structure", len(locations))
}

// TestRstspecComposableIDs tests that expected composables are present.
func TestRstspecComposableIDs(t *testing.T) {
	locations, err := FetchRstspecComposables()
	if err != nil {
		t.Fatalf("Failed to fetch rstspec composables: %v", err)
	}

	// Check for well-known composables
	expectedIDs := map[string]bool{
		"interface":         false,
		"language":          false,
		"deployment-type":   false,
		"cluster-topology":  false,
		"cloud-provider":    false,
		"operating-system":  false,
	}

	for _, loc := range locations {
		if _, exists := expectedIDs[loc.Composable.ID]; exists {
			expectedIDs[loc.Composable.ID] = true
		}
	}

	// Verify all expected IDs were found
	for id, found := range expectedIDs {
		if !found {
			t.Errorf("Expected composable ID %s not found", id)
		}
	}

	t.Logf("All expected composable IDs found")
}

// TestRstspecInterfaceComposableConversion tests the interface composable conversion.
func TestRstspecInterfaceComposableConversion(t *testing.T) {
	locations, err := FetchRstspecComposables()
	if err != nil {
		t.Fatalf("Failed to fetch rstspec composables: %v", err)
	}

	// Find the interface composable
	var interfaceLoc *ComposableLocation
	for i, loc := range locations {
		if loc.Composable.ID == "interface" {
			interfaceLoc = &locations[i]
			break
		}
	}

	if interfaceLoc == nil {
		t.Fatal("Interface composable not found")
	}

	// Verify it has expected options
	expectedOptions := []string{"atlas-ui", "driver", "mongosh"}
	foundOptions := make(map[string]bool)

	for _, opt := range interfaceLoc.Composable.Options {
		foundOptions[opt.ID] = true
	}

	for _, expected := range expectedOptions {
		if !foundOptions[expected] {
			t.Errorf("Expected option %s not found in interface composable", expected)
		}
	}

	// Verify the composable has a title
	if interfaceLoc.Composable.Title == "" {
		t.Error("Interface composable has empty title")
	}

	t.Logf("Interface composable converted correctly with %d options", len(interfaceLoc.Composable.Options))
}

