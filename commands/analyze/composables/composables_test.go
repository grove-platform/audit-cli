// Package composables provides tests for the composables analysis functionality.
package composables

import (
	"path/filepath"
	"testing"
)

// TestFindSnootyTOMLFiles tests finding snooty.toml files in the test monorepo.
func TestFindSnootyTOMLFiles(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "composables-test")

	locations, err := FindSnootyTOMLFiles(testDataDir, "", false)
	if err != nil {
		t.Fatalf("FindSnootyTOMLFiles failed: %v", err)
	}

	// Expected: project1 (2 composables) + project2/current (2) + project2/v1.0 (2) = 6 total
	expectedTotal := 6
	if len(locations) != expectedTotal {
		t.Errorf("Expected %d composables, got %d", expectedTotal, len(locations))
	}

	// Check that we have composables from both projects
	projectCounts := make(map[string]int)
	for _, loc := range locations {
		projectCounts[loc.Project]++
	}

	if projectCounts["project1"] != 2 {
		t.Errorf("Expected 2 composables from project1, got %d", projectCounts["project1"])
	}

	if projectCounts["project2"] != 4 {
		t.Errorf("Expected 4 composables from project2, got %d", projectCounts["project2"])
	}
}

// TestFindSnootyTOMLFilesForProject tests filtering by project.
func TestFindSnootyTOMLFilesForProject(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "composables-test")

	locations, err := FindSnootyTOMLFiles(testDataDir, "project1", false)
	if err != nil {
		t.Fatalf("FindSnootyTOMLFiles failed: %v", err)
	}

	// Expected: only project1 composables (2)
	expectedTotal := 2
	if len(locations) != expectedTotal {
		t.Errorf("Expected %d composables, got %d", expectedTotal, len(locations))
	}

	// All should be from project1
	for _, loc := range locations {
		if loc.Project != "project1" {
			t.Errorf("Expected all composables from project1, got %s", loc.Project)
		}
	}
}

// TestFindSnootyTOMLFilesCurrentOnly tests filtering to current versions only.
func TestFindSnootyTOMLFilesCurrentOnly(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "composables-test")

	locations, err := FindSnootyTOMLFiles(testDataDir, "", true)
	if err != nil {
		t.Fatalf("FindSnootyTOMLFiles failed: %v", err)
	}

	// Expected: project1 (2, non-versioned) + project2/current (2) = 4 total
	// Should NOT include project2/v1.0
	expectedTotal := 4
	if len(locations) != expectedTotal {
		t.Errorf("Expected %d composables, got %d", expectedTotal, len(locations))
	}

	// Check that we don't have v1.0
	for _, loc := range locations {
		if loc.Version == "v1.0" {
			t.Errorf("Expected no v1.0 composables with --current-only, got one from %s", loc.Project)
		}
	}
}

// TestParseSnootyTOML tests parsing a snooty.toml file.
func TestParseSnootyTOML(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "composables-test")
	snootyPath := filepath.Join(testDataDir, "content", "project1", "snooty.toml")

	composables, err := ParseSnootyTOML(snootyPath)
	if err != nil {
		t.Fatalf("ParseSnootyTOML failed: %v", err)
	}

	// Expected: 2 composables (interface and language)
	expectedCount := 2
	if len(composables) != expectedCount {
		t.Errorf("Expected %d composables, got %d", expectedCount, len(composables))
	}

	// Check interface composable
	var interfaceComp *Composable
	for i := range composables {
		if composables[i].ID == "interface" {
			interfaceComp = &composables[i]
			break
		}
	}

	if interfaceComp == nil {
		t.Fatal("Expected to find 'interface' composable")
	}

	if interfaceComp.Title != "Interface" {
		t.Errorf("Expected interface title 'Interface', got '%s'", interfaceComp.Title)
	}

	if interfaceComp.Default != "driver" {
		t.Errorf("Expected interface default 'driver', got '%s'", interfaceComp.Default)
	}

	// Check options
	expectedOptions := 3 // atlas-ui, driver, mongosh
	if len(interfaceComp.Options) != expectedOptions {
		t.Errorf("Expected %d options, got %d", expectedOptions, len(interfaceComp.Options))
	}
}

