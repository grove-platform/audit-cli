package config

import (
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

