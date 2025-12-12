package rst

import (
	"testing"
)

// TestFetchRstspec tests fetching and parsing the canonical rstspec.toml file.
// This is an integration test that makes a real HTTP request to GitHub.
func TestFetchRstspec(t *testing.T) {
	config, err := FetchRstspec()
	if err != nil {
		t.Fatalf("Failed to fetch rstspec.toml: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	// Verify we got composables
	if len(config.Composables) == 0 {
		t.Error("Expected at least one composable in rstspec.toml")
	}

	t.Logf("Successfully fetched rstspec.toml with %d composables", len(config.Composables))
}

// TestRstspecComposablesStructure tests that the composables have the expected structure.
func TestRstspecComposablesStructure(t *testing.T) {
	config, err := FetchRstspec()
	if err != nil {
		t.Fatalf("Failed to fetch rstspec.toml: %v", err)
	}

	// Check that we have some well-known composables
	expectedComposables := map[string]bool{
		"interface":         false,
		"language":          false,
		"deployment-type":   false,
		"cluster-topology":  false,
		"cloud-provider":    false,
		"operating-system":  false,
	}

	for _, comp := range config.Composables {
		// Verify required fields are present
		if comp.ID == "" {
			t.Error("Found composable with empty ID")
		}
		if comp.Title == "" {
			t.Errorf("Composable %s has empty title", comp.ID)
		}

		// Mark expected composables as found
		if _, exists := expectedComposables[comp.ID]; exists {
			expectedComposables[comp.ID] = true
		}

		// Verify options structure
		if len(comp.Options) == 0 {
			t.Errorf("Composable %s has no options", comp.ID)
		}

		for _, opt := range comp.Options {
			if opt.ID == "" {
				t.Errorf("Composable %s has option with empty ID", comp.ID)
			}
			if opt.Title == "" {
				t.Errorf("Composable %s has option %s with empty title", comp.ID, opt.ID)
			}
		}
	}

	// Verify we found all expected composables
	for id, found := range expectedComposables {
		if !found {
			t.Errorf("Expected composable %s not found in rstspec.toml", id)
		}
	}

	t.Logf("All expected composables found with valid structure")
}

// TestRstspecInterfaceComposable tests the interface composable specifically.
func TestRstspecInterfaceComposable(t *testing.T) {
	config, err := FetchRstspec()
	if err != nil {
		t.Fatalf("Failed to fetch rstspec.toml: %v", err)
	}

	// Find the interface composable
	var interfaceComp *RstspecComposable
	for i, comp := range config.Composables {
		if comp.ID == "interface" {
			interfaceComp = &config.Composables[i]
			break
		}
	}

	if interfaceComp == nil {
		t.Fatal("Interface composable not found")
	}

	// Verify it has expected options
	expectedOptions := []string{"atlas-ui", "driver", "mongosh"}
	foundOptions := make(map[string]bool)

	for _, opt := range interfaceComp.Options {
		foundOptions[opt.ID] = true
	}

	for _, expected := range expectedOptions {
		if !foundOptions[expected] {
			t.Errorf("Expected option %s not found in interface composable", expected)
		}
	}

	t.Logf("Interface composable has %d options", len(interfaceComp.Options))
}

// TestRstspecLanguageComposable tests the language composable specifically.
func TestRstspecLanguageComposable(t *testing.T) {
	config, err := FetchRstspec()
	if err != nil {
		t.Fatalf("Failed to fetch rstspec.toml: %v", err)
	}

	// Find the language composable
	var languageComp *RstspecComposable
	for i, comp := range config.Composables {
		if comp.ID == "language" {
			languageComp = &config.Composables[i]
			break
		}
	}

	if languageComp == nil {
		t.Fatal("Language composable not found")
	}

	// Verify it has expected options (using actual IDs from rstspec.toml)
	expectedOptions := []string{"python", "nodejs", "go", "csharp", "java-sync"}
	foundOptions := make(map[string]bool)

	for _, opt := range languageComp.Options {
		foundOptions[opt.ID] = true
	}

	for _, expected := range expectedOptions {
		if !foundOptions[expected] {
			t.Errorf("Expected option %s not found in language composable", expected)
		}
	}

	t.Logf("Language composable has %d options", len(languageComp.Options))
}