// TestAnalyzeComposables tests the analysis functionality.
func TestAnalyzeComposables(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "composables-test")

	locations, err := FindSnootyTOMLFiles(testDataDir, "", false)
	if err != nil {
		t.Fatalf("FindSnootyTOMLFiles failed: %v", err)
	}

	result := AnalyzeComposables(locations)

	// Check total composables
	if len(result.AllComposables) != 6 {
		t.Errorf("Expected 6 total composables, got %d", len(result.AllComposables))
	}
}

// TestIdenticalComposables tests detection of identical composables.
func TestIdenticalComposables(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "composables-test")

	locations, err := FindSnootyTOMLFiles(testDataDir, "", false)
	if err != nil {
		t.Fatalf("FindSnootyTOMLFiles failed: %v", err)
	}

	result := AnalyzeComposables(locations)

	// Expected: "interface" composable appears 3 times identically
	// (project1, project2/current, project2/v1.0)
	if len(result.IdenticalGroups) != 1 {
		t.Errorf("Expected 1 identical group, got %d", len(result.IdenticalGroups))
	}

	if len(result.IdenticalGroups) > 0 {
		interfaceGroup := result.IdenticalGroups[0]
		if interfaceGroup.ID != "interface" {
			t.Errorf("Expected identical group ID 'interface', got '%s'", interfaceGroup.ID)
		}

		if len(interfaceGroup.Locations) != 3 {
			t.Errorf("Expected 3 locations for interface composable, got %d", len(interfaceGroup.Locations))
		}
	}
}

// TestSimilarComposables tests detection of similar composables with different IDs.
// Note: The current test data doesn't have composables with different IDs but similar options,
// so we don't expect any similar groups. This test verifies the analysis runs without error.
func TestSimilarComposables(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "composables-test")

	locations, err := FindSnootyTOMLFiles(testDataDir, "", false)
	if err != nil {
		t.Fatalf("FindSnootyTOMLFiles failed: %v", err)
	}

	result := AnalyzeComposables(locations)

	// With current test data, we don't expect similar groups
	// (no composables with different IDs but similar option sets)
	if len(result.SimilarGroups) != 0 {
		t.Errorf("Expected 0 similar groups, got %d", len(result.SimilarGroups))
	}

	// Verify we still have the expected identical groups
	if len(result.IdenticalGroups) != 1 {
		t.Errorf("Expected 1 identical group, got %d", len(result.IdenticalGroups))
	}

	if len(result.IdenticalGroups) > 0 {
		interfaceGroup := result.IdenticalGroups[0]
		if interfaceGroup.ID != "interface" {
			t.Errorf("Expected identical group ID 'interface', got '%s'", interfaceGroup.ID)
		}

		if len(interfaceGroup.Locations) != 3 {
			t.Errorf("Expected 3 locations for interface composable, got %d", len(interfaceGroup.Locations))
		}
	}
}

// TestCalculateOptionSimilarity tests the Jaccard similarity calculation.
func TestCalculateOptionSimilarity(t *testing.T) {
	// Test identical option sets
	comp1 := Composable{
		ID:    "test1",
		Title: "Test 1",
		Options: []ComposableOption{
			{ID: "a", Title: "A"},
			{ID: "b", Title: "B"},
			{ID: "c", Title: "C"},
		},
	}

	comp2 := Composable{
		ID:    "test2",
		Title: "Test 2",
		Options: []ComposableOption{
			{ID: "a", Title: "A"},
			{ID: "b", Title: "B"},
			{ID: "c", Title: "C"},
		},
	}

	similarity := calculateOptionSimilarity(comp1, comp2)
	if similarity != 1.0 {
		t.Errorf("Expected similarity 1.0 for identical options, got %f", similarity)
	}

	// Test partial overlap
	comp3 := Composable{
		ID:    "test3",
		Title: "Test 3",
		Options: []ComposableOption{
			{ID: "a", Title: "A"},
			{ID: "b", Title: "B"},
		},
	}

	// comp1 has {a, b, c}, comp3 has {a, b}
	// intersection = 2, union = 3, similarity = 2/3 = 0.667
	similarity = calculateOptionSimilarity(comp1, comp3)
	expected := 2.0 / 3.0
	tolerance := 0.01
	if similarity < expected-tolerance || similarity > expected+tolerance {
		t.Errorf("Expected similarity %.3f, got %.3f", expected, similarity)
	}

	// Test no overlap
	comp4 := Composable{
		ID:    "test4",
		Title: "Test 4",
		Options: []ComposableOption{
			{ID: "x", Title: "X"},
			{ID: "y", Title: "Y"},
		},
	}

	similarity = calculateOptionSimilarity(comp1, comp4)
	if similarity != 0.0 {
		t.Errorf("Expected similarity 0.0 for no overlap, got %f", similarity)
	}
}

