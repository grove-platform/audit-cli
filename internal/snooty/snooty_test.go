package snooty

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	// Create a temporary snooty.toml file
	tempDir := t.TempDir()
	snootyPath := filepath.Join(tempDir, "snooty.toml")

	content := `
name = "test-project"
title = "Test Project"

[[composables]]
id = "language"
title = "Language"
default = "python"

[[composables.options]]
id = "python"
title = "Python"

[[composables.options]]
id = "javascript"
title = "JavaScript"

[[composables]]
id = "interface"
title = "Interface"
default = "atlas"

[[composables.options]]
id = "atlas"
title = "Atlas"

[[composables.options]]
id = "mongosh"
title = "MongoDB Shell"
`
	if err := os.WriteFile(snootyPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	config, err := ParseFile(snootyPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if config.Name != "test-project" {
		t.Errorf("config.Name = %q, want %q", config.Name, "test-project")
	}

	if config.Title != "Test Project" {
		t.Errorf("config.Title = %q, want %q", config.Title, "Test Project")
	}

	if len(config.Composables) != 2 {
		t.Fatalf("len(config.Composables) = %d, want 2", len(config.Composables))
	}

	// Check first composable
	lang := config.Composables[0]
	if lang.ID != "language" {
		t.Errorf("Composables[0].ID = %q, want %q", lang.ID, "language")
	}
	if len(lang.Options) != 2 {
		t.Errorf("len(Composables[0].Options) = %d, want 2", len(lang.Options))
	}
	if lang.Options[0].ID != "python" || lang.Options[0].Title != "Python" {
		t.Errorf("Composables[0].Options[0] = {%q, %q}, want {python, Python}",
			lang.Options[0].ID, lang.Options[0].Title)
	}
}

func TestParseFile_InvalidFile(t *testing.T) {
	_, err := ParseFile("/nonexistent/path/snooty.toml")
	if err == nil {
		t.Error("ParseFile() expected error for nonexistent file, got nil")
	}
}

func TestParseFile_InvalidTOML(t *testing.T) {
	tempDir := t.TempDir()
	snootyPath := filepath.Join(tempDir, "snooty.toml")

	// Write invalid TOML
	if err := os.WriteFile(snootyPath, []byte("invalid = [toml"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := ParseFile(snootyPath)
	if err == nil {
		t.Error("ParseFile() expected error for invalid TOML, got nil")
	}
}

func TestFindProjectSnootyTOML(t *testing.T) {
	// Create a mock project structure
	tempDir := t.TempDir()

	// Create: content/atlas/snooty.toml
	atlasDir := filepath.Join(tempDir, "content", "atlas")
	if err := os.MkdirAll(atlasDir, 0755); err != nil {
		t.Fatalf("Failed to create atlas dir: %v", err)
	}
	atlasSnootyPath := filepath.Join(atlasDir, "snooty.toml")
	if err := os.WriteFile(atlasSnootyPath, []byte("name = \"atlas\""), 0644); err != nil {
		t.Fatalf("Failed to write snooty.toml: %v", err)
	}

	// Create: content/atlas/source/getting-started.txt
	atlasSourceDir := filepath.Join(atlasDir, "source")
	if err := os.MkdirAll(atlasSourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	sourceFile := filepath.Join(atlasSourceDir, "getting-started.txt")
	if err := os.WriteFile(sourceFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Test finding snooty.toml from source file
	found, err := FindProjectSnootyTOML(sourceFile)
	if err != nil {
		t.Fatalf("FindProjectSnootyTOML() error = %v", err)
	}
	if found != atlasSnootyPath {
		t.Errorf("FindProjectSnootyTOML() = %q, want %q", found, atlasSnootyPath)
	}
}

func TestFindProjectSnootyTOML_VersionedProject(t *testing.T) {
	// Create a mock versioned project structure
	tempDir := t.TempDir()

	// Create: content/manual/v8.0/snooty.toml
	versionDir := filepath.Join(tempDir, "content", "manual", "v8.0")
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		t.Fatalf("Failed to create version dir: %v", err)
	}
	snootyPath := filepath.Join(versionDir, "snooty.toml")
	if err := os.WriteFile(snootyPath, []byte("name = \"manual\""), 0644); err != nil {
		t.Fatalf("Failed to write snooty.toml: %v", err)
	}

	// Create: content/manual/v8.0/source/tutorial/install.txt
	sourceDir := filepath.Join(versionDir, "source", "tutorial")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	sourceFile := filepath.Join(sourceDir, "install.txt")
	if err := os.WriteFile(sourceFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Test finding snooty.toml from nested source file
	found, err := FindProjectSnootyTOML(sourceFile)
	if err != nil {
		t.Fatalf("FindProjectSnootyTOML() error = %v", err)
	}
	if found != snootyPath {
		t.Errorf("FindProjectSnootyTOML() = %q, want %q", found, snootyPath)
	}
}

func TestFindProjectSnootyTOML_NotFound(t *testing.T) {
	// Create a directory structure without snooty.toml
	tempDir := t.TempDir()
	contentDir := filepath.Join(tempDir, "content", "project", "source")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}
	sourceFile := filepath.Join(contentDir, "test.txt")
	if err := os.WriteFile(sourceFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	found, err := FindProjectSnootyTOML(sourceFile)
	if err != nil {
		t.Fatalf("FindProjectSnootyTOML() error = %v", err)
	}
	if found != "" {
		t.Errorf("FindProjectSnootyTOML() = %q, want empty string", found)
	}
}

func TestBuildComposableIDToTitleMap(t *testing.T) {
	composables := []Composable{
		{
			ID:    "language",
			Title: "Language",
			Options: []ComposableOption{
				{ID: "python", Title: "Python"},
				{ID: "javascript", Title: "JavaScript"},
				{ID: "go", Title: "Go"},
			},
		},
		{
			ID:    "interface",
			Title: "Interface",
			Options: []ComposableOption{
				{ID: "atlas", Title: "Atlas"},
				{ID: "mongosh", Title: "MongoDB Shell"},
			},
		},
	}

	// Test extracting language composable
	langMap := BuildComposableIDToTitleMap(composables, "language")
	if len(langMap) != 3 {
		t.Errorf("len(langMap) = %d, want 3", len(langMap))
	}
	if langMap["python"] != "Python" {
		t.Errorf("langMap[python] = %q, want %q", langMap["python"], "Python")
	}
	if langMap["javascript"] != "JavaScript" {
		t.Errorf("langMap[javascript] = %q, want %q", langMap["javascript"], "JavaScript")
	}

	// Test extracting interface composable
	ifaceMap := BuildComposableIDToTitleMap(composables, "interface")
	if len(ifaceMap) != 2 {
		t.Errorf("len(ifaceMap) = %d, want 2", len(ifaceMap))
	}
	if ifaceMap["mongosh"] != "MongoDB Shell" {
		t.Errorf("ifaceMap[mongosh] = %q, want %q", ifaceMap["mongosh"], "MongoDB Shell")
	}

	// Test non-existent composable
	emptyMap := BuildComposableIDToTitleMap(composables, "nonexistent")
	if len(emptyMap) != 0 {
		t.Errorf("len(emptyMap) = %d, want 0", len(emptyMap))
	}
}

func TestExtractProjectAndVersion(t *testing.T) {
	tests := []struct {
		name        string
		relPath     string
		wantProject string
		wantVersion string
	}{
		{
			name:        "versioned project",
			relPath:     "manual/v8.0/snooty.toml",
			wantProject: "manual",
			wantVersion: "v8.0",
		},
		{
			name:        "non-versioned project",
			relPath:     "atlas/snooty.toml",
			wantProject: "atlas",
			wantVersion: "",
		},
		{
			name:        "current version",
			relPath:     "node/current/snooty.toml",
			wantProject: "node",
			wantVersion: "current",
		},
		{
			name:        "too short path",
			relPath:     "snooty.toml",
			wantProject: "",
			wantVersion: "",
		},
		{
			name:        "not snooty.toml",
			relPath:     "manual/v8.0/source/index.txt",
			wantProject: "",
			wantVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProject, gotVersion := ExtractProjectAndVersion(tt.relPath)
			if gotProject != tt.wantProject {
				t.Errorf("ExtractProjectAndVersion(%q) project = %q, want %q",
					tt.relPath, gotProject, tt.wantProject)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("ExtractProjectAndVersion(%q) version = %q, want %q",
					tt.relPath, gotVersion, tt.wantVersion)
			}
		})
	}
}

func TestIsCurrentVersion(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"current", true},
		{"manual", true},
		{"v8.0", false},
		{"v7.0", false},
		{"master", false},
		{"upcoming", false},
		{"latest", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := IsCurrentVersion(tt.version)
			if got != tt.want {
				t.Errorf("IsCurrentVersion(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

