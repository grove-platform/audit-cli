package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestGetMonorepoPath_CommandLineArg tests that command-line argument has highest priority.
func TestGetMonorepoPath_CommandLineArg(t *testing.T) {
	// Set environment variable
	os.Setenv(envVarMonorepoPath, "/env/path")
	defer os.Unsetenv(envVarMonorepoPath)

	// Command-line argument should override environment
	path, err := GetMonorepoPath("/cmd/path")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if path != "/cmd/path" {
		t.Errorf("Expected '/cmd/path', got '%s'", path)
	}
}

// TestGetMonorepoPath_EnvironmentVariable tests environment variable fallback.
func TestGetMonorepoPath_EnvironmentVariable(t *testing.T) {
	// Set environment variable
	os.Setenv(envVarMonorepoPath, "/env/path")
	defer os.Unsetenv(envVarMonorepoPath)

	// No command-line argument, should use environment
	path, err := GetMonorepoPath("")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if path != "/env/path" {
		t.Errorf("Expected '/env/path', got '%s'", path)
	}
}

// TestGetMonorepoPath_ConfigFile tests config file fallback.
func TestGetMonorepoPath_ConfigFile(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create config file
	configPath := filepath.Join(tempDir, configFileName)
	configContent := "monorepo_path: /config/path\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Ensure environment variable is not set
	os.Unsetenv(envVarMonorepoPath)

	// No command-line argument or environment, should use config file
	path, err := GetMonorepoPath("")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if path != "/config/path" {
		t.Errorf("Expected '/config/path', got '%s'", path)
	}
}

// TestGetMonorepoPath_NoConfig tests error when no configuration is provided.
func TestGetMonorepoPath_NoConfig(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory (no config file)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Ensure environment variable is not set
	os.Unsetenv(envVarMonorepoPath)

	// No configuration should return error
	_, err := GetMonorepoPath("")
	if err == nil {
		t.Error("Expected error when no configuration is provided")
	}
}

// TestCreateSampleConfig tests creating a sample config file.
func TestCreateSampleConfig(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create sample config
	if err := CreateSampleConfig("/sample/path"); err != nil {
		t.Fatalf("Failed to create sample config: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tempDir, configFileName)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify content
	config := &Config{}
	if err := parseConfigFile(configPath, config); err != nil {
		t.Fatalf("Failed to parse created config: %v", err)
	}

	if config.MonorepoPath != "/sample/path" {
		t.Errorf("Expected '/sample/path', got '%s'", config.MonorepoPath)
	}
}

// TestLoadConfig_InvalidYAML tests handling of invalid YAML.
func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create invalid config file
	configPath := filepath.Join(tempDir, configFileName)
	invalidContent := "monorepo_path: [invalid: yaml\n"
	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Should return error for invalid YAML
	_, err := LoadConfig()
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestResolveFilePath_AbsolutePath(t *testing.T) {
	// Create a temp file
	tempFile, err := os.CreateTemp("", "test-file-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Test with absolute path
	resolved, err := ResolveFilePath(tempFile.Name())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if resolved != tempFile.Name() {
		t.Errorf("Expected %s, got %s", tempFile.Name(), resolved)
	}
}

func TestResolveFilePath_AbsolutePath_NotExists(t *testing.T) {
	// Test with non-existent absolute path
	nonExistentPath := "/tmp/this-file-does-not-exist-12345.txt"
	_, err := ResolveFilePath(nonExistentPath)
	if err == nil {
		t.Error("Expected error for non-existent absolute path, got nil")
	}
}

func TestResolveFilePath_RelativeToMonorepo(t *testing.T) {
	// Create temp directory structure
	tempDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a file in the temp directory
	testFile := filepath.Join(tempDir, "test-file.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Change to a different directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	otherDir, err := os.MkdirTemp("", "other-dir-*")
	if err != nil {
		t.Fatalf("Failed to create other dir: %v", err)
	}
	defer os.RemoveAll(otherDir)

	if err := os.Chdir(otherDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create config file pointing to temp directory
	configContent := fmt.Sprintf("monorepo_path: %s\n", tempDir)
	if err := os.WriteFile(".audit-cli.yaml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test resolving relative path (should resolve relative to monorepo)
	resolved, err := ResolveFilePath("test-file.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedPath, _ := filepath.Abs(testFile)
	if resolved != expectedPath {
		t.Errorf("Expected %s, got %s", expectedPath, resolved)
	}
}

func TestResolveFilePath_RelativeToCurrentDir(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Resolve symlinks in temp dir (macOS uses /private/var instead of /var)
	tempDir, err = filepath.EvalSymlinks(tempDir)
	if err != nil {
		t.Fatalf("Failed to resolve symlinks: %v", err)
	}

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create a file in current directory
	testFile := "test-file.txt"
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test resolving relative path (no monorepo configured)
	resolved, err := ResolveFilePath(testFile)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedPath := filepath.Join(tempDir, testFile)
	if resolved != expectedPath {
		t.Errorf("Expected %s, got %s", expectedPath, resolved)
	}
}

func TestResolveFilePath_NotFound(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test with non-existent relative path
	_, err = ResolveFilePath("non-existent-file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestResolveFilePath_PriorityOrder(t *testing.T) {
	// Create temp directories
	monorepoDir, err := os.MkdirTemp("", "monorepo-*")
	if err != nil {
		t.Fatalf("Failed to create monorepo dir: %v", err)
	}
	defer os.RemoveAll(monorepoDir)

	currentDir, err := os.MkdirTemp("", "current-*")
	if err != nil {
		t.Fatalf("Failed to create current dir: %v", err)
	}
	defer os.RemoveAll(currentDir)

	// Create files with same name in both directories
	monorepoFile := filepath.Join(monorepoDir, "test.txt")
	if err := os.WriteFile(monorepoFile, []byte("monorepo"), 0644); err != nil {
		t.Fatalf("Failed to write monorepo file: %v", err)
	}

	currentFile := filepath.Join(currentDir, "test.txt")
	if err := os.WriteFile(currentFile, []byte("current"), 0644); err != nil {
		t.Fatalf("Failed to write current file: %v", err)
	}

	// Change to current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(currentDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create config file pointing to monorepo
	configContent := fmt.Sprintf("monorepo_path: %s\n", monorepoDir)
	if err := os.WriteFile(".audit-cli.yaml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test resolving relative path - should prefer monorepo over current dir
	resolved, err := ResolveFilePath("test.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedPath, _ := filepath.Abs(monorepoFile)
	if resolved != expectedPath {
		t.Errorf("Expected monorepo path %s, got %s", expectedPath, resolved)
	}

	// Read the file to verify it's the monorepo version
	content, err := os.ReadFile(resolved)
	if err != nil {
		t.Fatalf("Failed to read resolved file: %v", err)
	}

	if string(content) != "monorepo" {
		t.Errorf("Expected 'monorepo' content, got '%s'", string(content))
	}
}