// TestComposablesEqual tests the composable equality function.
func TestComposablesEqual(t *testing.T) {
	comp1 := Composable{
		ID:      "test",
		Title:   "Test",
		Default: "option1",
		Options: []ComposableOption{
			{ID: "option1", Title: "Option 1"},
			{ID: "option2", Title: "Option 2"},
		},
	}

	comp2 := Composable{
		ID:      "test",
		Title:   "Test",
		Default: "option1",
		Options: []ComposableOption{
			{ID: "option1", Title: "Option 1"},
			{ID: "option2", Title: "Option 2"},
		},
	}

	comp3 := Composable{
		ID:      "test",
		Title:   "Test",
		Default: "option1",
		Options: []ComposableOption{
			{ID: "option1", Title: "Option 1"},
			{ID: "option3", Title: "Option 3"},
		},
	}

	if !composablesEqual(comp1, comp2) {
		t.Error("Expected comp1 and comp2 to be equal")
	}

	if composablesEqual(comp1, comp3) {
		t.Error("Expected comp1 and comp3 to be different")
	}
}

// TestExtractProjectAndVersion tests the project and version extraction.
func TestExtractProjectAndVersion(t *testing.T) {
	tests := []struct {
		path            string
		expectedProject string
		expectedVersion string
	}{
		{
			path:            "project1/snooty.toml",
			expectedProject: "project1",
			expectedVersion: "",
		},
		{
			path:            "project2/v1.0/snooty.toml",
			expectedProject: "project2",
			expectedVersion: "v1.0",
		},
		{
			path:            "project2/current/snooty.toml",
			expectedProject: "project2",
			expectedVersion: "current",
		},
	}

	for _, tt := range tests {
		project, version := extractProjectAndVersion(tt.path)
		if project != tt.expectedProject {
			t.Errorf("For path %s, expected project '%s', got '%s'", tt.path, tt.expectedProject, project)
		}
		if version != tt.expectedVersion {
			t.Errorf("For path %s, expected version '%s', got '%s'", tt.path, tt.expectedVersion, version)
		}
	}
}

// TestFormatOptionsAsBullets tests the bullet formatting function.
func TestFormatOptionsAsBullets(t *testing.T) {
	options := []ComposableOption{
		{ID: "option1", Title: "Option 1"},
		{ID: "option2", Title: "Option 2"},
		{ID: "option3", Title: "Option 3"},
	}

	lines := formatOptionsAsBullets(options)

	expectedCount := 3
	if len(lines) != expectedCount {
		t.Errorf("Expected %d lines, got %d", expectedCount, len(lines))
	}

	// Check format of first line
	expected := "• option1: Option 1"
	if lines[0] != expected {
		t.Errorf("Expected first line '%s', got '%s'", expected, lines[0])
	}
}

// TestFormatOptions tests the comma-separated formatting function.
func TestFormatOptions(t *testing.T) {
	options := []ComposableOption{
		{ID: "option1", Title: "Option 1"},
		{ID: "option2", Title: "Option 2"},
		{ID: "option3", Title: "Option 3"},
	}

	result := formatOptions(options)

	expected := "option1, option2, option3"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

