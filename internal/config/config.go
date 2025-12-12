// Package config provides configuration management for audit-cli.
//
// This package handles loading configuration from multiple sources:
//   - Config file (.audit-cli.yaml in current directory or home directory)
//   - Environment variables (AUDIT_CLI_MONOREPO_PATH)
//   - Command-line arguments (highest priority)
//
// The monorepo path is resolved in the following order (highest to lowest priority):
//  1. Command-line argument (if provided)
//  2. Environment variable AUDIT_CLI_MONOREPO_PATH
//  3. Config file .audit-cli.yaml (current directory)
//  4. Config file .audit-cli.yaml (home directory)
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the audit-cli configuration.
type Config struct {
	MonorepoPath string `yaml:"monorepo_path"`
}

// configFileName is the name of the config file.
const configFileName = ".audit-cli.yaml"

// envVarMonorepoPath is the environment variable name for monorepo path.
const envVarMonorepoPath = "AUDIT_CLI_MONOREPO_PATH"

// LoadConfig loads configuration from file and environment variables.
// Returns a Config struct with values populated from available sources.
func LoadConfig() (*Config, error) {
	config := &Config{}

	// Try to load from config file
	if err := loadFromFile(config); err != nil {
		// Config file is optional, so we don't return error if it doesn't exist
		// Only return error if file exists but can't be parsed
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variable if set
	if envPath := os.Getenv(envVarMonorepoPath); envPath != "" {
		config.MonorepoPath = envPath
	}

	return config, nil
}

// loadFromFile loads configuration from a YAML file.
// Searches in the following order:
//  1. .audit-cli.yaml in current directory
//  2. .audit-cli.yaml in home directory
func loadFromFile(config *Config) error {
	// Try current directory first
	if _, err := os.Stat(configFileName); err == nil {
		return parseConfigFile(configFileName, config)
	}

	// Try home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	homeConfigPath := filepath.Join(homeDir, configFileName)
	if _, err := os.Stat(homeConfigPath); err == nil {
		return parseConfigFile(homeConfigPath, config)
	}

	// No config file found
	return os.ErrNotExist
}

// parseConfigFile parses a YAML config file.
func parseConfigFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return nil
}

// GetMonorepoPath resolves the monorepo path from multiple sources.
// Priority order (highest to lowest):
//  1. cmdLineArg (if non-empty)
//  2. Environment variable AUDIT_CLI_MONOREPO_PATH
//  3. Config file .audit-cli.yaml
//
// Returns:
//   - string: The resolved monorepo path
//   - error: Error if no path is configured
func GetMonorepoPath(cmdLineArg string) (string, error) {
	// Command-line argument has highest priority
	if cmdLineArg != "" {
		return cmdLineArg, nil
	}

	// Load config from file and environment
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}

	// Check if we got a path from config or environment
	if config.MonorepoPath != "" {
		return config.MonorepoPath, nil
	}

	// No path configured
	return "", fmt.Errorf("no monorepo path configured\n\n" +
		"Please configure the monorepo path using one of the following methods:\n" +
		"  1. Pass as command-line argument: audit-cli <command> /path/to/monorepo\n" +
		"  2. Set environment variable: export AUDIT_CLI_MONOREPO_PATH=/path/to/monorepo\n" +
		"  3. Create config file .audit-cli.yaml with:\n" +
		"     monorepo_path: /path/to/monorepo\n\n" +
		"The config file can be placed in:\n" +
		"  - Current directory: ./.audit-cli.yaml\n" +
		"  - Home directory: ~/.audit-cli.yaml")
}

// CreateSampleConfig creates a sample config file in the current directory.
func CreateSampleConfig(monorepoPath string) error {
	config := &Config{
		MonorepoPath: monorepoPath,
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFileName, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ResolveFilePath resolves a file path that could be:
//  1. An absolute path (used as-is)
//  2. A relative path from monorepo root (if monorepo is configured)
//  3. A relative path from current directory (fallback)
//
// Priority order:
//  1. If path is absolute, return it as-is
//  2. If monorepo is configured and path exists relative to monorepo, use that
//  3. Otherwise, resolve relative to current directory
//
// Parameters:
//   - pathArg: The path argument from the command line
//
// Returns:
//   - string: The resolved absolute path
//   - error: Error if path cannot be resolved or doesn't exist
func ResolveFilePath(pathArg string) (string, error) {
	// If path is absolute, use it as-is
	if filepath.IsAbs(pathArg) {
		// Verify it exists
		if _, err := os.Stat(pathArg); err != nil {
			return "", fmt.Errorf("path does not exist: %s", pathArg)
		}
		return pathArg, nil
	}

	// Try to load config to get monorepo path
	config, err := LoadConfig()
	if err == nil && config.MonorepoPath != "" {
		// Try relative to monorepo
		monorepoRelative := filepath.Join(config.MonorepoPath, pathArg)
		if _, err := os.Stat(monorepoRelative); err == nil {
			// Path exists relative to monorepo, use it
			absPath, err := filepath.Abs(monorepoRelative)
			if err != nil {
				return "", fmt.Errorf("failed to get absolute path: %w", err)
			}
			return absPath, nil
		}
	}

	// Fallback to relative from current directory
	absPath, err := filepath.Abs(pathArg)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// Verify it exists
	if _, err := os.Stat(absPath); err != nil {
		// Provide helpful error message showing what we tried
		if config != nil && config.MonorepoPath != "" {
			return "", fmt.Errorf("path not found: %s\n\nTried:\n  - Relative to monorepo: %s\n  - Relative to current directory: %s",
				pathArg,
				filepath.Join(config.MonorepoPath, pathArg),
				absPath)
		}
		return "", fmt.Errorf("path does not exist: %s (resolved to: %s)", pathArg, absPath)
	}

	return absPath, nil
}
